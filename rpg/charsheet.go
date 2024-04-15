package rpg

import (
	"cmp"
	"fmt"
	"slices"
)

type Modifier interface {
	Description() string
	Apply(int) int
	IsApplicable() bool
	IsPersistent() bool
	SortOrder() int
}

// track points spent

type Character struct {
	// Bought stats
	levelAdjustments map[Stat]int
	skillLevels      map[SkillName]int

	// Current state
	resources map[Stat]int

	// Lookups
	statModifiers  map[Stat][]Modifier
	skillModifiers map[SkillName][]Modifier

	characterPointsReceived int
	statChangedHandler      func()
	resourceChangedHandler  func()
}

func NewCharacterSheet() *Character {
	return &Character{
		levelAdjustments: map[Stat]int{
			Strength:      0,
			Dexterity:     0,
			Intelligence:  0,
			Health:        0,
			Will:          0,
			Perception:    0,
			BasicSpeed:    0,
			FatiguePoints: 0,
			HitPoints:     0,
			BasicLift:     0,
		},
		skillLevels: map[SkillName]int{
			SkillNameMeleeWeapons: 0,
			SkillNameShield:       0,
		},
		resources: map[Stat]int{
			HitPoints:     10,
			FatiguePoints: 3,
		},
		statModifiers:  make(map[Stat][]Modifier),
		skillModifiers: make(map[SkillName][]Modifier),
	}
}
func (c *Character) SetStatChangedHandler(f func()) {
	c.statChangedHandler = f
}
func (c *Character) SetResourceChangedHandler(f func()) {
	c.resourceChangedHandler = f

}
func (c *Character) AddStatModifier(stat Stat, modifier Modifier) {
	c.statModifiers[stat] = append(c.statModifiers[stat], modifier)
}
func (c *Character) AddSkillModifier(skill SkillName, modifier Modifier) {
	c.skillModifiers[skill] = append(c.skillModifiers[skill], modifier)
}
func (c *Character) GetStat(stat Stat) int {
	var statValue int
	if stat.IsDerived() {
		statValue = stat.GetDerivedValueFromLevelAdjustments(c.levelAdjustments[stat], c.GetStat, c.GetSkill)
	} else {
		statValue = stat.GetNonDerivedValueFromLevelAdjustment(c.levelAdjustments[stat])
	}
	return c.ApplyStatModifiers(stat, statValue)
}
func (c *Character) GetSkill(skill SkillName) int {
	level, hasSkill := c.skillLevels[skill]
	if !hasSkill {
		return skill.GetDefaultValue(c.GetStat)
	}
	skillValue := skill.GetValueFromLevel(level, c.GetStat)
	return c.applySkillModifiers(skill, skillValue)
}
func (c *Character) GetTotalCost() int {
	return GetTotalCostOfAdjustments(c.levelAdjustments) + GetTotalCostOfSkills(c.skillLevels)
}

func (c *Character) GetCharacterPointsBalance() int {
	return c.characterPointsReceived - c.GetTotalCost()
}

func (c *Character) IsBalanceValid() bool {
	return c.GetCharacterPointsBalance() >= 0
}

func (c *Character) SetCharacterPointsReceived(points int) {
	c.characterPointsReceived = points
}

func (c *Character) AddCharacterPoints(points int) {
	c.characterPointsReceived += points
}

func GetTotalCostOfSkills(levels map[SkillName]int) int {
	totalCost := 0
	for skill, level := range levels {
		totalCost += skill.PointsSpentFromLevel(level)
	}
	return totalCost
}
func (c *Character) GetPointsSpentOnSkill(skill SkillName) int {
	level := c.skillLevels[skill]
	return skill.PointsSpentFromLevel(level)
}
func (c *Character) GetPointsSpentOnStat(stat Stat) int {
	level := c.levelAdjustments[stat]
	return stat.CostPerLevel() * level
}
func (c *Character) Increment(stat Stat) {
	c.levelAdjustments[stat]++
	c.onStatChanged(stat)
}

func (c *Character) Decrement(stat Stat) {
	c.levelAdjustments[stat]--
	c.onStatChanged(stat)
}

func (c *Character) SetAdjustment(stat Stat, value int) {
	c.levelAdjustments[stat] = value
	c.onStatChanged(stat)
}

func (c *Character) IncreaseSkillLevel(skill SkillName) {
	c.skillLevels[skill]++
}

func (c *Character) SetSkillLevel(skill SkillName, level int) {
	c.skillLevels[skill] = level
}

