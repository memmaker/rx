package dice_curve

import (
	"math/rand"
	"strings"
)

type Counter int

const (
	CounterGold Counter = iota
	CounterHunger
)

type Stat int

func (s Stat) IsDerived() bool {
	return s == Will ||
		s == Perception ||
		s == BasicSpeed ||
		s == FatiguePoints ||
		s == HitPoints ||
		s == BasicLift ||
		s == Dodge ||
		s == Block ||
		s == Parry ||
		s == MaximumLoad
}

func (s Stat) CanBeBought() bool {
	return s == Strength ||
		s == Dexterity ||
		s == Intelligence ||
		s == Health ||
		s == Will ||
		s == Perception ||
		s == BasicSpeed ||
		s == FatiguePoints ||
		s == HitPoints
}

func (s Stat) AdjustmentPerLevel() int {
	return 1
}
func (s Stat) getAdjustment(levels int) int {
	totalAdjustment := int(levels) * s.AdjustmentPerLevel()
	return totalAdjustment
}

func (s Stat) GetDerivedValueFromLevelAdjustments(levels int, getCharStat func(stat Stat) int, getSkill func(skill SkillName) int) int {
	adjustment := s.getAdjustment(levels)
	switch s {
	case Will:
		will := getCharStat(Intelligence) + adjustment
		return will
	case Perception:
		perception := getCharStat(Intelligence) + adjustment
		return perception
	case BasicSpeed:
		speed := ((getCharStat(Dexterity) + getCharStat(Health)) / 2) + adjustment
		return speed
	case FatiguePoints:
		fatigue := 1 + (getCharStat(Health) / 5) + adjustment
		return fatigue
	case HitPoints:
		hp := getCharStat(Health) + adjustment
		return hp
	case Dodge:
		dodge := 3 + (getCharStat(BasicSpeed) / 2)
		return dodge
	case Block:
		block := 3 + (getSkill(SkillNameShield) / 2)
		return block
	case BasicLift:
		str := getCharStat(Strength)
		basicLift := (str * str) / 5
		return basicLift
	}
	return 0
}

func (s Stat) GetNonDerivedValueFromLevelAdjustment(levels int) int {
	totalAdjustment := s.getAdjustment(levels)
	defaultValue := 10
	return int(defaultValue + totalAdjustment)
}

func (s Stat) CostPerLevel() int {
	switch s {
	case Strength:
		return 10
	case Dexterity:
		return 20
	case Intelligence:
		return 20
	case Health:
		return 10
	case Will:
		return 5
	case Perception:
		return 5
	case BasicSpeed:
		return 20
	case FatiguePoints:
		return 3
	case HitPoints:
		return 2
	}
	return 0
}

func (s Stat) ToString() string {
	switch s {
	case Strength:
		return "Strength"
	case Dexterity:
		return "Dexterity"
	case Intelligence:
		return "Intelligence"
	case Health:
		return "Health"
	case Will:
		return "Will"
	case Perception:
		return "Perception"
	case BasicSpeed:
		return "Basic Speed"
	case FatiguePoints:
		return "Fatigue Points"
	case HitPoints:
		return "Hit Points"
	case Dodge:
		return "Dodge"
	case Block:
		return "Block"
	case Parry:
		return "Parry"
	case MaximumLoad:
		return "Maximum Load"
	}
	return "Unknown"
}

func (s Stat) ToShortString() string {
	switch s {
	case Strength:
		return "ST"
	case Dexterity:
		return "DX"
	case Intelligence:
		return "IQ"
	case Health:
		return "HT"
	case Will:
		return "Will"
	case Perception:
		return "Per"
	case BasicSpeed:
		return "Spd"
	case FatiguePoints:
		return "FP"
	case HitPoints:
		return "HP"
	case Dodge:
		return "Dodge"
	case Block:
		return "Block"
	case Parry:
		return "Parry"
	case MaximumLoad:
		return "Max Load"
	}
	return "Unknown"
}

// normal range 1-20
// 10 = average human
// ST can go much higher
const (
	Strength Stat = iota
	Dexterity
	Intelligence
	Health
	// Will = IQ
	Will
	// Perception = IQ
	Perception
	BasicSpeed
	BasicLift
	// BasicMove = BasicSpeed
	FatiguePoints
	HitPoints
	Dodge
	Block
	Parry
	MaximumLoad
	// Pseudo Stats
	ActiveDefense
	DamageResistance
)

func GetRandomStat() Stat {
	return Stat(rand.Intn(int(MaximumLoad) + 1))
}

func StatFromString(stat string) Stat {
	stat = strings.ToLower(stat)
	switch stat {
	case "strength":
		return Strength
	case "dexterity":
		return Dexterity
	case "intelligence":
		return Intelligence
	case "health":
		return Health
	case "will":
		return Will
	case "perception":
		return Perception
	case "basicspeed":
		return BasicSpeed
	case "fatiguepoints":
		return FatiguePoints
	case "hitpoints":
		return HitPoints
	case "dodge":
		return Dodge
	case "block":
		return Block
	case "parry":
		return Parry
	case "damage_resistance":
		return DamageResistance
	case "maximumload":
		return MaximumLoad
	}
	panic("Invalid stat: " + stat)
	return -1
}

var probabilityTable = map[int]float64{
	3:  0.0046,
	4:  0.0185,
	5:  0.0463,
	6:  0.0926,
	7:  0.162,
	8:  0.2593,
	9:  0.3750,
	10: 0.5,
	11: 0.625,
	12: 0.7407,
	13: 0.8380,
	14: 0.9074,
	15: 0.9537,
	16: 0.9815,
	17: 0.9948,
	18: 1,
}

func ChanceOfSuccess(effectiveSkill int) float64 {
	if effectiveSkill < 3 {
		return 0
	}
	if effectiveSkill > 18 {
		return 1
	}
	return probabilityTable[effectiveSkill]
}
