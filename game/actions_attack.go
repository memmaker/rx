package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
	"github.com/memmaker/go/fxtools"
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
	mainHandItem, hasWeapon := g.Player.GetEquipment().GetMainHandItem()
	if !hasWeapon || !mainHandItem.IsRangedWeapon() {
		g.msg(foundation.Msg("You have no suitable weapon equipped"))
		return
	}
	weapon := mainHandItem.GetWeapon()
	if !weapon.HasAmmo() {
		g.ui.PlayCue(weapon.GetOutOfAmmoAudioCue())
		g.msg(foundation.Msg("You have no ammo"))
		return
	}
	attackMode := mainHandItem.GetCurrentAttackMode()
	if attackMode.IsAimed {
		g.ui.SelectBodyPart(func(victim foundation.ActorForUI, bodyPart int) {
			g.msg(foundation.HiLite("You aim at %s's %s", victim.Name(), victim.GetBodyPartByIndex(bodyPart)))
			target := victim.(*Actor)
			shotAnim := g.actorRangedAttack(g.Player, mainHandItem, attackMode.Mode, target, bodyPart)
			g.ui.AddAnimations(shotAnim)
			g.endPlayerTurn(attackMode.TUCost)
		})
	} else {
		g.ui.SelectTarget(func(targetPos geometry.Point, bodyPart int) {
			if g.currentMap().IsActorAt(targetPos) {
				target := g.currentMap().ActorAt(targetPos)
				shotAnim := g.actorRangedAttack(g.Player, mainHandItem, attackMode.Mode, target, bodyPart)
				g.ui.AddAnimations(shotAnim)
				g.endPlayerTurn(attackMode.TUCost)
			} else {
				shotAnim := g.actorRangedAttackLocation(g.Player, mainHandItem, attackMode.Mode, targetPos)
				g.ui.AddAnimations(shotAnim)
				g.endPlayerTurn(attackMode.TUCost)
			}
		})
	}

}

func (g *GameState) PlayerQuickRangedAttack() {
	enemies := g.playerVisibleEnemiesByDistance()
	equipment := g.Player.GetEquipment()
	mainHandItem, hasWeapon := equipment.GetMainHandItem()

	if !hasWeapon || !mainHandItem.IsRangedWeapon() {
		g.msg(foundation.Msg("You have no suitable weapon equipped"))
		return
	}
	weapon := mainHandItem.GetWeapon()
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
	shotAnim := g.actorRangedAttack(g.Player, mainHandItem, mode.Mode, enemies[0], 2)
	g.ui.AddAnimations(shotAnim)
	g.endPlayerTurn(mode.TUCost)
}

func (g *GameState) QuickThrow() {
	enemies := g.playerVisibleEnemiesByDistance()
	preselectedTarget := g.Player.Position()
	equipment := g.Player.GetEquipment()
	weapon, hasWeapon := equipment.GetMainHandItem()
	if hasWeapon || !weapon.IsMissile() {
		g.msg(foundation.Msg("You have no suitable weapon equipped"))
		return
	}
	if len(enemies) == 0 {
		g.msg(foundation.Msg("No enemies in sight"))
		return
	}
	preselectedTarget = enemies[0].Position()
	g.actorThrowItem(g.Player, weapon, g.Player.Position(), preselectedTarget)
}

func (g *GameState) playerBackstab(defender *Actor) {
	attackerLuckChance := special.Percentage(g.Player.GetCharSheet().GetStat(special.Luck))
	defenderLuckChance := special.Percentage(defender.GetCharSheet().GetStat(special.Luck))

	attackerStealth := special.Percentage(g.Player.GetCharSheet().GetSkill(special.Sneak))
	defenderAwareness := special.Percentage(defender.GetCharSheet().GetStat(special.Perception) * 10)

	contestResult := special.SkillContest(attackerStealth, attackerLuckChance, defenderAwareness, defenderLuckChance)

	if defender.IsSleeping() || contestResult == 0 {
		sourcedDamage := SourcedDamage{
			NameOfThing:     "backstab",
			Attacker:        g.Player,
			IsObviousAttack: true,
			AttackMode:      special.TargetingModeFireSingle,
			DamageType:      special.DamageTypeNormal,
			DamageAmount:    defender.GetHitPointsMax(),
		}
		g.msg(foundation.HiLite("You stab %s in the back", defender.Name()))
		g.ui.AddAnimations(g.damageActor(sourcedDamage, defender))
		g.endPlayerTurn(10)
	} else {
		g.msg(foundation.HiLite("You fail to sneak up on %s", defender.Name()))
		g.playerMeleeAttack(defender)
	}
}

