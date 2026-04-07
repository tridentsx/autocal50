package store

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
)

// ExportCSV writes all measurements from a session to a CSV file.
func ExportCSV(sess *Session, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{
		"Step", "Pattern", "Target Label", "Target x", "Target y", "Target Lum",
		"Measured x", "Measured y", "Measured Y", "Measured Lum", "Measured CCT",
		"DeltaE", "Passed", "Timestamp",
	})

	// Sort steps for deterministic output.
	steps := make([]string, 0, len(sess.Steps))
	for k := range sess.Steps {
		steps = append(steps, k)
	}
	sort.Strings(steps)

	for _, step := range steps {
		for _, m := range sess.Steps[step] {
			w.Write([]string{
				step,
				m.PatternID,
				m.Target.Label,
				fmt.Sprintf("%.4f", m.Target.X),
				fmt.Sprintf("%.4f", m.Target.Y),
				fmt.Sprintf("%.2f", m.Target.Luminance),
				fmt.Sprintf("%.4f", m.Reading.X),
				fmt.Sprintf("%.4f", m.Reading.Y),
				fmt.Sprintf("%.4f", m.Reading.Z),
				fmt.Sprintf("%.2f", m.Reading.Luminance),
				fmt.Sprintf("%.0f", m.Reading.CCT),
				fmt.Sprintf("%.2f", m.DeltaE),
				fmt.Sprintf("%t", m.Passed),
				m.Timestamp.Format("2006-01-02T15:04:05"),
			})
		}
	}
	return nil
}
