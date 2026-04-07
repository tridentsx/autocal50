//go:build linux

package transport

import "autocal50/internal/patternout"

func newPlatformOutput() patternout.Output { return patternout.NewDRMOutput() }
