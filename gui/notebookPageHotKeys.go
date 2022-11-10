package gui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"linux-windows-switcher/keyboard"
	"linux-windows-switcher/libs/glibown"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type contentTabAtajos struct {
	MainGUI
	listBoxHotKeys *gtk.ListBox
	listHotKeys    []*keyboard.HotKey
}

const (
	// Section from config file related with the global hotkeys
	sectionHotKeys = "hotkeys"

	// Signal used to set the global hotkeys
	signalSetHotKeys = "app-listener-set-hotkeys"

	// Resources used
	iconKeyBoard = "keyboard.png"
	iconKeys     = "keys.png"

	// Default representation for empty Hot Key
	emtpyHotKey = "---"

	// Key limit to every single global hotkey
	keyLimit = 3
)

var (
	// Global Hotkeys names
	moveForwards  string
	moveBackwards string

	// Information of all the global hotkeys. Structure: key: Name, value: Name of option in config file
	infoGlobalHotKeys = map[string]string{}
)

// newContentTabAtajos Constructor
func (mainGUI *MainGUI) newContentTabAtajos() *contentTabAtajos {
	contentTabAtajos := &contentTabAtajos{MainGUI: *mainGUI}

	contentTabAtajos.initLocale()
	contentTabAtajos.setupContentTabAtajos()
	return contentTabAtajos
}

// Config function to assign all UI strings to appropriate locale
func (contentTabAtajos *contentTabAtajos) initLocale() {
	obj, _ := contentTabAtajos.MainGUI.builder.GetObject("labelTitleGlobalHotKeys")
	labelTitleGlobalHotKeys := obj.(*gtk.Label)
	labelTitleGlobalHotKeys.SetMarkup(funcGetStringResource("gui_title_global_hotkeys"))

	obj, _ = contentTabAtajos.MainGUI.builder.GetObject("labelExplanationGlobalHotKeys")
	labelExplanationGlobalHotKeys := obj.(*gtk.Label)
	labelExplanationGlobalHotKeys.SetMarkup(funcGetStringResource("gui_explanation_global_hotkeys"))
}

// Config function
func (contentTabAtajos *contentTabAtajos) setupContentTabAtajos() {
	// Initialize vars
	moveForwards = funcGetStringResource("hotkey_move_forwards")
	moveBackwards = funcGetStringResource("hotkey_move_backwards")
	infoGlobalHotKeys[moveForwards] = "move_forwards"
	infoGlobalHotKeys[moveBackwards] = "move_backwards"

	// Create signal to update the ListBoxRow containing the global hotkeys whenever a hotkey is set/modified
	_, _ = glib.SignalNew("listbox-update-hotkey")

	obj, _ := contentTabAtajos.MainGUI.builder.GetObject("containerTitleHotKeys")
	containerTitleHotKeys := obj.(*gtk.Box)
	imageContainerTitleHotKeys := getGtkImageFromResource(iconKeyBoard, 48, 48)
	imageContainerTitleHotKeys.SetMarginBottom(10)
	containerTitleHotKeys.PackEnd(imageContainerTitleHotKeys, false, false, 0)

	// ------------------------------------------- GLOBAL HOTKEYS ------------------------------------

	// Anonymous function that creates an instance of *keyboard.Hotkey and appends it to the hotkey list
	funcAddHotKey := func(name string, callback func()) {
		contentTabAtajos.listHotKeys = append(contentTabAtajos.listHotKeys, keyboard.NewHotKey(name, callback))
	}

	// Default global hotkeys are added to the hotkey list
	funcAddHotKey(moveForwards, contentTabAtajos.MainGUI.moveForwards)
	funcAddHotKey(moveBackwards, contentTabAtajos.MainGUI.moveBackwards)

	obj, _ = contentTabAtajos.MainGUI.builder.GetObject("listBoxGlobalHotKeys")
	contentTabAtajos.listBoxHotKeys = obj.(*gtk.ListBox)

	// Current config of global hotkeys from config file is read and every hotkey is added to the *gtk.ListBox
	for _, hotKey := range contentTabAtajos.listHotKeys {
		contentTabAtajos.getConfigHotKey(hotKey)
		contentTabAtajos.listBoxHotKeys.Add(contentTabAtajos.createListBoxRowHotKey(hotKey))
	}
	// ----------------------------------------------------------------------------------------------------

	// Handler of signal "app-listener-set-hotkeys"
	contentTabAtajos.MainGUI.application.Connect(signalSetHotKeys, func(application *gtk.Application) {
		keyboard.SetHotKeys(contentTabAtajos.listHotKeys)
	})

	// Emit signal to activate the global hotkey listener when the app starts
	_, _ = contentTabAtajos.MainGUI.application.Emit(signalControlListener, true, true)
}

