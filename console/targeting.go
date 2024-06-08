package console

import (
	"RogueUI/cview"
	"RogueUI/foundation"
	"RogueUI/geometry"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"strconv"
	"strings"
)

func (u *UI) SelectDirection(onSelected func(direction geometry.CompassDirection)) {
	u.pages.SetCurrentPanel("main")
	message := "Direction?"
	u.Print(foundation.Msg(message))
	u.mapWindow.SetInputCapture(u.handleDirectionalTargetingInput(onSelected))
}

func (u *UI) SelectBodyPart(onSelected func(victim foundation.ActorForUI, part int)) {
	// we want advanced targeting but only on tiles with actors
	// we also want to show the body part selection whenever the current target is updated
	// when the user has confirmed a body part of the currently selected target, we're done
	u.onTargetUpdated = func(targetPos geometry.Point) {
		actorAt := u.game.ActorAt(targetPos)
		if actorAt != nil {
			hitZones := u.game.GetBodyPartsAndHitChances(actorAt)
			var text string
			for i, hitZone := range hitZones {
				keys := u.GetKeysForCommandAsString(KeyLayerAdvancedTargeting, fmt.Sprintf("target_confirm_body_part_%d", i))
				text += fmt.Sprintf("%s - %s (%d%%)\n", keys, hitZone.Item1, hitZone.Item2)
			}
			u.Print(foundation.Msg("Select body part"))
			u.rightPanel.SetText(text)
		}
	}
	u.state = StateTargetingBodyPart
	u.beginTargeting(func(targetPos geometry.Point, hitZone int) {
		actorAt := u.game.ActorAt(targetPos)
		if actorAt != nil {
			onSelected(actorAt, hitZone)
		}
	})
}

func (u *UI) SelectTarget(onSelected func(targetPos geometry.Point, hitZone int)) {
	u.state = StateTargeting
	u.onTargetUpdated = func(targetPos geometry.Point) {
		actorAt := u.game.ActorAt(targetPos)
		if actorAt != nil {
			labelBelow := u.game.GetPlayerPosition().Y < targetPos.Y
			yPos := -1
			if labelBelow {
				yPos = 1
			}
			hitChance := u.game.GetRangedHitChance(actorAt)
			labelPos := targetPos.Add(geometry.Point{X: -1, Y: yPos})
			u.mapOverlay.Print(labelPos.X, labelPos.Y, fmt.Sprintf("%d%%", hitChance))
		}
	}
	u.beginTargeting(onSelected)
}

func (u *UI) beginTargeting(onSelected func(targetPos geometry.Point, hitZone int)) {
	listOfVisibleEnemies := u.game.GetVisibleEnemies()
	preselected := u.game.GetPlayerPosition()
	if len(listOfVisibleEnemies) > 0 {
		preselected = listOfVisibleEnemies[0].Position()
	}
	u.updateTarget(preselected)
	u.mapWindow.SetInputCapture(u.handleAdvancedTargetingInput(listOfVisibleEnemies, onSelected))
	u.application.SetMouseCapture(func(event *tcell.EventMouse, action cview.MouseAction) (*tcell.EventMouse, cview.MouseAction) {
		if action == cview.MouseMove {
			newX, newY := event.Position()
			if newX != u.currentMouseX || newY != u.currentMouseY {
				u.currentMouseX = newX
				u.currentMouseY = newY
				mapPos := u.ScreenToMap(geometry.Point{X: newX, Y: newY})
				if u.state != StateTargetingBodyPart || u.game.ActorAt(mapPos) != nil {
					u.updateTarget(mapPos)
					return nil, action
				}
			}
		} else if action == cview.MouseLeftClick {
			newX, newY := event.Position()
			mapPos := u.ScreenToMap(geometry.Point{X: newX, Y: newY})
			if u.state != StateTargetingBodyPart || u.game.ActorAt(mapPos) != nil {
				u.updateTarget(mapPos)
				u.cancelTargeting()
				onSelected(u.targetPos, 0)
				return nil, action
			}
		}
		return event, action
	})
}

