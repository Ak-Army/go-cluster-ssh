package internal

import (
	"github.com/Ak-Army/xlog"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type newHostDialog struct {
	mainWindow *gtk.Window
	saveFunc   func(string)
	entryBox   *gtk.Entry
}

func NewHostDialog(saveFunc func(string)) {
	xlog.Debug("Add new host")
	c := &newHostDialog{
		saveFunc: saveFunc,
	}
	var err error
	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	c.mainWindow, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		xlog.Fatal("Unable to create window:", err)
	}
	c.mainWindow.SetModal(true)

	mainBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	if err != nil {
		xlog.Fatal("Unable to create main box:", err)
	}
	err = mainBox.SetProperty("border_width", 5)
	if err != nil {
		xlog.Fatal("Unable to set property:", err)
	}
	c.mainWindow.Add(mainBox)

	label, _ := gtk.LabelNew("Add new hostname:")
	mainBox.PackStart(label, false, false, 0)

	c.entryBox, _ = gtk.EntryNew()
	c.entryBox.SetProperty("has_focus", true)
	c.entryBox.Connect("key_press_event", func(o interface{}, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{Event: ev}
		if keyEvent.Type() == gdk.EVENT_KEY_PRESS &&
			keyEvent.HardwareKeyCode() == 36 {
			c.saveHook(o)
		}
	})
	mainBox.PackStart(c.entryBox, false, false, 0)

	cancelButton, _ := gtk.ButtonNewWithLabel("Close")
	mainBox.PackStart(cancelButton, false, false, 0)
	saveButton, _ := gtk.ButtonNewWithLabel("Add")
	mainBox.PackStart(saveButton, false, false, 0)

	cancelButton.Connect("clicked", func(_ *gtk.Button) { c.mainWindow.Destroy() })
	saveButton.Connect("clicked", c.saveHook)
	// Recursively show all widgets contained in this window.
	c.mainWindow.ShowAll()
}

func (e *newHostDialog) saveHook(_ interface{}) {
	b, _ := e.entryBox.GetBuffer()
	t, _ := b.GetText()
	if t != "" {
		e.saveFunc(t)
	}
}
