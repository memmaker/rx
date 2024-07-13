package game

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"math/rand"
)

func (g *GameState) ManualMovePlayer(direction geometry.CompassDirection) {
	if !g.config.DiagonalMovementEnabled && direction.IsDiagonal() {
		return
	}

	player := g.Player
	oldPos := player.Position()

	// adapted from: https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/move.c#L56
	if player.HasFlag(foundation.FlagConfused) && rand.Intn(5) != 0 {
		direction = geometry.RandomDirection()
	}

	newPos := oldPos.Add(direction.ToPoint())

	if !g.gridMap.Contains(newPos) || !g.gridMap.IsTileWalkable(newPos) {
		if !g.config.WallSlide {
			return
		}
		// check for wallslide
		var forwardLeft, forwardRight geometry.Point
		var forwardLeftTest, forwardRightTest geometry.Point
		if direction.IsDiagonal() {
			dirVec := direction.ToPoint()
			leftDir := geometry.Point{0, dirVec.Y}
			rightDir := geometry.Point{dirVec.X, 0}
			forwardLeft = oldPos.Add(leftDir)
			forwardRight = oldPos.Add(rightDir)
			forwardLeftTest = forwardLeft
			forwardRightTest = forwardRight
		} else { // cardinal case
			rightDir := direction.TurnRightBy90()
			leftDir := direction.TurnLeftBy90()
			forwardLeft = oldPos.Add(leftDir.ToPoint())
			forwardRight = oldPos.Add(rightDir.ToPoint())

			forwardLeftTest = forwardLeft.Add(direction.ToPoint())
			forwardRightTest = forwardRight.Add(direction.ToPoint())
			if g.config.DiagonalMovementEnabled {
				forwardLeft = forwardLeftTest
				forwardRight = forwardRightTest
			}
		}

		if g.gridMap.IsCurrentlyPassable(forwardLeftTest) && !g.gridMap.IsTileWalkable(forwardRightTest) {
			newPos = forwardLeft // we can slide :)
		} else if g.gridMap.IsCurrentlyPassable(forwardRightTest) && !g.gridMap.IsTileWalkable(forwardLeftTest) {
			newPos = forwardRight // we can slide :)
		} else {
			return // no slide :(, no move
		}
	}

	if actorAt, exists := g.gridMap.TryGetActorAt(newPos); exists {
		if actorAt.IsHostile() {
			g.playerMeleeAttack(actorAt)
		} else {
			g.StartDialogue(actorAt.GetInternalName(), false)
		}
		return
	}

	if downedActorAt, exists := g.gridMap.TryGetDownedActorAt(newPos); exists {
		g.openInventoryOf(downedActorAt)
	}
	if objectAt, exists := g.gridMap.TryGetObjectAt(newPos); exists && !objectAt.IsWalkable(g.Player) {
		objectAt.OnBump(g.Player)
		return
	}
	direction = newPos.Sub(oldPos).ToDirection()
	g.playerMove(newPos)
	g.ui.AfterPlayerMoved(foundation.MoveInfo{
		Direction: direction,
		OldPos:    oldPos,
		NewPos:    newPos,
		Mode:      foundation.PlayerMoveModeManual,
	})
}

func (g *GameState) openInventoryOf(actor *Actor) {
	inventory := actor.GetInventory()

	stackRef := itemStacksForUI(inventory.StackedItemsWithFilter(func(item *Item) bool {
		return !item.HasTag(foundation.TagNoLoot)
	}))

	if inventory.IsEmpty() || len(stackRef) == 0 {
		g.msg(foundation.HiLite("There is nothing to pick up"))
		return
	}

	transfer := func(itemUI foundation.ItemForUI) {
		itemStack := itemUI.(*InventoryStack)
		for _, item := range itemStack.GetItems() {
			inventory.Remove(item)
			g.Player.GetInventory().Add(item)
		}

		if !inventory.IsEmpty() {
			g.openInventoryOf(actor)
		}
	}
	g.ui.ShowContainer(actor.Name(), stackRef, transfer)

}
func (g *GameState) afterPlayerMoved() {
	// explore the map
	// print "You see.." message
	if g.gridMap.IsItemAt(g.Player.Position()) && g.config.AutoPickup {
		g.PickupItem()
	}

	g.msg(g.GetMapInfoForMovement(g.Player.Position()))
	g.exploreMap()
	g.updateDijkstraMap()

	if g.Player.HasFlag(foundation.FlagCurseTeleportitis) && rand.Intn(100) < 5 {
		g.ui.AddAnimations(OneAnimation(teleportWithAnimation(g, g.Player, g.gridMap.RandomSpawnPosition())))
	}
}

func (g *GameState) updateDijkstraMap() {
	g.playerDijkstraMap = g.gridMap.GetDijkstraMapWithActorsNotBlocking(g.Player.Position(), 1000)
}

func (g *GameState) exploreMap() {
	if g.currentDungeonLevel == 0 {
		return
	}
	if g.dungeonLayout == nil { // no dungeon as a base
		g.updatePlayerFoVAndApplyExploration()
		return
	}

	playerRoom := g.getPlayerRoom()

	if playerRoom == nil {
		g.applyCorridorExploration()
	} else if playerRoom.IsLit() {
		g.applyLightRoomExploration(playerRoom)
		if g.dungeonLayout.IsDoorAt(g.Player.Position()) {
			g.applyDarkRoomExploration()
		}
		if g.TurnsTaken-playerRoom.LastSeenTurn > 50 {
			g.checkTilesForHiddenObjects(playerRoom.GetAbsoluteFloorTiles())
			playerRoom.LastSeenTurn = g.TurnsTaken
		}
	} else { // dark room, light the walls and explore the area around the player
		g.applyDarkRoomExploration()
	}
}
