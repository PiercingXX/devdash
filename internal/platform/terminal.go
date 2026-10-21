package platform

import (
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v3"
)

// gridColumns is the number of columns of the grid; the size of a column is a fraction of it.
const gridColumns = 12

// surface is where the widgets are drawn.
type surface interface {
	PutStrStyled(x int, y int, str string, style tcell.Style)
}

// widget is an element of the grid with a fixed height.
type widget interface {
	// height of the widget for a given width.
	height(width int) int
	draw(s surface, x int, y int, width int)
}

type col struct {
	span    int
	widgets []widget
}

type row struct {
	cols []col
}

// colWidths of a row. The last column fills the remaining space if the row is complete.
func (r row) colWidths(width int) []int {
	widths := make([]int, len(r.cols))
	spans, x := 0, 0
	for i, c := range r.cols {
		spans += c.span
		widths[i] = c.span * width / gridColumns
		if i == len(r.cols)-1 && spans == gridColumns {
			widths[i] = width - x
		}
		x += widths[i]
	}

	return widths
}

// draw the row and return its height.
func (r row) draw(s surface, y int, width int) int {
	h, x := 0, 0
	for i, w := range r.colWidths(width) {
		cy := y
		for _, wi := range r.cols[i].widgets {
			wi.draw(s, x, cy, w)
			cy += wi.height(w)
		}
		h = max(h, cy-y)
		x += w
	}

	return h
}

// terminal is the terminal user interface: a grid of rows, each row having columns of widgets.
type terminal struct {
	mu     sync.Mutex
	screen tcell.Screen

	body    []row
	cols    []col
	widgets []widget

	keys map[string]func()
	quit chan struct{}
}

// NewTerminal initializes the terminal.
func NewTerminal() (*terminal, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	return newTerminal(s), nil
}

func newTerminal(s tcell.Screen) *terminal {
	s.SetStyle(tcell.StyleDefault)
	s.Clear()

	return &terminal{
		screen: s,
		keys:   map[string]func(){},
		quit:   make(chan struct{}),
	}
}

// AddCol to the grid, with the widgets added since the last column.
func (t *terminal) AddCol(size int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cols = append(t.cols, col{span: size, widgets: t.widgets})
	t.widgets = nil
}

// AddRow to the grid, with the columns added since the last row.
func (t *terminal) AddRow() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.body = append(t.body, row{cols: t.cols})
	t.cols = nil
}

func (t *terminal) addWidget(w widget) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.widgets = append(t.widgets, w)
}

// TextBox widget type.
func (t *terminal) TextBox(
	data string,
	textColor uint16,
	borderColor uint16,
	title string,
	tc uint16,
	height int,
	multiline bool,
	bold bool,
) {
	t.addWidget(&textBox{
		block:     newBlock(title, borderColor, tc, height),
		text:      data,
		style:     style(textColor).Bold(bold),
		multiline: multiline,
	})
}

// Gauge widget type.
func (t *terminal) Gauge(
	data float64,
	textColor uint16,
	barColor uint16,
	borderColor uint16,
	title string,
	tc uint16,
	height int,
) {
	t.addWidget(&gauge{
		block:    newBlock(title, borderColor, tc, height),
		percent:  data,
		barColor: barColor,
		numColor: textColor,
	})
}

