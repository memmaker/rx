package console

import (
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"RogueUI/special"
	"cmp"
	"fmt"
	"github.com/0xcafed00d/joystick"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/audio"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math"
	"math/rand"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type UIState int

func (s UIState) IsTargeting() bool {
	return s == StateTargeting || s == StateLookTargeting || s == StateTargetingBodyPart
}

const (
	StateNormal UIState = iota
	StateLookTargeting
	StateTargeting
	StateTargetingBodyPart
)

type IntroState int

const (
	NoIntro IntroState = iota
	Fading
	TitleScreen
	InGame
)

type UI struct {
	settings *foundation.Configuration
	game     foundation.GameForUI

	audioPlayer *audio.Player

	uiTheme Theme

	mapOverlay *Overlay

	mainGrid        *cview.Grid
	lowerRightPanel *cview.TextView
	messageLabel    *cview.TextView
	statusBar       *cview.TextView
	rightPanel      *cview.TextView
	pages           *cview.Panels
	application     *cview.Application
	mapWindow       *cview.Box
	currentMouseX   int
	currentMouseY   int
	state           UIState
	targetingTiles  map[geometry.Point]bool

	animator  *Animator
	targetPos geometry.Point

	listTable map[string]*cview.List

	sentUIReady     bool
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
	lastHudStats       map[foundation.HudValue]int
	dialogueText       *cview.TextView
	dialogueOptions    *cview.List
	dialogueIsTerminal bool
	sentUIRunning      bool
	shouldPlayIntro    bool

	stateOfIntro IntroState
}

func (u *UI) GetKeybindingsAsString(command string) string {
	if command == "move" {
		allKeys := []string{
			u.GetKeysForCommandAsString(KeyLayerMain, "north"),
			u.GetKeysForCommandAsString(KeyLayerMain, "west"),
			u.GetKeysForCommandAsString(KeyLayerMain, "south"),
			u.GetKeysForCommandAsString(KeyLayerMain, "east"),
		}
		return fmt.Sprintf("[%s]", strings.Join(allKeys, ", "))
	}
	return u.GetKeysForCommandAsPrettyString(KeyLayerMain, command)
}

func (u *UI) AskForConfirmation(title, message string, choice func(didConfirm bool)) {
	dialogue := OpenConfirmDialogue(u.application, u.pages, title, message, choice)
	buttonOne := dialogue.GetForm().GetButton(0)
	oldBlurOne := buttonOne.GetInputCapture()
	buttonOne.SetInputCapture(u.directionalWrapper(oldBlurOne))

	buttonTwo := dialogue.GetForm().GetButton(1)
	oldBlurTwo := buttonTwo.GetInputCapture()
	buttonTwo.SetInputCapture(u.directionalWrapper(oldBlurTwo))
}

func (u *UI) AskForString(prompt string, prefill string, result func(entered string)) {
	cview.AskForString(u.application, u.pages, prompt, prefill, result)
}

func (u *UI) FadeToBlack() {
	cview.FadeToBlack(u.application, u.settings.AnimationDelay/2, 10, false)
}

func (u *UI) FadeFromBlack() {
	cview.FadeFromBlack(u.application, u.settings.AnimationDelay/2, 10, false)
}

func (u *UI) OpenKeypad(correctSequence []rune, onCompletion func(success bool)) {
	width, height := u.application.GetScreen().Size()
	keyPad := NewKeyPad(geometry.Point{X: width, Y: height})
	keyPad.SetCorrectSequence(correctSequence)
	keyPad.SetAudioPlayer(u.audioPlayer)
	keyPad.SetOnCompletion(func(success bool) {
		u.closeModal()
		onCompletion(success)
	})
	keyPad.SetVisible(true)
	origCapt := keyPad.GetInputCapture()
	keyPad.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		command := u.getCommandForKey(toUIKey(event))
		if command == "pickup" || command == "map_interaction" {
			u.closeModal()
		}
		return origCapt(event)
	})
	u.pages.AddPanel("modal", keyPad, false, true)
	u.lockFocusToPrimitive(keyPad)
}

func (u *UI) SetColors(palette textiles.ColorPalette, colors map[foundation.ItemCategory]color.RGBA) {
	u.uiTheme = NewUIThemeFromDataDir(u.settings.DataRootDir, palette, colors)
	u.setTheme()
}

func (u *UI) ShowTakeOnlyContainer(name string, containedItems []foundation.ItemForUI, transfer func(item foundation.ItemForUI)) {
	var menuItems []foundation.MenuItem
	menuLabels := u.menuLabelsFor(containedItems)
	for index, i := range containedItems {
		item := i
		menuItems = append(menuItems, foundation.MenuItem{
			Name: menuLabels[index],
			Action: func() {
				transfer(item)
				u.application.QueueUpdateDraw(u.UpdateLogWindow)
			},
			CloseMenus: true,
		})
	}

	menu := u.openSimpleMenu(menuItems)
	menu.SetTitle(name)
	keyForTakeAll := u.GetKeysForCommandAsString(KeyLayerMain, "pickup")
	u.Print(foundation.HiLite("Press %s to take all items", keyForTakeAll))
	originalCapture := menu.GetInputCapture() // will manage pressing escape
	menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		uiKey := toUIKey(event)
		command := u.getCommandForKey(uiKey)
		if command == "pickup" {
			for _, item := range containedItems {
				transfer(item)
			}
			originalCapture(EscapeKeyEvent())
			u.application.QueueUpdateDraw(u.UpdateLogWindow)
			return nil
		}
		return originalCapture(event)
	})
}

func (u *UI) openAmountWidget(itemName string, maxAmount int, onAmountSelected func(amount int)) {
	previousFocusHandler := u.application.GetBeforeFocusFunc()
	previousFocus := u.application.GetFocus()

	closeAmount := func() {
		u.pages.RemovePanel("amountWidget")
		u.application.SetBeforeFocusFunc(nil)
		u.application.SetFocus(previousFocus)
		u.application.SetBeforeFocusFunc(previousFocusHandler)
	}
	amountWidget := NewAmountWidget(itemName, maxAmount, func(amount int) {
		closeAmount()
		onAmountSelected(amount)
	}, u.application.SetFocus)

	screenWidth, _ := u.application.GetScreen().Size()
	amountWidget.SetRect(screenWidth/2-12, 4, 24, 7)

	originalCapture := amountWidget.GetInputCapture()
	amountWidget.SetInputCapture(u.directionalWrapperWithoutNumbers(originalCapture))

	u.pages.AddPanel("amountWidget", amountWidget, false, true)

	u.application.SetBeforeFocusFunc(nil)
	u.application.SetFocus(amountWidget)
	u.application.SetBeforeFocusFunc(func(p cview.Primitive) bool {
		if p == amountWidget || p == amountWidget.cancelButton || p == amountWidget.doneButton {
			return true
		}
		return false
	})

}

func (u *UI) ShowGiveAndTakeContainer(leftName string, leftItems []foundation.ItemForUI, rightName string, rightItems []foundation.ItemForUI, transferToLeft func(itemTaken foundation.ItemForUI, stackCount int), transferToRight func(itemTaken foundation.ItemForUI, stackCount int)) {
	var leftMenuItems []foundation.MenuItem
	var rightMenuItems []foundation.MenuItem
	var leftMenuLabels []string
	var rightMenuLabels []string
	if len(leftItems) > 0 {
		leftMenuLabels = u.menuLabelsFor(leftItems)
	}
	if len(rightItems) > 0 {
		rightMenuLabels = u.menuLabelsFor(rightItems)
	}

	for index, i := range leftItems {
		item := i
		leftMenuItems = append(leftMenuItems, foundation.MenuItem{
			Name: leftMenuLabels[index],
			Action: func() {
				if item.IsMultipleStacks() {
					u.openAmountWidget(item.Name(), item.GetStackSize(), func(amount int) {
						transferToRight(item, amount)
					})
				} else {
					transferToRight(item, 1)
				}
				u.application.QueueUpdateDraw(u.UpdateLogWindow)
			},
			CloseMenus: true,
		})
	}
	for index, i := range rightItems {
		item := i
		rightMenuItems = append(rightMenuItems, foundation.MenuItem{
			Name: rightMenuLabels[index],
			Action: func() {
				if item.IsMultipleStacks() {
					u.openAmountWidget(item.Name(), item.GetStackSize(), func(amount int) {
						transferToLeft(item, amount)
					})
				} else {
					transferToLeft(item, 1)
				}
				u.application.QueueUpdateDraw(u.UpdateLogWindow)
			},
			CloseMenus: true,
		})
	}

	leftMenu, longestItemLeft := u.createSimpleMenu(leftMenuItems)
	longestItemLeft = max(longestItemLeft, len(leftName))
	leftWidth := longestItemLeft + 2
	leftMenu.SetTitle(leftName)
	leftMenu.SetSelectedFocusOnly(true)

	if len(rightItems) > 0 {
		keyForTakeAll := u.GetKeysForCommandAsString(KeyLayerMain, "pickup")
		u.Print(foundation.HiLite("Press %s to take all items", keyForTakeAll))
	}

	rightMenu, longestItemRight := u.createSimpleMenu(rightMenuItems)
	longestItemRight = max(longestItemRight, len(rightName))
	rightWidth := longestItemRight + 2
	rightMenu.SetTitle(rightName)
	rightMenu.ShowFocus(true)
	rightMenu.SetSelectedFocusOnly(true)

	closeContainer := func() {
		u.pages.RemovePanel("leftModal")
		u.pages.RemovePanel("rightModal")
		u.resetFocusToMain()
	}

	screenWidth, screenHeight := u.application.GetScreen().Size()

	height := min(screenHeight-4, max(len(leftItems)+2, len(rightItems)+2))
	centerGap := 4
	//borderPadding := 2

	//screenRemaining := screenWidth - centerGap - (2 * borderPadding)
	equalWidth := max(leftWidth, rightWidth)
	leftWidth = equalWidth
	rightWidth = equalWidth

	halfCenterGap := centerGap / 2
	centerScreen := screenWidth / 2

	leftListStart := centerScreen - leftWidth - halfCenterGap
	rightListStart := leftListStart + leftWidth + centerGap
	leftMenu.ShowFocus(true)
	leftMenu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeContainer()
			return nil
		}
		uiKey := toUIKey(event)
		command := u.getCommandForKey(uiKey)
		if command == "west" || command == "east" {
			u.application.SetFocus(rightMenu)
			return nil
		} else if command == "wait" {
			closeContainer()
			return nil
		} else if command == "north" {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		} else if command == "south" {
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		}
		return event
	})

	rightMenu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeContainer()
			return nil
		}
		uiKey := toUIKey(event)
		command := u.getCommandForKey(uiKey)
		if command == "pickup" {
			for _, item := range rightItems {
				transferToLeft(item, item.GetStackSize())
			}
			u.application.QueueUpdateDraw(u.UpdateLogWindow)
			closeContainer()
			return nil
		} else if command == "west" || command == "east" {
			u.application.SetFocus(leftMenu)
			return nil
		} else if command == "wait" {
			closeContainer()
			return nil
		} else if command == "north" {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		} else if command == "south" {
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		}
		return event
	})

	// place both lists vertically centered
	// but side by side

	leftMenu.SetRect(leftListStart, 2, leftWidth, height)
	rightMenu.SetRect(rightListStart, 2, rightWidth, height)

	u.pages.AddPanel("leftModal", leftMenu, false, true)
	u.pages.AddPanel("rightModal", rightMenu, false, true)

	u.application.SetBeforeFocusFunc(nil)

	if len(rightItems) > 0 {
		u.application.SetFocus(rightMenu)
	} else {
		u.application.SetFocus(leftMenu)
	}

	u.application.SetBeforeFocusFunc(func(p cview.Primitive) bool {
		if p == leftMenu || p == rightMenu {
			return true
		}
		return false
	})
}

func (u *UI) menuLabelsFor(items []foundation.ItemForUI) []string {
	tablerows := make([]fxtools.TableRow, len(items))
	for index, i := range items {
		item := i

		itemName := item.InventoryNameWithColors(u.uiTheme.GetInventoryItemColorCode(item.GetCategory()))
		itemWeight := fmt.Sprintf("%dlbs", item.GetCarryWeight())

		tablerows[index] = fxtools.NewTableRow(itemName, itemWeight)
	}
	return fxtools.TableLayoutLastRight(tablerows)

}

func EscapeKeyEvent() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
}

func (u *UI) PlayMusic(fileName string) {
	u.audioPlayer.StopAll()
	u.audioPlayer.Stream(fileName)
}
func (u *UI) PlayCue(cueName string) {
	u.audioPlayer.PlayCue(cueName)
}

func (u *UI) OpenVendorMenu(itemsForSale []fxtools.Tuple[foundation.ItemForUI, int], buyItem func(ui foundation.ItemForUI, price int)) {
	var menuItems []foundation.MenuItem
	for _, i := range itemsForSale {
		item := i.GetItem1()
		price := i.GetItem2()
		menuItems = append(menuItems, foundation.MenuItem{
			Name: fmt.Sprintf("%s (%d)", item.InventoryNameWithColors(u.uiTheme.GetInventoryItemColorCode(item.GetCategory())), price),
			Action: func() {
				buyItem(item, price)
			},
		})
	}
	u.OpenMenu(menuItems)
}

func (u *UI) GetAnimBackgroundColor(position geometry.Point, colorName string, frameCount int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}
	iconAtLocation, _ := u.mapLookup(position)
	bgColor := u.uiTheme.GetColorByName(colorName)
	return NewCoverAnimation(position, iconAtLocation.WithBg(bgColor), frameCount, done)
}

func (u *UI) HighlightStatChange(stat dice_curve.Stat) {
	//TODO
}

