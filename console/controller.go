package console

import (
	"RogueUI/audio"
	"RogueUI/cview"
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/util"
	"cmp"
	"fmt"
	"github.com/0xcafed00d/joystick"
	"github.com/gdamore/tcell/v2"
	"image/color"
	"math"
	"math/rand"
	"path"
	"slices"
	"strings"
	"time"
	"unicode"
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

type UI struct {
	settings *foundation.Configuration
	game     foundation.GameForUI

	audioPlayer *audio.Player

	currentTheme Theme

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

	isMonochrome bool

	listTable map[string]*cview.List

	gameIsReady     bool
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
}

func (u *UI) ShowContainer(name string, containedItems []foundation.ItemForUI, transfer func(item foundation.ItemForUI)) {
	var menuItems []foundation.MenuItem
	closeContainer := func() {
		u.pages.HidePanel("contextMenu")
		u.UpdateLogWindow()
	}
	for _, i := range containedItems {
		item := i
		menuItems = append(menuItems, foundation.MenuItem{
			Name: item.InventoryNameWithColors(RGBAToFgColorCode(u.currentTheme.GetInventoryItemColor(item.GetCategory()))),
			Action: func() {
				closeContainer()
				transfer(item)
			},
			CloseMenus: false,
		})
	}

	menu := u.openSimpleMenu(menuItems)
	menu.SetTitle(name)
	keyForTakeAll := u.GetKeysForCommandAsString(KeyLayerMain, "use")
	u.Print(foundation.HiLite("Press %s to take all items", keyForTakeAll))
	menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeContainer()
			return nil
		}
		uiKey := toUIKey(event)
		command := u.getCommandForKey(uiKey)

		if command == "use" {
			for _, item := range containedItems {
				transfer(item)
			}
			closeContainer()
			return nil
		}
		return event
	})
}

func (u *UI) PlayMusic(fileName string) {
	u.audioPlayer.StopAll()
	u.audioPlayer.Stream(fileName)
}

