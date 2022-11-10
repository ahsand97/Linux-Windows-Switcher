package gui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"linux-windows-switcher/libs/glibown"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type contentTabVentanas struct {
	MainGUI
	listWindowClass            []window
	listPreferredClasses       []string
	listExcludedClasses        []string
	expander                   *gtk.Expander
	listBoxActiveWindowClasses *gtk.ListBox
	listBoxPreferredClasses    *gtk.ListBox
	listBoxExcludedClasses     *gtk.ListBox
	windowList                 *listaVentanas
}

const (
	// Section from config file related with the preferred/excluded classes config
	sectionClasses = "classes"

	// Options inside config file
	optionPreferredClasses = "preferred_classes"
	optionExcludedClasses  = "excluded_classes"
)

// newContentTabVentanas Constructor
func (mainGUI *MainGUI) newContentTabVentanas() *contentTabVentanas {
	contentTabVentanas := &contentTabVentanas{MainGUI: *mainGUI}

	contentTabVentanas.windowList = contentTabVentanas.newListaVentanas()
	contentTabVentanas.initLocale()
	contentTabVentanas.setupContentTabVentanas()
	return contentTabVentanas
}

// Config function to assign all UI strings to appropriate locale
func (contentTabVentanas *contentTabVentanas) initLocale() {
	obj, _ := contentTabVentanas.MainGUI.builder.GetObject("expanderLabel")
	expanderLabel := obj.(*gtk.Label)
	expanderLabel.SetMarkup(funcGetStringResource("gui_expander_label"))
	expanderLabel.SetTooltipText(funcGetStringResource("gui_expander_label_tooltip"))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelTitleListBoxActiveWindowClasses")
	labelTitleListBoxActiveWindowClasses := obj.(*gtk.Label)
	labelTitleListBoxActiveWindowClasses.SetMarkup(funcGetStringResource("gui_label_title_listbox_active_classes"))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelTitleListBoxPreferredClasses")
	labelTitleListBoxPreferredClasses := obj.(*gtk.Label)
	labelTitleListBoxPreferredClasses.SetMarkup(funcGetStringResource("gui_label_title_listbox_preferred_classes"))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelTitleListBoxExcludedClasses")
	labelTitleListBoxExcludedClasses := obj.(*gtk.Label)
	labelTitleListBoxExcludedClasses.SetMarkup(funcGetStringResource("gui_label_title_listbox_excluded_classes"))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelButtonRefreshWindowClasses")
	labelButtonRefreshWindowClasses := obj.(*gtk.Label)
	labelButtonRefreshWindowClasses.SetMarkup(funcGetStringResource("refresh"))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelTitleWindowsOrder")
	labelTitleWindowsOrder := obj.(*gtk.Label)
	labelTitleWindowsOrder.SetMarkup(funcGetStringResource("gui_title_windows_order"))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelExplanationWindowsOrder")
	labelExplanationWindowsOrder := obj.(*gtk.Label)
	labelExplanationWindowsOrder.SetMarkup(funcGetStringResource("gui_explanation_windows_order"))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelCurrentOrderTitle")
	labelCurrentOrderTitle := obj.(*gtk.Label)
	labelCurrentOrderTitle.SetMarkup(fmt.Sprintf("%s:", funcGetStringResource("gui_label_current_order")))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelDefaultOrderTitle")
	labelDefaultOrderTitle := obj.(*gtk.Label)
	labelDefaultOrderTitle.SetMarkup(fmt.Sprintf("%s:", funcGetStringResource("gui_label_default_order")))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelRefreshOpenWindows")
	labelRefreshOpenWindows := obj.(*gtk.Label)
	labelRefreshOpenWindows.SetMarkup(funcGetStringResource("gui_label_refresh_open_windows"))

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelRestoreDefaultOrder")
	labelRestoreDefaultOrder := obj.(*gtk.Label)
	labelRestoreDefaultOrder.SetMarkup(funcGetStringResource("gui_label_restore_default_order"))
}

