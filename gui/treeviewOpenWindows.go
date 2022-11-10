package gui

import (
	"fmt"
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
	contentTabVentanas
	listStoreActiveWindows     *gtk.ListStore
	treeViewActiveWindows      *gtk.TreeView
	treeSelectionActiveWindows *gtk.TreeSelection
	signalHandlerRowDeleted    glib.SignalHandle
	signalHandlerRowInserted   glib.SignalHandle
	windowList                 []window
}

const (
	// Columns associated to the *gtk.ListModel
	columnOrder = iota
	columnId
	columnClass
	columnTitle
	columnExcluded
	columnCloned
	columnPadding
	columnEllipsize
	columnFontWeight
	columnDeletedWindow
	columnIcon

	// Default values to columns from model
	valuecolumnPadding            = 6
	valueColumnPangoEllipsizeMode = pango.ELLIPSIZE_END
	valuecolumnFontWeight         = pango.WEIGHT_BOLD
)

// Prefix used in class of cloned windows
var (
	prefixClonedWindow string
	prefixClosedWindow string
)

// newListaVentanas Constructor
func (contentTabVentanas *contentTabVentanas) newListaVentanas() *listaVentanas {
	listaVentanas := &listaVentanas{contentTabVentanas: *contentTabVentanas}

	listaVentanas.initLocale()
	listaVentanas.setupLista()
	return listaVentanas
}

// Config function to assign all UI strings to appropriate locale
func (listaVentanas *listaVentanas) initLocale() {
	// Get default prefix of cloned and deleted windows
	prefixClonedWindow = fmt.Sprintf("<b><i>(%s)</i></b> ", funcGetStringResource("gui_treeview_prefix_cloned_window"))
	prefixClosedWindow = fmt.Sprintf(
		"<b><i>(%s)</i></b> ",
		funcGetStringResource("gui_treeview_prefix_closed_window"),
	)

	obj, _ := listaVentanas.MainGUI.builder.GetObject("columnClass")
	columnClass_ := obj.(*gtk.TreeViewColumn)
	columnClass_.SetTitle(funcGetStringResource("gui_treeview_column_class"))

	obj, _ = listaVentanas.MainGUI.builder.GetObject("columnTitle")
	columnTitle_ := obj.(*gtk.TreeViewColumn)
	columnTitle_.SetTitle(funcGetStringResource("gui_treeview_column_title"))

	obj, _ = listaVentanas.MainGUI.builder.GetObject("columnExclude")
	columnExclude_ := obj.(*gtk.TreeViewColumn)
	columnExclude_.SetTitle(funcGetStringResource("gui_treeview_column_exclude"))
}

