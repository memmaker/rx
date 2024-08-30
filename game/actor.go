package game

import (
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"RogueUI/special"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"strings"
)

type CharacterMood uint8

const (
	Neutral CharacterMood = iota
	Hostile
	Panic
)

func PlayerRelationFromString(str string) CharacterMood {
	str = strings.ToLower(str)
	switch str {
	case "neutral":
		return Neutral
	case "hostile":
		return Hostile
	case "ally":
		return Panic
	}
	return Neutral
}

type Actor struct {
	internalName string
	name         string
	charSheet    *special.CharSheet
	position     geometry.Point

	inventory *Inventory
	equipment *Equipment

	statusFlags *foundation.ActorFlags

	intrinsicZapEffects []string
	intrinsicUseEffects []string

	icon         textiles.TextIcon
	sizeModifier int
	timeEnergy   int
	body         []*foundation.BodyPart

	mood         CharacterMood
	dialogueFile string
	teamName     string

	enemyActors map[string]bool
	enemyTeams  map[string]bool

	xp int
}

// GobEncode encodes the Actor struct into a byte slice.
func (a *Actor) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Encode each field of the struct
	err := encoder.Encode(a.internalName)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.name)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.charSheet)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.position)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.inventory)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.equipment)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.statusFlags)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.intrinsicZapEffects)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.intrinsicUseEffects)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.icon)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.sizeModifier)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.timeEnergy)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.body)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.mood)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.dialogueFile)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.teamName)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.enemyActors)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.enemyTeams)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.xp)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GobDecode decodes a byte slice into an Actor struct.
func (a *Actor) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)

	// Decode each field of the struct
	err := decoder.Decode(&a.internalName)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.name)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.charSheet)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.position)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.inventory)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.equipment)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.statusFlags)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.intrinsicZapEffects)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.intrinsicUseEffects)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.icon)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.sizeModifier)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.timeEnergy)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.body)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.mood)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.dialogueFile)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.teamName)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.enemyActors)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.enemyTeams)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.xp)
	if err != nil {
		return err
	}

	return nil
}

func NewPlayer(name string, icon textiles.TextIcon, character *special.CharSheet) *Actor {
	player := NewActor()
	player.SetCharSheet(character)
	player.SetDisplayName(name)
	player.SetIcon(icon)
	player.SetInternalName("player")
	return player
}

func NewActor() *Actor {
	sheet := special.NewCharSheet()
	body := foundation.BodyByName("human", sheet.GetHitPointsMax())

	a := &Actor{
		name: "Unknown",
		icon: textiles.TextIcon{
			Char: '0',
			Fg:   color.RGBA{255, 255, 255, 255},
			Bg:   color.RGBA{0, 0, 0, 255},
		},
		inventory:   NewInventory(23),
		equipment:   NewEquipment(),
		charSheet:   sheet,
		body:        body,
		mood:        Neutral,
		statusFlags: foundation.NewActorFlags(),
		enemyActors: make(map[string]bool),
		enemyTeams:  make(map[string]bool),
	}
	return a
}

func (a *Actor) SetDialogueFile(scriptName string) {
	a.dialogueFile = scriptName
}

func (a *Actor) GetBodyPartsAndHitChances(attackerSkill int, defenderSkill int) []fxtools.Tuple[string, int] {
	var result []fxtools.Tuple[string, int]
	defenseChance := dice_curve.ChanceOfSuccess(defenderSkill)
	for _, part := range a.body {
		effectiveSkill := attackerSkill + part.SizeModifier
		hitChance := dice_curve.ChanceOfSuccess(effectiveSkill)
		combinedChanceToHit := hitChance * (1 - defenseChance)
		chanceAsInt := int(combinedChanceToHit * 100)
		result = append(result, fxtools.Tuple[string, int]{Item1: part.Name, Item2: chanceAsInt})
	}
	return result
}

