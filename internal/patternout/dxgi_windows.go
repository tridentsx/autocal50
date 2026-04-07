//go:build windows

package patternout

import (
	"autocal50/internal/pattern"
	"fmt"
	"syscall"
	"unsafe"
)

// WindowsOutput renders patterns via DXGI exclusive fullscreen swap chain.
type WindowsOutput struct {
	hwnd      uintptr
	factory   *comObj
	adapter   *comObj
	device    *comObj
	devCtx    *comObj
	swapChain *comObj
	output    *comObj
	width     int
	height    int
	stride    int
	bitDepth  int
	format    uint32
	modes     []Mode
	connector string
}

func NewDXGIOutput() Output { return &WindowsOutput{} }

func (w *WindowsOutput) Open(connector string) error {
	w.connector = connector

	// Create DXGI Factory2.
	var factory *comObj
	hr, _, _ := procCreateDXGIFactory2.Call(0, uintptr(unsafe.Pointer(&iidIDXGIFactory2)), uintptr(unsafe.Pointer(&factory)))
	if int32(hr) < 0 {
		return fmt.Errorf("CreateDXGIFactory2 failed: %#x", hr)
	}
	w.factory = factory

	// Find the target output by enumerating adapters → outputs.
	if err := w.findOutput(connector); err != nil {
		w.Close()
		return err
	}

	// Create D3D11 device.
	var device, devCtx uintptr
	hr, _, _ = procD3D11CreateDevice.Call(
		uintptr(unsafe.Pointer(w.adapter)), // pAdapter
		0,                                   // DriverType (0 = unknown, adapter provided)
		0,                                   // Software
		0,                                   // Flags
		0,                                   // pFeatureLevels
		0,                                   // FeatureLevels count
		7,                                   // SDK version
		uintptr(unsafe.Pointer(&device)),
		0, // pFeatureLevel out
		uintptr(unsafe.Pointer(&devCtx)),
	)
	if int32(hr) < 0 {
		w.Close()
		return fmt.Errorf("D3D11CreateDevice failed: %#x", hr)
	}
	w.device = (*comObj)(unsafe.Pointer(device))
	w.devCtx = (*comObj)(unsafe.Pointer(devCtx))

	// Create a minimal hidden window for the swap chain.
	if err := w.createWindow(); err != nil {
		w.Close()
		return err
	}

	return nil
}

func (w *WindowsOutput) Modes() []Mode { return w.modes }

func (w *WindowsOutput) SetMode(m Mode) error {
	w.width = m.Width
	w.height = m.Height
	w.bitDepth = m.BitDepth
	if w.bitDepth <= 0 {
		w.bitDepth = 8
	}

	// Pick DXGI format based on bit depth.
	switch w.bitDepth {
	case 10:
		w.format = dxgiFormatR10G10B10A2Unorm
	default:
		w.format = dxgiFormatB8G8R8A8Unorm
	}

	// Release old swap chain if resizing.
	if w.swapChain != nil {
		w.swapChain.call(vIDXGISwapChain_SetFullscreenState, 0, 0)
		w.swapChain.release()
		w.swapChain = nil
	}

	// Create swap chain.
	desc := dxgiSwapChainDesc1{
		Width:       uint32(w.width),
		Height:      uint32(w.height),
		Format:      w.format,
		SampleCount: 1,
		BufferUsage: dxgiUsageRenderTargetOutput,
		BufferCount: 2,
		SwapEffect:  dxgiSwapEffectFlipDiscard,
	}

	var sc *comObj
	_, err := w.factory.call(vIDXGIFactory2_CreateSwapChainForHwnd,
		uintptr(unsafe.Pointer(w.device)),
		w.hwnd,
		uintptr(unsafe.Pointer(&desc)),
		0, // pFullscreenDesc (nil = windowed initially)
		0, // pRestrictToOutput
		uintptr(unsafe.Pointer(&sc)),
	)
	if err != nil {
		return fmt.Errorf("CreateSwapChainForHwnd: %w", err)
	}
	w.swapChain = sc
	w.stride = w.width * 4

	// Go exclusive fullscreen on the target output.
	_, err = w.swapChain.call(vIDXGISwapChain_SetFullscreenState, 1, uintptr(unsafe.Pointer(w.output)))
	if err != nil {
		// Non-fatal — some drivers don't support exclusive, fall back to borderless.
	}

	// Resize buffers to match mode.
	w.swapChain.call(vIDXGISwapChain_ResizeBuffers, 0,
		uintptr(w.width), uintptr(w.height), uintptr(w.format), 0)

	// Set target mode (resolution + refresh).
	target := dxgiModeDesc{
		Width:      uint32(w.width),
		Height:     uint32(w.height),
		RefreshNum: uint32(m.RefreshHz * 1000),
		RefreshDen: 1000,
		Format:     w.format,
	}
	w.swapChain.call(vIDXGISwapChain_ResizeTarget, uintptr(unsafe.Pointer(&target)))

	return nil
}

