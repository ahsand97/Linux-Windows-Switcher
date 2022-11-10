package glibown

// #cgo pkg-config: gdk-3.0
// #include <gdk/gdk.h>
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

type Signal struct {
	name     string
	signalId C.guint
}

func (s *Signal) String() string {
	return s.name
}

/*
SignalNewV is a wrapper around g_signal_newv().

Parameters:

  - signalName          : The name for the signal.

  - returnType          : The type of return value, or TYPE_NONE for a signal without a return value.

  - nParams             : Amount of extra parameters the signal is going to recieve (the object who emits the signal does not count).

  - paramsTypes...      : Datatypes of the parameters (amount of elements must match nParams, except when nParams is 0).
    If nParams is 0 then paramsTypes has to be TYPE_NONE.
*/
func SignalNewV(signalName string, returnType glib.Type, nParams uint, paramsTypes ...glib.Type) (*Signal, error) {
	if nParams == 0 {
		if paramsTypes[0] != glib.TYPE_NONE || len(paramsTypes) != 1 {
			return nil, fmt.Errorf(
				"invalid Types: the amount of parameters is %d, paramsTypes must be TYPE_NONE",
				nParams,
			)
		}
	} else if nParams == 1 {
		if paramsTypes[0] == glib.TYPE_NONE || len(paramsTypes) != 1 {
			return nil, fmt.Errorf("invalid Types: the amount of parameters is %d, paramsTypes must be different from TYPE_NONE", nParams)
		}
	} else {
		if len(paramsTypes) != int(nParams) {
			return nil, fmt.Errorf("invalid Types: The amount of elements of paramsTypes has to be equal to %d", nParams)
		}
	}
	cstr := C.CString(signalName)
	defer C.free(unsafe.Pointer(cstr))

	var sliceOfGTypes []C.GType
	for _, paramType := range paramsTypes {
		sliceOfGTypes = append(sliceOfGTypes, C.ulong(paramType))
	}

	signalId := C.g_signal_newv(
		(*C.gchar)(cstr),
		C.G_TYPE_OBJECT,
		C.G_SIGNAL_ACTION,
		(*C.GClosure)(C.NULL),
		(*[0]byte)(C.NULL),
		C.gpointer(C.NULL),
		(*[0]byte)(C.g_cclosure_marshal_VOID__POINTER),
		C.gulong(returnType),
		C.guint(nParams),
		(*C.GType)(&sliceOfGTypes[0]),
	)
	return &Signal{name: signalName, signalId: signalId}, nil
}

// KeyValName is a wrapper around gdk_keyval_name()
func KeyValName(keyval uint) string {
	return C.GoString(C.gdk_keyval_name(C.guint(keyval)))
}
