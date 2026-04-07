// Package store persists calibration sessions, measurements, and profiles
// as JSON files in a user data directory.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"autocal50/internal/calibration"
	"autocal50/internal/meter"
)

// Measurement is a single pattern→reading pair with evaluation.
type Measurement struct {
	PatternID string              `json:"patternId"`
	Target    calibration.ColorTarget `json:"target"`
	Reading   meter.Reading       `json:"reading"`
	DeltaE    float64             `json:"deltaE"`
	Passed    bool                `json:"passed"`
	Timestamp time.Time           `json:"timestamp"`
}

// PreCalData holds pre-calibration diagnostic results.
type PreCalData struct {
	BlackLevel    float64 `json:"blackLevel"`
	PeakLuminance float64 `json:"peakLuminance"`
	ContrastRatio float64 `json:"contrastRatio"`
	ClipBlack     bool    `json:"clipBlack"`
	ClipWhite     bool    `json:"clipWhite"`
}

// ControlSnapshot captures projector settings at a point in time.
type ControlSnapshot struct {
	Label    string         `json:"label"` // e.g. "before", "after whitebal"
	Controls map[string]any `json:"controls"`
	Timestamp time.Time     `json:"timestamp"`
}

// Session is a complete calibration session.
type Session struct {
	ID          string            `json:"id"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
	Standard    string            `json:"standard"`    // rec709, dcip3, etc.
	Projector   string            `json:"projector"`   // driver name or model
	Meter       string            `json:"meter"`       // driver name
	Connector   string            `json:"connector"`   // HDMI output used
	SignalMode  string            `json:"signalMode"`  // e.g. "3840x2160@24 RGB 10bit"
	PreCal      *PreCalData       `json:"preCal,omitempty"`
	Snapshots   []ControlSnapshot `json:"snapshots,omitempty"`
	Steps       map[string][]Measurement `json:"steps"` // keyed by step kind
	ProfilePath string            `json:"profilePath,omitempty"`
	Notes       string            `json:"notes,omitempty"`
}

// Store manages session persistence.
type Store struct {
	dir string
}

// New creates a store rooted at the given directory.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// DefaultDir returns the platform-appropriate data directory.
func DefaultDir() string {
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "autocal50", "sessions")
	}
	return filepath.Join(".", "autocal50-data", "sessions")
}

// Save writes a session to disk.
func (s *Store) Save(sess *Session) error {
	sess.Updated = time.Now()
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, sess.ID+".json")
	return os.WriteFile(path, data, 0644)
}

// Load reads a session by ID.
func (s *Store) Load(id string) (*Session, error) {
	path := filepath.Join(s.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// Delete removes a session.
func (s *Store) Delete(id string) error {
	return os.Remove(filepath.Join(s.dir, id+".json"))
}

// List returns all sessions, newest first.
func (s *Store) List() ([]Session, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var sessions []Session
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		var sess Session
		if json.Unmarshal(data, &sess) == nil {
			sessions = append(sessions, sess)
		}
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Created.After(sessions[j].Created)
	})
	return sessions, nil
}

// NewSession creates a new session with a unique ID.
func NewSession(standard, projector, meterName string) *Session {
	return &Session{
		ID:       fmt.Sprintf("%d", time.Now().UnixMilli()),
		Created:  time.Now(),
		Updated:  time.Now(),
		Standard: standard,
		Projector: projector,
		Meter:    meterName,
		Steps:    make(map[string][]Measurement),
	}
}

// AddMeasurement appends a measurement to a step.
func (sess *Session) AddMeasurement(step string, m Measurement) {
	sess.Steps[step] = append(sess.Steps[step], m)
}

// AddSnapshot records a projector settings snapshot.
func (sess *Session) AddSnapshot(label string, controls map[string]any) {
	sess.Snapshots = append(sess.Snapshots, ControlSnapshot{
		Label:    label,
		Controls: controls,
		Timestamp: time.Now(),
	})
}
