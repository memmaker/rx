package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
    "image/color"
    "strconv"
)

type IconDataSource struct {
    IsHighlighted func(UIElement) bool
    GetElements   func() []UIElement
    OnSelection   func(UIElement)
}

type IconOverlay struct {
    infoLabel *BasicLabel

    getElements  func() []UIElement
    dimensions   geometry.Point
    margin       geometry.Point
    uiController *UIController
    topLeft      geometry.Point

    isHighlighted    func(UIElement) bool
    tileSizeInPixels geometry.Point
    bounds           geometry.Rect
    selectedIndex    int
    scale            float64

    showEmptySlots bool
    showShortCuts  bool
    showCounter    bool

    onSelection                 func(UIElement)
    isActive                    bool
    offsetFromBottomRightCorner geometry.Point
    tileSize                    geometry.Point
    onClose                     func()
    shouldClose                 bool
    canBeClosed                 bool
    infoLabelHandler            func(UIElement) string
    onCancel                    func()
}

func NewIconOverlay(uiController *UIController, offsetFromBottomRightCorner, tileSize geometry.Point, dataSource IconDataSource) *IconOverlay {
    i := &IconOverlay{
        uiController:                uiController,
        isHighlighted:               dataSource.IsHighlighted,
        getElements:                 dataSource.GetElements,
        onSelection:                 dataSource.OnSelection,
        selectedIndex:               -1,
        offsetFromBottomRightCorner: offsetFromBottomRightCorner,
        tileSize:                    tileSize,
    }
    i.infoLabel = NewCustomLabel(uiController, i.layoutLabel)
    i.Layout()
    return i
}

func (o *IconOverlay) SetInfoLabelHandler(handler func(UIElement) string) {
    o.infoLabelHandler = handler
}

func (o *IconOverlay) Layout() {
    scale := float64(4)
    tileSize := o.tileSize
    offsetFromBottomRightCorner := o.offsetFromBottomRightCorner
    screenSize := o.uiController.GetScreenSize()
    totalScale := scale

    tileSizeInPixels := geometry.Point{
        X: int(float64(tileSize.X) * totalScale),
        Y: int(float64(tileSize.Y) * totalScale),
    }
    dimensions := geometry.Point{X: 8, Y: 4}
    margin := geometry.Point{X: 1, Y: 1}

    width := dimensions.X*tileSizeInPixels.X + (dimensions.X-1)*margin.X
    height := dimensions.Y*tileSizeInPixels.Y + (dimensions.Y-1)*margin.Y
    // position contains the distance in x & axis from the bottom right corner of the screen
    bottomRight := geometry.Point{X: screenSize.X - offsetFromBottomRightCorner.X, Y: screenSize.Y - offsetFromBottomRightCorner.Y}
    topLeft := geometry.Point{X: bottomRight.X - width, Y: bottomRight.Y - height}

    o.offsetFromBottomRightCorner = offsetFromBottomRightCorner
    o.tileSize = tileSize

    o.topLeft = topLeft
    o.tileSizeInPixels = tileSizeInPixels
    o.dimensions = dimensions
    o.margin = margin
    o.bounds = geometry.NewRect(topLeft.X, topLeft.Y, topLeft.X+width, topLeft.Y+height)
    o.scale = scale

    o.layoutLabel(o.uiController, o.infoLabel)
}

func (o *IconOverlay) OnCommand(command CommandType) bool {
    if command == PlayerCommandCancel {
        if o.onCancel != nil {
            o.onCancel()
        }
        o.Close()
        return true
    }
    return false
}

func (o *IconOverlay) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
    if button == ebiten.MouseButtonRight {
        if o.onCancel != nil {
            o.onCancel()
        }
        if o.canBeClosed {
            o.Close()
            return true
        }
    }

    o.selectedIndex = -1
    point := geometry.Point{X: x, Y: y}
    if !o.bounds.Contains(point) {
        return false
    }
    o.selectedIndex = o.getItemIndexFromScreenPosition(point)
    o.onItemSelected(o.selectedIndex)
    return true
}

func (o *IconOverlay) OnMouseMoved(x int, y int) (bool, Tooltip) {
    o.selectedIndex = -1
    point := geometry.Point{X: x, Y: y}
    o.infoLabel.ClearText()
    if !o.bounds.Contains(point) {
        return false, EmptyTooltip{}
    }
    o.selectedIndex = o.getItemIndexFromScreenPosition(point)
    item, isValid := o.getSelectedItem(o.selectedIndex)
    if !isValid {
        return true, EmptyTooltip{}
    }
    infoLabelText := item.Name()

    if o.infoLabelHandler != nil {
        infoLabelText = o.infoLabelHandler(item)
    }

    o.infoLabel.SetText([]LabelText{
        {
            Text:      infoLabelText,
            TextColor: color.White,
        },
    })
    //o.layoutLabel(o.uiController, o.infoLabel)
    return true, NewTextTooltip(o.uiController, item.GetTooltipLines(o.uiController.ParseColorCodedText), point)
}
func (o *IconOverlay) GetBounds() geometry.Rect {
    return o.bounds
}
func (o *IconOverlay) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (o *IconOverlay) ShouldClose() bool {
    return o.shouldClose
}

