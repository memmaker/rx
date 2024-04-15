package console

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"
	"strings"
)

func (u *UI) SelectTarget(origin geometry.Point, onSelected func(targetPos geometry.Point)) {
	u.pages.SetCurrentPanel("main")
	message := "<Direction> = Fire, <Space> = Advanced Targeting"
	u.Print(foundation.Msg(message))
	u.mapWindow.SetInputCapture(u.handleDirectionalTargetingInput(origin, true, onSelected))
}

func (u *UI) SelectDirection(origin geometry.Point, onSelected func(direction geometry.CompassDirection)) {
	u.pages.SetCurrentPanel("main")
	message := "Direction?"
	u.Print(foundation.Msg(message))
	u.mapWindow.SetInputCapture(u.handleDirectionalTargetingInput(origin, false, func(targetPos geometry.Point) {
		directionVector := targetPos.Sub(origin)
		direction := directionVector.AsSigns().ToDirection()
		onSelected(direction)
	}))
}

func (u *UI) LookTargeting() {
	u.onTargetUpdated = func(targetPos geometry.Point) {
		mapInfo := u.game.GetMapInfo(targetPos)
		u.Print(mapInfo)
	}
	u.beginAdvancedTargeting(func(targetPos geometry.Point) {})
}
func (u *UI) handleDirectionalTargetingInput(origin geometry.Point, allowAdvancedTargeting bool, onSelected func(targetPos geometry.Point)) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		_, _, ch := ev.Modifiers(), ev.Key(), ev.Rune()
		if ev.Key() == tcell.KeyCtrlC {
			return ev
		}
		if ev.Key() == tcell.KeyEscape {
			u.cancelTargeting()
			u.UpdateLogWindow()
			return nil
		}
		handled := false
		if strings.ContainsRune("wasd12346789", ch) {
			onSelected(origin.Add(runeToDirection(ch).ToPoint().Mul(10)))
			handled = true
		} else if allowAdvancedTargeting && ch == ' ' {
			u.beginAdvancedTargeting(onSelected)
		}

		if handled {
			u.mapWindow.SetInputCapture(u.handleMainInput)
		}
		return nil
	}
}

func (u *UI) handleAdvancedTargetingInput(listOfVisibleEnemies []foundation.ActorForUI, onSelected func(targetPos geometry.Point)) func(ev *tcell.EventKey) *tcell.EventKey {
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
		mod, key, ch := ev.Modifiers(), ev.Key(), ev.Rune()
		if ev.Key() == tcell.KeyCtrlC {
			return ev
		}
		if key == tcell.KeyEscape {
			u.cancelTargeting()
			u.UpdateLogWindow()
			return nil
		} else if key == tcell.KeyTab {
			if mod == tcell.ModShift {
				choosePreviousEnemy()
			} else {
				chooseNextEnemy()
			}
		} else if strings.ContainsRune("wasd12346789", ch) {
			newTarget := u.targetPos.Add(runeToDirection(ch).ToPoint())
			u.updateTarget(newTarget)
		} else if ch == ' ' {
			u.cancelTargeting()
			u.SelectTarget(u.game.GetPlayerPosition(), onSelected)
		} else if ch == 'e' {
			u.onTargetSelected(onSelected)
		}

		return nil
	}
}
func (u *UI) onTargetSelected(onSelected func(targetPos geometry.Point)) {
	u.cancelTargeting()
	onSelected(u.targetPos)
}

func (u *UI) cancelTargeting() {
	u.mapWindow.SetInputCapture(u.handleMainInput)
	u.application.SetMouseCapture(nil)
	u.state = StateNormal
	clear(u.targetingTiles)
	u.Print(foundation.NoMsg())
}

func (u *UI) beginAdvancedTargeting(onSelected func(targetPos geometry.Point)) {
	u.state = StateTargeting
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
				u.updateTarget(mapPos)
				return nil, action
			}
		} else if action == cview.MouseLeftClick {
			newX, newY := event.Position()
			mapPos := u.ScreenToMap(geometry.Point{X: newX, Y: newY})
			u.updateTarget(mapPos)
			u.onTargetSelected(onSelected)
			return nil, action
		}
		return event, action
	})
}

func (u *UI) updateTarget(targetPos geometry.Point) {
	origin := u.game.GetPlayerPosition()
	clear(u.targetingTiles)
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
