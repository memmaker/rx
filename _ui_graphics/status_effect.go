package ui_graphics

import (
    "fmt"
    "image/color"
)

type StatusEffectWidget struct {
    statusEffect string
    turnsLeft    int
}

func (s *StatusEffectWidget) String() string {
    return s.statusEffect
}

func NewStatusEffectWidget(statusName string, turnsLeft int) *StatusEffectWidget {
    return &StatusEffectWidget{
        statusEffect: statusName,
        turnsLeft:    turnsLeft,
    }
}
func (s *StatusEffectWidget) Name() string {
    return s.statusEffect
}

func (s *StatusEffectWidget) GetTooltipLines(func(string) []ColoredTextPart) []LabelText {
    return []LabelText{
        {
            Text:      fmt.Sprintf("%s for %d turns", s.statusEffect, s.turnsLeft),
            TextColor: color.White,
        },
    }
}

func (s *StatusEffectWidget) InventoryIcon() int32 {
    return 0
}

func (s *StatusEffectWidget) TintColor() color.RGBA {
    return color.RGBA{255, 255, 255, 255}
}

func (s *StatusEffectWidget) Counter() int {
    return s.turnsLeft
}
