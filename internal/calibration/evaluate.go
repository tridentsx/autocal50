package calibration

import (
	"fmt"
	"math"

	"autocal50/internal/meter"
)

// Evaluate compares a measurement against a target and returns advice.
func Evaluate(target ColorTarget, reading meter.Reading, tolerance float64) Advice {
	// Normalize luminance: reading.Luminance is in cd/m²
	// For ΔE we need Y in 0-100 range
	tY := target.Luminance
	mY := reading.Luminance
	if tY == 0 {
		tY = mY // don't penalize luminance if target doesn't specify
	}

	de := DeltaE2000(target.X, target.Y, tY, reading.X, reading.Y, mY)
	dx := reading.X - target.X
	dy := reading.Y - target.Y
	var dlum float64
	if target.Luminance > 0 {
		dlum = (mY - tY) / tY * 100
	}

	adv := Advice{
		DeltaE:   math.Round(de*100) / 100,
		DeltaX:   math.Round(dx*10000) / 10000,
		DeltaY:   math.Round(dy*10000) / 10000,
		DeltaLum: math.Round(dlum*10) / 10,
		Passed:   de <= tolerance,
	}

	if !adv.Passed {
		adv.Adjustments = suggestAdjustments(dx, dy, dlum)
	}
	return adv
}

func suggestAdjustments(dx, dy, dlum float64) []Adjustment {
	var adj []Adjustment

	// xy error → RGB gain/bias advice
	// Positive dx (too red/yellow) → decrease red gain or increase blue gain
	// Positive dy (too green) → decrease green gain
	// These are simplified heuristics for the common 2-point white balance case

	mag := func(v float64) string {
		a := math.Abs(v)
		if a > 0.01 {
			return "significantly"
		}
		if a > 0.005 {
			return "moderately"
		}
		return "slightly"
	}

	if math.Abs(dx) > 0.002 {
		if dx > 0 {
			adj = append(adj, Adjustment{
				ControlID: "rgb_gain_r",
				Direction: "decrease",
				Magnitude: mag(dx),
				Reason:    "x is above target — too much red",
			})
		} else {
			adj = append(adj, Adjustment{
				ControlID: "rgb_gain_r",
				Direction: "increase",
				Magnitude: mag(dx),
				Reason:    "x is below target — not enough red",
			})
		}
	}

	if math.Abs(dy) > 0.002 {
		if dy > 0 {
			adj = append(adj, Adjustment{
				ControlID: "rgb_gain_g",
				Direction: "decrease",
				Magnitude: mag(dy),
				Reason:    "y is above target — too much green",
			})
		} else {
			adj = append(adj, Adjustment{
				ControlID: "rgb_gain_g",
				Direction: "increase",
				Magnitude: mag(dy),
				Reason:    "y is below target — not enough green",
			})
		}
	}

	if math.Abs(dlum) > 5 {
		dir := "increase"
		if dlum > 0 {
			dir = "decrease"
		}
		adj = append(adj, Adjustment{
			ControlID: "brightness",
			Direction: dir,
			Magnitude: mag(dlum / 100),
			Reason:    fmt.Sprintf("luminance is %.0f%% off target", dlum),
		})
	}

	return adj
}