// Config function
func (contentTabVentanas *contentTabVentanas) setupContentTabVentanas() {
	obj, _ := contentTabVentanas.MainGUI.builder.GetObject("listBoxActiveWindowClasses")
	contentTabVentanas.listBoxActiveWindowClasses = obj.(*gtk.ListBox)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("listBoxPreferredClasses")
	contentTabVentanas.listBoxPreferredClasses = obj.(*gtk.ListBox)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("listBoxExcludedClasses")
	contentTabVentanas.listBoxExcludedClasses = obj.(*gtk.ListBox)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelCurrentOrder")
	labelCurrentOrder := obj.(*gtk.Label)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("labelDefaultOrder")
	labelDefaultOrder := obj.(*gtk.Label)

	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("buttonRefreshWindowClasses")
	buttonRefreshWindowClasses := obj.(*gtk.Button)
	buttonRefreshWindowClasses.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() {
				contentTabVentanas.listWindowClass = nil
				contentTabVentanas.listBoxActiveWindowClasses.GetChildren().Foreach(func(item interface{}) {
					contentTabVentanas.listBoxActiveWindowClasses.Remove(item.(*gtk.Widget))
				})
				button.SetSensitive(false)
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				contentTabVentanas.getClassesCurrentWindows()
				button.SetSensitive(true)
			})
		}()
	})

	// Container of *gtk.TreeView of opened windows
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("containerTreeViewWindows")
	containerTreeViewWindows := obj.(*gtk.Box)

	// Container of all the section related with the order both current and default
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("containerSectionOrder")
	containerSectionOrder := obj.(*gtk.Box)

	// Auxiliar slices used when expander expands to check if there was any change in the configuration
	var listPreferredClassesPrevious []string
	var listExcludedClassesPrevious []string

	// Expander
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("expander")
	contentTabVentanas.expander = obj.(*gtk.Expander)
	contentTabVentanas.expander.Connect("activate", func(expander *gtk.Expander) {
		expanded := expander.GetExpanded()

		contentTabVentanas.MainGUI.buttonControlListener.SetSensitive(expanded)
		containerTreeViewWindows.SetSensitive(expanded)
		containerSectionOrder.SetSensitive(expanded)

		if expanded {
			clasesPreferidasIguales := strings.Join(
				contentTabVentanas.listPreferredClasses,
				",",
			) == strings.Join(
				listPreferredClassesPrevious,
				",",
			)
			clasesExcluidasIguales := strings.Join(
				contentTabVentanas.listExcludedClasses,
				",",
			) == strings.Join(
				listExcludedClassesPrevious,
				",",
			)
			// If there was any change in classes config reset the *gtk.Treeview
			if !clasesPreferidasIguales || !clasesExcluidasIguales {
				contentTabVentanas.windowList.clear()
				contentTabVentanas.getActiveWindows(true, true)
			}

			// Emit signal to start global hotkey listener
			_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, true, false)
		} else {
			listPreferredClassesPrevious = nil
			listPreferredClassesPrevious = append(listPreferredClassesPrevious, contentTabVentanas.listPreferredClasses...)

			listExcludedClassesPrevious = nil
			listExcludedClassesPrevious = append(listExcludedClassesPrevious, contentTabVentanas.listExcludedClasses...)

			// Emit signal to stop global hotkey listener
			_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, false, false)
			contentTabVentanas.listWindowClass = nil

			// Remove all items from the *gtk.ListBox of active window-classes
			contentTabVentanas.listBoxActiveWindowClasses.GetChildren().Foreach(func(item interface{}) {
				contentTabVentanas.listBoxActiveWindowClasses.Remove(item.(*gtk.Widget))
			})
			// Get all active window-classes
			contentTabVentanas.getClassesCurrentWindows()
		}
	})
	contentTabVentanas.expander.ConnectAfter("activate", func(expander *gtk.Expander) {
		width, height := contentTabVentanas.MainGUI.window.GetSize()
		if !contentTabVentanas.MainGUI.window.IsMaximized() && (width > 0 && height > 0) {
			contentTabVentanas.MainGUI.window.Resize(width, 1)
		}
	})

	// Button refesh open windows
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("buttonRefreshOpenWindows")
	buttonRefreshOpenWindows := obj.(*gtk.Button)
	buttonRefreshOpenWindows.Connect("clicked", func(button *gtk.Button) {
		go func() {
			glib.IdleAdd(func() {
				// Emit signal to stop global hotkey listener
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, false, false)
				// Empty *gtk.TreeView of active windows
				contentTabVentanas.windowList.clear()
				// Clear content of current/default order
				labelCurrentOrder.SetMarkup("")
				labelDefaultOrder.SetMarkup("")
				button.SetSensitive(false)
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				// Get all active window-classes
				contentTabVentanas.getActiveWindows(false, true)
				// Emit signal to start global hotkey listener
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, true, false)
				button.SetSensitive(true)
			})
		}()
	})

	// Button restore default order
	obj, _ = contentTabVentanas.MainGUI.builder.GetObject("buttonRestoreOrder")
	buttonRestoreOrder := obj.(*gtk.Button)
	buttonRestoreOrder.Connect("clicked", func(button *gtk.Button) {
		go func() {
			// If both current and default order are the same the button gets animated and that's it
			if funcTestEq(currentOrder, defaultOrder) {
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
				// Emit signal to stop global hotkey listener
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, false, false)
				// Clear *gtk.TreeView
				contentTabVentanas.windowList.clear()
				// Clear content of current order
				labelCurrentOrder.SetMarkup("")
			})
			time.Sleep(time.Second / 3)
			glib.IdleAdd(func() {
				contentTabVentanas.windowList.windowList = defaultOrder
				// All active windows are added to the *gtk.TreeView based on default order
				for _, window := range defaultOrder {
					contentTabVentanas.windowList.addRow(window, false, nil)
				}
				// Emit signal to stablish order
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalSetOrder, true, false, false)
				// Emit signal to start global hotkey listener
				_, _ = contentTabVentanas.MainGUI.application.Emit(signalControlListener, true, false)
				button.SetSensitive(true)
			})
		}()
	})

	// Map containing currently active windows, it is used to query the icon when loading
	// config from file
	activeWindows := map[string]window{}
	for _, activeWindow := range listWindows(true) {
		activeWindow.class = getClass(activeWindow.class)
		activeWindows[activeWindow.class] = activeWindow
	}

	/*
		Anonymous function that load the preferred/excluded classes from config file and add them to their
		associated list and *gtk.ListBox.

		Param:
			- option: Option to search from config file
			- list: List where to add the class
			- box: *gtk.ListBox where to add the class
	*/
	getClasses := func(option string, list *[]string, box *gtk.ListBox) {
		result, _ := contentTabVentanas.MainGUI.application.Emit(signalGetConfig, sectionClasses, option)
		classString := result.(string)
		if len(classString) > 0 {
			for _, class := range strings.Split(classString, ",") {
				class = strings.TrimSpace(class)
				if !contains(*list, class) {
					*list = append(*list, class)
					window := window{class: class}
					if val, ok := activeWindows[class]; ok {
						window.icon = val.icon
					}
					// Add item to its ListBox, wether it is the one with preferred classes or the one with excluded classes
					box.Add(contentTabVentanas.createListBoxRow(window, false, box))
				}
			}
		}
	}

	// Get preferred classes from config file
	getClasses(
		optionPreferredClasses,
		&contentTabVentanas.listPreferredClasses,
		contentTabVentanas.listBoxPreferredClasses,
	)
	// Get excluded classes from config file
	getClasses(
		optionExcludedClasses,
		&contentTabVentanas.listExcludedClasses,
		contentTabVentanas.listBoxExcludedClasses,
	)

	// Signal to update the text on the GUI related with current/default order of windows
	_, _ = glibown.SignalNewV("gui-update-tex-order", glib.TYPE_NONE, 2, glib.TYPE_STRING, glib.TYPE_STRING)
	// Handler
	contentTabVentanas.MainGUI.application.Connect(
		"gui-update-tex-order",
		func(application *gtk.Application, textLabelCurrentOrder string, textLabelDefaultOrder string) {
			if len(textLabelCurrentOrder) > 0 {
				if textLabelCurrentOrder == "-1" {
					labelCurrentOrder.SetMarkup("")
				} else {
					labelCurrentOrder.SetMarkup(fmt.Sprintf("[%s]", textLabelCurrentOrder))
				}
			}
			if len(textLabelDefaultOrder) > 0 {
				if textLabelDefaultOrder == "-1" {
					labelDefaultOrder.SetMarkup("")
				} else {
					labelDefaultOrder.SetMarkup(fmt.Sprintf("[%s]", textLabelDefaultOrder))
				}
			}
		},
	)

	// Signal to enable the button associated with the *gtk.ListBoxRow of active window-classes
	_, _ = glib.SignalNew("listBoxActiveWindowClasses-enable-button")

	// Get all active windows taking into consideration the config of preferred/excluded classes
	contentTabVentanas.getActiveWindows(true, true)
}

