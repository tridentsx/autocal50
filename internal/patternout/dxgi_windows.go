//go:build windows

package patternout

import (
	"autocal50/internal/pattern"
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

var (
	user32w   = syscall.NewLazyDLL("user32.dll")
	gdi32w    = syscall.NewLazyDLL("gdi32.dll")
	kernel32w = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassEx     = user32w.NewProc("RegisterClassExW")
	procCreateWindowEx      = user32w.NewProc("CreateWindowExW")
	procDestroyWindow       = user32w.NewProc("DestroyWindow")
	procShowWindow          = user32w.NewProc("ShowWindow")
	procSetForegroundWindow = user32w.NewProc("SetForegroundWindow")
	procGetDC               = user32w.NewProc("GetDC")
	procReleaseDC           = user32w.NewProc("ReleaseDC")
	procDefWindowProc       = user32w.NewProc("DefWindowProcW")
	procPeekMessage         = user32w.NewProc("PeekMessageW")
	procDispatchMessage     = user32w.NewProc("DispatchMessageW")
	procEnumDisplayDevices  = user32w.NewProc("EnumDisplayDevicesW")
	procEnumDisplaySettings = user32w.NewProc("EnumDisplaySettingsW")
	procShowCursor          = user32w.NewProc("ShowCursor")

	procCreateDIBSection    = gdi32w.NewProc("CreateDIBSection")
	procCreateCompatibleDC  = gdi32w.NewProc("CreateCompatibleDC")
	procSelectObject        = gdi32w.NewProc("SelectObject")
	procBitBlt              = gdi32w.NewProc("BitBlt")
	procDeleteDC            = gdi32w.NewProc("DeleteDC")
	procDeleteObject        = gdi32w.NewProc("DeleteObject")

	procGetModuleHandle = kernel32w.NewProc("GetModuleHandleW")
)

// Win32 constants.
const (
	wsPopup       = 0x80000000
	wsVisible     = 0x10000000
	wsExTopmost   = 0x00000008
	wsExToolWindow = 0x00000080
	swShow        = 5
	srccopy       = 0x00CC0020
	biRGB         = 0
	dibRGBColors  = 0
	pmRemove      = 1
)

type bitmapInfoHeader struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type bitmapInfo struct {
	bmiHeader bitmapInfoHeader
}

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  uintptr
	lpszClassName uintptr
	hIconSm       uintptr
}

type point struct{ x, y int32 }
type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

// WindowsOutput renders patterns via a fullscreen Win32 popup window + GDI DIB.
type WindowsOutput struct {
	hwnd      uintptr
	hdc       uintptr
	memDC     uintptr
	hBitmap   uintptr
	pixels    unsafe.Pointer
	width     int
	height    int
	stride    int
	modes     []Mode
	connector string
	mu        sync.Mutex
}

var classRegistered bool

func NewDXGIOutput() Output { return &WindowsOutput{} }

func (w *WindowsOutput) Open(connector string) error {
	w.connector = connector

	// Enumerate modes for this display.
	w.modes = enumModes(connector)

	// Find monitor position.
	x, y, width, height, err := getMonitorRect(connector)
	if err != nil {
		return err
	}

	// Register window class once.
	if !classRegistered {
		if err := registerClass(); err != nil {
			return err
		}
		classRegistered = true
	}

	// Create fullscreen popup window on the target monitor.
	className, _ := syscall.UTF16PtrFromString("AutoCal50Pattern")
	title, _ := syscall.UTF16PtrFromString("")

	hwnd, _, _ := procCreateWindowEx.Call(
		uintptr(wsExTopmost|wsExToolWindow),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		uintptr(wsPopup|wsVisible),
		uintptr(x), uintptr(y), uintptr(width), uintptr(height),
		0, 0, getHInstance(), 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowEx failed")
	}
	w.hwnd = hwnd
	w.width = width
	w.height = height

	procShowWindow.Call(hwnd, swShow)
	procSetForegroundWindow.Call(hwnd)
	procShowCursor.Call(0) // hide cursor

	// Create DIB section for pixel buffer.
	hdc, _, _ := procGetDC.Call(hwnd)
	w.hdc = hdc

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	w.memDC = memDC

	bi := bitmapInfo{
		bmiHeader: bitmapInfoHeader{
			biSize:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
			biWidth:       int32(width),
			biHeight:      -int32(height), // top-down DIB
			biPlanes:      1,
			biBitCount:    32,
			biCompression: biRGB,
		},
	}

	var bits unsafe.Pointer
	hBitmap, _, _ := procCreateDIBSection.Call(
		memDC,
		uintptr(unsafe.Pointer(&bi)),
		dibRGBColors,
		uintptr(unsafe.Pointer(&bits)),
		0, 0,
	)
	if hBitmap == 0 {
		return fmt.Errorf("CreateDIBSection failed")
	}
	w.hBitmap = hBitmap
	w.pixels = bits
	w.stride = width * 4

	procSelectObject.Call(memDC, hBitmap)

	return nil
}

