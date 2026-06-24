package service

import (
	"context"
	"fmt"
	"time"

	"navisha-prayer/internal/client"
	"navisha-prayer/internal/config"
	"navisha-prayer/internal/repository"
)

// PrayerService defines the prayer time business logic interface.
type PrayerService interface {
	GetTodayPrayerTimes(ctx context.Context) (*PrayerTimesResponse, error)
	GetPrayerTimesWithLocation(ctx context.Context, lat, lon float64, method int) (*PrayerTimesResponse, error)
	GetPrayerTimesByDate(ctx context.Context, date string) (*PrayerTimesResponse, error)
	GetMonthlyPrayerTimes(ctx context.Context, year, month int) ([]PrayerTimesResponse, error)
	GetNextPrayer(ctx context.Context) (*NextPrayerEntry, error)
	RefreshCache(ctx context.Context) error
}

// NextPrayerEntry is a PrayerEntry plus seconds remaining until it begins.
type NextPrayerEntry struct {
	Name                 string `json:"name"`
	Time                 string `json:"time"`
	IsNext               bool   `json:"is_next"`
	TimeRemainingSeconds int    `json:"time_remaining_seconds"`
}

// PrayerTimesResponse is the canonical API response for prayer times.
type PrayerTimesResponse struct {
	Date    DateInfo      `json:"date"`
	Lat     float64       `json:"latitude"`
	Lon     float64       `json:"longitude"`
	Method  int           `json:"method"`
	Prayers []PrayerEntry `json:"prayers"`
}

// StaleCacheError indicates the response is from stale cache.
type StaleCacheError struct {
	Response *PrayerTimesResponse
	Err      error
}

func (e *StaleCacheError) Error() string {
	return fmt.Errorf("stale cache: %w", e.Err).Error()
}

// NoCacheError indicates no cache was available at all.
type NoCacheError struct {
	Err error
}

func (e *NoCacheError) Error() string {
	return fmt.Errorf("no cache available: %w", e.Err).Error()
}

type DateInfo struct {
	Hijri        string `json:"hijri"`
	HijriMonthNo int    `json:"hijri_month_number"`
	Gregorian    string `json:"gregorian"`
}

type PrayerEntry struct {
	Name   string `json:"name"`
	Time   string `json:"time"` // "HH:MM" in 24h format
	IsNext bool   `json:"is_next"`
}

// prayerService implements PrayerService.
type prayerService struct {
	client       client.AladhanClient
	repo         *repository.PrayerCacheRepository
	settingsRepo *repository.SettingsRepository
	cfg          *config.Config
}

func NewPrayerService(client client.AladhanClient, repo *repository.PrayerCacheRepository, settingsRepo *repository.SettingsRepository, cfg *config.Config) PrayerService {
	return &prayerService{
		client:       client,
		repo:         repo,
		settingsRepo: settingsRepo,
		cfg:          cfg,
	}
}

// loc returns the configured *time.Location, falling back to UTC if the
// timezone name is invalid or empty.
func (s *prayerService) loc() *time.Location {
	if s.cfg != nil && s.cfg.Defaults.Timezone != "" {
		if l, err := time.LoadLocation(s.cfg.Defaults.Timezone); err == nil {
			return l
		}
	}
	return time.UTC
}

// GetTodayPrayerTimes returns today's prayer times.
func (s *prayerService) GetTodayPrayerTimes(ctx context.Context) (*PrayerTimesResponse, error) {
	now := time.Now().In(s.loc())
	dateStr := now.Format("02-01-2006")
	lat, lon, method := s.getLocation()

	// Check cache first
	if valid, cache := s.repo.IsCacheValid(lat, lon, dateStr, method); valid {
		return s.buildResponseFromCache(cache, lat, lon, method), nil
	}

	// Fetch from API
	aladhanResp, err := s.client.GetPrayerTimes(ctx, dateStr, lat, lon, method)
	if err != nil {
		// Try stale cache as fallback
		if cache, err2 := s.repo.GetByLocationAndDate(lat, lon, dateStr, method); err2 == nil {
			resp := s.buildResponseFromCache(cache, lat, lon, method)
			return nil, &StaleCacheError{Response: resp, Err: err}
		}
		return nil, &NoCacheError{Err: err}
	}

	// Cache the result
	cache := &repository.PrayerCache{
		Lat:         lat,
		Lon:         lon,
		Date:        dateStr,
		Method:      method,
		Fajr:        aladhanResp.Data.Timings.Fajr,
		Sunrise:     aladhanResp.Data.Timings.Sunrise,
		Dhuhr:       aladhanResp.Data.Timings.Dhuhr,
		Asr:         aladhanResp.Data.Timings.Asr,
		Maghrib:     aladhanResp.Data.Timings.Maghrib,
		Isha:        aladhanResp.Data.Timings.Isha,
		RawResponse: fmt.Sprintf("%v", aladhanResp.Data),
	}
	s.repo.Upsert(cache)

	return s.buildResponse(aladhanResp, lat, lon, method), nil
}

