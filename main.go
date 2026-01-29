package main

import (
	"4corners/gui"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	gui.ShowMainWindow(a)
	a.Run()
}
