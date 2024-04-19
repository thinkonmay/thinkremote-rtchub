package hid

const (
    LBUTTON = 0x01
    RBUTTON = 0x02
    CANCEL = 0x03
    MBUTTON = 0x04
    XBUTTON1 = 0x05
    XBUTTON2 = 0x06
    BACK = 0x08
    TAB = 0x09
    CLEAR = 0x0C
    RETURN = 0x0D
    SHIFT = 0x10
    CONTROL = 0x11
    MENU = 0x12
    PAUSE = 0x13
    CAPITAL = 0x14
    KANA = 0x15
    HANGUL = 0x15
    JUNJA = 0x17
    FINAL = 0x18
    HANJA = 0x19
    KANJI = 0x19
    ESCAPE = 0x1B
    CONVERT = 0x1C
    NONCONVERT = 0x1D
    ACCEPT = 0x1E
    MODECHANGE = 0x1F
    SPACE = 0x20
    PRIOR = 0x21
    NEXT = 0x22
    END = 0x23
    HOME = 0x24
    LEFT = 0x25
    UP = 0x26
    RIGHT = 0x27
    DOWN = 0x28
    SELECT = 0x29
    PRINT = 0x2A
    EXECUTE = 0x2B
    SNAPSHOT = 0x2C
    INSERT = 0x2D
    DELETE = 0x2E
    HELP = 0x2F
    KEY_0 = 0x30
    KEY_1 = 0x31
    KEY_2 = 0x32
    KEY_3 = 0x33
    KEY_4 = 0x34
    KEY_5 = 0x35
    KEY_6 = 0x36
    KEY_7 = 0x37
    KEY_8 = 0x38
    KEY_9 = 0x39
    KEY_A = 0x41
    KEY_B = 0x42
    KEY_C = 0x43
    KEY_D = 0x44
    KEY_E = 0x45
    KEY_F = 0x46
    KEY_G = 0x47
    KEY_H = 0x48
    KEY_I = 0x49
    KEY_J = 0x4A
    KEY_K = 0x4B
    KEY_L = 0x4C
    KEY_M = 0x4D
    KEY_N = 0x4E
    KEY_O = 0x4F
    KEY_P = 0x50
    KEY_Q = 0x51
    KEY_R = 0x52
    KEY_S = 0x53
    KEY_T = 0x54
    KEY_U = 0x55
    KEY_V = 0x56
    KEY_W = 0x57
    KEY_X = 0x58
    KEY_Y = 0x59
    KEY_Z = 0x5A
    LWIN = 0x5B
    RWIN = 0x5C
    APPS = 0x5D
    SLEEP = 0x5F
    NUMPAD0 = 0x60
    NUMPAD1 = 0x61
    NUMPAD2 = 0x62
    NUMPAD3 = 0x63
    NUMPAD4 = 0x64
    NUMPAD5 = 0x65
    NUMPAD6 = 0x66
    NUMPAD7 = 0x67
    NUMPAD8 = 0x68
    NUMPAD9 = 0x69
    MULTIPLY = 0x6A
    ADD = 0x6B
    SEPARATOR = 0x6C
    SUBTRACT = 0x6D
    DECIMAL = 0x6E
    DIVIDE = 0x6F
    F1 = 0x70
    F2 = 0x71
    F3 = 0x72
    F4 = 0x73
    F5 = 0x74
    F6 = 0x75
    F7 = 0x76
    F8 = 0x77
    F9 = 0x78
    F10 = 0x79
    F11 = 0x7A
    F12 = 0x7B
    F13 = 0x7C
    F14 = 0x7D
    F15 = 0x7E
    F16 = 0x7F
    F17 = 0x80
    F18 = 0x81
    F19 = 0x82
    F20 = 0x83
    F21 = 0x84
    F22 = 0x85
    F23 = 0x86
    F24 = 0x87
    NUMLOCK = 0x90
    SCROLL = 0x91
    LSHIFT = 0xA0
    RSHIFT = 0xA1
    LCONTROL = 0xA2
    RCONTROL = 0xA3
    LMENU = 0xA4
    RMENU = 0xA5
    BROWSER_BACK = 0xA6
    BROWSER_FORWARD = 0xA7
    BROWSER_REFRESH = 0xA8
    BROWSER_STOP = 0xA9
    BROWSER_SEARCH = 0xAA
    BROWSER_FAVORITES = 0xAB
    BROWSER_HOME = 0xAC
    VOLUME_MUTE = 0xAD
    VOLUME_DOWN = 0xAE
    VOLUME_UP = 0xAF
    MEDIA_NEXT_TRACK = 0xB0
    MEDIA_PREV_TRACK = 0xB1
    MEDIA_STOP = 0xB2
    MEDIA_PLAY_PAUSE = 0xB3
    LAUNCH_MAIL = 0xB4
    LAUNCH_MEDIA_SELECT = 0xB5
    LAUNCH_APP1 = 0xB6
    LAUNCH_APP2 = 0xB7
    OEM_1 = 0xBA
    OEM_PLUS = 0xBB
    OEM_COMMA = 0xBC
    OEM_MINUS = 0xBD
    OEM_PERIOD = 0xBE
    OEM_2 = 0xBF
    OEM_3 = 0xC0
    OEM_4 = 0xDB
    OEM_5 = 0xDC
    OEM_6 = 0xDD
    OEM_7 = 0xDE
    OEM_8 = 0xDF
    OEM_102 = 0xE2
    PROCESSKEY = 0xE5
    PACKET = 0xE7
    ATTN = 0xF6
    CRSEL = 0xF7
    EXSEL = 0xF8
    EREOF = 0xF9
    PLAY = 0xFA
    ZOOM = 0xFB
    NONAME = 0xFC
    PA1 = 0xFD
    OEM_CLEAR = 0xFE
)

