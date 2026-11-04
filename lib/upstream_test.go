package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHTTPAPIWithFrozenUpstreams(t *testing.T) {
	data, err := os.ReadFile("testdata/outerspace.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures map[string]json.RawMessage
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			t.Errorf("upstream method: %s", req.Method)
		}
		if req.URL.Query().Get("api_key") != "" {
			t.Error("NASA credential sent to demo backend")
		}
		body, ok := fixtures[strings.TrimPrefix(req.URL.Path, "/outerspace/")]
		if !ok {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	defer upstream.Close()
	t.Setenv("DEMO_UPSTREAM_BASE_URL", upstream.URL+"/outerspace/")
	for _, name := range []string{"SPACEX_BASE_URL", "NUMBERS_BASE_URL", "NASA_BASE_URL"} {
		t.Setenv(name, "")
	}
	t.Setenv("NASA_API_KEY", "must-not-reach-demo")
	space, numbers, nasa := NewSpaceXClient(), NewNumbersClient(), NewNASAClient()
	cases := []struct {
		path    string
		handler http.HandlerFunc
		want    string
	}{
		{"/api/latest-launch", HandleLatestLaunch(space), `"name":"Demo mission two"`},
		{"/api/rockets", HandleListRockets(space), `"name":"Demo Falcon 9"`},
		{"/api/rocket?id=5e9d0d96eda699382d09d1ee", HandleRocket(space), `"kg":1420788`},
		{"/api/numbers", HandleNumbers(numbers), `"number":42`},
		{"/api/nasa", HandleNASA(nasa), `"title":"Frozen demo astronomy entry"`},
		{"/api/launches-summary", HandleLaunchesSummary(space), `"by_year":`},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			tc.handler.ServeHTTP(res, httptest.NewRequest("GET", tc.path, nil))
			if res.Code != 200 || !strings.Contains(res.Body.String(), tc.want) || !json.Valid(res.Body.Bytes()) {
				t.Fatalf("HTTP %d: %s", res.Code, res.Body.String())
			}
		})
	}
}

func TestUpstreamHTTPFailuresAreNotSuccess(t *testing.T) {
	for _, code := range []int{404, 429, 503} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(code)
				fmt.Fprint(w, `{"name":"must not be accepted","number":42}`)
			}))
			defer backend.Close()
			t.Setenv("DEMO_UPSTREAM_BASE_URL", backend.URL)
			for _, name := range []string{"SPACEX_BASE_URL", "NUMBERS_BASE_URL", "NASA_BASE_URL"} {
				t.Setenv(name, "")
			}
			checks := []func() error{
				func() error { _, err := NewSpaceXClient().GetLatestLaunch(); return err },
				func() error { _, err := NewSpaceXClient().GetAllRockets(); return err },
				func() error { _, err := NewSpaceXClient().GetRocket("missing"); return err },
				func() error { _, err := NewSpaceXClient().GetLaunchesSummary(); return err },
				func() error { _, err := NewNumbersClient().GetMathFact(); return err },
				func() error { _, err := NewNASAClient().GetAPOD(); return err },
			}
			for _, check := range checks {
				if err := check(); err == nil || !strings.Contains(err.Error(), fmt.Sprint(code)) {
					t.Fatalf("expected HTTP %d failure, got %v", code, err)
				}
			}
		})
	}
}

func TestNASAUsesFinalQuotaResponse(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("api_key") != "test&key" {
			t.Error("NASA API key was not encoded")
		}
		w.Header().Set("X-RateLimit-Remaining", "0")
		fmt.Fprint(w, `{"title":"Final allowed response"}`)
	}))
	defer backend.Close()
	t.Setenv("NASA_BASE_URL", backend.URL)
	t.Setenv("NASA_API_KEY", "test&key")
	data, err := NewNASAClient().GetAPOD()
	if err != nil || data.Title != "Final allowed response" {
		t.Fatalf("response = %v, %v", data, err)
	}
}

func TestNASARequestErrorsDoNotExposeAPIKey(t *testing.T) {
	t.Setenv("NASA_BASE_URL", "://bad-url")
	t.Setenv("NASA_API_KEY", "private-test-key")
	_, err := NewNASAClient().GetAPOD()
	if err == nil || strings.Contains(err.Error(), "private-test-key") {
		t.Fatalf("unsafe error: %v", err)
	}
}

func TestLaunchSummaryCountsOnlyValidRecentLaunches(t *testing.T) {
	recent := time.Now().AddDate(0, -1, 0).UTC().Format(time.RFC3339)
	old := time.Now().AddDate(-4, 0, 0).UTC().Format(time.RFC3339)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" || req.URL.Path != "/spacex/launches.json" {
			t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		fmt.Fprintf(w, `[{"date_utc":%q,"ships":["demo-recovery"]},{"date_utc":%q},{"date_utc":"bad"}]`, recent, old)
	}))
	defer backend.Close()
	t.Setenv("DEMO_UPSTREAM_BASE_URL", backend.URL)
	t.Setenv("SPACEX_BASE_URL", "")
	summary, err := NewSpaceXClient().GetLaunchesSummary()
	if err != nil || summary.Total != 1 || summary.ByShip["demo-recovery"] != 1 {
		t.Fatalf("summary = %+v, %v", summary, err)
	}
}
