package calibration

import (
	"autocal50/internal/icc"
	"autocal50/internal/meter"
	"fmt"
	"os"
	"path/filepath"
)

// ProfileResult holds the generated ICC profile and metadata.
type ProfileResult struct {
	Path     string  `json:"path"`
	AvgDE    float64 `json:"avgDeltaE"`
	MaxDE    float64 `json:"maxDeltaE"`
	Primaries [6]float64 `json:"primaries"` // Rx,Ry,Gx,Gy,Bx,By measured
	WhiteXY  [2]float64 `json:"whiteXY"`
}

// GenerateProfile creates an ICC display profile from calibration measurements.
// colorReadings: R, G, B meter readings (from full-field primary patterns).
// whiteReading: white (100% gray) meter reading.
// gammaPoints: measured grayscale ramp from RunGammaSweep.
// desc: profile description string.
// outDir: directory to write the .icc file.
func GenerateProfile(
	colorReadings [3]meter.Reading, // R, G, B
	whiteReading meter.Reading,
	gammaPoints []GammaPoint,
	desc string,
	outDir string,
) (*ProfileResult, error) {

	// Convert meter XYZ readings to icc.XYZ.
	// meter.Reading has X,Y,Z as CIE XYZ (Y = luminance in cd/m²).
	// ICC needs Y-normalized (white Y = 1.0).
	wY := whiteReading.Y
	if wY <= 0 {
		wY = whiteReading.Luminance
	}
	if wY <= 0 {
		return nil, fmt.Errorf("white luminance is zero")
	}

	normalize := func(r meter.Reading) icc.XYZ {
		return icc.XYZ{X: r.X / wY, Y: r.Y / wY, Z: r.Z / wY}
	}

	rXYZ := normalize(colorReadings[0])
	gXYZ := normalize(colorReadings[1])
	bXYZ := normalize(colorReadings[2])
	white := normalize(whiteReading)

	// Build TRC from gamma measurements.
	trc := buildTRC(gammaPoints)

	// Create the ICC profile. Use same TRC for all channels
	// (per-channel TRC would need per-channel gamma sweeps).
	prof := icc.NewDisplayProfileXYZ(desc, rXYZ, gXYZ, bXYZ, white, trc, trc, trc)

	// Write to disk.
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, err
	}
	data, err := prof.Bytes()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(outDir, sanitizeFilename(desc)+".icc")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, err
	}

	// Extract measured primaries as xy for the result.
	rR := colorReadings[0]
	rG := colorReadings[1]
	rB := colorReadings[2]

	return &ProfileResult{
		Path: path,
		Primaries: [6]float64{
			xFromXYZ(rR), yFromXYZ(rR),
			xFromXYZ(rG), yFromXYZ(rG),
			xFromXYZ(rB), yFromXYZ(rB),
		},
		WhiteXY: [2]float64{xFromXYZ(whiteReading), yFromXYZ(whiteReading)},
	}, nil
}

func buildTRC(pts []GammaPoint) icc.Curve {
	if len(pts) == 0 {
		return icc.GammaCurve(2.2) // fallback
	}
	// Build a 256-entry LUT from the measured points.
	table := make([]uint16, 256)
	for i := range table {
		input := float64(i) / 255.0
		// Interpolate measured points.
		val := interpolateGamma(pts, input)
		if val < 0 {
			val = 0
		}
		if val > 1 {
			val = 1
		}
		table[i] = uint16(val * 65535)
	}
	return icc.Curve{Table: table}
}

func interpolateGamma(pts []GammaPoint, x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	for i := 1; i < len(pts); i++ {
		if pts[i].Input >= x {
			t := (x - pts[i-1].Input) / (pts[i].Input - pts[i-1].Input)
			return pts[i-1].Measured + t*(pts[i].Measured-pts[i-1].Measured)
		}
	}
	return pts[len(pts)-1].Measured
}

func xFromXYZ(r meter.Reading) float64 {
	sum := r.X + r.Y + r.Z
	if sum <= 0 {
		return 0
	}
	return r.X / sum
}

func yFromXYZ(r meter.Reading) float64 {
	sum := r.X + r.Y + r.Z
	if sum <= 0 {
		return 0
	}
	return r.Y / sum
}

func sanitizeFilename(s string) string {
	out := make([]byte, 0, len(s))
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == ' ' {
			out = append(out, byte(c))
		}
	}
	if len(out) == 0 {
		return "profile"
	}
	return string(out)
}