func (w *WindowsOutput) Modes() []Mode { return w.modes }

func (w *WindowsOutput) SetMode(m Mode) error {
	// On Windows, mode changes require ChangeDisplaySettingsEx.
	// For now, we just validate the mode exists and recreate the buffer if size changed.
	if m.Width == w.width && m.Height == w.height {
		return nil
	}
	// Would need to destroy and recreate window + DIB at new size.
	// For initial implementation, require the mode to match current display settings.
	return fmt.Errorf("mode change to %dx%d not yet supported — set display resolution in Windows Settings first", m.Width, m.Height)
}

func (w *WindowsOutput) Render(p pattern.Pattern) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.pixels == nil {
		return fmt.Errorf("no pixel buffer — call Open first")
	}

	// Build a Go slice over the DIB pixel memory.
	bufSize := w.stride * w.height
	buf := unsafe.Slice((*byte)(w.pixels), bufSize)

	// Draw pattern into the DIB. DIB pixel format is BGRX (same as DRM XRGB8888).
	DrawRGBA(buf, w.width, w.height, w.stride, p)

	// BitBlt from memory DC to window DC.
	procBitBlt.Call(
		w.hdc, 0, 0, uintptr(w.width), uintptr(w.height),
		w.memDC, 0, 0, srccopy,
	)

	// Pump messages to keep the window responsive.
	pumpMessages()

	return nil
}

func (w *WindowsOutput) Close() error {
	procShowCursor.Call(1) // restore cursor
	if w.hBitmap != 0 {
		procDeleteObject.Call(w.hBitmap)
		w.hBitmap = 0
	}
	if w.memDC != 0 {
		procDeleteDC.Call(w.memDC)
		w.memDC = 0
	}
	if w.hdc != 0 {
		procReleaseDC.Call(w.hwnd, w.hdc)
		w.hdc = 0
	}
	if w.hwnd != 0 {
		procDestroyWindow.Call(w.hwnd)
		w.hwnd = 0
	}
	w.pixels = nil
	return nil
}

// --- helpers ---

func registerClass() error {
	className, _ := syscall.UTF16PtrFromString("AutoCal50Pattern")
	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   procDefWindowProc.Addr(),
		hInstance:     getHInstance(),
		lpszClassName: uintptr(unsafe.Pointer(className)),
	}
	r, _, _ := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		return fmt.Errorf("RegisterClassEx failed")
	}
	return nil
}

func getHInstance() uintptr {
	h, _, _ := procGetModuleHandle.Call(0)
	return h
}

func pumpMessages() {
	var m msg
	for {
		r, _, _ := procPeekMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, pmRemove)
		if r == 0 {
			break
		}
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}

// getMonitorRect finds the position and size of the named display device.
func getMonitorRect(name string) (x, y, w, h int, err error) {
	devName, _ := syscall.UTF16PtrFromString(name)
	var dm [256]byte
	*(*uint16)(unsafe.Pointer(&dm[36])) = 256 // dmSize
	r, _, _ := procEnumDisplaySettings.Call(
		uintptr(unsafe.Pointer(devName)),
		uintptr(0xFFFFFFFF), // ENUM_CURRENT_SETTINGS
		uintptr(unsafe.Pointer(&dm[0])),
	)
	if r == 0 {
		return 0, 0, 0, 0, fmt.Errorf("EnumDisplaySettings failed for %s", name)
	}
	x = int(*(*int32)(unsafe.Pointer(&dm[44])))   // dmPosition.x
	y = int(*(*int32)(unsafe.Pointer(&dm[48])))   // dmPosition.y
	w = int(*(*uint32)(unsafe.Pointer(&dm[108])))  // dmPelsWidth
	h = int(*(*uint32)(unsafe.Pointer(&dm[112])))  // dmPelsHeight
	return
}

type winDisplayDevice struct {
	cb           uint32
	deviceName   [32]uint16
	deviceString [128]uint16
	stateFlags   uint32
	deviceID     [128]uint16
	deviceKey    [128]uint16
}

// enumModes lists available display modes for a device.
func enumModes(name string) []Mode {
	devName, _ := syscall.UTF16PtrFromString(name)
	seen := make(map[[3]int]bool)
	var modes []Mode

	for i := uint32(0); ; i++ {
		var dm [256]byte
		*(*uint16)(unsafe.Pointer(&dm[36])) = 256
		r, _, _ := procEnumDisplaySettings.Call(
			uintptr(unsafe.Pointer(devName)),
			uintptr(i),
			uintptr(unsafe.Pointer(&dm[0])),
		)
		if r == 0 {
			break
		}
		w := int(*(*uint32)(unsafe.Pointer(&dm[108])))
		h := int(*(*uint32)(unsafe.Pointer(&dm[112])))
		hz := int(*(*uint32)(unsafe.Pointer(&dm[120])))
		key := [3]int{w, h, hz}
		if seen[key] || w == 0 {
			continue
		}
		seen[key] = true
		modes = append(modes, Mode{Width: w, Height: h, RefreshHz: float64(hz)})
	}
	return modes
}
