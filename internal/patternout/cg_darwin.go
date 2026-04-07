//go:build darwin

package patternout

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Metal -framework QuartzCore -framework CoreGraphics

#import <Cocoa/Cocoa.h>
#import <Metal/Metal.h>
#import <QuartzCore/CAMetalLayer.h>
#import <CoreGraphics/CoreGraphics.h>

// --- Display helpers ---

static int screenCount(void) {
	return (int)[[NSScreen screens] count];
}

static CGDirectDisplayID displayIDForScreen(int idx) {
	NSArray *screens = [NSScreen screens];
	if (idx < 0 || idx >= (int)[screens count]) return 0;
	NSScreen *screen = [screens objectAtIndex:idx];
	return [[[screen deviceDescription] objectForKey:@"NSScreenNumber"] unsignedIntValue];
}

static void screenInfo(int idx, int *w, int *h, double *hz) {
	CGDirectDisplayID did = displayIDForScreen(idx);
	if (!did) return;
	*w = (int)CGDisplayPixelsWide(did);
	*h = (int)CGDisplayPixelsHigh(did);
	CGDisplayModeRef mode = CGDisplayCopyDisplayMode(did);
	if (mode) {
		*hz = CGDisplayModeGetRefreshRate(mode);
		CGDisplayModeRelease(mode);
	}
}

// --- Metal output ---

typedef struct {
	void *window;       // NSWindow*
	void *metalLayer;   // CAMetalLayer*
	void *device;       // id<MTLDevice>
	void *cmdQueue;     // id<MTLCommandQueue>
	CGDirectDisplayID displayID;
	int width, height;
} MetalOutput;

