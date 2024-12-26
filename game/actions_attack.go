package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

// melee attacks with and without weapons
// firearms with different modes of fire
// throwing weapons

// Logic & Animation
// - Melee Attack (with and without weapons)
// - Throw Item
// - Ranged Attack with fire mode

// OFFENSIVE ACTIONS

func (g *GameState) PlayerRangedAttack() {

	rangedWeapon, quickDraw := g.playerDrawnWeapon()
	canAttack := g.CanPlayerAttackAtRange()
	if !canAttack || rangedWeapon == nil {
		return
	}
	combatMods := d100.NoCombatModifier
	if quickDraw {
		combatMods = d100.CombatModifiers{DamageMods: []d100.Modifier{d100.QuickDrawModifier}}
	}
	attackMode := rangedWeapon.GetCurrentAttackMode()
	if attackMode.IsAimed {
		g.ui.SelectBodyPart(g.playerLastAimedAt, func(victim foundation.ActorForUI, bodyPart d100.BodyPart) {
			g.playerLastAimedAt = bodyPart
			target := victim.(*Actor)
			shotAnim := g.actorRangedAttack(g.Player, rangedWeapon, attackMode, target, bodyPart, combatMods)
			g.ui.AddAnimations(shotAnim)
			g.endPlayerTurn(attackMode.TUCost)
		})
	} else {
		g.ui.SelectTarget(g.getPlayerRangedChanceToHitForUI(combatMods), func(targetPos geometry.Point) {
			if g.currentMap().IsActorAt(targetPos) {
				target := g.currentMap().ActorAt(targetPos)
				shotAnim := g.actorRangedAttack(g.Player, rangedWeapon, attackMode, target, d100.Body, combatMods)
				g.ui.AddAnimations(shotAnim)
				g.endPlayerTurn(attackMode.TUCost)
			} else {
				shotAnim := g.actorRangedAttackLocation(g.Player, rangedWeapon, attackMode, targetPos)
				g.ui.AddAnimations(shotAnim)
				g.endPlayerTurn(attackMode.TUCost)
			}
		})
	}

}
func (g *GameState) playerDrawnWeapon() (drawnWeapon *Weapon, isQuickDraw bool) {
	mainHandItem, hasWeapon := g.Player.GetInventory().GetEquippedWeapon()
	if !hasWeapon && g.Player.HasPerk(d100.PerkQuickDraw) && g.Player.GetInventory().HasExactlyOneRangedWeapon() {
		if !g.Player.HasActionPoints() {
			g.msg(foundation.Msg("Not enough action points for quick draw"))
			return nil, false
		}
		mainHandItem = g.Player.GetInventory().GetBestWeapon()
		g.Player.GetInventory().Equip(mainHandItem)
		g.Player.GetCharSheet().LooseActionPoints(1)
		g.ui.UpdateStats()
		g.msg(foundation.HiLite("You quickly equip your %s", mainHandItem.Name()))
		return mainHandItem, true
	}
	return mainHandItem, false
}

func (g *GameState) CanPlayerAttackAtRange() (canAttack bool) {
	mainHandItem, hasWeapon := g.Player.GetInventory().GetEquippedWeapon()

	if hasWeapon && mainHandItem.GetCurrentAttackMode().IsThrow() {
		g.startThrowItem(mainHandItem)
		return false
	}

	if hasWeapon && mainHandItem.IsMeleeWeapon() && !mainHandItem.IsRangedWeapon() {
		g.ui.SelectDirection(func(direction geometry.CompassDirection) {
			targetPos := g.Player.Position().Add(direction.ToPoint())
			g.playerMeleeAttackLocation(targetPos)
		})
		return false
	}
	if !hasWeapon || !mainHandItem.IsRangedWeapon() {
		g.msg(foundation.Msg("You have no suitable weapon equipped"))
		return false
	}
	weapon := mainHandItem
	if !weapon.HasAmmo() {
		g.ui.PlayCue(weapon.GetOutOfAmmoAudioCue())
		g.msg(foundation.Msg("You have no ammo"))
		return false
	}
	if weapon.IsBroken() {
		g.msg(foundation.Msg("Your weapon is broken"))
		return false
	}

	if weapon.IsJammed() {
		g.ui.AskForConfirmation(weapon.Name(), "Your weapon is jammed. Do you want to unjam it?", func(confirmed bool) {
			if confirmed {
				weapon.Unjam()
				g.msg(foundation.Msg("You unjam your weapon"))
				g.endPlayerTurn(g.Player.TimeNeededForActions())
			}
		})
		return false
	}
	return true
}

