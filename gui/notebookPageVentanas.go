package gui

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

import (
	"linux-windows-switcher/libs/glibown"
)

type contenTabVentanas struct {
	MainGUI
	listaClasesAbiertas     []string
	listaClasesPreferidas   []string
	listaClasesExcluidas    []string
	expansor                *gtk.Expander
	listBoxClasesAbiertas   *gtk.ListBox
	listBoxClasesPreferidas *gtk.ListBox
	listBoxClasesExcluidas  *gtk.ListBox
	listaVentanas           *listaVentanas
}

const (
	// Sección del archivo de configuración para las clases
	sectionClases = "clases"

	// Opciones dentro del archivo de configuración
	optionClasesPreferidas = "clases_preferidas"
	optionClasesExcluidas  = "clases_excluidas"
)

// newContentTabVentanas Constructor
func (mainGUI *MainGUI) newContentTabVentanas() *contenTabVentanas {
	contentTabVentanas := &contenTabVentanas{MainGUI: *mainGUI}

	contentTabVentanas.listaVentanas = contentTabVentanas.newListaVentanas()
	contentTabVentanas.setupContentTabVentanas()
	return contentTabVentanas
}

// Función de configuración
func (contentTabVentanas *contenTabVentanas) setupContentTabVentanas() {
	obj, _ := contentTabVentanas.MainGUI.builder.GetObject("listBoxClasesAbiertas")
	contentTabVentanas.listBoxClasesAbiertas = obj.(*gtk.ListBox)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("listBoxClasesPreferidas")
	contentTabVentanas.listBoxClasesPreferidas = obj.(*gtk.ListBox)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("listBoxClasesExcluidas")
	contentTabVentanas.listBoxClasesExcluidas = obj.(*gtk.ListBox)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelOrdenActual")
	labelOrdenActual := obj.(*gtk.Label)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelOrdenPorDefecto")
	labelOrdenPorDefecto := obj.(*gtk.Label)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("botonRefrescarClasesVentanasAbiertas")
	botonRefrescarClasesVentanasAbiertas := obj.(*gtk.Button)
	botonRefrescarClasesVentanasAbiertas.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() {
				contentTabVentanas.listaClasesAbiertas = nil
				contentTabVentanas.listBoxClasesAbiertas.GetChildren().Foreach(func(item interface{}) {
					contentTabVentanas.listBoxClasesAbiertas.Remove(item.(*gtk.Widget))
				})
				button.SetSensitive(false)
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				contentTabVentanas.getClasesVentanasAbiertas()
				button.SetSensitive(true)
			})
		}()
	})

	// Contenedor del TreeView de las ventanas abiertas
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("contenedorTreeViewVentanas")
	contenedorTreeViewVentanas := obj.(*gtk.Box)

	// Contenedor de la sección que tiene el Label del orden actual, botón "Refrescar Ventanas" y botón "Restaurar Orden por Defecto"
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("contenedorSeccionOrdenActual")
	contenedorSeccionOrdenActual := obj.(*gtk.Box)

	// Lista que se setea cuando se abre el expansor, indican si hubo cambios o no en la configuración
	var listaClasesPreferidasPrevio []string
	var listaClasesExcluidasPrevio []string

	// Expansor
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("expansor")
	contentTabVentanas.expansor = obj.(*gtk.Expander)
	contentTabVentanas.expansor.Connect("activate", func(expander *gtk.Expander) {
		expanded := expander.GetExpanded()

		contentTabVentanas.MainGUI.botonControlListener.SetSensitive(expanded)
		contenedorTreeViewVentanas.SetSensitive(expanded)
		contenedorSeccionOrdenActual.SetSensitive(expanded)

		if expanded {
			clasesPreferidasIguales := strings.Join(
				contentTabVentanas.listaClasesPreferidas,
				",",
			) == strings.Join(
				listaClasesPreferidasPrevio,
				",",
			)
			clasesExcluidasIguales := strings.Join(
				contentTabVentanas.listaClasesExcluidas,
				",",
			) == strings.Join(
				listaClasesExcluidasPrevio,
				",",
			)
			// Si hubo cambios en la configuración de clases, resetear el TreeView de las ventanas abiertas
			if !clasesPreferidasIguales || !clasesExcluidasIguales {
				contentTabVentanas.listaVentanas.clear()
				contentTabVentanas.getVentanasAbiertas(true, true)
			}

			// Se emite la señal para encender el listener del teclado
			_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, true, false)
		} else {
			// Se setea el contenido de la configuración cuando se abre el expansor
			listaClasesPreferidasPrevio = nil
			listaClasesPreferidasPrevio = append(listaClasesPreferidasPrevio, contentTabVentanas.listaClasesPreferidas...)

			listaClasesExcluidasPrevio = nil
			listaClasesExcluidasPrevio = append(listaClasesExcluidasPrevio, contentTabVentanas.listaClasesExcluidas...)

			// Se emite la señal para apagar el listener del teclado
			_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, false, false)
			contentTabVentanas.listaClasesAbiertas = nil
			contentTabVentanas.listBoxClasesAbiertas.GetChildren().Foreach(func(item interface{}) {
				contentTabVentanas.listBoxClasesAbiertas.Remove(item.(*gtk.Widget))
			})
			contentTabVentanas.getClasesVentanasAbiertas()
		}
	})
	contentTabVentanas.expansor.ConnectAfter("activate", func(expander *gtk.Expander) {
		width, height := contentTabVentanas.MainGUI.window.GetSize()
		if !contentTabVentanas.MainGUI.window.IsMaximized() && (width > 0 && height > 0) {
			contentTabVentanas.MainGUI.window.Resize(width, 1)
		}
	})

	// Botón refrescar ventanas
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("botonRefrescarVentanasAbiertas")
	botonRefrescarVentanasAbiertas := obj.(*gtk.Button)
	botonRefrescarVentanasAbiertas.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() {
				// Se emite la señal para apagar el listener del teclado
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, false, false)
				// Se vacía el TreeView de las ventanas
				contentTabVentanas.listaVentanas.clear()
				// Se elimina el contenido de orden actual y por defecto
				labelOrdenActual.SetText("")
				labelOrdenPorDefecto.SetText("")
				button.SetSensitive(false)
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				// Se consulta de nuevo por las clases abiertas
				contentTabVentanas.getVentanasAbiertas(false, true)
				// Se emite la señal para encender el listener del teclado
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, true, false)
				button.SetSensitive(true)
			})
		}()
	})

	// Función anónima que mira la igualdad de 2 slices de ventana
	funcTestEq := func(first []ventana, second []ventana) bool {
		if len(first) != len(second) {
			return false
		}
		for i := range first {
			if first[i].orden != second[i].orden {
				return false
			}
		}
		return true
	}

	// Botón restaurar orden
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("botonRestaurar")
	botonRestaurar := obj.(*gtk.Button)
	botonRestaurar.Connect("clicked", func(button *gtk.Button) {
		go func() {
			// Si el orden actual y el orden por defecto son iguales solo se anima el botón y se sale de la función
			if funcTestEq(ordenActual, ordenPorDefecto) {
				glib.IdleAdd(func() {
					button.SetSensitive(false)
				})
				time.Sleep(time.Second / 3)
				glib.IdleAdd(func() {
					button.SetSensitive(true)
				})
				return
			}
			glib.IdleAdd(func() {
				button.SetSensitive(false)
				// Se emite la señal para apagar el listener del teclado
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, false, false)
				// Se vacía el TreeView de las ventanas
				contentTabVentanas.listaVentanas.clear()
				// Se elimina el contenido de orden actual
				labelOrdenActual.SetText("")
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				contentTabVentanas.listaVentanas.listaVentanas = ordenPorDefecto
				// Se vuelve a añadir al TreeView las ventanas basándose en el orden por defecto
				for _, ventana := range ordenPorDefecto {
					contentTabVentanas.listaVentanas.addRow(ventana, false, nil)
				}
				// Se emite la señal para establecer el orden
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalEstablecerOrden, true, false, false)
				// Se emite la señal para encender el listener del teclado
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, true, false)
				button.SetSensitive(true)
			})
		}()
	})

	// Función anónima que consulta las clases preferidas/excluidas del archivo de configuración y las añade respectivamente
	// a la lista asociada y LisBox asociado
	getClases := func(option string, lista *[]string, box *gtk.ListBox) {
		result, _ := contentTabVentanas.MainGUI.application.Emit(signalGetConfig, sectionClases, option)
		cadenaClases := result.(string)
		if len(cadenaClases) > 0 {
			for _, clase := range strings.Split(cadenaClases, ",") {
				clase = strings.TrimSpace(clase)
				if !contains(*lista, clase) {
					*lista = append(*lista, clase)
					box.Add(contentTabVentanas.createListBoxRow(clase, false, box))
				}
			}
		}
	}

	// Se consultan las clases preferidas del archivo de configuración
	getClases(
		optionClasesPreferidas,
		&contentTabVentanas.listaClasesPreferidas,
		contentTabVentanas.listBoxClasesPreferidas,
	)
	// Se consultan las clases excluídas del archivo de configuración
	getClases(
		optionClasesExcluidas,
		&contentTabVentanas.listaClasesExcluidas,
		contentTabVentanas.listBoxClasesExcluidas,
	)

	// Señal para actualizar los textos del orden actual y por defecto
	_, _ = glibown.SignalNewV("app-update-text-atajos", glib.TYPE_NONE, 2, glib.TYPE_STRING, glib.TYPE_STRING)
	// Handler
	contentTabVentanas.MainGUI.application.Connect(
		"app-update-text-atajos",
		func(application *gtk.Application, textLabelOrdenActual string, textLabelOrdenPorDefecto string) {
			if len(textLabelOrdenActual) > 0 {
				if textLabelOrdenActual == "-1" {
					labelOrdenActual.SetText("")
				} else {
					labelOrdenActual.SetText(fmt.Sprintf("[%s]", textLabelOrdenActual))
				}
			}
			if len(textLabelOrdenPorDefecto) > 0 {
				if textLabelOrdenPorDefecto == "-1" {
					labelOrdenPorDefecto.SetText("")
				} else {
					labelOrdenPorDefecto.SetText(fmt.Sprintf("[%s]", textLabelOrdenPorDefecto))
				}
			}
		},
	)

	// Señal para encender el botón asociado al ListBoxRow del ListBox de clases abiertas
	_, _ = glib.SignalNew("listboxClasesAbiertas-enable-boton")

	// Se consulta las ventanas abiertas teniendo en cuenta que sean instancias de la configuración de clases preferidas/excluídas
	contentTabVentanas.getVentanasAbiertas(true, true)
}

