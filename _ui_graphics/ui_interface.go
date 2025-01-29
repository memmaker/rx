package ui_graphics

import (
    "image/color"
)

// TODO: DO we need this? should only need the tilerenderer??
type UIElement interface {
    Name() string
    GetTooltipLines(textParser func(string) []ColoredTextPart) []LabelText
    InventoryIcon() int32
    TintColor() color.RGBA
    Counter() int
    String() string
}