func (u *UI) ShowGameOver(scoreInfo foundation.ScoreInfo, highScores []foundation.ScoreInfo) {
	u.animator.CancelAll()
	u.gameIsOver = true
	cview.FadeToBlack(u.application, u.settings.AnimationDelay, 10, false)

	if scoreInfo.Escaped {
		u.showWinScreen(scoreInfo, highScores)
	} else {
		u.showDeathScreen(scoreInfo, highScores)
	}
}
func (u *UI) showWinScreen(scoreInfo foundation.ScoreInfo, highScores []foundation.ScoreInfo) {
	textView := cview.NewTextView()
	textView.SetBorder(true)
	textView.SetScrollable(false)
	textView.SetScrollBarVisibility(cview.ScrollBarNever)
	textView.SetTextAlign(cview.AlignCenter)
	textView.SetTitleAlign(cview.AlignCenter)

	winMessage := fxtools.ReadFileAsLines(path.Join(u.settings.DataRootDir, "win.txt"))

	gameOverMessage := []string{
		"",
		"",
		"",
		fmt.Sprintf("%s", scoreInfo.PlayerName),
		fmt.Sprintf("Gold: %d", scoreInfo.Gold),
		fmt.Sprintf("%s", scoreInfo.DescriptiveMessage),
		"",
		"",
		"",
	}
	gameOverMessage = append(gameOverMessage, winMessage...)

	pressSpace := []string{
		"",
		"",
		"",
		"",
		fmt.Sprintf("Press [#FFFFFF::b]SPACE[-:-:-] to continue"),
	}
	gameOverMessage = append(gameOverMessage, pressSpace...)
	u.setColoredText(textView, strings.Join(gameOverMessage, "\n"))

	panelName := "modal"

	textView.SetInputCapture(u.popOnSpaceWithNotification(panelName, func() {
		u.resetFocusToMain()
		u.showHighscoresAndRestart(highScores)
	}))
	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.lockFocusToPrimitive(textView)
}
func (u *UI) showDeathScreen(scoreInfo foundation.ScoreInfo, highScores []foundation.ScoreInfo) {
	textView := cview.NewTextView()
	textView.SetBorder(false)
	textView.SetTextAlign(cview.AlignCenter)
	textView.SetTitleAlign(cview.AlignCenter)
	textView.SetBorder(true)
	textView.SetScrollable(false)
	textView.SetScrollBarVisibility(cview.ScrollBarNever)
	textView.SetTitle("You died")

	gameOverMessage := []string{
		"",
		fmt.Sprintf("%s", scoreInfo.PlayerName),
		fmt.Sprintf("Gold: %d", scoreInfo.Gold),
		fmt.Sprintf("Cause of Death: %s", scoreInfo.DescriptiveMessage),
	}
	restartText := []string{
		"",
		"",
		"[#fccc2b::b]Do you want to play again? (y/n)[-:-:-]",
		"",
		"",
	}
	scoreTable := toLinesOfText(highScores)

	gameOverMessage = append(gameOverMessage, restartText...)
	gameOverMessage = append(gameOverMessage, scoreTable...)

	u.setColoredText(textView, strings.Join(gameOverMessage, "\n"))

	panelName := "modal"

	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.lockFocusToPrimitive(textView)
	textView.SetInputCapture(u.yesNoReceiver(u.reset, u.application.Stop))
}

func (u *UI) reset() {
	u.mapOverlay.ClearAll()
	u.pages.RemovePanel("modal")
	u.resetFocusToMain()
	u.gameIsOver = false
	u.game.Reset()
}
func (u *UI) showHighscoresAndRestart(highScores []foundation.ScoreInfo) {
	textView := cview.NewTextView()
	textView.SetBorder(false)
	textView.SetTextAlign(cview.AlignCenter)
	textView.SetTitleAlign(cview.AlignCenter)
	textView.SetTitle("Game Over")

	restartText := []string{
		"",
		"[#fccc2b::b]Do you want to play again? (y/n)[-:-:-]",
		"",
	}
	scoreTable := toLinesOfText(highScores)
	gameOverMessage := append(restartText, scoreTable...)

	u.setColoredText(textView, strings.Join(gameOverMessage, "\n"))

	panelName := "modal"

	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(textView)
	textView.SetInputCapture(u.yesNoReceiver(u.reset, u.application.Stop))
}

func toLinesOfText(highScores []foundation.ScoreInfo) []string {
	scoreTable := []string{
		"= Top 10 Dungeon Crawlers =",
		"",
	}
	for i, highScore := range highScores {
		if i == 10 {
			break
		}
		scoreLine := ""
		if highScore.Escaped {
			scoreLine = fmt.Sprintf("[#c9c54d::b]%d. %s: %d Gold, %s[-:-:-]", i+1, highScore.PlayerName, highScore.Gold, highScore.DescriptiveMessage)
		} else {
			scoreLine = fmt.Sprintf("%d. %s: %d Gold, CoD: %s", i+1, highScore.PlayerName, highScore.Gold, highScore.DescriptiveMessage)
		}
		scoreTable = append(scoreTable, scoreLine)
	}
	return scoreTable
}

func (u *UI) ShowHighScoresOnly(highScores []foundation.ScoreInfo) {
	textView := cview.NewTextView()
	textView.SetBorder(false)
	textView.SetTextAlign(cview.AlignCenter)
	textView.SetTitleAlign(cview.AlignCenter)
	textView.SetTitle("High Scores")

	scoreTable := toLinesOfText(highScores)
	u.setColoredText(textView, strings.Join(scoreTable, "\n"))

	panelName := "main"
	if u.pages.HasPanel("main") {
		panelName = "fullscreen"
	}

	textView.SetInputCapture(u.popOnAnyKeyWithNotification(panelName, u.application.Stop))
	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(textView)
}

func (u *UI) AddAnimations(animations []foundation.Animation) {
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
func (u *UI) AfterPlayerMoved(moveInfo foundation.MoveInfo) {
	if moveInfo.Mode == foundation.PlayerMoveModePath {
		u.application.QueueEvent(tcell.NewEventKey(tcell.KeyF40, ' ', 128))
		u.autoRun = true
	} else if moveInfo.Mode == foundation.PlayerMoveModeRun && u.autoRun {
		u.application.QueueEvent(tcell.NewEventKey(tcell.KeyRune, directionToRune(moveInfo.Direction), 64))
	}
}

func (u *UI) GetAnimMove(actor foundation.ActorForUI, old geometry.Point, new geometry.Point) foundation.Animation {
	if u.settings.AnimationsEnabled && u.settings.AnimateMovement {
		return NewMovementAnimation(u.getIconForActor(actor), old, new, u.uiTheme.GetColorByName, nil)
	}
	return nil
}

func (u *UI) getIconForActor(actor foundation.ActorForUI) textiles.TextIcon {
	mapIconHere := u.game.MapAt(actor.Position())

	isHallucinating := u.isPlayerHallucinating()
	if isHallucinating {
		randomLetter := rune('A' + rand.Intn(26))
		if rand.Intn(2) == 0 {
			randomLetter = unicode.ToLower(randomLetter)
		}
		return textiles.TextIcon{
			Char: randomLetter,
			Fg:   u.uiTheme.GetRandomColor(),
			Bg:   mapIconHere.Bg,
		}
	}

	var backGroundColor color.RGBA

	if actor.HasFlag(special.FlagHeld) {
		return textiles.TextIcon{
			Char: actor.Icon().Char,
			Fg:   u.uiTheme.GetColorByName("Blue_1"),
			Bg:   u.uiTheme.GetColorByName("White"),
		}
	} else {
		backGroundColor = mapIconHere.Bg
	}
	if !actor.IsAlive() {
		return actor.TextIcon(backGroundColor).WithRune('%')
	}
	return actor.TextIcon(backGroundColor)
}

func (u *UI) isPlayerHallucinating() bool {
	flags := u.game.GetHudFlags()
	_, isHallucinating := flags[special.FlagHallucinating]
	return isHallucinating
}

func (u *UI) GetAnimQuickMove(actor foundation.ActorForUI, path []geometry.Point) foundation.Animation {
	if u.settings.AnimationsEnabled && u.settings.AnimateMovement {
		animation := NewMovementAnimation(u.getIconForActor(actor), actor.Position(), path[len(path)-1], u.uiTheme.GetColorByName, nil)
		animation.EnableQuickMoveMode(path)
		return animation
	}
	return nil
}
func (u *UI) GetAnimCover(loc geometry.Point, icon textiles.TextIcon, turns int, done func()) foundation.Animation {
	if u.settings.AnimationsEnabled && u.settings.AnimateMovement {
		return NewCoverAnimation(loc, icon, turns, done)
	}
	return nil
}

func (u *UI) GetAnimAttack(attacker, defender foundation.ActorForUI) foundation.Animation {
	return nil
}

func (u *UI) GetAnimDamage(spreadBlood func(mapPos geometry.Point), defenderPos geometry.Point, damage int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateDamage {
		return nil
	}
	bloodColors := []color.RGBA{
		u.uiTheme.palette.Get("red_5"),
		u.uiTheme.palette.Get("red_6"),
		u.uiTheme.palette.Get("red_7"),
		u.uiTheme.palette.Get("red_8"),
		u.uiTheme.palette.Get("red_9"),
		u.uiTheme.palette.Get("red_10"),
		u.uiTheme.palette.Get("red_11"),
		u.uiTheme.palette.Get("red_12"),
		u.uiTheme.palette.Get("red_13"),
		u.uiTheme.palette.Get("red_14"),
		u.uiTheme.palette.Get("red_15"),
	}
	animation := NewDamageAnimation(spreadBlood, defenderPos, u.game.GetPlayerPosition(), damage, bloodColors)
	animation.SetDoneCallback(done)
	return animation
}
func (u *UI) GetAnimTiles(positions []geometry.Point, frames []textiles.TextIcon, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}
	return NewTilesAnimation(positions, frames, done)
}

func (u *UI) GetAnimRadialReveal(position geometry.Point, dijkstra map[geometry.Point]int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}

	animation := NewRadialAnimation(position, dijkstra, u.uiTheme.GetColorByName, u.mapLookup, done)
	animation.SetKeepDrawingCoveredGround(true)
	animation.SetUseIconColors(false)
	return animation
}

func (u *UI) GetAnimRadialAlert(position geometry.Point, dijkstra map[geometry.Point]int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}
	lookup := func(loc geometry.Point) (textiles.TextIcon, bool) {
		return textiles.TextIcon{
			Char: '‼',
			Fg:   u.uiTheme.GetColorByName("Black"),
			Bg:   u.uiTheme.GetColorByName("Red_1"),
		}, true
	}
	animation := NewRadialAnimation(position, dijkstra, u.uiTheme.GetColorByName, lookup, done)
	animation.SetUseIconColors(true)
	return animation
}

func (u *UI) GetAnimTeleport(user foundation.ActorForUI, origin, targetPos geometry.Point, appearOnMap func()) (foundation.Animation, foundation.Animation) {
	originalIcon := u.getIconForActor(user)
	mapBackground := u.getMapTileBackgroundColor(origin)
	lightCyan := u.uiTheme.GetColorByName("LightCyan")
	white := u.uiTheme.GetColorByName("White")
	lightGray := u.uiTheme.GetColorByName("light_gray_5")
	vanishAnim := u.GetAnimTiles([]geometry.Point{origin}, []textiles.TextIcon{
		originalIcon.WithFg(white),
		originalIcon.WithFg(white),
		originalIcon.WithFg(lightCyan),
		{Char: '*', Fg: lightCyan, Bg: mapBackground},
		{Char: '*', Fg: lightCyan, Bg: mapBackground},
		{Char: '+', Fg: lightCyan, Bg: mapBackground},
		{Char: '+', Fg: lightCyan, Bg: mapBackground},
		{Char: '|', Fg: lightCyan, Bg: mapBackground},
		{Char: '|', Fg: lightCyan, Bg: mapBackground},
		{Char: '∙', Fg: lightCyan, Bg: mapBackground},
		{Char: '.', Fg: lightCyan, Bg: mapBackground},
		{Char: '.', Fg: lightGray, Bg: mapBackground},
		{Char: '.', Fg: u.uiTheme.GetColorByName("dark_gray_3"), Bg: mapBackground},
	}, nil)
	vanishAnim.RequestMapUpdateOnFinish()

	appearAnim := u.GetAnimAppearance(user, targetPos, appearOnMap)
	vanishAnim.SetFollowUp([]foundation.Animation{appearAnim})
	return vanishAnim, appearAnim
}

func (u *UI) GetAnimAppearance(actor foundation.ActorForUI, targetPos geometry.Point, done func()) foundation.Animation {
	originalIcon := u.getIconForActor(actor)
	mapBackground := u.getMapTileBackgroundColor(targetPos)
	lightCyan := u.uiTheme.GetColorByName("LightCyan")
	white := u.uiTheme.GetColorByName("White")
	lightGray := u.uiTheme.GetColorByName("light_gray_5")
	appearAnim := u.GetAnimTiles([]geometry.Point{targetPos}, []textiles.TextIcon{
		{Char: '.', Fg: u.uiTheme.GetColorByName("dark_gray_3"), Bg: mapBackground},
		{Char: '.', Fg: lightGray, Bg: mapBackground},
		{Char: '.', Fg: lightGray, Bg: mapBackground},
		{Char: '.', Fg: lightCyan, Bg: mapBackground},
		{Char: '∙', Fg: lightCyan, Bg: mapBackground},
		{Char: '|', Fg: lightCyan, Bg: mapBackground},
		{Char: '|', Fg: lightCyan, Bg: mapBackground},
		{Char: '+', Fg: lightCyan, Bg: mapBackground},
		{Char: '+', Fg: lightCyan, Bg: mapBackground},
		{Char: '*', Fg: lightCyan, Bg: mapBackground},
		{Char: '*', Fg: lightCyan, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: white, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: white, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: lightCyan, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: lightCyan, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: white, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: white, Bg: mapBackground},
	}, done)
	return appearAnim
}

