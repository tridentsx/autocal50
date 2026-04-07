package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"autocal50/internal/calibration"
	"autocal50/internal/core"
	"autocal50/internal/display"
	"autocal50/internal/events"
	"autocal50/internal/meter"
	"autocal50/internal/pattern"
	"autocal50/internal/projector"
	"autocal50/internal/serialutil"
	"autocal50/internal/store"
	"autocal50/internal/transport"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	core           *core.Manager
	store          *store.Store
	mu             sync.RWMutex
	activePattern  *pattern.Pattern
	activeSession  *store.Session
	patternDisplay string
}

func NewApp() *App {
	st, _ := store.New(store.DefaultDir())
	return &App{core: core.NewManager(), store: st}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.emitMeasurements()
}

func (a *App) emitMeasurements() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if !a.core.HasMeter() {
			continue
		}
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

func (a *App) DarkCalMeter() error {
	return a.core.DarkCalMeter(a.ctx)
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
	a.mu.Lock()
	a.activePattern = &p
	a.mu.Unlock()

	// Try native output first (DRM/DXGI/Metal), fall back to browser popup.
	if !a.core.RenderPattern(p) {
		wailsRuntime.EventsEmit(a.ctx, "pattern:update", p)
	}
}

func (a *App) GetActivePattern() *pattern.Pattern {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.activePattern
}

// Display

func (a *App) ListDisplays() ([]display.Output, error) {
	return display.ListOutputs()
}

func (a *App) SetPatternDisplay(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.patternDisplay = name
}

func (a *App) GetPatternDisplay() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.patternDisplay
}

func (a *App) GetDisplayEDID(name string) (*display.EDID, error) {
	return display.ReadEDID(name)
}

// Serial Ports

func (a *App) ListSerialPorts() ([]string, error) {
	return serialutil.ListPorts()
}

// Calibration

func (a *App) GetCalibrationSteps(standard string) ([]calibration.Step, error) {
	caps, err := a.core.ProjectorCapabilities(a.ctx)
	if err != nil {
		// No projector — return all steps (user may adjust via OSD)
		caps = projector.Capabilities{}
	}
	return calibration.BuildRoutine(standard, caps), nil
}

func (a *App) EvaluateMeasurement(target calibration.ColorTarget, tolerance float64) (calibration.Advice, error) {
	reading, err := a.core.Measure(a.ctx)
	if err != nil {
		return calibration.Advice{}, err
	}
	return calibration.Evaluate(target, reading, tolerance), nil
}

func (a *App) GetCalibrationStandards() map[string]calibration.Standard {
	return calibration.Standards
}

// Sessions

func (a *App) StartSession(standard string) (*store.Session, error) {
	projName := ""
	meterName := ""
	sess := store.NewSession(standard, projName, meterName)
	a.mu.Lock()
	a.activeSession = sess
	a.mu.Unlock()
	return sess, nil
}

func (a *App) GetActiveSession() *store.Session {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.activeSession
}

func (a *App) SaveSession() error {
	a.mu.RLock()
	sess := a.activeSession
	a.mu.RUnlock()
	if sess == nil {
		return fmt.Errorf("no active session")
	}
	return a.store.Save(sess)
}

func (a *App) ListSessions() ([]store.Session, error) {
	return a.store.List()
}

func (a *App) LoadSession(id string) (*store.Session, error) {
	sess, err := a.store.Load(id)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.activeSession = sess
	a.mu.Unlock()
	return sess, nil
}

func (a *App) DeleteSession(id string) error {
	return a.store.Delete(id)
}

func (a *App) RecordMeasurement(step string, target calibration.ColorTarget, tolerance float64) (*store.Measurement, error) {
	reading, err := a.core.Measure(a.ctx)
	if err != nil {
		return nil, err
	}
	advice := calibration.Evaluate(target, reading, tolerance)
	m := store.Measurement{
		PatternID: "",
		Target:    target,
		Reading:   reading,
		DeltaE:    advice.DeltaE,
		Passed:    advice.Passed,
		Timestamp: time.Now(),
	}
	a.mu.RLock()
	pat := a.activePattern
	a.mu.RUnlock()
	if pat != nil {
		m.PatternID = pat.ID
	}

	a.mu.Lock()
	if a.activeSession != nil {
		a.activeSession.AddMeasurement(step, m)
	}
	a.mu.Unlock()

	// Auto-save after each measurement.
	if a.activeSession != nil {
		_ = a.store.Save(a.activeSession)
	}

	return &m, nil
}

func (a *App) SnapshotControls(label string) error {
	caps, err := a.core.ProjectorCapabilities(a.ctx)
	if err != nil {
		return err
	}
	controls := make(map[string]any)
	for _, c := range caps.Controls {
		if v, err := a.core.GetProjectorControl(a.ctx, c.ID); err == nil {
			controls[c.ID] = v
		}
	}
	a.mu.Lock()
	if a.activeSession != nil {
		a.activeSession.AddSnapshot(label, controls)
	}
	a.mu.Unlock()
	return nil
}

// Automated Calibration

