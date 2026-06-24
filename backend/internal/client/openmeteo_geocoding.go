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

// OpenMeteoGeocodingClient handles city search via Open-Meteo Geocoding API.
// This replaces Aladhan's cityInfo endpoint which does not support autocomplete
// or proper geocoding (it only accepts city+country and returns a single object).
type OpenMeteoGeocodingClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOpenMeteoGeocodingClient creates a new Open-Meteo geocoding client.
func NewOpenMeteoGeocodingClient(baseURL string) *OpenMeteoGeocodingClient {
	return &OpenMeteoGeocodingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GeocodingResponse is the response from Open-Meteo geocoding API.
type GeocodingResponse struct {
	Results          []GeocodingResult `json:"results"`
	GenerationTimeMs float64           `json:"generationtime_ms"`
}

// GeocodingResult represents a single city search result.
type GeocodingResult struct {
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Lat         float64 `json:"latitude"`
	Lon         float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	Admin1      string  `json:"admin1,omitempty"`
	Admin2      string  `json:"admin2,omitempty"`
	Population  int64   `json:"population,omitempty"`
}

// SearchCities searches for cities by name using Open-Meteo geocoding API.
func (c *OpenMeteoGeocodingClient) SearchCities(ctx context.Context, query string, count int) ([]GeocodingResult, error) {
	endpoint := fmt.Sprintf("%s/search", c.baseURL)
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := reqURL.Query()
	q.Add("name", query)
	q.Add("count", fmt.Sprintf("%d", count))
	q.Add("language", "en")
	q.Add("format", "json")
	reqURL.RawQuery = q.Encode()

	resp, err := c.doRequest(ctx, reqURL.String())
	if err != nil {
		return nil, err
	}

	var result GeocodingResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal geocoding response: %w", err)
	}

	if result.Results == nil {
		return []GeocodingResult{}, nil
	}

	return result.Results, nil
}

func (c *OpenMeteoGeocodingClient) doRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "NavishaPrayer/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("open-meteo geocoding API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("open-meteo geocoding API returned status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
