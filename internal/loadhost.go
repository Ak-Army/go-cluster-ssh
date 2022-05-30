package internal

import (
	"bufio"
	"os"

	"github.com/Ak-Army/xlog"
	"github.com/gotk3/gotk3/gtk"
)

func LoadHostsDialog(ts *AllTerminal, w *gtk.Window) {
	entryBox, _ := gtk.FileChooserDialogNewWith2Buttons("Save host to:", w, gtk.FILE_CHOOSER_ACTION_OPEN, "Load", gtk.RESPONSE_OK, "Cancel", gtk.RESPONSE_CANCEL)
	entryBox.SetProperty("has_focus", true)

	defer entryBox.Destroy()
	resp := entryBox.Run()
	if resp == gtk.RESPONSE_CANCEL {
		entryBox.Destroy()
		return
	}
	xlog.Debug("Load from: ", entryBox.GetFilename())
	// hostnames are assumed to be whitespace separated
	file, err := os.Open(entryBox.GetFilename())
	if err != nil {
		xlog.Error("Unable to load hosts from file", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		ts.AddHost(entryBox.GetFilename(), scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		xlog.Error("Error occured when hosts load from file", err)
	}
}