func (u *UI) OpenVendorMenu(itemsForSale []util.Tuple[foundation.ItemForUI, int], buyItem func(ui foundation.ItemForUI, price int)) {
	var menuItems []foundation.MenuItem
	for _, i := range itemsForSale {
		item := i.GetItem1()
		price := i.GetItem2()
		menuItems = append(menuItems, foundation.MenuItem{
			Name: fmt.Sprintf("%s (%d)", item.InventoryNameWithColors(RGBAToFgColorCode(u.currentTheme.GetInventoryItemColor(item.GetCategory()))), price),
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
	bgColor := u.currentTheme.GetColorByName(colorName)
	return NewCoverAnimation(position, iconAtLocation.WithBg(bgColor), frameCount, done)
}

func (u *UI) HighlightStatChange(stat dice_curve.Stat) {
	//TODO
}

func (u *UI) ShowGameOver(scoreInfo foundation.ScoreInfo, highScores []foundation.ScoreInfo) {
	u.animator.CancelAll()
	u.gameIsOver = true
	u.FadeToBlack()

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

	winMessage := util.ReadFileAsLines(path.Join(u.settings.DataRootDir, "win.txt"))

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

	panelName := "blocker"

	textView.SetInputCapture(u.popOnSpaceWithNotification(panelName, func() {
		u.showHighscoresAndRestart(highScores)
	}))
	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(textView)
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
		fmt.Sprintf("Deepest Level: %d", scoreInfo.MaxLevel),
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

	panelName := "blocker"

	reset := func() {
		u.pages.HidePanel(panelName)
		u.gameIsOver = false
		u.game.Reset()
	}
	textView.SetInputCapture(u.yesNoReceiver(reset, u.application.Stop))
	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(textView)
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

	panelName := "blocker"

	reset := func() {
		u.pages.HidePanel(panelName)
		u.gameIsOver = false
		u.game.Reset()
	}
	textView.SetInputCapture(u.yesNoReceiver(reset, u.application.Stop))
	u.pages.AddPanel(panelName, textView, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(textView)
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
			scoreLine = fmt.Sprintf("[#c9c54d::b]%d. %s: %d Gold, Lvl: %d, %s[-:-:-]", i+1, highScore.PlayerName, highScore.Gold, highScore.MaxLevel, highScore.DescriptiveMessage)
		} else {
			scoreLine = fmt.Sprintf("%d. %s: %d Gold, Lvl: %d, CoD: %s", i+1, highScore.PlayerName, highScore.Gold, highScore.MaxLevel, highScore.DescriptiveMessage)
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

func (u *UI) GetMapWindowGridSize() (int, int) {
	_, _, w, h := u.mapWindow.GetInnerRect()
	return w, h
}
func (u *UI) AfterPlayerMoved(moveInfo foundation.MoveInfo) {
	if moveInfo.Mode == foundation.PlayerMoveModeRun && u.autoRun {
		u.application.QueueEvent(tcell.NewEventKey(tcell.KeyRune, directionToRune(moveInfo.Direction), 64))
	}
}

func (u *UI) GetAnimMove(actor foundation.ActorForUI, old geometry.Point, new geometry.Point) foundation.Animation {
	if u.settings.AnimationsEnabled && u.settings.AnimateMovement {
		return NewMovementAnimation(u.getIconForActor(actor), old, new, u.currentTheme.GetColorByName, nil)
	}
	return nil
}

func (u *UI) getIconForActor(actor foundation.ActorForUI) foundation.TextIcon {
	mapIconHere := u.currentTheme.GetIconForMap(u.game.MapAt(actor.Position()))

	isHallucinating := u.isPlayerHallucinating()
	if isHallucinating {
		randomLetter := rune('A' + rand.Intn(26))
		if rand.Intn(2) == 0 {
			randomLetter = unicode.ToLower(randomLetter)
		}
		return foundation.TextIcon{
			Rune: randomLetter,
			Fg:   u.currentTheme.GetRandomColor(),
			Bg:   mapIconHere.Bg,
		}
	}

	var backGroundColor color.RGBA

	if actor.HasFlag(foundation.FlagHeld) {
		return foundation.TextIcon{
			Rune: actor.Icon(),
			Fg:   u.currentTheme.GetColorByName("Blue"),
			Bg:   u.currentTheme.GetColorByName("White"),
		}
	} else {
		backGroundColor = mapIconHere.Bg
	}
	if !actor.IsAlive() {
		return actor.TextIcon(backGroundColor, u.currentTheme.GetColorByName).WithRune('%').WithFg(u.currentTheme.GetColorByName("Red"))
	}
	return actor.TextIcon(backGroundColor, u.currentTheme.GetColorByName)
}

func (u *UI) isPlayerHallucinating() bool {
	flags := u.game.GetHudFlags()
	_, isHallucinating := flags[foundation.FlagHallucinating]
	return isHallucinating
}

func (u *UI) GetAnimQuickMove(actor foundation.ActorForUI, path []geometry.Point) foundation.Animation {
	if u.settings.AnimationsEnabled && u.settings.AnimateMovement {
		animation := NewMovementAnimation(u.getIconForActor(actor), actor.Position(), path[len(path)-1], u.currentTheme.GetColorByName, nil)
		animation.EnableQuickMoveMode(path)
		return animation
	}
	return nil
}
func (u *UI) GetAnimCover(loc geometry.Point, icon foundation.TextIcon, turns int, done func()) foundation.Animation {
	if u.settings.AnimationsEnabled && u.settings.AnimateMovement {
		return NewCoverAnimation(loc, icon, turns, done)
	}
	return nil
}

func (u *UI) GetAnimAttack(attacker, defender foundation.ActorForUI) foundation.Animation {
	return nil
}

func (u *UI) GetAnimDamage(defenderPos geometry.Point, damage int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateDamage {
		return nil
	}
	animation := NewDamageAnimation(defenderPos, u.game.GetPlayerPosition(), damage)
	animation.SetDoneCallback(done)
	return animation
}
func (u *UI) GetAnimTiles(positions []geometry.Point, frames []foundation.TextIcon, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}
	return NewTilesAnimation(positions, frames, done)
}

func (u *UI) GetAnimRadialReveal(position geometry.Point, dijkstra map[geometry.Point]int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}

	animation := NewRadialAnimation(position, dijkstra, u.currentTheme.GetColorByName, u.mapLookup, done)
	animation.SetKeepDrawingCoveredGround(true)
	animation.SetUseIconColors(false)
	return animation
}

func (u *UI) GetAnimRadialAlert(position geometry.Point, dijkstra map[geometry.Point]int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}
	lookup := func(loc geometry.Point) (foundation.TextIcon, bool) {
		return foundation.TextIcon{
			Rune: '‼',
			Fg:   u.currentTheme.GetColorByName("Black"),
			Bg:   u.currentTheme.GetColorByName("Red"),
		}, true
	}
	animation := NewRadialAnimation(position, dijkstra, u.currentTheme.GetColorByName, lookup, done)
	animation.SetUseIconColors(true)
	return animation
}

func (u *UI) GetAnimTeleport(user foundation.ActorForUI, origin, targetPos geometry.Point, appearOnMap func()) (foundation.Animation, foundation.Animation) {
	originalIcon := u.getIconForActor(user)
	mapBackground := u.getMapTileBackgroundColor(origin)
	lightCyan := u.currentTheme.GetColorByName("LightCyan")
	white := u.currentTheme.GetColorByName("White")
	lightGray := u.currentTheme.GetColorByName("LightGray")
	vanishAnim := u.GetAnimTiles([]geometry.Point{origin}, []foundation.TextIcon{
		originalIcon.WithFg(white),
		originalIcon.WithFg(white),
		originalIcon.WithFg(lightCyan),
		{Rune: '*', Fg: lightCyan, Bg: mapBackground},
		{Rune: '*', Fg: lightCyan, Bg: mapBackground},
		{Rune: '+', Fg: lightCyan, Bg: mapBackground},
		{Rune: '+', Fg: lightCyan, Bg: mapBackground},
		{Rune: '|', Fg: lightCyan, Bg: mapBackground},
		{Rune: '|', Fg: lightCyan, Bg: mapBackground},
		{Rune: '∙', Fg: lightCyan, Bg: mapBackground},
		{Rune: '.', Fg: lightCyan, Bg: mapBackground},
		{Rune: '.', Fg: lightGray, Bg: mapBackground},
		{Rune: '.', Fg: u.currentTheme.GetColorByName("DarkGray"), Bg: mapBackground},
	}, nil)
	vanishAnim.RequestMapUpdateOnFinish()

	appearAnim := u.GetAnimAppearance(user, targetPos, appearOnMap)
	vanishAnim.SetFollowUp([]foundation.Animation{appearAnim})
	return vanishAnim, appearAnim
}

func (u *UI) GetAnimAppearance(actor foundation.ActorForUI, targetPos geometry.Point, done func()) foundation.Animation {
	originalIcon := u.getIconForActor(actor)
	mapBackground := u.getMapTileBackgroundColor(targetPos)
	lightCyan := u.currentTheme.GetColorByName("LightCyan")
	white := u.currentTheme.GetColorByName("White")
	lightGray := u.currentTheme.GetColorByName("LightGray")
	appearAnim := u.GetAnimTiles([]geometry.Point{targetPos}, []foundation.TextIcon{
		{Rune: '.', Fg: u.currentTheme.GetColorByName("DarkGray"), Bg: mapBackground},
		{Rune: '.', Fg: lightGray, Bg: mapBackground},
		{Rune: '.', Fg: lightGray, Bg: mapBackground},
		{Rune: '.', Fg: lightCyan, Bg: mapBackground},
		{Rune: '∙', Fg: lightCyan, Bg: mapBackground},
		{Rune: '|', Fg: lightCyan, Bg: mapBackground},
		{Rune: '|', Fg: lightCyan, Bg: mapBackground},
		{Rune: '+', Fg: lightCyan, Bg: mapBackground},
		{Rune: '+', Fg: lightCyan, Bg: mapBackground},
		{Rune: '*', Fg: lightCyan, Bg: mapBackground},
		{Rune: '*', Fg: lightCyan, Bg: mapBackground},
		{Rune: originalIcon.Rune, Fg: white, Bg: mapBackground},
		{Rune: originalIcon.Rune, Fg: white, Bg: mapBackground},
		{Rune: originalIcon.Rune, Fg: lightCyan, Bg: mapBackground},
		{Rune: originalIcon.Rune, Fg: lightCyan, Bg: mapBackground},
		{Rune: originalIcon.Rune, Fg: white, Bg: mapBackground},
		{Rune: originalIcon.Rune, Fg: white, Bg: mapBackground},
	}, done)
	return appearAnim
}

func (u *UI) GetAnimEvade(defender foundation.ActorForUI, done func()) foundation.Animation {
	actorIcon := u.getIconForActor(defender)
	return u.GetAnimTiles([]geometry.Point{defender.Position()}, []foundation.TextIcon{
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
	yellow := u.currentTheme.GetColorByName("Yellow")
	var prevAnim foundation.Animation
	var rootAnim foundation.Animation
	runeCount := len(wakeUpRunes)
	for i := 0; i < runeCount; i++ {

		cycleIcon := foundation.TextIcon{
			Rune: wakeUpRunes[i],
			Fg:   u.currentTheme.GetUIColor(UIColorUIForeground),
			Bg:   u.currentTheme.GetColorByName("black"),
		}

		frames := []foundation.TextIcon{
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
	confuseColors := []color.RGBA{u.currentTheme.GetColorByName("LightMagenta"), u.currentTheme.GetColorByName("LightRed"), u.currentTheme.GetColorByName("Yellow"), u.currentTheme.GetColorByName("LightGreen"), u.currentTheme.GetColorByName("LightBlue")}
	randomColor := func() color.RGBA {
		return confuseColors[rand.Intn(len(confuseColors))]
	}
	cycleCount := 4

	var prevAnim foundation.Animation
	var rootAnim foundation.Animation
	for i := 0; i < cycleCount; i++ {

		cycleIcon := foundation.TextIcon{
			Rune: randomRune(),
			Fg:   u.currentTheme.GetUIColor(UIColorUIForeground),
			Bg:   u.currentTheme.GetUIColor(UIColorUIBackground),
		}

		frames := []foundation.TextIcon{
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
	projAnim := u.GetAnimTiles(path, []foundation.TextIcon{
		{Rune: '.', Fg: u.currentTheme.GetColorByName("White"), Bg: u.currentTheme.GetColorByName("Black")},
		{Rune: '∙', Fg: u.currentTheme.GetColorByName("White"), Bg: u.currentTheme.GetColorByName("Black")},
		{Rune: '*', Fg: u.currentTheme.GetColorByName("White"), Bg: u.currentTheme.GetColorByName("Black")},
		{Rune: '*', Fg: u.currentTheme.GetColorByName("Yellow"), Bg: u.currentTheme.GetColorByName("Black")},
		{Rune: '*', Fg: u.currentTheme.GetColorByName("Red"), Bg: u.currentTheme.GetColorByName("Black")},
		{Rune: '*', Fg: u.currentTheme.GetColorByName("LightGray"), Bg: u.currentTheme.GetColorByName("Black")},
		{Rune: '*', Fg: u.currentTheme.GetColorByName("DarkGray"), Bg: u.currentTheme.GetColorByName("Black")},
	}, done)
	return projAnim
}
func (u *UI) GetAnimVorpalizeWeapon(origin geometry.Point, done func()) []foundation.Animation {
	effectIcon := foundation.TextIcon{
		Rune: '+',
		Fg:   u.currentTheme.GetColorByName("White"),
		Bg:   u.currentTheme.GetColorByName("Black"),
	}
	outmostPositions := geometry.CircleAround(origin, 2)
	outerPositions := geometry.CircleAround(origin, 1)

	animationInner := u.GetAnimTiles([]geometry.Point{origin}, []foundation.TextIcon{
		effectIcon.WithBg(u.currentTheme.GetColorByName("White")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("White")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("LightGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")).WithFg(u.currentTheme.GetColorByName("DarkGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")).WithFg(u.currentTheme.GetColorByName("Black")),
	}, done)
	animationCenter := u.GetAnimTiles(outerPositions, []foundation.TextIcon{
		effectIcon.WithBg(u.currentTheme.GetColorByName("White")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("LightGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")).WithFg(u.currentTheme.GetColorByName("DarkGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")).WithFg(u.currentTheme.GetColorByName("Black")),
	}, nil)

	animationOuter := u.GetAnimTiles(outmostPositions, []foundation.TextIcon{
		effectIcon.WithBg(u.currentTheme.GetColorByName("LightGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")).WithFg(u.currentTheme.GetColorByName("DarkGray")),
		effectIcon.WithBg(u.currentTheme.GetColorByName("Black")).WithFg(u.currentTheme.GetColorByName("Black")),
	}, nil)

	return []foundation.Animation{animationInner, animationCenter, animationOuter}
}
func (u *UI) GetAnimEnchantWeapon(player foundation.ActorForUI, location geometry.Point, done func()) foundation.Animation {
	playerIcon := u.getIconForActor(player)
	frames := []foundation.TextIcon{
		playerIcon.WithBg(u.currentTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightBlue")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightBlue")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightBlue")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightBlue")).WithFg(u.currentTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightBlue")).WithFg(u.currentTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightCyan")).WithFg(u.currentTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightCyan")).WithFg(u.currentTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightCyan")).WithFg(u.currentTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightBlue")).WithFg(u.currentTheme.GetColorByName("DarkGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightBlue")).WithFg(u.currentTheme.GetColorByName("DarkGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightBlue")).WithFg(u.currentTheme.GetColorByName("DarkGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("Blue")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("Blue")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("Blue")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("Blue")).WithFg(u.currentTheme.GetColorByName("LightGray")),
	}
	return u.GetAnimTiles([]geometry.Point{location}, frames, done)
}
func (u *UI) GetAnimEnchantArmor(player foundation.ActorForUI, location geometry.Point, done func()) foundation.Animation {
	playerIcon := u.getIconForActor(player)
	frames := []foundation.TextIcon{
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightGray")).WithFg(u.currentTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightGray")).WithFg(u.currentTheme.GetColorByName("White")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightGray")).WithFg(u.currentTheme.GetColorByName("White")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightGray")).WithFg(u.currentTheme.GetColorByName("White")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("LightGray")).WithFg(u.currentTheme.GetColorByName("White")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")).WithFg(u.currentTheme.GetColorByName("LightGray")),
		playerIcon.WithBg(u.currentTheme.GetColorByName("DarkGray")).WithFg(u.currentTheme.GetColorByName("LightGray")),
	}

	return u.GetAnimTiles([]geometry.Point{location}, frames, done)
}
func (u *UI) GetAnimThrow(item foundation.ItemForUI, origin geometry.Point, target geometry.Point) (foundation.Animation, int) {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateProjectiles {
		return nil, 0
	}
	textIcon := u.getIconForItem(item.GetCategory())

	return u.GetAnimProjectileWithIcon(textIcon, origin, target, nil)
}

func (u *UI) GetAnimProjectile(icon rune, fgColor string, origin geometry.Point, target geometry.Point, done func()) (foundation.Animation, int) {
	textIcon := foundation.TextIcon{
		Rune: icon,
		Fg:   u.currentTheme.GetColorByName(fgColor),
		Bg:   u.getMapTileBackgroundColor(origin), // TODO: this won't work when there is a transition
	}
	return u.GetAnimProjectileWithIcon(textIcon, origin, target, done)
}
func (u *UI) GetAnimProjectileWithIcon(textIcon foundation.TextIcon, origin geometry.Point, target geometry.Point, done func()) (foundation.Animation, int) {
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

	var trailIcons []foundation.TextIcon

	for i, cName := range colorNames {
		if i == 0 {
			trailIcons = append(trailIcons, foundation.TextIcon{
				Rune: leadIcon,
				Fg:   u.currentTheme.GetColorByName(cName),
				Bg:   u.currentTheme.GetColorByName("Black"),
			})
		} else {
			trailIcons = append(trailIcons, foundation.TextIcon{
				Rune: ' ',
				Fg:   u.currentTheme.GetColorByName("Black"),
				Bg:   u.currentTheme.GetColorByName(cName),
			})
		}
	}

	animation := NewProjectileAnimation(pathOfFlight, trailIcons[0], u.mapLookup, done)
	animation.SetTrail(trailIcons[1:])
	return animation, len(pathOfFlight)
}

func (u *UI) updateUntilDone() bool {
	screen := u.application.GetScreen()
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
			time.Sleep(10 * time.Millisecond)
			waited += 10 * time.Millisecond
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

func (u *UI) FadeToBlack() {
	screen := u.application.GetScreen()
	var breakingKey *tcell.EventKey
outerLoop:
	for i := 0; i < 100; i++ {
		if !darkenScreen(screen) {
			break outerLoop
		}
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
			time.Sleep(10 * time.Millisecond)
			waited += 10 * time.Millisecond
		}
	}
	if breakingKey != nil {
		u.application.QueueEvent(breakingKey)
	}
}

func darkenScreen(screen tcell.Screen) bool {
	darkenAmount := int32(10)
	w, h := screen.Size()
	centerPos := geometry.Point{X: w / 2, Y: h / 2}
	maxDist := geometry.Distance(centerPos, geometry.Point{X: 0, Y: 0})
	workLeft := false
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dist := geometry.Distance(centerPos, geometry.Point{X: x, Y: y})
			percent := util.Clamp(0.2, 1.0, (float64(dist)/float64(maxDist))+0.5)
			workDone := darkenScreenLocation(screen, x, y, int32(float64(darkenAmount)*percent))
			if workDone {
				workLeft = true
			}
		}
	}
	return workLeft
}

