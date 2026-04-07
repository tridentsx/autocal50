package display

import "fmt"

// Mode is a display mode (resolution + refresh rate).
type Mode struct {
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	RefreshHz float64 `json:"refreshHz"`
	Preferred bool    `json:"preferred"`
	Current   bool    `json:"current"`
}

type Output struct {
	Name      string  `json:"name"`
	Connected bool    `json:"connected"`
	Primary   bool    `json:"primary"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	RefreshHz float64 `json:"refreshHz"`
	Modes     []Mode  `json:"modes,omitempty"`
}

type EDID struct {
	Manufacturer string   `json:"manufacturer"`
	ProductCode  uint16   `json:"productCode"`
	Serial       uint32   `json:"serial"`
	Year         int      `json:"year"`
	Week         int      `json:"week"`
	Version      string   `json:"version"`
	DisplayName  string   `json:"displayName"`
	SerialString string   `json:"serialString"`
	MaxHRes      int      `json:"maxHRes"`
	MaxVRes      int      `json:"maxVRes"`
	BitDepth     int      `json:"bitDepth"`
	DigitalInput bool     `json:"digitalInput"`
	HDR          bool     `json:"hdr"`
	BT2020       bool     `json:"bt2020"`
	P3           bool     `json:"p3"`
	RawHex       string   `json:"rawHex"`
}

// ParseEDID parses a 128+ byte EDID base block.
func ParseEDID(data []byte) *EDID {
	if len(data) < 128 {
		return nil
	}
	// Verify header
	header := []byte{0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00}
	for i, b := range header {
		if data[i] != b {
			return nil
		}
	}

	e := &EDID{}

	// Manufacturer ID (bytes 8-9): 3 letters packed in 5-bit codes
	m := uint16(data[8])<<8 | uint16(data[9])
	e.Manufacturer = string([]byte{
		byte((m>>10)&0x1F) + 'A' - 1,
		byte((m>>5)&0x1F) + 'A' - 1,
		byte(m&0x1F) + 'A' - 1,
	})

	e.ProductCode = uint16(data[10]) | uint16(data[11])<<8
	e.Serial = uint32(data[12]) | uint32(data[13])<<8 | uint32(data[14])<<16 | uint32(data[15])<<24
	e.Week = int(data[16])
	e.Year = int(data[17]) + 1990
	e.Version = fmt.Sprintf("%d.%d", data[18], data[19])

	// Video input (byte 20)
	e.DigitalInput = data[20]&0x80 != 0
	if e.DigitalInput {
		switch (data[20] >> 4) & 0x07 {
		case 1:
			e.BitDepth = 6
		case 2:
			e.BitDepth = 8
		case 3:
			e.BitDepth = 10
		case 4:
			e.BitDepth = 12
		case 5:
			e.BitDepth = 14
		case 6:
			e.BitDepth = 16
		}
	}

	// Max resolution from detailed timing (bytes 54-71)
	if data[54] != 0 || data[55] != 0 {
		e.MaxHRes = int(data[56]) | int(data[58]>>4)<<8
		e.MaxVRes = int(data[59]) | int(data[61]>>4)<<8
	}

	// Descriptor blocks (4 × 18 bytes starting at byte 54)
	for i := 0; i < 4; i++ {
		offset := 54 + i*18
		// Check if it's a display descriptor (first 2 bytes = 0)
		if data[offset] == 0 && data[offset+1] == 0 {
			tag := data[offset+3]
			text := trimEDIDString(data[offset+5 : offset+18])
			switch tag {
			case 0xFC: // Display name
				e.DisplayName = text
			case 0xFF: // Serial string
				e.SerialString = text
			}
		}
	}

	// Parse CTA-861 extension blocks for HDR/gamut info
	numExt := int(data[126])
	for ext := 0; ext < numExt && (ext+1)*128+128 <= len(data); ext++ {
		block := data[(ext+1)*128 : (ext+2)*128]
		if block[0] == 0x02 { // CTA-861 extension
			parseCTABlock(block, e)
		}
	}

	// Raw hex (first 128 bytes)
	e.RawHex = fmt.Sprintf("%X", data[:128])

	return e
}

func parseCTABlock(block []byte, e *EDID) {
	dtdStart := int(block[2])
	if dtdStart < 4 || dtdStart > 127 {
		return
	}
	offset := 4
	for offset < dtdStart-1 {
		header := block[offset]
		tag := (header >> 5) & 0x07
		length := int(header & 0x1F)
		if tag == 7 && offset+1 < dtdStart { // Extended tag
			if offset+2+length <= dtdStart {
				extTag := block[offset+1]
				switch extTag {
				case 6: // HDR Static Metadata
					e.HDR = true
				case 5: // Colorimetry
					if length >= 3 {
						flags := block[offset+2]
						e.BT2020 = flags&(1<<5|1<<6) != 0
						e.P3 = flags&(1<<7) != 0
					}
				}
			}
		}
		offset += 1 + length
	}
}

func trimEDIDString(b []byte) string {
	s := string(b)
	for i, c := range s {
		if c == '\n' || c == 0 {
			return s[:i]
		}
	}
	return s
}
