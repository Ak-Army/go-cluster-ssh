package internal

import (
	"fmt"
	"unsafe"

	"github.com/electricface/go-gir/gdk-3.0"
	"github.com/electricface/go-gir/gi"
	"github.com/electricface/go-gir/gtk-3.0"
)

type addHostWindow struct {
	mainWindow gtk.Window
	saveFunc   func(string)
	entry      string
	entryBox   gtk.Entry
}

var addHost *addHostWindow

func AddHostDialog(b gtk.Builder, saveFunc func(string)) {
	if addHost != nil {
		addHost.mainWindow.ShowAll()
		return
	}

	addHost = &addHostWindow{
		saveFunc:   saveFunc,
		mainWindow: gtk.WrapWindow(b.GetObject("windowAddHost").P),
		entryBox:   gtk.WrapEntry(b.GetObject("windowAddHost.Entry").P),
	}
	addHost.entryBox.Connect("key-press-event", addHost.entryKeyPress)
	gtk.WrapButton(b.GetObject("windowAddHost.Close").P).Connect("clicked", addHost.close)
	gtk.WrapButton(b.GetObject("windowAddHost.Add").P).Connect("clicked", addHost.add)

	addHost.mainWindow.ShowAll()
}

func (c *addHostWindow) add() {
	if c.entry != "" {
		c.saveFunc(c.entry)
	}
}

func (c *addHostWindow) close() {
	c.entryBox.DeleteText(0, -1)
	c.entryBox.GrabFocus()
	c.entry = ""
	c.mainWindow.Hide()
}

func (c *addHostWindow) entryKeyPress(p gi.ParamBox) {
	ev := gdk.Event{}
	ev.P = p.Params[1].(unsafe.Pointer)
	_, code := ev.GetKeycode()
	if ev.GetEventType() == gdk.EventTypeKeyPress && code == 36 {
		c.add()
	} else {
		t := c.entryBox.GetText()
		_, kv := ev.GetKeyval()
		c.entry = fmt.Sprintf("%s%c", t, kv)
	}
}
