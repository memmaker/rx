package game

import (
	"contractor/foundation"
	"contractor/fsmai"
)

type IdleBehaviour struct {
	InitEvent fsmai.TransitionEvent
}

func (b IdleBehaviour) IsCombatBehavior() bool {
	return false
}

func (b IdleBehaviour) IsHostilityTowards(other *Actor) bool {
	return false
}

func (b IdleBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
	return IdleBehaviour{InitEvent: event}
}

func (b IdleBehaviour) AssociatedState() fsmai.StateName { return fsmai.StateIdle }

func (b IdleBehaviour) Init(state *GameState, actor *Actor) {

}

func (b IdleBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
	// barks
	if g.shouldActorBark(actor) && g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
		actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
	}

	// random animal movement
	if actor.HasFlag(foundation.FlagAnimal) {
		consequencesOfConfusion := g.actConfused(actor)
		if len(consequencesOfConfusion) > 0 {
			g.ui.AddAnimations(consequencesOfConfusion)
			return fsmai.NoEvent, actor.maximalTimeNeededForActions()
		}
	}

	// run schedule
	if actor.Schedule != nil {
		return b.actOnTimeSlot(g, actor, actor.Schedule.CurrentTimeSlot())
	}

	// just go back to spawn
	return runTowards(g, actor, actor.SpawnPosition)
}

func (b IdleBehaviour) actOnTimeSlot(g *GameState, actor *Actor, slot TimeSlot) (fsmai.TransitionEvent, int) {
	if slot.MapName != g.currentMapName {
		transitionLoc := g.currentMap().GetNamedLocation(slot.UseTransition)
		if actor.Position() == transitionLoc {
			actor.SetFlag(foundation.FlagWantsToTransition)
			return fsmai.NoEvent, actor.RawTimeEnergy
		} else {
			return walkTowards(g, actor, transitionLoc)
		}
	}

	loc := g.currentMap().GetNamedLocation(slot.Location)
	if actor.Position() != loc {
		// walk to location
		return walkTowards(g, actor, loc)
	}

	// we are at the scheduled location
	return fsmai.NoEvent, actor.RawTimeEnergy
}