func (g *GameState) playerNonLethalTakedown(victim *Actor) {
	attackerLuckChance := special.Percentage(g.Player.GetCharSheet().GetStat(special.Luck))
	defenderLuckChance := special.Percentage(victim.GetCharSheet().GetStat(special.Luck))

	attackerStealth := special.Percentage(g.Player.GetCharSheet().GetStat(special.Strength) * 10)
	defenderAwareness := special.Percentage(victim.GetCharSheet().GetStat(special.Strength) * 10)

	contestResult := special.SkillContest(attackerStealth, attackerLuckChance, defenderAwareness, defenderLuckChance)

	if contestResult == 0 {
		victim.SetSleeping()
		g.msg(foundation.HiLite("You knock out %s", victim.Name()))
		g.endPlayerTurn(10)
	} else {
		g.msg(foundation.HiLite("%s is able to resist your attempt", victim.Name()))
		g.playerMeleeAttack(victim)
	}
}

func (g *GameState) playerMeleeAttack(defender *Actor) {
	consequences := g.actorMeleeAttack(g.Player, defender)
	if !g.Player.HasFlag(foundation.FlagInvisible) {
		defender.GetFlags().Set(foundation.FlagAwareOfPlayer)
	}
	g.ui.AddAnimations(consequences)
	g.endPlayerTurn(10)
}

func (g *GameState) actorMeleeAttack(attacker *Actor, defender *Actor) []foundation.Animation {
	if !defender.IsAlive() {
		return nil
	}
	var afterAttackAnimations []foundation.Animation

	attackerSkill := special.Unarmed
	targetingMode := special.TargetingModePunch
	damageType := special.DamageTypeNormal
	damage := fxtools.NewInterval(1, 4)
	attackAudioCue := attacker.GetMeleeAudioCue(false)

	itemInHand, hasItem := attacker.GetEquipment().GetMeleeWeapon()

	if hasItem && itemInHand.IsMeleeWeapon() {
		weapon := itemInHand.GetWeapon()
		attackerSkill = special.MeleeWeapons
		damage = weapon.GetDamage()
		damageType = weapon.GetDamageType()
		attackAudioCue = weapon.GetFireAudioCue(special.TargetingModeFireSingle)
	}

	chanceToHit := special.MeleeChanceToHit(attacker.GetCharSheet(), attackerSkill, defender.GetCharSheet(), special.Body)

	animAttackerIndicator := g.ui.GetAnimBackgroundColor(attacker.Position(), "dark_gray_6", 4, nil)
	animAttackerIndicator.SetAudioCue(attackAudioCue)

	afterAttackAnimations = append(afterAttackAnimations, animAttackerIndicator)

	damageWithSource := SourcedDamage{
		NameOfThing:     "melee",
		Attacker:        attacker,
		IsObviousAttack: true,
		AttackMode:      targetingMode,
		DamageType:      damageType,
		DamageAmount:    damage.Roll(),
	}

	damageWithSource = defender.ModifyDamageByArmor(damageWithSource, 0, 0, 0)

	if rand.Intn(100) < chanceToHit && damageWithSource.DamageAmount > 0 {
		damageAnims := g.damageActor(damageWithSource, defender)
		afterAttackAnimations = append(afterAttackAnimations, damageAnims...)
	} else {
		evade := g.ui.GetAnimEvade(defender, nil)
		evade.SetAudioCue(defender.GetDodgedAudioCue())
		afterAttackAnimations = append(afterAttackAnimations, evade)
	}

	if damageWithSource.IsObviousAttack {
		g.trySetHostile(defender, damageWithSource.Attacker)
	}

	return afterAttackAnimations
}