// Get current active window-classes and add them to the *gtk.ListBox "listBoxActiveWindowClasses"
func (contentTabVentanas *contentTabVentanas) getClassesCurrentWindows() {
	windowClasses := map[window]bool{}

	for _, windowActive := range listWindows(true) {
		if strings.Contains(windowActive.class, contentTabVentanas.MainGUI.application.GetApplicationID()) ||
			windowActive.desktop == -1 || windowClasses[windowActive] {
			continue
		}
		alreadyAdded := false
		for windowKlass := range windowClasses {
			if strings.Contains(windowActive.class, windowKlass.class) {
				alreadyAdded = true
				break
			}
		}
		if alreadyAdded {
			continue
		}
		windowClasses[windowActive] = true
	}
	for windowKlass, active := range windowClasses {
		if !active {
			continue
		}
		windowKlass.class = getClass(windowKlass.class)
		contentTabVentanas.listWindowClass = append(contentTabVentanas.listWindowClass, windowKlass)
	}
	sort.Slice(contentTabVentanas.listWindowClass, func(i, j int) bool {
		return strings.ToLower(
			contentTabVentanas.listWindowClass[i].class,
		) < strings.ToLower(
			contentTabVentanas.listWindowClass[j].class,
		)
	})
	for _, windowClass := range contentTabVentanas.listWindowClass {
		// Add items to main ListBox, showing all currently active window-classes
		contentTabVentanas.listBoxActiveWindowClasses.Add(contentTabVentanas.createListBoxRow(windowClass, true, nil))
	}
}

