package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// UserAgentHeader is the header name for client version
	UserAgentHeader = "User-Agent"
	// ClientVersionHeader is a custom header for the client version
	ClientVersionHeader = "X-Client-Version"
)

// Client represents an HTTP client for the outerspace API
type Client struct {
	baseURL    string
	httpClient *http.Client
	version    string
}

// NewClient creates a new HTTP client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		version: getVersion(),
	}
}

// getVersion reads the version from VERSION file or returns default
func getVersion() string {
	if data, err := os.ReadFile("VERSION"); err == nil {
		return strings.TrimSpace(string(data))
	}
	return "dev"
}

// Close is a no-op for HTTP client (for compatibility with gRPC client interface)
func (c *Client) Close() error {
	return nil
}

// Launch represents the latest launch data
type Launch struct {
	FlightNumber int32  `json:"flight_number"`
	MissionName  string `json:"mission_name"`
	DateUtc      string `json:"date_utc"`
	Success      bool   `json:"success"`
	Details      string `json:"details"`
}

// Rocket represents rocket data
type Rocket struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	HeightMeters float32 `json:"height_meters"`
	MassKg       int32   `json:"mass_kg"`
}

// GetRocketsResponse represents the response from getting all rockets
type GetRocketsResponse struct {
	Rockets []*Rocket `json:"rockets"`
}

// MathFact represents a math fact
type MathFact struct {
	Number int32  `json:"number"`
	Type   string `json:"type"`
	Text   string `json:"text"`
	Found  bool   `json:"found"`
}

// NASAData represents NASA's Astronomy Picture of the Day
type NASAData struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Explanation string `json:"explanation"`
	URL         string `json:"url"`
	MediaType   string `json:"media_type"`
}

// GetLatestLaunch calls the HTTP API to get the latest launch
func (c *Client) GetLatestLaunch(ctx context.Context) (*Launch, error) {
	url := fmt.Sprintf("%s/api/latest-launch", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.addVersionHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var launch Launch
	if err := json.NewDecoder(resp.Body).Decode(&launch); err != nil {
		return nil, err
	}

	return &launch, nil
}

// GetRocket calls the HTTP API to get a specific rocket
func (c *Client) GetRocket(ctx context.Context, id string) (*Rocket, error) {
	u := fmt.Sprintf("%s/api/rocket?id=%s", c.baseURL, url.QueryEscape(id))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.addVersionHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var rocket Rocket
	if err := json.NewDecoder(resp.Body).Decode(&rocket); err != nil {
		return nil, err
	}

	return &rocket, nil
}

// GetRockets calls the HTTP API to get all rockets
func (c *Client) GetRockets(ctx context.Context) (*GetRocketsResponse, error) {
	url := fmt.Sprintf("%s/api/rockets", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.addVersionHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var rockets []*Rocket
	if err := json.NewDecoder(resp.Body).Decode(&rockets); err != nil {
		return nil, err
	}

	return &GetRocketsResponse{Rockets: rockets}, nil
}

// GetMathFact calls the HTTP API to get a math fact
func (c *Client) GetMathFact(ctx context.Context) (*MathFact, error) {
	url := fmt.Sprintf("%s/api/numbers", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.addVersionHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var mathFact MathFact
	if err := json.NewDecoder(resp.Body).Decode(&mathFact); err != nil {
		return nil, err
	}

	return &mathFact, nil
}

// GetNASAData calls the HTTP API to get NASA's Astronomy Picture of the Day
func (c *Client) GetNASAData(ctx context.Context) (*NASAData, error) {
	url := fmt.Sprintf("%s/api/nasa", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.addVersionHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var nasaData NASAData
	if err := json.NewDecoder(resp.Body).Decode(&nasaData); err != nil {
		return nil, err
	}

	return &nasaData, nil
}

// addVersionHeaders adds version-related headers to the request
func (c *Client) addVersionHeaders(req *http.Request) {
	userAgent := fmt.Sprintf("outerspace-go/%s", c.version)
	req.Header.Set(UserAgentHeader, userAgent)
	req.Header.Set(ClientVersionHeader, c.version)
}