func (a *Actor) GetBodyPartByIndex(part int) string {
	if part < 0 || part >= len(a.body) {
		return "None"
	}
	return a.body[part].Name
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

func (a *Actor) Icon() textiles.TextIcon {
	if a.IsSleeping() || a.IsKnockedDown() {
		originalRune := a.icon.Char
		asLower := strings.ToLower(string(originalRune))
		return a.icon.WithRune([]rune(asLower)[0])
	}
	return a.icon
}
func (a *Actor) GetListInfo() string {
	hp := a.charSheet.GetHitPoints()
	hpMax := a.charSheet.GetHitPointsMax()
	item, hasMainHandItem := a.GetEquipment().GetMainHandItem()
	damage := "0"
	if hasMainHandItem && item.IsWeapon() {
		damage = item.GetWeapon().GetDamage().ShortString()
	}
	return fmt.Sprintf("%s HP: %d/%d Dmg: %s DR: %d", a.name, hp, hpMax, damage, a.GetDamageResistance())
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

func (a *Actor) IsVisible(playerCanSeeInvisible bool) bool {
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

func (a *Actor) GetFlags() *foundation.ActorFlags {
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

func (a *Actor) TextIcon(bg color.RGBA) textiles.TextIcon {
	return a.Icon().WithBg(bg)
}

func (a *Actor) GetDetailInfo() string {

	var result []string
	result = append(result, fmt.Sprintf("Name: %s", a.Name()))

	// melee attack
	statRows := []fxtools.TableRow{
		fxtools.TableRow{Columns: []string{"Str:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Strength))}},
		fxtools.TableRow{Columns: []string{"Per:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Perception))}},
		fxtools.TableRow{Columns: []string{"End:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Endurance))}},
		fxtools.TableRow{Columns: []string{"Cha:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Charisma))}},
		fxtools.TableRow{Columns: []string{"Int:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Intelligence))}},
		fxtools.TableRow{Columns: []string{"Agi:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Agility))}},
		fxtools.TableRow{Columns: []string{"Lck:", fmt.Sprintf("%d", a.charSheet.GetStat(special.Luck))}},
	}

	derivedStatRows := []fxtools.TableRow{
		{Columns: []string{"HP:", fmt.Sprintf("%d/%d", a.GetHitPoints(), a.GetHitPointsMax())}},
		{Columns: []string{"AP:", fmt.Sprintf("%d/%d", a.charSheet.GetActionPoints(), a.charSheet.GetActionPointsMax())}},
		{Columns: []string{"Speed:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.Speed))}},
		{Columns: []string{"Dodge:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.Dodge))}},
		{Columns: []string{"Crit. Chance:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.CriticalChance))}},
		{Columns: []string{"Carry Weight:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.CarryWeight))}},
	}

	resistanceRows := []fxtools.TableRow{
		{Columns: []string{"Physical:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.DamageResistance))}},
		{Columns: []string{"Energy:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.EnergyResistance))}},
		{Columns: []string{"Poison :", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.PoisonResistance))}},
		{Columns: []string{"Radiation:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(special.RadiationResistance))}},
	}

	skillRows := []fxtools.TableRow{
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

	statLines := fxtools.TableLayout(statRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	skillLines := fxtools.TableLayout(skillRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	derivedLines := fxtools.TableLayout(derivedStatRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	resistanceLines := fxtools.TableLayout(resistanceRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	result = append(result, "", "> Stats:")
	result = append(result, statLines...)
	result = append(result, "", "> Derived Stats:")
	result = append(result, derivedLines...)
	result = append(result, "", "> Resistances:")
	result = append(result, resistanceLines...)
	result = append(result, "", "> Skills:")
	result = append(result, skillLines...)
	return strings.Join(result, "\n")
}

func (a *Actor) GetStatModifier(stat special.Stat) int {
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

func (a *Actor) Kill() {
	a.charSheet.Kill()
}

func (a *Actor) HasKey(identifier string) bool {
	return a.GetInventory().HasKey(identifier)
}

func (a *Actor) IsHostile() bool {
	return a.mood == Hostile
}

func (a *Actor) SetRelationToPlayer(relation CharacterMood) {
	a.mood = relation
}

func (a *Actor) SetDisplayName(name string) {
	a.name = name
}

func (a *Actor) GetDialogueFile() string {
	return a.dialogueFile
}

func (a *Actor) SetHostile() {
	a.mood = Hostile
	a.tryEquipWeapon()
}

func (a *Actor) tryEquipWeapon() {
	if !a.GetEquipment().HasWeaponEquipped() {
		weapon := a.GetInventory().GetBestWeapon()
		if weapon != nil {
			a.GetEquipment().Equip(weapon)
		}
	}
}

func (a *Actor) GetHitAudioCue(isMelee bool) string {
	audioName := a.getAudioName()
	hitType := "HIT"
	if isMelee {
		hitType = "MELEE_HIT"
	}
	return fmt.Sprintf("critters/%s/%s", audioName, hitType)
}

func (a *Actor) GetDeathAudioCue() string {
	audioName := a.getAudioName()
	return fmt.Sprintf("critters/%s/FALLING", audioName)
}
func (a *Actor) GetDeathCriticalAudioCue(mode special.TargetingMode, damageType special.DamageType) string {
	audioName := a.getAudioName()
	actionName := "FALLING"
	switch damageType {
	case special.DamageTypeNormal:
		switch mode {
		case special.TargetingModeFireBurst:
			actionName = "PERFORATED_DEATH"
		default:
			if rand.Intn(2) == 0 {
				actionName = "HOLE_IN_BODY"
			} else {
				actionName = "RIPPING_APART"
			}
		}
	case special.DamageTypeLaser:
		actionName = "SLICE_IN_TWO"
	case special.DamageTypeFire:
		if rand.Intn(2) == 0 { // TODO: not both always available, fallbacks or tests needed..
			actionName = "BURNED"
		} else {
			actionName = "BURNING_DANCE"
		}
	case special.DamageTypeExplosive:
		actionName = "BLOW_EXPLOSION"
	case special.DamageTypeElectrical:
		if rand.Intn(2) == 0 {
			actionName = "ELECTRIC_BURNED"
		} else {
			actionName = "ELECTRIC_BURNED_TO_ASHES"
		}
	case special.DamageTypePlasma:
		actionName = "MELTDOWN"
	default:
		actionName = "FALLING"
	}
	return fmt.Sprintf("critters/%s/%s", audioName, actionName)
}
func (a *Actor) GetDodgedAudioCue() string {
	audioName := a.getAudioName()
	return fmt.Sprintf("critters/%s/DODGE", audioName)
}

func (a *Actor) GetMeleeAudioCue(isKick bool) string {
	audioName := a.getAudioName()
	hitType := "PUNCH"
	if isKick {
		hitType = "KICK"
	}
	return fmt.Sprintf("critters/%s/%s", audioName, hitType)
}
func (a *Actor) getAudioName() string {
	return "human_male"
}

func (a *Actor) GetTeam() string {
	return a.teamName
}

func (a *Actor) AddToEnemyActors(name string) {
	if a.internalName == name {
		return
	}
	a.enemyActors[name] = true
}

func (a *Actor) AddToEnemyTeams(name string) {
	if a.teamName == name {
		return
	}
	a.enemyTeams[name] = true
}

func (a *Actor) IsHostileTowards(attacker *Actor) bool {
	if a.mood != Hostile {
		return false
	}
	if _, exists := a.enemyActors[attacker.GetInternalName()]; exists {
		return true
	}
	if _, exists := a.enemyTeams[attacker.GetTeam()]; exists {
		return true
	}
	return false
}

func (a *Actor) IsPanicking() bool {
	return a.mood == Panic
}

func (a *Actor) LookInfo() string {
	if !a.IsAlive() {
		return fmt.Sprintf("%s (dead)", a.Name())
	}
	if a.IsSleeping() {
		return fmt.Sprintf("%s (sleeping)", a.Name())
	}
	if a.IsKnockedDown() {
		return fmt.Sprintf("%s (knocked down)", a.Name())
	}
	if a.NeedsHealing() {
		return fmt.Sprintf("%s (%s)", a.Name(), a.injuredString())
	}
	return a.Name()
}

func (a *Actor) ModifyDamageByArmor(damage SourcedDamage, bodyPart int) SourcedDamage {
	if !a.GetEquipment().HasArmorEquipped() {
		return damage
	}

	armor := a.GetEquipment().GetArmor()
	armoInfo := armor.GetArmor()

	var reduction int
	var threshold int

	if damage.DamageType.IsEnergy() {
		protection := armoInfo.GetProtection(special.DamageTypeLaser)
		threshold = protection.DamageThreshold
		reduction = protection.DamageReduction
	} else {
		protection := armoInfo.GetProtection(special.DamageTypeNormal)
		threshold = protection.DamageThreshold
		reduction = protection.DamageReduction
	}
	originalDamageAmount := damage.DamageAmount

	newDamageAmount := max(0, originalDamageAmount-threshold)

	if newDamageAmount > 0 {
		cappedReduction := min(reduction, 90)
		reductionFactor := (100 - cappedReduction) / 100.0
		newDamageAmount = max(1, newDamageAmount*reductionFactor)
	}
	damage.DamageAmount = newDamageAmount
	return damage
}

func (a *Actor) HasDialogue() bool {
	return a.dialogueFile != ""
}

func (a *Actor) HasStealableItems() bool {
	return a.GetInventory().HasStealableItems(a.GetEquipment().IsNotEquipped)
}

func (a *Actor) IsKnockedDown() bool {
	return a.HasFlag(foundation.FlagKnockedDown)
}

func (a *Actor) injuredString() string {
	percent := a.GetHitPoints() * 100 / a.GetHitPointsMax()
	if percent < 10 {
		return "near death"
	}
	if percent < 25 {
		return "severely wounded"
	}
	if percent < 50 {
		return "badly injured"
	}
	if percent < 75 {
		return "injured"
	}
	return "scratched"
}

func (a *Actor) RemoveEnemy(other *Actor) {
	delete(a.enemyActors, other.GetInternalName())
}

func (a *Actor) SetNeutral() {
	a.mood = Neutral
}

func (a *Actor) SetIcon(icon textiles.TextIcon) {
	a.icon = icon
}

func (a *Actor) SetCharSheet(character *special.CharSheet) {
	a.charSheet = character
}

func (a *Actor) SetXP(xp int) {
	a.xp = xp
}

func (a *Actor) ToRecord() recfile.Record {
	actorRecord := append(recfile.Record{
		recfile.Field{Name: "Name", Value: a.name},
		recfile.Field{Name: "InternalName", Value: a.internalName},
		recfile.Field{Name: "Icon", Value: string(a.icon.Char)},
		recfile.Field{Name: "Fg", Value: recfile.RGBStr(a.icon.Fg)},
		recfile.Field{Name: "Bg", Value: recfile.RGBStr(a.icon.Bg)},
		recfile.Field{Name: "Mood", Value: recfile.IntStr(int(a.mood))},
		recfile.Field{Name: "DialogueFile", Value: a.dialogueFile},
		recfile.Field{Name: "Team", Value: a.teamName},
		recfile.Field{Name: "XP", Value: recfile.IntStr(a.xp)},
	}, a.charSheet.ToRecord()...)
	return actorRecord
}
