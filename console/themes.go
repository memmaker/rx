package console

import (
	"RogueUI/foundation"
	"RogueUI/recfile"
	"RogueUI/util"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"image/color"
	"math/rand"
	"strings"
)

type ColorTheme map[string]color.RGBA

func (c ColorTheme) GetByName(name string) color.RGBA {
	name = strings.ToLower(name)
	return c[name]
}
func (c ColorTheme) GetAsFgColorCode(name string) string {
	name = strings.ToLower(name)
	return RGBAToFgColorCode(c[name])
}
func (c ColorTheme) GetAsBgColorCode(name string) string {
	name = strings.ToLower(name)
	return RGBAToBgColorCode(c[name])
}
func RGBAToFgColorCode(color color.RGBA) string {
	hexFormat := fmt.Sprintf("#%02x%02x%02x", color.R, color.G, color.B)
	return fmt.Sprintf("[%s]", hexFormat)
}
func RGBAToColorCodes(fg, bg color.RGBA) string {
	bgHex := fmt.Sprintf("#%02x%02x%02x", bg.R, bg.G, bg.B)
	fgHex := fmt.Sprintf("#%02x%02x%02x", fg.R, fg.G, fg.B)

	return fmt.Sprintf("[%s:%s]", fgHex, bgHex)
}
func RGBAToBgColorCode(color color.RGBA) string {
	hexFormat := fmt.Sprintf("#%02x%02x%02x", color.R, color.G, color.B)
	return fmt.Sprintf("[:%s]", hexFormat)
}

type Theme struct {
	colorDefs  ColorTheme
	playerIcon foundation.TextIcon

	uiColors map[UIColor]color.RGBA

	uiStyles            map[UIStyle]tcell.Style
	inventoryItemColors map[foundation.ItemCategory]color.RGBA

	uiBorder map[BorderCases]rune

	iconsForItems   map[foundation.ItemCategory]foundation.TextIcon
	iconsForObjects map[foundation.ObjectCategory]foundation.TextIcon
	iconsForMap     map[foundation.TileType]foundation.TextIcon
	defaultStyle    tcell.Style
	isMonoChrome    bool
}

func (t Theme) GetIconForItem(category foundation.ItemCategory) foundation.TextIcon {
	return t.iconsForItems[category]
}

func (t Theme) GetIconForMap(tileType foundation.TileType) foundation.TextIcon {
	return t.iconsForMap[tileType]
}

func (t Theme) GetIconForObject(object foundation.ObjectCategory) foundation.TextIcon {
	return t.iconsForObjects[object]
}

func (t Theme) GetInventoryItemColor(category foundation.ItemCategory) color.RGBA {
	return t.inventoryItemColors[category]
}

func NewThemeFromFile(filename string) Theme {
	file := util.MustOpen(filename)
	defer file.Close()
	records := recfile.ReadMulti(file)

	colors := loadColors(records["colors"][0])

	//uiColors, uiStyles, inventoryItemColors := loadUIStyles(records["ui"])

	inventoryItemColors := loadInventoryColors(records["inventory"][0], colors)
	uiBorders := loadBorders(records["borders"][0])
	uiColors := loadUIColors(records["ui"][0], colors)

	iconsForMap := loadIconsForMap(records["map"][0], colors)
	iconsForItems := loadIconsForItems(records["items"][0], colors)
	iconsForObjects := loadIconsForObjects(records["objects"][0], colors)

	var defaultStyle tcell.Style
	defaultStyle = defaultStyle.Foreground(toTcellColor(uiColors[UIColorUIForeground])).Background(toTcellColor(uiColors[UIColorUIBackground]))

	return Theme{
		colorDefs: colors,

		uiColors: uiColors,
		//uiStyles:            uiStyles,
		inventoryItemColors: inventoryItemColors,
		iconsForItems:       iconsForItems,
		iconsForObjects:     iconsForObjects,
		iconsForMap:         iconsForMap,
		uiBorder:            uiBorders,
		defaultStyle:        defaultStyle,
	}
}

func (t Theme) GetUIColor(foreground UIColor) color.RGBA {
	return t.uiColors[foreground]
}

func (t Theme) GetUIColorForTcell(foreground UIColor) tcell.Color {
	return toTcellColor(t.uiColors[foreground])
}

