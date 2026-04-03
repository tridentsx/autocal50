package transport

import "context"

type MockDriver struct {
	current SignalFormat
}

func NewMockDriver() Driver { return &MockDriver{} }

func (m *MockDriver) Name() string { return "mock-transport" }

func (m *MockDriver) Connect(_ context.Context, _ map[string]any) error {
	m.current = SignalFormat{Resolution: "3840x2160", RefreshHz: 60, Encoding: "RGB", BitDepth: 10, HDR: false}
	return nil
}

func (m *MockDriver) Disconnect() error { return nil }

func (m *MockDriver) GetFormats(_ context.Context) ([]SignalFormat, error) {
	return []SignalFormat{
		{Resolution: "3840x2160", RefreshHz: 60, Encoding: "RGB", BitDepth: 10},
		{Resolution: "3840x2160", RefreshHz: 24, Encoding: "RGB", BitDepth: 12},
		{Resolution: "1920x1080", RefreshHz: 60, Encoding: "RGB", BitDepth: 8},
		{Resolution: "1920x1080", RefreshHz: 120, Encoding: "YCbCr422", BitDepth: 10},
	}, nil
}

func (m *MockDriver) ApplyFormat(_ context.Context, f SignalFormat) error {
	m.current = f
	return nil
}

func (m *MockDriver) CurrentFormat(_ context.Context) (SignalFormat, error) {
	return m.current, nil
}
