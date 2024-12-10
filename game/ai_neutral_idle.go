package game

import (
	"contractor/foundation"
	"contractor/fsmai"
)

type NeutralBehaviour struct {
	InitEvent fsmai.TransitionEvent
}

func (b NeutralBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
	return NeutralBehaviour{InitEvent: event}
}

func (b NeutralBehaviour) AssociatedState() fsmai.StateName {return fsmai.StateNeutral}

func (b NeutralBehaviour) Init(state *GameState, actor *Actor) {
	actor.SetGoal(GoalMoveToSpawn())
}

func (b NeutralBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
	// act on goals
	if actor.HasActiveGoal() {
		return actor.ActOnGoal(g)
	}

	// barks
	if g.shouldActorBark(actor) && g.tryAddRandomChatter(actor, foundation.ChatterBeingAroundPlayer) {
		actor.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
	}

	// random animal movement
	if actor.HasFlag(foundation.FlagAnimal) {
		consequencesOfConfusion := g.actConfused(actor)
		if len(consequencesOfConfusion) > 0 {
			g.ui.AddAnimations(consequencesOfConfusion)
			return fsmai.NoEvent, actor.maximalTimeNeededForActions()
		}
	}

	return fsmai.NoEvent, actor.RawTimeEnergy
}