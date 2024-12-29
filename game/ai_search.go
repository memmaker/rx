package game

import (
	"contractor/foundation"
	"contractor/fsmai"
	"github.com/memmaker/go/geometry"
)

type SearchBehaviour struct {
	InitEvent         fsmai.TransitionEvent
	LastKnownPosition geometry.Point
}

func (b SearchBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
	return SearchBehaviour{InitEvent: event, LastKnownPosition: event.(ActorEvent).Actor.Position()}
}

func (b SearchBehaviour) AssociatedState() fsmai.StateName { return fsmai.StateSearch }

func (b SearchBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
	actorEvent := b.InitEvent.(ActorEvent)
	perpetrator := actorEvent.Actor

	if actor.CanSee(perpetrator.Position()) {
		return NewEnemySightedEvent(perpetrator), actor.TimeNeededForActions()
	}

	// barks
	if g.shouldActorBark(actor) && g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
		actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
	}

	return g.actorMakeMove(actor, actor.getMoveTowards(g.currentMap(), b.LastKnownPosition), false)
	//return g.actorTakeStepTowardsOther(actor, leader, 2)
}

func (b SearchBehaviour) Init(state *GameState, actor *Actor) {

}

func (b SearchBehaviour) IsCombatBehavior() bool {
	return false
}

func (b SearchBehaviour) IsHostilityTowards(other *Actor) bool {
	return false
}
