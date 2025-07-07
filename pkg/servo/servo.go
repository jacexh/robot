package servo

import "go.bug.st/serial"

const (
	DeviceName = "/dev/ttyAMA0"
	BaudRate   = 115200
)

type Servo struct {
	BaudRate int
	Port     serial.Port
}

func (s *Servo) Connect() error {
	mode := &serial.Mode{BaudRate: s.BaudRate, DataBits: 8}
	port, err := serial.Open(DeviceName, mode)
	if err != nil {
		return err
	}
	s.Port = port
	return nil
}
