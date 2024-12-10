package game

import (
	"cmp"
	"contractor/d100"
	"contractor/foundation"
	"contractor/fsmai"
	"contractor/gridmap"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)

type TimedTransition struct {
	Destination gridmap.Transition
	Time        time.Time
}

type GameState struct {
	// Global State (Needs to be saved)
	gameTime             PointInTime
	gameFlags            *fxtools.StringFlags
	timeTracker          TimeTracker
	logBuffer            []foundation.HiLiteString
	journal              *Journal
	showEverything       bool
	flagsChangedThisTurn bool

	// Player State (Needs to be saved)
	Player            *Actor
	playerLightSource *gridmap.LightSource
	playerLastAimedAt d100.BodyPart

	// Map State
	mapLoader MapLoader

	// Scripts
	scriptRunner *ScriptRunner
	metronome    *Metronome

	// Maps (Needs to be saved)
	activeMaps     map[string]*gridmap.GridMap[*Actor, foundation.Item, Object]
	currentMapName string

	// outOfGame is a set of actors that have been removed from the game
	outOfGame map[*Actor]bool

	visionRange int

	// Input Config, UI, and bookkeeping
	config                *foundation.Configuration
	ui                    foundation.GameUI
	afterAnimationActions []func()

	// Colors & Icons
	palette         textiles.ColorPalette
	inventoryColors map[foundation.ItemCategory]color.RGBA
	iconsForItems   map[foundation.ItemCategory]textiles.TextIcon

	// Templates
	globalItemTemplates  map[string]recfile.Record
	globalActorTemplates map[string]recfile.Record
	globalTeamTemplates  map[string]recfile.Record

	// Temporary State
	chatterCache map[*Actor]map[foundation.ChatterType][]EntriesWithCondition
	quips        []string
	quipFile     string

	// Once a trap type has been successfully disarmed once, it can be disarmed again
	knownTraps        map[string]bool
	randomCodes       map[string][]rune
	mapContainsPlayer bool
	actorsComputed    int
}

func (g *GameState) GetMapSize() geometry.Point {
	return g.currentMap().MapSize()
}

func (g *GameState) ExecuteOnMap(mapName string, action func()) {
	currentMapName := g.currentMapName
	isForOtherMap := mapName != "" && mapName != currentMapName
	if isForOtherMap {
		g.mapContainsPlayer = false
		g.ensureMapIsLoaded(mapName)
		g.currentMapName = mapName
	}

	action()

	if isForOtherMap {
		g.mapContainsPlayer = true
		g.currentMapName = currentMapName
	}
}

func (g *GameState) MapContainsPlayer() bool {
	return g.mapContainsPlayer
}

func (g *GameState) Palette() textiles.ColorPalette {
	return g.palette
}

func (g *GameState) InventoryColors() map[foundation.ItemCategory]color.RGBA {
	return g.inventoryColors
}

func (g *GameState) TurnCount() int {
	return g.gameTime.Turns
}

func (g *GameState) GetPlayerFashionStyle() foundation.FashionStyle {
	return g.Player.OutfitStyle()
}

func (g *GameState) OpenPerkSelection(done func()) {
	var menuItems []foundation.MenuItem
	grayedOut := func(text string) string {
		return fmt.Sprintf("%s%s[-]", textiles.RGBAToFgColorCode(g.palette.Get("light_gray_5")), text)
	}
	for p := d100.Perk(0); p < d100.PerkCount; p++ {
		if !couldIncreasePerk(g.Player, p) {
			continue
		}
		requirements := d100.GetPerkRequirements(p)
		meetsRequirements := g.Player.GetCharSheet().MeetsRequirements(requirements)

		label := p.String()
		var action func()
		closeMenus := true
		if meetsRequirements {
			action = func() {
				g.ui.AskForConfirmation("Choose this?", p.Description(), func(didConfirm bool) {
					if didConfirm {
						g.Player.GetCharSheet().AddPerk(p)
						g.msg(foundation.HiLite("You have received the %s perk", p.String()))
						if p == d100.PerkOneLiners {
							g.chooseQuipStyle()
						}
						done()
					}
				})
			}
		} else {
			label = grayedOut(label)
			action = func() {
				g.ui.OpenTextWindow(fmt.Sprintf("%s Requires:\n%s", p.String(), requirements.String()))
			}
			closeMenus = false
		}

		menuItems = append(menuItems, foundation.MenuItem{
			Name:       label,
			Action:     action,
			CloseMenus: closeMenus,
		})
	}
	if len(menuItems) == 0 {
		g.msg(foundation.Msg("You have no perks to select"))
		return
	}

	boolAsInt := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}

	slices.SortStableFunc(menuItems, func(i, j foundation.MenuItem) int {
		if i.CloseMenus == j.CloseMenus {
			return strings.Compare(i.Name, j.Name)
		}
		return cmp.Compare(boolAsInt(j.CloseMenus), boolAsInt(i.CloseMenus))
	})

	g.ui.OpenMenu(menuItems)
}

func (g *GameState) chooseQuipStyle() {
	var quipChoices []foundation.MenuItem
	quipPath := path.Join(g.config.DataRootDir, "quips")
	entries, _ := os.ReadDir(quipPath)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".txt") {
			labelForStyle := strings.TrimSuffix(strings.ToLower(entry.Name()), ".txt")
			quipChoices = append(quipChoices, foundation.MenuItem{
				Name: labelForStyle,
				Action: func() {
					g.quipFile = entry.Name()
					g.msg(foundation.HiLite("You have selected the '%s' style", labelForStyle))
				},
				CloseMenus: true,
			})
		}
	}
	g.ui.OpenMenuWithTitle("Your style?", quipChoices)
}

