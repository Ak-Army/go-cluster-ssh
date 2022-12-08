package internal

import (
	"github.com/Ak-Army/xlog"
	"github.com/electricface/go-gir/g-2.0"
	"github.com/electricface/go-gir/gi"
	"github.com/electricface/go-gir/gtk-3.0"
)

func ActiveHostsDialog(ts *AllTerminal) {
	mainWindow := gtk.NewWindow(gtk.WindowTypeToplevel)
	mainWindow.SetModal(true)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 5)
	mainBox.SetBorderWidth(5)

	for _, term := range ts.terminals {
		hostsConfFrame := gtk.NewAspectFrame(term.Name, 0, 0, 1, true)
		hostsConfTable := gtk.NewGrid()
		hostsConfTable.SetBorderWidth(5)
		hostsConfTable.SetRowSpacing(5)
		hostsConfTable.SetColumnSpacing(5)
		col := int32(1)
		row := int32(1)
		for _, t := range term.terminals {
			host := t.Host
			hostTable := gtk.NewGrid()
			hostTable.SetColumnSpacing(2)
			label := gtk.NewLabel(host)
			hostTable.Attach(label, 1, 1, 1, 1)
			hostCheckbox := gtk.NewCheckButton()
			hostCheckbox.SetActive(t.CopyInput)
			hostCheckbox.Connect("toggled", func(p gi.ParamBox) {
				button := gtk.WrapCheckButton(p.Params[0].(g.Object).P)
				xlog.Debugf("set %s host to active %t", t, button.GetActive())
				ts.Activate(host, button.GetActive())
			})
			hostTable.Attach(hostCheckbox, 2, 1, 1, 1)
			hostsConfTable.Attach(hostTable, col, row, 1, 1)
			if col == term.Cols {
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
	okButton.Connect("clicked", mainWindow.Destroy)
	mainWindow.Add(mainBox)
	mainWindow.ShowAll()
}
