package game

import (
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/special"
	"RogueUI/util"
	"fmt"
	"image/color"
	"math/rand"
	"strings"
)

// Player Starting Stats in Rogue
// Strength: 16
// Experience: 0
// Level: 1
// Armor Class: 10
// Hit Points: 12/12
// Damage: 1d4

type Actor struct {
	internalName string
	name         string
	charSheet    *special.CharSheet
	position     geometry.Point

	inventory *Inventory
	equipment *Equipment

	statusFlags *foundation.MapFlags

	intrinsicAttacks    []IntrinsicAttack
	intrinsicZapEffects []string
	intrinsicUseEffects []string

	icon                   rune
	color                  string
	currentIntrinsicAttack int
	sizeModifier           int
	timeEnergy             int
	body                   []*foundation.BodyPart
}

func NewPlayer(name string, playerIcon rune, playerColor string, character *special.CharSheet) *Actor {
	player := NewActor(name, playerIcon, playerColor, character)
	return player
}

func NewActor(name string, icon rune, color string, character *special.CharSheet) *Actor {

	body := foundation.BodyByName("human", character.GetHitPointsMax())

	a := &Actor{
		name:        name,
		icon:        icon,
		color:       color,
		inventory:   NewInventory(23),
		equipment:   NewEquipment(),
		charSheet:   character,
		body:        body,
		statusFlags: foundation.NewMapFlags(),
	}

	// add persistent modifiers, that always apply when a condition is met here

	// injured -> 1/2 dodge

	return a
}

func (a *Actor) GetBodyPartsAndHitChances(attackerSkill int, defenderSkill int) []util.Tuple[string, int] {
	var result []util.Tuple[string, int]
	defenseChance := dice_curve.ChanceOfSuccess(defenderSkill)
	for _, part := range a.body {
		effectiveSkill := attackerSkill + part.SizeModifier
		hitChance := dice_curve.ChanceOfSuccess(effectiveSkill)
		combinedChanceToHit := hitChance * (1 - defenseChance)
		chanceAsInt := int(combinedChanceToHit * 100)
		result = append(result, util.Tuple[string, int]{Item1: part.Name, Item2: chanceAsInt})
	}
	return result
}

func (a *Actor) GetBodyPartByIndex(part int) string {
	if part < 0 || part >= len(a.body) {
		return "None"
	}
	return a.body[part].Name
}

func (a *Actor) Color() string {
	return a.color
}

type CapModifier struct {
	maxValue    int
	doesApply   func() bool
	persistent  bool
	description string
}

func (c CapModifier) Description() string {
	return c.description
}

func (c CapModifier) Apply(i int) int {
	if i > c.maxValue {
		return c.maxValue
	}
	return i
}

func (c CapModifier) IsApplicable() bool {
	return c.doesApply()
}

func (c CapModifier) IsPersistent() bool {
	return c.persistent
}

func (c CapModifier) SortOrder() int {
	return 3
}

func ModCapWhen(maxValue int, description string, conditionForApplication func() bool) dice_curve.Modifier {
	return CapModifier{
		maxValue:    maxValue,
		doesApply:   conditionForApplication,
		description: description,
		persistent:  true,
	}
}
func ModCap(maxValue int, reason string) dice_curve.Modifier {
	return CapModifier{
		maxValue:    maxValue,
		doesApply:   func() bool { return true },
		description: fmt.Sprintf("Cap %d - %s", maxValue, reason),
		persistent:  true,
	}
}

type FlatModifier struct {
	flatMod     int
	doesApply   func() bool
	persistent  bool
	description string
}

func (f FlatModifier) Description() string {
	return f.description
}

func (f FlatModifier) Apply(i int) int {
	return i + f.flatMod
}

func (f FlatModifier) IsApplicable() bool {
	return f.doesApply()
}

func (f FlatModifier) IsPersistent() bool {
	return f.persistent
}

func (f FlatModifier) SortOrder() int {
	return 1
}

func ModFlatWhen(flatMod int, reason string, conditionForApplication func() bool) dice_curve.Modifier {
	return FlatModifier{
		flatMod:     flatMod,
		doesApply:   conditionForApplication,
		description: fmt.Sprintf("%d - %s", flatMod, reason),
		persistent:  true,
	}
}
func ModFlat(flatMod int, reason string) dice_curve.Modifier {
	return FlatModifier{
		flatMod:     flatMod,
		doesApply:   func() bool { return true },
		description: fmt.Sprintf("%d - %s", flatMod, reason),
		persistent:  false,
	}
}

