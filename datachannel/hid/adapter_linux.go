package hid


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
    return 0,0,0,0,nil
}

func GetVirtualDisplay() (x, y int) {
    return 0,0
}
