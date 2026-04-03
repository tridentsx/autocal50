//go:build darwin

package display

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ReadEDID reads the EDID for a given output name on macOS.
func ReadEDID(outputName string) (*EDID, error) {
	out, err := exec.Command("ioreg", "-l", "-d0", "-w", "0", "-r", "-c", "IODisplayConnect").Output()
	if err != nil {
		return nil, fmt.Errorf("ioreg: %w", err)
	}

	// Find IODisplayEDID entries — they're hex-encoded byte arrays
	for _, chunk := range bytes.Split(out, []byte("IODisplayEDID")) {
		if len(chunk) < 10 {
			continue
		}
		// Extract hex data between < and >
		start := bytes.IndexByte(chunk, '<')
		end := bytes.IndexByte(chunk, '>')
		if start < 0 || end <= start {
			continue
		}
		hex := chunk[start+1 : end]
		data := make([]byte, len(hex)/2)
		for i := 0; i < len(data); i++ {
			fmt.Sscanf(string(hex[i*2:i*2+2]), "%02x", &data[i])
		}
		if len(data) >= 128 {
			edid := ParseEDID(data)
			if edid != nil {
				return edid, nil
			}
		}
	}
	return nil, fmt.Errorf("EDID not found for %s", outputName)
}
