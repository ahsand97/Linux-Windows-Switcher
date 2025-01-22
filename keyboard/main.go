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

type HotKey struct {
	Name            string
	HotKeys         []string
	HotKeysKeyCodes []uint
	Disabled        bool
	Callback        func() // Callback func, it gets called when the HotKeys of the HotKey are pressed
}

const (
	// Signals of application used
	signalControlListener   = "app-listener-keyboard"
	signalSetHotKeys        = "app-listener-set-hotkeys"
	signalSyncStateListener = "app-listener-sync-state"
)

var (
	// Slice of global hotkey objects
	hotKeys []*HotKey

	// Pressed keys
	keys = map[uint16]bool{}
	// Channel used to communicate with the goroutine listening to keyboard events
	mainChannel = make(chan bool)
)

// NewListenerKeyBoard constructor
func NewListenerKeyBoard(application *gtk.Application) *ListenerKeyboard {
	listenerKeyboard := &ListenerKeyboard{application: application}

	listenerKeyboard.setupContentKeyboard()
	return listenerKeyboard
}

// Configuration function
func (listenerKeyboard *ListenerKeyboard) setupContentKeyboard() {
	// Handler of signal to manage the global hotkey listener
	listenerKeyboard.application.Connect(
		signalControlListener,
		func(application *gtk.Application, active bool, updateCtrl bool) {
			glib.IdleAdd(func() {
				if updateCtrl { // Wether the state should be reflected on the UI (Window and indicator)
					listenerKeyboard.listenerState = !listenerKeyboard.listenerState
					_, _ = listenerKeyboard.application.Emit(
						signalSyncStateListener,
						glib.TYPE_NONE,
						listenerKeyboard.listenerState,
					)
				}
				if active { // If listener is active a signal is emitted to set the global hotkeys
					_, _ = application.Emit(signalSetHotKeys, glib.TYPE_NONE)
				} else { // If listener is inactive the map containing the pressed keys gets cleaned
					if len(keys) > 0 {
						keys = map[uint16]bool{}
					}
				}
				listenerKeyboard.active = active
			})
		},
	)
	listenerKeyboard.startKeyboardListener()
	mainChannel <- true
}

// Initializes the keyboard listener
func (listenerKeyboard *ListenerKeyboard) startKeyboardListener() {
	go func() {
		listenerActive := false
		for statListener := range mainChannel {
			if !statListener {
				if !listenerActive { // Listener already inactive
					return
				}
				listenerActive = false
			} else {
				if listenerActive { // Listener already active
					return
				}
				go func() {
					var teclaDownOrHold uint16
					channel := hook.Start()
					listenerActive = true
					for evento := range channel {
						if (evento.Kind == hook.KeyDown || evento.Kind == hook.KeyHold) && listenerKeyboard.active {
							// 2 events (KeyDown and KeyHold) for the same key can't be reported, one is ignored
							if evento.Rawcode == teclaDownOrHold {
								continue
							}
							teclaDownOrHold = evento.Rawcode
							keys[evento.Rawcode] = true // Update the map to indicate the key is pressed
							// Every time a key is pressed we check if the global hotkey was activated
							checkKeysPressed()
						} else if (evento.Kind == hook.KeyUp) && listenerKeyboard.active {
							keys[evento.Rawcode] = false // Update the map to indicate the key is not pressed anymore
							teclaDownOrHold = 0
							// Delete all non-active entries from the map
							for index, key := range keys {
								if !key {
									delete(keys, index)
								}
							}
						}
					}
				}()
			}
		}
	}()
}

// ExitListener Stops Keyboard Listener
func ExitListener() {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		fmt.Println("RECOVER: ", err)
	}()
	mainChannel <- false
	close(mainChannel)
	hook.End()
}

// NewHotKey Constructor HotKey
func NewHotKey(name string, callback func()) *HotKey {
	hotkey := &HotKey{
		Name:     name,
		Disabled: false,
		Callback: callback,
	}
	return hotkey
}

// SetHotKeys sets the hotkeys for the keyboard listener
func SetHotKeys(hotKeysInput []*HotKey) {
	hotKeys = hotKeysInput
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
func checkKeysPressed() {
	for _, hotKey := range hotKeys {
		// If the global hotkey is not disabled and the keys are pressed its callback gets triggered
		if !hotKey.Disabled && allPressed(keys, hotKey.HotKeysKeyCodes) {
			fmt.Printf("\nGLOBAL HOTKEY PRESSED: %s\n", hotKey.HotKeys)
			hotKey.Callback()
			break
		}
	}
}
