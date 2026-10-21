package platform

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/gdamore/tcell/v3/vt"
)

// fakeSurface records what is drawn, one grapheme per cell.
type fakeSurface struct {
	width, height int
	cells         [][]string
	styles        [][]tcell.Style
}

func newFakeSurface(width, height int) *fakeSurface {
	f := &fakeSurface{width: width, height: height}
	for range height {
		f.cells = append(f.cells, strings.Split(strings.Repeat(" ", width), ""))
		f.styles = append(f.styles, make([]tcell.Style, width))
	}

	return f
}

func (f *fakeSurface) PutStrStyled(x int, y int, str string, st tcell.Style) {
	g := displaywidth.StringGraphemes(str)
	w := 0
	for g.Next() {
		if y >= 0 && y < f.height && x+w >= 0 && x+w < f.width {
			f.cells[y][x+w] = g.Value()
			f.styles[y][x+w] = st
		}
		w += max(g.Width(), 1)
	}
}

func (f *fakeSurface) lines() []string {
	l := []string{}
	for _, c := range f.cells {
		l = append(l, strings.TrimRight(strings.Join(c, ""), " "))
	}

	return l
}

func (f *fakeSurface) String() string {
	return strings.Join(f.lines(), "\n")
}

func assertLines(t *testing.T, s *fakeSurface, expected []string) {
	t.Helper()
	actual := s.lines()[:len(expected)]
	for i := range expected {
		if actual[i] != expected[i] {
			t.Fatalf("line %d: expected %q, actual %q\n%s", i, expected[i], actual[i], s)
		}
	}
}

func Test_textBox(t *testing.T) {
	s := newFakeSurface(12, 5)
	tb := &textBox{block: newBlock(" T ", 0, 0, 3), text: "hello"}
	tb.draw(s, 0, 0, 12)

	assertLines(t, s, []string{
		"┌ T ───────┐",
		"│hello     │",
		"└──────────┘",
		"",
	})
}

func Test_textBoxTooLong(t *testing.T) {
	s := newFakeSurface(8, 4)
	tb := &textBox{block: newBlock("", 0, 0, 3), text: "a very long text"}
	tb.draw(s, 0, 0, 8)

	assertLines(t, s, []string{
		"┌──────┐",
		"│a ver…│",
		"└──────┘",
	})
}

func Test_textBoxMultiline(t *testing.T) {
	tb := &textBox{block: newBlock("", 0, 0, 3), text: "abcdefgh\nij", multiline: true}
	if h := tb.height(6); h != 5 {
		t.Fatalf("expected height 5, actual %d", h)
	}

	s := newFakeSurface(6, 6)
	tb.draw(s, 0, 0, 6)
	assertLines(t, s, []string{
		"┌────┐",
		"│abcd│",
		"│efgh│",
		"│ij  │",
		"└────┘",
	})
}

func Test_textBoxLines(t *testing.T) {
	testCases := []struct {
		text     string
		width    int
		expected []string
	}{
		{"wrap between words", 10, []string{"wrap", "between", "words"}},
		{"a b c", 3, []string{"a b", "c"}},
		{"a  b", 10, []string{"a  b"}},
		{"cut averylongword", 6, []string{"cut", "averyl", "ongwor", "d"}},
		{"one\ntwo", 10, []string{"one", "two"}},
	}

	for _, tc := range testCases {
		t.Run(tc.text, func(t *testing.T) {
			tb := &textBox{text: tc.text}
			if actual := tb.lines(tc.width); !slices.Equal(actual, tc.expected) {
				t.Errorf("expected %q, actual %q", tc.expected, actual)
			}
		})
	}
}

func Test_barChart(t *testing.T) {
	s := newFakeSurface(14, 7)
	bc := &barChart{
		block:    newBlock(" B ", 0, 0, 7),
		data:     []int{4, 2, 0},
		labels:   []string{"aa", "bb", "cc"},
		barWidth: 3,
		barGap:   1,
		barColor: 3,
	}
	bc.draw(s, 0, 0, 14)

	assertLines(t, s, []string{
		"┌ B ─────────┐",
		"│            │",
		"│            │",
		"│            │",
		"│ 4   2   0  │",
		"│aa  bb  cc  │",
		"└────────────┘",
	})

	// The first bar takes the whole height, the second one half of it, the last one nothing.
	green := color.PaletteColor(2)
	for y := 1; y <= 4; y++ {
		if s.styles[y][1].GetBackground() != green {
			t.Errorf("expected the first bar at line %d", y)
		}
	}
	if s.styles[2][5].GetBackground() == green || s.styles[3][5].GetBackground() != green {
		t.Errorf("expected the second bar to be half the height\n%s", s)
	}
	if s.styles[4][9].GetBackground() == green {
		t.Errorf("expected no bar for an empty value")
	}
}

