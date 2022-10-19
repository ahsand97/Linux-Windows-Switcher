package gui

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gdk"

	"github.com/gotk3/gotk3/glib"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)

type listaVentanas struct {
	contenTabVentanas
	listStoreventanasAbiertas     *gtk.ListStore
	treeViewVentanasAbiertas      *gtk.TreeView
	treeSelectionVentanasAbiertas *gtk.TreeSelection
	signalHandlerRowDeleted       glib.SignalHandle
	signalHandlerRowInserted      glib.SignalHandle
	listaVentanas                 []ventana // Lista de ventanas auxiliar
}

const (
	// Columnas del modelo
	columnaOrden = iota
	columnaId
	columnaClase
	columnaTitulo
	columnaExcluir
	columnaRepetida
	columnaSeparacionVertical
	columnaEllipsize
	columnaFontWeight
	columnaVentanaEliminada

	// Valores por defecto de las columnas del modelo
	valueColumnaSeparacionVertical = 6
	valueColumnaPangoEllipsizeMode = pango.ELLIPSIZE_END
	valueColumnaFontWeight         = pango.WEIGHT_BOLD
	prefijoVentanaRepetida         = "<b><i>(repetida)</i></b> "
)

// newListaVentanas Constructor
func (contentTabVentanas *contenTabVentanas) newListaVentanas() *listaVentanas {
	listaVentanas := &listaVentanas{contenTabVentanas: *contentTabVentanas}

	listaVentanas.setupLista()
	return listaVentanas
}

