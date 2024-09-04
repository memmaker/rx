package foundation

import (
	"RogueUI/special"
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
	HasFlag(held special.ActorFlag) bool
	GetDetailInfo() string
	GetInternalName() string
	IsAlive() bool
	GetBodyPart(index int) special.BodyPart
	GetBodyPartIndex(aim special.BodyPart) int
}
