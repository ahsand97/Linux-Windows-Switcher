package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bigkevmcd/go-configparser"
	"github.com/chigopher/pathlib"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

import (
	"linux-windows-switcher/appindicator"
	"linux-windows-switcher/gui"
	"linux-windows-switcher/keyboard"
	"linux-windows-switcher/libs/glibown"
)

// Main struct, it has the main components of the application
type mainApplication struct {
	application      *gtk.Application
	appIndicator     *appindicator.Indicator
	gui              *gui.MainGUI
	config           *configparser.ConfigParser
	keyboardListener *keyboard.ListenerKeyboard
}

// Constructor mainApplication
func newApplication(application *gtk.Application) *mainApplication {
	return &mainApplication{application: application}
}

// This function is a setup where all the custom signals are created and their callbacks
// Callback of "startup" signal of the application.
func (app *mainApplication) startup() {
	if result, _ := configFile.IsFile(); result {
		app.config, _ = configparser.NewConfigParserFromFile(configFile.String())
	} else {
		app.config = configparser.New()
	}

	// --------------------------------- SEÑALES DE LA APLICACIÓN -----------------------------

	// Señal para abrir la ventana principal
	_, _ = glib.SignalNew("app-abrir-ventana")
	// Handler
	app.application.Connect("app-abrir-ventana", func(application *gtk.Application) {
		app.gui.PresentWindow()
	})

	// Señal para reiniciar la aplicación
	_, _ = glib.SignalNew("app-reiniciar")
	// Handler
	app.application.Connect("app-reiniciar", func(application *gtk.Application) {
		application.Quit()
		comando := fmt.Sprintf("'%s' %s", getPathExecutbale(false), strings.Join(os.Args[1:], " "))
		_ = exec.Command("bash", "-c", comando).Start()
	})

	// Signal to close app
	_, _ = glib.SignalNew("app-salir")
	// Handler
	app.application.Connect("app-salir", func(application *gtk.Application) {
		keyboard.ExitListener()
		application.Quit()
	})

	// Signal to get data from config file
	_, _ = glibown.SignalNewV("app-get-config", glib.TYPE_STRING, 2, glib.TYPE_STRING, glib.TYPE_STRING)
	// Handler
	app.application.Connect("app-get-config", func(application *gtk.Application, section string, option string) string {
		result := ""
		if exists, _ := app.config.HasOption(section, option); exists {
			result, _ = app.config.Get(section, option)
			result = strings.ReplaceAll(result, " ", "")
		}
		return result
	})

	// Signal to update config file
	_, _ = glibown.SignalNewV(
		"app-update-config",
		glib.TYPE_BOOLEAN,
		3,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
	)
	// Handler
	app.application.Connect(
		"app-update-config",
		func(application *gtk.Application, section string, option string, value string) bool {
			err := app.updateConfig(section, option, value)
			return err == nil
		},
	)

	// Signal to synchronize the keyboard listener's state with the UI and AppIndicator
	_, _ = glibown.SignalNewV("app-listener-sync-state", glib.TYPE_NONE, 1, glib.TYPE_BOOLEAN)
	// Handler
	app.application.Connect("app-listener-sync-state", func(application *gtk.Application, state bool) {
		app.gui.UpdateListenerState(state)
		app.appIndicator.UpdateIconState(state)
	})

	// Signal to manage the keyboard listener
	_, _ = glibown.SignalNewV("app-listener-keyboard", glib.TYPE_NONE, 2, glib.TYPE_BOOLEAN, glib.TYPE_BOOLEAN)

	// Signal to stabligh the global hotkey
	_, _ = glib.SignalNew("app-listener-set-atajos")

	// Signal to stablish the new windows order
	_, _ = glibown.SignalNewV(
		"app-establecer-orden",
		glib.TYPE_NONE,
		3,
		glib.TYPE_BOOLEAN,
		glib.TYPE_BOOLEAN,
		glib.TYPE_BOOLEAN,
	)

	// Signal to delete a window from the order
	_, _ = glibown.SignalNewV("app-delete-window-order", glib.TYPE_NONE, 1, glib.TYPE_STRING)
}

// Callback of signal "activate" of the application
func (app *mainApplication) activate() {
	if app.gui == nil {
		app.keyboardListener = keyboard.NewListenerKeyBoard(app.application)
		app.appIndicator = appindicator.NewAppIndicator(
			app.application,
			[]string{iconFileDisabled.String(), iconFile.String()},
			gui.GetTitle(),
		)
		app.gui = gui.NewMainGUI(app.application, getResource, mostrarVentana)
	} else {
		app.gui.PresentWindow()
	}
}

// Update config file
func (app *mainApplication) updateConfig(section string, option string, value string) error {
	_ = app.config.AddSection(section)
	_ = app.config.Set(section, option, value)
	return app.config.SaveWithDelimiter(configFile.String(), "=")
}

// Returns path of executable, if it's running via AppImage it returns the path of .AppImage file
//
// Param:
//   - appimage:bool Whether to return the path of the .AppImage or not
func getPathExecutbale(appimage bool) string {
	pathExecutable := ""
	if value, exists := os.LookupEnv("APPIMAGE"); exists {
		pathExecutable = value
		if appimage {
			if value, exists = os.LookupEnv("APPDIR"); exists {
				path_ := pathlib.NewPath(value).Join("usr", "src", "linux-windows-switcher")
				pathExecutable = path_.String()
			}
		}
	} else {
		path, _ := os.Executable()
		pathExecutable = path
	}
	return pathExecutable
}

// Returns a resource from the embed filesystem
func getResource(filename string) []byte {
	return func(content []byte, err error) []byte {
		return content
	}(resources.ReadFile(filepath.Join(resourcesFolderName, filename)))
}

// Constants
const (
	appId                = "ahsan.windows-switcher-gotk3"
	configFileName       = "config.ini"
	resourcesFolderName  = "resources"
	iconFileName         = "tabs.png"
	iconDisabledFileName = "tabs-disabled.png"
)

//go:embed resources/*
var resources embed.FS

// Globals
var (
	configFile       *pathlib.Path
	iconFile         *pathlib.Path
	iconFileDisabled *pathlib.Path
	mostrarVentana   = true
)

func main() {
	var args []string
	for index, value := range os.Args {
		if value == "--hide" {
			mostrarVentana = false
		} else {
			args = append(args, os.Args[index])
		}
	}

	// App creation
	application, err := gtk.ApplicationNew(appId, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("An error occurred creating the application. ", err)
	}

	glib.SetPrgname(appId) // Setting the property "WM_CLASS"
	configFile = pathlib.NewPath(getPathExecutbale(false)).Parent().Join(configFileName)
	iconFile = pathlib.NewPath(getPathExecutbale(true)).Parent().Join(resourcesFolderName, iconFileName)
	iconFileDisabled = pathlib.NewPath(getPathExecutbale(true)).Parent().Join(resourcesFolderName, iconDisabledFileName)

	mainApplication := newApplication(application)
	mainApplication.application.Connect("startup", func(application *gtk.Application) {
		mainApplication.startup()
	})
	mainApplication.application.Connect("activate", func(application *gtk.Application) {
		mainApplication.activate()
	})
	mainApplication.application.Run(args)
	// http://getfr.org/pub/dragonfly-release/usr-local-share/gtk-doc/html/libwnck-3.0/WnckWindow.html#wnck-window-get-icon FALTA
}