type PercentageModifier struct {
	factor      float64
	doesApply   func() bool
	persistent  bool
	description string
}

func (p PercentageModifier) Description() string {
	return p.description
}

func (p PercentageModifier) Apply(i int) int {
	return int(float64(i) * p.factor)
}

func (p PercentageModifier) IsApplicable() bool {
	return p.doesApply()
}

func (p PercentageModifier) IsPersistent() bool {
	return p.persistent
}

func (p PercentageModifier) SortOrder() int {
	return 0
}

func ModHalveWhen(reason string, isInjured func() bool) PercentageModifier {
	return PercentageModifier{
		factor:      0.5,
		doesApply:   isInjured,
		persistent:  true,
		description: fmt.Sprintf("1/2 - %s", reason),
	}
}

func (a *Actor) Icon() rune {
	return a.icon
}
func (a *Actor) GetListInfo() string {
	hp := a.charSheet.GetHitPoints()
	hpMax := a.charSheet.GetHitPointsMax()
	return fmt.Sprintf("%s HP: %d/%d Dmg: %s DR: %d", a.name, hp, hpMax, a.GetIntrinsicDamageAsString(), a.GetDamageResistance())
}

func (a *Actor) IsStunned() bool {
	return a.HasFlag(foundation.FlagStun)
}

func (a *Actor) Position() geometry.Point {
	return a.position
}
func (a *Actor) SetPosition(pos geometry.Point) {
	a.position = pos
}

func (a *Actor) Name() string {
	if a.HasFlag(foundation.FlagInvisible) {
		return "something"
	}
	return a.name
}

func (a *Actor) IsDrawn(playerCanSeeInvisible bool) bool {
	return !a.HasFlag(foundation.FlagInvisible) || playerCanSeeInvisible
}

func (a *Actor) GetInventory() *Inventory {
	return a.inventory
}

func (a *Actor) GetEquipment() *Equipment {
	return a.equipment
}

func (a *Actor) GetDamageResistance() int {
	return 0
}

func (a *Actor) GetFlags() *foundation.MapFlags {
	return a.statusFlags
}

func (a *Actor) IsAlive() bool {
	return a.charSheet.IsAlive()
}

func (a *Actor) HasFlag(flag foundation.ActorFlag) bool {
	return a.statusFlags.IsSet(flag) || a.GetEquipment().ContainsFlag(flag)
}

func (a *Actor) TakeDamage(amount int) {
	a.charSheet.TakeRawDamage(amount)
}

func (a *Actor) IsSleeping() bool {
	return a.HasFlag(foundation.FlagSleep)
}

func (a *Actor) WakeUp() {
	a.statusFlags.Unset(foundation.FlagSleep)
}

func (a *Actor) SetIntrinsicZapEffects(effects []string) {
	a.intrinsicZapEffects = effects
}

func (a *Actor) SetIntrinsicUseEffects(effects []string) {
	a.intrinsicUseEffects = effects
}

func (a *Actor) GetIntrinsicZapEffects() []string {
	return a.intrinsicZapEffects
}

func (a *Actor) GetIntrinsicUseEffects() []string {
	return a.intrinsicUseEffects
}
func (a *Actor) AddGold(i int) {
	a.statusFlags.Increase(foundation.FlagGold, i)
}

func (a *Actor) GetIntrinsicDamageAsString() string {
	var parts []string
	for _, dice := range a.intrinsicAttacks {
		parts = append(parts, dice.DamageDice.ShortString())
	}
	return strings.Join(parts, "/")
}

func (a *Actor) RemoveLevelStatusEffects() {
	a.statusFlags.Unset(foundation.FlagSeeFood)
	a.statusFlags.Unset(foundation.FlagSeeMagic)
	a.statusFlags.Unset(foundation.FlagSeeTraps)
}

func (a *Actor) Heal(amount int) {
	a.charSheet.Heal(amount)
}

func (a *Actor) GetInternalName() string {
	return a.internalName
}

