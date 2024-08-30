package game

import (
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"RogueUI/special"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"path"
	"time"
)

type GameState struct {
	// Global State (Needs to be saved)
	TurnsTaken      int
	gameTime        time.Time
	gameFlags       *fxtools.StringFlags
	logBuffer       []foundation.HiLiteString
	terminalGuesses map[string][]string
	journal         *Journal
	showEverything  bool

	// Player State (Needs to be saved)
	Player            *Actor
	playerFoV         *geometry.FOV
	playerLightSource *gridmap.LightSource
	playerDijkstraMap map[geometry.Point]int

	// Map State (Needs to be saved)
	mapLoader MapLoader

	activeMaps     map[string]*gridmap.GridMap[*Actor, *Item, Object]
	currentMapName string

	visionRange int

	// Input Config, UI, and bookkeeping
	config                *foundation.Configuration
	ui                    foundation.GameUI
	afterAnimationActions []func()

	// Colors & Icons
	palette         textiles.ColorPalette
	inventoryColors map[foundation.ItemCategory]color.RGBA
	iconsForItems   map[foundation.ItemCategory]textiles.TextIcon
	iconsForObjects map[string]textiles.TextIcon

	// Item Templates
	globalItemTemplates map[string]recfile.Record
	mapItemTemplates    map[string]recfile.Record
}

func (g *GameState) currentMap() *gridmap.GridMap[*Actor, *Item, Object] {
	return g.activeMaps[g.currentMapName]
}

func (g *GameState) WizardAdvanceTime() {
	g.gameTime = g.gameTime.Add(time.Minute * 30)
	g.msg(foundation.Msg(fmt.Sprintf("Time is now %s", g.gameTime.Format("15:04"))))
}

func (g *GameState) LightAt(p geometry.Point) fxtools.HDRColor {
	return g.currentMap().LightAt(p, g.gameTime)
}

func (g *GameState) PlayerInteractInDirection(direction geometry.CompassDirection) {
	positionOnMap := g.GetPlayerPosition().Add(direction.ToPoint())
	g.OpenContextMenuFor(positionOnMap)
}

func (g *GameState) IsInteractionAt(position geometry.Point) bool {
	return g.currentMap().IsTransitionAt(position)
}

func loadItemTemplates(dataRootDir string) map[string]recfile.Record {
	itemTemplates := make(map[string]recfile.Record)
	parts := []string{"weapons", "ammo", "armor", "food", "miscItems"}
	for _, part := range parts {
		itemTemplateFile := path.Join(dataRootDir, "definitions", part+".rec")
		records := recfile.Read(fxtools.MustOpen(itemTemplateFile))
		for _, record := range records {
			itemTemplates[record.FindValueForKeyIgnoreCase("name")] = record
		}
	}
	return itemTemplates
}

func NewGameState(ui foundation.GameUI, config *foundation.Configuration) *GameState {
	paletteFile := path.Join(config.DataRootDir, "definitions", "palette.rec")
	palette := textiles.ReadPaletteFileOrDefault(fxtools.MustOpen(paletteFile))
	g := &GameState{
		config:    config,
		ui:        ui,
		playerFoV: geometry.NewFOV(geometry.NewRect(0, 0, config.MapWidth, config.MapHeight)),
		playerLightSource: &gridmap.LightSource{
			Pos:          geometry.Point{},
			Radius:       5,
			Color:        fxtools.HDRColor{R: 1, G: 1, B: 1, A: 1},
			MaxIntensity: 1,
		},
		visionRange:         80,
		palette:             palette,
		globalItemTemplates: loadItemTemplates(config.DataRootDir),
		activeMaps:          make(map[string]*gridmap.GridMap[*Actor, *Item, Object]),
	}

	g.init()
	ui.SetGame(g)

	return g
}
func (g *GameState) GetPlayerNameAndIcon() (string, textiles.TextIcon) {
	return g.config.PlayerName, textiles.TextIcon{
		Char: g.config.PlayerChar,
		Fg:   g.palette.Get(g.config.PlayerColor),
	}
}
func (g *GameState) NewActor(rec recfile.Record) (*Actor, geometry.Point) {
	newActor := NewActorFromRecord(rec, g.palette, g.NewItemFromString)
	if newActor != nil {
		actorPos := newActor.Position()
		return newActor, actorPos
	}
	panic(fmt.Sprintf("Could not create actor from record: %v", rec))
	return nil, geometry.Point{}
}
func (g *GameState) NewItem(rec recfile.Record) (*Item, geometry.Point) {
	newItem := NewItemFromRecord(rec, g.iconForItem)
	if newItem != nil {
		itemPos := newItem.Position()
		return newItem, itemPos
	}
	panic(fmt.Sprintf("Could not create item from record: %v", rec))
	return nil, geometry.Point{}
}

func (g *GameState) NewObject(rec recfile.Record, iconsForObjects map[string]textiles.TextIcon, newMap *gridmap.GridMap[*Actor, *Item, Object]) (Object, geometry.Point) {
	object := g.NewObjectFromRecord(rec, g.palette, iconsForObjects, newMap)
	if object != nil {
		objectPos := object.Position()
		return object, objectPos
	}
	panic(fmt.Sprintf("Could not create object from record: %v", rec))
	return nil, geometry.Point{}
}

func (g *GameState) init() {
	g.mapLoader = gridmap.NewRecMapLoader(
		path.Join(g.config.DataRootDir, "maps"),
		g.palette,
		g.NewActor,
		g.NewItem,
		g.NewObject,
	)

	g.iconsForItems, g.inventoryColors = loadIconsForItems(path.Join(g.config.DataRootDir, "definitions"), g.palette)

	g.TurnsTaken = 0
	g.logBuffer = []foundation.HiLiteString{}
	g.showEverything = false

	g.gameFlags = fxtools.NewStringFlags()

	g.terminalGuesses = make(map[string][]string)

	g.journal = NewJournal(fxtools.MustOpen(path.Join(g.config.DataRootDir, "definitions", "journal.rec")), g.getConditionFuncs())
	g.journal.OnFlagsChanged()
	g.hookupJournalAndFlags()
}

