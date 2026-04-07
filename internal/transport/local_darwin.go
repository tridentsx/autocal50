//go:build darwin

package transport

import "autocal50/internal/patternout"

func newPlatformOutput() patternout.Output { return patternout.NewCGOutput() }
