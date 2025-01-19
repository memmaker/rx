package foundation

import (
	"contractor/d100"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"strings"
)

type ActorForUI interface {
	GetIcon() textiles.TextIcon
	TextIcon(background color.RGBA) textiles.TextIcon
	Name() string
	Position() geometry.Point
	GetListInfo() string
	GetHitPoints() int
	GetHitPointsMax() int
	HasFlag(held ActorFlag) bool
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

type ChatterTopic int

const (
	ChatterOnTheWayToAKill ChatterTopic = iota
	ChatterKillOneLiner
	ChatterBeingDamaged
	ChatterBeingAroundPlayer
	ChatterInvestigating
	ChatterMinorCrimeNoticed
	ChatterIWarnedYou
	ChatterTrespassing
	ChatterOpenCarryNoticed
	ChatterTargetLost
	ChatterAttackNoticed
	ChatterSneakyAttackNoticed
)

func NewChatterTopicFromString(str string) ChatterTopic {
	str = strings.ToLower(str)
	switch str {
	case "way_to_kill":
		return ChatterOnTheWayToAKill
	case "kill_one_liner":
		return ChatterKillOneLiner
	case "being_damaged":
		return ChatterBeingDamaged
	case "open_carry_noticed":
		return ChatterOpenCarryNoticed
	case "being_around_player":
		return ChatterBeingAroundPlayer
	case "investigating":
		return ChatterInvestigating
	case "attack_noticed":
		return ChatterAttackNoticed
	case "sneaky_attack_noticed":
		return ChatterSneakyAttackNoticed
	case "minor_crime_noticed":
		return ChatterMinorCrimeNoticed
	case "i_warned_you":
		return ChatterIWarnedYou
	case "trespassing":
		return ChatterTrespassing
	case "target_lost":
		return ChatterTargetLost
	}
	return ChatterOnTheWayToAKill
}
func (t ChatterTopic) DefaultChatter() string {
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
	case ChatterAttackNoticed:
		return "Stop the attack!"
	case ChatterSneakyAttackNoticed:
		return "Got you!"
	case ChatterIWarnedYou:
		return "I warned you!"
	case ChatterTrespassing:
		return "Get out of here!"
	case ChatterInvestigating:
		return "What's that?"
	case ChatterOpenCarryNoticed:
		return "Put that weapon away!"
	case ChatterTargetLost:
		return "Where did they go?"
	}
	return ""
}

func (t ChatterTopic) IsMajorCrime() bool {
	switch t {
	case ChatterAttackNoticed:
		return true
	case ChatterSneakyAttackNoticed:
		return true
	}
	return false
}
