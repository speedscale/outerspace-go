package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
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

// NewNASAClient uses frozen demo data unless NASA_BASE_URL is explicitly set.
func NewNASAClient() *NASAClient {
	apiKey := ""
	if os.Getenv("NASA_BASE_URL") != "" {
		apiKey = os.Getenv("NASA_API_KEY")
		if apiKey == "" {
			apiKey = "DEMO_KEY"
		}
	}
	return &NASAClient{
		baseURL: upstreamURL("NASA_BASE_URL", "nasa"),
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		apiKey: apiKey,
	}
}

func (c *NASAClient) makeRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid NASA upstream URL")
	}
	return upstreamRequest(c.httpClient, "NASA", req)
}

// GetAPOD fetches the Astronomy Picture of the Day
func (c *NASAClient) GetAPOD() (*APOD, error) {
	endpoint := c.baseURL + "/planetary/apod"
	if c.apiKey != "" {
		endpoint += "?api_key=" + url.QueryEscape(c.apiKey)
	}
	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apod APOD
	if err := json.NewDecoder(resp.Body).Decode(&apod); err != nil {
		return nil, err
	}
	return &apod, nil
}
