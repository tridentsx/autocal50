//go:build linux

package patternout

import (
	"autocal50/internal/pattern"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// --- DRM ioctl structs (from linux/drm.h and drm_mode.h) ---

type drmModeCardRes struct {
	fbIDPtr        uint64
	crtcIDPtr      uint64
	connectorIDPtr uint64
	encoderIDPtr   uint64
	countFBs       uint32
	countCRTCs     uint32
	countConnectors uint32
	countEncoders  uint32
	minWidth       uint32
	maxWidth       uint32
	minHeight      uint32
	maxHeight      uint32
}

type drmModeModeInfo struct {
	clock      uint32
	hdisplay   uint16
	hsyncStart uint16
	hsyncEnd   uint16
	htotal     uint16
	hskew      uint16
	vdisplay   uint16
	vsyncStart uint16
	vsyncEnd   uint16
	vtotal     uint16
	vscan      uint16
	vrefresh   uint32
	flags      uint32
	typ        uint32
	name       [32]byte
}

type drmModeGetConnector struct {
	encodersPtr   uint64
	modesPtr      uint64
	propsPtr      uint64
	propValuesPtr uint64
	countModes    uint32
	countProps    uint32
	countEncoders uint32
	encoderID     uint32
	connectorID   uint32
	connectorType uint32
	connectorTypeID uint32
	connection    uint32
	mmWidth       uint32
	mmHeight      uint32
	subpixel      uint32
	pad           uint32
}

type drmModeGetEncoder struct {
	encoderID   uint32
	encoderType uint32
	crtcID      uint32
	possibleCRTCs  uint32
	possibleClones uint32
}

type drmModeCRTC struct {
	setConnectorsPtr uint64
	countConnectors  uint32
	crtcID           uint32
	fbID             uint32
	x, y             uint32
	gammaSize        uint32
	modeValid        uint32
	mode             drmModeModeInfo
}

type drmModeCreateDumb struct {
	height uint32
	width  uint32
	bpp    uint32
	flags  uint32
	handle uint32
	pitch  uint32
	size   uint64
}

type drmModeMapDumb struct {
	handle uint32
	pad    uint32
	offset uint64
}

type drmModeDestroyDumb struct {
	handle uint32
}

type drmModeFBCmd struct {
	fbID   uint32
	width  uint32
	height uint32
	pitch  uint32
	bpp    uint32
	depth  uint32
	handle uint32
}

// --- ioctl number computation ---

func ioc(dir, typ, nr, size uintptr) uintptr {
	return dir<<30 | typ<<8 | nr | size<<16
}

func iowr(nr, size uintptr) uintptr { return ioc(3, 'd', nr, size) }

var (
	ioctlGetResources  = iowr(0xA0, unsafe.Sizeof(drmModeCardRes{}))
	ioctlGetCRTC       = iowr(0xA1, unsafe.Sizeof(drmModeCRTC{}))
	ioctlSetCRTC       = iowr(0xA2, unsafe.Sizeof(drmModeCRTC{}))
	ioctlGetEncoder    = iowr(0xA6, unsafe.Sizeof(drmModeGetEncoder{}))
	ioctlGetConnector  = iowr(0xA7, unsafe.Sizeof(drmModeGetConnector{}))
	ioctlAddFB         = iowr(0xAE, unsafe.Sizeof(drmModeFBCmd{}))
	ioctlRmFB          = iowr(0xAF, unsafe.Sizeof(uint32(0)))
	ioctlCreateDumb    = iowr(0xB2, unsafe.Sizeof(drmModeCreateDumb{}))
	ioctlMapDumb       = iowr(0xB3, unsafe.Sizeof(drmModeMapDumb{}))
	ioctlDestroyDumb   = iowr(0xB4, unsafe.Sizeof(drmModeDestroyDumb{}))
	ioctlDropMaster    = ioc(0, 'd', 0x1F, 0)
	ioctlSetMaster     = ioc(0, 'd', 0x1E, 0)
)

func drmIoctl(fd int, req uintptr, arg unsafe.Pointer) error {
	for {
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), req, uintptr(arg))
		if errno == 0 {
			return nil
		}
		if errno == syscall.EINTR || errno == syscall.EAGAIN {
			continue
		}
		return errno
	}
}

// --- DRM Output implementation ---

// DRMOutput renders patterns via DRM/KMS, bypassing the compositor.
type DRMOutput struct {
	fd          int
	connectorID uint32
	crtcID      uint32
	fbID        uint32
	dumbHandle  uint32
	mmap        []byte
	width       int
	height      int
	stride      int
	modes       []drmModeModeInfo
	savedCRTC   *drmModeCRTC
}

// NewDRMOutput creates a new DRM output renderer.
func NewDRMOutput() Output { return &DRMOutput{fd: -1} }

