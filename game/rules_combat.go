package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"math/rand"
)

var maxArmorDR = 85
var maxArmorDT = 30
var ablationWithoutPenetration = 1
var ablationWithPenetration = 2

func (g *GameState) getRangedChanceToHit(attacker *Actor, equippedWeapon *Weapon, defender *Actor, bulletsSpent *Ammo) (int, d100.RangedModifiers) {
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

	infos := d100.RangedModifiers{}.
		WithRangeMods(distance, weaponRange, isFullAuto, bulletsSpent.StackSize()).
		With(illuminationAtTarget)

	if attacker.HasFlag(foundation.FlagRunning) {
		infos = infos.With(d100.RngModFiringWhileRunning)
	}

	if attacker.HasFlag(foundation.FlagConcentratedAiming) {
		infos = infos.WithAimTurns(attacker.GetFlags().Get(foundation.FlagConcentratedAiming))
	}

	mods := infos

	weaponSkill := equippedWeapon.GetSkillUsed()
	baseSkill := attacker.GetCharSheet().GetSkill(weaponSkill)

	moddedSkill := baseSkill
	for _, mod := range mods {
		moddedSkill = mod.Apply(moddedSkill)
	}

	hitChance := min(d100.SuccessChanceCap, max(0, moddedSkill))

	return hitChance, mods
}

func (g *GameState) calculateRangedDamage(attacker *Actor, weaponItem *Weapon, attackMode AttackMode, bulletsSpent *Ammo, chanceToHit int, victim *Actor, bodyPart d100.BodyPart) SourcedDamage {
	weapon := weaponItem
	damage := weaponItem.GetWeaponDamage()
	totalDamage := 0
	damagePerBullet := make([]int, bulletsSpent.StackSize())
	for i := 0; i < bulletsSpent.StackSize(); i++ {
		damageDone := damage.Roll()
		if rand.Intn(100)+1 >= chanceToHit {
			damageDone = 0
		}
		damagePerBullet[i] = damageDone
		totalDamage += damageDone
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

	damageWithSource := SourcedDamage{
		NameOfThing:     "ranged_weapon_damage",
		Attacker:        attacker,
		IsObviousAttack: true,
		TargetingMode:   attackMode.Mode,
		DamageType:      weapon.GetDamageType(),
		DamageAmount:    int(float64(totalDamage)*damageFactor) + bonusDamage,
		BodyPart:        bodyPart,
		DamagePerBullet: damagePerBullet,
	}
	return damageWithSource
}

func (g *GameState) getMeleeChanceToHit(attacker *Actor, weaponItem *Weapon, defender *Actor) int {
	attackerSkill := d100.SkillForUnarmed
	if weaponItem != nil && weaponItem.IsMeleeWeapon() {
		attackerSkill = weaponItem.GetSkillUsed()
	}
	chanceToHit := d100.MeleeChanceToHit(attacker.GetCharSheet(), attackerSkill, defender.GetCharSheet())
	return chanceToHit
}

func (g *GameState) getMeleeDamage(attacker *Actor, cth int, part d100.BodyPart) SourcedDamage {
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

	isHit := rand.Intn(100)+1 < cth
	if !isHit {
		damage = 0
	}

	damageWithSource := SourcedDamage{
		NameOfThing:     "melee",
		Attacker:        attacker,
		IsObviousAttack: true,
		TargetingMode:   targetingMode,
		DamageType:      damageType,
		DamageAmount:    damage,
		BodyPart:        part,
	}
	return damageWithSource
}

func (a *Actor) ModifyDamageByArmor(damage SourcedDamage, dtModifierFromAttack int) SourcedDamage {
	if damage.DamageAmount == 0 {
		return damage
	}
	reduction := a.GetCharSheet().GetDerivedStat(d100.DamageResistance)
	threshold := 0
	originalDamageAmount := damage.DamageAmount

	if a.GetEquipment().HasArmorEquipped() {
		armor := a.GetEquipment().GetArmor()
		if damage.DamageType.IsEnergy() {
			protection := armor.GetArmorProtection(DamageTypeLaser)
			threshold = protection.DamageThreshold
			reduction += protection.DamageReduction
		} else {
			protection := armor.GetArmorProtection(DamageTypeNormal)
			threshold = protection.DamageThreshold
			reduction += protection.DamageReduction
		}

	}

	reduction = max(0, min(maxArmorDR, reduction))
	threshold = max(0, min(maxArmorDT, threshold+dtModifierFromAttack))

	reductionFactor := (100 - float64(reduction)) / 100.0
	var newDamageAmount int

	if len(damage.DamagePerBullet) > 0 {
		for _, bulletDamage := range damage.DamagePerBullet {
			bulletDamage = int(max(1, float64(bulletDamage)*reductionFactor))
			bulletDamage = max(0, bulletDamage-threshold)
			newDamageAmount += bulletDamage
		}
	} else {
		newDamageAmount = int(max(1, float64(originalDamageAmount)*reductionFactor))
		newDamageAmount = max(0, originalDamageAmount-threshold)
	}

	// degrade armor
	if a.GetEquipment().HasArmorEquipped() {
		ablation := ablationWithoutPenetration
		if newDamageAmount > 0 {
			ablation = ablationWithPenetration
		}
		armor := a.GetEquipment().GetArmor()
		armor.DegradeDT(ablation)
	}

	damage.DamageAmount = newDamageAmount
	return damage
}
