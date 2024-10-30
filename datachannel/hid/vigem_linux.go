package hid

/*
#include <libevdev/libevdev.h>
#include <libevdev/libevdev-uinput.h>
typedef struct input_absinfo absinfo;
#cgo pkg-config: libevdev
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

type gamepad_state struct {
	buttonStates map[int]C.int

	lt  C.int
	rt  C.int
	lsX C.int
	lsY C.int
	rsX C.int
	rsY C.int
}

const (
	DPAD_VERT = iota
	DPAD_HORI
	START
	GBACK
	LEFT_STICK
	RIGHT_STICK
	LEFT_BUTTON
	RIGHT_BUTTON
	GHOME
	A
	B
	X
	Y
	PADDLE1
	PADDLE2
	PADDLE3
	PADDLE4
	TOUCHPAD_BUTTON
	MISC_BUTTON

	MAX_BUTTON
)

var (
	gamepad_input *C.struct_libevdev_uinput
)

func _init_gamepad() error {
	gamepad_dev := x360()

	rv := C.libevdev_uinput_create_from_device(gamepad_dev, C.LIBEVDEV_UINPUT_OPEN_MANAGED, &gamepad_input)
	if rv > 0 || mouse_abs_input == nil {
		return errors.New("failed to create new abs device")
	}

	return nil
}
func init() {
	if err := _init_gamepad(); err != nil {
		fmt.Printf("failed to initialize hid for linux : %s\n", err.Error())
	}
}

type Emulator struct {
	onVibration func(vibration Vibration)
}

type Xbox360Controller struct {
	gamepad_state_old *gamepad_state
	channel           chan *gamepad_state
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

func (r *Xbox360Controller) pressButton(index int64, value bool) {
	new := *r.gamepad_state_old
	prev := new.buttonStates
	new.buttonStates = map[int]C.int{}
	for k, v := range prev {
		new.buttonStates[k] = v
	}
	defer func() { r.channel <- &new }()

	state := C.int(0)
	if value {
		state = 1
	}

	switch index {
	case 0:
		new.buttonStates[A] = state
	case 1:
		new.buttonStates[B] = state
	case 2:
		new.buttonStates[X] = state
	case 3:
		new.buttonStates[Y] = state
	case 4:
		new.buttonStates[LEFT_BUTTON] = state
	case 5:
		new.buttonStates[RIGHT_BUTTON] = state
	case 8:
		new.buttonStates[GBACK] = state
	case 9:
		new.buttonStates[START] = state
	case 10:
		new.buttonStates[LEFT_STICK] = state
	case 11:
		new.buttonStates[RIGHT_STICK] = state
	case 12:
		if value {
			state = -1
		}
		new.buttonStates[DPAD_VERT] = state
	case 13:
		if value {
			state = 1
		}
		new.buttonStates[DPAD_VERT] = state
	case 14:
		if value {
			state = -1
		}
		new.buttonStates[DPAD_HORI] = state
	case 15:
		if value {
			state = 1
		}
		new.buttonStates[DPAD_HORI] = state

	case 16:
		new.buttonStates[MISC_BUTTON] = state
	}

}

func (r *Xbox360Controller) pressAxis(index int64, value float64) {
	new := *r.gamepad_state_old
	defer func() { r.channel <- &new }()
	switch index {
	case 0:
		new.lsX = C.int(value * 32767)
	case 1:
		new.lsY = -C.int(value * 32767)
	case 2:
		new.rsX = C.int(value * 32767)
	case 3:
		new.rsY = -C.int(value * 32767)
	}

}

func (r *Xbox360Controller) pressSlider(index int64, value float64) {
	new := *r.gamepad_state_old
	defer func() { r.channel <- &new }()
	switch index {
	case 6:
		new.lt = C.int(value * 255)
	case 7:
		new.rt = C.int(value * 255)
	}
}

type Vibration struct {
	LargeMotor byte
	SmallMotor byte
}

func NewEmulator() (*Emulator, error) {
	return &Emulator{}, nil
}

func (e *Emulator) Close() error {
	return nil
}

func (e *Emulator) CreateXbox360Controller() (*Xbox360Controller, error) {
	ret := &Xbox360Controller{
		gamepad_state_old: &gamepad_state{
			buttonStates: map[int]C.int{},
		},
		channel: make(chan *gamepad_state, 16),
	}

	for i := 0; i < MAX_BUTTON; i++ {
		ret.gamepad_state_old.buttonStates[i] = 0
	}

	thread.SafeLoop(make(chan bool), 0, func() {
		ret.send(<-ret.channel)
	})
	return ret, nil
}

func (controller *Xbox360Controller) send(gamepad_state *gamepad_state) {
	fmt.Printf("old %v\n", controller.gamepad_state_old.buttonStates)
	fmt.Printf("new %v\n", gamepad_state.buttonStates)
	if gamepad_state.buttonStates[START] != controller.gamepad_state_old.buttonStates[START] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_START, gamepad_state.buttonStates[START])
	}
	if gamepad_state.buttonStates[GBACK] != controller.gamepad_state_old.buttonStates[GBACK] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_SELECT, gamepad_state.buttonStates[GBACK])
	}
	if gamepad_state.buttonStates[LEFT_STICK] != controller.gamepad_state_old.buttonStates[LEFT_STICK] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_THUMBL, gamepad_state.buttonStates[LEFT_STICK])
	}
	if gamepad_state.buttonStates[RIGHT_STICK] != controller.gamepad_state_old.buttonStates[RIGHT_STICK] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_THUMBR, gamepad_state.buttonStates[RIGHT_STICK])
	}
	if gamepad_state.buttonStates[LEFT_BUTTON] != controller.gamepad_state_old.buttonStates[LEFT_BUTTON] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_TL, gamepad_state.buttonStates[LEFT_BUTTON])
	}
	if gamepad_state.buttonStates[RIGHT_BUTTON] != controller.gamepad_state_old.buttonStates[RIGHT_BUTTON] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_TR, gamepad_state.buttonStates[RIGHT_BUTTON])
	}
	if gamepad_state.buttonStates[A] != controller.gamepad_state_old.buttonStates[A] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_SOUTH, gamepad_state.buttonStates[A])
	}
	if gamepad_state.buttonStates[B] != controller.gamepad_state_old.buttonStates[B] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_EAST, gamepad_state.buttonStates[B])
	}
	if gamepad_state.buttonStates[X] != controller.gamepad_state_old.buttonStates[X] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_NORTH, gamepad_state.buttonStates[X])
	}
	if gamepad_state.buttonStates[Y] != controller.gamepad_state_old.buttonStates[Y] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_WEST, gamepad_state.buttonStates[Y])
	}
	if gamepad_state.buttonStates[MISC_BUTTON] != controller.gamepad_state_old.buttonStates[MISC_BUTTON] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_MODE, gamepad_state.buttonStates[MISC_BUTTON])
	}

	if gamepad_state.buttonStates[DPAD_VERT] != controller.gamepad_state_old.buttonStates[DPAD_VERT] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_HAT0Y, gamepad_state.buttonStates[DPAD_VERT])
	}
	if gamepad_state.buttonStates[DPAD_HORI] != controller.gamepad_state_old.buttonStates[DPAD_HORI] {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_HAT0X, gamepad_state.buttonStates[DPAD_HORI])
	}
	if controller.gamepad_state_old.lt != gamepad_state.lt {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_Z, gamepad_state.lt)
	}
	if controller.gamepad_state_old.rt != gamepad_state.rt {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_RZ, gamepad_state.rt)
	}
	if controller.gamepad_state_old.lsX != gamepad_state.lsX {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_X, gamepad_state.lsX)
	}
	if controller.gamepad_state_old.lsY != gamepad_state.lsY {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_Y, -gamepad_state.lsY)
	}
	if controller.gamepad_state_old.rsX != gamepad_state.rsX {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_RX, gamepad_state.rsX)
	}
	if controller.gamepad_state_old.rsY != gamepad_state.rsY {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_RY, -gamepad_state.rsY)
	}

	controller.gamepad_state_old = gamepad_state
	C.libevdev_uinput_write_event(gamepad_input, C.EV_SYN, C.SYN_REPORT, 0)
}

func x360() evdev_t {
	dev := C.libevdev_new()

	stick := C.absinfo{
		0,
		-32768, 32767,
		16,
		128,
		0,
	}

	trigger := C.absinfo{
		0,
		0, 255,
		0,
		0,
		0,
	}

	dpad := C.absinfo{
		0,
		-1, 1,
		0,
		0,
		0,
	}

	C.libevdev_set_uniq(dev, C.CString("Sunshine Gamepad"))
	C.libevdev_set_id_product(dev, 0x28E)
	C.libevdev_set_id_vendor(dev, 0x45E)
	C.libevdev_set_id_bustype(dev, 0x3)
	C.libevdev_set_id_version(dev, 0x110)
	C.libevdev_set_name(dev, C.CString("Microsoft X-Box 360 pad"))

	C.libevdev_enable_event_type(dev, C.EV_KEY)
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_WEST, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_EAST, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_NORTH, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_SOUTH, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_THUMBL, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_THUMBR, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_TR, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_TL, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_SELECT, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_MODE, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_START, unsafe.Pointer(nil))

	C.libevdev_enable_event_type(dev, C.EV_ABS)
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_HAT0Y, unsafe.Pointer(&dpad))
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_HAT0X, unsafe.Pointer(&dpad))
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_Z, unsafe.Pointer(&trigger))
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_RZ, unsafe.Pointer(&trigger))
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_X, unsafe.Pointer(&stick))
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_RX, unsafe.Pointer(&stick))
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_Y, unsafe.Pointer(&stick))
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_RY, unsafe.Pointer(&stick))

	C.libevdev_enable_event_type(dev, C.EV_FF)
	C.libevdev_enable_event_code(dev, C.EV_FF, C.FF_RUMBLE, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_FF, C.FF_CONSTANT, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_FF, C.FF_PERIODIC, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_FF, C.FF_SINE, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_FF, C.FF_RAMP, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_FF, C.FF_GAIN, unsafe.Pointer(nil))
	_ = stick
	_ = dpad
	_ = trigger

	return dev
}
