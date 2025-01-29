package ui_graphics

import (
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
)

type Alignment int

const (
    AlignmentLeft Alignment = iota
    AlignmentRight
)

type PositionInfoForShops struct {
    SizeInPixels   geometry.Point
    LineDist       int
    LineHeight     float64
    ColumnXOffsets []int // relative to the left/right border of the menu
    ItemWidths     [][]float64
}

type ShopMenuItem struct {
    TextColumns []LabelText
    Tooltip     []LabelText
    Action      func()
    ShortCut    rune
    Icon        int32
}

type ShopMenu struct {
    uiController  *UIController
    topLeft       geometry.Point
    items         []ShopMenuItem
    borderPadding geometry.Point
    posInfo       PositionInfoForShops
    layoutFunc    func()
    shouldClose   bool

    selectedIndex        int
    closeOnSelect        bool
    bounds               geometry.Rect
    columnCount          int
    alignInfo            []Alignment
    colDistance          int
    lineDistance         int
    headerTextLeft       string
    headerTextRight      string
    headerTextRightWidth float64
    tileSize             geometry.Point
    iconScale            float64
}

func NewShopMenu(uiController *UIController, iconTileSize geometry.Point) *ShopMenu {
    b := &ShopMenu{
        uiController:  uiController,
        closeOnSelect: true,
        borderPadding: geometry.Point{X: 40, Y: 40},
        colDistance:   10,
        lineDistance:  10,
        tileSize:      iconTileSize,
    }
    b.layoutFunc = func() {
        b.posInfo = b.sizeNeededForMenu(b.uiController.GetRenderer().MeasureString, b.items, b.alignInfo, b.lineDistance, b.colDistance)
        b.layoutCentered()
    }
    return b
}

func (i *ShopMenu) SetHeader(textLeft, textRight string) {
    i.headerTextLeft = textLeft
    i.headerTextRight = textRight
}

func (i *ShopMenu) SetItems(items []ShopMenuItem, alignInfo []Alignment) {
    columnCount := len(items[0].TextColumns)
    if len(items) == 0 || columnCount != len(alignInfo) {
        return
    }
    i.items = items
    i.columnCount = columnCount
    i.alignInfo = alignInfo
    i.Layout()
}

func (i *ShopMenu) OnCommand(command CommandType) bool {
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
func (i *ShopMenu) OnKeyPressed(key ebiten.Key, shiftPressed bool) {
    itemRune := keyToInventoryRune(key, shiftPressed)
    for index, item := range i.items {
        if item.ShortCut == itemRune {
            i.selectedIndex = index
            i.confirmSelection()
            return
        }
    }
}
func (i *ShopMenu) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
    if button == ebiten.MouseButtonRight {
        i.close()
        return true
    }
    i.OnMouseMoved(x, y)
    i.confirmSelection()
    return true
}

func (i *ShopMenu) OnMouseMoved(x int, y int) (bool, Tooltip) {
    i.selectedIndex = -1

    mousePos := geometry.Point{X: x, Y: y}
    if !i.bounds.Contains(mousePos) {
        return false, EmptyTooltip{}
    }

    lowerY := float64(i.topLeft.Y) + 1
    upperY := lowerY + i.posInfo.LineHeight
    for index := 0; index < len(i.items); index++ {
        if float64(mousePos.Y) >= lowerY && float64(mousePos.Y) <= upperY {
            i.selectedIndex = index
            return true, NewTextTooltip(i.uiController, i.items[index].Tooltip, geometry.Point{X: x, Y: y})
        }

        lowerY = upperY + float64(i.posInfo.LineDist)
        upperY = lowerY + i.posInfo.LineHeight
    }
    return false, EmptyTooltip{}
}

func (i *ShopMenu) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (i *ShopMenu) ShouldClose() bool {
    return i.shouldClose
}

func (i *ShopMenu) Layout() {
    i.layoutFunc()
}

