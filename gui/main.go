package gui

import (
	"fmt"
	"time"

	"github.com/gotk3/gotk3/glib"

	"linux-windows-switcher/libs/xlib"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type MainGUI struct {
	application                *gtk.Application
	builder                    *gtk.Builder
	window                     *gtk.Window
	contentTabVentanas         *contentTabVentanas
	contentTabAtajos           *contentTabAtajos
	buttonControlListener      *gtk.Button
	labelButtonControlListener *gtk.Label
	imageButtonControlListener *gtk.Image
}

type window struct {
	id          string
	class       string
	title       string
	desktop     int
	desktopName string
	order       int
	icon        *gdk.Pixbuf
}

func (w window) windowToString() string {
	return fmt.Sprintf(
		"{ID: %s, DesktopNumber: %d, DesktopName: %s, Class: %s, Title: %s}",
		w.id,
		w.desktop,
		w.desktopName,
		w.class,
		w.title,
	)
}

// Globals
var (
	showWindow        bool
	uiFile            string
	defaultAppIcon    *gdk.Pixbuf // Default application's icon
	headerBargtkImage *gtk.Image  // HeaderBar image
	currentIndex      int         = 0
	currentOrder      []window
	defaultOrder      []window
	listenerState     bool // State of global hotkey listener
)

// Constants
const (
	// Title
	title = "Linux Windows Switcher"

	// Resources names
	uiFileName           = "Application.xml"
	iconFileName         = "tabs.png"
	iconFileNameDisabled = "tabs-disabled.png"

	// Signals of the application used
	signalReboot          = "app-restart"
	signalExit            = "app-exit"
	signalGetConfig       = "app-get-config"
	signalUpdateConfig    = "app-update-config"
	signalControlListener = "app-listener-keyboard"
	signalSetOrder        = "app-set-order"
	signalDeleteRow       = "app-delete-window-order"

	// Page where the configuration of global hotkeys is
	pageContentGlobalHotkeys = 1
)

// NewMainGUI Constructor MainGUI
func NewMainGUI(
	application *gtk.Application,
	showWindow_ bool,
	funcGetResource_ func(resource string) []byte,
	funcGetStringResource_ func(id string) string,
) *MainGUI {
	// Assign functions coming from main module
	funcGetResource = funcGetResource_
	funcGetStringResource = funcGetStringResource_

	// Fill global vars with their initial values
	uiFile = string(funcGetResource(uiFileName))
	defaultAppIcon = getPixBuf(funcGetResource(iconFileName))

	// Window should be shown or not
	showWindow = showWindow_

	// Creation of main GUI struct
	mainGui := &MainGUI{application: application, builder: getNewBuilder()}
	mainGui.initLocale()
	mainGui.setupUi()
	mainGui.setIconsUI()
	mainGui.contentTabVentanas = mainGui.newContentTabVentanas()
	mainGui.contentTabAtajos = mainGui.newContentTabAtajos()
	return mainGui
}

// Config function to assign all UI strings to appropriate locale
func (mainGUI *MainGUI) initLocale() {
	obj, _ := mainGUI.builder.GetObject("labelTitleTabActiveWindows")
	labelTitleTabActiveWindows := obj.(*gtk.Label)
	labelTitleTabActiveWindows.SetMarkup(funcGetStringResource("gui_notebook_title_open_windows"))

	obj, _ = mainGUI.builder.GetObject("labelTitleTabHotKeys")
	labelTitleTabHotKeys := obj.(*gtk.Label)
	labelTitleTabHotKeys.SetMarkup(funcGetStringResource("gui_notebook_title_hotkeys"))

	obj, _ = mainGUI.builder.GetObject("labelButtonHideWindow")
	labelButtonHideWindow := obj.(*gtk.Label)
	labelButtonHideWindow.SetMarkup(funcGetStringResource("gui_label_button_hide_window"))

	obj, _ = mainGUI.builder.GetObject("labelButtonRestart")
	labelButtonRestart := obj.(*gtk.Label)
	labelButtonRestart.SetMarkup(funcGetStringResource("restart"))

	obj, _ = mainGUI.builder.GetObject("labelButtonExit")
	labelButtonExit := obj.(*gtk.Label)
	labelButtonExit.SetMarkup(funcGetStringResource("exit"))
}

// Config function
func (mainGUI *MainGUI) setupUi() {
	// Main window
	obj, _ := mainGUI.builder.GetObject("mainWindow")
	mainGUI.window = obj.(*gtk.Window)
	mainGUI.window.SetApplication(mainGUI.application)
	mainGUI.window.SetIcon(defaultAppIcon)
	mainGUI.window.SetTitle(title)
	mainGUI.window.HideOnDelete() // Avoid exit app when closing window

	// It tells if the initial size of window is already set (when it opens for the first time)
	initialSizeSet := false
	mainGUI.window.Connect("show", func(window *gtk.Window) {
		if !initialSizeSet {
			window.Resize(1, 1) // Resize window to 1x1, so it forces it to have its minimum allowed size
			initialSizeSet = true
		}
	})

	// Title bar
	obj, _ = mainGUI.builder.GetObject("headerBarMainWindow")
	headerBar := obj.(*gtk.HeaderBar)

	// Header bar's image
	headerBargtkImage, _ = gtk.ImageNewFromPixbuf(func(pixbuf *gdk.Pixbuf, err error) *gdk.Pixbuf {
		return pixbuf
	}(defaultAppIcon.ScaleSimple(25, 25, gdk.INTERP_HYPER)))
	headerBargtkImage.SetMarginTop(5)
	headerBargtkImage.SetMarginBottom(5)

	// Container of header bar's image so it can listen to events
	eventBoxImageHeaderBar, _ := gtk.EventBoxNew()
	eventBoxImageHeaderBar.Connect("button-release-event", func(box *gtk.EventBox, event *gdk.Event) bool {
		eventButton := gdk.EventButtonNewFromEvent(event)
		if eventButton.Button() == gdk.BUTTON_PRIMARY {
			xlib.ClickWindow(xlib.CURRENTWINDOW, xlib.MBUTTON_RIGHT)
		}
		return false
	})

	// Add event box to image and then to header bar
	eventBoxImageHeaderBar.Add(headerBargtkImage)
	headerBar.Add(eventBoxImageHeaderBar)

	// *gtk.NoteBook Tab container
	obj, _ = mainGUI.builder.GetObject("noteBookApp")
	noteBookApp := obj.(*gtk.Notebook)
	noteBookApp.Connect("switch-page", func(notebook *gtk.Notebook, page *gtk.Widget, pageNum int) {
		if pageNum == pageContentGlobalHotkeys {
			if mainGUI.contentTabVentanas.expander.GetExpanded() {
				mainGUI.contentTabVentanas.expander.Activate()
			}
		}
	})

	// Control button to manage the global hotkey listener
	obj, _ = mainGUI.builder.GetObject("buttonControlListener")
	mainGUI.buttonControlListener = obj.(*gtk.Button)
	mainGUI.buttonControlListener.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() { button.SetSensitive(false) })
			time.Sleep(time.Second / 2)
			glib.IdleAdd(func() {
				// Emit signal to start/stop the global hotkey listener
				_, _ = mainGUI.application.Emit(signalControlListener, glib.TYPE_NONE, !listenerState, true)
				button.SetSensitive(true)
			})
		}()
	})

	// Label of button that manages the global hotkey listener
	obj, _ = mainGUI.builder.GetObject("labelButtonControlListener")
	mainGUI.labelButtonControlListener = obj.(*gtk.Label)

	// Image of button that manages the global hotkey listener
	obj, _ = mainGUI.builder.GetObject("imageButtonControlListener")
	mainGUI.imageButtonControlListener = obj.(*gtk.Image)

	// Hide window button
	obj, _ = mainGUI.builder.GetObject("buttonHideWindow")
	buttonHideWindow := obj.(*gtk.Button)
	buttonHideWindow.Connect("clicked", func(button *gtk.Button) { mainGUI.window.Hide() })

	// Reboot button
	obj, _ = mainGUI.builder.GetObject("buttonReboot")
	buttonReboot := obj.(*gtk.Button)
	buttonReboot.Connect(
		"clicked",
		func(button *gtk.Button) { _, _ = mainGUI.application.Emit(signalReboot, glib.TYPE_NONE) },
	)

	// Exit button
	obj, _ = mainGUI.builder.GetObject("buttonExit")
	buttonExit := obj.(*gtk.Button)
	buttonExit.Connect(
		"clicked",
		func(button *gtk.Button) { _, _ = mainGUI.application.Emit(signalExit, glib.TYPE_NONE) },
	)

	if showWindow {
		mainGUI.window.ShowAll()
	}
}

