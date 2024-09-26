package special

import (
	"bytes"
	"cmp"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/recfile"
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
const SkillCap = 200

type Stat int

func (s Stat) ToShortString() string {
	switch s {
	case Strength:
		return "STR"
	case Perception:
		return "PER"
	case Endurance:
		return "END"
	case Charisma:
		return "CHA"
	case Intelligence:
		return "INT"
	case Agility:
		return "AGI"
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

type DerivedStat int

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
	case DamageResistance:
		return "Damage Resistance"
	case EnergyResistance:
		return "Energy Resistance"
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
	case RadiationResistance:
		return "Radiation Resistance"
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
	case "damageresistance":
		return DamageResistance
	case "energyresistance":
		return EnergyResistance
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
	case "radiationresistance":
		return RadiationResistance
	case "speed":
		return Speed
	case "skillrate":
		return SkillRate
	}
	panic("invalid derived stat name")
	return -1
}

func (s DerivedStat) ToShortString() any {
	switch s {
	case ActionPoints:
		return "AP"
	case Dodge:
		return "DG"
	case CarryWeight:
		return "CW"
	case CriticalChance:
		return "CC"
	case DamageResistance:
		return "DR"
	case EnergyResistance:
		return "ER"
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
	case RadiationResistance:
		return "RR"
	case Speed:
		return "SP"
	case SkillRate:
		return "SR"
	}
	return ""
}

const (
	ActionPoints DerivedStat = iota
	HitPoints
	HealingRate
	Speed
	Dodge
	CarryWeight
	CriticalChance
	MeleeDamageBonus
	DamageResistance
	EnergyResistance
	PoisonResistance
	RadiationResistance
	PartyLimit
	SkillRate
	PerkRate
	DerivedStatCount
)

var VisibleDerivedStatCount = 12

type Skill int

const (
	MeleeCombat Skill = iota
	RangedCombat

	Social
	Intimidate

	Stealth
	Mechanics

	Biology
	Technology

	SkillCount
)

func (s Skill) ToShortString() string {
	switch s {
	case MeleeCombat:
		return "ML"
	case RangedCombat:
		return "RG"

	case Social:
		return "SO"
	case Intimidate:
		return "IN"

	case Stealth:
		return "ST"
	case Mechanics:
		return "MC"

	case Biology:
		return "BI"
	case Technology:
		return "TC"

	}
	return ""
}

func (s Skill) ToAdjustmentString() string {
	return fmt.Sprintf("%s_Adjustment", strings.ReplaceAll(s.String(), " ", ""))
}

func SkillFromAdjustmentString(name string) Skill {
	name = strings.TrimSuffix(strings.ToLower(name), "_adjustment")
	return skillFromShortString(name)
}

func (s Skill) String() string {
	switch s {
	case RangedCombat:
		return "Ranged Combat"
	case MeleeCombat:
		return "Melee Combat"
	case Biology:
		return "Biology"
	case Stealth:
		return "Stealth"
	case Technology:
		return "Technology"
	case Mechanics:
		return "Mechanics"
	case Intimidate:
		return "Intimidate"
	case Social:
		return "Social"
	}
	return ""
}

func skillFromShortString(name string) Skill {
	switch name {
	case "rangedcombat":
		return RangedCombat
	case "meleecombat":
		return MeleeCombat
	case "biology":
		return Biology
	case "stealth":
		return Stealth
	case "technology":
		return Technology
	case "mechanics":
		return Mechanics
	case "intimidate":
		return Intimidate
	case "social":
		return Social
	}
	panic(fmt.Sprintf("invalid skill name: '%s'", name))
	return -1
}

func SkillFromString(name string) Skill {
	name = strings.ReplaceAll(strings.ToLower(name), "_", "")
	return skillFromShortString(name)
}

func SkillFromBonusString(name string) Skill {
	name = strings.TrimPrefix(strings.ToLower(name), "skillbonus")
	return skillFromShortString(name)
}

func (s Skill) IsRangedAttackSkill() bool {
	return s == RangedCombat
}

func (s Skill) IsMeleeAttackSkill() bool {
	return s == MeleeCombat
}

func (cs *CharSheet) getSkillBase(skill Skill) int {
	// Str: 2x
	// Int: 2x
	// Per: 2x
	// End: 2x
	// Cha: 2x
	// Agi: 2x
	multiplier := 5
	switch skill {
	case MeleeCombat:
		return ((cs.GetStat(Strength)) + (cs.GetStat(Endurance))) * multiplier
	case RangedCombat:
		return (cs.GetStat(Perception) + 5) * multiplier

	case Social:
		return (cs.GetStat(Charisma) + 5) * multiplier
	case Intimidate:
		return ((cs.GetStat(Charisma)) + (cs.GetStat(Strength))) * multiplier

	case Stealth:
		return (cs.GetStat(Agility) + 5) * multiplier
	case Mechanics:
		return ((cs.GetStat(Agility)) + (cs.GetStat(Endurance))) * multiplier

	case Biology:
		return ((cs.GetStat(Intelligence)) + (cs.GetStat(Perception))) * multiplier
	case Technology:
		return (cs.GetStat(Intelligence) + 5) * multiplier
		/*
			case Barter:
				return 4 * cs.GetStat(Charisma)
			case Gambling:
				return 5 * cs.GetStat(Luck)
			case Outdoorsman:
				return 5 + cs.GetStat(Endurance) + cs.GetStat(Intelligence) + cs.GetStat(Luck)
		*/
	}
	panic("invalid skill")
	return 0
}

// CreateReports: If the task at hand is simply not possible for someone without a certain level of skill
// Roll: If the task at hand is in general possible, even if it's difficult

// tagging gives 20% bonus to skill
func NewCharSheet() *CharSheet {
	c := &CharSheet{
		stats: map[Stat]int{
			Strength:     5,
			Perception:   5,
			Endurance:    5,
			Charisma:     5,
			Intelligence: 5,
			Agility:      5,
		},
		availableStatPoints:    0,
		derivedStatAdjustments: make(map[DerivedStat]int),
		skillAdjustments:       make(map[Skill]int),
		taggedSkills:           make(map[Skill]bool),
		level:                  1,
	}
	c.HealAPAndHPCompletely()
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
	xp                          int
}

// GobEncode encodes the CharSheet struct into a byte slice.
func (cs *CharSheet) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Encode each field of the struct in order
	if err := encoder.Encode(cs.level); err != nil {
		return nil, err
	}
	if err := encoder.Encode(cs.availableStatPoints); err != nil {
		return nil, err
	}
	if err := encoder.Encode(cs.availableSkillPoints); err != nil {
		return nil, err
	}
	if err := encoder.Encode(cs.availablePerks); err != nil {
		return nil, err
	}
	if err := encoder.Encode(cs.stats); err != nil {
		return nil, err
	}
	if err := encoder.Encode(cs.derivedStatAdjustments); err != nil {
		return nil, err
	}
	if err := encoder.Encode(cs.skillAdjustments); err != nil {
		return nil, err
	}
	if err := encoder.Encode(cs.taggedSkills); err != nil {
		return nil, err
	}
	if err := encoder.Encode(cs.hitPointsCurrent); err != nil {
		return nil, err
	}
	if err := encoder.Encode(cs.actionPointsCurrent); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GobDecode decodes a byte slice into a CharSheet struct.
func (cs *CharSheet) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)

	// Decode each field of the struct in order
	if err := decoder.Decode(&cs.level); err != nil {
		return err
	}
	if err := decoder.Decode(&cs.availableStatPoints); err != nil {
		return err
	}
	if err := decoder.Decode(&cs.availableSkillPoints); err != nil {
		return err
	}
	if err := decoder.Decode(&cs.availablePerks); err != nil {
		return err
	}
	if err := decoder.Decode(&cs.stats); err != nil {
		return err
	}
	if err := decoder.Decode(&cs.derivedStatAdjustments); err != nil {
		return err
	}
	if err := decoder.Decode(&cs.skillAdjustments); err != nil {
		return err
	}
	if err := decoder.Decode(&cs.taggedSkills); err != nil {
		return err
	}
	if err := decoder.Decode(&cs.hitPointsCurrent); err != nil {
		return err
	}
	if err := decoder.Decode(&cs.actionPointsCurrent); err != nil {
		return err
	}

	return nil
}

