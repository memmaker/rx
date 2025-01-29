package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
)

type ButtonOverlay struct {
    uiController             *UIController
    topLeft                  geometry.Point
    offsetFromTopRightCorner geometry.Point
    tileSize                 geometry.Point
    layoutFunc               func()
    icon                     int32
    onConfirmAction          func()
    tooltipText              string
    bounds                   geometry.Rect
    isHidden                 bool
    tileScale                float64
    isMouseOver              bool
}

func (o *ButtonOverlay) OnCommand(command CommandType) bool {
    if o.isHidden {
        return false
    }
    return false
}

func (o *ButtonOverlay) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
    if o.isHidden {
        return false
    }
    mousePos := geometry.Point{X: x, Y: y}
    if !o.bounds.Contains(mousePos) {
        return false
    }
    if o.onConfirmAction != nil {
        o.onConfirmAction()
        return true
    }
    return false
}

func (o *ButtonOverlay) OnMouseMoved(x int, y int) (bool, Tooltip) {
    o.isMouseOver = false
    if o.isHidden {
        return false, EmptyTooltip{}
    }
    mousePos := geometry.Point{X: x, Y: y}
    if o.bounds.Contains(mousePos) {
        o.isMouseOver = true
        return true, NewTextTooltip(o.uiController, []LabelText{LabelTextFromString(o.tooltipText)}, mousePos)
    }
    return false, EmptyTooltip{}
}

func (o *ButtonOverlay) OnMouseWheel(x int, y int, dy float64) bool {
    if o.isHidden {
        return false
    }
    return false

}

func (o *ButtonOverlay) ShouldClose() bool {
    return false

}

func (o *ButtonOverlay) Layout() {
    o.layoutFunc()
}

func (o *ButtonOverlay) layoutFromTopRightCorner(offsetFromTopRightCorner geometry.Point) {
    o.topLeft = geometry.Point{
        X: o.uiController.GetScreenSize().X - offsetFromTopRightCorner.X,
        Y: offsetFromTopRightCorner.Y,
    }
    o.bounds = geometry.NewRect(o.topLeft.X, o.topLeft.Y, o.topLeft.X+o.tileSize.X, o.topLeft.Y+o.tileSize.Y)
}

func (o *ButtonOverlay) Draw() {
    if o.isHidden {
        return
    }
    targetingIcon := o.uiController.configuration.TargetingIconIndex
    renderer := o.uiController.tileRenderer
    renderer.DrawDefaultBorder(o.topLeft, o.tileSize)
    renderer.DrawOnScreenWithScale(o.topLeft.X, o.topLeft.Y, o.icon, o.tileScale)
    if o.isMouseOver {
        renderer.DrawOnScreenWithScale(o.topLeft.X, o.topLeft.Y, targetingIcon, o.tileScale)
    }

}

func (o *ButtonOverlay) SetAction(action func()) {
    o.onConfirmAction = action
}

func (o *ButtonOverlay) SetTooltip(tooltipText string) {
    o.tooltipText = tooltipText
}

func (o *ButtonOverlay) Hide() {
    o.isHidden = true
}

func (o *ButtonOverlay) Show() {
    o.isHidden = false
}

func (o *ButtonOverlay) IsVisible() bool {
    return !o.isHidden
}

func (o *ButtonOverlay) GetBounds() geometry.Rect {
    return o.bounds
}

func NewIconButton(uiController *UIController, offsetFromTopRightCorner, tileSize geometry.Point, icon int32) *ButtonOverlay {
    tileScale := 4.0
    i := &ButtonOverlay{
        uiController:             uiController,
        icon:                     icon,
        offsetFromTopRightCorner: offsetFromTopRightCorner,
        tileSize:                 tileSize.MulF(tileScale),
        tileScale:                tileScale,
    }
    i.layoutFunc = func() {
        i.layoutFromTopRightCorner(offsetFromTopRightCorner)
    }
    i.Layout()
    return i
}
