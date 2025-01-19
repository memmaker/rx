package game

import (
    "contractor/foundation"
)

type ScriptedBehaviour struct {
    InitEvent TransitionEvent
}

func (b ScriptedBehaviour) WithInitEvent(event TransitionEvent) ActorBehavior {
    return ScriptedBehaviour{InitEvent: event}
}

func (b ScriptedBehaviour) AssociatedState() StateName { return StateScripted }

func (b ScriptedBehaviour) Execute(g *GameState, actor *Actor) (TransitionEvent, int) {
    // barks
    if g.shouldActorBark(actor) && g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
        actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
    }

    switch b.InitEvent.(type) {
    case LocationEvent:
        locationEvent := b.InitEvent.(LocationEvent)
        return g.actorTakeStepToMapPosition(actor, locationEvent.Location, false)
    case ActorEvent:
        actorEvent := b.InitEvent.(ActorEvent)
        return g.actorTakeStepTowardsOther(actor, actorEvent.Actor, 2)
    }

    return NoEvent, actor.maximalTimeNeededForActions()
}

func (b ScriptedBehaviour) Init(state *GameState, actor *Actor) {

}

func (b ScriptedBehaviour) IsCombatBehavior() bool {
    return false
}

func (b ScriptedBehaviour) IsHostilityTowards(other *Actor) bool {
    return false
}
