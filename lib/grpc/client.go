package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a gRPC client for the LaunchService
type Client struct {
	conn   *grpc.ClientConn
	client LaunchServiceClient
}

// NewClient creates a new gRPC client
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := NewLaunchServiceClient(conn)
	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// GetLatestLaunch calls the GetLatestLaunch RPC
func (c *Client) GetLatestLaunch(ctx context.Context) (*Launch, error) {
	req := &LatestLaunchRequest{}
	return c.client.GetLatestLaunch(ctx, req)
}

// GetRocket calls the GetRocket RPC
func (c *Client) GetRocket(ctx context.Context, id string) (*Rocket, error) {
	req := &GetRocketRequest{Id: id}
	return c.client.GetRocket(ctx, req)
}

// GetRockets calls the GetRockets RPC
func (c *Client) GetRockets(ctx context.Context) (*GetRocketsResponse, error) {
	req := &GetRocketsRequest{}
	return c.client.GetRockets(ctx, req)
}

// GetMathFact calls the GetMathFact RPC
func (c *Client) GetMathFact(ctx context.Context) (*MathFact, error) {
	req := &GetMathFactRequest{}
	return c.client.GetMathFact(ctx, req)
}

// GetAPOD calls the GetAPOD RPC
func (c *Client) GetAPOD(ctx context.Context, date string) (*APOD, error) {
	req := &GetAPODRequest{Date: date}
	return c.client.GetAPOD(ctx, req)
}

// Example usage:
func Example() {
	// Create a new client
	client, err := NewClient("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Call GetLatestLaunch
	launch, err := client.GetLatestLaunch(context.Background())
	if err != nil {
		log.Fatalf("Failed to get latest launch: %v", err)
	}
	log.Printf("Latest launch: %v", launch)

	// Call GetRocket
	rocket, err := client.GetRocket(context.Background(), "falcon9")
	if err != nil {
		log.Fatalf("Failed to get rocket: %v", err)
	}
	log.Printf("Rocket: %v", rocket)

	// Call GetRockets
	rockets, err := client.GetRockets(context.Background())
	if err != nil {
		log.Fatalf("Failed to get rockets: %v", err)
	}
	log.Printf("Rockets: %v", rockets)

	// Call GetMathFact
	mathFact, err := client.GetMathFact(context.Background())
	if err != nil {
		log.Fatalf("Failed to get math fact: %v", err)
	}
	log.Printf("Math fact: %v", mathFact)
}