// Función de configuración
func (listaVentanas *listaVentanas) setupLista() {
	obj, _ := listaVentanas.MainGUI.builder.GetObject("modelVentanasAbiertas")
	listaVentanas.listStoreventanasAbiertas = obj.(*gtk.ListStore)

	devolverItemPosInicial := false
	obj, _ = listaVentanas.MainGUI.builder.GetObject("treeSelectionVentanasAbiertas")
	listaVentanas.treeSelectionVentanasAbiertas = obj.(*gtk.TreeSelection)
	signalChangedSelection := listaVentanas.treeSelectionVentanasAbiertas.Connect(
		"changed",
		func(selection *gtk.TreeSelection) {
			if model, iter, ok := selection.GetSelected(); ok {
				value, _ := model.ToTreeModel().GetValue(iter, columnaExcluir)
				goValue, _ := value.GoValue()
				excluida := goValue.(bool)

				value, _ = model.ToTreeModel().GetValue(iter, columnaVentanaEliminada)
				goValue, _ = value.GoValue()
				eliminada := goValue.(bool)

				listaVentanas.treeViewVentanasAbiertas.SetReorderable(!(excluida || eliminada))
				devolverItemPosInicial = excluida || eliminada
			}
		},
	)

	pathItemRowInserted := ""
	listaVentanas.signalHandlerRowDeleted = listaVentanas.listStoreventanasAbiertas.Connect(
		"row-deleted",
		func(store *gtk.ListStore, path *gtk.TreePath) {
			if len(pathItemRowInserted) > 0 && devolverItemPosInicial {
				iter, err := store.GetIterFromString(pathItemRowInserted)
				iter_, err_ := store.GetIter(path)
				if err == nil && err_ == nil {
					listaVentanas.treeSelectionVentanasAbiertas.SelectIter(iter)
					glib.IdleAdd(func() {
						time.Sleep(time.Second / 2)
						store.MoveBefore(iter, iter_)
					})
				}
			}
			// Se emite la señal para establecer el orden
			_, _ = listaVentanas.MainGUI.application.Emit(signalEstablecerOrden, true, false, true)
		},
	)
	listaVentanas.signalHandlerRowInserted = listaVentanas.listStoreventanasAbiertas.Connect(
		"row-inserted",
		func(store *gtk.ListStore, path *gtk.TreePath, iter *gtk.TreeIter) {
			devolverItemPosInicial = false
			var iterPrimerExcluido *gtk.TreeIter
			cantidadItems := 0
			store.ForEach(func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
				value, _ := model.GetValue(iter, columnaExcluir)
				if valBoolean, err := value.GoValue(); err == nil {
					if valBoolean.(bool) && iterPrimerExcluido == nil {
						iterPrimerExcluido = iter
					}
				}
				value, _ = model.GetValue(iter, columnaOrden)
				if valInt, err := value.GoValue(); err == nil {
					if valInt.(int) != 0 {
						cantidadItems++
					}
				}
				return false
			})
			pathItemRowInserted = path.String()
			pathNumerico, _ := strconv.Atoi(pathItemRowInserted)
			if pathNumerico == 0 {
				pathNumerico++
			} else if pathNumerico == cantidadItems {
				pathNumerico--
			}
			pathItemRowInserted = strconv.Itoa(pathNumerico)
			if iterPrimerExcluido != nil {
				// Si se trata de insertar un item después de los items excluidos, se marca el booleano como verdadero
				if pathPrimerItemExcluido, err := store.GetPath(iterPrimerExcluido); err == nil {
					devolverItemPosInicial = path.String() > pathPrimerItemExcluido.String()
				}
			} else {
				// No hay items excluidos, por ende se deja al item insertado en la posición deseada
				devolverItemPosInicial = false
			}
		},
	)

	obj, _ = listaVentanas.MainGUI.builder.GetObject("treeViewVentanasAbiertas")
	listaVentanas.treeViewVentanasAbiertas = obj.(*gtk.TreeView)
	_ = listaVentanas.treeViewVentanasAbiertas.SetProperty("has-tooltip", true)
	listaVentanas.treeViewVentanasAbiertas.Connect(
		"query-tooltip",
		func(view *gtk.TreeView, x int, y int, keyboardMode bool, tooltip *gtk.Tooltip) bool {
			if !keyboardMode {
				bx := new(int)
				by := new(int)
				view.ConvertWidgetToBinWindowCoords(x, y, bx, by)
				if path, column, _, _, exists := view.GetPathAtPos(*bx, *by); exists {
					if column.GetTitle() == "Clase" || column.GetTitle() == "Título" {
						iter, _ := listaVentanas.listStoreventanasAbiertas.GetIter(path)

						// Cell Renderer Asociado
						obj, _ = listaVentanas.MainGUI.builder.GetObject(column.GetTitle())
						cellRenderer := obj.(*gtk.CellRendererText)

						textoTooltip := ""
						if column.GetTitle() == "Clase" {
							value, _ := listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaClase)
							clase, _ := value.GoValue()
							textoTooltip = strings.TrimPrefix(clase.(string), prefijoVentanaRepetida)
						} else if column.GetTitle() == "Título" {
							value, _ := listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaTitulo)
							titulo, _ := value.GoValue()
							textoTooltip = titulo.(string)
						}
						if len(textoTooltip) > 0 && cellRenderer != nil {
							tooltip.SetText(textoTooltip)
							view.SetTooltipCell(
								tooltip,
								path,
								column,
								&cellRenderer.CellRenderer,
							)
							return true
						}
					}
				}
			}
			return false
		},
	)

	idFilaSelected := 0
	listaVentanas.treeViewVentanasAbiertas.Connect("drag-begin", func(view *gtk.TreeView, ctx *gdk.DragContext) {
		// Se emite la señal para apagar el listener del teclado
		_, _ = listaVentanas.MainGUI.application.Emit(signalControlListener, false, false)
		if model, iter, ok := listaVentanas.treeSelectionVentanasAbiertas.GetSelected(); ok {
			val, _ := model.ToTreeModel().GetValue(iter, columnaOrden)
			if valInt, err := val.GoValue(); err == nil {
				idFilaSelected = valInt.(int)
			}
		}
		listaVentanas.treeSelectionVentanasAbiertas.HandlerBlock(signalChangedSelection)
	})
	listaVentanas.treeViewVentanasAbiertas.Connect("drag-end", func(view *gtk.TreeView, ctx *gdk.DragContext) {
		var iterFilaASeleccionar *gtk.TreeIter
		listaVentanas.listStoreventanasAbiertas.ForEach(
			func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
				value, _ := model.GetValue(iter, columnaOrden)
				if valInt, err := value.GoValue(); err == nil {
					if valInt.(int) == idFilaSelected {
						iterFilaASeleccionar = iter
						return true
					}
				}
				return false
			},
		)
		listaVentanas.treeSelectionVentanasAbiertas.HandlerUnblock(signalChangedSelection)
		if iterFilaASeleccionar != nil {
			if path, err := listaVentanas.listStoreventanasAbiertas.GetPath(iterFilaASeleccionar); err == nil {
				view.SetCursor(path, nil, false)
				listaVentanas.treeSelectionVentanasAbiertas.SelectIter(iterFilaASeleccionar)
			}
		}
		devolverItemPosInicial = false
		// Se emite la señal para encender el listener del teclado
		_, _ = listaVentanas.MainGUI.application.Emit(signalControlListener, true, false)
	})

	excludedWindow := false
	obj, _ = listaVentanas.MainGUI.builder.GetObject("Excluir")
	cellRendererToggleColumnaExcluir := obj.(*gtk.CellRendererToggle)
	cellRendererToggleColumnaExcluir.Connect(
		"toggled",
		func(toggle *gtk.CellRendererToggle, path string) {
			// Se hace el treeview no ordenable y se bloquea la señal de cambio de selección
			listaVentanas.treeViewVentanasAbiertas.SetReorderable(false)
			listaVentanas.treeSelectionVentanasAbiertas.HandlerBlock(signalChangedSelection)

			// *gtk.Iter asociado a la fila donde se desea hacer la exlusión
			iter, _ := listaVentanas.listStoreventanasAbiertas.GetIterFromString(path)

			// Valor actual de la fila donde se desea hacer toggle en la columna excluír
			value, _ := listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaExcluir)
			goValue, _ := value.GoValue()
			excluida := goValue.(bool)

			// Valor actual de la fila donde se desea hacer toggle en la columna oculta eliminada
			value, _ = listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaVentanaEliminada)
			goValue, _ = value.GoValue()
			eliminada := goValue.(bool)

			// Indica si se debe finalizar la ejecución, es decir no hacer toggle
			finalize := false
			if excluida && eliminada {
				finalize = true // Si la fila está excluida y eliminada, no se hace nada
			}
			// Se hace toggle si finalize = false
			if !finalize {
				newVal := !excluida
				_ = listaVentanas.listStoreventanasAbiertas.SetValue(iter, columnaExcluir, newVal)
				glib.IdleAdd(func() {
					actualizaOrdenPorDefecto := false
					if excludedWindow {
						_ = listaVentanas.listStoreventanasAbiertas.SetValue(iter, columnaVentanaEliminada, true)
						actualizaOrdenPorDefecto = true
						excludedWindow = false
					} else {
						time.Sleep(time.Second / 2) // Si se hace toggle por UI se añade éste timeout, si es por señal no
					}
					if newVal {
						listaVentanas.listStoreventanasAbiertas.MoveBefore(iter, nil) // Se mueve la fila al final
					} else { // Se mueve el item a donde sea que esté el primer excluído (si es que hay) - 1, sino, se conserva su posición
						var iterPrimerExcluido *gtk.TreeIter
						listaVentanas.listStoreventanasAbiertas.ForEach(func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
							// Valor columna excluír
							value, _ = model.GetValue(iter, columnaExcluir)
							goValue, _ = value.GoValue()
							excluida = goValue.(bool)

							if excluida {
								iterPrimerExcluido = iter
								return true
							}
							return false
						})
						if iterPrimerExcluido != nil {
							listaVentanas.listStoreventanasAbiertas.MoveBefore(iter, iterPrimerExcluido)
						}
					}
					// Se emite la señal para establecer el orden, ya que hubo una exclusión de ventana
					_, _ = listaVentanas.MainGUI.application.Emit(
						signalEstablecerOrden,
						true,
						actualizaOrdenPorDefecto,
						true,
					)
				})
			}

			// Se hace el treeview ordenable y se desbloquea la señal de cambio de selección
			listaVentanas.treeSelectionVentanasAbiertas.HandlerUnblock(signalChangedSelection)
			listaVentanas.treeViewVentanasAbiertas.SetReorderable(true)
		},
	)

	listaVentanas.treeViewVentanasAbiertas.Connect("key-press-event", func(view *gtk.TreeView, event *gdk.Event) bool {
		eventKey := gdk.EventKeyNewFromEvent(event)
		if eventKey.KeyVal() == gdk.KEY_Return {
			if model, iter, ok := listaVentanas.treeSelectionVentanasAbiertas.GetSelected(); ok {
				path, _ := model.ToTreeModel().GetPath(iter)
				_, _ = cellRendererToggleColumnaExcluir.Emit("toggled", path.String())
				return true
			}
		}
		return false
	})

	quitarSeleccion := false
	mostrarMenuContextual := false

	// Función que quita la selección visible del treeview
	funcQuitarSeleccion := func() {
		cantidadItems := 0
		listaVentanas.listStoreventanasAbiertas.ForEach(
			func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
				val, _ := model.ToTreeModel().GetValue(iter, columnaOrden)
				if valInt, err := val.GoValue(); err == nil {
					if valInt.(int) != 0 {
						cantidadItems++
					}
				}
				return false
			},
		)
		listaVentanas.treeViewVentanasAbiertas.GrabFocus()
		if cantidadItems > 0 {
			path, _ := gtk.TreePathNewFromString(strconv.Itoa(cantidadItems + 1))
			listaVentanas.treeViewVentanasAbiertas.SetCursor(path, nil, false)
			listaVentanas.treeSelectionVentanasAbiertas.UnselectAll()
		}
	}

	listaVentanas.treeViewVentanasAbiertas.Connect(
		"button-press-event",
		func(view *gtk.TreeView, event *gdk.Event) bool {
			eventButton := gdk.EventButtonNewFromEvent(event)
			if _, _, _, _, exists := view.GetPathAtPos(int(math.Round(eventButton.X())), int(math.Round(eventButton.Y()))); exists {
				if eventButton.Button() == gdk.BUTTON_SECONDARY {
					mostrarMenuContextual = true
				}
			} else {
				quitarSeleccion = true
				return true
			}
			return false
		},
	)

	listaVentanas.treeViewVentanasAbiertas.Connect(
		"button-release-event",
		func(view *gtk.TreeView, event *gdk.Event) bool {
			if quitarSeleccion {
				funcQuitarSeleccion()
				quitarSeleccion = false
			} else if mostrarMenuContextual {
				desactivarContextMenu := true
				if eventButton := gdk.EventButtonNewFromEvent(event); eventButton.Button() == gdk.BUTTON_SECONDARY {
					if path, _, _, _, exists := view.GetPathAtPos(int(math.Round(eventButton.X())), int(math.Round(eventButton.Y()))); exists {
						iter, _ := listaVentanas.listStoreventanasAbiertas.GetIter(path)

						value, _ := listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaExcluir)
						goValue, _ := value.GoValue()
						excluida := goValue.(bool)

						value, _ = listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaRepetida)
						goValue, _ = value.GoValue()
						repetida := goValue.(bool)

						value, _ = listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaVentanaEliminada)
						goValue, _ = value.GoValue()
						eliminada := goValue.(bool)

						if !excluida || (excluida && repetida || excluida && eliminada) {
							listaVentanas.createContextMenuTreeview(iter, !(repetida || eliminada), eventButton.Event)
							desactivarContextMenu = false
						}
					}
				}
				if desactivarContextMenu {
					mostrarMenuContextual = false
				}
			}
			return false
		},
	)
	listaVentanas.treeViewVentanasAbiertas.Connect("focus-in-event", func(view *gtk.TreeView, event *gdk.Event) bool {
		if path, _ := view.GetCursor(); path != nil && !quitarSeleccion {
			view.SetCursor(path, nil, false)
			listaVentanas.treeSelectionVentanasAbiertas.SelectPath(path)
		}
		return false
	})
	listaVentanas.treeViewVentanasAbiertas.Connect("focus-out-event", func(view *gtk.TreeView, event *gdk.Event) bool {
		if !mostrarMenuContextual {
			listaVentanas.treeSelectionVentanasAbiertas.UnselectAll()
		}
		return false
	})

	// Señal para desactivar el boolean mostrarMenuContextual
	_, _ = glib.SignalNew("desactivar-menu-contextual")
	listaVentanas.treeViewVentanasAbiertas.Connect("desactivar-menu-contextual", func(view *gtk.TreeView) {
		mostrarMenuContextual = false
	})

	// Se aplcica la fuente negrilla a la cabecera del treeview
	cssProvider, _ := gtk.CssProviderNew()
	_ = cssProvider.LoadFromData("#treeViewVentanasAbiertas.view header button { font-weight: bold; }")
	screen, _ := gdk.ScreenGetDefault()
	gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	// Se conecta la señal para establecer el orden basado en el treeview
	listaVentanas.MainGUI.application.Connect(
		signalEstablecerOrden,
		func(application *gtk.Application, resetOrdenActual bool, resetOrdenPorDefecto bool, useGUI bool) {
			var ventanasOrdenActual []ventana
			var ventanasOrdenPorDefecto []ventana
			if useGUI {
				listaVentanas.listStoreventanasAbiertas.ForEach(
					func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
						value, _ := model.GetValue(iter, columnaExcluir)
						goValue, _ := value.GoValue()
						excluded := goValue.(bool)

						value, _ = model.GetValue(iter, columnaVentanaEliminada)
						goValue, _ = value.GoValue()
						eliminada := goValue.(bool)

						value, _ = model.GetValue(iter, columnaRepetida)
						goValue, _ = value.GoValue()
						repetida := goValue.(bool)

						if !eliminada {
							ventana := listaVentanas.getVentanaFromRow(iter)
							if !repetida {
								ventanasOrdenPorDefecto = append(ventanasOrdenPorDefecto, *ventana)
							}
							if !excluded {
								ventanasOrdenActual = append(ventanasOrdenActual, *ventana)
							}
						}
						return false
					},
				)
			} else {
				ventanasOrdenActual = listaVentanas.listaVentanas
				ventanasOrdenPorDefecto = listaVentanas.listaVentanas
			}
			var ordenActualText []string
			var ordenPorDefectoText []string
			if resetOrdenActual {
				ordenActual = ventanasOrdenActual
			}
			for _, ventana := range ordenActual {
				ordenActualText = append(ordenActualText, strconv.Itoa(ventana.orden))
			}
			if resetOrdenPorDefecto {
				ordenPorDefecto = ventanasOrdenPorDefecto
			}
			for _, ventana := range ordenPorDefecto {
				ordenPorDefectoText = append(ordenPorDefectoText, strconv.Itoa(ventana.orden))
			}
			textoASetearOrdenActual := strings.Join(ordenActualText, ", ")
			textoASetearOrdenPorDefecto := strings.Join(ordenPorDefectoText, ", ")
			if len(ordenActualText) == 0 {
				textoASetearOrdenActual = "-1"
			}
			if len(ordenPorDefectoText) == 0 {
				textoASetearOrdenPorDefecto = "-1"
			}
			_, _ = listaVentanas.MainGUI.application.Emit(
				"app-update-text-atajos",
				textoASetearOrdenActual,
				textoASetearOrdenPorDefecto,
			)
		},
	)

	// Se conecta la señal para borrar una ventana que ya no es válida del treeview
	listaVentanas.MainGUI.application.Connect(signalDeleteRow, func(application *gtk.Application, path string) {
		// Indica si se debe excluír una ventana porque no es seleccionable
		excludedWindow = true
		_, _ = cellRendererToggleColumnaExcluir.Emit("toggled", path)
		funcQuitarSeleccion()
	})
}