func (d *DRMOutput) Open(connector string) error {
	// Find the DRM device.
	devPath, err := findDRMDevice(connector)
	if err != nil {
		return err
	}

	fd, err := syscall.Open(devPath, syscall.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("open %s: %w", devPath, err)
	}
	d.fd = fd

	// Try to become DRM master.
	_ = drmIoctl(fd, ioctlSetMaster, nil)

	// Get resources.
	var res drmModeCardRes
	if err := drmIoctl(fd, ioctlGetResources, unsafe.Pointer(&res)); err != nil {
		d.Close()
		return fmt.Errorf("get resources: %w", err)
	}

	connIDs := make([]uint32, res.countConnectors)
	crtcIDs := make([]uint32, res.countCRTCs)
	res.connectorIDPtr = uint64(uintptr(unsafe.Pointer(&connIDs[0])))
	res.crtcIDPtr = uint64(uintptr(unsafe.Pointer(&crtcIDs[0])))
	if err := drmIoctl(fd, ioctlGetResources, unsafe.Pointer(&res)); err != nil {
		d.Close()
		return fmt.Errorf("get resources (fill): %w", err)
	}

	// Find the target connector.
	found := false
	for _, cid := range connIDs {
		conn, modes, err := d.getConnector(cid)
		if err != nil {
			continue
		}
		name := connectorTypeName(conn.connectorType, conn.connectorTypeID)
		if name != connector && !strings.EqualFold(name, connector) {
			continue
		}
		if conn.connection != 1 { // DRM_MODE_CONNECTED
			d.Close()
			return fmt.Errorf("connector %s is not connected", connector)
		}
		d.connectorID = cid
		d.modes = modes
		// Find CRTC via current encoder.
		if conn.encoderID != 0 {
			var enc drmModeGetEncoder
			enc.encoderID = conn.encoderID
			if err := drmIoctl(fd, ioctlGetEncoder, unsafe.Pointer(&enc)); err == nil {
				d.crtcID = enc.crtcID
			}
		}
		// Fallback: use first available CRTC.
		if d.crtcID == 0 && len(crtcIDs) > 0 {
			d.crtcID = crtcIDs[0]
		}
		found = true
		break
	}
	if !found {
		d.Close()
		return fmt.Errorf("connector %q not found", connector)
	}

	// Save current CRTC for restore on Close.
	d.savedCRTC = &drmModeCRTC{}
	d.savedCRTC.crtcID = d.crtcID
	_ = drmIoctl(fd, ioctlGetCRTC, unsafe.Pointer(d.savedCRTC))

	return nil
}

func (d *DRMOutput) Modes() []Mode {
	var out []Mode
	for _, m := range d.modes {
		out = append(out, Mode{
			Width:     int(m.hdisplay),
			Height:    int(m.vdisplay),
			RefreshHz: float64(m.vrefresh),
		})
	}
	return out
}

func (d *DRMOutput) SetMode(m Mode) error {
	// Find matching DRM mode.
	var mode *drmModeModeInfo
	for i := range d.modes {
		dm := &d.modes[i]
		if int(dm.hdisplay) == m.Width && int(dm.vdisplay) == m.Height {
			if m.RefreshHz == 0 || int(dm.vrefresh) == int(m.RefreshHz) {
				mode = dm
				break
			}
		}
	}
	if mode == nil {
		return fmt.Errorf("mode %dx%d@%.0f not found", m.Width, m.Height, m.RefreshHz)
	}

	// Destroy old buffer if exists.
	d.destroyBuffer()

	d.width = int(mode.hdisplay)
	d.height = int(mode.vdisplay)

	// Create dumb buffer.
	create := drmModeCreateDumb{
		width:  uint32(d.width),
		height: uint32(d.height),
		bpp:    32,
	}
	if err := drmIoctl(d.fd, ioctlCreateDumb, unsafe.Pointer(&create)); err != nil {
		return fmt.Errorf("create dumb buffer: %w", err)
	}
	d.dumbHandle = create.handle
	d.stride = int(create.pitch)

	// Add framebuffer.
	fb := drmModeFBCmd{
		width:  uint32(d.width),
		height: uint32(d.height),
		pitch:  create.pitch,
		bpp:    32,
		depth:  24,
		handle: create.handle,
	}
	if err := drmIoctl(d.fd, ioctlAddFB, unsafe.Pointer(&fb)); err != nil {
		return fmt.Errorf("add framebuffer: %w", err)
	}
	d.fbID = fb.fbID

	// Map the buffer.
	mapReq := drmModeMapDumb{handle: create.handle}
	if err := drmIoctl(d.fd, ioctlMapDumb, unsafe.Pointer(&mapReq)); err != nil {
		return fmt.Errorf("map dumb buffer: %w", err)
	}
	var err error
	d.mmap, err = syscall.Mmap(d.fd, int64(mapReq.offset), int(create.size),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("mmap: %w", err)
	}

	// Set CRTC.
	crtc := drmModeCRTC{
		crtcID:           d.crtcID,
		fbID:             d.fbID,
		setConnectorsPtr: uint64(uintptr(unsafe.Pointer(&d.connectorID))),
		countConnectors:  1,
		modeValid:        1,
		mode:             *mode,
	}
	if err := drmIoctl(d.fd, ioctlSetCRTC, unsafe.Pointer(&crtc)); err != nil {
		return fmt.Errorf("set CRTC: %w", err)
	}

	return nil
}

