package util

import (
	"image/color"
)

type ColoredTextPart struct {
	Text    string
	Color   color.Color
	XOffset float64
	Height  float64
}
