package d100

import (
	"github.com/memmaker/go/fxtools"
	"strconv"
	"strings"
)

type Perk int

const (
	PerkPickpocket Perk = iota
	PerkDisarm
	PerkNonLethalTakeDown
	PerkBackstab
	PerkQuickDraw
	PerkCharge
	PerkOneLiners
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
	case PerkQuickDraw:
		return "Quick Draw"
	case PerkOneLiners:
		return "One-Liners"
	case PerkCharge:
		return "Charge"
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
	case PerkQuickDraw:
		return "When you have only one holstered weapon, you can quickly draw it for a 50% damage bonus."
	case PerkOneLiners:
		return "You can drop one-liners to distract and intimidate enemies."
	case PerkCharge:
		return "You can charge at enemies for a forceful attack."
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
	case "quickdraw":
		return PerkQuickDraw
	case "one-liners":
		return PerkOneLiners
	case "charge":
		return PerkCharge
	}
	return PerkCount
}
func (p Perk) MaxLevel() int {
	return 1
}

type CharacterRequirement struct {
	Stats        map[Stat]int
	Skills       map[Skill]int
	DerivedStats map[DerivedStat]int
	Perks        map[Perk]int
	Level        int
}

func (r CharacterRequirement) String() string {
	var tableRows []fxtools.TableRow
	for s := Stat(0); s < StatCount; s++ {
		if v, ok := r.Stats[s]; ok {
			tableRows = append(tableRows, fxtools.NewTableRow(s.String(), strconv.Itoa(v)))
		}
	}
	for s := Skill(0); s < Skill(SkillCount()); s++ {
		if v, ok := r.Skills[s]; ok {
			tableRows = append(tableRows, fxtools.NewTableRow(s.String(), strconv.Itoa(v)))
		}
	}
	for s := DerivedStat(0); s < DerivedStatCount; s++ {
		if v, ok := r.DerivedStats[s]; ok {
			tableRows = append(tableRows, fxtools.NewTableRow(s.String(), strconv.Itoa(v)))
		}
	}
	for s := Perk(0); s < PerkCount; s++ {
		if v, ok := r.Perks[s]; ok {
			tableRows = append(tableRows, fxtools.NewTableRow(s.String(), strconv.Itoa(v)))
		}
	}
	lines := fxtools.TableLayoutLastRight(tableRows)
	return strings.Join(lines, "\n")
}

func LoadPerkRequirements(requirements map[Perk]CharacterRequirement) {
	defaultPerkRequirements = requirements
}

var defaultPerkRequirements = map[Perk]CharacterRequirement{}

func GetPerkRequirements(p Perk) CharacterRequirement {
	return defaultPerkRequirements[p]
}
