# Assisted Calibration — Implementation Plan

## Overview

A wizard-style guided calibration flow with polished animations that runs on the laptop screen while test patterns display on the projector via a separate HDMI output. The user follows step-by-step instructions, adjusts projector controls, and watches live visual feedback converge on targets.

---

## 1. New Dependency

**framer-motion** — React animation library for layout transitions, spring physics, shared element animations, and gesture support. This is what enables the polished feel that differentiates from CalMAN/ChromaPure/HCFR.

```
npm install framer-motion
```

---

## 2. Backend — `internal/calibration/`

### 2.1 Types (`types.go`)

```go
type ColorTarget struct {
    X, Y       float64  // CIE xy chromaticity
    Luminance  float64  // cd/m² (0 = don't care)
    Label      string   // "20% IRE", "Red Primary", etc.
}

type Advice struct {
    DeltaE     float64            // ΔE2000 from target
    DeltaX     float64            // signed error in x
    DeltaY     float64            // signed error in y
    DeltaLum   float64            // signed luminance error %
    Passed     bool               // within tolerance
    Adjustments []Adjustment      // what to change
}

type Adjustment struct {
    ControlID  string   // projector control ID ("rgb_gain_r", etc.)
    Direction  string   // "increase" or "decrease"
    Magnitude  string   // "slightly", "moderately", "significantly"
    Reason     string   // "Red is too high at this stimulus level"
}
```

### 2.2 Target Generation (`targets.go`)

Functions that produce calibration targets from a color standard:

- `GrayscaleTargets(standard, numSteps) → []ColorTarget`
  Generates xy + luminance targets for each IRE level based on the standard's white point and gamma/EOTF.

- `PrimaryTargets(standard) → []ColorTarget`
  Returns the 3 primary + 3 secondary xy targets from the standard's gamut definition.

- `GammaTargets(standard, numSteps) → []GammaPoint`
  Returns input→output mapping for the target EOTF (power gamma or PQ).

### 2.3 Evaluation (`evaluate.go`)

- `Evaluate(target ColorTarget, reading meter.Reading) → Advice`
  Computes ΔE2000, determines which controls to adjust and by how much. Uses the projector's capabilities to generate relevant adjustments (e.g., only suggests RGB Gain if the projector exposes it).

- `DeltaE2000(x1, y1, Y1, x2, y2, Y2) → float64`
  CIE ΔE2000 computation (xy + Y → Lab → ΔE2000).

### 2.4 Routine (`routine.go`)

```go
type StepKind string
const (
    StepPreCal     StepKind = "precal"
    StepWhiteBal   StepKind = "whitebalance"
    StepGamma      StepKind = "gamma"
    StepCMS        StepKind = "cms"
    StepVerify     StepKind = "verify"
)

type Step struct {
    Kind        StepKind
    Label       string
    Description string
    Patterns    []pattern.Pattern   // patterns to cycle through
    Targets     []ColorTarget       // corresponding targets
    Tolerance   float64             // ΔE pass threshold
    Optional    bool                // skip if projector lacks controls
    RequiresControls []string       // projector control IDs needed
}

func BuildRoutine(std Standard, caps projector.Capabilities) []Step
```

`BuildRoutine` generates the step list, filtering out steps that require controls the projector doesn't have (e.g., skip CMS if no CMS controls).

### 2.5 New App Bindings (`app.go`)

```go
func (a *App) GetCalibrationSteps(standard string) ([]Step, error)
func (a *App) EvaluateMeasurement(target ColorTarget) (Advice, error)
    // takes a measurement and evaluates against the given target
```

---

## 3. Frontend — Animation System

### 3.1 Core Animation Patterns

All animations use framer-motion with spring physics for a natural, premium feel.

**Page transitions** — Steps slide in/out horizontally with a shared layout animation:
```
← Previous step slides out left | New step slides in from right →
```
Using `AnimatePresence` + `motion.div` with `initial/animate/exit` variants.