func Test_barChartAllZero(t *testing.T) {
	s := newFakeSurface(10, 5)
	bc := &barChart{block: newBlock("", 0, 0, 5), data: []int{0, 0}, labels: []string{"a", "b"}, barWidth: 3, barGap: 1}
	bc.draw(s, 0, 0, 10)

	assertLines(t, s, []string{
		"┌────────┐",
		"│        │",
		"│ 0   0  │",
		"│a   b   │",
	})
}

func Test_stackedBarChart(t *testing.T) {
	s := newFakeSurface(10, 7)
	sb := &stackedBarChart{
		block:    newBlock("", 0, 0, 7),
		data:     [][]int{{2, 1}, {2, 1}},
		labels:   []string{"x", "y"},
		colors:   []uint16{2, 3},
		barWidth: 3,
		barGap:   1,
	}
	sb.draw(s, 0, 0, 10)

	assertLines(t, s, []string{
		"┌────────┐",
		"│        │",
		"│ 2      │",
		"│     1  │",
		"│ 2   1  │",
		"│ x   y  │",
		"└────────┘",
	})

	red, green := color.PaletteColor(1), color.PaletteColor(2)
	if s.styles[4][1].GetBackground() != red || s.styles[2][1].GetBackground() != green {
		t.Errorf("expected the first serie at the bottom\n%s", s)
	}
}

func Test_table(t *testing.T) {
	s := newFakeSurface(22, 6)
	ta := &table{block: newBlock(" T ", 0, 0, 0), rows: [][]string{{"name", "n"}, {"devdash", "42"}}}
	if h := ta.height(22); h != 5 {
		t.Fatalf("expected height 5, actual %d", h)
	}
	ta.draw(s, 0, 0, 22)

	assertLines(t, s, []string{
		"┌ T ─────────────────┐",
		"│  name    | n       │",
		"│────────────────────│",
		"│  devdash | 42      │",
		"└────────────────────┘",
	})
}

func Test_tableTooWide(t *testing.T) {
	s := newFakeSurface(24, 6)
	ta := &table{block: newBlock("", 0, 0, 0), rows: [][]string{{"title", "state"}, {"a very long pull request", "open"}}}
	ta.draw(s, 0, 0, 24)

	// The widest column is shrunk to keep the others.
	assertLines(t, s, []string{
		"┌──────────────────────┐",
		"│  title        | state│",
		"│──────────────────────│",
		"│  a very long… | open │",
	})
}

func Test_barChartTooManyBars(t *testing.T) {
	s := newFakeSurface(10, 4)
	bc := &barChart{block: newBlock("", 0, 0, 4), data: []int{1, 2, 3}, labels: []string{"a", "b", "c"}, barWidth: 3, barGap: 1}
	bc.draw(s, 0, 0, 10)

	// Two bars fit: the last ones are displayed.
	assertLines(t, s, []string{
		"┌────────┐",
		"│ 2   3  │",
		"│b   c   │",
	})
}

func Test_gauge(t *testing.T) {
	s := newFakeSurface(12, 3)
	g := &gauge{block: newBlock("", 0, 0, 3), percent: 50, barColor: 3}
	g.draw(s, 0, 0, 12)

	assertLines(t, s, []string{
		"┌──────────┐",
		"│  50.00%  │",
		"└──────────┘",
	})
	if s.styles[1][1].GetBackground() != color.PaletteColor(2) || s.styles[1][8].GetBackground() == color.PaletteColor(2) {
		t.Errorf("expected half of the gauge to be filled")
	}

	// The default text color is black on the colored bar, and the default color elsewhere.
	if s.styles[1][3].GetForeground() != color.PaletteColor(0) || s.styles[1][7].GetForeground() != color.Default {
		t.Errorf("expected a black text on the bar only, actual %v and %v", s.styles[1][3].GetForeground(), s.styles[1][7].GetForeground())
	}
}

