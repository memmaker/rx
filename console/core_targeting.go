package console

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/geometry"
	"strconv"
	"strings"
)

func (u *UI) SelectDirection(onSelected func(direction geometry.CompassDirection)) {
	u.pages.SetCurrentPanel("main")
	u.Print(foundation.Msg("Direction?"))
	u.mapWindow.SetInputCapture(u.handleDirectionalTargetingInput(onSelected))
}

func (u *UI) SelectBodyPart(previousAim d100.BodyPart, onSelected func(victim foundation.ActorForUI, hitZone d100.BodyPart)) {
	// we want advanced targeting but only on tiles with actors
	// we also want to show the body part selection whenever the current target is updated
	// when the user has confirmed a body part of the currently selected target, we're done
	u.onTargetUpdated = func(targetPos geometry.Point) {
		actorAt := u.game.ActorAt(targetPos)
		if actorAt != nil {
			if u.hoveredActor != actorAt {
				u.hoveredActor = actorAt
				u.UpdateVisibleActors()
			}
			hitZones := u.game.GetBodyPartsAndHitChances(actorAt)
			var text string
			for i, hitZone := range hitZones {
				keys := u.GetKeysForCommandAsString(KeyLayerAdvancedTargeting, fmt.Sprintf("target_confirm_body_part_%d", i))
				text += fmt.Sprintf("%s - %s (%d%%)\n", keys, hitZone.Item1, hitZone.Item3)
			}
			u.Print(foundation.Msg("Select body part"))
			u.rightPanel.SetText(text)
		}
	}
	u.state = StateTargetingBodyPart
	u.beginTargeting(func(targetPos geometry.Point, hitZone int) {
		u.lastTarget = [2]geometry.Point{u.game.GetPlayerPosition(), targetPos}
		actorAt := u.game.ActorAt(targetPos)
		u.OpenAimedShotPicker(actorAt, previousAim, onSelected)
	})
}

func (u *UI) OpenAimedShotPicker(actorAt foundation.ActorForUI, previousAim d100.BodyPart, onSelected func(victim foundation.ActorForUI, hitZone d100.BodyPart)) {
	if actorAt != nil {
		var items []foundation.MenuItem
		for i, tuple := range u.game.GetBodyPartsAndHitChances(actorAt) {
			statusString := "ok"
			if tuple.Item2 { // iscrippled
				statusString = "CR"
			}
			hitString := fmt.Sprintf("<%s> %s (%d%%)", statusString, tuple.Item1, tuple.Item3)
			items = append(items, foundation.MenuItem{
				Name: hitString,
				Action: func() {
					onSelected(actorAt, actorAt.GetBodyPart(i))
				},
				CloseMenus: true,
			})
		}
		menu := u.openSimpleMenu(items, nil)
		menu.SetCurrentItem(actorAt.GetBodyPartIndex(previousAim))
	}
}

func (u *UI) SelectTarget(getAttackInfo func(target foundation.ActorForUI) foundation.AttackInfo, onSelected func(targetPos geometry.Point)) {
	u.state = StateTargeting
	u.onTargetUpdated = func(targetPos geometry.Point) {
		actorAt := u.game.ActorAt(targetPos)
		if actorAt != nil {
			if u.hoveredActor != actorAt {
				u.hoveredActor = actorAt
				u.UpdateVisibleActors()
			}
			attackInfo := getAttackInfo(actorAt)
			hitChance := attackInfo.SuccessChance()
			cthString := fmt.Sprintf("%d%%", hitChance)
			placeBelow := u.game.GetPlayerPosition().Y < targetPos.Y
			if placeBelow {
				u.mapOverlay.AddBelow(actorAt.Position(), cthString)
			} else {
				u.mapOverlay.AddAbove(actorAt.Position(), cthString)
			}

			//u.Print(foundation.Msg("Select body part"))
			cthModString := attackInfo.Modifiers().ChanceToHitMods.String()

			cthText := fmt.Sprintf("Attack on %s\n%s: %d%%\n%s", actorAt.Name(), attackInfo.Skill().String(), attackInfo.SkillBase(), cthModString)

			var damageText string
			if len(attackInfo.Modifiers().DamageMods) > 0 {
				baseDamage := fmt.Sprintf("Base damage: %s", attackInfo.DamageBase().ShortString())
				damageModString := attackInfo.Modifiers().DamageMods.String()
				finalDamage := fmt.Sprintf("Damage Range: %s", attackInfo.DamageWithMods().ShortString())
				damageText = fmt.Sprintf("%s\n%s\n%s", baseDamage, damageModString, finalDamage)
			} else {
				damageText = fmt.Sprintf("Damage Range: %s", attackInfo.DamageBase().ShortString())
			}
			totalCth := attackInfo.SuccessChance()
			totalCthString := fmt.Sprintf("Total CTH: %d%%", totalCth)
			fullText := fmt.Sprintf("%s\n%s\n\n%s", cthText, totalCthString, damageText)

			u.rightPanel.SetText(fullText)
		}
	}
	u.beginTargeting(func(targetPos geometry.Point, hitZone int) {
		u.lastTarget = [2]geometry.Point{u.game.GetPlayerPosition(), targetPos}
		onSelected(targetPos)
	})
}