// Función que obtiene un listado de clases de las ventanas abiertas y las añade al ListBox "listBoxClasesAbiertas"
func (contentTabVentanas *contenTabVentanas) getClasesVentanasAbiertas() {
	comando := fmt.Sprintf(
		"wmctrl -lx | tr -s ' ' | grep --invert-match \"%s\" | "+
			"awk '{if ($2 == \"-1\") { next; } print $3}' | sort -u",
		contentTabVentanas.MainGUI.application.GetApplicationID(),
	)
	if cadenaVentanasAbiertas, err := exec.Command("bash", "-c", comando).Output(); err == nil &&
		len(cadenaVentanasAbiertas) > 0 {
		for _, value := range strings.Split(strings.TrimSuffix(string(cadenaVentanasAbiertas), "\n"), "\n") {
			if len(value) > 0 {
				contentTabVentanas.listaClasesAbiertas = append(contentTabVentanas.listaClasesAbiertas, getClass(value))
			}
		}
		sort.Slice(contentTabVentanas.listaClasesAbiertas, func(i, j int) bool {
			return strings.ToLower(
				contentTabVentanas.listaClasesAbiertas[i],
			) < strings.ToLower(
				contentTabVentanas.listaClasesAbiertas[j],
			)
		})
		for _, clase := range contentTabVentanas.listaClasesAbiertas {
			contentTabVentanas.listBoxClasesAbiertas.Add(contentTabVentanas.createListBoxRow(clase, true, nil))
		}
	}
}