func (g *GameState) DebugClickAt(pos geometry.Point) {
	var lines []string
	lines = append(lines, fmt.Sprintf("Clicked; %s", pos.String()))
	lines = append(lines, fmt.Sprintf("Light Brightness: %.2f", g.LightAt(pos).Brightness()))

	g.msg(foundation.Msg(strings.Join(lines, "\n")))
}

func (g *GameState) CanActorAttackNextTurn(enemy foundation.ActorForUI) bool {
	enemyActor := enemy.(*Actor)
	if !enemyActor.IsHostileTowards(g.Player) {
		return false
	}
	minTime := g.Player.minimalTimeNeededForAction()
	enemyTU := enemy.TimeEnergy() + minTime
	enemyNeeds := enemy.TimeNeededForAttack()

	return enemyTU >= enemyNeeds
}

func (g *GameState) PlayerQuip() {
	if !g.Player.HasPerk(d100.PerkOneLiners) {
		g.msg(foundation.Msg("You are missing the 'One-Liners' perk"))
		return
	}

	if !g.Player.HasActionPoints() {
		g.msg(foundation.Msg("You are too tired to say anything, need action points"))
		return
	}

	if len(g.quips) == 0 {
		quipFile := path.Join(g.config.DataRootDir, "quips", g.quipFile)
		playerQuips := fxtools.ReadFileAsLines(quipFile)
		g.quips = playerQuips
	}

	if len(g.quips) == 0 {
		g.tryAddChatter(g.Player, "I have nothing to say.")
		return
	}

	g.Player.GetCharSheet().LooseActionPoints(1)

	randomQuip := g.quips[rand.Intn(len(g.quips))]
	g.tryAddChatter(g.Player, randomQuip)

	g.Player.AddTemporaryStatChange(&TemporaryStatChange{
		StatChange: StatChange{
			SkillChanges: map[d100.Skill]int{
				d100.SkillForIntimidation: 20,
			},
		},
		Name:      "Cool Quip",
		TurnsLeft: 1,
	})

	for _, observer := range g.getObservers(g.Player.Position()) {
		if observer == g.Player {
			continue
		}
		observer.AddTemporaryStatChange(&TemporaryStatChange{
			StatChange: StatChange{
				StatChanges: map[d100.Stat]int{
					d100.Perception: -2,
				},
			},
			Name:      "Distracted by Cool Quip",
			TurnsLeft: 1,
		})
	}

	g.endPlayerTurn(5)
}
func (g *GameState) PlayerToggleRun() {
	if g.Player.HasFlag(foundation.FlagRunning) {
		g.Player.UnsetFlag(foundation.FlagRunning)
	} else if g.Player.GetCharSheet().GetActionPoints() > 0 {
		g.Player.SetFlag(foundation.FlagRunning)
		g.Player.UnsetFlag(foundation.FlagSneaking)
	}
	g.ui.UpdateVisibleActors()
}

func (g *GameState) PlayerToggleSneak() {
	if g.Player.HasFlag(foundation.FlagSneaking) {
		g.Player.UnsetFlag(foundation.FlagSneaking)
	} else if g.Player.GetCharSheet().GetActionPoints() > 0 {
		g.Player.UnsetFlag(foundation.FlagRunning)
		g.Player.SetFlag(foundation.FlagSneaking)
	}
	g.ui.UpdateVisibleActors()
}

func (g *GameState) SetIronMan() {
	g.gameFlags.SetFlag("IronMan")
}

func (g *GameState) IsIronMan() bool {
	return g.gameFlags.HasFlag("IronMan")
}

func (g *GameState) IsPlayerAndMapInitialized() bool {
	return g.Player != nil && g.currentMap() != nil
}

func (g *GameState) IsActorHostileTowardsPlayer(enemy foundation.ActorForUI) bool {
	actor := enemy.(*Actor)
	return actor.IsHostileTowards(g.Player)
}

func (g *GameState) IsActorAlliedWithPlayer(ally foundation.ActorForUI) bool {
	actor := ally.(*Actor)
	return actor.IsAlliedWith(g.Player)
}

func (g *GameState) PlayerInteractAtPosition(pos geometry.Point) {
	if g.OpenContextMenuFor(pos) {
		return
	}

	if g.currentMap().IsActorAt(pos) {
		actor := g.currentMap().ActorAt(pos)
		g.ui.OpenTextWindow(actor.GetDetailInfo())
		return
	}
	// Move..

	moveDistance := g.currentMap().MoveDistance(g.Player.Position(), pos)
	if moveDistance == 1 {
		direction := pos.Sub(g.Player.Position())
		g.ManualMovePlayer(direction.ToDirection())
		return
	}
	/*
		if g.currentMap().IsObjectAt(pos) {
			neighbors := g.currentMap().GetFilteredCardinalNeighbors(pos, func(neighbor geometry.Point) bool {
				return g.currentMap().IsCurrentlyPassable(neighbor)
			})
			if len(neighbors) == 0 {
				return
			}
			pos = nearestPosition(g.Player.Position(), neighbors)
		}
	*/
	if g.currentMap().IsCurrentlyPassable(pos) || g.currentMap().IsObjectAt(pos) {
		g.Player.RemoveGoal()

		pathTo := g.currentMap().GetJPSPath(g.Player.Position(), pos, func(point geometry.Point) bool {
			if point != pos {
				return g.currentMap().IsWalkableFor(point, g.Player)
			}
			return g.currentMap().IsWalkableFor(point, g.Player) || g.currentMap().IsObjectAt(point)
		})

		if len(pathTo) > 0 {
			g.Player.CurrentPath = pathTo
			g.Player.CurrentPathIndex = 0
			g.Player.CurrentPathBlockedCount = 0
			g.Player.SetGoal(GoalWalkToLocation(pos))
			g.RunPlayerPath()
		}
	}
	// assume movement to this position..
}

func nearestPosition(origin geometry.Point, other []geometry.Point) geometry.Point {
	nearest := other[0]
	nearestDistance := geometry.Distance(origin, nearest)
	for _, point := range other {
		distance := geometry.Distance(origin, point)
		if distance < nearestDistance {
			nearest = point
			nearestDistance = distance
		}
	}
	return nearest
}

