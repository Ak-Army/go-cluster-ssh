package internal

import (
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/Ak-Army/xlog"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

func SaveHostsDialog(b *gtk.Builder, ts *AllTerminal) {
	windowSaveHost := b.GetObject("windowSaveHost").Cast().(*gtk.FileChooserDialog)

	resp := windowSaveHost.Run()
	if resp == int(gtk.ResponseCancel) {
		windowSaveHost.Hide()
		return
	}
	xlog.Debug("Save to: ", windowSaveHost.Filename())
	err := ioutil.WriteFile(windowSaveHost.Filename(), []byte(strings.Join(ts.Names(), " ")), fs.ModePerm)
	if err != nil {
		xlog.Error("Unable to save hosts to file", err)
	}
	windowSaveHost.Hide()
}
