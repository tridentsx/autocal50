package icc

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
)

// EncodeTextDescTag encodes a string as textDescriptionType (v2, ICC.1:2001 §6.5.17).
// This is the v2 'desc' tag: ASCII count + string + Unicode count + string + ScriptCode.
func EncodeTextDescTag(s string) []byte {
	var buf bytes.Buffer
	writeBytes(&buf, TypeTextDesc, uint32(0)) // type sig + reserved

	// ASCII.
	ascii := []byte(s)
	writeBytes(&buf, uint32(len(ascii)+1)) // count includes null
	buf.Write(ascii)
	buf.WriteByte(0) // null terminator

	// Unicode (language code 0, count, UTF-16BE).
	runes := utf16.Encode([]rune(s))
	writeBytes(&buf, uint32(0))              // Unicode language code
	writeBytes(&buf, uint32(len(runes)+1))   // count includes null
	for _, r := range runes {
		writeBytes(&buf, r)
	}
	writeBytes(&buf, uint16(0)) // null terminator

	// ScriptCode (unused, 67 bytes of zeros: 2 code + 1 count + 64 data).
	buf.Write(make([]byte, 67))

	return buf.Bytes()
}

// EncodeMLUCTag encodes a string as multiLocalizedUnicodeType (v4, ICC.1:2022 §10.15).
// Stores a single en/US record.
func EncodeMLUCTag(s string) []byte {
	var buf bytes.Buffer
	writeBytes(&buf, TypeMLUC, uint32(0)) // type sig + reserved

	// Record count and record size.
	writeBytes(&buf, uint32(1), uint32(12)) // 1 record, 12 bytes each

	// Record: language 'en', country 'US', length, offset.
	runes := utf16.Encode([]rune(s))
	strBytes := len(runes) * 2
	dataOffset := uint32(16 + 12) // header(16) + 1 record(12)

	buf.Write([]byte("enUS"))
	writeBytes(&buf, uint32(strBytes), dataOffset)

	// String data (UTF-16BE, no null terminator per spec).
	for _, r := range runes {
		writeBytes(&buf, r)
	}

	return buf.Bytes()
}

// DecodeTextDescTag decodes a textDescriptionType tag, returning the ASCII portion.
func DecodeTextDescTag(data []byte) (string, error) {
	if len(data) < 12 {
		return "", errShort("desc")
	}
	count := be.Uint32(data[8:])
	if count == 0 {
		return "", nil
	}
	end := 12 + int(count)
	if end > len(data) {
		return "", errShort("desc")
	}
	s := data[12:end]
	// Strip null terminator.
	if len(s) > 0 && s[len(s)-1] == 0 {
		s = s[:len(s)-1]
	}
	return string(s), nil
}

// DecodeMLUCTag decodes a multiLocalizedUnicodeType tag, returning the first string.
func DecodeMLUCTag(data []byte) (string, error) {
	if len(data) < 16 {
		return "", errShort("mluc")
	}
	count := binary.BigEndian.Uint32(data[8:])
	if count == 0 {
		return "", nil
	}
	// First record at offset 16.
	if len(data) < 28 {
		return "", errShort("mluc")
	}
	// lang(2) + country(2) + length(4) + offset(4)
	strLen := binary.BigEndian.Uint32(data[20:])
	strOff := binary.BigEndian.Uint32(data[24:])
	end := strOff + strLen
	if int(end) > len(data) {
		return "", errShort("mluc")
	}
	// Decode UTF-16BE.
	raw := data[strOff:end]
	u16s := make([]uint16, len(raw)/2)
	for i := range u16s {
		u16s[i] = binary.BigEndian.Uint16(raw[i*2:])
	}
	return string(utf16.Decode(u16s)), nil
}