// Función que elimina todos los elementos del TreeView de las ventanas abiertas
func (listaVentanas *listaVentanas) clear() {
	listaVentanas.listStoreventanasAbiertas.HandlerBlock(listaVentanas.signalHandlerRowDeleted)
	listaVentanas.listStoreventanasAbiertas.Clear()
	listaVentanas.listStoreventanasAbiertas.HandlerUnblock(listaVentanas.signalHandlerRowDeleted)
}

// Función que invoca un menú contextual en el TreeView de las ventanas abiertas
func (listaVentanas *listaVentanas) createContextMenuTreeview(iter *gtk.TreeIter, repetir bool, event *gdk.Event) {
	menu, _ := gtk.MenuNew()

	menu.Connect("deactivate", func(menu *gtk.Menu) {
		_, _ = listaVentanas.treeViewVentanasAbiertas.Emit("desactivar-menu-contextual")
	})

	repetirItem, _ := gtk.MenuItemNewWithLabel("Repetir")
	repetirItem.Connect("activate", func(menuItem *gtk.MenuItem) {
		listaVentanas.clonarRow(iter)
	})

	eliminarItem, _ := gtk.MenuItemNewWithLabel("Eliminar")
	eliminarItem.Connect("activate", func(item *gtk.MenuItem) {
		listaVentanas.borrarRow(iter)
	})

	if repetir {
		menu.Add(repetirItem)
	} else {
		menu.Add(eliminarItem)
	}
	menu.ShowAll()
	menu.PopupAtPointer(event)
}

