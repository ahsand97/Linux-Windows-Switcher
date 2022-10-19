package keyboard

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"

	hook "github.com/robotn/gohook"

	"github.com/gotk3/gotk3/gtk"
)

type ListenerKeyboard struct {
	application   *gtk.Application
	active        bool
	listenerState bool
}

type Atajo struct {
	Nombre        string
	Atajo         []string
	AtajoKeyCodes []uint
	Disabled      bool
	Callback      func()
}

const (
	// Signals of application used
	signalControlListener   = "app-listener-keyboard"
	signalSetAtajos         = "app-listener-set-atajos"
	signalSyncStateListener = "app-listener-sync-state"
)

var (
	// Slice of global hotkey objects
	atajos []*Atajo

	// Pressed keys
	teclas = map[uint16]bool{}
)

// NewListenerKeyBoard constructor
func NewListenerKeyBoard(application *gtk.Application) *ListenerKeyboard {
	listenerKeyboard := &ListenerKeyboard{application: application}

	listenerKeyboard.setupContentKeyboard()
	return listenerKeyboard
}

// Configuration function
func (listenerKeyboard *ListenerKeyboard) setupContentKeyboard() {
	// Handler señal para el control del listener del teclado
	listenerKeyboard.application.Connect(
		signalControlListener,
		func(application *gtk.Application, active bool, updateCtrl bool) {
			glib.IdleAdd(func() {
				if updateCtrl {
					listenerKeyboard.listenerState = !listenerKeyboard.listenerState
					_, _ = listenerKeyboard.application.Emit(signalSyncStateListener, listenerKeyboard.listenerState)
				}
				if active { // Si se activa el listener se emite la señal para setear los atajos
					_, _ = application.Emit(signalSetAtajos)
				} else { // Si se desactiva el listener se vacía el mapa con las teclas presionadas
					if len(teclas) > 0 {
						teclas = map[uint16]bool{}
					}
				}
				listenerKeyboard.active = active
			})
		},
	)
	listenerKeyboard.startKeyboardListener()
}

// Initializes the keyboard listener
func (listenerKeyboard *ListenerKeyboard) startKeyboardListener() {
	go func() {
		var teclaDownOrHold uint16
		channel := hook.Start()
		for evento := range channel {
			if (evento.Kind == hook.KeyDown || evento.Kind == hook.KeyHold) && listenerKeyboard.active {
				// 2 events (KeyDown and KeyHold) for the same key can't be reported, one is ignored
				if evento.Rawcode == teclaDownOrHold {
					continue
				}
				teclaDownOrHold = evento.Rawcode
				teclas[evento.Rawcode] = true // Update the map to indicate the key is pressed
				// Every time a key is pressed we check if the global hotkey was activated
				verificaActivacionAtajo()
			} else if evento.Kind == hook.KeyUp && listenerKeyboard.active {
				teclas[evento.Rawcode] = false // Update the map to indicate the key is not pressed anymore
				teclaDownOrHold = 0
				// Delete all non-active entries from the map
				for index, tecla := range teclas {
					if !tecla {
						delete(teclas, index)
					}
				}
			}
		}
	}()
}

func ExitListener() {
	hook.End()
}

// NewAtajo Constructor atajo
func NewAtajo(nombre string, callback func()) *Atajo {
	atajo := &Atajo{
		Nombre:   nombre,
		Disabled: false,
		Callback: callback,
	}
	return atajo
}

func SetAtajos(atajosInput []*Atajo) {
	atajos = atajosInput
}

// Function that checks if any global hotkey was pressed to trigger its callback
func allPressed(pressed map[uint16]bool, keys []uint) bool {
	if len(keys) == 0 {
		return false
	}
	for _, key := range keys {
		if !pressed[uint16(key)] {
			return false
		}
	}
	return true
}

// Loop through the hotkeys to find out if any has been activated to trigger its callback
func verificaActivacionAtajo() {
	for _, atajo := range atajos {
		// If the global hotkey is not disabled and the keys are pressed its callback gets triggered
		if !atajo.Disabled && allPressed(teclas, atajo.AtajoKeyCodes) {
			fmt.Println("-------------------------------------------------------")
			fmt.Printf("GLOBAL HOTKEY PRESSED: %s\n", atajo.Atajo)
			atajo.Callback()
			break
		}
	}
}