// actorRangedAttack logic and animation of a ranged attack with the equipped weapon
func (g *GameState) actorRangedAttack(attacker *Actor, weaponItem *Item, mode special.TargetingMode, defender *Actor, bodyPart int) []foundation.Animation {
	if !defender.IsAlive() {
		return nil
	}

	bulletsSpent := 1
	weapon := weaponItem.GetWeapon()
	if mode == special.TargetingModeFireBurst {
		bulletsSpent = min(weapon.GetLoadedBullets(), weapon.GetBurstRounds())
	} else if mode == special.TargetingModeFireFullAuto {
		bulletsSpent = weapon.GetLoadedBullets()
	}

	weapon.RemoveBullets(bulletsSpent)

	animAttackerIndicator := g.ui.GetAnimMuzzleFlash(attacker.Position(), fxtools.NewColorFromRGBA(g.palette.Get("White")).MultiplyWithScalar(5), 4, bulletsSpent, nil)
	animAttackerIndicator.SetAudioCue(weapon.GetFireAudioCue(mode))

	var onAttackAnims []foundation.Animation
	onAttackAnims = append(onAttackAnims, animAttackerIndicator)
	chanceToHit := g.getRangedChanceToHit(attacker, weaponItem, defender)
	damage := weapon.GetDamage()
	var damageAnims []foundation.Animation
	totalDamage := 0
	for i := 0; i < bulletsSpent; i++ {
		damageDone := damage.Roll()

		if rand.Intn(100) < chanceToHit {
			totalDamage += damageDone
		}
	}

	drModifier := 0
	damageFactor := 1.0
	if weapon.NeedsAmmo() {
		ammoItem := weapon.GetAmmo()
		ammo := ammoItem.GetAmmo()
		drModifier = ammo.DRModifier
		damageFactor = float64(ammo.DamageMultiplier) / float64(ammo.DamageDivisor)
	}

	damageWithSource := SourcedDamage{
		NameOfThing:     "",
		Attacker:        attacker,
		IsObviousAttack: true,
		AttackMode:      mode,
		DamageType:      weapon.GetDamageType(),
		DamageAmount:    int(float64(totalDamage) * damageFactor),
	}
	damageWithSource = defender.ModifyDamageByArmor(damageWithSource, drModifier, 0, bodyPart)

	if damageWithSource.DamageAmount > 0 {
		if weaponItem.IsZappable() {
			weaponZapEffect := ZapEffectFromName(weaponItem.GetZapEffectName())
			damageAnims = weaponZapEffect(g, attacker, defender.Position())
		} else {
			damageAnims = g.damageActor(damageWithSource, defender)
		}
	} else {
		if damageWithSource.IsObviousAttack {
			g.trySetHostile(defender, damageWithSource.Attacker)
		}
		evade := g.ui.GetAnimEvade(defender, func() {
			g.ui.PlayCue(weapon.GetMissAudioCue())
		})
		evade.SetAudioCue(defender.GetDodgedAudioCue())
		damageAnims = []foundation.Animation{evade}
	}
	return append(onAttackAnims, damageAnims...)
}

