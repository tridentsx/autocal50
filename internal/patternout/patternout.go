// Package patternout renders test patterns directly to display outputs,
// bypassing the desktop compositor for accurate calibration measurements.
package patternout

import "autocal50/internal/pattern"

// Mode describes a display mode.
type Mode struct {
	Width     int
	Height    int
	RefreshHz float64
}

// Output renders patterns to a physical display connector.
type Output interface {
	// Open acquires the display connector for exclusive output.
	Open(connector string) error

	// Modes returns available display modes for the opened connector.
	Modes() []Mode

	// SetMode configures resolution and refresh rate.
	SetMode(m Mode) error

	// Render draws a pattern to the display.
	Render(p pattern.Pattern) error

	// Close releases the display connector and restores previous state.
	Close() error
}