// GetPrayerTimesWithLocation returns today's prayer times using the provided GPS coordinates and method,
// bypassing the stored location settings.
func (s *prayerService) GetPrayerTimesWithLocation(ctx context.Context, lat, lon float64, method int) (*PrayerTimesResponse, error) {
	now := time.Now().In(s.loc())
	dateStr := now.Format("02-01-2006")

	// Use default method if not specified
	if method == 0 {
		method = s.getSettingInt("calculation.method", s.cfg.Defaults.Method)
	}

	// Check cache first
	if valid, cache := s.repo.IsCacheValid(lat, lon, dateStr, method); valid {
		return s.buildResponseFromCache(cache, lat, lon, method), nil
	}

	// Fetch from API
	aladhanResp, err := s.client.GetPrayerTimes(ctx, dateStr, lat, lon, method)
	if err != nil {
		if cache, err2 := s.repo.GetByLocationAndDate(lat, lon, dateStr, method); err2 == nil {
			resp := s.buildResponseFromCache(cache, lat, lon, method)
			return nil, &StaleCacheError{Response: resp, Err: err}
		}
		return nil, &NoCacheError{Err: err}
	}

	// Cache the result
	cache := &repository.PrayerCache{
		Lat:         lat,
		Lon:         lon,
		Date:        dateStr,
		Method:      method,
		Fajr:        aladhanResp.Data.Timings.Fajr,
		Sunrise:     aladhanResp.Data.Timings.Sunrise,
		Dhuhr:       aladhanResp.Data.Timings.Dhuhr,
		Asr:         aladhanResp.Data.Timings.Asr,
		Maghrib:     aladhanResp.Data.Timings.Maghrib,
		Isha:        aladhanResp.Data.Timings.Isha,
		RawResponse: fmt.Sprintf("%v", aladhanResp.Data),
	}
	s.repo.Upsert(cache)

	return s.buildResponse(aladhanResp, lat, lon, method), nil
}

// GetPrayerTimesByDate returns prayer times for a specific date (DD-MM-YYYY).
func (s *prayerService) GetPrayerTimesByDate(ctx context.Context, date string) (*PrayerTimesResponse, error) {
	lat, lon, method := s.getLocation()

	// Check cache first
	if valid, cache := s.repo.IsCacheValid(lat, lon, date, method); valid {
		return s.buildResponseFromCache(cache, lat, lon, method), nil
	}

	// Fetch from API
	aladhanResp, err := s.client.GetPrayerTimes(ctx, date, lat, lon, method)
	if err != nil {
		if cache, err2 := s.repo.GetByLocationAndDate(lat, lon, date, method); err2 == nil {
			resp := s.buildResponseFromCache(cache, lat, lon, method)
			return nil, &StaleCacheError{Response: resp, Err: err}
		}
		return nil, &NoCacheError{Err: err}
	}

	// Cache the result
	cache := &repository.PrayerCache{
		Lat:         lat,
		Lon:         lon,
		Date:        date,
		Method:      method,
		Fajr:        aladhanResp.Data.Timings.Fajr,
		Sunrise:     aladhanResp.Data.Timings.Sunrise,
		Dhuhr:       aladhanResp.Data.Timings.Dhuhr,
		Asr:         aladhanResp.Data.Timings.Asr,
		Maghrib:     aladhanResp.Data.Timings.Maghrib,
		Isha:        aladhanResp.Data.Timings.Isha,
		RawResponse: fmt.Sprintf("%v", aladhanResp.Data),
	}
	s.repo.Upsert(cache)

	return s.buildResponse(aladhanResp, lat, lon, method), nil
}

