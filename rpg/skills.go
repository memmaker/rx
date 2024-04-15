package rpg

import (
	"fmt"
	"strings"
)

const (
	SkillNameShield         SkillName = "Shield"
	SkillNameMeleeWeapons   SkillName = "Melee Weapons"
	SkillNameBrawling       SkillName = "Brawling"
	SkillNameThrowing       SkillName = "Throwing"
	SkillNameMissileWeapons SkillName = "Missile Weapons"
	SkillNameStealth        SkillName = "Stealth"
)

func SkillNameFromString(s string) SkillName {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "shield":
		return SkillNameShield
	case "melee":
		return SkillNameMeleeWeapons
	case "brawling":
		return SkillNameBrawling
	case "throwing":
		return SkillNameThrowing
	case "missile":
		return SkillNameMissileWeapons
	case "stealth":
		return SkillNameStealth
	}
	panic(fmt.Sprintf("Unknown skill name: %s", s))
	return ""
}
func (n SkillName) GetDefaultValue(getCharStat func(stat Stat) int) int {
	switch n {
	case SkillNameShield:
		return getCharStat(n.ControllingAttribute()) - 6
	case SkillNameMeleeWeapons:
		return getCharStat(n.ControllingAttribute()) - 6
	case SkillNameBrawling:
		return getCharStat(n.ControllingAttribute()) - 6
	case SkillNameThrowing:
		return getCharStat(n.ControllingAttribute()) - 6
	case SkillNameMissileWeapons:
		return getCharStat(n.ControllingAttribute()) - 6
	case SkillNameStealth:
		return getCharStat(n.ControllingAttribute()) - 6
	}
	return 0
}

func (n SkillName) Difficulty() SkillDifficulty {
	switch n {
	case SkillNameShield:
		return DiffAverage
	case SkillNameMeleeWeapons:
		return DiffAverage
	case SkillNameBrawling:
		return DiffAverage
	case SkillNameThrowing:
		return DiffAverage
	case SkillNameMissileWeapons:
		return DiffAverage
	case SkillNameStealth:
		return DiffAverage
	}
	return DiffAverage
}

func (n SkillName) ControllingAttribute() Stat {
	switch n {
	case SkillNameShield:
		return Strength
	case SkillNameMeleeWeapons:
		return Dexterity
	case SkillNameBrawling:
		return Strength
	case SkillNameThrowing:
		return Strength
	case SkillNameMissileWeapons:
		return Dexterity
	case SkillNameStealth:
		return Dexterity
	}
	return Strength
}

type SkillName string

func (n SkillName) PointsSpentFromLevel(level int) int {
	if level == 0 {
		return 0
	}
	if level == 1 || level == 2 {
		return 1 * level
	}
	if level == 3 {
		return 4
	}
	return 4 + ((level - 3) * 4)
}

func (n SkillName) GetValueFromLevel(level int, getCharStat func(stat Stat) int) int {
	if level == 0 {
		return n.GetDefaultValue(getCharStat)
	}
	relativeSkillLevel := getRelativeSkillLevel(n, level)
	return relativeSkillLevel + getCharStat(n.ControllingAttribute())
}

type SkillDifficulty int

const (
	DiffEasy SkillDifficulty = iota
	DiffAverage

	DiffHard
	DiffVeryHard
)

func getRelativeSkillLevel(skillInfo SkillName, levelsAboveStart int) int {
	startOffset := 0
	if skillInfo.Difficulty() == DiffEasy {
		startOffset = 0
	} else if skillInfo.Difficulty() == DiffAverage {
		startOffset = -1
	} else if skillInfo.Difficulty() == DiffHard {
		startOffset = -2
	} else if skillInfo.Difficulty() == DiffVeryHard {
		startOffset = -3
	}

	currentLevel := startOffset + levelsAboveStart
	return currentLevel
}
