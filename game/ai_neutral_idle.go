package game

import (
	"RogueUI/foundation"
	"RogueUI/fsmai"
)

func BehaviourNeutralIdleInit(g *GameState, actor *Actor, event fsmai.TransitionEvent) {
	actor.SetGoal(GoalMoveToSpawn())
}

func BehaviourNeutralIdle(g *GameState, actor *Actor, event fsmai.TransitionEvent) (fsmai.TransitionEvent, int) {
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

	return fsmai.NoEvent, actor.timeEnergy
}
