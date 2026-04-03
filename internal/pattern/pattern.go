package pattern

import "fmt"

// Pattern describes a test pattern to render on the projector output.
type Pattern struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Group    string  `json:"group"`
	Type     string  `json:"type"` // solid, window, gradient, grid, checker
	Color    *RGB    `json:"color,omitempty"`
	BgColor  *RGB    `json:"bgColor,omitempty"`
	WindowPc float64 `json:"windowPc,omitempty"` // window size as % of screen
	Steps    []RGB   `json:"steps,omitempty"`     // for gradients
	Cols     int     `json:"cols,omitempty"`       // for grid/checker
	Rows     int     `json:"rows,omitempty"`
}

type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

func ire(pct int) uint8 { return uint8(pct * 255 / 100) }

func solid(id, name, group string, r, g, b uint8) Pattern {
	return Pattern{ID: id, Name: name, Group: group, Type: "solid", Color: &RGB{r, g, b}}
}

func window(id, name, group string, r, g, b uint8, pct float64) Pattern {
	return Pattern{ID: id, Name: name, Group: group, Type: "window", Color: &RGB{r, g, b}, BgColor: &RGB{0, 0, 0}, WindowPc: pct}
}

// Battery returns the full set of calibration test patterns.
func Battery() []Pattern {
	var p []Pattern

	// === Grayscale: full-field IRE 0-100 in 5% steps ===
	for i := 0; i <= 100; i += 5 {
		v := ire(i)
		p = append(p, solid(
			fmt.Sprintf("gray-%d", i),
			fmt.Sprintf("%d%% IRE", i),
			"Grayscale",
			v, v, v,
		))
	}

	// === Grayscale: 18% windows for meter measurement ===
	for i := 0; i <= 100; i += 5 {
		v := ire(i)
		p = append(p, window(
			fmt.Sprintf("gray-win-%d", i),
			fmt.Sprintf("%d%% Window", i),
			"Grayscale Windows",
			v, v, v, 18,
		))
	}

	// === Clipping: near-black 0-5% in 1% steps ===
	for i := 0; i <= 5; i++ {
		v := ire(i)
		p = append(p, solid(
			fmt.Sprintf("clip-black-%d", i),
			fmt.Sprintf("Black %d%%", i),
			"Clipping – Black",
			v, v, v,
		))
	}

	// === Clipping: near-white 95-100% in 1% steps ===
	for i := 95; i <= 100; i++ {
		v := ire(i)
		p = append(p, solid(
			fmt.Sprintf("clip-white-%d", i),
			fmt.Sprintf("White %d%%", i),
			"Clipping – White",
			v, v, v,
		))
	}

	// === Primaries ===
	p = append(p,
		solid("red-100", "Red", "Primaries", 255, 0, 0),
		solid("green-100", "Green", "Primaries", 0, 255, 0),
		solid("blue-100", "Blue", "Primaries", 0, 0, 255),
	)

	// === Secondaries ===
	p = append(p,
		solid("cyan-100", "Cyan", "Secondaries", 0, 255, 255),
		solid("magenta-100", "Magenta", "Secondaries", 255, 0, 255),
		solid("yellow-100", "Yellow", "Secondaries", 255, 255, 0),
	)

	// === Color windows (18%) for meter measurement ===
	for _, c := range []struct {
		name string
		r, g, b uint8
	}{
		{"Red", 255, 0, 0}, {"Green", 0, 255, 0}, {"Blue", 0, 0, 255},
		{"Cyan", 0, 255, 255}, {"Magenta", 255, 0, 255}, {"Yellow", 255, 255, 0},
	} {
		p = append(p, window(
			fmt.Sprintf("win-%s", c.name),
			fmt.Sprintf("%s Window", c.name),
			"Color Windows",
			c.r, c.g, c.b, 18,
		))
	}

	// === Saturation sweeps at 20/40/60/80/100% ===
	// Saturation is mixed toward white at 100% stimulus
	for _, sat := range []int{20, 40, 60, 80, 100} {
		g := fmt.Sprintf("Saturation %d%%", sat)
		s := float64(sat) / 100.0
		lo := uint8((1.0 - s) * 255) // the desaturated channel value
		p = append(p,
			solid(fmt.Sprintf("sat-red-%d", sat), fmt.Sprintf("Red %d%%", sat), g, 255, lo, lo),
			solid(fmt.Sprintf("sat-green-%d", sat), fmt.Sprintf("Green %d%%", sat), g, lo, 255, lo),
			solid(fmt.Sprintf("sat-blue-%d", sat), fmt.Sprintf("Blue %d%%", sat), g, lo, lo, 255),
			solid(fmt.Sprintf("sat-cyan-%d", sat), fmt.Sprintf("Cyan %d%%", sat), g, lo, 255, 255),
			solid(fmt.Sprintf("sat-magenta-%d", sat), fmt.Sprintf("Magenta %d%%", sat), g, 255, lo, 255),
			solid(fmt.Sprintf("sat-yellow-%d", sat), fmt.Sprintf("Yellow %d%%", sat), g, 255, 255, lo),
		)
	}

	// === Peak vs Window Size (ABL detection) ===
	for _, pct := range []float64{1, 2, 5, 10, 25, 50, 75, 100} {
		id := fmt.Sprintf("abl-%g", pct)
		name := fmt.Sprintf("White %g%%", pct)
		if pct == 100 {
			p = append(p, solid(id, name, "Peak vs Size", 255, 255, 255))
		} else {
			p = append(p, window(id, name, "Peak vs Size", 255, 255, 255, pct))
		}
	}

	// === HDR EOTF Steps (PQ code values for specific nit levels) ===
	// ST.2084 PQ: nits → 10-bit code value → 8-bit approximation
	for _, nits := range []struct {
		label string
		code  uint8
	}{
		{"1 nit", 15}, {"10 nit", 30}, {"100 nit", 64},
		{"203 nit", 75}, {"350 nit", 84}, {"600 nit", 95},
		{"1000 nit", 109}, {"2000 nit", 127}, {"4000 nit", 148},
	} {
		p = append(p, window(
			fmt.Sprintf("hdr-%s", nits.label),
			nits.label,
			"HDR EOTF Steps",
			nits.code, nits.code, nits.code, 18,
		))
	}

	// === Contrast Ratio ===
	p = append(p,
		solid("cr-black", "Black (0%)", "Contrast Ratio", 0, 0, 0),
		solid("cr-white", "White (100%)", "Contrast Ratio", 255, 255, 255),
		Pattern{ID: "cr-checker-4x4", Name: "4×4 Checkerboard", Group: "Contrast Ratio", Type: "checker",
			Cols: 4, Rows: 4, Steps: checkerboard(4, 4, RGB{0, 0, 0}, RGB{255, 255, 255})},
		Pattern{ID: "cr-checker-1px", Name: "1px Checkerboard", Group: "Contrast Ratio", Type: "checker",
			Cols: 16, Rows: 9, Steps: checkerboard(16, 9, RGB{0, 0, 0}, RGB{255, 255, 255})},
	)

	// === Window patterns ===
	for _, pct := range []float64{10, 18, 50} {
		p = append(p, window(
			fmt.Sprintf("window-white-%g", pct),
			fmt.Sprintf("White Window %g%%", pct),
			"Windows",
			255, 255, 255, pct,
		))
	}

	// === Gradients ===
	steps := make([]RGB, 256)
	for i := range steps {
		steps[i] = RGB{uint8(i), uint8(i), uint8(i)}
	}
	p = append(p, Pattern{ID: "gradient-gray", Name: "Gray Gradient", Group: "Gradients", Type: "gradient", Steps: steps})

	for _, ch := range []struct {
		name    string
		r, g, b bool
	}{
		{"Red", true, false, false}, {"Green", false, true, false}, {"Blue", false, false, true},
		{"Cyan", false, true, true}, {"Magenta", true, false, true}, {"Yellow", true, true, false},
	} {
		gs := make([]RGB, 256)
		for i := range gs {
			var rv, gv, bv uint8
			if ch.r { rv = uint8(i) }
			if ch.g { gv = uint8(i) }
			if ch.b { bv = uint8(i) }
			gs[i] = RGB{rv, gv, bv}
		}
		p = append(p, Pattern{
			ID: fmt.Sprintf("gradient-%s", ch.name), Name: fmt.Sprintf("%s Gradient", ch.name),
			Group: "Gradients", Type: "gradient", Steps: gs,
		})
	}

	// === Geometry ===
	p = append(p, Pattern{ID: "crosshatch", Name: "Crosshatch", Group: "Geometry", Type: "grid",
		Color: &RGB{255, 255, 255}, BgColor: &RGB{0, 0, 0}, Cols: 16, Rows: 9})

	// === ColorChecker ===
	p = append(p, Pattern{ID: "checker", Name: "Color Checker", Group: "ColorChecker", Type: "checker",
		Cols: 6, Rows: 4, Steps: colorCheckerPalette()})

	return p
}

func checkerboard(cols, rows int, a, b RGB) []RGB {
	out := make([]RGB, cols*rows)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if (r+c)%2 == 0 {
				out[r*cols+c] = a
			} else {
				out[r*cols+c] = b
			}
		}
	}
	return out
}

func colorCheckerPalette() []RGB {
	// Simplified ColorChecker Classic 24-patch palette
	return []RGB{
		{115, 82, 68}, {194, 150, 130}, {98, 122, 157}, {87, 108, 67},
		{133, 128, 177}, {103, 189, 170}, {214, 126, 44}, {80, 91, 166},
		{193, 90, 99}, {94, 60, 108}, {157, 188, 64}, {224, 163, 46},
		{56, 61, 150}, {70, 148, 73}, {175, 54, 60}, {231, 199, 31},
		{187, 86, 149}, {8, 133, 161}, {243, 243, 242}, {200, 200, 200},
		{160, 160, 160}, {122, 122, 121}, {85, 85, 85}, {52, 52, 52},
	}
}
