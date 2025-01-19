package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
)

type KillBehaviour struct {
	InitEvent TransitionEvent // state that needs saving..
}

func (b KillBehaviour) IsCombatBehavior() bool {
	return true
}

func (b KillBehaviour) IsHostilityTowards(other *Actor) bool {
	actorEvent := b.InitEvent.(ActorEvent)
	return other == actorEvent.Actor
}

func (b KillBehaviour) WithInitEvent(event TransitionEvent) ActorBehavior {
	return KillBehaviour{InitEvent: event}
}

func (b KillBehaviour) AssociatedState() StateName { return StateKill }

func (b KillBehaviour) Init(g *GameState, actor *Actor) {
	g.ui.AddAnimations(OneAnimation(g.ui.GetAnimEnterCombat(actor, nil)))
}

func (b KillBehaviour) Execute(g *GameState, actor *Actor) (TransitionEvent, int) {
	actorEvent := b.InitEvent.(ActorEvent)
	victim := actorEvent.Actor

	if !victim.IsAlive() {
		return NewTargetDiedEvent(victim), actor.TimeNeededForActions()
	}

	if victim.HasFlag(foundation.FlagActiveCamouflage) && !g.isDetectedByObserver(victim, actor) {
		g.tryAddRandomChatter(actor, foundation.ChatterTargetLost)
		return NewTargetLostEvent(victim), actor.TimeNeededForMovement()
	}

	distanceToTarget := g.currentMap().MoveDistance(actor.Position(), victim.Position())
	if !actor.CanSee(victim.Position()) { // ensure visibility, else -> target lost

		targetIsMuchFaster := float64(victim.MovementSpeed()) > float64(actor.MovementSpeed())*2
		fartherThanCanBeDeduced := distanceToTarget > (actor.GetCharSheet().GetStat(d100.Intelligence) * 3)

		if fartherThanCanBeDeduced || targetIsMuchFaster {
			return NewTargetLostEvent(victim), actor.TimeNeededForMovement()
		}
	}

	if !actor.GetInventory().HasRangedWeaponEquipped() {
		actor.tryEquipRangedWeapon()
	}

	if !g.IsInShootingRange(actor, victim) { // ensure shooting range
		return g.actorTakeStepTowardsOther(actor, victim, 1)
	}

	mainHandItem, hasMainHandItem := actor.GetInventory().GetEquippedWeapon()

	if hasMainHandItem && mainHandItem.IsRangedWeapon() {
		isLoaded := mainHandItem.IsLoadedWeapon()
		canBeReloaded := actor.GetInventory().HasAmmoWithCaliber(mainHandItem.GetCaliber())
		doesntNeedAmmo := !mainHandItem.NeedsAmmo()
		if !isLoaded && canBeReloaded { // reload
			g.actorReloadMainHandWeapon(actor)
			return NoEvent, actor.TimeNeededForActions()
		}

		if mainHandItem.IsJammed() {
			mainHandItem.Unjam()
			g.msg(foundation.Msg(fmt.Sprintf("%s unjams %s", actor.Name(), mainHandItem.Name())))
			return NoEvent, actor.TimeNeededForActions()
		}

		if isLoaded || doesntNeedAmmo { // ranged attack
			g.ui.AddAnimations(g.actorRangedAttack(actor, mainHandItem, mainHandItem.GetCurrentAttackMode(), victim, d100.Body, d100.NoCombatModifier))
			event := TransitionEvent(NoEvent)
			if !victim.IsAlive() {
				event = NewTargetDiedEvent(victim)
			}
			return event, mainHandItem.GetCurrentAttackMode().TUCost
		}
	}

	if !actor.GetInventory().HasMeleeWeaponEquipped() {
		actor.tryEquipMeleeWeapon()
	}

	if distanceToTarget > 1 { // ensure melee range
		return g.actorTakeStepTowardsOther(actor, victim, 1)
	}

	// melee attack
	consequencesOfMonsterAttack := g.actorMeleeAttack(actor, victim, d100.Body, d100.NoCombatModifier)
	g.ui.AddAnimations(consequencesOfMonsterAttack)
	event := TransitionEvent(NoEvent)

	return event, actor.GetMeleeTUCost()
}
