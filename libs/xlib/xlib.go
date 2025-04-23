package xlib

//#cgo pkg-config: x11
//#include <X11/Xlib.h>
//#include <X11/extensions/XTest.h>
//
//int screen_count(Display *display) {
//    return ScreenCount(display);
//}
//
//Screen *screen_of_display(Display *display, int index) {
//    return ScreenOfDisplay(display, index);
//}
//
//Window default_root_window(Display *display) {
//    return DefaultRootWindow(display);
//}
import "C"

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
)

type (
	Window  = C.Window
	Display = C.Display
	Button  = int
)

var display *Display

const (
	CURRENTWINDOW Window = iota
	MBUTTON_LEFT  Button = iota
	MBUTTON_MIDDLE
	MBUTTON_RIGHT
	MWHEELUP
	MWHEELDOWN
)

// PropertyResult Struct representing the result of XGetWindowProperty call.
type PropertyResult struct {
	Format         int     // Format of result, it can be 8 (char *), 16 (short *) or 32 (long *)
	TypeOfProperty string  // Actual type (name) of the property
	NumberOfItems  uint64  // Number of total items returned by XGetWindowProperty
	charResult     []int8  // Result of casting buffer from uchar * to a slice of int8 (char *). This slice is null if Format is not 8
	shortResult    []int16 // Result of casting buffer from uchar * to a slice of int16 (short *). This slice is null if Format is not 16
	longResult     []int64 // Result of casting buffer from uchar * to a slice of int64 (long *). This slice is null if Format is not 32
}

// GetChar Get a slice of []int8 (char *) from result.
func (propertyResult *PropertyResult) GetChar() []int8 {
	if propertyResult.Format == 8 {
		return propertyResult.charResult
	}
	return nil
}

// GetString Get a slice of []string from result.
func (propertyResult *PropertyResult) GetString() []string {
	if propertyResult.Format != 8 {
		return nil
	}
	var words []string
	sliceOfSliceOfChars := [][]int8{{}}
	i := 0
	for _, char := range propertyResult.charResult {
		if char == 0 { // When 0 found it means end of word
			i++
			sliceOfSliceOfChars = append(sliceOfSliceOfChars, []int8{})
			continue
		}
		sliceOfSliceOfChars[i] = append(sliceOfSliceOfChars[i], char)
	}
	for _, sliceOfChars := range sliceOfSliceOfChars {
		wordInSlice := unsafe.String((*byte)(unsafe.Pointer(unsafe.SliceData(sliceOfChars))), len(sliceOfChars))
		if len(wordInSlice) == 0 || len(strings.TrimSpace(wordInSlice)) == 0 {
			continue
		}
		words = append(words, wordInSlice)
	}
	return words
}

// GetShort Get a slice of []int16 (short *) from result.
func (propertyResult *PropertyResult) GetShort() []int16 {
	if propertyResult.Format == 16 {
		return propertyResult.shortResult
	}
	return nil
}

// GetLong Get a slice of []int64 (long *) from result.
func (propertyResult *PropertyResult) GetLong() []int64 {
	if propertyResult.Format == 32 {
		return propertyResult.longResult
	}
	return nil
}

type QueryPointerResult struct {
	Root       Window // Root window that the pointer is in
	Child      Window // Child window that the pointer is located in, if any
	RootX      int    // Pointer coordinates relative to the root window's origin
	RootY      int    // Pointer coordinates relative to the root window's origin
	WinX       int    // Pointer coordinates relative to the specified window
	WinY       int    // Pointer coordinates relative to the specified window
	MaskReturn int    // Current state of the modifier keys and pointer buttons
}

type QueryTreeResult struct {
	Root     Window   // Root Window
	Parent   Window   // Parent Window
	Children []Window // List of children windows
}

