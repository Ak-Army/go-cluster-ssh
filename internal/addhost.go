package internal

import (
	"github.com/Ak-Army/xlog"
	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
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
	addHost = &addHostWindow{
		saveFunc:   saveFunc,
		mainWindow: b.GetObject("windowAddHost").Cast().(*gtk.Window),
		entryBox:   b.GetObject("windowAddHostEntry").Cast().(*gtk.Entry),
	}
	addHost.signals(b)

	addHost.mainWindow.ShowAll()
}

func (c *addHostWindow) signals(b *gtk.Builder) {
	xlog.Info("signals")
	c.entryBox.Connect("windowAddHost.Entry.KeyPress", c.entryKeyPress)
	c.mainWindow.Connect("windowAddHost.Close", c.close)
	c.mainWindow.Connect("windowAddHost.Add", c.add)
}

func (c *addHostWindow) add(_ interface{}) {
	xlog.Info("save")
	if c.entry != "" {
		c.saveFunc(c.entry)
	}
}

func (c *addHostWindow) close(_ interface{}) {
	xlog.Info("hide")
	c.entryBox.DeleteText(0, -1)
	c.entryBox.SetObjectProperty("has_focus", true)
	c.entry = ""
	c.mainWindow.Hide()
}

func (c *addHostWindow) entryKeyPress(e *gtk.Entry, ev *gdk.Event) {
	xlog.Info("entryKeyPress")
	keyEvent := ev.AsKey()
	if keyEvent.Type() == gdk.KeyPressType &&
		keyEvent.HardwareKeycode() == 36 {
		c.add(e)
	} else {
		b := e.Buffer()
		c.entry = b.Text()
	}
}
