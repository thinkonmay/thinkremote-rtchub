package hid

/*
#include <libevdev/libevdev.h>
#include <libevdev/libevdev-uinput.h>
typedef struct libevdev _evdev;
typedef struct input_absinfo absinfo;

#cgo pkg-config: libevdev
#include <X11/Xutil.h>
#include <X11/keysym.h>
#include <X11/keysymdef.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"math"
	"unsafe"
)

type KeyCode struct {
	wincode C.uint
	linuxcode  C.uint
	scancode C.uint
}
const UNKNOWN = 0;
var linuxCodeMap = make(map[C.uint]C.uint)
var linuxScanCodeMap = make(map[C.uint]C.uint)

type touch_port_t struct  {
	offset_x 	float64
	offset_y 	float64
	height 		float64
	width 		float64
};

type mouse_position struct {
	mode 		string  // Either "abs" or "rel"
	offset_x    float64
	offset_y    float64
}

var mouse_pos = mouse_position {
	mode: "abs",
	offset_x: 100,
	offset_y: 100,
}

var target_touch_port = touch_port_t {
	offset_x: 120,
	offset_y: 400,
	height: 19200,
	width: 12000,
  };

var touch_port = touch_port_t {
	offset_x: 0,
	offset_y: 0,
	height: 19200,
	width: 12000,
}

var last_mouse_device_used any;
var last_mouse_device_buttons_down *uint8 = nil;
var mouse_abs_buttons_down uint8 = 0;
var mouse_rel_buttons_down uint8 = 0;


var (
	keycodes = []KeyCode{
		{0x01 /*LBUTTON */, LINUX_LBUTTON, 90001},
		{0x02 /*RBUTTON */, LINUX_RBUTTON, 90002},
		// {0x03 /*CANCEL */, CANCEL, UNKNOWN},
		{0x04 /*MBUTTON */, LINUX_MBUTTON, 90003},
		// {0x05 /*XBUTTON1 */, LINUX_XBUTTON1, UNKNOWN},
		// {0x06 /*XBUTTON2 */, LINUX_XBUTTON2, UNKNOWN},
		{0x08 /* VKEY_BACK */, LINUX_BACK, 0x7002A},
		{0x09 /* VKEY_TAB */, LINUX_TAB, 0x7002B},
		{0x0C /* VKEY_CLEAR */, LINUX_CLEAR, UNKNOWN},
		{0x0D /* VKEY_RETURN */, LINUX_RETURN, 0x70028},
		{0x10 /* VKEY_SHIFT */, LINUX_SHIFT, 0x700E1},
		{0x11 /* VKEY_CONTROL */, LINUX_CONTROL, 0x700E0},
		{0x12 /* VKEY_MENU */, LINUX_MENU, UNKNOWN},
		{0x13 /* VKEY_PAUSE */, LINUX_PAUSE, UNKNOWN},
		{0x14 /* VKEY_CAPITAL */, LINUX_CAPITAL, 0x70039},
		{0x15 /* VKEY_KANA */, LINUX_KANA, UNKNOWN},
		{0x16 /* VKEY_HANGUL */, LINUX_HANGUL, UNKNOWN},
		{0x17 /* VKEY_JUNJA */, LINUX_HANJA, UNKNOWN},
		{0x19 /* VKEY_KANJI */, LINUX_KANJI, UNKNOWN},
		{0x1B /* VKEY_ESCAPE */, LINUX_ESCAPE, 0x70029},
		{0x20 /* VKEY_SPACE */, LINUX_SPACE, 0x7002C},
		{0x21 /* VKEY_PRIOR */, LINUX_PRIOR, 0x7004B},
		{0x22 /* VKEY_NEXT */, LINUX_NEXT, 0x7004E},
		{0x23 /* VKEY_END */, LINUX_END, 0x7004D},
		{0x24 /* VKEY_HOME */, LINUX_HOME, 0x7004A},
		{0x25 /* VKEY_LEFT */, LINUX_LEFT, 0x70050},
		{0x26 /* VKEY_UP */, LINUX_UP, 0x70052},
		{0x27 /* VKEY_RIGHT */, LINUX_RIGHT, 0x7004F},
		{0x28 /* VKEY_DOWN */, LINUX_DOWN, 0x70051},
		{0x29 /* VKEY_SELECT */, LINUX_SELECT, UNKNOWN},
		{0x2A /* VKEY_PRINT */, LINUX_PRINT, UNKNOWN},
		{0x2C /* VKEY_SNAPSHOT */, LINUX_SNAPSHOT, 0x70046},
		{0x2D /* VKEY_INSERT */, LINUX_INSERT, 0x70049},
		{0x2E /* VKEY_DELETE */, LINUX_DELETE, 0x7004C},
		{0x2F /* VKEY_HELP */, LINUX_HELP, UNKNOWN},
		{0x30 /* VKEY_0 */, LINUX_KEY_0, 0x70027},
		{0x31 /* VKEY_1 */, LINUX_KEY_1, 0x7001E},
		{0x32 /* VKEY_2 */, LINUX_KEY_2, 0x7001F},
		{0x33 /* VKEY_3 */, LINUX_KEY_3, 0x70020},
		{0x34 /* VKEY_4 */, LINUX_KEY_4, 0x70021},
		{0x35 /* VKEY_5 */, LINUX_KEY_5, 0x70022},
		{0x36 /* VKEY_6 */, LINUX_KEY_6, 0x70023},
		{0x37 /* VKEY_7 */, LINUX_KEY_7, 0x70024},
		{0x38 /* VKEY_8 */, LINUX_KEY_8, 0x70025},
		{0x39 /* VKEY_9 */, LINUX_KEY_9, 0x70026},
		{0x41 /* VKEY_A */, LINUX_KEY_A, 0x70004},
		{0x42 /* VKEY_B */, LINUX_KEY_B, 0x70005},
		{0x43 /* VKEY_C */, LINUX_KEY_C, 0x70006},
		{0x44 /* VKEY_D */, LINUX_KEY_D, 0x70007},
		{0x45 /* VKEY_E */, LINUX_KEY_E, 0x70008},
		{0x46 /* VKEY_F */, LINUX_KEY_F, 0x70009},
		{0x47 /* VKEY_G */, LINUX_KEY_G, 0x7000A},
		{0x48 /* VKEY_H */, LINUX_KEY_H, 0x7000B},
		{0x49 /* VKEY_I */, LINUX_KEY_I, 0x7000C},
		{0x4A /* VKEY_J */, LINUX_KEY_J, 0x7000D},
		{0x4B /* VKEY_K */, LINUX_KEY_K, 0x7000E},
		{0x4C /* VKEY_L */, LINUX_KEY_L, 0x7000F},
		{0x4D /* VKEY_M */, LINUX_KEY_M, 0x70010},
		{0x4E /* VKEY_N */, LINUX_KEY_N, 0x70011},
		{0x4F /* VKEY_O */, LINUX_KEY_O, 0x70012},
		{0x50 /* VKEY_P */, LINUX_KEY_P, 0x70013},
		{0x51 /* VKEY_Q */, LINUX_KEY_Q, 0x70014},
		{0x52 /* VKEY_R */, LINUX_KEY_R, 0x70015},
		{0x53 /* VKEY_S */, LINUX_KEY_S, 0x70016},
		{0x54 /* VKEY_T */, LINUX_KEY_T, 0x70017},
		{0x55 /* VKEY_U */, LINUX_KEY_U, 0x70018},
		{0x56 /* VKEY_V */, LINUX_KEY_V, 0x70019},
		{0x57 /* VKEY_W */, LINUX_KEY_W, 0x7001A},
		{0x58 /* VKEY_X */, LINUX_KEY_X, 0x7001B},
		{0x59 /* VKEY_Y */, LINUX_KEY_Y, 0x7001C},
		{0x5A /* VKEY_Z */, LINUX_KEY_Z, 0x7001D},
		{0x5B /* VKEY_LWIN */, LINUX_LWIN, 0x700E3},
		{0x5C /* VKEY_RWIN */, LINUX_RWIN, 0x700E7},
		{0x5F /* VKEY_SLEEP */, LINUX_SLEEP, UNKNOWN},
		{0x60 /* VKEY_NUMPAD0 */, LINUX_NUMPAD0, 0x70062},
		{0x61 /* VKEY_NUMPAD1 */, LINUX_NUMPAD1, 0x70059},
		{0x62 /* VKEY_NUMPAD2 */, LINUX_NUMPAD2, 0x7005A},
		{0x63 /* VKEY_NUMPAD3 */, LINUX_NUMPAD3, 0x7005B},
		{0x64 /* VKEY_NUMPAD4 */, LINUX_NUMPAD4, 0x7005C},
		{0x65 /* VKEY_NUMPAD5 */, LINUX_NUMPAD5, 0x7005D},
		{0x66 /* VKEY_NUMPAD6 */, LINUX_NUMPAD6, 0x7005E},
		{0x67 /* VKEY_NUMPAD7 */, LINUX_NUMPAD7, 0x7005F},
		{0x68 /* VKEY_NUMPAD8 */, LINUX_NUMPAD8, 0x70060},
		{0x69 /* VKEY_NUMPAD9 */, LINUX_NUMPAD9, 0x70061},
		{0x6A /* VKEY_MULTIPLY */, LINUX_MULTIPLY, 0x70055},
		{0x6B /* VKEY_ADD */, LINUX_ADD, 0x70057},
		{0x6C /* VKEY_SEPARATOR */, LINUX_SEPARATOR, UNKNOWN},
		{0x6D /* VKEY_SUBTRACT */, LINUX_SUBTRACT, 0x70056},
		{0x6E /* VKEY_DECIMAL */, LINUX_DECIMAL, 0x70063},
		{0x6F /* VKEY_DIVIDE */, LINUX_DIVIDE, 0x70054},
		{0x70 /* VKEY_F1 */, LINUX_F1, 0x70046},
		{0x71 /* VKEY_F2 */, LINUX_F2, 0x70047},
		{0x72 /* VKEY_F3 */, LINUX_F3, 0x70048},
		{0x73 /* VKEY_F4 */, LINUX_F4, 0x70049},
		{0x74 /* VKEY_F5 */, LINUX_F5, 0x7004a},
		{0x75 /* VKEY_F6 */, LINUX_F6, 0x7004b},
		{0x76 /* VKEY_F7 */, LINUX_F7, 0x7004c},
		{0x77 /* VKEY_F8 */, LINUX_F8, 0x7004d},
		{0x78 /* VKEY_F9 */, LINUX_F9, 0x7004e},
		{0x79 /* VKEY_F10 */, LINUX_F10, 0x70044},
		{0x7A /* VKEY_F11 */, LINUX_F11, 0x70044},
		{0x7B /* VKEY_F12 */, LINUX_F12, 0x70045},
		{0x7C /* VKEY_F13 */, LINUX_F13, 0x7003a},
		{0x7D /* VKEY_F14 */, LINUX_F14, 0x7003b},
		{0x7E /* VKEY_F15 */, LINUX_F15, 0x7003c},
		{0x7F /* VKEY_F16 */, LINUX_F16, 0x7003d},
		{0x80 /* VKEY_F17 */, LINUX_F17, 0x7003e},
		{0x81 /* VKEY_F18 */, LINUX_F18, 0x7003f},
		{0x82 /* VKEY_F19 */, LINUX_F19, 0x70040},
		{0x83 /* VKEY_F20 */, LINUX_F20, 0x70041},
		{0x84 /* VKEY_F21 */, LINUX_F21, 0x70042},
		{0x85 /* VKEY_F22 */, LINUX_F12, 0x70043},
		{0x86 /* VKEY_F23 */, LINUX_F23, 0x70044},
		{0x87 /* VKEY_F24 */, LINUX_F24, 0x70045},
		{0x90 /* VKEY_NUMLOCK */, LINUX_NUMLOCK, 0x70053},
		{0x91 /* VKEY_SCROLL */, LINUX_SCROLL, 0x70047},
		{0xA0 /* VKEY_LSHIFT */, LINUX_LSHIFT, 0x700E1},
		{0xA1 /* VKEY_RSHIFT */, LINUX_RSHIFT, 0x700E5},
		{0xA2 /* VKEY_LCONTROL */, LINUX_LCONTROL, 0x700E0},
		{0xA3 /* VKEY_RCONTROL */, LINUX_RCONTROL, 0x700E4},
		{0xA4 /* VKEY_LMENU */, LINUX_LMENU, 0x7002E},
		{0xA5 /* VKEY_RMENU */, LINUX_RMENU, 0x700E6},
		{0xBA /* VKEY_OEM_1 */, LINUX_OEM_1, 0x70033},

		{0xBB /* VKEY_OEM_PLUS */, LINUX_OEM_PLUS, 0x7002E},
		{0xBC /* VKEY_OEM_COMMA */, LINUX_OEM_COMMA, 0x70036},
		{0xBD /* VKEY_OEM_MINUS */, LINUX_OEM_MINUS, 0x7002D},
		{0xBE /* VKEY_OEM_PERIOD */, LINUX_OEM_PERIOD, 0x70037},
		{0xBF /* VKEY_OEM_2 */, LINUX_OEM_2, 0x70038},
		{0xC0 /* VKEY_OEM_3 */, LINUX_OEM_3, 0x70035},
		{0xDB /* VKEY_OEM_4 */, LINUX_OEM_4, 0x7002F},
		{0xDC /* VKEY_OEM_5 */, LINUX_OEM_5, 0x70031},
		{0xDD /* VKEY_OEM_6 */, LINUX_OEM_6, 0x70030},
		{0xDE /* VKEY_OEM_7 */, LINUX_OEM_7, 0x70034},
		{0xE2 /* VKEY_NON_US_BACKSLASH */, LINUX_OEM_102, 0x70064},
	}
	mouse_abs_input *C.struct_libevdev_uinput
	mouse_rel_input *C.struct_libevdev_uinput
	keyboard_input *C.struct_libevdev_uinput
	input *C.struct_libevdev_uinput
);

