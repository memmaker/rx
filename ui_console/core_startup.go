package ui_console

import (
	"contractor/foundation"
	"contractor/game"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/audio"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"os"
	"path/filepath"
)

func NewTextUI(state *game.GameState, uiImpl UILifeCycler, settings *foundation.Configuration) *UI {
	u := &UI{
		targetingTiles: make(map[geometry.Point]rune),
		animator:       NewAnimator(),
		audioPlayer:    audio.NewPlayer(),
		listTable:      make(map[string]*cview.List),
		cursorStyle:    tcell.CursorStyleSteadyBlock,
		gamma:          1.0,
		game:           state,
		lifeCycle:      uiImpl,
		settings:       settings,
		keyTable:       make(map[KeyLayer]map[UIKey]string),
		lastFrameIcons: make(map[geometry.Point]rune),
		lastFrameStyle: make(map[geometry.Point]tcell.Style),
	}
	u.loadAudioSfxInBackground()

	cview.TrueColorTags = true
	cview.ColorUnset = tcell.ColorBlack

	u.application = cview.NewApplication()
	//u.application.SetUnknownEventCapture(u.handleUnknownEvent)
	u.application.SetAfterResizeFunc(u.onTerminalResized)
	u.application.SetInputCapture(u.appInputCapture)

	u.pages = cview.NewPanels()

	u.application.SetRoot(u.pages, true)

	u.initUI(state.Palette(), state.InventoryColors())

	return u
}

func (u *UI) loadAudioSfxInBackground() {
	if !u.settings.SoundEffectsEnabled && u.settings.AudioEnabled {
		return
	}
	if !u.settings.AudioEnabled {
		u.audioPlayer.Disable()
		return
	}
	go func() {
		u.audioPlayer.LoadCuesFromDir(filepath.Join(u.settings.DataRootDir, "audio", "weapons"), "")
		u.audioPlayer.LoadCuesFromDir(filepath.Join(u.settings.DataRootDir, "audio", "ui"), "")
		u.audioPlayer.LoadCuesFromDir(filepath.Join(u.settings.DataRootDir, "audio", "world"), "")
		enemySfxDir := filepath.Join(u.settings.DataRootDir, "audio", "critters")
		entries, _ := os.ReadDir(enemySfxDir)
		for _, entry := range entries {
			if entry.IsDir() {
				enemyName := entry.Name()
				u.audioPlayer.LoadCuesFromDir(filepath.Join(enemySfxDir, enemyName), "critters")
			}
		}
		u.audioPlayer.SoundsLoaded()
	}()

	u.animator.SetAudioCuePlayer(u)
}

func (u *UI) initUI(palette textiles.ColorPalette, invColors map[foundation.ItemCategory]color.RGBA) {
	if u.mainGrid != nil {
		return
	}

	u.setupCommandTable()
	u.loadKeyMap(filepath.Join(u.settings.DataRootDir, "keymaps", u.settings.KeyMap+".txt"))

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
	u.rightPanel.SetWordWrap(false)

	u.lowerRightPanel = cview.NewTextView()
	u.lowerRightPanel.SetScrollable(true)
	u.lowerRightPanel.SetScrollBarVisibility(cview.ScrollBarNever)
	u.lowerRightPanel.SetDynamicColors(true)
	u.lowerRightPanel.SetWrap(false)
	u.lowerRightPanel.SetWordWrap(false)

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
}

func (u *UI) StartWithIntro() {
	u.pages.HidePanel("main")
	u.shouldPlayIntro = true
	u.StartGameLoop()
}
func (u *UI) StartGameLoop() {
	u.game.UIReady(u)
	u.lifeCycle.StartGameLoop(u.settings, u.application, u.afterScreenReady)
}

func (u *UI) afterScreenReady() {
	u.application.GetScreen().HideCursor()
	u.application.GetScreen().SetCursorStyle(tcell.CursorStyleSteadyBlock)

	if u.shouldPlayIntro && u.stateOfIntro == NoIntro {
		u.startIntro()
	}
}

func (u *UI) startIntro() {
	backGroundOverlay := u.addFullScreenTextOverlay()

	u.stateOfIntro = Fading
	FadeToWhite(u.application, u.settings.AnimationDelay, 5)

	u.stateOfIntro = TitleScreen
	backGroundOverlay.SetBackgroundColor(tcell.ColorWhite)
	backGroundOverlay.SetTextColor(tcell.ColorBlack)
	titleText := "C O N T R A C T O R\na game by Felix Ruzzoli"
	backGroundOverlay.SetText(titleText)

	u.application.Draw(backGroundOverlay)
}

func (u *UI) endIntro() {
	u.stateOfIntro = MainMenu
	u.showMainMenu()
}

func (u *UI) moveInGame() {
	u.stateOfIntro = InGame
	u.pages.RemovePanel("mainMenu")
	u.pages.RemovePanel("fullscreen")
	u.pages.ShowPanel("main")
	u.resetFocusToMain()
}

// helper functions

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
	return textView
}
