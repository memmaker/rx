package util

import (
	"image/color"
)

type MenuItem struct {
	MainText               string
	MainTextWithColorCodes []ColoredTextPart
	MainTextColor          color.Color

	UseColorCodes bool
	Action        func()
	CharIcon      int32
	TooltipText   []LabelText
	ActionLeft    func() string
	ActionRight   func() string
}

func (m MenuItem) String() string {
	return m.MainText
}

type LabelText struct {
	Text               string
	TextColor          color.Color
	UseColorCodes      bool
	TextWithColorCodes []ColoredTextPart
}

func (t LabelText) String() string {
	return t.Text
}

func (t LabelText) WithColor(textColor color.Color) LabelText {
	t.TextColor = textColor
	return t
}

func ToLabelText(buffer []string) []LabelText {
	var result []LabelText
	for _, line := range buffer {
		result = append(result, LabelTextFromString(line))
	}
	return result
}
func ToLabelTextWithColor(buffer []string, textColor color.Color) []LabelText {
	var result []LabelText
	for _, line := range buffer {
		result = append(result, LabelTextFromString(line).WithColor(textColor))
	}
	return result
}
func LabelTextFromString(text string) LabelText {
	return LabelText{Text: text, TextColor: color.White}
}
func LabelTextFromStringWithColor(text string, textColor color.Color) LabelText {
	return LabelText{Text: text, TextColor: textColor}
}
func IndexToXY(index int, width int) (int, int) {
	return index % width, index / width
}

func XYToIndex(x int, y int, width int) int {
	return y*width + x
}
