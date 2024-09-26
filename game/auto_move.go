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

	if currentPos == targetPos {
		return false
	}

	currentMap := g.currentMap()

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

	g.playerMove(currentPos, targetPos)
	g.gameFlags.Increment("playerRunSteps")
	g.ui.AfterPlayerMoved(foundation.MoveInfo{
		Direction: currentDirection,
		OldPos:    currentPos,
		NewPos:    targetPos,
		Mode:      foundation.PlayerMoveModeRun,
	})
	return true
}

func (g *GameState) RunPlayerPath() bool {
	player := g.Player
	if player.HasFlag(foundation.FlagConfused) {
		g.Player.RemoveGoal()
		g.msg(foundation.Msg("You cannot run while confused"))
		return false
	}

	if len(g.GetVisibleEnemies()) > 0 {
		g.Player.RemoveGoal()
		g.msg(foundation.Msg("You cannot run while enemies are near"))

		return false
	}

	if !g.Player.HasActiveGoal() || g.Player.cannotFindPath() {
		return false
	}

	oldPos := g.Player.Position()
	g.Player.ActOnGoal(g)
	newPos := g.Player.Position()

	if oldPos == newPos {
		//g.ui.StopAutoMove()
		return false
	}

	direction := newPos.Sub(oldPos).ToDirection()
	g.afterPlayerMoved(oldPos, false)
	g.endPlayerTurn(g.Player.timeNeededForMovement())
	g.gameFlags.Increment("playerRunSteps")
	g.ui.AfterPlayerMoved(foundation.MoveInfo{
		Direction: direction,
		OldPos:    oldPos,
		NewPos:    newPos,
		Mode:      foundation.PlayerMoveModePath,
	})
	return true
}