func (g *GameState) PlayerQuickRangedAttack() {
	enemies := g.playerVisibleActorsByDistance()

	if len(enemies) == 0 {
		g.msg(foundation.Msg("No enemies in sight"))
		return
	}

	rangedWeapon, quickDraw := g.playerDrawnWeapon()
	canAttack := g.CanPlayerAttackAtRange()

	if !canAttack || rangedWeapon == nil {
		return
	}
	combatMods := d100.NoCombatModifier
	if quickDraw {
		combatMods = d100.CombatModifiers{DamageMods: []d100.Modifier{d100.QuickDrawModifier}}
	}
	mode := rangedWeapon.GetCurrentAttackMode()
	shotAnim := g.actorRangedAttack(g.Player, rangedWeapon, mode, enemies[0], d100.Body, combatMods)
	g.ui.AddAnimations(shotAnim)
	g.endPlayerTurn(mode.TUCost)
}

func (g *GameState) QuickThrow() {
	enemies := g.playerVisibleActorsByDistance()
	preselectedTarget := g.Player.Position()
	equipment := g.Player.GetInventory()
	item, hasItem := equipment.GetMainHandItem()
	if hasItem || !item.IsMissile() {
		g.msg(foundation.Msg("You have no suitable item equipped"))
		return
	}
	if len(enemies) == 0 {
		g.msg(foundation.Msg("No enemies in sight"))
		return
	}
	preselectedTarget = enemies[0].Position()
	g.actorThrowItem(g.Player, item.(foundation.Item), g.Player.Position(), preselectedTarget)
}

func (g *GameState) playerDrown(defender *Actor) {
	attackerLuckChance := d100.Percentage(g.Player.GetCharSheet().GetDerivedStat(d100.CriticalChance))
	defenderLuckChance := d100.Percentage(defender.GetCharSheet().GetDerivedStat(d100.CriticalChance))

	attackerStrength := d100.Percentage(g.Player.GetCharSheet().GetStat(d100.Strength) * 10)
	defenderStrength := d100.Percentage(defender.GetCharSheet().GetStat(d100.Strength) * 10)

	contestResult := d100.SkillContest(attackerStrength, attackerLuckChance, defenderStrength, defenderLuckChance)

	if defender.IsSleeping() || contestResult == 0 {
		sourcedDamage := SourcedDamage{
			NameOfThing:     "drowning",
			Attacker:        g.Player,
			IsObviousAttack: true,
			TargetingMode:   TargetingModeFireSingle,
			DamageType:      DamageTypeNormal,
			DamageAmount:    defender.GetHitPointsMax(),
			BodyPart:        d100.Body,
		}
		g.msg(foundation.HiLite("You drown %s", defender.Name()))
		g.ui.AddAnimations(OneAnimation(g.damageActor(sourcedDamage, defender)))
		g.endPlayerTurn(g.Player.timeNeededForMeleeAttack())
	} else {
		g.msg(foundation.HiLite("You fail to sneak up on %s", defender.Name()))
		g.playerMeleeAttack(defender)
	}
}

