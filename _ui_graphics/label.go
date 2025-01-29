package ui_graphics

import (
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
    "image/color"
    "strings"
)

type BasicLabel struct {
    uiController   *UIController
    topLeft        geometry.Point
    textLines      []LabelText
    posInfo        PositionInfo
    centered       bool
    layoutFunc     func()
    isTooltip      bool
    drawBackground bool
    shouldClose    bool
    borderPadding  geometry.Point
    bounds         geometry.Rect
    canBeClosed    bool
}

func (i *BasicLabel) OnCommand(command CommandType) bool {
    switch command {
    case PlayerCommandConfirm:
        fallthrough
    case PlayerCommandCancel:
        i.shouldClose = true
        return true

    }
    return false
}

func (i *BasicLabel) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
    mousePos := geometry.Point{X: x, Y: y}
    if i.canBeClosed && (!i.bounds.Contains(mousePos) || button == ebiten.MouseButtonRight) {
        i.shouldClose = true
        return true
    }
    return false
}

type Tooltip interface {
    IsNull() bool
    Draw()
}

type EmptyTooltip struct {
}

func (n EmptyTooltip) Draw() {

}

func (n EmptyTooltip) IsNull() bool {
    return true
}

func (i *BasicLabel) OnMouseMoved(x int, y int) (bool, Tooltip) {
    return false, EmptyTooltip{}
}

func (i *BasicLabel) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (i *BasicLabel) ShouldClose() bool {
    return i.shouldClose
}

func (i *BasicLabel) Layout() {
    i.layoutFunc()
}

func (i *BasicLabel) Draw() {
    uiRenderer := i.uiController.GetRenderer()
    if i.drawBackground {
        borderTopLeft := i.topLeft.Sub(i.borderPadding.Div(2))
        borderSize := i.posInfo.SizeInPixels.Add(i.borderPadding).Add(geometry.Point{X: 4, Y: 4})
        uiRenderer.DrawDefaultBorder(borderTopLeft, borderSize)
    }
    i.drawStringArray(i.textLines)
}

func (i *BasicLabel) IsNull() bool {
    return len(i.textLines) == 0
}

func (i *BasicLabel) drawStringArray(lines []LabelText) {
    if len(lines) == 0 {
        return
    }
    uiRenderer := i.uiController.GetRenderer()
    //startPos := geometry.Point{X: i.topLeft.X + 2, Y: startAtY}
    //tWidth, tHeight := uiRenderer.MeasureString(linesAsOneString)
    drawY := i.topLeft.Y + i.posInfo.YOffsetForFirstLine + int(2.0)
    for lineIndex, line := range lines {
        if line.Text == "" {
            drawY += i.posInfo.YOffsetForFirstLine + i.posInfo.LineDist
            continue
        }
        itemSize := i.posInfo.ItemSizes[lineIndex]
        sizeX, sizeY := itemSize.X, itemSize.Y
        drawX := i.topLeft.X
        if i.centered {
            drawX = i.topLeft.X + (i.posInfo.SizeInPixels.X-sizeX)/2
        }
        if line.UseColorCodes {
            uiRenderer.DrawTTFOnScreenWithColorCodes(float64(drawX), float64(drawY), line.TextWithColorCodes)
        } else {
            drawColor := line.TextColor
            if drawColor == nil {
                drawColor = color.White
            }
            uiRenderer.DrawTTFOnScreen(float64(drawX), float64(drawY), line.Text, drawColor)
        }
        drawY += sizeY + i.posInfo.LineDist
    }

}

func (i *BasicLabel) SetText(name []LabelText) {
    i.textLines = name
    i.Layout()
}

type PositionInfo struct {
    SizeInPixels        geometry.Point
    YOffsetForFirstLine int
    LineDist            int
    ItemSizes           []geometry.Point
}

func (i PositionInfo) ToString() string {
    return fmt.Sprintf("SizeInPixels: %v, YOffsetForFirstLine: %v", i.SizeInPixels, i.YOffsetForFirstLine)
}