func (c *Character) GetFlatAttributes() map[Stat]int {
	flatAttributes := make(map[Stat]int)
	for stat, _ := range c.levelAdjustments {
		flatAttributes[stat] = c.GetStat(stat)
	}
	return flatAttributes
}

func (c *Character) GetFlatSkills() map[SkillName]int {
	skills := make(map[SkillName]int)
	for skill, _ := range c.skillLevels {
		skills[skill] = c.GetSkill(skill)
	}
	return skills
}

func (c *Character) GetOverview() []string {
	attribs := c.GetFlatAttributes()
	skills := c.GetFlatSkills()
	overview := make([]string, 0)
	overview = append(overview, fmt.Sprintf("ST: %d [%d]", attribs[Strength], c.pointsSpent(Strength)))
	overview = append(overview, fmt.Sprintf("DX: %d [%d]", attribs[Dexterity], c.pointsSpent(Dexterity)))
	overview = append(overview, fmt.Sprintf("IQ: %d [%d]", attribs[Intelligence], c.pointsSpent(Intelligence)))
	overview = append(overview, fmt.Sprintf("HT: %d [%d]", attribs[Health], c.pointsSpent(Health)))
	overview = append(overview, fmt.Sprintf("Will: %d [%d]", attribs[Will], c.pointsSpent(Will)))
	overview = append(overview, fmt.Sprintf("Per: %d [%d]", attribs[Perception], c.pointsSpent(Perception)))
	overview = append(overview, fmt.Sprintf("Basic Speed: %d [%d]", attribs[BasicSpeed], c.pointsSpent(BasicSpeed)))
	overview = append(overview, fmt.Sprintf("Basic Move: %d [%d]", attribs[BasicSpeed], c.pointsSpent(BasicSpeed)))
	overview = append(overview, fmt.Sprintf("Max Fatigue: %d [%d]", attribs[FatiguePoints], c.pointsSpent(FatiguePoints)))
	overview = append(overview, fmt.Sprintf("Max HP: %d [%d]", attribs[HitPoints], c.pointsSpent(HitPoints)))
	overview = append(overview, "")
	overview = append(overview, fmt.Sprintf("Brawling: %d [%d]", skills[SkillNameBrawling], c.pointsSpentSkill(SkillNameShield)))
	overview = append(overview, fmt.Sprintf("Melee: %d [%d]", skills[SkillNameMeleeWeapons], c.pointsSpentSkill(SkillNameMeleeWeapons)))
	overview = append(overview, fmt.Sprintf("Shield: %d [%d]", skills[SkillNameShield], c.pointsSpentSkill(SkillNameShield)))
	overview = append(overview, fmt.Sprintf("Throwing: %d [%d]", skills[SkillNameThrowing], c.pointsSpentSkill(SkillNameShield)))
	overview = append(overview, fmt.Sprintf("Missile: %d [%d]", skills[SkillNameMissileWeapons], c.pointsSpentSkill(SkillNameShield)))
	overview = append(overview, "")
	overview = append(overview, fmt.Sprintf("Total Points: %d", c.GetTotalCost()))

	return overview
}

func (c *Character) GetApplicableModifiersDescription() []string {
	descriptions := make([]string, 0)
	for stat, modifiers := range c.statModifiers {
		for _, modifier := range modifiers {
			if modifier.IsApplicable() {
				descriptions = append(descriptions, fmt.Sprintf("%s: %s", stat.ToString(), modifier.Description()))
			}
		}
	}
	for skillName, modifiers := range c.skillModifiers {
		for _, modifier := range modifiers {
			if modifier.IsApplicable() {
				descriptions = append(descriptions, fmt.Sprintf("%s: %s", skillName, modifier.Description()))
			}
		}
	}
	return descriptions
}

func (c *Character) ResetResources() {
	c.resources[HitPoints] = c.GetStat(HitPoints)
	c.resources[FatiguePoints] = c.GetStat(FatiguePoints)
}

func (c *Character) GetResource(stat Stat) int {
	return c.resources[stat]
}

func (c *Character) GetResourceMax(stat Stat) int {
	return c.GetStat(stat)
}

func (c *Character) IncreaseResourceBy(stat Stat, amount int) {
	c.resources[stat] += amount
	c.onResourceChanged(stat)
}

func (c *Character) DecreaseResourceBy(stat Stat, amount int) {
	c.resources[stat] -= amount
	c.onResourceChanged(stat)
}