type DefaultModifier struct {
	Source    string
	Modifier  int
	Order     int
	IsPercent bool
	Suffix    string
}

func (d DefaultModifier) Description() string {
	var line string
	if d.IsPercent {
		line = fmt.Sprintf("%s: %+d%%", d.Source, d.Modifier)
	}
	line = fmt.Sprintf("%s: %+d", d.Source, d.Modifier)
	if d.Suffix != "" {
		line += " " + d.Suffix
	}
	return line
}

func (d DefaultModifier) Apply(i int) int {
	return i + d.Modifier
}

func (d DefaultModifier) SortOrder() int {
	return d.Order
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

func (cs *CharSheet) GetStatWithModInfo(stat Stat) (int, []Modifier) {
	baseValue := cs.getStatBaseValue(stat)
	return cs.getModifiedStatWithInfo(stat, baseValue)
}

func (cs *CharSheet) GetDerivedStatWithModInfo(ds DerivedStat) (int, []Modifier) {
	baseValue := cs.getDerivedStatBaseValue(ds)
	return cs.getModifiedDerivedStatWithInfo(ds, baseValue)
}

func (cs *CharSheet) getStatBaseValue(stat Stat) int {
	return cs.stats[stat]
}

func (cs *CharSheet) SetStat(stat Stat, value int) {
	cs.stats[stat] = value
}

func (cs *CharSheet) GetSkill(skill Skill) int {
	baseValue := cs.GetUnmodifiedSkill(skill)
	skillValue := cs.onRetrieveSkillHook(skill, baseValue)
	return skillValue
}
func (cs *CharSheet) IsSkillAtCap(skill Skill) bool {
	return cs.GetUnmodifiedSkill(skill) >= SkillCap
}
func (cs *CharSheet) GetSkillWithModInfo(skill Skill) (int, []Modifier) {
	baseValue := cs.GetUnmodifiedSkill(skill)
	return cs.getModifiedSkillWithInfo(skill, baseValue)
}

func (cs *CharSheet) GetUnmodifiedSkill(skill Skill) int {
	return cs.getSkillBase(skill) + cs.getSkillAdjustment(skill) + cs.getTagSkillBonus(skill)
}

func (cs *CharSheet) getTagSkillBonus(skill Skill) int {
	if cs.taggedSkills[skill] {
		return 20
	}
	return 0
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
		return 5 + cs.GetStat(Agility) + (cs.GetStat(Endurance) / 2)
	case Dodge:
		return cs.GetStat(Agility) // Needs to factor in armor..
	case CarryWeight:
		return 20 + 15*cs.GetStat(Strength)
	case CriticalChance:
		return 5
	case DamageResistance:
		return 0
	case EnergyResistance:
		return 0
	case HealingRate:
		return max(1, cs.GetStat(Endurance)/3)
	case HitPoints:
		return 15 + (2 * cs.GetStat(Endurance)) + cs.GetStat(Strength)
	case MeleeDamageBonus:
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
		return 2 * cs.GetStat(Agility)
	case SkillRate:
		return 3 + (cs.GetStat(Intelligence) * 2)
	}
	panic("invalid derived stat")
	return 0
}