func Test_rowColWidths(t *testing.T) {
	testCases := []struct {
		name     string
		spans    []int
		width    int
		expected []int
	}{
		{"complete row fills the width", []int{4, 4, 4}, 100, []int{33, 33, 34}},
		{"incomplete row", []int{6}, 100, []int{50}},
		{"full", []int{12}, 81, []int{81}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := row{}
			for _, s := range tc.spans {
				r.cols = append(r.cols, col{span: s})
			}

			actual := r.colWidths(tc.width)
			for i := range tc.expected {
				if actual[i] != tc.expected[i] {
					t.Fatalf("expected %v, actual %v", tc.expected, actual)
				}
			}
		})
	}
}

func Test_gridLayout(t *testing.T) {
	s := newFakeSurface(20, 10)
	box := func(text string, h int) widget {
		return &textBox{block: newBlock("", 0, 0, h), text: text}
	}

	r := row{cols: []col{
		{span: 6, widgets: []widget{box("a", 3), box("b", 3)}},
		{span: 6, widgets: []widget{box("c", 4)}},
	}}
	if h := r.draw(s, 0, 20); h != 6 {
		t.Fatalf("expected the row height to be the highest column, actual %d", h)
	}

	assertLines(t, s, []string{
		"┌────────┐┌────────┐",
		"│a       ││c       │",
		"└────────┘│        │",
		"┌────────┐└────────┘",
		"│b       │",
		"└────────┘",
	})
}

func Test_keyName(t *testing.T) {
	testCases := []struct {
		ev       *tcell.EventKey
		expected string
	}{
		{tcell.NewEventKey(tcell.KeyCtrlC, "", tcell.ModCtrl), "C-c"},
		{tcell.NewEventKey(tcell.KeyRune, "r", tcell.ModCtrl), "C-r"},
		{tcell.NewEventKeyEx(tcell.KeyRune, "e", tcell.ModCtrl, true, 0, 1), "C-e"},
		{tcell.NewEventKey(tcell.KeyRune, "q", tcell.ModNone), "q"},
		{tcell.NewEventKey(tcell.KeyRune, "x", tcell.ModAlt), "M-x"},
		{tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone), "<space>"},
		{tcell.NewEventKey(tcell.KeyF5, "", tcell.ModNone), "<f5>"},
		{tcell.NewEventKey(tcell.KeyEsc, "", tcell.ModNone), "<escape>"},
		{tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone), "<enter>"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if actual := keyName(tc.ev); actual != tc.expected {
				t.Errorf("expected %q, actual %q", tc.expected, actual)
			}
		})
	}
}

func Test_normalizeKey(t *testing.T) {
	for in, expected := range map[string]string{
		"C-c":   "C-c",
		"c-c":   "C-c",
		" q ":   "q",
		"<F5>":  "<f5>",
		"m-x":   "M-x",
		"<tab>": "<tab>",
	} {
		if actual := normalizeKey(in); actual != expected {
			t.Errorf("normalizeKey(%q): expected %q, actual %q", in, expected, actual)
		}
	}
}

// Test the terminal end to end with a mock terminal: render, reload and quit with keystrokes.
func Test_terminal(t *testing.T) {
	mt := vt.NewMockTerm(vt.MockOptSize{X: 30, Y: 10})
	s, err := tcell.NewTerminfoScreenFromTty(mt)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	term := newTerminal(s)
	defer term.Close()

	term.Title("devdash", 0, 0, true, 3, 12)
	term.TextBox("hello", 0, 0, " box ", 0, 3, false, false)
	term.AddCol(6)
	term.AddRow()
	term.Render()

	text := func(y int) string {
		line := ""
		for x := range 30 {
			c := mt.GetCell(vt.Coord{X: vt.Col(x), Y: vt.Row(y)})
			if c.C == "" {
				line += " "
				continue
			}
			line += c.C
		}
		return strings.TrimRight(line, " ")
	}

	if l := text(1); l != "│devdash                     │" {
		t.Errorf("expected the title, actual %q", l)
	}
	if l := text(3); l != "┌ box ────────┐" {
		t.Errorf("expected the box, actual %q", l)
	}
	if l := text(4); l != "│hello        │" {
		t.Errorf("expected the box content, actual %q", l)
	}

	reload := make(chan time.Time, 1)
	term.KHotReload("C-r", reload)
	term.KQuit("C-c")

	done := make(chan struct{})
	go func() {
		term.Loop()
		close(done)
	}()

	mt.KeyTap(vt.KeyLCtrl, vt.KeyR)
	select {
	case <-reload:
	case <-time.After(5 * time.Second):
		t.Fatal("expected a reload after C-r")
	}

	mt.KeyTap(vt.KeyLCtrl, vt.KeyC)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("expected the loop to stop after C-c")
	}
}
