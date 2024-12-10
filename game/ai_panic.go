package game

import (
	"contractor/foundation"
	"contractor/fsmai"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

type PanicBehaviour struct {
	InitEvent fsmai.TransitionEvent
}
func (b PanicBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
	return PanicBehaviour{InitEvent: event}
}
func (b PanicBehaviour) AssociatedState() fsmai.StateName { return fsmai.StateKill }

func (b PanicBehaviour) Init(state *GameState, actor *Actor) {
	actorEvent := b.InitEvent.(ActorEvent)
	actor.SetGoal(GoalFleeFromActor(actorEvent.Actor))
}

func (b PanicBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
	// act on goals
	if actor.HasActiveGoal() {
		return actor.ActOnGoal(g)
	}

	distanceToPlayer := geometry.DistanceChebyshev(actor.Position(), g.Player.Position())
	nearEachOther := distanceToPlayer <= 7

	// barks
	if nearEachOther && g.canPlayerSee(actor.Position()) && actor.ChatterFile != "" && actor.GetFlags().Get(foundation.FlagTurnsSinceLastIdleChatter) > 40 && rand.Intn(4) == 0 {
		if g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
			actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
		}
	}

	return fsmai.NoEvent, actor.RawTimeEnergy
}