func (g *GameState) playerBackstab(defender *Actor) {
	attackerLuckChance := d100.Percentage(g.Player.GetCharSheet().GetDerivedStat(d100.CriticalChance))
	defenderLuckChance := d100.Percentage(defender.GetCharSheet().GetDerivedStat(d100.CriticalChance))

	attackerStealth := d100.Percentage(g.Player.GetCharSheet().GetSkill(d100.SkillForBackstabbing))
	defenderAwareness := d100.Percentage(defender.GetCharSheet().GetStat(d100.Perception) * 10)

	contestResult := d100.SkillContest(attackerStealth, attackerLuckChance, defenderAwareness, defenderLuckChance)

	if defender.IsSleeping() || contestResult == 0 {
		sourcedDamage := SourcedDamage{
			NameOfThing:     "backstab",
			Attacker:        g.Player,
			IsObviousAttack: true,
			TargetingMode:   TargetingModeFireSingle,
			DamageType:      DamageTypeNormal,
			DamageAmount:    defender.GetHitPointsMax(),
			BodyPart:        d100.Body,
		}
		g.msg(foundation.HiLite("You stab %s in the back", defender.Name()))
		g.ui.AddAnimations(OneAnimation(g.damageActor(sourcedDamage, defender)))
		g.endPlayerTurn(g.Player.timeNeededForMeleeAttack())
	} else {
		g.msg(foundation.HiLite("You fail to sneak up on %s", defender.Name()))
		g.playerMeleeAttack(defender)
	}
}
func (g *GameState) actorDisarm(attacker, victim *Actor) {
	weaponTaken, hasWeapon := victim.GetInventory().GetEquippedWeapon()
	if !hasWeapon {
		g.msg(foundation.Msg("Target has no weapon equipped"))
		return
	}

	attackerLuckChance := d100.Percentage(attacker.GetCharSheet().GetDerivedStat(d100.CriticalChance))
	defenderLuckChance := d100.Percentage(victim.GetCharSheet().GetDerivedStat(d100.CriticalChance))

	atkSkill := d100.Percentage(attacker.GetCharSheet().GetSkill(d100.SkillForUnarmed))
	defSkill := d100.Percentage(victim.GetCharSheet().GetSkill(d100.SkillForUnarmed))

	contestResult := d100.SkillContest(atkSkill, attackerLuckChance, defSkill, defenderLuckChance)

	if contestResult == 0 {

		victim.GetInventory().RemoveItem(weaponTaken)
		attacker.GetInventory().AddItem(weaponTaken)
		attacker.GetInventory().Equip(weaponTaken)

		if attacker == g.Player {
			g.trySetHostile(victim, attacker)
			g.msg(foundation.HiLite("You take the %s from %s", weaponTaken.Name(), victim.Name()))
			g.endPlayerTurn(attacker.timeNeededForMeleeAttack())
		} else if victim == g.Player {
			g.msg(foundation.HiLite("%s takes the %s from you", attacker.Name(), weaponTaken.Name()))
		} else {
			g.trySetHostile(victim, attacker)
			g.msg(foundation.HiLite("%s takes the %s from %s", attacker.Name(), weaponTaken.Name(), victim.Name()))
		}
	} else {
		if attacker == g.Player {
			g.trySetHostile(victim, attacker)
			g.msg(foundation.HiLite("%s is able to resist your attempt", victim.Name()))
			g.endPlayerTurn(attacker.timeNeededForMeleeAttack())
		} else if victim == g.Player {
			g.msg(foundation.HiLite("You defend against %s's disarm attempt", attacker.Name()))
		} else {
			g.trySetHostile(victim, attacker)
			g.msg(foundation.HiLite("%s is resisting %s's disarm attempt", victim.Name(), attacker.Name()))
		}
	}
}
func (g *GameState) playerNonLethalTakedown(victim *Actor) {
	attackerLuckChance := d100.Percentage(g.Player.GetCharSheet().GetDerivedStat(d100.CriticalChance))
	defenderLuckChance := d100.Percentage(victim.GetCharSheet().GetDerivedStat(d100.CriticalChance))

	attackerStealth := d100.Percentage(g.Player.GetCharSheet().GetStat(d100.Strength) * 10)
	defenderAwareness := d100.Percentage(victim.GetCharSheet().GetStat(d100.Strength) * 10)

	contestResult := d100.SkillContest(attackerStealth, attackerLuckChance, defenderAwareness, defenderLuckChance)

	if contestResult == 0 {
		victim.SetSleeping()
		g.msg(foundation.HiLite("You knock out %s", victim.Name()))
		g.endPlayerTurn(g.Player.timeNeededForMeleeAttack())
	} else {
		g.msg(foundation.HiLite("%s is able to resist your attempt", victim.Name()))
		g.playerMeleeAttack(victim)
	}
}
func (g *GameState) playerMeleeAttackLocation(targetPos geometry.Point) {
	if g.currentMap().IsActorAt(targetPos) {
		defender := g.currentMap().ActorAt(targetPos)
		g.playerMeleeAttack(defender)
	} else if g.currentMap().IsObjectAt(targetPos) {
		objectAt := g.currentMap().ObjectAt(targetPos)
		damageWithSource := g.getMeleeDamage(g.Player, 100, nil, d100.Body, nil)
		var attackAudioCue string
		if weapon, hasMeleeWeapon := g.Player.GetInventory().GetMeleeWeapon(); hasMeleeWeapon {
			weapon.Degrade(1)
			attackAudioCue = weapon.GetFireAudioCue(damageWithSource.TargetingMode)
		} else {
			attackAudioCue = g.Player.GetMeleeAudioCue(damageWithSource.TargetingMode == TargetingModeKick)
		}
		objectAt.OnDamage(damageWithSource)
		animAttackerIndicator := g.ui.GetAnimBackgroundColor(g.Player.Position(), "dark_gray_6", 4, nil)
		animAttackerIndicator.SetAudioCue(attackAudioCue)

		g.ui.AddAnimations([]foundation.Animation{animAttackerIndicator})
		g.endPlayerTurn(g.Player.timeNeededForMeleeAttack())
	} else {
		g.msg(foundation.Msg("Nothing to attack"))
	}
}
func (g *GameState) playerMeleeAttack(defender *Actor) {
	doMeleeAttack := func(part d100.BodyPart) {
		consequences := g.actorMeleeAttack(g.Player, defender, part, d100.NoCombatModifier)
		g.ui.AddAnimations(consequences)
		g.endPlayerTurn(g.Player.timeNeededForMeleeAttack())
	}

	mainhandItem, hasWeapon := g.Player.GetInventory().GetEquippedWeapon()
	if hasWeapon && !mainhandItem.HasAmmo() && mainhandItem.IsMeleeWeapon() {
		g.ui.PlayCue(mainhandItem.GetOutOfAmmoAudioCue())
		g.msg(foundation.Msg("You have no ammo"))
		return
	}
	if hasWeapon && mainhandItem.IsMeleeWeapon() && mainhandItem.GetCurrentAttackMode().IsAimed {
		g.ui.OpenAimedShotPicker(defender, g.playerLastAimedAt, func(victim foundation.ActorForUI, bodyPart d100.BodyPart) {
			doMeleeAttack(bodyPart)
		})
	} else {
		doMeleeAttack(d100.Body)
	}
}

