package game

import (
	"contractor/foundation"
	"contractor/fsmai"
)

type FollowBehaviour struct {
	InitEvent fsmai.TransitionEvent
}

func (b FollowBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
	return FollowBehaviour{InitEvent: event}
}

func (b FollowBehaviour) AssociatedState() fsmai.StateName { return fsmai.StateFollow }

func (b FollowBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
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