func (g *GameState) IsPlayerOverEncumbered() bool {
	return g.Player.IsOverEncumbered()
}

func (g *GameState) GetPlayerName() string {
	return g.Player.Name()
}

func (g *GameState) PlayerGold() int {
	return g.Player.GetGold()
}

func (g *GameState) GetPlayerCharSheet() *d100.CharSheet {
	return g.Player.CharSheet
}

func (g *GameState) GetMapDisplayName() string {
	return g.currentMap().GetDisplayName()
}

func (g *GameState) currentMap() *gridmap.GridMap[*Actor, foundation.Item, Object] {
	return g.activeMaps[g.currentMapName]
}

func (g *GameState) WizardAdvanceTime() {
	g.advanceTime(time.Minute * 30)
	g.printTime()
}

func (g *GameState) printTime() {
	g.msg(foundation.Msg(fmt.Sprintf("Time is now %s", g.gameTime.Time.Format("Monday 15:04, 2006-01-02"))))
}

func (g *GameState) LightAt(p geometry.Point) fxtools.HDRColor {
	return g.currentMap().LightAt(p, g.Player.Position(), g.Player.CanSee(p), g.gameTime.Time)
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
	parts := []string{"weapons", "ammo", "armor", "food", "consumables", "miscItems"}
	for _, part := range parts {
		itemTemplateFile := path.Join(dataRootDir, "definitions", part+".rec")
		records, _ := recfile.ReadAndClose(fxtools.MustOpen(itemTemplateFile))
		for _, record := range records {
			itemTemplates[record.FindValueForKeyIgnoreCase("name")] = record
		}
	}
	return itemTemplates
}

func loadActorTemplates(dataRootDir string) map[string]recfile.Record {
	actorTemplates := make(map[string]recfile.Record)
	records, _ := recfile.ReadAndClose(fxtools.MustOpen(path.Join(dataRootDir, "definitions", "actors.rec")))
	for _, record := range records {
		actorTemplates[record.FindValueForKeyIgnoreCase("name")] = record
	}
	return actorTemplates
}

func loadSpawnedTeamsTemplates(dataRootDir string) map[string]recfile.Record {
	actorTemplates := make(map[string]recfile.Record)
	records, _ := recfile.ReadAndClose(fxtools.MustOpen(path.Join(dataRootDir, "definitions", "spawned_teams.rec")))
	for _, record := range records {
		actorTemplates[record.FindValueForKeyIgnoreCase("name")] = record
	}
	return actorTemplates
}

func NewGameState(config *foundation.Configuration) *GameState {
	loadD100Rules(path.Join(config.DataRootDir, "definitions"))

	paletteFile := path.Join(config.DataRootDir, "definitions", "palette.rec")
	palette := textiles.ReadPaletteFileOrDefault(fxtools.MustOpen(paletteFile))
	g := &GameState{
		config: config,
		playerLightSource: &gridmap.LightSource{
			Pos:          geometry.Point{},
			Radius:       5,
			Color:        fxtools.HDRColor{R: 1, G: 1, B: 1, A: 1},
			MaxIntensity: 1,
		},
		mapContainsPlayer:    true,
		timeTracker:          make(TimeTracker),
		visionRange:          80,
		palette:              palette,
		globalItemTemplates:  loadItemTemplates(config.DataRootDir),
		globalActorTemplates: loadActorTemplates(config.DataRootDir),
		globalTeamTemplates:  loadSpawnedTeamsTemplates(config.DataRootDir),
	}

	g.init()
	return g
}

func (g *GameState) GetPlayerNameAndIcon() (string, textiles.TextIcon) {
	return g.config.PlayerName, textiles.TextIcon{
		Char: g.config.PlayerChar,
		Fg:   g.palette.Get(g.config.PlayerColor),
	}
}

// NewActorFromName creates a new actor from a template name
// NOTE: These actors won't have a valid SpawnPosition set
func (g *GameState) NewActorFromName(actorName string) *Actor {
	actorRec, exists := g.globalActorTemplates[actorName]
	if !exists {
		return nil
	}
	actor, _ := g.NewActor(actorRec)
	return actor
}

// NewActor creates a new actor from a record
// It will also load the schedule for the actor, initialize the FSM and set and return the spawn position
func (g *GameState) NewActor(rec recfile.Record) (*Actor, geometry.Point) {
	newActor := NewActorFromRecord(rec, g.palette, g.NewItemFromString)
	if newActor != nil {
		g.loadSchedule(newActor)
		g.FSMInit(newActor)
		spawnPos := newActor.Position()
		newActor.SpawnPosition = spawnPos
		return newActor, spawnPos
	}
	panic(fmt.Sprintf("Could not create actor from record: %v", rec))
	return nil, geometry.Point{}
}

func (g *GameState) FSMInit(newActor *Actor) {
	defaultState := fsmai.StateNeutral
	if newActor.IsAggressive() {
		defaultState = fsmai.StateAggressive
	}
	newActor.FSM = NewActorFSM(g, newActor, defaultState, DefaultBehaviorFactory)
}
func (g *GameState) NewItem(rec recfile.Record) (foundation.Item, geometry.Point) {
	newItem := NewItemFromRecord(rec, g.NewItemFromString, g.iconForItem)
	if newItem != nil {
		itemPos := newItem.Position()
		return newItem, itemPos
	}
	panic(fmt.Sprintf("Could not create item from record: %v", rec))
	return nil, geometry.Point{}
}
func (g *GameState) NewObject(rec recfile.Record, newMap *gridmap.GridMap[*Actor, foundation.Item, Object], iconsForObjects func(objType string) textiles.TextIcon) (Object, geometry.Point) {
	object := g.NewObjectFromRecord(rec, newMap, iconsForObjects)
	if object != nil {
		objectPos := object.Position()
		return object, objectPos
	}
	panic(fmt.Sprintf("Could not create object from record: %v", rec))
	return nil, geometry.Point{}
}

