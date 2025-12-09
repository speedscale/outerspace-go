package lib

// SpaceXClientInterface defines the interface for SpaceX API client
type SpaceXClientInterface interface {
	GetAllRockets() ([]RocketSummary, error)
	GetRocket(id string) (*Rocket, error)
	GetLatestLaunch() (*Launch, error)
	GetLaunchesSummary() (*LaunchesSummary, error)
}

// NumbersClientInterface defines the interface for Numbers API client
type NumbersClientInterface interface {
	GetMathFact() (*MathFact, error)
}

// NASAClientInterface defines the interface for NASA API client
type NASAClientInterface interface {
	GetAPOD() (*APOD, error)
}
