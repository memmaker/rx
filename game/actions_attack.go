package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
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
	mainHandItem, hasWeapon := g.Player.GetEquipment().GetMainHandWeapon()
	if hasWeapon && mainHandItem.GetCurrentAttackMode().IsThrow() {
		g.startThrowItem(mainHandItem)
		return
	}
	if hasWeapon && mainHandItem.IsMeleeWeapon() && !mainHandItem.IsRangedWeapon() {
		g.ui.SelectDirection(func(direction geometry.CompassDirection) {
			targetPos := g.Player.Position().Add(direction.ToPoint())
			g.playerMeleeAttackLocation(targetPos)
		})
		return
	}
	if !hasWeapon || !mainHandItem.IsRangedWeapon() {
		g.msg(foundation.Msg("You have no suitable weapon equipped"))
		return
	}
	weapon := mainHandItem
	if !weapon.HasAmmo() {
		g.ui.PlayCue(weapon.GetOutOfAmmoAudioCue())
		g.msg(foundation.Msg("You have no ammo"))
		return
	}
	attackMode := mainHandItem.GetCurrentAttackMode()
	if attackMode.IsAimed {
		g.ui.SelectBodyPart(g.playerLastAimedAt, func(victim foundation.ActorForUI, bodyPart special.BodyPart) {
			g.playerLastAimedAt = bodyPart
			target := victim.(*Actor)
			shotAnim := g.actorRangedAttack(g.Player, mainHandItem, attackMode, target, bodyPart)
			g.ui.AddAnimations(shotAnim)
			g.endPlayerTurn(attackMode.TUCost)
		})
	} else {
		g.ui.SelectTarget(func(targetPos geometry.Point) {
			if g.currentMap().IsActorAt(targetPos) {
				target := g.currentMap().ActorAt(targetPos)
				shotAnim := g.actorRangedAttack(g.Player, mainHandItem, attackMode, target, special.Body)
				g.ui.AddAnimations(shotAnim)
				g.endPlayerTurn(attackMode.TUCost)
			} else {
				shotAnim := g.actorRangedAttackLocation(g.Player, mainHandItem, attackMode, targetPos)
				g.ui.AddAnimations(shotAnim)
				g.endPlayerTurn(attackMode.TUCost)
			}
		})
	}

}

func (g *GameState) PlayerQuickRangedAttack() {
	enemies := g.playerVisibleActorsByDistance()
	equipment := g.Player.GetEquipment()
	mainHandItem, hasWeapon := equipment.GetMainHandWeapon()

	if !hasWeapon || !mainHandItem.IsRangedWeapon() {
		g.msg(foundation.Msg("You have no suitable weapon equipped"))
		return
	}
	weapon := mainHandItem
	if !weapon.HasAmmo() {
		g.msg(foundation.Msg("You have no ammo"))
		g.ui.PlayCue(weapon.GetOutOfAmmoAudioCue())
		return
	}
	if len(enemies) == 0 {
		g.msg(foundation.Msg("No enemies in sight"))
		return
	}

	mode := mainHandItem.GetCurrentAttackMode()
	shotAnim := g.actorRangedAttack(g.Player, mainHandItem, mode, enemies[0], special.Body)
	g.ui.AddAnimations(shotAnim)
	g.endPlayerTurn(mode.TUCost)
}

func (g *GameState) QuickThrow() {
	enemies := g.playerVisibleActorsByDistance()
	preselectedTarget := g.Player.Position()
	equipment := g.Player.GetEquipment()
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
	attackerLuckChance := special.Percentage(g.Player.GetCharSheet().GetDerivedStat(special.CriticalChance))
	defenderLuckChance := special.Percentage(defender.GetCharSheet().GetDerivedStat(special.CriticalChance))

	attackerStrength := special.Percentage(g.Player.GetCharSheet().GetStat(special.Strength) * 10)
	defenderStrength := special.Percentage(defender.GetCharSheet().GetStat(special.Strength) * 10)

	contestResult := special.SkillContest(attackerStrength, attackerLuckChance, defenderStrength, defenderLuckChance)

	if defender.IsSleeping() || contestResult == 0 {
		sourcedDamage := SourcedDamage{
			NameOfThing:     "drowning",
			Attacker:        g.Player,
			IsObviousAttack: true,
			TargetingMode:   special.TargetingModeFireSingle,
			DamageType:      special.DamageTypeNormal,
			DamageAmount:    defender.GetHitPointsMax(),
			BodyPart:        special.Body,
		}
		g.msg(foundation.HiLite("You drown %s", defender.Name()))
		g.ui.AddAnimations(g.damageActor(sourcedDamage, defender))
		g.endPlayerTurn(g.Player.timeNeededForMeleeAttack())
	} else {
		g.msg(foundation.HiLite("You fail to sneak up on %s", defender.Name()))
		g.playerMeleeAttack(defender)
	}
}

