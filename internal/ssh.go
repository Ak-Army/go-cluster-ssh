package internal

import (
	_ "embed"
	"strings"
	"time"
	"unsafe"

	"github.com/Ak-Army/xlog"
	"github.com/electricface/go-gir/g-2.0"
	"github.com/electricface/go-gir/gdk-3.0"
	"github.com/electricface/go-gir/gi"
	"github.com/electricface/go-gir/gtk-3.0"
)

//go:embed ui/main.glade
var gladeFile string

type SSH struct {
	terminals *AllTerminal

	config        *ConfigStore
	scrollWin     gtk.ScrolledWindow
	mainWin       gtk.ApplicationWindow
	termMinWidth  int32
	termMinHeight int32
	builder       gtk.Builder
	entryBox      gtk.Entry
}

func New(hosts []*HostGroup, sshCmd string, sshArgs []string) {
	s := &SSH{}
	gtk.Init(0, 0)
	var err error
	s.builder = gtk.NewBuilder()
	_, err = s.builder.AddFromString(gladeFile, uint64(len(gladeFile)))
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
	//s.mainWin.SetSizeRequest(0, 0)
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
	s.mainWin = gtk.WrapApplicationWindow(s.builder.GetObject("windowMain").P)
	s.scrollWin = gtk.WrapScrolledWindow(s.builder.GetObject("terminals").P)
	s.scrollWin.Add(s.terminals.Layout())
	//s.scrollWin.SetSizeRequest(s.termMinWidth, s.termMinHeight)

	s.initEntryBox()
}

func (s *SSH) initEntryBox() {
	s.entryBox = gtk.WrapEntry(s.builder.GetObject("entry").P)
	// feed GNOME clipboard to all active terminals
	feedPaste := func() {
		s.terminals.PasteClipboard()
		buffer := s.entryBox.GetBuffer()
		buffer.DeleteText(0, -1)
	}
	// forward key events to all terminals with copy_input set
	feedInput := func(p gi.ParamBox) interface{} {
		ev := gdk.Event{}
		ev.P = p.Params[1].(unsafe.Pointer)

		buffer := s.entryBox.GetBuffer()
		buffer.DeleteText(0, -1)

		_, mod := ev.GetState()
		if ev.GetEventType() == gdk.EventTypeKeyPress &&
			mod&gdk.ModifierTypeControlMask == gdk.ModifierTypeControlMask &&
			mod&gdk.ModifierTypeShiftMask == gdk.ModifierTypeShiftMask {
			feedPaste()
		} else {
			s.terminals.Event(ev)
		}
		// this stops regular handler from firing, switching focus.
		return gdk.EVENT_STOP
	}
	s.entryBox.Connect("key_press_event", feedInput)
	s.entryBox.Connect("key_release_event", feedInput)
	s.entryBox.Connect("paste_clipboard", feedPaste)
	s.entryBox.Connect("button_press_event", func(p gi.ParamBox) {
		ev := gdk.Event{}
		ev.P = p.Params[1].(unsafe.Pointer)
		_, button := ev.GetButton()
		if button == gdk.BUTTON_MIDDLE {
			feedInput(p)
		}
	})
}

func (s *SSH) initMainMenuBar() {
	for k, fn := range map[string]func(menu gi.ParamBox){
		"menu.AddHost": func(_ gi.ParamBox) {
			AddHostDialog(s.builder, func(hostName string) {
				s.scrollWin.Remove(s.terminals.mainBox)
				s.terminals.AddHost("Default", hostName)
				s.scrollWin.Add(s.terminals.Layout())
				s.reflow(true)
				s.scrollWin.QueueDraw()
			})
		},
		"menu.SaveHost": func(_ gi.ParamBox) {
			SaveHostsDialog(s.builder, s.terminals)
		},
		"menu.LoadHost": func(_ gi.ParamBox) {
			LoadHostsDialog(s.builder, s.terminals)
			s.scrollWin.Remove(s.terminals.mainBox)
			s.scrollWin.Add(s.terminals.Layout())
			s.reflow(true)
			s.scrollWin.QueueDraw()
		},
		"menu.Quit": func(_ gi.ParamBox) {
			gtk.MainQuit()
		},
		"menu.ActiveHost": func(_ gi.ParamBox) {
			ActiveHostsDialog(s.terminals)
		},
		"menu.RemoveClosed": func(_ gi.ParamBox) {
			s.terminals.RemoveClosedHost()
			s.reflow(true)
		},
		"menu.Preferences": func(_ gi.ParamBox) {
			NewConfigDialog(s.builder, s.config, func() {
				s.reflow(true)
			})
		},
		"menu.Ascend": func(_ gi.ParamBox) {
			s.terminals.OrderAsc()
			s.reflow(true)
			v, _ := g.NewValue()
			v.SetBoolean(true)
			s.entryBox.GrabFocus()
		},
		"menu.Descend": func(_ gi.ParamBox) {
			s.terminals.OrderDesc()
			s.reflow(true)
			v, _ := g.NewValue()
			v.SetBoolean(true)
			s.entryBox.GrabFocus()
		},
	} {
		m := gtk.WrapMenuItem(s.builder.GetObject(k).P)
		m.Connect("activate", fn)
	}
}

func (s *SSH) createSignals() {
	s.mainWin.Connect("delete-event", func() {
		gtk.MainQuit()
	})
	s.mainWin.Connect("size-allocate", func(p gi.ParamBox) {
		window := gtk.WrapApplicationWindow(p.Params[0].(g.Object).P)
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
		xlog.Info(w, h)
		if newWidth != w || newHeight != h {
			s.mainWin.SetSizeRequest(newWidth, newHeight)
		} else {
			s.reflow(false)
		}
	})
}
