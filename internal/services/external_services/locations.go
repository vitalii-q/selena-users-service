package external_services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type LocationCountry struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Code   string         `json:"code"`
	Cities []LocationCity `json:"cities"`
}

type LocationCity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LocationsClient struct {
	BaseURL string
	Client  *http.Client
}

func NewLocationsClient() *LocationsClient {
	baseURL := os.Getenv("HOTELS_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://hotels-service:9064"
	}

	return &LocationsClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *LocationsClient) GetLocations() ([]LocationCountry, error) {
	resp, err := c.Client.Get(c.BaseURL + "/api/v1/locations")
	if err != nil {
		return nil, fmt.Errorf("failed to get locations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("locations service returned %d", resp.StatusCode)
	}

	var locations []LocationCountry
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return nil, fmt.Errorf("decode locations error: %w", err)
	}

	return locations, nil
}