func (g *GameState) playerBackstab(defender *Actor) {
	attackerLuckChance := special.Percentage(g.Player.GetCharSheet().GetDerivedStat(special.CriticalChance))
	defenderLuckChance := special.Percentage(defender.GetCharSheet().GetDerivedStat(special.CriticalChance))

	attackerStealth := special.Percentage(g.Player.GetCharSheet().GetSkill(special.Stealth))
	defenderAwareness := special.Percentage(defender.GetCharSheet().GetStat(special.Perception) * 10)

	contestResult := special.SkillContest(attackerStealth, attackerLuckChance, defenderAwareness, defenderLuckChance)

	if defender.IsSleeping() || contestResult == 0 {
		sourcedDamage := SourcedDamage{
			NameOfThing:     "backstab",
			Attacker:        g.Player,
			IsObviousAttack: true,
			TargetingMode:   special.TargetingModeFireSingle,
			DamageType:      special.DamageTypeNormal,
			DamageAmount:    defender.GetHitPointsMax(),
			BodyPart:        special.Body,
		}
		g.msg(foundation.HiLite("You stab %s in the back", defender.Name()))
		g.ui.AddAnimations(g.damageActor(sourcedDamage, defender))
		g.endPlayerTurn(g.Player.timeNeededForMeleeAttack())
	} else {
		g.msg(foundation.HiLite("You fail to sneak up on %s", defender.Name()))
		g.playerMeleeAttack(defender)
	}
}

