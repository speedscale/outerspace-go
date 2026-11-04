package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NumbersClient handles API calls to Numbers API
type NumbersClient struct {
	baseURL    string
	httpClient *http.Client
}

// RocketSummary provides a simplified view of rocket data
type MathFact struct {
	Text   string `json:"text"`
	Number int    `json:"number"`
	Found  bool   `json:"found"`
	Type   string `json:"type"`
}

// NewNumbersClient creates a new Numbers API client
func NewNumbersClient() *NumbersClient {
	return &NumbersClient{
		baseURL: upstreamURL("NUMBERS_BASE_URL", "numbers"),
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *NumbersClient) makeRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid Numbers upstream URL")
	}
	return upstreamRequest(c.httpClient, "Numbers", req)
}

func (c *NumbersClient) GetMathFact() (*MathFact, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("%s/random/math?json", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var mathFact MathFact
	if err := json.NewDecoder(resp.Body).Decode(&mathFact); err != nil {
		return nil, err
	}
	return &mathFact, nil
}
