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
)

type gamepad_state struct {
	buttonFlags uint32
	lt          C.int
	rt          C.int
	lsX         C.int
	lsY         C.int
	rsX         C.int
	rsY         C.int
}

const (
	DPAD_UP         = 0x0001
	DPAD_DOWN       = 0x0002
	DPAD_LEFT       = 0x0004
	DPAD_RIGHT      = 0x0008
	START           = 0x0010
	GBACK           = 0x0020
	LEFT_STICK      = 0x0040
	RIGHT_STICK     = 0x0080
	LEFT_BUTTON     = 0x0100
	RIGHT_BUTTON    = 0x0200
	GHOME           = 0x0400
	A               = 0x1000
	B               = 0x2000
	X               = 0x4000
	Y               = 0x8000
	PADDLE1         = 0x010000
	PADDLE2         = 0x020000
	PADDLE3         = 0x040000
	PADDLE4         = 0x080000
	TOUCHPAD_BUTTON = 0x100000
	MISC_BUTTON     = 0x200000
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
	gamepad_state_old gamepad_state
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
	new := r.gamepad_state_old
	defer r.send(&new)
	switch index {
		case 0:
			new.buttonFlags = A
		case 1:
			new.buttonFlags = B
		case 2:
			new.buttonFlags = X
		case 3:
			new.buttonFlags = Y
		case 4:
			new.buttonFlags = LEFT_BUTTON
		case 5:
			new.buttonFlags = RIGHT_BUTTON
		case 8:
			new.buttonFlags = GBACK
		case 9:
			new.buttonFlags = START
		case 10:
			new.buttonFlags = LEFT_STICK
		case 11:
			new.buttonFlags = RIGHT_STICK
		case 12:
			new.buttonFlags = UP
		case 13:
			new.buttonFlags = DOWN
		case 14:
			new.buttonFlags = LEFT
		case 15:
			new.buttonFlags = RIGHT
		case 16:
			new.buttonFlags = MISC_BUTTON
	}
}

func (r *Xbox360Controller) pressAxis(index int64, value float64) {
	new := r.gamepad_state_old
	defer r.send(&new)
	switch index {
		case 0:
			new.lsX = C.int(value * 32767);
		case 1:
			new.lsY = -C.int(value * 32767);
		case 2:
			new.rsX = C.int(value * 32767);
		case 3:
			new.rsY = -C.int(value * 32767);
	}

}

func (r *Xbox360Controller) pressSlider(index int64, value float64) {
	new := r.gamepad_state_old
	defer r.send(&new)
	switch index {
		case 6:
			new.lt = C.int(value * 255);
		case 7:
			new.rt = C.int(value * 255);
	}
}

type Vibration struct {
	LargeMotor byte
	SmallMotor byte
}

func NewEmulator(onVibration func(vibration Vibration)) (*Emulator, error) {
	return &Emulator{}, nil
}

func (e *Emulator) Close() error {
	return nil
}

func (e *Emulator) CreateXbox360Controller() (*Xbox360Controller, error) {
	return &Xbox360Controller{
		gamepad_state_old: gamepad_state{},
	}, nil
}

func (controller *Xbox360Controller) send(gamepad_state *gamepad_state) {
	gamepad_state_old := &controller.gamepad_state_old
	bf := gamepad_state.buttonFlags ^ gamepad_state_old.buttonFlags
	bf_new := gamepad_state.buttonFlags

	if bf != 0 {
		// up pressed == -1, down pressed == 1, else 0
		if (DPAD_UP|DPAD_DOWN)&bf != 0 {
			// button_state := bf_new & DPAD_UP ? -1 : (bf_new & DPAD_DOWN ? 1 : 0);
			button_state := C.int(0)
			if bf_new&DPAD_UP != 0 {
				button_state = -1
			} else if bf_new&DPAD_DOWN != 0 {
				button_state = 1
			} else {
				button_state = 0
			}

			C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_HAT0Y, C.int(button_state))
		}

		if (DPAD_LEFT|DPAD_RIGHT)&bf != 0 {
			// int button_state = bf_new & DPAD_LEFT ? -1 : (bf_new & DPAD_RIGHT ? 1 : 0);
			button_state := C.int(0)
			if bf_new&DPAD_LEFT != 0 {
				button_state = -1
			} else if bf_new&DPAD_RIGHT != 0 {
				button_state = 1
			} else {
				button_state = 0
			}

			C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_HAT0X, button_state)
		}

		state := C.int(0)
		if START&bf != 0 {
			if bf_new&START != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_START, state)
		}
		if GBACK&bf != 0 {
			if bf_new&GBACK != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_SELECT, state)
		}
		if LEFT_STICK&bf != 0 {
			if bf_new&LEFT_STICK != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_THUMBL, state)
		}
		if RIGHT_STICK&bf != 0 {
			if bf_new&RIGHT_STICK != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_THUMBR, state)
		}
		if LEFT_BUTTON&bf != 0 {
			if bf_new&LEFT_BUTTON != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_TL, state)
		}
		if RIGHT_BUTTON&bf != 0 {
			if bf_new&RIGHT_BUTTON != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_TR, state)
		}
		if (GHOME|MISC_BUTTON)&bf != 0 {
			if bf_new&(GHOME|MISC_BUTTON) != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_MODE, state)
		}
		if A&bf != 0 {
			if bf_new&A != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_SOUTH, state)
		}
		if B&bf != 0 {
			if bf_new&B != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_EAST, state)
		}
		if X&bf != 0 {
			if bf_new&X != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_NORTH, state)
		}
		if Y&bf != 0 {
			if bf_new&Y != 0 {
				state = 1
			}
			C.libevdev_uinput_write_event(gamepad_input, C.EV_KEY, C.BTN_WEST, state)
		}
	}

	if gamepad_state_old.lt != gamepad_state.lt {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_Z, gamepad_state.lt)
	}

	if gamepad_state_old.rt != gamepad_state.rt {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_RZ, gamepad_state.rt)
	}

	if gamepad_state_old.lsX != gamepad_state.lsX {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_X, gamepad_state.lsX)
	}

	if gamepad_state_old.lsY != gamepad_state.lsY {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_Y, -gamepad_state.lsY)
	}

	if gamepad_state_old.rsX != gamepad_state.rsX {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_RX, gamepad_state.rsX)
	}

	if gamepad_state_old.rsY != gamepad_state.rsY {
		C.libevdev_uinput_write_event(gamepad_input, C.EV_ABS, C.ABS_RY, -gamepad_state.rsY)
	}

	controller.gamepad_state_old = *gamepad_state
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
