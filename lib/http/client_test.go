package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientDecodesServerFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/latest-launch":
			fmt.Fprint(w, `{"name":"Demo mission","flight_number":2}`)
		case "/api/rocket":
			fmt.Fprint(w, `{"id":"demo","name":"Demo rocket","height":{"meters":70},"mass":{"kg":549054}}`)
		}
	}))
	defer server.Close()
	client := NewClient(server.URL)
	launch, err := client.GetLatestLaunch(context.Background())
	if err != nil || launch.MissionName != "Demo mission" {
		t.Fatalf("launch = %+v, %v", launch, err)
	}
	rocket, err := client.GetRocket(context.Background(), "demo")
	if err != nil || rocket.HeightMeters != 70 || rocket.MassKg != 549054 {
		t.Fatalf("rocket = %+v, %v", rocket, err)
	}
}
