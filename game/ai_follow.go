package game

import (
	"contractor/foundation"
)

type FollowBehaviour struct {
	InitEvent TransitionEvent
}

func (b FollowBehaviour) WithInitEvent(event TransitionEvent) ActorBehavior {
	return FollowBehaviour{InitEvent: event}
}

func (b FollowBehaviour) AssociatedState() StateName { return StateFollow }

func (b FollowBehaviour) Execute(g *GameState, actor *Actor) (TransitionEvent, int) {
	// barks
	if g.shouldActorBark(actor) && g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
		actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
	}

	actorEvent := b.InitEvent.(ActorEvent)
	leader := actorEvent.Actor

	if leader.IsInCombat() {
		return NewProvokedEvent(leader.GetOpponent()), actor.TimeNeededForActions()
	}

	return g.actorTakeStepTowardsOther(actor, leader, 2)
}

func (b FollowBehaviour) Init(state *GameState, actor *Actor) {

}

func (b FollowBehaviour) IsCombatBehavior() bool {
	return false
}

func (b FollowBehaviour) IsHostilityTowards(other *Actor) bool {
	return false
}