func (g *GameState) actorMeleeAttack(attacker *Actor, defender *Actor, part d100.BodyPart, mods d100.CombatModifiers) []foundation.Animation {
	if !defender.IsAlive() {
		return nil
	}
	var afterAttackAnimations []foundation.Animation

	mainHandItem, hasMeleeWeapon := attacker.GetInventory().GetMeleeWeapon()

	chanceToHit, _ := g.getMeleeChanceToHit(attacker, mainHandItem, defender, mods.ChanceToHitMods)
	chanceToHit += part.AimPenalty() / 2 // melee attacks are more precise

	damageWithSource := g.getMeleeDamage(attacker, chanceToHit, defender, part, mods.DamageMods)

	var attackAudioCue string
	if hasMeleeWeapon {
		mainHandItem.Degrade(0.1)
		attackAudioCue = mainHandItem.GetFireAudioCue(damageWithSource.TargetingMode)
	} else {
		attackAudioCue = attacker.GetMeleeAudioCue(damageWithSource.TargetingMode == TargetingModeKick)
	}

	animAttackerIndicator := g.ui.GetAnimBackgroundColor(attacker.Position(), "dark_gray_6", 4, nil)
	animAttackerIndicator.SetAudioCue(attackAudioCue)

	afterAttackAnimations = append(afterAttackAnimations, animAttackerIndicator)

	damageAnims := g.applyDamageToActorAnimated(attacker, mainHandItem, damageWithSource, defender)
	afterAttackAnimations = append(afterAttackAnimations, damageAnims...)

	attacker.GetFlags().Unset(foundation.FlagConcentratedAiming)

	return afterAttackAnimations
}

