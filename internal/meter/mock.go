package meter

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type MockDriver struct{}

func NewMockDriver() Driver { return &MockDriver{} }

func (m *MockDriver) Name() string                                       { return "mock-meter" }
func (m *MockDriver) Connect(_ context.Context, _ map[string]any) error  { return nil }
func (m *MockDriver) Disconnect() error                                  { return nil }

func (m *MockDriver) Measure(_ context.Context) (Reading, error) {
	// Simulate D65-ish readings with noise
	jitter := rand.Float64()*0.01 - 0.005
	return Reading{
		X:         0.3127 + jitter,
		Y:         0.3290 + jitter,
		Z:         0.3583 + jitter,
		Luminance: 40 + math.Sin(float64(time.Now().UnixMilli())/1000)*5 + rand.Float64()*2,
		CCT:       6500 + (rand.Float64()*200 - 100),
		Timestamp: time.Now().UnixMilli(),
	}, nil
}