// GetMonthlyPrayerTimes returns prayer times for an entire month.
// It fetches a single calendar response from the Aladhan API and caches each day.
func (s *prayerService) GetMonthlyPrayerTimes(ctx context.Context, year, month int) ([]PrayerTimesResponse, error) {
	lat, lon, method := s.getLocation()

	aladhanResp, err := s.client.GetMonthlyPrayerTimes(ctx, lat, lon, month, year, method)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monthly prayer times: %w", err)
	}

	results := make([]PrayerTimesResponse, 0, len(aladhanResp.Data))
	var errs []string

	for _, day := range aladhanResp.Data {
		cache := &repository.PrayerCache{
			Lat:         lat,
			Lon:         lon,
			Date:        day.Date.Gregorian.Date,
			Method:      method,
			Fajr:        day.Timings.Fajr,
			Sunrise:     day.Timings.Sunrise,
			Dhuhr:       day.Timings.Dhuhr,
			Asr:         day.Timings.Asr,
			Maghrib:     day.Timings.Maghrib,
			Isha:        day.Timings.Isha,
			RawResponse: fmt.Sprintf("%v", day),
		}
		if err := s.repo.Upsert(cache); err != nil {
			errs = append(errs, fmt.Sprintf("cache day %s: %v", cache.Date, err))
		}

		results = append(results, *s.buildResponse(&client.AladhanResponse{Data: day}, lat, lon, method))
	}

	if len(errs) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("all days failed: %v", errs)
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("partial data: %d/%d days succeeded; errors: %v", len(results), len(aladhanResp.Data), errs)
	}

	return results, nil
}

// GetNextPrayer returns the next upcoming prayer along with seconds remaining.
func (s *prayerService) GetNextPrayer(ctx context.Context) (*NextPrayerEntry, error) {
	resp, err := s.GetTodayPrayerTimes(ctx)
	if err != nil {
		// Surface stale-cache responses so the caller can still compute the next prayer.
		if e, ok := err.(*StaleCacheError); ok && e.Response != nil {
			resp = e.Response
		} else {
			return nil, err
		}
	}

	var next *PrayerEntry
	for i := range resp.Prayers {
		if resp.Prayers[i].IsNext {
			next = &resp.Prayers[i]
			break
		}
	}
	if next == nil {
		return nil, fmt.Errorf("could not determine next prayer")
	}

	now := time.Now().In(s.loc())
	h, m, perr := parseTime(next.Time)
	remaining := 0
	if perr == nil {
		target := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
		diff := int(target.Sub(now).Seconds())
		if diff < 0 {
			diff += 24 * 60 * 60 // next occurrence is tomorrow (e.g. after Isha)
		}
		remaining = diff
	}

	return &NextPrayerEntry{
		Name:                 next.Name,
		Time:                 next.Time,
		IsNext:               true,
		TimeRemainingSeconds: remaining,
	}, nil
}

// RefreshCache forces a refresh of the prayer times cache.
func (s *prayerService) RefreshCache(ctx context.Context) error {
	lat, lon, method := s.getLocation()
	now := time.Now().In(s.loc())
	dateStr := now.Format("02-01-2006")

	aladhanResp, err := s.client.GetPrayerTimes(ctx, dateStr, lat, lon, method)
	if err != nil {
		return fmt.Errorf("failed to refresh cache: %w", err)
	}

	cache := &repository.PrayerCache{
		Lat:         lat,
		Lon:         lon,
		Date:        dateStr,
		Method:      method,
		Fajr:        aladhanResp.Data.Timings.Fajr,
		Sunrise:     aladhanResp.Data.Timings.Sunrise,
		Dhuhr:       aladhanResp.Data.Timings.Dhuhr,
		Asr:         aladhanResp.Data.Timings.Asr,
		Maghrib:     aladhanResp.Data.Timings.Maghrib,
		Isha:        aladhanResp.Data.Timings.Isha,
		RawResponse: fmt.Sprintf("%v", aladhanResp.Data),
	}
	return s.repo.Upsert(cache)
}

