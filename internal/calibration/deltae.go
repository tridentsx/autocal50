package calibration

import "math"

// DeltaE2000 computes CIE ΔE2000 between two colors given in xyY.
func DeltaE2000(x1, y1, Y1, x2, y2, Y2 float64) float64 {
	L1, a1, b1 := xyYToLab(x1, y1, Y1)
	L2, a2, b2 := xyYToLab(x2, y2, Y2)
	return de2000(L1, a1, b1, L2, a2, b2)
}

func xyYToLab(x, y, Y float64) (L, a, b float64) {
	if y == 0 {
		return 0, 0, 0
	}
	// xyY → XYZ
	X := (Y / y) * x
	Z := (Y / y) * (1 - x - y)
	// D65 reference white
	Xn, Yn, Zn := 95.047, 100.0, 108.883
	// XYZ → Lab
	fx := labF(X / Xn)
	fy := labF(Y * 100 / Yn) // Y is in cd/m², normalize
	fz := labF(Z / Zn)
	L = 116*fy - 16
	a = 500 * (fx - fy)
	b = 200 * (fy - fz)
	return
}

func labF(t float64) float64 {
	if t > 0.008856 {
		return math.Cbrt(t)
	}
	return 7.787*t + 16.0/116.0
}

func de2000(L1, a1, b1, L2, a2, b2 float64) float64 {
	// CIE ΔE2000 implementation
	C1 := math.Sqrt(a1*a1 + b1*b1)
	C2 := math.Sqrt(a2*a2 + b2*b2)
	Cb := (C1 + C2) / 2

	Cb7 := math.Pow(Cb, 7)
	G := 0.5 * (1 - math.Sqrt(Cb7/(Cb7+math.Pow(25, 7))))

	a1p := a1 * (1 + G)
	a2p := a2 * (1 + G)

	C1p := math.Sqrt(a1p*a1p + b1*b1)
	C2p := math.Sqrt(a2p*a2p + b2*b2)

	h1p := hpF(b1, a1p)
	h2p := hpF(b2, a2p)

	dLp := L2 - L1
	dCp := C2p - C1p
	dhp := dhpF(C1p, C2p, h1p, h2p)
	dHp := 2 * math.Sqrt(C1p*C2p) * math.Sin(dhp/2*math.Pi/180)

	Lbp := (L1 + L2) / 2
	Cbp := (C1p + C2p) / 2

	hbp := hbpF(C1p, C2p, h1p, h2p)

	T := 1 - 0.17*math.Cos((hbp-30)*math.Pi/180) +
		0.24*math.Cos((2*hbp)*math.Pi/180) +
		0.32*math.Cos((3*hbp+6)*math.Pi/180) -
		0.20*math.Cos((4*hbp-63)*math.Pi/180)

	SL := 1 + 0.015*(Lbp-50)*(Lbp-50)/math.Sqrt(20+(Lbp-50)*(Lbp-50))
	SC := 1 + 0.045*Cbp
	SH := 1 + 0.015*Cbp*T

	Cbp7 := math.Pow(Cbp, 7)
	RT := -2 * math.Sqrt(Cbp7/(Cbp7+math.Pow(25, 7))) *
		math.Sin(60*math.Exp(-((hbp-275)/25)*((hbp-275)/25))*math.Pi/180)

	return math.Sqrt(
		(dLp/SL)*(dLp/SL) +
			(dCp/SC)*(dCp/SC) +
			(dHp/SH)*(dHp/SH) +
			RT*(dCp/SC)*(dHp/SH))
}

func hpF(b, ap float64) float64 {
	if b == 0 && ap == 0 {
		return 0
	}
	h := math.Atan2(b, ap) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return h
}

func dhpF(C1p, C2p, h1p, h2p float64) float64 {
	if C1p*C2p == 0 {
		return 0
	}
	d := h2p - h1p
	if d > 180 {
		d -= 360
	} else if d < -180 {
		d += 360
	}
	return d
}

func hbpF(C1p, C2p, h1p, h2p float64) float64 {
	if C1p*C2p == 0 {
		return h1p + h2p
	}
	if math.Abs(h1p-h2p) <= 180 {
		return (h1p + h2p) / 2
	}
	if h1p+h2p < 360 {
		return (h1p + h2p + 360) / 2
	}
	return (h1p + h2p - 360) / 2
}