func (i *ShopMenu) Draw() {
    uiRenderer := i.uiController.GetRenderer()

    borderTopLeft := i.topLeft.Sub(i.borderPadding.Div(2))
    borderSize := i.posInfo.SizeInPixels.Add(i.borderPadding)
    uiRenderer.DrawDefaultBorder(borderTopLeft, borderSize)

    i.drawStringArray(i.items)

    if i.headerTextLeft == "" && i.headerTextRight == "" {
        return
    }
    // draw a single line above the rest, with its own background
    topLeftHeader := i.topLeft.Sub(geometry.Point{Y: int(i.borderPadding.Y) + int(i.posInfo.LineHeight) + i.lineDistance})
    headerSize := geometry.Point{X: i.posInfo.SizeInPixels.X, Y: int(i.posInfo.LineHeight)}

    headerBorderTopLeft := topLeftHeader.Sub(i.borderPadding.Div(2))
    headerBorderSize := headerSize.Add(i.borderPadding)
    uiRenderer.DrawDefaultBorder(headerBorderTopLeft, headerBorderSize)

    uiRenderer.DrawTTFOnScreen(float64(topLeftHeader.X)+1, float64(topLeftHeader.Y)+i.posInfo.LineHeight+1, i.headerTextLeft, defs.DarkTextColor)
    // right align
    headerTextRightX := float64(i.topLeft.X) + float64(i.posInfo.SizeInPixels.X) - i.headerTextRightWidth

    uiRenderer.DrawTTFOnScreen(headerTextRightX, float64(topLeftHeader.Y)+i.posInfo.LineHeight+1, i.headerTextRight, defs.DarkTextColor)
}

func (i *ShopMenu) IsNull() bool {
    return len(i.items) == 0
}

func (i *ShopMenu) drawStringArray(lines []ShopMenuItem) {
    if len(lines) == 0 {
        return
    }
    uiRenderer := i.uiController.GetRenderer()

    drawY := float64(i.topLeft.Y) + float64(i.posInfo.LineHeight) + 1

    for lineIndex, line := range lines {

        lineStartX := i.topLeft.X + 1
        iconDrawY := drawY - float64(i.tileSize.Y)*i.iconScale
        uiRenderer.DrawOnScreenWithScale(lineStartX, int(iconDrawY), line.Icon, i.iconScale)

        colOffset := i.posInfo.ColumnXOffsets[1]
        letterDrawX := 1.0 + float64(i.topLeft.X) + float64(colOffset)

        uiRenderer.DrawTTFOnScreen(letterDrawX, drawY, fmt.Sprintf("%c)", line.ShortCut), defs.White)
        for ci, column := range line.TextColumns {
            columnIndex := ci + 2 // +2 for the icon + letter shortcut column
            colOffset = i.posInfo.ColumnXOffsets[columnIndex]

            drawX := 1.0
            if i.alignInfo[ci] == AlignmentRight {
                itemWidth := i.posInfo.ItemWidths[lineIndex][columnIndex]
                drawX += float64(i.topLeft.X) + float64(i.posInfo.SizeInPixels.X) - float64(colOffset) - itemWidth
            } else {
                drawX += float64(i.topLeft.X) + float64(colOffset)
            }

            if column.UseColorCodes {
                textWithColorCodes := column.TextWithColorCodes
                uiRenderer.DrawTTFOnScreenWithColorCodes(drawX, float64(drawY), textWithColorCodes)
            } else {
                drawColor := column.TextColor
                if drawColor == nil {
                    drawColor = defs.DarkTextColor
                }
                if lineIndex == i.selectedIndex {
                    drawColor = defs.LightTextColor
                }

                uiRenderer.DrawTTFOnScreen(float64(drawX), float64(drawY), column.Text, drawColor)
            }
        }
        drawY += float64(i.posInfo.LineHeight) + float64(i.posInfo.LineDist)
    }

}

func (i *ShopMenu) layoutCentered() {
    screenSize := i.uiController.GetScreenSize()
    topLeft := geometry.Point{X: (screenSize.X - i.posInfo.SizeInPixels.X) / 2, Y: (screenSize.Y - i.posInfo.SizeInPixels.Y) / 2}
    i.topLeft = topLeft
    halfPadding := i.borderPadding.Div(2)
    halfPadX, halfPadY := halfPadding.X, halfPadding.Y
    i.bounds = geometry.NewRect(i.topLeft.X-halfPadX, i.topLeft.Y-halfPadY, i.topLeft.X+i.posInfo.SizeInPixels.X+halfPadX, i.topLeft.Y+i.posInfo.SizeInPixels.Y+halfPadY)

    //i.headerTextRight = fmt.Sprintf("%d gold", i.goldOfPlayer)
    i.headerTextRightWidth, _ = i.uiController.GetRenderer().MeasureString(i.headerTextRight)
}