func darkenScreenLocation(screen tcell.Screen, x int, y int, darkenAmount int32) bool {
	icon, _, style, _ := screen.GetContent(x, y)
	fg, bg, _ := style.Decompose()
	fR, fG, fB := fg.RGB()
	bR, bG, bB := bg.RGB()
	hadWorkLeft := fR > 0 || fG > 0 || fB > 0 || bR > 0 || bG > 0 || bB > 0
	newFG := tcell.NewRGBColor(max(0, fR-darkenAmount), max(0, fG-darkenAmount), max(0, fB-darkenAmount))
	newBG := tcell.NewRGBColor(max(0, bR-darkenAmount), max(0, bG-darkenAmount), max(0, bB-darkenAmount))
	screen.SetContent(x, y, icon, nil, style.Background(newBG).Foreground(newFG))
	return hadWorkLeft
}

func (u *UI) ShowTextFile(fileName string) {
	lines := util.ReadFileAsLines(fileName)
	u.OpenTextWindow(lines)
}
func (u *UI) OpenTextWindow(description []string) {
	u.openTextModal(description)
}

func (u *UI) ShowTextFileFullscreen(filename string, onClose func()) {
	lines := util.ReadFileAsLines(filename)
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

func (u *UI) openTextModal(description []string) *cview.TextView {
	textView := u.newTextModal(description)
	textView.SetMouseCapture(u.closeOnAnyClickInside)
	u.makeCenteredModal("textModal", textView, len(description), longestLineWithoutColorCodes(description))
	return textView
}

func (u *UI) closeOnAnyClickInside(action cview.MouseAction, event *tcell.EventMouse) (outAction cview.MouseAction, outEvent *tcell.EventMouse) {
	if action == cview.MouseLeftClick || action == cview.MouseRightClick {
		u.pages.SetCurrentPanel("main")
		return action, nil
	}
	return action, event
}

func (u *UI) newTextModal(description []string) *cview.TextView {
	textView := cview.NewTextView()
	textView.SetBorder(true)

	textView.SetTextColor(u.currentTheme.GetUIColorForTcell(UIColorUIForeground))
	textView.SetBackgroundColor(u.currentTheme.GetUIColorForTcell(UIColorUIBackground))

	textView.SetBorderColor(u.currentTheme.GetUIColorForTcell(UIColorBorderForeground))

	u.setColoredText(textView, strings.Join(description, "\n"))
	return textView
}

func (u *UI) setColoredText(view *cview.TextView, text string) {
	if u.isMonochrome {
		stripped := cview.StripTags([]byte(text), true, true)
		view.SetDynamicColors(false)
		view.SetBytes(stripped)
	} else {
		view.SetDynamicColors(true)
		view.SetText(text)
	}
}

func (u *UI) UpdateLogWindow() {
	logMessages := u.game.GetLog()
	/*
		_, _, _, windowHeight := u.messageLabel.GetInnerRect()

		if len(logMessages) > windowHeight { // get just the last lines
			logMessages = logMessages[len(logMessages)-windowHeight:]
		}

	*/
	var asColoredStrings []string
	for i, message := range logMessages {
		fadePercent := util.Clamp(0.2, 1.0, float64(i+1)/float64(len(logMessages)))
		asColoredStrings = append(asColoredStrings, u.ToColoredText(message, fadePercent))
	}

	u.setColoredText(u.messageLabel, strings.Join(asColoredStrings, "\n"))
}

func (u *UI) ToColoredText(h foundation.HiLiteString, intensity float64) string {
	if h.IsEmpty() {
		return ""
	}
	textColor := u.currentTheme.GetUIColor(UIColorUIForeground)
	hiLiteColor := u.currentTheme.GetUIColor(UIColorTextForegroundHighlighted)
	if intensity < 1.0 {
		textColor = util.SetBrightness(textColor, intensity)
		hiLiteColor = util.SetBrightness(hiLiteColor, intensity)
	}
	textColorCode := RGBAToFgColorCode(textColor)
	if h.FormatString == "" {
		return fmt.Sprintf("%s%s", textColorCode, h.Value[0])
	}
	hiLiteColorCode := RGBAToFgColorCode(hiLiteColor)
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
	u.application.SetAfterResizeFunc(u.onTerminalResized)
	//go u.joystickLoop()
	if err := u.application.Run(); err != nil {
		panic(err)
	}
}

func (u *UI) initCoreUI() {
	cview.TrueColorTags = true
	cview.ColorUnset = tcell.ColorBlack

	u.application = cview.NewApplication()
	u.application.SetUnknownEventCapture(u.handleUnknownEvent)
	u.application.SetAfterResizeFunc(u.onTerminalResized)
	u.application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
func (u *UI) InitDungeonUI() {
	if u.mainGrid != nil {
		return
	}
	disableMouseFocus := func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		u.application.SetFocus(u.mapWindow) // Don't switch input focus here by clicking
		return action, nil
	}

	u.setupCommandTable()
	u.loadKeyMap(u.settings.KeyMapFileFullPath())

	u.application.GetScreen().SetCursorStyle(tcell.CursorStyleSteadyBlock)

	u.application.EnableMouse(true)

	u.application.SetMouseCapture(u.handleMainMouse)

	u.mapWindow = cview.NewBox()
	u.mapWindow.SetDrawFunc(u.drawMap)
	u.mapWindow.SetInputCapture(u.handleMainInput)

	u.messageLabel = cview.NewTextView()
	u.messageLabel.SetMouseCapture(disableMouseFocus)

	u.statusBar = cview.NewTextView()
	u.statusBar.SetDynamicColors(true)
	u.statusBar.SetScrollable(false)
	u.statusBar.SetMouseCapture(disableMouseFocus)
	u.statusBar.SetScrollBarVisibility(cview.ScrollBarNever)

	u.rightPanel = cview.NewTextView()
	u.rightPanel.SetScrollable(false)
	u.rightPanel.SetDynamicColors(true)
	u.rightPanel.SetWrap(false)
	u.rightPanel.SetMouseCapture(disableMouseFocus)

	u.lowerRightPanel = cview.NewTextView()
	u.lowerRightPanel.SetScrollable(false)
	u.lowerRightPanel.SetDynamicColors(true)
	u.lowerRightPanel.SetMouseCapture(disableMouseFocus)
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

	u.setTheme(u.settings.ThemeFullPath())
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
	u.isMonochrome = u.currentTheme.IsMonochrome()

	fg := u.currentTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.currentTheme.GetUIColorForTcell(UIColorUIBackground)
	u.statusBar.SetTextColor(fg)
	u.statusBar.SetBackgroundColor(bg)

	u.messageLabel.SetTextColor(fg)
	u.messageLabel.SetBackgroundColor(bg)
	u.messageLabel.SetBorderColor(fg)
	u.messageLabel.SetScrollBarColor(fg)
	u.messageLabel.SetDynamicColors(!u.isMonochrome)

	u.rightPanel.SetTextColor(fg)
	u.rightPanel.SetBorderColor(fg)
	u.rightPanel.SetBackgroundColor(bg)
	u.rightPanel.SetDynamicColors(!u.isMonochrome)
	u.rightPanel.SetTextAlign(cview.AlignRight)

	u.lowerRightPanel.SetTextColor(fg)
	u.lowerRightPanel.SetBorderColor(fg)
	u.lowerRightPanel.SetBackgroundColor(bg)
	u.lowerRightPanel.SetDynamicColors(!u.isMonochrome)
	u.lowerRightPanel.SetTextAlign(cview.AlignLeft)

	u.mapOverlay.SetDefaultColors(tcellColorToRGBA(bg), tcellColorToRGBA(fg))
}
func (u *UI) setTheme(fileName string) {
	u.currentTheme = NewThemeFromFile(fileName)
	u.currentTheme.SetBorders(&cview.Borders)
	u.applyStylingToUI()
	u.UpdateInventory()
	u.UpdateVisibleEnemies()
	u.UpdateStats()
	u.UpdateLogWindow()
}

func (u *UI) drawMap(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	if !u.gameIsReady {
		return x, y, width, height
	}
	defaultMapStyle := u.currentTheme.GetMapDefaultStyle()
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
	var textIcon foundation.TextIcon
	var isPositionAnimated bool
	foundIcon := false

	if animIcon, exists := u.animator.animationState[mapPos]; isAnimationFrame && exists {
		textIcon = animIcon
		foundIcon = true
		isPositionAnimated = true
	} else if u.mapOverlay.IsSet(mapPos.X, mapPos.Y) {
		textIcon = u.mapOverlay.Get(mapPos.X, mapPos.Y)
		foundIcon = true
	} else {
		textIcon, foundIcon = u.mapLookup(mapPos)
	}

	var fg, bg color.RGBA

	if foundIcon {
		ch, fg, bg = textIcon.Rune, textIcon.Fg, textIcon.Bg
		style = style.Attributes(textIcon.Attributes)
	} else {
		ch = ' '
		fg = u.currentTheme.GetUIColor(UIColorUIForeground)
		bg = u.currentTheme.GetUIColor(UIColorUIBackground)
	}

	if !u.isMonochrome { // apply gamma
		style = style.Foreground(tcell.NewRGBColor(int32(applyGamma(fg.R, u.gamma)), int32(applyGamma(fg.G, u.gamma)), int32(applyGamma(fg.B, u.gamma))))
		style = style.Background(tcell.NewRGBColor(int32(applyGamma(bg.R, u.gamma)), int32(applyGamma(bg.G, u.gamma)), int32(applyGamma(bg.B, u.gamma))))
	}

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
	gammaCorrected := util.Clamp(0, 1, math.Pow(colorAsFloat, gamma))
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
	items := u.game.GetInventory()
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
			itemIcon := u.currentTheme.GetIconForItem(item.GetCategory()).WithFg(u.currentTheme.GetInventoryItemColor(item.GetCategory())).WithBg(u.currentTheme.GetUIColor(UIColorUIBackground))
			if isEquipped {
				itemIcon = itemIcon.Reversed()
			}
			iconString := IconAsString(itemIcon)
			return iconString
		}
	} else {
		getItemName = func(item foundation.ItemForUI, isEquipped bool) string {
			nameWithColorsAndShortcut := item.InventoryNameWithColorsAndShortcut(RGBAToFgColorCode(u.currentTheme.GetInventoryItemColor(item.GetCategory())))
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

func IconAsString(icon foundation.TextIcon) string {
	code := RGBAToColorCodes(icon.Fg, icon.Bg)
	return fmt.Sprintf("%s%s[-:-]", code, string(icon.Rune))
}

func (u *UI) UpdateVisibleEnemies() {
	visibleEnemies := u.game.GetVisibleEnemies()
	//longest := longestInventoryLineWithoutColorCodes(visibleEnemies)
	var asString []string
	for _, enemy := range visibleEnemies {
		icon := u.getIconForActor(enemy)
		iconColor := RGBAToFgColorCode(icon.Fg)
		iconString := fmt.Sprintf("%s%s[-]", iconColor, string(icon.Rune))
		hp, hpMax := enemy.GetHitPoints(), enemy.GetHitPointsMax()
		asPercent := float64(hp) / float64(hpMax)

		hallucinating := u.isPlayerHallucinating()
		if hallucinating {
			asPercent = rand.Float64()
		}
		barIcon := '*'
		if enemy.HasFlag(foundation.FlagSleep) {
			barIcon = 'z'
		} else if !enemy.HasFlag(foundation.FlagAwareOfPlayer) {
			barIcon = '?'
		}
		hpBarString := fmt.Sprintf("[%s]", u.RuneBarFromPercent(barIcon, asPercent, 5))
		name := enemy.Name()
		if hallucinating {
			name = u.game.GetRandomEnemyName()
		}
		enemyLine := fmt.Sprintf(" %s %s %s", iconString, hpBarString, name)
		asString = append(asString, enemyLine)
	}
	u.lowerRightPanel.SetText(strings.Join(asString, "\n"))
}

func (u *UI) FullColorBarFromPercent(currentVal, maxVal, width int) string {
	percent := float64(currentVal) / float64(maxVal)
	colorChangeIndex := int(math.Round(percent * float64(width)))
	white := u.currentTheme.GetColorByName("White")
	colorCode := RGBAToColorCodes(u.currentTheme.GetColorByName("Green"), white)
	if percent < 0.50 {
		colorCode = RGBAToColorCodes(u.currentTheme.GetColorByName("Red"), white)
	} else if percent < 0.75 {
		colorCode = RGBAToColorCodes(u.currentTheme.GetColorByName("Yellow"), u.currentTheme.GetColorByName("Black"))
	}
	darkGrayCode := RGBAToColorCodes(u.currentTheme.GetColorByName("DarkGray"), white)

	valString := fmt.Sprintf("%d/%d", currentVal, maxVal)
	xForCenter := (width - len(valString)) / 2
	prefix := strings.Repeat(" ", xForCenter)
	suffix := strings.Repeat(" ", width-len(valString)-xForCenter)
	barString := fmt.Sprintf("%s%s%s", prefix, valString, suffix)

	barString = colorCode + barString[:colorChangeIndex] + darkGrayCode + barString[colorChangeIndex:] + "[-:-]"
	return barString
}

func (u *UI) RuneBarWithColor(icon rune, fgColorName, bgColorName string, current, max int) string {
	colorCode := RGBAToColorCodes(u.currentTheme.GetColorByName(fgColorName), u.currentTheme.GetColorByName(bgColorName))
	darkGrayCode := RGBAToFgColorCode(u.currentTheme.GetColorByName("DarkGray"))
	return colorCode + strings.Repeat(string(icon), current) + "[-:-]" + darkGrayCode + strings.Repeat(" ", max-current) + "[-]"
}

func (u *UI) RuneBarFromPercent(icon rune, percent float64, width int) string {
	repeats := int(math.Round(percent * float64(width)))
	colorCode := RGBAToFgColorCode(u.currentTheme.GetColorByName("Green"))
	if percent < 0.50 {
		colorCode = RGBAToFgColorCode(u.currentTheme.GetColorByName("Red"))
	} else if percent < 0.75 {
		colorCode = RGBAToFgColorCode(u.currentTheme.GetColorByName("Yellow"))
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
		itemName = "| " + equippedItem.LongNameWithColors(RGBAToFgColorCode(u.currentTheme.GetInventoryItemColor(equippedItem.GetCategory()))) + " |"
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
		fatigueBarContent := u.RuneBarWithColor('!', "VeryLightBlue", "Blue", fatigueCurrent, fatigueMax)
		fpBarStr := fmt.Sprintf("FP [%s]", fatigueBarContent)

		longFlags := FlagStringLong(flags)

		width, _ := u.application.GetScreenSize()

		lineTwo := fmt.Sprintf("%s %s %s", hpBarStr, fpBarStr, longFlags)

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

func FlagStringLong(flags map[foundation.ActorFlag]int) string {
	flagOrder := foundation.AllFlagsExceptGoldOrdered()
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

func FlagStringShort(flags map[foundation.ActorFlag]int) string {
	flagOrder := foundation.AllFlagsExceptGoldOrdered()
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
	hiCode := RGBAToFgColorCode(u.currentTheme.GetColorByName("Yellow"))
	return fmt.Sprintf("%s%s[-]", hiCode, statStr)
}
func (u *UI) getSingleLineStatus(statusValues map[foundation.HudValue]int, flags map[foundation.ActorFlag]int, multiLine bool, equippedItem string) string {

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
		statusStr = util.RightPadCount(statusStr, width-statusWidth)
	}
	return statusStr
}

func (u *UI) openInventory(items []foundation.ItemForUI) *TextInventory {
	list := NewTextInventory()
	list.SetLineColor(u.currentTheme.GetInventoryItemColor)
	list.SetEquippedTest(u.game.IsEquipped)
	list.SetStyle(u.currentTheme.defaultStyle)

	list.SetItems(items)

	//u.setupListForUI("inventory", list)
	panelName := "inventory"

	list.SetCloseHandler(func() {
		u.pages.HidePanel(panelName)
	})
	u.pages.AddPanel(panelName, list, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(list)

	//u.makeTopRightModal(panelName, list, len(inventoryItems), longestItem)
	return list
}

func (u *UI) OpenInventoryForManagement(items []foundation.ItemForUI) {
	inv := u.openInventory(items)
	inv.SetTitle("Inventory")
	inv.SetDefaultSelection(u.game.EquipToggle)
	inv.SetShiftSelection(u.game.DropItem)
	inv.SetControlSelection(u.game.PlayerApplyItem)

	inv.SetCloseOnControlSelection(true)
	inv.SetCloseOnShiftSelection(true)
}
func (u *UI) OpenInventoryForSelection(itemStacks []foundation.ItemForUI, prompt string, onSelected func(item foundation.ItemForUI)) {
	inv := u.openInventory(itemStacks)
	inv.SetSelectionMode()
	inv.SetTitle(prompt)
	inv.SetDefaultSelection(onSelected)
	inv.SetCloseOnSelection(true)
}
func (u *UI) makeCenteredModal(panelName string, primitive cview.Primitive, contentHeight, contentWidth int) {
	u.makeModal(wrapPrimitiveForModalCentered, panelName, primitive, contentHeight, contentWidth)
}
func (u *UI) makeModal(wrapperFunc func(p cview.Primitive, contentHeight int, contentWidth int) cview.Primitive, panelName string, primitive cview.Primitive, contentHeight int, contentWidth int) {
	w, h := u.application.GetScreenSize()
	height := contentHeight + 2
	horizontalSpaceForBorder := 2
	if height > h-4 { // needs scrolling
		height = h - 4
		horizontalSpaceForBorder += 1
	}
	width := min(contentWidth+horizontalSpaceForBorder, w-4)
	modalContainer := wrapperFunc(primitive, width, height)

	if inputCapturer, ok := primitive.(InputCapturer); ok {
		inputCapturer.SetInputCapture(u.popOnEscape)
	}
	u.pages.AddPanel(panelName, modalContainer, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(primitive)
}
func (u *UI) makeSideBySideModal(panelName string, primitive, qPrimitive cview.Primitive, contentHeight int, contentWidth int) {
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
	u.pages.AddPanel(panelName, modalContainer, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(qPrimitive)
}

func (u *UI) makeTopAndBottomModal(panelName string, primitive, qPrimitive cview.Primitive) {
	modalContainer := wrapPrimitivesTopToBottom(primitive, qPrimitive)

	if inputCapturer, ok := primitive.(InputCapturer); ok {
		inputCapturer.SetInputCapture(u.popOnEscape)
	}

	if inputCapturer, ok := qPrimitive.(InputCapturer); ok {
		inputCapturer.SetInputCapture(u.popOnEscape)
	}
	u.pages.AddPanel(panelName, modalContainer, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(qPrimitive)
}
func (u *UI) StartLockpickGame(difficulty foundation.Difficulty, getLockpickCount func() int, removeLockpick func(), onCompletion func(result foundation.InteractionResult)) {

	panelName := "lockpickGame"
	lockpickGame := NewLockpickGame(rand.Int63(), difficulty, getLockpickCount, removeLockpick, func(result foundation.InteractionResult) {
		u.pages.RemovePanel(panelName)
		u.pages.SetCurrentPanel("main")
		onCompletion(result)
	})
	lockpickGame.SetAudioPlayer(u.audioPlayer)
	u.pages.AddPanel(panelName, lockpickGame, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(lockpickGame)
}
func (u *UI) StartHackingGame(identifier uint64, difficulty foundation.Difficulty, previousGuesses []string, onCompletion func(previousGuesses []string, success foundation.InteractionResult)) {
	passwordFile := "4-letter-words.txt"
	fakeCount := 4
	switch difficulty {
	case foundation.VeryEasy:
		passwordFile = "4-letter-words.txt"
		fakeCount = 4
	case foundation.Easy:
		passwordFile = "5-letter-words.txt"
		fakeCount = 4
	case foundation.Medium:
		passwordFile = "6-letter-words.txt"
		fakeCount = 5
	case foundation.Hard:
		passwordFile = "7-letter-words.txt"
		fakeCount = 5
	case foundation.VeryHard:
		passwordFile = "8-letter-words.txt"
		fakeCount = 6
	}
	passFile := path.Join(u.settings.DataRootDir, "wordlists", passwordFile)
	passwords := util.ReadFileAsLines(passFile)
	rnd := rand.New(rand.NewSource(int64(identifier)))

	// shuffle the passwords
	permutedIndexes := rnd.Perm(len(passwords))

	var fakes []string
	var correct string

	for i := 0; i < fakeCount; i++ {
		fakes = append(fakes, passwords[permutedIndexes[i]])
	}
	correct = passwords[permutedIndexes[fakeCount]]

	hackingGame := NewHackingGame(correct, fakes, func(previousGuesses []string, result foundation.InteractionResult) {
		u.pages.HidePanel("hackingGame")
		onCompletion(previousGuesses, result)
	})
	hackingGame.SetAudioPlayer(u.audioPlayer)
	hackingGame.SetGuesses(previousGuesses)
	panelName := "hackingGame"
	u.pages.AddPanel(panelName, hackingGame, true, true)
	u.pages.ShowPanel(panelName)
	u.application.SetFocus(hackingGame)
}

func (u *UI) SetConversationState(starterText string, starterOptions []foundation.MenuItem, isTerminal bool) {
	u.dialogueIsTerminal = isTerminal
	if !u.pages.HasPanel("conversation") && isTerminal {
		u.audioPlayer.PlayCue("ui/terminal_poweron")
	}
	// text field
	if u.dialogueText == nil {
		textField := cview.NewTextView()

		textField.SetBorder(true)
		textField.SetScrollable(true)
		textField.SetDynamicColors(true)
		textField.SetWordWrap(true)
		u.dialogueText = textField
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
			u.pages.SetCurrentPanel("main")
		}
		action.Action()
	})
	setListItemsFromMenuItems(u.dialogueOptions, starterOptions)
	u.makeTopAndBottomModal("conversation", u.dialogueText, u.dialogueOptions)
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
	list := cview.NewList()
	u.applyListStyle(list)

	list.SetSelectedFunc(func(index int, listItem *cview.ListItem) {
		action := menuItems[index]
		if action.CloseMenus {
			u.pages.SetCurrentPanel("main")
		}
		action.Action()
	})

	longestItem := setListItemsFromMenuItems(list, menuItems)
	u.makeCenteredModal("contextMenu", list, len(menuItems), longestItem)
	return list
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
		itemLength := len(action.Name) + 4
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
	monsterLore := util.ReadFileAsLines(lorePath)
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
		u.pages.SetCurrentPanel("main")
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
	if event == nil || u.gameIsOver {
		return nil, action
	}
	newX, newY := event.Position()
	if newX != u.currentMouseX || newY != u.currentMouseY {
		u.currentMouseX = newX
		u.currentMouseY = newY
	}

	if action == cview.MouseLeftDown || action == cview.MouseRightDown {
		panelName, prim := u.pages.GetFrontPanel()
		if panelName != "main" && panelName != "blocker" {
			x, y, w, h := prim.GetRect()
			if newX < x || newY < y || newX >= x+w || newY >= y+h || action == cview.MouseRightDown {
				u.pages.SetCurrentPanel("main")
				return nil, action
			}
		}
	}

	if action == cview.MouseLeftDown {
		mousePos := geometry.Point{X: newX, Y: newY}
		if u.currentMouseX >= u.settings.MapWidth {
			// clicked on right panel
			u.onRightPanelClicked(mousePos)
			return nil, action
		}

		mapPos := u.ScreenToMap(mousePos)
		mapInfo := u.game.GetMapInfo(mapPos)
		if !mapInfo.IsEmpty() {
			u.Print(mapInfo)
		}

		return nil, action
	} else if action == cview.MouseRightDown {
		mapPos := u.ScreenToMap(geometry.Point{X: newX, Y: newY})
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
	fg := u.currentTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.currentTheme.GetUIColorForTcell(UIColorUIBackground)

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

	list.SetBackgroundColor(bg)

	list.SetShortcutColor(fg)

	list.SetSelectedTextColor(fg)
	list.SetSelectedBackgroundColor(bg)
	list.SetSelectedTextAttributes(tcell.AttrReverse)
}

func (u *UI) ShowLog() {
	logLines := u.game.GetLog()
	logTexts := make([]string, len(logLines))
	for i, line := range logLines {
		logTexts[i] = u.ToColoredText(line, 1)
	}
	textView := u.openTextModal(logTexts)
	textView.ScrollToEnd()
}

type OverlayDrawInfo struct {
	Text       string
	Pos        geometry.Point
	Connectors []geometry.Point
	SourcePos  geometry.Point
}

func (u *UI) ShowEnemyOverlay() {
	listOfEnemies := u.game.GetVisibleEnemies()

	if len(listOfEnemies) == 0 {
		u.Print(foundation.Msg("No enemies in sight"))
		return
	}
	u.mapOverlay.ClearAll()

	for _, enemy := range listOfEnemies {
		name := enemy.Name()
		pos, connectors := u.calculateOverlayPos(enemy.Position(), len(enemy.Name()))
		if pos == enemy.Position() {
			continue
		}
		u.mapOverlay.Print(pos.X, pos.Y, name)
		u.mapOverlay.AsciiLine(enemy.Position(), pos, connectors)
	}

}

func (u *UI) ShowItemOverlay() {
	listOfItems := u.game.GetVisibleItems()

	if len(listOfItems) == 0 {
		u.Print(foundation.Msg("No items in sight"))
		return
	}
	u.mapOverlay.ClearAll()

	for _, items := range listOfItems {
		name := items.Name()
		pos, connectors := u.calculateOverlayPos(items.Position(), len(items.Name()))
		if pos == items.Position() {
			continue
		}
		u.mapOverlay.Print(pos.X, pos.Y, name)
		u.mapOverlay.AsciiLine(items.Position(), pos, connectors)
	}
}

func (u *UI) ShowVisibleEnemies() {
	listOfEnemies := u.game.GetVisibleEnemies()
	if len(listOfEnemies) == 0 {
		u.Print(foundation.Msg("No enemies in sight"))
		return

	}
	var infoTexts []string
	for _, enemy := range listOfEnemies {
		info := enemy.GetListInfo()
		info = fmt.Sprintf("%c - %s", u.getIconForActor(enemy).Rune, info)
		infoTexts = append(infoTexts, info)
	}
	u.OpenTextWindow(infoTexts)
}

func (u *UI) ShowVisibleItems() {
	listOfItems := u.game.GetVisibleItems()
	if len(listOfItems) == 0 {
		u.Print(foundation.Msg("No items in sight"))
		return

	}
	var infoTexts []string
	for _, item := range listOfItems {
		info := item.GetListInfo()
		info = fmt.Sprintf("%c - %s", u.getIconForItem(item.GetCategory()).Rune, info)
		infoTexts = append(infoTexts, info)
	}
	u.OpenTextWindow(infoTexts)
}

func (u *UI) calculateOverlayPos(position geometry.Point, widthNeeded int) (labelPos geometry.Point, connectors []geometry.Point) {
	sW, sH := u.GetMapWindowGridSize()
	locIsBlocked := func(pos geometry.Point) bool {
		return u.game.IsSomethingInterestingAtLoc(pos) || u.mapOverlay.IsSet(pos.X, pos.Y)
	}
	isPosForLabelValid := func(pos geometry.Point) bool {
		if pos.X < 0 || pos.Y < 0 || pos.X+widthNeeded >= sW || pos.Y >= sH {
			return false
		}
		for x := 0; x < widthNeeded; x++ {
			curPos := geometry.Point{X: pos.X + x, Y: pos.Y}
			if locIsBlocked(curPos) {
				return false
			}
		}
		return true
	}

	simpleRightConnector := position.Add(geometry.Point{X: 1, Y: 0})
	simpleRightLabelPos := position.Add(geometry.Point{X: 2, Y: 0})
	if isPosForLabelValid(simpleRightLabelPos) && !locIsBlocked(simpleRightConnector) {
		return simpleRightLabelPos, []geometry.Point{simpleRightConnector}
	}

	simpleLeftConnector := position.Add(geometry.Point{X: -1, Y: 0})
	simpleLeftLabelPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: 0})
	if isPosForLabelValid(simpleLeftLabelPos) && !locIsBlocked(simpleLeftConnector) {
		return simpleLeftLabelPos, []geometry.Point{simpleLeftConnector}
	}

	topRightConnector := position.Add(geometry.Point{X: 1, Y: -1})
	topRightLabelPos := position.Add(geometry.Point{X: 2, Y: -1})
	if isPosForLabelValid(topRightLabelPos) && !locIsBlocked(topRightConnector) {
		return topRightLabelPos, []geometry.Point{topRightConnector}
	}

	topLeftConnector := position.Add(geometry.Point{X: -1, Y: -1})
	topLeftLabelPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: -1})
	if isPosForLabelValid(topLeftLabelPos) && !locIsBlocked(topLeftConnector) {
		return topLeftLabelPos, []geometry.Point{topLeftConnector}
	}

	bottomRightConnector := position.Add(geometry.Point{X: 1, Y: 1})
	bottomRightLabelPos := position.Add(geometry.Point{X: 2, Y: 1})
	if isPosForLabelValid(bottomRightLabelPos) && !locIsBlocked(bottomRightConnector) {
		return bottomRightLabelPos, []geometry.Point{bottomRightConnector}
	}

	bottomLeftConnector := position.Add(geometry.Point{X: -1, Y: 1})
	bottomLeftLabelPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: 1})
	if isPosForLabelValid(bottomLeftLabelPos) && !locIsBlocked(bottomLeftConnector) {
		return bottomLeftLabelPos, []geometry.Point{bottomLeftConnector}
	}

	twoDownConnector := position.Add(geometry.Point{X: 0, Y: 2})
	twoDownLabelRightPos := position.Add(geometry.Point{X: 1, Y: 2})
	if isPosForLabelValid(twoDownLabelRightPos) && !locIsBlocked(twoDownConnector) {
		return twoDownLabelRightPos, []geometry.Point{twoDownConnector}
	}

	twoDownLabelLeftPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: 2})
	if isPosForLabelValid(twoDownLabelLeftPos) && !locIsBlocked(twoDownConnector) {
		return twoDownLabelLeftPos, []geometry.Point{twoDownConnector}
	}

	twoUpConnector := position.Add(geometry.Point{X: 0, Y: -2})
	twoUpLabelRightPos := position.Add(geometry.Point{X: 1, Y: -2})
	if isPosForLabelValid(twoUpLabelRightPos) && !locIsBlocked(twoUpConnector) {
		return twoUpLabelRightPos, []geometry.Point{twoUpConnector}
	}

	twoUpLabelLeftPos := position.Add(geometry.Point{X: -widthNeeded - 1, Y: -2})
	if isPosForLabelValid(twoUpLabelLeftPos) && !locIsBlocked(twoUpConnector) {
		return twoUpLabelLeftPos, []geometry.Point{twoUpConnector}
	}

	return position, nil
}