func (w *WindowsOutput) Render(p pattern.Pattern) error {
	if w.swapChain == nil {
		return fmt.Errorf("no swap chain — call SetMode first")
	}

	// Get back buffer.
	var backBuf *comObj
	_, err := w.swapChain.call(vIDXGISwapChain_GetBuffer, 0,
		uintptr(unsafe.Pointer(&iidID3D11Texture2D)),
		uintptr(unsafe.Pointer(&backBuf)))
	if err != nil {
		return fmt.Errorf("GetBuffer: %w", err)
	}
	defer backBuf.release()

	// Map the back buffer for CPU write.
	var mapped d3d11MappedSubresource
	_, err = w.devCtx.call(vID3D11DeviceContext_Map,
		uintptr(unsafe.Pointer(backBuf)),
		0,                                          // subresource
		4,                                          // D3D11_MAP_WRITE_DISCARD
		0,                                          // flags
		uintptr(unsafe.Pointer(&mapped)),
	)
	if err != nil {
		return fmt.Errorf("Map: %w", err)
	}

	// Write pixels.
	bufSize := int(mapped.RowPitch) * w.height
	buf := unsafe.Slice((*byte)(unsafe.Pointer(mapped.pData)), bufSize)

	if w.bitDepth == 10 {
		DrawRGBA10(buf, w.width, w.height, int(mapped.RowPitch), p)
	} else {
		DrawRGBA(buf, w.width, w.height, int(mapped.RowPitch), p)
	}

	// Unmap.
	w.devCtx.call(vID3D11DeviceContext_Unmap, uintptr(unsafe.Pointer(backBuf)), 0)

	// Present.
	w.swapChain.call(vIDXGISwapChain_Present, 1, 0) // vsync

	return nil
}

func (w *WindowsOutput) SetHDRMetadata(meta *HDRMetadata) error {
	if w.swapChain == nil {
		return fmt.Errorf("no swap chain")
	}

	// Query IDXGISwapChain3 for SetColorSpace1.
	var sc3 *comObj
	if err := w.swapChain.queryInterface(&iidIDXGISwapChain3, &sc3); err == nil {
		defer sc3.release()
		if meta != nil {
			sc3.call(vIDXGISwapChain3_SetColorSpace1, uintptr(dxgiColorSpaceRGBFullG2084NoneP2020))
		} else {
			sc3.call(vIDXGISwapChain3_SetColorSpace1, uintptr(dxgiColorSpaceRGBFullG22NoneP709))
		}
	}

	// Query IDXGISwapChain4 for SetHDRMetaData.
	var sc4 *comObj
	if err := w.swapChain.queryInterface(&iidIDXGISwapChain4, &sc4); err != nil {
		return fmt.Errorf("IDXGISwapChain4 not available: %w", err)
	}
	defer sc4.release()

	if meta == nil {
		sc4.call(vIDXGISwapChain4_SetHDRMetaData, 0, 0, 0) // clear
		return nil
	}

	hdr := dxgiHDR10Metadata{
		RedPrimaryX:           meta.Rx,
		RedPrimaryY:           meta.Ry,
		GreenPrimaryX:         meta.Gx,
		GreenPrimaryY:         meta.Gy,
		BluePrimaryX:          meta.Bx,
		BluePrimaryY:          meta.By,
		WhitePointX:           meta.Wx,
		WhitePointY:           meta.Wy,
		MaxMasteringLuminance: meta.MaxLuminance,
		MinMasteringLuminance: meta.MinLuminance,
		MaxContentLightLevel:  meta.MaxCLL,
		MaxFrameAvgLightLevel: meta.MaxFALL,
	}
	_, err := sc4.call(vIDXGISwapChain4_SetHDRMetaData,
		uintptr(dxgiHDRMetadataTypeHDR10),
		unsafe.Sizeof(hdr),
		uintptr(unsafe.Pointer(&hdr)),
	)
	return err
}

