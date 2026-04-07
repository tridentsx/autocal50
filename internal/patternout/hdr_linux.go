//go:build linux

package patternout

import (
	"fmt"
	"unsafe"
)

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
	eotf          uint8
	metadataType1 uint8
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
	drmObjConnector uint32 = 0xC0C0C0C0 + 1
	eotfST2084      uint8  = 2
)

var (
	ioctlObjGetProperties = iowr(0xB9, unsafe.Sizeof(drmModeObjGetProperties{}))
	ioctlObjSetProperty   = iowr(0xBA, unsafe.Sizeof(drmModeObjSetProperty{}))
	ioctlGetProperty      = iowr(0xAA, unsafe.Sizeof(drmModeGetProperty{}))
	ioctlCreateBlob       = iowr(0xBD, unsafe.Sizeof(drmModeCreateBlob{}))
	ioctlDestroyBlob      = iowr(0xBE, unsafe.Sizeof(drmModeDestroyBlob{}))
)

// SetHDRMetadata sets or clears HDR10 static metadata on the connector.
func (d *DRMOutput) SetHDRMetadata(meta *HDRMetadata) error {
	if d.fd < 0 {
		return fmt.Errorf("DRM device not open")
	}

	propID, err := d.findConnectorProperty("HDR_OUTPUT_METADATA")
	if err != nil {
		return fmt.Errorf("HDR_OUTPUT_METADATA property not found: %w", err)
	}

	if meta == nil {
		set := drmModeObjSetProperty{
			value: 0, propID: propID,
			objID: d.connectorID, objType: drmObjConnector,
		}
		return drmIoctl(d.fd, ioctlObjSetProperty, unsafe.Pointer(&set))
	}

	hdr := hdrOutputMetadata{
		metadataType:  0,
		eotf:          eotfST2084,
		metadataType1: 0,
		whitePointX:   meta.Wx,
		whitePointY:   meta.Wy,
		maxLuminance:  uint16(meta.MaxLuminance / 10000),
		minLuminance:  uint16(meta.MinLuminance),
		maxCLL:        meta.MaxCLL,
		maxFALL:       meta.MaxFALL,
	}
	hdr.displayPrimariesX = [3]uint16{meta.Rx, meta.Gx, meta.Bx}
	hdr.displayPrimariesY = [3]uint16{meta.Ry, meta.Gy, meta.By}

	blob := drmModeCreateBlob{
		data:   uint64(uintptr(unsafe.Pointer(&hdr))),
		length: uint32(unsafe.Sizeof(hdr)),
	}
	if err := drmIoctl(d.fd, ioctlCreateBlob, unsafe.Pointer(&blob)); err != nil {
		return fmt.Errorf("create HDR blob: %w", err)
	}

	set := drmModeObjSetProperty{
		value: uint64(blob.blobID), propID: propID,
		objID: d.connectorID, objType: drmObjConnector,
	}
	if err := drmIoctl(d.fd, ioctlObjSetProperty, unsafe.Pointer(&set)); err != nil {
		return fmt.Errorf("set HDR_OUTPUT_METADATA: %w", err)
	}
	return nil
}

func (d *DRMOutput) findConnectorProperty(name string) (uint32, error) {
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
		if cString(prop.name[:]) == name {
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
