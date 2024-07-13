package cview

import (
	"RogueUI/geometry"
	"RogueUI/util"
	"github.com/gdamore/tcell/v2"
)

// Borders defines various borders used when primitives are drawn.
// These may be changed to accommodate a different look and feel.
type BorderDef struct {
	Horizontal  rune
	Vertical    rune
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune

	LeftT   rune
	RightT  rune
	TopT    rune
	BottomT rune
	Cross   rune

	HorizontalFocus  rune
	VerticalFocus    rune
	TopLeftFocus     rune
	TopRightFocus    rune
	BottomLeftFocus  rune
	BottomRightFocus rune
}

var Borders = BorderDef{
	Horizontal:  BoxDrawingsLightHorizontal,
	Vertical:    BoxDrawingsLightVertical,
	TopLeft:     BoxDrawingsLightDownAndRight,
	TopRight:    BoxDrawingsLightDownAndLeft,
	BottomLeft:  BoxDrawingsLightUpAndRight,
	BottomRight: BoxDrawingsLightUpAndLeft,

	LeftT:   BoxDrawingsLightVerticalAndRight,
	RightT:  BoxDrawingsLightVerticalAndLeft,
	TopT:    BoxDrawingsLightDownAndHorizontal,
	BottomT: BoxDrawingsLightUpAndHorizontal,
	Cross:   BoxDrawingsLightVerticalAndHorizontal,

	HorizontalFocus:  BoxDrawingsDoubleHorizontal,
	VerticalFocus:    BoxDrawingsDoubleVertical,
	TopLeftFocus:     BoxDrawingsDoubleDownAndRight,
	TopRightFocus:    BoxDrawingsDoubleDownAndLeft,
	BottomLeftFocus:  BoxDrawingsDoubleUpAndRight,
	BottomRightFocus: BoxDrawingsDoubleUpAndLeft,
}

func DrawCircle(s tcell.Screen, x, y, radius int, fill rune, style tcell.Style) {
	util.DrawCircleSqrt(x, y, radius, func(x, y int) {
		s.SetContent(x, y, fill, nil, style)
	})
}
func DrawCircleSegmented(s tcell.Screen, x, y, radius int, angle float64, fill rune, style tcell.Style) {
	util.DrawCircleSqrtSegmented(x, y, radius, angle, func(x, y int) {
		s.SetContent(x, y, fill, nil, style)
	})
}
func DrawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, borders BorderDef, backgroundFill rune) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, borders.Horizontal, nil, style)
		s.SetContent(col, y2, borders.Horizontal, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, borders.Vertical, nil, style)
		s.SetContent(x2, row, borders.Vertical, nil, style)
	}
	if y1 != y2 && x1 != x2 {
		// Only add corners if we need to
		s.SetContent(x1, y1, borders.TopLeft, nil, style)
		s.SetContent(x2, y1, borders.TopRight, nil, style)
		s.SetContent(x1, y2, borders.BottomLeft, nil, style)
		s.SetContent(x2, y2, borders.BottomRight, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		for col := x1 + 1; col < x2; col++ {
			s.SetContent(col, row, backgroundFill, nil, style)
		}
	}
}

func DrawLine(s tcell.Screen, x1, y1, x2, y2 int, r rune, style tcell.Style) {
	line := geometry.BresenhamLine(geometry.Point{x1, y1}, geometry.Point{x2, y2}, func(x, y int) bool {
		return true
	})
	for _, p := range line {
		s.SetContent(p.X, p.Y, r, nil, style)
	}
}
