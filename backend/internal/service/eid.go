package service

import (
	"context"
	"fmt"
	"time"

	"navisha-prayer/internal/client"
	"navisha-prayer/internal/config"
	"navisha-prayer/internal/repository"
)

// EidService defines Eid al-Fitr countdown business logic.
type EidService interface {
	GetEidAlFitrCountdown(ctx context.Context) (*EidCountdownResponse, error)
}

// EidCountdownResponse contains Eid al-Fitr countdown information.
type EidCountdownResponse struct {
	IsToday       bool   `json:"is_islamic_today"`
	HijriDate     string `json:"hijri_date"`
	EidDate       string `json:"eid_date"`       // ISO 8601
	EidDateHijri  string `json:"eid_date_hijri"` // e.g. "1-10-1446"
	DaysRemaining int    `json:"days_remaining"`
	Message       string `json:"message"` // e.g. "Eid Mubarak!" or countdown text
}

// eidService implements EidService.
type eidService struct {
	client       client.AladhanClient
	settingsRepo *repository.SettingsRepository
	cfg          *config.Config
}

func NewEidService(client client.AladhanClient, settingsRepo *repository.SettingsRepository, cfg *config.Config) EidService {
	return &eidService{
		client:       client,
		settingsRepo: settingsRepo,
		cfg:          cfg,
	}
}

// loc returns the configured *time.Location, falling back to UTC.
func (s *eidService) loc() *time.Location {
	if s.cfg != nil && s.cfg.Defaults.Timezone != "" {
		if l, err := time.LoadLocation(s.cfg.Defaults.Timezone); err == nil {
			return l
		}
	}
	return time.UTC
}

// location reads the user's configured lat/lon, falling back to defaults.
func (s *eidService) location() (float64, float64) {
	lat := s.cfg.Defaults.Lat
	lon := s.cfg.Defaults.Lon
	if s.settingsRepo == nil {
		return lat, lon
	}
	if v, err := s.settingsRepo.Get("location.lat"); err == nil && v != "" {
		var f float64
		if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
			lat = f
		}
	}
	if v, err := s.settingsRepo.Get("location.lon"); err == nil && v != "" {
		var f float64
		if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
			lon = f
		}
	}
	return lat, lon
}

// GetEidAlFitrCountdown calculates the countdown to the next Eid al-Fitr.
// Eid al-Fitr falls on 1 Shawwal (10th month of Hijri calendar).
func (s *eidService) GetEidAlFitrCountdown(ctx context.Context) (*EidCountdownResponse, error) {
	now := time.Now().In(s.loc())
	lat, lon := s.location()

	// Fetch Hijri calendar data to find the next 1 Shawwal
	// We check a range of months to find the next Eid
	hijriDate := ""
	var eidDate time.Time
	found := false

	// Check current month and next 12 months (up to a full year ahead)
	for monthOffset := 0; monthOffset < 12; monthOffset++ {
		targetDate := now.AddDate(0, monthOffset, 0)
		month := int(targetDate.Month())
		year := targetDate.Year()

		resp, err := s.client.GetHijriCalendar(ctx, lat, lon, month, year)
		if err != nil {
			continue
		}

		for _, day := range resp.Data {
			if day.Hijri.Month.Number == 10 && day.Hijri.Date == "01" {
				// Parse the Gregorian date from the API response
				eidDateParsed, err := time.Parse("2006-01-02", day.Gregorian.Date)
				if err == nil && !eidDateParsed.Before(now.Truncate(24*time.Hour)) {
					hijriDate = day.Hijri.Date
					eidDate = eidDateParsed
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}

	if !found {
		// Fallback: estimate next Eid al-Fitr based on current Hijri year
		// The Hijri year is approximately 354 days, so Eid shifts ~11 days earlier each Gregorian year
		eidDate = time.Date(now.Year(), 3, 1, 0, 0, 0, 0, now.Location())
		if eidDate.Before(now) {
			eidDate = time.Date(now.Year()+1, 3, 1, 0, 0, 0, 0, now.Location())
		}
		hijriDate = "01-10"
	}

	daysRemaining := int(eidDate.Sub(now).Hours() / 24)
	isToday := daysRemaining == 0 && now.Day() == eidDate.Day()

	message := ""
	if isToday {
		message = "Eid Mubarak!"
	} else if daysRemaining > 0 {
		message = fmt.Sprintf("%d hari menuju Idul Fitri", daysRemaining)
	} else {
		message = fmt.Sprintf("Idul Fitri telah lewat %d hari yang lalu", -daysRemaining)
	}

	return &EidCountdownResponse{
		IsToday:       isToday,
		HijriDate:     hijriDate,
		EidDate:       eidDate.Format(time.RFC3339),
		EidDateHijri:  fmt.Sprintf("1-10-%d", getHijriYear(eidDate)),
		DaysRemaining: daysRemaining,
		Message:       message,
	}, nil
}

// getHijriYear estimates the Hijri year from a Gregorian date.
func getHijriYear(gregorianDate time.Time) int {
	// Approximate: Hijri year = (Gregorian year - 622) * 1.030684
	hijriYear := float64(gregorianDate.Year()-622) * 1.030684
	return int(hijriYear) + 1
}
