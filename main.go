package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

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

	d := dialog.NewForm(title, "Save", "Cancel", []*widget.FormItem{
		{Text: "Name", Widget: nameEntry},
		{Text: "IP", Widget: ipEntry},
	}, func(b bool) {
		if !b {
			return
		}
		onSave(Profile{Name: nameEntry.Text, IPAddress: ipEntry.Text})
	}, w)
	d.Resize(fyne.NewSize(400, d.MinSize().Height))
	d.Show()
}

// showTunnelDialog displays a dialog for creating or editing a tunnel configuration.
// If t is non-nil, its fields populate the dialog entries. onSave is called with the
// resulting tunnel when the user confirms the dialog.
func showTunnelDialog(w fyne.Window, title string, t *Tunnel, onSave func(Tunnel)) {
	nameEntry := widget.NewEntry()
	sshServerEntry := widget.NewEntry()
	sshPortEntry := widget.NewEntry()
	sshUserEntry := widget.NewEntry()
	sshKeyPathEntry := widget.NewEntry()
	remoteHostEntry := widget.NewEntry()
	remotePortEntry := widget.NewEntry()
	domainEntry := widget.NewEntry()
	localPortEntry := widget.NewEntry()

	if t != nil {
		nameEntry.SetText(t.Name)
		sshServerEntry.SetText(t.SSHServer)
		sshPortEntry.SetText(fmt.Sprintf("%d", t.SSHPort))
		sshUserEntry.SetText(t.SSHUser)
		sshKeyPathEntry.SetText(t.SSHKeyPath)
		remoteHostEntry.SetText(t.RemoteHost)
		remotePortEntry.SetText(fmt.Sprintf("%d", t.RemotePort))
		domainEntry.SetText(t.Domain)
		localPortEntry.SetText(fmt.Sprintf("%d", t.LocalPort))
	}

	d := dialog.NewForm(title, "Save", "Cancel", []*widget.FormItem{
		{Text: "Name", Widget: nameEntry},
		{Text: "SSH Server", Widget: sshServerEntry},
		{Text: "SSH Port", Widget: sshPortEntry},
		{Text: "SSH User", Widget: sshUserEntry},
		{Text: "SSH Key Path", Widget: sshKeyPathEntry},
		{Text: "Remote Host", Widget: remoteHostEntry},
		{Text: "Remote Port", Widget: remotePortEntry},
		{Text: "Domain", Widget: domainEntry},
		{Text: "Local Port", Widget: localPortEntry},
	}, func(b bool) {
		if !b {
			return
		}
		sshPort, _ := strconv.Atoi(sshPortEntry.Text)
		remotePort, _ := strconv.Atoi(remotePortEntry.Text)
		localPort, _ := strconv.Atoi(localPortEntry.Text)
		onSave(Tunnel{
			Name:       nameEntry.Text,
			SSHServer:  sshServerEntry.Text,
			SSHPort:    sshPort,
			SSHUser:    sshUserEntry.Text,
			SSHKeyPath: sshKeyPathEntry.Text,
			RemoteHost: remoteHostEntry.Text,
			RemotePort: remotePort,
			Domain:     domainEntry.Text,
			LocalPort:  localPort,
		})
	}, w)
	d.Resize(fyne.NewSize(600, d.MinSize().Height))
	d.Show()
}

