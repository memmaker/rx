package console

import (
	"RogueUI/foundation"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"path"
	"strings"
)

type Theme struct {
	uiColors map[UIColor]color.RGBA

	uiBorder     map[BorderCases]rune
	defaultStyle tcell.Style

	palette textiles.ColorPalette

	inventoryColors map[foundation.ItemCategory]color.RGBA
}

func NewUIThemeFromDataDir(dataDirectory string, palette textiles.ColorPalette, inventory map[foundation.ItemCategory]color.RGBA) Theme {
	uiThemeFile := path.Join(dataDirectory, "themes", "ui.rec")
	file := fxtools.MustOpen(uiThemeFile)
	defer file.Close()
	records := recfile.ReadMulti(file)

	uiBorders := loadBorders(records["borders"][0])
	uiColors := loadUIColors(records["ui"][0])

	var defaultStyle tcell.Style
	defaultStyle = defaultStyle.Foreground(toTcellColor(uiColors[UIColorUIForeground])).Background(toTcellColor(uiColors[UIColorUIBackground]))

	return Theme{
		palette:         palette,
		uiColors:        uiColors,
		uiBorder:        uiBorders,
		defaultStyle:    defaultStyle,
		inventoryColors: inventory,
	}
}

func (t Theme) GetUIColor(foreground UIColor) color.RGBA {
	return t.uiColors[foreground]
}

func (t Theme) GetUIColorForTcell(foreground UIColor) tcell.Color {
	return toTcellColor(t.uiColors[foreground])
}

func loadUIColors(record recfile.Record) map[UIColor]color.RGBA {
	uiColors := make(map[UIColor]color.RGBA)
	for _, field := range record {
		colorName := UIColorFromString(field.Name)
		colorValue := field.AsRGB("|")
		uiColors[colorName] = colorValue
	}
	return uiColors

}

func loadBorders(record recfile.Record) map[BorderCases]rune {
	borders := make(map[BorderCases]rune)
	for _, field := range record {
		borderCase := BorderCaseFromString(strings.TrimPrefix(field.Name, "UIBorder_"))
		borders[borderCase] = field.AsRune()
	}
	return borders

}

type UIColor int

const (
	UIColorUIBackground UIColor = iota
	UIColorUIForeground
	UIColorBorderBackground
	UIColorBorderForeground

	UIColorTextForegroundHighlighted
)

func UIColorFromString(s string) UIColor {
	s = strings.ToLower(s)
	switch s {
	case "uiforeground":
		return UIColorUIForeground
	case "uiforegroundhighlighted":
		return UIColorTextForegroundHighlighted
	case "uibackground":
		return UIColorUIBackground
	case "borderbackground":
		return UIColorBorderBackground
	case "borderforeground":
		return UIColorBorderForeground
	}
	println("WARNING: Unknown color: ", s)
	return UIColorUIForeground
}

type BorderCases int

const (
	BorderHorizontal BorderCases = iota
	BorderVertical
	BorderTopLeft
	BorderTopRight
	BorderBottomLeft
	BorderBottomRight
	BorderLeftT
	BorderRightT
	BorderTopT
	BorderBottomT
	BorderCross
	BorderHorizontalFocus
	BorderVerticalFocus
	BorderTopLeftFocus
	BorderTopRightFocus
	BorderBottomLeftFocus
	BorderBottomRightFocus
)

func BorderCaseFromString(s string) BorderCases {
	s = strings.ToLower(s)
	switch s {
	case "horizontal":
		return BorderHorizontal
	case "vertical":
		return BorderVertical
	case "topleft":
		return BorderTopLeft
	case "topright":
		return BorderTopRight
	case "bottomleft":
		return BorderBottomLeft
	case "bottomright":
		return BorderBottomRight
	case "leftt":
		return BorderLeftT
	case "rightt":
		return BorderRightT
	case "topt":
		return BorderTopT
	case "bottomt":
		return BorderBottomT
	case "cross":
		return BorderCross
	case "horizontalfocus":
		return BorderHorizontalFocus
	case "verticalfocus":
		return BorderVerticalFocus
	case "topleftfocus":
		return BorderTopLeftFocus
	case "toprightfocus":
		return BorderTopRightFocus
	case "bottomleftfocus":
		return BorderBottomLeftFocus
	case "bottomrightfocus":
		return BorderBottomRightFocus
	}
	println("WARNING: Unknown border case: ", s)
	return BorderHorizontal
}

func (t Theme) SetBorders(s *cview.BorderDef) {
	s.Horizontal = t.uiBorder[BorderHorizontal]
	s.Vertical = t.uiBorder[BorderVertical]
	s.TopLeft = t.uiBorder[BorderTopLeft]
	s.TopRight = t.uiBorder[BorderTopRight]
	s.BottomLeft = t.uiBorder[BorderBottomLeft]
	s.BottomRight = t.uiBorder[BorderBottomRight]
	s.LeftT = t.uiBorder[BorderLeftT]
	s.RightT = t.uiBorder[BorderRightT]
	s.TopT = t.uiBorder[BorderTopT]
	s.BottomT = t.uiBorder[BorderBottomT]
	s.Cross = t.uiBorder[BorderCross]
	s.HorizontalFocus = t.uiBorder[BorderHorizontalFocus]
	s.VerticalFocus = t.uiBorder[BorderVerticalFocus]
	s.TopLeftFocus = t.uiBorder[BorderTopLeftFocus]
	s.TopRightFocus = t.uiBorder[BorderTopRightFocus]
	s.BottomLeftFocus = t.uiBorder[BorderBottomLeftFocus]
	s.BottomRightFocus = t.uiBorder[BorderBottomRightFocus]
}

func (t Theme) GetMapDefaultStyle() tcell.Style {
	return t.defaultStyle
}

func (t Theme) GetInventoryItemColorCode(category foundation.ItemCategory) string {
	return textiles.RGBAToFgColorCode(t.inventoryColors[category])
}

func (t Theme) GetColorByName(name string) color.RGBA {
	return t.palette.Get(name)
}

func (t Theme) GetRandomColor() color.RGBA {
	return t.palette.GetRandomColor()
}

func (t Theme) GetInventoryItemColor(category foundation.ItemCategory) color.RGBA {
	return t.inventoryColors[category]
}

func (t Theme) WithInventoryColors(colors map[foundation.ItemCategory]color.RGBA) Theme {
	t.inventoryColors = colors
	return t
}
