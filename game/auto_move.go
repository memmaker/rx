package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
)

func (g *GameState) RunPlayer(direction geometry.CompassDirection, isStarting bool) bool {
	player := g.Player
	if player.HasFlag(foundation.FlagConfused) {
		g.msg(foundation.Msg("You cannot run while confused"))
		return false
	}

	if len(g.GetVisibleEnemies()) > 0 {
		g.msg(foundation.Msg("You cannot run while enemies are near"))
		return false
	}

	currentPos := player.Position()
	targetPos := currentPos.Add(direction.ToPoint())
	currentMap := g.gridMap

	if !isStarting {
		if !currentMap.IsTileWalkable(targetPos) && !direction.IsDiagonal() {

			leftDir := direction.TurnLeftBy90()
			rightDir := direction.TurnRightBy90()
			leftTarget := currentPos.Add(leftDir.ToPoint())
			rightTarget := currentPos.Add(rightDir.ToPoint())
			isFreeLeft := currentMap.IsTileWalkable(leftTarget)
			isFreeRight := currentMap.IsTileWalkable(rightTarget)

			if (isFreeLeft && isFreeRight) || (!isFreeLeft && !isFreeRight) {
				return false
			}
			if isFreeLeft {
				targetPos = leftTarget
			} else if isFreeRight {
				targetPos = rightTarget
			}
		}
	}

	if !currentMap.IsCurrentlyPassable(targetPos) {
		return false
	}

	currentDirection := targetPos.Sub(currentPos).ToDirection()
	if g.dungeonLayout != nil && g.dungeonLayout.IsCorridor(currentPos) && g.dungeonLayout.IsDoorAt(targetPos) {
		firstRoomTileAfterDoor := targetPos.Add(currentDirection.ToPoint())
		if !currentMap.IsExplored(firstRoomTileAfterDoor) {
			return false
		}
	}
	g.playerMove(currentPos, targetPos)

	g.ui.AfterPlayerMoved(foundation.MoveInfo{
		Direction: currentDirection,
		OldPos:    currentPos,
		NewPos:    targetPos,
		Mode:      foundation.PlayerMoveModeRun,
	})
	return true
}
