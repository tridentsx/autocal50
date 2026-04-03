//go:build darwin

package display

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type spDisplay struct {
	Name       string `json:"_name"`
	Resolution string `json:"_spdisplays_resolution"`
	PixelW     string `json:"_spdisplays_pixels"`
	Main       string `json:"spdisplays_main,omitempty"`
}

type spEntry struct {
	Displays []spDisplay `json:"spdisplays_ndrvs"`
}

func ListOutputs() ([]Output, error) {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType", "-json").Output()
	if err != nil {
		return nil, fmt.Errorf("system_profiler: %w", err)
	}

	var entries []spEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, err
	}

	var outputs []Output
	for _, e := range entries {
		for _, d := range e.Displays {
			o := Output{
				Name:      d.Name,
				Connected: true,
				Primary:   d.Main == "spdisplays_yes",
			}
			// Parse resolution like "3840 x 2160 @ 60 Hz"
			fmt.Sscanf(d.Resolution, "%d x %d @ %f", &o.Width, &o.Height, &o.RefreshHz)
			outputs = append(outputs, o)
		}
	}

	return outputs, nil
}
