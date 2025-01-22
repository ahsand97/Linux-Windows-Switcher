package appindicator

import (
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type Indicator struct {
	application *gtk.Application
	indicator   *appindicator.Indicator
}

var (
	icons                 []string
	optionControlListener *gtk.MenuItem
	funcGetStringResource func(id string) string // Anonymous function that returns a string from the localizer
)

const (
	// Signals used
	signalOpenMainWindow  = "app-open-window"
	signalReboot          = "app-restart"
	signalExit            = "app-exit"
	signalControlListener = "app-listener-keyboard"

	// Index of icons inside slice "icons"
	iconInactive = 0
	iconActive   = 1
)

// NewAppIndicator Constructor AppIndicator
func NewAppIndicator(
	application *gtk.Application,
	iconPaths []string,
	title string,
	funcGetStringResource_ func(id string) string,
) *Indicator {
	// Assign function coming from main module
	funcGetStringResource = funcGetStringResource_

	// Fill global vars with their initial values
	icons = iconPaths

	// Creation of indicator object
	indicator := appindicator.New(
		application.GetApplicationID(),
		icons[iconInactive],
		appindicator.CategoryApplicationStatus,
	)
	indicator.SetTitle(title)
	indicator.SetStatus(appindicator.StatusActive)

	// Creation of main struct to handle the appIndicator logic
	indicator_ := &Indicator{application: application, indicator: indicator}
	indicator_.setupIndicator()
	return indicator_
}

// Config function
func (indicator *Indicator) setupIndicator() {
	// appindicator menu
	menu, _ := gtk.MenuNew()
	defer menu.ShowAll()

	openMainWindow, _ := gtk.MenuItemNewWithLabel(funcGetStringResource("indicator_menu_open_main_window"))
	openMainWindow.Connect(
		"activate",
		func(menuItem *gtk.MenuItem) { _, _ = indicator.application.Emit(signalOpenMainWindow, glib.TYPE_NONE) },
	)

	optionControlListener, _ = gtk.MenuItemNewWithLabel(funcGetStringResource("enable_keyboard_listener"))
	optionControlListener.Connect("activate", func(menuItem *gtk.MenuItem) {
		newState := menuItem.GetLabel() == funcGetStringResource("enable_keyboard_listener")
		// Emit signal to start/stop the keyboard listener
		_, _ = indicator.application.Emit(signalControlListener, glib.TYPE_NONE, newState, true)
	})

	restartApp, _ := gtk.MenuItemNewWithLabel(funcGetStringResource("restart"))
	restartApp.Connect("activate", func(menuItem *gtk.MenuItem) {
		// Emit signal to restart  application
		_, _ = indicator.application.Emit(signalReboot, glib.TYPE_NONE)
	})

	exitApp, _ := gtk.MenuItemNewWithLabel(funcGetStringResource("exit"))
	exitApp.Connect("activate", func(menuItem *gtk.MenuItem) {
		// Emit signal to exit application
		_, _ = indicator.application.Emit(signalExit, glib.TYPE_NONE)
	})

	// Add subitems to appindicator menu
	menu.Add(openMainWindow)
	menu.Add(optionControlListener)
	menu.Add(restartApp)
	menu.Add(exitApp)

	// Set menu to appindicator
	indicator.indicator.SetMenu(menu)
}

// UpdateIconState Update the icon state
func (indicator *Indicator) UpdateIconState(newState bool) {
	icon := icons[iconActive]
	if !newState {
		icon = icons[iconInactive]
	}
	indicator.indicator.SetIcon(icon)

	label := funcGetStringResource("disable_keyboard_listener")
	if !newState {
		label = funcGetStringResource("enable_keyboard_listener")
	}
	optionControlListener.SetLabel(label)
}
