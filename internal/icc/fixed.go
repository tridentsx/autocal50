// Package icc implements ICC profile reading, writing, and creation
// for display (projector) calibration. Implemented from the ICC.1:2022
// specification (https://www.color.org/specification/ICC.1-2022-05.pdf).
package icc

import (
	"encoding/binary"
	"io"
	"math"
)

// s15Fixed16 is a signed fixed-point number with 15 integer bits and 16 fractional bits.
type s15Fixed16 int32

func s15Fixed16FromFloat(f float64) s15Fixed16 {
	return s15Fixed16(math.Round(f * 65536))
}

func (s s15Fixed16) Float64() float64 {
	return float64(s) / 65536
}

// u16Fixed16 is an unsigned fixed-point number with 16 integer bits and 16 fractional bits.
type u16Fixed16 uint32

func u16Fixed16FromFloat(f float64) u16Fixed16 {
	return u16Fixed16(math.Round(f * 65536))
}

func (u u16Fixed16) Float64() float64 {
	return float64(u) / 65536
}

// u8Fixed8 is an unsigned fixed-point number with 8 integer bits and 8 fractional bits.
type u8Fixed8 uint16

func u8Fixed8FromFloat(f float64) u8Fixed8 {
	return u8Fixed8(math.Round(f * 256))
}

func (u u8Fixed8) Float64() float64 {
	return float64(u) / 256
}

// Signature is a 4-byte ICC signature.
type Signature [4]byte

func sig(s string) Signature {
	var v Signature
	copy(v[:], s)
	return v
}

func (s Signature) String() string { return string(s[:]) }

// Binary helpers.
var be = binary.BigEndian

func writeBytes(w io.Writer, data ...any) error {
	for _, d := range data {
		if err := binary.Write(w, be, d); err != nil {
			return err
		}
	}
	return nil
}

func pad4(n int) int {
	if m := n % 4; m != 0 {
		return 4 - m
	}
	return 0
}
