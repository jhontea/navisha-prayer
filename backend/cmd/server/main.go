package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"navisha-prayer/internal/client"
	"navisha-prayer/internal/config"
	"navisha-prayer/internal/database"
	"navisha-prayer/internal/handler"
	"navisha-prayer/internal/middleware"
	"navisha-prayer/internal/repository"
	"navisha-prayer/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Extract *sql.DB for graceful close on shutdown
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to extract sql.DB: %v", err)
	}

	// Initialize repositories
	prayerCacheRepo := repository.NewPrayerCacheRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	fastingRepo := repository.NewFastingRepository(db)

	// Initialize API clients
	aladhanClient := client.NewAladhanAPIClient(cfg.Aladhan.BaseURL, cfg.Aladhan.APIKey)
	geocodingClient := client.NewOpenMeteoGeocodingClient(cfg.Geocoding.BaseURL)

	// Initialize services
	prayerService := service.NewPrayerService(aladhanClient, prayerCacheRepo, settingsRepo, cfg)
	fastingService := service.NewFastingService(fastingRepo, prayerService, cfg)
	eidService := service.NewEidService(aladhanClient, settingsRepo, cfg)
	locationService := service.NewLocationService(geocodingClient, settingsRepo, cfg.Defaults)

	// Initialize handlers
	prayerHandler := handler.NewPrayerHandler(prayerService)
	fastingHandler := handler.NewFastingHandler(fastingService)
	eidHandler := handler.NewEidHandler(eidService)
	locationHandler := handler.NewLocationHandler(locationService)
	healthHandler := handler.NewHealthHandler()

	// Setup Gin router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.ErrorRecovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.RequestTimer())
	router.Use(middleware.CORS(cfg.Server.AllowOrigins))

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", healthHandler.GetHealth)

		// Prayer times
		prayer := api.Group("/prayer-times")
		{
			prayer.GET("/today", prayerHandler.GetTodayPrayerTimes)
			prayer.GET("/date/:date", prayerHandler.GetPrayerTimesByDate)
			prayer.GET("/monthly", prayerHandler.GetMonthlyPrayerTimes)
			prayer.GET("/next", prayerHandler.GetNextPrayer)
		}

		// Fasting
		fasting := api.Group("/fasting")
		{
			fasting.GET("/status", fastingHandler.GetFastingStatus)
			fasting.POST("/type", fastingHandler.UpdateFastingType)
			fasting.GET("/schedule", fastingHandler.GetFastingSchedule)
		}

		// Eid
		eid := api.Group("/eid")
		{
			eid.GET("/countdown", eidHandler.GetEidCountdown)
		}

		// Location config
		location := api.Group("/config/location")
		{
			location.GET("", locationHandler.GetLocation)
			location.POST("/gps", locationHandler.SetLocationGPS)
			location.POST("/city", locationHandler.SetLocationCity)
			location.GET("/search", locationHandler.SearchCities)
		}
	}

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: router,
	}

	go func() {
		log.Printf("🕌 Navisha Prayer API starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close database connection to flush writes and release file lock
	if err := sqlDB.Close(); err != nil {
		log.Printf("Warning: failed to close database: %v", err)
	}

	log.Println("Server exited gracefully")
}