func SizeNeededForText[T fmt.Stringer](measureString func(line string) (float64, float64), text []T, lineDist int) PositionInfo {
    maxLength := 0
    totalHeight := 0
    yOffFirstLine := 0
    //previousHeight := 0
    var itemSizes []geometry.Point
    for i, line := range text {
        w, h := measureString(line.String())
        width := int(w)
        height := int(h)
        itemSizes = append(itemSizes, geometry.Point{X: width, Y: height})
        if strings.TrimSpace(line.String()) == "" {
            height = yOffFirstLine
        }
        if i == 0 {
            yOffFirstLine = height
        }
        if width > maxLength {
            maxLength = width
        }
        totalHeight += height
        //previousHeight = height
    }
    totalHeight += lineDist * (len(text) - 1)
    return PositionInfo{
        SizeInPixels:        geometry.Point{X: maxLength, Y: totalHeight},
        YOffsetForFirstLine: yOffFirstLine,
        LineDist:            lineDist,
        ItemSizes:           itemSizes,
    }
}
func NewTextLabel(uiController *UIController, topLeftPositionAsOffsetFromTopRight geometry.Point) *BasicLabel {
    b := &BasicLabel{
        uiController:   uiController,
        centered:       false,
        drawBackground: false,
    }
    b.layoutFunc = func() {
        b.posInfo = SizeNeededForText(b.uiController.GetRenderer().MeasureString, b.textLines, 10)
        b.layoutFromTopLeftPos(topLeftPositionAsOffsetFromTopRight)
        b.bounds = geometry.NewRect(b.topLeft.X, b.topLeft.Y, b.topLeft.X+b.posInfo.SizeInPixels.X, b.topLeft.Y+b.posInfo.SizeInPixels.Y)
    }
    return b
}
func NewCenteredTextLabel(uiController *UIController) *BasicLabel {
    b := &BasicLabel{
        uiController:   uiController,
        centered:       false,
        drawBackground: true,
    }
    b.layoutFunc = func() {
        b.posInfo = SizeNeededForText(b.uiController.GetRenderer().MeasureString, b.textLines, 10)
        b.layoutCentered()
        b.bounds = geometry.NewRect(b.topLeft.X, b.topLeft.Y, b.topLeft.X+b.posInfo.SizeInPixels.X, b.topLeft.Y+b.posInfo.SizeInPixels.Y)
    }
    return b
}
func NewCustomLabel(uiController *UIController, layoutFunc func(uiController *UIController, label *BasicLabel)) *BasicLabel {
    b := &BasicLabel{
        uiController:   uiController,
        centered:       false,
        drawBackground: false,
    }
    b.layoutFunc = func() {
        layoutFunc(uiController, b)
        b.bounds = geometry.NewRect(b.topLeft.X, b.topLeft.Y, b.topLeft.X+b.posInfo.SizeInPixels.X, b.topLeft.Y+b.posInfo.SizeInPixels.Y)
    }
    return b
}
func NewTextTooltip(uiController *UIController, textLines []LabelText, mousePosOnScreen geometry.Point) *BasicLabel {
    b := &BasicLabel{
        uiController:   uiController,
        textLines:      textLines,
        centered:       true,
        isTooltip:      true,
        drawBackground: true,
    }
    b.layoutFunc = func() {
        b.posInfo = SizeNeededForText(b.uiController.GetRenderer().MeasureString, b.textLines, 10)
        b.layoutFromMousePos(mousePosOnScreen)
        b.bounds = geometry.NewRect(b.topLeft.X, b.topLeft.Y, b.topLeft.X+b.posInfo.SizeInPixels.X, b.topLeft.Y+b.posInfo.SizeInPixels.Y)
    }
    b.SetBorderPadding(geometry.Point{X: 15, Y: 15})
    b.Layout()
    return b
}
func (i *BasicLabel) layoutCentered() {
    screenSize := i.uiController.GetScreenSize()
    topLeft := geometry.Point{X: (screenSize.X - i.posInfo.SizeInPixels.X) / 2, Y: (screenSize.Y - i.posInfo.SizeInPixels.Y) / 2}
    i.topLeft = topLeft
}

func (i *BasicLabel) layoutFromMousePos(mousePosOnScreen geometry.Point) {
    var topLeftX, topLeftY int
    // we want the tooltip's bottom right corner to be at the mousePos
    posInfo := i.posInfo
    uiController := i.uiController

    paddingX := uiController.GetUIBarWidth() + 20
    paddingY := 50
    topLeftX = mousePosOnScreen.X - posInfo.SizeInPixels.X/2
    topLeftY = mousePosOnScreen.Y - posInfo.SizeInPixels.Y - paddingY

    // if the tooltip would be offscreen, move it to the left
    if topLeftX+posInfo.SizeInPixels.X > uiController.GetScreenSize().X-paddingX {
        topLeftX = uiController.GetScreenSize().X - posInfo.SizeInPixels.X - paddingX
    }

    // if it is above the screen, move it down
    if topLeftY < 0 {
        topLeftY = 10
    }
    i.topLeft = geometry.Point{X: topLeftX, Y: topLeftY}
}

func (i *BasicLabel) layoutFromTopLeftPos(topLeftPos geometry.Point) {
    screenSize := i.uiController.GetScreenSize()
    topLeft := geometry.Point{X: screenSize.X - int(float64(topLeftPos.X)), Y: int(float64(topLeftPos.Y))}
    i.topLeft = topLeft
}

func (i *BasicLabel) ClearText() {
    i.textLines = []LabelText{}
}

func (i *BasicLabel) SetBorderPadding(borderPadding geometry.Point) {
    i.borderPadding = borderPadding
}

func (i *BasicLabel) GetBorderPadding() geometry.Point {
    return i.borderPadding
}

func (i *BasicLabel) SetCentered() {
    i.centered = true
}

func (i *BasicLabel) SetCanBeClosed(value bool) {
    i.canBeClosed = value
}

func (i *BasicLabel) GetBounds() geometry.Rect {
    return i.bounds
}
