package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
)

type MinimapOverlay struct {
    uiController             *UIController
    topLeft                  geometry.Point
    offsetFromTopRightCorner geometry.Point
    bounds                   geometry.Rect
    isHidden                 bool
    isMouseOver              bool
    miniMapImage             *ebiten.Image
    availableWidth           int
    scaleFactor              float64
    onMapClickedHandler      func(button ebiten.MouseButton, mapPos geometry.Point)
    maxHeight                int
}

func (o *MinimapOverlay) OnCommand(command CommandType) bool {
    if o.isHidden {
        return false
    }
    return false
}

func (o *MinimapOverlay) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
    if o.isHidden {
        return false
    }
    mousePos := geometry.Point{X: x, Y: y}
    if !o.bounds.Contains(mousePos) {
        return false
    }

    if o.onMapClickedHandler != nil {
        relativeCoordsInMap := mousePos.Sub(o.topLeft)
        mapPos := geometry.Point{
            X: int(float64(relativeCoordsInMap.X) / o.scaleFactor),
            Y: int(float64(relativeCoordsInMap.Y) / o.scaleFactor),
        }
        o.onMapClickedHandler(button, mapPos)
    }

    return false
}

func (o *MinimapOverlay) OnMouseMoved(x int, y int) (bool, Tooltip) {
    o.isMouseOver = false
    if o.isHidden {
        return false, EmptyTooltip{}
    }
    mousePos := geometry.Point{X: x, Y: y}
    if o.bounds.Contains(mousePos) {
        o.isMouseOver = true
    }
    return false, EmptyTooltip{}
}

func (o *MinimapOverlay) OnMouseWheel(x int, y int, dy float64) bool {
    if o.isHidden {
        return false
    }
    return false

}

func (o *MinimapOverlay) ShouldClose() bool {
    return false

}

func (o *MinimapOverlay) Layout() {
    o.layoutFromTopRightCorner(o.offsetFromTopRightCorner)
}

func (o *MinimapOverlay) layoutFromTopRightCorner(offsetFromTopRightCorner geometry.Point) {
    screenSize := o.uiController.GetScreenSize()
    image := o.miniMapImage
    mapSizeX := image.Bounds().Dx()
    mapSizeY := image.Bounds().Dy()

    o.scaleFactor = float64(o.availableWidth) / float64(mapSizeX)
    mapSizeX = int(float64(mapSizeX) * o.scaleFactor)
    mapSizeY = int(float64(mapSizeY) * o.scaleFactor)

    if mapSizeY > o.maxHeight {
        o.scaleFactor = float64(o.maxHeight) / float64(mapSizeY)
        mapSizeX = int(float64(mapSizeX) * o.scaleFactor)
        mapSizeY = int(float64(mapSizeY) * o.scaleFactor)
    }

    o.topLeft = geometry.Point{
        X: screenSize.X - offsetFromTopRightCorner.X,
        Y: offsetFromTopRightCorner.Y,
    }
    o.bounds = geometry.NewRect(o.topLeft.X, o.topLeft.Y, o.topLeft.X+mapSizeX, o.topLeft.Y+mapSizeY)
}

func (o *MinimapOverlay) Draw() {
    if o.isHidden {
        return
    }
    renderer := o.uiController.tileRenderer

    mapSizeP := o.bounds.Size()
    //renderer.DrawDefaultBorder(o.topLeft, mapSizeP)
    image := o.miniMapImage
    renderer.DrawImageOnScreen(o.topLeft.X, o.topLeft.Y, mapSizeP, image)
}

func (o *MinimapOverlay) SetAction(action func()) {
    //o.onConfirmAction = action
}

func (o *MinimapOverlay) SetTooltip(tooltipText string) {
    //o.tooltipText = tooltipText
}

func (o *MinimapOverlay) Hide() {
    o.isHidden = true
}

func (o *MinimapOverlay) Show() {
    o.isHidden = false
}

func (o *MinimapOverlay) SetMiniMapImage(image *ebiten.Image) {
    o.miniMapImage = image
}
func (o *MinimapOverlay) SetOnMapClickedHandler(handler func(button ebiten.MouseButton, mapPos geometry.Point)) {
    o.onMapClickedHandler = handler
}
func NewMiniMap(uiController *UIController, offsetFromTopRightCorner geometry.Point, maxHeight, availableWidth int, miniMapImage *ebiten.Image) *MinimapOverlay {
    i := &MinimapOverlay{
        uiController:             uiController,
        miniMapImage:             miniMapImage,
        offsetFromTopRightCorner: offsetFromTopRightCorner,
        availableWidth:           availableWidth,
        maxHeight:                maxHeight,
    }
    i.Layout()
    return i
}
