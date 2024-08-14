package special

import (
	"RogueUI/foundation"
	"cmp"
	"math"
	"slices"
	"strings"
)

// Requirements
// Factor in modifiers to all these stats by..
// 1. Equipment
// 2. Perks
// 3. Traits
// 4. Drugs
// 5. Environment
// 6. Status effects
// 7. Party members

func CostForOnePercentSkill(skillValue int) int {
	switch {
	case skillValue <= 100:
		return 1
	case skillValue >= 101 && skillValue <= 125:
		return 2
	case skillValue >= 126 && skillValue <= 150:
		return 3
	case skillValue >= 151 && skillValue <= 175:
		return 4
	case skillValue >= 176 && skillValue <= 200:
		return 5
	case skillValue >= 201 && skillValue <= 300:
		return 6
	}
	return 1000
}

type Stat int

const (
	Strength Stat = iota
	Perception
	Endurance
	Charisma
	Intelligence
	Agility
	Luck
)

type DerivedStat int

const (
	ActionPoints DerivedStat = iota
	Dodge
	CarryWeight
	CriticalChance
	DamageResistance
	HealingRate
	HitPoints
	MeleeDamage
	PartyLimit
	PerkRate
	EnergyResistance
	PoisonResistance
	RadiationResistance
	Speed
	SkillRate
)

type Skill int

func (s Skill) IsRangedAttackSkill() bool {
	return s == SmallGuns || s == BigGuns || s == EnergyWeapons || s == Throwing
}

func (s Skill) IsMeleeAttackSkill() bool {
	return s == Unarmed || s == MeleeWeapons
}

const (
	SmallGuns Skill = iota
	BigGuns
	EnergyWeapons
	Unarmed
	MeleeWeapons
	Throwing
	Doctor
	Sneak
	Lockpick
	Steal
	Traps
	Science
	Repair
	Speech
	Barter
	Gambling
	Outdoorsman
)

func SkillFromName(name string) Skill {
	name = strings.ToLower(name)
	switch name {
	case "small_guns":
		return SmallGuns
	case "big_guns":
		return BigGuns
	case "energy_weapons":
		return EnergyWeapons
	case "unarmed":
		return Unarmed
	case "melee_weapons":
		return MeleeWeapons
	case "throwing":
		return Throwing
	case "doctor":
		return Doctor
	case "sneak":
		return Sneak
	case "lockpick":
		return Lockpick
	case "steal":
		return Steal
	case "traps":
		return Traps
	case "science":
		return Science
	case "repair":
		return Repair
	case "speech":
		return Speech
	case "barter":
		return Barter
	case "gambling":
		return Gambling
	case "outdoorsman":
		return Outdoorsman
	}
	panic("invalid skill name")
	return 0
}
func NewCharSheet() *CharSheet {
	c := &CharSheet{
		stats: map[Stat]int{
			Strength:     5,
			Perception:   5,
			Endurance:    5,
			Charisma:     5,
			Intelligence: 5,
			Agility:      5,
			Luck:         5,
		},
		availableStatPoints:    5,
		derivedStatAdjustments: make(map[DerivedStat]int),
		skillAdjustments:       make(map[Skill]int),
		taggedSkills:           make(map[Skill]bool),
	}
	c.HealCompletely()
	return c
}

type CharSheet struct {
	level int

	availableStatPoints  int
	availableSkillPoints int
	availablePerks       int

	stats map[Stat]int

	derivedStatAdjustments map[DerivedStat]int
	skillAdjustments       map[Skill]int

	taggedSkills map[Skill]bool

	hitPointsCurrent int

	actionPointsCurrent int

	getStatMods        func(Stat) []Modifier
	getDerivedStatMods func(DerivedStat) []Modifier
	getSkillMods       func(Skill) []Modifier

	onStatChangedHandler        func(Stat)
	onDerivedStatChangedHandler func(DerivedStat)
	onSkillChangedHandler       func(Skill)
}

type Modifier interface {
	Description() string
	Apply(int) int
	SortOrder() int
}

func (cs *CharSheet) GetLevel() int {
	return cs.level
}

func (cs *CharSheet) LevelUp() {
	cs.level++
	cs.availableSkillPoints += cs.GetDerivedStat(SkillRate)
	cs.derivedStatAdjustments[HitPoints] = cs.getDerivedStatAdjustment(HitPoints) + cs.getHitPointIncrease()
	if cs.level%cs.GetDerivedStat(PerkRate) == 0 {
		cs.availablePerks++
	}
}

func (cs *CharSheet) getHitPointIncrease() int {
	return int(math.Floor(float64(cs.getStatBaseValue(Endurance))/2.0)) + 2
}
func (cs *CharSheet) GetStat(stat Stat) int {
	baseValue := cs.getStatBaseValue(stat)
	statValue := cs.onRetrieveStatHook(stat, baseValue)
	return statValue
}

