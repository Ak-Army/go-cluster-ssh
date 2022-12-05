package internal

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type addHostWindow struct {
	mainWindow *gtk.Window
	saveFunc   func(string)
	entry      string
	entryBox   *gtk.Entry
}

var addHost *addHostWindow

func AddHostDialog(b *gtk.Builder, saveFunc func(string)) {
	if addHost != nil {
		addHost.mainWindow.ShowAll()
		return
	}
	mv, _ := b.GetObject("windowAddHost")
	eb, _ := b.GetObject("windowAddHost.Entry")
	addHost = &addHostWindow{
		saveFunc:   saveFunc,
		mainWindow: mv.(*gtk.Window),
		entryBox:   eb.(*gtk.Entry),
	}
	addHost.entryBox.Connect("key-press-event", addHost.entryKeyPress)
	c, _ := b.GetObject("windowAddHost.Close")
	a, _ := b.GetObject("windowAddHost.Add")
	c.(*gtk.Button).Connect("clicked", addHost.close)
	a.(*gtk.Button).Connect("clicked", addHost.add)

	addHost.mainWindow.ShowAll()
}

func (c *addHostWindow) add(_ interface{}) {
	if c.entry != "" {
		c.saveFunc(c.entry)
	}
}

func (c *addHostWindow) close(_ interface{}) {
	c.entryBox.DeleteText(0, -1)
	c.entryBox.SetProperty("has_focus", true)
	c.entry = ""
	c.mainWindow.Hide()
}

func (c *addHostWindow) entryKeyPress(e *gtk.Entry, ev *gdk.Event) {
	keyEvent := &gdk.EventKey{Event: ev}
	if keyEvent.Type() == gdk.EVENT_KEY_PRESS &&
		keyEvent.HardwareKeyCode() == 36 {
		c.add(e)
	} else {
		t, _ := e.GetText()
		c.entry = fmt.Sprintf("%s%c", t, keyEvent.KeyVal())
	}
}
