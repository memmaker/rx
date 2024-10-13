package game

import (
	"RogueUI/d100"
	"RogueUI/fsmai"
)

func BehaviourKillInit(g *GameState, actor *Actor, event fsmai.TransitionEvent) {
	actorEvent := event.(ActorEvent)
	actor.SetGoal(GoalKillActor(actor, actorEvent.Actor))
}

func BehaviourKill(g *GameState, actor *Actor, event fsmai.TransitionEvent) (fsmai.TransitionEvent, int) {
	// act on goals
	if actor.HasActiveGoal() {
		return actor.ActOnGoal(g)
	}
	actorEvent := event.(ActorEvent)

	return NewTargetLostEvent(actorEvent.Actor), actor.TimeNeededForActions()
}

func tryKill(g *GameState, a *Actor, target *Actor) (fsmai.TransitionEvent, int) {
	distanceToTarget := g.currentMap().MoveDistance(a.Position(), target.Position())
	if !a.CanSee(target.Position()) { // ensure visibility, else -> target lost

		targetIsMuchFaster := float64(target.MovementSpeed()) > float64(a.MovementSpeed())*2
		fartherThanCanBeDeduced := distanceToTarget > (a.GetCharSheet().GetStat(d100.Intelligence) * 3)

		if fartherThanCanBeDeduced || targetIsMuchFaster {
			return NewTargetLostEvent(target), a.TimeNeededForMovement()
		}
	}

	if !g.IsInShootingRange(a, target) { // ensure shooting range
		return moveTowardsActor(g, a, target)
	}

	if !a.GetEquipment().HasRangedWeaponInMainHand() {
		a.tryEquipRangedWeapon()
	}
	mainHandItem, hasMainHandItem := a.GetEquipment().GetMainHandWeapon()

	if hasMainHandItem && mainHandItem.IsRangedWeapon() {
		isLoaded := mainHandItem.IsLoadedWeapon()
		canBeReloaded := a.GetInventory().HasAmmoWithCaliber(mainHandItem.GetCaliber())
		doesntNeedAmmo := !mainHandItem.NeedsAmmo()
		if !isLoaded && canBeReloaded { // reload
			g.actorReloadMainHandWeapon(a)
			return fsmai.NoEvent, a.TimeNeededForActions()
		}
		if isLoaded || doesntNeedAmmo { // ranged attack
			g.ui.AddAnimations(g.actorRangedAttack(a, mainHandItem, mainHandItem.GetCurrentAttackMode(), target, d100.Body, d100.NoCombatModifier))
			event := fsmai.TransitionEvent(fsmai.NoEvent)
			if !target.IsAlive() {
				event = NewTargetDiedEvent(target)
			}
			return event, mainHandItem.GetCurrentAttackMode().TUCost
		}
	}

	if !a.GetEquipment().HasMeleeWeaponEquipped() {
		a.tryEquipMeleeWeapon()
	}

	if distanceToTarget > 1 { // ensure melee range
		return moveTowardsActor(g, a, target)
	}

	// melee attack
	consequencesOfMonsterAttack := g.actorMeleeAttack(a, target, d100.Body, d100.NoCombatModifier)
	g.ui.AddAnimations(consequencesOfMonsterAttack)
	event := fsmai.TransitionEvent(fsmai.NoEvent)
	if !target.IsAlive() {
		event = NewTargetDiedEvent(target)
	}
	return event, a.GetMeleeTUCost()
}