// Config function to set all icons
func (mainGUI *MainGUI) setIconsUI() {
	// Image refresh classes button
	obj, _ := mainGUI.builder.GetObject("imageRefreshClasses")
	image := obj.(*gtk.Image)
	image.SetFromPixbuf(getPixBufAtSize("refresh-pink.png", 24, 24))

	// Image refresh windows button
	obj, _ = mainGUI.builder.GetObject("imageRefreshWindows")
	image = obj.(*gtk.Image)
	image.SetFromPixbuf(getPixBufAtSize("refresh-green.png", 24, 24))

	// Image restore default order button
	obj, _ = mainGUI.builder.GetObject("imageRestoreDefaultOrder")
	image = obj.(*gtk.Image)
	image.SetFromPixbuf(getPixBufAtSize("restore.png", 24, 24))

	// Image hide window button
	obj, _ = mainGUI.builder.GetObject("imageHideWindow")
	image = obj.(*gtk.Image)
	image.SetFromPixbuf(getPixBufAtSize("hide_window.png", 24, 24))

	// Image reset button
	obj, _ = mainGUI.builder.GetObject("imageRestartApp")
	image = obj.(*gtk.Image)
	image.SetFromPixbuf(getPixBufAtSize("restart.png", 24, 24))

	// Image exit button
	obj, _ = mainGUI.builder.GetObject("imageExitApp")
	image = obj.(*gtk.Image)
	image.SetFromPixbuf(getPixBufAtSize("exit.png", 24, 24))
}

