package gui

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gotk3/gotk3/glib"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

import (
	"linux-windows-switcher/libs/goxdo"
)

type MainGUI struct {
	getResource                func(resource string) []byte // Función que retorna el recurso especificado en un slice de bytes
	application                *gtk.Application
	builder                    *gtk.Builder
	window                     *gtk.Window
	contentTabVentanas         *contenTabVentanas
	contentTabAtajos           *contentTabAtajos
	botonControlListener       *gtk.Button
	labelBotonControlListener  *gtk.Label
	imagenBotonControlListener *gtk.Image
	xdotool                    *goxdo.Xdo
}

type ventana struct {
	id     string
	clase  string
	titulo string
	orden  int
}

// Globals
var (
	mostrarVentana  bool
	uiFile          string
	icon            *gdk.Pixbuf // Icono por defecto de la aplicación
	gtkImage        *gtk.Image  // Imagen de la cabecera
	ordenActual     []ventana
	ordenPorDefecto []ventana
	listenerState   bool
	awkFile         string
)

// Constants
const (
	// Nombre de los resources
	uiFileName           = "Application.xml"
	iconFileName         = "tabs.png"
	iconFileNameDisabled = "tabs-disabled.png"
	consultaAwkFileName  = "consultaGetVentanasAbiertas.awk"

	// Título de la aplicación
	title = "Linux Windows Switcher"

	// Nombres de las señales de la aplicación *gtk.Application usadas
	signalReiniciar       = "app-reiniciar"
	signalSalir           = "app-salir"
	signalGetConfig       = "app-get-config"
	signalUpdateConfig    = "app-update-config"
	signalControlListener = "app-listener-keyboard"
	signalEstablecerOrden = "app-establecer-orden"
	signalDeleteRow       = "app-delete-window-order"

	// Indice de las página del gtkNotebook para la configuración de atajos del teclado
	pageContenidoAtajosTeclado = 1
)

// NewMainGUI Constructor MainGUI
func NewMainGUI(application *gtk.Application, funcGetResource func(resource string) []byte, showWindow bool) *MainGUI {
	mostrarVentana = showWindow

	mainGui := &MainGUI{application: application, getResource: funcGetResource, xdotool: goxdo.NewXdo()}
	uiFile = string(mainGui.getResource(uiFileName))
	icon = getPixBuf(mainGui.getResource(iconFileName))
	awkFile = string(mainGui.getResource(consultaAwkFileName))
	mainGui.builder = mainGui.getNewBuilder()

	mainGui.setupUi()
	mainGui.contentTabVentanas = mainGui.newContentTabVentanas()
	mainGui.contentTabAtajos = mainGui.newContentTabAtajos()
	return mainGui
}

