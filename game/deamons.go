package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/fxtools"
	"path"
	"time"
)

func (g *GameState) enemyMovement(playerTimeSpent int) {
	gridMap := g.currentMap()
	allEnemies := gridMap.Actors()
	for _, enemy := range allEnemies {
		if enemy == g.Player {
			continue
		}
		if !enemy.IsAlive() {
			continue
		}

		// TODO: Fix the time energy system
		// Currently actors go into negative energy, which is not intended
		// Also taxi driver skips a turn and then moves twice, when chasing the player

		// IMPORTANT:
		// Actions of enemies should never remove actors from the game directly
		enemy.AddTimeEnergy(playerTimeSpent)
		hasActions := true
		for hasActions {
			tuSpent := g.TryAIAction(enemy)
			if tuSpent == 0 {
				hasActions = false
			} else {
				enemy.SpendTimeEnergy(tuSpent)
				//g.msg(foundation.HiLite("%s spent %s time units", enemy.Name(), strconv.Itoa(tuSpent)))
			}
		}
	}
}

func (g *GameState) isPlayerHungry() bool {
	lastEaten := g.GetNamedTime("PlayerLastAteAt")
	hours := 24
	if g.Player.HasFlag(foundation.FlagSlowDigestion) {
		hours = 48
	}
	// more than 24 hours since last meal
	if g.gameTime.Time.Sub(lastEaten.Time) > time.Duration(hours)*time.Hour {
		return true
	}
	return false
}

func (g *GameState) isPlayerStarving() bool {
	lastEaten := g.GetNamedTime("PlayerLastAteAt")
	hours := 72
	if g.Player.HasFlag(foundation.FlagSlowDigestion) {
		hours = 144
	}
	if g.gameTime.Time.Sub(lastEaten.Time) > time.Duration(hours)*time.Hour {
		return true
	}
	return false
}

func (g *GameState) removeDeadAndApplyTurnCounters() {
	if g.isPlayerStarving() {

		g.Player.SetFlag(foundation.FlagStarving)
		g.Player.GetFlags().Unset(foundation.FlagHunger)

		if g.Player.HasActionPoints() {
			g.Player.GetCharSheet().LooseActionPoints(1)
		} else if g.Player.GetHitPoints() > 1 {
			g.Player.GetCharSheet().TakeRawDamage(1)
		}
	} else if g.isPlayerHungry() {
		g.Player.SetFlag(foundation.FlagHunger)
		g.Player.GetFlags().Unset(foundation.FlagStarving)
	}

	allActorsOnThisMap := g.currentMap().Actors()
	for i := len(allActorsOnThisMap) - 1; i >= 0; i-- {
		actor := allActorsOnThisMap[i]
		wornOff := actor.AfterTurn()
		if actor == g.Player && len(wornOff) > 0 {
			for _, effect := range wornOff {
				g.msg(effect)
			}
		}
		if actor.HasFlag(foundation.FlagRegenerating) && actor.IsWounded() {
			actor.Heal(1)
		}
	}

}

func (g *GameState) updateAllSchedules() {
	allActorsOnThisMap := g.currentMap().Actors()
	for i := len(allActorsOnThisMap) - 1; i >= 0; i-- {
		actor := allActorsOnThisMap[i]

		if actor.schedule == nil || actor == g.Player {
			continue
		}
		// update scheduled actions
		if g.hasNewGoalFromSchedule(actor) {
			return
		}

		// remove actors that are transitioning to another map via schedule
		isAtScheduledLocation := g.isAtScheduledLocation(actor)
		if !isAtScheduledLocation {
			continue
		}

		if transition, isTransition := g.currentMap().GetTransitionAt(actor.Position()); isTransition && actor.HasFlag(foundation.FlagWantsToTransition) {
			g.removeActorFromMap(g.currentMap(), actor, transition)
			actor.UnsetFlag(foundation.FlagWantsToTransition)
			continue
		}
	}
	// PROBLEM: We are missing those actors, that only decided to leave the map when the player already left.
	for activeMap, actors := range g.allActorsInOtherActiveMaps() {
		for _, actor := range actors {
			if actor.schedule == nil {
				continue
			}

			var currentSlot TimeSlot
			if slot, move := actor.MoveToNextTimeSlot(g.gameTime.Time); move {
				currentSlot = slot
			} else {
				currentSlot = actor.schedule.CurrentTimeSlot()
			}

			// check if this slot was more than 2 minute ago
			if g.gameTime.Time.Sub(currentSlot.Time) <= 2*time.Minute {
				continue
			}

			locationName := currentSlot.Location
			location := activeMap.GetNamedLocation(locationName)
			transition, isTransition := activeMap.GetTransitionAt(location)
			if !isTransition {
				actor.SetPosition(location)
				continue
			}

			targetMap := transition.TargetMap

			g.setScheduleFromMapName(actor, targetMap)
			g.trySetGoalFromSchedule(actor)

			if targetMap != g.currentMap().GetName() {
				g.removeActorFromMap(activeMap, actor, transition)
				continue
			}

			// actor wants to transition to the current map
			// remove the actor from the outOfGame map
			activeMap.RemoveActor(actor)

			spawnPos := g.currentMap().GetNamedLocation(transition.TargetLocation)
			g.currentMap().AddActorWithDisplacement(actor, spawnPos)
		}
	}
	// check for schedules that would make an actor transition to another map
	// this does not care about actor's that were already trying to transition
	// to the current map, this is handled after map loading.
	for actor, _ := range g.outOfGame {
		if actor.schedule == nil {
			continue
		}

		var currentSlot TimeSlot
		if slot, move := actor.MoveToNextTimeSlot(g.gameTime.Time); move {
			currentSlot = slot
		} else {
			currentSlot = actor.schedule.CurrentTimeSlot()
		}

		// check if this slot was more than 2 minute ago
		if g.gameTime.Time.Sub(currentSlot.Time) <= 2*time.Minute {
			continue
		}

		locationName := currentSlot.Location
		location := g.currentMap().GetNamedLocation(locationName)
		transition, isTransition := g.currentMap().GetTransitionAt(location)
		if !isTransition {
			actor.SetPosition(location)
			continue
		}

		targetMap := transition.TargetMap

		g.setScheduleFromMapName(actor, targetMap)

		if targetMap != g.currentMap().GetName() {
			// update target map and schedule
			g.outOfGame[actor] = TimedTransition{
				Destination: transition,
				Time:        g.gameTime.Time,
			}
			continue
		}

		// actor wants to transition to the current map
		// remove the actor from the outOfGame map
		delete(g.outOfGame, actor)

		spawnPos := g.currentMap().GetNamedLocation(transition.TargetLocation)
		g.currentMap().AddActorWithDisplacement(actor, spawnPos)
	}
}

func (g *GameState) isAtScheduledLocation(actor *Actor) bool {
	if actor.schedule == nil {
		return false
	}
	location := actor.schedule.CurrentTimeSlot().Location
	if location == "" {
		return false
	}
	return actor.Position() == g.currentMap().GetNamedLocation(location)
}

func (g *GameState) setScheduleFromMapName(actor *Actor, mapName string) {
	schedulePath := path.Join(g.config.DataRootDir, "maps", mapName, "schedules", actor.GetInternalName()+".rec")
	if fxtools.FileExists(schedulePath) {
		actor.schedule = NewScheduleFromFile(schedulePath)
	}
}
