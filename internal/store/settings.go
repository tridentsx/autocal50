package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings holds user preferences persisted between sessions.
type Settings struct {
	LastProjectorDriver string         `json:"lastProjectorDriver,omitempty"`
	LastProjectorCfg    map[string]any `json:"lastProjectorCfg,omitempty"`
	LastMeterDriver     string         `json:"lastMeterDriver,omitempty"`
	LastMeterCfg        map[string]any `json:"lastMeterCfg,omitempty"`
	LastTransportDriver string         `json:"lastTransportDriver,omitempty"`
	LastTransportCfg    map[string]any `json:"lastTransportCfg,omitempty"`
	LastStandard        string         `json:"lastStandard,omitempty"`
	LastConnector       string         `json:"lastConnector,omitempty"`
}

func settingsPath() string {
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "autocal50", "settings.json")
	}
	return filepath.Join(".", "autocal50-data", "settings.json")
}

// LoadSettings reads saved preferences.
func LoadSettings() Settings {
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return Settings{}
	}
	var s Settings
	json.Unmarshal(data, &s)
	return s
}

// SaveSettings writes preferences to disk.
func SaveSettings(s Settings) error {
	path := settingsPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
