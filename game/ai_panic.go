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

func (b PanicBehaviour) IsCombatBehavior() bool {
    return false
}

func (b PanicBehaviour) IsHostilityTowards(other *Actor) bool {
    return false
}

func (b PanicBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
    return PanicBehaviour{InitEvent: event}
}
func (b PanicBehaviour) AssociatedState() fsmai.StateName { return fsmai.StatePanic }

func (b PanicBehaviour) Init(state *GameState, actor *Actor) {
	
}

func (b PanicBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
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
