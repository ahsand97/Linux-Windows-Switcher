package gui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"linux-windows-switcher/libs/xlib"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var (
	funcGetResource       func(resource string) []byte // Anonymous function that returns a slice of bytes from a needed resource
	funcGetStringResource func(id string) string       // Anonymous function that returns a string from the localizer
)

// GetTitle Get title of the application
func GetTitle() string {
	return title
}

// Returns a *gdk.Pixbuf from a slice of bytes
func getPixBuf(content []byte) *gdk.Pixbuf {
	return func(pixbuf *gdk.Pixbuf, err error) *gdk.Pixbuf {
		return pixbuf
	}(gdk.PixbufNewFromBytesOnly(content))
}

// Returns a *gdk.Pixbuf from a slice of bytes at desired size
func getPixBufAtSize(iconName string, width int, height int) *gdk.Pixbuf {
	iconOriginalSize := getPixBuf(funcGetResource(iconName))
	iconScaled, _ := iconOriginalSize.ScaleSimple(width, height, gdk.INTERP_HYPER)
	return iconScaled
}

// This function returns the current active windows using Xlib
func listWindows(includeIcons bool) []window {
	var desktopNames []string
	var windows []window
	lookForAlternativeProperty := false

	// Client List
	netClientListResult, err := xlib.GetWindowProperty(xlib.GetRootWindow(), "_NET_CLIENT_LIST")
	if err != nil {
		lookForAlternativeProperty = true
	} else if netClientListResult != nil && len(netClientListResult.GetLong()) == 0 {
		lookForAlternativeProperty = true
	}
	if lookForAlternativeProperty {
		// GNOME Spec property "_WIN_CLIENT_LIST"
		netClientListResult, err = xlib.GetWindowProperty(xlib.GetRootWindow(), "_WIN_CLIENT_LIST")
		if err != nil {
			return windows
		} else if netClientListResult != nil && len(netClientListResult.GetLong()) == 0 {
			return windows
		}
	}

	// Get desktop information
	netDesktopNamesPropertyResult, err := xlib.GetWindowProperty(xlib.GetRootWindow(), "_NET_DESKTOP_NAMES")
	if netDesktopNamesPropertyResult != nil && netDesktopNamesPropertyResult.NumberOfItems > 0 {
		for _, desktopName := range netDesktopNamesPropertyResult.GetString() {
			desktopNames = append(desktopNames, desktopName)
		}
	}

	// Loop to get data of every window
	for _, windowId := range netClientListResult.GetLong() {
		// Window
		win := xlib.Window(windowId)

		// Window Desktop
		lookForAlternativeProperty = false
		desktop_, err := xlib.GetWindowProperty(win, "_NET_WM_DESKTOP")
		if err != nil {
			lookForAlternativeProperty = true
		} else if desktop_ != nil && len(desktop_.GetLong()) == 0 {
			lookForAlternativeProperty = true
		}
		if lookForAlternativeProperty {
			// GNOME Spec property "_WIN_WORKSPACE"
			desktop_, err = xlib.GetWindowProperty(win, "_WIN_WORKSPACE")
			if err != nil {
				continue
			} else if desktop_ != nil && len(desktop_.GetLong()) == 0 {
				continue
			}
		}
		desktop := int(desktop_.GetLong()[0])

		// Window class
		class_, err := xlib.GetWindowProperty(win, "WM_CLASS")
		if err != nil {
			continue
		} else if class_ != nil && len(class_.GetString()) == 0 {
			continue
		}
		class := strings.Join(class_.GetString(), ".")
		class = strings.TrimSpace(strings.TrimSuffix(class, "."))

		// Window Title
		lookForAlternativeProperty = false
		title_, err := xlib.GetWindowProperty(win, "_NET_WM_NAME")
		if err != nil {
			lookForAlternativeProperty = true
		} else if title_ != nil && len(title_.GetString()) == 0 {
			lookForAlternativeProperty = true
		}
		if lookForAlternativeProperty {
			title_, err = xlib.GetWindowProperty(win, "WM_NAME")
			if err != nil {
				continue
			} else if title_ != nil && len(title_.GetString()) == 0 {
				continue
			}
		}
		title := strings.Join(title_.GetString(), " ")

		// Window Icon
		var windowIcon *gdk.Pixbuf
		if includeIcons {
			originalIcon_ := xlib.GetWindowIcon(win)
			if originalIcon_ != nil {
				scaledIcon, err := originalIcon_.ScaleSimple(24, 24, gdk.INTERP_HYPER)
				if err == nil {
					windowIcon = scaledIcon
				}
			}
		}

		window := &window{
			id:      fmt.Sprint(windowId),
			class:   class,
			title:   title,
			desktop: desktop,
			desktopName: func() string {
				if desktop == -1 {
					return ""
				}
				if len(desktopNames) >= 0 && len(desktopNames) >= desktop {
					return desktopNames[desktop]
				}
				return strconv.Itoa(desktop)
			}(),
			icon: windowIcon,
		}
		windows = append(windows, *window)
	}
	return windows
}