const (
    // LINUX_LBUTTON = 0x01 
    // LINUX_RBUTTON = 0x02
    // LINUX_CANCEL = 0x03
    // LINUX_MBUTTON = 0x04
    // LINUX_XBUTTON1 = 0x05
    // LINUX_XBUTTON2 = 0x06
    LINUX_BACK = 14
    LINUX_TAB = 15
    LINUX_CLEAR = 0x163
    LINUX_RETURN = 28
    LINUX_SHIFT = 0x10
    LINUX_CONTROL = 0x11
    LINUX_MENU = 0x12
    LINUX_PAUSE = 42
    LINUX_CAPITAL = 58
    LINUX_KANA = 93
    LINUX_HANGUL = 93
    LINUX_JUNJA = 123
    // LINUX_FINAL = 0x18
    LINUX_HANJA = 123
    LINUX_KANJI = 90
    LINUX_ESCAPE = 1
    // LINUX_CONVERT = 0x1C
    // LINUX_NONCONVERT = 0x1D
    // LINUX_ACCEPT = 0x1E
    // LINUX_MODECHANGE = 0x1F
    LINUX_SPACE = 57
    LINUX_PRIOR = 104
    LINUX_NEXT = 109
    LINUX_END = 107
    LINUX_HOME = 102
    LINUX_LEFT = 105
    LINUX_UP = 103
    LINUX_RIGHT = 106
    LINUX_DOWN = 108
    LINUX_SELECT = 0x161
    LINUX_PRINT = 210	/* AC Print */
    // LINUX_EXECUTE = 0x2B
    LINUX_SNAPSHOT = 99
    LINUX_INSERT = 110
    LINUX_DELETE = 111
    LINUX_HELP = 138
    LINUX_KEY_0 = 11
    LINUX_KEY_1 = 2
    LINUX_KEY_2 = 3
    LINUX_KEY_3 = 4
    LINUX_KEY_4 = 5
    LINUX_KEY_5 = 6
    LINUX_KEY_6 = 7
    LINUX_KEY_7 = 8
    LINUX_KEY_8 = 9
    LINUX_KEY_9 = 10
    LINUX_KEY_A = 30
    LINUX_KEY_B = 48
    LINUX_KEY_C = 46
    LINUX_KEY_D = 32
    LINUX_KEY_E = 18
    LINUX_KEY_F = 33
    LINUX_KEY_G = 34
    LINUX_KEY_H = 35
    LINUX_KEY_I = 23
    LINUX_KEY_J = 36
    LINUX_KEY_K = 37
    LINUX_KEY_L = 38
    LINUX_KEY_M = 50
    LINUX_KEY_N = 49
    LINUX_KEY_O = 24
    LINUX_KEY_P = 25
    LINUX_KEY_Q = 16
    LINUX_KEY_R = 19
    LINUX_KEY_S = 31
    LINUX_KEY_T = 20
    LINUX_KEY_U = 22
    LINUX_KEY_V = 47
    LINUX_KEY_W = 17
    LINUX_KEY_X = 45
    LINUX_KEY_Y = 21
    LINUX_KEY_Z = 44
    LINUX_LWIN = 125
    LINUX_RWIN = 126
    // LINUX_APPS = 0x5D
    LINUX_SLEEP = 142 /* SC System Sleep */
    LINUX_NUMPAD0 = 82
    LINUX_NUMPAD1 = 79
    LINUX_NUMPAD2 = 80
    LINUX_NUMPAD3 = 81
    LINUX_NUMPAD4 = 75
    LINUX_NUMPAD5 = 76
    LINUX_NUMPAD6 = 77
    LINUX_NUMPAD7 = 71
    LINUX_NUMPAD8 = 72
    LINUX_NUMPAD9 = 73
    LINUX_MULTIPLY = 55 /* KEY_KPASTERISK */
    LINUX_ADD = 78
    LINUX_SEPARATOR = 121 /* KEY_KPCOMMA */
    LINUX_SUBTRACT = 74 /* KEY_KPMINUS */
    LINUX_DECIMAL = 83
    LINUX_DIVIDE = 98
    LINUX_F1 = 59
    LINUX_F2 = 60
    LINUX_F3 = 61
    LINUX_F4 = 62
    LINUX_F5 = 63
    LINUX_F6 = 64
    LINUX_F7 = 65
    LINUX_F8 = 66
    LINUX_F9 = 67
    LINUX_F10 = 68
    LINUX_F11 = 87
    LINUX_F12 = 88
    LINUX_F13 = 183
    LINUX_F14 = 184
    LINUX_F15 = 185
    LINUX_F16 = 186
    LINUX_F17 = 187
    LINUX_F18 = 188
    LINUX_F19 = 189
    LINUX_F20 = 190
    LINUX_F21 = 191
    LINUX_F22 = 192
    LINUX_F23 = 193
    LINUX_F24 = 194
    LINUX_NUMLOCK = 69
    LINUX_SCROLL = 70
    LINUX_LSHIFT = 42
    LINUX_RSHIFT = 54
    LINUX_LCONTROL = 29
    LINUX_RCONTROL = 97
    LINUX_LMENU = 56


    LINUX_RMENU = 100
    // LINUX_BROWSER_BACK = 0xA6
    // LINUX_BROWSER_FORWARD = 0xA7
    // LINUX_BROWSER_REFRESH = 0xA8
    // LINUX_BROWSER_STOP = 0xA9
    // LINUX_BROWSER_SEARCH = 0xAA
    // LINUX_BROWSER_FAVORITES = 0xAB
    // LINUX_BROWSER_HOME = 0xAC
    // LINUX_VOLUME_MUTE = 0xAD
    // LINUX_VOLUME_DOWN = 0xAE
    // LINUX_VOLUME_UP = 0xAF
    // LINUX_MEDIA_NEXT_TRACK = 0xB0
    // LINUX_MEDIA_PREV_TRACK = 0xB1
    LINUX_MEDIA_STOP = 0xB2
    LINUX_MEDIA_PLAY_PAUSE = 0xB3
    LINUX_LAUNCH_MAIL = 0xB4
    LINUX_LAUNCH_MEDIA_SELECT = 0xB5
    LINUX_LAUNCH_APP1 = 0xB6
    LINUX_LAUNCH_APP2 = 0xB7
    LINUX_OEM_1 = 39 /* KEY_SEMICOLON */
    LINUX_OEM_PLUS = 13 /*KEY_EQUAL*/
    LINUX_OEM_COMMA = 51
    LINUX_OEM_MINUS = 12
    LINUX_OEM_PERIOD = 52
    LINUX_OEM_2 = 53
    LINUX_OEM_3 = 41
    LINUX_OEM_4 = 26
    LINUX_OEM_5 = 43
    LINUX_OEM_6 = 27
    LINUX_OEM_7 = 40
    // LINUX_OEM_8 = 0xDF
    LINUX_OEM_102 = 86
    // LINUX_PROCESSKEY = 0xE5
    // LINUX_PACKET = 0xE7
    // LINUX_ATTN = 0xF6
    // LINUX_CRSEL = 0xF7
    // LINUX_EXSEL = 0xF8
    // LINUX_EREOF = 0xF9
    // LINUX_PLAY = 0xFA
    // LINUX_ZOOM = 0xFB
    // LINUX_NONAME = 0xFC
    // LINUX_PA1 = 0xFD
    // LINUX_OEM_CLEAR = 0xFE
)