func (u *UI) GetAnimEvade(defender foundation.ActorForUI, done func()) foundation.Animation {
	actorIcon := u.getIconForActor(defender)
	return u.GetAnimTiles([]geometry.Point{defender.Position()}, []textiles.TextIcon{
		actorIcon.WithItalic(),
		actorIcon.WithItalic().WithBold(),
		actorIcon.WithItalic(),
	},
		done)
}

func (u *UI) GetAnimWakeUp(location geometry.Point, done func()) foundation.Animation {
	keepAllNeighbors := func(point geometry.Point) bool { return true }

	neigh := geometry.Neighbors{}
	cardinalNeighbors := neigh.Cardinal(location, keepAllNeighbors)
	diagonalNeighbors := neigh.Diagonal(location, keepAllNeighbors)

	wakeUpRunes := []rune("????")
	yellow := u.uiTheme.GetColorByName("Yellow_1")
	var prevAnim foundation.Animation
	var rootAnim foundation.Animation
	runeCount := len(wakeUpRunes)
	for i := 0; i < runeCount; i++ {

		cycleIcon := textiles.TextIcon{
			Char: wakeUpRunes[i],
			Fg:   u.uiTheme.GetUIColor(UIColorUIForeground),
			Bg:   u.uiTheme.GetColorByName("black"),
		}

		frames := []textiles.TextIcon{
			cycleIcon.WithFg(yellow),
			cycleIcon.WithFg(yellow),
		}

		var neighbors []geometry.Point
		if i%2 == 0 {
			neighbors = cardinalNeighbors
		} else {
			neighbors = diagonalNeighbors
		}
		var doneCall func()
		if i == runeCount-1 {
			doneCall = done
		}
		anim := u.GetAnimTiles(neighbors, frames, doneCall)
		if rootAnim == nil {
			rootAnim = anim
		}

		if prevAnim != nil {
			prevAnim.SetFollowUp([]foundation.Animation{anim})
		}

		prevAnim = anim
	}
	return rootAnim
}
func (u *UI) GetAnimConfuse(location geometry.Point, done func()) foundation.Animation {
	keepAllNeighbors := func(point geometry.Point) bool { return true }

	neigh := geometry.Neighbors{}
	cardinalNeighbors := neigh.Cardinal(location, keepAllNeighbors)
	diagonalNeighbors := neigh.Diagonal(location, keepAllNeighbors)

	confuseRune := []rune("?¿¡!")
	randomRune := func() rune {
		return confuseRune[rand.Intn(len(confuseRune))]
	}
	confuseColors := []color.RGBA{u.uiTheme.GetColorByName("LightMagenta"), u.uiTheme.GetColorByName("LightRed"), u.uiTheme.GetColorByName("Yellow_1"), u.uiTheme.GetColorByName("LightGreen"), u.uiTheme.GetColorByName("light_blue_3")}
	randomColor := func() color.RGBA {
		return confuseColors[rand.Intn(len(confuseColors))]
	}
	cycleCount := 4

	var prevAnim foundation.Animation
	var rootAnim foundation.Animation
	for i := 0; i < cycleCount; i++ {

		cycleIcon := textiles.TextIcon{
			Char: randomRune(),
			Fg:   u.uiTheme.GetUIColor(UIColorUIForeground),
			Bg:   u.uiTheme.GetUIColor(UIColorUIBackground),
		}

		frames := []textiles.TextIcon{
			cycleIcon.WithFg(randomColor()),
			cycleIcon.WithFg(randomColor()),
			cycleIcon.WithFg(randomColor()),
			cycleIcon.WithFg(randomColor()),
		}

		var neighbors []geometry.Point
		if i%2 == 0 {
			neighbors = cardinalNeighbors
		} else {
			neighbors = diagonalNeighbors
		}
		var doneCall func()
		if i == cycleCount-1 {
			doneCall = done
		}
		anim := u.GetAnimTiles(neighbors, frames, doneCall)
		if rootAnim == nil {
			rootAnim = anim
		}

		if prevAnim != nil {
			prevAnim.SetFollowUp([]foundation.Animation{anim})
		}

		prevAnim = anim
	}
	return rootAnim
}
func (u *UI) GetAnimBreath(path []geometry.Point, done func()) foundation.Animation {
	projAnim := u.GetAnimTiles(path, []textiles.TextIcon{
		{Char: '.', Fg: u.uiTheme.GetColorByName("White"), Bg: u.uiTheme.GetColorByName("Black")},
		{Char: '∙', Fg: u.uiTheme.GetColorByName("White"), Bg: u.uiTheme.GetColorByName("Black")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("White"), Bg: u.uiTheme.GetColorByName("Black")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("Yellow"), Bg: u.uiTheme.GetColorByName("Black")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("Red"), Bg: u.uiTheme.GetColorByName("Black")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("light_gray_5"), Bg: u.uiTheme.GetColorByName("Black")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("dark_gray_3"), Bg: u.uiTheme.GetColorByName("Black")},
	}, done)
	return projAnim
}
func (u *UI) GetAnimVorpalizeWeapon(origin geometry.Point, done func()) []foundation.Animation {
	effectIcon := textiles.TextIcon{
		Char: '+',
		Fg:   u.uiTheme.GetColorByName("White"),
		Bg:   u.uiTheme.GetColorByName("Black"),
	}
	outmostPositions := geometry.CircleAround(origin, 2)
	outerPositions := geometry.CircleAround(origin, 1)

	animationInner := u.GetAnimTiles([]geometry.Point{origin}, []textiles.TextIcon{
		effectIcon.WithBg(u.uiTheme.GetColorByName("White")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("White")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("Black")),
	}, done)
	animationCenter := u.GetAnimTiles(outerPositions, []textiles.TextIcon{
		effectIcon.WithBg(u.uiTheme.GetColorByName("White")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("Black")),
	}, nil)

	animationOuter := u.GetAnimTiles(outmostPositions, []textiles.TextIcon{
		effectIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("Black")),
	}, nil)

	return []foundation.Animation{animationInner, animationCenter, animationOuter}
}
func (u *UI) GetAnimEnchantWeapon(player foundation.ActorForUI, location geometry.Point, done func()) foundation.Animation {
	playerIcon := u.getIconForActor(player)
	frames := []textiles.TextIcon{
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("LightCyan")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("LightCyan")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("LightCyan")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue_1")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
	}
	return u.GetAnimTiles([]geometry.Point{location}, frames, done)
}
func (u *UI) GetAnimEnchantArmor(player foundation.ActorForUI, location geometry.Point, done func()) foundation.Animation {
	playerIcon := u.getIconForActor(player)
	frames := []textiles.TextIcon{
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("White")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("White")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("White")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("White")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
	}

	return u.GetAnimTiles([]geometry.Point{location}, frames, done)
}
func (u *UI) GetAnimThrow(item foundation.ItemForUI, origin geometry.Point, target geometry.Point) (foundation.Animation, int) {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateProjectiles {
		return nil, 0
	}
	textIcon := item.GetIcon()

	return u.GetAnimProjectileWithIcon(textIcon, origin, target, nil)
}

func (u *UI) GetAnimProjectile(icon rune, fgColor string, origin geometry.Point, target geometry.Point, done func()) (foundation.Animation, int) {
	textIcon := textiles.TextIcon{
		Char: icon,
		Fg:   u.uiTheme.GetColorByName(fgColor),
	}
	return u.GetAnimProjectileWithIcon(textIcon, origin, target, done)
}
func (u *UI) GetAnimProjectileWithIcon(textIcon textiles.TextIcon, origin geometry.Point, target geometry.Point, done func()) (foundation.Animation, int) {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateProjectiles {
		return nil, 0
	}
	pathOfFlight := geometry.BresenhamLine(origin, target, func(x, y int) bool {
		return true
	})

	if len(pathOfFlight) == 0 {
		return nil, 0
	}

	return NewProjectileAnimation(pathOfFlight, textIcon, u.mapLookup, done), len(pathOfFlight)
}

func (u *UI) GetAnimProjectileWithTrail(leadIcon rune, colorNames []string, pathOfFlight []geometry.Point, done func()) (foundation.Animation, int) {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateProjectiles {
		return nil, 0
	}

	if len(pathOfFlight) == 0 {
		return nil, 0
	}

	var trailIcons []textiles.TextIcon

	for i, cName := range colorNames {
		if i == 0 {
			trailIcons = append(trailIcons, textiles.TextIcon{
				Char: leadIcon,
				Fg:   u.uiTheme.GetColorByName(cName),
			})
		} else {
			trailIcons = append(trailIcons, textiles.TextIcon{
				Char: '█',
				Fg:   u.uiTheme.GetColorByName(cName),
			})
		}
	}

	animation := NewProjectileAnimation(pathOfFlight, trailIcons[0], u.mapLookup, done)
	animation.SetTrail(trailIcons[1:])
	return animation, len(pathOfFlight)
}

func (u *UI) GetAnimProjectileWithLight(leadIcon rune, lightColorName string, pathOfFlight []geometry.Point, done func()) (foundation.Animation, int) {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateProjectiles {
		return nil, 0
	}

	if len(pathOfFlight) == 0 {
		return nil, 0
	}

	lightColor := u.uiTheme.GetColorByName(lightColorName)
	leadTextIcon := textiles.TextIcon{
		Char: leadIcon,
		Fg:   lightColor,
	}

	animation := NewProjectileAnimation(pathOfFlight, leadTextIcon, u.mapLookup, done)
	animation.SetLightSource(&gridmap.LightSource{
		Pos:          pathOfFlight[0],
		Radius:       1,
		Color:        fxtools.NewColorFromRGBA(lightColor),
		MaxIntensity: 3,
	})
	return animation, len(pathOfFlight)
}

func (u *UI) updateUntilDone() bool {
	screen := u.application.GetScreen()
	duration := 2 * time.Millisecond

	u.application.Lock()
	defer u.application.Unlock()

	u.isAnimationFrame = true
	var breakingKey *tcell.EventKey
outerLoop:
	for len(u.animator.runningAnimations) > 0 {
		u.mapWindow.Draw(screen)
		screen.Show()

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

	u.mapWindow.Draw(screen)
	screen.Show()

	return false
}

func (u *UI) ShowTextFile(fileName string) {
	lines := fxtools.ReadFile(fileName)
	u.OpenTextWindow(lines)
}
func (u *UI) OpenTextWindow(description string) {
	u.openTextModal(description)
}

func (u *UI) ShowTextFileFullscreen(filename string, onClose func()) {
	lines := fxtools.ReadFileAsLines(filename)
	textView := cview.NewTextView()
	textView.SetBorder(false)
	u.setColoredText(textView, strings.Join(lines, "\n"))

	panelName := "main"
	if u.pages.HasPanel("main") {
		panelName = "fullscreen"
	}

	textView.SetInputCapture(u.popOnAnyKeyWithNotification(panelName, onClose))
	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(textView)
}

func (u *UI) openTextModal(description string) *cview.TextView {
	textView := u.newTextModal(description)
	w, h := widthAndHeightFromString(description)
	u.makeCenteredModal(textView, w, h)

	originalInputCapture := textView.GetInputCapture()
	textView.SetInputCapture(u.directionalWrapper(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEnter {
			u.closeModal()
			return nil
		}
		if originalInputCapture != nil {
			return originalInputCapture(event)
		}
		return event
	}))
	return textView
}

func (u *UI) newTextModal(description string) *cview.TextView {
	textView := cview.NewTextView()
	textView.SetWrap(true)
	textView.SetWrapWidth(0)
	textView.SetWordWrap(true)
	textView.SetBorder(true)

	textView.SetTextColor(toTcellColor(u.uiTheme.GetColorByName("light_gray_5")))
	textView.SetBackgroundColor(toTcellColor(u.uiTheme.GetColorByName("black")))

	textView.SetBorderColor(u.uiTheme.GetUIColorForTcell(UIColorBorderForeground))

	u.setColoredText(textView, description)
	return textView
}

func (u *UI) setColoredText(view *cview.TextView, text string) {
	view.SetDynamicColors(true)
	view.SetText(text)
}

func (u *UI) UpdateLogWindow() {
	logMessages := u.game.GetLog()
	var asColoredStrings []string
	for i, message := range logMessages {
		fadePercent := fxtools.Clamp(0.2, 1.0, float64(i+1)/float64(len(logMessages)))
		asColoredStrings = append(asColoredStrings, u.ToColoredText(message, fadePercent))
	}

	u.setColoredText(u.messageLabel, strings.Join(asColoredStrings, "\n"))
}

func (u *UI) ToColoredText(h foundation.HiLiteString, intensity float64) string {
	if h.IsEmpty() {
		return ""
	}
	textColor := u.uiTheme.GetUIColor(UIColorUIForeground)
	hiLiteColor := u.uiTheme.GetUIColor(UIColorTextForegroundHighlighted)
	if intensity < 1.0 {
		textColor = fxtools.SetBrightness(textColor, intensity)
		hiLiteColor = fxtools.SetBrightness(hiLiteColor, intensity)
	}
	if len(h.Value) == 0 {
		return h.AppendRepetitions(fmt.Sprintf("%s%s", textiles.RGBAToFgColorCode(hiLiteColor), h.FormatString))
	}
	textColorCode := textiles.RGBAToFgColorCode(textColor)
	if h.FormatString == "" {
		return h.AppendRepetitions(fmt.Sprintf("%s%s", textColorCode, h.Value[0]))
	}
	hiLiteColorCode := textiles.RGBAToFgColorCode(hiLiteColor)
	anyValues := make([]interface{}, len(h.Value)+1)
	anyValues[0] = textColorCode
	for i, v := range h.Value {
		anyValues[i+1] = fmt.Sprintf("%s%s%s", hiLiteColorCode, v, textColorCode)
	}
	return h.AppendRepetitions(fmt.Sprintf("%s"+h.FormatString, anyValues...))
}