func (g *GameState) actorRangedAttackLocation(attacker *Actor, weaponItem *Item, mode special.TargetingMode, loc geometry.Point) []foundation.Animation {
	bulletsSpent := 1
	weapon := weaponItem.GetWeapon()
	if mode == special.TargetingModeFireBurst {
		bulletsSpent = min(weapon.GetLoadedBullets(), weapon.GetBurstRounds())
	} else if mode == special.TargetingModeFireFullAuto {
		bulletsSpent = weapon.GetLoadedBullets()
	}

	weapon.RemoveBullets(bulletsSpent)

	animAttackerIndicator := g.ui.GetAnimMuzzleFlash(attacker.Position(), fxtools.NewColorFromRGBA(g.palette.Get("White")).MultiplyWithScalar(5), 4, bulletsSpent, nil)
	animAttackerIndicator.SetAudioCue(weapon.GetFireAudioCue(mode))

	var onAttackAnims []foundation.Animation
	onAttackAnims = append(onAttackAnims, animAttackerIndicator)

	damage := weapon.GetDamage()
	var damageAnims []foundation.Animation
	totalDamage := 0
	for i := 0; i < bulletsSpent; i++ {
		damageDone := damage.Roll()
		totalDamage += damageDone
	}
	damageWithSource := SourcedDamage{
		NameOfThing:     "",
		Attacker:        attacker,
		IsObviousAttack: true,
		AttackMode:      mode,
		DamageType:      weapon.GetDamageType(),
		DamageAmount:    totalDamage,
	}
	if totalDamage > 0 {
		if weaponItem.IsZappable() {
			weaponZapEffect := ZapEffectFromName(weaponItem.GetZapEffectName())
			damageAnims = weaponZapEffect(g, attacker, loc)
		} else {
			damageAnims = g.damageLocation(damageWithSource, loc)
		}
	}
	return append(onAttackAnims, damageAnims...)
}

// Validation for Player Commands
func (g *GameState) Throw() {
	equipment := g.Player.GetEquipment()
	weapon, hasWeapon := equipment.GetMainHandItem()
	if !hasWeapon || !weapon.IsMissile() {
		g.msg(foundation.Msg("You have no suitable weapon equipped"))
		return
	}
	g.startThrowItem(weapon)
}

// Target Selection Stage
func (g *GameState) startThrowItem(item *Item) {
	g.ui.SelectTarget(func(targetPos geometry.Point, hitZone int) {
		g.actorThrowItem(g.Player, item, g.Player.Position(), targetPos)
	})
}

// Logic And Animation
func (g *GameState) actorThrowItem(thrower *Actor, missile *Item, origin, targetPos geometry.Point) {
	pathOfFlight := geometry.BresenhamLine(origin, targetPos, func(x, y int) bool {
		if origin.X == x && origin.Y == y {
			return true
		}
		return g.currentMap().IsCurrentlyPassable(geometry.Point{X: x, Y: y})
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

	if !missile.IsBreakingNow() {
		g.addItemToMap(missile, targetPos)
	}

	throwAnim, _ := g.ui.GetAnimThrow(missile, origin, targetPos)

	if g.currentMap().IsActorAt(targetPos) {
		defender := g.currentMap().ActorAt(targetPos)
		//isLaunch := thrower.IsLaunching(missile) // otherwise it's a throw
		consequenceOfActorHit := g.actorRangedAttack(thrower, missile, special.TargetingModeFireSingle, defender, 0)
		onHitAnimations = append(onHitAnimations, consequenceOfActorHit...)
	} else if g.currentMap().IsObjectAt(targetPos) {
		object := g.currentMap().ObjectAt(targetPos)
		consequenceOfObjectHit := object.OnDamage(SourcedDamage{
			NameOfThing:     "",
			Attacker:        thrower,
			IsObviousAttack: true,
			AttackMode:      special.TargetingModeThrow,
			DamageType:      special.DamageTypeNormal,
			DamageAmount:    missile.GetThrowDamage().Roll(),
		})
		onHitAnimations = append(onHitAnimations, consequenceOfObjectHit...)
	}
	// explosion/fragmentation
	// fire
	// emp
	// plasma
	if missile.GetZapEffectName() != "" {
		zapEffect := ZapEffectFromName(missile.GetZapEffectName())
		itemHitEffect := zapEffect(g, thrower, targetPos)
		onHitAnimations = append(onHitAnimations, itemHitEffect...)
	}

	if throwAnim != nil {
		throwAnim.SetFollowUp(onHitAnimations)
	}

	g.ui.AddAnimations([]foundation.Animation{throwAnim})

	if thrower == g.Player {
		g.endPlayerTurn(10)
	}
}
