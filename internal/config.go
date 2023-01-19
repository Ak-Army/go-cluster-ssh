package internal

import (
	"context"
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
	"github.com/gotk3/gotk3/gtk"
	"github.com/juju/errors"
)

type Config struct {
	StartMaximized bool   `config:"startMaximized" json:"startMaximized"`
	Font           string `config:"font" json:"font"`
	MinWidth       int    `config:"minWidth" json:"minWidth"`
	MinHeight      int    `config:"minHeight" json:"minHeight"`
}

type ConfigStore struct {
	mu         sync.RWMutex
	config     *Config
	err        error
	configFile string
	encoder    encoder.Encoder
}

type element struct {
	StartMaximized *gtk.CheckButton
	Font           *gtk.FontButton
	MinWidth       *gtk.SpinButton
	MinHeight      *gtk.SpinButton
}

type configDialog struct {
	mainWindow *gtk.Window
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

	return c, err
}

func NewConfigDialog(b *gtk.Builder, c *ConfigStore, saveFunc func()) {
	if configWindow != nil {
		configWindow.mainWindow.ShowAll()
		return
	}
	sm, _ := b.GetObject("windowConfig.Maximized")
	f, _ := b.GetObject("windowConfig.Font")
	mw, _ := b.GetObject("windowConfig.MinWidth")
	mh, _ := b.GetObject("windowConfig.MinHeight")
	w, _ := b.GetObject("windowConfig")
	configWindow = &configDialog{
		saveFunc: saveFunc,
		config:   c,
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

func (c *configDialog) signals(b *gtk.Builder) {
	c.mainWindow.Connect("show", c.show)
	wClose, _ := b.GetObject("windowConfig.Close")
	s, _ := b.GetObject("windowConfig.Save")
	wClose.(*gtk.Button).Connect("clicked", c.mainWindow.Hide)
	s.(*gtk.Button).Connect("clicked", c.save)
}

func (c *configDialog) show(_ interface{}) {
	c.element.StartMaximized.SetActive(c.config.Config().StartMaximized)

	c.element.Font.SetFont(c.config.Config().Font)
	c.element.MinWidth.SetValue(float64(c.config.Config().MinWidth))
	c.element.MinHeight.SetValue(float64(c.config.Config().MinHeight))
}

func (c *configDialog) save(_ interface{}) {
	conf := c.config.NewSnapshot().(*Config)
	conf.StartMaximized = c.element.StartMaximized.Activate()

	conf.Font = c.element.Font.GetFont()
	conf.MinWidth = int(c.element.MinWidth.GetValue())
	conf.MinHeight = int(c.element.MinHeight.GetValue())
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