func (o *IconOverlay) Draw() {
    blackTransparent := color.RGBA{R: 0, G: 0, B: 0, A: 128}
    targetingIcon := o.uiController.configuration.TargetingTileIndex

    renderer := o.uiController.GetRenderer()
    items := o.getElements()
    for i := 0; i < o.dimensions.X*o.dimensions.Y; i++ {
        itemPosition := o.positionForItem(i)
        if i >= len(items) { // an empty slot
            if !o.showEmptySlots {
                continue
            }
            renderer.DrawDefaultBorder(itemPosition, o.tileSizeInPixels)
            //renderer.DrawOnScreenWithScale(itemPosition.X, itemPosition.Y, o.normalBackgroundIcon, o.scale)
            if i == o.selectedIndex {
                renderer.DrawOnScreenWithScale(itemPosition.X, itemPosition.Y, targetingIcon, o.scale)
            }
            continue
        }
        item := items[i]
        if o.isHighlighted(item) {
            renderer.DrawColoredBorder(itemPosition, o.tileSizeInPixels, color.RGBA{12, 111, 12, 255}, color.RGBA{111, 111, 12, 255})
        } else {
            renderer.DrawDefaultBorder(itemPosition, o.tileSizeInPixels)
            //renderer.DrawOnScreenWithScale(itemPosition.X, itemPosition.Y, o.normalBackgroundIcon, o.scale)
        }
        renderer.DrawOnScreenWithScale(itemPosition.X, itemPosition.Y, item.InventoryIcon(), o.scale)
        if i == o.selectedIndex {
            renderer.DrawOnScreenWithScale(itemPosition.X, itemPosition.Y, targetingIcon, o.scale)
        }
        if o.showShortCuts {
            renderer.DrawCharOnScreen(itemPosition.X, itemPosition.Y, GetShortCutForIndex(i), 1, color.White)
        } else if o.showCounter {
            counter := item.Counter()
            if counter > 1 {
                counterString := strconv.Itoa(counter)
                _, cHeight := renderer.MeasureString(counterString)
                tileSize := o.tileSizeInPixels
                // NOTE: we have to specify the lower left corner where we want to start drawing.
                // in our case, that means we want to draw at the lower right corner of our tile
                // itemPosition is currently the top left corner

                drawPos := geometry.Point{
                    X: itemPosition.X,
                    Y: itemPosition.Y + tileSize.Y - 2,
                }
                borderTopLeft := geometry.Point{
                    X: itemPosition.X,
                    Y: itemPosition.Y + (tileSize.Y / 2),
                }
                borderWidth := tileSize.X //max(tileSize.X / 2, cWidth)
                borderHeight := max(tileSize.Y/2, int(cHeight))
                renderer.DrawColoredRect(borderTopLeft, geometry.Point{X: borderWidth, Y: borderHeight}, blackTransparent)
                renderer.DrawTTFOnScreen(float64(drawPos.X), float64(drawPos.Y), counterString, color.White)
            }
        }
    }
    o.infoLabel.Draw()
}

func (o *IconOverlay) positionForItem(i int) geometry.Point {
    itemPosition := geometry.Point{
        X: o.topLeft.X + o.margin.X + (i%o.dimensions.X)*(o.tileSizeInPixels.X+o.margin.X),
        Y: o.topLeft.Y + o.margin.Y + (i/o.dimensions.X)*(o.tileSizeInPixels.Y+o.margin.Y),
    }
    return itemPosition
}

func (o *IconOverlay) getItemIndexFromScreenPosition(point geometry.Point) int {
    point.X -= o.topLeft.X
    point.Y -= o.topLeft.Y
    point.X /= o.tileSizeInPixels.X + o.margin.X
    point.Y /= o.tileSizeInPixels.Y + o.margin.Y
    return point.Y*o.dimensions.X + point.X
}

func (o *IconOverlay) onItemSelected(index int) {
    item, isValid := o.getSelectedItem(index)
    if !isValid {
        return
    }
    o.onSelection(item)
}

func (o *IconOverlay) getSelectedItem(index int) (UIElement, bool) {
    items := o.getElements()
    if index >= len(items) || index < 0 {
        return nil, false
    }
    item := items[index]
    return item, true
}

func (o *IconOverlay) IsActive() bool {
    return o.isActive
}

func (o *IconOverlay) Activate() {
    o.isActive = true
    o.showShortCuts = true
}

func (o *IconOverlay) Deactivate() {
    o.isActive = false
    o.showShortCuts = false
}

func (o *IconOverlay) HasItems() bool {
    return len(o.getElements()) > 0
}

func (o *IconOverlay) SetShowEmptySlots(value bool) {
    o.showEmptySlots = value
}

func (o *IconOverlay) SetShowCounter(value bool) {
    o.showCounter = value
}

func (o *IconOverlay) layoutLabel(controller *UIController, label *BasicLabel) {
    label.posInfo = SizeNeededForText(controller.GetRenderer().MeasureString, label.textLines, 0)
    label.topLeft = geometry.Point{
        X: o.topLeft.X, Y: o.topLeft.Y - label.posInfo.YOffsetForFirstLine - label.posInfo.SizeInPixels.Y + 4,
    }
    //        b.layoutFromTopLeftPos(topLeftPositionAsOffsetFromTopRight)
}

func (o *IconOverlay) GetOnSelectedHandler() func(uiItem UIElement) {
    return o.onSelection
}

func (o *IconOverlay) SetOnSelectedHandler(handler func(uiItem UIElement)) {
    o.onSelection = handler
}

func (o *IconOverlay) SetOnCloseHandler(onClose func()) {
    o.onClose = onClose
}

func (o *IconOverlay) SetCanBeClosed(canBeClosed bool) {
    o.canBeClosed = canBeClosed
}

func (o *IconOverlay) Close() {
    if o.shouldClose == true || !o.canBeClosed {
        return
    }
    o.shouldClose = true
    if o.onClose != nil {
        o.onClose()
    }
}

func (o *IconOverlay) SetOnCancelHandler(cancelHandler func()) {
    o.onCancel = cancelHandler
}
