package calibration

// ColorTarget is a calibration target point.
type ColorTarget struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Luminance float64 `json:"luminance"` // cd/m², 0 = don't care
	Label     string  `json:"label"`
}

// Adjustment tells the user (or auto-loop) what to change.
type Adjustment struct {
	ControlID string `json:"controlId"`
	Direction string `json:"direction"` // "increase" or "decrease"
	Magnitude string `json:"magnitude"` // "slightly", "moderately", "significantly"
	Reason    string `json:"reason"`
}

// Advice is the result of evaluating a measurement against a target.
type Advice struct {
	DeltaE      float64      `json:"deltaE"`
	DeltaX      float64      `json:"deltaX"`
	DeltaY      float64      `json:"deltaY"`
	DeltaLum    float64      `json:"deltaLum"` // signed % error
	Passed      bool         `json:"passed"`
	Adjustments []Adjustment `json:"adjustments"`
}

// Standard defines a color space calibration target.
type Standard struct {
	Label   string     `json:"label"`
	R, G, B [2]float64 `json:"-"` // xy primaries
	W       [2]float64 `json:"-"` // white point xy
	Gamma   float64    `json:"gamma"`
	HDR     bool       `json:"hdr"`
}

// GammaPoint is one point on the target EOTF curve.
type GammaPoint struct {
	Input    float64 `json:"input"`
	Target   float64 `json:"target"`
	Measured float64 `json:"measured"`
}

// StepKind identifies a calibration step type.
type StepKind string

const (
	StepPreCal   StepKind = "precal"
	StepWhiteBal StepKind = "whitebalance"
	StepGamma    StepKind = "gamma"
	StepCMS      StepKind = "cms"
	StepVerify   StepKind = "verify"
)

// Step describes one phase of the calibration routine.
type Step struct {
	Kind             StepKind      `json:"kind"`
	Label            string        `json:"label"`
	Description      string        `json:"description"`
	Targets          []ColorTarget `json:"targets"`
	Tolerance        float64       `json:"tolerance"`
	Optional         bool          `json:"optional"`
	RequiresControls []string      `json:"requiresControls,omitempty"`
}

// PreCalResult holds the results of pre-calibration checks.
type PreCalResult struct {
	BlackLevel  float64 `json:"blackLevel"`
	PeakLum     float64 `json:"peakLum"`
	ContrastRat float64 `json:"contrastRatio"`
	ClipBlack   bool    `json:"clipBlack"`
	ClipWhite   bool    `json:"clipWhite"`
}

// Known standards.
var Standards = map[string]Standard{
	"rec709":  {Label: "Rec. 709", R: [2]float64{0.64, 0.33}, G: [2]float64{0.30, 0.60}, B: [2]float64{0.15, 0.06}, W: [2]float64{0.3127, 0.329}, Gamma: 2.2},
	"dcip3":   {Label: "DCI-P3", R: [2]float64{0.68, 0.32}, G: [2]float64{0.265, 0.69}, B: [2]float64{0.15, 0.06}, W: [2]float64{0.3127, 0.329}, Gamma: 2.6},
	"p3d65":   {Label: "Display P3", R: [2]float64{0.68, 0.32}, G: [2]float64{0.265, 0.69}, B: [2]float64{0.15, 0.06}, W: [2]float64{0.3127, 0.329}, Gamma: 2.2},
	"bt2020":  {Label: "BT.2020", R: [2]float64{0.708, 0.292}, G: [2]float64{0.17, 0.797}, B: [2]float64{0.131, 0.046}, W: [2]float64{0.3127, 0.329}, Gamma: 2.2},
	"hdr10":   {Label: "HDR10 (PQ)", R: [2]float64{0.708, 0.292}, G: [2]float64{0.17, 0.797}, B: [2]float64{0.131, 0.046}, W: [2]float64{0.3127, 0.329}, Gamma: 0, HDR: true},
}
