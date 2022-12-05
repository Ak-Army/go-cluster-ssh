package internal

import (
	"bufio"
	"os"

	"github.com/Ak-Army/xlog"
	"github.com/gotk3/gotk3/gtk"
)

func LoadHostsDialog(b *gtk.Builder, ts *AllTerminal) {
	w, _ := b.GetObject("windowLoadHost")
	windowLoadHost := w.(*gtk.FileChooserDialog)

	defer windowLoadHost.Hide()
	resp := windowLoadHost.Run()
	if resp == gtk.RESPONSE_CANCEL {
		return
	}
	xlog.Debug("Load from: ", windowLoadHost.GetFilename())
	// hostnames are assumed to be whitespace separated
	file, err := os.Open(windowLoadHost.GetFilename())
	if err != nil {
		xlog.Error("Unable to load hosts from file", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		ts.AddHost(windowLoadHost.GetFilename(), scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		xlog.Error("Error occured when hosts load from file", err)
	}
}
