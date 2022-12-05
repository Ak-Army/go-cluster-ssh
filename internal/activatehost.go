package internal

import (
	"log"

	"github.com/Ak-Army/xlog"
	"github.com/gotk3/gotk3/gtk"
)

func ActiveHostsDialog(ts *AllTerminal) {
	mainWindow, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	mainWindow.SetModal(true)

	mainBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	mainBox.SetProperty("border_width", 5)

	for _, term := range ts.terminals {
		hostsConfFrame, _ := gtk.AspectFrameNew(term.Name, 0, 0, 1, true)
		hostsConfTable, _ := gtk.GridNew()
		hostsConfTable.SetBorderWidth(5)
		hostsConfTable.SetRowSpacing(5)
		hostsConfTable.SetColumnSpacing(5)
		col := 1
		row := 1
		for _, t := range term.terminals {
			host := t.Host
			hostTable, _ := gtk.GridNew()
			hostTable.SetColumnSpacing(2)
			label, _ := gtk.LabelNew(host)
			hostTable.Attach(label, 1, 1, 1, 1)
			hostCheckbox, _ := gtk.CheckButtonNew()
			hostCheckbox.SetActive(t.CopyInput)
			hostCheckbox.Connect("toggled", func(button *gtk.CheckButton) {
				xlog.Debugf("set %s host to active %t", t, button.GetActive())
				ts.Activate(host, button.GetActive())
			})
			hostTable.Attach(hostCheckbox, 2, 1, 1, 1)
			hostsConfTable.Attach(hostTable, col, row, 1, 1)
			if col == int(term.Cols) {
				col = 0
				row++
			}
			col++
		}
		hostsConfFrame.Add(hostsConfTable)
		mainBox.PackStart(hostsConfFrame, true, false, 0)
	}

	okButton, _ := gtk.ButtonNewWithLabel("Ok")
	mainBox.PackStart(okButton, false, false, 0)

	// wire up behaviour
	okButton.Connect("clicked", func(_ *gtk.Button) { mainWindow.Destroy() })
	mainWindow.Add(mainBox)
	mainWindow.ShowAll()
}
