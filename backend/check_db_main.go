package main

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/glebarez/sqlite"
)

type PrayerCache struct {
	ID        uint      `gorm:"primaryKey"`
	Lat       float64   `gorm:"column:lat"`
	Lon       float64   `gorm:"column:lon"`
	Date      string    `gorm:"column:date"`
	Method    int       `gorm:"column:method"`
	Fajr      string    `gorm:"column:fajr"`
	Sunrise   string    `gorm:"column:sunrise"`
	Dhuhr     string    `gorm:"column:dhuhr"`
	Asr       string    `gorm:"column:asr"`
	Maghrib   string    `gorm:"column:maghrib"`
	Isha      string    `gorm:"column:isha"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (PrayerCache) TableName() string {
	return "prayer_cache"
}

func main() {
	db, err := gorm.Open(sqlite.Open("data/navisha.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	var caches []PrayerCache
	db.Find(&caches)
	for _, c := range caches {
		fmt.Printf("ID=%d Lat=%f Lon=%f Date=%s Method=%d CreatedAt=%v\n", c.ID, c.Lat, c.Lon, c.Date, c.Method, c.CreatedAt)
	}
}
