package internal

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
)

type Config struct {
	StartMaximized bool
	Font           string
	MinWidth       int
	MinHeight      int
}

type ConfigDialog struct {
	mainWindow *gtk.Window
	saveFunc   func(Config)
	config     Config
}

func NewConfigDialog(config Config, saveFunc func(Config)) {
	c := &ConfigDialog{
		saveFunc: saveFunc,
		config:   config,
	}
	var err error
	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	c.mainWindow, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	c.mainWindow.SetModal(true)

	mainBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	if err != nil {
		log.Fatal("Unable to create main box:", err)
	}
	err = mainBox.SetProperty("border_width", 5)
	if err != nil {
		log.Fatal("Unable to set property:", err)
	}
	c.mainWindow.Add(mainBox)
	globalConfFrame, err := gtk.FrameNew("Global Options")
	if err != nil {
		log.Fatal("Unable to create conf frame:", err)
	}

	globalConfTable, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	err = globalConfTable.SetProperty("border_width", 5)
	if err != nil {
		log.Fatal("Unable to set property:", err)
	}
	err = globalConfTable.SetProperty("row_spacing", 5)
	if err != nil {
		log.Fatal("Unable to set property:", err)
	}
	err = globalConfTable.SetProperty("column_spacing", 5)
	if err != nil {
		log.Fatal("Unable to set property:", err)
	}
	globalConfFrame.Add(globalConfTable)
	mainBox.PackStart(globalConfFrame, false, false, 0)

	label, _ := gtk.LabelNew("Start Maximized:")
	globalConfTable.Attach(label, 1, 2, 1, 2)
	maximizedConf, err := gtk.CheckButtonNew()
	maximizedConf.SetActive(config.StartMaximized)
	maximizedConf.Connect("toggled", c.maximizedHook)
	globalConfTable.Attach(maximizedConf, 2, 2, 1, 2)

	termConfFrame, _ := gtk.FrameNew("Terminal Options")
	termConfTable, _ := gtk.GridNew()
	termConfTable.SetProperty("border_width", 5)
	termConfTable.SetProperty("row_spacing", 5)
	termConfTable.SetProperty("column_spacing", 5)
	termConfFrame.Add(termConfTable)
	mainBox.PackStart(termConfFrame, false, false, 0)

	label2, _ := gtk.LabelNew("Font:")
	termConfTable.Attach(label2, 1, 1, 1, 2)
	fontConf, _ := gtk.FontButtonNewWithFont(config.Font)
	fontConf.Connect("font-set", c.fontHook)
	termConfTable.Attach(fontConf, 2, 1, 1, 2)

	sizeBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	termConfTable.Attach(sizeBox, 1, 3, 2, 2)
	label3, _ := gtk.LabelNew("Min Width:")
	sizeBox.PackStart(label3, false, false, 0)
	widthEntry, _ := gtk.SpinButtonNewWithRange(1, 9999, 1)
	widthEntry.SetValue(float64(c.config.MinWidth))
	widthEntry.Connect("value-changed", c.widthHook)
	sizeBox.PackStart(widthEntry, false, false, 0)

	label4, _ := gtk.LabelNew("Min Height:")
	sizeBox.PackStart(label4, false, false, 0)
	heightEntry, _ := gtk.SpinButtonNewWithRange(1, 9999, 1)
	heightEntry.SetValue(float64(c.config.MinHeight))
	heightEntry.Connect("value-changed", c.heightHook)
	sizeBox.PackStart(heightEntry, false, false, 0)

	confirmBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	cancelButton, _ := gtk.ButtonNewWithLabel("Cancel")
	confirmBox.PackStart(cancelButton, false, false, 0)
	saveButton, _ := gtk.ButtonNewWithLabel("Save")
	confirmBox.PackStart(saveButton, false, false, 0)
	mainBox.PackStart(confirmBox, false, false, 0)

	cancelButton.Connect("clicked", func(_ *gtk.Button) { c.mainWindow.Destroy() })
	saveButton.Connect("clicked", c.saveHook)
	// Recursively show all widgets contained in this window.
	c.mainWindow.ShowAll()
}

func (c *ConfigDialog) maximizedHook(button *gtk.CheckButton) {
	c.config.StartMaximized = button.GetActive()
}

func (c *ConfigDialog) fontHook(button *gtk.FontButton) {
	c.config.Font = button.GetFont()
}

func (c *ConfigDialog) widthHook(button *gtk.SpinButton) {
	c.config.MinWidth = int(button.GetValue())
}

func (c *ConfigDialog) heightHook(button *gtk.SpinButton) {
	c.config.MinHeight = int(button.GetValue())
}

func (c *ConfigDialog) saveHook(_ *gtk.Button) {
	c.mainWindow.Destroy()
	if c.saveFunc != nil {
		c.saveFunc(c.config)
	}
}
