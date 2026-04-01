package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// NASAClient handles API calls to NASA
type NASAClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// Response structures for NASA API
type APOD struct {
	Title          string `json:"title"`
	Date           string `json:"date"`
	Explanation    string `json:"explanation"`
	URL            string `json:"url"`
	MediaType      string `json:"media_type"`
	ServiceVersion string `json:"service_version"`
}

// NewNASAClient creates a new NASA API client.
// Uses NASA_API_KEY env if set; otherwise falls back to DEMO_KEY (strict rate limits).
func NewNASAClient() *NASAClient {
	apiKey := os.Getenv("NASA_API_KEY")
	if apiKey == "" {
		apiKey = "DEMO_KEY"
	}
	return &NASAClient{
		baseURL: "https://api.nasa.gov",
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		apiKey: apiKey,
	}
}

// Add logging to API calls
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
			Msg("API request failed")
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
		Msg("API request completed")

	return resp, nil
}

// GetAPOD fetches the Astronomy Picture of the Day
func (c *NASAClient) GetAPOD() (*APOD, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("%s/planetary/apod?api_key=%s", c.baseURL, c.apiKey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for rate limiting headers
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if remaining == "0" {
			return nil, fmt.Errorf("NASA API rate limit exceeded")
		}
	}

	// Also check for 429 status code (Too Many Requests)
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("NASA API rate limit exceeded (429)")
	}

	// Check for other non-success status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NASA API error: HTTP %d", resp.StatusCode)
	}

	var apod APOD
	if err := json.NewDecoder(resp.Body).Decode(&apod); err != nil {
		return nil, err
	}
	return &apod, nil
}
