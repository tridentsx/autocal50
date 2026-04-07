//go:build linux

package display

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ListOutputs enumerates display connectors via sysfs.
// Works on X11, Wayland, and bare TTY — no dependency on xrandr.
func ListOutputs() ([]Output, error) {
	dirs, err := filepath.Glob("/sys/class/drm/card*-*")
	if err != nil {
		return nil, err
	}

	var outputs []Output
	for _, dir := range dirs {
		base := filepath.Base(dir)
		// card0-HDMI-A-1 → HDMI-A-1
		parts := strings.SplitN(base, "-", 2)
		if len(parts) < 2 {
			continue
		}
		name := parts[1]

		status := readFile(filepath.Join(dir, "status"))
		connected := strings.TrimSpace(status) == "connected"

		o := Output{
			Name:      name,
			Connected: connected,
		}

		if connected {
			o.Modes = parseModes(filepath.Join(dir, "modes"))
			for _, m := range o.Modes {
				if m.Preferred || m.Current {
					o.Width = m.Width
					o.Height = m.Height
					o.RefreshHz = m.RefreshHz
					break
				}
			}
			// Fallback to first mode if none marked preferred.
			if o.Width == 0 && len(o.Modes) > 0 {
				o.Width = o.Modes[0].Width
				o.Height = o.Modes[0].Height
				o.RefreshHz = o.Modes[0].RefreshHz
			}
		}

		outputs = append(outputs, o)
	}

	return outputs, nil
}

// parseModes reads /sys/class/drm/card*-*/modes which lists one mode per line
// in the format "3840x2160" (sysfs doesn't include refresh in the modes file,
// but we can get it from the EDID or DRM ioctls later).
// For now we parse what sysfs gives us.
func parseModes(path string) []Mode {
	data := readFile(path)
	if data == "" {
		return nil
	}
	seen := make(map[string]bool)
	var modes []Mode
	for i, line := range strings.Split(strings.TrimSpace(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true
		parts := strings.SplitN(line, "x", 2)
		if len(parts) != 2 {
			continue
		}
		w, _ := strconv.Atoi(parts[0])
		h, _ := strconv.Atoi(parts[1])
		if w == 0 || h == 0 {
			continue
		}
		modes = append(modes, Mode{
			Width:     w,
			Height:    h,
			Preferred: i == 0, // first mode in sysfs is typically preferred
		})
	}
	return modes
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
