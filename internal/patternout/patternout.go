// Package patternout renders test patterns directly to display outputs,
// bypassing the desktop compositor for accurate calibration measurements.
package patternout

import "autocal50/internal/pattern"

// Mode describes a display mode.
type Mode struct {
	Width     int
	Height    int
	RefreshHz float64
	BitDepth  int  // 8, 10, or 12
	HDR       bool // ST.2084 PQ capable
}

// HDRMetadata describes HDR10 static metadata (CTA-861.3 / SMPTE ST 2086).
type HDRMetadata struct {
	// Mastering display primaries in 0.00002 units (CIE 1931 xy × 50000).
	Rx, Ry uint16
	Gx, Gy uint16
	Bx, By uint16
	Wx, Wy uint16
	// Luminance in units of 0.0001 cd/m².
	MaxLuminance uint32
	MinLuminance uint32
	// Content light level.
	MaxCLL  uint16 // nits
	MaxFALL uint16 // nits
}

// BT2020HDR10 returns HDR10 metadata for BT.2020 primaries with the given peak luminance.
func BT2020HDR10(peakNits int) HDRMetadata {
	return HDRMetadata{
		Rx: 35400, Ry: 14600,
		Gx: 8500, Gy: 39850,
		Bx: 6550, By: 2300,
		Wx: 15635, Wy: 16450,
		MaxLuminance: uint32(peakNits) * 10000,
		MinLuminance: 500,
		MaxCLL:       uint16(peakNits),
		MaxFALL:      uint16(peakNits / 4),
	}
}

// Output renders patterns to a physical display connector.
type Output interface {
	// Open acquires the display connector for exclusive output.
	Open(connector string) error

	// Modes returns available display modes for the opened connector.
	Modes() []Mode

	// SetMode configures resolution, refresh rate, and bit depth.
	SetMode(m Mode) error

	// Render draws a pattern to the display.
	Render(p pattern.Pattern) error

	// SetHDRMetadata sends HDR10 static metadata to the display.
	// Pass nil to clear HDR metadata.
	SetHDRMetadata(meta *HDRMetadata) error

	// Close releases the display connector and restores previous state.
	Close() error
}