// Función de configuración
func (mainGUI *MainGUI) setupUi() {
	// Main window
	obj, _ := mainGUI.builder.GetObject("mainWindow")
	mainGUI.window = obj.(*gtk.Window)
	mainGUI.window.SetApplication(mainGUI.application)
	mainGUI.window.SetIcon(icon)
	mainGUI.window.SetTitle(title)
	mainGUI.window.HideOnDelete() // Añade evento para que la ventana se oculte y no se cierre la aplicación

	// Indica si el tamaño inicial de la ventana ya fue establecido (cuando se abre)
	sizeInicialEstablecido := false
	mainGUI.window.Connect("show", func(window *gtk.Window) {
		if !sizeInicialEstablecido {
			// Se redimensiona la ventana a 1x1, así la obliga a obtener su tamaño mínimo permitido
			window.Resize(1, 1)
			sizeInicialEstablecido = true
		}
	})

	// Barra de título
	obj, _ = mainGUI.builder.GetObject("headerBarMainWindow")
	headerBar := obj.(*gtk.HeaderBar)

	// Imagen que aparece al principio de la barra de título
	gtkImage, _ = gtk.ImageNewFromPixbuf(func(pixbuf *gdk.Pixbuf, err error) *gdk.Pixbuf {
		return pixbuf
	}(icon.ScaleSimple(25, 25, gdk.INTERP_HYPER)))
	gtkImage.SetMarginTop(5)
	gtkImage.SetMarginBottom(5)

	// EventBox que será el contenedor de gtkImage para poder recibir eventos
	eventBoxImageHeaderBar, _ := gtk.EventBoxNew()
	eventBoxImageHeaderBar.Connect("button-release-event", func(box *gtk.EventBox, event *gdk.Event) bool {
		eventButton := gdk.EventButtonNewFromEvent(event)
		if eventButton.Button() == gdk.BUTTON_PRIMARY {
			mainGUI.xdotool.ClickWindow(goxdo.CURRENTWINDOW, goxdo.MBUTTON_RIGHT)
		}
		return false
	})

	// Se añade gtkImage al contenedor eventBoxImageHeaderBar
	eventBoxImageHeaderBar.Add(gtkImage)
	// Se añade eventBoxImageHeaderBar al headerBar
	headerBar.Add(eventBoxImageHeaderBar)

	// gtkNoteBook (Contenedor de los Tabs)
	obj, _ = mainGUI.builder.GetObject("noteBookApp")
	noteBookApp := obj.(*gtk.Notebook)
	noteBookApp.Connect("switch-page", func(notebook *gtk.Notebook, page *gtk.Widget, pageNum int) {
		if pageNum == pageContenidoAtajosTeclado {
			if mainGUI.contentTabVentanas.expansor.GetExpanded() {
				mainGUI.contentTabVentanas.expansor.Activate()
			}
		}
	})

	// Botón de control del Listener del Teclado
	obj, _ = mainGUI.builder.GetObject("botonControlListener")
	mainGUI.botonControlListener = obj.(*gtk.Button)
	mainGUI.botonControlListener.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() {
				button.SetSensitive(false)
			})
			time.Sleep(time.Second / 2)
			glib.IdleAdd(func() {
				// Se emite la señal para iniciar/detener el listener del teclado
				_, _ = mainGUI.application.Emit(signalControlListener, !listenerState, true)
				button.SetSensitive(true)
			})
		}()
	})

	// Label del botón del control del Listener del Teclado
	obj, _ = mainGUI.builder.GetObject("labelBotonControlListener")
	mainGUI.labelBotonControlListener = obj.(*gtk.Label)

	// Imágen del botón del control del Listener del Teclado
	obj, _ = mainGUI.builder.GetObject("imagenBotonControlListener")
	mainGUI.imagenBotonControlListener = obj.(*gtk.Image)

	// Botón ocultar ventana
	obj, _ = mainGUI.builder.GetObject("botonOcultarVentana")
	botonOcultarVentana := obj.(*gtk.Button)
	botonOcultarVentana.Connect("clicked", func(button *gtk.Button) {
		mainGUI.window.Hide()
	})

	// Botón reiniciar
	obj, _ = mainGUI.builder.GetObject("botonReiniciar")
	botonReiniciar := obj.(*gtk.Button)
	botonReiniciar.Connect("clicked", func(button *gtk.Button) {
		_, _ = mainGUI.application.Emit(signalReiniciar)
	})

	// Botón salir
	obj, _ = mainGUI.builder.GetObject("botonSalir")
	botonSalir := obj.(*gtk.Button)
	botonSalir.Connect("clicked", func(button *gtk.Button) {
		_, _ = mainGUI.application.Emit(signalSalir)
	})

	// Si se pasó el parámetro --abrir cuando se abrió la aplicación se procede a mostrar la ventana
	if mostrarVentana {
		mainGUI.window.ShowAll()
	}
}

// Función que retorna una nueva instancia de un *gtk.Builder usando el
// mismo archivo de definición de la interfaz
func (mainGUI *MainGUI) getNewBuilder() *gtk.Builder {
	return func(builder *gtk.Builder, err error) *gtk.Builder {
		return builder
	}(gtk.BuilderNewFromString(uiFile))
}

// Función que muestra un mensaje en un Gtk.MessageDialog
func (mainGUI *MainGUI) mostrarMensajeDialog(messageType gtk.MessageType, msg string, msg2 string) {
	dialog := gtk.MessageDialogNew(mainGUI.window, gtk.DIALOG_MODAL, messageType, gtk.BUTTONS_OK, msg)
	if len(msg2) > 0 {
		dialog.FormatSecondaryText(msg2)
	}
	dialog.SetPosition(gtk.WIN_POS_CENTER)
	dialog.SetTitle(title)
	dialog.SetIcon(icon)
	dialog.Connect("response", func(dialog *gtk.MessageDialog, response int) {
		dialog.Destroy()
	})
	dialog.Run()
}

// PresentWindow Función expuesta wrapper del método gtk.Window.Present
func (mainGUI *MainGUI) PresentWindow() {
	mainGUI.window.Present()
}

// UpdateListenerState Actualiza el estado del boolean listenerState
func (mainGUI *MainGUI) UpdateListenerState(newState bool) {
	listenerState = newState
	if mainGUI != nil {
		iconName := iconFileName
		if !newState {
			iconName = iconFileNameDisabled
		}
		icon = getPixBuf(mainGUI.getResource(iconName))
		mainGUI.window.SetIcon(icon)
		gtkImage.SetFromPixbuf(func(pixbuf *gdk.Pixbuf, err error) *gdk.Pixbuf {
			return pixbuf
		}(icon.ScaleSimple(25, 25, gdk.INTERP_HYPER)))

		textLabel := "Desactivar Listener del Teclado"
		if !newState {
			textLabel = "Activar Listener del Teclado"
		}
		mainGUI.labelBotonControlListener.SetText(textLabel)

		iconName = "emblem-unreadable"
		if !newState {
			iconName = "emblem-default"
		}
		mainGUI.imagenBotonControlListener.SetFromIconName(iconName, gtk.ICON_SIZE_BUTTON)
	}
}

