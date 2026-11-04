package lib

import (
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
	ByYear map[string]int `json:"by_year"`
	ByShip map[string]int `json:"by_ship"`
	Total  int            `json:"total"`
}

// NewSpaceXClient creates a new SpaceX API client
func NewSpaceXClient() *SpaceXClient {
	return &SpaceXClient{
		baseURL: upstreamURL("SPACEX_BASE_URL", "spacex"),
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *SpaceXClient) makeRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid SpaceX upstream URL")
	}
	return upstreamRequest(c.httpClient, "SpaceX", req)
}

// GetAllRockets fetches all rocket summaries
func (c *SpaceXClient) GetAllRockets() ([]RocketSummary, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("%s/rockets.json", c.baseURL))
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
	threeYearsAgo := time.Now().AddDate(-3, 0, 0)
	resp, err := c.makeRequest("GET", c.baseURL+"/launches.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var launches []Launch
	if err := json.NewDecoder(resp.Body).Decode(&launches); err != nil {
		return nil, err
	}

	// Initialize summary maps
	summary := &LaunchesSummary{
		ByYear: make(map[string]int),
		ByShip: make(map[string]int),
		Total:  0,
	}

	// Process each launch
	for _, launch := range launches {
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
			summary.Total++

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
