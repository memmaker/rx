package game

import (
	"RogueUI/foundation"
	"RogueUI/fsmai"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

func BehaviourAggressiveIdleInit(g *GameState, actor *Actor, event fsmai.TransitionEvent) {
	actor.SetGoal(GoalMoveToSpawn())
}

func BehaviourAggressiveIdle(g *GameState, actor *Actor, event fsmai.TransitionEvent) (fsmai.TransitionEvent, int) {
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

	// random animal movement
	if actor.HasFlag(foundation.FlagAnimal) {
		consequencesOfConfusion := g.actConfused(actor)
		if len(consequencesOfConfusion) > 0 {
			g.ui.AddAnimations(consequencesOfConfusion)
			return fsmai.NoEvent, actor.maximalTimeNeededForActions()
		}
	}

	// scheduled actions
	if slot, move := actor.MoveToNextTimeSlot(g.gameTime.Time); move {
		loc := g.currentMap().GetNamedLocation(slot.Location)
		g.msg(foundation.HiLite("%s moves to %s", actor.Name(), slot.Location))
		actor.SetGoal(GoalMoveToLocation(loc))
	}

	return fsmai.NoEvent, actor.timeEnergy
}