// newVentana Constructor ventana
func newVentana(id string, clase string, titulo string, orden int) *ventana {
	ventana := &ventana{
		id:     id,
		clase:  clase,
		titulo: titulo,
		orden:  orden,
	}
	return ventana
}

// GetTitle retorna el titulo de la aplicación
func GetTitle() string {
	return title
}

// Función que retorna un *gdk.Pixbuf a partur de un slice de bytes
func getPixBuf(content []byte) *gdk.Pixbuf {
	return func(pixbuf *gdk.Pixbuf, err error) *gdk.Pixbuf {
		return pixbuf
	}(gdk.PixbufNewFromBytesOnly(content))
}

//-------------------------------------------------- CALLBACKS ATAJOS DE TECLADO -------------------------------------------

// Función para validar y activar el foco de la siguiente ventana (ya sea hacia adelante o hacia atrás)
func (mainGUI *MainGUI) moveEntreVentanas(backwards bool) {
	result, windowCurrent := mainGUI.xdotool.GetActiveWindow()
	fmt.Printf("(Callback) moveEntreVentanas(%t)\n", backwards)
	fmt.Printf("(Callback) ventanaActual: %d\n", windowCurrent)
	if result != 0 || int(windowCurrent) == 0 {
		fmt.Println("ERROR OBTENIENDO VENTANA ACTUAL, se llama de nuevo a la función")
		mainGUI.moveEntreVentanas(backwards)
	} else {
		ventanaActual := strconv.Itoa(int(windowCurrent))
		if len(ordenActual) == 1 && ventanaActual == ordenActual[0].id {
			return
		}
		indexActual := -1
		var indexSiguiente int
		for index, ventana := range ordenActual {
			if ventana.id == ventanaActual {
				indexActual = index
				break
			}
		}
		if backwards {
			indexSiguiente = indexActual - 1
			if indexSiguiente < 0 {
				indexSiguiente = len(ordenActual) - 1
			}
		} else {
			indexSiguiente = indexActual + 1
			if indexSiguiente >= len(ordenActual) {
				indexSiguiente = 0
			}
		}
		proximaVentanaIsValid := false
		comando := "wmctrl -lx | awk --non-decimal-data '{if ($2 == \"-1\") { next; } printf \"%d\\n\", $1; }'"
		res, _ := exec.Command("bash", "-c", comando).Output()
		for _, v := range strings.Split(strings.TrimSuffix(string(res), "\n"), "\n") {
			if v == ordenActual[indexSiguiente].id {
				proximaVentanaIsValid = true
				break
			}
		}
		llamadoRecursivo := false
		if proximaVentanaIsValid {
			fmt.Println("(Callback) Próxima ventana:", ordenActual[indexSiguiente])
			// Activación ventana por medio de xdotool
			windowId, _ := strconv.Atoi(ordenActual[indexSiguiente].id)
			mainGUI.xdotool.WindowActivate(goxdo.Window(windowId))
			mainGUI.xdotool.WaitForWindowActivate(goxdo.Window(windowId), true)
		} else {
			fmt.Println("(Callback) Próxima ventana:", ordenActual[indexSiguiente], "no es válida")
			llamadoRecursivo = true
			fmt.Println("-------------------------------------------------------")
			fmt.Println("ATAJO ACTIVADO por llamado recursivo")
		}
		if llamadoRecursivo {
			glib.IdleAdd(func() {
				_, _ = mainGUI.application.Emit(signalDeleteRow, strconv.Itoa(indexSiguiente))
				mainGUI.moveEntreVentanas(backwards)
			})
		}
	}
}

// Función para avanzar a la ventana siguiente siguiendo el orden, es el callback del atajo "Avanzar a la ventana siguiente"
func (mainGUI *MainGUI) avanzarDeVentana() {
	if len(ordenActual) > 0 {
		mainGUI.moveEntreVentanas(false)
	}
}

// Función para retroceder a la ventana anterior siguiendo el orden, es el callback del atajo "Retroceder a la ventana anterior"
func (mainGUI *MainGUI) retrocederDeVentana() {
	if len(ordenActual) > 0 {
		mainGUI.moveEntreVentanas(true)
	}
}
