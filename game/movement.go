package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"RogueUI/fov"
	"RogueUI/gridmap"
	"fmt"
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

	if oldPos == newPos {
		return
	}

	if objectAt, exists := g.currentMap().TryGetObjectAt(newPos); exists {
		if door, isDoor := objectAt.(*Door); isDoor && door.IsLocked() && player.HasKey(door.GetLockFlag()) { // THIS IS STILL HACKY; DOOR OBJECT ALSO IMPLEMENTS THIS
			door.Unlock()
			g.msg(foundation.Msg("You unlocked the door using the key"))
			g.ui.PlayCue("world/PICKKEYS")
			return
		}

		if trap, isTrap := objectAt.(*Trap); isTrap && trap.IsArmedOrTriggered() {
			if g.playerStopBecauseOfTrap(trap) {
				return
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

	positionBehind := newPos.Add(direction.ToPoint())

	if !g.currentMap().IsTileWalkable(newPos) &&
		g.currentMap().IsTransparent(newPos) &&
		!g.currentMap().IsObjectAt(newPos) &&
		g.currentMap().IsActorAt(positionBehind) {
		actorBehindCounter := g.currentMap().ActorAt(positionBehind)
		if actorBehindCounter.IsAlive() && actorBehindCounter.HasDialogue() {
			g.OpenContextMenuFor(positionBehind)
			return
		}
	}

	if !g.currentMap().IsTileWalkable(newPos) &&
		g.currentMap().IsCurrentlyMountable(newPos) &&
		(!g.currentMap().IsTileWithFlagAt(oldPos, gridmap.TileFlagMountable) && !g.currentMap().IsTileWithFlagAt(oldPos, gridmap.TileFlagCrawlable)) {
		message := fmt.Sprintf("Do you want to climb onto the %s?", g.currentMap().GetCell(newPos).TileType.Description())
		g.ui.AskForConfirmation("Confirm", message, func(confirmed bool) {
			if confirmed {
				g.msg(foundation.Msg(fmt.Sprintf("You climb onto the %s", g.currentMap().GetCell(newPos).TileType.Description())))
				g.gameFlags.Increment("playerClimbs")
				g.playerMove(newPos)
				g.Player.SetStance(Mounted)
				g.ui.AfterPlayerMoved(foundation.MoveInfo{
					Direction: direction,
					OldPos:    oldPos,
					NewPos:    newPos,
					Mode:      foundation.PlayerMoveModeManual,
				})
			}
		})
		return
	}

	if !g.currentMap().IsTileWalkable(newPos) &&
		g.currentMap().IsCurrentlyCrawlable(newPos) &&
		(!g.currentMap().IsTileWithFlagAt(oldPos, gridmap.TileFlagMountable) && !g.currentMap().IsTileWithFlagAt(oldPos, gridmap.TileFlagCrawlable)) {
		message := fmt.Sprintf("Do you want to crawl under the %s?", g.currentMap().GetCell(newPos).TileType.Description())
		g.ui.AskForConfirmation("Confirm", message, func(confirmed bool) {
			if confirmed {
				g.msg(foundation.Msg(fmt.Sprintf("You crawl under the %s", g.currentMap().GetCell(newPos).TileType.Description())))
				g.gameFlags.Increment("playerCrawls")
				g.playerMove(newPos)
				g.Player.SetStance(Crawling)
				g.ui.AfterPlayerMoved(foundation.MoveInfo{
					Direction: direction,
					OldPos:    oldPos,
					NewPos:    newPos,
					Mode:      foundation.PlayerMoveModeManual,
				})
			}
		})
		return
	}

	if !g.currentMap().Contains(newPos) || (!g.currentMap().IsTileWalkable(newPos) && !g.currentMap().IsCurrentlyMountable(newPos)) {
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
		} else {
			g.OpenContextMenuFor(actorAt.Position())
		}
		return
	}

	if downedActorAt, exists := g.currentMap().TryGetDownedActorAt(newPos); exists {
		g.openInventoryOf(downedActorAt)
	}
	direction = newPos.Sub(oldPos).ToDirection()
	g.playerMove(newPos)
	g.gameFlags.Increment("playerSteps")
	g.ui.AfterPlayerMoved(foundation.MoveInfo{
		Direction: direction,
		OldPos:    oldPos,
		NewPos:    newPos,
		Mode:      foundation.PlayerMoveModeManual,
	})
}

func (g *GameState) playerStopBecauseOfTrap(trap *Trap) bool {
	player := g.Player
	skill := player.GetCharSheet().GetSkill(d100.SkillForTraps)
	if skill < trap.MinimalSkillNeededForDisarm() {
		message := fmt.Sprintf("You need a higher %s skill (%d%%) to disarm this trap.\nDo you want to kick it instead?", d100.SkillForTraps.String(), trap.MinimalSkillNeededForDisarm())
		g.ui.AskForConfirmation("Kick the trap?", message, func(confirmed bool) {
			if confirmed {
				g.ui.ForceUIRedraw()
				g.msg(foundation.Msg("You kicked the trap"))
				g.ui.AddAnimations(trap.Activate())
				g.endPlayerTurn(player.TimeNeededForActions())
			}
		})
		return true
	}
	_, knownTrap := g.knownTraps[trap.GetInternalName()]
	if knownTrap {
		skill = 100
	}
	if trap.IsPlacedByPlayer() {
		skill = 100
	}
	chanceString := fmt.Sprintf("Disarm the trap? (%d%% chance)", skill)
	g.ui.AskForConfirmation("Disarm", chanceString, func(confirmed bool) {
		if confirmed {
			result := d100.SuccessRoll(d100.Percentage(skill), 5)
			if result.Success {
				trap.Disarm()
				if trap.IsPlacedByPlayer() {
					g.msg(foundation.HiLite("You disabled the %s", trap.displayName))
				} else if !knownTrap {
					g.knownTraps[trap.GetInternalName()] = true
					g.msg(foundation.HiLite("You've learned how to disarm %s", trap.displayName))
				} else {
					g.msg(foundation.HiLite("You disarmed the %s", trap.displayName))
				}
				g.endPlayerTurn(player.TimeNeededForActions())
			} else {
				g.ui.ForceUIRedraw()
				g.msg(foundation.Msg("You failed to disarm the trap"))
				g.ui.AddAnimations(trap.Activate())
				g.endPlayerTurn(player.TimeNeededForActions())
			}
		}
	})
	return true
}

func (g *GameState) openInventoryOf(actor *Actor) {
	inventory := actor.GetInventory()

	actorItems := inventory.StackedItemsWithFilter(func(item foundation.Item) bool {
		return !item.HasTag(foundation.TagNoLoot)
	})

	if (inventory.IsEmpty() || len(actorItems) == 0) && !actor.IsAlive() {
		g.msg(foundation.Msg("There is nothing to pick up"))
		return
	}

	rightToLeft := func(itemUI foundation.Item, amount int) {
		if amount > 0 {
			itemStack := itemUI

			stackTransfer(inventory, g.Player.GetInventory(), itemStack, amount)

			g.ui.PlayCue("world/pickup")
		}

		g.openInventoryOf(actor)
	}

	if !actor.IsAlive() {
		g.ui.ShowTakeOnlyContainer(actor.Name(), actorItems, func(uiItem foundation.Item) {
			rightToLeft(uiItem, uiItem.StackSize())
		})
		return
	}

	leftToRight := func(itemUI foundation.Item, amount int) {
		if amount > 0 {
			itemStack := itemUI
			stackTransfer(g.Player.GetInventory(), inventory, itemStack, amount)
			g.ui.PlayCue("world/drop")
		}

		g.openInventoryOf(actor)
	}
	takeAll := func() {
		loot := actor.GetInventory().StackedItemsWithFilter(func(item foundation.Item) bool {
			return !item.HasTag(foundation.TagNoLoot)
		})
		for _, item := range loot {
			stackTransfer(actor.GetInventory(), g.Player.GetInventory(), item, item.StackSize())
		}
		g.openInventoryOf(actor)
	}
	playerItems := g.Player.GetInventory().Items()
	g.ui.ShowGiveAndTakeContainer(g.Player.Name(), playerItems, actor.Name(), actorItems, rightToLeft, leftToRight, takeAll)
}

func (g *GameState) getObservers(mapPos geometry.Point) []*Actor {
	var watchers []*Actor
	for _, actor := range g.currentMap().Actors() {
		if actor.CanSee(mapPos) {
			watchers = append(watchers, actor)
		}
	}
	return watchers
}

func (g *GameState) updateFoVAndDijkstraMap(actor *Actor) {
	actor.DijkstraMap = g.currentMap().GetDijkstraMapWithActorsNotBlocking(actor.Position(), 2000)

	fovRange := g.visionRange

	fovRangeRangeSquared := fovRange * fovRange

	// new fov
	actor.ResetFov()
	fov.ComputeFov([2]int{actor.Position().X, actor.Position().Y}, func(x, y int) bool {
		p := geometry.Point{X: x, Y: y}
		if !g.currentMap().Contains(p) || !g.currentMap().IsTransparent(p) {
			return true
		}
		return geometry.DistanceSquared(p, actor.Position()) > fovRangeRangeSquared
	}, func(x, y int) {
		point := geometry.Point{X: x, Y: y}
		actor.SetVisible(point)
	})

	// old fov

	/*
		actor.FoV.RemoveFromVisibles(func(p geometry.Point) bool {
			return g.currentMap().IsDarknessAt(g.gameTime.Time, p) && actor.Position() != p
		})

	*/

	if actor == g.Player {
		for _, pos := range g.Player.Visibles() {
			g.currentMap().SetExplored(pos)
		}
	}
}