func (cs *CharSheet) getStatBaseValue(stat Stat) int {
	return cs.stats[stat]
}

func (cs *CharSheet) SetStat(stat Stat, value int) {
	cs.stats[stat] = value
}

func (cs *CharSheet) GetSkill(skill Skill) int {
	baseValue := cs.getSkillBase(skill) + cs.getSkillAdjustment(skill)
	skillValue := cs.onRetrieveSkillHook(skill, baseValue)
	return skillValue
}

func (cs *CharSheet) getSkillAdjustment(skill Skill) int {
	if value, ok := cs.skillAdjustments[skill]; ok {
		return value
	}
	return 0
}

func (cs *CharSheet) getDerivedStatAdjustment(ds DerivedStat) int {
	if value, ok := cs.derivedStatAdjustments[ds]; ok {
		return value
	}
	return 0
}

func (cs *CharSheet) SetSkillAdjustment(skill Skill, value int) {
	cs.skillAdjustments[skill] = value
}

func (cs *CharSheet) GetDerivedStat(ds DerivedStat) int {
	baseValue := cs.getDerivedStatBaseValue(ds) + cs.getDerivedStatAdjustment(ds)
	derivedStatValue := cs.onRetrieveDerivedStatHook(ds, baseValue)
	return derivedStatValue
}

func (cs *CharSheet) getDerivedStatBaseValue(ds DerivedStat) int {
	switch ds {
	case ActionPoints:
		return 5 + cs.GetStat(Agility)/2
	case Dodge:
		return cs.GetStat(Agility) // Needs to factor in armor..
	case CarryWeight:
		return 25 + 25*cs.GetStat(Strength)
	case CriticalChance:
		return cs.GetStat(Luck)
	case DamageResistance:
		return 0
	case EnergyResistance:
		return 0
	case HealingRate:
		return max(1, cs.GetStat(Endurance)/3)
	case HitPoints:
		return 15 + (2 * cs.GetStat(Endurance)) + cs.GetStat(Strength)
	case MeleeDamage:
		return max(1, cs.GetStat(Strength)-5)
	case PartyLimit:
		return int(math.Floor(float64(cs.GetStat(Charisma)) / 2.0))
	case PerkRate:
		return 3
	case PoisonResistance:
		return cs.GetStat(Endurance) * 5
	case RadiationResistance:
		return cs.GetStat(Endurance) * 2
	case Speed:
		return 2 * cs.GetStat(Perception)
	case SkillRate:
		return 5 + (cs.GetStat(Intelligence) * 2)
	}
	panic("invalid derived stat")
	return 0
}

func (cs *CharSheet) getSkillBase(skill Skill) int {
	switch skill {
	case SmallGuns:
		return 5 + (cs.GetStat(Perception) * 4)
	case BigGuns:
		return cs.GetStat(Strength) + cs.GetStat(Perception) + 5
	case EnergyWeapons:
		return cs.GetStat(Perception) * 2
	case Unarmed:
		return 30 + (2 * (cs.GetStat(Agility) + cs.GetStat(Strength)))
	case MeleeWeapons:
		return 20 + (2 * (cs.GetStat(Agility) + cs.GetStat(Strength)))
	case Throwing:
		return 4 * cs.GetStat(Agility)
	case Doctor:
		return 5 + cs.GetStat(Perception) + cs.GetStat(Intelligence)
	case Sneak:
		return 5 + (3 * cs.GetStat(Agility))
	case Lockpick:
		return 10 + (cs.GetStat(Perception) + cs.GetStat(Agility))
	case Steal:
		return 3 * cs.GetStat(Agility)
	case Traps:
		return 10 + (cs.GetStat(Perception) + cs.GetStat(Agility))
	case Science:
		return 4 * cs.GetStat(Intelligence)
	case Repair:
		return 2*cs.GetStat(Intelligence) + cs.GetStat(Luck)
	case Speech:
		return 3*cs.GetStat(Charisma) + 2*cs.GetStat(Intelligence)
	case Barter:
		return 4 * cs.GetStat(Charisma)
	case Gambling:
		return 5 * cs.GetStat(Luck)
	case Outdoorsman:
		return 5 + cs.GetStat(Endurance) + cs.GetStat(Intelligence) + cs.GetStat(Luck)
	}
	panic("invalid skill")
	return 0
}

func (cs *CharSheet) HealCompletely() {
	cs.hitPointsCurrent = cs.GetDerivedStat(HitPoints)
	cs.onDerivedStatChanged(HitPoints)
}

