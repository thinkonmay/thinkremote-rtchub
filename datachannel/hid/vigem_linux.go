package hid

import (
	"fmt"

)

type Emulator struct {
	handle      uintptr
	onVibration func(vibration Vibration)
}

type Xbox360Controller struct {
}
func (c *Xbox360Controller) Close() error {
	return nil
}

func (c *Xbox360Controller) Connect() error {
	return nil
}

func (c *Xbox360Controller) Disconnect() error {
	return nil
}

// func (c *Xbox360Controller) Send(report *Xbox360ControllerReport) error {
// 	return nil
// }

func (r *Xbox360Controller) pressButton(index int64, value bool) {
}

func (r *Xbox360Controller) pressAxis(index int64, value float64) {
}

func (r *Xbox360Controller) pressSlider(index int64, value float64) {
}



type Vibration struct {
	LargeMotor byte
	SmallMotor byte
}

func NewEmulator(onVibration func(vibration Vibration)) (*Emulator, error) {
	return nil, fmt.Errorf("linux not support")
}

func (e *Emulator) Close() error {
	return fmt.Errorf("linux not support")
}

func (e *Emulator) CreateXbox360Controller() (*Xbox360Controller, error) {
	return nil,fmt.Errorf("linux not supported")
}
