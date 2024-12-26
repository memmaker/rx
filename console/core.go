package console

import (
	"contractor/foundation"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/audio"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
)

type UI struct {
	settings *foundation.Configuration
	game     foundation.GameForUI

	lifeCycle UILifeCycler

	audioPlayer *audio.Player

	uiTheme Theme

	mapOverlay *Overlay

	mapScroll geometry.Point

	mainGrid        *cview.Grid
	lowerRightPanel *cview.TextView
	messageLabel    *cview.TextView
	statusBar       *cview.TextView
	rightPanel      *cview.TextView
	pages           *cview.Panels
	application     *cview.Application
	mapWindow       *cview.Box

	currentMouseX  int
	currentMouseY  int
	state          UIState
	targetingTiles map[geometry.Point]rune

	animator  *Animator
	targetPos geometry.Point

	listTable map[string]*cview.List

	gameIsOver      bool
	autoRun         bool
	onTargetUpdated func(targetPos geometry.Point)
	showCursor      bool
	cursorStyle     tcell.CursorStyle
	tooSmall        bool
	gamma           float64
	commandTable    map[string]func()
	keyTable        map[KeyLayer]map[UIKey]string

	lastFrameIcons     map[geometry.Point]rune
	lastFrameStyle     map[geometry.Point]tcell.Style
	isAnimationFrame   bool
	lastHudStats       foundation.HudValueMap
	dialogueText       *cview.TextView
	dialogueOptions    *cview.List
	dialogueIsTerminal bool
	shouldPlayIntro    bool

	stateOfIntro IntroState
	focusStack   []FocusInfo
	tileColors   map[geometry.Point]fxtools.HDRColor
	lastTarget   [2]geometry.Point
	graphicsMode bool

	onAnyKey func()
}

func (u *UI) IndicateConversationStartByNPC(partner foundation.ChatterSource, done func()) {
	labelText := fmt.Sprintf("<%s is addressing you>", partner.Name())
	u.mapOverlay.TryAddOverlay(partner.Position(), labelText, u.GetMapWindowGridSize(), u.game.IsSomethingInterestingAtLoc)
	u.Print(foundation.HiLite("%s is addressing you. %s", partner.Name(), cview.Escape("[MORE]")))

	u.onAnyKey = func() {
		u.mapOverlay.ClearAll()
		if done != nil {
			done()
		}
	}
}

func (u *UI) SetSneakOverlay(overlay map[geometry.Point]fxtools.HDRColor) {
	u.tileColors = overlay
}

func (u *UI) isMapScrolling() bool {
	mapSize := u.game.GetMapSize()
	return mapSize.X > u.settings.MapWidth || mapSize.Y > u.settings.MapHeight
}

func (u *UI) SetColors(palette textiles.ColorPalette, colors map[foundation.ItemCategory]color.RGBA) {
	u.uiTheme = NewUIThemeFromDataDir(u.settings.DataRootDir, palette, colors)
	u.applyStylingToUI()
}

func (u *UI) SetShowCursor(show bool) {
	u.showCursor = show
	screen := u.application.GetScreen()
	if show {
		screen.SetCursorStyle(u.cursorStyle)
	} else {
		screen.HideCursor()
	}
}

func (u *UI) GetMapWindowGridSize() geometry.Point {
	_, _, w, h := u.mapWindow.GetInnerRect()
	return geometry.Point{X: w, Y: h}
}

func (u *UI) isPlayerHallucinating() bool {
	flags := u.game.GetHudFlags()
	_, isHallucinating := flags[foundation.FlagHallucinating]
	return isHallucinating
}

func (u *UI) setColoredText(view *cview.TextView, text string) {
	view.SetDynamicColors(true)
	view.SetText(text)
}

// Print prints a message to the screen.
// Should only be called by the game
func (u *UI) Print(message foundation.HiLiteString) {
	if message.IsEmpty() {
		return
	}
	u.application.QueueUpdateDraw(func() {
		textColor := u.uiTheme.GetUIColor(UIColorUIForeground)
		hiLiteColor := u.uiTheme.GetUIColor(UIColorTextForegroundHighlighted)
		u.setColoredText(u.messageLabel, ToColoredText(message, 1, textColor, hiLiteColor))
	})
}

func (u *UI) ChooseDirectionForRun() {
	u.SelectDirection(func(direction geometry.CompassDirection) {
		u.startAutoRun(direction)
	})
}

func (u *UI) GenericInteraction() {
	u.SelectDirection(func(direction geometry.CompassDirection) {
		u.game.PlayerInteractInDirection(direction)
	})
}

func (u *UI) startAutoRun(direction geometry.CompassDirection) {
	u.autoRun = true
	u.game.RunPlayer(direction, true)
}

func (u *UI) isRightPanelWidthAtLeast(width int) bool {
	panelWidth := u.getRightPanelWidth()
	return panelWidth >= width
}

func (u *UI) getRightPanelWidth() int {
	w, _ := u.application.GetScreenSize()
	wNeeded, _ := u.settings.GetMinTerminalSize()
	panelWidth := w - wNeeded
	return panelWidth
}

func (u *UI) colorIfDiff(statStr string, stat foundation.HudValue, currentValue interface{}) string {
	lastValue, ok := u.lastHudStats[stat]
	if !ok {
		return statStr
	}
	if lastValue == currentValue {
		return statStr
	}
	hiCode := textiles.RGBAToFgColorCode(u.uiTheme.GetColorByName("Yellow_3"))
	return fmt.Sprintf("%s%s[-]", hiCode, statStr)
}

