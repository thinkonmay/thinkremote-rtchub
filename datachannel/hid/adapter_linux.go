package hid

/*
#include <libevdev/libevdev.h>
#include <libevdev/libevdev-uinput.h>

typedef struct libevdev _evdev;
typedef struct input_absinfo absinfo;

#cgo pkg-config: libevdev
*/
import "C"
import (
	"errors"
	"unsafe"
)

type KeyCode struct {
	keycode  C.uint
	scancode C.uint
}


var (
	keycodes = []KeyCode{
		// {0x08 /* VKEY_BACK */, BACKSPACE},
		{0x09 /* VKEY_TAB */, TAB},
		{0x0C /* VKEY_CLEAR */, CLEAR},
		// {0x0D /* VKEY_RETURN */, ENTER},
		// {0x10 /* VKEY_SHIFT */, LEFTSHIFT},
		// {0x11 /* VKEY_CONTROL */, LEFTCTRL},
		// {0x12 /* VKEY_MENU */, LEFTALT},
		{0x13 /* VKEY_PAUSE */, PAUSE},
		// {0x14 /* VKEY_CAPITAL */, CAPSLOCK},
		// {0x15 /* VKEY_KANA */, KATAKANAHIRAGANA},
		// {0x16 /* VKEY_HANGUL */, HANGEUL},
		{0x17 /* VKEY_JUNJA */, HANJA},
		// {0x19 /* VKEY_KANJI */, KATAKANA},
		// {0x1B /* VKEY_ESCAPE */, ESC},
		{0x20 /* VKEY_SPACE */, SPACE},
		// {0x21 /* VKEY_PRIOR */, PAGEUP},
		// {0x22 /* VKEY_NEXT */, PAGEDOWN},
		{0x23 /* VKEY_END */, END},
		{0x24 /* VKEY_HOME */, HOME},
		{0x25 /* VKEY_LEFT */, LEFT},
		{0x26 /* VKEY_UP */, UP},
		{0x27 /* VKEY_RIGHT */, RIGHT},
		{0x28 /* VKEY_DOWN */, DOWN},
		{0x29 /* VKEY_SELECT */, SELECT},
		{0x2A /* VKEY_PRINT */, PRINT},
		// {0x2C /* VKEY_SNAPSHOT */, SYSRQ},
		{0x2D /* VKEY_INSERT */, INSERT},
		{0x2E /* VKEY_DELETE */, DELETE},
		{0x2F /* VKEY_HELP */, HELP},
		{0x30 /* VKEY_0 */, KEY_0},
		{0x31 /* VKEY_1 */, KEY_1},
		{0x32 /* VKEY_2 */, KEY_2},
		{0x33 /* VKEY_3 */, KEY_3},
		{0x34 /* VKEY_4 */, KEY_4},
		{0x35 /* VKEY_5 */, KEY_5},
		{0x36 /* VKEY_6 */, KEY_6},
		{0x37 /* VKEY_7 */, KEY_7},
		{0x38 /* VKEY_8 */, KEY_8},
		{0x39 /* VKEY_9 */, KEY_9},
		{0x41 /* VKEY_A */, KEY_A},
		{0x42 /* VKEY_B */, KEY_B},
		{0x43 /* VKEY_C */, KEY_C},
		{0x44 /* VKEY_D */, KEY_D},
		{0x45 /* VKEY_E */, KEY_E},
		{0x46 /* VKEY_F */, KEY_F},
		{0x47 /* VKEY_G */, KEY_G},
		{0x48 /* VKEY_H */, KEY_H},
		{0x49 /* VKEY_I */, KEY_I},
		{0x4A /* VKEY_J */, KEY_J},
		{0x4B /* VKEY_K */, KEY_K},
		{0x4C /* VKEY_L */, KEY_L},
		{0x4D /* VKEY_M */, KEY_M},
		{0x4E /* VKEY_N */, KEY_N},
		{0x4F /* VKEY_O */, KEY_O},
		{0x50 /* VKEY_P */, KEY_P},
		{0x51 /* VKEY_Q */, KEY_Q},
		{0x52 /* VKEY_R */, KEY_R},
		{0x53 /* VKEY_S */, KEY_S},
		{0x54 /* VKEY_T */, KEY_T},
		{0x55 /* VKEY_U */, KEY_U},
		{0x56 /* VKEY_V */, KEY_V},
		{0x57 /* VKEY_W */, KEY_W},
		{0x58 /* VKEY_X */, KEY_X},
		{0x59 /* VKEY_Y */, KEY_Y},
		{0x5A /* VKEY_Z */, KEY_Z},
		// {0x5B /* VKEY_LWIN */, LEFTMETA},
		// {0x5C /* VKEY_RWIN */, RIGHTMETA},
		{0x5F /* VKEY_SLEEP */, SLEEP},
		// {0x60 /* VKEY_NUMPAD0 */, KP0},
		// {0x61 /* VKEY_NUMPAD1 */, KP1},
		// {0x62 /* VKEY_NUMPAD2 */, KP2},
		// {0x63 /* VKEY_NUMPAD3 */, KP3},
		// {0x64 /* VKEY_NUMPAD4 */, KP4},
		// {0x65 /* VKEY_NUMPAD5 */, KP5},
		// {0x66 /* VKEY_NUMPAD6 */, KP6},
		// {0x67 /* VKEY_NUMPAD7 */, KP7},
		// {0x68 /* VKEY_NUMPAD8 */, KP8},
		// {0x69 /* VKEY_NUMPAD9 */, KP9},
		// {0x6A /* VKEY_MULTIPLY */, KPASTERISK},
		// {0x6B /* VKEY_ADD */, KPPLUS},
		// {0x6C /* VKEY_SEPARATOR */, KPCOMMA},
		// {0x6D /* VKEY_SUBTRACT */, KPMINUS},
		// {0x6E /* VKEY_DECIMAL */, KPDOT},
		// {0x6F /* VKEY_DIVIDE */, KPSLASH},
		{0x70 /* VKEY_F1 */, F1},
		{0x71 /* VKEY_F2 */, F2},
		{0x72 /* VKEY_F3 */, F3},
		{0x73 /* VKEY_F4 */, F4},
		{0x74 /* VKEY_F5 */, F5},
		{0x75 /* VKEY_F6 */, F6},
		{0x76 /* VKEY_F7 */, F7},
		{0x77 /* VKEY_F8 */, F8},
		{0x78 /* VKEY_F9 */, F9},
		{0x79 /* VKEY_F10 */, F10},
		{0x7A /* VKEY_F11 */, F11},
		{0x7B /* VKEY_F12 */, F12},
		{0x7C /* VKEY_F13 */, F13},
		{0x7D /* VKEY_F14 */, F14},
		{0x7E /* VKEY_F15 */, F15},
		{0x7F /* VKEY_F16 */, F16},
		{0x80 /* VKEY_F17 */, F17},
		{0x81 /* VKEY_F18 */, F18},
		{0x82 /* VKEY_F19 */, F19},
		{0x83 /* VKEY_F20 */, F20},
		{0x84 /* VKEY_F21 */, F21},
		{0x85 /* VKEY_F22 */, F12},
		{0x86 /* VKEY_F23 */, F23},
		{0x87 /* VKEY_F24 */, F24},
		{0x90 /* VKEY_NUMLOCK */, NUMLOCK},
		// {0x91 /* VKEY_SCROLL */, SCROLLLOCK},
		// {0xA0 /* VKEY_LSHIFT */, LEFTSHIFT},
		// {0xA1 /* VKEY_RSHIFT */, RIGHTSHIFT},
		// {0xA2 /* VKEY_LCONTROL */, LEFTCTRL},
		// {0xA3 /* VKEY_RCONTROL */, RIGHTCTRL},
		// {0xA4 /* VKEY_LMENU */, LEFTALT},
		// {0xA5 /* VKEY_RMENU */, RIGHTALT},
		// {0xBA /* VKEY_OEM_1 */, SEMICOLON},
		// {0xBB /* VKEY_OEM_PLUS */, EQUAL},
		// {0xBC /* VKEY_OEM_COMMA */, COMMA},
		// {0xBD /* VKEY_OEM_MINUS */, MINUS},
		// {0xBE /* VKEY_OEM_PERIOD */, DOT},
		// {0xBF /* VKEY_OEM_2 */, SLASH},
		// {0xC0 /* VKEY_OEM_3 */, GRAVE},
		// {0xDB /* VKEY_OEM_4 */, LEFTBRACE},
		// {0xDC /* VKEY_OEM_5 */, BACKSLASH},
		// {0xDD /* VKEY_OEM_6 */, RIGHTBRACE},
		// {0xDE /* VKEY_OEM_7 */, APOSTROPHE},
		// {0xE2 /* VKEY_NON_US_BACKSLASH */, 102ND},
	}

	mouse_abs_input *C.struct_libevdev_uinput
	mouse_rel_input *C.struct_libevdev_uinput
	keyboard_input *C.struct_libevdev_uinput
	input *C.struct_libevdev_uinput
)

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
		C.libevdev_enable_event_code(dev, C.EV_KEY, keycode.keycode, unsafe.Pointer(nil))
	}
	C.libevdev_enable_event_type(dev, C.EV_MSC)
	C.libevdev_enable_event_code(dev, C.EV_MSC, C.MSC_SCAN, unsafe.Pointer(nil))

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
	panic(_init())
}

func SendMouseRelative(x float32, y float32) {
}

func SendMouseAbsolute(x float32, y float32) {
}

func SendMouseWheel(wheel float64) {
}

func SendMouseButton(button int, is_up bool) {
}

func SendKeyboard(keycode int, is_up bool, scan_code bool) {
}

func SetClipboard(text string) {
}

func DisplayPosition(name string) (x, y, width, height int, err error) {
	return 0, 0, 0, 0, nil
}

func GetVirtualDisplay() (x, y int) {
	return 0, 0
}