func (cs *CharSheet) HealAPAndHPCompletely() {
	cs.hitPointsCurrent = cs.GetDerivedStat(HitPoints)
	cs.actionPointsCurrent = cs.GetDerivedStat(ActionPoints)

	cs.onDerivedStatChanged(HitPoints)
	cs.onDerivedStatChanged(ActionPoints)
}

func (cs *CharSheet) TakeRawDamage(damage int) {
	cs.hitPointsCurrent -= damage
	cs.onDerivedStatChanged(HitPoints)
}

func (cs *CharSheet) IsAlive() bool {
	return cs.hitPointsCurrent > 0
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

func (cs *CharSheet) getModifiedStatWithInfo(stat Stat, value int) (int, []Modifier) {
	if cs.getStatMods != nil {
		mods := cs.getStatMods(stat)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value, mods
	}
	return value, nil
}
func (cs *CharSheet) getModifiedDerivedStatWithInfo(ds DerivedStat, value int) (int, []Modifier) {
	if cs.getDerivedStatMods != nil {
		mods := cs.getDerivedStatMods(ds)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value, mods
	}
	return value, nil
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

func (cs *CharSheet) getModifiedSkillWithInfo(skill Skill, value int) (int, []Modifier) {
	if cs.getSkillMods != nil {
		mods := cs.getSkillMods(skill)

		slices.SortStableFunc(mods, func(i, j Modifier) int {
			return cmp.Compare(i.SortOrder(), j.SortOrder())
		})

		for _, mod := range mods {
			value = mod.Apply(value)
		}
		return value, mods
	}
	return value, nil
}

func (cs *CharSheet) GetHitPointsMax() int {
	return cs.GetDerivedStat(HitPoints)
}

func (cs *CharSheet) GetHitPoints() int {
	return min(cs.hitPointsCurrent, cs.GetHitPointsMax())
}

func (cs *CharSheet) Heal(amount int) {
	cs.hitPointsCurrent = min(cs.hitPointsCurrent+amount, cs.GetHitPointsMax())
}

func (cs *CharSheet) GetActionPointsMax() int {
	return cs.GetDerivedStat(ActionPoints)
}

func (cs *CharSheet) GetActionPoints() int {
	return min(cs.actionPointsCurrent, cs.GetActionPointsMax())
}

func (cs *CharSheet) LooseActionPoints(amount int) {
	cs.actionPointsCurrent = max(0, cs.actionPointsCurrent-amount)
	cs.onDerivedStatChanged(ActionPoints)
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
func (cs *CharSheet) SkillRoll(skill Skill, modifiers int) CheckResult {
	critChance := cs.GetDerivedStat(CriticalChance)
	return SuccessRoll(Percentage(max(0, min(95, cs.GetSkill(skill)+modifiers))), Percentage(critChance))
}

func (cs *CharSheet) StatRoll(stat Stat, modifiers int) CheckResult {
	critChance := cs.GetDerivedStat(CriticalChance)
	statSkill := max(0, min(95, (cs.GetStat(stat)*10)+modifiers))
	return SuccessRoll(Percentage(statSkill), Percentage(critChance))
}
func (cs *CharSheet) IsStatHigherOrEqual(stat Stat, difficulty int) bool {
	return cs.GetStat(stat) >= difficulty
}

func (cs *CharSheet) GetHitPointsString() string {
	return fmt.Sprintf("%d/%d", cs.GetHitPoints(), cs.GetHitPointsMax())
}

func (cs *CharSheet) ToRecord() recfile.Record {
	record := recfile.Record{
		recfile.Field{Name: "Level", Value: recfile.IntStr(cs.GetLevel())},
		recfile.Field{Name: "AvailableStatPoints", Value: recfile.IntStr(cs.availableStatPoints)},
		recfile.Field{Name: "AvailableSkillPoints", Value: recfile.IntStr(cs.availableSkillPoints)},
		recfile.Field{Name: "AvailablePerks", Value: recfile.IntStr(cs.availablePerks)},
		recfile.Field{Name: "Strength", Value: recfile.IntStr(cs.GetStat(Strength))},
		recfile.Field{Name: "Perception", Value: recfile.IntStr(cs.GetStat(Perception))},
		recfile.Field{Name: "Endurance", Value: recfile.IntStr(cs.GetStat(Endurance))},
		recfile.Field{Name: "Charisma", Value: recfile.IntStr(cs.GetStat(Charisma))},
		recfile.Field{Name: "Intelligence", Value: recfile.IntStr(cs.GetStat(Intelligence))},
		recfile.Field{Name: "Agility", Value: recfile.IntStr(cs.GetStat(Agility))},
		recfile.Field{Name: "HitPoints", Value: recfile.IntStr(cs.GetHitPoints())},
		recfile.Field{Name: "ActionPoints", Value: recfile.IntStr(cs.GetActionPoints())},
	}
	// add skills
	for skillNo := 0; skillNo < int(SkillCount); skillNo++ {
		skill := Skill(skillNo)
		record = append(record, recfile.Field{Name: skill.ToAdjustmentString(), Value: recfile.IntStr(cs.getSkillAdjustment(skill))})
	}

	return record
}

func (cs *CharSheet) HasStatPointsToSpend() bool {
	return cs.availableStatPoints > 0
}

func (cs *CharSheet) GetStatPointsToSpend() int {
	return cs.availableStatPoints
}

func (cs *CharSheet) IsTagSkill(skill Skill) bool {
	return cs.taggedSkills[skill]
}

func (cs *CharSheet) HasSkillPointsToSpend() bool {
	return cs.availableSkillPoints > 0
}

func (cs *CharSheet) TagSkill(skill Skill) {
	if cs.GetTagSkillCount() >= 3 {
		return
	}
	cs.taggedSkills[skill] = true
}

func (cs *CharSheet) UntagSkill(skill Skill) {
	delete(cs.taggedSkills, skill)
}

func (cs *CharSheet) GetTagSkillCount() int {
	return len(cs.taggedSkills)
}

func (cs *CharSheet) SpendStatPoint(stat Stat) {
	if cs.availableStatPoints <= 0 {
		return
	}
	if cs.stats[stat] >= 10 {
		return
	}
	cs.stats[stat]++
	cs.availableStatPoints--
	cs.onStatChanged(stat)
}

func (cs *CharSheet) RefundStatPoint(stat Stat) {
	if cs.stats[stat] <= 1 {
		return
	}
	cs.stats[stat]--
	cs.availableStatPoints++
	cs.onStatChanged(stat)
}

func (cs *CharSheet) GetSkillPointsToSpend() int {
	return cs.availableSkillPoints
}

func (cs *CharSheet) ResetStatPoints() {
	for stat := range cs.stats {
		cs.stats[stat] = 5
	}
	cs.availableStatPoints = 5
}

func (cs *CharSheet) ResetTagSkills() {
	cs.taggedSkills = make(map[Skill]bool)
}

func (cs *CharSheet) SpendSkillPoints(skill Skill, points int) {
	if cs.availableSkillPoints < points {
		return
	}
	if cs.IsSkillAtCap(skill) {
		return
	}
	cs.availableSkillPoints -= points
	isTagSkill := cs.IsTagSkill(skill)
	if isTagSkill {
		points = points * 2
	}
	cs.skillAdjustments[skill] += points
	cs.onSkillChanged(skill)
}

func (cs *CharSheet) AddXP(xp int) bool {
	if xp <= 0 {
		return false
	}
	wasAbleToLevelBefore := cs.CanLevelUp()
	cs.xp += xp

	if cs.CanLevelUp() && !wasAbleToLevelBefore {
		cs.LevelUp()
		return true
	}
	return false
}

func (cs *CharSheet) CanLevelUp() bool {
	return cs.xp >= cs.GetTotalXPForNextLevel(cs.level)
}

func (cs *CharSheet) GetTotalXPForNextLevel(currentLevel int) int {
	if currentLevel > 21 {
		// (n*(n-1)/2) * 1,000 XP
		return (currentLevel * (currentLevel - 1) / 2) * 1000
	}
	return []int{
		0,
		1000,
		3000,
		6000,
		10000,
		15000,
		21000,
		28000,
		36000,
		45000,
		55000,
		66000,
		78000,
		91000,
		105000,
		120000,
		136000,
		153000,
		171000,
		190000,
		210000,
	}[currentLevel]
}

func (cs *CharSheet) GetCurrentXP() int {
	return cs.xp
}

func (cs *CharSheet) GetXPNeededForNextLevel() int {
	return cs.GetTotalXPForNextLevel(cs.level) - cs.xp
}

func (cs *CharSheet) SetSkillModifierHandler(handler func(skill Skill) []Modifier) {
	cs.getSkillMods = handler
}

func (cs *CharSheet) SetStatModifierHandler(handler func(stat Stat) []Modifier) {
	cs.getStatMods = handler
}

func (cs *CharSheet) SetDerivedStatModifierHandler(handler func(ds DerivedStat) []Modifier) {
	cs.getDerivedStatMods = handler
}

func (cs *CharSheet) AddSkillPointsTo(skill Skill, increase int) {
	if cs.IsSkillAtCap(skill) {
		return
	}
	cs.skillAdjustments[skill] = cs.skillAdjustments[skill] + increase
}