static void ensureApp(void) {
	if (NSApp == nil) {
		[NSApplication sharedApplication];
		[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
	}
}

static MetalOutput metalOpen(int screenIdx, int use10bit) {
	ensureApp();
	MetalOutput mo = {0};

	CGDirectDisplayID did = displayIDForScreen(screenIdx);
	if (!did) return mo;

	// Capture the display for exclusive access.
	if (CGDisplayCapture(did) != kCGErrorSuccess) return mo;
	mo.displayID = did;

	mo.width = (int)CGDisplayPixelsWide(did);
	mo.height = (int)CGDisplayPixelsHigh(did);

	// Create Metal device.
	id<MTLDevice> device = MTLCreateSystemDefaultDevice();
	if (!device) { CGDisplayRelease(did); return mo; }
	mo.device = (void *)device;

	id<MTLCommandQueue> queue = [device newCommandQueue];
	mo.cmdQueue = (void *)queue;

	// Create fullscreen window on captured display.
	NSArray *screens = [NSScreen screens];
	NSScreen *screen = [screens objectAtIndex:screenIdx];
	NSRect frame = [screen frame];

	NSWindow *win = [[NSWindow alloc]
		initWithContentRect:frame
		styleMask:NSWindowStyleMaskBorderless
		backing:NSBackingStoreBuffered
		defer:NO
		screen:screen];
	[win setLevel:CGShieldingWindowLevel()];
	[win setBackgroundColor:[NSColor blackColor]];

	// Set up Metal layer.
	NSView *view = [[NSView alloc] initWithFrame:frame];
	[view setWantsLayer:YES];

	CAMetalLayer *layer = [CAMetalLayer layer];
	layer.device = device;
	layer.pixelFormat = use10bit ? MTLPixelFormatBGR10A2Unorm : MTLPixelFormatBGRA8Unorm;
	layer.framebufferOnly = NO;
	layer.drawableSize = CGSizeMake(mo.width, mo.height);
	layer.wantsExtendedDynamicRangeContent = YES;
	[view setLayer:layer];

	[win setContentView:view];
	[win makeKeyAndOrderFront:nil];
	[NSCursor hide];

	mo.window = (void *)win;
	mo.metalLayer = (void *)layer;
	return mo;
}

static int metalRender(MetalOutput *mo, const unsigned char *data, int w, int h, int stride) {
	CAMetalLayer *layer = (CAMetalLayer *)mo->metalLayer;
	id<MTLCommandQueue> queue = (id<MTLCommandQueue>)mo->cmdQueue;

	id<CAMetalDrawable> drawable = [layer nextDrawable];
	if (!drawable) return -1;

	id<MTLTexture> tex = [drawable texture];
	MTLRegion region = MTLRegionMake2D(0, 0, w, h);
	[tex replaceRegion:region mipmapLevel:0 withBytes:data bytesPerRow:stride];

	id<MTLCommandBuffer> cmdBuf = [queue commandBuffer];
	[cmdBuf presentDrawable:drawable];
	[cmdBuf commit];
	[cmdBuf waitUntilCompleted];
	return 0;
}

static void metalClose(MetalOutput *mo) {
	[NSCursor unhide];
	if (mo->window) {
		NSWindow *win = (NSWindow *)mo->window;
		[win close];
		mo->window = NULL;
	}
	if (mo->displayID) {
		CGDisplayRelease(mo->displayID);
		mo->displayID = 0;
	}
	mo->metalLayer = NULL;
	mo->device = NULL;
	mo->cmdQueue = NULL;
}

static int metalSetMode(CGDirectDisplayID did, int w, int h, double hz) {
	CFArrayRef allModes = CGDisplayCopyAllDisplayModes(did, NULL);
	if (!allModes) return -1;

	CGDisplayModeRef best = NULL;
	for (CFIndex i = 0; i < CFArrayGetCount(allModes); i++) {
		CGDisplayModeRef mode = (CGDisplayModeRef)CFArrayGetValueAtIndex(allModes, i);
		int mw = (int)CGDisplayModeGetWidth(mode);
		int mh = (int)CGDisplayModeGetHeight(mode);
		double mhz = CGDisplayModeGetRefreshRate(mode);
		if (mw == w && mh == h) {
			if (hz <= 0 || (int)mhz == (int)hz) {
				best = mode;
				break;
			}
		}
	}

	int result = -1;
	if (best) {
		CGDisplayConfigRef config;
		if (CGBeginDisplayConfiguration(&config) == kCGErrorSuccess) {
			CGConfigureDisplayWithDisplayMode(config, did, best, NULL);
			if (CGCompleteDisplayConfiguration(config, kCGConfigureForSession) == kCGErrorSuccess) {
				result = 0;
			}
		}
	}
	CFRelease(allModes);
	return result;
}

static void metalSetEDR(void *layerPtr, int enable) {
	CAMetalLayer *layer = (CAMetalLayer *)layerPtr;
	layer.wantsExtendedDynamicRangeContent = enable ? YES : NO;
	if (enable) {
		layer.colorspace = CGColorSpaceCreateWithName(kCGColorSpaceExtendedLinearDisplayP3);
	} else {
		layer.colorspace = CGColorSpaceCreateDeviceRGB();
	}
}
*/
import "C"
import (
	"autocal50/internal/pattern"
	"fmt"
	"unsafe"
)

// DarwinOutput renders patterns via Metal on a captured display.
type DarwinOutput struct {
	mo       C.MetalOutput
	buf      []byte
	width    int
	height   int
	stride   int
	bitDepth int
	modes    []Mode
	idx      int
}

func NewCGOutput() Output { return &DarwinOutput{idx: -1} }

func (d *DarwinOutput) Open(connector string) error {
	count := int(C.screenCount())
	d.idx = -1

	// Parse connector as screen index.
	var targetIdx int
	if _, err := fmt.Sscanf(connector, "%d", &targetIdx); err == nil {
		if targetIdx >= 0 && targetIdx < count {
			d.idx = targetIdx
		} else if targetIdx > 0 && targetIdx <= count {
			d.idx = targetIdx - 1
		}
	}
	// Fallback: last non-primary screen.
	if d.idx < 0 {
		for i := count - 1; i > 0; i-- {
			d.idx = i
			break
		}
	}
	if d.idx < 0 {
		if count > 0 {
			d.idx = 0
		} else {
			return fmt.Errorf("no screens found")
		}
	}

	// Enumerate modes.
	d.modes = nil
	var w, h C.int
	var hz C.double
	C.screenInfo(C.int(d.idx), &w, &h, &hz)
	d.modes = append(d.modes,
		Mode{Width: int(w), Height: int(h), RefreshHz: float64(hz), BitDepth: 8},
		Mode{Width: int(w), Height: int(h), RefreshHz: float64(hz), BitDepth: 10},
	)

	// Open Metal output with display capture.
	d.mo = C.metalOpen(C.int(d.idx), 0)
	if d.mo.window == nil {
		return fmt.Errorf("failed to open Metal output on screen %d", d.idx)
	}
	d.width = int(d.mo.width)
	d.height = int(d.mo.height)
	d.bitDepth = 8
	d.stride = d.width * 4
	d.buf = make([]byte, d.stride*d.height)

	return nil
}

func (d *DarwinOutput) Modes() []Mode { return d.modes }

func (d *DarwinOutput) SetMode(m Mode) error {
	if m.Width != d.width || m.Height != d.height {
		rc := C.metalSetMode(d.mo.displayID, C.int(m.Width), C.int(m.Height), C.double(m.RefreshHz))
		if rc != 0 {
			return fmt.Errorf("mode %dx%d@%.0f not available", m.Width, m.Height, m.RefreshHz)
		}
		d.width = m.Width
		d.height = m.Height
		d.stride = d.width * 4
		d.buf = make([]byte, d.stride*d.height)
	}
	d.bitDepth = m.BitDepth
	if d.bitDepth <= 0 {
		d.bitDepth = 8
	}
	return nil
}

func (d *DarwinOutput) Render(p pattern.Pattern) error {
	if d.buf == nil {
		return fmt.Errorf("not open")
	}

	if d.bitDepth == 10 {
		DrawRGBA10(d.buf, d.width, d.height, d.stride, p)
	} else {
		DrawRGBA(d.buf, d.width, d.height, d.stride, p)
		// Swizzle BGRX → BGRA (Metal BGRA8Unorm expects B,G,R,A).
		// DrawRGBA already writes B,G,R,0xFF — compatible with BGRA8.
	}

	rc := C.metalRender(&d.mo,
		(*C.uchar)(unsafe.Pointer(&d.buf[0])),
		C.int(d.width), C.int(d.height), C.int(d.stride))
	if rc != 0 {
		return fmt.Errorf("Metal render failed")
	}
	return nil
}

func (d *DarwinOutput) SetHDRMetadata(meta *HDRMetadata) error {
	if d.mo.metalLayer == nil {
		return fmt.Errorf("not open")
	}
	if meta != nil {
		C.metalSetEDR(d.mo.metalLayer, 1)
	} else {
		C.metalSetEDR(d.mo.metalLayer, 0)
	}
	// macOS doesn't expose HDMI InfoFrame metadata directly —
	// EDR + colorspace is how it signals HDR to the display.
	return nil
}

func (d *DarwinOutput) Close() error {
	if d.mo.window != nil {
		C.metalClose(&d.mo)
	}
	d.buf = nil
	return nil
}
