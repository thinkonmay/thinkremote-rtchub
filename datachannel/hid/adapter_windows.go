package hid

/*
#include <Windows.h>
#include "windows.h"
#include "winuser.h"
#include "string.h"
#include <direct.h>

typedef enum
{
	KEYUP = 100,
	KEYDOWN,

	MOUSE_WHEEL,
	MOUSE_MOVE,
	MOUSE_UP,
	MOUSE_DOWN,
}JavaScriptOpcode;

HDESK _lastKnownInputDesktop = NULL;


int DisplayPosition(char* display_name, int* x, int* y, int* width, int* height) {
	HRESULT result = 1;
	int deviceIndex = 0;
	do
	{
		DISPLAY_DEVICEA dpd = {0};
		PDISPLAY_DEVICEA displayDevice = &dpd;
		displayDevice->cb = sizeof(DISPLAY_DEVICEA);

		result = EnumDisplayDevicesA(NULL,
			deviceIndex++, displayDevice, 0);
        if ((displayDevice->StateFlags & DISPLAY_DEVICE_ACTIVE) &&
			 !strcmp(display_name,displayDevice->DeviceName)) {

			DEVMODEA dm = {};
			if (!EnumDisplaySettingsA(displayDevice->DeviceName, ENUM_CURRENT_SETTINGS, &dm) )
				continue;

            *x = dm.dmPosition.x;
            *y = dm.dmPosition.y;
            *width  = dm.dmPelsWidth;
			*height = dm.dmPelsHeight;
            return 1;
		}
	} while (result);
    return 0;
}

HDESK
syncThreadDesktop() {
    HDESK hDesk = OpenInputDesktop(DF_ALLOWOTHERACCOUNTHOOK, FALSE, GENERIC_ALL);
    if (!hDesk) {
        return NULL;
    }

    if (!SetThreadDesktop(hDesk)) {
        return NULL;
    }

    CloseDesktop(hDesk);
    return hDesk;
}

void
handle_mouse_javascript(int opcode,
                        int button,
                        float dX,
                        float dY,
                        float wheel,
						int relative_mouse)
{
    INPUT window_input;
    memset(&window_input,0, sizeof(window_input));

    if(opcode == MOUSE_DOWN || opcode == MOUSE_UP || opcode == MOUSE_MOVE)
    {
        window_input.type = INPUT_MOUSE;
        window_input.mi.mouseData = 0;
        window_input.mi.time = 0;
        window_input.mi.dx = dX * (!relative_mouse ? 65535 : 1 );
        window_input.mi.dy = dY * (!relative_mouse ? 65535 : 1 );
    }


    if (opcode == MOUSE_UP)
    {
        if(relative_mouse)
        {
            if(button == 0)
                window_input.mi.dwFlags =  MOUSEEVENTF_LEFTUP | MOUSEEVENTF_VIRTUALDESK;
            else if(button == 1)
                window_input.mi.dwFlags =  MOUSEEVENTF_MIDDLEUP | MOUSEEVENTF_VIRTUALDESK;
            else if (button == 2)
                window_input.mi.dwFlags =  MOUSEEVENTF_RIGHTUP | MOUSEEVENTF_VIRTUALDESK;
        }
        else
        {
            if(button == 0)
                window_input.mi.dwFlags = MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_LEFTUP | MOUSEEVENTF_VIRTUALDESK;
            else if(button == 1)
                window_input.mi.dwFlags = MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_MIDDLEUP | MOUSEEVENTF_VIRTUALDESK;
            else if (button == 2)
                window_input.mi.dwFlags = MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_RIGHTUP | MOUSEEVENTF_VIRTUALDESK;
        }
    }
    else if (opcode == MOUSE_DOWN)
    {
        if(relative_mouse)
        {
            if(button == 0)
                window_input.mi.dwFlags =  MOUSEEVENTF_LEFTDOWN | MOUSEEVENTF_VIRTUALDESK;
            else if(button == 1)
                window_input.mi.dwFlags =  MOUSEEVENTF_MIDDLEDOWN | MOUSEEVENTF_VIRTUALDESK;
            else if (button == 2)
                window_input.mi.dwFlags =  MOUSEEVENTF_RIGHTDOWN | MOUSEEVENTF_VIRTUALDESK;
        }
        else
        {
            if(button == 0)
                window_input.mi.dwFlags = MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_LEFTDOWN | MOUSEEVENTF_VIRTUALDESK;
            else if(button == 1)
                window_input.mi.dwFlags = MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_MIDDLEDOWN | MOUSEEVENTF_VIRTUALDESK;
            else if (button == 2)
                window_input.mi.dwFlags = MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_RIGHTDOWN | MOUSEEVENTF_VIRTUALDESK;
        }
    }
    else if (opcode == MOUSE_MOVE)
    {
        if(relative_mouse)
            window_input.mi.dwFlags = MOUSEEVENTF_MOVE ;
        else
            window_input.mi.dwFlags = MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_MOVE | MOUSEEVENTF_VIRTUALDESK;
    }
    else if(opcode == MOUSE_WHEEL)
    {
        window_input.mi.dwFlags = MOUSEEVENTF_WHEEL;
        window_input.mi.mouseData = wheel;
    }

    UINT send;
    retry:
    send = SendInput(1, &window_input, sizeof(window_input));
    if (send != 1) {
        HDESK hDesk = syncThreadDesktop();
        if (_lastKnownInputDesktop != hDesk) {
            _lastKnownInputDesktop = hDesk;
            goto retry;
        }
    }
}




void
handle_keyboard_javascript(int opcode,
                           int key,
                           int extended,
                           int scankey)
{
    UINT send;
    INPUT window_input = {0};
    window_input.type = INPUT_KEYBOARD;

    if(scankey) {
        window_input.ki.wScan   = MapVirtualKeyEx(key, MAPVK_VK_TO_VSC, LoadKeyboardLayoutA("00000409", 0));
        window_input.ki.dwFlags = KEYEVENTF_SCANCODE;
    } else {
        window_input.ki.wVk = key;
    }


    if (extended)
        window_input.ki.dwFlags |= KEYEVENTF_EXTENDEDKEY;
    if (opcode == KEYUP)
        window_input.ki.dwFlags |= KEYEVENTF_KEYUP ;


    retry:
    send = SendInput(1, &window_input, sizeof(window_input));
    if (send != 1) {
        HDESK hDesk = syncThreadDesktop();
        if (_lastKnownInputDesktop != hDesk) {
            _lastKnownInputDesktop = hDesk;
            goto retry;
        }
    }
}


void
SetClipboard(char* output) {
    const size_t len = strlen(output) + 1;
    HGLOBAL hMem =  GlobalAlloc(GMEM_MOVEABLE, len);
    memcpy(GlobalLock(hMem), output, len);
    GlobalUnlock(hMem);
    OpenClipboard(0);
    EmptyClipboard();
    SetClipboardData(CF_TEXT, hMem);
    CloseClipboard();
}

*/
import "C"
import "fmt"