type evdev_t *C.struct_libevdev

func keyboard() evdev_t {
	dev := C.libevdev_new()

	C.libevdev_set_uniq(dev, C.CString("Sunshine Keyboard"))
	C.libevdev_set_id_product(dev, 0xDEAD)
	C.libevdev_set_id_vendor(dev, 0xBEEF)
	C.libevdev_set_id_bustype(dev, 0x3)
	C.libevdev_set_id_version(dev, 0x111)
	C.libevdev_set_name(dev, C.CString("Keyboard passthrough"))

	C.libevdev_enable_event_type(dev, C.EV_KEY)
	for _, keycode := range keycodes {
		C.libevdev_enable_event_code(dev, C.EV_KEY, C.uint(keycode.linuxcode), unsafe.Pointer(nil))
	}
	C.libevdev_enable_event_type(dev, C.EV_MSC)
	C.libevdev_enable_event_code(dev, C.EV_MSC, C.MSC_SCAN, unsafe.Pointer(nil))

	for _, kc := range keycodes {
		linuxCodeMap[kc.wincode] = kc.linuxcode
		linuxScanCodeMap[kc.wincode] = kc.scancode
	}

	return dev
}

func mouse_rel() evdev_t {
	dev := C.libevdev_new()

	C.libevdev_set_uniq(dev, C.CString("Sunshine Mouse (Rel)"))
	C.libevdev_set_id_product(dev, 0x4038)
	C.libevdev_set_id_vendor(dev, 0x46D)
	C.libevdev_set_id_bustype(dev, 0x3)
	C.libevdev_set_id_version(dev, 0x111)
	C.libevdev_set_name(dev, C.CString("Logitech Wireless Mouse PID:4038"))

	C.libevdev_enable_event_type(dev, C.EV_KEY)
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_LEFT, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_RIGHT, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_MIDDLE, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_SIDE, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_EXTRA, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_FORWARD, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_BACK, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_TASK, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, 280, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, 281, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, 282, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, 283, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, 284, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, 285, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, 286, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, 287, unsafe.Pointer(nil))

	C.libevdev_enable_event_type(dev, C.EV_REL)
	C.libevdev_enable_event_code(dev, C.EV_REL, C.REL_X, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_REL, C.REL_Y, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_REL, C.REL_WHEEL, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_REL, C.REL_WHEEL_HI_RES, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_REL, C.REL_HWHEEL, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_REL, C.REL_HWHEEL_HI_RES, unsafe.Pointer(nil))

	C.libevdev_enable_event_type(dev, C.EV_MSC)
	C.libevdev_enable_event_code(dev, C.EV_MSC, C.MSC_SCAN, unsafe.Pointer(nil))

	return dev
}

