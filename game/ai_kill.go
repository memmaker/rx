package game

import (
	"contractor/d100"
	"contractor/foundation"
	"contractor/fsmai"
	"fmt"
)

type KillBehaviour struct {
	InitEvent fsmai.TransitionEvent
}

func (b KillBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
	return KillBehaviour{InitEvent: event}
}

func (b KillBehaviour) AssociatedState() fsmai.StateName { return fsmai.StateKill }

func (b KillBehaviour) Init(state *GameState, actor *Actor) {
	actorEvent := b.InitEvent.(ActorEvent)
	actor.SetGoal(GoalKillActor(actorEvent.Actor))
}

func (b KillBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {

	if actor.CanSee(b.InitEvent.(ActorEvent).Actor.Position()) {
		// Would have to remember last known position in order to pass it to
		// the target lost event..
	}
	// act on goals
	if actor.HasActiveGoal() {
		return actor.ActOnGoal(g)
	}
	actorEvent := b.InitEvent.(ActorEvent)

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

	if !a.GetInventory().HasRangedWeaponEquipped() {
		a.tryEquipRangedWeapon()
	}

	if !g.IsInShootingRange(a, target) { // ensure shooting range
		return moveTowardsActor(g, a, target, 1)
	}

	mainHandItem, hasMainHandItem := a.GetInventory().GetEquippedWeapon()

	if hasMainHandItem && mainHandItem.IsRangedWeapon() {
		isLoaded := mainHandItem.IsLoadedWeapon()
		canBeReloaded := a.GetInventory().HasAmmoWithCaliber(mainHandItem.GetCaliber())
		doesntNeedAmmo := !mainHandItem.NeedsAmmo()
		if !isLoaded && canBeReloaded { // reload
			g.actorReloadMainHandWeapon(a)
			return fsmai.NoEvent, a.TimeNeededForActions()
		}

		if mainHandItem.IsJammed() {
			mainHandItem.Unjam()
			g.msg(foundation.Msg(fmt.Sprintf("%s unjams %s", a.Name(), mainHandItem.Name())))
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

	if !a.GetInventory().HasMeleeWeaponEquipped() {
		a.tryEquipMeleeWeapon()
	}

	if distanceToTarget > 1 { // ensure melee range
		return moveTowardsActor(g, a, target, 1)
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
