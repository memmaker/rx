package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/fxtools"
	"path"
	"time"
)

// endPlayerTurn is called by game actions that end the player's turn.
// It will
// - animate the player's actions
// - run the AI for all enemies
// - advance time and turn counter
// - trigger "before enemy turn" effects
// - animate the enemies' actions
// - trigger "after enemy turn" effects
// - simulate all other active maps
// - execute any actions that were queued to be executed after animations
// - update the UI status
// - check if the player can act
func (g *GameState) endPlayerTurn(playerTimeTakenForTurn int) {
	// player has changed the game state..
	g.actorsComputed = 0

	didCancel := g.ui.AnimatePending() // animate player actions..

	// AI Actions (incl. Behaviours, Goals, Schedules) and State Changes happen here
	g.enemyMovement(playerTimeTakenForTurn)

	if didCancel {
		g.ui.SkipAnimations()
	} else {
		didCancel = g.ui.AnimatePending() // animate enemy actions
	}

	// EXPERIMENTAL and dangerous..
	// we simulate all actors on all loaded maps..
	if g.config.SimulateAllLoadedMaps {
		g.onOtherMaps(func(mapName string) {
			g.enemyMovement(playerTimeTakenForTurn)
			g.ui.SkipAnimations()
		})
	}

	// advancing the time will run scripts
	g.advanceTimeAndTurn(time.Second * time.Duration(float64(playerTimeTakenForTurn)/10))

	// EXPERIMENTAL and dangerous..
	// we simulate all actors on all loaded maps..
	if g.config.SimulateAllLoadedMaps {
		g.onOtherMaps(func(mapName string) { g.afterTurnEffectsForActors() })
	}

	g.ui.SkipAnimations()

	// trigger turn based events
	g.afterTurn()

	if didCancel {
		g.ui.SkipAnimations()
	} else {
		g.ui.AnimatePending() // animate afterTurn effects
	}

	// This is where level transitions are handled
	for _, action := range g.afterAnimationActions {
		action()
	}

	g.afterAnimationActions = nil

	g.checkPlayerCanAct()

	g.gameFlags.Set("ActorsComputed", g.actorsComputed)

	if g.Player.HasFlag(foundation.FlagSneaking) {
		g.ui.SetSneakOverlay(g.createSneakOverlay())
	} else {
		g.ui.SetSneakOverlay(nil)
	}

	g.updateUIStatus()
}

func (g *GameState) onOtherMaps(call func(string)) {
	for mapName, _ := range g.activeMaps {
		if mapName == g.currentMapName {
			continue
		}
		g.ExecuteOnMap(mapName, func() {
			call(mapName)
		})
	}
}

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

func (g *GameState) afterTurn() {
	g.metronome.Tick(g)

	g.checkJournal()

	g.afterTurnEffectsForActors()

	// handle player hunger
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

func (g *GameState) afterTurnEffectsForActors() {
	// apply after turn effects for all actors
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

func (g *GameState) isAtScheduledLocation(actor *Actor, slot TimeSlot) bool {
	location := slot.Location
	if location == "" {
		return false
	}
	return actor.Position() == g.currentMap().GetNamedLocation(location)
}

func (g *GameState) loadSchedule(actor *Actor) {
	schedulePath := path.Join(g.config.DataRootDir, "schedules", actor.GetInternalName()+".rec")
	if fxtools.FileExists(schedulePath) {
		actor.Schedule = NewScheduleFromFile(schedulePath, func() time.Time {
			return g.gameTime.Time
		})
	}
}
