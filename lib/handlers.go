package lib

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// LoggingMiddleware wraps an http.HandlerFunc and logs request details
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next(w, r)

		// Log the request details after it's completed
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", r.URL.RawQuery).
			Dur("latency", time.Since(start)).
			Msg("Inbound")
	}
}

func HandleLatestLaunch(client SpaceXClientInterface) http.HandlerFunc {
	return LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		launch, err := client.GetLatestLaunch()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(launch)
	})
}

func HandleRocket(client SpaceXClientInterface) http.HandlerFunc {
	return LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		rocketID := r.URL.Query().Get("id")
		if rocketID == "" {
			http.Error(w, "rocket ID is required", http.StatusBadRequest)
			return
		}

		rocket, err := client.GetRocket(rocketID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rocket)
	})
}

func HandleListRockets(client SpaceXClientInterface) http.HandlerFunc {
	return LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		rockets, err := client.GetAllRockets()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rockets)
	})
}

func HandleNumbers(client NumbersClientInterface) http.HandlerFunc {
	return LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		mathFact, err := client.GetMathFact()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mathFact)
	})
}

func HandleNASA(client NASAClientInterface) http.HandlerFunc {
	return LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		apod, err := client.GetAPOD()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errorResponse := map[string]string{"error": "unknown error"}
			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apod)
	})
}

func HandleLaunchesSummary(client SpaceXClientInterface) http.HandlerFunc {
	return LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		summary, err := client.GetLaunchesSummary()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	})
}

func HandleRoot() http.HandlerFunc {
	return LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		endpoints := map[string]string{
			"/":                      "Shows this list of available endpoints",
			"/api/latest-launch":     "Get the latest SpaceX launch",
			"/api/rocket":            "Get a specific rocket by ID (use ?id=[rocket_id])",
			"/api/rockets":           "Get a list of all SpaceX rockets",
			"/api/numbers":           "Get a random math fact",
			"/api/nasa":              "Get NASA's Astronomy Picture of the Day",
			"/api/launches-summary":   "Get summary of launches by year and ship over the last 3 years",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(endpoints)
	})
}
