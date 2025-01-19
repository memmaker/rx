package game

import (
	"contractor/foundation"
	"contractor/fsmai"
	"github.com/memmaker/go/geometry"
)

type HuntBehaviour struct {
	InitEvent         fsmai.TransitionEvent
	LastKnownPosition geometry.Point
}

func (b HuntBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
	return HuntBehaviour{InitEvent: event, LastKnownPosition: event.(ActorEvent).Actor.Position()}
}

func (b HuntBehaviour) AssociatedState() fsmai.StateName { return fsmai.StateHunt }

func (b HuntBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
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

func (b HuntBehaviour) Init(state *GameState, actor *Actor) {

}

func (b HuntBehaviour) IsCombatBehavior() bool {
	return false
}

func (b HuntBehaviour) IsHostilityTowards(other *Actor) bool {
	return false
}
