package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// showProfileDialog displays a dialog for creating or editing a profile.
// If p is non-nil, its fields are used to pre-populate the dialog entries.
// onSave is called with the resulting profile when the user confirms the dialog.
func showProfileDialog(w fyne.Window, title string, p *Profile, onSave func(Profile)) {
	nameEntry := widget.NewEntry()
	ipEntry := widget.NewEntry()

	if p != nil {
		nameEntry.SetText(p.Name)
		ipEntry.SetText(p.IPAddress)
	}

	dialog.ShowForm(title, "Save", "Cancel", []*widget.FormItem{
		{Text: "Name", Widget: nameEntry},
		{Text: "IP", Widget: ipEntry},
	}, func(b bool) {
		if !b {
			return
		}
		onSave(Profile{Name: nameEntry.Text, IPAddress: ipEntry.Text})
	}, w)
}

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
		selected := -1
		list.OnSelected = func(id widget.ListItemID) {
			selected = int(id)
		}

		createButton := widget.NewButton("Create", func() {
			showProfileDialog(w, "New Profile", nil, func(p Profile) {
				profiles = append(profiles, p)
				names = append(names, p.Name)
				if err := SaveProfiles(profiles); err != nil {
					log.Println("failed to save profiles:", err)
				}
				list.Refresh()
			})
		})

		renameButton := widget.NewButton("Rename", func() {
			if selected >= 0 && selected < len(profiles) {
				existing := profiles[selected]
				showProfileDialog(w, "Edit Profile", &existing, func(p Profile) {
					profiles[selected] = p
					names[selected] = p.Name
					if err := SaveProfiles(profiles); err != nil {
						log.Println("failed to save profiles:", err)
					}
					list.Refresh()
				})
			}
		})

		deleteButton := widget.NewButton("Delete", func() {
			if selected >= 0 && selected < len(profiles) {
				profiles = append(profiles[:selected], profiles[selected+1:]...)
				names = append(names[:selected], names[selected+1:]...)
				if err := SaveProfiles(profiles); err != nil {
					log.Println("failed to save profiles:", err)
				}
				list.UnselectAll()
				selected = -1
				list.Refresh()
			}
		})

		activateButton := widget.NewButton("Activate", func() {
			if selected >= 0 && selected < len(profiles) {
				if err := ActivateProfile(profiles[selected]); err != nil {
					log.Println("activate profile:", err)
				}
			}
		})

		w.SetContent(container.NewVBox(
			widget.NewLabel("Profiles"),
			list,
			container.NewHBox(createButton, renameButton, deleteButton, activateButton),
		))
	}
	w.ShowAndRun()
}
