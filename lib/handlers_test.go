package lib

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Stub SpaceX client
type StubSpaceXClient struct {
	rockets              []RocketSummary
	rocket               *Rocket
	launch               *Launch
	launchesSummary      *LaunchesSummary
	rocketError          error
	launchError          error
	rocketsError         error
	launchesSummaryError error
}

func (s *StubSpaceXClient) GetAllRockets() ([]RocketSummary, error) {
	return s.rockets, s.rocketsError
}

func (s *StubSpaceXClient) GetRocket(id string) (*Rocket, error) {
	return s.rocket, s.rocketError
}

func (s *StubSpaceXClient) GetLatestLaunch() (*Launch, error) {
	return s.launch, s.launchError
}

func (s *StubSpaceXClient) GetLaunchesSummary() (*LaunchesSummary, error) {
	return s.launchesSummary, s.launchesSummaryError
}

// Stub Numbers client
type StubNumbersClient struct {
	fact  *MathFact
	error error
}

func (s *StubNumbersClient) GetMathFact() (*MathFact, error) {
	return s.fact, s.error
}

func TestHandleRoot(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	HandleRoot()(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var endpoints map[string]string
	json.Unmarshal(body, &endpoints)

	assert.Contains(t, endpoints, "/")
	assert.Contains(t, endpoints, "/api/rockets")
	assert.Contains(t, endpoints, "/api/rocket")
	assert.Contains(t, endpoints, "/api/latest-launch")
	assert.Contains(t, endpoints, "/api/numbers")
	assert.Contains(t, endpoints, "/api/launches-summary")
}

func TestHandleListRockets(t *testing.T) {
	stubClient := &StubSpaceXClient{
		rockets: []RocketSummary{
			{ID: "123", Name: "Falcon 9"},
			{ID: "456", Name: "Falcon Heavy"},
		},
	}

	req := httptest.NewRequest("GET", "/api/rockets", nil)
	w := httptest.NewRecorder()

	handler := HandleListRockets(stubClient)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var rockets []RocketSummary
	json.Unmarshal(body, &rockets)

	assert.Len(t, rockets, 2)
	assert.Equal(t, "Falcon 9", rockets[0].Name)
	assert.Equal(t, "456", rockets[1].ID)
}

func TestHandleRocket(t *testing.T) {
	stubClient := &StubSpaceXClient{
		rocket: &Rocket{
			ID:          "123",
			Name:        "Falcon 9",
			Description: "Orbital rocket",
			Height: struct {
				Meters float64 `json:"meters"`
			}{Meters: 70},
			Mass: struct {
				Kg int `json:"kg"`
			}{Kg: 549054},
		},
	}

	req := httptest.NewRequest("GET", "/api/rocket?id=123", nil)
	w := httptest.NewRecorder()

	handler := HandleRocket(stubClient)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var rocket Rocket
	json.Unmarshal(body, &rocket)

	assert.Equal(t, "Falcon 9", rocket.Name)
	assert.Equal(t, "123", rocket.ID)
}

func TestHandleRocket_MissingID(t *testing.T) {
	stubClient := &StubSpaceXClient{}

	req := httptest.NewRequest("GET", "/api/rocket", nil)
	w := httptest.NewRecorder()

	handler := HandleRocket(stubClient)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleRocket_Error(t *testing.T) {
	stubClient := &StubSpaceXClient{
		rocketError: errors.New("not found"),
	}

	req := httptest.NewRequest("GET", "/api/rocket?id=999", nil)
	w := httptest.NewRecorder()

	handler := HandleRocket(stubClient)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestHandleLatestLaunch(t *testing.T) {
	stubClient := &StubSpaceXClient{
		launch: &Launch{
			FlightNumber: 100,
			MissionName:  "Mission X",
			DateUTC:      "2023-01-01T12:00:00Z",
			Success:      true,
			Details:      "Test mission",
		},
	}

	req := httptest.NewRequest("GET", "/api/latest-launch", nil)
	w := httptest.NewRecorder()

	handler := HandleLatestLaunch(stubClient)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var launch Launch
	json.Unmarshal(body, &launch)

	assert.Equal(t, 100, launch.FlightNumber)
}

func TestHandleNumbers(t *testing.T) {
	stubClient := &StubNumbersClient{
		fact: &MathFact{
			Text:   "42 is the meaning of life",
			Number: 42,
			Found:  true,
			Type:   "math",
		},
	}

	req := httptest.NewRequest("GET", "/api/numbers", nil)
	w := httptest.NewRecorder()

	handler := HandleNumbers(stubClient)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var fact MathFact
	json.Unmarshal(body, &fact)

	assert.Equal(t, "42 is the meaning of life", fact.Text)
	assert.Equal(t, 42, fact.Number)
}

func TestHandleLaunchesSummary(t *testing.T) {
	stubClient := &StubSpaceXClient{
		launchesSummary: &LaunchesSummary{
			ByYear: map[string]int{
				"2023": 10,
				"2024": 15,
				"2025": 5,
			},
			ByShip: map[string]int{
				"ship1": 5,
				"ship2": 8,
				"none":  17,
			},
			Total: 30,
		},
	}

	req := httptest.NewRequest("GET", "/api/launches-summary", nil)
	w := httptest.NewRecorder()

	handler := HandleLaunchesSummary(stubClient)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var summary LaunchesSummary
	json.Unmarshal(body, &summary)

	assert.Equal(t, 30, summary.Total)
	assert.Equal(t, 10, summary.ByYear["2023"])
	assert.Equal(t, 15, summary.ByYear["2024"])
	assert.Equal(t, 5, summary.ByYear["2025"])
	assert.Equal(t, 5, summary.ByShip["ship1"])
	assert.Equal(t, 8, summary.ByShip["ship2"])
	assert.Equal(t, 17, summary.ByShip["none"])
}

func TestHandleLaunchesSummary_Error(t *testing.T) {
	stubClient := &StubSpaceXClient{
		launchesSummaryError: errors.New("API error"),
	}

	req := httptest.NewRequest("GET", "/api/launches-summary", nil)
	w := httptest.NewRecorder()

	handler := HandleLaunchesSummary(stubClient)
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
