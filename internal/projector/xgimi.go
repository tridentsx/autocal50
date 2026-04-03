package projector

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// XGIMIDriver implements the Driver interface for XGIMI projectors
// using the RS232 command protocol (header 0x2A 0x2A).
type XGIMIDriver struct {
	mu     sync.Mutex
	port   io.ReadWriteCloser
	state  map[string]any
	opener SerialOpener
}

// SerialOpener abstracts serial port creation for testability.
type SerialOpener func(device string, baud int) (io.ReadWriteCloser, error)

func NewXGIMIDriver(opener SerialOpener) DriverFactory {
	return func() Driver { return &XGIMIDriver{opener: opener, state: make(map[string]any)} }
}

// --- protocol ---

func xgimiChecksum(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return sum
}

func xgimiFrame(payload []byte) []byte {
	frame := make([]byte, 0, 2+len(payload)+1)
	frame = append(frame, 0x2A, 0x2A)
	frame = append(frame, payload...)
	frame = append(frame, xgimiChecksum(payload))
	return frame
}

func (d *XGIMIDriver) send(payload []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.port == nil {
		return fmt.Errorf("not connected")
	}
	_, err := d.port.Write(xgimiFrame(payload))
	// Small delay between commands to avoid overwhelming the UART
	time.Sleep(50 * time.Millisecond)
	return err
}

// --- Driver interface ---

func (d *XGIMIDriver) Name() string { return "xgimi-rs232" }

func (d *XGIMIDriver) Connect(ctx context.Context, cfg map[string]any) error {
	device, _ := cfg["device"].(string)
	if device == "" {
		device = "/dev/ttyUSB0"
	}
	baud := 115200
	if b, ok := cfg["baud"].(float64); ok {
		baud = int(b)
	}
	port, err := d.opener(device, baud)
	if err != nil {
		return fmt.Errorf("open %s: %w", device, err)
	}
	d.port = port
	// Set defaults
	d.state = map[string]any{
		"input":        "HDMI1",
		"image_mode":   "Movie",
		"brightness":   float64(5),
		"screen":       true,
		"high_refresh": false,
	}
	return nil
}

func (d *XGIMIDriver) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.port != nil {
		err := d.port.Close()
		d.port = nil
		return err
	}
	return nil
}

func (d *XGIMIDriver) Capabilities(_ context.Context) (Capabilities, error) {
	return Capabilities{
		Controls: []Control{
			{ID: "input", Label: "Input Source", Type: "enum", Options: []string{"HDMI1", "HDMI2", "USB"}},
			{ID: "image_mode", Label: "Image Mode", Type: "enum", Options: []string{
				"Vivid", "Movie", "IMAX Enhanced", "Performance", "TV", "Sport", "Filmmaker",
			}},
			{ID: "brightness", Label: "Brightness", Type: "slider", Min: 1, Max: 10, Step: 1},
			{ID: "screen", Label: "Screen", Type: "toggle"},
			{ID: "high_refresh", Label: "High Refresh Rate", Type: "toggle"},
			// Key simulation controls
			{ID: "key_power", Label: "Power", Type: "toggle"},
			{ID: "key_auto_focus", Label: "Auto Focus", Type: "toggle"},
			{ID: "key_mute", Label: "Mute", Type: "toggle"},
		},
		SupportsModes: []string{"Vivid", "Movie", "IMAX Enhanced", "Performance", "TV", "Sport", "Filmmaker"},
	}, nil
}

// --- command tables ---

var xgimiInputCmds = map[string][]byte{
	"HDMI1": {0x02, 0x01, 0x01},
	"HDMI2": {0x02, 0x01, 0x02},
	"USB":   {0x02, 0x01, 0x14},
}

var xgimiImageModeCmds = map[string][]byte{
	"Vivid":         {0x02, 0x03, 0x00},
	"Movie":         {0x02, 0x03, 0x01},
	"IMAX Enhanced": {0x02, 0x03, 0x02},
	"Performance":   {0x02, 0x03, 0x05},
	"TV":            {0x02, 0x03, 0x07},
	"Sport":         {0x02, 0x03, 0x09},
	"Filmmaker":     {0x02, 0x03, 0x1B},
}

var xgimiKeyCmds = map[string][]byte{
	"key_power":      {0x02, 0x07, 0x00},
	"key_source":     {0x02, 0x07, 0x08},
	"key_up":         {0x02, 0x07, 0x09},
	"key_down":       {0x02, 0x07, 0x0A},
	"key_right":      {0x02, 0x07, 0x0B},
	"key_left":       {0x02, 0x07, 0x0C},
	"key_ok":         {0x02, 0x07, 0x0D},
	"key_back":       {0x02, 0x07, 0x0E},
	"key_setting":    {0x02, 0x07, 0x0F},
	"key_home":       {0x02, 0x07, 0x10},
	"key_vol_up":     {0x02, 0x07, 0x11},
	"key_vol_down":   {0x02, 0x07, 0x12},
	"key_auto_focus": {0x02, 0x07, 0x13},
	"key_man_focus":  {0x02, 0x07, 0x14},
	"key_mute":       {0x02, 0x07, 0x15},
}

func (d *XGIMIDriver) GetControl(_ context.Context, id string) (any, error) {
	v, ok := d.state[id]
	if !ok {
		return nil, fmt.Errorf("unknown control: %s", id)
	}
	return v, nil
}

func (d *XGIMIDriver) SetControl(_ context.Context, id string, value any) error {
	switch id {
	case "input":
		s, _ := value.(string)
		cmd, ok := xgimiInputCmds[s]
		if !ok {
			return fmt.Errorf("unknown input: %s", s)
		}
		if err := d.send(cmd); err != nil {
			return err
		}
		d.state[id] = s

	case "image_mode":
		s, _ := value.(string)
		cmd, ok := xgimiImageModeCmds[s]
		if !ok {
			return fmt.Errorf("unknown image mode: %s", s)
		}
		if err := d.send(cmd); err != nil {
			return err
		}
		d.state[id] = s

	case "brightness":
		var level int
		switch v := value.(type) {
		case float64:
			level = int(v)
		case int:
			level = v
		default:
			return fmt.Errorf("brightness must be a number")
		}
		if level < 1 || level > 10 {
			return fmt.Errorf("brightness must be 1-10")
		}
		if err := d.send([]byte{0x02, 0x05, byte(level)}); err != nil {
			return err
		}
		d.state[id] = float64(level)

	case "screen":
		on, _ := value.(bool)
		param := byte(0x00)
		if on {
			param = 0x01
		}
		if err := d.send([]byte{0x02, 0x0D, param}); err != nil {
			return err
		}
		d.state[id] = on

	case "high_refresh":
		on, _ := value.(bool)
		param := byte(0x00)
		if on {
			param = 0x01
		}
		if err := d.send([]byte{0x02, 0x0E, param}); err != nil {
			return err
		}
		d.state[id] = on

	default:
		// Key press commands (fire-and-forget)
		if cmd, ok := xgimiKeyCmds[id]; ok {
			return d.send(cmd)
		}
		return fmt.Errorf("unknown control: %s", id)
	}

	return nil
}

// Wakeup sends the standby wakeup command ("wakeup" in ASCII).
func (d *XGIMIDriver) Wakeup() error {
	return d.send([]byte{0x07, 0x09, 0x77, 0x61, 0x6B, 0x65, 0x75, 0x70})
}
