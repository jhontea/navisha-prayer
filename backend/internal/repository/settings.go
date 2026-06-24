package repository

import (
	"time"

	"gorm.io/gorm"
)

// DB is the database connection instance.
var DB *gorm.DB

// PrayerCacheRepository handles prayer cache database operations.
type PrayerCacheRepository struct {
	db *gorm.DB
}

func NewPrayerCacheRepository(db *gorm.DB) *PrayerCacheRepository {
	return &PrayerCacheRepository{db: db}
}

// GetByLocationAndDate retrieves cached prayer times for a location and date.
func (r *PrayerCacheRepository) GetByLocationAndDate(lat, lon float64, date string, method int) (*PrayerCache, error) {
	var cache PrayerCache
	err := r.db.Where("lat = ? AND lon = ? AND date = ? AND method = ?", lat, lon, date, method).First(&cache).Error
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

// Create stores a new prayer cache entry.
func (r *PrayerCacheRepository) Create(cache *PrayerCache) error {
	return r.db.Create(cache).Error
}

// Upsert creates or updates a prayer cache entry.
func (r *PrayerCacheRepository) Upsert(cache *PrayerCache) error {
	return r.db.Where("lat = ? AND lon = ? AND date = ? AND method = ?", cache.Lat, cache.Lon, cache.Date, cache.Method).
		Assign(cache).
		FirstOrCreate(cache).Error
}

// IsCacheValid checks if a cache entry exists and is still valid (within 1 hour).
func (r *PrayerCacheRepository) IsCacheValid(lat, lon float64, date string, method int) (bool, *PrayerCache) {
	cache, err := r.GetByLocationAndDate(lat, lon, date, method)
	if err != nil {
		return false, nil
	}
	// Cache is valid if created within the last hour
	if time.Since(cache.CreatedAt) < time.Hour {
		return true, cache
	}
	return false, cache
}

// SettingsRepository handles user settings database operations.
type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get retrieves a setting by key.
func (r *SettingsRepository) Get(key string) (string, error) {
	var setting UserSettings
	err := r.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

// Set creates or updates a setting.
func (r *SettingsRepository) Set(key, value string) error {
	return r.db.Where(UserSettings{Key: key}).
		Assign(UserSettings{Value: value}).
		FirstOrCreate(&UserSettings{Key: key, Value: value}).Error
}

// GetAll retrieves all settings as a map.
func (r *SettingsRepository) GetAll() (map[string]string, error) {
	var settings []UserSettings
	err := r.db.Find(&settings).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

// FastingRepository handles fasting schedule database operations.
type FastingRepository struct {
	db *gorm.DB
}

func NewFastingRepository(db *gorm.DB) *FastingRepository {
	return &FastingRepository{db: db}
}

// GetEnabled returns the currently enabled fasting type.
func (r *FastingRepository) GetEnabled() (*FastingSchedule, error) {
	var schedule FastingSchedule
	err := r.db.Where("enabled = ?", true).First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// GetAll returns all fasting schedules.
func (r *FastingRepository) GetAll() ([]FastingSchedule, error) {
	var schedules []FastingSchedule
	err := r.db.Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

// UpdateEnabled enables a fasting type and disables all others atomically.
func (r *FastingRepository) UpdateEnabled(fastingType string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Disable all currently enabled types
		if err := tx.Model(&FastingSchedule{}).Where("enabled = ?", true).Update("enabled", false).Error; err != nil {
			return err
		}
		// Enable the specified type
		return tx.Model(&FastingSchedule{}).Where("type = ?", fastingType).Update("enabled", true).Error
	})
}

// DisableAll disables all fasting types.
func (r *FastingRepository) DisableAll() error {
	return r.db.Model(&FastingSchedule{}).Where("enabled = ?", true).Update("enabled", false).Error
}
