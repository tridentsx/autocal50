//go:build darwin

package patternout

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics

#import <Cocoa/Cocoa.h>

// Ensure NSApplication is initialized.
static void ensureApp(void) {
	if (NSApp == nil) {
		[NSApplication sharedApplication];
		[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
	}
}

// PatternView draws from a pixel buffer.
@interface PatternView : NSView {
	@public
	unsigned char *pixelData;
	int pixWidth, pixHeight, pixStride;
}
@end

@implementation PatternView
- (void)drawRect:(NSRect)dirtyRect {
	if (!pixelData) return;
	NSBitmapImageRep *rep = [[NSBitmapImageRep alloc]
		initWithBitmapDataPlanes:&pixelData
		pixelsWide:pixWidth
		pixelsHigh:pixHeight
		bitsPerSample:8
		samplesPerPixel:4
		hasAlpha:YES
		isPlanar:NO
		colorSpaceName:NSDeviceRGBColorSpace
		bitmapFormat:0
		bytesPerRow:pixStride
		bitsPerPixel:32];
	NSImage *img = [[NSImage alloc] initWithSize:NSMakeSize(pixWidth, pixHeight)];
	[img addRepresentation:rep];
	[img drawInRect:self.bounds fromRect:NSZeroRect operation:NSCompositingOperationCopy fraction:1.0];
}
@end

typedef struct {
	void *window;
	void *view;
	int width;
	int height;
} PatternWindow;

// Create a fullscreen borderless window on the screen at index screenIdx.
static PatternWindow createPatternWindow(int screenIdx) {
	ensureApp();
	PatternWindow pw = {0};

	NSArray *screens = [NSScreen screens];
	if (screenIdx < 0 || screenIdx >= (int)[screens count]) {
		return pw;
	}
	NSScreen *screen = [screens objectAtIndex:screenIdx];
	NSRect frame = [screen frame];

	NSWindow *win = [[NSWindow alloc]
		initWithContentRect:frame
		styleMask:NSWindowStyleMaskBorderless
		backing:NSBackingStoreBuffered
		defer:NO
		screen:screen];
	[win setLevel:NSScreenSaverWindowLevel + 1];
	[win setBackgroundColor:[NSColor blackColor]];
	[win setCollectionBehavior:NSWindowCollectionBehaviorFullScreenPrimary |
		NSWindowCollectionBehaviorStationary];
	[NSCursor hide];

	PatternView *view = [[PatternView alloc] initWithFrame:frame];
	[win setContentView:view];
	[win makeKeyAndOrderFront:nil];

	pw.window = (__bridge_retained void *)win;
	pw.view = (__bridge_retained void *)view;
	pw.width = (int)frame.size.width;
	pw.height = (int)frame.size.height;
	return pw;
}

static void updatePatternView(void *viewPtr, unsigned char *data, int w, int h, int stride) {
	PatternView *view = (__bridge PatternView *)viewPtr;
	view->pixelData = data;
	view->pixWidth = w;
	view->pixHeight = h;
	view->pixStride = stride;
	dispatch_async(dispatch_get_main_queue(), ^{
		[view setNeedsDisplay:YES];
	});
}

static void destroyPatternWindow(void *winPtr, void *viewPtr) {
	[NSCursor unhide];
	if (winPtr) {
		NSWindow *win = (__bridge_transfer NSWindow *)winPtr;
		dispatch_async(dispatch_get_main_queue(), ^{
			[win close];
		});
	}
	if (viewPtr) {
		(void)(__bridge_transfer PatternView *)viewPtr;
	}
}

static int screenCount(void) {
	ensureApp();
	return (int)[[NSScreen screens] count];
}

static void screenInfo(int idx, int *w, int *h, double *hz) {
	NSArray *screens = [NSScreen screens];
	if (idx < 0 || idx >= (int)[screens count]) return;
	NSScreen *screen = [screens objectAtIndex:idx];
	NSRect frame = [screen frame];
	*w = (int)frame.size.width;
	*h = (int)frame.size.height;
	// Refresh rate from display mode.
	CGDirectDisplayID displayID = [[[screen deviceDescription]
		objectForKey:@"NSScreenNumber"] unsignedIntValue];
	CGDisplayModeRef mode = CGDisplayCopyDisplayMode(displayID);
	if (mode) {
		*hz = CGDisplayModeGetRefreshRate(mode);
		CGDisplayModeRelease(mode);
	}
}
*/
import "C"
import (
	"autocal50/internal/pattern"
	"fmt"
	"unsafe"
)

// DarwinOutput renders patterns via a fullscreen NSWindow on macOS.
type DarwinOutput struct {
	pw     C.PatternWindow
	buf    []byte
	width  int
	height int
	stride int
	modes  []Mode
	idx    int // screen index
}

func NewCGOutput() Output { return &DarwinOutput{idx: -1} }

func (d *DarwinOutput) Open(connector string) error {
	// Find the screen. On macOS, connector names aren't as standardized;
	// accept index ("1", "2") or match by resolution.
	count := int(C.screenCount())
	d.idx = -1

	// Try parsing as index (0-based or 1-based).
	var targetIdx int
	if _, err := fmt.Sscanf(connector, "%d", &targetIdx); err == nil {
		if targetIdx >= 0 && targetIdx < count {
			d.idx = targetIdx
		} else if targetIdx > 0 && targetIdx <= count {
			d.idx = targetIdx - 1 // 1-based
		}
	}

	// Fallback: use the last non-primary screen (index > 0).
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
	for i := 0; i < count; i++ {
		var w, h C.int
		var hz C.double
		C.screenInfo(C.int(i), &w, &h, &hz)
		if i == d.idx {
			d.modes = append(d.modes, Mode{
				Width: int(w), Height: int(h), RefreshHz: float64(hz),
			})
		}
	}

	// Create the fullscreen window.
	d.pw = C.createPatternWindow(C.int(d.idx))
	if d.pw.window == nil {
		return fmt.Errorf("failed to create pattern window on screen %d", d.idx)
	}
	d.width = int(d.pw.width)
	d.height = int(d.pw.height)
	d.stride = d.width * 4
	d.buf = make([]byte, d.stride*d.height)

	return nil
}

func (d *DarwinOutput) Modes() []Mode { return d.modes }

func (d *DarwinOutput) SetMode(m Mode) error {
	if m.Width == d.width && m.Height == d.height {
		return nil
	}
	return fmt.Errorf("mode change to %dx%d not yet supported — set display resolution in System Settings first", m.Width, m.Height)
}

func (d *DarwinOutput) Render(p pattern.Pattern) error {
	if d.buf == nil {
		return fmt.Errorf("no buffer — call Open first")
	}

	// DrawRGBA writes BGRX, but NSBitmapImageRep expects RGBA.
	// Draw into buffer then swizzle.
	DrawRGBA(d.buf, d.width, d.height, d.stride, p)

	// Swizzle BGRX → RGBA in-place.
	for i := 0; i < len(d.buf); i += 4 {
		d.buf[i], d.buf[i+2] = d.buf[i+2], d.buf[i] // swap B↔R
	}

	C.updatePatternView(d.pw.view,
		(*C.uchar)(unsafe.Pointer(&d.buf[0])),
		C.int(d.width), C.int(d.height), C.int(d.stride))

	return nil
}

func (d *DarwinOutput) Close() error {
	if d.pw.window != nil {
		C.destroyPatternWindow(d.pw.window, d.pw.view)
		d.pw.window = nil
		d.pw.view = nil
	}
	d.buf = nil
	return nil
}