func (u *UI) ScreenToMap(point geometry.Point) geometry.Point {
	x, y, _, _ := u.mapWindow.GetInnerRect()
	offsetX, offsetY := u.mapScroll.X, u.mapScroll.Y
	mapX := point.X - x + offsetX
	mapY := point.Y - y + offsetY
	return geometry.Point{X: mapX, Y: mapY}
}

func (u *UI) TryAddChatter(source foundation.ChatterSource, text string) bool {
	windowSize := u.GetMapWindowGridSize()
	if u.mapOverlay.TryAddOverlayColored(source.Position(), text, source.GetIcon().Fg, windowSize, u.game.IsSomethingInterestingAtLoc) {
		return true
	}
	return false
}

func (u *UI) onTerminalResized(width int, height int) {
	if u.mainGrid == nil {
		return
	}

	tSizeX, tSizeY := u.settings.GetMinTerminalSize()

	if height <= tSizeY {
		u.mainGrid.SetRows(1, 0, 1)
		u.messageLabel.SetScrollable(false)
		u.messageLabel.SetScrollBarVisibility(cview.ScrollBarNever)
	} else if height == tSizeY+1 {
		u.mainGrid.SetRows(1, 0, 2)
		u.messageLabel.SetScrollable(false)
		u.messageLabel.SetScrollBarVisibility(cview.ScrollBarNever)
	} else if height > tSizeY+1 {
		additionalHeight := height - tSizeY - 1
		u.mainGrid.SetRows(0, 1+additionalHeight, 2)
		u.messageLabel.SetScrollable(true)
		u.messageLabel.SetScrollBarVisibility(cview.ScrollBarAuto)
	}
	u.pages.SetRect(0, 0, width, height)
	if width < tSizeX || height < tSizeY {
		u.tooSmall = true
		view := cview.NewTextView()
		view.SetText(fmt.Sprintf("Min. terminal size is %dx%d", tSizeX, tSizeY))
		u.pages.AddPanel("tooSmall", view, true, true)
	} else if u.tooSmall {
		u.pages.HidePanel("tooSmall")
		u.tooSmall = false
	}
	u.application.QueueUpdateDraw(func() {
		u.UpdateLogWindow()
		u.UpdateInventory()
		u.UpdateStats()
	})
}

func (u *UI) onRightPanelClicked(clickPos geometry.Point, isRightClick bool, modified bool) {
	itemIndex := clickPos.Y - 1

	inv := u.game.GetInventoryForUI()

	if itemIndex < 0 || itemIndex >= len(inv) {
		return
	}

	item := inv[itemIndex]

	if isRightClick {
		u.game.PlayerExamineItem(item)
	} else {
		if modified {
			u.dropItemWithAmountSelection(item, nil)
		} else {
			if item.IsEquippable() {
				u.game.EquipToggle(item)
			} else {
				u.game.PlayerApplyItem(item)
			}
		}
	}

}
func (u *UI) isModalOpen() bool {
	_, front := u.pages.GetFrontPanel()
	return front != u.mainGrid
}

func (u *UI) saveGamesExist() bool {
	return fxtools.DirExists(u.settings.SaveGameDir) && fxtools.DirHasSubDirs(u.settings.SaveGameDir)
}

func (u *UI) getSingleLineStatus(statusValues foundation.HudValueMap, flags map[foundation.ActorFlag]int, multiLine bool, equippedItem string) string {
	armorStr := statusValues.GetString(foundation.HudArmorString)
	armorStr = u.colorIfDiff(armorStr, foundation.HudArmorString, armorStr)

	var statusStr string
	if !multiLine {
		hp := statusValues.GetInt(foundation.HudHitPoints)
		hpMax := statusValues.GetInt(foundation.HudHitPointsMax)
		hpValString := fmt.Sprintf("%d/%d", hp, hpMax)
		hpStr := fmt.Sprintf("HP: %-7s", hpValString)
		hpStr = u.colorIfDiff(hpStr, foundation.HudHitPoints, hp)

		fatigueCurrent := statusValues.GetInt(foundation.HudActionPoints)
		fatigueMax := statusValues.GetInt(foundation.HudActionPointsMax)
		fpValString := fmt.Sprintf("%d/%d", fatigueCurrent, fatigueMax)
		fpStr := fmt.Sprintf("FP: %-7s", fpValString)
		fpStr = u.colorIfDiff(fpStr, foundation.HudActionPoints, fatigueCurrent)

		flagString := PlayerFlagStringShort(flags)

		statusStr = fmt.Sprintf("%s %s %s %s %s", hpStr, fpStr, armorStr, equippedItem, flagString)
	} else {
		statusStr = fmt.Sprintf("%s %s", armorStr, equippedItem)
	}

	width, _ := u.application.GetScreenSize()
	statusStr = expandToWidth(statusStr, width)
	return statusStr
}

func (u *UI) PlayMusic(fileName string) {
	if !u.settings.AudioEnabled || !u.settings.MusicEnabled {
		return
	}
	u.audioPlayer.StopAll()
	u.audioPlayer.StreamLoop(fileName)
}
func (u *UI) PlayCue(cueName string) {
	if !u.settings.AudioEnabled || !u.settings.SoundEffectsEnabled || !u.game.MapContainsPlayer() {
		return
	}
	u.audioPlayer.PlayCue(cueName)
}
