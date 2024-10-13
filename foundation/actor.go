package foundation

import (
	"RogueUI/d100"
	"RogueUI/fsmai"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"strings"
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
	GetState() fsmai.StateName
	GetDetailInfo() string
	GetInternalName() string
	IsAlive() bool
	GetBodyPart(index int) d100.BodyPart
	GetBodyPartIndex(aim d100.BodyPart) int
	GetDamageResistance() int
	GetMainHandDamageAsString() string
	TimeEnergy() int
	TimeNeededForMovement() int
	TimeNeededForActions() int
	TimeNeededForAttack() int
	ActionDescription(name string) string
}

type ChatterType int

const (
	ChatterOnTheWayToAKill ChatterType = iota
	ChatterKillOneLiner
	ChatterBeingDamaged
	ChatterBeingAroundPlayer
)

func NewChatterTypeFromString(str string) ChatterType {
	str = strings.ToLower(str)
	switch str {
	case "way_to_kill":
		return ChatterOnTheWayToAKill
	case "kill_one_liner":
		return ChatterKillOneLiner
	case "being_damaged":
		return ChatterBeingDamaged
	case "being_around_player":
		return ChatterBeingAroundPlayer
	}
	return ChatterOnTheWayToAKill
}
