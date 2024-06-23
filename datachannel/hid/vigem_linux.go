package hid

/*
*/
import "C"
import (
	"fmt"
	"errors"

)

type gamepad_state struct {
   buttonFlags uint32
    lt uint8
    rt uint8
    lsX int16
    lsY int16
    rsX int16
    rsY int16
}

const (
	DPAD_UP = 0x0001
	DPAD_DOWN = 0x0002;
	DPAD_LEFT = 0x0004
	DPAD_RIGHT = 0x0008
	START = 0x0010
	BACK = 0x0020
	LEFT_STICK = 0x0040
	RIGHT_STICK = 0x0080
	LEFT_BUTTON = 0x0100
	RIGHT_BUTTON = 0x0200
	HOME = 0x0400
	A = 0x1000
	B = 0x2000
	X = 0x4000
	Y = 0x8000
	PADDLE1 = 0x010000
	PADDLE2 = 0x020000
	PADDLE3 = 0x040000
	PADDLE4 = 0x080000
	TOUCHPAD_BUTTON = 0x100000
	MISC_BUTTON = 0x200000
)

var (
	gamepad_input   *C.struct_libevdev_uinput
)


func _init_gamepad() error {
	gamepad_dev := x360()

	rv := C.libevdev_uinput_create_from_device(&gamepad_dev, C.LIBEVDEV_UINPUT_OPEN_MANAGED, &gamepad_input)
	if rv > 0 || mouse_abs_input == nil {
		return errors.New("failed to create new abs device")
	}

	return nil
}
func init() {
	if err := _init_gamepad();err != nil {
		fmt.Printf("failed to initialize hid for linux : %s\n", err.Error())
	}
}

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



func gamepadaction(gamepad_state gamepad_state) {
    bf := gamepad_state.buttonFlags ^ gamepad_state_old.buttonFlags;
    bf_new := gamepad_state.buttonFlags;

    if bf {
      // up pressed == -1, down pressed == 1, else 0
      if ((DPAD_UP | DPAD_DOWN) & bf) {
        button_state := bf_new & DPAD_UP ? -1 : (bf_new & DPAD_DOWN ? 1 : 0);
        libevdev_uinput_write_event(uinput.get(), EV_ABS, ABS_HAT0Y, button_state);
      }

      if ((DPAD_LEFT | DPAD_RIGHT) & bf) {
        int button_state = bf_new & DPAD_LEFT ? -1 : (bf_new & DPAD_RIGHT ? 1 : 0);

        libevdev_uinput_write_event(uinput.get(), EV_ABS, ABS_HAT0X, button_state);
      }

      if START & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_START, bf_new & START ? 1 : 0);
	  }
      if BACK & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_SELECT, bf_new & BACK ? 1 : 0);
	  }
      if LEFT_STICK & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_THUMBL, bf_new & LEFT_STICK ? 1 : 0);
	  }
      if RIGHT_STICK & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_THUMBR, bf_new & RIGHT_STICK ? 1 : 0);
	  }
      if LEFT_BUTTON & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_TL, bf_new & LEFT_BUTTON ? 1 : 0);
	  }
      if RIGHT_BUTTON & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_TR, bf_new & RIGHT_BUTTON ? 1 : 0);
	  }
      if (HOME | MISC_BUTTON) & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_MODE, bf_new & (HOME | MISC_BUTTON) ? 1 : 0);
	  }
      if A & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_SOUTH, bf_new & A ? 1 : 0);
	  }
      if B & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_EAST, bf_new & B ? 1 : 0);
	  }
      if X & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_NORTH, bf_new & X ? 1 : 0);
	  }
      if Y & bf {
		libevdev_uinput_write_event(uinput.get(), EV_KEY, BTN_WEST, bf_new & Y ? 1 : 0);
	  }
    }

    if (gamepad_state_old.lt != gamepad_state.lt) {
      libevdev_uinput_write_event(uinput.get(), EV_ABS, ABS_Z, gamepad_state.lt);
    }

    if (gamepad_state_old.rt != gamepad_state.rt) {
      libevdev_uinput_write_event(uinput.get(), EV_ABS, ABS_RZ, gamepad_state.rt);
    }

    if (gamepad_state_old.lsX != gamepad_state.lsX) {
      libevdev_uinput_write_event(uinput.get(), EV_ABS, ABS_X, gamepad_state.lsX);
    }

    if (gamepad_state_old.lsY != gamepad_state.lsY) {
      libevdev_uinput_write_event(uinput.get(), EV_ABS, ABS_Y, -gamepad_state.lsY);
    }

    if (gamepad_state_old.rsX != gamepad_state.rsX) {
      libevdev_uinput_write_event(uinput.get(), EV_ABS, ABS_RX, gamepad_state.rsX);
    }

    if (gamepad_state_old.rsY != gamepad_state.rsY) {
      libevdev_uinput_write_event(uinput.get(), EV_ABS, ABS_RY, -gamepad_state.rsY);
    }

    gamepad_state_old = gamepad_state;
    libevdev_uinput_write_event(uinput.get(), EV_SYN, SYN_REPORT, 0);
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