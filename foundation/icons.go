package foundation

import (
	"image/color"
)

type TextIcon struct {
	Rune   rune
	Fg, Bg color.RGBA
}

func (i TextIcon) Reversed() TextIcon {
	return TextIcon{i.Rune, i.Bg, i.Fg}
}

func (i TextIcon) WithFg(newFg color.RGBA) TextIcon {
	return TextIcon{i.Rune, newFg, i.Bg}
}

func (i TextIcon) WithBg(newBg color.RGBA) TextIcon {
	return TextIcon{i.Rune, i.Fg, newBg}
}

func (i TextIcon) WithColors(fgColor color.RGBA, bgColor color.RGBA) TextIcon {
	return TextIcon{i.Rune, fgColor, bgColor}
}

func (i TextIcon) WithRune(r rune) TextIcon {
	return TextIcon{r, i.Fg, i.Bg}
}

