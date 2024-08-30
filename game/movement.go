package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

func (g *GameState) TryGetDoorAt(mapPos geometry.Point) (*Door, bool) {
	if obj, exists := g.currentMap().TryGetObjectAt(mapPos); exists {
		if door, isDoor := obj.(*Door); isDoor {
			return door, true
		}
	}
	return nil, false
}
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

	if objectAt, exists := g.currentMap().TryGetObjectAt(newPos); exists {
		if door, isDoor := objectAt.(*Door); isDoor { // THIS IS STILL HACKY; DOOR OBJECT ALSO IMPLEMENTS THIS
			if door.IsLocked() {
				if player.HasKey(door.GetLockFlag()) {
					door.Unlock()
					g.msg(foundation.Msg("You unlocked the door"))
					g.ui.PlayCue("world/PICKKEYS")
					return
				}
			}
		}

		if !objectAt.IsWalkable(g.Player) {
			objectAt.OnBump(g.Player)
			return
		}
	}

	if _, exists := g.currentMap().TryGetItemAt(newPos); exists && !g.currentMap().IsCurrentlyPassable(newPos) {
		g.PlayerPickupItemAt(newPos)
		return
	}

	if !g.currentMap().Contains(newPos) || !g.currentMap().IsTileWalkable(newPos) {
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

		if g.currentMap().IsCurrentlyPassable(forwardLeftTest) && !g.currentMap().IsTileWalkable(forwardRightTest) {
			newPos = forwardLeft // we can slide :)
		} else if g.currentMap().IsCurrentlyPassable(forwardRightTest) && !g.currentMap().IsTileWalkable(forwardLeftTest) {
			newPos = forwardRight // we can slide :)
		} else {
			return // no slide :(, no move
		}
	}

	if actorAt, exists := g.currentMap().TryGetActorAt(newPos); exists {
		if actorAt.IsHostileTowards(g.Player) {
			g.playerMeleeAttack(actorAt)
		} else if !actorAt.IsSleeping() && actorAt.HasDialogue() {
			g.StartDialogue(actorAt.GetDialogueFile(), actorAt, false)
		} else if actorAt.IsSleeping() && actorAt.HasStealableItems() {
			g.StartPickpocket(actorAt)
		} else {
			g.OpenContextMenuFor(actorAt.Position())
		}
		return
	}

	if downedActorAt, exists := g.currentMap().TryGetDownedActorAt(newPos); exists {
		g.openInventoryOf(downedActorAt)
	}
	direction = newPos.Sub(oldPos).ToDirection()
	g.playerMove(oldPos, newPos)
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

		g.ui.PlayCue("world/pickup")

		if !inventory.IsEmpty() {
			g.openInventoryOf(actor)
		}
	}
	g.ui.ShowContainer(actor.Name(), stackRef, transfer)

}
func (g *GameState) afterPlayerMoved(oldPos geometry.Point, wasMapTransition bool) {
	// explore the map
	// print "You see.." message
	if g.currentMap().IsItemAt(g.Player.Position()) && g.config.AutoPickup {
		g.PlayerPickupItem()
	}

	g.currentMap().MoveLightSource(g.playerLightSource, g.Player.Position())
	g.msg(g.GetMapInfoForMovement(g.Player.Position()))
	g.updatePlayerFoVAndApplyExploration()
	g.updateDijkstraMap()

	if g.Player.HasFlag(foundation.FlagCurseTeleportitis) && rand.Intn(100) < 5 {
		g.ui.AddAnimations(OneAnimation(teleportWithAnimation(g, g.Player, g.currentMap().RandomSpawnPosition())))
	}

	// automatic door opening/closing sfx handling
	if !wasMapTransition {
		if door, exists := g.TryGetDoorAt(g.Player.Position()); exists {
			if door.IsClosedButNotLocked() {
				door.PlayOpenSfx()
			}
		}
		if door, exists := g.TryGetDoorAt(oldPos); exists {
			if door.IsClosedButNotLocked() {
				door.PlayCloseSfx()
			}
		}
	}
}

func (g *GameState) updateDijkstraMap() {
	g.playerDijkstraMap = g.currentMap().GetDijkstraMapWithActorsNotBlocking(g.Player.Position(), 1000)
}