func init() {
	C.syncThreadDesktop()
}

func SendMouseRelative(x float32, y float32) {
	C.handle_mouse_javascript(
		C.MOUSE_MOVE,
		0,
		C.float(x),
		C.float(y),
		0,
		1,
	)
}

func SendMouseAbsolute(wx, wy, lx, ly float32) {
	C.handle_mouse_javascript(
		C.MOUSE_MOVE,
		0,
		C.float(wx),
		C.float(wy),
		0,
		0,
	)
}

func SendMouseWheel(wheel float64) {
	C.handle_mouse_javascript(
		C.MOUSE_WHEEL,
		0,
		0,
		0,
		C.float(wheel),
		0,
	)
}

func SendMouseButton(button int, is_up bool) {
	code := C.MOUSE_UP
	if !is_up {
		code = C.MOUSE_DOWN
	}
	C.handle_mouse_javascript(
		C.int(code),
		C.int(button),
		0,
		0,
		0,
		0,
	)
}

func SendKeyboard(keycode int,
	is_up bool,
	scan_code bool) {
	code := C.KEYUP
	if !is_up {
		code = C.KEYDOWN
	}

	scankey := 0
	if scan_code {
		scankey = 1
	}
	C.handle_keyboard_javascript(
		C.int(code),
		C.int(keycode),
		C.int(ExtendedFlag(keycode)),
		C.int(scankey),
	)
}

func SetClipboard(text string) {
	C.SetClipboard(C.CString(text))
}

func DisplayPosition(name string) (x, y, width, height int, err error) {
	a, b, c, d := C.int(0), C.int(0), C.int(0), C.int(0)
	if C.DisplayPosition(C.CString(name), &a, &b, &c, &d) > 0 {
		x, y, width, height = int(a), int(b), int(c), int(d)
		return
	}

	err = fmt.Errorf("")
	return
}

func GetVirtualDisplay() (x, y int) {
	return int(C.GetSystemMetrics(C.SM_CXVIRTUALSCREEN)),int(C.GetSystemMetrics(C.SM_CYVIRTUALSCREEN))
}
