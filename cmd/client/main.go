package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"outerspace-go/lib/grpc"
)

var (
	Version   string = "dev"
	BuildTime string = "unknown"
)

func main() {
	fmt.Printf("outerspace-go client version %s (built at %s)\n", Version, BuildTime)
	
	// Get server address from environment variable or use default
	serverAddr := os.Getenv("GRPC_SERVER_ADDR")
	if serverAddr == "" {
		host, err := os.Hostname()
		if err != nil {
			// grpc will not use a proxy on localhost so we want to use a real hostname
			host = "127.0.0.1"
		}
		serverAddr = host + ":50053"
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

	// Start APOD polling in a separate goroutine (every 10 minutes)
	apodInterval := 10 * time.Minute
	go func() {
		apodTicker := time.NewTicker(apodInterval)
		defer apodTicker.Stop()
		
		// Call APOD immediately on start
		if err := callAPOD(serverAddr); err != nil {
			log.Printf("Initial APOD call failed: %v", err)
		}
		
		for range apodTicker.C {
			if err := callAPOD(serverAddr); err != nil {
				log.Printf("APOD call failed: %v", err)
			}
		}
	}()

	// Main loop for other calls
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
	client, err := grpc.NewClient(serverAddr)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get latest launch
	fmt.Println("\n=== Getting Latest Launch ===")
	launch, err := client.GetLatestLaunch(ctx)
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
	rockets, err := client.GetRockets(ctx)
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
		rocket, err := client.GetRocket(ctx, rocketID)
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
	mathFact, err := client.GetMathFact(ctx)
	if err != nil {
		log.Printf("Failed to get math fact: %v", err)
	} else {
		fmt.Printf("Number: %d\n", mathFact.Number)
		fmt.Printf("Type: %s\n", mathFact.Type)
		fmt.Printf("Fact: %s\n", mathFact.Text)
		fmt.Printf("Found: %v\n", mathFact.Found)
	}

	return nil
}

func callAPOD(serverAddr string) error {
	client, err := grpc.NewClient(serverAddr)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get today's date in YYYY-MM-DD format
	today := time.Now().Format("2006-01-02")
	
	fmt.Printf("\n[%s] === Getting NASA APOD for %s ===\n", time.Now().Format(time.RFC3339), today)
	apod, err := client.GetAPOD(ctx, today)
	if err != nil {
		return fmt.Errorf("failed to get APOD: %w", err)
	}

	fmt.Printf("Date: %s\n", apod.Date)
	fmt.Printf("Title: %s\n", apod.Title)
	fmt.Printf("Media Type: %s\n", apod.MediaType)
	fmt.Printf("URL: %s\n", apod.Url)
	if apod.Hdurl != "" {
		fmt.Printf("HD URL: %s\n", apod.Hdurl)
	}
	fmt.Printf("Explanation: %.200s", apod.Explanation)
	if len(apod.Explanation) > 200 {
		fmt.Print("...")
	}
	fmt.Println()

	return nil
}