// Function that configures every *keyboard.Hotkey, its keys and its state
func (contentTabAtajos *contentTabAtajos) getConfigHotKey(hotKey *keyboard.HotKey) {
	result, _ := contentTabAtajos.MainGUI.application.Emit(
		signalGetConfig,
		sectionHotKeys,
		infoGlobalHotKeys[hotKey.Name],
	)
	cadenaInfoAtajo := result.(string)
	sliceInfoAtajo := strings.Split(cadenaInfoAtajo, ":")
	lengthSliceinfoGlobalHotKeys := len(sliceInfoAtajo)
	if lengthSliceinfoGlobalHotKeys > 0 && lengthSliceinfoGlobalHotKeys <= 2 {
		if lengthSliceinfoGlobalHotKeys == 2 {
			if strings.ToLower(sliceInfoAtajo[1]) == "disabled" {
				hotKey.Disabled = true
			}
		}
		sliceKeysAtajo := strings.Split(sliceInfoAtajo[0], ",")
		if len(sliceKeysAtajo) <= keyLimit {
			for _, key := range sliceKeysAtajo {
				key = strings.TrimSpace(key)
				keyVal := gdk.KeyvalFromName(key)
				if keyVal != gdk.KEY_VoidSymbol && !contains(hotKey.HotKeys, key) {
					hotKey.HotKeys = append(hotKey.HotKeys, key)
					hotKey.HotKeysKeyCodes = append(hotKey.HotKeysKeyCodes, keyVal)
				}
			}
		}
	}
}

// Function that creates a *gtk.ListBoxRow ready to be added to the *gtk.ListBox of global hotkeys
func (contentTabAtajos *contentTabAtajos) createListBoxRowHotKey(hotKey *keyboard.HotKey) *gtk.ListBoxRow {
	builder := getNewBuilder()

	// Config anonymous function to assign all UI strings to appropriate locale
	initLocale := func() {
		obj, _ := builder.GetObject("labelButtonDisableHotKey")
		labelButtonDisableHotKey := obj.(*gtk.Label)
		labelButtonDisableHotKey.SetMarkup(funcGetStringResource("disable"))

		obj, _ = builder.GetObject("labelButtonChangeHotKey")
		labelButtonChangeHotKey := obj.(*gtk.Label)
		labelButtonChangeHotKey.SetMarkup(funcGetStringResource("change"))
	}

	// Config anonymous function to assign the appropiate icons
	setIcons := func() {
		obj, _ := builder.GetObject("imageDisableHotKey")
		image := obj.(*gtk.Image)
		image.SetFromPixbuf(getPixBufAtSize("stop-listener.png", 24, 24))

		obj, _ = builder.GetObject("imageChangeHotKey")
		image = obj.(*gtk.Image)
		image.SetFromPixbuf(getPixBufAtSize("edit-hotkey.png", 24, 24))
	}

	initLocale()
	setIcons()

	obj, _ := builder.GetObject("listBoxRowGlobalHotKey")
	listBoxRowGlobalHotKey := obj.(*gtk.ListBoxRow)

	obj, _ = builder.GetObject("labelNameHotKey")
	labelNameHotKey := obj.(*gtk.Label)
	labelNameHotKey.SetMarkup(hotKey.Name)
	labelNameHotKey.SetTooltipText(hotKey.Name)

	obj, _ = builder.GetObject("labelKeysHotKey")
	labelKeysHotKey := obj.(*gtk.Label)
	if len(hotKey.HotKeys) == 0 {
		labelKeysHotKey.SetMarkup(emtpyHotKey)
	} else {
		labelKeysHotKey.SetMarkup(strings.Join(hotKey.HotKeys, " + "))
	}

	// Anonymous function that updates the state of a global hotkey whenever is enabled/disabled
	functionUpdateHotKey := func(hotkey keyboard.HotKey) bool {
		valueNewHotKey := strings.Join(hotkey.HotKeys, ",")
		if hotkey.Disabled {
			valueNewHotKey += ":disabled"
		}
		result, _ := contentTabAtajos.MainGUI.application.Emit(
			signalUpdateConfig,
			sectionHotKeys,
			infoGlobalHotKeys[hotkey.Name],
			valueNewHotKey,
		)
		return result.(bool)
	}

	// Button change hotkey
	obj, _ = builder.GetObject("buttonChangeHotKey")
	buttonChangeHotKey := obj.(*gtk.Button)
	buttonChangeHotKey.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() {
				button.SetSensitive(false)
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				contentTabAtajos.showDialogChangeHotKey(button, hotKey)
			})
		}()
	})

	// Button disable hotkey
	obj, _ = builder.GetObject("buttonDisableHotKey")
	buttonDisableHotKey := obj.(*gtk.ToggleButton)
	buttonDisableHotKey.Connect("toggled", func(button *gtk.ToggleButton) {
		hotKey.Disabled = button.GetActive()
		result := functionUpdateHotKey(*hotKey)
		if result {
			_, _ = contentTabAtajos.MainGUI.application.Emit(signalControlListener, true, false)
		}
		buttonChangeHotKey.SetSensitive(!button.GetActive())
	})

	// Handler of signal "listbox-update-hotkey" used to update the label of a hotkey
	buttonChangeHotKey.Connect(
		"listbox-update-hotkey", func(button *gtk.Button) {
			result := functionUpdateHotKey(*hotKey)
			if result {
				labelKeysHotKey.SetMarkup(strings.Join(hotKey.HotKeys, " + "))
				buttonDisableHotKey.SetSensitive(true)
			}
		},
	)

	if hotKey.Disabled && !buttonDisableHotKey.GetActive() {
		buttonDisableHotKey.SetActive(true)
	}

	buttonChangeHotKey.SetSensitive(!buttonDisableHotKey.GetActive())
	buttonDisableHotKey.SetSensitive(len(hotKey.HotKeysKeyCodes) > 0)
	return listBoxRowGlobalHotKey
}

