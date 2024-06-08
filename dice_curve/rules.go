package dice_curve

import "strings"

func GetTotalCostOfAdjustments(adjustments map[Stat]int) int {
	totalCost := 0
	for stat, levels := range adjustments {
		totalCost += stat.CostPerLevel() * levels
	}
	return totalCost
}

func GetStatModifier(stat Stat, sheet *Character, encumbrance Encumbrance) int {
	switch stat {
	case Dodge:
		return -int(encumbrance)
	}
	return 0
}

type Encumbrance int

const (
	EncumbranceNone Encumbrance = iota
	EncumbranceLight
	EncumbranceMedium
	EncumbranceHeavy
	EncumbranceExtraHeavy
	EncumbranceOverloaded
)

func EncumbranceFromString(s string) Encumbrance {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "none":
		return EncumbranceNone
	case "light":
		return EncumbranceLight
	case "medium":
		return EncumbranceMedium
	case "heavy":
		return EncumbranceHeavy
	case "extra_heavy":
		return EncumbranceExtraHeavy
	case "overloaded":
		return EncumbranceOverloaded
	default:
		return EncumbranceNone
	}
}
func GetEncumbrance(basicLift int, carriedWeight int) Encumbrance {
	switch {
	case carriedWeight <= basicLift:
		return EncumbranceNone
	case carriedWeight <= basicLift*2:
		return EncumbranceLight
	case carriedWeight <= basicLift*3:
		return EncumbranceMedium
	case carriedWeight <= basicLift*6:
		return EncumbranceHeavy
	case carriedWeight <= basicLift*10:
		return EncumbranceExtraHeavy
	default:
		return EncumbranceOverloaded
	}
}

func GetDistanceModifier(dist int) int {
	switch {
	case dist <= 2:
		return 0
	case dist <= 3:
		return -1
	case dist <= 5:
		return -2
	case dist <= 7:
		return -3
	case dist <= 10:
		return -4
	case dist <= 15:
		return -5
	case dist <= 20:
		return -6
	case dist <= 30:
		return -7
	case dist <= 50:
		return -8
	case dist <= 70:
		return -9
	case dist <= 100:
		return -10
	}
	return -30
}