type WindowAttributesResult struct {
	X                int        // Location of window
	Y                int        // Location of window
	Width            int        // Width of window
	Height           int        // Height of window
	BorderWidth      int        // Border width of window
	Visual           *C.Visual  // The associated visual structure
	Root             Window     // Root of screen containing window
	Colormap         C.Colormap // Color map to be associated with window
	MapInstalled     bool       // Is color map currently installed
	MapState         C.int      // C.IsUnmapped, C.IsUnviewable, C.IsViewable
	OverrideRedirect bool       // Value for override-redirect
	Screen           *C.Screen  // Pointer to correct screen
}

// MouseLocation
type MouseLocation struct {
	X            int    // Coordenate X where the mouse is
	Y            int    // Coordenate Y where the mouse is
	ScreenNumber int    // Screen number where the mouse is
	Window       Window // Window where the mouse is
}

// OpenDisplay Open connection to the X server
func OpenDisplay() {
	display = C.XOpenDisplay(nil)
}

// CloseDisplay Close connection to the X server, wrapper for XCloseDisplay
func CloseDisplay() {
	if display != nil {
		C.XCloseDisplay(display)
	}
}

// GetRootWindow Get root window, wrapper for XDefaultRootWindow
func GetRootWindow() Window {
	return C.XDefaultRootWindow(display)
}

/*
GetWindowProperty This function is a wrapper around XGetWindowProperty.

Parameters:
  - window: Specifies the window whose property you want to obtain.
  - property: Specifies the property name.

Returns:
  - A PropertyResult object with the result
  - Possible error or nil
*/
func GetWindowProperty(window Window, property string) (*PropertyResult, error) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		fmt.Println("RECOVER: ", err)
	}()

	// Default values
	offset := C.long(0)
	length := C.long(1024)

	// INPUT PARAMETERS
	window_ := C.Window(window)
	propertyName := C.CString(property)
	atomProperty := C.XInternAtom(display, propertyName, C.False)
	defer C.XFree(unsafe.Pointer(propertyName))

	// OUTPUT PARAMETERS
	var actualTypeReturn C.Atom
	var actualFormatReturn C.int
	var nitemsReturn C.ulong
	var bytesAfterReturn C.ulong
	var propReturn *C.uchar
	defer C.XFree(unsafe.Pointer(propReturn))

	// Result
	var charResult []int8
	var shortResult []int16
	var longResult []int64
	var actualTypeOfTheProperty string
	finalNItems := uint64(0)

	for {
		// Call C Function XGetWindowProperty
		result := C.XGetWindowProperty(
			display,
			window_,
			atomProperty,
			offset,
			length,
			C.False,
			C.AnyPropertyType,
			&actualTypeReturn,
			&actualFormatReturn,
			&nitemsReturn,
			&bytesAfterReturn,
			&propReturn,
		)
		if result != C.Success || (actualFormatReturn == 0 && bytesAfterReturn == 0) {
			return nil, fmt.Errorf("property \"%s\" not found on window %d", property, window)
		}
		if len(actualTypeOfTheProperty) == 0 {
			actualTypeOfTheProperty = C.GoString(C.XGetAtomName(display, actualTypeReturn))
		}
		finalNItems += uint64(nitemsReturn)
		switch actualFormatReturn {
		case 8:
			charResult = append(charResult, unsafe.Slice((*int8)(unsafe.Pointer(propReturn)), int(nitemsReturn))...)
		case 16:
			shortResult = append(shortResult, unsafe.Slice((*int16)(unsafe.Pointer(propReturn)), int(nitemsReturn))...)
		case 32:
			longResult = append(longResult, unsafe.Slice((*int64)(unsafe.Pointer(propReturn)), int(nitemsReturn))...)
		}
		if bytesAfterReturn > 0 { // buffer again till there's no remaining bytes
			C.XFree(unsafe.Pointer(propReturn))
			offset = length
			length = C.long(bytesAfterReturn/4 + 1)
		} else { // No buffer remaining
			break
		}
	}
	propertyResult := &PropertyResult{
		Format:         int(actualFormatReturn),
		TypeOfProperty: actualTypeOfTheProperty,
		NumberOfItems:  finalNItems,
		charResult:     charResult,
		shortResult:    shortResult,
		longResult:     longResult,
	}
	return propertyResult, nil
}