// Función que obtiene un listado de ventanas abiertas tomando en cuenta que sean instancias de las clases preferidas
// en caso de haber, sino, se toman en cuenta todas las ventanas abiertas
func (contentTabVentanas *contenTabVentanas) getVentanasAbiertas(resetOrdenActual bool, resetOrdenPorDefecto bool) {
	comando := fmt.Sprintf(
		"wmctrl -lx | tr -s ' ' | awk --non-decimal-data '%s' | grep --invert-match \"%s\"",
		awkFile,
		contentTabVentanas.MainGUI.application.GetApplicationID(),
	)
	if len(contentTabVentanas.listaClasesPreferidas) > 0 {
		comando += " | grep -E \"" + strings.Join(contentTabVentanas.listaClasesPreferidas, "|") + "\""
	}
	if len(contentTabVentanas.listaClasesExcluidas) > 0 {
		comando += " | grep --invert-match -E \"" + strings.Join(contentTabVentanas.listaClasesExcluidas, "|") + "\""
	}
	var ventanas []ventana
	if cadenaVentanasAbiertas, err := exec.Command("bash", "-c", comando).Output(); err == nil {
		if len(cadenaVentanasAbiertas) > 0 {
			for index, cadenaventana := range strings.Split(strings.TrimSuffix(string(cadenaVentanasAbiertas), "\n"), "\n") {
				datosVentana := strings.Split(cadenaventana, "|<>|")
				ventana := newVentana(datosVentana[0], getClass(datosVentana[1]), datosVentana[2], index+1)
				glib.IdleAdd(func() {
					contentTabVentanas.listaVentanas.addRow(*ventana, false, nil)
				})
				ventanas = append(ventanas, *ventana)
			}
		}
	}
	contentTabVentanas.listaVentanas.listaVentanas = ventanas

	// Si el tamaño de las ventanas del orden actual y el tamaño de las ventanas del orden por defecto son diferentes
	// entonces se reinicia el orden actual también o si el orden actual es igual a 0
	if !resetOrdenActual {
		if len(ordenActual) != len(ventanas) || len(ordenActual) == 0 {
			resetOrdenActual = true
		}
		// Si el orden actual y por defecto tienen el mismo tamaño se verifica que las ventanas en el orden actual
		// existan en el orden por defecto también, sino, se resetea el orden actual
		for _, ventanaOrdenActual := range ordenActual {
			ventanaExiste := false
			for _, ventanaNueva := range ventanas {
				if ventanaNueva.id == ventanaOrdenActual.id {
					ventanaExiste = true
					break
				}
			}
			if !ventanaExiste {
				resetOrdenActual = true
				break
			}
		}
	}
	// Se emite la señal para establecer el orden
	_, _ = contentTabVentanas.MainGUI.application.Emit(
		signalEstablecerOrden,
		resetOrdenActual,
		resetOrdenPorDefecto,
		false,
	)
}

