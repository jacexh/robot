package servo_test

import (
	"testing"

	"github.com/jacexh/robot/pkg/servo"
	"github.com/stretchr/testify/assert"
)

func TestPort(t *testing.T) {
	ser := &servo.Servo{
		BaudRate: 115200,
	}
	err := ser.Connect()
	assert.NoError(t, err)
}