// Config function
func (listaVentanas *listaVentanas) setupLista() {
	obj, _ := listaVentanas.MainGUI.builder.GetObject("modelActiveWindows")
	listaVentanas.listStoreActiveWindows = obj.(*gtk.ListStore)

	returnItemToInitialPos := false // wether an item should be returned to its initial pos
	obj, _ = listaVentanas.MainGUI.builder.GetObject("treeSelectionActiveWindows")
	listaVentanas.treeSelectionActiveWindows = obj.(*gtk.TreeSelection)
	// Handler of signal "changed", triggered when the visible selection of the *gtk.TreeView changes
	signalChangedSelection := listaVentanas.treeSelectionActiveWindows.Connect(
		"changed",
		func(selection *gtk.TreeSelection) {
			if model, iter, ok := selection.GetSelected(); ok {
				value, _ := model.ToTreeModel().GetValue(iter, columnExcluded)
				goValue, _ := value.GoValue()
				excluded := goValue.(bool)

				value, _ = model.ToTreeModel().GetValue(iter, columnDeletedWindow)
				goValue, _ = value.GoValue()
				deleted := goValue.(bool)

				listaVentanas.treeViewActiveWindows.SetReorderable(!(excluded || deleted))
				returnItemToInitialPos = excluded || deleted
			}
		},
	)

	var pathItemRowInserted string // Path where new item was inserted
	// Handler of signal "row-deleted". This signal is emitted when a row has been deleted (drag-n-drop)
	listaVentanas.signalHandlerRowDeleted = listaVentanas.listStoreActiveWindows.Connect(
		"row-deleted",
		func(store *gtk.ListStore, path *gtk.TreePath) {
			if len(pathItemRowInserted) > 0 && returnItemToInitialPos {
				iter, err := store.GetIterFromString(pathItemRowInserted)
				iter_, err_ := store.GetIter(path)
				if err == nil && err_ == nil {
					listaVentanas.treeSelectionActiveWindows.SelectIter(iter)
					glib.IdleAdd(func() {
						time.Sleep(time.Second / 2)
						store.MoveBefore(iter, iter_)
					})
				}
			}
			// Emit signal to stablish order
			_, _ = listaVentanas.MainGUI.application.Emit(signalSetOrder, true, false, true)
		},
	)
	// Handler of signal "row-inserted". This signal is emitted when a row has been inserted (drag-n-drop)
	listaVentanas.signalHandlerRowInserted = listaVentanas.listStoreActiveWindows.Connect(
		"row-inserted",
		func(store *gtk.ListStore, path *gtk.TreePath, iter *gtk.TreeIter) {
			returnItemToInitialPos = false
			var iterFirstexcludeWindow *gtk.TreeIter
			amountOfItems := 0
			store.ForEach(func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
				value, _ := model.GetValue(iter, columnExcluded)
				if valBoolean, err := value.GoValue(); err == nil {
					if valBoolean.(bool) && iterFirstexcludeWindow == nil {
						iterFirstexcludeWindow = iter
					}
				}
				value, _ = model.GetValue(iter, columnOrder)
				if valInt, err := value.GoValue(); err == nil {
					if valInt.(int) != 0 {
						amountOfItems++
					}
				}
				return false // loop through all items
			})
			pathItemRowInserted = path.String()
			numericPath, _ := strconv.Atoi(pathItemRowInserted)
			if numericPath == 0 {
				numericPath++
			} else if numericPath == amountOfItems {
				numericPath--
			}
			pathItemRowInserted = strconv.Itoa(numericPath)
			if iterFirstexcludeWindow != nil {
				// If an item is moved below excluded items then the boolean "returnItemToInitialPos" is set to true
				if pathPrimerItemExcluido, err := store.GetPath(iterFirstexcludeWindow); err == nil {
					returnItemToInitialPos = path.String() > pathPrimerItemExcluido.String()
				}
			} else {
				// There's no excluded items, new item can be on desired position
				returnItemToInitialPos = false
			}
		},
	)

	obj, _ = listaVentanas.MainGUI.builder.GetObject("treeViewActiveWindows")
	listaVentanas.treeViewActiveWindows = obj.(*gtk.TreeView)
	_ = listaVentanas.treeViewActiveWindows.SetProperty("has-tooltip", true)
	/*
		Handler of signal "query-tooltip"
		Emitted when GtkWidget:has-tooltip is TRUE and the hover timeout has expired with the cursor hovering “above” widget;
		or emitted when widget got focus in keyboard mode.
	*/
	listaVentanas.treeViewActiveWindows.Connect(
		"query-tooltip",
		func(view *gtk.TreeView, x int, y int, keyboardMode bool, tooltip *gtk.Tooltip) bool {
			if !keyboardMode {
				bx := new(int)
				by := new(int)
				view.ConvertWidgetToBinWindowCoords(x, y, bx, by)
				if path, column, _, _, exists := view.GetPathAtPos(*bx, *by); exists {
					if column.GetTitle() == funcGetStringResource("gui_treeview_column_class") ||
						column.GetTitle() == funcGetStringResource("gui_treeview_column_title") {
						iter, _ := listaVentanas.listStoreActiveWindows.GetIter(path)

						// *gtk.CellRendererText where to show the tooltip
						var obj glib.IObject
						if column.GetTitle() == funcGetStringResource("gui_treeview_column_class") {
							obj, _ = listaVentanas.MainGUI.builder.GetObject("Class")
						} else if column.GetTitle() == funcGetStringResource("gui_treeview_column_title") {
							obj, _ = listaVentanas.MainGUI.builder.GetObject("Title")
						}
						cellRenderer := obj.(*gtk.CellRendererText)

						toolTipText := ""
						if column.GetTitle() == funcGetStringResource("gui_treeview_column_class") {
							value, _ := listaVentanas.listStoreActiveWindows.GetValue(iter, columnClass)
							class, _ := value.GoValue()
							toolTipText = strings.TrimPrefix(class.(string), prefixClonedWindow)
						} else if column.GetTitle() == funcGetStringResource("gui_treeview_column_title") {
							value, _ := listaVentanas.listStoreActiveWindows.GetValue(iter, columnTitle)
							title, _ := value.GoValue()
							toolTipText = title.(string)
						}
						if len(toolTipText) > 0 && cellRenderer != nil {
							tooltip.SetMarkup(toolTipText)
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

	var idRowSelected int // id of selected row when dragging
	var originalListenerState bool
	// Handler of signal "drag-begin". This signal is emitted when drag starts on the *gtk.TreeView
	listaVentanas.treeViewActiveWindows.Connect("drag-begin", func(view *gtk.TreeView, ctx *gdk.DragContext) {
		originalListenerState = listenerState
		// Emit signal to stop global hotkey listener
		_, _ = listaVentanas.MainGUI.application.Emit(signalControlListener, false, false)

		// Get selected row
		if model, iter, ok := listaVentanas.treeSelectionActiveWindows.GetSelected(); ok {
			val, _ := model.ToTreeModel().GetValue(iter, columnOrder)
			if valInt, err := val.GoValue(); err == nil {
				idRowSelected = valInt.(int)
			}
		}
		listaVentanas.treeSelectionActiveWindows.HandlerBlock(signalChangedSelection)
	})
	// Handler of signal "drag-end". This signal is emitted when drag ends on the *gtk.TreeView
	listaVentanas.treeViewActiveWindows.Connect("drag-end", func(view *gtk.TreeView, ctx *gdk.DragContext) {
		var iterRowDesiredToSelect *gtk.TreeIter
		listaVentanas.listStoreActiveWindows.ForEach(
			func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
				value, _ := model.GetValue(iter, columnOrder)
				if valInt, err := value.GoValue(); err == nil {
					if valInt.(int) == idRowSelected {
						iterRowDesiredToSelect = iter
						return true
					}
				}
				return false
			},
		)
		listaVentanas.treeSelectionActiveWindows.HandlerUnblock(signalChangedSelection)
		if iterRowDesiredToSelect != nil {
			if path, err := listaVentanas.listStoreActiveWindows.GetPath(iterRowDesiredToSelect); err == nil {
				view.SetCursor(path, nil, false)
				listaVentanas.treeSelectionActiveWindows.SelectIter(iterRowDesiredToSelect)
			}
		}
		returnItemToInitialPos = false
		// Emit signal to start global hotkey listener
		_, _ = listaVentanas.MainGUI.application.Emit(signalControlListener, originalListenerState, false)
	})

	excludeWindow := false // Wether a window should be excluded or not (cuz it's invalid)
	obj, _ = listaVentanas.MainGUI.builder.GetObject("cellRenderExclude")
	cellRendererToglleColumnExcluded := obj.(*gtk.CellRendererToggle)
	// Handler of signal "toggled". This signal is emitted when selecting the column "exclude" on a row on the *gtk.TreeView
	// or directly emitting the signal
	cellRendererToglleColumnExcluded.Connect(
		"toggled",
		func(toggle *gtk.CellRendererToggle, path string) {
			// Make *gtk.TreeView unorderable and block signal when selection changes
			listaVentanas.treeViewActiveWindows.SetReorderable(false)
			listaVentanas.treeSelectionActiveWindows.HandlerBlock(signalChangedSelection)

			// *gtk.Iter (row) that is going to be excluded
			iter, _ := listaVentanas.listStoreActiveWindows.GetIterFromString(path)

			// Current value of column "Excluded" on desired row (iter)
			value, _ := listaVentanas.listStoreActiveWindows.GetValue(iter, columnExcluded)
			goValue, _ := value.GoValue()
			excluded := goValue.(bool)

			// Current value of column "Deleted" on desired row (iter)
			value, _ = listaVentanas.listStoreActiveWindows.GetValue(iter, columnDeletedWindow)
			goValue, _ = value.GoValue()
			deleted := goValue.(bool)

			handleToggle := true
			if excluded && deleted {
				handleToggle = false // If window is already excluded and deleted then no toggle
			}
			// Toggle handler
			if handleToggle {
				newVal := !excluded
				_ = listaVentanas.listStoreActiveWindows.SetValue(iter, columnExcluded, newVal) // Set new value

				glib.IdleAdd(func() {
					updateDefaultOrder := false
					if excludeWindow { // Window is not valid
						_ = listaVentanas.listStoreActiveWindows.SetValue(iter, columnDeletedWindow, true)

						value, _ := listaVentanas.listStoreActiveWindows.GetValue(iter, columnClass)
						goValue, _ := value.GoValue()
						class := goValue.(string)
						_ = listaVentanas.listStoreActiveWindows.SetValue(iter, columnClass, prefixClosedWindow+class)

						updateDefaultOrder = true
						excludeWindow = false
					} else {
						time.Sleep(time.Second / 4) // If the toggle was triggered by the UI we add this timeout
					}
					if newVal {
						// Move iter to end of the *gtk.TreeView
						listaVentanas.listStoreActiveWindows.MoveBefore(iter, nil)
					} else {
						// Search for the first excluded *gtk.Iter (window)
						var iterFirstexcludeWindow *gtk.TreeIter
						listaVentanas.listStoreActiveWindows.ForEach(func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
							// Value comun excluded
							value, _ = model.GetValue(iter, columnExcluded)
							goValue, _ = value.GoValue()
							excluded = goValue.(bool)

							if excluded {
								iterFirstexcludeWindow = iter
								return true
							}
							return false
						})
						// Move iter below the first excluded item if exists, otherwise move to the end of *gtk.TreeView
						if iterFirstexcludeWindow != nil {
							listaVentanas.listStoreActiveWindows.MoveBefore(iter, iterFirstexcludeWindow)
						}
					}
					// Emit signal to stablish order, a window was excluded/included
					_, _ = listaVentanas.MainGUI.application.Emit(signalSetOrder, true, updateDefaultOrder, true)
				})
			}

			// Make *gtk.TreeView reorderable and unblock signal when selection changes
			listaVentanas.treeSelectionActiveWindows.HandlerUnblock(signalChangedSelection)
			listaVentanas.treeViewActiveWindows.SetReorderable(!(excluded && deleted))
		},
	)

	listaVentanas.treeViewActiveWindows.Connect("key-press-event", func(view *gtk.TreeView, event *gdk.Event) bool {
		eventKey := gdk.EventKeyNewFromEvent(event)
		if eventKey.KeyVal() == gdk.KEY_Return {
			if model, iter, ok := listaVentanas.treeSelectionActiveWindows.GetSelected(); ok {
				path, _ := model.ToTreeModel().GetPath(iter)
				_, _ = cellRendererToglleColumnExcluded.Emit("toggled", path.String())
				return true // Stop propagation of event (default behaviour)
			}
		}
		return false
	})

	takeOffSelection := false // Wether the selection of the *gtk.TreeView should be removed
	showContextMenu := false  // Wether the contextual menu should be shown

	// Anonymous function that takes off the visible selection of the *gtk.TreeView
	functakeOffSelection := func() {
		amountOfItems := 0
		listaVentanas.listStoreActiveWindows.ForEach(
			func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
				val, _ := model.ToTreeModel().GetValue(iter, columnOrder)
				if valInt, err := val.GoValue(); err == nil {
					if valInt.(int) != 0 {
						amountOfItems++
					}
				}
				return false
			},
		)
		listaVentanas.treeViewActiveWindows.GrabFocus()
		if amountOfItems > 0 {
			path, _ := gtk.TreePathNewFromString(strconv.Itoa(amountOfItems + 1))
			listaVentanas.treeViewActiveWindows.SetCursor(path, nil, false)
			listaVentanas.treeSelectionActiveWindows.UnselectAll()
		}
	}

	listaVentanas.treeViewActiveWindows.Connect(
		"button-press-event",
		func(view *gtk.TreeView, event *gdk.Event) bool {
			eventButton := gdk.EventButtonNewFromEvent(event)
			if _, _, _, _, exists := view.GetPathAtPos(int(math.Round(eventButton.X())), int(math.Round(eventButton.Y()))); exists {
				if eventButton.Button() == gdk.BUTTON_SECONDARY {
					showContextMenu = true
				}
			} else {
				takeOffSelection = true
				return true // Stop propagation of event (default behaviour)
			}
			return false
		},
	)
	listaVentanas.treeViewActiveWindows.Connect(
		"button-release-event",
		func(view *gtk.TreeView, event *gdk.Event) bool {
			if takeOffSelection {
				functakeOffSelection()
				takeOffSelection = false
			} else if showContextMenu {
				disableContextMenu := true
				if eventButton := gdk.EventButtonNewFromEvent(event); eventButton.Button() == gdk.BUTTON_SECONDARY {
					if path, _, _, _, exists := view.GetPathAtPos(int(math.Round(eventButton.X())), int(math.Round(eventButton.Y()))); exists {
						iter, _ := listaVentanas.listStoreActiveWindows.GetIter(path)

						// Value of column excluded
						value, _ := listaVentanas.listStoreActiveWindows.GetValue(iter, columnExcluded)
						goValue, _ := value.GoValue()
						excluded := goValue.(bool)

						// Value of column cloned
						value, _ = listaVentanas.listStoreActiveWindows.GetValue(iter, columnCloned)
						goValue, _ = value.GoValue()
						cloned := goValue.(bool)

						// Value of column deleted
						value, _ = listaVentanas.listStoreActiveWindows.GetValue(iter, columnDeletedWindow)
						goValue, _ = value.GoValue()
						deleted := goValue.(bool)

						if !excluded || (excluded && cloned || excluded && deleted) {
							listaVentanas.createContextMenuTreeview(iter, !(cloned || deleted), eventButton.Event)
							disableContextMenu = false
						}
					}
				}
				if disableContextMenu {
					showContextMenu = false
				}
			}
			return false
		},
	)

	listaVentanas.treeViewActiveWindows.Connect("focus-in-event", func(view *gtk.TreeView, event *gdk.Event) bool {
		if path, _ := view.GetCursor(); path != nil && !takeOffSelection {
			view.SetCursor(path, nil, false)
			listaVentanas.treeSelectionActiveWindows.SelectPath(path)
		}
		return false
	})
	listaVentanas.treeViewActiveWindows.Connect("focus-out-event", func(view *gtk.TreeView, event *gdk.Event) bool {
		if !showContextMenu {
			listaVentanas.treeSelectionActiveWindows.UnselectAll()
		}
		return false
	})

	// Signal to set boolean "showContextMenu" to false
	_, _ = glib.SignalNew("disable-context-menu")
	// Handler
	listaVentanas.treeViewActiveWindows.Connect("disable-context-menu", func(view *gtk.TreeView) {
		showContextMenu = false
	})

	// Bold font in the *gtk.TreeView header
	cssProvider, _ := gtk.CssProviderNew()
	_ = cssProvider.LoadFromData("#treeViewActiveWindows.view header button { font-weight: bold; }")
	screen, _ := gdk.ScreenGetDefault()
	gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	// Handler of signal "app-set-order", it sets the current/default order based on the *gtk.TreeView of active windows
	listaVentanas.MainGUI.application.Connect(
		signalSetOrder,
		func(application *gtk.Application, resetCurrentOrder bool, resetDefaultOrder bool, useGUI bool) {
			var windowsCurrentOrder []window
			var windowsDefaultOrder []window
			if useGUI { // Set order from the UI
				listaVentanas.listStoreActiveWindows.ForEach(
					func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
						// Value of column excluded
						value, _ := model.GetValue(iter, columnExcluded)
						goValue, _ := value.GoValue()
						excluded := goValue.(bool)

						// Value of column deleted
						value, _ = model.GetValue(iter, columnDeletedWindow)
						goValue, _ = value.GoValue()
						deleted := goValue.(bool)

						// Value of column cloned
						value, _ = model.GetValue(iter, columnCloned)
						goValue, _ = value.GoValue()
						cloned := goValue.(bool)

						if !deleted {
							window := listaVentanas.getWindowFromRowIter(iter)
							if !cloned {
								windowsDefaultOrder = append(windowsDefaultOrder, *window)
							}
							if !excluded {
								windowsCurrentOrder = append(windowsCurrentOrder, *window)
							}
						}
						return false
					},
				)
			} else { // Set order from signal
				windowsCurrentOrder = listaVentanas.windowList
				windowsDefaultOrder = listaVentanas.windowList
			}
			var currentOrderText []string
			var defaultOrderText []string
			if resetCurrentOrder {
				currentOrder = windowsCurrentOrder
			}
			for _, window := range currentOrder {
				currentOrderText = append(currentOrderText, strconv.Itoa(window.order))
			}
			if resetDefaultOrder {
				defaultOrder = windowsDefaultOrder
			}
			for _, window := range defaultOrder {
				defaultOrderText = append(defaultOrderText, strconv.Itoa(window.order))
			}
			textToSetCurrentOrder := strings.Join(currentOrderText, ", ")
			textToSetDefaultOrder := strings.Join(defaultOrderText, ", ")
			if len(currentOrderText) == 0 {
				textToSetCurrentOrder = "-1"
			}
			if len(defaultOrderText) == 0 {
				textToSetDefaultOrder = "-1"
			}
			// Emit signal to stablish text of current/default order
			_, _ = listaVentanas.MainGUI.application.Emit(
				"gui-update-tex-order",
				textToSetCurrentOrder,
				textToSetDefaultOrder,
			)
		},
	)

	// Handler of signal "app-delete-window-order" used to delete a window that is no longer valid from the *gtk.TreeView
	listaVentanas.MainGUI.application.Connect(signalDeleteRow, func(application *gtk.Application, path string) {
		excludeWindow = true                                          // The window should be excluded, is not selectable
		_, _ = cellRendererToglleColumnExcluded.Emit("toggled", path) // Emit toggle signal
		functakeOffSelection()
	})
}

// Function that deletes all the rows from the *gtk.TreeView
func (listaVentanas *listaVentanas) clear() {
	listaVentanas.listStoreActiveWindows.HandlerBlock(listaVentanas.signalHandlerRowDeleted)
	listaVentanas.listStoreActiveWindows.Clear()
	listaVentanas.listStoreActiveWindows.HandlerUnblock(listaVentanas.signalHandlerRowDeleted)
}

// Function that pop ups a context menu on the *gtk.TreeView at a specific row cell
func (listaVentanas *listaVentanas) createContextMenuTreeview(iter *gtk.TreeIter, clone bool, event *gdk.Event) {
	menu, _ := gtk.MenuNew()

	menu.Connect("deactivate", func(menu *gtk.Menu) {
		_, _ = listaVentanas.treeViewActiveWindows.Emit("disable-context-menu")
	})

	cloneItem, _ := gtk.MenuItemNewWithLabel(funcGetStringResource("gui_treeview_context_menu_clone"))
	cloneItem.Connect("activate", func(menuItem *gtk.MenuItem) {
		listaVentanas.cloneRow(iter)
	})

	deleteItem, _ := gtk.MenuItemNewWithLabel(funcGetStringResource("delete"))
	deleteItem.Connect("activate", func(item *gtk.MenuItem) {
		listaVentanas.deleteRow(iter)
	})

	if clone {
		menu.Add(cloneItem)
	} else {
		menu.Add(deleteItem)
	}
	menu.ShowAll()
	menu.PopupAtPointer(event)
}

// Function that clones a row (*gtk.TreeIter) from the *gtk.TreeView of opened windows
// and puts it just below the original one
func (listaVentanas *listaVentanas) cloneRow(iter *gtk.TreeIter) {
	window := listaVentanas.getWindowFromRowIter(iter)
	window.class = prefixClonedWindow + window.class
	newOrder := 0
	listaVentanas.listStoreActiveWindows.ForEach(
		func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
			value, _ := model.GetValue(iter, columnOrder)
			goValue, _ := value.GoValue()
			order := goValue.(int)

			if order != 0 {
				newOrder++
			}
			return false
		},
	)
	if newOrder != 0 {
		window.order = newOrder + 1
	}
	listaVentanas.addRow(*window, true, iter)
	// Emit signal to stablish order
	_, _ = listaVentanas.MainGUI.application.Emit(signalSetOrder, true, false, true)
}

