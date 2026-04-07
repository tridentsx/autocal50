//go:build windows

package transport

import "autocal50/internal/patternout"

func newPlatformOutput() patternout.Output { return patternout.NewDXGIOutput() }
