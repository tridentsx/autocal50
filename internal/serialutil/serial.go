package serialutil

import (
	"io"

	"go.bug.st/serial"
)

// Open opens a serial port with the given baud rate and 8N1 settings.
func Open(device string, baud int) (io.ReadWriteCloser, error) {
	return serial.Open(device, &serial.Mode{
		BaudRate: baud,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
}

// ListPorts returns available serial port names.
func ListPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	return ports, nil
}
