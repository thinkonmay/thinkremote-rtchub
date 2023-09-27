package hid

/*
#include <Windows.h>
#include "windows.h"
#include "winuser.h"
#include "string.h"
#include <direct.h>
#include <glib.h>


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
            window_input.mi.dwFlags = MOUSEEVENTF_MOVE | MOUSEEVENTF_VIRTUALDESK;
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
                           int lrkey)
{
    INPUT window_input;
    memset(&window_input,0, sizeof(window_input));

    if(opcode == KEYUP || opcode == KEYDOWN)
    {
        window_input.type = INPUT_KEYBOARD;
        window_input.ki.time  = 0;
        window_input.ki.wVk = key;
        window_input.ki.dwExtraInfo = GetMessageExtraInfo();
    }


    if (opcode == KEYUP)
        window_input.ki.dwFlags = KEYEVENTF_KEYUP ;
    else if (opcode == KEYDOWN)
        window_input.ki.dwFlags = 0;

    if (extended)
        window_input.ki.dwFlags |= KEYEVENTF_EXTENDEDKEY;

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

#cgo pkg-config: glib-2.0
*/
import "C"


func init() {
    C.syncThreadDesktop()
}


func SendMouseRelative(x float32,y float32) {
    C.handle_mouse_javascript(
        C.MOUSE_MOVE,
        0,
        C.float(x),
        C.float(y),
        0,
        1,
    )
}

func SendMouseAbsolute(x float32,y float32) {
    C.handle_mouse_javascript(
        C.MOUSE_MOVE,
        0,
        C.float(x),
        C.float(y),
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
    if !is_up { code = C.MOUSE_DOWN }
    C.handle_mouse_javascript(
        C.int(code),
        C.int(button),
        0,
        0,
        0,
        0,
    )
}

func SendKeyboard(keycode int, is_up bool) {
    code := C.KEYUP
    if !is_up { code = C.KEYDOWN }
    C.handle_keyboard_javascript(
        C.int(code),
        C.int(keycode),
        C.int(ExtendedFlag(keycode)),
        C.int(LRKey(keycode)),
    )
}

func SetClipboard(text string) {
    C.SetClipboard(C.CString(text))
}