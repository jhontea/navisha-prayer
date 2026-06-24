package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"navisha-prayer/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Init initializes the SQLite database with GORM auto-migration.
func Init(dbPath string) (*gorm.DB, error) {
	// Ensure the data directory exists
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&repository.PrayerCache{},
		&repository.UserSettings{},
		&repository.FastingSchedule{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Seed default settings
	seedDefaults(db)

	log.Printf("Database initialized at %s", dbPath)
	return db, nil
}

func seedDefaults(db *gorm.DB) {
	defaults := []repository.UserSettings{
		{Key: "location.lat", Value: "-6.2088"},
		{Key: "location.lon", Value: "106.8456"},
		{Key: "location.city", Value: "Jakarta"},
		{Key: "location.country", Value: "ID"},
		{Key: "location.source", Value: "default"},
		{Key: "calculation.method", Value: "3"},
	}

	for _, d := range defaults {
		db.Where(repository.UserSettings{Key: d.Key}).
			Assign(repository.UserSettings{Value: d.Value}).
			FirstOrCreate(&d)
	}

	fastingTypes := []string{"ramadan", "syawal", "sawm"}
	for _, t := range fastingTypes {
		db.Where(repository.FastingSchedule{Type: t}).
			FirstOrCreate(&repository.FastingSchedule{Type: t, Enabled: false})
	}
}