// Function that shows a *gtk.Dialog to change a global hotkey keys
func (contentTabAtajos *contentTabAtajos) showDialogChangeHotKey(sourceButton *gtk.Button, hotKey *keyboard.HotKey) {
	// Emit signal to disable global hotkey listener
	_, _ = contentTabAtajos.MainGUI.application.Emit(signalControlListener, false, false)

	builder := getNewBuilder()

	// Config anonymous function to assign all UI strings to appropriate locale
	initLocale := func() {
		obj, _ := builder.GetObject("labelTitleDialogChangeHotKey")
		labelTitleDialogChangeHotKey := obj.(*gtk.Label)
		labelTitleDialogChangeHotKey.SetMarkup(funcGetStringResource("gui_config_hotkey_title"))

		obj, _ = builder.GetObject("labelInfoDialogChangeHotKeys")
		labelInfoDialogChangeHotKeys := obj.(*gtk.Label)
		labelInfoDialogChangeHotKeys.SetMarkup(
			fmt.Sprintf(
				"%s <b>\"%s\"</b>\n%s",
				funcGetStringResource("gui_config_hotkey_label"),
				hotKey.Name,
				strings.ReplaceAll(funcGetStringResource("gui_config_hotkey_label_info"), "%d", strconv.Itoa(keyLimit)),
			),
		)

		obj, _ = builder.GetObject("labelButtonOkChangeHotKeys")
		labelButtonOkChangeHotKeys := obj.(*gtk.Label)
		labelButtonOkChangeHotKeys.SetMarkup(funcGetStringResource("accept"))

		obj, _ = builder.GetObject("labelButtonDiscardChangeHotKeys")
		labelButtonDiscardChangeHotKeys := obj.(*gtk.Label)
		labelButtonDiscardChangeHotKeys.SetMarkup(funcGetStringResource("discard"))

		obj, _ = builder.GetObject("labelButtonCancelChangeHotKeys")
		labelButtonCancelChangeHotKeys := obj.(*gtk.Label)
		labelButtonCancelChangeHotKeys.SetMarkup(funcGetStringResource("cancel"))
	}

	// Config anonymous function to assign the appropiate icons
	setIcons := func() {
		obj, _ := builder.GetObject("imageAcceptHotKey")
		image := obj.(*gtk.Image)
		image.SetFromPixbuf(getPixBufAtSize("accept.png", 18, 18))

		obj, _ = builder.GetObject("imageDiscardHotKey")
		image = obj.(*gtk.Image)
		image.SetFromPixbuf(getPixBufAtSize("discard.png", 18, 18))

		obj, _ = builder.GetObject("imageCancelHotKey")
		image = obj.(*gtk.Image)
		image.SetFromPixbuf(getPixBufAtSize("cancel.png", 18, 18))
	}

	initLocale()
	setIcons()

	obj, _ := builder.GetObject("dialogChangeHotKey")
	dialogChangeHotKey := obj.(*gtk.Dialog)
	dialogChangeHotKey.SetTitle(fmt.Sprintf("%s - %s", title, funcGetStringResource("hotkey_title_config")))
	dialogChangeHotKey.SetIcon(defaultAppIcon)
	dialogChangeHotKey.SetTransientFor(contentTabAtajos.MainGUI.window)

	obj, _ = builder.GetObject("containerTitleDialogChangeHotKey")
	containerTitleDialogChangeHotKey := obj.(*gtk.Box)
	image := getGtkImageFromResource(iconKeys, 40, 40)
	containerTitleDialogChangeHotKey.PackEnd(image, false, false, 10)

	obj, _ = builder.GetObject("buttonCancelNewHotKey")
	buttonCancelNewHotKey := obj.(*gtk.Button)

	obj, _ = builder.GetObject("labelDialogChangeHotKey")
	labelDialogChangeHotKey := obj.(*gtk.Label)

	obj, _ = builder.GetObject("buttonAcceptNewHotKey")
	buttonAcceptNewHotKey := obj.(*gtk.Button)

	obj, _ = builder.GetObject("buttonDiscardNewHotKey")
	buttonDiscardNewHotKey := obj.(*gtk.Button)

	obj, _ = builder.GetObject("labelErrorNewHotKey")
	labelErrorNewHotKey := obj.(*gtk.Label)

	obj, _ = builder.GetObject("containerErrorDialogNewHotKey")
	containerErrorDialogNewHotKey := obj.(*gtk.Box)

	// Dialog vars
	var sliceOfKeys []string
	var sliceofKeyVals []uint
	amountOfKeysPressed := 0
	canEdit := true

	// Handler button okay
	buttonAcceptNewHotKey.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() {
				button.SetSensitive(false)
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				invalidHotKey := false
				error_ := ""
				if strings.Join(hotKey.HotKeys, " + ") == strings.Join(sliceOfKeys, " + ") {
					invalidHotKey = true
					error_ = funcGetStringResource("hotkey_keys_didnt_change")
				} else {
					for _, hotkey_ := range contentTabAtajos.listHotKeys {
						if strings.Join(hotkey_.HotKeys, " + ") == strings.Join(sliceOfKeys, " + ") {
							invalidHotKey = true
							error_ = fmt.Sprintf("%s\n<b>\"%s\"</b>.", funcGetStringResource("hotkey_keys_already_used"), hotkey_.Name)
							break
						}
					}
				}
				if !invalidHotKey { // Valid new HotKey, the signal is emitted to stablish it
					hotKey.HotKeys = sliceOfKeys
					hotKey.HotKeysKeyCodes = sliceofKeyVals
					_, _ = sourceButton.Emit("listbox-update-hotkey")
					dialogChangeHotKey.Response(gtk.RESPONSE_CLOSE)
				} else {
					labelErrorNewHotKey.SetMarkup(fmt.Sprintf("<span color='tomato'><b>%s:</b></span> %s", funcGetStringResource("error"), error_))
					containerErrorDialogNewHotKey.Show()
				}
			})
		}()
	})

	// Handler button discard
	buttonDiscardNewHotKey.Connect("clicked", func(button *gtk.Button) {
		labelDialogChangeHotKey.SetMarkup(emtpyHotKey)
		canEdit = true
		amountOfKeysPressed = 0
		sliceOfKeys = nil
		sliceofKeyVals = nil

		labelErrorNewHotKey.SetMarkup("")
		containerErrorDialogNewHotKey.Hide()

		button.SetSensitive(false)
		buttonAcceptNewHotKey.SetSensitive(false)
	})

	// Handler button cancel
	buttonCancelNewHotKey.Connect("clicked", func(button *gtk.Button) {
		dialogChangeHotKey.Response(gtk.RESPONSE_CLOSE)
	})

	dialogChangeHotKey.Connect("key-press-event", func(dialog *gtk.Dialog, event *gdk.Event) bool {
		if canEdit && amountOfKeysPressed < keyLimit {
			eventKey := gdk.EventKeyNewFromEvent(event)
			if keyName := glibown.KeyValName(eventKey.KeyVal()); len(keyName) > 0 {
				if !contains(sliceOfKeys, keyName) {
					sliceOfKeys = append(sliceOfKeys, keyName)
					sliceofKeyVals = append(sliceofKeyVals, eventKey.KeyVal())
					glib.IdleAdd(func() {
						labelDialogChangeHotKey.SetMarkup(strings.Join(sliceOfKeys, " + "))
					})
					amountOfKeysPressed++
				}
			}
		}
		return true
	})

	dialogChangeHotKey.Connect("key-release-event", func(dialog *gtk.Dialog, event *gdk.Event) bool {
		if canEdit {
			canEdit = false
			buttonAcceptNewHotKey.SetSensitive(true)
			buttonDiscardNewHotKey.SetSensitive(true)
		}
		return true
	})

	dialogChangeHotKey.Connect("response", func(dialog *gtk.Dialog, response int) {
		_, _ = contentTabAtajos.MainGUI.application.Emit(signalControlListener, true, false)
		dialog.Destroy()
		contentTabAtajos.MainGUI.window.Present()
		sourceButton.SetSensitive(true)
	})
	dialogChangeHotKey.Run()
}
