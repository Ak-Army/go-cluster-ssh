package internal

import (
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/Ak-Army/xlog"
	"github.com/gotk3/gotk3/gtk"
)

func SaveHostsDialog(b *gtk.Builder, ts *AllTerminal) {
	w, _ := b.GetObject("windowSaveHost")
	windowSaveHost := w.(*gtk.FileChooserDialog)

	resp := windowSaveHost.Run()
	if resp == gtk.RESPONSE_CANCEL {
		windowSaveHost.Hide()
		return
	}
	xlog.Debug("Save to: ", windowSaveHost.GetFilename())
	err := ioutil.WriteFile(windowSaveHost.GetFilename(), []byte(strings.Join(ts.Names(), " ")), fs.ModePerm)
	if err != nil {
		xlog.Error("Unable to save hosts to file", err)
	}
	windowSaveHost.Hide()
}