func (c *Character) onResourceChanged(stat Stat) {
	if c.resources[stat] > c.GetStat(stat) {
		c.resources[stat] = c.GetStat(stat)
	}
	if c.resources[stat] < 0 {
		c.resources[stat] = 0
	}
	if c.resourceChangedHandler != nil {
		c.resourceChangedHandler()
	}
}

func (c *Character) onStatChanged(stat Stat) {
	if c.statChangedHandler != nil {
		defer c.statChangedHandler()
	}
	if stat == HitPoints {
		newMaxHitPoints := c.GetStat(HitPoints)
		if c.resources[HitPoints] > newMaxHitPoints {
			c.resources[HitPoints] = newMaxHitPoints
		}
		return
	}

	if stat == FatiguePoints {
		newMaxFatiguePoints := c.GetStat(FatiguePoints)
		if c.resources[FatiguePoints] > newMaxFatiguePoints {
			c.resources[FatiguePoints] = newMaxFatiguePoints
		}
		return
	}
}

func (c *Character) pointsSpent(state Stat) int {
	return c.levelAdjustments[state] * state.CostPerLevel()
}

func (c *Character) pointsSpentSkill(name SkillName) int {
	return name.PointsSpentFromLevel(c.skillLevels[name])
}

func (c *Character) SetStat(stat Stat, value int) {
	currentValue := c.GetStat(stat)
	if value == currentValue {
		return
	}
	if value < currentValue {
		adjPerLevel := stat.AdjustmentPerLevel()
		adjustmentsNeeded := (currentValue - value) / adjPerLevel
		c.DecreaseStatBy(stat, adjustmentsNeeded)
		return
	}
	if value > currentValue {
		adjPerLevel := stat.AdjustmentPerLevel()
		adjustmentsNeeded := (value - currentValue) / adjPerLevel
		c.IncreaseStatBy(stat, adjustmentsNeeded)
	}
}

func (c *Character) DecreaseStatBy(stat Stat, levels int) {
	c.levelAdjustments[stat] -= levels
}

func (c *Character) IncreaseStatBy(stat Stat, levels int) {
	c.levelAdjustments[stat] += levels
}

func (c *Character) GetCharacterPointsReceived() int {
	return c.characterPointsReceived
}

func (c *Character) ApplyStatModifiers(stat Stat, value int) int {
	applicableModifiers := make([]Modifier, 0)
	if c.statModifiers == nil {
		return value
	}
	modifiers, ok := c.statModifiers[stat]
	if !ok {
		return value
	}

	applicableModifiers = FilterModifiers(modifiers)

	for _, modifier := range applicableModifiers {
		value = modifier.Apply(value)
	}

	return value
}
func (c *Character) applySkillModifiers(skill SkillName, value int) int {
	applicableModifiers := make([]Modifier, 0)
	if c.skillModifiers == nil {
		return value
	}
	modifiers, ok := c.skillModifiers[skill]
	if !ok {
		return value
	}

	applicableModifiers = FilterModifiers(modifiers)

	for _, modifier := range applicableModifiers {
		value = modifier.Apply(value)
	}

	return value
}
func FilterModifiers(modifiers []Modifier) []Modifier {
	if len(modifiers) == 0 {
		return modifiers
	}
	applicableModifiers := make([]Modifier, 0)
	for _, modifier := range modifiers {
		if modifier.IsApplicable() {
			applicableModifiers = append(applicableModifiers, modifier)
		}
	}
	slices.SortStableFunc(applicableModifiers, func(i, j Modifier) int {
		return cmp.Compare(i.SortOrder(), j.SortOrder())
	})
	return applicableModifiers
}

func (c *Character) GetActiveDefenseScore(defense ActiveDefenseType, parryDefenseScore int) int {
	baseDefenseScore := 0
	switch defense {
	case ActiveDefenseDodge:
		baseDefenseScore = c.GetStat(Dodge) // depends on stat
	case ActiveDefenseBlock:
		baseDefenseScore = c.GetStat(Block) // depends on skill
	case ActiveDefenseParry:
		baseDefenseScore = parryDefenseScore // depends on variable skill
	}

	return c.ApplyStatModifiers(ActiveDefense, baseDefenseScore)
}

func (c *Character) HasCharPointsLeft() bool {
	return c.GetCharacterPointsBalance() > 0
}

func (c *Character) GetLevelAdjustments(stat Stat) int {
	return c.levelAdjustments[stat]
}

type ActiveDefenseType int

const (
	ActiveDefenseDodge ActiveDefenseType = iota
	ActiveDefenseBlock
	ActiveDefenseParry
)
