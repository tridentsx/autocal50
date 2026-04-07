//go:build windows

package patternout

import (
	"syscall"
	"unsafe"
)

// COM helper — call a method on a COM vtable.
type comObj struct {
	vtbl *[1024]uintptr
}

func (c *comObj) call(method int, args ...uintptr) (uintptr, error) {
	a := append([]uintptr{uintptr(unsafe.Pointer(c))}, args...)
	r, _, _ := syscall.SyscallN(c.vtbl[method], a...)
	if hr := int32(r); hr < 0 {
		return r, syscall.Errno(hr)
	}
	return r, nil
}

func (c *comObj) release() {
	if c != nil && c.vtbl != nil {
		syscall.SyscallN(c.vtbl[2], uintptr(unsafe.Pointer(c))) // IUnknown::Release
	}
}

func (c *comObj) queryInterface(iid *guid, out **comObj) error {
	r, _, _ := syscall.SyscallN(c.vtbl[0], // IUnknown::QueryInterface
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(out)),
	)
	if hr := int32(r); hr < 0 {
		return syscall.Errno(hr)
	}
	return nil
}

// GUID
type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// GUIDs for DXGI/D3D11 interfaces.
var (
	iidIDXGIFactory2 = guid{0x50c83a1c, 0xe072, 0x4c48, [8]byte{0x87, 0xb0, 0x36, 0x30, 0xfa, 0x36, 0xa6, 0xd0}}
	iidIDXGIDevice   = guid{0x54ec77fa, 0x1377, 0x44e6, [8]byte{0x8c, 0x32, 0x88, 0xfd, 0x5f, 0x44, 0xc8, 0x4c}}
	iidIDXGIAdapter  = guid{0x2411e7e1, 0x12ac, 0x4ccf, [8]byte{0xbd, 0x14, 0x97, 0x98, 0xe8, 0x53, 0x4d, 0xc0}}
	iidIDXGIOutput   = guid{0xae02eedb, 0xc735, 0x4690, [8]byte{0x8d, 0x52, 0x5a, 0x8d, 0xc2, 0x02, 0x13, 0xaa}}
	iidIDXGISwapChain3 = guid{0x94d99bdb, 0xf1f8, 0x4ab0, [8]byte{0xb2, 0x36, 0x7d, 0xa0, 0x17, 0x0e, 0xda, 0xb1}}
	iidIDXGISwapChain4 = guid{0x3d585d5a, 0xbd4a, 0x489e, [8]byte{0xb1, 0xf4, 0x3d, 0xbc, 0xb6, 0x45, 0x2f, 0xfb}}
	iidID3D11Texture2D = guid{0x6f15aaf2, 0xd208, 0x4e89, [8]byte{0x9a, 0xb4, 0x48, 0x95, 0x35, 0xd3, 0x4f, 0x9c}}
)

// DXGI/D3D11 constants.
const (
	dxgiFormatB8G8R8A8Unorm    = 87
	dxgiFormatR10G10B10A2Unorm = 24
	dxgiFormatR16G16B16A16Float = 10

	dxgiUsageRenderTargetOutput = 1 << 5

	dxgiSwapEffectFlipDiscard = 4

	dxgiColorSpaceRGBFullG22NoneP709  = 0  // sRGB
	dxgiColorSpaceRGBFullG2084NoneP2020 = 12 // HDR10 PQ BT.2020

	dxgiHDRMetadataTypeHDR10 = 1

	d3dDriverTypeHardware = 1
	d3d11CreateDeviceFlag  = 0 // no debug, no singlethreaded

	d3d11MapRead = 1

	// IDXGIFactory2 vtable offsets.
	vIDXGIFactory2_EnumAdapters          = 7  // IDXGIFactory::EnumAdapters
	vIDXGIFactory2_CreateSwapChainForHwnd = 15 // IDXGIFactory2::CreateSwapChainForHwnd

	// IDXGIAdapter vtable offsets.
	vIDXGIAdapter_EnumOutputs = 7

	// IDXGIOutput vtable offsets.
	vIDXGIOutput_GetDesc = 7

	// IDXGISwapChain vtable offsets.
	vIDXGISwapChain_Present           = 8
	vIDXGISwapChain_GetBuffer         = 9
	vIDXGISwapChain_SetFullscreenState = 10
	vIDXGISwapChain_ResizeTarget       = 14
	vIDXGISwapChain_GetContainingOutput = 15
	vIDXGISwapChain_ResizeBuffers      = 13

	// IDXGISwapChain3 vtable offsets (extends IDXGISwapChain2).
	vIDXGISwapChain3_SetColorSpace1 = 22

	// IDXGISwapChain4 vtable offsets (extends IDXGISwapChain3).
	vIDXGISwapChain4_SetHDRMetaData = 23

	// ID3D11DeviceContext vtable offsets.
	vID3D11DeviceContext_Map   = 14
	vID3D11DeviceContext_Unmap = 15

	// ID3D11Texture2D — we just need it as a resource for Map/Unmap.
)

// DXGI structs.

type dxgiSwapChainDesc1 struct {
	Width       uint32
	Height      uint32
	Format      uint32
	Stereo      int32
	SampleCount uint32
	SampleQual  uint32
	BufferUsage uint32
	BufferCount uint32
	Scaling     uint32
	SwapEffect  uint32
	AlphaMode   uint32
	Flags       uint32
}

type dxgiSwapChainFullscreenDesc struct {
	RefreshNum   uint32
	RefreshDen   uint32
	ScanlineMode uint32
	Scaling      uint32
	Windowed     int32
}

type dxgiOutputDesc struct {
	DeviceName     [32]uint16
	DesktopCoords  [4]int32 // left, top, right, bottom
	AttachedToDesk int32
	Rotation       uint32
	Monitor        uintptr
}

type dxgiModeDesc struct {
	Width       uint32
	Height      uint32
	RefreshNum  uint32
	RefreshDen  uint32
	Format      uint32
	ScanlineMode uint32
	Scaling     uint32
}

type d3d11MappedSubresource struct {
	pData      uintptr
	RowPitch   uint32
	DepthPitch uint32
}

type dxgiHDR10Metadata struct {
	RedPrimaryX   uint16
	RedPrimaryY   uint16
	GreenPrimaryX uint16
	GreenPrimaryY uint16
	BluePrimaryX  uint16
	BluePrimaryY  uint16
	WhitePointX   uint16
	WhitePointY   uint16
	MaxMasteringLuminance uint32
	MinMasteringLuminance uint32
	MaxContentLightLevel  uint16
	MaxFrameAvgLightLevel uint16
}

// DLL procs.
var (
	d3d11dll  = syscall.NewLazyDLL("d3d11.dll")
	dxgidll   = syscall.NewLazyDLL("dxgi.dll")

	procD3D11CreateDevice   = d3d11dll.NewProc("D3D11CreateDevice")
	procCreateDXGIFactory2  = dxgidll.NewProc("CreateDXGIFactory2")
)