func (u *UI) LookTargeting() {
	u.state = StateLookTargeting
	u.onTargetUpdated = func(targetPos geometry.Point) {
		mapInfo := u.game.GetMapInfo(targetPos)
		u.Print(mapInfo)
	}

	u.beginTargeting(func(targetPos geometry.Point, hitZone int) {
		actorAt := u.game.ActorAt(targetPos)
		if actorAt != nil {
			u.ShowMonsterInfo(actorAt)
		}
	})
}
func (u *UI) handleDirectionalTargetingInput(onSelected func(targetDir geometry.CompassDirection)) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		//_, _, ch := ev.Modifiers(), ev.Key(), ev.Rune()
		if ev.Key() == tcell.KeyCtrlC {
			return ev
		}
		command := u.getDirectionalTargetingCommandForKey(toUIKey(ev))

		if command == "target_cancel" {
			u.cancelTargeting()
			u.UpdateLogWindow()
			return nil
		}
		handled := false
		if direction, isDirectionalCommand := directionFromCommand(command); isDirectionalCommand {
			onSelected(direction)
			handled = true
		}

		if handled {
			u.mapWindow.SetInputCapture(u.handleMainInput)
		}
		return nil
	}
}

func (u *UI) handleAdvancedTargetingInput(listOfVisibleEnemies []foundation.ActorForUI, onSelected func(targetPos geometry.Point, hitZone int)) func(ev *tcell.EventKey) *tcell.EventKey {
	enemyIndex := 0
	chooseNextEnemy := func() {
		if len(listOfVisibleEnemies) == 0 {
			return
		}
		enemyIndex = (enemyIndex + 1) % len(listOfVisibleEnemies)
		u.updateTarget(listOfVisibleEnemies[enemyIndex].Position())
	}
	choosePreviousEnemy := func() {
		if len(listOfVisibleEnemies) == 0 {
			return
		}
		enemyIndex = (enemyIndex - 1 + len(listOfVisibleEnemies)) % len(listOfVisibleEnemies)
		u.updateTarget(listOfVisibleEnemies[enemyIndex].Position())
	}
	return func(ev *tcell.EventKey) *tcell.EventKey {
		//mod, key, ch := ev.Modifiers(), ev.Key(), ev.Rune()
		if ev.Key() == tcell.KeyCtrlC {
			return ev
		}

		command := u.getAdvancedTargetingCommandForKey(toUIKey(ev))
		// Damn, this is a second layer of keymaps..or is it? probably is
		if command == "target_cancel" {
			u.cancelTargeting()
			u.UpdateLogWindow()
			return nil
		} else if command == "target_next" {
			chooseNextEnemy()
		} else if command == "target_previous" {
			choosePreviousEnemy()
		} else if direction, isDirectionalCommand := directionFromCommand(command); isDirectionalCommand {
			newTarget := u.targetPos.Add(direction.ToPoint())
			u.updateTarget(newTarget)
		} else if command == "target_confirm" {
			u.cancelTargeting()
			onSelected(u.targetPos, 0)
		} else if strings.HasPrefix(command, "target_confirm_body_part_") {
			index, _ := strconv.Atoi(command[len("target_confirm_body_part_"):])
			u.cancelTargeting()
			onSelected(u.targetPos, index)
		}

		return nil
	}
}

func (u *UI) cancelTargeting() {
	u.mapWindow.SetInputCapture(u.handleMainInput)
	u.application.SetMouseCapture(nil)
	u.state = StateNormal
	u.mapOverlay.ClearAll()
	clear(u.targetingTiles)
	u.Print(foundation.NoMsg())
}

func (u *UI) updateTarget(targetPos geometry.Point) {
	origin := u.game.GetPlayerPosition()
	clear(u.targetingTiles)
	if origin == targetPos {
		return
	}
	line := geometry.BresenhamLine(origin, targetPos, func(x, y int) bool {
		curPos := geometry.Point{X: x, Y: y}
		if origin == curPos {
			return true
		}
		return !u.game.IsSomethingBlockingTargetingAtLoc(curPos)
	})
	if len(line) > 1 {
		line = line[1:]
		targetPos = line[len(line)-1]
	}
	for _, point := range line {
		u.targetingTiles[point] = true
	}
	u.targetPos = targetPos
	if u.onTargetUpdated != nil {
		u.onTargetUpdated(targetPos)
	}
}
