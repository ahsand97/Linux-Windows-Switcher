package gui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

import (
	"linux-windows-switcher/keyboard"
	"linux-windows-switcher/libs/glibown"
)

type contentTabAtajos struct {
	MainGUI
	listBoxAtajos *gtk.ListBox
	listaAtajos   []*keyboard.Atajo
}

const (
	// Sección del archivo de configuración para los atajos
	sectionAtajos = "atajos"

	// Nombre de las señales de la *gtk.Application usadas
	signalSetAtajos = "app-listener-set-atajos"

	// Nombre de los resources
	iconoTeclado = "keyboard.png"
	iconoTeclas  = "keys.png"

	// Nombre de los atajos
	avanzarVentana    = "Avanzar a la ventana siguiente"
	retrocederVentana = "Retroceder a la ventana anterior"
	atajoEmpty        = "---"

	// Límite de teclas del atajo
	limiteTeclas = 3
)

// Información de los atajos de la aplicación (Título - nombre de la opción del archivo de configuración)
var infoAtajos = map[string]string{avanzarVentana: "avanzar_ventana", retrocederVentana: "retroceder_ventana"}

// newContentTabAtajos Constructor
func (mainGUI *MainGUI) newContentTabAtajos() *contentTabAtajos {
	contentTabAtajos := &contentTabAtajos{MainGUI: *mainGUI}

	contentTabAtajos.setupContentTabAtajos()
	return contentTabAtajos
}

// Función de configuración
func (contentTabAtajos *contentTabAtajos) setupContentTabAtajos() {
	// Señal para actualizar los datos del ListBoxRow cuando se establezca/modifique algún atajo
	_, _ = glib.SignalNew("listbox-update-atajo")

	obj, _ := contentTabAtajos.MainGUI.builder.GetObject("contenedorTituloAtajos")
	contenedorTituloAtajos := obj.(*gtk.Box)
	imageContenedorTituloAtajos := contentTabAtajos.getGtkImageFromResource(iconoTeclado, 48, 48)
	imageContenedorTituloAtajos.SetMarginBottom(10)
	contenedorTituloAtajos.PackEnd(imageContenedorTituloAtajos, false, false, 0)

	// ------------------------------------------- ATAJOS DE LA APLICACIÓN ------------------------------------
	// Función anónima que crea un *keyboard.Atajo y lo añade a la lista de Atajos
	funcAddAtajo := func(atajo string, callback func()) {
		contentTabAtajos.listaAtajos = append(contentTabAtajos.listaAtajos, keyboard.NewAtajo(atajo, callback))
	}

	// Se crean los objetos tipo *keyboard.Atajo y se añaden a la lista de atajos
	funcAddAtajo(avanzarVentana, contentTabAtajos.MainGUI.avanzarDeVentana)
	funcAddAtajo(retrocederVentana, contentTabAtajos.MainGUI.retrocederDeVentana)

	obj, _ = contentTabAtajos.MainGUI.builder.GetObject("listBoxAtajosTeclado")
	contentTabAtajos.listBoxAtajos = obj.(*gtk.ListBox)

	// Se consulta la configuración actual de los atajos en el archivo de configuración y se añaden al listbox
	for _, atajo := range contentTabAtajos.listaAtajos {
		contentTabAtajos.getConfigAtajo(atajo)
		contentTabAtajos.listBoxAtajos.Add(contentTabAtajos.createListBoxRowAtajo(atajo))
	}
	// ----------------------------------------------------------------------------------------------------

	// Se conecta la señal para establecer los atajos en el listener
	contentTabAtajos.MainGUI.application.Connect(signalSetAtajos, func(application *gtk.Application) {
		keyboard.SetAtajos(contentTabAtajos.listaAtajos)
	})

	// Se emite la señal para poner activo el listener del teclado cuando la aplicación inicia
	_, _ = contentTabAtajos.MainGUI.application.Emit(signalControlListener, true, true)
}

