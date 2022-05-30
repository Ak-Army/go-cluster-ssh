package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ak-Army/xlog"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"gopkg.in/yaml.v3"
)

type SSH struct {
	terminals *AllTerminal

	configFile    string
	config        Config
	scrollWin     *gtk.ScrolledWindow
	mainWin       *gtk.Window
	termMinWidth  int
	termMinHeight int
	entryBox      *gtk.Entry
}

func New(hosts []*HostGroup, sshCmd string, sshArgs []string) {
	gtk.Init(nil)
	s := &SSH{}
	home, err := os.UserHomeDir()
	if err != nil {
		xlog.Warn("Unable to determine home dir", err)
	} else {
		s.configFile = filepath.Join(home, ".config", "go-cluster-ssh.yml")
	}
	if sshCmd == "" {
		sshCmd = "/usr/bin/ssh"
	}
	s.config = Config{
		StartMaximized: true,
		Font:           "Ubuntu Mono,monospace Bold 10",
		MinWidth:       250,
		MinHeight:      250,
	}
	b, err := os.ReadFile(s.configFile)
	if err == nil {
		yaml.Unmarshal(b, &s.config)
	}
	s.termMinWidth = 1
	s.termMinHeight = 1

	// GUI Objects
	s.mainWin, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	s.scrollWin, _ = gtk.ScrolledWindowNew(nil, nil)
	s.scrollWin.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	s.scrollWin.SetShadowType(gtk.SHADOW_ETCHED_OUT)

	s.terminals = NewAllTerminals(sshCmd, sshArgs, hosts, func(t gtk.IWidget) {
		s.scrollWin.Remove(t)
		xlog.Debug("removed")
		s.scrollWin.Add(s.terminals.Layout())
		xlog.Debug("Added")
		s.reflow(true)
		s.scrollWin.QueueDraw()
	})
	s.configTerminals()
	s.initGUI()
	if s.config.StartMaximized {
		s.mainWin.Maximize()
	}
	s.reflow(true)
	s.mainWin.ShowAll()
	gtk.Main()
}

func (s *SSH) reflow(force bool) {
	// force redraw
	w, h := s.mainWin.GetSizeRequest()
	s.mainWin.SetSizeRequest(0, 0)
	time.Sleep(time.Millisecond)
	defer func() {
		s.mainWin.SetSizeRequest(w, h)
		s.mainWin.ShowAll()
	}()

	// reconfigure before updating Rows and columns
	s.configTerminals()
	numTerms := s.terminals.Len()
	if numTerms < 1 {
		gtk.MainQuit()
		return
	}
	width, _ := s.mainWin.GetSize()
	s.terminals.Reflow(width, force, s.config)

	title := "go-cluster-ssh:" + strings.Join(s.terminals.Names(), " ")

	s.mainWin.SetTitle(title)
}

func (s *SSH) configTerminals() {
	s.terminals.Each(func(t *Terminal) {
		t.SetScrollbackLines(-1)
		t.SetSizeRequest(s.config.MinWidth, s.config.MinHeight)
		t.SetFontFromString(s.config.Font)
		if s.termMinWidth < s.config.MinWidth {
			s.termMinWidth = s.config.MinWidth
		}
		s.termMinHeight = t.GetAllocatedHeight()
		if s.termMinHeight < s.config.MinHeight {
			s.termMinHeight = s.config.MinHeight
		}
	})
}

func (s *SSH) initGUI() {
	theme, _ := gtk.IconThemeGetDefault()
	if theme.HasIcon("Terminal") {
		icon, _ := theme.LoadIcon("Terminal", 128, gtk.ICON_LOOKUP_USE_BUILTIN)
		if icon != nil {
			s.mainWin.SetIcon(icon)
		}
	}
	s.mainWin.SetRole("go_cluster_ssh_main_win")
	s.mainWin.Connect("delete-event", func(_ *gtk.Window) {
		gtk.MainQuit()
	})
	mainBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	s.mainWin.Add(mainBox)

	mainMenuBar := s.initMainMenuBar()
	mainBox.PackStart(mainMenuBar, false, false, 0)

	mainBox.PackStart(s.scrollWin, true, true, 0)
	s.scrollWin.Add(s.terminals.Layout())
	//s.scrollWin.SetSizeRequest(s.termMinWidth, s.termMinHeight)

	s.entryBox = s.initEntryBox()
	mainBox.PackStart(s.entryBox, false, false, 0)

	// reflow layout on size change.
	s.mainWin.Connect("size-allocate", func(window *gtk.Window) {
		w, h := window.GetSize()
		xlog.Debugf("Size %dx%d", w, h)
		newWidth := w
		newHeight := h
		if w < s.config.MinWidth {
			newWidth = s.config.MinWidth
		}
		if h < s.config.MinHeight {
			newHeight = s.config.MinHeight
		}
		if newWidth != w || newHeight != h {
			window.SetSizeRequest(newWidth, newHeight)
		} else {
			s.reflow(false)
		}
	})
}