func (u *UI) SetGame(game foundation.GameForUI) {
	u.game = game
}

// Print prints a message to the screen.
// Should only be called by the game
func (u *UI) Print(message foundation.HiLiteString) {
	if message.IsEmpty() {
		return
	}
	u.application.QueueUpdateDraw(func() {
		u.setColoredText(u.messageLabel, u.ToColoredText(message, 1))
	})
}
func (u *UI) StartGameLoop() {
	u.application.QueueUpdate(func() {
		w, h := u.application.GetScreen().Size()
		resize := tcell.NewEventResize(w, h)
		u.application.QueueEvent(resize)
	})
	if err := u.application.Run(); err != nil {
		panic(err)
	}
}

func (u *UI) initCoreUI() {
	cview.TrueColorTags = true
	cview.ColorUnset = tcell.ColorBlack

	u.application = cview.NewApplication()
	u.application.SetAfterDrawFunc(func(screen tcell.Screen) {
		if u.mainGrid != nil && u.sentUIRunning && !u.sentUIReady && u.GetMapWindowGridSize().X >= 80 && u.GetMapWindowGridSize().Y >= 23 {

			u.sentUIReady = true

			u.application.SetAfterDrawFunc(nil)

			u.application.QueueUpdateDraw(func() {
				u.game.UIReady()
			})

			//u.application.QueueEvent(tcell.NewEventKey(tcell.KeyRune, ' ', 0)) // WTF DOESNT THIS WORK?
		}
	})
	u.application.SetUnknownEventCapture(u.handleUnknownEvent)
	u.application.SetAfterResizeFunc(u.onTerminalResized)
	u.application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if u.stateOfIntro == Fading {
			return event
		}
		if u.stateOfIntro == TitleScreen {
			//u.endIntro()
			u.endIntro()
		}
		u.animator.CancelAll()
		if frontName, frontPanel := u.pages.GetFrontPanel(); frontName == "inventory" && event.Key() == tcell.KeyCtrlC {
			inventory := frontPanel.(*TextInventory)
			inventory.handleInput(event)
			return nil // don't forward, or else we will quit
		}
		return event
	})

	u.pages = cview.NewPanels()

	u.application.SetRoot(u.pages, true)
}
func (u *UI) InitDungeonUI(palette textiles.ColorPalette, invColors map[foundation.ItemCategory]color.RGBA) {
	if u.mainGrid != nil {
		return
	}

	u.setupCommandTable()
	u.loadKeyMap(path.Join(u.settings.DataRootDir, "keymaps", u.settings.KeyMap+".txt"))

	u.application.GetScreen().SetCursorStyle(tcell.CursorStyleSteadyBlock)

	u.application.EnableMouse(true)

	u.application.SetMouseCapture(u.handleMainMouse)
	u.application.SetBeforeFocusFunc(u.defaultFocusHandler)

	u.mapWindow = cview.NewBox()
	u.mapWindow.SetDrawFunc(u.drawMap)
	u.mapWindow.SetInputCapture(u.handleMainInput)

	u.messageLabel = cview.NewTextView()

	u.statusBar = cview.NewTextView()
	u.statusBar.SetDynamicColors(true)
	u.statusBar.SetScrollable(false)
	u.statusBar.SetScrollBarVisibility(cview.ScrollBarNever)

	u.rightPanel = cview.NewTextView()
	u.rightPanel.SetScrollable(false)
	u.rightPanel.SetScrollBarVisibility(cview.ScrollBarNever)
	u.rightPanel.SetDynamicColors(true)
	u.rightPanel.SetWrap(false)
	u.lowerRightPanel = cview.NewTextView()
	u.lowerRightPanel.SetScrollable(false)
	u.lowerRightPanel.SetDynamicColors(true)
	u.lowerRightPanel.SetWordWrap(true)

	grid := cview.NewGrid()

	grid.SetRows(1, 0, 1)
	grid.SetColumns(u.settings.MapWidth, 0)
	//SetColumns(30, 0, 30).
	//SetBorders(true).
	panelThreshold := u.settings.MapWidth + 1
	logThreshold := u.settings.MapHeight + 4
	grid.AddItem(u.messageLabel, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(u.messageLabel, 0, 0, 1, 1, 0, panelThreshold, false)
	grid.AddItem(u.messageLabel, 1, 0, 1, 2, logThreshold, 0, false)
	grid.AddItem(u.messageLabel, 1, 0, 1, 1, logThreshold, panelThreshold, false)
	grid.AddItem(u.mapWindow, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(u.mapWindow, 0, 0, 1, 1, logThreshold, 0, true)
	grid.AddItem(u.rightPanel, 0, 1, 2, 1, 0, panelThreshold, false)
	grid.AddItem(u.rightPanel, 0, 1, 1, 1, logThreshold, panelThreshold, false)
	grid.AddItem(u.lowerRightPanel, 1, 1, 1, 1, logThreshold, panelThreshold, false)
	grid.AddItem(u.statusBar, 2, 0, 1, 2, 0, 0, false)

	u.mainGrid = grid

	u.pages.AddPanel("main", grid, true, true)

	u.application.SetFocus(grid)

	u.mapOverlay = NewOverlay(u.settings.MapWidth, u.settings.MapHeight)

	u.SetColors(palette, invColors)

	if u.shouldPlayIntro && u.stateOfIntro == NoIntro {
		u.startIntro()
	}
}
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
		u.game.RunPlayerPath()
		return nil
	}
	if mod == 64 && u.autoRun && strings.ContainsRune("12346789", ch) {
		direction := runeToDirection(ch)
		u.continueAutoRun(direction)
		return nil
	}
	u.autoRun = false

	uiKey := toUIKey(ev)
	playerCommand := u.getCommandForKey(uiKey)
	u.executePlayerCommand(playerCommand)

	return nil
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

func (u *UI) continueAutoRun(direction geometry.CompassDirection) {
	time.Sleep(64 * time.Millisecond)
	canRun := u.game.RunPlayer(direction, false)
	if !canRun {
		u.autoRun = false
	}
}

func (u *UI) applyStylingToUI() {

	fg := u.uiTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.uiTheme.GetUIColorForTcell(UIColorUIBackground)
	u.statusBar.SetTextColor(tcell.ColorLightSlateGray)
	u.statusBar.SetBackgroundColor(tcell.ColorBlack)

	u.messageLabel.SetTextColor(fg)
	u.messageLabel.SetBackgroundColor(bg)
	u.messageLabel.SetBorderColor(fg)
	u.messageLabel.SetScrollBarColor(fg)
	u.messageLabel.SetDynamicColors(true)

	u.rightPanel.SetTextColor(fg)
	u.rightPanel.SetBorderColor(fg)
	u.rightPanel.SetBackgroundColor(bg)
	u.rightPanel.SetDynamicColors(true)
	u.rightPanel.SetTextAlign(cview.AlignRight)

	u.lowerRightPanel.SetTextColor(fg)
	u.lowerRightPanel.SetBorderColor(fg)
	u.lowerRightPanel.SetBackgroundColor(bg)
	u.lowerRightPanel.SetDynamicColors(true)
	u.lowerRightPanel.SetTextAlign(cview.AlignLeft)

	u.mapOverlay.SetDefaultColors(tcellColorToRGBA(bg), tcellColorToRGBA(fg))
}

func (u *UI) setTheme() {
	u.uiTheme.SetBorders(&cview.Borders)
	u.applyStylingToUI()
}

func (u *UI) drawMap(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	if !u.sentUIReady {
		return x, y, width, height
	}
	defaultMapStyle := u.uiTheme.GetMapDefaultStyle()
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {

			mapPosX := col - x
			mapPosY := row - y

			mapPos := geometry.Point{X: mapPosX, Y: mapPosY}

			ch, style := u.renderMapPosition(mapPos, u.isAnimationFrame, defaultMapStyle)

			screen.SetContent(col, row, ch, nil, style)
		}
	}
	if u.showCursor {
		screen.ShowCursor(u.game.GetPlayerPosition().X+x, u.game.GetPlayerPosition().Y+y)
	}
	// Space for other content.
	return x, y, width, height
}

func (u *UI) renderMapPosition(mapPos geometry.Point, isAnimationFrame bool, style tcell.Style) (rune, tcell.Style) {
	var ch rune
	var textIcon textiles.TextIcon
	isPositionAnimated := u.animator.isPositionAnimated(mapPos)
	foundIcon := false

	if animIcon, exists := u.animator.animationState[mapPos]; isAnimationFrame && exists {
		foundIcon = true
		if animIcon.HasBackground() {
			textIcon = animIcon
		} else {
			mapIcon, _ := u.mapLookup(mapPos)
			textIcon = animIcon.WithBg(mapIcon.Bg)
		}
	} else if u.mapOverlay.IsSet(mapPos.X, mapPos.Y) {
		textIcon = u.mapOverlay.Get(mapPos.X, mapPos.Y)
		foundIcon = true
	} else {
		textIcon, foundIcon = u.mapLookup(mapPos)
	}

	var fg, bg color.RGBA

	if foundIcon {
		ch, fg, bg = textIcon.Char, textIcon.Fg, textIcon.Bg
		style = style.Attributes(textIcon.Attributes)
	} else {
		ch = ' '
		fg = u.uiTheme.GetUIColor(UIColorUIForeground)
		bg = u.uiTheme.GetUIColor(UIColorUIBackground)
	}

	style = style.Foreground(tcell.NewRGBColor(int32(applyGamma(fg.R, u.gamma)), int32(applyGamma(fg.G, u.gamma)), int32(applyGamma(fg.B, u.gamma))))
	style = style.Background(tcell.NewRGBColor(int32(applyGamma(bg.R, u.gamma)), int32(applyGamma(bg.G, u.gamma)), int32(applyGamma(bg.B, u.gamma))))

	if isAnimationFrame && !isPositionAnimated {
		ch = u.lastFrameIcons[mapPos]
		style = u.lastFrameStyle[mapPos]
	}

	if !isAnimationFrame {
		u.lastFrameStyle[mapPos] = style
		u.lastFrameIcons[mapPos] = ch
	}

	if _, ok := u.targetingTiles[mapPos]; (u.state.IsTargeting()) && ok {
		if mapPos == u.targetPos {
			ch = 'X'
		}
		style = style.Reverse(true)
	}
	return ch, style
}

func applyGamma(colorChannel uint8, gamma float64) uint8 {
	colorAsFloat := float64(colorChannel) / 255.0
	gammaCorrected := fxtools.Clamp(0, 1, math.Pow(colorAsFloat, gamma))
	asEightBit := uint8(gammaCorrected * 255.0)
	return asEightBit
}