func (u *UI) beginTargeting(onSelected func(targetPos geometry.Point, hitZone int)) {
	listOfVisibleActors := u.game.GetVisibleActors()
	preselected := u.game.GetPlayerPosition()
	if u.lastTarget != [2]geometry.Point{} && u.game.GetPlayerPosition() == u.lastTarget[0] {
		preselected = u.lastTarget[1]
	} else if len(listOfVisibleActors) > 0 {
		preselected = listOfVisibleActors[0].Position()
	}
	u.updateTarget(preselected)
	u.mapWindow.SetInputCapture(u.handleAdvancedTargetingInput(listOfVisibleActors, onSelected))
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
		} else if action == cview.MouseRightClick {
			u.cancelTargeting()
			u.UpdateLogWindow()
			return nil, action
		}
		return event, action
	})
}

func (u *UI) LookTargeting() {
	u.state = StateLookTargeting
	u.onTargetUpdated = func(targetPos geometry.Point) {
		mapInfo := u.game.GetMapInfo(targetPos)
		u.Print(mapInfo)
		actorAt := u.game.ActorAt(targetPos)
		if u.hoveredActor != actorAt {
			u.hoveredActor = actorAt
			u.UpdateVisibleActors()
		}
	}

	u.beginTargeting(func(targetPos geometry.Point, hitZone int) {
		u.lastTarget = [2]geometry.Point{u.game.GetPlayerPosition(), targetPos}
		u.game.PlayerInteractAtPosition(targetPos)
	})
}
func (u *UI) handleDirectionalTargetingInput(onSelected func(targetDir geometry.CompassDirection)) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		//_, _, ch := ev.ModList(), ev.Key(), ev.Rune()
		if ev.Key() == tcell.KeyCtrlC {
			return ev
		}
		command := u.getDirectionalTargetingCommandForKey(toUIKey(ev))

		if command == "target_cancel" || command == "wait" {
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
			u.UpdateLogWindow()
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
		//mod, key, ch := ev.ModList(), ev.Key(), ev.Rune()
		if ev.Key() == tcell.KeyCtrlC {
			return ev
		}
		uiKey := toUIKey(ev)
		targetingCommand := u.getAdvancedTargetingCommandForKey(uiKey)
		// Damn, this is a second layer of keymaps..or is it? probably is
		if targetingCommand == "target_cancel" {
			u.cancelTargeting()
			u.UpdateLogWindow()
			return nil
		} else if targetingCommand == "target_next" {
			chooseNextEnemy()
		} else if targetingCommand == "target_previous" {
			choosePreviousEnemy()
		} else if direction, isDirectionalCommand := directionFromCommand(targetingCommand); isDirectionalCommand {
			newTarget := u.targetPos.Add(direction.ToPoint())
			u.updateTarget(newTarget)
		} else if targetingCommand == "target_confirm" {
			u.cancelTargeting()
			onSelected(u.targetPos, 0)
		} else if strings.HasPrefix(targetingCommand, "target_confirm_body_part_") {
			index, _ := strconv.Atoi(targetingCommand[len("target_confirm_body_part_"):])
			u.cancelTargeting()
			onSelected(u.targetPos, index)
		}

		return nil
	}
}

func (u *UI) cancelTargeting() {
	u.mapWindow.SetInputCapture(u.handleMainInput)
	u.application.SetMouseCapture(u.handleMainMouse)
	u.state = StateNormal
	u.UpdateInventory()
	u.Print(foundation.NoMsg())
	u.mapOverlay.ClearAll()
	clear(u.targetingTiles)
}

func (u *UI) updateTarget(targetPos geometry.Point) {
	origin := u.game.GetPlayerPosition()

	clear(u.targetingTiles)

	if origin == targetPos {
		u.targetingTiles[targetPos] = 'X'
		u.targetPos = targetPos
		return
	}

	line := u.game.GetFlightPath(origin, targetPos)

	if len(line) == 0 {
		return
	}

	line = line[1:] // remove origin

	if len(line) == 0 {
		return
	}

	targetPos = line[len(line)-1]

	if len(line) == 1 {
		onePointLine := line[0]
		if onePointLine == origin {
			u.targetPos = targetPos
			return
		}
		dir := onePointLine.Sub(origin)
		u.targetingTiles[onePointLine] = charFromDirection(dir)
	} else {
		for i, point := range line {
			var directionOfLine geometry.Point
			if i == 0 {
				directionOfLine = line[i+1].Sub(line[i])
			} else {
				directionOfLine = line[i].Sub(line[i-1])
			}
			u.targetingTiles[point] = charFromDirection(directionOfLine)
		}
	}

	u.targetPos = targetPos
	if u.onTargetUpdated != nil {
		u.onTargetUpdated(targetPos)
	}
}
