package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
)

func (g *GameState) enemyMovement(playerTimeSpent int) {
	if g.currentDungeonLevel <= 2 {
		playerTimeSpent /= 2
	}
	gridMap := g.gridMap
	allEnemies := gridMap.Actors()
	for _, enemy := range allEnemies {
		if enemy == g.Player {
			continue
		}
		if !enemy.IsAlive() {
			continue
		}

		// IMPORTANT:
		// Actions of enemies should never remove actors from the game directly
		enemy.AddTimeEnergy(playerTimeSpent)
		for enemy.HasEnergyForActions() {
			enemy.SpendTimeEnergy()
			g.aiAct(enemy)
		}
	}
}
func (g *GameState) removeDeadAndApplyRegeneration() {
	healInterval := 2 + (100 / g.Player.charSheet.GetStat(special.Endurance))
	hungerInterval := 300

	g.Player.decrementStatusEffectCounters()

	if !g.Player.HasFlag(foundation.FlagSlowDigestion) || g.TurnsTaken%2 == 0 {
		g.Player.GetFlags().Increment(foundation.FlagTurnsSinceEating)
	}

	turnsSinceEating := g.Player.GetFlags().Get(foundation.FlagTurnsSinceEating)

	if turnsSinceEating%hungerInterval == 0 {
		wasHungry := g.Player.IsHungry()
		g.Player.GetFlags().Increment(foundation.FlagHunger)
		if g.Player.IsHungry() && !wasHungry {
			g.msg(foundation.Msg("You are hungry."))
		}
	}

	if g.Player.IsHungry() && turnsSinceEating%(healInterval*3) == 0 {
		//g.Player.LooseActionPoints(1)
	}

	if !g.Player.IsHungry() && g.TurnsTaken%healInterval == 0 && len(g.playerVisibleEnemiesByDistance()) == 0 {
		if g.Player.NeedsHealing() {
			g.Player.Heal(1)
		} else {
			//g.Player.AddFatiguePoints(1)
		}
	}

	for i := len(g.gridMap.Actors()) - 1; i >= 0; i-- {
		actor := g.gridMap.Actors()[i]
		if !actor.IsAlive() {
			g.gridMap.RemoveActor(actor)
		} else {
			actor.AfterTurn()
			if actor.HasFlag(foundation.FlagRegenerating) && actor.NeedsHealing() {
				actor.Heal(1)
			}
		}
	}
	for i := len(g.gridMap.Objects()) - 1; i >= 0; i-- {
		object := g.gridMap.Objects()[i]
		if !object.IsAlive() {
			g.gridMap.RemoveObject(object)
		}
	}
}
