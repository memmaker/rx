package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
)

type BasicMenu struct {
    uiController   *UIController
    topLeft        geometry.Point
    items          []MenuItem
    itemBounds     []geometry.Rect
    borderPadding  geometry.Point
    posInfo        PositionInfo
    centeredText   bool
    layoutFunc     func()
    drawBackground bool
    shouldClose    bool

    selectedIndex int
    closeOnSelect bool
    bounds        geometry.Rect
}

func NewBasicMenu(uiController *UIController) *BasicMenu {
    b := &BasicMenu{
        uiController:   uiController,
        drawBackground: true,
        centeredText:   true,
        closeOnSelect:  true,
        borderPadding:  geometry.Point{X: 40, Y: 40},
    }
    b.layoutFunc = func() {
        b.posInfo = SizeNeededForText(b.uiController.GetRenderer().MeasureString, b.items, 10)
        b.layoutCentered()
    }
    return b
}
func (i *BasicMenu) OnCommand(command CommandType) bool {
    switch command {
    case PlayerCommandCancel:
        i.close()
        return true
    case PlayerCommandUp:
        i.selectionUp()
        return true
    case PlayerCommandDown:
        i.selectionDown()
        return true
    case PlayerCommandConfirm:
        i.confirmSelection()
        return true
    }
    return false
}

func (i *BasicMenu) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
    if button == ebiten.MouseButtonRight {
        i.close()
        return true
    }
    i.OnMouseMoved(x, y)
    i.confirmSelection()
    return true
}

func (i *BasicMenu) OnMouseMoved(x int, y int) (bool, Tooltip) {
    mousePos := geometry.Point{X: x, Y: y}
    i.selectedIndex = -1

    if !i.bounds.Contains(mousePos) {
        return false, EmptyTooltip{}
    }
    for index, itemBounds := range i.itemBounds {
        if itemBounds.Contains(mousePos) {
            i.selectedIndex = index
            return true, EmptyTooltip{}
        }
    }
    return false, EmptyTooltip{}
}

func (i *BasicMenu) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (i *BasicMenu) ShouldClose() bool {
    return i.shouldClose
}

func (i *BasicMenu) Layout() {
    i.layoutFunc()
}

func (i *BasicMenu) Draw() {
    uiRenderer := i.uiController.GetRenderer()
    if i.drawBackground {
        borderTopLeft := i.topLeft.Sub(i.borderPadding.Div(2))
        borderSize := i.posInfo.SizeInPixels.Add(i.borderPadding)
        uiRenderer.DrawDefaultBorder(borderTopLeft, borderSize)
    }
    i.drawStringArray(i.items)
}

func (i *BasicMenu) IsNull() bool {
    return len(i.items) == 0
}

func (i *BasicMenu) drawStringArray(lines []MenuItem) {
    if len(lines) == 0 {
        return
    }
    uiRenderer := i.uiController.GetRenderer()
    //startPos := geometry.Point{X: i.topLeft.X + 2, Y: startAtY}
    //tWidth, tHeight := uiRenderer.MeasureString(linesAsOneString)
    drawY := i.topLeft.Y + i.posInfo.YOffsetForFirstLine
    for lineIndex, line := range lines {
        if line.MainText == "" {
            drawY += i.posInfo.YOffsetForFirstLine + i.posInfo.LineDist
            continue
        }
        itemSize := i.posInfo.ItemSizes[lineIndex]
        sizeX, sizeY := itemSize.X, itemSize.Y
        drawX := i.topLeft.X
        if i.centeredText {
            drawX = i.topLeft.X + (i.posInfo.SizeInPixels.X-sizeX)/2
        }

        if line.UseColorCodes {
            textWithColorCodes := line.MainTextWithColorCodes
            if lineIndex == i.selectedIndex {
                textWithColorCodes[0].Color = defs.LightTextColor
                //textWithColorCodes = "[:lightText]" + textWithColorCodes
            } else {
                //textWithColorCodes = "[:darkText]" + textWithColorCodes
                textWithColorCodes[0].Color = defs.DarkTextColor
            }
            uiRenderer.DrawTTFOnScreenWithColorCodes(float64(drawX), float64(drawY), textWithColorCodes)
        } else {
            drawColor := line.MainTextColor
            if drawColor == nil {
                drawColor = defs.DarkTextColor
            }
            if lineIndex == i.selectedIndex {
                drawColor = defs.LightTextColor
            }

            uiRenderer.DrawTTFOnScreen(float64(drawX), float64(drawY), line.MainText, drawColor)
        }

        drawY += sizeY + i.posInfo.LineDist
    }

}

func (i *BasicMenu) layoutCentered() {
    screenSize := i.uiController.GetScreenSize()
    topLeft := geometry.Point{X: (screenSize.X - i.posInfo.SizeInPixels.X) / 2, Y: (screenSize.Y - i.posInfo.SizeInPixels.Y) / 2}
    i.topLeft = topLeft
    i.itemBounds = i.calcItemBounds()
    halfPadding := i.borderPadding.Div(2)
    halfPadX, halfPadY := halfPadding.X, halfPadding.Y
    i.bounds = geometry.NewRect(i.topLeft.X-halfPadX, i.topLeft.Y-halfPadY, i.topLeft.X+i.posInfo.SizeInPixels.X+halfPadX, i.topLeft.Y+i.posInfo.SizeInPixels.Y+halfPadY)
}

func (i *BasicMenu) SetItems(items []MenuItem) {
    i.items = items
    i.Layout()
}

func (i *BasicMenu) selectionUp() {
    if i.selectedIndex == 0 {
        i.selectedIndex = len(i.items) - 1
    } else {
        i.selectedIndex--
    }
}

func (i *BasicMenu) selectionDown() {
    if i.selectedIndex == len(i.items)-1 {
        i.selectedIndex = 0
    } else {
        i.selectedIndex++
    }
}

func (i *BasicMenu) confirmSelection() {
    if !i.hasValidSelection() {
        return
    }
    action := i.items[i.selectedIndex].Action
    if action != nil {
        action()
    }
    if i.closeOnSelect {
        i.close()
    }
}

func (i *BasicMenu) hasValidSelection() bool {
    return i.selectedIndex >= 0 && i.selectedIndex < len(i.items)
}

func (i *BasicMenu) close() {
    i.shouldClose = true
}

func (i *BasicMenu) calcItemBounds() []geometry.Rect {
    itemBounds := make([]geometry.Rect, len(i.items))
    drawY := i.topLeft.Y
    padding := i.borderPadding.Div(2)
    for lineIndex, line := range i.items {
        if line.MainText == "" {
            drawY += i.posInfo.YOffsetForFirstLine + i.posInfo.LineDist
            continue
        }
        itemSize := i.posInfo.ItemSizes[lineIndex]
        _, sizeY := itemSize.X, itemSize.Y
        // let's simplify this and use the whole width of the menu
        itemBounds[lineIndex] = geometry.NewRect(i.topLeft.X-padding.X, drawY, i.topLeft.X+i.posInfo.SizeInPixels.X+i.borderPadding.X, drawY+sizeY)
        drawY += sizeY + i.posInfo.LineDist
    }
    return itemBounds
}

func (i *BasicMenu) SetCentered(centered bool) {
    i.centeredText = centered
}
