package icc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Profile class signatures (ICC.1:2022 §7.2.5).
var (
	ClassDisplay = sig("mntr")
	ClassInput   = sig("scnr")
	ClassOutput  = sig("prtr")
	ClassLink    = sig("link")
)

// Color space signatures (ICC.1:2022 §7.2.6).
var (
	SpaceXYZ = sig("XYZ ")
	SpaceLab = sig("Lab ")
	SpaceRGB = sig("RGB ")
	SpaceGray = sig("GRAY")
)

// Platform signatures.
var (
	PlatformApple     = sig("APPL")
	PlatformMicrosoft = sig("MSFT")
	PlatformNone      = Signature{}
)

// Rendering intent (ICC.1:2022 §7.2.15).
type RenderingIntent uint32

const (
	IntentPerceptual RenderingIntent = 0
	IntentRelative   RenderingIntent = 1
	IntentSaturation RenderingIntent = 2
	IntentAbsolute   RenderingIntent = 3
)

// Header is the 128-byte ICC profile header (ICC.1:2022 §7.2).
type Header struct {
	Size            uint32
	PreferredCMM    Signature
	Version         uint32 // e.g. 0x04400000 for v4.4
	DeviceClass     Signature
	ColorSpace      Signature
	PCS             Signature
	Created         time.Time
	ProfileSig      Signature // always 'acsp'
	Platform        Signature
	Flags           uint32
	Manufacturer    Signature
	Model           uint32
	Attributes      uint64
	RenderingIntent RenderingIntent
	Illuminant      [3]s15Fixed16 // PCS illuminant (D50)
	Creator         Signature
	ProfileID       [16]byte
}

// D50 illuminant in XYZ (ICC spec mandates this for PCS).
var D50 = [3]s15Fixed16{
	s15Fixed16FromFloat(0.9642),
	s15Fixed16FromFloat(1.0),
	s15Fixed16FromFloat(0.8249),
}

func defaultHeader() Header {
	return Header{
		Version:         0x02400000, // v2.4 for max compatibility
		ProfileSig:      sig("acsp"),
		PCS:             SpaceXYZ,
		RenderingIntent: IntentPerceptual,
		Illuminant:      D50,
	}
}

// tagEntry is one entry in the tag table.
type tagEntry struct {
	Sig    Signature
	Offset uint32
	Size   uint32
}

// Profile is an ICC color profile.
type Profile struct {
	Header Header
	tags   map[Signature][]byte // raw tag data keyed by signature
	order  []Signature          // tag insertion order for deterministic output
}

// NewProfile creates an empty profile.
func NewProfile() *Profile {
	return &Profile{
		Header: defaultHeader(),
		tags:   make(map[Signature][]byte),
	}
}

// SetTag stores raw tag data.
func (p *Profile) SetTag(sig Signature, data []byte) {
	if _, exists := p.tags[sig]; !exists {
		p.order = append(p.order, sig)
	}
	p.tags[sig] = data
}

// GetTag returns raw tag data or nil.
func (p *Profile) GetTag(sig Signature) []byte {
	return p.tags[sig]
}

// HasTag checks if a tag exists.
func (p *Profile) HasTag(sig Signature) bool {
	_, ok := p.tags[sig]
	return ok
}

