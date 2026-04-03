//go:build linux

package display

import (
	"os"
	"path/filepath"
	"strings"
)

// ReadEDID reads the EDID for a given output name (e.g. "HDMI-1", "DP-1").
func ReadEDID(outputName string) (*EDID, error) {
	// DRM connector names: /sys/class/drm/card*-<name>/edid
	matches, _ := filepath.Glob("/sys/class/drm/card*-*/edid")
	for _, path := range matches {
		dir := filepath.Base(filepath.Dir(path))
		// card0-HDMI-A-1 → HDMI-A-1, card0-DP-1 → DP-1
		parts := strings.SplitN(dir, "-", 2)
		if len(parts) < 2 {
			continue
		}
		connector := parts[1]
		// xrandr uses HDMI-1 but DRM uses HDMI-A-1; normalize
		normalized := strings.ReplaceAll(connector, "-A-", "-")
		normalized = strings.ReplaceAll(normalized, "-B-", "-")
		if normalized == outputName || connector == outputName {
			data, err := os.ReadFile(path)
			if err != nil || len(data) < 128 {
				continue
			}
			return ParseEDID(data), nil
		}
	}
	return nil, os.ErrNotExist
}