**Number transitions** — All numeric values (ΔE, xy, luminance, CCT) animate smoothly between values using `motion.span` with `animate={{ opacity }}` and a custom counting hook that interpolates between old and new values over ~400ms.

**Convergence indicators** — Animated rings/arcs that fill as measurements approach targets. Spring-animated so they overshoot slightly and settle (feels alive, not mechanical).

**Success celebrations** — When a step passes:
- The convergence ring pulses green with a scale spring
- A subtle confetti burst (small colored dots that fade out)
- The "Next" button animates in with a bounce

**Live crosshair on chromaticity** — An animated dot on a mini CIE diagram that moves with spring physics toward the target crosshair. Shows the measurement converging in real-time.

### 3.2 Color Palette for States

| State | Color | Usage |
|-------|-------|-------|
| Measuring | `#6366f1` (indigo) | Pulsing ring, active indicators |
| Converging | `#f59e0b` (amber) | Getting closer but not there yet |
| Passed | `#22c55e` (green) | Step complete, within tolerance |
| Failed | `#ef4444` (red) | Out of tolerance, needs adjustment |
| Idle | `var(--muted)` | Waiting for action |

---

## 4. Frontend — Components

### 4.1 `CalibrationWizard.tsx` — Main Stepper

The top-level component that manages the calibration flow.

**Layout:**
```
┌──────────────────────────────────────────────────┐
│  Step Progress Bar (animated, shows all steps)   │
├──────────────────────────────────────────────────┤
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │                                            │  │
│  │         Active Step Content                │  │
│  │         (slides in/out with animation)     │  │
│  │                                            │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
├──────────────────────────────────────────────────┤
│  [← Back]              Step 2 of 5    [Next →]   │
└──────────────────────────────────────────────────┘
```

**Progress bar:** Connected dots with animated fill between them. Completed steps have a green check with a pop-in animation. Current step pulses. Future steps are dimmed.

**Step transitions:** `AnimatePresence` with directional slide — going forward slides left-to-right, going back slides right-to-left.

### 4.2 `ConvergenceRing.tsx` — Animated Target Indicator

A circular gauge that shows how close the current measurement is to the target.

```
        ╭───────╮
      ╱   ΔE 1.2  ╲        ← Large number in center, animates between values
     │    ────────  │
     │   Target: ≤2 │       ← Small target label
      ╲            ╱
        ╰───────╯
     [████████░░░░]         ← Arc fills based on proximity (100% = on target)
```

- Arc animated with spring physics via framer-motion's `motion.circle` with `pathLength`
- Color transitions smoothly between red → amber → green as ΔE decreases
- Pulses gently while actively measuring

### 4.3 `LiveCrosshair.tsx` — Mini CIE Diagram with Animated Dot

A small (200×200) CIE 1931 diagram showing:
- Target point (static crosshair, white)
- Measured point (animated dot, accent color, spring-animated position)
- Trail of last N measurements (fading dots showing convergence path)

The measured dot moves with `motion.div` using `spring` transition so it overshoots slightly and settles — gives a visceral sense of the measurement changing.

### 4.4 `AdjustmentCard.tsx` — What to Change

Shows the current adjustment advice with animated entry:

```
┌─────────────────────────────────┐
│  ↑ Increase Red Gain            │
│    slightly                     │
│    Red is 0.008 above target    │
│                                 │
│  ↓ Decrease Blue Gain           │
│    moderately                   │
│    Blue is 0.015 above target   │
└─────────────────────────────────┘
```

Cards stagger-animate in using `motion.div` with `variants` and `staggerChildren`. Direction arrows animate (bounce up/down). Cards fade out and new ones fade in when advice changes.

### 4.5 `MeasureButton.tsx` — Animated Measure Trigger

A large button that:
- Shows "Measure" in idle state
- Animates to a spinning ring while measuring
- Pops to a checkmark on success
- Shakes on error

Uses `AnimatePresence` to swap between states with crossfade.

### 4.6 `StepHeader.tsx` — Step Title + Description

