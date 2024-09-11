package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
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
			}
		}
	}
}
func (g *GameState) removeDeadAndApplyRegeneration() {
	healInterval := 2 + (100 / g.Player.charSheet.GetStat(special.Endurance))
	hungerInterval := 300

	g.Player.decrementStatusEffectCounters()

	if !g.Player.HasFlag(special.FlagSlowDigestion) || g.TurnsTaken()%2 == 0 {
		g.Player.GetFlags().Increment(special.FlagTurnsSinceEating)
	}

	turnsSinceEating := g.Player.GetFlags().Get(special.FlagTurnsSinceEating)

	if turnsSinceEating%hungerInterval == 0 {
		wasHungry := g.Player.IsHungry()
		g.Player.GetFlags().Increment(special.FlagHunger)
		if g.Player.IsHungry() && !wasHungry {
			g.msg(foundation.Msg("You are hungry."))
		}
	}

	if g.Player.IsHungry() && turnsSinceEating%(healInterval*3) == 0 {
		//g.Player.LooseActionPoints(1)
	}

	if !g.Player.IsHungry() && g.TurnsTaken()%healInterval == 0 && len(g.playerVisibleEnemiesByDistance()) == 0 {
		if g.Player.NeedsHealing() {
			g.Player.Heal(1)
		} else {
			//g.Player.AddFatiguePoints(1)
		}
	}

	for i := len(g.currentMap().Actors()) - 1; i >= 0; i-- {
		actor := g.currentMap().Actors()[i]
		actor.AfterTurn()
		if actor.HasFlag(special.FlagRegenerating) && actor.NeedsHealing() {
			actor.Heal(1)
		}
	}
	for i := len(g.currentMap().Objects()) - 1; i >= 0; i-- {
		object := g.currentMap().Objects()[i]
		if !object.IsAlive() {
			g.currentMap().RemoveObject(object)
		}
	}
}