// Función que obtiene la configuración de un atajo
func (contentTabAtajos *contentTabAtajos) getConfigAtajo(atajo *keyboard.Atajo) {
	result, _ := contentTabAtajos.MainGUI.application.Emit(
		signalGetConfig,
		sectionAtajos,
		infoAtajos[atajo.Nombre],
	)
	cadenaInfoAtajo := result.(string)
	sliceInfoAtajo := strings.Split(cadenaInfoAtajo, ":")
	lengthSliceInfoAtajos := len(sliceInfoAtajo)
	if lengthSliceInfoAtajos > 0 && lengthSliceInfoAtajos <= 2 {
		if lengthSliceInfoAtajos == 2 {
			if strings.ToLower(sliceInfoAtajo[1]) == "disabled" {
				atajo.Disabled = true
			}
		}
		sliceKeysAtajo := strings.Split(sliceInfoAtajo[0], ",")
		if len(sliceKeysAtajo) <= limiteTeclas {
			for _, key := range sliceKeysAtajo {
				key = strings.TrimSpace(key)
				keyVal := gdk.KeyvalFromName(key)
				if keyVal != gdk.KEY_VoidSymbol && !contains(atajo.Atajo, key) {
					atajo.Atajo = append(atajo.Atajo, key)
					atajo.AtajoKeyCodes = append(atajo.AtajoKeyCodes, keyVal)
				}
			}
		}
	}
}

// Función que crea un ListBoxRow listo para ser añadido a un ListBox de atajos
func (contentTabAtajos *contentTabAtajos) createListBoxRowAtajo(atajo *keyboard.Atajo) *gtk.ListBoxRow {
	builder := contentTabAtajos.MainGUI.getNewBuilder()

	obj, _ := builder.GetObject("listBoxRowAtajoTeclado")
	listBoxRowAtajoTeclado := obj.(*gtk.ListBoxRow)

	obj, _ = builder.GetObject("labelNombreAtajo")
	labelNombreAtajo := obj.(*gtk.Label)
	labelNombreAtajo.SetText(atajo.Nombre)
	labelNombreAtajo.SetTooltipText(atajo.Nombre)

	obj, _ = builder.GetObject("labelTeclasAtajo")
	labelTeclasAtajo := obj.(*gtk.Label)
	if len(atajo.Atajo) == 0 {
		labelTeclasAtajo.SetText(atajoEmpty)
	} else {
		labelTeclasAtajo.SetText(strings.Join(atajo.Atajo, " + "))
	}

	// Función anónima que actualiza el estado de un atajo cuando se habilita/deshabilita
	funcionUpdateAtajo := func(atajo keyboard.Atajo) bool {
		valorNuevoAtajo := strings.Join(atajo.Atajo, ",")
		if atajo.Disabled {
			valorNuevoAtajo += ":disabled"
		}
		result, _ := contentTabAtajos.MainGUI.application.Emit(
			signalUpdateConfig,
			sectionAtajos,
			infoAtajos[atajo.Nombre],
			valorNuevoAtajo,
		)
		return result.(bool)
	}

	// Botón cambiar atajo
	obj, _ = builder.GetObject("botonCambiarAtajo")
	botonCambiarAtajo := obj.(*gtk.Button)
	botonCambiarAtajo.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() {
				button.SetSensitive(false)
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				contentTabAtajos.mostrarDialogCambiarAtajo(button, atajo)
			})
		}()
	})

	// Botón desactivar atajo
	obj, _ = builder.GetObject("botonDesactivarAtajo")
	botonDesactivarAtajo := obj.(*gtk.ToggleButton)
	botonDesactivarAtajo.Connect("toggled", func(button *gtk.ToggleButton) {
		atajo.Disabled = button.GetActive()
		result := funcionUpdateAtajo(*atajo)
		if result {
			_, _ = contentTabAtajos.MainGUI.application.Emit(signalControlListener, true, false)
		}
		botonCambiarAtajo.SetSensitive(!button.GetActive())
	})

	// Se conecta la señal al botón cambiar atajo para actualizar un atajo
	botonCambiarAtajo.Connect(
		"listbox-update-atajo", func(button *gtk.Button) {
			result := funcionUpdateAtajo(*atajo)
			if result {
				labelTeclasAtajo.SetText(strings.Join(atajo.Atajo, " + "))
				botonDesactivarAtajo.SetSensitive(true)
			}
		},
	)

	// Si el atajo está deshabilitado desde el archivo de configuración y su botón asociado está apagado se enciende
	if atajo.Disabled && !botonDesactivarAtajo.GetActive() {
		botonDesactivarAtajo.SetActive(true)
	}

	botonCambiarAtajo.SetSensitive(!botonDesactivarAtajo.GetActive())
	botonDesactivarAtajo.SetSensitive(len(atajo.AtajoKeyCodes) > 0)
	return listBoxRowAtajoTeclado
}