func (a *App) buildEngine() *calibration.Engine {
	return &calibration.Engine{
		ShowPattern: func(patternID string) error {
			// Find pattern by ID from battery.
			for _, p := range pattern.Battery() {
				if p.ID == patternID {
					a.SetActivePattern(p)
					return nil
				}
			}
			return fmt.Errorf("pattern %s not found", patternID)
		},
		Measure: func(ctx context.Context) (meter.Reading, error) {
			return a.core.Measure(ctx)
		},
		GetControl: func(ctx context.Context, id string) (float64, error) {
			v, err := a.core.GetProjectorControl(ctx, id)
			if err != nil {
				return 0, err
			}
			switch val := v.(type) {
			case float64:
				return val, nil
			case int:
				return float64(val), nil
			case json.Number:
				return val.Float64()
			default:
				return 0, fmt.Errorf("control %s is not numeric", id)
			}
		},
		SetControl: func(ctx context.Context, id string, value float64) error {
			return a.core.SetProjectorControl(ctx, id, value)
		},
		OnProgress: func(ev calibration.ProgressEvent) {
			wailsRuntime.EventsEmit(a.ctx, events.CalibrationProgress, ev)
		},
	}
}

func (a *App) RunPreCal() (*calibration.PreCalResult, error) {
	eng := a.buildEngine()
	result, err := eng.RunPreCal(a.ctx)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	if a.activeSession != nil {
		a.activeSession.PreCal = &store.PreCalData{
			BlackLevel:    result.BlackLevel,
			PeakLuminance: result.PeakLum,
			ContrastRatio: result.ContrastRat,
			ClipBlack:     result.ClipBlack,
			ClipWhite:     result.ClipWhite,
		}
	}
	a.mu.Unlock()
	_ = a.SaveSession()
	return result, nil
}

func (a *App) RunWhiteBalance(standard string, tolerance float64) error {
	std, ok := calibration.Standards[standard]
	if !ok {
		return fmt.Errorf("unknown standard: %s", standard)
	}
	_ = a.SnapshotControls("before whitebalance")
	eng := a.buildEngine()
	err := eng.RunWhiteBalance(a.ctx, std, tolerance)
	_ = a.SnapshotControls("after whitebalance")
	_ = a.SaveSession()
	return err
}

func (a *App) RunGammaSweep(standard string) ([]calibration.GammaPoint, error) {
	std, ok := calibration.Standards[standard]
	if !ok {
		return nil, fmt.Errorf("unknown standard: %s", standard)
	}
	eng := a.buildEngine()
	pts, err := eng.RunGammaSweep(a.ctx, std, 20)
	_ = a.SaveSession()
	return pts, err
}

func (a *App) RunVerification(standard string) (map[string]float64, error) {
	std, ok := calibration.Standards[standard]
	if !ok {
		return nil, fmt.Errorf("unknown standard: %s", standard)
	}
	eng := a.buildEngine()
	avg, max, err := eng.RunVerification(a.ctx, std)
	if err != nil {
		return nil, err
	}
	_ = a.SaveSession()
	return map[string]float64{"avgDeltaE": avg, "maxDeltaE": max}, nil
}

// Profile Generation

func (a *App) GenerateICCProfile(standard string) (*calibration.ProfileResult, error) {
	std, ok := calibration.Standards[standard]
	if !ok {
		return nil, fmt.Errorf("unknown standard: %s", standard)
	}

	eng := a.buildEngine()

	// Measure white.
	eng.ShowPattern("gray-win-100")
	time.Sleep(time.Second)
	white, err := a.core.Measure(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("measure white: %w", err)
	}

	// Measure primaries.
	var primaries [3]meter.Reading
	for i, pid := range []string{"win-Red", "win-Green", "win-Blue"} {
		eng.ShowPattern(pid)
		time.Sleep(time.Second)
		r, err := a.core.Measure(a.ctx)
		if err != nil {
			return nil, fmt.Errorf("measure %s: %w", pid, err)
		}
		primaries[i] = r
	}

	// Gamma sweep.
	gamma, err := eng.RunGammaSweep(a.ctx, std, 20)
	if err != nil {
		return nil, fmt.Errorf("gamma sweep: %w", err)
	}

	// Generate profile.
	desc := fmt.Sprintf("AutoCal50 %s", std.Label)
	outDir := store.DefaultDir() + "/../profiles"
	result, err := calibration.GenerateProfile(primaries, white, gamma, desc, outDir)
	if err != nil {
		return nil, err
	}

	// Store path in session.
	a.mu.Lock()
	if a.activeSession != nil {
		a.activeSession.ProfilePath = result.Path
	}
	a.mu.Unlock()
	_ = a.SaveSession()

	return result, nil
}

func (a *App) InstallICCProfile() (*calibration.InstallResult, error) {
	a.mu.RLock()
	sess := a.activeSession
	a.mu.RUnlock()
	if sess == nil || sess.ProfilePath == "" {
		return nil, fmt.Errorf("no profile generated — run GenerateICCProfile first")
	}
	return calibration.InstallICCProfile(sess.ProfilePath)
}

func (a *App) ExportSessionCSV(sessionID, path string) error {
	sess, err := a.store.Load(sessionID)
	if err != nil {
		return err
	}
	return store.ExportCSV(sess, path)
}

// Settings

func (a *App) GetSettings() store.Settings {
	return store.LoadSettings()
}

func (a *App) SaveSettings(s store.Settings) error {
	return store.SaveSettings(s)
}
