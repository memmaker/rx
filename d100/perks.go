package d100

import "strings"

type Perk int

const (
	PerkPickpocket Perk = iota
	PerkDisarm
	PerkNonLethalTakeDown
	PerkBackstab
	PerkCount
)

func (p Perk) String() string {
	switch p {
	case PerkPickpocket:
		return "Pickpocket"
	case PerkDisarm:
		return "Disarm"
	case PerkNonLethalTakeDown:
		return "Takedown"
	case PerkBackstab:
		return "Backstab"
	}
	return "Unknown"
}

func (p Perk) Description() string {
	switch p {
	case PerkPickpocket:
		return "You can steal from enemies."
	case PerkDisarm:
		return "You can disarm enemies."
	case PerkNonLethalTakeDown:
		return "You can perform non-lethal takedowns."
	case PerkBackstab:
		return "You can backstab enemies."
	}
	return "Unknown"
}
func PerkFromString(s string) Perk {
	switch strings.ToLower(s) {
	case "pickpocket":
		return PerkPickpocket
	case "disarm":
		return PerkDisarm
	case "takedown":
		return PerkNonLethalTakeDown
	case "backstab":
		return PerkBackstab
	}
	return PerkCount
}
func (p Perk) MaxLevel() int {
	return 1
}

type PerkRequirements struct {
	Stats        map[Stat]int
	Skills       map[Skill]int
	DerivedStats map[DerivedStat]int
	Perks        map[Perk]int
}

func LoadPerkRequirements(requirements map[Perk]PerkRequirements) {
	defaultPerkRequirements = requirements
}

var defaultPerkRequirements = map[Perk]PerkRequirements{}

func GetPerkRequirements(p Perk) PerkRequirements {
	return defaultPerkRequirements[p]
}