func tcellColorToRGBA(tColor tcell.Color) color.RGBA {
	rF, gF, bF := tColor.RGB()
	return color.RGBA{R: uint8(rF), G: uint8(gF), B: uint8(bF), A: 255}
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

func (u *UI) UpdateInventory() {
	items := u.game.GetInventoryForUI()
	if len(items) == 0 {
		u.rightPanel.Clear()
		return
	}
	longest := longestInventoryLineWithoutColorCodes(items)

	var getItemName func(item foundation.ItemForUI, isEquipped bool) string

	if !u.isRightPanelWidthAtLeast(longest) {
		if u.getRightPanelWidth() == 0 {
			return
		}
		getItemName = func(item foundation.ItemForUI, isEquipped bool) string {
			itemIcon := item.GetIcon().WithFg(u.uiTheme.GetInventoryItemColor(item.GetCategory())).WithBg(u.uiTheme.GetUIColor(UIColorUIBackground))
			if isEquipped {
				itemIcon = itemIcon.Reversed()
			}
			iconString := IconAsString(itemIcon)
			return iconString
		}
	} else {
		getItemName = func(item foundation.ItemForUI, isEquipped bool) string {
			nameWithColorsAndShortcut := item.InventoryNameWithColorsAndShortcut(u.uiTheme.GetInventoryItemColorCode(item.GetCategory()))
			if isEquipped {
				nameWithColorsAndShortcut = nameWithColorsAndShortcut[:2] + "+" + nameWithColorsAndShortcut[3:]
			}
			appendString := RightPadColored(nameWithColorsAndShortcut, longest)
			return appendString
		}
	}

	var asString []string
	for _, item := range items {
		isEquipped := u.game.IsEquipped(item)
		appendString := getItemName(item, isEquipped)
		asString = append(asString, appendString)
	}
	u.rightPanel.SetText("\n" + strings.Join(asString, "\n"))
}

func IconAsString(icon textiles.TextIcon) string {
	code := textiles.RGBAToColorCodes(icon.Fg, icon.Bg)
	return fmt.Sprintf("%s%s[-:-]", code, string(icon.Char))
}

func (u *UI) UpdateVisibleActors() {
	visibleEnemies := u.game.GetVisibleActors()
	//longest := longestInventoryLineWithoutColorCodes(visibleEnemies)
	var asString []string
	for _, enemy := range visibleEnemies {
		icon := u.getIconForActor(enemy)
		iconColor := textiles.RGBAToFgColorCode(icon.Fg)
		iconString := fmt.Sprintf("%s%s[-]", iconColor, string(icon.Char))
		hp, hpMax := enemy.GetHitPoints(), enemy.GetHitPointsMax()
		asPercent := float64(hp) / float64(hpMax)

		hallucinating := u.isPlayerHallucinating()
		if hallucinating {
			asPercent = rand.Float64()
		}
		barIcon := '*'
		if enemy.HasFlag(special.FlagSleep) {
			barIcon = 'z'
		} else if !enemy.HasFlag(special.FlagAwareOfPlayer) {
			barIcon = '?'
		}
		hpBarString := fmt.Sprintf("[%s]", u.RuneBarFromPercent(barIcon, asPercent, 5))
		name := enemy.Name()

		enemyLine := fmt.Sprintf(" %s %s %s", iconString, hpBarString, name)
		asString = append(asString, enemyLine)
	}
	u.lowerRightPanel.SetText(strings.Join(asString, "\n"))
}

func (u *UI) FullColorBarFromPercent(currentVal, maxVal, width int) string {
	percent := float64(currentVal) / float64(maxVal)
	colorChangeIndex := int(math.Round(percent * float64(width)))
	white := u.uiTheme.GetColorByName("White")
	colorCode := textiles.RGBAToColorCodes(u.uiTheme.GetColorByName("green_4"), white)
	if percent < 0.50 {
		colorCode = textiles.RGBAToColorCodes(u.uiTheme.GetColorByName("dark_red_1"), white)
	} else if percent < 0.75 {
		colorCode = textiles.RGBAToColorCodes(u.uiTheme.GetColorByName("yellow_4"), u.uiTheme.GetColorByName("Black"))
	}
	darkGrayCode := textiles.RGBAToColorCodes(u.uiTheme.GetColorByName("dark_gray_3"), white)

	valString := fmt.Sprintf("%d/%d", currentVal, maxVal)
	xForCenter := (width - len(valString)) / 2
	prefix := strings.Repeat(" ", xForCenter)
	suffix := strings.Repeat(" ", width-len(valString)-xForCenter)
	barString := fmt.Sprintf("%s%s%s", prefix, valString, suffix)

	if colorChangeIndex > len(barString) {
		colorChangeIndex = len(barString) - 1
	} else if colorChangeIndex < 0 {
		colorChangeIndex = 0
	}
	barString = colorCode + barString[:colorChangeIndex] + darkGrayCode + barString[colorChangeIndex:] + "[-:-]"
	return barString
}

func (u *UI) RuneBarWithColor(icon rune, fgColorName, bgColorName string, current, max int) string {
	colorCode := textiles.RGBAToColorCodes(u.uiTheme.GetColorByName(fgColorName), u.uiTheme.GetColorByName(bgColorName))
	darkGrayCode := textiles.RGBAToFgColorCode(u.uiTheme.GetColorByName("dark_gray_3"))
	return colorCode + strings.Repeat(string(icon), current) + "[-:-]" + darkGrayCode + strings.Repeat(" ", max-current) + "[-]"
}

func (u *UI) RuneBarFromPercent(icon rune, percent float64, width int) string {
	repeats := int(math.Round(percent * float64(width)))
	colorCode := textiles.RGBAToFgColorCode(u.uiTheme.GetColorByName("Green_1"))
	if percent < 0.50 {
		colorCode = textiles.RGBAToFgColorCode(u.uiTheme.GetColorByName("Red_1"))
	} else if percent < 0.75 {
		colorCode = textiles.RGBAToFgColorCode(u.uiTheme.GetColorByName("Yellow_1"))
	}
	return colorCode + strings.Repeat(string(icon), repeats) + "[-]" + strings.Repeat(" ", width-repeats)
}
func (u *UI) isStatusBarMultiLine() bool {
	_, h := u.application.GetScreenSize()
	_, hNeeded := u.settings.GetMinTerminalSize()
	return h >= hNeeded+1
}
func (u *UI) UpdateStats() {
	statusValues := u.game.GetHudStats()
	flags := u.game.GetHudFlags()
	if len(statusValues) == 0 {
		return
	}

	equippedItem, isEquipped := u.game.GetItemInMainHand()

	itemName := "| none |"
	if isEquipped {
		itemName = "| " + equippedItem.LongNameWithColors(textiles.RGBAToFgColorCode(u.uiTheme.GetInventoryItemColor(equippedItem.GetCategory()))) + " |"
	}

	multiLine := u.isStatusBarMultiLine()

	statusStr := u.getSingleLineStatus(statusValues, flags, multiLine, itemName)

	if multiLine {
		hp := statusValues[foundation.HudHitPoints]
		hpMax := statusValues[foundation.HudHitPointsMax]

		playerBar := u.FullColorBarFromPercent(hp, hpMax, 11)
		hpBarStr := fmt.Sprintf("HP [%s]", playerBar)

		fatigueCurrent := statusValues[foundation.HudFatiguePoints]
		fatigueMax := statusValues[foundation.HudFatiguePointsMax]

		// display as bar
		fatigueBarContent := u.RuneBarWithColor('!', "light_blue_1", "light_blue_5", fatigueCurrent, fatigueMax)
		fpBarStr := fmt.Sprintf("FP [%s]", fatigueBarContent)

		longFlags := FlagStringLong(flags)

		width, _ := u.application.GetScreenSize()

		mapFriendlyName := u.game.GetMapDisplayName()

		lineTwo := fmt.Sprintf("%s %s %s %s", hpBarStr, fpBarStr, longFlags, mapFriendlyName)

		if cview.TaggedStringWidth(lineTwo) > width {
			shortFlags := FlagStringShort(flags)
			lineTwo = fmt.Sprintf("%s %s %s", hpBarStr, fpBarStr, shortFlags)
		}

		lineTwo = expandToWidth(lineTwo, width)

		statusStr = fmt.Sprintf("%s\n%s", lineTwo, statusStr)
	}

	u.statusBar.SetText(fmt.Sprintf("[::r]%s[-:-:-]", statusStr))

	if !u.isAnimationFrame {
		u.lastHudStats = statusValues
	}
}

func FlagStringLong(flags map[special.ActorFlag]int) string {
	flagOrder := special.AllFlagsExceptGoldOrdered()
	var flagStrings []string
	for _, flag := range flagOrder {
		if count, ok := flags[flag]; ok {
			var flagLine string
			if count > 1 {
				flagLine = fmt.Sprintf("%s(%d)", flag.String(), count)
			} else {
				flagLine = fmt.Sprintf("%s", flag.String())
			}

			flagStrings = append(flagStrings, flagLine)
		}
	}
	return strings.Join(flagStrings, " | ")
}

func FlagStringShort(flags map[special.ActorFlag]int) string {
	flagOrder := special.AllFlagsExceptGoldOrdered()
	var flagStrings []string
	for _, flag := range flagOrder {
		if count, ok := flags[flag]; ok {
			var flagLine string
			if count > 1 {
				flagLine = fmt.Sprintf("%s(%d)", flag.StringShort(), count)
			} else {
				flagLine = fmt.Sprintf("%s", flag.StringShort())
			}

			flagStrings = append(flagStrings, flagLine)
		}
	}
	return strings.Join(flagStrings, " ")
}

func mapStrings(listOfStrings []string, mapper func(arg string) string) []string {
	var mapped []string
	for _, s := range listOfStrings {
		mapped = append(mapped, mapper(s))
	}
	return mapped
}
func (u *UI) colorIfDiff(statStr string, stat foundation.HudValue, currentValue int) string {
	lastValue, ok := u.lastHudStats[stat]
	if !ok {
		return statStr
	}
	if lastValue == currentValue {
		return statStr
	}
	hiCode := textiles.RGBAToFgColorCode(u.uiTheme.GetColorByName("Yellow"))
	return fmt.Sprintf("%s%s[-]", hiCode, statStr)
}
func (u *UI) getSingleLineStatus(statusValues map[foundation.HudValue]int, flags map[special.ActorFlag]int, multiLine bool, equippedItem string) string {

	damageResistance := statusValues[foundation.HudDamageResistance]
	armorStr := fmt.Sprintf("DR: %-3d", damageResistance)
	armorStr = u.colorIfDiff(armorStr, foundation.HudDamageResistance, damageResistance)

	var statusStr string
	if !multiLine {
		hp := statusValues[foundation.HudHitPoints]
		hpMax := statusValues[foundation.HudHitPointsMax]
		hpValString := fmt.Sprintf("%d/%d", hp, hpMax)
		hpStr := fmt.Sprintf("HP: %-7s", hpValString)
		hpStr = u.colorIfDiff(hpStr, foundation.HudHitPoints, hp)

		fatigueCurrent := statusValues[foundation.HudFatiguePoints]
		fatigueMax := statusValues[foundation.HudFatiguePointsMax]
		fpValString := fmt.Sprintf("%d/%d", fatigueCurrent, fatigueMax)
		fpStr := fmt.Sprintf("FP: %-7s", fpValString)
		fpStr = u.colorIfDiff(fpStr, foundation.HudFatiguePoints, fatigueCurrent)

		flagString := FlagStringShort(flags)

		statusStr = fmt.Sprintf("%s %s %s %s %s", hpStr, fpStr, equippedItem, armorStr, flagString)
	} else {
		statusStr = fmt.Sprintf("%s %s", equippedItem, armorStr)
	}

	width, _ := u.application.GetScreenSize()
	statusStr = expandToWidth(statusStr, width)
	return statusStr
}

func expandToWidth(statusStr string, width int) string {
	statusWidth := cview.TaggedStringWidth(statusStr)
	if statusWidth < width {
		statusStr = fxtools.RightPadCount(statusStr, width-statusWidth)
	}
	return statusStr
}
func (u *UI) openCharSheet() {
	charSheet := NewCharsheetViewer(u.game.GetPlayerName(), u.game.GetPlayerCharSheet(), u.closeModal)
	charSheet.SetConfirmer(u)
	originalInputCapture := charSheet.GetInputCapture()
	charSheet.SetInputCapture(u.directionalWrapper(originalInputCapture))

	u.makeCenteredModal(charSheet, 80, 25)
}
func (u *UI) openInventory(items []foundation.ItemForUI) *TextInventory {
	inventory := NewTextInventory(u.game.IsPlayerOverEncumbered)
	inventory.SetLineColor(u.uiTheme.GetInventoryItemColor)
	inventory.SetEquippedTest(u.game.IsEquipped)
	inventory.SetStyle(u.uiTheme.defaultStyle)

	inventory.SetItems(items)

	inventory.SetCloseHandler(u.closeModal)
	u.pages.AddPanel("modal", inventory, true, true)
	u.pages.ShowPanel("modal")
	u.lockFocusToPrimitive(inventory)

	originalInputCapture := inventory.GetInputCapture()
	inventory.SetInputCapture(u.directionalWrapperWithoutAlphabet(originalInputCapture))

	//u.makeTopRightModal(panelName, list, len(inventoryItems), longestItem)
	return inventory
}

func (u *UI) OpenInventoryForManagement(items []foundation.ItemForUI) {
	inv := u.openInventory(items)
	inv.SetTitle("Inventory")
	inv.SetDefaultSelection(func(item foundation.ItemForUI) {
		if item.IsEquippable() {
			u.game.EquipToggle(item)
		} else {
			inv.Close()
			u.game.PlayerApplyItem(item)
		}
	})
	inv.SetShiftSelection(u.game.DropItemFromInventory)
	inv.SetControlSelection(u.game.PlayerApplyItem)

	inv.SetCloseOnControlSelection(true)
	inv.SetCloseOnShiftSelection(true)
}
func (u *UI) OpenInventoryForSelection(itemStacks []foundation.ItemForUI, prompt string, onSelected func(item foundation.ItemForUI)) {
	u.rightPanel.Clear()
	inv := u.openInventory(itemStacks)
	inv.SetSelectionMode()
	inv.SetTitle(prompt)
	inv.SetDefaultSelection(onSelected)
	inv.SetCloseOnSelection(true)
	inv.SetAfterClose(func() {
		u.UpdateInventory()
	})
}

type InputPrimitive interface {
	cview.Primitive
	SetInputCapture(f func(event *tcell.EventKey) *tcell.EventKey)
	GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey
}

func (u *UI) closeModal() {
	u.pages.RemovePanel("modal")
	u.pages.SetCurrentPanel("main")
	u.resetFocusToMain()
}

func (u *UI) defaultFocusHandler(p cview.Primitive) bool {
	if p == u.mainGrid || p == u.mapWindow {
		return true
	}
	return false
}

func (u *UI) makeCenteredModal(modal InputPrimitive, w, h int) {
	w = w + 2
	h = h + 2
	screenW, screenH := u.application.GetScreen().Size()
	if h > screenH {
		h = screenH
		w = w + 1 // scrollbar
	}
	if w > screenW {
		w = screenW
	}
	x, y := (screenW-w)/2, (screenH-h)/2
	modal.SetRect(x, y, w, h)
	/*
		originalInputCapture := modal.GetInputCapture()
		modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			command := u.getAdvancedTargetingCommandForKey(toUIKey(event))
			if command == "target_cancel" {
				u.closeModal()
			}
			return originalInputCapture(event)
		})

	*/
	u.pages.AddPanel("modal", modal, false, true)
	u.lockFocusToPrimitive(modal)
}
func (u *UI) makeSideBySideModal(primitive, qPrimitive cview.Primitive, contentHeight int, contentWidth int) {
	w, h := u.application.GetScreenSize()
	height := contentHeight + 2
	horizontalSpaceForBorder := 2
	if height > h-4 { // needs scrolling
		height = h - 4
		horizontalSpaceForBorder += 1
	}
	width := min(contentWidth+horizontalSpaceForBorder, w-4)
	modalContainer := wrapPrimitivesSideBySide(primitive, qPrimitive, width, height)

	if inputCapturer, ok := primitive.(InputCapturer); ok {
		inputCapturer.SetInputCapture(u.popOnEscape)
	}

	if inputCapturer, ok := qPrimitive.(InputCapturer); ok {
		inputCapturer.SetInputCapture(u.popOnEscape)
	}
	u.pages.AddPanel("modal", modalContainer, true, true)
	u.pages.ShowPanel("modal")
	u.lockFocusToPrimitive(qPrimitive)
}

func (u *UI) makeTopAndBottomModal(primitive, qPrimitive cview.Primitive) {
	modalContainer := wrapPrimitivesTopToBottom(primitive, qPrimitive)

	if inputCapturer, ok := primitive.(InputCapturer); ok {
		inputCapturer.SetInputCapture(u.popOnEscape)
	}

	if inputCapturer, ok := qPrimitive.(InputCapturer); ok {
		inputCapturer.SetInputCapture(u.popOnEscape)
	}
	u.pages.AddPanel("modal", modalContainer, true, true)
	u.pages.ShowPanel("modal")
	u.lockFocusToPrimitive(qPrimitive)
}
func (u *UI) StartLockpickGame(difficulty foundation.Difficulty, getLockpickCount func() int, removeLockpick func(), onCompletion func(result foundation.InteractionResult)) {
	lockpickGame := NewLockpickGame(rand.Int63(), difficulty, getLockpickCount, removeLockpick, func(result foundation.InteractionResult) {
		u.closeModal()
		onCompletion(result)
	})
	origCapt := lockpickGame.GetInputCapture()
	lockpickGame.SetInputCapture(u.directionalWrapper(origCapt))
	lockpickGame.SetAudioPlayer(u.audioPlayer)
	u.pages.AddPanel("modal", lockpickGame, true, true)
	u.pages.ShowPanel("modal")
	u.lockFocusToPrimitive(lockpickGame)
}
func (u *UI) resetFocusToMain() {
	u.application.SetBeforeFocusFunc(nil)
	u.application.SetFocus(u.mainGrid)
	u.application.SetBeforeFocusFunc(u.defaultFocusHandler)
}
func (u *UI) lockFocusToPrimitive(p cview.Primitive) {
	u.application.SetBeforeFocusFunc(nil)
	u.application.SetFocus(p)
	u.application.SetBeforeFocusFunc(func(p cview.Primitive) bool { return false })
}
func hammingDistance(a, b string) int {
	if utf8.RuneCountInString(a) != utf8.RuneCountInString(b) {
		panic("strings must be same length")
	}
	aRunes := []rune(a)
	bRunes := []rune(b)
	distance := 0
	for i := 0; i < len(aRunes); i++ {
		if aRunes[i] != bRunes[i] {
			distance++
		}
	}
	return distance
}
func (u *UI) StartHackingGame(identifier uint64, difficulty foundation.Difficulty, previousGuesses []string, onCompletion func(previousGuesses []string, success foundation.InteractionResult)) {
	letterCount := 4
	fakeCount := 4
	switch difficulty {
	case foundation.VeryEasy:
		letterCount = 4
		fakeCount = 4
	case foundation.Easy:
		letterCount = 5
		fakeCount = 4
	case foundation.Medium:
		letterCount = 6
		fakeCount = 5
	case foundation.Hard:
		letterCount = 7
		fakeCount = 5
	case foundation.VeryHard:
		letterCount = 8
		fakeCount = 6
	}
	passwordFile := fmt.Sprintf("%d-letter-words.txt", letterCount)
	passFile := path.Join(u.settings.DataRootDir, "wordlists", passwordFile)
	passwords := fxtools.ReadFileAsLines(passFile)
	rnd := rand.New(rand.NewSource(int64(identifier)))

	// shuffle the passwords
	permutedIndexes := rnd.Perm(len(passwords))

	correct := passwords[0]

	var fakes []string
	for i := 0; i < len(passwords)-1; i++ {
		nextPossiblePassword := passwords[permutedIndexes[i+1]]
		distance := hammingDistance(correct, nextPossiblePassword)
		if distance >= letterCount-1 {
			continue
		}
		fakes = append(fakes, nextPossiblePassword)
		if len(fakes) >= fakeCount {
			break
		}
	}

	hackingGame := NewHackingGame(correct, fakes, func(previousGuesses []string, result foundation.InteractionResult) {
		u.closeModal()
		onCompletion(previousGuesses, result)
	})
	originalCapture := hackingGame.GetInputCapture()
	hackingGame.SetInputCapture(u.directionalWrapper(originalCapture))

	hackingGame.SetAudioPlayer(u.audioPlayer)
	hackingGame.SetGuesses(previousGuesses)
	panelName := "modal"

	u.pages.AddPanel(panelName, hackingGame, true, true)
	u.lockFocusToPrimitive(hackingGame)
}

func (u *UI) SetConversationState(starterText string, starterOptions []foundation.MenuItem, chatterSource foundation.ChatterSource, isTerminal bool) {
	u.dialogueIsTerminal = isTerminal

	// text field
	if u.dialogueText == nil {
		textField := cview.NewTextView()
		textField.SetTitle(chatterSource.Name())
		textField.SetBorder(true)
		textField.SetBorderColor(toTcellColor(u.uiTheme.GetColorByName("neon_green_2")))
		textField.SetScrollable(true)
		textField.SetDynamicColors(true)
		textField.SetWordWrap(true)
		u.dialogueText = textField
		if isTerminal {
			u.audioPlayer.PlayCue("ui/terminal_poweron")
		}
	}
	u.dialogueText.SetText(starterText)

	// menu
	if u.dialogueOptions == nil {
		choicesMenu := cview.NewList()
		u.applyListStyle(choicesMenu)
		u.dialogueOptions = choicesMenu
	}

	u.dialogueOptions.SetSelectedFunc(func(index int, listItem *cview.ListItem) {
		action := starterOptions[index]
		if action.CloseMenus {
			u.closeModal()
		}
		action.Action()
	})

	u.makeTopAndBottomModal(u.dialogueText, u.dialogueOptions)
	u.lockFocusToPrimitive(u.dialogueOptions)

	originalCapture := u.dialogueOptions.GetInputCapture()

	if u.settings.DialogueShortcutsAreNumbers {
		u.dialogueOptions.SetInputCapture(u.directionalWrapperWithoutNumbers(originalCapture))
		setListItemsFromMenuItemsWithNumbers(u.dialogueOptions, starterOptions)
	} else {
		u.dialogueOptions.SetInputCapture(u.directionalWrapperWithoutAlphabet(originalCapture))
		setListItemsFromMenuItems(u.dialogueOptions, starterOptions)
	}

}

func (u *UI) directionalWrapperWithoutAlphabet(originalCapture func(event *tcell.EventKey) *tcell.EventKey) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && (event.Rune() >= 'a' && event.Rune() <= 'z') || (event.Rune() >= 'A' && event.Rune() <= 'Z') {
			return originalCapture(event)
		}
		return u.directionalWrapper(originalCapture)(event)
	}
}

