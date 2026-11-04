package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpaceXClient_GetAllRockets(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/rockets.json", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Return a sample response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"123","name":"Falcon 9"},{"id":"456","name":"Falcon Heavy"}]`))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewSpaceXClient()
	client.baseURL = server.URL + "/v4"

	// Call the method
	rockets, err := client.GetAllRockets()

	// Assert results
	assert.NoError(t, err)
	assert.Len(t, rockets, 2)
	assert.Equal(t, "Falcon 9", rockets[0].Name)
	assert.Equal(t, "123", rockets[0].ID)
	assert.Equal(t, "Falcon Heavy", rockets[1].Name)
	assert.Equal(t, "456", rockets[1].ID)
}

func TestSpaceXClient_GetRocket(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/rockets/123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Return a sample response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","name":"Falcon 9","description":"Orbital rocket","height":{"meters":70},"mass":{"kg":549054}}`))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewSpaceXClient()
	client.baseURL = server.URL + "/v4"

	// Call the method
	rocket, err := client.GetRocket("123")

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, "Falcon 9", rocket.Name)
	assert.Equal(t, "123", rocket.ID)
	assert.Equal(t, "Orbital rocket", rocket.Description)
	assert.Equal(t, float64(70), rocket.Height.Meters)
	assert.Equal(t, 549054, rocket.Mass.Kg)
}

func TestSpaceXClient_GetLatestLaunch(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/launches/latest", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Return a sample response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"flight_number":100,"name":"Mission X","date_utc":"2023-01-01T12:00:00Z","success":true,"details":"Test mission"}`))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewSpaceXClient()
	client.baseURL = server.URL + "/v4"

	// Call the method
	launch, err := client.GetLatestLaunch()

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, 100, launch.FlightNumber)
	assert.Equal(t, "Mission X", launch.MissionName)
	assert.Equal(t, "2023-01-01T12:00:00Z", launch.DateUTC)
	assert.True(t, launch.Success)
	assert.Equal(t, "Test mission", launch.Details)
}
