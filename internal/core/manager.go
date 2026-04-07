package core

import (
	"context"
	"fmt"
	"sync"

	"autocal50/internal/meter"
	"autocal50/internal/pattern"
	"autocal50/internal/projector"
	"autocal50/internal/serialutil"
	"autocal50/internal/transport"
)

type Manager struct {
	ProjectorDrivers map[string]projector.DriverFactory
	MeterDrivers     map[string]meter.DriverFactory
	TransportDrivers map[string]transport.DriverFactory

	mu        sync.RWMutex
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
			"mock-meter":      meter.NewMockDriver,
			"argyll-spotread": meter.NewArgyllDriver,
		},
		TransportDrivers: map[string]transport.DriverFactory{
			"mock-transport": transport.NewMockDriver,
			"local-output":   transport.NewLocalDriver,
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
	m.mu.Lock()
	m.projector = d
	m.mu.Unlock()
	return nil
}

func (m *Manager) DisconnectProjector() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.projector == nil {
		return nil
	}
	err := m.projector.Disconnect()
	m.projector = nil
	return err
}

func (m *Manager) ProjectorCapabilities(ctx context.Context) (projector.Capabilities, error) {
	m.mu.RLock()
	p := m.projector
	m.mu.RUnlock()
	if p == nil {
		return projector.Capabilities{}, fmt.Errorf("no projector connected")
	}
	return p.Capabilities(ctx)
}

func (m *Manager) GetProjectorControl(ctx context.Context, id string) (any, error) {
	m.mu.RLock()
	p := m.projector
	m.mu.RUnlock()
	if p == nil {
		return nil, fmt.Errorf("no projector connected")
	}
	return p.GetControl(ctx, id)
}

func (m *Manager) SetProjectorControl(ctx context.Context, id string, value any) error {
	m.mu.RLock()
	p := m.projector
	m.mu.RUnlock()
	if p == nil {
		return fmt.Errorf("no projector connected")
	}
	return p.SetControl(ctx, id, value)
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
	m.mu.Lock()
	m.meter = d
	m.mu.Unlock()
	return nil
}

func (m *Manager) DisconnectMeter() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.meter == nil {
		return nil
	}
	err := m.meter.Disconnect()
	m.meter = nil
	return err
}

func (m *Manager) HasMeter() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.meter != nil
}

func (m *Manager) Measure(ctx context.Context) (meter.Reading, error) {
	m.mu.RLock()
	mt := m.meter
	m.mu.RUnlock()
	if mt == nil {
		return meter.Reading{}, fmt.Errorf("no meter connected")
	}
	return mt.Measure(ctx)
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
	m.mu.Lock()
	m.transport = d
	m.mu.Unlock()
	return nil
}

func (m *Manager) DisconnectTransport() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.transport == nil {
		return nil
	}
	err := m.transport.Disconnect()
	m.transport = nil
	return err
}

func (m *Manager) GetSignalFormats(ctx context.Context) ([]transport.SignalFormat, error) {
	m.mu.RLock()
	t := m.transport
	m.mu.RUnlock()
	if t == nil {
		return nil, fmt.Errorf("no transport connected")
	}
	return t.GetFormats(ctx)
}

func (m *Manager) ApplySignalFormat(ctx context.Context, f transport.SignalFormat) error {
	m.mu.RLock()
	t := m.transport
	m.mu.RUnlock()
	if t == nil {
		return fmt.Errorf("no transport connected")
	}
	return t.ApplyFormat(ctx, f)
}

func (m *Manager) CurrentSignalFormat(ctx context.Context) (transport.SignalFormat, error) {
	m.mu.RLock()
	t := m.transport
	m.mu.RUnlock()
	if t == nil {
		return transport.SignalFormat{}, fmt.Errorf("no transport connected")
	}
	return t.CurrentFormat(ctx)
}

// RenderPattern sends a pattern to the native output if a local transport
// is connected. Returns false if no native output is available (caller
// should fall back to the browser popup).
func (m *Manager) RenderPattern(p pattern.Pattern) bool {
	m.mu.RLock()
	t := m.transport
	m.mu.RUnlock()
	if t == nil {
		return false
	}
	local, ok := t.(*transport.LocalDriver)
	if !ok {
		return false
	}
	return local.Output().Render(p) == nil
}