/*
ChangeWindowProperty This function is a wrapper around XChangeProperty.

Parameters:

  - window: Specifies the window whose property you want to change.

  - property: Specifies the property name you want to change.

  - typeOfProperty: Specifies the type of the property.

  - format: Possible values are 8, 16, and 32. Specifies whether the data should be viewed as a slice of int8, int16 or int64.

  - mode: Specifies the mode. It has to be a string with any of these values: PropModeReplace, PropModePrepend or PropModeAppend.

  - newValue: Specifies the new value for the property.

Returns:
  - A boolean indicating if the property was changed succesfully or not
  - Possible error or nil
*/
func ChangeWindowProperty[T string | []int8 | []int16 | []int64](
	window Window,
	property string,
	typeOfProperty string,
	format int,
	mode string,
	newValue T,
) (bool, error) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		fmt.Println("RECOVER: ", err)
	}()

	var mode_ C.int
	switch mode {
	case "PropModeReplace":
		mode_ = C.PropModeReplace
	case "PropModeAppend":
		mode_ = C.PropModeAppend
	case "PropModePrepend":
		mode_ = C.PropModePrepend
	default:
		return false, fmt.Errorf("the requested mode does not exist")
	}

	if format != 8 && format != 16 && format != 32 {
		return false, fmt.Errorf("the format %d is not valid, possible values are 8, 16 or 32", format)
	}

	propertyName := C.CString(property)
	atomProperty := C.XInternAtom(display, propertyName, C.False)
	defer C.XFree(unsafe.Pointer(propertyName))

	propertyTypeName := C.CString(typeOfProperty)
	atomPropertyType := C.XInternAtom(display, propertyTypeName, C.False)
	defer C.XFree(unsafe.Pointer(propertyTypeName))

	result := C.XChangeProperty(
		display,
		window,
		atomProperty,
		atomPropertyType,
		C.int(format),
		mode_,
		(*C.uchar)(reflect.ValueOf(newValue).UnsafePointer()),
		C.int(len(newValue)),
	)
	C.XFlush(display)
	if result == C.True {
		return true, nil
	} else {
		return false, fmt.Errorf("an error occurred changing the window property")
	}
}

// QueryPointer Wrapper around XQueryPointer
func QueryPointer(window Window) (*QueryPointerResult, error) {
	var rootReturn Window
	var childReturn Window
	var rootXReturn C.int
	var rootYReturn C.int
	var winXReturn C.int
	var winYReturn C.int
	var maskReturn C.uint
	result := C.XQueryPointer(
		display,
		window,
		&rootReturn,
		&childReturn,
		&rootXReturn,
		&rootYReturn,
		&winXReturn,
		&winYReturn,
		&maskReturn,
	)
	if result == C.False {
		return nil, fmt.Errorf(
			"an error occurred querying the pointer position for window %d with XQueryPointer",
			window,
		)
	}
	return &QueryPointerResult{
		Root:       Window(rootReturn),
		Child:      Window(childReturn),
		RootX:      int(rootXReturn),
		RootY:      int(rootYReturn),
		WinX:       int(winXReturn),
		WinY:       int(winYReturn),
		MaskReturn: int(maskReturn),
	}, nil
}

// QueryTree Wrapper around XQueryTree
func QueryTree(window Window) (*QueryTreeResult, error) {
	var rootReturn Window
	var parentReturn Window
	var childrenReturn *Window
	var nchildrenReturn C.uint
	result := C.XQueryTree(display, window, &rootReturn, &parentReturn, &childrenReturn, &nchildrenReturn)
	if result == C.False {
		return nil, fmt.Errorf("an error occurred querying the window tree for window %d with XQueryTree", window)
	}
	var finalListOfChildren []Window
	if nchildrenReturn > 0 {
		finalListOfChildren = append(
			finalListOfChildren,
			unsafe.Slice((*Window)(unsafe.Pointer(childrenReturn)), int(nchildrenReturn))...)
	}
	return &QueryTreeResult{Root: rootReturn, Parent: parentReturn, Children: finalListOfChildren}, nil
}