Animated text that slides in with the step. Includes:
- Step number badge (animated counter)
- Title (slides in)
- Description (fades in with delay)
- "What you'll do" bullet points (stagger in)

---

## 5. Calibration Steps — Detail

### Step 1: Pre-Cal Checks

**Purpose:** Establish baseline before calibration.

**Patterns:** Black (0%), White (100%), 1% steps near black, 99% steps near white, 10% window, 100% field.

**UI:** Automated sweep — the app cycles through patterns, takes measurements, and presents results as animated cards that flip in one by one:
- Black level: `0.032 cd/m²` ✓
- Peak luminance: `48 cd/m² (10% win)` ✓
- Clipping: `Level 1 visible` ✓
- ABL drop: `12%` ⚠
- Contrast ratio: `1500:1` ✓

Each card has a status icon that animates (checkmark draws itself, warning pulses).

**No user adjustments needed** — this is informational. "Next" is always available.

### Step 2: White Balance (2-Point)

**Purpose:** Get the grayscale tracking to D65 (or target white point).

**Patterns:** 20% IRE window, 80% IRE window (alternating).

**UI flow:**
1. Display 80% IRE → "Adjust RGB Gain to match target"
2. Live measurement loop:
   - ConvergenceRing shows ΔE from D65
   - LiveCrosshair shows xy moving toward target
   - AdjustmentCards show which gain to adjust
3. When 80% passes (ΔE < 1.5), auto-switch to 20% IRE
4. "Adjust RGB Bias (offset) to match target"
5. Same live loop for bias
6. Iterate: recheck 80% (gain may have drifted), then 20% again
7. Both pass → step complete with celebration animation

**This is the most interactive step** — the user is actively turning gain/bias controls on the projector while watching the crosshair converge. The spring-animated dot on the CIE diagram is the hero visual here.

### Step 3: Gamma / EOTF Verification

**Purpose:** Verify gamma tracking matches target curve.

**Patterns:** 0–100% IRE in 5% steps (windows).

**UI flow:**
1. Automated sweep — app cycles through all IRE levels, measuring each
2. Animated chart builds up in real-time: each point appears with a pop animation as it's measured
3. Shows target curve (dashed) vs measured curve (solid, drawing itself left to right)
4. If projector has gamma presets, suggest the best-matching preset
5. User selects preset → re-sweep to verify

**Less interactive** — mostly watching the chart build. The animation of the curve drawing itself point-by-point is the visual highlight.

### Step 4: CMS (Color Management)

**Purpose:** Adjust primary/secondary colors to match target gamut.

**Patterns:** R, G, B, C, M, Y full-field and windows.

**Conditional:** Only shown if projector has CMS controls (hue/saturation/luminance per color).

**UI flow:**
1. For each color (R, G, B, C, M, Y):
   - Display color pattern
   - Measure
   - Show on mini CIE diagram: target point vs measured point
   - AdjustmentCards for hue/sat/lum
   - Live loop until ΔE < 2
2. Mini gamut diagram animates: as each color is calibrated, the measured gamut triangle/hexagon morphs toward the target

**Visual highlight:** The animated gamut hexagon morphing as each color gets dialed in.

### Step 5: Verification

**Purpose:** Full measurement sweep to confirm calibration quality.

**Patterns:** Full battery — grayscale, primaries, secondaries, saturation sweeps, ColorChecker.

**UI flow:**
1. Automated sweep (30-60 measurements)
2. Progress ring fills as measurements complete
3. Results appear as an animated scorecard:

```
┌─────────────────────────────────┐
│         CALIBRATION REPORT      │
│                                 │
│    Average ΔE:  1.2  ✓         │  ← number counts up from 0
│    Max ΔE:      2.8  ✓         │
│    Gamma:       2.18 ✓         │
│    White Point: 6504K ✓        │
│    Grayscale:   PASS           │  ← badge animates in
│    Gamut:       98.2% Rec.709  │
│                                 │
│    ★★★★☆  Excellent            │  ← stars pop in one by one
│                                 │
│    [Save Profile]  [Export]     │
└─────────────────────────────────┘
```

