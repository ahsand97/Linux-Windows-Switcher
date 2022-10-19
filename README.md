# Linux-Windows-Switcher
Linux desktop application to cycle through windows like when using ALT + TAB but it allows to specify the windows, reorder them, exclude them and set two global hotkeys to go forwards and backwards to cycle through the desired windows. 

<img src="https://user-images.githubusercontent.com/32344641/196622904-7769213b-cb4c-46c5-b715-44b3a714e517.png" width="510" height="468" />
<img src="https://user-images.githubusercontent.com/32344641/196622923-e07b60c5-4d43-46cb-ad56-41eaa9718b10.png" width="509.4" height="466.2" />


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
