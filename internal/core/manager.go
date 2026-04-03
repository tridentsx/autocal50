package core

import (
	"context"
	"fmt"

	"autocal50/internal/meter"
	"autocal50/internal/projector"
	"autocal50/internal/serialutil"
	"autocal50/internal/transport"
)

type Manager struct {
	ProjectorDrivers map[string]projector.DriverFactory
	MeterDrivers     map[string]meter.DriverFactory
	TransportDrivers map[string]transport.DriverFactory

	projector projector.Driver
	meter     meter.Driver
	transport transport.Driver
}

func NewManager() *Manager {
	return &Manager{
		ProjectorDrivers: map[string]projector.DriverFactory{
			"mock-projector": projector.NewMockDriver,
			"xgimi-rs232":   projector.NewXGIMIDriver(serialutil.Open),
		},
		MeterDrivers: map[string]meter.DriverFactory{
			"mock-meter": meter.NewMockDriver,
		},
		TransportDrivers: map[string]transport.DriverFactory{
			"mock-transport": transport.NewMockDriver,
		},
	}
}

// Projector

func (m *Manager) ListProjectorDrivers() []string {
	keys := make([]string, 0, len(m.ProjectorDrivers))
	for k := range m.ProjectorDrivers {
		keys = append(keys, k)
	}
	return keys
}

func (m *Manager) ConnectProjector(ctx context.Context, name string, cfg map[string]any) error {
	f, ok := m.ProjectorDrivers[name]
	if !ok {
		return fmt.Errorf("unknown projector driver: %s", name)
	}
	d := f()
	if err := d.Connect(ctx, cfg); err != nil {
		return err
	}
	m.projector = d
	return nil
}

func (m *Manager) DisconnectProjector() error {
	if m.projector == nil {
		return nil
	}
	err := m.projector.Disconnect()
	m.projector = nil
	return err
}

func (m *Manager) ProjectorCapabilities(ctx context.Context) (projector.Capabilities, error) {
	if m.projector == nil {
		return projector.Capabilities{}, fmt.Errorf("no projector connected")
	}
	return m.projector.Capabilities(ctx)
}

func (m *Manager) GetProjectorControl(ctx context.Context, id string) (any, error) {
	if m.projector == nil {
		return nil, fmt.Errorf("no projector connected")
	}
	return m.projector.GetControl(ctx, id)
}

func (m *Manager) SetProjectorControl(ctx context.Context, id string, value any) error {
	if m.projector == nil {
		return fmt.Errorf("no projector connected")
	}
	return m.projector.SetControl(ctx, id, value)
}

// Meter

func (m *Manager) ListMeterDrivers() []string {
	keys := make([]string, 0, len(m.MeterDrivers))
	for k := range m.MeterDrivers {
		keys = append(keys, k)
	}
	return keys
}

func (m *Manager) ConnectMeter(ctx context.Context, name string, cfg map[string]any) error {
	f, ok := m.MeterDrivers[name]
	if !ok {
		return fmt.Errorf("unknown meter driver: %s", name)
	}
	d := f()
	if err := d.Connect(ctx, cfg); err != nil {
		return err
	}
	m.meter = d
	return nil
}

func (m *Manager) DisconnectMeter() error {
	if m.meter == nil {
		return nil
	}
	err := m.meter.Disconnect()
	m.meter = nil
	return err
}

func (m *Manager) Measure(ctx context.Context) (meter.Reading, error) {
	if m.meter == nil {
		return meter.Reading{}, fmt.Errorf("no meter connected")
	}
	return m.meter.Measure(ctx)
}

// Transport

func (m *Manager) ListTransportDrivers() []string {
	keys := make([]string, 0, len(m.TransportDrivers))
	for k := range m.TransportDrivers {
		keys = append(keys, k)
	}
	return keys
}

func (m *Manager) ConnectTransport(ctx context.Context, name string, cfg map[string]any) error {
	f, ok := m.TransportDrivers[name]
	if !ok {
		return fmt.Errorf("unknown transport driver: %s", name)
	}
	d := f()
	if err := d.Connect(ctx, cfg); err != nil {
		return err
	}
	m.transport = d
	return nil
}

func (m *Manager) DisconnectTransport() error {
	if m.transport == nil {
		return nil
	}
	err := m.transport.Disconnect()
	m.transport = nil
	return err
}

func (m *Manager) GetSignalFormats(ctx context.Context) ([]transport.SignalFormat, error) {
	if m.transport == nil {
		return nil, fmt.Errorf("no transport connected")
	}
	return m.transport.GetFormats(ctx)
}

func (m *Manager) ApplySignalFormat(ctx context.Context, f transport.SignalFormat) error {
	if m.transport == nil {
		return fmt.Errorf("no transport connected")
	}
	return m.transport.ApplyFormat(ctx, f)
}

func (m *Manager) CurrentSignalFormat(ctx context.Context) (transport.SignalFormat, error) {
	if m.transport == nil {
		return transport.SignalFormat{}, fmt.Errorf("no transport connected")
	}
	return m.transport.CurrentFormat(ctx)
}
