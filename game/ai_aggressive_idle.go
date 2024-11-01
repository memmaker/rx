package game

import (
	"RogueUI/foundation"
	"RogueUI/fsmai"
)

func BehaviourAggressiveIdleInit(g *GameState, actor *Actor, event fsmai.TransitionEvent) {
	actor.SetGoal(GoalMoveToSpawn())
}

func BehaviourAggressiveIdle(g *GameState, actor *Actor, event fsmai.TransitionEvent) (fsmai.TransitionEvent, int) {
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

func (g *GameState) trySetGoalFromSchedule(actor *Actor) {
	if actor.schedule == nil || actor == g.Player {
		return
	}

	if g.hasNewGoalFromSchedule(actor) {
		return
	}

	if !actor.HasActiveGoal() {
		currentSlot := actor.schedule.CurrentTimeSlot()
		loc := g.currentMap().GetNamedLocation(currentSlot.Location)
		if actor.Position() != loc {
			g.msg(foundation.HiLite("%s moves to %s", actor.Name(), currentSlot.Location))
			actor.SetGoal(GoalStrideToLocation(loc))
		}
	}
}

func (g *GameState) hasNewGoalFromSchedule(actor *Actor) bool {
	if slot, move := actor.MoveToNextTimeSlot(g.gameTime.Time); move {
		loc := g.currentMap().GetNamedLocation(slot.Location)
		g.msg(foundation.HiLite("%s moves to %s", actor.Name(), slot.Location))
		actor.SetGoal(GoalStrideToLocation(loc))
		if g.currentMap().IsTransitionAt(loc) {
			actor.SetFlag(foundation.FlagWantsToTransition)
		}
		if actor.IsSleeping() {
			actor.WakeUp()
		}
		return true
	}
	return false
}
