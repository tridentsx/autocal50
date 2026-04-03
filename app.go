package main

import (
	"context"
	"time"

	"autocal50/internal/core"
	"autocal50/internal/display"
	"autocal50/internal/events"
	"autocal50/internal/meter"
	"autocal50/internal/pattern"
	"autocal50/internal/projector"
	"autocal50/internal/serialutil"
	"autocal50/internal/transport"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	core           *core.Manager
	activePattern  *pattern.Pattern
	patternDisplay string // xrandr output name for pattern window
}

func NewApp() *App {
	return &App{core: core.NewManager()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.emitMeasurements()
}

func (a *App) emitMeasurements() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		r, err := a.core.Measure(a.ctx)
		if err == nil {
			wailsRuntime.EventsEmit(a.ctx, events.MeasurementUpdate, r)
		}
	}
}

// Projector

func (a *App) ListProjectorDrivers() []string {
	return a.core.ListProjectorDrivers()
}

func (a *App) ConnectProjector(driver string, cfg map[string]any) error {
	return a.core.ConnectProjector(a.ctx, driver, cfg)
}

func (a *App) DisconnectProjector() error {
	return a.core.DisconnectProjector()
}

func (a *App) GetProjectorCapabilities() (projector.Capabilities, error) {
	return a.core.ProjectorCapabilities(a.ctx)
}

func (a *App) GetProjectorControl(id string) (any, error) {
	return a.core.GetProjectorControl(a.ctx, id)
}

func (a *App) SetProjectorControl(id string, value any) error {
	return a.core.SetProjectorControl(a.ctx, id, value)
}

// Meter

func (a *App) ListMeterDrivers() []string {
	return a.core.ListMeterDrivers()
}

func (a *App) ConnectMeter(driver string, cfg map[string]any) error {
	return a.core.ConnectMeter(a.ctx, driver, cfg)
}

func (a *App) DisconnectMeter() error {
	return a.core.DisconnectMeter()
}

func (a *App) Measure() (meter.Reading, error) {
	return a.core.Measure(a.ctx)
}

// Transport

func (a *App) ListTransportDrivers() []string {
	return a.core.ListTransportDrivers()
}

func (a *App) ConnectTransport(driver string, cfg map[string]any) error {
	return a.core.ConnectTransport(a.ctx, driver, cfg)
}

func (a *App) DisconnectTransport() error {
	return a.core.DisconnectTransport()
}

func (a *App) GetSignalFormats() ([]transport.SignalFormat, error) {
	return a.core.GetSignalFormats(a.ctx)
}

func (a *App) ApplySignalFormat(f transport.SignalFormat) error {
	return a.core.ApplySignalFormat(a.ctx, f)
}

func (a *App) CurrentSignalFormat() (transport.SignalFormat, error) {
	return a.core.CurrentSignalFormat(a.ctx)
}

// Patterns

func (a *App) GetPatternBattery() []pattern.Pattern {
	return pattern.Battery()
}

func (a *App) SetActivePattern(p pattern.Pattern) {
	a.activePattern = &p
	wailsRuntime.EventsEmit(a.ctx, "pattern:update", p)
}

func (a *App) GetActivePattern() *pattern.Pattern {
	return a.activePattern
}

// Display

func (a *App) ListDisplays() ([]display.Output, error) {
	return display.ListOutputs()
}

func (a *App) SetPatternDisplay(name string) {
	a.patternDisplay = name
}

func (a *App) GetPatternDisplay() string {
	return a.patternDisplay
}

func (a *App) GetDisplayEDID(name string) (*display.EDID, error) {
	return display.ReadEDID(name)
}

// Serial Ports

func (a *App) ListSerialPorts() ([]string, error) {
	return serialutil.ListPorts()
}