// Función que crea un ListBoxRow listo para ser añadido a un ListBox, el parámetro booleano "add"
// indica la funcionalidad del botón que tiene el ListBoxRow
func (contentTabVentanas *contenTabVentanas) createListBoxRow(
	class string,
	add bool,
	box *gtk.ListBox,
) *gtk.ListBoxRow {
	builder := contentTabVentanas.MainGUI.getNewBuilder()

	obj, _ := builder.GetObject("listBoxRowClass")
	listboxRow := obj.(*gtk.ListBoxRow)
	listboxRow.SetName(class)

	obj, _ = builder.GetObject("labelClassListBoxRowClass")
	labelClass := obj.(*gtk.Label)
	labelClass.SetText(class)
	labelClass.SetTooltipText(class)

	var boton_ *gtk.Widget
	if add {
		obj, _ = builder.GetObject("botonAñadirListBoxRowClass")
		boton := obj.(*gtk.MenuButton)
		boton.SetTooltipText("Añadir clase a las clases preferidas o excluídas")
		boton_ = &boton.Widget

		menu, _ := gtk.MenuNew()
		menu.SetVAlign(gtk.ALIGN_CENTER)

		// Función anónima que añade un item a los ListBox de clases preferidas o excluídas
		addItemToListBox := func(section string, listaAsociada *[]string, listBoxAsociado *gtk.ListBox) {
			contentTabVentanas.addITemToListBox(class, section, listaAsociada, boton_, listBoxAsociado)
		}

		addClasesPreferidas, _ := gtk.MenuItemNewWithLabel("Añadir a las clases preferidas")
		addClasesPreferidas.Connect("activate", func(menuItem *gtk.MenuItem) {
			addItemToListBox(
				optionClasesPreferidas,
				&contentTabVentanas.listaClasesPreferidas,
				contentTabVentanas.listBoxClasesPreferidas,
			)
		})

		addClasesExcluidas, _ := gtk.MenuItemNewWithLabel("Añadir a las clases excluídas")
		addClasesExcluidas.Connect("activate", func(menuItem *gtk.MenuItem) {
			addItemToListBox(
				optionClasesExcluidas,
				&contentTabVentanas.listaClasesExcluidas,
				contentTabVentanas.listBoxClasesExcluidas,
			)
		})

		menu.Add(addClasesPreferidas)
		menu.Add(addClasesExcluidas)
		menu.ShowAll()
		boton.SetPopup(menu)

		modificaBoton := func() {
			boton.SetSensitive(
				!(contains(contentTabVentanas.listaClasesPreferidas, class) ||
					contains(contentTabVentanas.listaClasesExcluidas, class)),
			)
		}
		modificaBoton()
		listboxRow.Connect("listboxClasesAbiertas-enable-boton", func(row *gtk.ListBoxRow) {
			modificaBoton()
		})
	} else {
		obj, _ = builder.GetObject("botonEliminarListBoxRowClass")
		boton := obj.(*gtk.Button)
		boton_ = &boton.Widget

		textoTooltip := "Eliminar clase de las clases preferidas"
		listaAsociada := &contentTabVentanas.listaClasesPreferidas
		name, _ := box.GetName()
		if name == "listBoxClasesExcluidas" {
			textoTooltip = strings.Replace(textoTooltip, "preferidas", "excluídas", 1)
			listaAsociada = &contentTabVentanas.listaClasesExcluidas
		}
		boton.SetTooltipText(textoTooltip)
		boton.Connect("clicked", func(button *gtk.Button) {
			option := optionClasesPreferidas
			if name == "listBoxClasesExcluidas" {
				option = optionClasesExcluidas
			}
			contentTabVentanas.deleteItemFromListBox(class, option, listaAsociada, boton_, box, listboxRow)
		})
	}
	boton_.SetVisible(true)
	return listboxRow
}

