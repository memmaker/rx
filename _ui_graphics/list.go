package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"

    "github.com/memmaker/go/geometry"
    "strings"
)

type IconListElement interface {
    Name() string
    GetTooltipLines(func(string) []ColoredTextPart) []LabelText
    InventoryIcon() int32
    TintColor() color.RGBA
    Shortcut() rune
    OnSelected()
    String() string
}

type Keylistener interface {
    OnKeyPressed(key ebiten.Key, shiftPressed bool)
}
type IconList struct {
    uiController    *UIController
    getElements     func() []IconListElement
    selectedIndex   int
    tileSize        geometry.Point
    needsScrolling  bool
    topLeft         geometry.Point
    borderPadding   int
    lineSpacing     int
    posInfo         PositionInfo
    shouldClose     bool
    currentElements []IconListElement
    iconScale       float64
    iconPadding     int
}

func (i *IconList) OnCommand(command CommandType) bool {
    if command == PlayerCommandCancel {
        i.Close()
    }
    return false
}
func (i *IconList) OnKeyPressed(key ebiten.Key, shiftPressed bool) {
    asRune := keyToInventoryRune(key, shiftPressed)

    defer i.Layout()

    for _, element := range i.currentElements {
        if element.Shortcut() == asRune {
            element.OnSelected()
            return
        }
    }
}

func (i *IconList) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
    if !i.contains(x, y) {
        return false
    }
    i.OnMouseMoved(x, y)
    if button == ebiten.MouseButtonLeft {
        if i.selectedIndex >= 0 && i.selectedIndex < len(i.currentElements) {
            i.currentElements[i.selectedIndex].OnSelected()
            return true
        }
    }
    return false
}

func (i *IconList) OnMouseMoved(x int, y int) (bool, Tooltip) {
    if !i.contains(x, y) {
        return false, EmptyTooltip{}
    }
    yStart := i.topLeft.Y + i.borderPadding
    yEnd := i.topLeft.Y + i.borderPadding + i.posInfo.ItemSizes[0].Y
    for index, element := range i.currentElements {

        // we only check the y coordinate, because the x coordinate is always the same
        if y >= yStart && y <= yEnd {
            i.selectedIndex = index
            return true, NewTextTooltip(i.uiController, element.GetTooltipLines(i.uiController.ParseColorCodedText), geometry.Point{X: x, Y: y})
        }

        yStart = yStart + int(i.lineSpacing) + i.posInfo.ItemSizes[index].Y
        yEnd = yEnd + int(i.lineSpacing) + i.posInfo.ItemSizes[index].Y
    }
    return false, EmptyTooltip{}
}

func (i *IconList) contains(x int, y int) bool {
    size := i.posInfo.SizeInPixels
    return x >= i.topLeft.X && x <= i.topLeft.X+size.X && y >= i.topLeft.Y && y <= i.topLeft.Y+size.Y
}

func (i *IconList) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (i *IconList) Draw() {
    renderer := i.uiController.GetRenderer()

    renderer.DrawDefaultBorder(i.topLeft, i.posInfo.SizeInPixels)

    itemDrawX := i.topLeft.X + i.borderPadding
    itemDrawY := i.topLeft.Y + i.borderPadding + i.posInfo.YOffsetForFirstLine

    for eIndex, element := range i.currentElements {
        icon := element.InventoryIcon()
        iconWidth := int(float64(i.tileSize.X) * i.iconScale)
        iconHeight := int(float64(i.tileSize.Y) * i.iconScale)
        iconWidthWithPadding := iconWidth + i.iconPadding

        renderer.DrawOnScreenWithScale(itemDrawX, itemDrawY-iconHeight, icon, i.iconScale)

        textDrawX := itemDrawX + iconWidthWithPadding
        drawColor := element.TintColor()
        if eIndex == i.selectedIndex {
            drawColor = defs.LightTextColor
        }
        renderer.DrawTTFOnScreen(float64(textDrawX), float64(itemDrawY), element.Name(), drawColor)
        itemDrawY += i.posInfo.ItemSizes[eIndex].Y + i.lineSpacing
    }
}

func (i *IconList) ShouldClose() bool {
    return i.shouldClose
}

func (i *IconList) Layout() {
    i.updateElements()
    iconWidthWithPadding := int(float64(i.tileSize.X)*i.iconScale) + i.iconPadding
    i.topLeft, i.posInfo = GetAutoFitRectWithExtraSpace(i.currentElements, i.uiController.GetScreenSize(), i.lineSpacing, geometry.Point{X: i.borderPadding*2 + iconWidthWithPadding, Y: i.borderPadding * 2}, i.uiController.GetRenderer().MeasureString)

    tileHeight := i.tileSize.Y
    lineHeight := i.posInfo.YOffsetForFirstLine
    // make the tileheight the same as the line height
    i.iconScale = float64(lineHeight) / float64(tileHeight)
}

func (i *IconList) updateElements() {
    i.currentElements = i.getElements()
    if len(i.currentElements) == 0 {
        i.Close()
    }
}

func (i *IconList) Close() {
    i.shouldClose = true
}

type ListDataSource struct {
    GetElements func() []IconListElement
}

func NewIconList(uiController *UIController, tileSize geometry.Point, dataSource ListDataSource) *IconList {
    i := &IconList{
        uiController:  uiController,
        getElements:   dataSource.GetElements,
        selectedIndex: -1,
        tileSize:      tileSize,
        borderPadding: 12,
        lineSpacing:   10.0,
        iconScale:     2,
        iconPadding:   6,
    }

    //i.infoLabel = NewCustomLabel(uiController, i.layoutLabel)
    i.Layout()
    return i
}

func keyToInventoryRune(key ebiten.Key, shiftPressed bool) rune {
    asString := key.String()
    if len(asString) != 1 {
        return 0
    }
    if shiftPressed {
        asString = strings.ToUpper(asString)
    } else {
        asString = strings.ToLower(asString)
    }
    asRune := rune(asString[0])
    return asRune
}