4. Star rating based on average ΔE:
   - ★★★★★ avg ΔE < 1.0 (Reference)
   - ★★★★☆ avg ΔE < 2.0 (Excellent)
   - ★★★☆☆ avg ΔE < 3.0 (Good)
   - ★★☆☆☆ avg ΔE < 5.0 (Fair)
   - ★☆☆☆☆ avg ΔE ≥ 5.0 (Poor)

---

## 6. Implementation Order

### Phase 1: Backend Foundation
1. `internal/calibration/types.go` — types
2. `internal/calibration/deltae.go` — ΔE2000 math (xy,Y → Lab → ΔE)
3. `internal/calibration/targets.go` — target generation from standards
4. `internal/calibration/evaluate.go` — measurement evaluation + advice
5. `internal/calibration/routine.go` — step builder
6. `app.go` — new bindings: `GetCalibrationSteps`, `EvaluateMeasurement`
7. `frontend/src/api/index.ts` — add API wrappers

### Phase 2: Frontend Skeleton
8. `npm install framer-motion`
9. `CalibrationWizard.tsx` — stepper with step transitions
10. `ConvergenceRing.tsx` — animated ΔE gauge
11. `LiveCrosshair.tsx` — mini CIE with animated dot
12. `AdjustmentCard.tsx` — advice display
13. `MeasureButton.tsx` — animated measure trigger

### Phase 3: Step Implementations
14. `steps/PreCalStep.tsx` — automated sweep with card reveals
15. `steps/WhiteBalanceStep.tsx` — interactive gain/bias loop
16. `steps/GammaStep.tsx` — sweep with animated chart
17. `steps/CMSStep.tsx` — per-color adjustment with gamut morph
18. `steps/VerifyStep.tsx` — final sweep with scorecard

### Phase 4: Polish
19. Number transition hook (`useAnimatedValue`)
20. Confetti/celebration on step completion
21. Sound effects (optional — subtle click on measure, chime on pass)
22. Keyboard shortcuts (Space = measure, Enter = next, Esc = back)

---

## 7. File Structure

```
internal/calibration/
  types.go
  deltae.go
  targets.go
  evaluate.go
  routine.go

frontend/src/pages/
  Calibration.tsx              → routes to CalibrationWizard

frontend/src/components/cal/
  CalibrationWizard.tsx        → main stepper
  ConvergenceRing.tsx          → animated ΔE ring
  LiveCrosshair.tsx            → mini CIE + animated dot
  AdjustmentCard.tsx           → advice cards
  MeasureButton.tsx            → animated button
  StepProgress.tsx             → top progress bar
  ScoreCard.tsx                → final report
  useAnimatedValue.ts          → smooth number transitions

frontend/src/components/cal/steps/
  PreCalStep.tsx
  WhiteBalanceStep.tsx
  GammaStep.tsx
  CMSStep.tsx
  VerifyStep.tsx
```


---

## 8. Automatic Calibration — Reuse Strategy

The assisted and automatic flows share the same backend engine and most of the same frontend components. The difference is who performs the adjustments: the user (assisted) or the app via `SetProjectorControl` (automatic).

### 8.1 Shared Backend (100% reuse)

Everything in `internal/calibration/` is mode-agnostic:

| Module | Used by Assisted | Used by Automatic |
|--------|:---:|:---:|
| `types.go` — targets, advice | ✓ | ✓ |
| `deltae.go` — ΔE2000 math | ✓ | ✓ |
| `targets.go` — target generation | ✓ | ✓ |
| `evaluate.go` — measurement → advice | ✓ | ✓ |
| `routine.go` — step builder | ✓ | ✓ |

The `Evaluate()` function returns `Advice` with `Adjustments` — in assisted mode these are displayed to the user; in automatic mode they're fed directly to `SetProjectorControl`.

### 8.2 Automatic Control Loop (`internal/calibration/autoloop.go`)

