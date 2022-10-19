# Linux-Windows-Switcher
Linux Desktop Application to cycle through windows the same way that you would do with ALT + TAB but it allows you to specify the windows, reorder them, exclude them and set two global hoykeys to go forwards and backwards to cycle through the desired windows.

## Features
- Written in Go, very fast
- User Interface done with GTK3 (gotk3)
- Include/exclude windows by their class
- Define custom global hotkeys to go forwards or backwards
- Configuration of preferred classes can be saved
- AppIndicator on tray so the main window can be closed

# Usage
## From source
This application requires `wmctrl` and `awk`.
```bash
git clone https://github.com/ahsand97/Linux-Windows-Switcher.git
cd Linux-Windows-Switcher
go mod tidy
go build -o linux-windows-switcher *.go
./linux-windows-switcher
```

#
## TODO
- English traslation
- AppImage generation