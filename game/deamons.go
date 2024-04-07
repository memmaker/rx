package game

import (
	"RogueUI/foundation"
	"RogueUI/rpg"
)

func (g *GameState) enemyMovement() {
	gridMap := g.gridMap
	allEnemies := gridMap.Actors()
	for _, enemy := range allEnemies {
		if enemy == g.Player {
			continue
		}
		// IMPORTANT:
		// Actions of enemies may never remove actors from the game directly
		g.executeAI(enemy)
	}
}
func (g *GameState) removeDeadAndApplyRegeneration() {
	healEvery := 10
	if g.TurnsTaken%healEvery == 0 {
		if g.Player.NeedsHealing() && len(g.playerVisibleEnemiesByDistance()) == 0 {
			g.Player.Heal(1)
		}
		g.Player.charSheet.AddToCounter(rpg.CounterHunger, 1)
	}

	for i := len(g.gridMap.Actors()) - 1; i >= 0; i-- {
		actor := g.gridMap.Actors()[i]
		if !actor.IsAlive() {
			g.gridMap.RemoveActor(actor)
		} else if actor.HasFlag(foundation.IsRegenerating) && actor.NeedsHealing() {
			actor.Heal(1)
		}
	}
	for i := len(g.gridMap.Objects()) - 1; i >= 0; i-- {
		object := g.gridMap.Objects()[i]
		if !object.IsAlive() {
			g.gridMap.RemoveObject(object)
		}
	}
}