// Function that get all currently active windows taking into consideration config of preferred/excluded classes
func (contentTabVentanas *contentTabVentanas) getActiveWindows(resetCurrentOrder bool, resetDefaultOrder bool) {
	validWindows := []window{}
	for _, windowActive := range listWindows(true) {
		if strings.Contains(windowActive.class, contentTabVentanas.MainGUI.application.GetApplicationID()) ||
			windowActive.desktop == -1 {
			continue
		}
		valid := true
		if len(contentTabVentanas.listPreferredClasses) > 0 {
			valid = false
			for _, preferredClass := range contentTabVentanas.listPreferredClasses {
				if strings.Contains(windowActive.class, preferredClass) {
					valid = true
					break
				}
			}
		}
		if len(contentTabVentanas.listExcludedClasses) > 0 {
			for _, excludedClass := range contentTabVentanas.listExcludedClasses {
				if strings.Contains(windowActive.class, excludedClass) {
					valid = false
					break
				}
			}
		}
		if valid {
			validWindows = append(validWindows, windowActive)
		}
	}
	contentTabVentanas.windowList.windowList = []window{}
	for index, validWindow := range validWindows {
		validWindow.class = getClass(validWindow.class)
		validWindow.order = index + 1
		contentTabVentanas.windowList.windowList = append(contentTabVentanas.windowList.windowList, validWindow)
		contentTabVentanas.windowList.addRow(validWindow, false, nil)
	}

	if !resetCurrentOrder {
		// If the length of current order and default order are different or if length of current order is 0 then it gets reset
		if len(currentOrder) != len(contentTabVentanas.windowList.windowList) || len(currentOrder) == 0 {
			resetCurrentOrder = true
		}
		// If both current order and default order have same length then we check if both have same windows, if not
		// current order is reset
		for _, windowCurrentOrder := range currentOrder {
			windowExist := false
			for _, newWindow := range contentTabVentanas.windowList.windowList {
				if newWindow.id == windowCurrentOrder.id {
					windowExist = true
					break
				}
			}
			if !windowExist {
				resetCurrentOrder = true
				break
			}
		}
	}
	// Emit signal to stablish order
	_, _ = contentTabVentanas.MainGUI.application.Emit(
		signalSetOrder,
		resetCurrentOrder,
		resetDefaultOrder,
		false,
	)
}