func (w *WindowsOutput) Close() error {
	if w.swapChain != nil {
		w.swapChain.call(vIDXGISwapChain_SetFullscreenState, 0, 0)
		w.swapChain.release()
		w.swapChain = nil
	}
	if w.devCtx != nil {
		w.devCtx.release()
		w.devCtx = nil
	}
	if w.device != nil {
		w.device.release()
		w.device = nil
	}
	if w.output != nil {
		w.output.release()
		w.output = nil
	}
	if w.adapter != nil {
		w.adapter.release()
		w.adapter = nil
	}
	if w.factory != nil {
		w.factory.release()
		w.factory = nil
	}
	if w.hwnd != 0 {
		procDestroyWindow.Call(w.hwnd)
		w.hwnd = 0
	}
	procShowCursor.Call(1)
	return nil
}

// --- internal helpers ---

func (w *WindowsOutput) findOutput(connector string) error {
	for ai := uint32(0); ; ai++ {
		var adapter *comObj
		_, err := w.factory.call(vIDXGIFactory2_EnumAdapters, uintptr(ai), uintptr(unsafe.Pointer(&adapter)))
		if err != nil {
			break
		}

		for oi := uint32(0); ; oi++ {
			var output *comObj
			_, err := adapter.call(vIDXGIAdapter_EnumOutputs, uintptr(oi), uintptr(unsafe.Pointer(&output)))
			if err != nil {
				break
			}

			var desc dxgiOutputDesc
			output.call(vIDXGIOutput_GetDesc, uintptr(unsafe.Pointer(&desc)))
			name := syscall.UTF16ToString(desc.DeviceName[:])

			// Collect modes.
			mw := int(desc.DesktopCoords[2] - desc.DesktopCoords[0])
			mh := int(desc.DesktopCoords[3] - desc.DesktopCoords[1])

			if name == connector || connector == "" {
				w.adapter = adapter
				w.output = output
				w.modes = append(w.modes, Mode{Width: mw, Height: mh, RefreshHz: 60})
				// Also add common modes.
				if mw >= 3840 {
					w.modes = append(w.modes,
						Mode{Width: 3840, Height: 2160, RefreshHz: 60, BitDepth: 10},
						Mode{Width: 3840, Height: 2160, RefreshHz: 24, BitDepth: 10},
					)
				}
				return nil
			}
			output.release()
		}
		adapter.release()
	}
	return fmt.Errorf("output %q not found", connector)
}

var (
	procRegisterClassExW = user32w.NewProc("RegisterClassExW")
	procCreateWindowExW  = user32w.NewProc("CreateWindowExW")
	procDestroyWindow    = user32w.NewProc("DestroyWindow")
	procShowCursor       = user32w.NewProc("ShowCursor")
	procDefWindowProcW   = user32w.NewProc("DefWindowProcW")

	user32w   = syscall.NewLazyDLL("user32.dll")
	kernel32w = syscall.NewLazyDLL("kernel32.dll")
	procGetModuleHandleW = kernel32w.NewProc("GetModuleHandleW")
)

var dxgiClassRegistered bool

func (w *WindowsOutput) createWindow() error {
	if !dxgiClassRegistered {
		className, _ := syscall.UTF16PtrFromString("AutoCal50DXGI")
		wc := struct {
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
		}{
			lpfnWndProc:   procDefWindowProcW.Addr(),
			lpszClassName: uintptr(unsafe.Pointer(className)),
		}
		wc.cbSize = uint32(unsafe.Sizeof(wc))
		h, _, _ := procGetModuleHandleW.Call(0)
		wc.hInstance = h
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		dxgiClassRegistered = true
	}

	className, _ := syscall.UTF16PtrFromString("AutoCal50DXGI")
	title, _ := syscall.UTF16PtrFromString("")
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		0x80000000|0x10000000, // WS_POPUP | WS_VISIBLE
		0, 0, uintptr(w.width), uintptr(w.height),
		0, 0, 0, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowEx failed")
	}
	w.hwnd = hwnd
	procShowCursor.Call(0)
	return nil
}