// Función que clona una fila del TreeView de las ventanas abiertas y la ubica justo debajo
// de la original
func (listaVentanas *listaVentanas) clonarRow(iter *gtk.TreeIter) {
	ventana := listaVentanas.getVentanaFromRow(iter)
	ventana.clase = prefijoVentanaRepetida + ventana.clase
	newOrden := 0
	listaVentanas.listStoreventanasAbiertas.ForEach(
		func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
			value, _ := model.GetValue(iter, columnaOrden)
			goValue, _ := value.GoValue()
			orden := goValue.(int)

			if orden != 0 {
				newOrden++
			}
			return false
		},
	)
	if newOrden != 0 {
		ventana.orden = newOrden + 1
	}
	listaVentanas.addRow(*ventana, true, iter)
	// Se emite la señal para establecer el orden
	_, _ = listaVentanas.MainGUI.application.Emit(signalEstablecerOrden, true, false, true)
}

// Función que elimina una fila del TreeView de las ventanas abiertas
func (listaVentanas *listaVentanas) borrarRow(iter *gtk.TreeIter) {
	ventana := listaVentanas.getVentanaFromRow(iter)
	listaVentanas.listStoreventanasAbiertas.ForEach(
		func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
			value, _ := model.GetValue(iter, columnaOrden)
			goValue, _ := value.GoValue()
			orden := goValue.(int)

			if orden > ventana.orden {
				_ = listaVentanas.listStoreventanasAbiertas.SetValue(iter, columnaOrden, orden-1)
			}
			return false
		},
	)
	listaVentanas.listStoreventanasAbiertas.Remove(iter)
	// Se emite la señal para establecer el orden
	_, _ = listaVentanas.MainGUI.application.Emit(signalEstablecerOrden, true, false, true)
}

