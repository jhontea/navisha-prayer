package repository

import (
	"time"
)

// PrayerCache stores fetched prayer times to avoid repeated API calls.
type PrayerCache struct {
	ID          uint      `gorm:"primaryKey"`
	Lat         float64   `gorm:"not null"`
	Lon         float64   `gorm:"not null"`
	Date        string    `gorm:"not null;index"` // "DD-MM-YYYY"
	Method      int       `gorm:"not null"`
	Fajr        string    `gorm:"not null"`
	Sunrise     string    `gorm:"not null"`
	Dhuhr       string    `gorm:"not null"`
	Asr         string    `gorm:"not null"`
	Maghrib     string    `gorm:"not null"`
	Isha        string    `gorm:"not null"`
	RawResponse string    `gorm:"type:text"` // raw JSON for debugging
	CreatedAt   time.Time `gorm:"autoCreatedAt"`
}

func (PrayerCache) TableName() string { return "prayer_cache" }

// UserSettings persists user preferences.
type UserSettings struct {
	ID        uint      `gorm:"primaryKey"`
	Key       string    `gorm:"uniqueIndex;not null"`
	Value     string    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"autoUpdatedAt"`
}

func (UserSettings) TableName() string { return "user_settings" }

// FastingSchedule stores fasting preferences.
type FastingSchedule struct {
	ID        uint      `gorm:"primaryKey"`
	Type      string    `gorm:"not null"` // "ramadan" | "syawal" | "sawm" | "none"
	Enabled   bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreatedAt"`
	UpdatedAt time.Time `gorm:"autoUpdatedAt"`
}

func (FastingSchedule) TableName() string { return "fasting_schedule" }