func (u *UI) directionalWrapperWithoutNumbers(originalCapture func(event *tcell.EventKey) *tcell.EventKey) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && (event.Rune() >= '0' && event.Rune() <= '9') {
			return originalCapture(event)
		}
		return u.directionalWrapper(originalCapture)(event)
	}
}
func (u *UI) directionalWrapper(originalCapture func(event *tcell.EventKey) *tcell.EventKey) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		command := u.getCommandForKey(toUIKey(event))

		if command == "wait" {
			event = tcell.NewEventKey(tcell.KeyEscape, ' ', tcell.ModNone)
			return originalCapture(event)
		}

		if command == "map_interaction" {
			event = tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone)
			return originalCapture(event)
		}

		direction, possible := directionFromCommand(command)
		if !possible {
			return originalCapture(event)
		}
		if direction == geometry.North {
			event = tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
		} else if direction == geometry.South {
			event = tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
		} else if direction == geometry.West {
			event = tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone)
		} else if direction == geometry.East {
			event = tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone)
		}

		return originalCapture(event)
	}
}

func (u *UI) CloseConversation() {
	if u.dialogueIsTerminal {
		u.audioPlayer.PlayCue("ui/terminal_poweroff")
	}
	u.pages.RemovePanel("conversation")
	u.pages.SetCurrentPanel("main")
	u.dialogueOptions = nil
	u.dialogueText = nil
	u.dialogueIsTerminal = false
}

func (u *UI) OpenMenu(actions []foundation.MenuItem) {
	u.openSimpleMenu(actions)
}

func (u *UI) openSimpleMenu(menuItems []foundation.MenuItem) *cview.List {
	list, longestItem := u.createSimpleMenu(menuItems)

	originalCapture := list.GetInputCapture()
	list.SetInputCapture(u.directionalWrapper(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			u.closeModal()
			return nil
		}
		if originalCapture != nil {
			return originalCapture(event)
		}
		return event
	}))
	u.makeCenteredModal(list, longestItem, len(menuItems))

	return list
}

func (u *UI) createSimpleMenu(menuItems []foundation.MenuItem) (*cview.List, int) {
	list := cview.NewList()
	u.applyListStyle(list)
	list.SetSelectedFunc(func(index int, listItem *cview.ListItem) {
		action := menuItems[index]
		if action.CloseMenus {
			u.closeModal()
		}
		action.Action()
	})

	longestItem := setListItemsFromMenuItems(list, menuItems)
	return list, longestItem
}
func setListItemsFromMenuItemsWithNumbers(list *cview.List, menuItems []foundation.MenuItem) int {
	list.Clear()
	longestItem := 0
	for index, a := range menuItems {
		action := a
		listItem := cview.NewListItem(action.Name)
		// we need the runes 0-9
		asRune := '0' + rune(index+1)
		listItem.SetShortcut(asRune)
		list.AddItem(listItem)
		itemLength := cview.TaggedStringWidth(action.Name) + 4
		longestItem = max(longestItem, itemLength)
	}
	return longestItem
}
func setListItemsFromMenuItems(list *cview.List, menuItems []foundation.MenuItem) int {
	list.Clear()
	longestItem := 0
	for index, a := range menuItems {
		action := a
		shortcut := foundation.ShortCutFromIndex(index)
		listItem := cview.NewListItem(action.Name)
		listItem.SetShortcut(shortcut)
		list.AddItem(listItem)
		itemLength := cview.TaggedStringWidth(action.Name) + 4
		longestItem = max(longestItem, itemLength)
	}
	return longestItem
}

func (u *UI) ShowMonsterInfo(monster foundation.ActorForUI) {
	monsterNameInternalName := monster.GetInternalName()
	lorePath := path.Join(u.settings.DataRootDir, "lore", "monsters", monsterNameInternalName+".txt")
	panels := cview.NewTabbedPanels()
	panels.SetFullScreen(true)
	panels.SetTabSwitcherDivider("|", "|", "|")
	monsterInfo := monster.GetDetailInfo()
	monsterLore := fxtools.ReadFile(lorePath)
	if len(monsterLore) == 0 {
		u.openTextModal(monsterInfo)
		return
	}
	monsterStats := u.newTextModal(monsterInfo)
	monsterLoreText := u.newTextModal(monsterLore)
	monsterLoreText.SetWrap(true)
	monsterLoreText.SetWordWrap(true)

	panels.AddTab("stats", "Stats", monsterStats)
	panels.AddTab("lore", "Lore", monsterLoreText)
	inputHandler := func(nextTab string) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				panels.SetCurrentTab(nextTab)
				return nil
			}
			return u.popOnEscape(event)
		}
	}
	monsterStats.SetInputCapture(inputHandler("lore"))
	monsterLoreText.SetInputCapture(inputHandler("stats"))
	//panels.SetInputCapture(u.popOnEscape)

	panelName := "monsterInfo"
	u.pages.AddPanel(panelName, panels, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(panels)
}
func (u *UI) getListForPanel(panelName string) (*cview.List, bool) {
	list, exists := u.listTable[panelName]
	return list, exists
}

func (u *UI) popOnEscape(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEscape {
		u.closeModal()
	}
	return event
}

func (u *UI) yesNoReceiver(yes, no func()) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'y' || event.Rune() == 'Y' {
			yes()
			return nil
		}
		if event.Rune() == 'n' || event.Rune() == 'N' {
			no()
			return nil
		}
		return event
	}
}

func (u *UI) popOnAnyKeyWithNotification(currentPage string, onClose func()) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		u.pages.HidePanel(currentPage)
		onClose()
		return nil
	}
}

func (u *UI) popOnSpaceWithNotification(currentPage string, onClose func()) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == ' ' {
			u.pages.HidePanel(currentPage)
			onClose()
			return nil
		}
		return event
	}
}

func (u *UI) ScreenToMap(point geometry.Point) geometry.Point {
	x, y, _, _ := u.mapWindow.GetInnerRect()

	mapX := point.X - x
	mapY := point.Y - y
	return geometry.Point{X: mapX, Y: mapY}
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
		mapPos := u.ScreenToMap(mousePos)
		mapInfo := u.game.GetMapInfo(mapPos)
		if !mapInfo.IsEmpty() {
			u.Print(mapInfo)
		} else {
			u.application.QueueUpdateDraw(u.UpdateLogWindow)
		}
	}
	mapPos := u.ScreenToMap(geometry.Point{X: newX, Y: newY})
	if action == cview.MouseLeftDown {
		u.autoRun = false
		if u.currentMouseX >= u.settings.MapWidth {
			// clicked on right panel
			u.onRightPanelClicked(mousePos)
		} else {
			u.game.PlayerInteractAtPosition(mapPos)
			//u.game.OpenContextMenuFor(mapPos)
		}
		return nil, action
	} else if action == cview.MouseRightDown {
		u.autoRun = false
		actorAt := u.game.ActorAt(mapPos)
		if actorAt != nil {
			u.ShowMonsterInfo(actorAt)
		}

		return nil, action
	}
	return event, action
}

func (u *UI) setupListForUI(panelName string, list *cview.List) {
	u.applyListStyle(list)

	u.listTable[panelName] = list
}

