package foundation

import (
	"github.com/gdamore/tcell/v2"
	"image/color"
)

type TextIcon struct {
	Rune       rune
	Fg, Bg     color.RGBA
	Attributes tcell.AttrMask
}

func (i TextIcon) Reversed() TextIcon {
	return TextIcon{i.Rune, i.Bg, i.Fg, i.Attributes}
}
func (i TextIcon) WithFg(newFg color.RGBA) TextIcon {
	return TextIcon{i.Rune, newFg, i.Bg, i.Attributes}
}

func (i TextIcon) WithBg(newBg color.RGBA) TextIcon {
	return TextIcon{i.Rune, i.Fg, newBg, i.Attributes}
}

func (i TextIcon) WithColors(fgColor color.RGBA, bgColor color.RGBA) TextIcon {
	return TextIcon{i.Rune, fgColor, bgColor, i.Attributes}
}

func (i TextIcon) WithRune(r rune) TextIcon {
	return TextIcon{r, i.Fg, i.Bg, i.Attributes}
}

func (i TextIcon) WithItalic() TextIcon {
	i.Attributes |= tcell.AttrItalic
	return i
}

func (i TextIcon) WithBold() TextIcon {
	i.Attributes |= tcell.AttrBold
	return i
}