// GetWindowAttributes Wrapper around XGetWindowAttributes
func GetWindowAttributes(window Window) (*WindowAttributesResult, error) {
	var windowAttributes C.XWindowAttributes
	result := C.XGetWindowAttributes(display, window, &windowAttributes)
	if result == C.False {
		return nil, fmt.Errorf(
			"an error occurred getting window attributes for window %d with XGetWindowAttributes",
			window,
		)
	}
	return &WindowAttributesResult{
		X:                int(windowAttributes.x),
		Y:                int(windowAttributes.y),
		Width:            int(windowAttributes.width),
		Height:           int(windowAttributes.height),
		BorderWidth:      int(windowAttributes.border_width),
		Visual:           windowAttributes.visual,
		Root:             windowAttributes.root,
		Colormap:         windowAttributes.colormap,
		MapInstalled:     func() bool { return windowAttributes.map_installed == C.True }(),
		MapState:         windowAttributes.map_state,
		OverrideRedirect: func() bool { return windowAttributes.override_redirect == C.True }(),
		Screen:           windowAttributes.screen,
	}, nil
}

// GetWindowIcon Function that returns a *gdk.Pixbuf containing the window icon based on the property "_NET_WM_ICON"
func GetWindowIcon(window Window) *gdk.Pixbuf {
	// Internal struct representing a window icon
	type windowIcon struct {
		width  int
		height int
		size   int
		data   []int64
	}
	var icons []windowIcon // Slice with all available icons for a window

	// Get all buffer from property "_NET_WM_ICON"
	icons_, err := GetWindowProperty(window, "_NET_WM_ICON")
	if icons_ == nil && err != nil {
		return nil
	} else if len(icons_.longResult) == 0 {
		return nil
	}

	// Loop through result to append all available icons for the window
	for len(icons_.longResult) > 0 {
		// icon structure is represented by:
		// position 0: icon width
		// position 1: icon height
		// position 2...width*height: icon data
		// repeat till there's no data
		width := int(icons_.longResult[0])
		if len(icons_.longResult) < 2 {
			break
		}
		height := int(icons_.longResult[1])
		size := width * height
		if len(icons_.longResult) < size+2 {
			break
		}
		icons = append(icons, windowIcon{width: width, height: height, size: size, data: icons_.longResult[2 : size+2]})
		icons_.longResult = icons_.longResult[size+2:]
	}
	if len(icons) <= 0 {
		return nil
	}

	// Sort slice to let the biggest icon be the first one (best match)
	sort.Slice(icons, func(i, j int) bool {
		return icons[i].width > icons[j].width
	})
	bestMatch := icons[0]

	// Convert data slice from int64 (C.long) to uint8 (C.uchar) to obtain the BGRA format.
	// BGRA format comes with 4 bytes of extra padding: B, G, R, A, 4 extra bytes, then next values
	data := unsafe.Slice(
		(*C.uchar)(unsafe.Pointer(&bestMatch.data[0])),
		len(bestMatch.data)*int(unsafe.Sizeof(bestMatch.data[0])/unsafe.Sizeof(C.uchar(0))),
	)

	// Transform BGRA format to RGBA to create the *gdk.Pixbuf
	var bytes []byte // Slice of bytes to create the *gdk.Pixbuf
	for i := 0; i < len(data); {
		alpha := int(data[i+3])
		red := int(data[i+2])
		green := int(data[i+1])
		blue := int(data[i])
		i += 8 // Skip 4 bytes of extra padding
		// Append every byte of red, green, blue and alpha
		bytes = append(bytes, byte(red), byte(green), byte(blue), byte(alpha))
	}

	// Creation of *gdk.Pixbuf from bytes
	pixbuff, err := gdk.PixbufNewFromData(
		bytes,
		gdk.COLORSPACE_RGB,
		true,
		8,
		bestMatch.width,
		bestMatch.height,
		bestMatch.width*4,
	)
	if err != nil {
		return nil
	}
	return pixbuff
}

