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
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

const UNKNOWN = 0

type KeyCode struct {
	linuxcode C.uint
	scancode  C.uint
}

type evdev_t *C.struct_libevdev

type touch_port_t struct {
	offset_x float64
	offset_y float64
	height   float64
	width    float64
}

var (
	use_mouse_abs = true

	target_touch_port = touch_port_t{
		height: 1920,
		width:  1200,
	}

	last_mouse_device_used         any
	last_mouse_device_buttons_down *uint8 = nil
	mouse_abs_buttons_down         uint8  = 0
	mouse_rel_buttons_down         uint8  = 0

	display  = touch_port_t{}
	keycodes = map[int]KeyCode{
		0x08 /* VKEY_BACK */ :             {C.KEY_BACKSPACE, 0x7002A},
		0x09 /* VKEY_TAB */ :              {C.KEY_TAB, 0x7002B},
		0x0C /* VKEY_CLEAR */ :            {C.KEY_CLEAR, UNKNOWN},
		0x0D /* VKEY_RETURN */ :           {C.KEY_ENTER, 0x70028},
		0x10 /* VKEY_SHIFT */ :            {C.KEY_LEFTSHIFT, 0x700E1},
		0x11 /* VKEY_CONTROL */ :          {C.KEY_LEFTCTRL, 0x700E0},
		0x12 /* VKEY_MENU */ :             {C.KEY_LEFTALT, UNKNOWN},
		0x13 /* VKEY_PAUSE */ :            {C.KEY_PAUSE, UNKNOWN},
		0x14 /* VKEY_CAPITAL */ :          {C.KEY_CAPSLOCK, 0x70039},
		0x15 /* VKEY_KANA */ :             {C.KEY_KATAKANAHIRAGANA, UNKNOWN},
		0x16 /* VKEY_HANGUL */ :           {C.KEY_HANGEUL, UNKNOWN},
		0x17 /* VKEY_JUNJA */ :            {C.KEY_HANJA, UNKNOWN},
		0x19 /* VKEY_KANJI */ :            {C.KEY_KATAKANA, UNKNOWN},
		0x1B /* VKEY_ESCAPE */ :           {C.KEY_ESC, 0x70029},
		0x20 /* VKEY_SPACE */ :            {C.KEY_SPACE, 0x7002C},
		0x21 /* VKEY_PRIOR */ :            {C.KEY_PAGEUP, 0x7004B},
		0x22 /* VKEY_NEXT */ :             {C.KEY_PAGEDOWN, 0x7004E},
		0x23 /* VKEY_END */ :              {C.KEY_END, 0x7004D},
		0x24 /* VKEY_HOME */ :             {C.KEY_HOME, 0x7004A},
		0x25 /* VKEY_LEFT */ :             {C.KEY_LEFT, 0x70050},
		0x26 /* VKEY_UP */ :               {C.KEY_UP, 0x70052},
		0x27 /* VKEY_RIGHT */ :            {C.KEY_RIGHT, 0x7004F},
		0x28 /* VKEY_DOWN */ :             {C.KEY_DOWN, 0x70051},
		0x29 /* VKEY_SELECT */ :           {C.KEY_SELECT, UNKNOWN},
		0x2A /* VKEY_PRINT */ :            {C.KEY_PRINT, UNKNOWN},
		0x2C /* VKEY_SNAPSHOT */ :         {C.KEY_SYSRQ, 0x70046},
		0x2D /* VKEY_INSERT */ :           {C.KEY_INSERT, 0x70049},
		0x2E /* VKEY_DELETE */ :           {C.KEY_DELETE, 0x7004C},
		0x2F /* VKEY_HELP */ :             {C.KEY_HELP, UNKNOWN},
		0x30 /* VKEY_0 */ :                {C.KEY_0, 0x70027},
		0x31 /* VKEY_1 */ :                {C.KEY_1, 0x7001E},
		0x32 /* VKEY_2 */ :                {C.KEY_2, 0x7001F},
		0x33 /* VKEY_3 */ :                {C.KEY_3, 0x70020},
		0x34 /* VKEY_4 */ :                {C.KEY_4, 0x70021},
		0x35 /* VKEY_5 */ :                {C.KEY_5, 0x70022},
		0x36 /* VKEY_6 */ :                {C.KEY_6, 0x70023},
		0x37 /* VKEY_7 */ :                {C.KEY_7, 0x70024},
		0x38 /* VKEY_8 */ :                {C.KEY_8, 0x70025},
		0x39 /* VKEY_9 */ :                {C.KEY_9, 0x70026},
		0x41 /* VKEY_A */ :                {C.KEY_A, 0x70004},
		0x42 /* VKEY_B */ :                {C.KEY_B, 0x70005},
		0x43 /* VKEY_C */ :                {C.KEY_C, 0x70006},
		0x44 /* VKEY_D */ :                {C.KEY_D, 0x70007},
		0x45 /* VKEY_E */ :                {C.KEY_E, 0x70008},
		0x46 /* VKEY_F */ :                {C.KEY_F, 0x70009},
		0x47 /* VKEY_G */ :                {C.KEY_G, 0x7000A},
		0x48 /* VKEY_H */ :                {C.KEY_H, 0x7000B},
		0x49 /* VKEY_I */ :                {C.KEY_I, 0x7000C},
		0x4A /* VKEY_J */ :                {C.KEY_J, 0x7000D},
		0x4B /* VKEY_K */ :                {C.KEY_K, 0x7000E},
		0x4C /* VKEY_L */ :                {C.KEY_L, 0x7000F},
		0x4D /* VKEY_M */ :                {C.KEY_M, 0x70010},
		0x4E /* VKEY_N */ :                {C.KEY_N, 0x70011},
		0x4F /* VKEY_O */ :                {C.KEY_O, 0x70012},
		0x50 /* VKEY_P */ :                {C.KEY_P, 0x70013},
		0x51 /* VKEY_Q */ :                {C.KEY_Q, 0x70014},
		0x52 /* VKEY_R */ :                {C.KEY_R, 0x70015},
		0x53 /* VKEY_S */ :                {C.KEY_S, 0x70016},
		0x54 /* VKEY_T */ :                {C.KEY_T, 0x70017},
		0x55 /* VKEY_U */ :                {C.KEY_U, 0x70018},
		0x56 /* VKEY_V */ :                {C.KEY_V, 0x70019},
		0x57 /* VKEY_W */ :                {C.KEY_W, 0x7001A},
		0x58 /* VKEY_X */ :                {C.KEY_X, 0x7001B},
		0x59 /* VKEY_Y */ :                {C.KEY_Y, 0x7001C},
		0x5A /* VKEY_Z */ :                {C.KEY_Z, 0x7001D},
		0x5B /* VKEY_LWIN */ :             {C.KEY_LEFTMETA, 0x700E3},
		0x5C /* VKEY_RWIN */ :             {C.KEY_RIGHTMETA, 0x700E7},
		0x5F /* VKEY_SLEEP */ :            {C.KEY_SLEEP, UNKNOWN},
		0x60 /* VKEY_NUMPAD0 */ :          {C.KEY_KP0, 0x70062},
		0x61 /* VKEY_NUMPAD1 */ :          {C.KEY_KP1, 0x70059},
		0x62 /* VKEY_NUMPAD2 */ :          {C.KEY_KP2, 0x7005A},
		0x63 /* VKEY_NUMPAD3 */ :          {C.KEY_KP3, 0x7005B},
		0x64 /* VKEY_NUMPAD4 */ :          {C.KEY_KP4, 0x7005C},
		0x65 /* VKEY_NUMPAD5 */ :          {C.KEY_KP5, 0x7005D},
		0x66 /* VKEY_NUMPAD6 */ :          {C.KEY_KP6, 0x7005E},
		0x67 /* VKEY_NUMPAD7 */ :          {C.KEY_KP7, 0x7005F},
		0x68 /* VKEY_NUMPAD8 */ :          {C.KEY_KP8, 0x70060},
		0x69 /* VKEY_NUMPAD9 */ :          {C.KEY_KP9, 0x70061},
		0x6A /* VKEY_MULTIPLY */ :         {C.KEY_KPASTERISK, 0x70055},
		0x6B /* VKEY_ADD */ :              {C.KEY_KPPLUS, 0x70057},
		0x6C /* VKEY_SEPARATOR */ :        {C.KEY_KPCOMMA, UNKNOWN},
		0x6D /* VKEY_SUBTRACT */ :         {C.KEY_KPMINUS, 0x70056},
		0x6E /* VKEY_DECIMAL */ :          {C.KEY_KPDOT, 0x70063},
		0x6F /* VKEY_DIVIDE */ :           {C.KEY_KPSLASH, 0x70054},
		0x70 /* VKEY_F1 */ :               {C.KEY_F1, 0x70046},
		0x71 /* VKEY_F2 */ :               {C.KEY_F2, 0x70047},
		0x72 /* VKEY_F3 */ :               {C.KEY_F3, 0x70048},
		0x73 /* VKEY_F4 */ :               {C.KEY_F4, 0x70049},
		0x74 /* VKEY_F5 */ :               {C.KEY_F5, 0x7004a},
		0x75 /* VKEY_F6 */ :               {C.KEY_F6, 0x7004b},
		0x76 /* VKEY_F7 */ :               {C.KEY_F7, 0x7004c},
		0x77 /* VKEY_F8 */ :               {C.KEY_F8, 0x7004d},
		0x78 /* VKEY_F9 */ :               {C.KEY_F9, 0x7004e},
		0x79 /* VKEY_F10 */ :              {C.KEY_F10, 0x70044},
		0x7A /* VKEY_F11 */ :              {C.KEY_F11, 0x70044},
		0x7B /* VKEY_F12 */ :              {C.KEY_F12, 0x70045},
		0x7C /* VKEY_F13 */ :              {C.KEY_F13, 0x7003a},
		0x7D /* VKEY_F14 */ :              {C.KEY_F14, 0x7003b},
		0x7E /* VKEY_F15 */ :              {C.KEY_F15, 0x7003c},
		0x7F /* VKEY_F16 */ :              {C.KEY_F16, 0x7003d},
		0x80 /* VKEY_F17 */ :              {C.KEY_F17, 0x7003e},
		0x81 /* VKEY_F18 */ :              {C.KEY_F18, 0x7003f},
		0x82 /* VKEY_F19 */ :              {C.KEY_F19, 0x70040},
		0x83 /* VKEY_F20 */ :              {C.KEY_F20, 0x70041},
		0x84 /* VKEY_F21 */ :              {C.KEY_F21, 0x70042},
		0x85 /* VKEY_F22 */ :              {C.KEY_F12, 0x70043},
		0x86 /* VKEY_F23 */ :              {C.KEY_F23, 0x70044},
		0x87 /* VKEY_F24 */ :              {C.KEY_F24, 0x70045},
		0x90 /* VKEY_NUMLOCK */ :          {C.KEY_NUMLOCK, 0x70053},
		0x91 /* VKEY_SCROLL */ :           {C.KEY_SCROLLLOCK, 0x70047},
		0xA0 /* VKEY_LSHIFT */ :           {C.KEY_LEFTSHIFT, 0x700E1},
		0xA1 /* VKEY_RSHIFT */ :           {C.KEY_RIGHTSHIFT, 0x700E5},
		0xA2 /* VKEY_LCONTROL */ :         {C.KEY_LEFTCTRL, 0x700E0},
		0xA3 /* VKEY_RCONTROL */ :         {C.KEY_RIGHTCTRL, 0x700E4},
		0xA4 /* VKEY_LMENU */ :            {C.KEY_LEFTALT, 0x7002E},
		0xA5 /* VKEY_RMENU */ :            {C.KEY_RIGHTALT, 0x700E6},
		0xBA /* VKEY_OEM_1 */ :            {C.KEY_SEMICOLON, 0x70033},
		0xBB /* VKEY_OEM_PLUS */ :         {C.KEY_EQUAL, 0x7002E},
		0xBC /* VKEY_OEM_COMMA */ :        {C.KEY_COMMA, 0x70036},
		0xBD /* VKEY_OEM_MINUS */ :        {C.KEY_MINUS, 0x7002D},
		0xBE /* VKEY_OEM_PERIOD */ :       {C.KEY_DOT, 0x70037},
		0xBF /* VKEY_OEM_2 */ :            {C.KEY_SLASH, 0x70038},
		0xC0 /* VKEY_OEM_3 */ :            {C.KEY_GRAVE, 0x70035},
		0xDB /* VKEY_OEM_4 */ :            {C.KEY_LEFTBRACE, 0x7002F},
		0xDC /* VKEY_OEM_5 */ :            {C.KEY_BACKSLASH, 0x70031},
		0xDD /* VKEY_OEM_6 */ :            {C.KEY_RIGHTBRACE, 0x70030},
		0xDE /* VKEY_OEM_7 */ :            {C.KEY_APOSTROPHE, 0x70034},
		0xE2 /* VKEY_NON_US_BACKSLASH */ : {C.KEY_102ND, 0x70064},
	}
	mouse_abs_input *C.struct_libevdev_uinput
	mouse_rel_input *C.struct_libevdev_uinput
	keyboard_input  *C.struct_libevdev_uinput
	gamepad_input   *C.struct_libevdev_uinput
)

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
	rv = C.libevdev_uinput_create_from_device(gamepad_dev, C.LIBEVDEV_UINPUT_OPEN_MANAGED, &gamepad_input)
	if rv > 0 || gamepad_input == nil {
		return errors.New("failed to create new gamepad device")
	}

	go func() {
		for {
			_, _, x, y, err := DisplayPosition("")
			if err != nil {
				panic(err)
			}

			display.height = float64(y)
			display.width = float64(x)
		}
	}()

	return nil
}