// Función que añade un item a los ListBox de Clases Preferidas o Clases Excluídas
func (contentTabVentanas *contenTabVentanas) addITemToListBox(
	class string,
	option string,
	listaAsociada *[]string,
	botonAsociado *gtk.Widget,
	listBoxDestino *gtk.ListBox,
) {
	if !contains(*listaAsociada, class) {
		*listaAsociada = append(*listaAsociada, class)
		result, _ := contentTabVentanas.MainGUI.application.Emit(
			signalUpdateConfig,
			sectionClases,
			option,
			strings.Join(*listaAsociada, ","),
		)
		go func() {
			glib.IdleAdd(func() {
				botonAsociado.SetSensitive(false)
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				if result.(bool) {
					listBoxDestino.Add(contentTabVentanas.createListBoxRow(class, false, listBoxDestino))
				} else {
					msg2 := "Se produjo un error al guardar las clases preferidas en el archivo de configuración."
					if option == optionClasesExcluidas {
						msg2 = strings.Replace(msg2, "preferidas", "excluídas", 1)
					}
					contentTabVentanas.MainGUI.mostrarMensajeDialog(gtk.MESSAGE_ERROR, "Ocurrió un error actualizando la configuación.", msg2)
					*listaAsociada = removeItem(*listaAsociada, class)
					botonAsociado.SetSensitive(true)
				}
			})
		}()
	}
}

