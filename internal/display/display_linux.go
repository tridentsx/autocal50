//go:build linux

package display

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// re matches lines like: HDMI-1 connected primary 3840x2160+0+0 ...
// or: DP-1 connected 1920x1080+3840+0 ...
var outputRe = regexp.MustCompile(
	`^(\S+)\s+(connected|disconnected)\s*(primary)?\s*(?:(\d+)x(\d+)\+(\d+)\+(\d+))?`,
)

// modeRe matches the active mode line like: 3840x2160     60.00*+
var modeRe = regexp.MustCompile(`^\s+(\d+)x(\d+)\s+([\d.]+)\*`)

func ListOutputs() ([]Output, error) {
	out, err := exec.Command("xrandr", "--current").Output()
	if err != nil {
		return nil, fmt.Errorf("xrandr: %w", err)
	}

	var outputs []Output
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var current *Output

	for scanner.Scan() {
		line := scanner.Text()

		if m := outputRe.FindStringSubmatch(line); m != nil {
			o := Output{
				Name:      m[1],
				Connected: m[2] == "connected",
				Primary:   m[3] == "primary",
			}
			if m[4] != "" {
				o.Width, _ = strconv.Atoi(m[4])
				o.Height, _ = strconv.Atoi(m[5])
				o.X, _ = strconv.Atoi(m[6])
				o.Y, _ = strconv.Atoi(m[7])
			}
			outputs = append(outputs, o)
			current = &outputs[len(outputs)-1]
			continue
		}

		// Grab refresh rate from the active mode line
		if current != nil && current.RefreshHz == 0 {
			if m := modeRe.FindStringSubmatch(line); m != nil {
				current.RefreshHz, _ = strconv.ParseFloat(m[3], 64)
			}
		}
	}

	return outputs, nil
}
