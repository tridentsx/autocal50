//go:build windows

package display

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	enumDisplayMonitors = user32.NewProc("EnumDisplayMonitorsW")
	getMonitorInfo      = user32.NewProc("GetMonitorInfoW")
	enumDisplayDevices  = user32.NewProc("EnumDisplayDevicesW")
	enumDisplaySettings = user32.NewProc("EnumDisplaySettingsW")
)

type rect struct{ left, top, right, bottom int32 }

type monitorInfoEx struct {
	cbSize    uint32
	rcMonitor rect
	rcWork    rect
	dwFlags   uint32
	szDevice  [32]uint16
}

type displayDevice struct {
	cb           uint32
	deviceName   [32]uint16
	deviceString [128]uint16
	stateFlags   uint32
	deviceID     [128]uint16
	deviceKey    [128]uint16
}

func ListOutputs() ([]Output, error) {
	var outputs []Output

	// Enumerate display adapters via EnumDisplayDevices — works even
	// when EnumDisplayMonitors has callback issues with CGO.
	var idx uint32
	for idx = 0; ; idx++ {
		var dd displayDevice
		dd.cb = uint32(unsafe.Sizeof(dd))
		r, _, _ := enumDisplayDevices.Call(0, uintptr(idx), uintptr(unsafe.Pointer(&dd)), 0)
		if r == 0 {
			break
		}

		active := dd.stateFlags&1 != 0 // DISPLAY_DEVICE_ATTACHED_TO_DESKTOP
		primary := dd.stateFlags&4 != 0 // DISPLAY_DEVICE_PRIMARY_DEVICE
		name := syscall.UTF16ToString(dd.deviceName[:])
		friendlyName := syscall.UTF16ToString(dd.deviceString[:])

		o := Output{
			Name:      name,
			Connected: active,
			Primary:   primary,
		}

		if active {
			// Get current resolution and refresh rate
			devName, _ := syscall.UTF16PtrFromString(name)
			var dm [256]byte // oversized buffer to avoid layout issues
			*(*uint16)(unsafe.Pointer(&dm[36])) = 256 // dmSize at offset 36
			r2, _, _ := enumDisplaySettings.Call(
				uintptr(unsafe.Pointer(devName)),
				uintptr(0xFFFFFFFF), // ENUM_CURRENT_SETTINGS
				uintptr(unsafe.Pointer(&dm[0])),
			)
			if r2 != 0 {
				o.Width = int(*(*uint32)(unsafe.Pointer(&dm[108])))  // dmPelsWidth
				o.Height = int(*(*uint32)(unsafe.Pointer(&dm[112]))) // dmPelsHeight
				o.RefreshHz = float64(*(*uint32)(unsafe.Pointer(&dm[120]))) // dmDisplayFrequency
				o.X = int(*(*int32)(unsafe.Pointer(&dm[44])))  // dmPosition.x
				o.Y = int(*(*int32)(unsafe.Pointer(&dm[48])))  // dmPosition.y
			}
		}

		// Use friendly name if available (e.g. "Intel UHD Graphics" → not great,
		// but the monitor name comes from EDID which we show separately)
		if friendlyName != "" && o.Name == "" {
			o.Name = friendlyName
		}

		outputs = append(outputs, o)
	}

	if len(outputs) == 0 {
		return nil, fmt.Errorf("no display devices found")
	}

	return outputs, nil
}
