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
	GetMainHandDamageAsString() string
	TimeEnergy() int
	TimeNeededForMovement() int
	TimeNeededForActions() int
	TimeNeededForAttack() int
	ActionDescription() string
	GetArmorProtectionString() string
}

type ChatterType int

const (
	ChatterOnTheWayToAKill ChatterType = iota
	ChatterKillOneLiner
	ChatterBeingDamaged
	ChatterBeingAroundPlayer
	ChatterMinorCrimeNoticed
	ChatterIWarnedYou
	ChatterTrespassing
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
	case "minor_crime_noticed":
		return ChatterMinorCrimeNoticed
	case "i_warned_you":
		return ChatterIWarnedYou
	case "trespassing":
		return ChatterTrespassing
	}
	return ChatterOnTheWayToAKill
}
func (t ChatterType) DefaultChatter() string {
	switch t {
	case ChatterOnTheWayToAKill:
		return "On my way.."
	case ChatterKillOneLiner:
		return "You're dead."
	case ChatterBeingDamaged:
		return "Ouch!"
	case ChatterBeingAroundPlayer:
		return ""
	case ChatterMinorCrimeNoticed:
		return "Stop that!"
	case ChatterIWarnedYou:
		return "I warned you!"
	case ChatterTrespassing:
		return "Get out of here!"
	}
	return ""
}
