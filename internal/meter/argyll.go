package meter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ArgyllDriver implements the Driver interface using ArgyllCMS spotread.
type ArgyllDriver struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	cancel context.CancelFunc
}

func NewArgyllDriver() Driver {
	return &ArgyllDriver{}
}

func (d *ArgyllDriver) Name() string { return "argyll-spotread" }

func (d *ArgyllDriver) Connect(_ context.Context, cfg map[string]any) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	args := []string{"-e", "-x", "-T"}

	// Optional: instrument port selection
	if port, ok := cfg["port"].(string); ok && port != "" {
		args = append(args, "-c", port)
	}
	// Optional: CCMX correction file
	if ccmx, ok := cfg["ccmx"].(string); ok && ccmx != "" {
		args = append(args, "-X", ccmx)
	}
	// Optional: refresh mode override
	if refresh, ok := cfg["refresh"].(string); ok && refresh != "" {
		args = append(args, "-Y", refresh)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel

	d.cmd = exec.CommandContext(ctx, "spotread", args...)

	var err error
	d.stdin, err = d.cmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdoutPipe, err := d.cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("stdout pipe: %w", err)
	}
	d.stdout = bufio.NewReader(stdoutPipe)

	// Merge stderr into stdout so we can read prompts
	d.cmd.Stderr = d.cmd.Stdout

	if err := d.cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("start spotread: %w", err)
	}

	// Wait for the initial prompt (spotread prints calibration info then waits)
	// Read until we see the measurement prompt
	if err := d.waitForPrompt(5 * time.Second); err != nil {
		d.kill()
		return fmt.Errorf("spotread init: %w", err)
	}

	return nil
}

func (d *ArgyllDriver) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.kill()
}

func (d *ArgyllDriver) kill() error {
	if d.stdin != nil {
		// Try graceful quit
		d.stdin.Write([]byte("q\n"))
		d.stdin.Close()
		d.stdin = nil
	}
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}
	if d.cmd != nil {
		d.cmd.Wait()
		d.cmd = nil
	}
	return nil
}

// DarkCal triggers a dark calibration on the instrument.
// The sensor should be covered or pointed at a dark surface.
func (d *ArgyllDriver) DarkCal(_ context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stdin == nil {
		return fmt.Errorf("not connected")
	}
	// spotread: 'k' triggers calibration
	if _, err := d.stdin.Write([]byte("k")); err != nil {
		return fmt.Errorf("trigger dark cal: %w", err)
	}
	// Wait for calibration to complete (spotread prints prompts).
	if err := d.waitForPrompt(30 * time.Second); err != nil {
		return fmt.Errorf("dark cal: %w", err)
	}
	return nil
}

// Yxy regex: matches "Yxy: 13.456 0.3127 0.3290" or similar
var yxyRe = regexp.MustCompile(`Yxy:\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)`)

// CCT regex: matches "CCT = 6504K" or "Correlated Color Temp = 6504K"
var cctRe = regexp.MustCompile(`(?:CCT|Correlated Color Temp)\s*=\s*([\d.]+)\s*K`)

func (d *ArgyllDriver) Measure(_ context.Context) (Reading, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stdin == nil {
		return Reading{}, fmt.Errorf("not connected")
	}

	// Trigger a reading by sending a space (or enter)
	if _, err := d.stdin.Write([]byte(" ")); err != nil {
		return Reading{}, fmt.Errorf("trigger reading: %w", err)
	}

	// Read output lines until we find Yxy data
	var reading Reading
	reading.Timestamp = time.Now().UnixMilli()
	gotYxy := false

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		line, err := d.readLineTimeout(30 * time.Second)
		if err != nil {
			return Reading{}, fmt.Errorf("read: %w", err)
		}

		// Parse Yxy
		if m := yxyRe.FindStringSubmatch(line); m != nil {
			reading.Luminance, _ = strconv.ParseFloat(m[1], 64)
			reading.X, _ = strconv.ParseFloat(m[2], 64)
			reading.Y, _ = strconv.ParseFloat(m[3], 64)
			gotYxy = true
		}

		// Parse CCT
		if m := cctRe.FindStringSubmatch(line); m != nil {
			reading.CCT, _ = strconv.ParseFloat(m[1], 64)
		}

		// If we got Yxy and hit a prompt or empty line after data, we're done
		if gotYxy && (strings.Contains(line, "Press") || line == "" || reading.CCT > 0) {
			break
		}
	}

	if !gotYxy {
		return Reading{}, fmt.Errorf("no measurement data received")
	}

	return reading, nil
}

func (d *ArgyllDriver) waitForPrompt(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		line, err := d.readLineTimeout(timeout)
		if err != nil {
			return err
		}
		// spotread shows "Press" when ready for a reading
		if strings.Contains(strings.ToLower(line), "press") {
			return nil
		}
	}
	return fmt.Errorf("timeout waiting for prompt")
}

func (d *ArgyllDriver) readLineTimeout(timeout time.Duration) (string, error) {
	ch := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		line, err := d.stdout.ReadString('\n')
		if err != nil {
			errCh <- err
			return
		}
		ch <- strings.TrimSpace(line)
	}()

	select {
	case line := <-ch:
		return line, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("read timeout")
	}
}
