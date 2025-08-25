package grpc

import (
	"context"
	"log"
	"net"

	"outerspace-go/lib"

	"google.golang.org/grpc"
)

// Server implements the LaunchService
type Server struct {
	UnimplementedLaunchServiceServer
	spaceClient   *lib.SpaceXClient
	numbersClient *lib.NumbersClient
	nasaClient    *lib.NASAClient
}

// NewServer creates a new gRPC server
func NewServer(spaceClient *lib.SpaceXClient, numbersClient *lib.NumbersClient, nasaClient *lib.NASAClient) *Server {
	return &Server{
		spaceClient:   spaceClient,
		numbersClient: numbersClient,
		nasaClient:    nasaClient,
	}
}

// GetLatestLaunch implements the LaunchService interface
func (s *Server) GetLatestLaunch(ctx context.Context, req *LatestLaunchRequest) (*Launch, error) {
	launch, err := s.spaceClient.GetLatestLaunch()
	if err != nil {
		return nil, err
	}

	return &Launch{
		FlightNumber: int32(launch.FlightNumber),
		MissionName:  launch.MissionName,
		DateUtc:      launch.DateUTC,
		Success:      launch.Success,
		Details:      launch.Details,
	}, nil
}

// GetRocket implements the LaunchService interface
func (s *Server) GetRocket(ctx context.Context, req *GetRocketRequest) (*Rocket, error) {
	rocket, err := s.spaceClient.GetRocket(req.Id)
	if err != nil {
		return nil, err
	}

	return &Rocket{
		Id:           rocket.ID,
		Name:         rocket.Name,
		Description:  rocket.Description,
		HeightMeters: rocket.Height.Meters,
		MassKg:       int32(rocket.Mass.Kg),
	}, nil
}

// GetRockets implements the LaunchService interface
func (s *Server) GetRockets(ctx context.Context, req *GetRocketsRequest) (*GetRocketsResponse, error) {
	rockets, err := s.spaceClient.GetAllRockets()
	if err != nil {
		return nil, err
	}

	response := &GetRocketsResponse{
		Rockets: make([]*RocketSummary, len(rockets)),
	}
	for i, rocket := range rockets {
		response.Rockets[i] = &RocketSummary{
			Id:   rocket.ID,
			Name: rocket.Name,
		}
	}
	return response, nil
}

// GetMathFact implements the LaunchService interface
func (s *Server) GetMathFact(ctx context.Context, req *GetMathFactRequest) (*MathFact, error) {
	mathFact, err := s.numbersClient.GetMathFact()
	if err != nil {
		return nil, err
	}

	return &MathFact{
		Text:   mathFact.Text,
		Number: int32(mathFact.Number),
		Found:  mathFact.Found,
		Type:   mathFact.Type,
	}, nil
}

// GetAPOD implements the LaunchService interface
func (s *Server) GetAPOD(ctx context.Context, req *GetAPODRequest) (*APOD, error) {
	apod, err := s.nasaClient.GetAPOD(req.Date)
	if err != nil {
		return nil, err
	}

	return &APOD{
		Date:           apod.Date,
		Title:          apod.Title,
		Explanation:    apod.Explanation,
		Url:            apod.Url,
		Hdurl:          apod.HDUrl,
		MediaType:      apod.MediaType,
		ServiceVersion: apod.ServiceVersion,
	}, nil
}

// StartServer starts the gRPC server
func StartServer(spaceClient *lib.SpaceXClient, numbersClient *lib.NumbersClient, nasaClient *lib.NASAClient, port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	RegisterLaunchServiceServer(s, NewServer(spaceClient, numbersClient, nasaClient))

	log.Printf("Starting gRPC server on %s", port)
	return s.Serve(lis)
}