New file, only used by automatic mode:

```go
type AutoLoop struct {
    manager    *core.Manager
    routine    []Step
    maxIter    int           // max iterations per step (prevent infinite loops)
    tolerance  float64       // ΔE threshold to pass
}

func (l *AutoLoop) Run(ctx context.Context, onProgress func(StepProgress)) error
```

The loop per step:
1. Set pattern via `SetActivePattern`
2. Measure via `Measure`
3. Evaluate via `Evaluate`
4. If passed → next step
5. If not → apply adjustments via `SetProjectorControl`, go to 2
6. If max iterations reached → mark step as best-effort, move on

The `onProgress` callback emits events to the frontend for live UI updates.

### 8.3 Shared Frontend Components (90% reuse)

| Component | Assisted | Automatic | Difference |
|-----------|:---:|:---:|------------|
| `CalibrationWizard.tsx` | ✓ | ✓ | Automatic hides Back/Next, auto-advances |
| `StepProgress.tsx` | ✓ | ✓ | Identical |
| `ConvergenceRing.tsx` | ✓ | ✓ | Identical — shows live ΔE in both modes |
| `LiveCrosshair.tsx` | ✓ | ✓ | Identical — animated dot converging |
| `AdjustmentCard.tsx` | ✓ | ✓ | Assisted: "please adjust X". Automatic: "adjusting X…" |
| `MeasureButton.tsx` | ✓ | ✗ | Automatic measures automatically |
| `ScoreCard.tsx` | ✓ | ✓ | Identical |
| Step components | ✓ | ✓ | Each step accepts a `mode` prop |

### 8.4 Mode Prop Pattern

Each step component receives a `mode: "assisted" | "automatic"` prop that controls behavior:

```tsx
// In WhiteBalanceStep.tsx
if (mode === "automatic") {
  // Auto-loop: measure → evaluate → apply → repeat
  useEffect(() => {
    const interval = setInterval(async () => {
      const reading = await api.measure();
      const advice = await api.evaluate(target, reading);
      if (advice.passed) { onComplete(); return; }
      for (const adj of advice.adjustments) {
        const current = await api.getControl(adj.controlID);
        const delta = adj.direction === "increase" ? adj.step : -adj.step;
        await api.setControl(adj.controlID, current + delta);
      }
    }, 1500);
    return () => clearInterval(interval);
  }, []);
}

if (mode === "assisted") {
  // Show MeasureButton + AdjustmentCards, user does the work
}
```

### 8.5 What's Different in Automatic Mode

| Aspect | Assisted | Automatic |
|--------|----------|-----------|
| Who adjusts controls | User | App via `SetProjectorControl` |
| Measure trigger | User clicks button | Automatic on timer |
| Step advancement | User clicks Next | Auto-advance on pass |
| Back button | Available | Hidden |
| Skip button | Available | Hidden (auto-skips on max iterations) |
| AdjustmentCard text | "Increase Red Gain" | "Adjusting Red Gain…" with spinner |
| Duration | 15–30 min (user speed) | 5–10 min (automated) |
| Projector requirement | Any (user adjusts OSD) | Must have RS232/IP control for all needed settings |

### 8.6 Automatic Mode Prerequisites

Before starting automatic calibration, the app checks:
1. Projector is connected and supports programmatic control
2. Projector capabilities include the controls needed for each step
3. Meter is connected and responding

If the projector lacks certain controls (e.g., no CMS over RS232), those steps are skipped automatically — same logic as assisted mode's `BuildRoutine` filtering.

### 8.7 Implementation Order

Automatic mode is built after assisted mode since it reuses everything:

1. Build assisted mode fully (Phase 1–4 from section 6)
2. Add `autoloop.go` to backend
3. Add `mode` prop to each step component
4. Add automatic-specific UI tweaks (auto-advance, spinner on adjustments)
5. Wire up the `/calibration/automatic` route

Estimated additional work beyond assisted: ~20% — the architecture is designed so automatic is a thin layer on top of assisted.