// actorRangedAttack logic and animation of a ranged attack with the equipped weapon
func (g *GameState) actorRangedAttack(attacker *Actor, weaponItem *Weapon, attackMode AttackMode, defender *Actor, bodyPart d100.BodyPart, situationalMods d100.CombatModifiers) []foundation.Animation {
	if !defender.IsAlive() {
		return nil
	}

	// Jamming?
	isJammed := weaponItem.DoesJam()
	if isJammed {
		attacker.GetFlags().Unset(foundation.FlagConcentratedAiming)
		g.ui.PlayCue(weaponItem.GetOutOfAmmoAudioCue())
		if attacker == g.Player {
			g.msg(foundation.Msg("Your weapon jams"))
		} else {
			g.msg(foundation.Msg(fmt.Sprintf("%s's weapon jams", attacker.Name())))
		}
		return nil
	}

	// Bookkeeping
	bulletsSpent, weapon := g.removeBulletsFromWeapon(weaponItem, attackMode)

	// Generate the attack animation based on the weapon and damage type
	attackAnimations, isProjectileAnimation := g.getWeaponAttackAnim(attacker, defender.Position(), weaponItem, attackMode, bulletsSpent.GetStackSize())

	// Calculate the chance to hit
	baseChanceToHit, _ := g.getRangedChanceToHit(attacker, weaponItem, defender, bulletsSpent, situationalMods.ChanceToHitMods)
	chanceToHit := baseChanceToHit + bodyPart.AimPenalty()

	// Pass the weapon damage to the effect
	weaponEffectParams := weaponItem.GetEffectParameters()

	var hitAnimations []foundation.Animation
	//var rollResult d100.CheckResult
	if weapon.GetDamageType() == DamageTypeExplosive { // this is a bit wacky. these will always hit all the time..
		//rollResult = d100.SuccessRoll(d100.Percentage(chanceToHit), d100.Percentage(attacker.GetCharSheet().GetDerivedStat(d100.CriticalChance)))
		hitAnimations = explosion(g, attacker, defender.Position(), weaponEffectParams)
	} else if weapon.GetDamageType() == DamageTypeFire && attackMode.Mode == TargetingModeFlame {
		//rollResult = d100.SuccessRoll(d100.Percentage(chanceToHit), d100.Percentage(attacker.GetCharSheet().GetDerivedStat(d100.CriticalChance)))
		hitAnimations = fireBreath(g, attacker, defender.Position(), weaponEffectParams)
	} else {
		// Calculate the default bullet damage
		damageWithSource, normalRollResult := g.calculateRangedDamage(attacker, weaponItem, attackMode, bulletsSpent, chanceToHit, defender, bodyPart, situationalMods.DamageMods)
		if normalRollResult.IsCriticalFailure() {
			// TODO: Apply critical fail effect..
		}
		hitAnimations = g.applyDamageToActorAnimated(attacker, weaponItem, damageWithSource, defender)
	}

	// Reset aiming flag
	attacker.GetFlags().Unset(foundation.FlagConcentratedAiming)

	// apply weapon degradation by shooting
	baseDegrade := 0.1 * float64(min(5, bulletsSpent.GetStackSize()))
	weaponItem.Degrade(baseDegrade)

	// Return animations based on whether there is a projectile or not
	if isProjectileAnimation {
		attackAnimations.SetFollowUp(hitAnimations)
		return []foundation.Animation{attackAnimations}
	} else {
		return append(hitAnimations, attackAnimations)
	}
}

