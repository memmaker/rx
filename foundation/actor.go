package foundation

import (
	"RogueUI/geometry"
	"image/color"
)

type ActorForUI interface {
	Icon() rune
	TextIcon(background color.RGBA, getColor func(string) color.RGBA) TextIcon
	Color() string
	Name() string
	Position() geometry.Point
	GetListInfo() string
	GetHitPoints() int
	GetHitPointsMax() int
	HasFlag(held ActorFlag) bool
	GetDetailInfo() []string
	GetInternalName() string
}
