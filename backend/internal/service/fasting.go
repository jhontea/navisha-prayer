package service

import (
	"context"
	"fmt"
	"time"

	"navisha-prayer/internal/config"
	"navisha-prayer/internal/repository"
)

// FastingService defines fasting-related business logic.
type FastingService interface {
	GetCurrentFasting(ctx context.Context) (*FastingResponse, error)
	UpdateFastingType(ctx context.Context, fastingType string) error
}

// FastingResponse contains countdown info for Suhoor and Iftar.
type FastingResponse struct {
	IsFasting     bool   `json:"is_fasting"`
	FastingType   string `json:"fasting_type"`   // "ramadan" | "syawal" | "sawm" | "none"
	SuhoorTime    string `json:"suhoor_time"`    // "HH:MM"
	IftarTime     string `json:"iftar_time"`     // "HH:MM"
	SuhoorIn      string `json:"suhoor_in"`      // e.g. "2h 15m"
	IftarIn       string `json:"iftar_in"`       // e.g. "5h 30m"
	SuhoorSeconds int    `json:"suhoor_seconds"` // seconds until suhoor
	IftarSeconds  int    `json:"iftar_seconds"`  // seconds until iftar
}

// fastingService implements FastingService.
type fastingService struct {
	repo      *repository.FastingRepository
	prayerSvc PrayerService
	cfg       *config.Config
}

func NewFastingService(repo *repository.FastingRepository, prayerSvc PrayerService, cfg *config.Config) FastingService {
	return &fastingService{
		repo:      repo,
		prayerSvc: prayerSvc,
		cfg:       cfg,
	}
}

// loc returns the configured *time.Location, falling back to UTC.
func (s *fastingService) loc() *time.Location {
	if s.cfg != nil && s.cfg.Defaults.Timezone != "" {
		if l, err := time.LoadLocation(s.cfg.Defaults.Timezone); err == nil {
			return l
		}
	}
	return time.UTC
}

// GetCurrentFasting returns the current fasting status and countdown.
func (s *fastingService) GetCurrentFasting(ctx context.Context) (*FastingResponse, error) {
	schedule, err := s.repo.GetEnabled()
	if err != nil {
		// No fasting enabled
		return &FastingResponse{
			IsFasting:   false,
			FastingType: "none",
		}, nil
	}

	// Get today's prayer times for Suhoor (Fajr) and Iftar (Maghrib)
	prayerTimes, err := s.prayerSvc.GetTodayPrayerTimes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get prayer times for fasting: %w", err)
	}

	suhoorTime := ""
	iftarTime := ""
	for _, p := range prayerTimes.Prayers {
		if p.Name == "Fajr" {
			suhoorTime = p.Time
		}
		if p.Name == "Maghrib" {
			iftarTime = p.Time
		}
	}

	now := time.Now().In(s.loc())
	suhoorSeconds := timeToSeconds(suhoorTime, now) // next occurrence
	iftarSeconds := timeToSeconds(iftarTime, now)   // next occurrence

	return &FastingResponse{
		IsFasting:     true,
		FastingType:   schedule.Type,
		SuhoorTime:    suhoorTime,
		IftarTime:     iftarTime,
		SuhoorIn:      formatDuration(suhoorSeconds),
		IftarIn:       formatDuration(iftarSeconds),
		SuhoorSeconds: suhoorSeconds,
		IftarSeconds:  iftarSeconds,
	}, nil
}

// UpdateFastingType updates the active fasting type.
func (s *fastingService) UpdateFastingType(ctx context.Context, fastingType string) error {
	validTypes := map[string]bool{
		"ramadan": true,
		"syawal":  true,
		"sawm":    true,
		"none":    true,
	}

	if !validTypes[fastingType] {
		return fmt.Errorf("invalid fasting type: %s", fastingType)
	}

	if fastingType == "none" {
		return s.repo.DisableAll()
	}

	return s.repo.UpdateEnabled(fastingType)
}

// timeToSeconds calculates seconds until a given HH:MM time.
// If the time has passed today, returns seconds until tomorrow.
// The caller should pass a `now` already in the location's timezone.
func timeToSeconds(timeStr string, now time.Time) int {
	var h, m int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &h, &m)
	if err != nil {
		return 0
	}

	target := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
	diff := int(target.Sub(now).Seconds())

	if diff < 0 {
		// Time has passed, calculate for tomorrow
		diff += 24 * 60 * 60
	}

	return diff
}

// formatDuration converts seconds to a human-readable format like "2h 15m".
func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "now"
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