func init() {
	err := _init()
	if err != nil {
		fmt.Printf("failed to initialize hid for linux : %s\n", err.Error())
	}
}

func SendMouseRelative(x, y float32) {

	C.libevdev_uinput_write_event(mouse_rel_input, C.EV_REL, C.REL_X, C.int(x))
	C.libevdev_uinput_write_event(mouse_rel_input, C.EV_REL, C.REL_Y, C.int(y))
	C.libevdev_uinput_write_event(mouse_rel_input, C.EV_SYN, C.SYN_REPORT, 0)

	use_mouse_abs = false
}

func SendMouseAbsolute(x, y float32) {
	C.libevdev_uinput_write_event(mouse_abs_input, C.EV_ABS, C.ABS_X, C.int(x*float32(display.width)))
	C.libevdev_uinput_write_event(mouse_abs_input, C.EV_ABS, C.ABS_Y, C.int(y*float32(display.height)))
	C.libevdev_uinput_write_event(mouse_abs_input, C.EV_SYN, C.SYN_REPORT, 0)

	// Remember this was the last device we sent input on
	use_mouse_abs = true
}

func SendMouseWheel(wheel float64) {
}

func SendMouseButton(button int, is_up bool) {
	var btn_type int
	var scan int
	var chosen_mouse_dev *C.struct_libevdev_uinput
	if button == 1 {
		btn_type = C.BTN_LEFT
		scan = 90001
	} else if button == 2 {
		btn_type = C.BTN_MIDDLE
		scan = 90003
	} else if button == 3 {
		btn_type = C.BTN_RIGHT
		scan = 90002
	} else if button == 4 {
		btn_type = C.BTN_SIDE
		scan = 90004
	} else {
		btn_type = C.BTN_EXTRA
		scan = 90005
	}

	if use_mouse_abs {
		chosen_mouse_dev = mouse_abs_input
	} else {
		chosen_mouse_dev = mouse_rel_input
	}

	code := 0
	if !is_up {
		code = 1

	}

	C.libevdev_uinput_write_event(chosen_mouse_dev, C.EV_MSC, C.MSC_SCAN, C.int(scan))
	C.libevdev_uinput_write_event(chosen_mouse_dev, C.EV_KEY, C.uint(btn_type), C.int(code))
	C.libevdev_uinput_write_event(chosen_mouse_dev, C.EV_SYN, C.SYN_REPORT, 0)
}