func (g *GameState) actorRangedAttackLocation(attacker *Actor, weaponItem *Weapon, attackMode AttackMode, targetPos geometry.Point) []foundation.Animation {

	// Jamming?
	isJammed := weaponItem.DoesJam()
	if isJammed {
		g.ui.PlayCue(weaponItem.GetOutOfAmmoAudioCue())
		if attacker == g.Player {
			g.msg(foundation.Msg("Your weapon jams"))
		} else {
			g.msg(foundation.Msg(fmt.Sprintf("%s's weapon jams", attacker.Name())))
		}
		return nil
	}

	bulletsSpent, weapon := g.removeBulletsFromWeapon(weaponItem, attackMode)

	onAttackAnims, isProjectileAnimation := g.getWeaponAttackAnim(attacker, targetPos, weaponItem, attackMode, bulletsSpent.GetStackSize())

	chanceToHit := 100

	damageWithSource, _ := g.calculateRangedDamage(attacker, weaponItem, attackMode, bulletsSpent, chanceToHit, nil, d100.Body, nil)
	weaponEffectParams := weaponItem.GetEffectParameters()
	var consequenceOfHit []foundation.Animation
	if weapon.GetDamageType() == DamageTypeExplosive {
		consequenceOfHit = explosion(g, attacker, targetPos, weaponEffectParams)
	} else {
		if damageWithSource.DamageAmount > 0 {
			if weaponItem.IsZappable() {
				weaponZapEffect := ZapEffectFromName(weaponItem.ZapEffect())
				consequenceOfHit = weaponZapEffect(g, attacker, targetPos, weaponEffectParams)
			} else {
				consequenceOfHit = g.damageLocation(damageWithSource, targetPos)
			}
		}
	}

	// apply weapon degradation by shooting
	baseDegrade := 0.1 * float64(min(5, bulletsSpent.GetStackSize()))
	weaponItem.Degrade(baseDegrade)

	if isProjectileAnimation {
		onAttackAnims.SetFollowUp(consequenceOfHit)
		return []foundation.Animation{onAttackAnims}
	} else {
		return append(consequenceOfHit, onAttackAnims)
	}
}

func (g *GameState) removeBulletsFromWeapon(weaponItem *Weapon, attackMode AttackMode) (*Ammo, *Weapon) {
	bulletsSpent := 1
	weapon := weaponItem
	if attackMode.Mode == TargetingModeFireBurst {
		bulletsSpent = min(weapon.GetLoadedBullets(), weapon.GetBurstRounds())
	} else if attackMode.Mode == TargetingModeFireFullAuto {
		bulletsSpent = weapon.GetLoadedBullets()
	}

	bulletsRemoved := weapon.RemoveBullets(bulletsSpent)

	return bulletsRemoved, weapon
}

