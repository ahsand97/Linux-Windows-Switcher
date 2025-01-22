# Linux Windows Switcher
Linux desktop application to cycle through windows like when using ALT + TAB but it allows to specify the windows, reorder them, exclude them and set two global hotkeys to go forwards and backwards to cycle through the desired windows.

![Screenshot_20221110_003508](https://user-images.githubusercontent.com/32344641/201009983-692b2f9f-1001-4929-a46d-910aa085eb22.png)
![Screenshot_20221110_003657](https://user-images.githubusercontent.com/32344641/201009987-f5c006f4-49ac-4349-9eae-d5cb3e9f7c12.png)


## Features
- English, Spanish and French translation available based on locale
- Written in Go, very fast
- User Interface done with GTK3 (gotk3)
- Include/exclude windows by their class
- Define custom global hotkeys to go forwards or backwards
- Configuration of preferred classes can be saved
- AppIndicator on tray so the main window can be closed
- Change window title

# Usage
## From source
```bash
git clone https://github.com/ahsand97/Linux-Windows-Switcher.git
cd Linux-Windows-Switcher
go build -o linux-windows-switcher *.go
./linux-windows-switcher
```
## AppImage
An AppImage is provided to use the application. You can download it from the [releases](https://github.com/ahsand97/Linux-Windows-Switcher/releases).
#
## TODO
- Wayland support
