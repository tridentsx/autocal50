package projector

import (
	"context"
	"fmt"
)

type MockDriver struct {
	controls map[string]any
}

func NewMockDriver() Driver { return &MockDriver{controls: make(map[string]any)} }

func (m *MockDriver) Name() string { return "mock-projector" }

func (m *MockDriver) Connect(_ context.Context, _ map[string]any) error {
	m.controls = map[string]any{
		"brightness": float64(50), "contrast": float64(50),
		"rgb_gain_r": float64(128), "rgb_gain_g": float64(128), "rgb_gain_b": float64(128),
		"rgb_bias_r": float64(0), "rgb_bias_g": float64(0), "rgb_bias_b": float64(0),
		"gamma_preset": "2.2", "picture_mode": "Cinema", "dynamic_contrast": false,
	}
	return nil
}

func (m *MockDriver) Disconnect() error { m.controls = nil; return nil }

func (m *MockDriver) Capabilities(_ context.Context) (Capabilities, error) {
	return Capabilities{
		Controls: []Control{
			{ID: "brightness", Label: "Brightness", Type: "slider", Min: 0, Max: 100, Step: 1},
			{ID: "contrast", Label: "Contrast", Type: "slider", Min: 0, Max: 100, Step: 1},
			{ID: "rgb_gain_r", Label: "Red Gain", Type: "slider", Min: 0, Max: 255, Step: 1},
			{ID: "rgb_gain_g", Label: "Green Gain", Type: "slider", Min: 0, Max: 255, Step: 1},
			{ID: "rgb_gain_b", Label: "Blue Gain", Type: "slider", Min: 0, Max: 255, Step: 1},
			{ID: "rgb_bias_r", Label: "Red Bias", Type: "slider", Min: -50, Max: 50, Step: 1},
			{ID: "rgb_bias_g", Label: "Green Bias", Type: "slider", Min: -50, Max: 50, Step: 1},
			{ID: "rgb_bias_b", Label: "Blue Bias", Type: "slider", Min: -50, Max: 50, Step: 1},
			{ID: "gamma_preset", Label: "Gamma", Type: "enum", Options: []string{"1.8", "2.0", "2.2", "2.4", "BT.1886"}},
			{ID: "picture_mode", Label: "Picture Mode", Type: "enum", Options: []string{"Cinema", "Natural", "Vivid", "User"}},
			{ID: "dynamic_contrast", Label: "Dynamic Contrast", Type: "toggle"},
		},
		SupportsModes: []string{"Cinema", "Natural", "Vivid", "User"},
	}, nil
}

func (m *MockDriver) GetControl(_ context.Context, id string) (any, error) {
	v, ok := m.controls[id]
	if !ok {
		return nil, fmt.Errorf("unknown control: %s", id)
	}
	return v, nil
}

func (m *MockDriver) SetControl(_ context.Context, id string, value any) error {
	if _, ok := m.controls[id]; !ok {
		return fmt.Errorf("unknown control: %s", id)
	}
	m.controls[id] = value
	return nil
}