func (a *Actor) SetInternalName(name string) {
	a.internalName = name
}
func (a *Actor) ChooseIntrinsicAttack() (int, dice_curve.Dice) {
	if a.intrinsicAttacks == nil || len(a.intrinsicAttacks) == 0 {
		return 0, dice_curve.Dice{}
	}
	randomIndexOfIntrinsicDamage := rand.Intn(len(a.intrinsicAttacks))
	a.currentIntrinsicAttack = randomIndexOfIntrinsicDamage
	return a.intrinsicAttacks[randomIndexOfIntrinsicDamage].BaseSkill, a.intrinsicAttacks[randomIndexOfIntrinsicDamage].DamageDice
}
func (a *Actor) GetHitPoints() int {
	return a.charSheet.GetHitPoints()
}

func (a *Actor) GetHitPointsMax() int {
	return a.charSheet.GetHitPointsMax()
}

func (a *Actor) IsInjured() bool {
	hpMax := a.GetHitPointsMax()
	hpCurrent := a.GetHitPoints()
	belowOneThirdHitPoints := hpCurrent < hpMax/3
	return belowOneThirdHitPoints
}
func (a *Actor) IsFatigued() bool {
	fpMax := a.charSheet.GetActionPointsMax()
	fpCurrent := a.charSheet.GetActionPoints()
	belowOneThird := fpCurrent < fpMax/3
	return belowOneThird
}
func (a *Actor) GetColor() string {
	return a.color
}

func (a *Actor) TextIcon(bg color.RGBA, getColor func(string) color.RGBA) foundation.TextIcon {
	return foundation.TextIcon{
		Rune: a.icon,
		Fg:   getColor(a.color),
		Bg:   bg,
	}
}

