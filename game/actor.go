package game

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/rpg"
	"RogueUI/util"
	"fmt"
	"image/color"
	"math"
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
	charSheet    *rpg.Character
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
}

func (a *Actor) Color() string {
	return a.color
}

func NewPlayer(name string, playerIcon rune, playerColor string) *Actor {
	player := NewActor(name, playerIcon, playerColor)
	player.charSheet.SetCharacterPointsReceived(7)
	player.charSheet.SetAdjustment(rpg.Strength, 0)
	player.charSheet.SetAdjustment(rpg.Dexterity, 0)
	player.charSheet.SetAdjustment(rpg.Intelligence, 0)
	player.charSheet.SetAdjustment(rpg.Health, 0)
	player.charSheet.SetAdjustment(rpg.Will, 0)
	player.charSheet.SetAdjustment(rpg.Perception, 0)
	player.charSheet.SetSkillLevel(rpg.SkillNameMeleeWeapons, 2)
	player.charSheet.SetSkillLevel(rpg.SkillNameBrawling, 1)
	player.charSheet.SetSkillLevel(rpg.SkillNameShield, 1)
	player.charSheet.SetSkillLevel(rpg.SkillNameMissileWeapons, 1)
	player.charSheet.SetSkillLevel(rpg.SkillNameThrowing, 1)
	player.charSheet.SetSkillLevel(rpg.SkillNameStealth, 1)
	return player
}