/*
ClickWindow function simulates a click (mouse left button press / release) on a window.

If window 0 is passed, then the click will be performed using XTest library (XTestFakeButtonEvent), else, it will use XSendEvent.
*/
func ClickWindow(window Window, button Button) bool {
	if !MousePress(window, button) {
		fmt.Printf("MousePress on window %d failed, aborting click.\n", window)
		return false
	}
	time.Sleep(time.Microsecond * 12)
	return MouseRelease(window, button)
}

func SimulateMouseButton(window Window, button Button, pressed bool) bool {
	if window == CURRENTWINDOW { // Send event to current window using XTest library
		retCode := C.XTestFakeButtonEvent(
			display,
			C.uint(button),
			C.int(*(*byte)(unsafe.Pointer(&pressed))),
			C.CurrentTime,
		)
		C.XFlush(display)
		if retCode == C.True {
			return true
		}
	} else {
		// Send to specific window
		mouseLocation, err := GetMouseLocation()
		if mouseLocation == nil && err != nil {
			return false
		}
		var xButtonEvent C.XButtonEvent
		xButtonEvent.x_root = C.int(mouseLocation.X)
		xButtonEvent.y_root = C.int(mouseLocation.Y)
		xButtonEvent.window = window
		xButtonEvent.button = C.uint(button)
		xButtonEvent.display = display
		xButtonEvent.root = C.screen_of_display(display, C.int(mouseLocation.ScreenNumber)).root
		xButtonEvent.same_screen = C.True
		xButtonEvent.subwindow = C.None
		xButtonEvent.time = C.CurrentTime
		if queryPointerResult, err := QueryPointer(C.default_root_window(display)); queryPointerResult != nil && err == nil {
			xButtonEvent.state = C.uint(queryPointerResult.MaskReturn)
		}
		if pressed {
			xButtonEvent._type = C.ButtonPress
		} else {
			xButtonEvent._type = C.ButtonRelease
		}
		C.XTranslateCoordinates(display, xButtonEvent.root, xButtonEvent.window, xButtonEvent.x_root, xButtonEvent.y_root, &xButtonEvent.x, &xButtonEvent.y, &xButtonEvent.subwindow)
		if !pressed {
			switch button {
			case 1:
				xButtonEvent.state |= C.Button1MotionMask
			case 2:
				xButtonEvent.state |= C.Button2MotionMask
			case 3:
				xButtonEvent.state |= C.Button3MotionMask
			case 4:
				xButtonEvent.state |= C.Button4MotionMask
			case 5:
				xButtonEvent.state |= C.Button5MotionMask
			}
		}
		result := C.XSendEvent(display, window, C.True, C.ButtonPressMask, (*C.XEvent)(unsafe.Pointer(&xButtonEvent)))
		C.XFlush(display)
		return result != C.False
	}
	return false
}

func MousePress(window Window, button Button) bool {
	return SimulateMouseButton(window, button, true)
}

func MouseRelease(window Window, button Button) bool {
	return SimulateMouseButton(window, button, false)
}