func (s *SSH) initEntryBox() *gtk.Entry {
	entryBox, _ := gtk.EntryNew()
	// don't display chars while typing.
	entryBox.SetVisible(false)
	entryBox.SetInvisibleChar(' ')
	// feed GNOME clipboard to all active terminals
	feedPaste := func(widget *gtk.Entry, ev *gdk.Event) {
		s.terminals.PasteClipboard()
		buffer, _ := entryBox.GetBuffer()
		buffer.DeleteText(0, -1)
	}
	// forward key events to all terminals with copy_input set
	feedInput := func(widget *gtk.Entry, ev *gdk.Event) bool {
		buffer, _ := entryBox.GetBuffer()
		buffer.DeleteText(0, -1)

		keyEvent := &gdk.EventKey{Event: ev}
		// check for paste key shortcut (ctl-shift-v)
		if keyEvent.Type() == gdk.EVENT_KEY_PRESS &&
			keyEvent.State()&uint(gdk.CONTROL_MASK) == uint(gdk.CONTROL_MASK) &&
			keyEvent.State()&uint(gdk.SHIFT_MASK) == uint(gdk.SHIFT_MASK) {
			feedPaste(widget, ev)
		} else {
			s.terminals.Event(ev)
		}
		// this stops regular handler from firing, switching focus.
		return gdk.GDK_EVENT_STOP
	}
	entryBox.Connect("key_press_event", feedInput)
	entryBox.Connect("key_release_event", feedInput)
	entryBox.Connect("paste_clipboard", feedPaste)
	entryBox.Connect("button_press_event", func(widget *gtk.Entry, ev *gdk.Event) {
		if gdk.EventButtonNewFromEvent(ev).Button() == gdk.BUTTON_MIDDLE {
			feedInput(widget, ev)
		}
	})
	// give EntryBox default focus on init
	entryBox.SetProperty("has_focus", true)

	return entryBox
}

func (s *SSH) initMainMenuBar() *gtk.MenuBar {
	mainMenuBar, _ := gtk.MenuBarNew()

	fileItem, _ := gtk.MenuItemNewWithLabel("File")
	fileMenu, _ := gtk.MenuNew()
	fileItem.SetSubmenu(fileMenu)

	addHostItem, _ := gtk.MenuItemNewWithLabel("Add host")
	addHostItem.Connect("activate", func(_ *gtk.MenuItem) {
		NewHostDialog(func(hostName string) {
			s.terminals.AddHost("Default", hostName)
			s.reflow(true)
		})
	})
	fileMenu.Append(addHostItem)

	saveHostItem, _ := gtk.MenuItemNewWithLabel("Save hosts")
	saveHostItem.Connect("activate", func(_ *gtk.MenuItem) {
		SaveHostsDialog(s.terminals, s.mainWin)
	})
	fileMenu.Append(saveHostItem)

	loadHostItem, _ := gtk.MenuItemNewWithLabel("Load hosts")
	loadHostItem.Connect("activate", func(_ *gtk.MenuItem) {
		LoadHostsDialog(s.terminals, s.mainWin)
		s.scrollWin.Remove(s.terminals.mainBox)
		xlog.Debug("removed")
		s.scrollWin.Add(s.terminals.Layout())
		xlog.Debug("Added")
		s.reflow(true)
		s.scrollWin.QueueDraw()
	})
	fileMenu.Append(loadHostItem)

	separator, _ := gtk.SeparatorMenuItemNew()
	fileMenu.Append(separator)

	quitItem, _ := gtk.MenuItemNewWithLabel("Quit")
	quitItem.Connect("activate", func(_ *gtk.MenuItem) { gtk.MainQuit() })
	fileMenu.Append(quitItem)

	mainMenuBar.Append(fileItem)

	editItem, _ := gtk.MenuItemNewWithLabel("Edit")
	editMenu, _ := gtk.MenuNew()
	editItem.SetSubmenu(editMenu)

	activeHostsItem, _ := gtk.MenuItemNewWithLabel("Active Hosts")
	activeHostsItem.Connect("activate", func(_ *gtk.MenuItem) {
		ActiveHostsDialog(s.terminals)
	})
	editMenu.Append(activeHostsItem)

	removeClosedItem, _ := gtk.MenuItemNewWithLabel("Remove closed")
	removeClosedItem.Connect("activate", func(_ *gtk.MenuItem) {
		s.terminals.RemoveClosedHost()
		s.reflow(true)
	})
	editMenu.Append(removeClosedItem)

	pref, _ := gtk.MenuItemNewWithLabel("Preferences")
	pref.Connect("activate", func(_ *gtk.MenuItem) {
		NewConfigDialog(s.config, func(newConf Config) {
			if s.configFile == "" {
				return
			}
			xlog.Debug("Save config to: ", s.configFile)
			s.config = newConf
			s.reflow(true)
			b, err := yaml.Marshal(s.config)
			if err != nil {
				xlog.Error("Unable to marshal config", err)
				return
			}

			err = os.WriteFile(s.configFile, b, fs.ModePerm)
			if err != nil {
				xlog.Error("Unable to save config", err)
				return
			}
		})
	})
	editMenu.Append(pref)

	mainMenuBar.Append(editItem)

	orderItem, _ := gtk.MenuItemNewWithLabel("Order")
	orderMenu, _ := gtk.MenuNew()
	orderItem.SetSubmenu(orderMenu)

	ascItem, _ := gtk.MenuItemNewWithLabel("Ascending")
	ascItem.Connect("activate", func(_ *gtk.MenuItem) {
		s.terminals.OrderAsc()
		s.reflow(true)
		s.entryBox.SetProperty("has_focus", true)
	})
	orderMenu.Add(ascItem)

	descItem, _ := gtk.MenuItemNewWithLabel("Descending")
	descItem.Connect("activate", func(_ *gtk.MenuItem) {
		s.terminals.OrderDesc()
		s.reflow(true)
		s.entryBox.SetProperty("has_focus", true)
	})
	orderMenu.Add(descItem)
	mainMenuBar.Append(orderItem)

	return mainMenuBar
}
