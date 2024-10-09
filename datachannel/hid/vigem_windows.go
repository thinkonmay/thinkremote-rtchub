package hid

/*
#include <stdint.h>

typedef struct
{
	uint16_t wButtons;
	uint8_t bLeftTrigger;
	uint8_t bRightTrigger;
	int16_t sThumbLX;
	int16_t sThumbLY;
	int16_t sThumbRX;
	int16_t sThumbRY;
} xusb_report;
*/
import "C"

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	VIGEM_ERROR_NONE                        = 0x20000000
	VIGEM_ERROR_BUS_NOT_FOUND               = 0xE0000001
	VIGEM_ERROR_NO_FREE_SLOT                = 0xE0000002
	VIGEM_ERROR_INVALID_TARGET              = 0xE0000003
	VIGEM_ERROR_REMOVAL_FAILED              = 0xE0000004
	VIGEM_ERROR_ALREADY_CONNECTED           = 0xE0000005
	VIGEM_ERROR_TARGET_UNINITIALIZED        = 0xE0000006
	VIGEM_ERROR_TARGET_NOT_PLUGGED_IN       = 0xE0000007
	VIGEM_ERROR_BUS_VERSION_MISMATCH        = 0xE0000008
	VIGEM_ERROR_BUS_ACCESS_FAILED           = 0xE0000009
	VIGEM_ERROR_CALLBACK_ALREADY_REGISTERED = 0xE0000010
	VIGEM_ERROR_CALLBACK_NOT_FOUND          = 0xE0000011
	VIGEM_ERROR_BUS_ALREADY_CONNECTED       = 0xE0000012
	VIGEM_ERROR_BUS_INVALID_HANDLE          = 0xE0000013
	VIGEM_ERROR_XUSB_USERINDEX_OUT_OF_RANGE = 0xE0000014

	VIGEM_ERROR_MAX = VIGEM_ERROR_XUSB_USERINDEX_OUT_OF_RANGE + 1
)

var (
	client = windows.NewLazyDLL("ViGEmClient.dll")

	procAlloc                            = client.NewProc("vigem_alloc")
	procFree                             = client.NewProc("vigem_free")
	procConnect                          = client.NewProc("vigem_connect")
	procDisconnect                       = client.NewProc("vigem_disconnect")
	procTargetAdd                        = client.NewProc("vigem_target_add")
	procTargetFree                       = client.NewProc("vigem_target_free")
	procTargetRemove                     = client.NewProc("vigem_target_remove")
	procTargetX360Alloc                  = client.NewProc("vigem_target_x360_alloc")
	procTargetX360RegisterNotification   = client.NewProc("vigem_target_x360_register_notification")
	procTargetX360UnregisterNotification = client.NewProc("vigem_target_x360_unregister_notification")
	procTargetX360Update                 = client.NewProc("vigem_target_x360_update")
)

type VigemError struct {
	code uint
}

func NewVigemError(rawCode uintptr) *VigemError {
	code := uint(rawCode)

	if code == VIGEM_ERROR_NONE {
		return nil
	}

	return &VigemError{code}
}

func (err *VigemError) Error() string {
	switch err.code {
	case VIGEM_ERROR_BUS_NOT_FOUND:
		return "bus not found"
	case VIGEM_ERROR_NO_FREE_SLOT:
		return "no free slot"
	case VIGEM_ERROR_INVALID_TARGET:
		return "invalid target"
	case VIGEM_ERROR_REMOVAL_FAILED:
		return "removal failed"
	case VIGEM_ERROR_ALREADY_CONNECTED:
		return "already connected"
	case VIGEM_ERROR_TARGET_UNINITIALIZED:
		return "target uninitialized"
	case VIGEM_ERROR_TARGET_NOT_PLUGGED_IN:
		return "target not plugged in"
	case VIGEM_ERROR_BUS_VERSION_MISMATCH:
		return "bus version mismatch"
	case VIGEM_ERROR_BUS_ACCESS_FAILED:
		return "bus access failed"
	case VIGEM_ERROR_CALLBACK_ALREADY_REGISTERED:
		return "callback already registered"
	case VIGEM_ERROR_CALLBACK_NOT_FOUND:
		return "callback not found"
	case VIGEM_ERROR_BUS_ALREADY_CONNECTED:
		return "bus already connected"
	case VIGEM_ERROR_BUS_INVALID_HANDLE:
		return "bus invalid handle"
	case VIGEM_ERROR_XUSB_USERINDEX_OUT_OF_RANGE:
		return "xusb userindex out of range"
	default:
		return "invalid code returned by ViGEm"
	}
}

type Emulator struct {
	handle      uintptr
	onVibration func(vibration Vibration)
}

type Vibration struct {
	LargeMotor byte
	SmallMotor byte
}

func NewEmulator(onVibration func(vibration Vibration)) (*Emulator, error) {
	handle, _, err := procAlloc.Call()

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil, err
	}

	libErr, _, err := procConnect.Call(handle)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil, err
	}
	if err := NewVigemError(libErr); err != nil {
		return nil, err
	}

	return &Emulator{handle, onVibration}, nil
}

