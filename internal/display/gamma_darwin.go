//go:build darwin

package display

/*
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>

static int resetGamma(CGDirectDisplayID displayID, float *savedR, float *savedG, float *savedB, uint32_t *count) {
    // Get current gamma.
    uint32_t cap = 256;
    if (CGGetDisplayTransferByTable(displayID, cap, savedR, savedG, savedB, count) != kCGErrorSuccess) {
        return -1;
    }
    // Set identity.
    float identity[256];
    for (int i = 0; i < 256; i++) {
        identity[i] = (float)i / 255.0f;
    }
    if (CGSetDisplayTransferByTable(displayID, 256, identity, identity, identity) != kCGErrorSuccess) {
        return -2;
    }
    return 0;
}

static int restoreGamma(CGDirectDisplayID displayID, float *r, float *g, float *b, uint32_t count) {
    if (CGSetDisplayTransferByTable(displayID, count, r, g, b) != kCGErrorSuccess) {
        return -1;
    }
    return 0;
}
*/
import "C"
import "fmt"

type savedGamma struct {
	r, g, b [256]C.float
	count   C.uint32_t
}

var savedGammas = map[string]*savedGamma{}

// ResetGammaRamp sets the display's gamma to identity (linear passthrough).
func ResetGammaRamp(outputName string) error {
	// macOS uses display IDs; for now use main display as fallback.
	displayID := C.CGMainDisplayID()

	var saved savedGamma
	rc := C.resetGamma(displayID, &saved.r[0], &saved.g[0], &saved.b[0], &saved.count)
	if rc != 0 {
		return fmt.Errorf("gamma reset failed (rc=%d) for %s", rc, outputName)
	}
	savedGammas[outputName] = &saved
	return nil
}

// RestoreGammaRamp restores the previously saved gamma ramp.
func RestoreGammaRamp(outputName string) error {
	saved := savedGammas[outputName]
	if saved == nil {
		return nil
	}
	displayID := C.CGMainDisplayID()
	rc := C.restoreGamma(displayID, &saved.r[0], &saved.g[0], &saved.b[0], saved.count)
	if rc != 0 {
		return fmt.Errorf("gamma restore failed for %s", outputName)
	}
	delete(savedGammas, outputName)
	return nil
}
