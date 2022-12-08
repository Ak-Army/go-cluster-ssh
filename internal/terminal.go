package internal

import "C"
import (
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/Ak-Army/xlog"
	"github.com/electricface/go-gir/g-2.0"
	"github.com/electricface/go-gir/gdk-3.0"
	"github.com/electricface/go-gir/gi"
	"github.com/electricface/go-gir/gtk-3.0"

	"github.com/sqp/vte"
	"go.uber.org/atomic"
)

type HostGroup struct {
	Name  string
	Hosts []string
}

type AllTerminal struct {
	terminals []*Terminals
	sshCmd    string
	sshArgs   []string
	reflow    func(t gtk.IWidget)
	mainBox   gtk.Box
	inited    bool
}

func (t *AllTerminal) Len() int {
	var l int
	for _, term := range t.terminals {
		l += term.Len()
	}
	return l
}

func (t *AllTerminal) Reflow(width int32, force bool, c *Config) {
	for _, term := range t.terminals {
		term.Reflow(width, force, c)
	}
}

func (t *AllTerminal) Each(f func(t *Terminal)) {
	for _, term := range t.terminals {
		term.Each(f)
	}
}

func (t *AllTerminal) Names() []string {
	var names []string
	for _, term := range t.terminals {
		names = append(names, term.Names()...)
	}
	return names
}

func (t *AllTerminal) Layout() gtk.IWidget {
	if t.inited {
		t.mainBox.Destroy()
	}
	t.inited = true
	t.mainBox = gtk.NewBox(gtk.OrientationVertical, 0)
	t.mainBox.SetBorderWidth(0)

	l := len(t.terminals)
	xlog.Debug("terminals:", l)
	for _, term := range t.terminals {
		xlog.Debug("Add terminals ", term.Name)
		if l > 1 {
			header := gtk.NewHeaderBar()
			header.SetTitle(term.Name)
			theme := gtk.IconThemeGetDefault1()
			t.addButtons(theme, header, term)
			t.mainBox.PackStart(header, false, false, 0)
		}
		if !term.IsHidden() {
			t.mainBox.PackStart(term.Layout(), true, true, 0)
		}
	}
	return t.mainBox
}

func (t *AllTerminal) addButtons(theme gtk.IconTheme, header gtk.HeaderBar, term *Terminals) {
	upButton := gtk.NewButton()
	var upImage gtk.Image
	var downImage gtk.Image
	if theme.HasIcon("go-up") {
		icon, err := theme.LoadIcon("go-up", int32(gtk.IconSizeButton), gtk.IconLookupFlagsUseBuiltin)
		if err == nil {
			upImage = gtk.NewImageFromPixbuf(icon)
		}
	}
	if theme.HasIcon("go-down") {
		icon, err := theme.LoadIcon("go-down", int32(gtk.IconSizeButton), gtk.IconLookupFlagsUseBuiltin)
		if err == nil {
			downImage = gtk.NewImageFromPixbuf(icon)
		}
	}
	if term.IsHidden() {
		upButton.Add(downImage)
	} else {
		upButton.Add(upImage)
	}
	upButton.Connect("clicked", func() {
		xlog.Info("Hide: ", term.IsHidden())
		if term.IsHidden() {
			term.Show()
		} else {
			term.Hide()
		}
		t.reflow(t.mainBox)
	})
	header.PackEnd(upButton)

	closeButton := gtk.NewButton()
	if theme.HasIcon("window-close") {
		icon, err := theme.LoadIcon("window-close", int32(gtk.IconSizeButton), gtk.IconLookupFlagsUseBuiltin)
		if err == nil {
			closeImage := gtk.NewImageFromPixbuf(icon)
			closeButton.Add(closeImage)
		}
	}
	closeButton.Connect("clicked", func() {
		t.RemoveGroup(term.Name)
	})
	header.PackEnd(closeButton)
}

func (t *AllTerminal) PasteClipboard() {
	for _, term := range t.terminals {
		term.PasteClipboard()
	}
}

func (t *AllTerminal) Event(ev gdk.Event) {
	for _, term := range t.terminals {
		term.Event(ev)
	}
}

func (t *AllTerminal) AddHost(group string, name string) {
	for _, term := range t.terminals {
		if group == term.Name {
			term.AddHost(name)
			return
		}
	}
	terms := NewTerminals(t.sshCmd, t.sshArgs, group)
	terms.AddHost(name)
	terms.isHidden.Store(false)

	t.terminals = append(t.terminals, terms)
}