/*
GetMouseLocation Returns mouse location

Returns:

  - integer with the X coordinate
  - integer with the Y coordinate
  - integer with the screen number
  - Window where the mouse is
*/
func GetMouseLocation() (*MouseLocation, error) {
	var queryPointerResult *QueryPointerResult
	var errQueryPointer error
	var screenNumber int
	for i := range int(C.screen_count(display)) {
		screen := C.screen_of_display(display, C.int(i))
		if queryPointerResult, errQueryPointer = QueryPointer(screen.root); queryPointerResult != nil &&
			errQueryPointer == nil {
			screenNumber = i
			break
		}
	}
	if queryPointerResult == nil {
		return nil, fmt.Errorf("an error occurred getting the mouse location")
	}
	// Find the client window if we are not root.
	windowWhereMouseIs := queryPointerResult.Child
	if windowWhereMouseIs != queryPointerResult.Root && windowWhereMouseIs != CURRENTWINDOW {
		winn := Window(0)
		result := FindWindowClient(windowWhereMouseIs, true, &winn)
		if !result {
			result = FindWindowClient(windowWhereMouseIs, false, &winn)
		}
		if result {
			windowWhereMouseIs = winn
		}
	} else {
		windowWhereMouseIs = queryPointerResult.Root
	}
	return &MouseLocation{
		X:            int(queryPointerResult.RootX),
		Y:            int(queryPointerResult.RootY),
		ScreenNumber: screenNumber,
		Window:       windowWhereMouseIs,
	}, nil
}

/*
FindWindowClient Find a client window (child) in a given window.

If lookForParent is true, it will look for a client window that is a parent of the window given.
If lookForParent is false, it will look for a client window that is a child of the window given.
*/
func FindWindowClient(window Window, lookForParent bool, windowReturn *Window) bool {
	done := false
	for !done {
		if window == CURRENTWINDOW {
			return false
		}
		keepLooking := false
		lookForAlternativeProperty := false
		if netWMStateProperty, err := GetWindowProperty(window, "_NET_WM_STATE"); netWMStateProperty == nil &&
			err != nil {
			lookForAlternativeProperty = true
		} else if netWMStateProperty.NumberOfItems == 0 {
			keepLooking = true
		}
		if lookForAlternativeProperty {
			if wmStateProperty, err := GetWindowProperty(window, "WM_STATE"); wmStateProperty == nil &&
				err != nil {
				keepLooking = true
			} else if wmStateProperty.NumberOfItems == 0 {
				keepLooking = true
			}
		}
		if keepLooking {
			// This window doesn't have _NET_WM_STATE or WM_STATE property, keep searching
			queryTreeResult, errQueryTree := QueryTree(window)
			if queryTreeResult == nil && errQueryTree != nil {
				return false
			}
			if lookForParent {
				window = queryTreeResult.Parent
			} else {
				done = true
				for _, child := range queryTreeResult.Children {
					if FindWindowClient(child, lookForParent, &window) {
						*windowReturn = window
						break
					}
				}
				if len(queryTreeResult.Children) == 0 {
					return false
				}
			}
		} else {
			*windowReturn = window
			done = true
		}
	}
	return true
}

// GetActiveWindow gets the current actual window by querying the property "_NET_ACTIVE_WINDOW" on the root window.
func GetActiveWindow() (bool, Window) {
	activeWindow, err := GetWindowProperty(GetRootWindow(), "_NET_ACTIVE_WINDOW")
	if activeWindow == nil && err != nil {
		return false, Window(0)
	} else if activeWindow.NumberOfItems == 0 {
		return false, Window(0)
	}
	return true, Window(activeWindow.longResult[0])
}

