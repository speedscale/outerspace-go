package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// NASAClient handles API calls to NASA APOD
type NASAClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// APODResponse represents the NASA APOD API response
type APODResponse struct {
	Date           string `json:"date"`
	Explanation    string `json:"explanation"`
	HDUrl          string `json:"hdurl"`
	MediaType      string `json:"media_type"`
	ServiceVersion string `json:"service_version"`
	Title          string `json:"title"`
	Url            string `json:"url"`
}

// NewNASAClient creates a new NASA APOD API client
func NewNASAClient() *NASAClient {
	return &NASAClient{
		baseURL: "https://api.nasa.gov/planetary/apod",
		apiKey:  "DEMO_KEY", // Using demo key, replace with real key if needed
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// makeRequest makes an HTTP request with logging
func (c *NASAClient) makeRequest(method, url string) (*http.Response, error) {
	start := time.Now()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		log.Error().
			Str("method", method).
			Str("host", req.URL.Host).
			Dur("latency", duration).
			Err(err).
			Msg("NASA API request failed")
		return nil, err
	}

	// Log X-prefixed headers
	for header, values := range resp.Header {
		if len(header) > 0 && (header[0] == 'X' || header[0] == 'x') {
			log.Info().
				Str("header", header).
				Strs("values", values).
				Msg("X-Header found")
		}
	}

	log.Info().
		Str("method", method).
		Str("host", req.URL.Host).
		Int("status", resp.StatusCode).
		Dur("latency", duration).
		Msg("NASA API request completed")

	return resp, nil
}

// GetAPOD fetches the Astronomy Picture of the Day for a specific date
func (c *NASAClient) GetAPOD(date string) (*APODResponse, error) {
	url := fmt.Sprintf("%s?api_key=%s&date=%s", c.baseURL, c.apiKey, date)
	
	resp, err := c.makeRequest("GET", url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NASA API returned status %d", resp.StatusCode)
	}

	var apod APODResponse
	if err := json.NewDecoder(resp.Body).Decode(&apod); err != nil {
		return nil, fmt.Errorf("failed to decode APOD response: %w", err)
	}
	
	return &apod, nil
}