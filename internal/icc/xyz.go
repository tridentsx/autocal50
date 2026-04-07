package icc

import (
	"bytes"
	"math"
)

// XYZ is a CIE XYZ tristimulus value.
type XYZ struct {
	X, Y, Z float64
}

// D50XYZ is the ICC PCS illuminant.
var D50XYZ = XYZ{0.9642, 1.0, 0.8249}

// D65XYZ is the standard illuminant for sRGB/Rec709/etc.
var D65XYZ = XYZ{0.9505, 1.0, 1.0890}

// EncodeXYZTag encodes an XYZ value as an XYZType tag (ICC.1:2022 §10.31).
func EncodeXYZTag(v XYZ) []byte {
	var buf bytes.Buffer
	writeBytes(&buf, TypeXYZ, uint32(0), // type sig + reserved
		s15Fixed16FromFloat(v.X),
		s15Fixed16FromFloat(v.Y),
		s15Fixed16FromFloat(v.Z))
	return buf.Bytes()
}

// DecodeXYZTag decodes an XYZType tag. Returns the first XYZ value.
func DecodeXYZTag(data []byte) (XYZ, error) {
	if len(data) < 20 {
		return XYZ{}, errShort("XYZ")
	}
	return XYZ{
		X: s15Fixed16(be.Uint32(data[8:])).Float64(),
		Y: s15Fixed16(be.Uint32(data[12:])).Float64(),
		Z: s15Fixed16(be.Uint32(data[16:])).Float64(),
	}, nil
}

// EncodeS15Fixed16ArrayTag encodes a float64 slice as sf32 type.
func EncodeS15Fixed16ArrayTag(vals []float64) []byte {
	var buf bytes.Buffer
	writeBytes(&buf, TypeS15Fixed16Array, uint32(0))
	for _, v := range vals {
		writeBytes(&buf, s15Fixed16FromFloat(v))
	}
	return buf.Bytes()
}

// --- XYZ ↔ Lab (ICC.1:2022 §A.3, D50-relative) ---

func xyzToLab(v XYZ, wp XYZ) (L, a, b float64) {
	fx := labF(v.X / wp.X)
	fy := labF(v.Y / wp.Y)
	fz := labF(v.Z / wp.Z)
	L = 116*fy - 16
	a = 500 * (fx - fy)
	b = 200 * (fy - fz)
	return
}

func labToXYZ(L, a, b float64, wp XYZ) XYZ {
	fy := (L + 16) / 116
	fx := a/500 + fy
	fz := fy - b/200
	return XYZ{
		X: wp.X * labFInv(fx),
		Y: wp.Y * labFInv(fy),
		Z: wp.Z * labFInv(fz),
	}
}

const labDelta = 6.0 / 29.0

func labF(t float64) float64 {
	if t > labDelta*labDelta*labDelta {
		return math.Cbrt(t)
	}
	return t/(3*labDelta*labDelta) + 4.0/29.0
}

func labFInv(t float64) float64 {
	if t > labDelta {
		return t * t * t
	}
	return 3 * labDelta * labDelta * (t - 4.0/29.0)
}

// --- Chromatic adaptation (Bradford) ---

// Bradford cone response matrix and its inverse.
var (
	bradfordM = [9]float64{
		0.8951, 0.2664, -0.1614,
		-0.7502, 1.7135, 0.0367,
		0.0389, -0.0685, 1.0296,
	}
	bradfordMInv = [9]float64{
		0.9870, -0.1471, 0.1600,
		0.4323, 0.5184, 0.0493,
		-0.0085, 0.0400, 0.9685,
	}
)

// BradfordMatrix computes the 3×3 chromatic adaptation matrix from src to dst illuminant.
// Returns the matrix as 9 floats in row-major order.
func BradfordMatrix(src, dst XYZ) [9]float64 {
	// Transform illuminants to cone space.
	srcCone := mul3x1(bradfordM, src)
	dstCone := mul3x1(bradfordM, dst)

	// Diagonal scaling in cone space.
	scale := [9]float64{
		dstCone[0] / srcCone[0], 0, 0,
		0, dstCone[1] / srcCone[1], 0,
		0, 0, dstCone[2] / srcCone[2],
	}

	// M_adapt = M_inv * scale * M
	tmp := mul3x3(scale, bradfordM)
	return mul3x3(bradfordMInv, tmp)
}

func mul3x1(m [9]float64, v XYZ) [3]float64 {
	return [3]float64{
		m[0]*v.X + m[1]*v.Y + m[2]*v.Z,
		m[3]*v.X + m[4]*v.Y + m[5]*v.Z,
		m[6]*v.X + m[7]*v.Y + m[8]*v.Z,
	}
}

func mul3x3(a, b [9]float64) [9]float64 {
	var r [9]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				r[i*3+j] += a[i*3+k] * b[k*3+j]
			}
		}
	}
	return r
}

func errShort(name string) error {
	return &DecodeError{Tag: name, Msg: "data too short"}
}

// DecodeError is returned when tag data is malformed.
type DecodeError struct {
	Tag string
	Msg string
}

func (e *DecodeError) Error() string {
	return "icc: " + e.Tag + ": " + e.Msg
}