// Remove item from a slice of strings
func removeItem(list []string, item string) []string {
	for index, value := range list {
		if value == item {
			return slices.Delete(list, index, index+1)
		}
	}
	return list
}

// Function that returns true wether an item of a slice of strings contains a string
// or if the string contains the item
func contains(list []string, string string) bool {
	response := false
	for _, item := range list {
		if strings.Contains(item, string) || strings.Contains(string, item) {
			response = true
			break
		}
	}
	return response
}

// Formats the output of WM_CLASS
func getClass(classString string) string {
	value := ""
	isRepeated := false
	for _, v := range strings.Split(classString, ".") {
		if strings.EqualFold(v, value) {
			if unicode.IsUpper(rune(v[0])) {
				value = v
			}
			isRepeated = true
			break
		} else {
			value = v
		}
	}
	if !isRepeated {
		value = classString
	}
	return value
}

// Function that checks if two window slices are equal
func funcTestEq(first []window, second []window) bool {
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i].order != second[i].order {
			return false
		}
	}
	return true
}

// Function that creates and return a *gtk.Image from a resource and changes its size to a desired one
func getGtkImageFromResource(resource string, width int, height int) *gtk.Image {
	pixbuf := getPixBuf(funcGetResource(resource))
	var gtkImage *gtk.Image
	if width > 0 && height > 0 {
		pixbufResized, _ := pixbuf.ScaleSimple(width, height, gdk.INTERP_HYPER)
		gtkImage, _ = gtk.ImageNewFromPixbuf(pixbufResized)
	} else {
		gtkImage, _ = gtk.ImageNewFromPixbuf(pixbuf)
	}
	gtkImage.Show()
	return gtkImage
}

// Function that returns a new instance of a *gtk.Builder based on UI content file
func getNewBuilder() *gtk.Builder {
	return func(builder *gtk.Builder, err error) *gtk.Builder {
		return builder
	}(gtk.BuilderNewFromString(uiFile))
}

// Function that updates the property "_NET_WM_NAME" and "WM_NAME" for a window
func changeWindowTitle(windowId string, newTitle string) bool {
	windowId_ := xlib.Window(func() uint64 {
		id, _ := strconv.ParseUint(windowId, 10, 64)
		return id
	}())
	netWMProp := "_NET_WM_NAME"
	wmProp := "WM_NAME"
	utf8StringType := "UTF8_STRING"
	stringType := "STRING"

	originalValueNetWMNAmeProp, _ := xlib.GetWindowProperty(windowId_, netWMProp)
	originalValueWMNameProp, _ := xlib.GetWindowProperty(windowId_, wmProp)

	funcChangeWindowTitle := func(property string, typeOfProperty string, value string) bool {
		result, err := xlib.ChangeWindowProperty(windowId_, property, typeOfProperty, 8, "PropModeReplace", value)
		if err != nil {
			return false
		}
		return result
	}

	funcRestoreWindowTitle := func(propertyResult *xlib.PropertyResult, property string, typeOfProperty string) {
		title := strings.Join(propertyResult.GetString(), " ")
		if len(title) > 0 {
			funcChangeWindowTitle(property, typeOfProperty, title)
		}
	}

	netWMNameResult := funcChangeWindowTitle(netWMProp, utf8StringType, newTitle)
	WMNameResult := funcChangeWindowTitle(wmProp, stringType, newTitle)

	finalResult := true
	if !netWMNameResult && originalValueNetWMNAmeProp != nil {
		funcRestoreWindowTitle(originalValueNetWMNAmeProp, netWMProp, utf8StringType)
		finalResult = false
	}
	if !WMNameResult && originalValueWMNameProp != nil {
		funcRestoreWindowTitle(originalValueWMNameProp, wmProp, stringType)
		finalResult = false
	}
	return finalResult && netWMNameResult && WMNameResult
}

