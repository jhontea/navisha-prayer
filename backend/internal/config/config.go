package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Server    ServerConfig
	Aladhan   AladhanConfig
	Geocoding GeocodingConfig
	Database  DatabaseConfig
	Defaults  DefaultsConfig
}

type ServerConfig struct {
	Port         string // default ":8080"
	Environment  string // "development" | "production"
	AllowOrigins string // CORS allowed origins, e.g. "http://localhost:3000"
}

type AladhanConfig struct {
	BaseURL string // "https://api.aladhan.com/v1"
	APIKey  string // optional; Aladhan free tier may not require one
}

type GeocodingConfig struct {
	BaseURL string // "https://geocoding-api.open-meteo.com/v1"
}

type DatabaseConfig struct {
	Path string // default "./data/navisha.db"
}

type DefaultsConfig struct {
	Lat      float64
	Lon      float64
	Method   int
	Timezone string // IANA timezone, e.g. "Asia/Jakarta"
}

// Load loads configuration from environment variables with sensible defaults.
func Load() *Config {
	// Try to load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("NAVISHA_SERVER_PORT", ":8080"),
			Environment:  getEnv("NAVISHA_SERVER_ENV", "development"),
			AllowOrigins: getEnv("NAVISHA_SERVER_ALLOW_ORIGINS", "http://localhost:3000"),
		},
		Aladhan: AladhanConfig{
			BaseURL: getEnv("NAVISHA_ALADHAN_BASE_URL", "https://api.aladhan.com/v1"),
			APIKey:  getEnv("NAVISHA_ALADHAN_API_KEY", ""),
		},
		Geocoding: GeocodingConfig{
			BaseURL: getEnv("NAVISHA_GEOCODING_BASE_URL", "https://geocoding-api.open-meteo.com/v1"),
		},
		Database: DatabaseConfig{
			Path: getEnv("NAVISHA_DATABASE_PATH", "./data/navisha.db"),
		},
		Defaults: DefaultsConfig{
			Lat:      -6.2088,
			Lon:      106.8456,
			Method:   3,
			Timezone: "Asia/Jakarta",
		},
	}

	// Parse default lat/lon/method from env if present
	if lat := getEnv("NAVISHA_DEFAULT_LAT", ""); lat != "" {
		if v, err := parseFloat(lat); err == nil {
			cfg.Defaults.Lat = v
		} else {
			log.Printf("Warning: invalid NAVISHA_DEFAULT_LAT: %s", lat)
		}
	}
	if lon := getEnv("NAVISHA_DEFAULT_LON", ""); lon != "" {
		if v, err := parseFloat(lon); err == nil {
			cfg.Defaults.Lon = v
		} else {
			log.Printf("Warning: invalid NAVISHA_DEFAULT_LON: %s", lon)
		}
	}
	if method := getEnv("NAVISHA_DEFAULT_METHOD", ""); method != "" {
		if v, err := parseInt(method); err == nil {
			cfg.Defaults.Method = v
		} else {
			log.Printf("Warning: invalid NAVISHA_DEFAULT_METHOD: %s", method)
		}
	}
	if tz := getEnv("NAVISHA_DEFAULT_TIMEZONE", ""); tz != "" {
		cfg.Defaults.Timezone = tz
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func parseFloat(s string) (float64, error) {
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}

func parseInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
