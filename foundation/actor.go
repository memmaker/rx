package foundation

import (
	"RogueUI/special"
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
	GetState() AIState
	GetDetailInfo() string
	GetInternalName() string
	IsAlive() bool
	GetBodyPart(index int) special.BodyPart
	GetBodyPartIndex(aim special.BodyPart) int
	GetDamageResistance() int
	GetMainHandDamageAsString() string
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

type AIState uint8

const (
	Neutral AIState = iota
	AttackEverything
	AttackEnemies
	Panic
)

func AIStateFromString(str string) AIState {
	str = strings.ToLower(str)
	switch str {
	case "neutral":
		return Neutral
	case "hostile":
		return AttackEverything
	case "ally":
		return Panic
	}
	return Neutral
}
