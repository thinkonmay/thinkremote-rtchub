package hid

/*
#include <Windows.h>
*/
import "C"

const (
    LBUTTON = 0x01
    RBUTTON = 0x02
    //Control-break processing
    CANCEL = 0x03
    //Middle mouse button (three-button mouse)
    MBUTTON = 0x04
    //Windows 2000/XP: X1 mouse button
    XBUTTON1 = 0x05
    //Windows 2000/XP: X2 mouse button
    XBUTTON2 = 0x06
    //BACKSPACE key
    BACK = 0x08
    //TAB key
    TAB = 0x09
    //CLEAR key
    CLEAR = 0x0C
    //ENTER key
    RETURN = 0x0D
    //SHIFT key
    SHIFT = 0x10
    //CTRL key
    CONTROL = 0x11
    //ALT key
    MENU = 0x12
    //PAUSE key
    PAUSE = 0x13
    //CAPS LOCK key
    CAPITAL = 0x14
    //Input Method Editor (IME) Kana mode
    KANA = 0x15
    //IME Hangul mode
    HANGUL = 0x15
    //IME Junja mode
    JUNJA = 0x17
    //IME final mode
    FINAL = 0x18
    //IME Hanja mode
    HANJA = 0x19
    
    //IME Kanji mode
    KANJI = 0x19
    
    //ESC key
    ESCAPE = 0x1B
    
    //IME convert
    CONVERT = 0x1C
    
    //IME nonconvert
    NONCONVERT = 0x1D
    
    //IME accept
    ACCEPT = 0x1E
    
    //IME mode change request
    MODECHANGE = 0x1F
    
    //SPACEBAR
    SPACE = 0x20
    
    //PAGE UP key
    PRIOR = 0x21
    //PAGE DOWN key
    NEXT = 0x22
    
    END = 0x23
    
    HOME = 0x24
    LEFT = 0x25
    UP = 0x26
    RIGHT = 0x27
    DOWN = 0x28
    SELECT = 0x29
    PRINT = 0x2A
    
    //EXECUTE key
    EXECUTE = 0x2B
    
    //PRINT SCREEN key
    SNAPSHOT = 0x2C
    
    //INS key

    INSERT = 0x2D
    
    //DEL key

    DELETE = 0x2E
    
    //HELP key

    HELP = 0x2F
    
    //0 key

    KEY_0 = 0x30
    
    //1 key

    KEY_1 = 0x31
    
    //2 key

    KEY_2 = 0x32
    
    //3 key

    KEY_3 = 0x33
    
    //4 key

    KEY_4 = 0x34
    
    //5 key

    KEY_5 = 0x35
    
    //6 key

    KEY_6 = 0x36
    
    //7 key

    KEY_7 = 0x37
    
    //8 key

    KEY_8 = 0x38
    
    //9 key

    KEY_9 = 0x39
    
    //A key

    KEY_A = 0x41
    
    //B key

    KEY_B = 0x42
    
    //C key

    KEY_C = 0x43
    
    //D key

    KEY_D = 0x44
    
    //E key

    KEY_E = 0x45
    
    //F key

    KEY_F = 0x46
    
    //G key

    KEY_G = 0x47
    
    //H key

    KEY_H = 0x48
    
    //I key

    KEY_I = 0x49
    
    //J key

    KEY_J = 0x4A
    
    //K key

    KEY_K = 0x4B
    
    //L key

    KEY_L = 0x4C
    
    //M key

    KEY_M = 0x4D
    
    //N key

    KEY_N = 0x4E
    
    //O key

    KEY_O = 0x4F
    
    //P key

    KEY_P = 0x50
    
    //Q key

    KEY_Q = 0x51
    
    //R key

    KEY_R = 0x52
    
    //S key

    KEY_S = 0x53
    
    //T key

    KEY_T = 0x54
    
    //U key

    KEY_U = 0x55
    
    //V key

    KEY_V = 0x56
    
    //W key

    KEY_W = 0x57
    
    //X key

    KEY_X = 0x58
    
    //Y key

    KEY_Y = 0x59
    
    //Z key

    KEY_Z = 0x5A
    
    //Left Windows key (Microsoft Natural keyboard) 

    LWIN = 0x5B
    
    //Right Windows key (Natural keyboard)

    RWIN = 0x5C
    
    //Applications key (Natural keyboard)

    APPS = 0x5D
    
    //Computer Sleep key

    SLEEP = 0x5F
    
    //Numeric keypad 0 key

    NUMPAD0 = 0x60
    
    //Numeric keypad 1 key

    NUMPAD1 = 0x61
    
    //Numeric keypad 2 key

    NUMPAD2 = 0x62
    
    //Numeric keypad 3 key

    NUMPAD3 = 0x63
    
    //Numeric keypad 4 key

    NUMPAD4 = 0x64
    
    //Numeric keypad 5 key

    NUMPAD5 = 0x65
    
    //Numeric keypad 6 key

    NUMPAD6 = 0x66
    
    //Numeric keypad 7 key

    NUMPAD7 = 0x67
    
    //Numeric keypad 8 key

    NUMPAD8 = 0x68
    
    //Numeric keypad 9 key

    NUMPAD9 = 0x69
    
    //Multiply key

    MULTIPLY = 0x6A
    
    //Add key

    ADD = 0x6B
    
    //Separator key

    SEPARATOR = 0x6C
    
    //Subtract key

    SUBTRACT = 0x6D
    
    //Decimal key

    DECIMAL = 0x6E
    
    //Divide key

    DIVIDE = 0x6F
    
    //F1 key

    F1 = 0x70
    
    //F2 key

    F2 = 0x71
    
    //F3 key

    F3 = 0x72
    
    //F4 key

    F4 = 0x73
    
    //F5 key

    F5 = 0x74
    
    //F6 key

    F6 = 0x75
    
    //F7 key

    F7 = 0x76
    
    //F8 key

    F8 = 0x77
    
    //F9 key

    F9 = 0x78
    
    //F10 key

    F10 = 0x79
    
    //F11 key

    F11 = 0x7A
    
    //F12 key

    F12 = 0x7B
    
    //F13 key

    F13 = 0x7C
    
    //F14 key

    F14 = 0x7D
    
    //F15 key

    F15 = 0x7E
    
    //F16 key

    F16 = 0x7F
    
    //F17 key  

    F17 = 0x80
    
    //F18 key  

    F18 = 0x81
    
    //F19 key  

    F19 = 0x82
    
    //F20 key  

    F20 = 0x83
    
    //F21 key  

    F21 = 0x84
    
    //F22 key (PPC only) Key used to lock device.

    F22 = 0x85
    
    //F23 key  

    F23 = 0x86
    
    //F24 key  

    F24 = 0x87
    
    //NUM LOCK key

    NUMLOCK = 0x90
    
    //SCROLL LOCK key

    SCROLL = 0x91
    
    //Left SHIFT key

    LSHIFT = 0xA0
    
    //Right SHIFT key

    RSHIFT = 0xA1
    
    //Left CONTROL key

    LCONTROL = 0xA2
    
    //Right CONTROL key

    RCONTROL = 0xA3
    
    //Left MENU key

    LMENU = 0xA4
    
    //Right MENU key

    RMENU = 0xA5
    
    //Windows 2000/XP: Browser Back key

    BROWSER_BACK = 0xA6
    
    //Windows 2000/XP: Browser Forward key

    BROWSER_FORWARD = 0xA7
    
    //Windows 2000/XP: Browser Refresh key

    BROWSER_REFRESH = 0xA8
    
    //Windows 2000/XP: Browser Stop key

    BROWSER_STOP = 0xA9
    
    //Windows 2000/XP: Browser Search key 

    BROWSER_SEARCH = 0xAA
    
    //Windows 2000/XP: Browser Favorites key

    BROWSER_FAVORITES = 0xAB
    
    //Windows 2000/XP: Browser Start and Home key

    BROWSER_HOME = 0xAC
    
    //Windows 2000/XP: Volume Mute key

    VOLUME_MUTE = 0xAD
    
    //Windows 2000/XP: Volume Down key

    VOLUME_DOWN = 0xAE
    
    //Windows 2000/XP: Volume Up key

    VOLUME_UP = 0xAF
    
    //Windows 2000/XP: Next Track key

    MEDIA_NEXT_TRACK = 0xB0
    
    //Windows 2000/XP: Previous Track key

    MEDIA_PREV_TRACK = 0xB1
    
    //Windows 2000/XP: Stop Media key

    MEDIA_STOP = 0xB2
    
    //Windows 2000/XP: Play/Pause Media key

    MEDIA_PLAY_PAUSE = 0xB3
    
    //Windows 2000/XP: Start Mail key

    LAUNCH_MAIL = 0xB4
    
    //Windows 2000/XP: Select Media key

    LAUNCH_MEDIA_SELECT = 0xB5
    
    //Windows 2000/XP: Start Application 1 key

    LAUNCH_APP1 = 0xB6
    
    //Windows 2000/XP: Start Application 2 key

    LAUNCH_APP2 = 0xB7
    
    //Used for miscellaneous characters; it can vary by keyboard.

    OEM_1 = 0xBA
    
    //Windows 2000/XP: For any country/region the '+' key

    OEM_PLUS = 0xBB
    
    //Windows 2000/XP: For any country/region the '' key

    OEM_COMMA = 0xBC
    
    //Windows 2000/XP: For any country/region the '-' key

    OEM_MINUS = 0xBD
    
    //Windows 2000/XP: For any country/region the '.' key

    OEM_PERIOD = 0xBE
    
    //Used for miscellaneous characters; it can vary by keyboard.

    OEM_2 = 0xBF
    
    //Used for miscellaneous characters; it can vary by keyboard. 

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
func ConvertJavaScriptKeyToVirtualKey(key string) (int) {
    switch key {
    case "Down":
      return DOWN
    case "ArrowDown":
      return DOWN
    case "Up":
      return UP
    case "ArrowUp":
      return UP
    case "Left":
      return LEFT
    case "ArrowLeft":
      return LEFT
    case "Right": 
      return RIGHT
    case "ArrowRight":
      return RIGHT
    case "Enter":
      return RETURN
    case "Esc":
      return ESCAPE
	  case "Escape":
      return ESCAPE
    case "Alt":
      return MENU
    case "Control":
      return CONTROL
    case "Shift":
      return SHIFT
    case "PAUSE":
      return PAUSE
    case "BREAK":
      return PAUSE
    case "Backspace":
      return BACK
    case "Tab":
      return TAB
    case "CapsLock":
      return CAPITAL
    case "Delete":
      return DELETE
    case "Home":
      return HOME
    case "End":
      return END
    case "PageUp":
      return PRIOR
    case "PageDown":
      return NEXT
    case "NumLock":
      return NUMLOCK
    case "Insert":
      return INSERT
    case "ScrollLock":
      return SCROLL
    case "F1":
      return F1
    case "F2":
      return F2
    case "F3":
      return F3
    case "F4":
      return F4
    case "F5":
      return F5
    case "F6":
      return F6
    case "F7":
      return F7
    case "F8":
      return F8
    case "F9":
      return F9
    case "F10":
      return F10
    case "F11":
      return F11
    case "F12":
      return F12
    case "Meta":
      return LWIN
    case "ContextMenu":
      return MENU
    }

    if len(key) == 1 {
      return int(C.VkKeyScan(*C.CString(key)))
    }

    return -1
}
