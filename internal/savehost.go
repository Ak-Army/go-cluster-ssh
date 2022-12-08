package internal

import (
	"io/fs"
	"os"
	"strings"

	"github.com/Ak-Army/xlog"
	"github.com/electricface/go-gir/gtk-3.0"
)

func SaveHostsDialog(b gtk.Builder, ts *AllTerminal) {
	windowSaveHost := gtk.WrapFileChooserDialog(b.GetObject("windowSaveHost").P)

	resp := windowSaveHost.Run()
	if resp == int32(gtk.ResponseTypeCancel) {
		windowSaveHost.Hide()
		return
	}
	xlog.Debug("Save to: ", windowSaveHost.GetFilename())
	err := os.WriteFile(windowSaveHost.GetFilename(), []byte(strings.Join(ts.Names(), " ")), fs.ModePerm)
	if err != nil {
		xlog.Error("Unable to save hosts to file", err)
	}
	windowSaveHost.Hide()
}
