//go:build linux

package patternout

import (
	"fmt"
	"unsafe"
)

// HDRMetadata describes HDR10 static metadata (CTA-861.3 / SMPTE ST 2086).
type HDRMetadata struct {
	// Mastering display primaries in 0.00002 units (CIE 1931 xy × 50000).
	Rx, Ry uint16 // Red primary
	Gx, Gy uint16 // Green primary
	Bx, By uint16 // Blue primary
	Wx, Wy uint16 // White point
	// Luminance in units of 0.0001 cd/m².
	MaxLuminance uint32
	MinLuminance uint32
	// Content light level.
	MaxCLL  uint16 // Maximum Content Light Level (nits)
	MaxFALL uint16 // Maximum Frame-Average Light Level (nits)
}

// BT2020HDR10 returns HDR10 metadata for BT.2020 primaries with the given peak luminance.
func BT2020HDR10(peakNits int) HDRMetadata {
	return HDRMetadata{
		Rx: 35400, Ry: 14600, // 0.708, 0.292
		Gx: 8500, Gy: 39850,  // 0.170, 0.797
		Bx: 6550, By: 2300,   // 0.131, 0.046
		Wx: 15635, Wy: 16450, // 0.3127, 0.3290
		MaxLuminance: uint32(peakNits) * 10000,
		MinLuminance: 500, // 0.05 cd/m²
		MaxCLL:       uint16(peakNits),
		MaxFALL:      uint16(peakNits / 4),
	}
}

// DRM property ioctl structs.

type drmModeGetProperty struct {
	valuesPtr    uint64
	enumBlobPtr  uint64
	propID       uint32
	flags        uint32
	name         [32]byte
	countValues  uint32
	countEnumBlobs uint32
}

type drmModeObjGetProperties struct {
	propsPtr     uint64
	propValuesPtr uint64
	countProps   uint32
	objID        uint32
	objType      uint32
}

type drmModeObjSetProperty struct {
	value   uint64
	propID  uint32
	objID   uint32
	objType uint32
}

type drmModeCreateBlob struct {
	data   uint64
	length uint32
	blobID uint32
}

type drmModeDestroyBlob struct {
	blobID uint32
}

// DRM HDR output metadata struct (matches kernel's hdr_output_metadata).
type hdrOutputMetadata struct {
	metadataType uint32
	// Type 1 (HDMI) metadata follows.
	eotf          uint8
	metadataType1 uint8
	// Display primaries (CIE 1931 xy × 50000).
	displayPrimariesX [3]uint16
	displayPrimariesY [3]uint16
	whitePointX       uint16
	whitePointY       uint16
	maxLuminance      uint16
	minLuminance      uint16
	maxCLL            uint16
	maxFALL           uint16
}

const (
	drmObjConnector uint32 = 0xC0C0C0C0 + 1 // DRM_MODE_OBJECT_CONNECTOR
	eotfST2084      uint8  = 2               // SMPTE ST 2084 (PQ)
)

var (
	ioctlObjGetProperties = iowr(0xB9, unsafe.Sizeof(drmModeObjGetProperties{}))
	ioctlObjSetProperty   = iowr(0xBA, unsafe.Sizeof(drmModeObjSetProperty{}))
	ioctlGetProperty      = iowr(0xAA, unsafe.Sizeof(drmModeGetProperty{}))
	ioctlCreateBlob       = iowr(0xBD, unsafe.Sizeof(drmModeCreateBlob{}))
	ioctlDestroyBlob      = iowr(0xBE, unsafe.Sizeof(drmModeDestroyBlob{}))
)

// SetHDRMetadata sets HDR10 static metadata on the connector via DRM properties.
func (d *DRMOutput) SetHDRMetadata(meta HDRMetadata) error {
	if d.fd < 0 {
		return fmt.Errorf("DRM device not open")
	}

	// Build the kernel metadata struct.
	hdr := hdrOutputMetadata{
		metadataType:  0, // HDMI
		eotf:          eotfST2084,
		metadataType1: 0,
		whitePointX:   meta.Wx,
		whitePointY:   meta.Wy,
		maxLuminance:  uint16(meta.MaxLuminance / 10000),
		minLuminance:  uint16(meta.MinLuminance),
		maxCLL:        meta.MaxCLL,
		maxFALL:       meta.MaxFALL,
	}
	// Primaries order: R=0, G=1, B=2.
	hdr.displayPrimariesX = [3]uint16{meta.Rx, meta.Gx, meta.Bx}
	hdr.displayPrimariesY = [3]uint16{meta.Ry, meta.Gy, meta.By}

	// Create a blob for the metadata.
	blob := drmModeCreateBlob{
		data:   uint64(uintptr(unsafe.Pointer(&hdr))),
		length: uint32(unsafe.Sizeof(hdr)),
	}
	if err := drmIoctl(d.fd, ioctlCreateBlob, unsafe.Pointer(&blob)); err != nil {
		return fmt.Errorf("create HDR blob: %w", err)
	}

	// Find the HDR_OUTPUT_METADATA property on the connector.
	propID, err := d.findConnectorProperty("HDR_OUTPUT_METADATA")
	if err != nil {
		return fmt.Errorf("HDR_OUTPUT_METADATA property not found: %w", err)
	}

	// Set the property.
	set := drmModeObjSetProperty{
		value:   uint64(blob.blobID),
		propID:  propID,
		objID:   d.connectorID,
		objType: drmObjConnector,
	}
	if err := drmIoctl(d.fd, ioctlObjSetProperty, unsafe.Pointer(&set)); err != nil {
		return fmt.Errorf("set HDR_OUTPUT_METADATA: %w", err)
	}

	return nil
}

// ClearHDRMetadata removes HDR metadata (sets blob ID to 0).
func (d *DRMOutput) ClearHDRMetadata() error {
	propID, err := d.findConnectorProperty("HDR_OUTPUT_METADATA")
	if err != nil {
		return nil // property doesn't exist, nothing to clear
	}
	set := drmModeObjSetProperty{
		value:   0,
		propID:  propID,
		objID:   d.connectorID,
		objType: drmObjConnector,
	}
	return drmIoctl(d.fd, ioctlObjSetProperty, unsafe.Pointer(&set))
}

// findConnectorProperty finds a named property on the connector.
func (d *DRMOutput) findConnectorProperty(name string) (uint32, error) {
	// Get property list for connector.
	var objProps drmModeObjGetProperties
	objProps.objID = d.connectorID
	objProps.objType = drmObjConnector
	if err := drmIoctl(d.fd, ioctlObjGetProperties, unsafe.Pointer(&objProps)); err != nil {
		return 0, err
	}
	if objProps.countProps == 0 {
		return 0, fmt.Errorf("no properties")
	}

	propIDs := make([]uint32, objProps.countProps)
	propValues := make([]uint64, objProps.countProps)
	objProps.propsPtr = uint64(uintptr(unsafe.Pointer(&propIDs[0])))
	objProps.propValuesPtr = uint64(uintptr(unsafe.Pointer(&propValues[0])))
	if err := drmIoctl(d.fd, ioctlObjGetProperties, unsafe.Pointer(&objProps)); err != nil {
		return 0, err
	}

	for _, pid := range propIDs {
		var prop drmModeGetProperty
		prop.propID = pid
		if err := drmIoctl(d.fd, ioctlGetProperty, unsafe.Pointer(&prop)); err != nil {
			continue
		}
		propName := cString(prop.name[:])
		if propName == name {
			return pid, nil
		}
	}
	return 0, fmt.Errorf("property %q not found", name)
}

func cString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
