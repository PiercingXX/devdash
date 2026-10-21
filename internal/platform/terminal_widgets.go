package platform

import (
	"strconv"
	"strings"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// The colors of the configuration: 0 is the default color of the terminal, then the 8 basic colors.
const (
	defaultColor uint16 = iota
	blackColor
)

func toColor(c uint16) color.Color {
	if c == defaultColor || c > 8 {
		return color.Default
	}

	return color.PaletteColor(int(c - blackColor))
}

// style with a foreground color.
func style(fg uint16) tcell.Style {
	return tcell.StyleDefault.Foreground(toColor(fg))
}

// fillStyle of a bar. The default color is the reverse of the default background, to be visible.
func fillStyle(bg uint16) tcell.Style {
	if bg == defaultColor {
		return tcell.StyleDefault.Reverse(true)
	}

	return tcell.StyleDefault.Background(toColor(bg))
}

// onBarStyle of a text displayed on a bar. A text with the default color is black on a colored bar, to be readable.
func onBarStyle(bg uint16, fg uint16) tcell.Style {
	if fg == defaultColor && bg != defaultColor {
		fg = blackColor
	}

	return fillStyle(bg).Foreground(toColor(fg))
}

// put a string, cut to the given width. Return the width used.
func put(s surface, x int, y int, str string, maxWidth int, st tcell.Style) int {
	if maxWidth <= 0 {
		return 0
	}

	str = displaywidth.TruncateString(str, maxWidth, "")
	s.PutStrStyled(x, y, str, st)

	return displaywidth.String(str)
}

// fill an area with a style.
func fill(s surface, x int, y int, width int, height int, st tcell.Style) {
	if width <= 0 {
		return
	}

	line := strings.Repeat(" ", width)
	for i := range height {
		s.PutStrStyled(x, y+i, line, st)
	}
}

// block is a widget with a border and a title on the border.
type block struct {
	title       string
	borderStyle tcell.Style
	titleStyle  tcell.Style
	h           int
}

func newBlock(title string, borderColor uint16, titleColor uint16, height int) block {
	return block{
		title:       title,
		borderStyle: style(borderColor),
		titleStyle:  style(titleColor),
		h:           height,
	}
}

func (b block) height(int) int {
	return b.h
}

// draw the border and returns the inner area.
func (b block) draw(s surface, x int, y int, width int, height int) (ix int, iy int, iw int, ih int) {
	if width < 2 || height < 2 {
		return x, y, 0, 0
	}

	horizontal := strings.Repeat("─", width-2)
	s.PutStrStyled(x, y, "┌"+horizontal+"┐", b.borderStyle)
	for i := 1; i < height-1; i++ {
		s.PutStrStyled(x, y+i, "│", b.borderStyle)
		s.PutStrStyled(x+width-1, y+i, "│", b.borderStyle)
	}
	s.PutStrStyled(x, y+height-1, "└"+horizontal+"┘", b.borderStyle)

	put(s, x+1, y, b.title, width-2, b.titleStyle)

	return x + 1, y + 1, width - 2, height - 2
}

type textBox struct {
	block
	text      string
	style     tcell.Style
	multiline bool
}

// lines of the text, wrapped to the given width.
// The lines are wrapped between words; a word longer than the width is cut.
func (t *textBox) lines(width int) []string {
	lines := []string{}
	for l := range strings.SplitSeq(t.text, "\n") {
		line, w := "", 0
		// Splitting on every space keeps the spaces between the words.
		for i, word := range strings.Split(l, " ") {
			if i > 0 {
				if w+1+displaywidth.String(word) <= width {
					line, w = line+" ", w+1
				} else {
					lines = append(lines, line)
					line, w = "", 0
				}
			}

			g := displaywidth.StringGraphemes(word)
			for g.Next() {
				if w+g.Width() > width && w > 0 {
					lines = append(lines, line)
					line, w = "", 0
				}
				line += g.Value()
				w += g.Width()
			}
		}
		lines = append(lines, line)
	}

	return lines
}

// height of a multiline text box grows with its text.
func (t *textBox) height(width int) int {
	if !t.multiline || width < 3 {
		return t.h
	}

	return max(t.h, len(t.lines(width-2))+2)
}

func (t *textBox) draw(s surface, x int, y int, width int) {
	ix, iy, iw, ih := t.block.draw(s, x, y, width, t.height(width))
	if iw <= 0 || ih <= 0 {
		return
	}

	lines := t.lines(iw)
	for i, l := range lines {
		if i >= ih {
			// The text is too long for the box.
			s.PutStrStyled(ix+iw-1, iy+ih-1, "…", t.style)
			break
		}
		put(s, ix, iy+i, l, iw, t.style)
	}
}

type gauge struct {
	block
	percent  float64
	barColor uint16
	numColor uint16
}

func (g *gauge) draw(s surface, x int, y int, width int) {
	ix, iy, iw, ih := g.block.draw(s, x, y, width, g.h)
	if iw <= 0 || ih <= 0 {
		return
	}

	filled := min(max(int(g.percent*float64(iw)/100), 0), iw)
	fill(s, ix, iy, filled, ih, fillStyle(g.barColor))

	label := strconv.FormatFloat(g.percent, 'f', 2, 64) + "%"
	pos := max((iw-displaywidth.String(label))/2, 0)
	ly := iy + ih/2
	for i, r := range []rune(label) {
		if pos+i >= iw {
			break
		}

		st := style(g.numColor)
		if pos+i < filled {
			st = onBarStyle(g.barColor, g.numColor)
		}
		s.PutStrStyled(ix+pos+i, ly, string(r), st)
	}
}

type barChart struct {
	block
	data          []int
	labels        []string
	textColor     uint16
	numColor      uint16
	emptyNumColor uint16
	barColor      uint16
	barGap        int
	barWidth      int
}

func (b *barChart) draw(s surface, x int, y int, width int) {
	ix, iy, iw, ih := b.block.draw(s, x, y, width, b.h)
	if iw <= 0 || ih < 2 || b.barWidth <= 0 {
		return
	}

	maxValue := 0
	for _, v := range b.data {
		maxValue = max(maxValue, v)
	}

	// The last line of the chart is for the labels.
	// The bars are mostly time series: the last ones are displayed when they don't all fit.
	barsHeight := ih - 1
	numValues := min(len(b.data), len(b.labels))
	first := max(numValues-iw/(b.barGap+b.barWidth), 0)
	for i := first; i < numValues; i++ {
		bx := ix + (i-first)*(b.barWidth+b.barGap)

		h := 0
		if maxValue > 0 {
			h = b.data[i] * barsHeight / maxValue
		}
		fill(s, bx, iy+barsHeight-h, b.barWidth, h, fillStyle(b.barColor))

		put(s, bx, iy+ih-1, b.labels[i], b.barWidth, style(b.textColor))

		num := displaywidth.TruncateString(strconv.Itoa(b.data[i]), b.barWidth, "")
		numStyle := onBarStyle(b.barColor, b.numColor)
		if h == 0 {
			numStyle = style(b.emptyNumColor)
		}
		put(s, bx+(b.barWidth-displaywidth.String(num))/2, iy+barsHeight-1, num, b.barWidth, numStyle)
	}
}

type stackedBarChart struct {
	block
	data      [][]int
	labels    []string
	textColor uint16
	numColor  uint16
	colors    []uint16
	barGap    int
	barWidth  int
}

func (b *stackedBarChart) color(serie int) uint16 {
	if serie < len(b.colors) {
		return b.colors[serie]
	}

	return defaultColor
}

func (b *stackedBarChart) draw(s surface, x int, y int, width int) {
	ix, iy, iw, ih := b.block.draw(s, x, y, width, b.h)
	if iw <= 0 || ih < 2 || b.barWidth <= 0 || len(b.data) == 0 {
		return
	}

	// Only the bars with a value for every serie are displayed.
	numValues := len(b.labels)
	for _, d := range b.data {
		numValues = min(numValues, len(d))
	}

	maxValue := 0
	for i := range numValues {
		sum := 0
		for _, d := range b.data {
			sum += d[i]
		}
		maxValue = max(maxValue, sum)
	}

	// The last line of the chart is for the labels.
	// The bars are time series: the last ones are displayed when they don't all fit.
	barsHeight := ih - 1
	first := max(numValues-iw/(b.barGap+b.barWidth), 0)
	for i := first; i < numValues; i++ {
		bx := ix + (i-first)*(b.barWidth+b.barGap)

		stacked := 0
		for serie, d := range b.data {
			h := 0
			if maxValue > 0 {
				h = d[i] * barsHeight / maxValue
			}
			if h == 0 {
				continue
			}

			barStyle := fillStyle(b.color(serie))
			fill(s, bx, iy+barsHeight-stacked-h, b.barWidth, h, barStyle)

			num := displaywidth.TruncateString(strconv.Itoa(d[i]), b.barWidth, "")
			put(s, bx+(b.barWidth-displaywidth.String(num))/2, iy+barsHeight-stacked-1, num, b.barWidth, onBarStyle(b.color(serie), b.numColor))

			stacked += h
		}

		label := displaywidth.TruncateString(b.labels[i], b.barWidth, "")
		put(s, bx+(b.barWidth-displaywidth.String(label))/2, iy+ih-1, label, b.barWidth, style(b.textColor))
	}
}

// table with a separator between each row.
type table struct {
	block
	rows  [][]string
	style tcell.Style
}

func (t *table) height(int) int {
	return len(t.rows)*2 + 1
}

func (t *table) draw(s surface, x int, y int, width int) {
	ix, iy, iw, ih := t.block.draw(s, x, y, width, t.height(width))
	if iw <= 0 || ih <= 0 || len(t.rows) == 0 {
		return
	}

	cellWidths := t.cellWidths(iw)

	for r, cells := range t.rows {
		ry := iy + r*2
		cellStart := ix
		for c, cell := range cells {
			if c >= len(cellWidths) {
				break
			}
			if c != 0 {
				cellStart += cellWidths[c-1] + 3
				if cellStart >= ix+iw {
					break
				}
				s.PutStrStyled(cellStart, ry, "|", t.style)
			}
			put(s, cellStart+2, ry, displaywidth.TruncateString(cell, cellWidths[c], "…"), ix+iw-cellStart-2, t.style)
		}

		if r != len(t.rows)-1 {
			s.PutStrStyled(ix, ry+1, strings.Repeat("─", iw), t.style)
		}
	}
}

// cellWidths of the columns. When the table is too wide, the widest columns are shrunk to fit.
func (t *table) cellWidths(width int) []int {
	widths := make([]int, len(t.rows[0]))
	for _, r := range t.rows {
		for c, cell := range r {
			if c < len(widths) {
				widths[c] = max(widths[c], displaywidth.String(cell))
			}
		}
	}

	// Each column is preceded by 2 characters, the next ones by a separator too.
	total := 2 + 3*(len(widths)-1)
	for _, w := range widths {
		total += w
	}

	for ; total > width; total-- {
		widest := 0
		for c, w := range widths {
			if w > widths[widest] {
				widest = c
			}
		}
		if widths[widest] <= 4 {
			break
		}
		widths[widest]--
	}

	return widths
}