func (u *UI) applyListStyle(list *cview.List) {
	fg := u.uiTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.uiTheme.GetUIColorForTcell(UIColorUIBackground)

	borderFgFocus := u.uiTheme.GetUIColorForTcell(UIColorBorderForegroundFocus)

	list.SetBorder(true)
	list.SetWrapAround(true)
	list.SetHover(true)
	list.ShowSecondaryText(false)

	list.SetScrollBarColor(fg)
	//list.SetHighlightFullLine(true)

	list.SetTitleColor(fg)
	list.SetMainTextColor(fg)
	list.SetSecondaryTextColor(fg)

	list.SetBorderColor(fg)
	list.SetBorderColorFocused(borderFgFocus)

	list.SetBackgroundColor(bg)

	list.SetShortcutColor(fg)

	list.SetSelectedTextColor(bg)
	list.SetSelectedBackgroundColor(fg)
	list.SetForceSelectedTextColor(true)
	//list.SetSelectedTextAttributes(tcell.AttrReverse)
}

func (u *UI) ShowLog() {
	logLines := u.game.GetLog()
	if len(logLines) == 0 {
		u.Print(foundation.Msg("No log entries"))
		return
	}
	var sb strings.Builder
	for i, line := range logLines {
		text := u.ToColoredText(line, 1)
		sb.WriteString(text)
		if i < len(logLines)-1 {
			sb.WriteString("\n")
		}
	}
	textView := u.openTextModal(sb.String())
	textView.ScrollToEnd()
}

type OverlayDrawInfo struct {
	Text       string
	Pos        geometry.Point
	Connectors []geometry.Point
	SourcePos  geometry.Point
}

func (u *UI) ShowActorOverlay() {
	listOfEnemies := u.game.GetVisibleActors()

	if len(listOfEnemies) == 0 {
		u.Print(foundation.Msg("No actors in sight"))
		return
	}
	u.mapOverlay.ClearAll()

	for _, enemy := range listOfEnemies {
		u.mapOverlay.TryAddOverlay(enemy.Position(), enemy.Name(), u.GetMapWindowGridSize(), u.game.IsSomethingInterestingAtLoc)
	}
}

func (u *UI) TryAddChatter(source foundation.ChatterSource, text string) bool {
	windowSize := u.GetMapWindowGridSize()
	if u.mapOverlay.TryAddOverlayColored(source.Position(), text, source.Icon().Fg, windowSize, u.game.IsSomethingInterestingAtLoc) {
		return true
	}
	return false
}

func (u *UI) ShowItemOverlay() {
	listOfItems := u.game.GetVisibleItems()

	if len(listOfItems) == 0 {
		u.Print(foundation.Msg("No items in sight"))
		return
	}
	u.mapOverlay.ClearAll()

	for _, items := range listOfItems {
		u.mapOverlay.TryAddOverlay(items.Position(), items.Name(), u.GetMapWindowGridSize(), u.game.IsSomethingInterestingAtLoc)
	}
}

func (u *UI) ShowVisibleActors() {
	listOfEnemies := u.game.GetVisibleActors()
	if len(listOfEnemies) == 0 {
		u.Print(foundation.Msg("No actors in sight"))
		return

	}
	tableRows := make([]fxtools.TableRow, len(listOfEnemies)+1)
	header := fxtools.NewTableRow(
		"Icon",
		"Name",
		"HP",
		"Dmg",
		"DR",
	)
	tableRows[0] = header
	for i, enemy := range listOfEnemies {
		row := fxtools.NewTableRow(
			string(u.getIconForActor(enemy).Char),
			enemy.Name(),
			strconv.Itoa(enemy.GetHitPoints()),
			enemy.GetMainHandDamageAsString(),
			strconv.Itoa(enemy.GetDamageResistance()),
		)
		tableRows[i+1] = row
	}
	layout := fxtools.TableLayout(tableRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft, fxtools.AlignRight, fxtools.AlignRight, fxtools.AlignRight})
	u.OpenTextWindow(strings.Join(layout, "\n"))
}

func (u *UI) ShowVisibleItems() {
	listOfItems := u.game.GetVisibleItems()
	if len(listOfItems) == 0 {
		u.Print(foundation.Msg("No items in sight"))
		return

	}
	var infoTexts strings.Builder
	for i, item := range listOfItems {
		info := item.GetListInfo()
		info = fmt.Sprintf("%c - %s", item.GetIcon().Char, info)
		infoTexts.WriteString(info)
		if i < len(listOfItems)-1 {
			infoTexts.WriteString("\n")
		}
	}
	u.OpenTextWindow(infoTexts.String())
}

func (u *UI) onTerminalResized(width int, height int) {
	if !u.sentUIRunning {
		tty, isTerm := u.application.GetScreen().Tty()
		if isTerm {
			tty.Write([]byte{0x1B, 0x3E})
		}

		u.game.UIRunning()
		u.sentUIRunning = true
		return
	}

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
	if u.sentUIReady {
		u.application.QueueUpdateDraw(func() {
			u.UpdateLogWindow()
			u.UpdateInventory()
			u.UpdateStats()
		})
	}
}

func toTcellColor(rgba color.RGBA) tcell.Color {
	return tcell.NewRGBColor(int32(rgba.R), int32(rgba.G), int32(rgba.B))
}

func NewTextUI(settings *foundation.Configuration) *UI {
	u := &UI{
		targetingTiles: make(map[geometry.Point]bool),
		animator:       NewAnimator(),
		audioPlayer:    audio.NewPlayer(),
		listTable:      make(map[string]*cview.List),
		cursorStyle:    tcell.CursorStyleSteadyBlock,
		gamma:          1.0,
		settings:       settings,
		keyTable:       make(map[KeyLayer]map[UIKey]string),
		lastFrameIcons: make(map[geometry.Point]rune),
		lastFrameStyle: make(map[geometry.Point]tcell.Style),
	}
	u.initCoreUI()
	u.initAudio()
	return u
}

func runeToDirection(r rune) geometry.CompassDirection {
	switch r {
	case '8':
		fallthrough
	case 'w':
		return geometry.North
	case '2':
		fallthrough
	case 's':
		return geometry.South
	case '4':
		fallthrough
	case 'a':
		return geometry.West
	case '6':
		fallthrough
	case 'd':
		return geometry.East
	case '7':
		return geometry.NorthWest
	case '9':
		return geometry.NorthEast
	case '1':
		return geometry.SouthWest
	case '3':
		return geometry.SouthEast
	}
	return geometry.North
}

func upperRuneToDirection(r rune) geometry.CompassDirection {
	switch r {
	case 'W':
		return geometry.North
	case 'S':
		return geometry.South
	case 'A':
		return geometry.West
	case 'D':
		return geometry.East
	}
	return geometry.North
}

func directionToRune(dir geometry.CompassDirection) rune {
	switch dir {
	case geometry.North:
		return '8'
	case geometry.South:
		return '2'
	case geometry.West:
		return '4'
	case geometry.East:
		return '6'
	case geometry.NorthWest:
		return '7'
	case geometry.NorthEast:
		return '9'
	case geometry.SouthWest:
		return '1'
	case geometry.SouthEast:
		return '3'
	}
	return 'w'
}

func (u *UI) mapLookup(loc geometry.Point) (textiles.TextIcon, bool) {
	if u.game.IsVisibleToPlayer(loc) {
		mapIcon, found := u.visibleLookup(loc)
		return mapIcon, found
	} else if u.game.IsExplored(loc) {
		mapIcon, found := u.exploredLookup(loc)
		mapIcon.Fg = desaturate(mapIcon.Fg)
		mapIcon.Bg = desaturate(mapIcon.Bg)
		return mapIcon, found
	}
	return textiles.TextIcon{}, false
}

func desaturate(fg color.RGBA) color.RGBA {
	gray := uint8((uint16(fg.R) + uint16(fg.G) + uint16(fg.B)) / 3)
	return color.RGBA{R: gray, G: gray, B: gray, A: fg.A}
}
func (u *UI) getMapTileBackgroundColor(loc geometry.Point) color.RGBA {
	icon := u.getIconForMap(loc)
	return applyLightToMaterial(u.lightAt(loc), icon.Bg).ToRGBA()
}

func (u *UI) lightAt(loc geometry.Point) fxtools.HDRColor {
	if u.isAnimationFrame {
		return u.game.LightAt(loc).Add(u.animator.lightAt(loc))
	}
	return u.game.LightAt(loc)
}

func (u *UI) exploredLookup(loc geometry.Point) (textiles.TextIcon, bool) {
	conditionalBackgroundWrapper := func(i textiles.TextIcon) textiles.TextIcon {
		if i.HasBackground() {
			return i
		}
		mapIcon := u.getIconForMap(loc)
		return i.WithBg(mapIcon.Bg)
	}
	var icon textiles.TextIcon
	objectAtLoc := u.game.ObjectAt(loc)
	switch objectAtLoc != nil {
	case true:
		icon = conditionalBackgroundWrapper(objectAtLoc.Icon())
	default:
		icon = u.getIconForMap(loc)
	}
	fgWithLight, bgWithLight := u.ApplyLighting(loc, icon.Fg, icon.Bg)
	icon.Fg = fgWithLight
	icon.Bg = bgWithLight
	return icon, true
}

func (u *UI) visibleLookup(loc geometry.Point) (textiles.TextIcon, bool) {
	conditionalBackgroundWrapper := func(i textiles.TextIcon) textiles.TextIcon {
		if i.HasBackground() {
			return i
		}
		mapIcon := u.getIconForMap(loc)
		return i.WithBg(mapIcon.Bg)
	}
	var icon textiles.TextIcon
	entityType := u.game.TopEntityAt(loc)
	switch entityType {
	case foundation.EntityTypeActor:
		actor := u.game.ActorAt(loc)
		icon = u.getIconForActor(actor)
	case foundation.EntityTypeDownedActor:
		actor := u.game.DownedActorAt(loc)
		icon = u.getIconForActor(actor)
	case foundation.EntityTypeItem:
		item := u.game.ItemAt(loc)
		icon = conditionalBackgroundWrapper(item.GetIcon())
	case foundation.EntityTypeObject:
		object := u.game.ObjectAt(loc)
		icon = conditionalBackgroundWrapper(object.Icon())
	default:
		icon = u.getIconForMap(loc)
	}
	fgWithLight, bgWithLight := u.ApplyLighting(loc, icon.Fg, icon.Bg)
	icon.Fg = fgWithLight
	icon.Bg = bgWithLight
	return icon, true
}
func (u *UI) ApplyLighting(p geometry.Point, fg, bg color.RGBA) (color.RGBA, color.RGBA) {
	lightAtCell := u.lightAt(p)
	fgWithLight := applyLightToMaterial(lightAtCell, fg)
	bgWithLight := applyLightToMaterial(lightAtCell, bg)
	return fgWithLight.ToRGBA(), bgWithLight.ToRGBA()
}

func applyLightToMaterial(lightAtCell fxtools.HDRColor, material color.RGBA) fxtools.HDRColor {
	return lightAtCell.Multiply(fxtools.NewRGBColorFromBytes(material.R, material.G, material.B))
}

func (u *UI) ShowCharacterSheet() {
	var attributeActions []foundation.MenuItem

	statList := []dice_curve.Stat{
		dice_curve.Strength,
		dice_curve.Dexterity,
		dice_curve.Intelligence,
		dice_curve.Health,
		dice_curve.BasicSpeed,
		dice_curve.HitPoints,
		dice_curve.FatiguePoints,
		dice_curve.Perception,
		dice_curve.Will,
	}
	for _, s := range statList {
		//statInList := s
		attributeActions = append(attributeActions, foundation.MenuItem{
			Name: fmt.Sprintf("+ %s", s.ToString()),
			Action: func() {
				//u.game.IncreaseAttributeLevel(statInList)
				u.showCharacterActions(attributeActions)
			},
		})
	}

	var skillActions []foundation.MenuItem

	skillList := []dice_curve.SkillName{
		dice_curve.SkillNameBrawling,
		dice_curve.SkillNameMeleeWeapons,
		dice_curve.SkillNameShield,
		dice_curve.SkillNameThrowing,
		dice_curve.SkillNameMissileWeapons,
	}

	for _, s := range skillList {
		skillInList := s
		skillActions = append(skillActions, foundation.MenuItem{
			Name: fmt.Sprintf("+ %s", skillInList),
			Action: func() {
				//u.game.IncreaseSkillLevel(skillInList)
				u.showCharacterActions(skillActions)
			},
		})

	}

	baseActions := []foundation.MenuItem{
		{
			Name:       "Close",
			Action:     func() {},
			CloseMenus: true,
		},
		{
			Name: "Change base Attributes",
			Action: func() {
				u.showCharacterActions(attributeActions)
			},
			CloseMenus: true,
		},

		{
			Name: "Change Skills",
			Action: func() {
				u.showCharacterActions(skillActions)
			},
			CloseMenus: true,
		},
	}

	u.showCharacterActions(baseActions)
}

func (u *UI) showCharacterActions(actions []foundation.MenuItem) {
	list := cview.NewList()
	u.applyListStyle(list)

	list.SetSelectedFunc(func(index int, listItem *cview.ListItem) {
		action := actions[index]
		list.HideContextMenu(func(primitive cview.Primitive) {
			u.application.SetFocus(primitive)
		})
		if action.CloseMenus {
			u.closeModal()
		}
		action.Action()
	})

	longestItem := 0
	for index, a := range actions {
		action := a
		shortcut := foundation.ShortCutFromIndex(index)
		listItem := cview.NewListItem(action.Name)
		listItem.SetShortcut(shortcut)
		list.AddItem(listItem)
		itemLength := len(action.Name) + 4
		longestItem = max(longestItem, itemLength)
	}

	textView, playerInfo := u.charSheetView()
	w, h := widthAndHeightFromString(playerInfo)
	u.makeSideBySideModal(textView, list, h, w)
}

