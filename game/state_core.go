package game

import (
	"RogueUI/dungen"
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"RogueUI/special"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"path"
)

type GameState struct {
	Player *Actor

	TurnsTaken int

	ui foundation.GameUI

	textIcons []textiles.TextIcon

	logBuffer []foundation.HiLiteString

	gridMap       *gridmap.GridMap[*Actor, *Item, Object]
	dungeonLayout *dungen.DungeonMap

	tileStyle         int
	defaultBackground color.RGBA
	defaultForeground color.RGBA

	playerDijkstraMap     map[geometry.Point]int
	showEverything        bool
	playerName            string
	dataDefinitions       DataDefinitions
	afterAnimationActions []func()

	playerFoV               *geometry.FOV
	visionRange             int
	genericWallIconIndex    textiles.TextIcon
	playerIcon              textiles.TextIcon
	playerColor             string
	config                  *foundation.Configuration
	ascensionsWithoutAmulet int
	mapLoader               MapLoader
	gameFlags               fxtools.StringFlags
	terminalGuesses         map[string][]string
	iconsForItems           map[foundation.ItemCategory]textiles.TextIcon
	iconsForObjects         map[string]textiles.TextIcon
	spawnMap                string
	spawnLocation           string
	palette                 textiles.ColorPalette
	inventoryColors         map[foundation.ItemCategory]color.RGBA
}

func (g *GameState) PlayerInteractInDirection(direction geometry.CompassDirection) {
	positionOnMap := g.GetPlayerPosition().Add(direction.ToPoint())
	g.OpenContextMenuFor(positionOnMap)
}

func (g *GameState) IsInteractionAt(position geometry.Point) bool {
	return g.gridMap.IsTransitionAt(position)
}

func NewGameState(ui foundation.GameUI, config *foundation.Configuration) *GameState {
	paletteFile := path.Join(config.DataRootDir, "definitions", "palette.rec")
	palette := textiles.ReadPaletteFileOrDefault(fxtools.MustOpen(paletteFile))
	g := &GameState{
		config:      config,
		playerName:  config.PlayerName,
		playerColor: "White",
		playerIcon: textiles.TextIcon{
			Char: '@',
			Fg:   color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		ui:              ui,
		tileStyle:       0,
		gameFlags:       make(fxtools.StringFlags),
		dataDefinitions: GetDataDefinitions(config.DataRootDir, palette),
		playerFoV:       geometry.NewFOV(geometry.NewRect(0, 0, config.MapWidth, config.MapHeight)),
		visionRange:     14,
		palette:         palette,
		terminalGuesses: make(map[string][]string),
	}

	g.init()
	ui.SetGame(g)

	return g
}

func (g *GameState) init() {
	g.mapLoader = NewRecMapLoader(g, g.palette)

	g.iconsForItems, g.inventoryColors = loadIconsForItems(path.Join(g.config.DataRootDir, "definitions"), g.palette)

	g.TurnsTaken = 0
	g.logBuffer = []foundation.HiLiteString{}
	g.showEverything = false
}

func (g *GameState) initPlayerAndSpawnMap() {
	playerSheet := special.NewCharSheet()
	playerSheet.SetSkillAdjustment(special.SmallGuns, 50)
	g.Player = NewPlayer(g.playerName, g.playerIcon, playerSheet)

	playerStartInfo := path.Join(g.config.DataRootDir, "definitions", "player_start.rec")
	if fxtools.FileExists(playerStartInfo) {
		startGear := recfile.Read(fxtools.MustOpen(playerStartInfo))[0]
		for _, field := range startGear {
			if field.Name == "mapName" {
				g.spawnMap = field.Value
			} else if field.Name == "mapLocation" {
				g.spawnLocation = field.Value
			} else if field.Name == "item" {
				parts := field.AsList("|")
				amount := parts[0].AsInt()
				itemName := parts[1].Value
				for i := 0; i < amount; i++ {
					// these require icons to be loaded
					g.giveAndTryEquipItem(g.Player, g.NewItemFromName(itemName))
				}
			}
		}
	}

	loadedMap := g.mapLoader.LoadMap(g.spawnMap)

	if loadedMap == nil {
		g.msg(foundation.Msg("It's impossible to move there.."))
		return
	}
	namedLocation := loadedMap.GetNamedLocation(g.spawnLocation)
	loadedMap.AddActor(g.Player, namedLocation)

	g.gridMap = loadedMap

	equipment := g.Player.GetEquipment()
	g.Player.GetFlags().SetOnChangeHandler(func(flag foundation.ActorFlag, value int) {
		g.ui.UpdateStats()
	})
	g.Player.charSheet.SetOnStatChangeHandler(func(stat special.Stat) {
		g.ui.UpdateStats()
	})

	g.Player.GetInventory().SetOnChangeHandler(g.ui.UpdateInventory)

	g.Player.GetInventory().SetOnBeforeRemove(equipment.UnEquip)

	equipment.SetOnChangeHandler(g.updateUIStatus)
}

// UIReady is called by the UI when it has initialized itself
func (g *GameState) UIReady() {
	g.moveIntoDungeon()
	// ADD Banner
	//g.ui.ShowTextFileFullscreen(path.Join("data","banner.txt"), g.moveIntoDungeon)
}

// moveIntoDungeon requires the UI to be available. It will request a dungeon crawl UI
// and then moves the player into the loaded map.
func (g *GameState) moveIntoDungeon() {

	// Since the player has equipment, we need the item
	g.initPlayerAndSpawnMap()

	g.afterPlayerMoved()

	g.ui.InitDungeonUI(g.palette, g.inventoryColors)

	g.updateUIStatus()
}

func (g *GameState) Reset() {
	g.init()
	g.moveIntoDungeon()
	g.ui.UpdateInventory()
}

func (g *GameState) QueueActionAfterAnimation(action func()) {
	g.afterAnimationActions = append(g.afterAnimationActions, action)
}

// endPlayerTurn is called by game actions that end the player's turn.
// It will
// - animate the player's actions
// - then the enemies' actions
// - remove dead actors and apply regeneration
// - execute any actions that were queued to be executed after animations
// - update the UI status
// - check if the player can act
func (g *GameState) endPlayerTurn(playerTimeTakenForTurn int) {
	// player has changed the game state..
	g.TurnsTaken++

	didCancel := g.ui.AnimatePending() // animate player actions..

	g.enemyMovement(playerTimeTakenForTurn)

	if didCancel {
		g.ui.SkipAnimations()
	} else {
		g.ui.AnimatePending() // animate enemy actions
	}

	g.removeDeadAndApplyRegeneration()

	for _, action := range g.afterAnimationActions {
		action()
	}
	g.afterAnimationActions = nil

	g.updateUIStatus()

	g.checkPlayerCanAct()
}

func (g *GameState) calculateTotalNetWorth() int {
	return g.Player.GetGold()
}

func (g *GameState) gameWon() {
	scoreInfo := foundation.ScoreInfo{
		PlayerName:         g.Player.Name(),
		Gold:               g.calculateTotalNetWorth(),
		DescriptiveMessage: "ESCAPED the dungeon",
		Escaped:            true,
	}
	highScores := g.writePlayerScore(scoreInfo)
	g.ui.ShowGameOver(scoreInfo, highScores)
}

func (g *GameState) gameOver(death string) {
	scoreInfo := foundation.ScoreInfo{
		PlayerName:         g.Player.Name(),
		Gold:               g.calculateTotalNetWorth(),
		DescriptiveMessage: death,
		Escaped:            false,
	}
	highScores := g.writePlayerScore(scoreInfo)
	g.ui.ShowGameOver(scoreInfo, highScores)
}
