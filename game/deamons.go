package game

import (
	"contractor/foundation"
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
		g.actorsComputed++
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

func (g *GameState) applyTurnCounters() {
	currentMap := g.currentMap()
	allActorsOnThisMap := currentMap.Actors()
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

func (g *GameState) applyPlayerHunger() {
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
}

// updateAllSchedules is called whenever time advances.
// It will set a new goal from a schedule, if applicable.
// It will also make an actor transition to another map, if applicable.
func (g *GameState) updateAllSchedules() {
	gridmap := g.currentMap()
	allActorsOnThisMap := gridmap.Actors()

	for i := len(allActorsOnThisMap) - 1; i >= 0; i-- {
		actor := allActorsOnThisMap[i]

		if actor.Schedule == nil || actor == g.Player {
			continue
		}
		// update scheduled actions
		if g.hasNewGoalFromSchedule(actor) {
			return
		}

		if transition, isTransition := gridmap.GetTransitionAt(actor.Position()); isTransition && actor.HasFlag(foundation.FlagWantsToTransition) {
			g.actorTransition(gridmap, actor, transition)
			actor.UnsetFlag(foundation.FlagWantsToTransition)
			continue
		}
	}
}

func (g *GameState) isAtScheduledLocation(actor *Actor) bool {
	if actor.Schedule == nil {
		return false
	}
	location := actor.Schedule.CurrentTimeSlot().Location
	if location == "" {
		return false
	}
	return actor.Position() == g.currentMap().GetNamedLocation(location)
}

func (g *GameState) loadSchedule(actor *Actor) {
	schedulePath := path.Join(g.config.DataRootDir, "schedules", actor.GetInternalName()+".rec")
	if fxtools.FileExists(schedulePath) {
		actor.Schedule = NewScheduleFromFile(schedulePath)
	}
}
