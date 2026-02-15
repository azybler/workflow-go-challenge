package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WeatherClient fetches current temperature for geographic coordinates.
type WeatherClient interface {
	GetTemperature(ctx context.Context, lat, lon float64) (float64, error)
}

// OpenMeteoClient calls the Open-Meteo public weather API.
type OpenMeteoClient struct {
	httpClient *http.Client
}

// NewOpenMeteoClient returns a client with a 10-second timeout.
func NewOpenMeteoClient() *OpenMeteoClient {
	return &OpenMeteoClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// openMeteoResponse is the relevant subset of the Open-Meteo API response.
type openMeteoResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
	} `json:"current_weather"`
}

// GetTemperature fetches the current temperature in Celsius for the given coordinates.
func (c *OpenMeteoClient) GetTemperature(ctx context.Context, lat, lon float64) (float64, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current_weather=true",
		lat, lon,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("weather API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var result openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode weather response: %w", err)
	}

	return result.CurrentWeather.Temperature, nil
}