// Función que crea un *gtk.Image a partir del nombre del recurso y se cambia el tamaño al especificado
func (contentTabAtajos *contentTabAtajos) getGtkImageFromResource(resource string, width int, height int) *gtk.Image {
	pixbuf := getPixBuf(contentTabAtajos.MainGUI.getResource(resource))
	var gtkImage *gtk.Image
	if width > 0 && height > 0 {
		pixbufResized, _ := pixbuf.ScaleSimple(width, height, gdk.INTERP_HYPER)
		gtkImage, _ = gtk.ImageNewFromPixbuf(pixbufResized)
	} else {
		gtkImage, _ = gtk.ImageNewFromPixbuf(pixbuf)
	}
	gtkImage.Show()
	return gtkImage
}

// Función que muestra el dialog para asignar/cambiar algún atajo
func (contentTabAtajos *contentTabAtajos) mostrarDialogCambiarAtajo(botonOrigen *gtk.Button, atajo *keyboard.Atajo) {
	// Se apaga el listener del teclado global
	_, _ = contentTabAtajos.MainGUI.application.Emit(signalControlListener, false, false)

	builder := contentTabAtajos.MainGUI.getNewBuilder()
	obj, _ := builder.GetObject("dialogCambiarAtajo")
	dialogCambiarAtajo := obj.(*gtk.Dialog)
	dialogCambiarAtajo.SetTitle(title + " - Atajo de Teclado")
	dialogCambiarAtajo.SetIcon(icon)

	obj, _ = builder.GetObject("contenedorTituloDialogCambiarAtajo")
	contenedorTituloDialogCambiarAtajo := obj.(*gtk.Box)
	image := contentTabAtajos.getGtkImageFromResource(iconoTeclas, 40, 40)
	contenedorTituloDialogCambiarAtajo.PackEnd(image, false, false, 10)

	obj, _ = builder.GetObject("labelInfoTeclasDialog")
	labelInfoTeclasDialog := obj.(*gtk.Label)

	textlabelInfoTeclasDialog, _ := labelInfoTeclasDialog.GetText()
	labelInfoTeclasDialog.SetText(textlabelInfoTeclasDialog + fmt.Sprintf("\"%s\".", atajo.Nombre))

	obj, _ = builder.GetObject("botonCancelarNuevoAtajo")
	botonCancelarNuevoAtajo := obj.(*gtk.Button)

	obj, _ = builder.GetObject("labelDialogCambiarAtajo")
	labelDialogCambiarAtajo := obj.(*gtk.Label)

	obj, _ = builder.GetObject("botonAceptarNuevoAtajo")
	botonAceptarNuevoAtajo := obj.(*gtk.Button)

	obj, _ = builder.GetObject("botonBorrarNuevoAtajo")
	botonBorrarNuevoAtajo := obj.(*gtk.Button)

	obj, _ = builder.GetObject("labelErrorNuevoAtajo")
	labelErrorNuevoAtajo := obj.(*gtk.Label)

	obj, _ = builder.GetObject("contenedorErrorDialogNuevoAtajo")
	contenedorErrorDialogNuevoAtajo := obj.(*gtk.Box)

	// Variables del dialog
	var sliceOfKeys []string
	var sliceofKeyVals []uint
	cantidadTeclasPressed := 0
	puedeEditar := true

	botonAceptarNuevoAtajo.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() {
				button.SetSensitive(false)
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				invalidAtajo := false
				error_ := ""
				if strings.Join(atajo.Atajo, " + ") == strings.Join(sliceOfKeys, " + ") {
					invalidAtajo = true
					error_ = "La combinación de teclas es la misma que ya hay configurada."
				} else {
					for _, atajo_ := range contentTabAtajos.listaAtajos {
						if strings.Join(atajo_.Atajo, " + ") == strings.Join(sliceOfKeys, " + ") {
							invalidAtajo = true
							error_ = fmt.Sprintf("La combinación de teclas ya está asignada al atajo \"%s\".", atajo_.Nombre)
							break
						}
					}
				}
				if !invalidAtajo { // Atajo válido, se procede a emitir la señal para establecerlo
					atajo.Atajo = sliceOfKeys
					atajo.AtajoKeyCodes = sliceofKeyVals
					_, _ = botonOrigen.Emit("listbox-update-atajo")
					dialogCambiarAtajo.Response(gtk.RESPONSE_CLOSE)
				} else {
					labelErrorNuevoAtajo.SetText(error_)
					contenedorErrorDialogNuevoAtajo.Show()
				}
			})
		}()
	})

	botonBorrarNuevoAtajo.Connect("clicked", func(button *gtk.Button) {
		labelDialogCambiarAtajo.SetText(atajoEmpty)
		puedeEditar = true
		cantidadTeclasPressed = 0
		sliceOfKeys = nil
		sliceofKeyVals = nil

		labelErrorNuevoAtajo.SetText("")
		contenedorErrorDialogNuevoAtajo.Hide()

		button.SetSensitive(false)
		botonAceptarNuevoAtajo.SetSensitive(false)
	})

	botonCancelarNuevoAtajo.Connect("clicked", func(button *gtk.Button) {
		dialogCambiarAtajo.Response(gtk.RESPONSE_CLOSE)
	})

	dialogCambiarAtajo.Connect("key-press-event", func(dialog *gtk.Dialog, event *gdk.Event) bool {
		if puedeEditar && cantidadTeclasPressed < limiteTeclas {
			eventKey := gdk.EventKeyNewFromEvent(event)
			if keyName := glibown.KeyValName(eventKey.KeyVal()); len(keyName) > 0 {
				if !contains(sliceOfKeys, keyName) {
					sliceOfKeys = append(sliceOfKeys, keyName)
					sliceofKeyVals = append(sliceofKeyVals, eventKey.KeyVal())
					glib.IdleAdd(func() {
						labelDialogCambiarAtajo.SetText(strings.Join(sliceOfKeys, " + "))
					})
					cantidadTeclasPressed++
				}
			}
		}
		return true
	})

	dialogCambiarAtajo.Connect("key-release-event", func(dialog *gtk.Dialog, event *gdk.Event) bool {
		if puedeEditar {
			puedeEditar = false
			botonAceptarNuevoAtajo.SetSensitive(true)
			botonBorrarNuevoAtajo.SetSensitive(true)
		}
		return true
	})

	dialogCambiarAtajo.Connect("response", func(dialog *gtk.Dialog, response int) {
		_, _ = contentTabAtajos.MainGUI.application.Emit(signalControlListener, true, false)
		dialog.Destroy()
		contentTabAtajos.MainGUI.window.Present()
		botonOrigen.SetSensitive(true)
	})
	dialogCambiarAtajo.Run()
}
