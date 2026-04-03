package projector

import "context"

type Control struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Type     string   `json:"type"` // slider, enum, toggle
	Min      float64  `json:"min"`
	Max      float64  `json:"max"`
	Step     float64  `json:"step"`
	Options  []string `json:"options"`
	ReadOnly bool     `json:"readOnly"`
}

type Capabilities struct {
	Controls      []Control `json:"controls"`
	SupportsModes []string  `json:"supportsModes"`
}

type Driver interface {
	Name() string
	Connect(ctx context.Context, cfg map[string]any) error
	Disconnect() error
	Capabilities(ctx context.Context) (Capabilities, error)
	GetControl(ctx context.Context, id string) (any, error)
	SetControl(ctx context.Context, id string, value any) error
}

type DriverFactory func() Driver