/*
Function that creates a *gtk.ListBoxRow ready to be added to a *gtk.ListBox.

Parameters:

  - windowClass: windowClass

  - add:   Handles the functionallity of the button inside the *gtk.ListBoxRow. If true, the button "add" is shown else "delete"

  - box:   *gtk.ListBox where to delete items when the callback of button "delete" is triggered. It can be nil
*/
func (contentTabVentanas *contentTabVentanas) createListBoxRow(
	windowClass window,
	add bool,
	box *gtk.ListBox,
) *gtk.ListBoxRow {
	builder := getNewBuilder()

	// Config anonymous function to assign all UI strings to appropriate locale
	initLocale := func() {
		obj, _ := builder.GetObject("labelButtonAddClassToListBox")
		labelButtonAddClassToListBox := obj.(*gtk.Label)
		labelButtonAddClassToListBox.SetMarkup(funcGetStringResource("add"))

		obj, _ = builder.GetObject("labelButtonDeleteClassFromListBox")
		labelButtonDeleteClassFromListBox := obj.(*gtk.Label)
		labelButtonDeleteClassFromListBox.SetMarkup(funcGetStringResource("delete"))
	}
	initLocale()

	obj, _ := builder.GetObject("listBoxRowClass")
	listboxRow := obj.(*gtk.ListBoxRow)
	listboxRow.SetName(windowClass.class)

	obj, _ = builder.GetObject("labelClassListBoxRowClass")
	labelClass := obj.(*gtk.Label)
	labelClass.SetMarkup(windowClass.class)
	labelClass.SetTooltipText(windowClass.class)

	if windowClass.icon != nil {
		obj, _ = builder.GetObject("imageRowWindowClass")
		imageRow := obj.(*gtk.Image)
		imageRow.SetFromPixbuf(windowClass.icon)
	}

	var button_ *gtk.Widget
	if add { // Show button "Add"
		obj, _ = builder.GetObject("buttonAddToListBox")
		button := obj.(*gtk.MenuButton)
		button.SetTooltipText(funcGetStringResource("gui_tooltip_add_class_to_listbox"))
		button_ = &button.Widget

		menu, _ := gtk.MenuNew()
		menu.SetVAlign(gtk.ALIGN_CENTER)

		addToPreferredClasses, _ := gtk.MenuItemNewWithLabel(funcGetStringResource("gui_class_add_to_preferred_list"))
		addToPreferredClasses.Connect("activate", func(menuItem *gtk.MenuItem) {
			contentTabVentanas.addITemToListBox(
				windowClass,
				optionPreferredClasses,
				&contentTabVentanas.listPreferredClasses,
				button_,
				contentTabVentanas.listBoxPreferredClasses,
			)
		})

		addClasesExcluidas, _ := gtk.MenuItemNewWithLabel(funcGetStringResource("gui_class_add_to_excluded_list"))
		addClasesExcluidas.Connect("activate", func(menuItem *gtk.MenuItem) {
			contentTabVentanas.addITemToListBox(
				windowClass,
				optionExcludedClasses,
				&contentTabVentanas.listExcludedClasses,
				button_,
				contentTabVentanas.listBoxExcludedClasses,
			)
		})

		menu.Add(addToPreferredClasses)
		menu.Add(addClasesExcluidas)
		menu.ShowAll()
		button.SetPopup(menu)

		// Anonymous function that checks if the list of preferred/excluded classes
		// already contains windowClass so make it non-sensitive
		checkButton := func() {
			button.SetSensitive(
				!(contains(contentTabVentanas.listPreferredClasses, windowClass.class) || contains(contentTabVentanas.listExcludedClasses, windowClass.class)),
			)
		}
		checkButton()

		// Handler of signal
		listboxRow.Connect("listBoxActiveWindowClasses-enable-button", func(row *gtk.ListBoxRow) {
			checkButton()
		})
	} else { // Show button "Delete"
		obj, _ = builder.GetObject("buttonDeleteFromListBox")
		button := obj.(*gtk.Button)
		button_ = &button.Widget

		textTooltip := funcGetStringResource("gui_class_delete_from_preferred_list")
		associatedList := &contentTabVentanas.listPreferredClasses
		name, _ := box.GetName()
		if name == "listBoxExcludedClasses" {
			textTooltip = funcGetStringResource("gui_class_delete_from_excluded_list")
			associatedList = &contentTabVentanas.listExcludedClasses
		}
		button.SetTooltipText(textTooltip)
		button.Connect("clicked", func(button *gtk.Button) {
			option := optionPreferredClasses
			if name == "listBoxExcludedClasses" {
				option = optionExcludedClasses
			}
			contentTabVentanas.deleteItemFromListBox(windowClass.class, option, associatedList, button_, box, listboxRow)
		})
	}
	button_.SetVisible(true) // Both buttons are not visible, only the one needed is shown
	return listboxRow
}

