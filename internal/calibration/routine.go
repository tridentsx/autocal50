package calibration

import "autocal50/internal/projector"

// BuildRoutine generates the calibration steps for a given standard,
// filtering out steps that require controls the projector doesn't have.
func BuildRoutine(stdKey string, caps projector.Capabilities) []Step {
	std, ok := Standards[stdKey]
	if !ok {
		std = Standards["rec709"]
	}

	hasControl := make(map[string]bool)
	for _, c := range caps.Controls {
		hasControl[c.ID] = true
	}

	steps := []Step{
		{
			Kind:        StepPreCal,
			Label:       "Pre-Cal Checks",
			Description: "Measure baseline black level, peak luminance, clipping, and contrast ratio before making any adjustments.",
			Tolerance:   0, // informational
		},
		{
			Kind:             StepWhiteBal,
			Label:            "White Balance",
			Description:      "Adjust RGB gain and bias to achieve the target white point (D65) at 80% and 20% IRE.",
			Targets:          WhiteBalanceTargets(std),
			Tolerance:        1.5,
			RequiresControls: []string{"rgb_gain_r", "rgb_gain_g", "rgb_gain_b"},
			Optional:         true,
		},
		{
			Kind:        StepGamma,
			Label:       "Gamma / EOTF",
			Description: "Verify the electro-optical transfer function matches the target curve. Select the best gamma preset if available.",
			Tolerance:   0.03, // gamma error tolerance
		},
		{
			Kind:             StepCMS,
			Label:            "Color Management",
			Description:      "Adjust primary and secondary colors to match the target gamut using the projector's CMS controls.",
			Targets:          PrimaryTargets(std),
			Tolerance:        2.0,
			RequiresControls: []string{"cms_red_hue"},
			Optional:         true,
		},
		{
			Kind:        StepVerify,
			Label:       "Verification",
			Description: "Run a full measurement sweep to confirm calibration quality and generate a final report.",
			Tolerance:   2.0,
		},
	}

	// Filter out optional steps that require missing controls
	var result []Step
	for _, s := range steps {
		if s.Optional && len(s.RequiresControls) > 0 {
			skip := false
			for _, id := range s.RequiresControls {
				if !hasControl[id] {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		result = append(result, s)
	}
	return result
}