func mouse_abs() evdev_t {
	dev := C.libevdev_new()

	C.libevdev_set_uniq(dev, C.CString("Sunshine Mouse (Abs)"))
	C.libevdev_set_id_product(dev, 0xDEAD)
	C.libevdev_set_id_vendor(dev, 0xBEEF)
	C.libevdev_set_id_bustype(dev, 0x3)
	C.libevdev_set_id_version(dev, 0x111)
	C.libevdev_set_name(dev, C.CString("Mouse passthrough"))

	C.libevdev_enable_property(dev, C.INPUT_PROP_DIRECT)

	C.libevdev_enable_event_type(dev, C.EV_KEY)
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_LEFT, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_RIGHT, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_MIDDLE, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_SIDE, unsafe.Pointer(nil))
	C.libevdev_enable_event_code(dev, C.EV_KEY, C.BTN_EXTRA, unsafe.Pointer(nil))

	C.libevdev_enable_event_type(dev, C.EV_MSC)
	C.libevdev_enable_event_code(dev, C.EV_MSC, C.MSC_SCAN, unsafe.Pointer(nil))

	absx := C.absinfo{
		0,
		0,
		1920,
		1,
		0,
		28,
	}

	absy := C.absinfo{
		0,
		0,
		1080,
		1,
		0,
		28,
	}
	C.libevdev_enable_event_type(dev, C.EV_ABS)
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_X, unsafe.Pointer(&absx))
	C.libevdev_enable_event_code(dev, C.EV_ABS, C.ABS_Y, unsafe.Pointer(&absy))
	_ = absx
	_ = absy

	return dev
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

