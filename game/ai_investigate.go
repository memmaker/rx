package game

import (
	"contractor/foundation"
	"fmt"
)

type InvestigateBehaviour struct {
	InitEvent TransitionEvent
}

func (b InvestigateBehaviour) WithInitEvent(event TransitionEvent) ActorBehavior {
	return InvestigateBehaviour{InitEvent: event}
}

func (b InvestigateBehaviour) AssociatedState() StateName { return StateInvestigate }

func (b InvestigateBehaviour) Execute(g *GameState, actor *Actor) (TransitionEvent, int) {
	locEvent := b.InitEvent.(LocationEvent)
	location := locEvent.Location
	if g.currentMapName != location.MapName {
		return g.actorTakeStepToMapPosition(actor, location, false)
	}

	interestingPos := location.Position

	distToPos := g.currentMap().MoveDistance(actor.Position(), interestingPos)

	if actor.CanSee(interestingPos) {
		if distToPos == 2 {
			g.tryAddRandomChatter(actor, foundation.ChatterInvestigating)
		}
		if distToPos <= 1 {
			// check if we can wake up a sleeping ally
			if actorAt, isActorHere := g.currentMap().TryGetActorAt(interestingPos); isActorHere {
				if actorAt != actor && actorAt.Faction == actor.Faction && actorAt.IsSleeping() {
					g.gameFlags.Increment(fmt.Sprintf("wake_ups(%s)", actorAt.Faction))
					actorAt.WakeUp()
					g.msg(foundation.HiLite("%s wakes up", actorAt.Name()))
					return NoEvent, actor.TimeNeededForMovement()
				}
			}

			return EmptyEvent{
				Event: EventCalmed,
			}, actor.TimeNeededForMovement()
		}
	}

	return g.actorMakeMove(actor, actor.getMoveTowards(g.currentMap(), interestingPos), false)
}

func (b InvestigateBehaviour) Init(g *GameState, actor *Actor) {
	g.ui.AddAnimations(OneAnimation(g.ui.GetAnimSuspicious(actor, nil)))
}

func (b InvestigateBehaviour) IsCombatBehavior() bool {
	return false
}

func (b InvestigateBehaviour) IsHostilityTowards(other *Actor) bool {
	return false
}
