package ui_graphics

import (
    "image/color"
)

type PrintMessage struct {
    rawText         string
    textForRenderer []ColoredTextPart
    ticksLeft       uint64
    ticksStart      uint64
}

func (m PrintMessage) LifeColor() color.Color {
    // at full life, it's white fully opaque
    // at 0 life, it's white fully transparent
    percentage := 1.0 - (float64(m.ticksLeft) / float64(m.ticksStart))
    easedPercentage := EaseInExpo(percentage)
    percentageLeft := 1.0 - easedPercentage
    return color.RGBA{
        R: 255,
        G: 255,
        B: 255,
        A: uint8(float64(255) * percentageLeft),
    }
}

func (m PrintMessage) GetText() string {
    return m.rawText
}

func (m PrintMessage) GetHeight() float64 {
    return m.textForRenderer[0].Height
}

func (m PrintMessage) GetColorcodedText() []ColoredTextPart {
    return m.textForRenderer
}