func (g *GameState) playerNonLethalTakedown(victim *Actor) {
	attackerLuckChance := special.Percentage(g.Player.GetCharSheet().GetDerivedStat(special.CriticalChance))
	defenderLuckChance := special.Percentage(victim.GetCharSheet().GetDerivedStat(special.CriticalChance))

	attackerStealth := special.Percentage(g.Player.GetCharSheet().GetStat(special.Strength) * 10)
	defenderAwareness := special.Percentage(victim.GetCharSheet().GetStat(special.Strength) * 10)

	contestResult := special.SkillContest(attackerStealth, attackerLuckChance, defenderAwareness, defenderLuckChance)

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
		attackAudioCue, damageWithSource := g.getMeleeDamage(g.Player, 100, special.Body)
		if weapon, hasMeleeWeapon := g.Player.GetEquipment().GetMeleeWeapon(); hasMeleeWeapon {
			weapon.Degrade(1)
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
	doMeleeAttack := func(part special.BodyPart) {
		consequences := g.actorMeleeAttack(g.Player, defender, part)
		if !g.Player.HasFlag(foundation.FlagInvisible) {
			defender.GetFlags().Set(foundation.FlagAwareOfPlayer)
		}
		g.ui.AddAnimations(consequences)
		g.endPlayerTurn(g.Player.timeNeededForMeleeAttack())
	}

	mainhandItem, hasWeapon := g.Player.GetEquipment().GetMainHandWeapon()
	if hasWeapon && mainhandItem.IsMeleeWeapon() && mainhandItem.GetCurrentAttackMode().IsAimed {
		g.ui.OpenAimedShotPicker(defender, g.playerLastAimedAt, func(victim foundation.ActorForUI, bodyPart special.BodyPart) {
			doMeleeAttack(bodyPart)
		})
	} else {
		doMeleeAttack(special.Body)
	}
}

func (g *GameState) actorMeleeAttack(attacker *Actor, defender *Actor, part special.BodyPart) []foundation.Animation {
	if !defender.IsAlive() {
		return nil
	}
	var afterAttackAnimations []foundation.Animation

	mainHandItem, hasMeleeWeapon := attacker.GetEquipment().GetMeleeWeapon()

	chanceToHit := g.getMeleeChanceToHit(attacker, mainHandItem, defender)
	chanceToHit += part.AimPenalty() / 2

	if hasMeleeWeapon {
		mainHandItem.Degrade(0.1)
	}

	attackAudioCue, damageWithSource := g.getMeleeDamage(attacker, chanceToHit, part)

	animAttackerIndicator := g.ui.GetAnimBackgroundColor(attacker.Position(), "dark_gray_6", 4, nil)
	animAttackerIndicator.SetAudioCue(attackAudioCue)

	afterAttackAnimations = append(afterAttackAnimations, animAttackerIndicator)

	damageAnims := g.applyDamageToActorAnimated(attacker, mainHandItem, damageWithSource, defender)
	afterAttackAnimations = append(afterAttackAnimations, damageAnims...)

	attacker.GetFlags().Unset(foundation.FlagConcentratedAiming)

	return afterAttackAnimations
}

func (g *GameState) getMeleeChanceToHit(attacker *Actor, weaponItem *Weapon, defender *Actor) int {
	attackerSkill := special.MeleeCombat
	if weaponItem != nil && weaponItem.IsMeleeWeapon() {
		attackerSkill = weaponItem.GetSkillUsed()
	}
	chanceToHit := special.MeleeChanceToHit(attacker.GetCharSheet(), attackerSkill, defender.GetCharSheet())
	return chanceToHit
}

func (g *GameState) getMeleeDamage(attacker *Actor, cth int, part special.BodyPart) (string, SourcedDamage) {
	targetingMode := special.TargetingModePunch
	if rand.Intn(100) < 50 {
		targetingMode = special.TargetingModeKick
	}
	damageType := special.DamageTypeNormal
	meleeDamageBonus := attacker.GetMeleeDamageBonus()
	attackAudioCue := attacker.GetMeleeAudioCue(false)
	damage := 3 + meleeDamageBonus

	if targetingMode == special.TargetingModeKick {
		damage = 5 + meleeDamageBonus
		attackAudioCue = attacker.GetMeleeAudioCue(true)
	}

	itemInHand, hasItem := attacker.GetEquipment().GetMeleeWeapon()

	if hasItem && itemInHand.IsMeleeWeapon() {
		weapon := itemInHand
		damage = meleeDamageBonus + itemInHand.GetWeaponDamage().Roll()
		damageType = weapon.GetDamageType()
		attackAudioCue = weapon.GetFireAudioCue(special.TargetingModeFireSingle)
	}

	isHit := rand.Intn(100) < cth
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
	return attackAudioCue, damageWithSource
}

// actorRangedAttack logic and animation of a ranged attack with the equipped weapon
func (g *GameState) actorRangedAttack(attacker *Actor, weaponItem *Weapon, attackMode AttackMode, defender *Actor, bodyPart special.BodyPart) []foundation.Animation {
	if !defender.IsAlive() {
		return nil
	}

	bulletsSpent, weapon := g.removeBulletsFromWeapon(weaponItem, attackMode)

	attackAnimations, isProjectileAnimation := g.getWeaponAttackAnim(attacker, defender.Position(), weaponItem, attackMode, bulletsSpent)

	baseChanceToHit := g.getRangedChanceToHit(attacker, weaponItem, defender)
	chanceToHit := baseChanceToHit + bodyPart.AimPenalty()

	damageWithSource := g.calculateRangedDamage(attacker, weaponItem, attackMode, bulletsSpent, chanceToHit, defender, bodyPart)

	weaponEffectParams := foundation.Params{
		"damage": damageWithSource.DamageAmount,
	}
	var hitAnimations []foundation.Animation
	if weapon.GetDamageType() == special.DamageTypeExplosive {
		hitAnimations = explosion(g, attacker, defender.Position(), weaponEffectParams)
	} else if weapon.GetDamageType() == special.DamageTypeFire && attackMode.Mode == special.TargetingModeFlame {
		hitAnimations = fireBreath(g, attacker, defender.Position(), weaponEffectParams)
	} else {
		hitAnimations = g.applyDamageToActorAnimated(attacker, weaponItem, damageWithSource, defender)
	}

	attacker.GetFlags().Unset(foundation.FlagConcentratedAiming)

	if isProjectileAnimation {
		attackAnimations.SetFollowUp(hitAnimations)
		return []foundation.Animation{attackAnimations}
	} else {
		return append(hitAnimations, attackAnimations)
	}
}

func (g *GameState) actorRangedAttackLocation(attacker *Actor, weaponItem *Weapon, attackMode AttackMode, targetPos geometry.Point) []foundation.Animation {

	bulletsSpent, weapon := g.removeBulletsFromWeapon(weaponItem, attackMode)

	onAttackAnims, isProjectileAnimation := g.getWeaponAttackAnim(attacker, targetPos, weaponItem, attackMode, bulletsSpent)

	chanceToHit := 100

	damageWithSource := g.calculateRangedDamage(attacker, weaponItem, attackMode, bulletsSpent, chanceToHit, nil, special.Body)
	weaponEffectParams := weaponItem.GetEffectParameters()
	var consequenceOfHit []foundation.Animation
	if weapon.GetDamageType() == special.DamageTypeExplosive {
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

	if isProjectileAnimation {
		onAttackAnims.SetFollowUp(consequenceOfHit)
		return []foundation.Animation{onAttackAnims}
	} else {
		return append(consequenceOfHit, onAttackAnims)
	}
}

func (g *GameState) calculateRangedDamage(attacker *Actor, weaponItem *Weapon, attackMode AttackMode, bulletsSpent int, chanceToHit int, victim *Actor, bodyPart special.BodyPart) SourcedDamage {
	weapon := weaponItem
	damage := weaponItem.GetWeaponDamage()
	totalDamage := 0
	damagePerBullet := make([]int, bulletsSpent)
	for i := 0; i < bulletsSpent; i++ {
		damageDone := damage.Roll()
		if rand.Intn(100)+1 >= chanceToHit {
			damageDone = 0
		}
		damagePerBullet[i] = damageDone
		totalDamage += damageDone
	}

	damageFactor := 1.0
	bonusDamage := 0

	if weapon.NeedsAmmo() && totalDamage > 0 {
		ammoItem := weapon.GetLoadedAmmo()
		ammo := ammoItem
		if victim != nil {
			for tags, dmgBonus := range ammo.BonusDamageAgainstActorWithTags {
				if victim.HasFlag(tags) {
					bonusDamage += dmgBonus
				}
			}
		}

		damageFactor = ammo.DamageFactor
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

func (g *GameState) removeBulletsFromWeapon(weaponItem *Weapon, attackMode AttackMode) (int, *Weapon) {
	bulletsSpent := 1
	weapon := weaponItem
	if attackMode.Mode == special.TargetingModeFireBurst {
		bulletsSpent = min(weapon.GetLoadedBullets(), weapon.GetBurstRounds())
	} else if attackMode.Mode == special.TargetingModeFireFullAuto {
		bulletsSpent = weapon.GetLoadedBullets()
	}

	weapon.RemoveBullets(bulletsSpent)

	// apply weapon degradation by shooting
	baseDegradePerBullet := 0.1

	factorFromWeaponAndAmmo := weaponItem.GetDegradationFactorOfAttack()

	totalDegrade := baseDegradePerBullet * float64(min(5, bulletsSpent)) * factorFromWeaponAndAmmo

	weaponItem.Degrade(totalDegrade)

	return bulletsSpent, weapon
}

func (g *GameState) applyDamageToActorAnimated(attacker *Actor, weaponItem *Weapon, damageWithSource SourcedDamage, defender *Actor) []foundation.Animation {
	var damageAnims []foundation.Animation

	dtMod := 0
	if weaponItem != nil {
		weapon := weaponItem
		dtMod = weapon.GetTargetDTModifier()
	}

	damageWithSource = defender.ModifyDamageByArmor(damageWithSource, dtMod)

	attackedFlag := fmt.Sprintf("WasAttacked(%s)", defender.GetInternalName())
	g.gameFlags.SetFlag(attackedFlag)

	if attacker == g.Player {
		attackedByPlayer := fmt.Sprintf("WasAttackedByPlayer(%s)", defender.GetInternalName())
		g.gameFlags.SetFlag(attackedByPlayer)
	}

	if damageWithSource.DamageAmount > 0 {
		if weaponItem != nil && weaponItem.IsZappable() {
			weaponZapEffect := ZapEffectFromName(weaponItem.ZapEffect())
			damageAnims = weaponZapEffect(g, attacker, defender.Position(), weaponItem.GetEffectParameters())
		} else {
			damageAnims = g.damageActor(damageWithSource, defender)
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

// Validation for Player Commands
func (g *GameState) Throw() {
	equipment := g.Player.GetEquipment()
	weapon, hasWeapon := equipment.GetMainHandItem()
	if !hasWeapon || !weapon.IsMissile() {
		g.msg(foundation.Msg("You have no suitable weapon equipped"))
		return
	}
	g.startThrowItem(weapon.(foundation.Item))
}

// Target Selection Stage
func (g *GameState) startThrowItem(item foundation.Item) {
	g.ui.SelectTarget(func(targetPos geometry.Point) {
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

	g.removeItemFromInventory(thrower, missile)

	if missile.IsBreakingNow() {
		missile.SetPosition(targetPos)
	} else {
		g.addItemToMap(missile, targetPos)
	}

	throwAnim, _ := g.ui.GetAnimThrow(missile, origin, targetPos)

	var attackMode AttackMode
	var damageType special.DamageType

	weapon, isWeapon := missile.(*Weapon)

	if isWeapon && weapon.GetCurrentAttackMode().IsThrow() {
		attackMode = weapon.GetCurrentAttackMode()
		damageType = weapon.GetDamageType()
	} else {
		attackMode = AttackMode{
			Mode:     special.TargetingModeThrow,
			TUCost:   thrower.timeNeededForActions(),
			MaxRange: thrower.GetMaxThrowRange(),
			IsAimed:  false,
		}
		damageType = special.DamageTypeNormal
	}

	damage := SourcedDamage{
		NameOfThing:     "throw",
		Attacker:        thrower,
		IsObviousAttack: true,
		TargetingMode:   attackMode.Mode,
		DamageType:      damageType,
		DamageAmount:    missile.GetThrowDamage().Roll(),
		BodyPart:        special.Body,
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
		g.endPlayerTurn(g.Player.timeNeededForActions())
	}
}