/*
Función que añade una fila al Treeview de las ventanas abiertas en la posición indicada por el atributo "orden"
del parámetro "ventana". Si el booleano "ventanaRepetida" es true y el atributo "iter" no es nulo
entonces la fila nueva se ubica justo debajo de "iter"
*/
func (listaVentanas *listaVentanas) addRow(ventana ventana, ventanaRepetida bool, iter *gtk.TreeIter) {
	listaVentanas.listStoreventanasAbiertas.HandlerBlock(listaVentanas.signalHandlerRowInserted)
	var iterNuevo *gtk.TreeIter
	if !ventanaRepetida {
		iterNuevo = listaVentanas.listStoreventanasAbiertas.Insert(ventana.orden - 1)
	} else {
		iterNuevo = listaVentanas.listStoreventanasAbiertas.InsertAfter(iter)
	}
	_ = listaVentanas.listStoreventanasAbiertas.Set(
		iterNuevo,
		[]int{
			columnaOrden,
			columnaId,
			columnaClase,
			columnaTitulo,
			columnaExcluir,
			columnaRepetida,
			columnaSeparacionVertical,
			columnaEllipsize,
			columnaFontWeight,
			columnaVentanaEliminada,
		},
		[]interface{}{
			ventana.orden,
			ventana.id,
			ventana.clase,
			ventana.titulo,
			false,
			ventanaRepetida,
			valueColumnaSeparacionVertical,
			valueColumnaPangoEllipsizeMode,
			valueColumnaFontWeight,
			false,
		},
	)
	listaVentanas.listStoreventanasAbiertas.HandlerUnblock(listaVentanas.signalHandlerRowInserted)
}

// Devuelve un objeto ventana a partir de una fila del TreeView de ventanas abiertas
func (listaVentanas *listaVentanas) getVentanaFromRow(iter *gtk.TreeIter) *ventana {
	value, _ := listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaOrden)
	goValue, _ := value.GoValue()
	orden := goValue.(int)

	value, _ = listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaId)
	goValue, _ = value.GoValue()
	id := goValue.(string)

	value, _ = listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaClase)
	goValue, _ = value.GoValue()
	clase := goValue.(string)

	value, _ = listaVentanas.listStoreventanasAbiertas.GetValue(iter, columnaTitulo)
	goValue, _ = value.GoValue()
	titulo := goValue.(string)

	return newVentana(id, clase, titulo, orden)
}
