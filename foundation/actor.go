package foundation

import (
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
)

type ActorForUI interface {
	Icon() textiles.TextIcon
	TextIcon(background color.RGBA) textiles.TextIcon
	Name() string
	Position() geometry.Point
	GetListInfo() string
	GetHitPoints() int
	GetHitPointsMax() int
	HasFlag(held ActorFlag) bool
	GetDetailInfo() string
	GetInternalName() string
	GetBodyPartByIndex(part int) string
	IsAlive() bool
}