func (u *UI) onTerminalResized(width int, height int) {
	tty, _ := u.application.GetScreen().Tty()
	tty.Write([]byte{0x1B, 0x3E}) // set keypad to numeric mode
	tSizeX, tSizeY := u.settings.GetMinTerminalSize()
	u.application.QueueUpdateDraw(func() {
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
		u.UpdateLogWindow()
		u.UpdateInventory()
		u.UpdateStats()
	})

	if !u.gameIsReady {
		u.game.UIReady()
		u.gameIsReady = true
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
		isMonochrome:   false,
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

func (u *UI) mapLookup(loc geometry.Point) (foundation.TextIcon, bool) {
	if u.game.IsVisibleToPlayer(loc) {
		return u.visibleLookup(loc)
	} else if u.game.IsExplored(loc) && u.game.IsLit(loc) {
		icon := u.getIconForMap(u.game.MapAt(loc))
		return icon, true
	}
	return foundation.TextIcon{}, false
}
func (u *UI) getMapTileBackgroundColor(loc geometry.Point) color.RGBA {
	icon := u.getIconForMap(u.game.MapAt(loc))
	return icon.Bg
}
func (u *UI) visibleLookup(loc geometry.Point) (foundation.TextIcon, bool) {
	entityType := u.game.TopEntityAt(loc)
	switch entityType {
	case foundation.EntityTypeActor:
		actor := u.game.ActorAt(loc)
		return u.getIconForActor(actor), true
	case foundation.EntityTypeDownedActor:
		actor := u.game.DownedActorAt(loc)
		return u.getIconForActor(actor), true
	case foundation.EntityTypeItem:
		item := u.game.ItemAt(loc)
		return u.getIconForItem(item.GetCategory()), true
	case foundation.EntityTypeObject:
		object := u.game.ObjectAt(loc)
		return u.getIconForObject(object), true
	}
	icon := u.getIconForMap(u.game.MapAt(loc))
	return icon, true
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
			u.pages.SetCurrentPanel("main")
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
	u.makeSideBySideModal("textModal", textView, list, len(playerInfo), longestLineWithoutColorCodes(playerInfo))
}

func (u *UI) charSheetView() (*cview.TextView, []string) {
	playerInfo := u.game.GetCharacterSheet()
	textView := cview.NewTextView()
	textView.SetBorder(true)

	fg := u.currentTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.currentTheme.GetUIColorForTcell(UIColorUIBackground)

	textView.SetTextColor(fg)
	textView.SetBorderColor(fg)
	textView.SetBackgroundColor(bg)
	textView.SetBorderColorFocused(fg)
	textView.SetBorderColor(fg) // TODO: darker style here..
	u.setColoredText(textView, strings.Join(playerInfo, "\n"))
	return textView, playerInfo
}

func (u *UI) onRightPanelClicked(clickPos geometry.Point) {
	itemIndex := clickPos.Y - 1

	inv := u.game.GetInventory()

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

func (u *UI) getIconForItem(itemCategory foundation.ItemCategory) foundation.TextIcon {
	if u.isPlayerHallucinating() {
		u.currentTheme.GetIconForItem(foundation.RandomItemCategory())
	}
	iconForItem := u.currentTheme.GetIconForItem(itemCategory)
	if iconForItem.HasEmptyBackground() {
		mapIconHere := u.currentTheme.GetIconForMap(foundation.TileFloor)
		return iconForItem.WithBg(mapIconHere.Bg)
	}
	return iconForItem
}

func (u *UI) getIconForMap(worldTileType foundation.TileType) foundation.TextIcon {
	return u.currentTheme.GetIconForMap(worldTileType)
}

func (u *UI) getIconForObject(object foundation.ObjectForUI) foundation.TextIcon {
	if u.isPlayerHallucinating() {
		u.currentTheme.GetIconForObject(foundation.RandomObjectCategory())
	}
	objectIcon := u.currentTheme.GetIconForObject(object.GetCategory())
	if objectIcon.HasEmptyBackground() {
		mapIconHere := u.currentTheme.GetIconForMap(u.game.MapAt(object.Position()))
		return objectIcon.WithBg(mapIconHere.Bg)
	}
	return objectIcon
}

func RightPadColored(s string, pLen int) string {
	return s + strings.Repeat(" ", pLen-cview.TaggedStringWidth(s))
}

func (u *UI) GetAnimExplosion(hitPositions []geometry.Point, done func()) foundation.Animation {
	white := u.currentTheme.GetColorByName("White")
	background := u.currentTheme.GetIconForMap(foundation.TileFloor).Bg
	yellow := u.currentTheme.GetColorByName("Yellow")
	red := u.currentTheme.GetColorByName("Red")
	lightGray := u.currentTheme.GetColorByName("LightGray")
	darkGray := u.currentTheme.GetColorByName("DarkGray")
	frames := []foundation.TextIcon{
		{Rune: '.', Fg: white, Bg: background},
		{Rune: '∙', Fg: white, Bg: background},
		{Rune: '*', Fg: white, Bg: background},
		{Rune: '*', Fg: yellow, Bg: background},
		{Rune: '*', Fg: red, Bg: background},
		{Rune: '*', Fg: lightGray, Bg: background},
		{Rune: '*', Fg: darkGray, Bg: background},
	}
	return u.GetAnimTiles(hitPositions, frames, done)
}

func (u *UI) GetAnimUncloakAtPosition(actor foundation.ActorForUI, uncloakLocation geometry.Point) (foundation.Animation, int) {
	actorIcon := u.getIconForActor(actor)
	tileIcon := u.currentTheme.GetIconForMap(u.game.MapAt(uncloakLocation))
	lightGray := u.currentTheme.GetColorByName("LightGray")
	darkGray := u.currentTheme.GetColorByName("DarkGray")
	black := u.currentTheme.GetColorByName("Black")
	frames := []foundation.TextIcon{
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

func (u *UI) OpenThemesMenu() {
	themesDir := path.Join(u.settings.DataRootDir, "themes")
	allThemes := util.FilesInDirByExtension(themesDir, "rec")

	actions := make([]foundation.MenuItem, 0)
	for _, t := range allThemes {
		themeFile := t
		themeName := strings.TrimSuffix(path.Base(themeFile), ".rec")
		actions = append(actions, foundation.MenuItem{
			Name: themeName,
			Action: func() {
				u.setTheme(themeFile)
			},
		})
	}

	u.OpenMenu(actions)
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
			u.renderMapPosition(pos, false, u.currentTheme.GetMapDefaultStyle())
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
		u.audioPlayer.LoadEnemyCuesFromDir(path.Join(u.settings.DataRootDir, "audio", "enemies"))
		u.audioPlayer.SoundsLoaded()
	}()

	u.animator.SetAudioCuePlayer(u.audioPlayer)
}

type EventJoy struct {
	AxisData   []int8
	Buttons    uint32
	OccurredAt time.Time
}

func (e EventJoy) IsButtonReleased(index int) bool {
	return e.Buttons&(1<<uint(index)) == 0
}
func (e EventJoy) IsButtonDown(index int) bool {
	return e.Buttons&(1<<uint(index)) > 0
}

func (e EventJoy) When() time.Time {
	return e.OccurredAt
}

func NewJoyStickEvent(data []int8, buttons uint32) tcell.Event {
	return &EventJoy{
		AxisData:   data,
		Buttons:    buttons,
		OccurredAt: time.Now(),
	}
}
