package icc

import (
	"bytes"
	"math"
)

// Curve represents a tone response curve (TRC).
type Curve struct {
	// For a simple gamma: Gamma > 0, Table == nil, Params == nil.
	// For a table TRC: Table is populated.
	// For parametric: Params is populated.
	Gamma  float64
	Table  []uint16
	Params *ParametricCurve
}

// GammaCurve creates a simple power-law curve.
func GammaCurve(gamma float64) Curve {
	return Curve{Gamma: gamma}
}

// ParametricCurve holds parameters for parametricCurveType (ICC.1:2022 §10.18).
// Function type 3 (IEC 61966-2-1, i.e. sRGB):
//
//	Y = (aX + b)^g + c   for X >= d
//	Y = eX + f            for X < d
type ParametricCurve struct {
	FuncType uint16
	G, A, B, C, D, E, F float64
}

// SRGBCurve returns the sRGB parametric TRC (function type 3).
func SRGBCurve() Curve {
	return Curve{Params: &ParametricCurve{
		FuncType: 3,
		G: 2.4, A: 1.0 / 1.055, B: 0.055 / 1.055,
		C: 1.0 / 12.92, D: 0.04045, E: 0, F: 0,
	}}
}

// Eval evaluates the curve at input x in [0,1], returning output in [0,1].
func (c Curve) Eval(x float64) float64 {
	if x < 0 {
		x = 0
	} else if x > 1 {
		x = 1
	}

	if c.Params != nil {
		return c.Params.eval(x)
	}
	if len(c.Table) > 0 {
		return evalTable(c.Table, x)
	}
	if c.Gamma == 0 {
		return x // identity
	}
	return math.Pow(x, c.Gamma)
}

func (p *ParametricCurve) eval(x float64) float64 {
	switch p.FuncType {
	case 0: // Y = X^g
		return math.Pow(x, p.G)
	case 1: // Y = (aX+b)^g for X >= -b/a, else 0
		if x >= -p.B/p.A {
			return math.Pow(p.A*x+p.B, p.G)
		}
		return 0
	case 2: // Y = (aX+b)^g + c for X >= -b/a, else c
		if x >= -p.B/p.A {
			return math.Pow(p.A*x+p.B, p.G) + p.C
		}
		return p.C
	case 3: // Y = (aX+b)^g + c for X >= d, else eX+f  (sRGB)
		if x >= p.D {
			return math.Pow(p.A*x+p.B, p.G) + p.C
		}
		return p.E*x + p.F
	case 4: // Y = (aX+b)^g + c for X >= d, else eX+f
		if x >= p.D {
			return math.Pow(p.A*x+p.B, p.G) + p.C
		}
		return p.E*x + p.F
	}
	return x
}

func evalTable(table []uint16, x float64) float64 {
	n := len(table) - 1
	if n <= 0 {
		return x
	}
	pos := x * float64(n)
	lo := int(pos)
	if lo >= n {
		return float64(table[n]) / 65535
	}
	frac := pos - float64(lo)
	v := float64(table[lo])*(1-frac) + float64(table[lo+1])*frac
	return v / 65535
}

// EncodeCurveTag encodes a Curve as a curveType tag (ICC.1:2022 §10.6).
func EncodeCurveTag(c Curve) []byte {
	var buf bytes.Buffer
	writeBytes(&buf, TypeCurve, uint32(0)) // type sig + reserved

	if c.Params != nil {
		// Encode as parametricCurveType instead.
		return EncodeParametricCurveTag(*c.Params)
	}

	if len(c.Table) > 0 {
		writeBytes(&buf, uint32(len(c.Table)))
		for _, v := range c.Table {
			writeBytes(&buf, v)
		}
		return buf.Bytes()
	}

	if c.Gamma == 0 {
		// Identity: count = 0.
		writeBytes(&buf, uint32(0))
	} else if c.Gamma == 1.0 {
		// Identity: count = 0.
		writeBytes(&buf, uint32(0))
	} else {
		// Single gamma: count = 1, value is u8Fixed8.
		writeBytes(&buf, uint32(1), u8Fixed8FromFloat(c.Gamma))
	}
	return buf.Bytes()
}

// EncodeParametricCurveTag encodes a parametricCurveType tag (ICC.1:2022 §10.18).
func EncodeParametricCurveTag(p ParametricCurve) []byte {
	var buf bytes.Buffer
	writeBytes(&buf, TypeParametricCurve, uint32(0)) // type sig + reserved
	writeBytes(&buf, p.FuncType, uint16(0))          // funcType + reserved

	writeBytes(&buf, s15Fixed16FromFloat(p.G))
	if p.FuncType >= 1 {
		writeBytes(&buf, s15Fixed16FromFloat(p.A), s15Fixed16FromFloat(p.B))
	}
	if p.FuncType >= 2 {
		writeBytes(&buf, s15Fixed16FromFloat(p.C))
	}
	if p.FuncType >= 3 {
		writeBytes(&buf, s15Fixed16FromFloat(p.D))
	}
	if p.FuncType >= 4 {
		writeBytes(&buf, s15Fixed16FromFloat(p.E), s15Fixed16FromFloat(p.F))
	}
	return buf.Bytes()
}

// DecodeCurveTag decodes a curveType or parametricCurveType tag.
func DecodeCurveTag(data []byte) (Curve, error) {
	if len(data) < 12 {
		return Curve{}, errShort("curve")
	}
	var typeSig Signature
	copy(typeSig[:], data[:4])

	if typeSig == TypeParametricCurve {
		return decodeParametricCurve(data)
	}

	// curveType.
	count := be.Uint32(data[8:])
	switch {
	case count == 0:
		return Curve{Gamma: 1.0}, nil // identity
	case count == 1:
		if len(data) < 14 {
			return Curve{}, errShort("curve")
		}
		g := u8Fixed8(be.Uint16(data[12:])).Float64()
		return Curve{Gamma: g}, nil
	default:
		need := 12 + int(count)*2
		if len(data) < need {
			return Curve{}, errShort("curve")
		}
		table := make([]uint16, count)
		for i := range table {
			table[i] = be.Uint16(data[12+i*2:])
		}
		return Curve{Table: table}, nil
	}
}

func decodeParametricCurve(data []byte) (Curve, error) {
	if len(data) < 16 {
		return Curve{}, errShort("parametricCurve")
	}
	funcType := be.Uint16(data[8:])
	p := ParametricCurve{FuncType: funcType}

	off := 12
	readF := func() float64 {
		v := s15Fixed16(be.Uint32(data[off:])).Float64()
		off += 4
		return v
	}

	// Number of params per function type: 0→1, 1→3, 2→4, 3→5, 4→7
	needed := []int{16, 24, 28, 32, 40}
	if int(funcType) >= len(needed) {
		return Curve{}, &DecodeError{Tag: "parametricCurve", Msg: "unknown function type"}
	}
	if len(data) < needed[funcType] {
		return Curve{}, errShort("parametricCurve")
	}

	p.G = readF()
	if funcType >= 1 {
		p.A = readF()
		p.B = readF()
	}
	if funcType >= 2 {
		p.C = readF()
	}
	if funcType >= 3 {
		p.D = readF()
	}
	if funcType >= 4 {
		p.E = readF()
		p.F = readF()
	}
	return Curve{Params: &p}, nil
}
