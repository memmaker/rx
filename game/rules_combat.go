package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"math/rand"
)

var maxArmorDR = 25
var maxArmorDT = 30
var ablationWithoutPenetration = 1
var ablationWithPenetration = 2

func (g *GameState) getRangedChanceToHit(attacker *Actor, equippedWeapon *Weapon, defender *Actor, bulletsSpent *Ammo, situationalMods []d100.Modifier) (int, d100.Modifiers) {
	distance := g.currentMap().MoveDistance(attacker.Position(), defender.Position())
	weaponRange := equippedWeapon.Range()

	isFullAuto := equippedWeapon.GetCurrentAttackMode().Mode.IsFullAuto()

	illuminationAtTarget := d100.RangedAttackModifier(0)

	brightnessAtTarget := g.LightAt(defender.Position()).Brightness()

	if brightnessAtTarget < 0.28 {
		illuminationAtTarget = d100.RngModLightDarkness
	} else if brightnessAtTarget < 0.48 {
		illuminationAtTarget = d100.RngModLightDim
	}

	infos := d100.NewRangedModsFromSituation(situationalMods).
		WithRangeMods(distance, weaponRange, isFullAuto, bulletsSpent.StackSize()).
		With(illuminationAtTarget)

	if attacker.HasFlag(foundation.FlagRunning) {
		infos = infos.With(d100.RngModFiringWhileRunning)
	}

	if attacker.HasFlag(foundation.FlagConcentratedAiming) {
		infos = infos.WithAimTurns(attacker.GetFlags().Get(foundation.FlagConcentratedAiming))
	} else if !defender.IsInCombat() {
		infos = infos.With(d100.RngModSurprised)
	}

	if defender.IsInCombat() && defender.GetBasicSpeed() > attacker.GetBasicSpeed() {
		delta := defender.GetBasicSpeed() - attacker.GetBasicSpeed()
		infos = infos.WithSpeedMods(delta)
	}

	if defender.HasFlag(foundation.FlagActiveCamouflage) {
		infos = infos.With(d100.RngModTargetCamouflaged)
	}

	mods := infos

	weaponSkill := equippedWeapon.GetSkillUsed()
	baseSkill := attacker.GetCharSheet().GetSkill(weaponSkill)

	moddedSkill := baseSkill
	for _, mod := range mods {
		moddedSkill = mod.Apply(moddedSkill)
	}

	hitChance := min(d100.SuccessChanceCap, max(0, moddedSkill))

	return hitChance, d100.Modifiers(mods)
}

func (g *GameState) calculateRangedDamage(attacker *Actor, weaponItem *Weapon, attackMode AttackMode, bulletsSpent *Ammo, chanceToHit int, victim *Actor, bodyPart d100.BodyPart, mods []d100.Modifier) (SourcedDamage, d100.CheckResult) {
	weapon := weaponItem
	damage := weaponItem.GetWeaponDamage()
	critChance := attacker.GetCharSheet().GetDerivedStat(d100.CriticalChance)
	totalDamage := 0
	damagePerBullet := make([]int, bulletsSpent.StackSize())
	attackResult := d100.SuccessRoll(d100.Percentage(chanceToHit), d100.Percentage(critChance))
	if attackResult.Success {
		firstBulletDamage := damage.Roll()
		totalDamage = firstBulletDamage
		damagePerBullet[0] = firstBulletDamage
		bulletsLeft := bulletsSpent.StackSize() - 1
		for i := 0; i < bulletsLeft; i++ {
			damageDone := 0
			if rand.Intn(100)+1 < chanceToHit {
				damageDone = damage.Roll()
			}
			damagePerBullet[i+1] = damageDone
			totalDamage += damageDone
		}
	}
	damageFactor := 1.0
	bonusDamage := 0

	if bulletsSpent != nil && bulletsSpent.StackSize() > 0 && totalDamage > 0 {
		if victim != nil {
			for tags, dmgBonus := range bulletsSpent.BonusDamageAgainstActorWithTags {
				if victim.HasFlag(tags) {
					bonusDamage += dmgBonus
				}
			}
		}

		damageFactor = bulletsSpent.DamageFactor
	}

	baseDamage := int(float64(totalDamage)*damageFactor) + bonusDamage

	// apply mods
	for _, mod := range mods {
		baseDamage = mod.Apply(baseDamage)
	}

	dtMod := 0
	if weaponItem != nil {
		dtMod = weapon.GetTargetDTModifier()
	}

	damageWithSource := SourcedDamage{
		NameOfThing:      "ranged_weapon_damage",
		Attacker:         attacker,
		IsObviousAttack:  true,
		TargetingMode:    attackMode.Mode,
		DamageType:       weapon.GetDamageType(),
		DamageAmount:     baseDamage,
		BodyPart:         bodyPart,
		DamagePerBullet:  damagePerBullet,
		AppliedMods:      mods,
		TargetDTModifier: dtMod,
	}

	if victim != nil {
		if attackResult.IsCriticalSuccess() {
			critMod := attacker.GetCharSheet().GetDerivedStat(d100.CriticalImpactModifier)
			damageWithSource = damageWithSource.MakeItCritical(critMod, victim)
		}

		damageWithSource = damageWithSource.ModifyDamageByArmor(victim)
	}

	return damageWithSource, attackResult
}

