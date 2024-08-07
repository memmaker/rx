package console

import (
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
)

type Overlay struct {
	width             int
	height            int
	grid              []textiles.TextIcon
	defaultBackground color.RGBA
	defaultForeground color.RGBA
}

func NewOverlay(width int, height int) *Overlay {
	grid := make([]textiles.TextIcon, width*height)
	o := &Overlay{width: width, height: height, grid: grid}
	o.ClearAll()
	return o
}

func (o *Overlay) SetDefaultColors(fg, bg color.RGBA) {
	o.defaultForeground = fg
	o.defaultBackground = bg
}

func (o *Overlay) Set(x, y int, icon textiles.TextIcon) {
	o.grid[y*o.width+x] = icon
}

func (o *Overlay) ClearAll() {
	for i := range o.grid {
		o.grid[i] = textiles.TextIcon{
			Char: -1,
		}
	}
}

func (o *Overlay) Print(x, y int, text string) {
	for i, r := range []rune(text) {
		if x+i >= o.width {
			break
		}
		o.grid[y*o.width+x+i] = textiles.TextIcon{Char: r, Fg: o.defaultForeground, Bg: o.defaultBackground}
	}
}

func (o *Overlay) Contains(x, y int) bool {
	return x >= 0 && x < o.width && y >= 0 && y < o.height
}
func (o *Overlay) IsSet(x, y int) bool {
	if !o.Contains(x, y) {
		return false
	}
	return o.grid[y*o.width+x].Char != -1
}
func (o *Overlay) Get(x, y int) textiles.TextIcon {
	return o.grid[y*o.width+x]
}

func (o *Overlay) AsciiLine(origin geometry.Point, dest geometry.Point, steps []geometry.Point) {
	if origin.X == dest.X || len(steps) == 0 {
		return
	}

	horzRune := textiles.TextIcon{Char: '-', Fg: o.defaultForeground, Bg: o.defaultBackground}
	vertRune := textiles.TextIcon{Char: '|', Fg: o.defaultForeground, Bg: o.defaultBackground}
	blToTrRune := textiles.TextIcon{Char: '/', Fg: o.defaultForeground, Bg: o.defaultBackground}
	tlToBrRune := textiles.TextIcon{Char: '\\', Fg: o.defaultForeground, Bg: o.defaultBackground}
	directionToLineRune := map[geometry.Point]textiles.TextIcon{
		geometry.Point{1, 0}:   horzRune,
		geometry.Point{-1, 0}:  horzRune,
		geometry.Point{0, 1}:   vertRune,
		geometry.Point{0, -1}:  vertRune,
		geometry.Point{1, 1}:   tlToBrRune,
		geometry.Point{-1, -1}: tlToBrRune,
		geometry.Point{1, -1}:  blToTrRune,
		geometry.Point{-1, 1}:  blToTrRune,
	}

	prev := origin
	for _, step := range steps {
		dir := step.Sub(prev)
		lineRune := directionToLineRune[dir]
		o.Set(step.X, step.Y, lineRune)
		prev = step
	}
}
