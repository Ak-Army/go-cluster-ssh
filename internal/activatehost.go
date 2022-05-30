package internal

import (
	"log"
	"math"

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
	mainWindow.Add(mainBox)

	// determine optimal table dimensions
	tLen := float64(ts.Len())
	cols := math.Sqrt(tLen)
	//Rows := math.Ceil(tLen / Cols)

	hostsConfFrame, _ := gtk.FrameNew("Active t")
	hostsConfTable, _ := gtk.GridNew()
	hostsConfTable.SetProperty("border_width", 5)
	hostsConfTable.SetProperty("row_spacing", 5)
	hostsConfTable.SetProperty("column_spacing", 5)
	hostsConfFrame.Add(hostsConfTable)

	i := float64(1)
	ts.Each(func(t *Terminal) {
		host := t.Host
		hostTable, _ := gtk.GridNew()
		hostTable.SetProperty("column_spacing", 2)
		label, _ := gtk.LabelNew(host)
		hostTable.Attach(label, 1, 1, 1, 1)
		hostCheckbox, _ := gtk.CheckButtonNew()
		hostCheckbox.SetActive(t.CopyInput)
		hostCheckbox.Connect("toggled", func(button *gtk.CheckButton) {
			xlog.Debugf("set %s host to active %t", t, button.GetActive())
			ts.Activate(host, button.GetActive())
		})
		hostTable.Attach(hostCheckbox, 2, 1, 1, 1)
		col := int(i / cols)
		row := int(i) % int(cols)
		hostsConfTable.Attach(hostTable, col+1, row, 1, 1)
		i++
	})

	mainBox.PackStart(hostsConfFrame, true, true, 0)

	okButton, _ := gtk.ButtonNewWithLabel("Ok")
	mainBox.PackStart(okButton, false, false, 0)

	// wire up behaviour
	okButton.Connect("clicked", func(_ *gtk.Button) { mainWindow.Destroy() })
	mainWindow.ShowAll()
}