func NewActor(name string, icon rune, color string) *Actor {
	characterSheet := rpg.NewCharacterSheet()

	a := &Actor{
		name:        name,
		icon:        icon,
		color:       color,
		inventory:   NewInventory(23),
		equipment:   NewEquipment(),
		charSheet:   characterSheet,
		statusFlags: foundation.NewMapFlags(),
	}

	// add persistent modifiers, that always apply when a condition is met here

	// injured -> 1/2 dodge
	characterSheet.AddStatModifier(rpg.Dodge, ModHalveWhen("injured", a.IsInjured))

	// fatigued -> 1/2 dodge & strength
	characterSheet.AddStatModifier(rpg.Dodge, ModHalveWhen("fatigued", a.IsFatigued))
	characterSheet.AddStatModifier(rpg.Strength, ModHalveWhen("fatigued", a.IsFatigued))

	characterSheet.AddStatModifier(rpg.Dodge, ModFlatWhen(-4, "stunned", a.IsStunned))
	characterSheet.AddStatModifier(rpg.Dodge, ModFlatWhen(-4, "stunned", a.IsStunned))
	characterSheet.AddStatModifier(rpg.Dodge, ModFlatWhen(-4, "stunned", a.IsStunned))

	characterSheet.AddStatModifier(rpg.Strength, ModFlatWhen(-1, "hungry", a.IsHungry))
	characterSheet.AddStatModifier(rpg.Strength, ModFlatWhen(-3, "starving", a.IsStarving))

	return a
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

func ModCapWhen(maxValue int, description string, conditionForApplication func() bool) rpg.Modifier {
	return CapModifier{
		maxValue:    maxValue,
		doesApply:   conditionForApplication,
		description: description,
		persistent:  true,
	}
}
func ModCap(maxValue int, reason string) rpg.Modifier {
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

func ModFlatWhen(flatMod int, reason string, conditionForApplication func() bool) rpg.Modifier {
	return FlatModifier{
		flatMod:     flatMod,
		doesApply:   conditionForApplication,
		description: fmt.Sprintf("%d - %s", flatMod, reason),
		persistent:  true,
	}
}
func ModFlat(flatMod int, reason string) rpg.Modifier {
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
	hp := a.charSheet.GetResource(rpg.HitPoints)
	hpMax := a.charSheet.GetResourceMax(rpg.HitPoints)
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
	totalDRFromArmor := 0
	if a.GetEquipment().HasArmorEquipped() {
		armorPieces := a.GetEquipment().GetArmor()
		for _, armorPiece := range armorPieces {
			totalDRFromArmor += armorPiece.GetArmor().GetDamageResistanceWithPlus()
		}
	}
	totalDRFromEquipment := a.GetEquipment().GetStatModifier(rpg.DamageResistance)

	return totalDRFromArmor + totalDRFromEquipment
}

func (a *Actor) GetFlags() *foundation.MapFlags {
	return a.statusFlags
}

func (a *Actor) IsAlive() bool {
	return a.charSheet.GetResource(rpg.HitPoints) > 0
}

func (a *Actor) HasFlag(flag foundation.ActorFlag) bool {
	return a.statusFlags.IsSet(flag) || a.GetEquipment().ContainsFlag(flag)
}

func (a *Actor) TakeDamage(amount int) {
	a.charSheet.DecreaseResourceBy(rpg.HitPoints, amount)
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
}

func (a *Actor) Heal(amount int) {
	a.charSheet.IncreaseResourceBy(rpg.HitPoints, amount)
}

func (a *Actor) GetInternalName() string {
	return a.internalName
}

func (a *Actor) SetInternalName(name string) {
	a.internalName = name
}
func (a *Actor) GetMelee(enemyInternalName string) (effectiveSkill int, damageDice rpg.Dice) {
	equipment := a.GetEquipment()
	if equipment.HasMeleeWeaponEquipped() {
		weapon := equipment.GetMainWeapon(MeleeAttack).GetWeapon()
		toHit, toDamage := weapon.GetVorpalBonus(enemyInternalName)
		strBonus := min(10, a.GetStrength()-10)
		toDamage += strBonus
		return a.GetSkill(weapon.GetSkillUsed()) + toHit, weapon.GetDamageDice().WithBonus(toDamage)
	}
	if len(a.intrinsicAttacks) > 0 {
		return a.ChooseIntrinsicAttack()
	}
	meleeDamage := rpg.GetBasicMeleeDamageFromStrength(a.charSheet.GetStat(rpg.Strength))
	return a.GetSkill(rpg.SkillNameBrawling), meleeDamage.ThrustDice
}
func (a *Actor) GetThrowing(enemyInternalName string, thrownItem *Item) (effectiveSkill int, damageDice rpg.Dice) {
	return a.GetSkill(rpg.SkillNameThrowing), thrownItem.GetThrowDamageDice()
}
func (a *Actor) GetRanged(enemyInternalName string, launcher, missile *Item) (effectiveSkill int, damageDice rpg.Dice) {
	missileWeapon := missile.GetWeapon()
	toHit, toDamage := missileWeapon.GetVorpalBonus(enemyInternalName)

	missileDamage := missileWeapon.GetDamageDice().WithBonus(toDamage)

	launcherWeapon := launcher.GetWeapon()
	return a.GetSkill(launcherWeapon.GetSkillUsed()) + toHit, missileDamage
}
func (a *Actor) ChooseIntrinsicAttack() (int, rpg.Dice) {
	if a.intrinsicAttacks == nil || len(a.intrinsicAttacks) == 0 {
		return 0, rpg.Dice{}
	}
	randomIndexOfIntrinsicDamage := rand.Intn(len(a.intrinsicAttacks))
	a.currentIntrinsicAttack = randomIndexOfIntrinsicDamage
	return a.intrinsicAttacks[randomIndexOfIntrinsicDamage].BaseSkill, a.intrinsicAttacks[randomIndexOfIntrinsicDamage].DamageDice
}
func (a *Actor) GetHitPoints() int {
	return a.charSheet.GetResource(rpg.HitPoints)
}

func (a *Actor) GetHitPointsMax() int {
	return a.charSheet.GetResourceMax(rpg.HitPoints)
}

func (a *Actor) GetActiveDefenseScore() int {
	chosenDefense := rpg.ActiveDefenseDodge
	return a.charSheet.GetActiveDefenseScore(chosenDefense, a.GetParryDefenseScore())
}
func (a *Actor) IsInjured() bool {
	hpMax := a.GetHitPointsMax()
	hpCurrent := a.GetHitPoints()
	belowOneThirdHitPoints := hpCurrent < hpMax/3
	return belowOneThirdHitPoints
}
func (a *Actor) IsFatigued() bool {
	fpMax := a.GetFatiguePointsMax()
	fpCurrent := a.GetFatiguePoints()
	belowOneThird := fpCurrent < fpMax/3
	return belowOneThird
}

func (a *Actor) GetBasicSpeed() int {
	factor := 1.0
	if a.HasFlag(foundation.FlagSlow) {
		factor = 0.5
	} else if a.HasFlag(foundation.FlagHaste) {
		factor = 2.0
	}
	return int(float64(a.charSheet.GetStat(rpg.BasicSpeed)) * factor)
}

func (a *Actor) GetBlockDefenseScore() int {
	return a.charSheet.GetStat(rpg.Block)
}
func (a *Actor) GetParryDefenseScore() int {
	parryBaseValue := 3 + (a.getMeleeSkillInUse() / 2)
	return a.charSheet.ApplyStatModifiers(rpg.Parry, parryBaseValue)
}
func (a *Actor) GetDodgeDefenseScore() int {
	return a.charSheet.GetStat(rpg.Dodge)
}
func (a *Actor) getMeleeSkillInUse() int {
	equipment := a.GetEquipment()
	if equipment.HasMeleeWeaponEquipped() {
		weapon := equipment.GetMainWeapon(MeleeAttack).GetWeapon()
		return a.GetSkill(weapon.GetSkillUsed())
	}
	if len(a.intrinsicAttacks) > 0 {
		return a.intrinsicAttacks[a.currentIntrinsicAttack].BaseSkill
	}
	return a.GetSkill(rpg.SkillNameBrawling)
}

func (a *Actor) GetSkill(name rpg.SkillName) int {
	return a.charSheet.GetSkill(name) + a.GetEquipment().GetSkillModifier(name)
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
	skillLineWithSuccess := func(skillName rpg.SkillName, skillValue int) []string {
		asFloatPercent := rpg.ChanceOfSuccess(skillValue)
		return []string{
			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnSkill(skillName)),
			fmt.Sprintf("%s:", string(skillName)),
			fmt.Sprintf("%d", skillValue),
			fmt.Sprintf("(%d%%)", int(asFloatPercent*100)),
		}
	}
	statLineWithSuccess := func(statName rpg.Stat, skillValue int) []string {
		asFloatPercent := rpg.ChanceOfSuccess(skillValue)
		return []string{
			fmt.Sprintf("%s:", statName.ToString()),
			fmt.Sprintf("%d", skillValue),
			fmt.Sprintf("(%d%%)", int(asFloatPercent*100)),
		}
	}
	rows := util.TableLayout([]util.TableRow{
		{Columns: []string{
			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnStat(rpg.Strength)),
			"ST:",
			fmt.Sprintf("%d", a.GetStrength()),

			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnStat(rpg.HitPoints)),
			"HP:",
			fmt.Sprintf("%d/%d", a.GetHitPoints(), a.GetHitPointsMax()),
		}},
		{Columns: []string{
			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnStat(rpg.Dexterity)),
			"DX:",
			fmt.Sprintf("%d", a.GetDexterity()),

			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnStat(rpg.FatiguePoints)),
			"FP:",
			fmt.Sprintf("%d/%d", a.GetFatiguePoints(), a.GetFatiguePointsMax()),
		}},
		{Columns: []string{
			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnStat(rpg.Intelligence)),
			"IN:",
			fmt.Sprintf("%d", a.GetIntelligence()),

			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnStat(rpg.Perception)),
			"Per:",
			fmt.Sprintf("%d", a.charSheet.GetStat(rpg.Perception)),
		}},
		{Columns: []string{
			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnStat(rpg.Health)),
			"HT:",
			fmt.Sprintf("%d", a.charSheet.GetStat(rpg.Health)),

			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnStat(rpg.Will)),
			"Will:",
			fmt.Sprintf("%d", a.charSheet.GetStat(rpg.Will)),
		}},
		{Columns: []string{
			fmt.Sprintf("[%d]", a.charSheet.GetPointsSpentOnStat(rpg.BasicSpeed)),
			"Speed:",
			fmt.Sprintf("%d", a.GetBasicSpeed()), "", "", ""}},
	}, []util.TextAlignment{util.AlignLeft, util.AlignLeft, util.AlignLeft, util.AlignLeft, util.AlignLeft, util.AlignLeft})

	var result []string
	result = append(result, fmt.Sprintf("Name: %s [%d/%d]", a.Name(), a.charSheet.GetCharacterPointsBalance(), a.charSheet.GetCharacterPointsReceived()), "")

	result = append(result, rows...)
	result = append(result, "", "> Attacks:")

	if len(a.intrinsicAttacks) > 0 {
		var attackRows []util.TableRow
		for _, intrinsicAttack := range a.intrinsicAttacks {
			asFloatPercent := rpg.ChanceOfSuccess(intrinsicAttack.BaseSkill)
			attackRows = append(attackRows, util.TableRow{
				Columns: []string{
					fmt.Sprintf("%s:", intrinsicAttack.AttackName),
					fmt.Sprintf("%d", intrinsicAttack.BaseSkill),
					fmt.Sprintf("(%d%%)", int(asFloatPercent*100)),
					fmt.Sprintf("Dmg: %s", intrinsicAttack.DamageDice.ShortString()),
				},
			})
		}
		attackLines := util.TableLayout(attackRows, []util.TextAlignment{util.AlignLeft, util.AlignLeft, util.AlignLeft, util.AlignLeft})
		result = append(result, attackLines...)
	} else {
		var attackRows []util.TableRow
		// melee attack
		meleeSkill, meleeDamage := a.GetMelee("")
		asFloatPercent := rpg.ChanceOfSuccess(meleeSkill)
		attackRows = append(attackRows, util.TableRow{
			Columns: []string{
				"melee:",
				fmt.Sprintf("%d", meleeSkill),
				fmt.Sprintf("(%d%%)", int(asFloatPercent*100)),
				fmt.Sprintf("Dmg: %s", meleeDamage.ShortString()),
			},
		})

		// ranged
		if a.GetEquipment().HasMissileLauncherEquipped() {
			launcher := a.GetEquipment().GetMissileLauncher()
			missile := a.GetEquipment().GetNextQuiveredMissile()
			if missile != nil {
				rangedSkill, rangedDamage := a.GetRanged("", launcher, missile)
				asFloatPercent = rpg.ChanceOfSuccess(rangedSkill)
				attackRows = append(attackRows, util.TableRow{
					Columns: []string{
						"ranged:",
						fmt.Sprintf("%d", rangedSkill),
						fmt.Sprintf("(%d%%)", int(asFloatPercent*100)),
						fmt.Sprintf("Dmg: %s", rangedDamage.ShortString()),
					},
				})
			}
		}
		attackLines := util.TableLayout(attackRows, []util.TextAlignment{util.AlignLeft, util.AlignLeft, util.AlignLeft, util.AlignLeft})
		result = append(result, attackLines...)
	}

	result = append(result, "", "> Defenses:")

	defenseRows := util.TableLayout([]util.TableRow{
		{Columns: statLineWithSuccess(rpg.Dodge, a.GetDodgeDefenseScore())},
		{Columns: statLineWithSuccess(rpg.Block, a.GetBlockDefenseScore())},
		{Columns: statLineWithSuccess(rpg.Parry, a.GetParryDefenseScore())},
	}, []util.TextAlignment{util.AlignLeft, util.AlignLeft, util.AlignLeft})

	result = append(result, defenseRows...)
	result = append(result, "", "> Skills:")

	skillRows := util.TableLayout([]util.TableRow{
		{Columns: skillLineWithSuccess(rpg.SkillNameBrawling, a.GetSkill(rpg.SkillNameBrawling))},
		{Columns: skillLineWithSuccess(rpg.SkillNameMeleeWeapons, a.GetSkill(rpg.SkillNameMeleeWeapons))},
		{Columns: skillLineWithSuccess(rpg.SkillNameShield, a.GetSkill(rpg.SkillNameShield))},
		{Columns: skillLineWithSuccess(rpg.SkillNameThrowing, a.GetSkill(rpg.SkillNameThrowing))},
		{Columns: skillLineWithSuccess(rpg.SkillNameMissileWeapons, a.GetSkill(rpg.SkillNameMissileWeapons))},
		{Columns: skillLineWithSuccess(rpg.SkillNameStealth, a.GetSkill(rpg.SkillNameStealth))},
	}, []util.TextAlignment{util.AlignLeft, util.AlignLeft, util.AlignLeft, util.AlignLeft})

	result = append(result, skillRows...)

	if len(a.intrinsicUseEffects) > 0 || len(a.intrinsicZapEffects) > 0 {
		result = append(result, "", "> Abilities:")
		for _, abilityName := range a.intrinsicUseEffects {
			result = append(result, abilityName)
		}
		for _, abilityName := range a.intrinsicZapEffects {
			result = append(result, abilityName)
		}
	}

	return result
}

