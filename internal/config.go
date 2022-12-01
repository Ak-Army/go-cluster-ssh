package internal

import (
	"github.com/Ak-Army/xlog"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
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
	configWindow = &ConfigDialog{
		saveFunc: saveFunc,
		config:   config,
		element: &element{
			StartMaximized: b.GetObject("windowConfigMaximized").Cast().(*gtk.CheckButton),
			Font:           b.GetObject("windowConfigFont").Cast().(*gtk.FontButton),
			MinWidth:       b.GetObject("windowConfigMinWidth").Cast().(*gtk.SpinButton),
			MinHeight:      b.GetObject("windowConfigMinHeight").Cast().(*gtk.SpinButton),
		},
		mainWindow: b.GetObject("windowConfig").Cast().(*gtk.Window),
	}
	configWindow.signals(b)

	configWindow.mainWindow.ShowAll()
}

func (c *ConfigDialog) signals(b *gtk.Builder) {
	xlog.Info("add signals")
	configWindow.mainWindow.Connect("windowConfig.Show", c.show)
	configWindow.mainWindow.Connect("windowConfig.Hide", c.mainWindow.Hide)
	configWindow.mainWindow.Connect("windowConfig.Save", c.save)
}

func (c *ConfigDialog) show(_ interface{}) {
	c.element.StartMaximized.SetActive(c.config.StartMaximized)

	c.element.Font.SetFont(c.config.Font)
	c.element.MinWidth.SetValue(float64(c.config.MinWidth))
	c.element.MinHeight.SetValue(float64(c.config.MinHeight))
}

func (c *ConfigDialog) save(_ interface{}) {
	c.config.StartMaximized = c.element.StartMaximized.Activate()

	c.config.Font = c.element.Font.Font()
	c.config.MinWidth = int(c.element.MinWidth.Value())
	c.config.MinHeight = int(c.element.MinHeight.Value())

	c.mainWindow.Hide()
	if c.saveFunc != nil {
		c.saveFunc(c.config)
	}
}