func (cs *CharSheet) TakeRawDamage(damage int) {
	cs.hitPointsCurrent -= damage
	cs.onDerivedStatChanged(HitPoints)
}

func (cs *CharSheet) IsAlive() bool {
	return cs.hitPointsCurrent > 0
}

func (cs *CharSheet) IsHelpless() bool {
	return false // TODO: return status effect knocked out or knocked down
}

func (cs *CharSheet) IsBlinded() bool {
	return false // TODO: return status effect blinded
}

func (cs *CharSheet) onRetrieveStatHook(stat Stat, value int) int {
	if cs.getStatMods != nil {
		mods := cs.getStatMods(stat)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value
	}
	return value
}

func (cs *CharSheet) onRetrieveDerivedStatHook(ds DerivedStat, value int) int {
	if cs.getDerivedStatMods != nil {
		mods := cs.getDerivedStatMods(ds)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value
	}
	return value
}

func (cs *CharSheet) onRetrieveSkillHook(skill Skill, value int) int {
	if cs.getSkillMods != nil {
		mods := cs.getSkillMods(skill)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value
	}
	return value
}

func (cs *CharSheet) GetHitPointsMax() int {
	return cs.GetDerivedStat(HitPoints)
}

func (cs *CharSheet) GetHitPoints() int {
	return cs.hitPointsCurrent
}

func (cs *CharSheet) Heal(amount int) {
	cs.hitPointsCurrent = min(cs.hitPointsCurrent+amount, cs.GetHitPointsMax())
}

func (cs *CharSheet) GetActionPointsMax() int {
	return cs.GetDerivedStat(ActionPoints)
}

func (cs *CharSheet) GetActionPoints() int {
	return cs.actionPointsCurrent
}

func (cs *CharSheet) LooseActionPoints(amount int) {
	cs.actionPointsCurrent = max(0, cs.actionPointsCurrent-amount)
	cs.onDerivedStatChanged(ActionPoints)
}

func (cs *CharSheet) HasFlag(flagName foundation.ActorFlag) bool {
	return false
}

func (cs *CharSheet) AddSkillPoints(amount int) {
	cs.availableSkillPoints += amount
}

func (cs *CharSheet) SetOnStatChangeHandler(changed func(Stat)) {
	cs.onStatChangedHandler = changed
}

func (cs *CharSheet) onStatChanged(stat Stat) {
	if cs.onStatChangedHandler != nil {
		cs.onStatChangedHandler(stat)
	}
}
func (cs *CharSheet) SetOnDerivedStatChangeHandler(changed func(DerivedStat)) {
	cs.onDerivedStatChangedHandler = changed
}
func (cs *CharSheet) SetOnSkillChangeHandler(changed func(Skill)) {
	cs.onSkillChangedHandler = changed
}
func (cs *CharSheet) onDerivedStatChanged(ds DerivedStat) {
	if cs.onDerivedStatChangedHandler != nil {
		cs.onDerivedStatChangedHandler(ds)
	}
}
func (cs *CharSheet) onSkillChanged(skill Skill) {
	if cs.onSkillChangedHandler != nil {
		cs.onSkillChangedHandler(skill)
	}
}

func (cs *CharSheet) SetDerivedStatAbsoluteValue(stat DerivedStat, value int) {
	currentValue := cs.GetDerivedStat(stat)
	if currentValue == value {
		return
	}
	delta := value - currentValue
	cs.derivedStatAdjustments[stat] = delta
}

func (cs *CharSheet) SetSkillAbsoluteValue(skill Skill, value int) {
	currentValue := cs.GetSkill(skill)
	if currentValue == value {
		return
	}
	delta := value - currentValue
	cs.skillAdjustments[skill] = delta
}

func (cs *CharSheet) Kill() {
	cs.hitPointsCurrent = 0
	cs.onDerivedStatChanged(HitPoints)
}

func (cs *CharSheet) IsSkillHigherOrEqual(skill Skill, difficulty int) bool {
	return cs.GetSkill(skill) >= difficulty
}
func (cs *CharSheet) SkillRoll(skill Skill, modifiers int) foundation.CheckResult {
	critChance := cs.GetStat(Luck)
	return SuccessRoll(Percentage(cs.GetSkill(skill)+modifiers), Percentage(critChance))
}

func (cs *CharSheet) StatRoll(stat Stat, modifiers int) foundation.CheckResult {
	critChance := cs.GetStat(Luck)
	statSkill := (cs.GetStat(stat) * 10) + modifiers
	return SuccessRoll(Percentage(statSkill), Percentage(critChance))
}
func (cs *CharSheet) IsStatHigherOrEqual(stat Stat, difficulty int) bool {
	return cs.GetStat(stat) >= difficulty
}