func (g *GameState) init() {
	g.iconsForItems, g.inventoryColors = loadIconsForItems(path.Join(g.config.DataRootDir, "definitions"), g.palette)

	g.mapLoader = gridmap.NewRecMapLoader(
		path.Join(g.config.DataRootDir, "maps"),
		g.palette,
		g.NewActor,
		g.NewItem,
		g.NewObject,
	)
	g.activeMaps = make(map[string]*gridmap.GridMap[*Actor, foundation.Item, Object])
	g.chatterCache = make(map[*Actor]map[foundation.ChatterType][]EntriesWithCondition)
	g.outOfGame = make(map[*Actor]bool)
	g.knownTraps = make(map[string]bool)

	g.gameTime = PointInTime{
		Turns: 0,
		Time:  time.Date(2077, 2, 5, 16, 20, 23, 0, time.UTC),
	}

	g.logBuffer = []foundation.HiLiteString{}
	g.showEverything = false

	g.gameFlags = fxtools.NewStringFlags()

	g.journal = NewJournal(fxtools.MustOpen(path.Join(g.config.DataRootDir, "definitions", "journal.rec")), g.GetScriptFuncs())
	g.hookupJournalAndFlags()

	g.scriptRunner = NewScriptRunner()
	g.metronome = &Metronome{}
}

func (g *GameState) hookupJournalAndFlags() {
	g.journal.SetIncrementFlagHandler(g.gameFlags.Increment)
	g.journal.SetChangeHandler(func() {
		g.ui.PlayCue("ui/journal")
		g.msg(foundation.HiLite(">>> Journal updated <<<"))
	})
	g.gameFlags.SetChangeHandler(func(flagName string, value int) {
		g.flagsChangedThisTurn = true
	})
}

