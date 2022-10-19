package appindicator

import (
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/gtk"
)

type Indicator struct {
	application *gtk.Application
	indicator   *appindicator.Indicator
}

var (
	icons                 []string
	opcionControlListener *gtk.MenuItem
)

const (
	// Nombres de las señales de la *gtk.Application usadas
	signalAbrirVentanaPrincipal = "app-abrir-ventana"
	signalReiniciar             = "app-reiniciar"
	signalSalir                 = "app-salir"
	signalControlListener       = "app-listener-keyboard"

	// Indices de los íconos dentro del slice de iconos
	iconInactive = 0
	iconActive   = 1
)

// NewAppIndicator Constructor AppIndicator
func NewAppIndicator(application *gtk.Application, iconPaths []string, title string) *Indicator {
	icons = iconPaths

	indicator := appindicator.New(
		application.GetApplicationID(),
		icons[iconInactive],
		appindicator.CategoryApplicationStatus,
	)
	indicator.SetTitle(title)
	indicator.SetStatus(appindicator.StatusActive)

	indicator_ := &Indicator{application: application, indicator: indicator}

	indicator_.setupIndicator()
	return indicator_
}

// Función de configuración
func (indicator *Indicator) setupIndicator() {
	menu, _ := gtk.MenuNew()
	defer menu.ShowAll()

	abrirVentanaPrincipal, _ := gtk.MenuItemNewWithLabel("Abrir ventana Principal")
	abrirVentanaPrincipal.Connect("activate", func(menuItem *gtk.MenuItem) {
		_, _ = indicator.application.Emit(signalAbrirVentanaPrincipal)
	})

	opcionControlListener, _ = gtk.MenuItemNewWithLabel("Activar Listener")
	opcionControlListener.Connect("activate", func(menuItem *gtk.MenuItem) {
		newState := false
		if menuItem.GetLabel() == "Activar Listener" {
			newState = true
		}
		// Se emite la señal para iniciar/detener el listener del teclado
		_, _ = indicator.application.Emit(signalControlListener, newState, true)
	})

	reiniciarAplicacion, _ := gtk.MenuItemNewWithLabel("Reiniciar Aplicación")
	reiniciarAplicacion.Connect("activate", func(menuItem *gtk.MenuItem) {
		_, _ = indicator.application.Emit(signalReiniciar)
	})

	salir, _ := gtk.MenuItemNewWithLabel("Salir")
	salir.Connect("activate", func(menuItem *gtk.MenuItem) {
		_, _ = indicator.application.Emit(signalSalir)
	})

	menu.Add(abrirVentanaPrincipal)
	menu.Add(opcionControlListener)
	menu.Add(reiniciarAplicacion)
	menu.Add(salir)
	indicator.indicator.SetMenu(menu)
}

// UpdateIconState Función que actualiza el icono del indicator
func (indicator *Indicator) UpdateIconState(newState bool) {
	icon := icons[iconActive]
	if !newState {
		icon = icons[iconInactive]
	}
	indicator.indicator.SetIcon(icon)

	label := "Desactivar Listener"
	if !newState {
		label = "Activar Listener"
	}
	opcionControlListener.SetLabel(label)
}
