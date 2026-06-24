package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"navisha-prayer/internal/client"
	"navisha-prayer/internal/config"
	"navisha-prayer/internal/repository"
)

// GeocodingClient defines the interface for geocoding/city search.
// We use Open-Meteo Geocoding API instead of Aladhan cityInfo because
// Aladhan's cityInfo does not support autocomplete or proper geocoding.
type GeocodingClient interface {
	SearchCities(ctx context.Context, query string, count int) ([]client.GeocodingResult, error)
}

// LocationService handles location resolution.
type LocationService interface {
	GetLocation(ctx context.Context) (*LocationResponse, error)
	SetLocationGPS(ctx context.Context, lat, lon float64) error
	SetLocationCity(ctx context.Context, cityName, countryCode string) error
	SearchCities(ctx context.Context, query string) ([]CityResult, error)
}

// LocationResponse contains the current location info.
type LocationResponse struct {
	Lat     float64 `json:"latitude"`
	Lon     float64 `json:"longitude"`
	City    string  `json:"city"`
	Country string  `json:"country"`
	Source  string  `json:"source"` // "gps" | "city_search" | "default"
}

// CityResult is a city search result.
type CityResult struct {
	Name    string  `json:"name"`
	Country string  `json:"country"`
	Lat     float64 `json:"latitude"`
	Lon     float64 `json:"longitude"`
}

// locationService implements LocationService.
type locationService struct {
	geoClient GeocodingClient
	repo      *repository.SettingsRepository
	defaults  config.DefaultsConfig
}

func NewLocationService(geoClient GeocodingClient, repo *repository.SettingsRepository, defaults config.DefaultsConfig) LocationService {
	return &locationService{
		geoClient: geoClient,
		repo:      repo,
		defaults:  defaults,
	}
}

// GetLocation returns the current stored location.
func (s *locationService) GetLocation(ctx context.Context) (*LocationResponse, error) {
	lat, _ := s.repo.Get("location.lat")
	lon, _ := s.repo.Get("location.lon")
	city, _ := s.repo.Get("location.city")
	country, _ := s.repo.Get("location.country")
	source, _ := s.repo.Get("location.source")

	// Use defaults if not set
	latVal := s.defaults.Lat
	if lat != "" {
		if v, err := strconv.ParseFloat(lat, 64); err == nil {
			latVal = v
		}
	}

	lonVal := s.defaults.Lon
	if lon != "" {
		if v, err := strconv.ParseFloat(lon, 64); err == nil {
			lonVal = v
		}
	}

	cityVal := city
	if cityVal == "" {
		cityVal = "Kuala Lumpur"
	}

	countryVal := country
	if countryVal == "" {
		countryVal = "MY"
	}

	sourceVal := source
	if sourceVal == "" {
		sourceVal = "default"
	}

	return &LocationResponse{
		Lat:     latVal,
		Lon:     lonVal,
		City:    cityVal,
		Country: countryVal,
		Source:  sourceVal,
	}, nil
}

// SetLocationGPS sets the location via GPS coordinates.
func (s *locationService) SetLocationGPS(ctx context.Context, lat, lon float64) error {
	// Reverse geocode to get city name using Open-Meteo
	cityName := ""
	countryCode := ""

	results, err := s.geoClient.SearchCities(ctx, fmt.Sprintf("%.4f,%.4f", lat, lon), 1)
	if err == nil && len(results) > 0 {
		cityName = results[0].Name
		countryCode = results[0].CountryCode
	}

	s.repo.Set("location.lat", strconv.FormatFloat(lat, 'f', 4, 64))
	s.repo.Set("location.lon", strconv.FormatFloat(lon, 'f', 4, 64))
	s.repo.Set("location.city", cityName)
	s.repo.Set("location.country", countryCode)
	s.repo.Set("location.source", "gps")

	return nil
}

// SetLocationCity sets the location via city name.
func (s *locationService) SetLocationCity(ctx context.Context, cityName, countryCode string) error {
	results, err := s.geoClient.SearchCities(ctx, cityName, 10)
	if err != nil {
		return fmt.Errorf("failed to find city: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("city not found: %s", cityName)
	}

	city := results[0]
	if countryCode != "" {
		// Find matching country
		for _, c := range results {
			if strings.EqualFold(c.CountryCode, countryCode) {
				city = c
				break
			}
		}
	}

	s.repo.Set("location.lat", strconv.FormatFloat(city.Lat, 'f', 4, 64))
	s.repo.Set("location.lon", strconv.FormatFloat(city.Lon, 'f', 4, 64))
	s.repo.Set("location.city", city.Name)
	s.repo.Set("location.country", city.CountryCode)
	s.repo.Set("location.source", "city_search")

	return nil
}

// SearchCities searches for cities by name using Open-Meteo geocoding API.
func (s *locationService) SearchCities(ctx context.Context, query string) ([]CityResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("city search query cannot be empty")
	}

	results, err := s.geoClient.SearchCities(ctx, query, 10)
	if err != nil {
		return nil, fmt.Errorf("city search failed: %w", err)
	}

	cityResults := make([]CityResult, 0, len(results))
	for _, city := range results {
		cityResults = append(cityResults, CityResult{
			Name:    city.Name,
			Country: city.CountryCode,
			Lat:     city.Lat,
			Lon:     city.Lon,
		})
	}

	return cityResults, nil
}
