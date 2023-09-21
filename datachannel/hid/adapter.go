package hid


/*
#include <Windows.h>
#include "windows.h"
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


unsigned short
convert_javascript_key_to_window_key(char* keycode)
{
        if(!g_strcmp0(keycode,"Backspace"))                 {return     0x08;}
   else if(!g_strcmp0(keycode,"Tab"))                       {return     0x09;}
   else if(!g_strcmp0(keycode,"Enter"))                     {return     0X0D;}
   else if(!g_strcmp0(keycode,"AltRight"))                  {return     0X12;}
   else if(!g_strcmp0(keycode,"AltLeft"))                   {return     0X12;}
   else if(!g_strcmp0(keycode,"Pause"))                     {return     0X13;}
   else if(!g_strcmp0(keycode,"CapsLock"))                  {return     0X14;}
   else if(!g_strcmp0(keycode,"Escape"))                    {return     0x1B;}
   else if(!g_strcmp0(keycode,"Space"))                     {return     0x20;}
   else if(!g_strcmp0(keycode,"PageUp"))                    {return     0x21;}
   else if(!g_strcmp0(keycode,"PageDown"))                  {return     0x22;}
   else if(!g_strcmp0(keycode,"End"))                       {return     0x23;}
   else if(!g_strcmp0(keycode,"Home"))                      {return     0x24;}
   else if(!g_strcmp0(keycode,"ArrowLeft"))                 {return     0x25;}
   else if(!g_strcmp0(keycode,"ArrowUp"))                   {return     0x26;}
   else if(!g_strcmp0(keycode,"ArrowRight"))                {return     0x27;}
   else if(!g_strcmp0(keycode,"ArrowDown"))                 {return     0x28;}
   else if(!g_strcmp0(keycode,"Insert"))                    {return     0x2D;}
   else if(!g_strcmp0(keycode,"Delete"))                    {return     0x2E;}
   else if(!g_strcmp0(keycode,"Digit0"))                    {return     0x30;}
   else if(!g_strcmp0(keycode,"Digit1"))                    {return     0x31;}
   else if(!g_strcmp0(keycode,"Digit2"))                    {return     0x32;}
   else if(!g_strcmp0(keycode,"Digit3"))                    {return     0x33;}
   else if(!g_strcmp0(keycode,"Digit4"))                    {return     0x34;}
   else if(!g_strcmp0(keycode,"Digit5"))                    {return     0x35;}
   else if(!g_strcmp0(keycode,"Digit6"))                    {return     0x36;}
   else if(!g_strcmp0(keycode,"Digit7"))                    {return     0x37;}
   else if(!g_strcmp0(keycode,"Digit8"))                    {return     0x38;}
   else if(!g_strcmp0(keycode,"Digit9"))                    {return     0x39;}
   else if(!g_strcmp0(keycode,"KeyA"))                      {return     0x41;}
   else if(!g_strcmp0(keycode,"KeyB"))                      {return     0x42;}
   else if(!g_strcmp0(keycode,"KeyC"))                      {return     0x43;}
   else if(!g_strcmp0(keycode,"KeyD"))                      {return     0x44;}
   else if(!g_strcmp0(keycode,"KeyE"))                      {return     0x45;}
   else if(!g_strcmp0(keycode,"KeyF"))                      {return     0x46;}
   else if(!g_strcmp0(keycode,"KeyG"))                      {return     0x47;}
   else if(!g_strcmp0(keycode,"KeyH"))                      {return     0x48;}
   else if(!g_strcmp0(keycode,"KeyI"))                      {return     0x49;}
   else if(!g_strcmp0(keycode,"KeyJ"))                      {return     0x4A;}
   else if(!g_strcmp0(keycode,"KeyK"))                      {return     0x4B;}
   else if(!g_strcmp0(keycode,"KeyL"))                      {return     0x4C;}
   else if(!g_strcmp0(keycode,"KeyM"))                      {return     0x4D;}
   else if(!g_strcmp0(keycode,"KeyN"))                      {return     0x4E;}
   else if(!g_strcmp0(keycode,"KeyO"))                      {return     0x4F;}
   else if(!g_strcmp0(keycode,"KeyP"))                      {return     0x50;}
   else if(!g_strcmp0(keycode,"KeyQ"))                      {return     0x51;}
   else if(!g_strcmp0(keycode,"KeyR"))                      {return     0x52;}
   else if(!g_strcmp0(keycode,"KeyS"))                      {return     0x53;}
   else if(!g_strcmp0(keycode,"KeyT"))                      {return     0x54;}
   else if(!g_strcmp0(keycode,"KeyU"))                      {return     0x55;}
   else if(!g_strcmp0(keycode,"KeyV"))                      {return     0x56;}
   else if(!g_strcmp0(keycode,"KeyW"))                      {return     0x57;}
   else if(!g_strcmp0(keycode,"KeyX"))                      {return     0x58;}
   else if(!g_strcmp0(keycode,"KeyY"))                      {return     0x59;}
   else if(!g_strcmp0(keycode,"KeyZ"))                      {return     0x5A;}
   else if(!g_strcmp0(keycode,"MetaLeft"))                  {return     0x5B;}
   else if(!g_strcmp0(keycode,"F1"))                        {return     0x70;}
   else if(!g_strcmp0(keycode,"F2"))                        {return     0x71;}
   else if(!g_strcmp0(keycode,"F2"))                        {return     0x72;}
   else if(!g_strcmp0(keycode,"F4"))                        {return     0x73;}
   else if(!g_strcmp0(keycode,"F5"))                        {return     0x74;}
   else if(!g_strcmp0(keycode,"F6"))                        {return     0x75;}
   else if(!g_strcmp0(keycode,"F7"))                        {return     0x76;}
   else if(!g_strcmp0(keycode,"F8"))                        {return     0x77;}
   else if(!g_strcmp0(keycode,"F9"))                        {return     0x78;}
   else if(!g_strcmp0(keycode,"F10"))                       {return     0x79;}
   else if(!g_strcmp0(keycode,"F11"))                       {return     0x7A;}
   else if(!g_strcmp0(keycode,"F12"))                       {return     0x7B;}
   else if(!g_strcmp0(keycode,"ScrollLock"))                {return     0x91;}
   else if(!g_strcmp0(keycode,"ShiftLeft"))                 {return     0xA0;}
   else if(!g_strcmp0(keycode,"ShiftRight"))                {return     0xA1;}
   else if(!g_strcmp0(keycode,"ControlLeft"))               {return     0xA2;}
   else if(!g_strcmp0(keycode,"ControlRight"))              {return     0xA3;}
   else if(!g_strcmp0(keycode,"ContextMenu"))               {return     0xA4;}
   else if(!g_strcmp0(keycode,"Semicolon"))                 {return     0xBA;}
   else if(!g_strcmp0(keycode,"Equal"))                     {return     0xBB;}
   else if(!g_strcmp0(keycode,"Comma"))                     {return     0xBC;}
   else if(!g_strcmp0(keycode,"Minus"))                     {return     0xBD;}
   else if(!g_strcmp0(keycode,"Period"))                    {return     0xBE;}
   else if(!g_strcmp0(keycode,"Slash"))                     {return     0xBF;}
   else if(!g_strcmp0(keycode,"Backquote"))                 {return     0xC0;}
   else if(!g_strcmp0(keycode,"BracketLeft"))               {return     0xDB;}
   else if(!g_strcmp0(keycode,"Backslash"))                 {return     0xDC;}
   else if(!g_strcmp0(keycode,"BracketRight"))              {return     0xDD;}
   else
       return 0;
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

    SendInput(1, &window_input, sizeof(window_input));
}



void
handle_keyboard_javascript(int opcode,
                           int key)
{
    INPUT window_input;
    memset(&window_input,0, sizeof(window_input));

    if(opcode == KEYUP || opcode == KEYDOWN)
    {
        window_input.type = INPUT_KEYBOARD;
        window_input.ki.time = 0;
        window_input.ki.wVk = key;
    }


    if (opcode == KEYUP)
        window_input.ki.dwFlags = KEYEVENTF_KEYUP | KEYEVENTF_EXTENDEDKEY;
    else if (opcode == KEYDOWN)
        window_input.ki.dwFlags = KEYEVENTF_EXTENDEDKEY;

    SendInput(1, &window_input, sizeof(window_input));
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

func SendKeyboard(key string, is_up bool) {
    code := C.KEYUP
    if !is_up { code = C.KEYDOWN }
    C.handle_keyboard_javascript(
        C.int(code),
        C.int(ConvertJavaScriptKeyToVirtualKey(key)),
    )
}

func SetClipboard(text string) {
    C.SetClipboard(C.CString(text))
}