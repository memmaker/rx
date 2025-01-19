package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

type PanicBehaviour struct {
	InitEvent TransitionEvent
}

func (b PanicBehaviour) IsCombatBehavior() bool {
	return false
}

func (b PanicBehaviour) IsHostilityTowards(other *Actor) bool {
	return false
}

func (b PanicBehaviour) WithInitEvent(event TransitionEvent) ActorBehavior {
	return PanicBehaviour{InitEvent: event}
}
func (b PanicBehaviour) AssociatedState() StateName { return StatePanic }

func (b PanicBehaviour) Init(state *GameState, actor *Actor) {

}

func (b PanicBehaviour) Execute(g *GameState, actor *Actor) (TransitionEvent, int) {
	// act on goals

	distanceToPlayer := geometry.DistanceChebyshev(actor.Position(), g.Player.Position())
	nearEachOther := distanceToPlayer <= 7

	// barks
	if nearEachOther && g.Player.CanSee(actor.Position()) && actor.ChatterFile != "" && actor.GetFlags().Get(foundation.FlagTurnsSinceLastIdleChatter) > 40 && rand.Intn(4) == 0 {
		if g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
			actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
		}
	}

	return moveAwayFromActor(g, actor, g.Player)
}