func (d *DRMOutput) Render(p pattern.Pattern) error {
	if d.mmap == nil {
		return fmt.Errorf("no buffer mapped — call SetMode first")
	}
	DrawRGBA(d.mmap, d.width, d.height, d.stride, p)
	return nil
}

func (d *DRMOutput) Close() error {
	// Restore saved CRTC.
	if d.savedCRTC != nil && d.fd >= 0 {
		_ = drmIoctl(d.fd, ioctlSetCRTC, unsafe.Pointer(d.savedCRTC))
	}
	d.destroyBuffer()
	if d.fd >= 0 {
		_ = drmIoctl(d.fd, ioctlDropMaster, nil)
		syscall.Close(d.fd)
		d.fd = -1
	}
	return nil
}

func (d *DRMOutput) destroyBuffer() {
	if d.mmap != nil {
		_ = syscall.Munmap(d.mmap)
		d.mmap = nil
	}
	if d.fbID != 0 {
		id := d.fbID
		_ = drmIoctl(d.fd, ioctlRmFB, unsafe.Pointer(&id))
		d.fbID = 0
	}
	if d.dumbHandle != 0 {
		destroy := drmModeDestroyDumb{handle: d.dumbHandle}
		_ = drmIoctl(d.fd, ioctlDestroyDumb, unsafe.Pointer(&destroy))
		d.dumbHandle = 0
	}
}

func (d *DRMOutput) getConnector(id uint32) (drmModeGetConnector, []drmModeModeInfo, error) {
	// First call: get counts.
	var conn drmModeGetConnector
	conn.connectorID = id
	if err := drmIoctl(d.fd, ioctlGetConnector, unsafe.Pointer(&conn)); err != nil {
		return conn, nil, err
	}
	if conn.countModes == 0 {
		return conn, nil, nil
	}
	// Second call: fill modes.
	modes := make([]drmModeModeInfo, conn.countModes)
	conn.modesPtr = uint64(uintptr(unsafe.Pointer(&modes[0])))
	if conn.countEncoders > 0 {
		encs := make([]uint32, conn.countEncoders)
		conn.encodersPtr = uint64(uintptr(unsafe.Pointer(&encs[0])))
	}
	if err := drmIoctl(d.fd, ioctlGetConnector, unsafe.Pointer(&conn)); err != nil {
		return conn, nil, err
	}
	return conn, modes, nil
}

// --- helpers ---

// connectorTypeName maps DRM connector type + ID to a name like "HDMI-A-1".
func connectorTypeName(typ, id uint32) string {
	names := map[uint32]string{
		1: "VGA", 2: "DVII", 3: "DVID", 4: "DVIA", 5: "Composite",
		6: "SVIDEO", 7: "LVDS", 8: "Component", 9: "9PinDIN",
		10: "DisplayPort", 11: "HDMI-A", 12: "HDMI-B", 13: "TV",
		14: "eDP", 15: "VIRTUAL", 16: "DSI", 17: "DPI",
		18: "WRITEBACK", 19: "SPI", 20: "USB",
	}
	n := names[typ]
	if n == "" {
		n = fmt.Sprintf("Unknown%d", typ)
	}
	return fmt.Sprintf("%s-%d", n, id)
}

// findDRMDevice finds the /dev/dri/cardN device that has the given connector.
func findDRMDevice(connector string) (string, error) {
	// Check which card has this connector via sysfs.
	matches, _ := filepath.Glob("/sys/class/drm/card*-*")
	for _, m := range matches {
		base := filepath.Base(m)
		parts := strings.SplitN(base, "-", 2)
		if len(parts) < 2 {
			continue
		}
		if parts[1] == connector || strings.EqualFold(parts[1], connector) {
			// Extract card number.
			card := parts[0] // "card0"
			return "/dev/dri/" + card, nil
		}
	}
	// Fallback to card0.
	if _, err := os.Stat("/dev/dri/card0"); err == nil {
		return "/dev/dri/card0", nil
	}
	return "", fmt.Errorf("no DRM device found for connector %q", connector)
}
