package d100

import "strings"

type Stat int

func (s Stat) ToShortString() string {
	switch s {
	case Strength:
		return "ST"
	case Perception:
		return "PE"
	case Endurance:
		return "EN"
	case Charisma:
		return "CH"
	case Intelligence:
		return "IN"
	case Agility:
		return "AG"
	}
	return ""
}

func (s Stat) GetDescription() string {
	switch s {
	case Strength:
		return "Strength measures the raw physical power of your character. It affects how much you can carry, and the damage of all melee attacks."
	case Perception:
		return "Perception affects your ranged combat skills, and your ability to detect traps and enemies."
	case Endurance:
		return "Endurance affects your Hit Points, Poison Resistance, and Radiation Resistance."
	case Charisma:
		return "Charisma affects your ability to negotiate, and the size of your party."
	case Intelligence:
		return "Intelligence affects the number of skill points you receive when you level up, and the number of new Perks you can choose."
	case Agility:
		return "Agility affects your Action Points, and your ability to dodge attacks."
	}
	return ""
}

func (s Stat) String() string {
	switch s {
	case Strength:
		return "Strength"
	case Perception:
		return "Perception"
	case Endurance:
		return "Endurance"
	case Charisma:
		return "Charisma"
	case Intelligence:
		return "Intelligence"
	case Agility:
		return "Agility"
	}
	return ""
}

const (
	Strength Stat = iota
	Perception
	Endurance
	Charisma
	Intelligence
	Agility
	StatCount
)

func StatFromString(name string) Stat {
	name = strings.ToLower(name)
	switch name {
	case "strength":
		return Strength
	case "perception":
		return Perception
	case "endurance":
		return Endurance
	case "charisma":
		return Charisma
	case "intelligence":
		return Intelligence
	case "agility":
		return Agility
	}
	panic("invalid stat name")
	return 0
}