func _init() error {
	keyboard_dev := keyboard()
	mouse_rel_dev := mouse_rel()
	mouse_abs_dev := mouse_abs()
	gamepad_dev := x360()

	rv := C.libevdev_uinput_create_from_device(mouse_abs_dev, C.LIBEVDEV_UINPUT_OPEN_MANAGED, &mouse_abs_input)
	if rv > 0 || mouse_abs_input == nil {
		return errors.New("failed to create new abs device")
	}
	rv = C.libevdev_uinput_create_from_device(mouse_rel_dev, C.LIBEVDEV_UINPUT_OPEN_MANAGED, &mouse_rel_input)
	if rv > 0 || mouse_rel_input == nil {
		return errors.New("failed to create new rel device")
	}
	rv = C.libevdev_uinput_create_from_device(keyboard_dev, C.LIBEVDEV_UINPUT_OPEN_MANAGED, &keyboard_input)
	if rv > 0 || keyboard_input == nil {
		return errors.New("failed to create new keyboard device")
	}
	rv = C.libevdev_uinput_create_from_device(gamepad_dev, C.LIBEVDEV_UINPUT_OPEN_MANAGED, &input);
	if rv > 0 || input == nil {
		return errors.New("failed to create new gamepad device")
	}

	return nil
}


func init(){
	_init()
}

