package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Lighthouse")

	profiles, err := LoadProfiles()
	if err != nil {
		log.Println("failed to load profiles:", err)
		w.SetContent(widget.NewLabel("Error loading profiles"))
	} else {
		names := make([]string, len(profiles))
		for i, p := range profiles {
			names[i] = p.Name
		}
		list := widget.NewList(
			func() int { return len(names) },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(i widget.ListItemID, o fyne.CanvasObject) {
				o.(*widget.Label).SetText(names[i])
			},
		)
		w.SetContent(container.NewVBox(
			widget.NewLabel("Profiles"),
			list,
		))
	}
	w.ShowAndRun()
}