func (g *GameState) applyDamageToActorAnimated(attacker *Actor, weaponItem *Weapon, damageWithSource SourcedDamage, defender *Actor) []foundation.Animation {
	var damageAnims []foundation.Animation

	attackedFlag := fmt.Sprintf("WasAttacked(%s)", defender.GetInternalName())
	g.gameFlags.Increment(attackedFlag)

	if attacker == g.Player {
		attackedByPlayer := fmt.Sprintf("WasAttackedByPlayer(%s)", defender.GetInternalName())
		g.gameFlags.Increment(attackedByPlayer)
	}

	if damageWithSource.DamageAmount > 0 {
		if weaponItem != nil && weaponItem.IsZappable() {
			weaponZapEffect := ZapEffectFromName(weaponItem.ZapEffect())
			parameters := weaponItem.GetEffectParameters()
			if damageWithSource.IsCritical {
				parameters = parameters.WithCritical()
			}
			if weaponZapEffect == nil {
				panic(fmt.Sprintf("Weapon zap effect not found: %s", weaponItem.ZapEffect()))
			}
			damageAnims = weaponZapEffect(g, attacker, defender.Position(), parameters)
			if attacker == g.Player {
				g.msg(foundation.Msg("You hit"))
			} else {
				g.msg(foundation.Msg(fmt.Sprintf("%s hits", attacker.Name())))
			}
		} else {
			damageAnims = OneAnimation(g.damageActor(damageWithSource, defender))
		}
	} else {
		if damageWithSource.IsObviousAttack {
			g.trySetHostile(defender, damageWithSource.Attacker)
		}
		var playMissSound func() = nil
		if weaponItem != nil && weaponItem.IsWeapon() {
			playMissSound = func() {
				g.ui.PlayCue(weaponItem.GetMissAudioCue())
			}
		}

		evade := g.ui.GetAnimEvade(defender, playMissSound)
		evade.SetAudioCue(defender.GetDodgedAudioCue())
		damageAnims = []foundation.Animation{evade}
		if attacker == g.Player {
			g.msg(foundation.Msg("You miss"))
		} else {
			g.msg(foundation.Msg(fmt.Sprintf("%s misses", attacker.Name())))
		}
	}
	return damageAnims
}

// damageActor applies damage to an actor and returns the animations for the damage.
// The animations will also paint blood on the map.
// It will also
// - set the hostile flag on the victim if the attack is obvious
// - set the global flags that are applicable
// - play the appropriate audio cues
// - emit chatter and log messages
// - check if the victim is dead and handle that
func (g *GameState) damageActor(damage SourcedDamage, victim *Actor) foundation.Animation {
	didCripple := victim.TakeDamage(damage)

	if !victim.IsAlive() {
		damage = damage.WithKillingBlow()
	}
	if didCripple {
		damage = damage.WithCrippling()
	}
	if damage.IsObviousAttack {
		g.trySetHostile(victim, damage.Attacker)
	}
	isOverKill := victim.GetHitPoints() <= (-victim.GetHitPointsMax() / 2)
	if isOverKill {
		damage = damage.WithOverkill()
	}
	var damageAnim foundation.Animation
	var damageAudioCue string

	hurtFlag := fmt.Sprintf("WasHurt(%s)", victim.GetInternalName())
	g.gameFlags.SetFlag(hurtFlag)

	if damage.Attacker == g.Player {
		hurtByPlayerFlag := fmt.Sprintf("WasHurtByPlayer(%s)", victim.GetInternalName())
		g.gameFlags.SetFlag(hurtByPlayerFlag)
	}

	g.actorHitMessage(victim, damage)

	if damage.IsKillingBlow {
		g.actorKilled(damage, victim)
		if damage.IsCritical {
			damageAudioCue = victim.GetDeathCriticalAudioCue(damage.TargetingMode, damage.DamageType)
		} else {
			damageAudioCue = victim.GetDeathAudioCue()
		}
		// TODO: replace this with cool matching death animations
		g.makeMapBloody(victim.Position())
		damageAnim = g.ui.GetAnimDamage(g.spreadBloodAround, victim.Position(), damage.DamageAmount, 4)
		//damageAnim.SetVictimSizeModifier(victim.GetSizeModifier())
		//damageAnim.SetFollowUp(followUps)
	} else { // only a flesh wound
		damageAudioCue = victim.GetHitAudioCue(damage.TargetingMode.IsMelee())

		//
		bullets := 1
		if damage.TargetingMode.IsBurstOrFullAuto() {
			bullets = 3
		}
		damageAnim = g.ui.GetAnimDamage(g.spreadBloodAround, victim.Position(), damage.DamageAmount, bullets)
		//damageAnim.SetVictimSizeModifier(victim.GetSizeModifier())
		//damageAnim.SetFollowUp(followUps)

		if victim != g.Player && rand.Intn(5) == 0 {
			g.tryAddRandomChatter(victim, foundation.ChatterBeingDamaged)
		}
	}

	damageAnim.SetAudioCue(damageAudioCue)
	return damageAnim
}