func (u *UI) charSheetView() (*cview.TextView, string) {
	playerInfo := u.game.GetCharacterSheet()
	textView := cview.NewTextView()
	textView.SetBorder(true)

	fg := u.uiTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.uiTheme.GetUIColorForTcell(UIColorUIBackground)

	textView.SetTextColor(fg)
	textView.SetBorderColor(fg)
	textView.SetBackgroundColor(bg)
	textView.SetBorderColorFocused(fg)
	textView.SetBorderColor(fg) // TODO: darker style here..
	u.setColoredText(textView, playerInfo)
	return textView, playerInfo
}

func (u *UI) onRightPanelClicked(clickPos geometry.Point) {
	itemIndex := clickPos.Y - 1

	inv := u.game.GetInventoryForUI()

	if itemIndex < 0 || itemIndex >= len(inv) {
		return
	}

	item := inv[itemIndex]

	if item.IsEquippable() {
		u.game.EquipToggle(item)
	} else {
		u.game.PlayerApplyItem(item)
	}
}

func RightPadColored(s string, pLen int) string {
	return s + strings.Repeat(" ", pLen-cview.TaggedStringWidth(s))
}

func (u *UI) GetAnimExplosion(hitPositions []geometry.Point, done func()) foundation.Animation {
	white := u.uiTheme.GetColorByName("white")
	yellow := u.uiTheme.GetColorByName("yellow_3")
	red := u.uiTheme.GetColorByName("red_8")
	lightGray := u.uiTheme.GetColorByName("light_gray_5")
	darkGray := u.uiTheme.GetColorByName("dark_gray_3")
	frames := []textiles.TextIcon{
		{Char: '.', Fg: white},
		{Char: '∙', Fg: white},
		{Char: '*', Fg: white},
		{Char: '*', Fg: yellow},
		{Char: '*', Fg: red},
		{Char: '☼', Fg: lightGray},
		{Char: '☼', Fg: darkGray},
	}
	return u.GetAnimTiles(hitPositions, frames, done)
}

func (u *UI) GetAnimUncloakAtPosition(actor foundation.ActorForUI, uncloakLocation geometry.Point) (foundation.Animation, int) {
	actorIcon := u.getIconForActor(actor)
	tileIcon := u.getIconForMap(uncloakLocation)
	lightGray := u.uiTheme.GetColorByName("light_gray_5")
	darkGray := u.uiTheme.GetColorByName("dark_gray_3")
	black := u.uiTheme.GetColorByName("Black")
	frames := []textiles.TextIcon{
		tileIcon,
		tileIcon.WithFg(lightGray),
		tileIcon.WithFg(lightGray),
		tileIcon.WithFg(darkGray),
		tileIcon.WithFg(darkGray),
		tileIcon.WithFg(black),
		tileIcon.WithFg(black),
		actorIcon.WithFg(black),
		actorIcon.WithFg(darkGray),
		actorIcon.WithFg(darkGray),
		actorIcon.WithFg(lightGray),
		actorIcon.WithFg(lightGray),
		actorIcon,
	}
	uncloakAnim := u.GetAnimTiles([]geometry.Point{uncloakLocation}, frames, nil)
	return uncloakAnim, len(frames)
}

func (u *UI) remapCommand(layer KeyLayer, command string) {
	u.Print(foundation.Msg("Press the key you want to bind to this command"))
	key := u.getPressedKey()
	u.keyTable[layer][key] = command
	u.Print(foundation.Msg(fmt.Sprintf("Bound %s to %s", key.name, command)))
}
func (u *UI) OpenKeyMapper(layer KeyLayer) {
	var commandMenu []foundation.MenuItem

	for key, c := range u.keyTable[layer] {
		command := c
		line := fmt.Sprintf("%s - %s", key.name, command)
		commandMenu = append(commandMenu, foundation.MenuItem{
			Name: line,
			Action: func() {
				u.remapCommand(layer, command)
				u.OpenKeyMapper(layer)
			},
		})
	}

	u.OpenMenu(commandMenu)
}

func (u *UI) ShowHelpScreen() {
	u.ShowTextFile(path.Join(u.settings.DataRootDir, "help.txt"))
}

func (u *UI) getCommandForKey(key UIKey) string {
	if command, ok := u.keyTable[KeyLayerMain][key]; ok {
		return command
	}
	//println("No command found for key %s", key.String())
	return ""
}

func (u *UI) getDirectionalTargetingCommandForKey(key UIKey) string {
	if command, ok := u.keyTable[KeyLayerDirectionalTargeting][key]; ok {
		return command
	}
	//println("No command found for key %s", key.String())
	return ""
}

func (u *UI) getAdvancedTargetingCommandForKey(key UIKey) string {
	if command, ok := u.keyTable[KeyLayerAdvancedTargeting][key]; ok {
		return command
	}
	//println("No command found for key %s", key.String())
	return ""
}

func (u *UI) updateLastFrame() {
	// iterate the map and force and update of the last frame
	for y := 0; y < u.settings.MapHeight; y++ {
		for x := 0; x < u.settings.MapWidth; x++ {
			pos := geometry.Point{X: x, Y: y}
			u.renderMapPosition(pos, false, u.uiTheme.GetMapDefaultStyle())
		}
	}
}

func (u *UI) Queue(f func()) {
	u.application.QueueUpdate(f)
}

func (u *UI) GetKeysForCommandAsString(layer KeyLayer, command string) string {
	var keys []string
	for key, c := range u.keyTable[layer] {
		if c == command && key.name != "" {
			keys = append(keys, key.name)
		}
	}
	if len(keys) == 0 {
		return ""
	}
	slices.SortStableFunc(keys, func(i, j string) int {
		return cmp.Compare(i, j)
	})
	return strings.Join(keys, ", ")
}

func (u *UI) GetKeysForCommandAsPrettyString(layer KeyLayer, command string) string {
	var keys []string
	for key, c := range u.keyTable[layer] {
		if c == command && key.name != "" {
			keys = append(keys, key.name)
		}
	}
	if len(keys) == 0 {
		return ""
	}
	slices.SortStableFunc(keys, func(i, j string) int {
		return cmp.Compare(i, j)
	})

	for i, k := range keys {
		keys[i] = fmt.Sprintf("[%s]", k)
	}
	return strings.Join(keys, ", ")
}

func (u *UI) joystickLoop() {
	updateRate := time.Millisecond * 16
	js, _ := joystick.Open(0)

	var lastAxisData []int8
	var newAxisData []int8
	var lastButtonData uint32
	analogDeadzone := int8(6)
	leftThumbAxisXIndex := 0
	leftThumbAxisYIndex := 1
	rightThumbAxisXIndex := 2
	rightThumbAxisYIndex := 3
	for {
		if js == nil {
			js, _ = joystick.Open(0)
		}
		if js != nil {
			state, err := js.Read()
			newAxisData = make([]int8, len(state.AxisData))
			for index, value := range state.AxisData {
				newAxisData[index] = int8(value / 256)
			}

			if newAxisData[leftThumbAxisXIndex] < analogDeadzone && newAxisData[leftThumbAxisXIndex] > -analogDeadzone {
				newAxisData[leftThumbAxisXIndex] = 0
			}
			if newAxisData[leftThumbAxisYIndex] < analogDeadzone && newAxisData[leftThumbAxisYIndex] > -analogDeadzone {
				newAxisData[leftThumbAxisYIndex] = 0
			}
			if newAxisData[rightThumbAxisXIndex] < analogDeadzone && newAxisData[rightThumbAxisXIndex] > -analogDeadzone {
				newAxisData[rightThumbAxisXIndex] = 0
			}
			if newAxisData[rightThumbAxisYIndex] < analogDeadzone && newAxisData[rightThumbAxisYIndex] > -analogDeadzone {
				newAxisData[rightThumbAxisYIndex] = 0
			}
			if err == nil {
				if slices.Equal(lastAxisData, newAxisData) && lastButtonData == state.Buttons {
					continue
				} else {
					u.application.QueueEvent(NewJoyStickEvent(newAxisData, state.Buttons))
					lastAxisData = slices.Clone(newAxisData)
					lastButtonData = state.Buttons
				}
			} else {
				js = nil
			}
		}
		time.Sleep(updateRate)
	}
}

func (u *UI) handleUnknownEvent(event tcell.Event) tcell.Event {
	switch e := event.(type) {
	case *EventJoy:
		command := u.getCommandForJoystick(e)
		u.executePlayerCommand(command)
		u.Print(foundation.Msg(fmt.Sprintf("Joystick event: %v, %v", e.Buttons, e.AxisData)))
	}
	return event
}

func (u *UI) getCommandForJoystick(e *EventJoy) string {
	if e.IsButtonDown(0) {
		return "confirm"
	}
	return ""
}

func (u *UI) openPipBoy() {
	pip := NewPipBoy()
	u.pages.AddPanel("PipBoy", pip, true, true)
	u.application.SetFocus(pip)
	pip.SetOnClose(func() {
		u.pages.HidePanel("PipBoy")
		u.pages.SetCurrentPanel("main")
		u.application.SetFocus(u.mapWindow)
	})
}

func (u *UI) initAudio() {
	go func() {
		u.audioPlayer.LoadCuesFromDir(path.Join(u.settings.DataRootDir, "audio", "weapons"), "")
		u.audioPlayer.LoadCuesFromDir(path.Join(u.settings.DataRootDir, "audio", "ui"), "")
		u.audioPlayer.LoadCuesFromDir(path.Join(u.settings.DataRootDir, "audio", "world"), "")
		enemySfxDir := path.Join(u.settings.DataRootDir, "audio", "critters")
		entries, _ := os.ReadDir(enemySfxDir)
		for _, entry := range entries {
			if entry.IsDir() {
				enemyName := entry.Name()
				u.audioPlayer.LoadCuesFromDir(path.Join(enemySfxDir, enemyName), "critters")
			}
		}
		u.audioPlayer.SoundsLoaded()
	}()

	u.animator.SetAudioCuePlayer(u.audioPlayer)
}

func (u *UI) getIconForMap(loc geometry.Point) textiles.TextIcon {
	colored := u.game.MapAt(loc)
	return textiles.TextIcon{
		Char: colored.Char,
		Fg:   colored.Fg,
		Bg:   colored.Bg,
	}
}

func (u *UI) isModalOpen() bool {
	_, front := u.pages.GetFrontPanel()
	return front != u.mainGrid
}

func (u *UI) waitForAnyKey(screen tcell.Screen) {
	for {
		event := screen.PollEvent()
		switch event.(type) {
		case *tcell.EventKey:
			return
		}
	}
}
func (u *UI) StartWithIntro() {
	u.shouldPlayIntro = true
	u.StartGameLoop()
}

func (u *UI) startIntro() {
	backGroundOverlay := u.addFullScreenTextOverlay()

	u.stateOfIntro = Fading
	FadeToWhite(u.application, u.settings.AnimationDelay, 5)

	u.stateOfIntro = TitleScreen
	backGroundOverlay.SetBackgroundColor(tcell.ColorWhite)
	backGroundOverlay.SetTextColor(tcell.ColorBlack)
	backGroundOverlay.SetText("C O N T R A C T O R\na game by Felix Ruzzoli")

	u.application.Draw(backGroundOverlay)
}

func (u *UI) endIntro() {
	u.stateOfIntro = InGame
	u.pages.RemovePanel("fullscreen")
	u.resetFocusToMain()
}
func (u *UI) addFullScreenTextOverlay() *cview.TextView {
	textView := cview.NewTextView()
	textView.SetBorder(false)
	textView.SetBorderColor(tcell.ColorDefault)
	textView.SetPadding(0, 0, 0, 0)
	textView.SetTextAlign(cview.AlignCenter)
	textView.SetTitleAlign(cview.AlignCenter)
	textView.SetVerticalAlign(cview.AlignMiddle)
	textView.SetWordWrap(true)
	textView.SetDynamicColors(true)
	textView.SetScrollable(false)
	textView.SetScrollBarVisibility(cview.ScrollBarNever)
	textView.SetBackgroundColor(tcell.ColorDefault)
	textView.SetTextColor(tcell.ColorDefault)
	u.pages.AddPanel("fullscreen", textView, true, true)
	u.lockFocusToPrimitive(textView)
	return textView
}

func FadeToWhite(app *cview.Application, animDelay time.Duration, stepSize int) {
	screen := app.GetScreen()
	duration := 10 * time.Millisecond
	var waited time.Duration
	for i := 0; i < 100; i++ {
		if !lightenScreen(screen, stepSize) {
			return
		}
		screen.Show()
		waited = 0
		for waited < animDelay {
			if screen.HasPendingEvent() {
				ev := screen.PollEvent()
				if _, ok := ev.(*tcell.EventKey); ok {
					return
				}
			}

			time.Sleep(duration)
			waited += duration
		}
	}
	return
}

func lightenScreen(screen tcell.Screen, size int) bool {
	w, h := screen.Size()
	workLeft := false
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			workDone := lightenScreenLocation(screen, x, y, int32(size))
			if workDone {
				workLeft = true
			}
		}
	}
	return workLeft
}

func lightenScreenLocation(screen tcell.Screen, x int, y int, amount int32) bool {
	icon, _, style, _ := screen.GetContent(x, y)
	fg, bg, _ := style.Decompose()
	fR, fG, fB := fg.RGB()
	bR, bG, bB := bg.RGB()
	hadWorkLeft := fR < 255 || fG < 255 || fB < 255 || bR < 255 || bG < 255 || bB < 255
	newFG := tcell.NewRGBColor(min(255, fR+amount), min(255, fG+amount), min(255, fB+amount))
	newBG := tcell.NewRGBColor(min(255, bR+amount), min(255, bG+amount), min(255, bB+amount))
	screen.SetContent(x, y, icon, nil, style.Background(newBG).Foreground(newFG))
	return hadWorkLeft
}