// Function that deletes a row (*gtk.TreeIter) from the *gtk.TreeView of opened windows
func (listaVentanas *listaVentanas) deleteRow(iter *gtk.TreeIter) {
	window := listaVentanas.getWindowFromRowIter(iter)
	listaVentanas.listStoreActiveWindows.ForEach(
		func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
			value, _ := model.GetValue(iter, columnOrder)
			goValue, _ := value.GoValue()
			order := goValue.(int)

			if order > window.order {
				_ = listaVentanas.listStoreActiveWindows.SetValue(iter, columnOrder, order-1)
			}
			return false
		},
	)
	listaVentanas.listStoreActiveWindows.Remove(iter)
	// Emit signal to stablish order
	_, _ = listaVentanas.MainGUI.application.Emit(signalSetOrder, true, false, true)
}

/*
Function that adds a row to the *gtk.TreeView of active windows in the position determined by
the field "order" of the parameter "window". If the boolean parameter "clonedWindow" is true and the parameter "iter"
is not nil then the new row is placed just below "iter".

Parameters:
  - window: Window to add
  - cloned: Wether the new window to add is cloned or not
  - iter: Iter where to put the new (cloned) window
*/
func (listaVentanas *listaVentanas) addRow(window window, clonedWindow bool, iter *gtk.TreeIter) {
	// Block signal "row-inserted"
	listaVentanas.listStoreActiveWindows.HandlerBlock(listaVentanas.signalHandlerRowInserted)

	var newIter *gtk.TreeIter
	if !clonedWindow {
		newIter = listaVentanas.listStoreActiveWindows.Insert(window.order - 1)
	} else {
		newIter = listaVentanas.listStoreActiveWindows.InsertAfter(iter)
	}
	_ = listaVentanas.listStoreActiveWindows.Set(
		newIter,
		[]int{
			columnOrder,
			columnId,
			columnClass,
			columnTitle,
			columnExcluded,
			columnCloned,
			columnPadding,
			columnEllipsize,
			columnFontWeight,
			columnDeletedWindow,
			columnIcon,
		},
		[]interface{}{
			window.order,
			window.id,
			window.class,
			window.title,
			false,
			clonedWindow,
			valuecolumnPadding,
			valueColumnPangoEllipsizeMode,
			valuecolumnFontWeight,
			false,
			window.icon,
		},
	)
	// Unblock signal "row-inserted"
	listaVentanas.listStoreActiveWindows.HandlerUnblock(listaVentanas.signalHandlerRowInserted)
}

// Function that returns a *window instance based on a row (*gtk.Iter) data
func (listaVentanas *listaVentanas) getWindowFromRowIter(iter *gtk.TreeIter) *window {
	value, _ := listaVentanas.listStoreActiveWindows.GetValue(iter, columnOrder)
	goValue, _ := value.GoValue()
	order := goValue.(int)

	value, _ = listaVentanas.listStoreActiveWindows.GetValue(iter, columnId)
	goValue, _ = value.GoValue()
	id := goValue.(string)

	value, _ = listaVentanas.listStoreActiveWindows.GetValue(iter, columnClass)
	goValue, _ = value.GoValue()
	class := goValue.(string)

	value, _ = listaVentanas.listStoreActiveWindows.GetValue(iter, columnTitle)
	goValue, _ = value.GoValue()
	title := goValue.(string)

	value, _ = listaVentanas.listStoreActiveWindows.GetValue(iter, columnIcon)
	goValue, _ = value.GoValue()
	icon := goValue.(*gdk.Pixbuf)

	return &window{id: id, class: class, title: title, order: order, icon: icon}
}