// Title is a TextBox displayed in its own row.
func (t *terminal) Title(
	title string,
	textColor uint16,
	borderColor uint16,
	bold bool,
	height int,
	size int,
) {
	tb := &textBox{
		block: newBlock("", borderColor, defaultColor, height),
		text:  title,
		style: style(textColor).Bold(bold),
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.body = append(t.body, row{cols: []col{{span: size, widgets: []widget{tb}}}})
}

// BarChart widget type.
func (t *terminal) BarChart(
	data []int,
	dimensions []string,
	title string,
	tc uint16,
	bd uint16,
	fg uint16,
	nc uint16,
	enc uint16,
	height int,
	gap int,
	barWidth int,
	barColor uint16,
) {
	t.addWidget(&barChart{
		block:         newBlock(title, bd, tc, height),
		data:          data,
		labels:        dimensions,
		textColor:     fg,
		numColor:      nc,
		emptyNumColor: enc,
		barColor:      barColor,
		barGap:        gap,
		barWidth:      barWidth,
	})
}

// StackedBarChart widget type.
func (t *terminal) StackedBarChart(
	data [8][]int,
	dimensions []string,
	title string,
	tc uint16,
	colors []uint16,
	bd uint16,
	fg uint16,
	nc uint16,
	height int,
	gap int,
	barWidth int,
) {
	series := [][]int{}
	for _, d := range data {
		if d == nil {
			break
		}
		series = append(series, d)
	}

	t.addWidget(&stackedBarChart{
		block:     newBlock(title, bd, tc, height),
		data:      series,
		labels:    dimensions,
		textColor: fg,
		numColor:  nc,
		colors:    colors,
		barGap:    gap,
		barWidth:  barWidth,
	})
}

// Table widget type.
func (t *terminal) Table(
	data [][]string,
	title string,
	tc uint16,
	bd uint16,
	fg uint16,
) {
	t.addWidget(&table{
		block: newBlock(title, bd, tc, 0),
		rows:  data,
		style: style(fg),
	})
}

// KQuit set a key to quit the application.
func (t *terminal) KQuit(key string) {
	t.handle(key, func() {
		select {
		case <-t.quit:
		default:
			close(t.quit)
		}
	})
}

// KHotReload set a key to reload the dashboard.
func (t *terminal) KHotReload(key string, c chan<- time.Time) {
	t.handle(key, func() {
		go func() {
			c <- time.Now()
		}()
	})
}

// KEdit set a key to edit a dashboard config.
// The terminal is given to the editor until it exits.
func (t *terminal) KEdit(key string, editDashboard func()) {
	t.handle(key, func() {
		if err := t.screen.Suspend(); err != nil {
			return
		}
		editDashboard()
		if err := t.screen.Resume(); err != nil {
			return
		}
		t.Render()
	})
}

// KAction set a key to run an action, without blocking the events.
func (t *terminal) KAction(key string, action func()) {
	t.handle(key, func() {
		go action()
	})
}

func (t *terminal) handle(key string, f func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.keys[normalizeKey(key)] = f
}

// Loop to receive the events of the terminal, until the quit key is pressed.
func (t *terminal) Loop() {
	for {
		select {
		case <-t.quit:
			return
		case ev, ok := <-t.screen.EventQ():
			if !ok {
				return
			}

			switch ev := ev.(type) {
			case *tcell.EventResize:
				t.screen.Sync()
				t.Render()
			case *tcell.EventKey:
				t.mu.Lock()
				f, ok := t.keys[keyName(ev)]
				t.mu.Unlock()
				if ok {
					f()
				}
			}
		}
	}
}

// Render the whole grid.
func (t *terminal) Render() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.screen.Clear()
	width, _ := t.screen.Size()
	y := 0
	for _, r := range t.body {
		y += r.draw(t.screen, y, width)
	}
	t.screen.Show()
}

// Clean and create a new empty grid.
func (t *terminal) Clean() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.body = nil
	t.cols = nil
	t.widgets = nil
}

// Close the terminal and give it back to the shell.
func (t *terminal) Close() {
	t.screen.Fini()
}

// HotReload empties the grid. The current widgets stay displayed until the next render.
func (t *terminal) HotReload() {
	t.Clean()
}

// normalizeKey from the configuration, which uses the termui notation: "C-c", "M-x", "<f5>", "q"...
func normalizeKey(key string) string {
	key = strings.TrimSpace(key)
	switch {
	case strings.HasPrefix(key, "c-"):
		return "C-" + key[2:]
	case strings.HasPrefix(key, "m-"):
		return "M-" + key[2:]
	case strings.HasPrefix(key, "<") && strings.HasSuffix(key, ">"):
		return strings.ToLower(key)
	}

	return key
}

var specialKeys = map[tcell.Key]string{
	tcell.KeyEnter:     "<enter>",
	tcell.KeyTab:       "<tab>",
	tcell.KeyBacktab:   "<backtab>",
	tcell.KeyBackspace: "<backspace>",
	tcell.KeyEsc:       "<escape>",
	tcell.KeyDelete:    "<delete>",
	tcell.KeyInsert:    "<insert>",
	tcell.KeyUp:        "<up>",
	tcell.KeyDown:      "<down>",
	tcell.KeyLeft:      "<left>",
	tcell.KeyRight:     "<right>",
	tcell.KeyHome:      "<home>",
	tcell.KeyEnd:       "<end>",
	tcell.KeyPgUp:      "<previous>",
	tcell.KeyPgDn:      "<next>",
	tcell.KeyF1:        "<f1>",
	tcell.KeyF2:        "<f2>",
	tcell.KeyF3:        "<f3>",
	tcell.KeyF4:        "<f4>",
	tcell.KeyF5:        "<f5>",
	tcell.KeyF6:        "<f6>",
	tcell.KeyF7:        "<f7>",
	tcell.KeyF8:        "<f8>",
	tcell.KeyF9:        "<f9>",
	tcell.KeyF10:       "<f10>",
	tcell.KeyF11:       "<f11>",
	tcell.KeyF12:       "<f12>",
}

// keyName of a key event, using the termui notation of the configuration.
func keyName(ev *tcell.EventKey) string {
	k := ev.Key()
	if k >= tcell.KeyCtrlA && k <= tcell.KeyCtrlZ {
		return "C-" + string(rune('a'+k-tcell.KeyCtrlA))
	}

	if k == tcell.KeyRune {
		s := ev.Str()
		if s == " " {
			s = "<space>"
		}

		switch {
		case ev.Modifiers()&tcell.ModCtrl != 0:
			return "C-" + strings.ToLower(s)
		case ev.Modifiers()&tcell.ModAlt != 0:
			return "M-" + s
		}

		return s
	}

	return specialKeys[k]
}