/*
Function that adds an item to the *gtk.ListBox of preferred/excluded classes.

Parameters:
  - windowClass: window containing the desired class
  - option: Option from config file to edit
  - associatedList: List to edit
  - associatedButton: Button with the item's functionallity
  - targetListBox: *gtk.ListBox target
*/
func (contentTabVentanas *contentTabVentanas) addITemToListBox(
	windowClass window,
	option string,
	associatedList *[]string,
	associatedButton *gtk.Widget,
	targetListBox *gtk.ListBox,
) {
	if contains(*associatedList, windowClass.class) {
		return
	}
	*associatedList = append(*associatedList, windowClass.class)

	// Update config file when adding a class
	result, _ := contentTabVentanas.MainGUI.application.Emit(
		signalUpdateConfig,
		sectionClasses,
		option,
		strings.Join(*associatedList, ","),
	)
	go func() {
		glib.IdleAdd(func() {
			associatedButton.SetSensitive(false)
		})
		time.Sleep(time.Second / 3)
		glib.IdleAdd(func() {
			targetListBox.Add(contentTabVentanas.createListBoxRow(windowClass, false, targetListBox))
			if !result.(bool) {
				msg2 := funcGetStringResource("gui_class_error_add_to_preferred_listbox")
				if option == optionExcludedClasses {
					msg2 = funcGetStringResource("gui_class_error_add_to_excluded_listbox")
				}
				contentTabVentanas.MainGUI.showMessageDialog(
					gtk.MESSAGE_ERROR,
					funcGetStringResource("config_error_update_file"),
					msg2,
				)
			}
		})
	}()
}

/*
Function that deletes an item from the *gtk.ListBox of preferred/excluded classes.

Parameters:
  - class: Class to delete
  - option: Option from config file to update
  - associatedList: Associated list to update
  - associatedListBox: *gtk.ListBox where to delete the item
  - associatedListBoxRow: *gtk.ListBoxRow to delete
*/
func (contentTabVentanas *contentTabVentanas) deleteItemFromListBox(
	class string,
	option string,
	associatedList *[]string,
	associatedButton *gtk.Widget,
	associatedListBox *gtk.ListBox,
	associatedListBoxRow *gtk.ListBoxRow,
) {
	*associatedList = removeItem(*associatedList, class)
	// Update config file when removing class
	result, _ := contentTabVentanas.MainGUI.application.Emit(
		signalUpdateConfig,
		sectionClasses,
		option,
		strings.Join(*associatedList, ","),
	)
	go func() {
		glib.IdleAdd(func() {
			associatedButton.SetSensitive(false)
		})
		time.Sleep(time.Second / 3)
		glib.IdleAdd(func() {
			var associatedItem *gtk.Widget
			isClassCurrentlyShown := false
			for _, windowKlass := range contentTabVentanas.listWindowClass {
				if strings.Contains(windowKlass.class, class) || strings.Contains(class, windowKlass.class) {
					isClassCurrentlyShown = true
					break
				}
			}
			// If class is currently visible in the *gtk.ListBox of active classes emit signal
			// to activate its button again
			if isClassCurrentlyShown {
				for list := contentTabVentanas.listBoxActiveWindowClasses.GetChildren(); list != nil; list = list.Next() {
					item := list.Data().(*gtk.Widget)
					if name, err := item.GetName(); err == nil {
						if name == class {
							associatedItem = item
							break
						}
					}
				}
			}
			// Remove *gtk.ListBoxRow from desired *gtk.ListBox
			associatedListBox.Remove(associatedListBoxRow)
			if associatedItem != nil {
				// Emit signal to enable button if item is in *gtk.ListBox of active window-classes
				_, _ = associatedItem.Emit("listBoxActiveWindowClasses-enable-button")
			}
			if !result.(bool) {
				msg2 := funcGetStringResource("gui_class_error_update_preferred_listbox")
				if option == optionExcludedClasses {
					msg2 = funcGetStringResource("gui_class_error_update_excluded_listbox")
				}
				contentTabVentanas.MainGUI.showMessageDialog(
					gtk.MESSAGE_ERROR,
					funcGetStringResource("config_error_update_file"),
					msg2,
				)
			}
			associatedButton.SetSensitive(true)
		})
	}()
}