func (g *GameState) hookupJournalAndFlags() {
	g.journal.SetChangeHandler(func() {
		g.ui.PlayCue("ui/journal")
		g.msg(foundation.HiLite(">>> %s <<<", "Journal updated"))
	})
	g.gameFlags.SetChangeHandler(func(flagName string, value int) {
		g.journal.OnFlagsChanged()
	})
}

func (g *GameState) initPlayerAndSpawnMap() {
	playerSheet := special.NewCharSheet()
	playerSheet.SetSkillAdjustment(special.SmallGuns, 50)
	playerName, playerIcon := g.GetPlayerNameAndIcon()
	g.Player = NewPlayer(playerName, playerIcon, playerSheet)
	var spawnMap, spawnLocation string
	playerStartInfo := path.Join(g.config.DataRootDir, "definitions", "player_start.rec")
	if fxtools.FileExists(playerStartInfo) {
		startGear := recfile.Read(fxtools.MustOpen(playerStartInfo))[0]
		for _, field := range startGear {
			if field.Name == "mapName" {
				spawnMap = field.Value
			} else if field.Name == "mapLocation" {
				spawnLocation = field.Value
			} else if field.Name == "item" {
				parts := field.AsList("|")
				amount := parts[0].AsInt()
				itemName := parts[1].Value
				for i := 0; i < amount; i++ {
					// these require icons to be loaded
					g.giveAndTryEquipItem(g.Player, g.NewItemFromString(itemName))
				}
			}
		}
	}

	loadedMapResult := g.mapLoader.LoadMap(spawnMap)

	loadedMap := loadedMapResult.Map
	if loadedMap == nil {
		g.msg(foundation.Msg("It's impossible to move there.."))
		return
	}
	namedLocation := loadedMap.GetNamedLocation(spawnLocation)
	loadedMap.AddActor(g.Player, namedLocation)
	// TODO: ADD LIGHT SOURCE
	//loadedMap.AddDynamicLightSource(namedLocation, g.playerLightSource)
	loadedMap.UpdateDynamicLights()

	g.setCurrentMap(loadedMap)

	g.iconsForObjects = loadedMapResult.IconsForObjects

	for flagName, flagValue := range loadedMapResult.FlagsOfMap {
		g.gameFlags.Set(flagName, flagValue)
	}

	g.attachHooksToPlayer()
}

func (g *GameState) attachHooksToPlayer() {
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

func (g *GameState) setCurrentMap(loadedMap *gridmap.GridMap[*Actor, *Item, Object]) {
	mapName := loadedMap.GetName()
	g.activeMaps[mapName] = loadedMap
	g.currentMapName = mapName
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

	g.afterPlayerMoved(geometry.Point{}, true)

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

	g.gameTime = g.gameTime.Add(time.Second * time.Duration(float64(playerTimeTakenForTurn)/10))

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

func (g *GameState) canActorSee(victim *Actor, position geometry.Point) bool {
	return g.currentMap().IsLineOfSightClear(victim.Position(), position)
}

func (g *GameState) tryAddChatter(actor *Actor, text string) bool {
	if g.canPlayerSee(actor.Position()) {
		if g.ui.TryAddChatter(actor, text) {
			g.msg(foundation.HiLite("%s: \"%s\"", actor.Name(), text))
			return true
		}
	}
	return false
}

func (g *GameState) StartPickpocket(actor *Actor) {
	actorEquipment := actor.GetEquipment()
	stealableItems := actor.GetInventory().StackedItemsWithFilter(actorEquipment.IsNotEquipped)

	var menuItems []foundation.MenuItem

	for _, i := range stealableItems {
		item := i
		itemStealModifier := 0
		if item.GetCategory().IsEasySteal() {
			itemStealModifier = 10
		} else if item.GetCategory().IsHardSteal() {
			itemStealModifier = -10
		}
		if actor.IsSleeping() {
			itemStealModifier += 75
		}
		chance := max(0, min(95, g.Player.GetCharSheet().GetSkill(special.Steal)+itemStealModifier))
		menuItems = append(menuItems, foundation.MenuItem{
			Name: fmt.Sprintf("Steal %s (%d%%)", item.Name(), chance),
			Action: func() {
				g.tryStealItem(actor, item.First(), itemStealModifier)
			},
			CloseMenus: true,
		})
	}

	g.ui.OpenMenu(menuItems)
}

func (g *GameState) tryStealItem(victim *Actor, item *Item, itemStealModifier int) {
	skillRoll := g.Player.GetCharSheet().SkillRoll(special.Steal, itemStealModifier)
	if skillRoll.Success {
		victim.GetInventory().Remove(item)
		g.Player.GetInventory().Add(item)
		g.msg(foundation.HiLite("You steal %s", item.Name()))
		g.ui.PlayCue("world/pickup")
	} else {
		g.msg(foundation.HiLite("%s notices you trying to steal %s", victim.Name(), item.Name()))
		if victim.IsSleeping() {
			victim.WakeUp()
		}
		victim.AddToEnemyActors(g.Player.GetInternalName())
		victim.SetHostile()
	}
}

func (g *GameState) getItemTemplateByName(shortString string) recfile.Record {
	if record, ok := g.mapItemTemplates[shortString]; ok {
		return record
	}
	if record, ok := g.globalItemTemplates[shortString]; ok {
		return record
	}
	return nil
}
