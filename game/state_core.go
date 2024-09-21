package game

import (
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"RogueUI/special"
	"cmp"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)

type PointInTime struct {
	Turns int
	Time  time.Time
}

func (t PointInTime) AddDuration(duration time.Duration) PointInTime {
	return PointInTime{
		Turns: t.Turns,
		Time:  t.Time.Add(duration),
	}
}

func (t PointInTime) AddDurationAndTurn(duration time.Duration) PointInTime {
	return PointInTime{
		Turns: t.Turns + 1,
		Time:  t.Time.Add(duration),
	}
}

func (t PointInTime) WithTurns(turns int) PointInTime {
	return PointInTime{
		Turns: turns,
		Time:  t.Time,
	}
}

func (t PointInTime) WithTime(time time.Time) PointInTime {
	return PointInTime{
		Turns: t.Turns,
		Time:  time,
	}
}

func (t PointInTime) TurnsSince(inTime PointInTime) int {
	return t.Turns - inTime.Turns
}

func (t PointInTime) MinutesSince(inTime PointInTime) int {
	return int(t.Time.Sub(inTime.Time).Minutes())
}

func (t PointInTime) HoursSince(inTime PointInTime) int {
	return int(t.Time.Sub(inTime.Time).Hours())
}

func (t PointInTime) DaysSince(inTime PointInTime) int {
	return int(t.Time.Sub(inTime.Time).Hours() / 24)
}

type TimeTracker map[string]PointInTime

func (t TimeTracker) String() string {
	if len(t) == 0 {
		return "No time tracked"
	}
	out := make([]string, len(t))
	i := 0
	for key, value := range t {
		out[i] = fmt.Sprintf("%s: %s", key, value.Time.Format("2006-01-02 15:04"))
		i++
	}
	slices.SortStableFunc(out, func(i, j string) int {
		return cmp.Compare(i, j)
	})
	return strings.Join(out, "\n")
}

type Timed interface {
	ShouldActivate(tickCount int) bool
	IsAlive(tickCount int) bool
	String() string
}
type TimedFunc struct {
	Timed
	ActionOnThing            func()
	TickCount                int
	ActivateBeforeLeavingMap bool
}

func (f TimedFunc) WithTickCount(i int) TimedFunc {
	f.TickCount = i
	return f
}

type Metronome struct {
	timed []TimedFunc
}

func (m *Metronome) AddTimed(timed Timed, activateOnLeave bool, action func()) {
	m.timed = append(m.timed, TimedFunc{
		Timed:                    timed,
		ActionOnThing:            action,
		TickCount:                0,
		ActivateBeforeLeavingMap: activateOnLeave,
	})
}
func (m *Metronome) LeavingMapEvents() bool {
	activate := fxtools.FilterSlice(m.timed, func(timed TimedFunc) bool {
		return timed.ActivateBeforeLeavingMap
	})
	if len(activate) == 0 {
		return false
	}
	for _, timed := range activate {
		timed.ActionOnThing()
	}
	m.timed = fxtools.FilterSlice(m.timed, func(timed TimedFunc) bool {
		return timed.IsAlive(timed.TickCount)
	})
	return true
}
func (m *Metronome) Tick() {
	m.timed = fxtools.FilterSlice(m.timed, func(timed TimedFunc) bool {
		return timed.IsAlive(timed.TickCount)
	})
	for i, timed := range m.timed {
		if timed.ShouldActivate(timed.TickCount) {
			timed.ActionOnThing()
		}
		m.timed[i] = timed.WithTickCount(timed.TickCount + 1)
	}
}

func (m *Metronome) HasTimed(item Timed) bool {
	for _, timed := range m.timed {
		if timed.Timed == item {
			return true
		}
	}
	return false
}

func (m *Metronome) String() string {
	if len(m.timed) == 0 {
		return "No timed things"
	}
	out := make([]string, len(m.timed))
	for i, timed := range m.timed {
		out[i] = fmt.Sprintf("%d: %s", timed.TickCount, timed.Timed.String())
	}
	return fmt.Sprintf("Timed things:\n%s", strings.Join(out, "\n"))
}