func (g *GameState) getMeleeChanceToHit(attacker *Actor, weaponItem *Weapon, defender *Actor, situationalMods []d100.Modifier) (int, []d100.Modifier) {
	attackerSkill := d100.SkillForUnarmed
	if weaponItem != nil && weaponItem.IsMeleeWeapon() {
		attackerSkill = weaponItem.GetSkillUsed()
	}

	baseSkill := attacker.GetCharSheet().GetSkill(attackerSkill)

	moddedSkill := baseSkill

	for _, mod := range situationalMods {
		moddedSkill = mod.Apply(moddedSkill)
	}

	computedCtH := max(0, moddedSkill)
	hitChance := min(d100.SuccessChanceCap, computedCtH)

	return hitChance, situationalMods
}

func (g *GameState) getMeleeDamage(attacker *Actor, cth int, victim *Actor, part d100.BodyPart, mods []d100.Modifier) SourcedDamage {
	punchBaseDamage := 3
	kickBaseDamage := 5

	targetingMode := TargetingModePunch
	if rand.Intn(100) < 50 {
		targetingMode = TargetingModeKick
	}
	damageType := DamageTypeNormal
	meleeDamageBonus := attacker.GetMeleeDamageBonus()

	damage := punchBaseDamage + meleeDamageBonus

	if targetingMode == TargetingModeKick {
		damage = kickBaseDamage + meleeDamageBonus
	}

	itemInHand, hasItem := attacker.GetEquipment().GetMeleeWeapon()

	if hasItem && itemInHand.IsMeleeWeapon() {
		weapon := itemInHand
		damage = meleeDamageBonus + itemInHand.GetWeaponDamage().Roll()
		damageType = weapon.GetDamageType()
	}

	critChance := attacker.GetCharSheet().GetDerivedStat(d100.CriticalChance)
	checkResult := d100.SuccessRoll(d100.Percentage(cth), d100.Percentage(critChance))
	if !checkResult.Success {
		damage = 0
	} else { // apply mods
		for _, mod := range mods {
			damage = mod.Apply(damage)
		}
	}

	damageWithSource := SourcedDamage{
		NameOfThing:     "melee",
		Attacker:        attacker,
		IsObviousAttack: true,
		TargetingMode:   targetingMode,
		DamageType:      damageType,
		DamageAmount:    damage,
		BodyPart:        part,
		AppliedMods:     mods,
	}

	if victim != nil {
		if checkResult.IsCriticalSuccess() {
			critMod := attacker.GetCharSheet().GetDerivedStat(d100.CriticalImpactModifier)
			damageWithSource = damageWithSource.MakeItCritical(critMod, victim)
		}
		damageWithSource = damageWithSource.ModifyDamageByArmor(victim)
	}

	return damageWithSource
}
