package meter

import "context"

type Reading struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Z         float64 `json:"z"`
	Luminance float64 `json:"luminance"`
	CCT       float64 `json:"cct"`
	Timestamp int64   `json:"timestamp"`
}

type Driver interface {
	Name() string
	Connect(ctx context.Context, cfg map[string]any) error
	Disconnect() error
	Measure(ctx context.Context) (Reading, error)
}

type DriverFactory func() Driver
