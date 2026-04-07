package icc

// Primaries holds CIE xy chromaticity coordinates for RGB primaries and white point.
type Primaries struct {
	Rx, Ry float64
	Gx, Gy float64
	Bx, By float64
	Wx, Wy float64
}

// Standard primaries.
var (
	PrimariesSRGB = Primaries{
		Rx: 0.64, Ry: 0.33, Gx: 0.30, Gy: 0.60, Bx: 0.15, By: 0.06,
		Wx: 0.3127, Wy: 0.3290,
	}
	PrimariesRec709  = PrimariesSRGB // same chromaticities
	PrimariesP3D65   = Primaries{
		Rx: 0.68, Ry: 0.32, Gx: 0.265, Gy: 0.69, Bx: 0.15, By: 0.06,
		Wx: 0.3127, Wy: 0.3290,
	}
	PrimariesBT2020  = Primaries{
		Rx: 0.708, Ry: 0.292, Gx: 0.170, Gy: 0.797, Bx: 0.131, By: 0.046,
		Wx: 0.3127, Wy: 0.3290,
	}
)

// NewDisplayProfile creates a v2.4 display (mntr) profile from primaries and a TRC.
// The primaries are chromatically adapted from their native white to D50 using Bradford.
func NewDisplayProfile(desc string, p Primaries, trc Curve) *Profile {
	prof := NewProfile()
	prof.Header.DeviceClass = ClassDisplay
	prof.Header.ColorSpace = SpaceRGB

	// Compute native white in XYZ (Y=1).
	white := xyToXYZ(p.Wx, p.Wy)

	// Compute unscaled primary XYZ from chromaticity.
	rRaw := xyToXYZ(p.Rx, p.Ry)
	gRaw := xyToXYZ(p.Gx, p.Gy)
	bRaw := xyToXYZ(p.Bx, p.By)

	// Solve for scaling factors: M * [Sr, Sg, Sb]^T = white.
	sr, sg, sb := solvePrimaryScaling(rRaw, gRaw, bRaw, white)
	rXYZ := XYZ{rRaw.X * sr, rRaw.Y * sr, rRaw.Z * sr}
	gXYZ := XYZ{gRaw.X * sg, gRaw.Y * sg, gRaw.Z * sg}
	bXYZ := XYZ{bRaw.X * sb, bRaw.Y * sb, bRaw.Z * sb}

	// Bradford adaptation from native white to D50.
	chad := BradfordMatrix(white, D50XYZ)

	// Adapt primaries to D50.
	rD50 := adaptXYZ(chad, rXYZ)
	gD50 := adaptXYZ(chad, gXYZ)
	bD50 := adaptXYZ(chad, bXYZ)

	// Set tags.
	prof.SetTag(TagProfileDesc, EncodeTextDescTag(desc))
	prof.SetTag(TagCopyright, EncodeTextDescTag("No copyright"))
	prof.SetTag(TagMediaWhitePoint, EncodeXYZTag(D50XYZ))
	prof.SetTag(TagChromAdapt, EncodeS15Fixed16ArrayTag(chad[:]))
	prof.SetTag(TagRedMatrixCol, EncodeXYZTag(rD50))
	prof.SetTag(TagGreenMatrixCol, EncodeXYZTag(gD50))
	prof.SetTag(TagBlueMatrixCol, EncodeXYZTag(bD50))

	trcData := EncodeCurveTag(trc)
	prof.SetTag(TagRedTRC, trcData)
	prof.SetTag(TagGreenTRC, trcData)
	prof.SetTag(TagBlueTRC, trcData)

	return prof
}

// NewDisplayProfileXYZ creates a display profile from pre-measured XYZ primaries
// (already at Y=1 white) and per-channel TRCs. Useful when you have separate
// R/G/B tone curves from calibration measurements.
func NewDisplayProfileXYZ(desc string, r, g, b, white XYZ, rTRC, gTRC, bTRC Curve) *Profile {
	prof := NewProfile()
	prof.Header.DeviceClass = ClassDisplay
	prof.Header.ColorSpace = SpaceRGB

	chad := BradfordMatrix(white, D50XYZ)
	rD50 := adaptXYZ(chad, r)
	gD50 := adaptXYZ(chad, g)
	bD50 := adaptXYZ(chad, b)

	prof.SetTag(TagProfileDesc, EncodeTextDescTag(desc))
	prof.SetTag(TagCopyright, EncodeTextDescTag("No copyright"))
	prof.SetTag(TagMediaWhitePoint, EncodeXYZTag(D50XYZ))
	prof.SetTag(TagChromAdapt, EncodeS15Fixed16ArrayTag(chad[:]))
	prof.SetTag(TagRedMatrixCol, EncodeXYZTag(rD50))
	prof.SetTag(TagGreenMatrixCol, EncodeXYZTag(gD50))
	prof.SetTag(TagBlueMatrixCol, EncodeXYZTag(bD50))
	prof.SetTag(TagRedTRC, EncodeCurveTag(rTRC))
	prof.SetTag(TagGreenTRC, EncodeCurveTag(gTRC))
	prof.SetTag(TagBlueTRC, EncodeCurveTag(bTRC))

	return prof
}

// Convenience constructors for standard profiles.

func NewSRGBProfile() *Profile {
	return NewDisplayProfile("sRGB IEC61966-2.1", PrimariesSRGB, SRGBCurve())
}

func NewRec709Profile() *Profile {
	return NewDisplayProfile("Rec. 709", PrimariesRec709, GammaCurve(2.2))
}

func NewP3D65Profile() *Profile {
	return NewDisplayProfile("Display P3", PrimariesP3D65, GammaCurve(2.2))
}

func NewBT2020Profile() *Profile {
	return NewDisplayProfile("BT.2020", PrimariesBT2020, GammaCurve(2.2))
}

// --- helpers ---

// xyToXYZ converts CIE xy chromaticity to XYZ with Y=1.
func xyToXYZ(x, y float64) XYZ {
	return XYZ{X: x / y, Y: 1, Z: (1 - x - y) / y}
}

// solvePrimaryScaling finds Sr, Sg, Sb such that Sr*R + Sg*G + Sb*B = W.
// This is a 3x3 linear system solve via Cramer's rule.
func solvePrimaryScaling(r, g, b, w XYZ) (sr, sg, sb float64) {
	// Matrix columns are the primaries.
	det := r.X*(g.Y*b.Z-g.Z*b.Y) - g.X*(r.Y*b.Z-r.Z*b.Y) + b.X*(r.Y*g.Z-r.Z*g.Y)
	sr = (w.X*(g.Y*b.Z-g.Z*b.Y) - g.X*(w.Y*b.Z-w.Z*b.Y) + b.X*(w.Y*g.Z-w.Z*g.Y)) / det
	sg = (r.X*(w.Y*b.Z-w.Z*b.Y) - w.X*(r.Y*b.Z-r.Z*b.Y) + b.X*(r.Y*w.Z-r.Z*w.Y)) / det
	sb = (r.X*(g.Y*w.Z-g.Z*w.Y) - g.X*(r.Y*w.Z-r.Z*w.Y) + w.X*(r.Y*g.Z-r.Z*g.Y)) / det
	return
}

// adaptXYZ applies a 3x3 adaptation matrix to an XYZ value.
func adaptXYZ(m [9]float64, v XYZ) XYZ {
	return XYZ{
		X: m[0]*v.X + m[1]*v.Y + m[2]*v.Z,
		Y: m[3]*v.X + m[4]*v.Y + m[5]*v.Z,
		Z: m[6]*v.X + m[7]*v.Y + m[8]*v.Z,
	}
}