func main() {
	logFile, err := os.OpenFile("lighthouse.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err == nil {
		log.SetOutput(io.MultiWriter(os.Stderr, logFile))
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		defer logFile.Close()
	} else {
		log.Printf("open log file: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v", r)
		}
		_ = DeactivateProfile()
	}()

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

		var tunnelList *widget.List
		tunnelList = widget.NewList(
			func() int {
				if selected >= 0 && selected < len(profiles) {
					return len(profiles[selected].Tunnels)
				}
				return 0
			},
			func() fyne.CanvasObject {
				name := widget.NewLabel("")
				edit := widget.NewButton("Edit", nil)
				del := widget.NewButton("Delete", nil)
				start := widget.NewButton("Start", nil)
				return container.NewHBox(name, edit, del, start)
			},
			func(i widget.ListItemID, obj fyne.CanvasObject) {
				if selected < 0 || selected >= len(profiles) {
					return
				}
				t := profiles[selected].Tunnels[i]
				c := obj.(*fyne.Container)
				name := c.Objects[0].(*widget.Label)
				edit := c.Objects[1].(*widget.Button)
				del := c.Objects[2].(*widget.Button)
				start := c.Objects[3].(*widget.Button)

				name.SetText(t.Name)

				edit.OnTapped = func() {
					existing := t
					showTunnelDialog(w, "Edit Tunnel", &existing, func(nt Tunnel) {
						profiles[selected].Tunnels[i] = nt
						if err := SaveProfile(profiles[selected]); err != nil {
							log.Println("failed to save profile:", err)
						}
						tunnelList.Refresh()
					})
				}

				del.OnTapped = func() {
					profiles[selected].Tunnels = append(profiles[selected].Tunnels[:i], profiles[selected].Tunnels[i+1:]...)
					if err := SaveProfile(profiles[selected]); err != nil {
						log.Println("failed to save profile:", err)
					}
					tunnelList.Refresh()
				}

				if IsTunnelRunning(t) {
					start.SetText("Stop")
					start.OnTapped = func() {
						_ = StopTunnel(profiles[selected], t)
						tunnelList.Refresh()
					}
				} else {
					start.SetText("Start")
					start.OnTapped = func() {
						if err := StartTunnel(profiles[selected], t); err != nil {
							log.Println("start tunnel:", err)
						}
						tunnelList.Refresh()
					}
				}

				if selected >= 0 && IsProfileActive(profiles[selected]) {
					edit.Disable()
					del.Disable()
				} else {
					edit.Enable()
					del.Enable()
				}
			},
		)

		addTunnelButton := widget.NewButton("Add Tunnel", func() {
			if selected < 0 || selected >= len(profiles) {
				return
			}
			showTunnelDialog(w, "New Tunnel", nil, func(t Tunnel) {
				if err := profiles[selected].AddTunnel(t); err != nil {
					log.Println("failed to add tunnel:", err)
				}
				tunnelList.Refresh()
			})
		})

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
				tunnelList.Refresh()
			}
		})

		activateButton := widget.NewButton("Activate", nil)

		var updateActivate func()
		var updateButtons func()

		updateActivate = func() {
			if selected >= 0 && selected < len(profiles) && IsProfileActive(profiles[selected]) {
				activateButton.SetText("Deactivate")
			} else {
				activateButton.SetText("Activate")
			}
		}

		updateButtons = func() {
			if selected >= 0 && selected < len(profiles) && IsProfileActive(profiles[selected]) {
				renameButton.Disable()
				deleteButton.Disable()
				addTunnelButton.Disable()
			} else {
				renameButton.Enable()
				deleteButton.Enable()
				addTunnelButton.Enable()
			}
		}

		list.OnSelected = func(id widget.ListItemID) {
			selected = int(id)
			updateActivate()
			updateButtons()
			tunnelList.Refresh()
		}

		activateButton.OnTapped = func() {
			if selected >= 0 && selected < len(profiles) {
				if IsProfileActive(profiles[selected]) {
					if err := DeactivateProfile(); err != nil {
						log.Println("deactivate profile:", err)
					}
				} else {
					if err := ActivateProfile(profiles[selected]); err != nil {
						log.Println("activate profile:", err)
					}
				}
				updateActivate()
				updateButtons()
				tunnelList.Refresh()
			}
		}

		w.SetContent(container.NewVBox(
			widget.NewLabel("Profiles"),
			list,
			container.NewHBox(createButton, renameButton, deleteButton, activateButton),
			widget.NewLabel("Tunnels"),
			tunnelList,
			addTunnelButton,
		))
	}
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