func loadUIColors(record recfile.Record, colors ColorTheme) map[UIColor]color.RGBA {
	uiColors := make(map[UIColor]color.RGBA)
	for _, field := range record {
		colorName := UIColorFromString(field.Name)
		colorValue := colors.GetByName(field.Value)
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

func loadInventoryColors(record recfile.Record, colors ColorTheme) map[foundation.ItemCategory]color.RGBA {
	inventoryColors := make(map[foundation.ItemCategory]color.RGBA)
	for _, field := range record {
		itemName := foundation.ItemCategoryFromString(field.Name)
		colorName := field.Value
		inventoryColors[itemName] = colors.GetByName(colorName)
	}
	return inventoryColors
}

func loadIconsForObjects(record recfile.Record, colors ColorTheme) map[foundation.ObjectCategory]foundation.TextIcon {
	icons := make(map[foundation.ObjectCategory]foundation.TextIcon)
	for _, field := range record {
		if strings.ContainsRune(field.Name, '_') {
			objectName, fgColor, bgColor := readColorField(field, colors)
			objectType := foundation.ObjectCategoryFromString(objectName)
			icons[objectType] = icons[objectType].WithColors(fgColor, bgColor)
		} else {
			objectType := foundation.ObjectCategoryFromString(field.Name)
			icons[objectType] = icons[objectType].WithRune([]rune(field.Value)[0])
		}
	}
	return icons

}

func loadIconsForItems(record recfile.Record, colors ColorTheme) map[foundation.ItemCategory]foundation.TextIcon {
	icons := make(map[foundation.ItemCategory]foundation.TextIcon)
	for _, field := range record {
		if strings.ContainsRune(field.Name, '_') {
			itemName, fgColor, bgColor := readColorField(field, colors)
			itemType := foundation.ItemCategoryFromString(itemName)
			icons[itemType] = icons[itemType].WithColors(fgColor, bgColor)
		} else {
			itemType := foundation.ItemCategoryFromString(field.Name)
			icons[itemType] = icons[itemType].WithRune([]rune(field.Value)[0])
		}
	}
	return icons

}

func loadIconsForMap(record recfile.Record, colors ColorTheme) map[foundation.TileType]foundation.TextIcon {
	icons := make(map[foundation.TileType]foundation.TextIcon)
	for _, field := range record {
		if strings.ContainsRune(field.Name, '_') {
			tileName, fgColor, bgColor := readColorField(field, colors)
			tileType := foundation.TileType(tileName)
			icons[tileType] = icons[tileType].WithColors(fgColor, bgColor)
		} else {
			tileType := foundation.TileType(field.Name)
			icons[tileType] = icons[tileType].WithRune([]rune(field.Value)[0])
		}
	}
	return icons

}

func readColorField(field recfile.Field, colors ColorTheme) (string, color.RGBA, color.RGBA) {
	tileName := strings.Split(field.Name, "_")[0]
	tileType := tileName
	// color def
	colorNames := field.AsList("|")
	fgColor := colors.GetByName(colorNames[0].Value)
	bgColor := colors.GetByName(colorNames[1].Value)
	return tileType, fgColor, bgColor
}

func loadColors(record recfile.Record) ColorTheme {
	colors := make(map[string]color.RGBA)
	for _, field := range record {
		colorName := strings.ToLower(field.Name) // case insensitive
		colorValue := field.AsRGB("|")
		colors[colorName] = colorValue
	}
	return colors
}

type UIStyle int

const (
	UIStyleNormal UIStyle = iota
	UIStyleHighlighted
	UIStyleSelected
	UIStyleBorder
	UIStyleBorderFocused
)

type UIColor int

const (
	UIColorMapDefaultBackground UIColor = iota
	UIColorMapDefaultForeground
	UIColorMapDefaultForegroundLight
	UIColorMapDefaultForegroundDark
	UIColorUIBackground
	UIColorUIForeground
	UIColorBorderBackground
	UIColorBorderForeground

	UIColorTextForegroundHighlighted
)

func UIColorFromString(s string) UIColor {
	s = strings.ToLower(s)
	switch s {
	case "mapdefaultbackground":
		return UIColorMapDefaultBackground
	case "mapdefaultforeground":
		return UIColorMapDefaultForeground
	case "mapdefaultforegroundlight":
		return UIColorMapDefaultForegroundLight
	case "mapdefaultforegrounddark":
		return UIColorMapDefaultForegroundDark
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
	return UIColorMapDefaultBackground
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

func (t Theme) SetBorders(s *struct {
	Horizontal       rune
	Vertical         rune
	TopLeft          rune
	TopRight         rune
	BottomLeft       rune
	BottomRight      rune
	LeftT            rune
	RightT           rune
	TopT             rune
	BottomT          rune
	Cross            rune
	HorizontalFocus  rune
	VerticalFocus    rune
	TopLeftFocus     rune
	TopRightFocus    rune
	BottomLeftFocus  rune
	BottomRightFocus rune
}) {
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

func (t Theme) GetColorByName(colorName string) color.RGBA {
	return t.colorDefs.GetByName(colorName)
}

func (t Theme) GetMapDefaultStyle() tcell.Style {
	return t.defaultStyle
}

func (t Theme) IsMonochrome() bool {
	return t.isMonoChrome
}

func (t Theme) GetRandomColor() color.RGBA {
	var colors []color.RGBA
	for _, c := range t.colorDefs {
		colors = append(colors, c)
	}
	return colors[rand.Intn(len(colors))]
}
