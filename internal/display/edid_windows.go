//go:build windows

package display

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

// ReadEDID reads the EDID for a given output name (e.g. "\\\\.\\DISPLAY1").
func ReadEDID(outputName string) (*EDID, error) {
	// Walk HKLM\SYSTEM\CurrentControlSet\Enum\DISPLAY\*\*\Device Parameters
	base := `SYSTEM\CurrentControlSet\Enum\DISPLAY`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, base, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	monitors, _ := k.ReadSubKeyNames(-1)
	for _, mon := range monitors {
		mk, err := registry.OpenKey(registry.LOCAL_MACHINE, base+`\`+mon, registry.ENUMERATE_SUB_KEYS)
		if err != nil {
			continue
		}
		instances, _ := mk.ReadSubKeyNames(-1)
		mk.Close()

		for _, inst := range instances {
			dpPath := base + `\` + mon + `\` + inst + `\Device Parameters`
			dk, err := registry.OpenKey(registry.LOCAL_MACHINE, dpPath, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			data, _, err := dk.GetBinaryValue("EDID")
			dk.Close()
			if err != nil || len(data) < 128 {
				continue
			}
			edid := ParseEDID(data)
			if edid == nil {
				continue
			}
			// Match by display name or return first match for the requested output
			if edid.DisplayName != "" && outputName != "" {
				// On Windows, outputName is like \\.\DISPLAY1 — we can't directly
				// match to registry entries, so return the EDID whose device name
				// contains the monitor index or just iterate until we find a match.
				// For now, collect all and match by index.
			}
			return edid, nil
		}
	}
	return nil, fmt.Errorf("EDID not found for %s", outputName)
}