// Función que elimina un item de los ListBox de Clases Preferidas o Clases Excluídas
func (contentTabVentanas *contenTabVentanas) deleteItemFromListBox(
	class string,
	option string,
	listaAsociada *[]string,
	botonAsociado *gtk.Widget,
	listBoxAsociado *gtk.ListBox,
	listBoxRowAsociado *gtk.ListBoxRow,
) {
	*listaAsociada = removeItem(*listaAsociada, class)
	result, _ := contentTabVentanas.MainGUI.application.Emit(
		signalUpdateConfig,
		sectionClases,
		option,
		strings.Join(*listaAsociada, ","),
	)
	go func() {
		glib.IdleAdd(func() {
			botonAsociado.SetSensitive(false)
		})
		time.Sleep(time.Second / 3)
		glib.IdleAdd(func() {
			if result.(bool) {
				var itemAsociado *gtk.Widget
				if contains(contentTabVentanas.listaClasesAbiertas, class) {
					for list := contentTabVentanas.listBoxClasesAbiertas.GetChildren(); list != nil; list = list.Next() {
						item := list.Data().(*gtk.Widget)
						if name, err := item.GetName(); err == nil {
							if name == class {
								itemAsociado = item
								break
							}
						}
					}
				}
				listBoxAsociado.Remove(
					listBoxRowAsociado,
				) // Se elimina el Row del ListBox asociado (de clases preferidas o excluídas)
				if itemAsociado != nil {
					_, _ = itemAsociado.Emit(
						"listboxClasesAbiertas-enable-boton",
					) // Si el ítem eliminado existe en el ListBox de clases abiertas se procede a habilitar su botón
				}
			} else {
				msg2 := "Se produjo un error al actualizar las clases preferidas en el archivo de configuración."
				if option == optionClasesExcluidas {
					msg2 = strings.Replace(msg2, "preferidas", "excluídas", 1)
				}
				contentTabVentanas.MainGUI.mostrarMensajeDialog(gtk.MESSAGE_ERROR, "Ocurrió un error actualizando la configuación.", msg2)
				if !contains(*listaAsociada, class) {
					*listaAsociada = append(*listaAsociada, class)
				}
				botonAsociado.SetSensitive(true)
			}
		})
	}()
}

// Función que remueve un item de un slice
func removeItem(list []string, item string) []string {
	for index, value := range list {
		if value == item {
			return append(list[:index], list[index+1:]...)
		}
	}
	return list
}

// Función que indica si un item de un slice contiene una cadena o si la cadena contiene el item
func contains(list []string, string string) bool {
	response := false
	for _, item := range list {
		if strings.Contains(item, string) || strings.Contains(string, item) {
			response = true
			break
		}
	}
	return response
}

// Función que recibe una clase y retorna ella misma pero más simplificada
func getClass(classString string) string {
	value := ""
	isRepeated := false
	for _, v := range strings.Split(classString, ".") {
		if strings.EqualFold(v, value) {
			if unicode.IsUpper(rune(v[0])) {
				value = v
			}
			isRepeated = true
			break
		} else {
			value = v
		}
	}
	if !isRepeated {
		value = classString
	}
	return value
}
