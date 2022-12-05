package internal

import (
	"github.com/gotk3/gotk3/gtk"
)

type Config struct {
	StartMaximized bool
	Font           string
	MinWidth       int
	MinHeight      int
}

type element struct {
	StartMaximized *gtk.CheckButton
	Font           *gtk.FontButton
	MinWidth       *gtk.SpinButton
	MinHeight      *gtk.SpinButton
}

type ConfigDialog struct {
	mainWindow *gtk.Window
	saveFunc   func(Config)
	config     Config
	element    *element
}

var configWindow *ConfigDialog

func NewConfigDialog(b *gtk.Builder, config Config, saveFunc func(Config)) {
	if configWindow != nil {
		configWindow.mainWindow.ShowAll()
		return
	}
	sm, _ := b.GetObject("windowConfig.Maximized")
	f, _ := b.GetObject("windowConfig.Font")
	mw, _ := b.GetObject("windowConfig.MinWidth")
	mh, _ := b.GetObject("windowConfig.MinHeight")
	w, _ := b.GetObject("windowConfig")
	configWindow = &ConfigDialog{
		saveFunc: saveFunc,
		config:   config,
		element: &element{
			StartMaximized: sm.(*gtk.CheckButton),
			Font:           f.(*gtk.FontButton),
			MinWidth:       mw.(*gtk.SpinButton),
			MinHeight:      mh.(*gtk.SpinButton),
		},
		mainWindow: w.(*gtk.Window),
	}
	configWindow.signals(b)

	configWindow.mainWindow.ShowAll()
}

func (c *ConfigDialog) signals(b *gtk.Builder) {
	c.mainWindow.Connect("show", c.show)
	close, _ := b.GetObject("windowConfig.Close")
	s, _ := b.GetObject("windowConfig.Save")
	close.(*gtk.Button).Connect("clicked", c.mainWindow.Hide)
	s.(*gtk.Button).Connect("clicked", c.save)
}

func (c *ConfigDialog) show(_ interface{}) {
	c.element.StartMaximized.SetActive(c.config.StartMaximized)

	c.element.Font.SetFont(c.config.Font)
	c.element.MinWidth.SetValue(float64(c.config.MinWidth))
	c.element.MinHeight.SetValue(float64(c.config.MinHeight))
}

func (c *ConfigDialog) save(_ interface{}) {
	c.config.StartMaximized = c.element.StartMaximized.Activate()

	c.config.Font = c.element.Font.GetFont()
	c.config.MinWidth = int(c.element.MinWidth.GetValue())
	c.config.MinHeight = int(c.element.MinHeight.GetValue())

	c.mainWindow.Hide()
	if c.saveFunc != nil {
		c.saveFunc(c.config)
	}
}
