package d100

import (
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"math"
	"slices"
	"strings"
)

func standardFunctions() map[string]govaluate.ExpressionFunction {
	var standardFuncs = map[string]govaluate.ExpressionFunction{
		"max": func(args ...interface{}) (interface{}, error) {
			intArgs := fxtools.MapSlice(args, func(i interface{}) int {
				return int(i.(float64))
			})
			maxValue := slices.Max(intArgs)
			return float64(maxValue), nil
		},
		"min": func(args ...interface{}) (interface{}, error) {
			intArgs := fxtools.MapSlice(args, func(i interface{}) int {
				return int(i.(float64))
			})
			maxValue := slices.Min(intArgs)
			return float64(maxValue), nil
		},
		"if": func(args ...interface{}) (interface{}, error) {
			expression := args[0].(bool)
			if expression {
				return args[1], nil
			} else {
				return args[2], nil
			}
		},
		"floor": func(args ...interface{}) (interface{}, error) {
			return math.Floor(args[0].(float64)), nil
		},
		"ceil": func(args ...interface{}) (interface{}, error) {
			return math.Ceil(args[0].(float64)), nil
		},
		"round": func(args ...interface{}) (interface{}, error) {
			return math.Round(args[0].(float64)), nil
		},
	}
	return standardFuncs
}

var derivedBaseValues = make(map[DerivedStat]*govaluate.EvaluableExpression)

var VisibleDerivedStatCount = 12

type DerivedStat int

const (
	ActionPoints DerivedStat = iota
	HitPoints
	HealingRate
	Speed
	Dodge
	CarryWeight
	CriticalChance
	CriticalImpactModifier
	MeleeDamageBonus
	DamageResistance
	PoisonResistance
	PartyLimit
	SkillRate
	PerkRate
	DerivedStatCount
)

func (s DerivedStat) String() string {
	switch s {
	case ActionPoints:
		return "Action Points"
	case Dodge:
		return "Dodge"
	case CarryWeight:
		return "Carry Weight"
	case CriticalChance:
		return "Critical Chance"
	case CriticalImpactModifier:
		return "Critical Modifier"
	case DamageResistance:
		return "Damage Resistance"
	case HealingRate:
		return "Healing Rate"
	case HitPoints:
		return "Hit Points"
	case MeleeDamageBonus:
		return "Melee Damage"
	case PartyLimit:
		return "Party Limit"
	case PerkRate:
		return "Perk Rate"
	case PoisonResistance:
		return "Poison Resistance"
	case Speed:
		return "Speed"
	case SkillRate:
		return "Skill Rate"
	}
	return ""
}

func DerivedStatFromString(name string) DerivedStat {
	name = strings.ReplaceAll(strings.ToLower(name), "_", "")
	switch name {
	case "actionpoints":
		return ActionPoints
	case "dodge":
		return Dodge
	case "carryweight":
		return CarryWeight
	case "criticalchance":
		return CriticalChance
	case "criticalimpactmodifier":
		return CriticalImpactModifier
	case "damageresistance":
		return DamageResistance
	case "healingrate":
		return HealingRate
	case "hitpoints":
		return HitPoints
	case "meleedamagebonus":
		return MeleeDamageBonus
	case "partylimit":
		return PartyLimit
	case "perkrate":
		return PerkRate
	case "poisonresistance":
		return PoisonResistance
	case "speed":
		return Speed
	case "skillrate":
		return SkillRate

	}
	panic("invalid derived stat name")
	return -1
}

func (s DerivedStat) ToShortString() string {
	switch s {
	case ActionPoints:
		return "AP"
	case Dodge:
		return "DG"
	case CarryWeight:
		return "CW"
	case CriticalChance:
		return "CC"
	case CriticalImpactModifier:
		return "CI"
	case DamageResistance:
		return "DR"
	case HealingRate:
		return "HR"
	case HitPoints:
		return "HP"
	case MeleeDamageBonus:
		return "MD"
	case PartyLimit:
		return "PL"
	case PerkRate:
		return "PR"
	case PoisonResistance:
		return "PR"
	case Speed:
		return "SP"
	case SkillRate:
		return "SR"
	}
	return ""
}

func (s DerivedStat) Source() string {
	return derivedBaseValues[s].String()
}
