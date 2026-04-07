package calibration

import (
	"autocal50/internal/meter"
	"context"
	"fmt"
	"math"
	"time"
)

// Engine runs automated or assisted calibration loops.
type Engine struct {
	ShowPattern func(patternID string) error
	Measure     func(ctx context.Context) (meter.Reading, error)
	GetControl  func(ctx context.Context, id string) (float64, error)
	SetControl  func(ctx context.Context, id string, value float64) error
	OnProgress  func(event ProgressEvent)

	MaxIterations int
	SettleTime    time.Duration // wait after adjusting before re-measuring
}

// ProgressEvent is emitted during calibration.
type ProgressEvent struct {
	Step       string  `json:"step"`
	Message    string  `json:"message"`
	PatternID  string  `json:"patternId"`
	DeltaE     float64 `json:"deltaE"`
	Passed     bool    `json:"passed"`
	Iteration  int     `json:"iteration"`
	TotalSteps int     `json:"totalSteps"`
	StepIndex  int     `json:"stepIndex"`
}

func (e *Engine) defaults() {
	if e.MaxIterations == 0 {
		e.MaxIterations = 20
	}
	if e.SettleTime == 0 {
		e.SettleTime = 500 * time.Millisecond
	}
}

func (e *Engine) emit(ev ProgressEvent) {
	if e.OnProgress != nil {
		e.OnProgress(ev)
	}
}

// RunPreCal measures baseline black level, peak luminance, and contrast ratio.
func (e *Engine) RunPreCal(ctx context.Context) (*PreCalResult, error) {
	e.emit(ProgressEvent{Step: "precal", Message: "Measuring black level..."})
	e.ShowPattern("gray-win-0")
	time.Sleep(e.SettleTime)
	black, err := e.Measure(ctx)
	if err != nil {
		return nil, err
	}

	e.emit(ProgressEvent{Step: "precal", Message: "Measuring peak white..."})
	e.ShowPattern("gray-win-100")
	time.Sleep(e.SettleTime)
	white, err := e.Measure(ctx)
	if err != nil {
		return nil, err
	}

	cr := 0.0
	if black.Luminance > 0 {
		cr = white.Luminance / black.Luminance
	}

	// Clipping detection: check if 1% and 99% are distinguishable.
	e.ShowPattern("clip-black-1")
	time.Sleep(e.SettleTime)
	b1, _ := e.Measure(ctx)
	clipBlack := math.Abs(b1.Luminance-black.Luminance) < 0.001

	e.ShowPattern("clip-white-99")
	time.Sleep(e.SettleTime)
	w99, _ := e.Measure(ctx)
	clipWhite := math.Abs(w99.Luminance-white.Luminance) < 0.01

	return &PreCalResult{
		BlackLevel:  black.Luminance,
		PeakLum:     white.Luminance,
		ContrastRat: math.Round(cr),
		ClipBlack:   clipBlack,
		ClipWhite:   clipWhite,
	}, nil
}

// RunWhiteBalance performs 2-point white balance using RGB gain (high) and bias (low).
func (e *Engine) RunWhiteBalance(ctx context.Context, std Standard, tolerance float64) error {
	e.defaults()
	targets := WhiteBalanceTargets(std)

	// Pass 1: Gain (80% IRE)
	if err := e.adjustWhitePoint(ctx, "gain", targets[0], tolerance,
		"gray-win-80", "rgb_gain_r", "rgb_gain_g", "rgb_gain_b"); err != nil {
		return err
	}

	// Pass 2: Bias (20% IRE)
	if err := e.adjustWhitePoint(ctx, "bias", targets[1], tolerance,
		"gray-win-20", "rgb_bias_r", "rgb_bias_g", "rgb_bias_b"); err != nil {
		return err
	}

	return nil
}

func (e *Engine) adjustWhitePoint(ctx context.Context, label string, target ColorTarget, tolerance float64,
	patternID, rCtrl, gCtrl, bCtrl string) error {

	e.ShowPattern(patternID)
	time.Sleep(e.SettleTime)

	for iter := 0; iter < e.MaxIterations; iter++ {
		reading, err := e.Measure(ctx)
		if err != nil {
			return err
		}
		advice := Evaluate(target, reading, tolerance)
		e.emit(ProgressEvent{
			Step: "whitebalance", Message: fmt.Sprintf("%s: ΔE=%.2f", label, advice.DeltaE),
			PatternID: patternID, DeltaE: advice.DeltaE, Passed: advice.Passed, Iteration: iter,
		})
		if advice.Passed {
			return nil
		}

		// Calculate corrections from xy error.
		dx := reading.X - target.X
		dy := reading.Y - target.Y

		// Heuristic: map xy error to RGB adjustments.
		// Positive dx → too red → decrease R or increase B.
		// Positive dy → too green → decrease G.
		stepSize := 1.0
		if math.Abs(dx) < 0.005 && math.Abs(dy) < 0.005 {
			stepSize = 0.5 // fine adjustment
		}

		if math.Abs(dx) > 0.002 {
			e.nudgeControl(ctx, rCtrl, -dx*stepSize*50)
			e.nudgeControl(ctx, bCtrl, dx*stepSize*30)
		}
		if math.Abs(dy) > 0.002 {
			e.nudgeControl(ctx, gCtrl, -dy*stepSize*50)
		}

		time.Sleep(e.SettleTime)
	}
	return fmt.Errorf("white balance %s did not converge after %d iterations", label, e.MaxIterations)
}