/*
Function that shows a message using a *gtk.MessageDialog.

Parameters:
  - messageType: gtk.MessageType
  - msg: Main message of dialog
  - msg2: Secondary message
*/
func (mainGUI *MainGUI) showMessageDialog(messageType gtk.MessageType, msg string, msg2 string) {
	dialog := gtk.MessageDialogNew(mainGUI.window, gtk.DIALOG_MODAL, messageType, gtk.BUTTONS_OK, "%s", msg)
	if len(msg2) > 0 {
		dialog.FormatSecondaryText("%s", msg2)
	}
	dialog.SetPosition(gtk.WIN_POS_CENTER)
	dialog.SetTransientFor(mainGUI.window)
	dialog.SetTitle(title)
	dialog.SetIcon(defaultAppIcon)
	dialog.Run()
	dialog.Destroy()
}

// PresentWindow Present Main Window
func (mainGUI *MainGUI) PresentWindow() {
	mainGUI.window.Present()
}

// UpdateListenerState Updates global hotkey listener state
func (mainGUI *MainGUI) UpdateListenerState(newState bool) {
	listenerState = newState
	if mainGUI == nil {
		return
	}
	// Window icon
	windowIconName := iconFileName
	if !newState {
		windowIconName = iconFileNameDisabled
	}
	defaultAppIcon = getPixBuf(funcGetResource(windowIconName))
	mainGUI.window.SetIcon(defaultAppIcon)
	headerBargtkImage.SetFromPixbuf(func(pixbuf *gdk.Pixbuf, err error) *gdk.Pixbuf {
		return pixbuf
	}(defaultAppIcon.ScaleSimple(25, 25, gdk.INTERP_HYPER)))

	// Label of button
	textLabel := funcGetStringResource("disable_keyboard_listener")
	if !newState {
		textLabel = funcGetStringResource("enable_keyboard_listener")
	}
	mainGUI.labelButtonControlListener.SetMarkup(textLabel)

	// Icon of button
	iconName := "stop-listener.png"
	if !newState {
		iconName = "start-listener.png"
	}
	mainGUI.imageButtonControlListener.SetFromPixbuf(getPixBufAtSize(iconName, 32, 32))
}
