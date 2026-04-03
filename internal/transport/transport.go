package transport

import "context"

type SignalFormat struct {
	Resolution string  `json:"resolution"`
	RefreshHz  float64 `json:"refreshHz"`
	Encoding   string  `json:"encoding"`
	BitDepth   int     `json:"bitDepth"`
	HDR        bool    `json:"hdr"`
}

type Driver interface {
	Name() string
	Connect(ctx context.Context, cfg map[string]any) error
	Disconnect() error
	GetFormats(ctx context.Context) ([]SignalFormat, error)
	ApplyFormat(ctx context.Context, f SignalFormat) error
	CurrentFormat(ctx context.Context) (SignalFormat, error)
}

type DriverFactory func() Driver
