package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/geometry"
)

type HuntBehaviour struct {
	InitEvent         TransitionEvent
	LastKnownPosition geometry.Point
}

func (b HuntBehaviour) WithInitEvent(event TransitionEvent) ActorBehavior {
	return HuntBehaviour{InitEvent: event, LastKnownPosition: event.(ActorEvent).Actor.Position()}
}

func (b HuntBehaviour) AssociatedState() StateName { return StateHunt }

func (b HuntBehaviour) Execute(g *GameState, actor *Actor) (TransitionEvent, int) {
	actorEvent := b.InitEvent.(ActorEvent)
	perpetrator := actorEvent.Actor

	if actor.CanSee(perpetrator.Position()) &&
		g.isDetectedByObserver(perpetrator, actor) {
		// hunt is over
		return NewEnemySightedEvent(perpetrator), actor.TimeNeededForActions()
	}

	// barks
	if g.shouldActorBark(actor) && g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
		actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
	}

	distance := g.currentMap().MoveDistance(actor.Position(), b.LastKnownPosition)
	if distance <= 5 {
		actor.StatusFlags.Increment(foundation.FlagTurnsHunting)
		if actor.StatusFlags.Get(foundation.FlagTurnsHunting) > 10 {
			return NewCalmedEvent(perpetrator), actor.TimeNeededForActions()
		}
	}

	return g.actorMakeMove(actor, actor.getMoveTowards(g.currentMap(), b.LastKnownPosition), true)
	//return g.actorTakeStepTowardsOther(actor, leader, 2)
}

func (b HuntBehaviour) Init(state *GameState, actor *Actor) {

}

func (b HuntBehaviour) IsCombatBehavior() bool {
	return true
}

func (b HuntBehaviour) IsHostilityTowards(other *Actor) bool {
	actorEvent := b.InitEvent.(ActorEvent)
	return other == actorEvent.Actor
}
