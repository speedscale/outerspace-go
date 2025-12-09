package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// SpaceXClient handles API calls to SpaceX
type SpaceXClient struct {
	baseURL    string
	httpClient *http.Client
}

// Response structures for SpaceX API
type Launch struct {
	FlightNumber int      `json:"flight_number"`
	MissionName  string   `json:"name"`
	DateUTC      string   `json:"date_utc"`
	Success      bool     `json:"success"`
	Details      string   `json:"details"`
	Ships        []string `json:"ships"`
}

type Rocket struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Height      struct {
		Meters float64 `json:"meters"`
	} `json:"height"`
	Mass struct {
		Kg int `json:"kg"`
	} `json:"mass"`
}

// RocketSummary provides a simplified view of rocket data
type RocketSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LaunchesSummary provides a summary of launches by year and by ship
type LaunchesSummary struct {
	ByYear map[string]int            `json:"by_year"`
	ByShip map[string]int            `json:"by_ship"`
	Total  int                       `json:"total"`
}

// NewSpaceXClient creates a new SpaceX API client
func NewSpaceXClient() *SpaceXClient {
	return &SpaceXClient{
		baseURL: "https://api.spacexdata.com/v4",
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// Add logging to API calls
func (c *SpaceXClient) makeRequest(method, url string) (*http.Response, error) {
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

// GetAllRockets fetches all rocket summaries
func (c *SpaceXClient) GetAllRockets() ([]RocketSummary, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("%s/rockets", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rockets []Rocket
	if err := json.NewDecoder(resp.Body).Decode(&rockets); err != nil {
		return nil, err
	}

	// Convert to summaries
	summaries := make([]RocketSummary, len(rockets))
	for i, rocket := range rockets {
		summaries[i] = RocketSummary{
			ID:   rocket.ID,
			Name: rocket.Name,
		}
	}
	return summaries, nil
}

// GetLatestLaunch fetches details of the latest SpaceX launch
func (c *SpaceXClient) GetLatestLaunch() (*Launch, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("%s/launches/latest", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var launch Launch
	if err := json.NewDecoder(resp.Body).Decode(&launch); err != nil {
		return nil, err
	}
	return &launch, nil
}

// GetRocket fetches details of a specific rocket by its ID
func (c *SpaceXClient) GetRocket(rocketID string) (*Rocket, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("%s/rockets/%s", c.baseURL, rocketID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rocket Rocket
	if err := json.NewDecoder(resp.Body).Decode(&rocket); err != nil {
		return nil, err
	}
	return &rocket, nil
}

// GetLaunchesSummary fetches launches from the last 3 years and summarizes by year and ship
func (c *SpaceXClient) GetLaunchesSummary() (*LaunchesSummary, error) {
	// Calculate the date 3 years ago
	threeYearsAgo := time.Now().AddDate(-3, 0, 0)
	startDate := threeYearsAgo.Format("2006-01-02T00:00:00.000Z")

	// Query launches from the last 3 years using SpaceX API query parameters
	// SpaceX API v4 supports querying with date_utc filter
	url := fmt.Sprintf("%s/launches/query", c.baseURL)
	
	// Create a POST request with query body for SpaceX API v4 query endpoint
	queryBody := map[string]interface{}{
		"query": map[string]interface{}{
			"date_utc": map[string]string{
				"$gte": startDate,
			},
		},
		"options": map[string]interface{}{
			"limit": 200, // Get up to 200 launches (should be enough for 3 years)
		},
	}

	jsonBody, err := json.Marshal(queryBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		log.Error().
			Str("method", "POST").
			Str("host", req.URL.Host).
			Dur("latency", duration).
			Err(err).
			Msg("API request failed")
		return nil, err
	}
	defer resp.Body.Close()

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
		Str("method", "POST").
		Str("host", req.URL.Host).
		Int("status", resp.StatusCode).
		Dur("latency", duration).
		Msg("API request completed")

	// Parse the response - SpaceX query endpoint returns {docs: [...], totalDocs: ..., ...}
	var queryResponse struct {
		Docs []Launch `json:"docs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&queryResponse); err != nil {
		return nil, err
	}

	// Initialize summary maps
	summary := &LaunchesSummary{
		ByYear: make(map[string]int),
		ByShip: make(map[string]int),
		Total:  len(queryResponse.Docs),
	}

	// Process each launch
	for _, launch := range queryResponse.Docs {
		// Parse the date to extract year
		launchDate, err := time.Parse(time.RFC3339, launch.DateUTC)
		if err != nil {
			// Try alternative format if RFC3339 fails
			launchDate, err = time.Parse("2006-01-02T15:04:05.000Z", launch.DateUTC)
			if err != nil {
				log.Warn().Str("date", launch.DateUTC).Msg("Failed to parse launch date")
				continue
			}
		}

		// Only count launches from the last 3 years (double-check filter)
		if launchDate.After(threeYearsAgo) || launchDate.Equal(threeYearsAgo) {
			year := launchDate.Format("2006")
			summary.ByYear[year]++

			// Count by ship
			if len(launch.Ships) > 0 {
				for _, shipID := range launch.Ships {
					if shipID != "" {
						summary.ByShip[shipID]++
					}
				}
			} else {
				// Count launches with no ships
				summary.ByShip["none"]++
			}
		}
	}

	return summary, nil
}
