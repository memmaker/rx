package game

import (
	"RogueUI/foundation"
	"RogueUI/fsmai"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

func BehaviourPanicInit(g *GameState, actor *Actor, event fsmai.TransitionEvent) {
	actorEvent := event.(ActorEvent)
	actor.SetGoal(GoalFleeFromActor(actor, actorEvent.Actor))
}

func BehaviourPanic(g *GameState, actor *Actor, event fsmai.TransitionEvent) (fsmai.TransitionEvent, int) {
	// act on goals
	if actor.HasActiveGoal() {
		return actor.ActOnGoal(g)
	}

	distanceToPlayer := geometry.DistanceChebyshev(actor.Position(), g.Player.Position())
	nearEachOther := distanceToPlayer <= 7

	// barks
	if nearEachOther && g.canPlayerSee(actor.Position()) && actor.chatterFile != "" && actor.GetFlags().Get(foundation.FlagTurnsSinceLastIdleChatter) > 40 && rand.Intn(4) == 0 {
		if g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
			actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
		}
	}

	return fsmai.NoEvent, actor.timeEnergy
}
