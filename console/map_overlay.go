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
	o.PrintColored(x, y, text, o.defaultForeground, o.defaultBackground)
}

func (o *Overlay) PrintColored(x, y int, text string, fg, bg color.RGBA) {
	for i, r := range []rune(text) {
		if x+i >= o.width {
			break
		}
		o.grid[y*o.width+x+i] = textiles.TextIcon{Char: r, Fg: fg, Bg: bg}
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

func (o *Overlay) asciiLine(origin geometry.Point, dest geometry.Point, steps []geometry.Point, fg, bg color.RGBA) {
	if origin.X == dest.X || len(steps) == 0 {
		return
	}

	horzRune := textiles.TextIcon{Char: '-', Fg: fg, Bg: bg}
	vertRune := textiles.TextIcon{Char: '|', Fg: fg, Bg: bg}
	blToTrRune := textiles.TextIcon{Char: '/', Fg: fg, Bg: bg}
	tlToBrRune := textiles.TextIcon{Char: '\\', Fg: fg, Bg: bg}
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
func (o *Overlay) TryAddOverlay(sourcePos geometry.Point, text string, windowSize geometry.Point, isMapFeatureAt func(position geometry.Point) bool) bool {
	pos, connectors := o.calculateOverlayPos(windowSize, sourcePos, len(text), isMapFeatureAt)
	if pos == sourcePos {
		return false
	}
	o.Print(pos.X, pos.Y, text)
	o.asciiLine(sourcePos, pos, connectors, o.defaultForeground, o.defaultBackground)
	return true
}

func (o *Overlay) TryAddOverlayColored(pos geometry.Point, text string, textColor color.RGBA, size geometry.Point, loc func(position geometry.Point) bool) bool {
	labelPos, connectors := o.calculateOverlayPos(size, pos, len(text), loc)
	if labelPos == pos {
		return false
	}
	bgColor := color.RGBA{R: 216, G: 216, B: 216, A: 255} // lightish
	if isMoreLightThanDark(textColor) {
		bgColor = color.RGBA{R: 40, G: 40, B: 40, A: 255} // darkish
	}
	o.PrintColored(labelPos.X, labelPos.Y, text, textColor, bgColor)
	o.asciiLine(pos, labelPos, connectors, textColor, bgColor)
	return true
}

func isMoreLightThanDark(textColor color.RGBA) bool {
	return int(textColor.R)+int(textColor.G)+int(textColor.B) > 383
}

func (o *Overlay) calculateOverlayPos(windowSize, position geometry.Point, widthNeeded int, isMapFeatureAt func(point geometry.Point) bool) (labelPos geometry.Point, connectors []geometry.Point) {
	sW, sH := windowSize.X, windowSize.Y
	locIsBlocked := func(pos geometry.Point) bool {
		return isMapFeatureAt(pos) || o.IsSet(pos.X, pos.Y)
	}
	isPosForLabelValid := func(pos geometry.Point) bool {
		if pos.X < 0 || pos.Y < 0 || pos.X+widthNeeded >= sW || pos.Y >= sH {
			return false
		}
		for x := 0; x < widthNeeded; x++ {
			curPos := geometry.Point{X: pos.X + x, Y: pos.Y}
			if locIsBlocked(curPos) {
				return false
			}
		}
		return true
	}

	centeredAbove := position.Add(geometry.Point{X: -widthNeeded / 2, Y: -1})
	if isPosForLabelValid(centeredAbove) {
		return centeredAbove, nil
	}

	simpleRightConnector := position.Add(geometry.Point{X: 1, Y: 0})
	simpleRightLabelPos := position.Add(geometry.Point{X: 2, Y: 0})
	if isPosForLabelValid(simpleRightLabelPos) && !locIsBlocked(simpleRightConnector) {
		return simpleRightLabelPos, []geometry.Point{simpleRightConnector}
	}

	simpleLeftConnector := position.Add(geometry.Point{X: -1, Y: 0})
	simpleLeftLabelPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: 0})
	if isPosForLabelValid(simpleLeftLabelPos) && !locIsBlocked(simpleLeftConnector) {
		return simpleLeftLabelPos, []geometry.Point{simpleLeftConnector}
	}

	topRightConnector := position.Add(geometry.Point{X: 1, Y: -1})
	topRightLabelPos := position.Add(geometry.Point{X: 2, Y: -1})
	if isPosForLabelValid(topRightLabelPos) && !locIsBlocked(topRightConnector) {
		return topRightLabelPos, []geometry.Point{topRightConnector}
	}

	topLeftConnector := position.Add(geometry.Point{X: -1, Y: -1})
	topLeftLabelPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: -1})
	if isPosForLabelValid(topLeftLabelPos) && !locIsBlocked(topLeftConnector) {
		return topLeftLabelPos, []geometry.Point{topLeftConnector}
	}

	bottomRightConnector := position.Add(geometry.Point{X: 1, Y: 1})
	bottomRightLabelPos := position.Add(geometry.Point{X: 2, Y: 1})
	if isPosForLabelValid(bottomRightLabelPos) && !locIsBlocked(bottomRightConnector) {
		return bottomRightLabelPos, []geometry.Point{bottomRightConnector}
	}

	bottomLeftConnector := position.Add(geometry.Point{X: -1, Y: 1})
	bottomLeftLabelPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: 1})
	if isPosForLabelValid(bottomLeftLabelPos) && !locIsBlocked(bottomLeftConnector) {
		return bottomLeftLabelPos, []geometry.Point{bottomLeftConnector}
	}

	twoDownConnector := position.Add(geometry.Point{X: 0, Y: 2})
	twoDownLabelRightPos := position.Add(geometry.Point{X: 1, Y: 2})
	if isPosForLabelValid(twoDownLabelRightPos) && !locIsBlocked(twoDownConnector) {
		return twoDownLabelRightPos, []geometry.Point{twoDownConnector}
	}

	twoDownLabelLeftPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: 2})
	if isPosForLabelValid(twoDownLabelLeftPos) && !locIsBlocked(twoDownConnector) {
		return twoDownLabelLeftPos, []geometry.Point{twoDownConnector}
	}

	twoUpConnector := position.Add(geometry.Point{X: 0, Y: -2})
	twoUpLabelRightPos := position.Add(geometry.Point{X: 1, Y: -2})
	if isPosForLabelValid(twoUpLabelRightPos) && !locIsBlocked(twoUpConnector) {
		return twoUpLabelRightPos, []geometry.Point{twoUpConnector}
	}

	twoUpLabelLeftPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: -2})
	if isPosForLabelValid(twoUpLabelLeftPos) && !locIsBlocked(twoUpConnector) {
		return twoUpLabelLeftPos, []geometry.Point{twoUpConnector}
	}

	return position, nil
}

func (o *Overlay) AddAbove(pos geometry.Point, cthString string) {
	xOffset := len(cthString) / 2
	labelPos := pos.Add(geometry.Point{X: -xOffset, Y: -1})
	o.Print(labelPos.X, labelPos.Y, cthString)
}

func (o *Overlay) AddBelow(pos geometry.Point, text string) {
	xOffset := len(text) / 2
	labelPos := pos.Add(geometry.Point{X: -xOffset, Y: 1})
	o.Print(labelPos.X, labelPos.Y, text)
}