func (s *prayerService) getLocation() (float64, float64, int) {
	lat := s.getSettingFloat("location.lat", s.cfg.Defaults.Lat)
	lon := s.getSettingFloat("location.lon", s.cfg.Defaults.Lon)
	method := s.getSettingInt("calculation.method", s.cfg.Defaults.Method)
	return lat, lon, method
}

func (s *prayerService) getSettingFloat(key string, fallback float64) float64 {
	if s.settingsRepo == nil {
		return fallback
	}
	val, err := s.settingsRepo.Get(key)
	if err != nil || val == "" {
		return fallback
	}
	var v float64
	if _, err := fmt.Sscanf(val, "%f", &v); err == nil {
		return v
	}
	return fallback
}

func (s *prayerService) getSettingInt(key string, fallback int) int {
	if s.settingsRepo == nil {
		return fallback
	}
	val, err := s.settingsRepo.Get(key)
	if err != nil || val == "" {
		return fallback
	}
	var v int
	if _, err := fmt.Sscanf(val, "%d", &v); err == nil {
		return v
	}
	return fallback
}

func (s *prayerService) buildResponse(resp *client.AladhanResponse, lat, lon float64, method int) *PrayerTimesResponse {
	timings := resp.Data.Timings
	hijriDate := resp.Data.Date

	prayers := []PrayerEntry{
		{Name: "Fajr", Time: cleanTime(timings.Fajr)},
		{Name: "Sunrise", Time: cleanTime(timings.Sunrise)},
		{Name: "Dhuhr", Time: cleanTime(timings.Dhuhr)},
		{Name: "Asr", Time: cleanTime(timings.Asr)},
		{Name: "Maghrib", Time: cleanTime(timings.Maghrib)},
		{Name: "Isha", Time: cleanTime(timings.Isha)},
	}

	markNextPrayer(prayers, time.Now().In(s.loc()))

	return &PrayerTimesResponse{
		Date: DateInfo{
			Hijri:        hijriDate.Hijri.Date,
			HijriMonthNo: hijriDate.Hijri.Month.Number,
			Gregorian:    hijriDate.Gregorian.Date,
		},
		Lat:     lat,
		Lon:     lon,
		Method:  method,
		Prayers: prayers,
	}
}

func (s *prayerService) buildResponseFromCache(cache *repository.PrayerCache, lat, lon float64, method int) *PrayerTimesResponse {
	prayers := []PrayerEntry{
		{Name: "Fajr", Time: cleanTime(cache.Fajr)},
		{Name: "Sunrise", Time: cleanTime(cache.Sunrise)},
		{Name: "Dhuhr", Time: cleanTime(cache.Dhuhr)},
		{Name: "Asr", Time: cleanTime(cache.Asr)},
		{Name: "Maghrib", Time: cleanTime(cache.Maghrib)},
		{Name: "Isha", Time: cleanTime(cache.Isha)},
	}

	markNextPrayer(prayers, time.Now().In(s.loc()))

	return &PrayerTimesResponse{
		Date: DateInfo{
			Hijri:        "",
			HijriMonthNo: 0,
			Gregorian:    cache.Date,
		},
		Lat:     lat,
		Lon:     lon,
		Method:  method,
		Prayers: prayers,
	}
}

// cleanTime removes timezone suffix from Aladhan time strings (e.g., "05:58 (MYT)" -> "05:58").
func cleanTime(t string) string {
	for i, c := range t {
		if c == ' ' {
			return t[:i]
		}
	}
	return t
}

// markNextPrayer marks the next upcoming prayer based on the provided current time.
// The caller is responsible for passing a time in the location's timezone.
func markNextPrayer(prayers []PrayerEntry, now time.Time) {
	currentMinutes := now.Hour()*60 + now.Minute()

	lastIdx := 0
	for i, p := range prayers {
		h, m, err := parseTime(p.Time)
		if err != nil {
			continue
		}
		if h*60+m <= currentMinutes {
			lastIdx = i
		}
	}

	// The next prayer is the one after the last past prayer
	nextIdx := (lastIdx + 1) % len(prayers)
	prayers[nextIdx].IsNext = true
}

func parseTime(t string) (int, int, error) {
	var h, m int
	_, err := fmt.Sscanf(t, "%d:%d", &h, &m)
	return h, m, err
}
