package calibration

import (
	"fmt"
	"math"
)

// GrayscaleTargets generates white-point xy + luminance targets for each IRE step.
func GrayscaleTargets(std Standard, steps int) []ColorTarget {
	targets := make([]ColorTarget, steps+1)
	for i := 0; i <= steps; i++ {
		pct := float64(i) / float64(steps)
		var lum float64
		if std.HDR {
			lum = pqEOTF(pct) * 10000 // nits
		} else {
			lum = math.Pow(pct, std.Gamma) * 100 // normalized to 100 cd/m² reference
		}
		targets[i] = ColorTarget{
			X:         std.W[0],
			Y:         std.W[1],
			Luminance: lum,
			Label:     fmt.Sprintf("%d%% IRE", int(pct*100)),
		}
	}
	return targets
}

// PrimaryTargets returns the 6 primary+secondary xy targets.
func PrimaryTargets(std Standard) []ColorTarget {
	return []ColorTarget{
		{X: std.R[0], Y: std.R[1], Label: "Red"},
		{X: std.G[0], Y: std.G[1], Label: "Green"},
		{X: std.B[0], Y: std.B[1], Label: "Blue"},
		{X: (std.G[0] + std.B[0]) / 2, Y: (std.G[1] + std.B[1]) / 2, Label: "Cyan"},
		{X: (std.R[0] + std.B[0]) / 2, Y: (std.R[1] + std.B[1]) / 2, Label: "Magenta"},
		{X: (std.R[0] + std.G[0]) / 2, Y: (std.R[1] + std.G[1]) / 2, Label: "Yellow"},
	}
}

// WhiteBalanceTargets returns the two IRE levels used for 2-point white balance.
func WhiteBalanceTargets(std Standard) []ColorTarget {
	return []ColorTarget{
		{X: std.W[0], Y: std.W[1], Label: "80% IRE (Gain)", Luminance: math.Pow(0.8, std.Gamma) * 100},
		{X: std.W[0], Y: std.W[1], Label: "20% IRE (Bias)", Luminance: math.Pow(0.2, std.Gamma) * 100},
	}
}

// GammaTargets generates input→output pairs for the target EOTF.
func GammaTargets(std Standard, steps int) []GammaPoint {
	pts := make([]GammaPoint, steps+1)
	for i := 0; i <= steps; i++ {
		input := float64(i) / float64(steps)
		var target float64
		if std.HDR {
			target = pqEOTF(input)
		} else {
			target = math.Pow(input, std.Gamma)
		}
		pts[i] = GammaPoint{Input: input, Target: target}
	}
	return pts
}

// ST.2084 PQ EOTF: normalized input (0-1) → normalized luminance (0-1, ×10000 = nits)
func pqEOTF(v float64) float64 {
	if v <= 0 {
		return 0
	}
	m1 := 0.1593017578125
	m2 := 78.84375
	c1 := 0.8359375
	c2 := 18.8515625
	c3 := 18.6875
	vp := math.Pow(v, 1.0/m2)
	num := math.Max(vp-c1, 0)
	return math.Pow(num/(c2-c3*vp), 1.0/m1)
}
