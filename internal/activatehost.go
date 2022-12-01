package internal

import (
	"github.com/Ak-Army/xlog"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

func ActiveHostsDialog(ts *AllTerminal) {
	mainWindow := gtk.NewWindow(gtk.WindowToplevel)
	mainWindow.SetModal(true)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 5)
	mainBox.SetObjectProperty("border_width", 5)

	for _, term := range ts.terminals {
		hostsConfFrame := gtk.NewAspectFrame(term.Name, 0, 0, 1, true)
		hostsConfTable := gtk.NewGrid()
		hostsConfTable.SetObjectProperty("border_width", 5)
		hostsConfTable.SetObjectProperty("row_spacing", 5)
		hostsConfTable.SetObjectProperty("column_spacing", 5)
		col := 1
		row := 1
		for _, t := range term.terminals {
			host := t.Host
			hostTable := gtk.NewGrid()
			hostTable.SetObjectProperty("column_spacing", 2)
			label := gtk.NewLabel(host)
			hostTable.Attach(label, 1, 1, 1, 1)
			hostCheckbox := gtk.NewCheckButton()
			hostCheckbox.SetActive(t.CopyInput)
			hostCheckbox.Connect("toggled", func(button *gtk.CheckButton) {
				xlog.Debugf("set %s host to active %t", t, button.Activate())
				ts.Activate(host, button.Activate())
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

	okButton := gtk.NewButtonWithLabel("Ok")
	mainBox.PackStart(okButton, false, false, 0)

	// wire up behaviour
	okButton.Connect("clicked", func(_ *gtk.Button) { mainWindow.Destroy() })
	mainWindow.Add(mainBox)
	mainWindow.ShowAll()
}
