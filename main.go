package main

import (
	"log"
	"net/http"

	"outerspace-go/lib"
	"outerspace-go/lib/grpc"
	"outerspace-go/lib/logger"
)

var (
	Version   string = "dev"
	BuildTime string = "unknown"
)

func main() {
	// Initialize logger
	logger.Init()

	log.Printf("outerspace-go version %s (built at %s)", Version, BuildTime)

	spaceClient := lib.NewSpaceXClient()
	numbersClient := lib.NewNumbersClient()
	nasaClient := lib.NewNASAClient()

	// Define routes
	http.HandleFunc("/", lib.HandleRoot())
	http.HandleFunc("/api/latest-launch", lib.HandleLatestLaunch(spaceClient))
	http.HandleFunc("/api/rocket", lib.HandleRocket(spaceClient))
	http.HandleFunc("/api/rockets", lib.HandleListRockets(spaceClient))
	http.HandleFunc("/api/numbers", lib.HandleNumbers(numbersClient))
	http.HandleFunc("/api/nasa", lib.HandleNASA(nasaClient))
	http.HandleFunc("/api/launches-summary", lib.HandleLaunchesSummary(spaceClient))

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Start gRPC server
	if err := grpc.StartServer(spaceClient, numbersClient, ":50053"); err != nil {
		log.Fatal(err)
	}
}
