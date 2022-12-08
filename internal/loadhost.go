package internal

import (
	"bufio"
	"os"

	"github.com/Ak-Army/xlog"
	"github.com/electricface/go-gir/gtk-3.0"
)

func LoadHostsDialog(b gtk.Builder, ts *AllTerminal) {
	windowLoadHost := gtk.WrapFileChooserDialog(b.GetObject("windowSaveHost").P)

	defer windowLoadHost.Hide()
	resp := windowLoadHost.Run()
	if resp == int32(gtk.ResponseTypeCancel) {
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