func (e *Emulator) Close() error {
	procDisconnect.Call(e.handle)
	_, _, err := procFree.Call(e.handle)

	return err
}

func (e *Emulator) CreateXbox360Controller() (*Xbox360Controller, error) {
	handle, _, err := procTargetX360Alloc.Call()

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil, err
	}

	notificationHandler := func(client, target uintptr, largeMotor, smallMotor, ledNumber byte) uintptr {
		e.onVibration(Vibration{largeMotor, smallMotor})

		return 0
	}
	callback := windows.NewCallback(notificationHandler)

	return &Xbox360Controller{e, handle, false, callback, Xbox360ControllerReport{}}, nil
}

type x360NotificationHandler func(client, target uintptr, largeMotor, smallMotor, ledNumber byte) uintptr

type Xbox360Controller struct {
	emulator            *Emulator
	handle              uintptr
	connected           bool
	notificationHandler uintptr

	slider Xbox360ControllerReport
}

func (c *Xbox360Controller) Close() error {
	_, _, err := procTargetFree.Call(c.handle)

	return err
}

func (c *Xbox360Controller) Connect() error {
	libErr, _, err := procTargetAdd.Call(c.emulator.handle, c.handle)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	libErr, _, err = procTargetX360RegisterNotification.Call(c.emulator.handle, c.handle, c.notificationHandler)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	c.connected = true

	return nil
}

func (c *Xbox360Controller) Disconnect() error {
	libErr, _, err := procTargetX360UnregisterNotification.Call(c.handle)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	libErr, _, err = procTargetRemove.Call(c.emulator.handle, c.handle)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	c.connected = false

	return nil
}

func (c *Xbox360Controller) send(report *Xbox360ControllerReport) error {
	libErr, _, err := procTargetX360Update.Call(c.emulator.handle, c.handle, uintptr(unsafe.Pointer(&report.native)))

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	return nil
}

type Xbox360ControllerReport struct {
	native    C.xusb_report
	Capture   bool
	Assistant bool
}

// Bits that correspond to the Xbox 360 controller buttons.
const (
	Xbox360ControllerButtonUp            = 1 << 0
	Xbox360ControllerButtonDown          = 1 << 1
	Xbox360ControllerButtonLeft          = 1 << 2
	Xbox360ControllerButtonRight         = 1 << 3
	Xbox360ControllerButtonStart         = 1 << 4
	Xbox360ControllerButtonBack          = 1 << 5
	Xbox360ControllerButtonLeftThumb     = 1 << 6
	Xbox360ControllerButtonRightThumb    = 1 << 7
	Xbox360ControllerButtonLeftShoulder  = 1 << 8
	Xbox360ControllerButtonRightShoulder = 1 << 9
	Xbox360ControllerButtonGuide         = 1 << 10
	Xbox360ControllerButtonA             = 1 << 12
	Xbox360ControllerButtonB             = 1 << 13
	Xbox360ControllerButtonX             = 1 << 14
	Xbox360ControllerButtonY             = 1 << 15
)

func (r *Xbox360Controller) pressButton(index int64, value bool) {
	var button C.ushort = 0
	switch index {
	case 0:
		button = Xbox360ControllerButtonA
	case 1:
		button = Xbox360ControllerButtonB
	case 2:
		button = Xbox360ControllerButtonX
	case 3:
		button = Xbox360ControllerButtonY
	case 4:
		button = Xbox360ControllerButtonLeftShoulder
	case 5:
		button = Xbox360ControllerButtonRightShoulder
	case 8:
		button = Xbox360ControllerButtonBack
	case 9:
		button = Xbox360ControllerButtonStart
	case 10:
		button = Xbox360ControllerButtonLeftThumb
	case 11:
		button = Xbox360ControllerButtonRightThumb
	case 12:
		button = Xbox360ControllerButtonUp
	case 13:
		button = Xbox360ControllerButtonDown
	case 14:
		button = Xbox360ControllerButtonLeft
	case 15:
		button = Xbox360ControllerButtonRight
	case 16:
		button = Xbox360ControllerButtonGuide
	}

	if value {
		r.slider.native.wButtons |= C.ushort(button)
	} else {
		r.slider.native.wButtons ^= C.ushort(button)
	}

	r.send(&r.slider)
}

func (r *Xbox360Controller) pressAxis(index int64, value float64) {
	switch index {
	case 0:
		r.slider.native.sThumbLX = C.short(value * 32767)
	case 1:
		r.slider.native.sThumbLY = -C.short(value * 32767)
	case 2:
		r.slider.native.sThumbRX = C.short(value * 32767)
	case 3:
		r.slider.native.sThumbRY = -C.short(value * 32767)
	}

	r.send(&r.slider)
}

func (r *Xbox360Controller) pressSlider(index int64, value float64) {
	switch index {
	case 6:
		r.slider.native.bLeftTrigger = C.uchar(value * 255)
	case 7:
		r.slider.native.bRightTrigger = C.uchar(value * 255)
	}

	r.send(&r.slider)
}
