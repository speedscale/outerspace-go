package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestClientCycleReachesEveryDependency(t *testing.T) {
	var paths []string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		paths = append(paths, req.URL.Path)
		switch req.URL.Path {
		case "/api/latest-launch":
			fmt.Fprint(w, `{"flight_number":2,"name":"Demo mission"}`)
		case "/api/rockets":
			fmt.Fprint(w, `[{"id":"demo-rocket","name":"Demo rocket"}]`)
		case "/api/rocket":
			if req.URL.Query().Get("id") != "demo-rocket" {
				t.Errorf("unexpected rocket id: %s", req.URL.RawQuery)
			}
			fmt.Fprint(w, `{"id":"demo-rocket","height":{"meters":70},"mass":{"kg":549054}}`)
		case "/api/numbers":
			fmt.Fprint(w, `{"number":42,"text":"Demo fact"}`)
		case "/api/nasa":
			fmt.Fprint(w, `{"title":"Demo astronomy"}`)
		default:
			http.NotFound(w, req)
		}
	}))
	defer backend.Close()
	if err := executeClientCycle(backend.URL); err != nil {
		t.Fatal(err)
	}
	want := []string{"/api/latest-launch", "/api/rockets", "/api/rocket", "/api/numbers", "/api/nasa"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
}