func (a *Actor) GetPointsInfo() []string {
	return a.charSheet.GetOverview()
}

func (a *Actor) SetIntrinsicAttacks(attacks []IntrinsicAttack) {
	a.intrinsicAttacks = attacks
}

func (a *Actor) GetStatModifier(stat rpg.Stat) int {
	equipMod := a.GetEquipment().GetStatModifier(stat)
	rulesMod := rpg.GetStatModifier(stat, a.charSheet, a.GetEncumbrance())
	return equipMod + rulesMod
}

func (a *Actor) GetEncumbrance() rpg.Encumbrance {
	if !a.GetEquipment().HasArmorEquipped() {
		return rpg.EncumbranceNone
	}
	armorEncumbrance := a.GetEquipment().GetEncumbranceFromArmor()
	return armorEncumbrance
}

func (a *Actor) GetFatiguePoints() int {
	return a.charSheet.GetResource(rpg.FatiguePoints)
}
func (a *Actor) GetFatiguePointsMax() int {
	return a.charSheet.GetResourceMax(rpg.FatiguePoints)
}

func (a *Actor) GetStrength() int {
	baseValue := a.charSheet.GetStat(rpg.Strength)
	if a.IsFatigued() {
		baseValue = int(math.Ceil(float64(baseValue) / 2.0))
	}
	return baseValue
}
func (a *Actor) GetDexterity() int {
	return a.charSheet.GetStat(rpg.Dexterity)

}
func (a *Actor) GetIntelligence() int {
	return a.charSheet.GetStat(rpg.Intelligence)

}
func (a *Actor) GetHealth() int {
	return a.charSheet.GetStat(rpg.Health)

}

func (a *Actor) IsLaunching(missile *Item) bool {
	return missile.IsMissile() && a.GetEquipment().HasMissileLauncherEquippedForMissile(missile.GetWeapon())
}

func (a *Actor) GetSizeModifier() int {
	return a.sizeModifier
}

func (a *Actor) SetSizeModifier(modifier int) {
	a.sizeModifier = modifier
}

func (a *Actor) AddCharacterPoints(amount int) {
	a.charSheet.AddCharacterPoints(amount)
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

func (a *Actor) LooseFatigue(amount int) {
	a.charSheet.DecreaseResourceBy(rpg.FatiguePoints, amount)
}

func (a *Actor) AddFatiguePoints(amount int) {
	a.charSheet.IncreaseResourceBy(rpg.FatiguePoints, amount)
}

func (a *Actor) GetPerception() int {
	return a.charSheet.GetStat(rpg.Perception)
}

func (a *Actor) CanPerceivePlayer(stealthSkillOfPlayer int, distance int) bool {
	winner := rpg.SkillContest(stealthSkillOfPlayer, a.GetPerception()-distance)
	return winner == 2
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