func SendMouseRelative(x float32, y float32) {

	C.libevdev_uinput_write_event(mouse_rel_input, C.EV_REL, C.REL_X, C.int(x));
	C.libevdev_uinput_write_event(mouse_rel_input, C.EV_REL, C.REL_Y, C.int(y));
    C.libevdev_uinput_write_event(mouse_rel_input, C.EV_SYN, C.SYN_REPORT, 0);

	mouse_pos = mouse_position {
		mode: "rel",
		offset_x: float64(x),
		offset_y: float64(y),
	}
}

func lround(x float64) int {
	if x < 0 {
		return int(math.Ceil(x - 0.5))
	}
	return int(math.Floor(x + 0.5))
}

func SendMouseAbsolute(x float32, y float32) {
    scaled_x := lround((float64(x) + touch_port.offset_x) * (target_touch_port.width / touch_port.width));
    scaled_y := lround((float64(y) + touch_port.offset_y) * (target_touch_port.height / touch_port.height));

    C.libevdev_uinput_write_event(mouse_abs_input, C.EV_ABS, C.ABS_X, C.int(scaled_x));
    C.libevdev_uinput_write_event(mouse_abs_input, C.EV_ABS, C.ABS_Y, C.int(scaled_y));
    C.libevdev_uinput_write_event(mouse_abs_input, C.EV_SYN, C.SYN_REPORT, 0);

    // Remember this was the last device we sent input on
	mouse_pos = mouse_position {
		mode: "abs",
		offset_x: float64(scaled_x),
		offset_y: float64(scaled_y),
	}
}