func (i *ShopMenu) selectionUp() {
    if i.selectedIndex == 0 {
        i.selectedIndex = len(i.items) - 1
    } else {
        i.selectedIndex--
    }
}

func (i *ShopMenu) selectionDown() {
    if i.selectedIndex == len(i.items)-1 {
        i.selectedIndex = 0
    } else {
        i.selectedIndex++
    }
}

func (i *ShopMenu) confirmSelection() {
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

func (i *ShopMenu) hasValidSelection() bool {
    return i.selectedIndex >= 0 && i.selectedIndex < len(i.items)
}

func (i *ShopMenu) close() {
    i.shouldClose = true
}

func (i *ShopMenu) sizeNeededForMenu(measureString func(line string) (float64, float64), items []ShopMenuItem, alignments []Alignment, lineDistance int, colDistance int) PositionInfoForShops {

    // we'll start by calculating the width needed for each column and the height needed for each row
    alignments = append([]Alignment{AlignmentLeft}, alignments...)
    columnCount := len(items[0].TextColumns) + 1 // +1 for the letter shortcut column
    columnWidths := make([]int, columnCount)
    var rowHeight float64
    var itemWidths [][]float64
    for lineIndex, item := range items {
        itemWidths = append(itemWidths, make([]float64, columnCount))
        wL, hL := measureString(fmt.Sprintf("%c)", item.ShortCut))
        if hL > rowHeight {
            rowHeight = hL
        }
        if wL > float64(columnWidths[0]) {
            columnWidths[0] = int(wL)
        }
        for ci, column := range item.TextColumns {
            columnIndex := ci + 1 // +1 for the letter shortcut column
            w, h := measureString(column.Text)
            width := int(w)
            itemWidths[lineIndex][columnIndex] = w
            if h > rowHeight {
                rowHeight = h
            }
            if width > columnWidths[columnIndex] {
                columnWidths[columnIndex] = width
            }
        }
    }

    /* Start Add ICON */
    tileHeight := i.tileSize.Y
    // make the tileheight the same as the line height
    i.iconScale = float64(rowHeight) / float64(tileHeight)
    iconWidth := int(float64(i.tileSize.X) * i.iconScale)

    columnCount++
    columnWidths = append([]int{iconWidth}, columnWidths...)
    alignments = append([]Alignment{AlignmentLeft}, alignments...)

    for lineIndex, item := range itemWidths {
        itemWidths[lineIndex] = append([]float64{float64(iconWidth)}, item...)
    }
    /* End Add ICON */

    totalHeight := (len(items)-1)*lineDistance + int(float64(len(items))*rowHeight) + (lineDistance / 2)
    totalWidth := (columnCount - 1) * colDistance

    for _, columnWidth := range columnWidths {
        totalWidth += columnWidth
    }

    var xOffsets []int
    var xAccumulator int
    for columnIndex, columnWidth := range columnWidths {
        alignment := alignments[columnIndex]
        if alignment == AlignmentLeft {
            // offset from left border
            drawX := xAccumulator
            xOffsets = append(xOffsets, drawX)
        } else if alignment == AlignmentRight {
            // drawX should now be the offset from the right border..
            drawX := totalWidth - (xAccumulator + columnWidth)
            xOffsets = append(xOffsets, drawX)
        }
        xAccumulator += columnWidth + colDistance
    }

    posInfo := PositionInfoForShops{
        SizeInPixels: geometry.Point{
            X: totalWidth,
            Y: totalHeight,
        },
        LineDist:       lineDistance,
        LineHeight:     rowHeight,
        ColumnXOffsets: xOffsets,
        ItemWidths:     itemWidths,
    }
    return posInfo
}