func (t *AllTerminal) RemoveClosedHost() {
	for _, term := range t.terminals {
		term.RemoveClosedHost()
	}
}

func (t *AllTerminal) OrderAsc() {
	for _, term := range t.terminals {
		term.OrderAsc()
	}
}

func (t *AllTerminal) OrderDesc() {
	for _, term := range t.terminals {
		term.OrderDesc()
	}
}

func (t *AllTerminal) Activate(host string, active bool) {
	for _, term := range t.terminals {
		term.Activate(host, active)
	}
}

func (t *AllTerminal) RemoveGroup(name string) {
	xlog.Debug("remove ", name)
	for i, term := range t.terminals {
		if term.Name == name {
			t.terminals = append(t.terminals[:i], t.terminals[i+1:]...)
			break
		}
	}
	t.reflow(t.mainBox)
}

func NewAllTerminals(sshCmd string, sshArgs []string, group []*HostGroup, reflow func(t gtk.IWidget)) *AllTerminal {
	t := &AllTerminal{
		sshCmd:  sshCmd,
		sshArgs: sshArgs,
		reflow:  reflow,
	}
	for _, g := range group {
		terms := NewTerminals(sshCmd, sshArgs, g.Name)
		for _, h := range g.Hosts {
			terms.AddHost(h)
		}
		t.terminals = append(t.terminals, terms)
	}
	return t
}

type Terminals struct {
	Name  string
	Hosts []string

	sshCmd    string
	sshArgs   []string
	terminals []*Terminal
	mu        sync.Mutex

	layoutTable gtk.Grid

	isHidden atomic.Bool
	Cols     int32
	Rows     int32
}

type Terminal struct {
	gtk.Widget
	*vte.Terminal

	Host      string
	CopyInput bool
	closeAble atomic.Bool
}

type reflowConfig struct {
	Cols        int32
	Rows        int32
	LayoutTable gtk.Grid
	MinWidth    int32
	MinHeight   int32
}

func NewTerminals(sshCmd string, sshArgs []string, name string) *Terminals {
	t := &Terminals{
		sshCmd:  sshCmd,
		sshArgs: sshArgs,
		Name:    name,
	}
	t.layoutTable = gtk.NewGrid()
	t.layoutTable.SetRowHomogeneous(true)
	t.layoutTable.SetColumnHomogeneous(true)
	t.layoutTable.SetRowSpacing(1)
	t.layoutTable.SetColumnSpacing(1)

	return t
}

func newTerminal(host string) *Terminal {
	t := vte.NewTerminal()
	if t == nil {
		return nil
	}
	return &Terminal{
		Widget: gtk.Widget{
			InitiallyUnowned: g.InitiallyUnowned{
				Object: g.WrapObject(unsafe.Pointer(t.Native())),
			},
		},
		Terminal: t,
		Host:     host,
	}
}

func (t *Terminals) Reflow(width int32, force bool, c *Config) {
	numTerms := int32(t.Len())
	cs, _ := t.layoutTable.Get("column_spacing")

	t.Cols = (width - cs.GetInt()) / c.MinWidth
	if t.Cols < 1 || numTerms == 1 {
		t.Cols = 1
	} else if t.Cols > numTerms {
		t.Cols = numTerms
	}
	t.Rows = numTerms / t.Cols
	if t.Rows < 1 {
		t.Rows = 1
	}
	// ensure we evenly distribute terminals per row.
	t.Cols = numTerms / t.Rows
	xlog.Debugf("Reflow width %s %d => cols: %.0f, rows: %.0f numTerms: %d", t.Name, width, t.Cols, t.Rows, numTerms)
	nc, _ := t.layoutTable.Get("n_columns")
	nr, _ := t.layoutTable.Get("n_rows")
	if nc.GetInt() != t.Rows || nr.GetInt() != t.Rows || force {
		t.ReflowTable(&reflowConfig{
			Cols:        t.Cols,
			Rows:        t.Rows,
			LayoutTable: t.layoutTable,
			MinWidth:    c.MinWidth,
			MinHeight:   c.MinHeight,
		})
	}
}