type GameState struct {
	// Global State (Needs to be saved)
	gameTime             PointInTime
	gameFlags            *fxtools.StringFlags
	timeTracker          TimeTracker
	logBuffer            []foundation.HiLiteString
	terminalGuesses      map[string][]string
	journal              *Journal
	rewardTracker        *RewardTracker
	showEverything       bool
	flagsChangedThisTurn bool

	// Player State (Needs to be saved)
	Player            *Actor
	playerFoV         *geometry.FOV
	playerLightSource *gridmap.LightSource
	playerDijkstraMap map[geometry.Point]int
	playerLastAimedAt special.BodyPart

	// Map State
	mapLoader MapLoader

	// Scripts
	scriptRunner *ScriptRunner
	metronome    *Metronome

	// Maps (Needs to be saved)
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

	// Temporary State
	chatterCache map[*Actor]map[foundation.ChatterType][]EntriesWithCondition
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
	isContextAvailable := g.OpenContextMenuFor(pos)

	if !isContextAvailable {
		moveDistance := g.currentMap().MoveDistance(g.Player.Position(), pos)
		if moveDistance == 1 {
			direction := pos.Sub(g.Player.Position())
			g.ManualMovePlayer(direction.ToDirection())
			return
		}
		if g.currentMap().IsCurrentlyPassable(pos) {
			g.Player.RemoveGoal()

			pathTo := g.currentMap().GetJPSPath(g.Player.Position(), pos, func(point geometry.Point) bool {
				return g.currentMap().IsWalkableFor(point, g.Player)
			})

			if len(pathTo) > 0 {
				g.Player.currentPath = pathTo
				g.Player.currentPathIndex = 0
				g.Player.currentPathBlockedCount = 0
				g.Player.SetGoal(ActorGoal{
					Action: func(g *GameState, a *Actor) int {
						return moveTowards(g, a, pos)
					},
					Achieved: func(g *GameState, a *Actor) bool {
						return a.Position() == pos || a.cannotFindPath()
					},
				})
				g.RunPlayerPath(true)
			}
		}
		// assume movement to this position..
	}
}

func (g *GameState) IsPlayerOverEncumbered() bool {
	return g.Player.IsOverEncumbered()
}

func (g *GameState) GetPlayerName() string {
	return g.Player.Name()
}

func (g *GameState) GetPlayerCharSheet() *special.CharSheet {
	return g.Player.charSheet
}

func (g *GameState) GetMapDisplayName() string {
	return g.currentMap().GetDisplayName()
}

func (g *GameState) currentMap() *gridmap.GridMap[*Actor, *Item, Object] {
	return g.activeMaps[g.currentMapName]
}

func (g *GameState) WizardAdvanceTime() {
	g.advanceTime(time.Minute * 30)
	g.msg(foundation.Msg(fmt.Sprintf("Time is now %s", g.gameTime.Time.Format("15:04"))))
}

