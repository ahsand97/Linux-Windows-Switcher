package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"linux-windows-switcher/appindicator"
	"linux-windows-switcher/gui"
	"linux-windows-switcher/keyboard"
	"linux-windows-switcher/libs/glibown"
	"linux-windows-switcher/libs/xlib"

	"github.com/Xuanwo/go-locale"
	"github.com/bigkevmcd/go-configparser"
	"github.com/chigopher/pathlib"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
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

// This function is for configuration where all the custom signals are created and their callbacks.
// Callback of "startup" signal of the application.
func (app *mainApplication) startup() {
	app.config = configparser.New()
	if result, _ := configFile.IsFile(); result {
		config, err := configparser.NewConfigParserFromFile(configFile.String())
		if err == nil {
			app.config = config
		}
	}

	// --------------------------------- APPLICATION CUSTOM SIGNALS -----------------------------

	// Signal to open main window
	_, _ = glib.SignalNew("app-open-window")
	// Handler
	app.application.Connect("app-open-window", func(application *gtk.Application) { app.gui.PresentWindow() })

	// Signal to restart application
	_, _ = glib.SignalNew("app-restart")
	// Handler
	app.application.Connect("app-restart", func(application *gtk.Application) {
		application.Quit()
		command := fmt.Sprintf("'%s' %s", getPathExecutbale(false), strings.Join(os.Args[1:], " "))
		fmt.Println("Restarting app, using command: ", command)
		_ = exec.Command("bash", "-c", command).Start()
	})

	// Signal to close app
	_, _ = glib.SignalNew("app-exit")
	// Handler
	app.application.Connect("app-exit", func(application *gtk.Application) {
		keyboard.ExitListener()
		xlib.CloseDisplay() // Close connection to X server
		application.Quit()
	})

	// Signal to get data from config file
	_, _ = glibown.SignalNewV(
		"app-get-config",
		glib.TYPE_STRING,
		2,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
	)
	// Handler
	app.application.Connect(
		"app-get-config",
		func(application *gtk.Application, section string, option string) string {
			result := ""
			if exists, _ := app.config.HasOption(section, option); exists {
				result, _ = app.config.Get(section, option)
				result = strings.ReplaceAll(result, " ", "")
			}
			return result
		},
	)

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
	app.application.Connect(
		"app-listener-sync-state",
		func(application *gtk.Application, state bool) {
			app.gui.UpdateListenerState(state)
			app.appIndicator.UpdateIconState(state)
		},
	)

	// Signal to manage the keyboard listener
	_, _ = glibown.SignalNewV(
		"app-listener-keyboard",
		glib.TYPE_NONE,
		2,
		glib.TYPE_BOOLEAN,
		glib.TYPE_BOOLEAN,
	)

	// Signal to establish the global hotkey
	_, _ = glib.SignalNew("app-listener-set-hotkeys")

	// Signal to establish the new windows order
	_, _ = glibown.SignalNewV(
		"app-set-order",
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
// This function initializes the UI and the app if it hasn't started yet, otherwise it shows the main window
func (app *mainApplication) activate(debug bool) {
	if app.gui == nil {
		app.keyboardListener = keyboard.NewListenerKeyBoard(app.application, debug)
		app.appIndicator = appindicator.NewAppIndicator(
			app.application,
			[]string{iconFileDisabled.String(), iconFile.String()},
			gui.GetTitle(),
			getStringResource,
		)
		app.gui = gui.NewMainGUI(app.application, showWindow, getResource, getStringResource)
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

/*
Function that returns path of executable.

Parameter:
  - appimage: Whether to return the path of the APPDIR if it's running from the .AppImage
*/
func getPathExecutbale(appimage bool) string {
	var pathExecutable string
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

// Returns a string using localizer
func getStringResource(id string) string {
	msg, err := localizer.LocalizeMessage(&i18n.Message{ID: id})
	if err != nil {
		return ""
	}
	return msg
}

// Constants
const (
	appId                = "ahsand97.linux-windows-switcher-gotk3"
	configFileName       = "linux-windows-switcher-config.ini"
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
	showWindow       = true
	localizer        *i18n.Localizer
)

// Function that sets-up the locale configuration
func initLocalization() {
	// Default language
	defaultLanguage := language.English

	// Bundle
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// String resources
	stringResources := map[language.Tag]string{
		language.English: "strings-en.json",
		language.Spanish: "strings-es.json",
		language.French:  "strings-fr.json",
	}

	// Load string resources
	for language_, stringResource := range stringResources {
		messageFile, _ := bundle.LoadMessageFileFS(resources, filepath.Join(resourcesFolderName, stringResource))
		_ = bundle.AddMessages(language_, messageFile.Messages...)
	}

	// Languages
	var languages []string

	// Get current locale
	tag, _ := locale.Detect()
	currentLanguage, _ := tag.Base()
	currentLanguageStr := currentLanguage.String()
	languages = append(languages, currentLanguageStr)

	// Get all allowed locales
	tags, _ := locale.DetectAll()
	for _, tag_ := range tags {
		lang, _ := tag_.Base()
		if lang.String() == currentLanguage.String() {
			continue
		}
		languages = append(languages, lang.String())
	}
	if !slices.Contains(languages, defaultLanguage.String()) {
		languages = append(languages, defaultLanguage.String())
	}

	// Supported languages
	supportedLanguages := []string{}
	for _, lang := range languages {
		for tag := range stringResources {
			if tag.String() == lang {
				supportedLanguages = append(supportedLanguages, lang)
				break
			}
		}
	}

	// New localizer to get strings based on locale language_
	localizer = i18n.NewLocalizer(bundle, supportedLanguages...)
}

func main() {
	// CLI Flags
	hideFlag := flag.Bool(
		"hide",
		false,
		"Start the application only in the tray area (appindicator), not showing the main window.",
	)
	debugFlag := flag.Bool("debug", false, "Display debug information, keyboard events.")

	// Parse the flags
	flag.Parse()

	showWindow = !*hideFlag

	// Init Localization
	initLocalization()

	// App creation
	application, err := gtk.ApplicationNew(appId, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("An error occurred creating the application. ", err)
	}

	xlib.OpenDisplay()     // Open connection to X server
	glib.SetPrgname(appId) // Setting the property "WM_CLASS"
	configFile = pathlib.NewPath(getPathExecutbale(false)).Parent().Join(configFileName)
	iconFile = pathlib.NewPath(getPathExecutbale(true)).Parent().Join(resourcesFolderName, iconFileName)
	iconFileDisabled = pathlib.NewPath(getPathExecutbale(true)).Parent().Join(resourcesFolderName, iconDisabledFileName)

	mainApplication := newApplication(application)
	mainApplication.application.Connect("startup", func(application *gtk.Application) { mainApplication.startup() })
	mainApplication.application.Connect(
		"activate",
		func(application *gtk.Application) { mainApplication.activate(*debugFlag) },
	)
	mainApplication.application.Run(nil)
}