func (e *Engine) nudgeControl(ctx context.Context, id string, delta float64) {
	if e.GetControl == nil || e.SetControl == nil {
		return
	}
	cur, err := e.GetControl(ctx, id)
	if err != nil {
		return
	}
	_ = e.SetControl(ctx, id, cur+delta)
}

// RunGammaSweep measures the grayscale ramp and returns measured gamma points.
func (e *Engine) RunGammaSweep(ctx context.Context, std Standard, steps int) ([]GammaPoint, error) {
	e.defaults()
	targets := GammaTargets(std, steps)

	// Measure 100% first for normalization.
	e.ShowPattern("gray-win-100")
	time.Sleep(e.SettleTime)
	peak, err := e.Measure(ctx)
	if err != nil {
		return nil, err
	}

	for i := range targets {
		pct := int(targets[i].Input * 100)
		patternID := fmt.Sprintf("gray-win-%d", pct)
		e.ShowPattern(patternID)
		time.Sleep(e.SettleTime)

		reading, err := e.Measure(ctx)
		if err != nil {
			return nil, err
		}

		if peak.Luminance > 0 {
			targets[i].Measured = reading.Luminance / peak.Luminance
		}

		e.emit(ProgressEvent{
			Step: "gamma", Message: fmt.Sprintf("%d%% IRE", pct),
			PatternID: patternID, StepIndex: i, TotalSteps: len(targets),
		})
	}
	return targets, nil
}

// RunColorSweep measures primaries and secondaries.
func (e *Engine) RunColorSweep(ctx context.Context, std Standard) ([]Advice, error) {
	e.defaults()
	targets := PrimaryTargets(std)
	patterns := []string{"win-Red", "win-Green", "win-Blue", "win-Cyan", "win-Magenta", "win-Yellow"}

	var results []Advice
	for i, t := range targets {
		e.ShowPattern(patterns[i])
		time.Sleep(e.SettleTime)

		reading, err := e.Measure(ctx)
		if err != nil {
			return nil, err
		}
		adv := Evaluate(t, reading, 2.0)
		results = append(results, adv)

		e.emit(ProgressEvent{
			Step: "cms", Message: t.Label, PatternID: patterns[i],
			DeltaE: adv.DeltaE, Passed: adv.Passed, StepIndex: i, TotalSteps: len(targets),
		})
	}
	return results, nil
}

// RunVerification runs a full grayscale + color sweep for final scoring.
func (e *Engine) RunVerification(ctx context.Context, std Standard) (avgDE float64, maxDE float64, err error) {
	e.defaults()

	// Grayscale 0-100 in 10% steps.
	grayTargets := GrayscaleTargets(std, 10)
	var allDE []float64

	e.ShowPattern("gray-win-100")
	time.Sleep(e.SettleTime)
	peak, err := e.Measure(ctx)
	if err != nil {
		return 0, 0, err
	}

	for i, t := range grayTargets {
		pct := i * 10
		patternID := fmt.Sprintf("gray-win-%d", pct)
		e.ShowPattern(patternID)
		time.Sleep(e.SettleTime)

		reading, err := e.Measure(ctx)
		if err != nil {
			return 0, 0, err
		}
		// Scale target luminance to actual peak.
		scaled := t
		if peak.Luminance > 0 {
			scaled.Luminance = t.Luminance / 100 * peak.Luminance
		}
		adv := Evaluate(scaled, reading, 2.0)
		allDE = append(allDE, adv.DeltaE)

		e.emit(ProgressEvent{
			Step: "verify", Message: fmt.Sprintf("Gray %d%%", pct),
			DeltaE: adv.DeltaE, Passed: adv.Passed, StepIndex: i, TotalSteps: 17,
		})
	}

	// Primaries + secondaries.
	colors, err := e.RunColorSweep(ctx, std)
	if err != nil {
		return 0, 0, err
	}
	for _, c := range colors {
		allDE = append(allDE, c.DeltaE)
	}

	// Compute stats.
	for _, de := range allDE {
		avgDE += de
		if de > maxDE {
			maxDE = de
		}
	}
	if len(allDE) > 0 {
		avgDE /= float64(len(allDE))
	}
	return math.Round(avgDE*100) / 100, math.Round(maxDE*100) / 100, nil
}
