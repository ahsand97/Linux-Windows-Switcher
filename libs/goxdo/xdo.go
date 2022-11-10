package goxdo

//#cgo pkg-config: libxdo
//#include <xdo.h>
import "C"

const (
	CURRENTWINDOW = iota
	MBUTTON_LEFT
	MBUTTON_MIDDLE
	MBUTTON_RIGHT
	MWHEELUP
	MWHEELDOWN
)

type Window int

type Xdo struct {
	xdo *C.xdo_t
}

func NewXdo() *Xdo {
	x := new(Xdo)
	x.xdo = C.xdo_new(nil)
	return x
}

func (t *Xdo) MoveMouse(x, y int, window Window) {
	C.xdo_move_mouse(t.xdo, C.int(x), C.int(y), C.int(window))
}

func (t *Xdo) MoveMouseRelativeToWindow(window Window, x, y int) {
	C.xdo_move_mouse_relative_to_window(t.xdo, C.Window(window), C.int(x), C.int(y))
}

func (t *Xdo) MoveMouseRelative(x, y int) {
	C.xdo_move_mouse_relative(t.xdo, C.int(x), C.int(y))
}

func (t *Xdo) MouseDown(window Window, button int) {
	C.xdo_mouse_down(t.xdo, C.Window(window), C.int(button))
}

func (t *Xdo) MouseUp(window Window, button int) {
	C.xdo_mouse_up(t.xdo, C.Window(window), C.int(button))
}

func (t *Xdo) GetMouseLocation() (int, int, int) {
	x := C.int(0)
	y := C.int(0)
	screen := C.int(0)

	C.xdo_get_mouse_location(t.xdo, &x, &y, &screen)
	return int(x), int(y), int(screen)
}

func (t *Xdo) GetWindowAtMouse() Window {
	var window C.Window
	C.xdo_get_window_at_mouse(t.xdo, &window)
	return Window(window)
}

func (t *Xdo) GetActiveWindow() (int, Window) {
	var window C.Window

	result := C.int(0)
	result = C.xdo_get_active_window(t.xdo, &window)
	return int(result), Window(window)
}

func (t *Xdo) WindowActivate(window Window) int {
	return int(C.xdo_activate_window(t.xdo, C.Window(window)))
}

func (t *Xdo) WaitForWindowActivate(window Window, active bool) int {
	x := C.int(0)
	if active {
		x = C.int(1)
	}
	return int(C.xdo_wait_for_window_active(t.xdo, C.Window(window), x))
}

func (t *Xdo) GetMouseLocation2() (int, int, int, Window) {
	var window C.Window
	x := C.int(0)
	y := C.int(0)
	screen := C.int(0)

	C.xdo_get_mouse_location2(t.xdo, &x, &y, &screen, &window)
	return int(x), int(y), int(screen), Window(window)
}

func (t *Xdo) WaitForMouseMoveFrom(x, y int) {
	C.xdo_wait_for_mouse_move_from(t.xdo, C.int(x), C.int(y))
}

func (t *Xdo) WaitForMouseMoveTo(x, y int) {
	C.xdo_wait_for_mouse_move_to(t.xdo, C.int(x), C.int(y))
}

func (t *Xdo) ClickWindow(window Window, button int) {
	C.xdo_click_window(t.xdo, C.Window(window), C.int(button))
}

func (t *Xdo) ClickWindowMultiple(window Window, button, repeat, useconds int) {
	C.xdo_click_window_multiple(t.xdo, C.Window(window), C.int(button), C.int(repeat),
		C.useconds_t(useconds))
}

func (t *Xdo) EnterTextWindow(window Window, text string, udelay int) {
	C.xdo_enter_text_window(t.xdo, C.Window(window), C.CString(text), C.useconds_t(udelay))
}

func (t *Xdo) SendKeysequenceWindow(window Window, sequence string, udelay int) {
	C.xdo_send_keysequence_window(t.xdo, C.Window(window), C.CString(sequence),
		C.useconds_t(udelay))
}

func (t *Xdo) SendKeysequenceWindowUp(window Window, sequence string, udelay int) {
	C.xdo_send_keysequence_window_up(t.xdo, C.Window(window), C.CString(sequence),
		C.useconds_t(udelay))
}

func (t *Xdo) SendKeysequenceWindowDown(window Window, sequence string, udelay int) {
	C.xdo_send_keysequence_window_down(t.xdo, C.Window(window), C.CString(sequence),
		C.useconds_t(udelay))
}