// Validation for Player Commands
func (g *GameState) Throw() {
	equipment := g.Player.GetInventory()
	weapon, hasWeapon := equipment.GetMainHandItem()
	if !hasWeapon || !weapon.IsMissile() {
		g.msg(foundation.Msg("You have no suitable weapon equipped"))
		return
	}
	g.startThrowItem(weapon.(foundation.Item))
}

// Target Selection Stage
func (g *GameState) startThrowItem(item foundation.Item) {
	g.ui.SelectTarget(g.getThrownChanceToHitForUI, func(targetPos geometry.Point) {
		g.actorThrowItem(g.Player, item, g.Player.Position(), targetPos)
	})
}

// Logic And Animation
func (g *GameState) actorThrowItem(thrower *Actor, missile foundation.Item, origin, targetPos geometry.Point) {
	pathOfFlight := geometry.BresenhamLine(origin, targetPos, func(x, y int) bool {
		if origin.X == x && origin.Y == y {
			return true
		}
		return !g.IsSomethingBlockingTargetingAtLoc(geometry.Point{X: x, Y: y})
	})
	if len(pathOfFlight) > 1 {
		// remove start
		pathOfFlight = pathOfFlight[1:]
	}
	targetPos = pathOfFlight[len(pathOfFlight)-1]
	if !g.currentMap().IsTileWalkable(targetPos) && len(pathOfFlight) > 1 {
		targetPos = pathOfFlight[len(pathOfFlight)-2]
	}
	var onHitAnimations []foundation.Animation

	thrower.Inventory.RemoveItem(missile)

	if missile.IsBreakingNow() {
		missile.SetPosition(targetPos)
	} else {
		g.addItemToMap(missile, targetPos)
	}

	throwAnim, _ := g.ui.GetAnimThrow(missile, origin, targetPos)

	var attackMode AttackMode
	var damageType DamageType

	weapon, isWeapon := missile.(*Weapon)

	if isWeapon && weapon.GetCurrentAttackMode().IsThrow() {
		attackMode = weapon.GetCurrentAttackMode()
		damageType = weapon.GetDamageType()
	} else {
		attackMode = AttackMode{
			Mode:     TargetingModeThrow,
			TUCost:   thrower.TimeNeededForActions(),
			MaxRange: thrower.GetMaxThrowRange(),
			IsAimed:  false,
		}
		damageType = DamageTypeNormal
	}

	damage := SourcedDamage{
		NameOfThing:     "throw",
		Attacker:        thrower,
		IsObviousAttack: true,
		TargetingMode:   attackMode.Mode,
		DamageType:      damageType,
		DamageAmount:    missile.GetThrowDamage().Roll(),
		BodyPart:        d100.Body,
	}
	onHitAnimations = append(onHitAnimations, g.damageLocation(damage, targetPos)...)
	// explosion/fragmentation
	// fire
	// emp
	// plasma
	if missile.ZapEffect() != "" && !g.metronome.HasTimed(missile) {
		zapEffect := ZapEffectFromName(missile.ZapEffect())
		itemHitEffect := zapEffect(g, thrower, targetPos, missile.GetEffectParameters())
		onHitAnimations = append(onHitAnimations, itemHitEffect...)
	}

	if throwAnim != nil {
		throwAnim.SetFollowUp(onHitAnimations)
	}

	g.ui.AddAnimations([]foundation.Animation{throwAnim})

	if thrower == g.Player {
		g.endPlayerTurn(g.Player.TimeNeededForActions())
	}
}
