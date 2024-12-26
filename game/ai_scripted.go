package game

import (
	"contractor/foundation"
	"contractor/fsmai"
)

type ScriptedBehaviour struct {
	InitEvent fsmai.TransitionEvent
}

func (b ScriptedBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
	return ScriptedBehaviour{InitEvent: event}
}

func (b ScriptedBehaviour) AssociatedState() fsmai.StateName { return fsmai.StateScripted }

func (b ScriptedBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
	// barks
	if g.shouldActorBark(actor) && g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
		actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
	}

	switch b.InitEvent.(type) {
	case LocationEvent:
		locationEvent := b.InitEvent.(LocationEvent)
		return runTowards(g, actor, locationEvent.Location)
	}

	return fsmai.NoEvent, actor.maximalTimeNeededForActions()
}

func (b ScriptedBehaviour) Init(state *GameState, actor *Actor) {

}

func (b ScriptedBehaviour) IsCombatBehavior() bool {
	return false
}

func (b ScriptedBehaviour) IsHostilityTowards(other *Actor) bool {
	return false
}
