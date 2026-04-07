//go:build linux

package display

// ResetGammaRamp is a no-op on Linux when using DRM/KMS direct output,
// since the compositor is bypassed entirely.
func ResetGammaRamp(outputName string) error { return nil }

// RestoreGammaRamp is a no-op on Linux when using DRM/KMS direct output.
func RestoreGammaRamp(outputName string) error { return nil }
