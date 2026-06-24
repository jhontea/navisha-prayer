package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// AladhanClient defines the interface for the external Aladhan API client.
type AladhanClient interface {
	GetPrayerTimes(ctx context.Context, date string, lat, lon float64, method int) (*AladhanResponse, error)
	GetMonthlyPrayerTimes(ctx context.Context, lat, lon float64, month, year, method int) (*MonthlyCalendarResponse, error)
	GetHijriCalendar(ctx context.Context, lat, lon float64, month, year int) (*HijriCalendarResponse, error)
	GetCities(ctx context.Context, query string) (*CitySearchResponse, error)
}

// AladhanResponse is the top-level JSON response from the Aladhan API.
type AladhanResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   AladhanData `json:"data"`
}

type AladhanData struct {
	Timings Timings     `json:"timings"`
	Date    AladhanDate `json:"date"`
}

type Timings struct {
	Fajr    string `json:"Fajr"`
	Sunrise string `json:"Sunrise"`
	Dhuhr   string `json:"Dhuhr"`
	Asr     string `json:"Asr"`
	Maghrib string `json:"Maghrib"`
	Isha    string `json:"Isha"`
}

type AladhanDate struct {
	Hijri     HijriDate     `json:"hijri"`
	Gregorian GregorianDate `json:"gregorian"`
}

type HijriDate struct {
	Date  string     `json:"date"`
	Month HijriMonth `json:"month"`
}

type HijriMonth struct {
	Number int    `json:"number"`
	En     string `json:"en"`
}

type GregorianDate struct {
	Date string `json:"date"`
}

// CitySearchResponse is used for city autocomplete.
type CitySearchResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   []CityEntry `json:"data"`
}

type CityEntry struct {
	Name    string  `json:"name"`
	Country string  `json:"country"`
	Lat     float64 `json:"latitude,string"`
	Lon     float64 `json:"longitude,string"`
}

// MonthlyCalendarResponse is the response from the monthly prayer-times calendar endpoint.
type MonthlyCalendarResponse struct {
	Code   int           `json:"code"`
	Status string        `json:"status"`
	Data   []AladhanData `json:"data"`
}

// HijriCalendarResponse is the response from the Hijri calendar endpoint.
type HijriCalendarResponse struct {
	Code   int                `json:"code"`
	Status string             `json:"status"`
	Data   []HijriCalendarDay `json:"data"`
}

type HijriCalendarDay struct {
	Gregorian GregorianDate `json:"gregorian"`
	Hijri     HijriDate     `json:"hijri"`
}

// AladhanAPIClient implements AladhanClient for the real Aladhan API.
type AladhanAPIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewAladhanAPIClient creates a new Aladhan API client.
func NewAladhanAPIClient(baseURL, apiKey string) *AladhanAPIClient {
	return &AladhanAPIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetPrayerTimes fetches prayer times for a specific date and location.
func (c *AladhanAPIClient) GetPrayerTimes(ctx context.Context, date string, lat, lon float64, method int) (*AladhanResponse, error) {
	endpoint := fmt.Sprintf("%s/timings/%s", c.baseURL, date)
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	q.Add("latitude", fmt.Sprintf("%f", lat))
	q.Add("longitude", fmt.Sprintf("%f", lon))
	q.Add("method", fmt.Sprintf("%d", method))
	if c.apiKey != "" {
		q.Add("api_key", c.apiKey)
	}
	reqURL.RawQuery = q.Encode()

	resp, err := c.doRequest(ctx, reqURL.String())
	if err != nil {
		return nil, err
	}

	var result AladhanResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prayer times response: %w", err)
	}

	return &result, nil
}

// GetMonthlyPrayerTimes fetches an entire month of prayer times in a single request
// using the Aladhan /calendar endpoint.
func (c *AladhanAPIClient) GetMonthlyPrayerTimes(ctx context.Context, lat, lon float64, month, year, method int) (*MonthlyCalendarResponse, error) {
	endpoint := fmt.Sprintf("%s/calendar/%d/%d", c.baseURL, year, month)
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	q.Add("latitude", fmt.Sprintf("%f", lat))
	q.Add("longitude", fmt.Sprintf("%f", lon))
	q.Add("method", fmt.Sprintf("%d", method))
	if c.apiKey != "" {
		q.Add("api_key", c.apiKey)
	}
	reqURL.RawQuery = q.Encode()

	resp, err := c.doRequest(ctx, reqURL.String())
	if err != nil {
		return nil, err
	}

	var result MonthlyCalendarResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal monthly prayer times response: %w", err)
	}

	return &result, nil
}

// GetHijriCalendar fetches the Hijri calendar for a specific month and year.
func (c *AladhanAPIClient) GetHijriCalendar(ctx context.Context, lat, lon float64, month, year int) (*HijriCalendarResponse, error) {
	endpoint := fmt.Sprintf("%s/hijriCalendar", c.baseURL)
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	q.Add("latitude", fmt.Sprintf("%f", lat))
	q.Add("longitude", fmt.Sprintf("%f", lon))
	q.Add("month", fmt.Sprintf("%d", month))
	q.Add("year", fmt.Sprintf("%d", year))
	if c.apiKey != "" {
		q.Add("api_key", c.apiKey)
	}
	reqURL.RawQuery = q.Encode()

	resp, err := c.doRequest(ctx, reqURL.String())
	if err != nil {
		return nil, err
	}

	var result HijriCalendarResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hijri calendar response: %w", err)
	}

	return &result, nil
}

// GetCities searches for cities by name.
func (c *AladhanAPIClient) GetCities(ctx context.Context, query string) (*CitySearchResponse, error) {
	endpoint := fmt.Sprintf("%s/cityInfo", c.baseURL)
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	q.Add("city", query)
	if c.apiKey != "" {
		q.Add("api_key", c.apiKey)
	}
	reqURL.RawQuery = q.Encode()

	resp, err := c.doRequest(ctx, reqURL.String())
	if err != nil {
		return nil, err
	}

	var result CitySearchResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal city search response: %w", err)
	}

	return &result, nil
}

func (c *AladhanAPIClient) doRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "NavishaPrayer/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aladhan API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("aladhan API returned status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
