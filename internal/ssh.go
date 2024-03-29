package internal

import (
	_ "embed"
	"strings"
	"time"

	"github.com/Ak-Army/xlog"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed ui/main.glade
var gladeFile string

type SSH struct {
	terminals *AllTerminal

	config        *ConfigStore
	scrollWin     *gtk.ScrolledWindow
	mainWin       *gtk.ApplicationWindow
	termMinWidth  int
	termMinHeight int
	builder       *gtk.Builder
	entryBox      *gtk.Entry
}

func New(hosts []*HostGroup, sshCmd string, sshArgs []string) {
	s := &SSH{}
	gtk.Init(nil)
	var err error
	s.builder, _ = gtk.BuilderNew()
	if s.builder == nil {
		xlog.Fatal("Unable to create new GTK builder")
	}
	err = s.builder.AddFromString(gladeFile)
	if err != nil {
		xlog.Fatal("Unable to load main.glade", err)
	}
	s.config, err = NewConfig()
	if err != nil {
		xlog.Warn("Unable to init config", err)
	}
	if sshCmd == "" {
		sshCmd = "/usr/bin/ssh"
	}

	s.termMinWidth = 1
	s.termMinHeight = 1

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
	s.createSignals()
	s.initMainMenuBar()
	if s.config.Config().StartMaximized {
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
	s.terminals.Reflow(width, force, s.config.Config())

	title := "go-cluster-ssh:" + strings.Join(s.terminals.Names(), " ")

	s.mainWin.SetTitle(title)
}

func (s *SSH) configTerminals() {
	s.terminals.Each(func(t *Terminal) {
		conf := s.config.Config()
		t.SetScrollbackLines(-1)
		t.SetSizeRequest(conf.MinWidth, conf.MinHeight)
		t.SetFontFromString(conf.Font)
		if s.termMinWidth < conf.MinWidth {
			s.termMinWidth = conf.MinWidth
		}
		s.termMinHeight = t.GetAllocatedHeight()
		if s.termMinHeight < conf.MinHeight {
			s.termMinHeight = conf.MinHeight
		}
	})
}

func (s *SSH) initGUI() {
	// GUI Objects
	mv, _ := s.builder.GetObject("windowMain")
	t, _ := s.builder.GetObject("terminals")
	s.mainWin = mv.(*gtk.ApplicationWindow)
	s.scrollWin = t.(*gtk.ScrolledWindow)
	s.scrollWin.Add(s.terminals.Layout())
	//s.scrollWin.SetSizeRequest(s.termMinWidth, s.termMinHeight)

	s.initEntryBox()
}

func (s *SSH) initEntryBox() {
	eb, _ := s.builder.GetObject("entry")
	s.entryBox = eb.(*gtk.Entry)
	// feed GNOME clipboard to all active terminals
	feedPaste := func(_ *gtk.Entry, _ *gdk.Event) {
		s.terminals.PasteClipboard()
		buffer, _ := s.entryBox.GetBuffer()
		buffer.DeleteText(0, -1)
	}
	// forward key events to all terminals with copy_input set
	feedInput := func(widget *gtk.Entry, ev *gdk.Event) bool {
		buffer, _ := s.entryBox.GetBuffer()
		buffer.DeleteText(0, -1)

		keyEvent := &gdk.EventKey{Event: ev}
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
	s.entryBox.Connect("key_press_event", feedInput)
	s.entryBox.Connect("key_release_event", feedInput)
	s.entryBox.Connect("paste_clipboard", feedPaste)
	s.entryBox.Connect("button_press_event", func(widget *gtk.Entry, ev *gdk.Event) {
		if gdk.EventButtonNewFromEvent(ev).Button() == gdk.BUTTON_MIDDLE {
			feedInput(widget, ev)
		}
	})
}

func (s *SSH) initMainMenuBar() {
	for k, fn := range map[string]func(menu *gtk.MenuItem){
		"menu.AddHost": func(_ *gtk.MenuItem) {
			AddHostDialog(s.builder, func(hostName string) {
				s.scrollWin.Remove(s.terminals.mainBox)
				s.terminals.AddHost("Default", hostName)
				s.scrollWin.Add(s.terminals.Layout())
				s.reflow(true)
				s.scrollWin.QueueDraw()
			})
		},
		"menu.SaveHost": func(_ *gtk.MenuItem) {
			SaveHostsDialog(s.builder, s.terminals)
		},
		"menu.LoadHost": func(_ *gtk.MenuItem) {
			LoadHostsDialog(s.builder, s.terminals)
			s.scrollWin.Remove(s.terminals.mainBox)
			s.scrollWin.Add(s.terminals.Layout())
			s.reflow(true)
			s.scrollWin.QueueDraw()
		},
		"menu.Quit": func(_ *gtk.MenuItem) {
			gtk.MainQuit()
		},
		"menu.ActiveHost": func(_ *gtk.MenuItem) {
			ActiveHostsDialog(s.terminals)
		},
		"menu.RemoveClosed": func(_ *gtk.MenuItem) {
			s.terminals.RemoveClosedHost()
			s.reflow(true)
		},
		"menu.Preferences": func(_ *gtk.MenuItem) {
			NewConfigDialog(s.builder, s.config, func() {
				s.reflow(true)
			})
		},
		"menu.Ascend": func(_ *gtk.MenuItem) {
			s.terminals.OrderAsc()
			s.reflow(true)
			s.entryBox.SetProperty("has_focus", true)
		},
		"menu.Descend": func(_ *gtk.MenuItem) {
			s.terminals.OrderDesc()
			s.reflow(true)
			s.entryBox.SetProperty("has_focus", true)
		},
	} {
		m, _ := s.builder.GetObject(k)
		m.(*gtk.MenuItem).Connect("activate", fn)
	}
}

func (s *SSH) createSignals() {
	s.mainWin.Connect("delete-event", func(_ interface{}) {
		gtk.MainQuit()
	})
	s.mainWin.Connect("size-allocate", func(window *gtk.ApplicationWindow) {
		conf := s.config.Config()
		w, h := window.GetSize()
		newWidth := w
		newHeight := h
		if w < conf.MinWidth {
			newWidth = conf.MinWidth
		}
		if h < conf.MinHeight {
			newHeight = conf.MinHeight
		}
		if newWidth != w || newHeight != h {
			window.SetSizeRequest(newWidth, newHeight)
		} else {
			s.reflow(false)
		}
	})
}