func (g *GameState) initPlayerAndMap() {
	playerSheet := d100.NewCharSheet()
	playerSheet.AddSkillPoints(0)

	//playerSheet.SetSkillAdjustment(special.SmallGuns, 50)
	playerName, playerIcon := g.GetPlayerNameAndIcon()
	g.Player = NewPlayer(playerName, playerIcon, playerSheet)
	g.FSMInit(g.Player)

	var spawnMap, spawnLocation string
	playerStartInfo := path.Join(g.config.DataRootDir, "definitions", "player_start.rec")
	if fxtools.FileExists(playerStartInfo) {
		records, _ := recfile.ReadAndClose(fxtools.MustOpen(playerStartInfo))
		startGear := records[0]
		for _, field := range startGear {
			if field.Name == "mapName" {
				spawnMap = field.Value
			} else if field.Name == "mapLocation" {
				spawnLocation = field.Value
			} else if field.Name == "item" {
				itemName := field.Value
				newItemFromString := g.NewItemFromString(itemName)
				g.giveAndTryEquipItem(g.Player, newItemFromString)
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
	g.Player.SetPosition(namedLocation)

	// TODO: ADD LIGHT SOURCE
	//loadedMap.AddDynamicLightSource(namedLocation, g.playerLightSource)

	g.setCurrentMap(loadedMap)

	for flagName, flagValue := range loadedMapResult.FlagsOfMap {
		g.gameFlags.Set(flagName, flagValue)
	}

	for _, script := range loadedMapResult.ScriptsToRun {
		g.RunScriptByName(script)
	}

	g.journal.Update()

	playerSheet.HealAPAndHPCompletely()

	g.playerSecondaryInit()

	g.afterMapLoad()

	g.SaveTimeNow("PlayerLastAteAt")
}

func (g *GameState) afterMapLoad() {
	g.currentMap().UpdateBakedLights()

	g.currentMap().UpdateDynamicLights()

	g.initAllActorSchedules()

	g.updateAllFoVsAndDijkstras()

	// Spawn Player
	g.currentMap().AddActorWithDisplacement(g.Player, g.Player.Position())

	g.ui.PlayMusic(path.Join(g.config.DataRootDir, "audio", "music", g.currentMap().GetMeta().MusicFile+".ogg"))

	g.advanceTime(time.Second * time.Duration(60))

	g.afterPlayerMoved(geometry.Point{}, true)

	g.updateUIStatus()
}

func (g *GameState) playerSecondaryInit() {
	g.Player.secondaryInit()

	g.Player.GetFlags().SetOnChangeHandler(func(flag foundation.ActorFlag, value int) {
		g.ui.UpdateStats()

		if flag == foundation.FlagSneaking {
			if g.Player.HasFlag(foundation.FlagSneaking) {
				g.ui.SetSneakOverlay(g.createSneakOverlay())
			} else {
				g.ui.SetSneakOverlay(nil)
			}
		}
	})
	g.Player.GetCharSheet().SetOnStatChangeHandler(func(stat d100.Stat) {
		g.ui.UpdateStats()
	})

	g.Player.GetInventory().SetOnChangeHandler(g.ui.UpdateInventory)

	equipment := g.Player.GetInventory()

	equipment.SetOnChangeHandler(g.updateUIStatus)

	g.Player.SetAugmentToggledHandler(func(augment CyberWare, enabled bool) {
		switch augment {
		case CyberWareLight:
			g.updatePlayerLightSource()
		}
	})
}

func (g *GameState) setCurrentMap(loadedMap *gridmap.GridMap[*Actor, foundation.Item, Object]) {
	mapName := loadedMap.GetName()
	g.activeMaps[mapName] = loadedMap
	g.currentMapName = mapName
}

// UIReady is called by the UI when it has initialized itself
func (g *GameState) UIReady(ui foundation.GameUI) {
	g.ui = ui

	g.moveIntoDungeon()
	// ADD Banner
	//g.ui.ShowTextFileFullscreen(path.Join("data","banner.txt"), g.moveIntoDungeon)
	g.scriptRunner.CheckAndRunFrames(g.currentMap().GetName())
}

// moveIntoDungeon requires the UI to be available. It will request a dungeon crawl UI
// and then moves the player into the loaded map.
func (g *GameState) moveIntoDungeon() {

	// Since the player has equipment, we need the item
	g.initPlayerAndMap()

	g.afterPlayerMoved(geometry.Point{}, true)

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
	g.actorsComputed = 0
	// advancing the time will run scripts
	g.advanceTimeAndTurn(time.Second * time.Duration(float64(playerTimeTakenForTurn)/10))

	// trigger turn based events
	g.metronome.Tick(g)

	g.applyPlayerHunger()

	didCancel := g.ui.AnimatePending() // animate player actions..

	// AI Actions (incl. Behaviours, Goals, Schedules) and State Changes happen here
	g.enemyMovement(playerTimeTakenForTurn)
	g.applyTurnCounters()
	if didCancel {
		g.ui.SkipAnimations()
	} else {
		g.ui.AnimatePending() // animate enemy actions
	}

	// EXPERIMENTAL and dangerous..
	// we simulate all actors on all loaded maps..
	if g.config.SimulateAllLoadedMaps {
		for mapName, _ := range g.activeMaps {
			if mapName == g.currentMapName {
				continue
			}
			g.ExecuteOnMap(mapName, func() {
				g.updateAllSchedules()
				g.scriptRunner.CheckAndRunFrames(mapName)
				g.enemyMovement(playerTimeTakenForTurn)
				g.applyTurnCounters()
				g.ui.SkipAnimations()
			})
		}
	}

	// This is where level transitions are handled
	for _, action := range g.afterAnimationActions {
		action()
	}
	g.afterAnimationActions = nil

	g.checkJournal()

	g.checkPlayerCanAct()

	if g.Player.HasFlag(foundation.FlagSneaking) {
		g.ui.SetSneakOverlay(g.createSneakOverlay())
	} else {
		g.ui.SetSneakOverlay(nil)
	}
	g.gameFlags.Set("ActorsComputed", g.actorsComputed)
	g.updateUIStatus()
}

func (g *GameState) checkJournal() {
	rewards := g.journal.Update()

	for _, reward := range rewards {
		g.awardXP(reward.XP, reward.Text)
	}
}

func (g *GameState) calculateTotalNetWorth() int {
	return g.Player.GetGold()
}

func (g *GameState) gameWon() {
	scoreInfo := foundation.ScoreInfo{
		PlayerName:         g.Player.Name(),
		Gold:               g.calculateTotalNetWorth(),
		DescriptiveMessage: "RETIRED - alive",
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

func (g *GameState) tryAddRandomChatter(actor *Actor, textType foundation.ChatterType) bool {
	chatter := g.GetRandomChatter(actor, textType)
	if chatter == "" {
		chatter = textType.DefaultChatter()
	}
	return g.tryAddChatter(actor, chatter)
}
func (g *GameState) tryAddChatter(actor *Actor, text string) bool {
	if text == "" || !g.MapContainsPlayer() {
		return false
	}
	if actor.IsAlive() && !actor.IsSleeping() && g.canPlayerSee(actor.Position()) {
		text = g.fillTemplatedText(text)
		if g.ui.TryAddChatter(actor, text) {
			g.msg(foundation.HiLite("%s: \"%s\"", actor.Name(), cview.Escape(text)))
			return true
		}
	}
	return false
}

func (g *GameState) StartPickpocket(actor *Actor) {
	actorEquipment := actor.GetInventory()
	stealableItems := actor.GetInventory().StackedItemsWithFilter(func(item foundation.Item) bool {
		return actorEquipment.IsNotEquipped(item)
	})

	rightToLeft := func(itemUI foundation.Item, amount int) {
		if amount == 0 {
			return
		}
		itemStack := itemUI
		if g.PlayerStealOrPlantItem(actor, itemStack, true) {
			g.StartPickpocket(actor)
		}
	}

	leftToRight := func(itemUI foundation.Item, amount int) {
		if amount == 0 {
			return
		}
		itemStack := itemUI
		if g.PlayerStealOrPlantItem(actor, itemStack, false) {
			g.StartPickpocket(actor)
		}
	}

	leftName := g.Player.Name()
	rightName := actor.Name()
	playerItems := g.Player.GetInventory().GetItems()

	stealAll := func() {
		loot := actor.GetInventory().StackedItemsWithFilter(func(item foundation.Item) bool {
			return actorEquipment.IsNotEquipped(item)
		})
		caught := false
		for _, lootItem := range loot {
			if !g.PlayerStealOrPlantItem(actor, lootItem, false) {
				caught = true
				break
			}
		}
		if !caught {
			g.StartPickpocket(actor)
		}
	}

	g.ui.ShowGiveAndTakeContainer(leftName, playerItems, rightName, stealableItems, rightToLeft, leftToRight, stealAll)
}

func (g *GameState) PlayerStealOrPlantItem(victim *Actor, item foundation.Item, isSteal bool) bool {
	itemStealModifier := 0
	if item.GetCategory().IsEasySteal() {
		itemStealModifier = 10
	} else if item.GetCategory().IsHardSteal() {
		itemStealModifier = -10
	}
	if victim.IsSleeping() {
		itemStealModifier += 75
	}

	var transferFunc func(foundation.Item)
	if isSteal {
		transferFunc = func(itemTaken foundation.Item) {
			victim.GetInventory().RemoveItem(item)
			g.Player.GetInventory().AddItem(item)
			g.msg(foundation.HiLite("You steal %s", item.Name()))
			g.ui.PlayCue("world/pickup")
		}
	} else {
		transferFunc = func(itemTaken foundation.Item) {
			g.Player.GetInventory().RemoveItem(item)
			victim.GetInventory().AddItem(item)
			g.msg(foundation.HiLite("You plant %s", item.Name()))
			g.ui.PlayCue("world/drop")
		}
	}

	skillRoll := g.Player.GetCharSheet().SkillRoll(d100.SkillForPickPockets, itemStealModifier)
	if skillRoll.Success {
		transferFunc(item)
		return true
	} else {
		g.msg(foundation.HiLite("%s notices your hands in his pockets", victim.Name()))
		if victim.IsSleeping() {
			victim.WakeUp()
		}
		g.trySetHostile(victim, g.Player)
		return false
	}
}

func (g *GameState) ShowDateTime() {
	// full date
	g.printTime()
}
func (g *GameState) getItemTemplateByName(shortString string) recfile.Record {
	if record, ok := g.globalItemTemplates[shortString]; ok {
		return record
	}
	return nil
}

func (g *GameState) IsInShootingRange(attacker *Actor, defender *Actor) bool {
	attackerPos := attacker.Position()
	defenderPos := defender.Position()
	moveDistance := geometry.Distance(attackerPos, defenderPos)

	attackerWeapon, hasWeapon := attacker.GetInventory().GetRangedWeapon()
	if !hasWeapon {
		return moveDistance <= 1
	}
	weaponRange := float64(attackerWeapon.GetCurrentAttackMode().MaxRange)

	inRange := moveDistance <= weaponRange

	visible := attacker.CanSee(defenderPos)

	return inRange && visible
}

func (g *GameState) IsInTalkingRange(one *Actor, two *Actor) bool {
	onePos := one.Position()
	twoPos := two.Position()
	moveDistance := geometry.Distance(onePos, twoPos)

	inRange := moveDistance <= 6

	visible := one.CanSee(twoPos)

	return inRange && visible
}

func (g *GameState) getShootingRangePosition(attacker *Actor, weaponRange int, victim *Actor) geometry.Point {
	attackerPos := attacker.Position()
	victimPos := victim.Position()

	mapAroundVictim := g.currentMap().GetDijkstraMap(victimPos, weaponRange, g.currentMap().IsCurrentlyPassable)
	mapAroundAttacker := g.currentMap().GetDijkstraMap(attackerPos, weaponRange+10, g.currentMap().IsCurrentlyPassable)

	// find the best position to shoot from
	bestPos := attackerPos
	bestDistance := geometry.Distance(attackerPos, victimPos)

	for pos, _ := range mapAroundVictim {
		if !victim.CanSee(pos) || pos == victimPos {
			continue
		}

		distForUs, canBeReached := mapAroundAttacker[pos]
		if !canBeReached {
			continue
		}
		distFloat := float64(distForUs) / 10

		if distFloat < bestDistance {
			bestPos = pos
			bestDistance = distFloat
		}
	}
	return bestPos
}
func (g *GameState) advanceTimeAndTurn(duration time.Duration) {
	g.gameTime = g.gameTime.AddDurationAndTurn(duration)
	g.scriptRunner.CheckAndRunFrames(g.currentMap().GetName())
	g.updateAllSchedules()
}
func (g *GameState) advanceTime(duration time.Duration) {
	g.gameTime = g.gameTime.AddDuration(duration)
	g.scriptRunner.CheckAndRunFrames(g.currentMap().GetName())
	g.updateAllSchedules()
	g.updateFoVAndDijkstraMap(g.Player)
}

func (g *GameState) awardXP(xp int, text string) {
	didLevelUpNow := g.Player.GetCharSheet().AddXP(xp)
	g.msg(foundation.HiLite("You received %s ("+text+")", strconv.Itoa(xp)+" XP"))
	if didLevelUpNow {
		g.ui.PlayCue("ui/LEVELUP")
		g.msg(foundation.HiLite(">>> You have gone up a level <<<"))
	}
}

func (g *GameState) actorHitMessage(victim *Actor, damage SourcedDamage) {
	if victim == g.Player {
		g.playerHitMessage(damage, damage.IsCrippling)
		return
	}
	baseMessage := fmt.Sprintf("%s was hit for %d hit points", victim.Name(), damage.DamageAmount)
	if damage.BodyPart != d100.Body {
		baseMessage += fmt.Sprintf("%s was hit in the %s for %d hit points", victim.Name(), damage.BodyPart.String(), damage.DamageAmount)
	}

	if damage.IsKillingBlow {
		baseMessage += fmt.Sprintf(", killing them")
	} else if damage.IsOverkill {
		baseMessage += fmt.Sprintf(", reducing them to a bloody pulp")
	} else if damage.IsCrippling {
		baseMessage += fmt.Sprintf(", crippling them")
	} else if damage.IsCritical {
		baseMessage += fmt.Sprintf(", dealing critical damage")
	}

	g.msg(foundation.Msg(baseMessage))
}

func (g *GameState) playerHitMessage(damage SourcedDamage, cripple bool) {
	baseMessage := fmt.Sprintf("You were hit for %d hit points", damage.DamageAmount)

	if cripple {
		baseMessage += fmt.Sprintf(", crippling your %s", damage.BodyPart.String())
	}

	g.msg(foundation.Msg(baseMessage))
}

func (g *GameState) TurnsTaken() int {
	return g.gameTime.Turns
}

func (g *GameState) GetRandomChatter(talker *Actor, chatterType foundation.ChatterType) string {
	if talker.ChatterFile == "" {
		return ""
	}
	var chatterForActor map[foundation.ChatterType][]EntriesWithCondition
	if _, hasCached := g.chatterCache[talker]; !hasCached {
		chatterFilePath := path.Join(g.config.DataRootDir, "dialogues", talker.ChatterFile+".rec")
		if !fxtools.FileExists(chatterFilePath) {
			return ""
		}
		records, _ := recfile.ReadAndClose(fxtools.MustOpen(chatterFilePath))
		g.chatterCache[talker] = NewChatterFromRecords(records, g.GetScriptFuncs())
	}

	chatterForActor = g.chatterCache[talker]

	if chatterList, hasChatter := chatterForActor[chatterType]; hasChatter {
		var chosenChatter EntriesWithCondition
		for _, chatter := range chatterList {
			if chatter.Condition == nil {
				chosenChatter = chatter
				break
			} else if condition, err := chatter.Condition.Evaluate(map[string]interface{}{
				"NPC": talker,
			}); err == nil && condition.(bool) {
				chosenChatter = chatter
				break
			}
		}
		return chosenChatter.Entries[rand.Intn(len(chosenChatter.Entries))]
	}
	return ""
}

type EntriesWithCondition struct {
	Condition *govaluate.EvaluableExpression
	Entries   []string
}

func NewChatterFromRecords(records []recfile.Record, condFuncs map[string]govaluate.ExpressionFunction) map[foundation.ChatterType][]EntriesWithCondition {
	chatter := make(map[foundation.ChatterType][]EntriesWithCondition)
	for _, record := range records {
		var chatterType foundation.ChatterType
		var entries []string
		var condition *govaluate.EvaluableExpression
		for _, field := range record {
			switch strings.ToLower(field.Name) {
			case "s":
				chatterType = foundation.NewChatterTypeFromString(field.Value)
			case "t":
				entries = append(entries, field.Value)
			case "c":
				condition, _ = govaluate.NewEvaluableExpressionWithFunctions(field.Value, condFuncs)
			}
		}
		chatter[chatterType] = append(chatter[chatterType], EntriesWithCondition{
			Condition: condition,
			Entries:   entries,
		})
	}
	return chatter
}

func (g *GameState) actorWithName(name string) *Actor {
	for _, actor := range g.currentMap().Actors() {
		if actor.GetInternalName() == name {
			return actor
		}
	}
	return nil
}

func (g *GameState) actorTransition(originMap *gridmap.GridMap[*Actor, foundation.Item, Object], actor *Actor, whereTo gridmap.Transition) {
	if whereTo.TargetMap == originMap.GetName() {
		// just teleport
		targetLocation := originMap.GetNamedLocation(whereTo.TargetLocation)
		originMap.MoveActor(actor, targetLocation)
		return
	}

	destMap := g.ensureMapIsLoaded(whereTo.TargetMap)
	if destMap == nil {
		return
	}

	originMap.RemoveActor(actor)

	targetLocation := destMap.GetNamedLocation(whereTo.TargetLocation)
	destMap.AddActorWithDisplacement(actor, targetLocation)

	g.ExecuteOnMap(whereTo.TargetMap, func() {
		g.trySetGoalFromSchedule(actor)
	})
}

func (g *GameState) removeItemFromGame(item foundation.Item) {
	item.SetAlive(false)
	itemPosition := item.Position()
	if itemAt, isItemAt := g.currentMap().TryGetItemAt(itemPosition); isItemAt && itemAt == item {
		g.currentMap().RemoveItem(item)
	} else if actorAt, isActorAt := g.currentMap().TryGetActorAt(itemPosition); isActorAt && actorAt.GetInventory().Has(item) {
		actorAt.GetInventory().RemoveItem(item)
		item.SetPosition(itemPosition)
	} else if downedActorAt, isDownedActorAt := g.currentMap().TryGetDownedActorAt(itemPosition); isDownedActorAt && downedActorAt.GetInventory().Has(item) {
		downedActorAt.GetInventory().RemoveItem(item)
	} else if objectAt, isObjectAt := g.currentMap().TryGetObjectAt(itemPosition); isObjectAt {
		if container, isContainer := objectAt.(*Container); isContainer && container.Has(item) {
			container.RemoveItem(item)
		}
	}
}

func (g *GameState) actorConsumeDrug(actor *Actor, item *GenericItem) {
	actor.AddTemporaryStatChange(&TemporaryStatChange{
		StatChange: item.StatChanges,
		Name:       item.Name(),
		TurnsLeft:  item.Charges,
	})

	if g.Player == actor {
		g.msg(foundation.HiLite("You consume %s", item.Name()))
	} else {
		g.msg(foundation.HiLite("%s consumes %s", actor.Name(), item.Name()))
	}
	if item.GetStackSize() > 1 {
		item.RemoveStacks(1)
		g.ui.UpdateInventory()
	} else {
		g.removeItemFromInventory(actor, item)
	}
}

func (g *GameState) NewAmmo(name string, bullets int) *Ammo {
	ammo := g.newItemFromName(name).(*Ammo)
	ammo.SetStackSize(bullets)
	return ammo
}

func (g *GameState) createSneakOverlay() map[geometry.Point]fxtools.HDRColor {
	canBeDetectedHere := func(pos geometry.Point) bool {
		for _, actor := range g.currentMap().Actors() {
			if actor != g.Player && actor.CanSee(pos) && actor.CanDetect(pos) {
				return true
			}
		}
		return false
	}
	var marked map[geometry.Point]fxtools.HDRColor
	for _, visPos := range g.Player.Visibles() {
		if !g.currentMap().IsTileWalkable(visPos) {
			continue
		}
		if canBeDetectedHere(visPos) {
			if marked == nil {
				marked = make(map[geometry.Point]fxtools.HDRColor)
			}
			// dark orange-red
			marked[visPos] = fxtools.HDRColor{R: 1.4, G: 0.3, B: 0.3, A: 1}
		}
	}
	return marked
}

func (g *GameState) updateAllFoVsAndDijkstras() {
	for _, actor := range g.currentMap().Actors() {
		g.updateFoVAndDijkstraMap(actor)
	}

	if g.Player.HasFlag(foundation.FlagSneaking) {
		g.ui.SetSneakOverlay(g.createSneakOverlay())
	}
}

func (g *GameState) initAllActorSchedules() {
	for _, actor := range g.currentMap().Actors() {
		g.trySetGoalFromSchedule(actor)
	}
}

type TimedSpawn struct {
	Actor     *Actor
	SpawnTime time.Time
	Location  geometry.Point
}

func (g *GameState) inventoryColorCode(item foundation.Item) string {
	return textiles.RGBAToFgColorCode(g.inventoryColors[item.GetCategory()])
}

func (g *GameState) SetNeutral(actor *Actor) {
	actor.SetNeutral()
	g.FSMInit(actor)
}

func (g *GameState) SetAggressive(actor *Actor) {
	actor.SetAggressive()
	g.FSMInit(actor)
}

func (g *GameState) isDetectedByObserver(actor *Actor, observer *Actor) bool {
	outsideDetectionRange := !observer.CanDetect(actor.Position())
	if outsideDetectionRange {
		return false
	}
	//SLEEPING?

	if observer.IsSleeping() {
		wakeMod := actor.GetCharSheet().GetSkill(d100.SkillForSneak)
		if !actor.HasFlag(foundation.FlagSneaking) {
			wakeMod -= 50
		}
		chanceToWake := (observer.GetCharSheet().GetStat(d100.Perception) * 10) - wakeMod
		awakened := d100.SuccessRoll(d100.Percentage(chanceToWake), 0)
		if awakened.Success {
			observer.GetFlags().Unset(foundation.FlagSleep)
			g.msg(foundation.HiLite("%s wakes up", observer.Name()))
		} else {
			return false
		}
	}

	if !actor.HasFlag(foundation.FlagSneaking) && !actor.HasFlag(foundation.FlagActiveCamouflage) {
		return true
	}

	// sneaking inside the detection range..

	actorPos := actor.Position()
	forcedDelta := 0
	lightAtActorPos := g.LightAt(actorPos).Brightness()
	if lightAtActorPos < 1.0 && !actor.HasFlag(foundation.FlagActiveCamouflage) {
		forcedDelta += int((1.0 - lightAtActorPos) * 50)
	}

	sneakSkill := actor.GetCharSheet().GetSkill(d100.SkillForSneak)
	perception := observer.GetCharSheet().GetStat(d100.Perception) * 10

	sneakSkill, perception = advantageForOne(sneakSkill, perception, forcedDelta)

	successCritChance := 0
	if actor.HasFlag(foundation.FlagActiveCamouflage) {
		successCritChance = 10
	}

	result := d100.SkillContest(d100.Percentage(sneakSkill), d100.Percentage(successCritChance), d100.Percentage(perception), 0)
	return result == 1
}

func couldIncreasePerk(player *Actor, perk d100.Perk) bool {
	currentLevel := player.GetPerkLevel(perk)
	if currentLevel >= perk.MaxLevel() {
		return false
	}
	return true
}

func (g *GameState) IsRemovedFromGame(leaver *Actor) bool {
	transition, exists := g.outOfGame[leaver]
	return exists && transition
}

func (g *GameState) getRandomNumberCode(door string) []rune {
	randomCode := g.randomCodes[door]
	if randomCode == nil {
		randomNumber := 1000 + rand.Intn(9000)
		randomCode = []rune(fmt.Sprintf("%d", randomNumber))
		if g.randomCodes == nil {
			g.randomCodes = make(map[string][]rune)
		}
		g.randomCodes[door] = randomCode
	}
	return randomCode
}

func (g *GameState) shouldActorBark(actor *Actor) bool {
	distanceToPlayer := geometry.DistanceChebyshev(actor.Position(), g.Player.Position())
	nearEachOther := distanceToPlayer <= 7
	return nearEachOther &&
		!g.Player.HasFlag(foundation.FlagSneaking) &&
		!g.Player.HasFlag(foundation.FlagActiveCamouflage) &&
		g.canPlayerSee(actor.Position()) &&
		actor.ChatterFile != "" &&
		actor.GetFlags().Get(foundation.FlagTurnsSinceLastIdleChatter) > 40 &&
		rand.Intn(4) == 0
}

func (g *GameState) setGameTime(start time.Time) {
	g.gameTime = g.gameTime.WithTime(start)
}

func (g *GameState) allActorsInOtherActiveMaps() map[*gridmap.GridMap[*Actor, foundation.Item, Object]][]*Actor {
	actors := make(map[*gridmap.GridMap[*Actor, foundation.Item, Object]][]*Actor)

	for _, activeMap := range g.activeMaps {
		if activeMap == g.currentMap() {
			continue
		}
		for _, actor := range activeMap.Actors() {
			actors[activeMap] = append(actors[activeMap], actor)
		}
	}

	return actors
}

func (g *GameState) isBedNear(position geometry.Point) (*Bed, bool) {
	neighborsWithObjects := g.currentMap().NeighborsAll(position, func(p geometry.Point) bool {
		return g.currentMap().IsObjectAt(p)
	})
	for _, neighbor := range neighborsWithObjects {
		if bed, isBed := g.currentMap().ObjectAt(neighbor).(*Bed); isBed {
			return bed, true
		}
	}
	return nil, false
}

func (g *GameState) getDialogueCheckMods(actor *Actor, opponent *Actor, skill d100.Skill, sitMods d100.ModList) d100.ModList {
	switch skill {
	case d100.SkillForIntimidation:
		return g.getIntimidateModifiers(actor, opponent, sitMods)
	}
	return sitMods
}

func (g *GameState) resolveActorID(id gridmap.ActorID) *Actor {
	return g.currentMap().GetActorByID(id)
}

func advantageForOne(one int, two int, advantage int) (int, int) {
	if advantage == 0 {
		return one, two
	}
	if two > advantage {
		two -= advantage
	} else if 100-one > advantage {
		one += advantage
	} else {
		one = 100
		two = 0
	}

	return one, two
}