func (t *Terminals) ReflowTable(rc *reflowConfig) {
	if t.isHidden.Load() {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	xlog.Debugf("Reflow table %d %d", rc.Cols, rc.Rows)
	//empty table and re-size
	for _, t := range t.terminals {
		rc.LayoutTable.Remove(t)
	}
	// layout terminals
	i := 0
	l := len(t.terminals)
	for r := int32(1); r <= rc.Rows; r++ {
		for c := int32(1); c <= rc.Cols; c++ {
			t.terminals[i].SetSizeRequest(rc.MinWidth, rc.MinHeight)
			t.terminals[i].SetTooltipText(t.terminals[i].Host)
			rc.LayoutTable.Attach(t.terminals[i], c, r, 1, 1)
			i++
			if l == i {
				return
			}
		}
	}
}

func (t *Terminals) AddHost(host string) {
	term := newTerminal(host)

	cmd := []string{t.sshCmd}
	cmd = append(cmd, t.sshArgs...)
	cmd = append(cmd, host)
	term.ExecAsync(vte.Cmd{
		Args:    cmd,
		Timeout: -1,
		OnExec: func(pid int, err error) {
			if err != nil {
				xlog.Error("Exit cause:", err)
				return
			}
			xlog.Infof("New terminal: %s args:%s host:%s, pid: %d, err: %#v",
				t.sshCmd,
				strings.Join(t.sshArgs, " "),
				host,
				pid,
				err)
		},
	})
	term.CopyInput = true
	// attach copy/paste handler
	term.Connect("key_press_event", func(p gi.ParamBox) {
		ev := gdk.Event{}
		ev.P = p.Params[1].(unsafe.Pointer)
		// check for paste key shortcut (ctl-shift-v/c)
		_, mod := ev.GetState()
		if ev.GetEventType() == gdk.EventTypeKeyPress &&
			mod&gdk.ModifierTypeControlMask == gdk.ModifierTypeControlMask &&
			mod&gdk.ModifierTypeShiftMask == gdk.ModifierTypeShiftMask {
			_, v := ev.GetKeyval()
			switch v {
			case gdk.KEY_V:
				term.PasteClipboard()
			case gdk.KEY_C:
				term.CopyClipboard()
			}
		}
	})
	t.mu.Lock()
	t.terminals = append(t.terminals, term)
	t.mu.Unlock()
	// hook terminals so they reflow layout on exit
	term.Connect("child-exited", func() {
		term.closeAble.Store(true)
	})
}

func (t *Terminals) RemoveClosedHost() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, term := range t.terminals {
		if term.closeAble.Load() {
			xlog.Info("Disconnected from: " + term.Host)
			t.terminals = append(t.terminals[:i], t.terminals[i+1:]...)
			t.layoutTable.Remove(term)
		}
	}
}

func (t *Terminals) Event(ev gdk.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, t := range t.terminals {
		if t.CopyInput {
			t.Event(ev)
		}
	}
}

func (t *Terminals) PasteClipboard() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, t := range t.terminals {
		if t.CopyInput {
			t.PasteClipboard()
		}
	}
}

func (t *Terminals) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.terminals)
}

func (t *Terminals) Names() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var names []string
	for _, t := range t.terminals {
		names = append(names, t.Host)
	}
	return names
}

func (t *Terminals) Activate(host string, active bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, t := range t.terminals {
		if t.Host == host {
			t.CopyInput = active
		}
	}
}

func (t *Terminals) Each(f func(t *Terminal)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, t := range t.terminals {
		f(t)
	}
}

func (t *Terminals) OrderAsc() {
	t.mu.Lock()
	defer t.mu.Unlock()
	sort.Slice(t.terminals, func(i, j int) bool {
		if t.terminals[i].Host != t.terminals[j].Host {
			return t.terminals[i].Host < t.terminals[j].Host
		}
		return false
	})
}

func (t *Terminals) OrderDesc() {
	t.mu.Lock()
	defer t.mu.Unlock()
	sort.Slice(t.terminals, func(i, j int) bool {
		if t.terminals[i].Host != t.terminals[j].Host {
			return t.terminals[i].Host > t.terminals[j].Host
		}
		return false
	})
}

func (t *Terminals) Layout() gtk.Grid {
	return t.layoutTable
}

func (t *Terminals) Hide() {
	t.isHidden.Store(true)
	t.layoutTable.Hide()
}

func (t *Terminals) Show() {
	t.isHidden.Store(false)
	t.layoutTable.Show()
}

func (t *Terminals) IsHidden() bool {
	return t.isHidden.Load()
}
