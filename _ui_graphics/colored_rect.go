package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
    "image/color"
)

type ColoredRect struct {
    uiController     *UIController
    distanceToScreen int
    topLeft          geometry.Point
    bottomRight      geometry.Point
    fillColor        color.Color
}

func (c *ColoredRect) OnCommand(command CommandType) bool {
    return false
}

func (c *ColoredRect) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
    return false
}

func (c *ColoredRect) OnMouseMoved(x int, y int) (bool, Tooltip) {
    return false, EmptyTooltip{}
}

func (c *ColoredRect) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (c *ColoredRect) ShouldClose() bool {
    return false
}

func (c *ColoredRect) Layout() {
    screenSize := c.uiController.GetScreenSize()

    c.topLeft = geometry.Point{X: screenSize.X - c.distanceToScreen}
    c.bottomRight = geometry.Point{X: screenSize.X, Y: screenSize.Y}
}

func NewColoredRect(uiController *UIController, distanceToScreen int) *ColoredRect {
    c := &ColoredRect{
        uiController:     uiController,
        distanceToScreen: distanceToScreen,
        fillColor:        color.White,
    }
    c.Layout()
    return c
}

func (c *ColoredRect) SetFillColor(fillColor color.Color) {
    c.fillColor = fillColor
}
func (c *ColoredRect) Draw() {
    renderer := c.uiController.GetRenderer()
    renderer.DrawColoredRect(c.topLeft, c.bottomRight, c.fillColor)
}
