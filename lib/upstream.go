package lib

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func upstreamURL(override, path string) string {
	if value := os.Getenv(override); value != "" {
		return strings.TrimRight(value, "/")
	}
	base := os.Getenv("DEMO_UPSTREAM_BASE_URL")
	if base == "" {
		base = "https://demo-api.trafficreplay.com/outerspace"
	}
	return strings.TrimRight(base, "/") + "/" + path
}

func upstreamRequest(client *http.Client, name string, req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		// Request errors can contain credential-bearing URLs.
		log.Error().Str("upstream", name).Str("host", req.URL.Host).
			Dur("latency", time.Since(start)).Msg("Upstream request failed")
		return nil, fmt.Errorf("%s upstream request failed", name)
	}
	log.Info().Str("upstream", name).Str("host", req.URL.Host).
		Int("status", resp.StatusCode).Dur("latency", time.Since(start)).Msg("Outbound")
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s upstream returned HTTP %d", name, resp.StatusCode)
	}
	return resp, nil
}
