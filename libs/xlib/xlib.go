package xlib

//#cgo pkg-config: x11
//#include <X11/Xlib.h>
import "C"

import (
	"fmt"
	"sort"
	"strings"
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
)

type Window = C.Window

var display *C.Display

// Struct representing the result of XGetWindowProperty call.
type PropertyResult struct {
	Format        int     // Format of result, it can be 8 (char *), 16 (short *) or 32 (long *)
	NumberOfItems uint64  // Number of total items returned by XGetWindowProperty
	charResult    []int8  // Result of casting buffer from char * to a slice of int8 (char *). This slice is empty if Format is not 8
	shortResult   []int16 // Result of casting buffer from char * to a slice of int16 (short *). This slice is empty if Format is not 16
	longResult    []int64 // Result of casting buffer from char * to a slice of int64 (long *). This slice is empty if Format is not 32
}

// Get a slice of []int8 (char *) from result.
func (propertyResult *PropertyResult) GetChar() []int8 {
	return propertyResult.charResult
}

// Get a slice of []string from result.
func (propertyResult *PropertyResult) GetString() []string {
	// slice containing the string result, it is grouped as a 2D array of words and letters
	strings_ := [][]string{{""}}
	i := 0
	result := []string{""}

	// Loop through every char
	for _, char := range propertyResult.GetChar() {
		if char == 0 { // When 0 found it means end of word or new word
			i++
			strings_ = append(strings_, []string{""})
			continue
		}
		char_ := C.char(char)                                 // Cast again to char
		strings_[i] = append(strings_[i], C.GoString(&char_)) // Get GoString from char
	}
	i = 0
	for index, letters := range strings_ {
		if !(len(letters) > 0) || !(len(strings.TrimSpace(strings.Join(letters, ""))) > 0) {
			continue
		}
		if index > 0 {
			i++
			result = append(result, "")
		}
		result[i] += strings.Join(letters, "")
	}
	return result
}

// Get a slice of []int16 (short *) from result.
func (propertyResult *PropertyResult) GetShort() []int16 {
	return propertyResult.shortResult
}

// Get a slice of []int64 (long *) from result.
func (propertyResult *PropertyResult) GetLong() []int64 {
	return propertyResult.longResult
}

// Open connection to the X server
func OpenDisplay() {
	display = C.XOpenDisplay(nil)
}

// Close connection to the X server
func CloseDisplay() {
	if display != nil {
		C.XCloseDisplay(display)
	}
}

// Get root window
func GetRootWindow() Window {
	return C.XDefaultRootWindow(display)
}

/*
This function is a wrapper around XGetWindowProperty.

Parameters:

  - window: Specifies the window whose property you want to obtain
  - property: Specifies the property name.
  - offset: Specifies the offset in the specified property (in 32-bit quantities) where the data is to be retrieved.
  - length: Specifies the length in 32-bit multiples of the data to be retrieved.
  - wholebuffer: Specifies whether to query all the remaining bytes or not, useful when the length is not enough
    to hold the whole data.
*/
func GetWindowProperty(
	window Window,
	property string,
	offset int,
	length int,
	wholebuffer bool,
) (*PropertyResult, error) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		fmt.Println("RECOVER: ", err)
	}()
	if length <= 0 {
		return nil, fmt.Errorf("length should be at least 1")
	}
	// INPUT PARAMETERS
	window_ := C.Window(window)
	property_name := C.CString(property)
	atomProperty := C.XInternAtom(display, property_name, C.False)
	defer C.XFree(unsafe.Pointer(property_name))

	// Buffer size
	default_bufer_size := C.long(length)
	long_offset_ := C.long(offset)
	long_length_ := default_bufer_size

	// OUTPUT PARAMETERS
	var actual_type_return_ C.Atom
	var actual_format_return_ C.int
	var nitems_return_ C.ulong
	var bytes_after_return_ C.ulong
	var prop_return_ *C.uchar
	defer C.XFree(unsafe.Pointer(prop_return_))

	// Result
	charResult := []int8{}
	shortResult := []int16{}
	longResult := []int64{}
	finalNItems := uint64(0)

	for {
		// Call C Function XGetWindowProperty
		result := C.XGetWindowProperty(
			display,
			window_,
			atomProperty,
			long_offset_,
			long_length_,
			C.False,
			C.AnyPropertyType,
			&actual_type_return_,
			&actual_format_return_,
			&nitems_return_,
			&bytes_after_return_,
			&prop_return_,
		)
		if result != C.Success || (actual_format_return_ == 0 && bytes_after_return_ == 0) {
			return nil, fmt.Errorf("property not found")
		}
		finalNItems += uint64(nitems_return_)
		switch int(actual_format_return_) {
		case 8:
			charResult = append(charResult, unsafe.Slice((*int8)(unsafe.Pointer(prop_return_)), nitems_return_)...)
		case 16:
			shortResult = append(shortResult, unsafe.Slice((*int16)(unsafe.Pointer(prop_return_)), nitems_return_)...)
		case 32:
			longResult = append(longResult, unsafe.Slice((*int64)(unsafe.Pointer(prop_return_)), nitems_return_)...)
		}
		if !wholebuffer {
			break
		}
		if bytes_after_return_ > 0 { // buffer again till there's no remaining bytes
			C.XFree(unsafe.Pointer(prop_return_))
			long_offset_ = default_bufer_size
			long_length_ = C.long(bytes_after_return_/4) + 1
		} else { // No buffer remaining
			break
		}
	}
	propertyResult := &PropertyResult{
		Format:        int(actual_format_return_),
		NumberOfItems: finalNItems,
		charResult:    charResult,
		shortResult:   shortResult,
		longResult:    longResult,
	}
	return propertyResult, nil
}

// Function that returns a *gdk.Pixbuf containing the window icon based on the property "_NET_WM_ICON"
func GetWindowIcon(window Window) *gdk.Pixbuf {
	propertyName := "_NET_WM_ICON"
	defaultBufferSize := 1024

	// Internal struct representing a window icon
	type windowIcon struct {
		width  int
		height int
		size   int
		data   []int64
	}
	icons := []windowIcon{} // Slice with all available icons for a window

	// Get all buffer from property "_NET_WM_ICON"
	icons_, err := GetWindowProperty(window, propertyName, 0, defaultBufferSize, true)
	if err != nil {
		return nil
	} else if icons_ == nil || len(icons_.longResult) == 0 {
		return nil
	}

	// Loop through result to append all available icons for the window
	for len(icons_.longResult) > 0 {
		// icon structure is represented by:
		// position 0: icon width
		// position 1: icon height
		// position 2...width*height: icon data
		// repeat again till there's no data
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
	if !(len(icons) > 0) {
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
	bytes := []byte{} // Slice of bytes to create the *gdk.Pixbuf
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