func SendMouseWheel(wheel float64) {
}

func SendMouseButton(button int, is_up bool) {
    // var btn_type int;
    // var scan int;
	var chosen_mouse_dev *C.struct_libevdev_uinput;

	code := 1
	if !is_up {
		code = 0
	}

	// TODO: double check in demo

    // if button == 1 {
    //   btn_type = LINUX_LBUTTON
    //   scan = 90001
    // } else if button == 2 {
    //   btn_type = LINUX_MBUTTON
    //   scan = 90003
    // } else if(button == 3) {
    //   btn_type = LINUX_RBUTTON
    //   scan = 90002
    // } else if (button == 4) {
    //   btn_type = LINUX_BTN_SIDE
    //   scan = 90004
    // } else {
    //   btn_type = LINUX_BTN_EXTRA
    //   scan = 90005
    // }

	linuxCode := linuxCodeMap[C.uint(button)]
	linuxScanCode := linuxScanCodeMap[C.uint(button)]

	if mouse_pos.mode == "abs"{
		chosen_mouse_dev = mouse_abs_input
	} else{
		chosen_mouse_dev = mouse_rel_input
	} 

    C.libevdev_uinput_write_event(chosen_mouse_dev, C.EV_MSC, C.MSC_SCAN, C.int(linuxScanCode));
    C.libevdev_uinput_write_event(chosen_mouse_dev, C.EV_KEY, C.uint(linuxCode), C.int(code));
    C.libevdev_uinput_write_event(chosen_mouse_dev, C.EV_SYN, C.SYN_REPORT, 0);

	fmt.Println("foo")
    // if (release) {
    //   *chosen_mouse_dev_buttons_down &= ~(1 << button);
    // } else {
    //   *chosen_mouse_dev_buttons_down |= (1 << button);
    // }
}

func SendKeyboard(keycode int, is_up bool, scan_code bool) {

	code := 1
	if !is_up {
		code = 0
	}

	keyboard_dev := keyboard()

	C.libevdev_uinput_create_from_device(keyboard_dev, C.LIBEVDEV_UINPUT_OPEN_MANAGED, &keyboard_input);


    if scan_code {
		linuxScanCode, ok := linuxScanCodeMap[C.uint(keycode)]
		if(ok){
			fmt.Println(linuxScanCode)
			// TODO: scancode is not work
			// libevdev_uinput_write_event(keyboard, EV_MSC, MSC_SCAN, C.uint(linuxScanCode));
			// C.libevdev_uinput_write_event(keyboard_input, C.EV_KEY, C.uint(linuxScanCode), C.int(code));
			return
		} else {
			fmt.Println("Not found scan keycode in linux", keycode)
		}
    }
	linuxCode, ok := linuxCodeMap[C.uint(keycode)]
	if ok {
		C.libevdev_uinput_write_event(keyboard_input, C.EV_KEY, C.uint(linuxCode), C.int(code));
	} else {
		fmt.Println("Not found keycode in linux", keycode)
	}
}

func SetClipboard(text string) {
}

func DisplayPosition(name string) (x, y, width, height int, err error) {
	return 0, 0, 0, 0, nil
}

func GetVirtualDisplay() (x, y int) {
	return 0, 0
}

