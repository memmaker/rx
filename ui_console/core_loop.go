package ui_console

import (
	"contractor/foundation"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/geometry"
	"strings"
	"time"
)

func (u *UI) handleMainInput(ev *tcell.EventKey) *tcell.EventKey {
	mod, _, ch := ev.Modifiers(), ev.Key(), ev.Rune()
	if ev.Key() == tcell.KeyCtrlC {
		return ev
	}
	if u.gameIsOver {
		return ev
	}

	u.mapOverlay.ClearAll()
	if u.autoRun && mod == 128 && ev.Key() == tcell.KeyF40 {
		time.Sleep(64 * time.Millisecond)
		u.autoRun = u.game.RunPlayerPath()
		return nil
	}
	if mod == 64 && u.autoRun && strings.ContainsRune("12346789", ch) {
		direction := runeToDirection(ch)
		time.Sleep(64 * time.Millisecond)
		u.autoRun = u.game.RunPlayer(direction, false)
		return nil
	}
	u.autoRun = false

	uiKey := toUIKey(ev)
	playerCommand := u.getCommandForKey(uiKey)
	u.executePlayerCommand(playerCommand)

	return nil
}

func (u *UI) handleMainMouse(event *tcell.EventMouse, action cview.MouseAction) (*tcell.EventMouse, cview.MouseAction) {
	if u.isModalOpen() {
		return event, action
	}

	if event == nil || u.gameIsOver {
		return nil, action
	}

	newX, newY := event.Position()
	mousePos := geometry.Point{X: newX, Y: newY}
	if newX != u.currentMouseX || newY != u.currentMouseY {
		u.currentMouseX = newX
		u.currentMouseY = newY
		if !u.autoRun && action == cview.MouseMove {
			if u.currentMouseX >= u.settings.MapWidth {
				// hovered over right panel
				u.onRightPanelHovered(mousePos)
				return nil, -1
			}
			mapPos := u.ScreenToMap(mousePos)
			mapInfo := u.game.GetMapInfo(mapPos)

			hoveredActor := u.game.ActorAt(mapPos)
			if hoveredActor != u.hoveredActor {
				u.hoveredActor = hoveredActor
				u.UpdateVisibleActors()
			}

			if mapInfo.IsEmpty() {
				u.application.QueueUpdateDraw(u.UpdateLogWindow)
			} else {
				u.Print(mapInfo)
			}
			return nil, -1
		}
	}
	if u.onMoreKey != nil {
		return nil, action
	}
	mapPos := u.ScreenToMap(geometry.Point{X: newX, Y: newY})
	isModified := event.Modifiers() != 0

	_, screenHeight := u.application.GetScreenSize()

	if action == cview.MouseLeftClick {
		u.autoRun = false
		if u.currentMouseX >= u.settings.MapWidth {
			// clicked on right panel
			u.onRightPanelClicked(mousePos, false, isModified)
		} else if u.currentMouseY == screenHeight-1 {
			// clicked status bar
			if isModified {
				u.game.PlayerReloadWeapon()
			} else {
				u.game.PlayerRangedAttack()
			}
		} else {
			if isModified {
				u.game.DebugClickAt(mapPos)
			} else {
				u.game.PlayerInteractAtPosition(mapPos)
			}
			//u.game.OpenContextMenuFor(mapPos)
		}
		return nil, -1
	} else if action == cview.MouseRightClick {
		u.autoRun = false
		if u.currentMouseX >= u.settings.MapWidth {
			// clicked on right panel
			u.onRightPanelClicked(mousePos, true, isModified)
		} else if u.currentMouseY == screenHeight-1 {
			// clicked status bar
			u.game.CycleTargetMode()
		} else {
			actorAt := u.game.ActorAt(mapPos)
			if actorAt != nil {
				if mapPos == u.game.GetPlayerPosition() {
					u.openCharSheet()
				} else {
					u.ShowMonsterInfo(actorAt)
				}
			}
		}
		return nil, -1
	}
	return event, action
}

// executePlayerCommand Is how the game is driven by the player.
// These are bound to function calls on the game in setupCommandTable().
func (u *UI) executePlayerCommand(command string) {
	if command == "" {
		return
	}
	if action, ok := u.commandTable[command]; ok {
		action()
	} else {
		u.Print(foundation.Msg(fmt.Sprintf("Unknown command: %s", command)))
	}
}

func (u *UI) AfterPlayerMoved(moveInfo foundation.MoveInfo) {
	if moveInfo.Mode == foundation.PlayerMoveModePath {
		u.autoRun = true
		u.application.QueueEvent(tcell.NewEventKey(tcell.KeyF40, ' ', 128))
	} else if moveInfo.Mode == foundation.PlayerMoveModeRun && u.autoRun {
		u.application.QueueEvent(tcell.NewEventKey(tcell.KeyRune, directionToRune(moveInfo.Direction), 64))
	}
	u.updateMapScrollPosition()
}

func (u *UI) AddAnimations(animations []foundation.Animation) {
	if !u.game.MapContainsPlayer() {
		return
	}
	for _, animation := range animations {
		if textAnim, isTextAnim := animation.(TextAnimation); isTextAnim && textAnim != nil {
			u.animator.AddAnimation(textAnim)
		}
	}
}

func (u *UI) AnimatePending() bool {
	if !u.settings.AnimationsEnabled {
		return true
	}

	return u.updateUntilDone()
}

func (u *UI) SkipAnimations() {
	u.animator.CancelAll()
}

func (u *UI) updateUntilDone() bool {
	duration := 2 * time.Millisecond

	//
	screen := u.application.GetScreen()
	//u.application.Unlock()

	u.isAnimationFrame = true
	var breakingKey *tcell.EventKey
outerLoop:
	for len(u.animator.runningAnimations) > 0 {
		u.application.Lock()
		u.mapWindow.Draw(screen)
		screen.Show()
		u.application.Unlock()

		var waited time.Duration
		for waited < u.settings.AnimationDelay {
			if screen.HasPendingEvent() {
				ev := screen.PollEvent()
				if keyEvent, ok := ev.(*tcell.EventKey); ok {
					breakingKey = keyEvent
					u.animator.CancelAll()
					break outerLoop
				}
			}

			time.Sleep(duration)
			waited += duration
		}

		shouldMapFrameBeUpdated := u.animator.Tick()

		if shouldMapFrameBeUpdated {
			u.updateLastFrame()
		}
	}
	u.isAnimationFrame = false

	if breakingKey != nil {
		u.application.QueueEvent(breakingKey)
		return true
	}

	u.application.Lock()
	u.mapWindow.Draw(screen)
	screen.Show()
	u.application.Unlock()

	return false
}