func SendKeyboard(keycode int, is_up bool, scan_code bool) {
	code := 0
	if !is_up {
		code = 1
	}

	linuxCode, ok := keycodes[keycode]
	if !ok {
		return
	}

	C.libevdev_uinput_write_event(keyboard_input, C.EV_KEY, C.uint(linuxCode.linuxcode), C.int(code))
	C.libevdev_uinput_write_event(keyboard_input, C.EV_SYN, C.SYN_REPORT, 0)
}

func SetClipboard(text string) {
}

func DisplayPosition(name string) (x, y, width, height int, err error) {
	out, err := exec.Command("xdpyinfo").Output()
	if err != nil {
		panic(err)
	}

	resx, resy := int64(0), int64(0)
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "dimensions") {
			word0 := []string{}
			for _, word := range strings.Split(line, " ") {
				if len(word) == 0 {
					continue
				}
				word0 = append(word0, word)
			}

			for index, val := range word0 {
				if val == "pixels" {
					if index == 0 {
						continue
					}

					res := strings.Split(word0[index-1], "x")
					if len(res) < 2 {
						continue
					}

					resx, err = strconv.ParseInt(res[0], 10, 32)
					if err != nil {
						panic(err)
					}
					resy, err = strconv.ParseInt(res[1], 10, 32)
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
	return 0, 0, int(resx), int(resy), nil

}

func GetVirtualDisplay() (x, y int) {
	return 0, 0
}
