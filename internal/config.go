package internal

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/Ak-Army/config"
	"github.com/Ak-Army/config/backend"
	"github.com/Ak-Army/config/backend/file"
	"github.com/Ak-Army/config/encoder"
	cyaml "github.com/Ak-Army/config/encoder/yaml"
	"github.com/Ak-Army/xlog"
	"github.com/electricface/go-gir/gtk-3.0"
	"github.com/juju/errors"
)

type Config struct {
	StartMaximized bool   `config:"startmaximized"`
	Font           string `config:"font"`
	MinWidth       int32  `config:"minwidth"`
	MinHeight      int32  `config:"minheight"`
}

type ConfigStore struct {
	mu         sync.RWMutex
	config     *Config
	err        error
	configFile string
	encoder    encoder.Encoder
}

type element struct {
	StartMaximized gtk.CheckButton
	Font           gtk.FontButton
	MinWidth       gtk.SpinButton
	MinHeight      gtk.SpinButton
}

type configDialog struct {
	mainWindow gtk.Window
	saveFunc   func()
	config     *ConfigStore
	element    *element
}

var configWindow *configDialog

func NewConfig() (*ConfigStore, error) {
	c := &ConfigStore{
		encoder: cyaml.New(),
	}
	c.config = c.NewSnapshot().(*Config)
	home, err := os.UserHomeDir()
	if err != nil {
		return c, errors.Annotate(err, "Unable to determine home dir")
	}
	c.configFile = filepath.Join(home, ".config", "go-cluster-ssh.yml")
	loader, err := config.NewLoader(context.Background(),
		file.New(file.WithPath(c.configFile), file.WithOption(backend.WithEncoder(c.encoder))),
	)
	if err != nil {
		log.Fatal(err)
	}
	err = loader.Load(c)

	fmt.Println(c.Config())
	return c, err
}

func NewConfigDialog(b gtk.Builder, c *ConfigStore, saveFunc func()) {
	if configWindow != nil {
		configWindow.mainWindow.ShowAll()
		return
	}
	configWindow = &configDialog{
		saveFunc: saveFunc,
		config:   c,
		element: &element{
			StartMaximized: gtk.WrapCheckButton(b.GetObject("windowConfig.Maximized").P),
			Font:           gtk.WrapFontButton(b.GetObject("windowConfig.Font").P),
			MinWidth:       gtk.WrapSpinButton(b.GetObject("windowConfig.MinWidth").P),
			MinHeight:      gtk.WrapSpinButton(b.GetObject("windowConfig.MinHeight").P),
		},
		mainWindow: gtk.WrapWindow(b.GetObject("windowConfig").P),
	}
	configWindow.signals(b)

	configWindow.mainWindow.ShowAll()
}

func (c *configDialog) signals(b gtk.Builder) {
	c.mainWindow.Connect("show", c.show)
	gtk.WrapButton(b.GetObject("windowConfig.Close").P).Connect("clicked", c.mainWindow.Hide)
	gtk.WrapButton(b.GetObject("windowConfig.Save").P).Connect("clicked", c.save)
}

func (c *configDialog) show() {
	c.element.StartMaximized.SetActive(c.config.Config().StartMaximized)

	c.element.Font.SetFont(c.config.Config().Font)
	c.element.MinWidth.SetValue(float64(c.config.Config().MinWidth))
	c.element.MinHeight.SetValue(float64(c.config.Config().MinHeight))
}

func (c *configDialog) save() {
	conf := c.config.NewSnapshot().(*Config)
	conf.StartMaximized = c.element.StartMaximized.Activate()

	conf.Font = c.element.Font.GetFont()
	conf.MinWidth = int32(c.element.MinWidth.GetValue())
	conf.MinHeight = int32(c.element.MinHeight.GetValue())
	c.config.SetSnapshot(conf, nil)
	if c.config.configFile != "" {
		b, err := c.config.encoder.Encode(conf)
		if err != nil {
			xlog.Error("Unable to marshal config", err)
			return
		}

		err = os.WriteFile(c.config.configFile, b, fs.ModePerm)
		if err != nil {
			xlog.Error("Unable to save config", err)
			return
		}
	}
	c.mainWindow.Hide()
	if c.saveFunc != nil {
		c.saveFunc()
	}
}

func (c *ConfigStore) NewSnapshot() interface{} {
	return &Config{
		StartMaximized: true,
		Font:           "Ubuntu Mono,monospace Bold 10",
		MinWidth:       250,
		MinHeight:      250,
	}
}

func (c *ConfigStore) SetSnapshot(i interface{}, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	conf := i.(*Config)
	c.config = conf
	c.err = err
}

func (c *ConfigStore) ConfigWithError() (*Config, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config, c.err
}

func (c *ConfigStore) Config() *Config {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config
}