func (g *GameState) LightAt(p geometry.Point) fxtools.HDRColor {
	return g.currentMap().LightAt(p, g.gameTime.Time)
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
		timeTracker:         make(TimeTracker),
		visionRange:         80,
		palette:             palette,
		globalItemTemplates: loadItemTemplates(config.DataRootDir),
		activeMaps:          make(map[string]*gridmap.GridMap[*Actor, *Item, Object]),
		chatterCache:        make(map[*Actor]map[foundation.ChatterType][]EntriesWithCondition),
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
		spawnPos := newActor.Position()
		newActor.SpawnPosition = spawnPos
		return newActor, spawnPos
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
func (g *GameState) NewObject(rec recfile.Record, newMap *gridmap.GridMap[*Actor, *Item, Object]) (Object, geometry.Point) {
	object := g.NewObjectFromRecord(rec, g.palette, newMap)
	if object != nil {
		objectPos := object.Position()
		return object, objectPos
	}
	panic(fmt.Sprintf("Could not create object from record: %v", rec))
	return nil, geometry.Point{}
}
func (g *GameState) setIconsForObjects(iconsForObjects map[string]textiles.TextIcon) {
	g.iconsForObjects = iconsForObjects
}
func (g *GameState) init() {
	g.iconsForItems, g.inventoryColors = loadIconsForItems(path.Join(g.config.DataRootDir, "definitions"), g.palette)

	g.mapLoader = gridmap.NewRecMapLoader(
		path.Join(g.config.DataRootDir, "maps"),
		g.palette,
		g.setIconsForObjects,
		g.NewActor,
		g.NewItem,
		g.NewObject,
	)

	g.gameTime = PointInTime{
		Turns: 0,
		Time:  time.Date(2077, 2, 5, 16, 20, 23, 0, time.UTC),
	}

	g.logBuffer = []foundation.HiLiteString{}
	g.showEverything = false

	g.gameFlags = fxtools.NewStringFlags()

	g.terminalGuesses = make(map[string][]string)

	g.rewardTracker = NewRewardTracker(fxtools.MustOpen(path.Join(g.config.DataRootDir, "definitions", "xp_rewards.rec")), g.getScriptFuncs())
	g.journal = NewJournal(fxtools.MustOpen(path.Join(g.config.DataRootDir, "definitions", "journal.rec")), g.getScriptFuncs())
	g.hookupJournalAndFlags()

	g.scriptRunner = NewScriptRunner()
	g.metronome = &Metronome{}
}

func (g *GameState) hookupJournalAndFlags() {
	g.journal.SetChangeHandler(func() {
		g.ui.PlayCue("ui/journal")
		g.msg(foundation.HiLite(">>> Journal updated <<<"))
	})
	g.gameFlags.SetChangeHandler(func(flagName string, value int) {
		g.flagsChangedThisTurn = true
	})
}

func (g *GameState) initPlayerAndMap() {
	playerSheet := special.NewCharSheet()
	playerSheet.AddSkillPoints(0)

	//playerSheet.SetSkillAdjustment(special.SmallGuns, 50)
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
				itemName := field.Value
				g.giveAndTryEquipItem(g.Player, g.NewItemFromString(itemName))
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
	loadedMap.UpdateBakedLights()
	loadedMap.UpdateDynamicLights()

	g.setCurrentMap(loadedMap)

	g.iconsForObjects = loadedMapResult.IconsForObjects

	for flagName, flagValue := range loadedMapResult.FlagsOfMap {
		g.gameFlags.Set(flagName, flagValue)
	}

	for _, script := range loadedMapResult.ScriptsToRun {
		g.RunScriptByName(script)
	}

	g.attachHooksToPlayer()

	g.journal.OnFlagsChanged()

	g.updatePlayerFoVAndApplyExploration()

	g.ui.PlayMusic(path.Join(g.config.DataRootDir, "audio", "music", loadedMap.GetMeta().MusicFile+".ogg"))
}
func (g *GameState) attachHooksToPlayer() {
	equipment := g.Player.GetEquipment()
	g.Player.GetFlags().SetOnChangeHandler(func(flag special.ActorFlag, value int) {
		g.ui.UpdateStats()
	})
	g.Player.GetCharSheet().SetOnStatChangeHandler(func(stat special.Stat) {
		g.ui.UpdateStats()
	})

	g.Player.GetInventory().SetOnChangeHandler(g.ui.UpdateInventory)

	g.Player.GetInventory().SetOnBeforeRemove(equipment.UnEquip)

	g.Player.GetCharSheet().SetSkillModifierHandler(g.Player.GetInventory().GetSkillModifiersFromItems)

	g.Player.GetCharSheet().SetStatModifierHandler(g.Player.GetInventory().GetStatModifiersFromItems)

	equipment.SetOnChangeHandler(g.updateUIStatus)
}

func (g *GameState) setCurrentMap(loadedMap *gridmap.GridMap[*Actor, *Item, Object]) {
	mapName := loadedMap.GetName()
	g.activeMaps[mapName] = loadedMap
	g.currentMapName = mapName
}

// UIRunning is called by the UI when it has started the game loop
func (g *GameState) UIRunning() {
	g.ui.InitDungeonUI(g.palette, g.inventoryColors)
}

// UIReady is called by the UI when it has initialized itself
func (g *GameState) UIReady() {
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
	g.advanceTimeAndTurn(time.Second * time.Duration(float64(playerTimeTakenForTurn)/10))

	didCancel := g.ui.AnimatePending() // animate player actions..

	g.scriptRunner.CheckAndRunFrames(g.currentMap().GetName())

	g.metronome.Tick()

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

	if g.flagsChangedThisTurn {
		g.checkJournalAndRewards()
		g.flagsChangedThisTurn = false
	}

	g.updateUIStatus()

	g.checkPlayerCanAct()

	g.updatePlayerFoVAndApplyExploration()
}

func (g *GameState) checkJournalAndRewards() {
	g.journal.OnFlagsChanged()

	rewards := g.rewardTracker.GetNewRewards(nil)
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
	if geometry.DistanceChebyshev(victim.Position(), position) <= 1 {
		return true
	}
	return g.currentMap().IsLineOfSightClear(victim.Position(), position, g.IsSomethingBlockingTargetingAtLoc)
}

func (g *GameState) tryAddRandomChatter(actor *Actor, textType foundation.ChatterType) bool {
	chatter := g.GetRandomChatter(actor, textType)
	return g.tryAddChatter(actor, chatter)
}
func (g *GameState) tryAddChatter(actor *Actor, text string) bool {
	if text == "" {
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
	actorEquipment := actor.GetEquipment()
	stealableItems := itemStacksForUI(actor.GetInventory().StackedItemsWithFilter(actorEquipment.IsNotEquipped))

	rightToLeft := func(itemUI foundation.ItemForUI, amount int) {
		if amount == 0 {
			return
		}
		itemStack := itemUI.(*InventoryStack)
		g.PlayerStealOrPlantItem(actor, itemStack.First(), true)
	}

	leftToRight := func(itemUI foundation.ItemForUI, amount int) {
		if amount == 0 {
			return
		}
		itemStack := itemUI.(*InventoryStack)
		g.PlayerStealOrPlantItem(actor, itemStack.First(), false)
	}

	leftName := g.Player.Name()
	rightName := actor.Name()
	playerItems := itemStacksForUI(g.Player.GetInventory().StackedItems())

	g.ui.ShowGiveAndTakeContainer(leftName, playerItems, rightName, stealableItems, rightToLeft, leftToRight)
}

func (g *GameState) PlayerStealOrPlantItem(victim *Actor, item *Item, isSteal bool) {
	itemStealModifier := 0
	if item.GetCategory().IsEasySteal() {
		itemStealModifier = 10
	} else if item.GetCategory().IsHardSteal() {
		itemStealModifier = -10
	}
	if victim.IsSleeping() {
		itemStealModifier += 75
	}

	var transferFunc func(*Item)
	if isSteal {
		transferFunc = func(itemTaken *Item) {
			victim.GetInventory().RemoveItem(item)
			g.Player.GetInventory().AddItem(item)
			g.msg(foundation.HiLite("You steal %s", item.Name()))
			g.ui.PlayCue("world/pickup")
		}
	} else {
		transferFunc = func(itemTaken *Item) {
			g.Player.GetInventory().RemoveItem(item)
			victim.GetInventory().AddItem(item)
			g.msg(foundation.HiLite("You plant %s", item.Name()))
			g.ui.PlayCue("world/drop")
		}
	}

	skillRoll := g.Player.GetCharSheet().SkillRoll(special.Steal, itemStealModifier)
	if skillRoll.Success {
		transferFunc(item)
		g.StartPickpocket(victim)
	} else {
		g.msg(foundation.HiLite("%s notices your hands in his pockets", victim.Name()))
		if victim.IsSleeping() {
			victim.WakeUp()
		}
		g.trySetHostile(victim, g.Player)
	}
}

func (g *GameState) ShowDateTime() {
	// full date
	g.msg(foundation.HiLite("The time is %s", g.gameTime.Time.Format("2006-01-02 15:04")))
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

func (g *GameState) IsInShootingRange(attacker *Actor, defender *Actor) bool {
	attackerPos := attacker.Position()
	defenderPos := defender.Position()
	moveDistance := geometry.Distance(attackerPos, defenderPos)

	attackerWeapon, hasWeapon := attacker.GetEquipment().GetRangedWeapon()
	if !hasWeapon {
		return moveDistance <= 1
	}
	weaponRange := float64(attackerWeapon.GetCurrentAttackMode().MaxRange)

	inRange := moveDistance <= weaponRange

	visible := g.canActorSee(attacker, defenderPos)

	return inRange && visible
}

func (g *GameState) IsInTalkingRange(one *Actor, two *Actor) bool {
	onePos := one.Position()
	twoPos := two.Position()
	moveDistance := geometry.Distance(onePos, twoPos)

	inRange := moveDistance <= 6

	visible := g.canActorSee(one, twoPos)

	return inRange && visible
}

func (g *GameState) getShootingRangePosition(attacker *Actor, weaponRange int, victim *Actor) geometry.Point {
	attackerPos := attacker.Position()
	victimPos := victim.Position()

	mapAroundVictim := g.currentMap().GetDijkstraMap(victimPos, weaponRange, g.currentMap().IsCurrentlyPassable)

	// find the best position to shoot from
	bestPos := attackerPos
	bestDistance := geometry.Distance(attackerPos, victimPos)

	for pos, dist := range mapAroundVictim {
		distFloat := float64(dist) / 10
		if !g.canActorSee(victim, pos) {
			continue
		}
		if distFloat < bestDistance {
			bestPos = pos
			bestDistance = distFloat
		}
	}
	return bestPos
}
func (g *GameState) advanceTimeAndTurn(duration time.Duration) {
	g.gameTime = g.gameTime.AddDurationAndTurn(duration)
}
func (g *GameState) advanceTime(duration time.Duration) {
	g.gameTime = g.gameTime.AddDuration(duration)
	g.scriptRunner.CheckAndRunFrames(g.currentMap().GetName())
	g.updatePlayerFoVAndApplyExploration()
}

func (g *GameState) awardXP(xp int, text string) {
	didLevelUpNow := g.Player.GetCharSheet().AddXP(xp)
	g.msg(foundation.HiLite("You received %s "+text, strconv.Itoa(xp)+" XP"))
	if didLevelUpNow {
		g.ui.PlayCue("ui/LEVELUP")
		g.msg(foundation.HiLite(">>> You have gone up a level <<<"))
	}
}

func (g *GameState) actorHitMessage(victim *Actor, damage SourcedDamage, cripple bool, kill bool, overKill bool) {
	if victim == g.Player {
		g.playerHitMessage(damage, cripple)
		return
	}
	baseMessage := fmt.Sprintf("%s was hit for %d hit points", victim.Name(), damage.DamageAmount)
	if damage.BodyPart != special.Body {
		baseMessage += fmt.Sprintf("%s was hit in the %s for %d hit points", victim.Name(), damage.BodyPart.String(), damage.DamageAmount)
	}

	if kill {
		baseMessage += fmt.Sprintf(", killing them")
	} else if overKill {
		baseMessage += fmt.Sprintf(", reducing them to a bloody pulp")
	} else if cripple {
		baseMessage += fmt.Sprintf(", crippling them")
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
	if talker.chatterFile == "" {
		return ""
	}
	var chatterForActor map[foundation.ChatterType][]EntriesWithCondition
	if _, hasCached := g.chatterCache[talker]; !hasCached {
		chatterFilePath := path.Join(g.config.DataRootDir, "dialogues", talker.chatterFile+".rec")
		records := recfile.Read(fxtools.MustOpen(chatterFilePath))
		g.chatterCache[talker] = NewChatterFromRecords(records, g.getScriptFuncs())
	}

	chatterForActor = g.chatterCache[talker]

	if chatterList, hasChatter := chatterForActor[chatterType]; hasChatter {
		var chosenChatter EntriesWithCondition
		for _, chatter := range chatterList {
			if chatter.Condition == nil {
				chosenChatter = chatter
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
			if field.Name == "s" {
				chatterType = foundation.NewChatterTypeFromString(field.Value)
			} else if field.Name == "t" {
				entries = append(entries, field.Value)
			} else if field.Name == "c" {
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

func (g *GameState) removeItemFromGame(item *Item) {
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
