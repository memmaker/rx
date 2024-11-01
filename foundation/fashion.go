package foundation

import "strings"

type FashionStyle int8

func (s FashionStyle) PriceMultiplier() int {
	switch s {
	case FashionStyleCombatGear:
		return 1
	case FashionStyleLowLife:
		return 1
	case FashionStyleChromeGang:
		return 1
	case FashionStyleNomadLeather:
		return 1
	case FashionStyleGenericChic:
		return 1
	case FashionStyleEdgerunner:
		return 3
	case FashionStyleBusiness:
		return 3
	case FashionStyleHighFashion:
		return 4
	}
	return 1
}

func (s FashionStyle) String() string {
	switch s {
	case FashionStyleCombatGear:
		return "Combat Gear"
	case FashionStyleLowLife:
		return "Low Life"
	case FashionStyleChromeGang:
		return "Gang Colors"
	case FashionStyleNomadLeather:
		return "Nomad Leather"
	case FashionStyleGenericChic:
		return "Generic Chic"
	case FashionStyleEdgerunner:
		return "Edgerunner"
	case FashionStyleBusiness:
		return "Business"
	case FashionStyleHighFashion:
		return "High Fashion"
	}
	return "Unknown"

}

const FashionStyleRandom FashionStyle = -1
const (
	FashionStyleCombatGear FashionStyle = iota
	FashionStyleLowLife
	FashionStyleNomadLeather
	FashionStyleChromeGang
	FashionStyleGenericChic
	FashionStyleEdgerunner
	FashionStyleBusiness
	FashionStyleHighFashion
	FashionStyleCount
)

func FashionStyleFromString(str string) FashionStyle {
	switch strings.ToLower(str) {
	case "random":
		return FashionStyleRandom
	case "combat_gear":
		return FashionStyleCombatGear
	case "low_life":
		return FashionStyleLowLife
	case "gang_colors":
		return FashionStyleChromeGang
	case "nomad_leather":
		return FashionStyleNomadLeather
	case "generic_chic":
		return FashionStyleGenericChic
	case "edgerunner":
		return FashionStyleEdgerunner
	case "business":
		return FashionStyleBusiness
	case "high_fashion":
		return FashionStyleHighFashion
	default:
		return FashionStyleRandom
	}
}
