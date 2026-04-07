package transport

import (
	"autocal50/internal/display"
	"autocal50/internal/patternout"
	"context"
	"fmt"
)

// LocalDriver is a transport driver that outputs patterns directly to a
// local display connector via DRM/KMS (Linux) or platform-native APIs.
// It implements the transport.Driver interface so the UI treats it like
// any other signal transport.
type LocalDriver struct {
	output    patternout.Output
	connector string
	current   SignalFormat
}

// NewLocalDriver creates a local output transport driver.
func NewLocalDriver() Driver {
	return &LocalDriver{output: newPlatformOutput()}
}

func (d *LocalDriver) Name() string { return "local-output" }

func (d *LocalDriver) Connect(_ context.Context, cfg map[string]any) error {
	connector, _ := cfg["connector"].(string)
	if connector == "" {
		// Try to find a connected non-primary output.
		outputs, err := display.ListOutputs()
		if err != nil {
			return fmt.Errorf("list outputs: %w", err)
		}
		for _, o := range outputs {
			if o.Connected && !o.Primary {
				connector = o.Name
				break
			}
		}
		if connector == "" {
			return fmt.Errorf("no secondary display connector found")
		}
	}

	if err := d.output.Open(connector); err != nil {
		return err
	}
	d.connector = connector

	// Set initial format from first available mode.
	modes := d.output.Modes()
	if len(modes) > 0 {
		m := modes[0]
		d.current = SignalFormat{
			Resolution: fmt.Sprintf("%dx%d", m.Width, m.Height),
			RefreshHz:  m.RefreshHz,
			Encoding:   "RGB",
			BitDepth:   8,
		}
	}
	return nil
}

func (d *LocalDriver) Disconnect() error {
	return d.output.Close()
}

func (d *LocalDriver) GetFormats(_ context.Context) ([]SignalFormat, error) {
	modes := d.output.Modes()
	var formats []SignalFormat
	for _, m := range modes {
		formats = append(formats, SignalFormat{
			Resolution: fmt.Sprintf("%dx%d", m.Width, m.Height),
			RefreshHz:  m.RefreshHz,
			Encoding:   "RGB",
			BitDepth:   8,
		})
	}
	return formats, nil
}

func (d *LocalDriver) ApplyFormat(_ context.Context, f SignalFormat) error {
	var w, h int
	fmt.Sscanf(f.Resolution, "%dx%d", &w, &h)
	if err := d.output.SetMode(patternout.Mode{
		Width: w, Height: h, RefreshHz: f.RefreshHz,
	}); err != nil {
		return err
	}
	d.current = f
	return nil
}

func (d *LocalDriver) CurrentFormat(_ context.Context) (SignalFormat, error) {
	return d.current, nil
}

// Output returns the underlying patternout.Output for direct pattern rendering.
func (d *LocalDriver) Output() patternout.Output {
	return d.output
}
