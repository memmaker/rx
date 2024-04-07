package console

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"image/color"
)

type Overlay struct {
	width             int
	height            int
	grid              []foundation.TextIcon
	defaultBackground color.RGBA
	defaultForeground color.RGBA
}

func NewOverlay(width int, height int) *Overlay {
	grid := make([]foundation.TextIcon, width*height)
	o := &Overlay{width: width, height: height, grid: grid}
	o.ClearAll()
	return o
}

func (o *Overlay) SetDefaultColors(fg, bg color.RGBA) {
	o.defaultForeground = fg
	o.defaultBackground = bg
}

func (o *Overlay) Set(x, y int, icon foundation.TextIcon) {
	o.grid[y*o.width+x] = icon
}

func (o *Overlay) ClearAll() {
	for i := range o.grid {
		o.grid[i] = foundation.TextIcon{
			Rune: -1,
		}
	}
}

func (o *Overlay) Print(x, y int, text string) {
	for i, r := range []rune(text) {
		if x+i >= o.width {
			break
		}
		o.grid[y*o.width+x+i] = foundation.TextIcon{Rune: r, Fg: o.defaultForeground, Bg: o.defaultBackground}
	}
}

func (o *Overlay) Contains(x, y int) bool {
	return x >= 0 && x < o.width && y >= 0 && y < o.height
}
func (o *Overlay) IsSet(x, y int) bool {
	if !o.Contains(x, y) {
		return false
	}
	return o.grid[y*o.width+x].Rune != -1
}
func (o *Overlay) Get(x, y int) foundation.TextIcon {
	return o.grid[y*o.width+x]
}

func (o *Overlay) AsciiLine(origin geometry.Point, dest geometry.Point, steps []geometry.Point) {
	if origin.X == dest.X || len(steps) == 0 {
		return
	}

	horzRune := foundation.TextIcon{Rune: '-', Fg: o.defaultForeground, Bg: o.defaultBackground}
	vertRune := foundation.TextIcon{Rune: '|', Fg: o.defaultForeground, Bg: o.defaultBackground}
	blToTrRune := foundation.TextIcon{Rune: '/', Fg: o.defaultForeground, Bg: o.defaultBackground}
	tlToBrRune := foundation.TextIcon{Rune: '\\', Fg: o.defaultForeground, Bg: o.defaultBackground}
	directionToLineRune := map[geometry.Point]foundation.TextIcon{
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
