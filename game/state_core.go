package game

import (
	"cmp"
	"contractor/d100"
	"contractor/foundation"
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

type GameMap = *gridmap.GridMap[*Actor, foundation.Item, Object]

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
	Scripts   *ScriptRunner
	metronome *Metronome

	pathfinder *Pathfinder

	// Maps (Needs to be saved)
	activeMaps     map[string]GameMap
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
	chatterCache map[*Actor]map[foundation.ChatterTopic][]EntriesWithCondition
	quips        []string
	quipFile     string

	// Once a trap type has been successfully disarmed once, it can be disarmed again
	knownTraps        map[string]bool
	randomCodes       map[string][]rune
	mapContainsPlayer bool
	actorsComputed    int
	userFunctions     map[string]*govaluate.EvaluableExpression
	allMapNames       []string
}

func NewGameState(config *foundation.Configuration) *GameState {
	// stuff initialised here will stay the same between resets
	loadD100Rules(path.Join(config.DataRootDir, "definitions"))

	paletteFile := path.Join(config.DataRootDir, "definitions", "palette.rec")
	palette := textiles.ReadPaletteFileOrDefault(fxtools.MustOpen(paletteFile))

	g := &GameState{
		config:               config,
		visionRange:          80,
		palette:              palette,
		globalItemTemplates:  loadItemTemplates(config.DataRootDir),
		globalActorTemplates: loadActorTemplates(config.DataRootDir),
		globalTeamTemplates:  loadSpawnedTeamsTemplates(config.DataRootDir),
		allMapNames:          loadMapNames(config.DataRootDir),
	}
	g.userFunctions = g.loadUserFuncs(g.config.DataRootDir)
	g.mapLoader = gridmap.NewRecMapLoader(
		path.Join(g.config.DataRootDir, "maps"),
		g.palette,
		g.NewActor,
		g.NewItem,
		g.NewObject,
	)
	g.pathfinder = NewPathfinder(g.ensureMapIsLoaded)
	// this will initialise the volatile state, which will be reset on each new game
	g.init()
	return g
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
	} else {
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

func (g *GameState) PlayerInteractAtPosition(interactionPoint geometry.Point) {
	if g.OpenContextMenuFor(interactionPoint) {
		return
	}

	if g.currentMap().IsActorAt(interactionPoint) {
		actor := g.currentMap().ActorAt(interactionPoint)
		g.ui.OpenTextWindow(actor.GetDetailInfo())
		return
	}

	// Move..
	moveDistance := g.currentMap().MoveDistance(g.Player.Position(), interactionPoint)
	if moveDistance == 1 {
		direction := interactionPoint.Sub(g.Player.Position())
		g.ManualMovePlayer(direction.ToDirection())
		return
	}

	if g.currentMap().IsCurrentlyPassable(interactionPoint) || g.currentMap().IsObjectAt(interactionPoint) {
		pathTo := g.currentMap().GetJPSPath(g.Player.Position(), interactionPoint, func(point geometry.Point) bool {
			if point != interactionPoint {
				return g.currentMap().IsWalkableFor(point, g.Player)
			}
			return g.currentMap().IsWalkableFor(point, g.Player) || g.currentMap().IsObjectAt(point)
		})

		if len(pathTo) > 0 {
			g.Player.CurrentPath = pathTo
			g.Player.CurrentPathBlockedCount = 0
			g.RunPlayerPath()
		}
	}
	// assume movement to this position..
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
	g.ShowDateTime()
}

func (g *GameState) ShowDateTime() {
	g.msg(foundation.Msg(fmt.Sprintf("Time is now %s", g.gameTime.Time.Format("Monday 15:04, 2006-01-02"))))
}

func (g *GameState) LightAt(p geometry.Point) fxtools.HDRColor {
	return g.currentMap().LightAt(p, g.Player.Position(), g.Player.CanSee(p), g.gameTime.Time)
}

func (g *GameState) PlayerInteractInDirection(direction geometry.CompassDirection) {
	positionOnMap := g.GetPlayerPosition().Add(direction.ToPoint())
	g.OpenContextMenuFor(positionOnMap)
}

func (g *GameState) loadUserFuncs(dir string) map[string]*govaluate.EvaluableExpression {
	records, _ := recfile.ReadAndClose(fxtools.MustOpen(path.Join(dir, "definitions", "userFuncs.rec")))

	queries := make(map[string]*govaluate.EvaluableExpression)
	for _, queryRecord := range records {
		var queryName string
		var queryExpression string
		for _, field := range queryRecord {
			switch strings.ToLower(field.Name) {
			case "name":
				queryName = field.Value
			case "expr":
				queryExpression = field.Value
			}
		}
		expression, _ := govaluate.NewEvaluableExpressionWithFunctions(queryExpression, g.GetScriptFuncs())
		queries[queryName] = expression
	}

	return queries
}

func (g *GameState) GetPlayerNameAndIcon() (string, textiles.TextIcon) {
	return g.config.PlayerName, textiles.TextIcon{
		Char: g.config.PlayerChar,
		Fg:   g.palette.Get(g.config.PlayerColor),
	}
}

// NewActorFromName creates a new actor from a template name
// NOTE: These actors won't have a valid SpawnPosition set
func (g *GameState) NewActorFromName(actorName string, spawnMapName string) *Actor {
	actorRec, exists := g.globalActorTemplates[actorName]
	if !exists {
		return nil
	}
	actor, _ := g.NewActor(actorRec, spawnMapName)
	return actor
}

// NewActor creates a new actor from a record
// It will also load the schedule for the actor, initialize the FSM and set and return the spawn position
func (g *GameState) NewActor(rec recfile.Record, mapName string) (*Actor, geometry.Point) {
	newActor := NewActorFromRecord(rec, g.palette, g.NewItemFromString)
	if newActor != nil {
		g.loadSchedule(newActor)
		g.FSMInit(newActor)
		newActor.InitWithGameState(g)
		spawnPos := newActor.Position()
		newActor.SpawnPosition = spawnPos
		newActor.SpawnMapName = mapName
		return newActor, spawnPos
	}
	panic(fmt.Sprintf("Could not create actor from record: %v", rec))
	return nil, geometry.Point{}
}

func (g *GameState) FSMInit(newActor *Actor) {
	newActor.FSM = NewActorFSM(g, newActor, DefaultBehaviorFactory)
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
	g.playerLightSource = &gridmap.LightSource{
		Pos:          geometry.Point{},
		Radius:       5,
		Color:        fxtools.HDRColor{R: 1, G: 1, B: 1, A: 1},
		MaxIntensity: 1,
	}
	g.mapContainsPlayer = true
	g.timeTracker = make(TimeTracker)
	g.iconsForItems, g.inventoryColors = loadIconsForItems(path.Join(g.config.DataRootDir, "definitions"), g.palette)

	g.activeMaps = make(map[string]GameMap)
	g.chatterCache = make(map[*Actor]map[foundation.ChatterTopic][]EntriesWithCondition)
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

	g.Scripts = NewScriptRunner()
	g.metronome = &Metronome{}
}

func loadMapNames(rootDir string) []string {
	mapDir := path.Join(rootDir, "maps")
	files, _ := os.ReadDir(mapDir)
	var mapNames []string
	for _, file := range files {
		if file.IsDir() {
			mapNames = append(mapNames, file.Name())
		}
	}
	return mapNames
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
		startRecord := records[0]
		for _, field := range startRecord {
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

	g.playerAttachHooks()

	playerSheet.HealAPAndHPCompletely()

	g.transitionToMapLocation(spawnMap, spawnLocation)

	g.journal.Update()

	g.SaveTimeNow("PlayerLastAteAt")
}

func (g *GameState) transitionToMapLocation(levelName string, location string) {
	// Remove Player from Old Map
	if g.currentMap() != nil && g.Player != nil {
		g.currentMap().RemoveActor(g.Player)
		g.Player.RemoveLevelStatusEffects()
		g.currentMap().SetLastVisited(g.gameTime.Time)
	}

	loadedMap := g.ensureMapIsLoaded(levelName)
	g.currentMapName = loadedMap.GetName()

	mapVisited := fmt.Sprintf("PlayerVisited(%s)", levelName)
	if !g.gameFlags.HasFlag(mapVisited) {
		g.awardXP(100, fmt.Sprintf("Discovered '%s'", loadedMap.GetDisplayName()))
	}
	g.gameFlags.Increment(mapVisited)

	// Ensure correct map state
	g.currentMap().UpdateBakedLights()

	g.currentMap().UpdateDynamicLights()

	g.ui.PlayMusic(path.Join(g.config.DataRootDir, "audio", "music", g.currentMap().GetMeta().MusicFile+".ogg"))

	// Spawn Player
	playerSpawnPosition := loadedMap.GetNamedLocation(location)
	g.currentMap().AddActorWithDisplacement(g.Player, playerSpawnPosition)

	g.afterPlayerMoved(geometry.Point{}, true)

	g.advanceTime(time.Second * time.Duration(60))

	g.updateUIStatus()
}

func (g *GameState) playerAttachHooks() {
	g.Player.InitWithGameState(g)

	// additional player only hooks
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

// UIReady is called by the UI when it has initialized itself
func (g *GameState) UIReady(ui foundation.GameUI) {
	g.ui = ui

	g.moveIntoDungeon()

	g.Scripts.CheckAndRunFrames()
}

func (g *GameState) Reset() {
	g.init()
	g.moveIntoDungeon()
}

// moveIntoDungeon requires the UI to be available. It will request a dungeon crawl UI
// and then moves the player into the loaded map.
func (g *GameState) moveIntoDungeon() {
	g.initPlayerAndMap()

	g.afterPlayerMoved(geometry.Point{}, true)

	g.updateUIStatus()
}

func (g *GameState) QueueActionAfterAnimation(action func()) {
	g.afterAnimationActions = append(g.afterAnimationActions, action)
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

func (g *GameState) tryAddRandomChatter(actor *Actor, textType foundation.ChatterTopic) bool {
	chatter := g.GetRandomChatter(actor, textType)
	if chatter == "" {
		chatter = textType.DefaultChatter()
	}
	return g.tryAddChatter(actor, chatter)
}

func (g *GameState) tryAddChatter(actor *Actor, text string) bool {
	if text == "" || actor.currentMapName != g.Player.currentMapName {
		return false
	}
	if actor.IsAlive() && !actor.IsSleeping() && g.Player.CanSee(actor.Position()) {
		text = g.FillTemplatedText(text)
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
		g.onActorCriminalActivity(g.Player, victim, foundation.ChatterMinorCrimeNoticed)
		return false
	}
}

func (g *GameState) getItemTemplateByName(shortString string) recfile.Record {
	if record, ok := g.globalItemTemplates[shortString]; ok {
		return record
	}
	return nil
}

func (g *GameState) IsInShootingRange(attacker *Actor, defender *Actor) bool {
	if attacker.currentMapName != defender.currentMapName {
		return false
	}
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
	if one.currentMapName != two.currentMapName {
		return false
	}
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

// advanceTimeAndTurn will advance the time and turn, and run any active scripts on the map and also update the schedules of all actors
func (g *GameState) advanceTimeAndTurn(duration time.Duration) {
	g.gameTime = g.gameTime.AddDurationAndTurn(duration)
	g.Scripts.CheckAndRunFrames()
}

// advanceTime will advance the time, and run any active scripts on the map and also update the schedules of all actors
// IMPORTANT: This function will also UPDATE the FOV and Dijkstra map of the player
func (g *GameState) advanceTime(duration time.Duration) {
	g.gameTime = g.gameTime.AddDuration(duration)
	g.Scripts.CheckAndRunFrames()
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

func (g *GameState) GetRandomChatter(talker *Actor, chatterType foundation.ChatterTopic) string {
	if talker.ChatterFile == "" {
		return ""
	}
	var chatterForActor map[foundation.ChatterTopic][]EntriesWithCondition
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

func NewChatterFromRecords(records []recfile.Record, condFuncs map[string]govaluate.ExpressionFunction) map[foundation.ChatterTopic][]EntriesWithCondition {
	chatter := make(map[foundation.ChatterTopic][]EntriesWithCondition)
	for _, record := range records {
		var chatterType foundation.ChatterTopic
		var entries []string
		var condition *govaluate.EvaluableExpression
		for _, field := range record {
			switch strings.ToLower(field.Name) {
			case "s":
				chatterType = foundation.NewChatterTopicFromString(field.Value)
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

	didRemove := originMap.RemoveActor(actor)
	if didRemove {
		targetLocation := destMap.GetNamedLocation(whereTo.TargetLocation)
		destMap.AddActorWithDisplacement(actor, targetLocation)
	}
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
		actor.Inventory.RemoveItem(item)
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
			if actor.HasFlag(foundation.FlagSleep) || !actor.IsAlive() {
				continue
			}
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

type TimedSpawn struct {
	Actor     *Actor
	SpawnTime time.Time
	Location  geometry.Point
}

func (g *GameState) inventoryColorCode(item foundation.Item) string {
	return textiles.RGBAToFgColorCode(g.inventoryColors[item.GetCategory()])
}

func (g *GameState) isDetectedByObserver(actor *Actor, observer *Actor) bool {
	outsideDetectionRange := !observer.CanDetect(actor.Position())

	if outsideDetectionRange {
		return false
	}

	if !actor.IsSneaky() {
		return true
	}

	//SLEEPING?
	if observer.IsSleeping() {
		if actor.IsSneaky() {
			return false
		}
		wakeMod := actor.GetCharSheet().GetSkill(d100.SkillForSneak)
		chanceToWake := (observer.GetCharSheet().GetStat(d100.Perception) * 5) - wakeMod
		awakened := d100.SuccessRoll(d100.Percentage(chanceToWake), 0)
		if awakened.Success {
			observer.GetFlags().Unset(foundation.FlagSleep)
			g.msg(foundation.HiLite("%s wakes up", observer.Name()))
		} else {
			return false
		}
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
		g.Player.CanSee(actor.Position()) &&
		actor.ChatterFile != "" &&
		actor.GetFlags().Get(foundation.FlagTurnsSinceLastIdleChatter) > 40 &&
		rand.Intn(4) == 0
}

func (g *GameState) setGameTime(start time.Time) {
	g.gameTime = g.gameTime.WithTime(start)
}

func (g *GameState) IterateAllActors(iterator func(mapName string, actor *Actor) bool) {
	for _, activeMap := range g.activeMaps {
		for _, actor := range activeMap.Actors() {
			if !iterator(activeMap.GetName(), actor) {
				return
			}
		}
	}
}

func (g *GameState) IterateAllObjects(iterator func(mapName string, object Object) bool) {
	for _, activeMap := range g.activeMaps {
		for _, object := range activeMap.Objects() {
			if !iterator(activeMap.GetName(), object) {
				return
			}
		}
	}
}

func (g *GameState) isBedAt(position geometry.Point) (*Bed, bool) {
	if obj, isObjectAt := g.currentMap().TryGetObjectAt(position); isObjectAt {
		if bed, isBed := obj.(*Bed); isBed {
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

func (g *GameState) OpenTeleportMenu() {
	var menuItems []foundation.MenuItem
	for _, mapName := range g.allMapNames {
		menuItems = append(menuItems, foundation.MenuItem{
			Name: mapName,
			Action: func() {
				g.openTeleportToLocationMenu(mapName)
			},
			CloseMenus: true,
		})
	}

	g.ui.OpenMenu(menuItems)
}

func (g *GameState) openTeleportToLocationMenu(mapName string) {
	var menuItems []foundation.MenuItem
	loadedMap := g.ensureMapIsLoaded(mapName)
	locations := loadedMap.GetNamedLocations()
	for locationName, _ := range locations {
		menuItems = append(menuItems, foundation.MenuItem{
			Name: locationName,
			Action: func() {
				g.transitionToMapLocation(mapName, locationName)
			},
			CloseMenus: true,
		})
	}

	g.ui.OpenMenu(menuItems)
}

func (g *GameState) updateFoVAndDijkstraMapForAllActors() {
	if g.currentMapName == "" || g.Player == nil {
		return
	}
	g.updateFoVAndDijkstraMap(g.Player)
	for _, actor := range g.currentMap().Actors() {
		actor.SetDijkstraMapDirty()
		actor.SetFoVDirty()
	}
}

func (g *GameState) TestForcedMonologue() {
	grim := g.actorWithName("grim_beard")
	g.ui.ForceListening(grim, []string{"Hi there", "Can you hear me?", "Are you fine?"}, nil)
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
