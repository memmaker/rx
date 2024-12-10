package game

import (
	"contractor/foundation"
	"contractor/fsmai"
)

type AggressiveBehaviour struct {
	InitEvent fsmai.TransitionEvent
}

func (b AggressiveBehaviour) WithInitEvent(event fsmai.TransitionEvent) ActorBehavior {
	return AggressiveBehaviour{InitEvent: event}
}
func (b AggressiveBehaviour) AssociatedState() fsmai.StateName { return fsmai.StateAggressive }

func (b AggressiveBehaviour) Init(state *GameState, actor *Actor) {
	actor.SetGoal(GoalMoveToSpawn())
}

func (b AggressiveBehaviour) Execute(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
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

func (g *GameState) trySetGoalFromSchedule(actor *Actor) {
	if actor.Schedule == nil || actor == g.Player {
		return
	}

	if g.hasNewGoalFromSchedule(actor) {
		return
	}

	if !actor.HasActiveGoal() {
		currentSlot := actor.Schedule.CurrentTimeSlot()
		loc := g.currentMap().GetNamedLocation(currentSlot.Location)
		if actor.Position() != loc {
			g.msg(foundation.HiLite("%s moves to %s", actor.Name(), currentSlot.Location))
			actor.SetGoal(GoalWalkToLocation(loc))
		}
	}
}

func (g *GameState) hasNewGoalFromSchedule(actor *Actor) bool {
	if slot, move := actor.MoveToNextTimeSlot(g.gameTime.Time); move {

		targetMap := slot.MapName
		if targetMap != "" && targetMap != g.currentMap().GetName() {
			g.ensureMapIsLoaded(targetMap)

			transitionLoc := g.currentMap().GetNamedLocation(slot.UseTransition)
			targetLoc := g.activeMaps[targetMap].GetNamedLocation(slot.Location)

			actor.SetGoal(GoalWalkToLocationOnMap(transitionLoc, targetMap, targetLoc))
		} else {
			loc := g.currentMap().GetNamedLocation(slot.Location)
			g.msg(foundation.HiLite("%s moves to %s", actor.Name(), slot.Location))
			actor.SetGoal(GoalWalkToLocation(loc))
		}
		if actor.IsSleeping() {
			actor.WakeUp()
		}
		return true
	}
	return false
}