//-------------------------------------------------- CALLBACKS GLOBAL HOTKEYS -------------------------------------------

// Function to move between windows following the current order, it can go backwards or forwards
func (mainGUI *MainGUI) moveNextWindow(backwards bool, times int) {
	fmt.Printf("(Callback) moveNextWindow(backwards: %t)\n", backwards)
	result, currentWindow_ := xlib.GetActiveWindow()
	if !result || currentWindow_ == xlib.CURRENTWINDOW {
		times++
		if times == 5 {
			return
		}
		fmt.Println("ERROR OBTAINING CURRENT WINDOW, calling itself again.")
		mainGUI.moveNextWindow(backwards, times)
	}
	fmt.Printf("(Callback) currentWindow: %d\n", currentWindow_)
	currentWindow := strconv.Itoa(int(currentWindow_))
	if len(currentOrder) == 1 && currentWindow == currentOrder[0].id {
		return
	}
	var nextIndex int
	if backwards {
		nextIndex = currentIndex - 1
		if nextIndex < 0 {
			nextIndex = len(currentOrder) - 1
		}
	} else {
		nextIndex = currentIndex + 1
		if nextIndex >= len(currentOrder) {
			nextIndex = 0
		}
	}
	if currentOrder[currentIndex].id == currentOrder[nextIndex].id &&
		strings.Contains(currentOrder[nextIndex].class, "cloned") {
		currentIndex = nextIndex
		return
	}
	isNextWindowValid := false
	for _, windowActive := range listWindows(false) {
		if strings.Contains(windowActive.class, mainGUI.application.GetApplicationID()) || windowActive.desktop == -1 {
			continue
		}
		if windowActive.id == currentOrder[nextIndex].id {
			isNextWindowValid = true
			break
		}
	}
	recursiveCall := false // Wether the function should call itself again
	if isNextWindowValid {
		fmt.Println("(Callback) Next window:", currentOrder[nextIndex].windowToString())
		// Activating window
		nextWindow := xlib.Window(func() int {
			id, _ := strconv.Atoi(currentOrder[nextIndex].id)
			return id
		}())
		if xlib.ActivateWindow(nextWindow) {
			if xlib.WaitForWindowActivate(nextWindow, true) {
				currentIndex = nextIndex
			}
		}
	} else {
		fmt.Println("(Callback) Next window:", currentOrder[nextIndex], "IS NOT VALID")
		recursiveCall = true
		fmt.Println("\nCALLING FUNCTION AGAIN, TRIGGERED BY A RECURSIVE CALL")
	}
	if recursiveCall {
		glib.IdleAdd(func() {
			// Emit signal to delete invalid window
			_, _ = mainGUI.application.Emit(signalDeleteRow, glib.TYPE_NONE, strconv.Itoa(nextIndex))
			// When it's done deleting the window call itself again
			glib.IdleAdd(func() { mainGUI.moveNextWindow(backwards, times) })
		})
	}
}

// Function to move to the next window (forwards). Callback of global hotkey
func (mainGUI *MainGUI) moveForwards() {
	times := 0
	if len(currentOrder) > 0 {
		mainGUI.moveNextWindow(false, times)
	}
}

// Function to move to the next window (backwards). Callback of global hotkey
func (mainGUI *MainGUI) moveBackwards() {
	times := 0
	if len(currentOrder) > 0 {
		mainGUI.moveNextWindow(true, times)
	}
}