func ActivateWindow(window Window) bool {
	// Try to change current desktop to the window's desktop if we're not in the same desktop
	if windowDesktopProp, _ := GetWindowProperty(window, "_NET_WM_DESKTOP"); windowDesktopProp != nil &&
		windowDesktopProp.NumberOfItems > 0 {
		if _, currentDesktop := GetCurrentDesktop(); currentDesktop >= 0 {
			windowDesktop := int(windowDesktopProp.longResult[0])
			if windowDesktop >= 0 && currentDesktop != windowDesktop {
				fmt.Printf("Changing current desktop from %d to %d\n", currentDesktop, windowDesktop)
				ChangeCurrentDesktop(int(windowDesktopProp.longResult[0]))
			}
		}
	}

	windowAttributes, err := GetWindowAttributes(window)
	if windowAttributes == nil && err != nil {
		return false
	} else if windowAttributes.Screen == nil {
		return false
	}

	var xClientMessageEvent C.XClientMessageEvent
	xClientMessageEvent._type = C.int(C.ClientMessage)
	xClientMessageEvent.display = display
	xClientMessageEvent.window = window
	netActiveWindow := C.CString("_NET_ACTIVE_WINDOW")
	defer C.XFree(unsafe.Pointer(netActiveWindow))
	xClientMessageEvent.message_type = C.XInternAtom(display, netActiveWindow, C.False)
	xClientMessageEvent.format = C.int(32)
	xClientMessageEvent.data = [40]byte{byte(2), byte(C.CurrentTime)}

	result := C.XSendEvent(
		display,
		windowAttributes.Screen.root,
		C.False,
		C.SubstructureNotifyMask|C.SubstructureRedirectMask,
		(*C.XEvent)(unsafe.Pointer(&xClientMessageEvent)),
	)
	C.XFlush(display)
	return result == C.True
}

// GetCurrentDesktop get current desktop based on property "_NET_CURRENT_DESKTOP" on the root window
func GetCurrentDesktop() (bool, int) {
	currentDesktopProp, err := GetWindowProperty(GetRootWindow(), "_NET_CURRENT_DESKTOP")
	if currentDesktopProp == nil && err != nil {
		return false, -1
	} else if currentDesktopProp.NumberOfItems == 0 {
		return false, -1
	}
	currentDesktop := int(currentDesktopProp.longResult[0])
	if currentDesktop >= 0 {
		return true, currentDesktop
	}
	return false, -1
}

func ChangeCurrentDesktop(desktop int) bool {
	if desktop < 0 {
		return false
	}

	amountOfAvailableDesktopsProp, err := GetWindowProperty(GetRootWindow(), "_NET_NUMBER_OF_DESKTOPS")
	if amountOfAvailableDesktopsProp == nil && err != nil {
		return false
	} else if amountOfAvailableDesktopsProp.NumberOfItems == 0 {
		return false
	}
	availableDesktops := int(amountOfAvailableDesktopsProp.longResult[0])
	if availableDesktops <= 0 || availableDesktops == 1 {
		return false
	}

	res, currentDesktop := GetCurrentDesktop()
	if !res || currentDesktop < 0 || desktop == currentDesktop {
		return false
	}

	var xClientMessageEvent C.XClientMessageEvent
	xClientMessageEvent._type = C.int(C.ClientMessage)
	xClientMessageEvent.display = display
	xClientMessageEvent.window = GetRootWindow()
	netCurrentDesktop := C.CString("_NET_CURRENT_DESKTOP")
	defer C.XFree(unsafe.Pointer(netCurrentDesktop))
	xClientMessageEvent.message_type = C.XInternAtom(display, netCurrentDesktop, C.False)
	xClientMessageEvent.format = C.int(32)
	xClientMessageEvent.data = [40]byte{byte(desktop), byte(C.CurrentTime)}

	result := C.XSendEvent(
		display,
		GetRootWindow(),
		C.False,
		C.SubstructureNotifyMask|C.SubstructureRedirectMask,
		(*C.XEvent)(unsafe.Pointer(&xClientMessageEvent)),
	)
	C.XFlush(display)
	return result == C.True
}

func WaitForWindowActivate(window Window, active bool) bool {
	var result bool
	activeWindow := Window(0)
	maxTries := 500

	for range maxTries {
		if active {
			if activeWindow == window {
				break
			}
		} else {
			if activeWindow != Window(0) && activeWindow != window {
				break
			}
		}
		result, activeWindow = GetActiveWindow()
		if !result {
			return false
		}
		time.Sleep(time.Microsecond * 30000)
	}
	return true
}
