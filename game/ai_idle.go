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
		return g.actOnTimeSlot(actor, actor.Schedule.CurrentTimeSlot())
	}

	// just go back to spawn
	spawnPos := MapPosition{MapName: actor.SpawnMapName, Position: actor.SpawnPosition}

	return g.actorTakeStepToMapPosition(actor, spawnPos, false)
}
