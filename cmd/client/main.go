package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"outerspace-go/lib/http"
)

var (
	Version   string = "dev"
	BuildTime string = "unknown"
)

func main() {
	fmt.Printf("outerspace-go client version %s (built at %s)\n", Version, BuildTime)

	// Get server address from environment variable or use default
	serverAddr := os.Getenv("HTTP_SERVER_ADDR")
	if serverAddr == "" {
		host, err := os.Hostname()
		if err != nil {
			host = "127.0.0.1"
		}
		serverAddr = "http://" + host + ":80"
	}

	// Get polling interval from environment variable or use default (30 minutes)
	intervalStr := os.Getenv("POLL_INTERVAL")
	interval := 30 * time.Minute
	if intervalStr != "" {
		if parsedInterval, err := time.ParseDuration(intervalStr); err == nil {
			interval = parsedInterval
		} else if seconds, err := strconv.Atoi(intervalStr); err == nil {
			interval = time.Duration(seconds) * time.Second
		}
	}

	fmt.Printf("Server: %s, Poll interval: %v\n", serverAddr, interval)

	// Main loop
	for {
		fmt.Printf("\n[%s] Starting client execution cycle\n", time.Now().Format(time.RFC3339))

		if err := executeClientCycle(serverAddr); err != nil {
			log.Printf("Client cycle failed: %v", err)
		}

		fmt.Printf("[%s] Client cycle completed, sleeping for %v\n", time.Now().Format(time.RFC3339), interval)
		time.Sleep(interval)
	}
}

func executeClientCycle(serverAddr string) error {
	// Create a new client
	client := http.NewClient(serverAddr)
	defer client.Close()

	// Get latest launch
	fmt.Println("\n=== Getting Latest Launch ===")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	launch, err := client.GetLatestLaunch(ctx)
	cancel()
	if err != nil {
		log.Printf("Failed to get latest launch: %v", err)
	} else {
		fmt.Printf("Flight Number: %d\n", launch.FlightNumber)
		fmt.Printf("Mission Name: %s\n", launch.MissionName)
		fmt.Printf("Date (UTC): %s\n", launch.DateUtc)
		fmt.Printf("Success: %v\n", launch.Success)
		fmt.Printf("Details: %s\n", launch.Details)
	}

	// Get all rockets
	fmt.Println("\n=== Getting All Rockets ===")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	rockets, err := client.GetRockets(ctx)
	cancel()
	if err != nil {
		log.Printf("Failed to get rockets: %v", err)
	} else {
		fmt.Println("Available Rockets:")
		for _, rocket := range rockets.Rockets {
			fmt.Printf("- %s (ID: %s)\n", rocket.Name, rocket.Id)
		}
	}

	// Get specific rocket details (using the first rocket ID from the list)
	if rockets != nil && len(rockets.Rockets) > 0 {
		fmt.Println("\n=== Getting Rocket Details ===")
		rocketID := rockets.Rockets[0].Id
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		rocket, err := client.GetRocket(ctx, rocketID)
		cancel()
		if err != nil {
			log.Printf("Failed to get rocket details: %v", err)
		} else {
			fmt.Printf("Rocket Name: %s\n", rocket.Name)
			fmt.Printf("Description: %s\n", rocket.Description)
			fmt.Printf("Height: %.2f meters\n", rocket.HeightMeters)
			fmt.Printf("Mass: %d kg\n", rocket.MassKg)
		}
	}

	// Get math fact
	fmt.Println("\n=== Getting Math Fact ===")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	mathFact, err := client.GetMathFact(ctx)
	cancel()
	if err != nil {
		log.Printf("Failed to get math fact: %v", err)
	} else {
		fmt.Printf("Number: %d\n", mathFact.Number)
		fmt.Printf("Type: %s\n", mathFact.Type)
		fmt.Printf("Fact: %s\n", mathFact.Text)
		fmt.Printf("Found: %v\n", mathFact.Found)
	}

	// Get NASA data
	fmt.Println("\n=== Getting NASA Astronomy Picture of the Day ===")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	nasaData, err := client.GetNASAData(ctx)
	cancel()
	if err != nil {
		log.Printf("Failed to get NASA data: %v", err)
	} else {
		fmt.Printf("Title: %s\n", nasaData.Title)
		fmt.Printf("Date: %s\n", nasaData.Date)
		fmt.Printf("Media Type: %s\n", nasaData.MediaType)
		fmt.Printf("URL: %s\n", nasaData.URL)
		fmt.Printf("Explanation: %s\n", nasaData.Explanation)
	}

	return nil
}