// Bytes serializes the profile to a byte slice.
func (p *Profile) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Write serializes the profile to w.
func (p *Profile) Write(w io.Writer) error {
	tagCount := len(p.tags)

	// Calculate offsets: header(128) + tag count(4) + tag entries(12 each).
	dataStart := 128 + 4 + tagCount*12
	// Pad data start to 4-byte boundary.
	dataStart += pad4(dataStart)

	// Build tag data and compute offsets.
	type tagOut struct {
		sig    Signature
		offset uint32
		size   uint32
		data   []byte
	}
	var tags []tagOut
	offset := uint32(dataStart)
	for _, s := range p.order {
		d := p.tags[s]
		tags = append(tags, tagOut{sig: s, offset: offset, size: uint32(len(d)), data: d})
		offset += uint32(len(d))
		offset += uint32(pad4(len(d)))
	}

	// Write header.
	h := p.Header
	h.Size = offset

	var hdr bytes.Buffer
	writeBytes(&hdr, h.Size, h.PreferredCMM, h.Version, h.DeviceClass,
		h.ColorSpace, h.PCS)
	// Date/time: 12 bytes (year, month, day, hour, minute, second as uint16).
	t := h.Created
	if t.IsZero() {
		t = time.Now().UTC()
	}
	writeBytes(&hdr, uint16(t.Year()), uint16(t.Month()), uint16(t.Day()),
		uint16(t.Hour()), uint16(t.Minute()), uint16(t.Second()))
	writeBytes(&hdr, h.ProfileSig, h.Platform, h.Flags, h.Manufacturer, h.Model,
		h.Attributes, h.RenderingIntent, h.Illuminant, h.Creator, h.ProfileID)
	// Pad header to 128 bytes.
	hdr.Write(make([]byte, 128-hdr.Len()))

	if _, err := w.Write(hdr.Bytes()); err != nil {
		return err
	}

	// Write tag table.
	if err := binary.Write(w, be, uint32(tagCount)); err != nil {
		return err
	}
	for _, t := range tags {
		if err := writeBytes(w, t.sig, t.offset, t.size); err != nil {
			return err
		}
	}
	// Pad to data start.
	padBytes := dataStart - (128 + 4 + tagCount*12)
	if padBytes > 0 {
		if _, err := w.Write(make([]byte, padBytes)); err != nil {
			return err
		}
	}

	// Write tag data.
	for _, t := range tags {
		if _, err := w.Write(t.data); err != nil {
			return err
		}
		if p := pad4(len(t.data)); p > 0 {
			if _, err := w.Write(make([]byte, p)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Parse reads an ICC profile from raw bytes.
func Parse(data []byte) (*Profile, error) {
	if len(data) < 132 {
		return nil, errors.New("icc: data too short for header + tag count")
	}
	r := bytes.NewReader(data)
	p := NewProfile()

	// Read header.
	h := &p.Header
	binary.Read(r, be, &h.Size)
	binary.Read(r, be, &h.PreferredCMM)
	binary.Read(r, be, &h.Version)
	binary.Read(r, be, &h.DeviceClass)
	binary.Read(r, be, &h.ColorSpace)
	binary.Read(r, be, &h.PCS)

	var year, month, day, hour, minute, second uint16
	binary.Read(r, be, &year)
	binary.Read(r, be, &month)
	binary.Read(r, be, &day)
	binary.Read(r, be, &hour)
	binary.Read(r, be, &minute)
	binary.Read(r, be, &second)
	h.Created = time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.UTC)

	binary.Read(r, be, &h.ProfileSig)
	binary.Read(r, be, &h.Platform)
	binary.Read(r, be, &h.Flags)
	binary.Read(r, be, &h.Manufacturer)
	binary.Read(r, be, &h.Model)
	binary.Read(r, be, &h.Attributes)
	binary.Read(r, be, &h.RenderingIntent)
	binary.Read(r, be, &h.Illuminant)
	binary.Read(r, be, &h.Creator)
	binary.Read(r, be, &h.ProfileID)

	// Skip to offset 128.
	r.Seek(128, io.SeekStart)

	if h.ProfileSig != sig("acsp") {
		return nil, fmt.Errorf("icc: bad signature %q", h.ProfileSig)
	}

	// Read tag table.
	var tagCount uint32
	binary.Read(r, be, &tagCount)
	if tagCount > 1000 {
		return nil, fmt.Errorf("icc: unreasonable tag count %d", tagCount)
	}

	entries := make([]tagEntry, tagCount)
	for i := range entries {
		binary.Read(r, be, &entries[i].Sig)
		binary.Read(r, be, &entries[i].Offset)
		binary.Read(r, be, &entries[i].Size)
	}

	// Extract tag data.
	for _, e := range entries {
		end := e.Offset + e.Size
		if int(end) > len(data) {
			return nil, fmt.Errorf("icc: tag %s overflows (offset=%d size=%d)", e.Sig, e.Offset, e.Size)
		}
		p.SetTag(e.Sig, data[e.Offset:end])
	}

	return p, nil
}
