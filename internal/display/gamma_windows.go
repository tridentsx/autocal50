//go:build windows

package display

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	procSetDeviceGammaRamp = gdi32.NewProc("SetDeviceGammaRamp")
	procGetDeviceGammaRamp = gdi32.NewProc("GetDeviceGammaRamp")
	procCreateDC           = gdi32.NewProc("CreateDCW")
	procDeleteDC           = gdi32.NewProc("DeleteDC")
)

// gammaRamp is a 256×3 array of uint16 (R, G, B channels).
type gammaRamp [3][256]uint16

var savedRamps = map[string]*gammaRamp{}

// ResetGammaRamp sets the display's gamma ramp to identity (linear passthrough).
// Call before calibration measurements.
func ResetGammaRamp(outputName string) error {
	dc, err := createDC(outputName)
	if err != nil {
		return err
	}
	defer deleteDC(dc)

	// Save current ramp.
	var saved gammaRamp
	r, _, _ := procGetDeviceGammaRamp.Call(dc, uintptr(unsafe.Pointer(&saved)))
	if r != 0 {
		savedRamps[outputName] = &saved
	}

	// Set identity ramp.
	var identity gammaRamp
	for i := 0; i < 256; i++ {
		v := uint16(i << 8)
		identity[0][i] = v
		identity[1][i] = v
		identity[2][i] = v
	}
	r, _, _ = procSetDeviceGammaRamp.Call(dc, uintptr(unsafe.Pointer(&identity)))
	if r == 0 {
		return fmt.Errorf("SetDeviceGammaRamp failed for %s", outputName)
	}
	return nil
}

// RestoreGammaRamp restores the previously saved gamma ramp.
func RestoreGammaRamp(outputName string) error {
	saved := savedRamps[outputName]
	if saved == nil {
		return nil
	}
	dc, err := createDC(outputName)
	if err != nil {
		return err
	}
	defer deleteDC(dc)

	r, _, _ := procSetDeviceGammaRamp.Call(dc, uintptr(unsafe.Pointer(saved)))
	if r == 0 {
		return fmt.Errorf("SetDeviceGammaRamp restore failed for %s", outputName)
	}
	delete(savedRamps, outputName)
	return nil
}

func createDC(name string) (uintptr, error) {
	n, _ := syscall.UTF16PtrFromString(name)
	dc, _, _ := procCreateDC.Call(uintptr(unsafe.Pointer(n)), 0, 0, 0)
	if dc == 0 {
		return 0, fmt.Errorf("CreateDC failed for %s", name)
	}
	return dc, nil
}

func deleteDC(dc uintptr) {
	procDeleteDC.Call(dc)
}