var (
	extendkeys = []int{    
        LWIN,
        RWIN,
        RMENU,
        RCONTROL,
        INSERT,
        DELETE,
        HOME,
        END,
        PRIOR,
        NEXT,
        UP,
        DOWN,
        LEFT,
        RIGHT,
        DIVIDE,
	}
    positionkeys = []int{    
        LSHIFT,
        RSHIFT,
        LCONTROL,
        RCONTROL,
        LMENU,
        RMENU,
    }
    scankeys = []int{
        KEY_0,
        KEY_1,
        KEY_2,
        KEY_3,
        KEY_4,
        KEY_5,
        KEY_6,
        KEY_7,
        KEY_8,
        KEY_9,
        KEY_A,
        KEY_B,
        KEY_C,
        KEY_D,
        KEY_E,
        KEY_F,
        KEY_G,
        KEY_H,
        KEY_I,
        KEY_J,
        KEY_K,
        KEY_L,
        KEY_M,
        KEY_N,
        KEY_O,
        KEY_P,
        KEY_Q,
        KEY_R,
        KEY_S,
        KEY_T,
        KEY_U,
        KEY_V,
        KEY_W,
        KEY_X,
        KEY_Y,
        KEY_Z,

        SPACE,
        ESCAPE,

        UP,
        DOWN,
        LEFT,
        RIGHT,
        RETURN,
    };
)

func ExtendedFlag(code int) int {
	for _,v := range extendkeys {
		if code == v {
			return 1
		}
	}

	return 0
}
