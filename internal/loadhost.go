package internal

import (
	"bufio"
	"os"

	"github.com/Ak-Army/xlog"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

func LoadHostsDialog(b *gtk.Builder, ts *AllTerminal) {
	windowLoadHost := b.GetObject("windowLoadHost").Cast().(*gtk.FileChooserDialog)

	defer windowLoadHost.Hide()
	resp := windowLoadHost.Run()
	if resp == int(gtk.ResponseCancel) {
		return
	}
	xlog.Debug("Load from: ", windowLoadHost.Filename())
	// hostnames are assumed to be whitespace separated
	file, err := os.Open(windowLoadHost.Filename())
	if err != nil {
		xlog.Error("Unable to load hosts from file", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		ts.AddHost(windowLoadHost.Filename(), scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		xlog.Error("Error occured when hosts load from file", err)
	}
}