func (a *Actor) GetDetailInfo() []string {

	var result []string
	result = append(result, fmt.Sprintf("Name: %s", a.Name()))

	// melee attack
	statRows := []util.TableRow{
		util.TableRow{Columns: []string{"Str:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Strength))}},
		util.TableRow{Columns: []string{"Per:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Perception))}},
		util.TableRow{Columns: []string{"End:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Endurance))}},
		util.TableRow{Columns: []string{"Cha:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Charisma))}},
		util.TableRow{Columns: []string{"Int:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Intelligence))}},
		util.TableRow{Columns: []string{"Agi:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Agility))}},
		util.TableRow{Columns: []string{"Lck:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Luck))}},
	}

	derivedStatRows := []util.TableRow{
		{Columns: []string{"HP:", fmt.Sprintf("%d/%d", a.GetHitPoints(), a.GetHitPointsMax())}},
		{Columns: []string{"AP:", fmt.Sprintf("%d/%d", a.charSheet.GetActionPoints(), a.charSheet.GetActionPointsMax())}},
		{Columns: []string{"Speed:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.Speed))}},
		{Columns: []string{"Dodge:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.Dodge))}},
		{Columns: []string{"Crit. Chance:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.CriticalChance))}},
		{Columns: []string{"Carry Weight:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.CarryWeight))}},
	}

	resistanceRows := []util.TableRow{
		{Columns: []string{"Physical:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.DamageResistance))}},
		{Columns: []string{"Energy:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.EnergyResistance))}},
		{Columns: []string{"Poison :", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.PoisonResistance))}},
		{Columns: []string{"Radiation:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.RadiationResistance))}},
	}

	skillRows := []util.TableRow{
		{Columns: []string{"Melee Weapons:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.MeleeWeapons))}},
		{Columns: []string{"Unarmed:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Unarmed))}},
		{Columns: []string{"Throwing:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Throwing))}},
		{Columns: []string{"Small Guns:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.SmallGuns))}},
		{Columns: []string{"Big Guns:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.BigGuns))}},
		{Columns: []string{"Energy Weapons:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.EnergyWeapons))}},
		{Columns: []string{"Doctor:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Doctor))}},
		{Columns: []string{"Sneak:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Sneak))}},
		{Columns: []string{"Lockpick:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Lockpick))}},
		{Columns: []string{"Science:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Science))}},
		{Columns: []string{"Repair:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Repair))}},
		{Columns: []string{"Speech:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Speech))}},
		{Columns: []string{"Barter:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Barter))}},
		{Columns: []string{"Outdoorsman:", fmt.Sprintf("%d", a.charSheet.GetSkill(special.Outdoorsman))}},
	}

	statLines := util.TableLayout(statRows, []util.TextAlignment{util.AlignLeft, util.AlignLeft})
	skillLines := util.TableLayout(skillRows, []util.TextAlignment{util.AlignLeft, util.AlignLeft})
	derivedLines := util.TableLayout(derivedStatRows, []util.TextAlignment{util.AlignLeft, util.AlignLeft})
	resistanceLines := util.TableLayout(resistanceRows, []util.TextAlignment{util.AlignLeft, util.AlignLeft})
	result = append(result, "", "> Stats:")
	result = append(result, statLines...)
	result = append(result, "", "> Derived Stats:")
	result = append(result, derivedLines...)
	result = append(result, "", "> Resistances:")
	result = append(result, resistanceLines...)
	result = append(result, "", "> Skills:")
	result = append(result, skillLines...)
	return result
}

func (a *Actor) SetIntrinsicAttacks(attacks []IntrinsicAttack) {
	a.intrinsicAttacks = attacks
}

func (a *Actor) GetStatModifier(stat dice_curve.Stat) int {
	equipMod := a.GetEquipment().GetStatModifier(stat)
	//rulesMod := dice_curve.GetStatModifier(stat, a.charSheet, a.GetEncumbrance())
	return equipMod
}

func (a *Actor) GetEncumbrance() int {
	if !a.GetEquipment().HasArmorEquipped() {
		return 0
	}
	armorEncumbrance := a.GetEquipment().GetEncumbranceFromArmor()
	return armorEncumbrance
}

func (a *Actor) GetSizeModifier() int {
	return a.sizeModifier
}

func (a *Actor) SetSizeModifier(modifier int) {
	a.sizeModifier = modifier
}
func (a *Actor) HasGold(price int) bool {
	return a.statusFlags.Get(foundation.FlagGold) >= price
}

func (a *Actor) RemoveGold(price int) {
	a.statusFlags.Decrease(foundation.FlagGold, price)
}

func (a *Actor) GetGold() int {
	return a.statusFlags.Get(foundation.FlagGold)
}

func (a *Actor) NeedsHealing() bool {
	return a.GetHitPoints() < a.GetHitPointsMax()
}

func (a *Actor) IsHungry() bool {
	return a.statusFlags.Get(foundation.FlagHunger) > 0
}

func (a *Actor) IsStarving() bool {
	return a.statusFlags.Get(foundation.FlagHunger) > 1
}

func (a *Actor) Satiate() {
	a.statusFlags.Unset(foundation.FlagHunger)
	a.statusFlags.Unset(foundation.FlagTurnsSinceEating)
}

func (a *Actor) SetSleeping() {
	flags := a.GetFlags()
	flags.Set(foundation.FlagSleep)
	flags.Unset(foundation.FlagAwareOfPlayer)
	flags.Unset(foundation.FlagScared)
}

func (a *Actor) SetUnwary() {
	flags := a.GetFlags()
	flags.Unset(foundation.FlagSleep)
	flags.Unset(foundation.FlagAwareOfPlayer)
}

func (a *Actor) SetAware() {
	flags := a.GetFlags()
	flags.Unset(foundation.FlagSleep)
	flags.Set(foundation.FlagAwareOfPlayer)
}

func (a *Actor) IsBlind() bool {
	return a.HasFlag(foundation.FlagBlind)
}

func (a *Actor) AddTimeEnergy(timeSpent int) {
	a.timeEnergy += timeSpent
}

func (a *Actor) HasEnergyForActions() bool {
	return a.timeEnergy >= a.timeNeededForActions()
}

func (a *Actor) timeNeededForActions() int {
	speed := a.GetBasicSpeed()
	timeNeeded := 100 / speed
	return timeNeeded
}

func (a *Actor) SpendTimeEnergy() {
	a.timeEnergy -= a.timeNeededForActions()
}

func (a *Actor) AfterTurn() {
	a.GetEquipment().AfterTurn()
}

func (a *Actor) decrementStatusEffectCounters() {
	flags := a.GetFlags()
	flags.Decrement(foundation.FlagHaste)
	flags.Decrement(foundation.FlagSlow)
	flags.Decrement(foundation.FlagConfused)
	flags.Decrement(foundation.FlagFly)
	flags.Decrement(foundation.FlagSeeInvisible)
	flags.Decrement(foundation.FlagHallucinating)
}

func (a *Actor) GetBasicSpeed() int {
	return 10 // TODO
}

func (a *Actor) GetCharSheet() *special.CharSheet {
	return a.charSheet
}

func (a *Actor) IsHuman() bool {
	return a.internalName == "human" || a.internalName == "player" || a.internalName == ""
}
