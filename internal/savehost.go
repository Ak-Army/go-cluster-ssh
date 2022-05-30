package internal

import (
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/Ak-Army/xlog"
	"github.com/gotk3/gotk3/gtk"
)

func SaveHostsDialog(ts *AllTerminal, w *gtk.Window) {
	entryBox, _ := gtk.FileChooserDialogNewWith2Buttons("Save host to:", w, gtk.FILE_CHOOSER_ACTION_SAVE, "Save", gtk.RESPONSE_OK, "Cancel", gtk.RESPONSE_CANCEL)
	entryBox.SetCreateFolders(true)
	entryBox.SetProperty("has_focus", true)

	resp := entryBox.Run()
	if resp == gtk.RESPONSE_CANCEL {
		entryBox.Destroy()
		return
	}
	xlog.Debug("Save to: ", entryBox.GetFilename())
	err := ioutil.WriteFile(entryBox.GetFilename(), []byte(strings.Join(ts.Names(), " ")), fs.ModePerm)
	if err != nil {
		xlog.Error("Unable to save hosts to file", err)
	}
	entryBox.Destroy()
}
