package game

import (
	"RogueUI/dungen"
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/gridmap"
	"RogueUI/rpg"
	"RogueUI/util"
	"cmp"
	"encoding/gob"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"
	"slices"
	"time"
)

// state changes / animations
// act_move(actor, from, to) & anim_move(actor, from, to)
// act_melee(actorOne, actorTwo) & anim_melee(actorOne, actorTwo)

// map window size in a console : 23x80

type GameState struct {
	Player *Actor

	TurnsTaken int

	ui foundation.GameUI

	textIcons []foundation.TextIcon

	logBuffer []foundation.HiLiteString

	gridMap       *gridmap.GridMap[*Actor, *Item, *Object]
	dungeonLayout *dungen.DungeonMap

	currentDungeonLevel              int
	maximumDungeonLevel              int
	deepestDungeonLevelPlayerReached int

	tileStyle         int
	defaultBackground color.RGBA
	defaultForeground color.RGBA

	playerDijkstraMap     map[geometry.Point]int
	showEverything        bool
	playerName            string
	dataDefinitions       DataDefinitions
	identification        *IdentificationKnowledge
	afterAnimationActions []func()

	playerFoV               *geometry.FOV
	visionRange             int
	genericWallIconIndex    foundation.TextIcon
	playerIcon              rune
	playerColor             string
	config                  *foundation.Configuration
	ascensionsWithoutAmulet int
}

func (g *GameState) IncreaseSkillLevel(skill rpg.SkillName) {
	if g.Player.charSheet.HasCharPointsLeft() {
		g.Player.charSheet.IncreaseSkillLevel(skill)
	} else {
		g.msg(foundation.Msg("You have no more character points to spend"))
	}

	g.updateUIStatus()
}

func (g *GameState) IncreaseAttributeLevel(stat rpg.Stat) {
	if stat == rpg.FatiguePoints && g.Player.charSheet.GetLevelAdjustments(rpg.FatiguePoints) >= 7 {
		g.msg(foundation.Msg("You cannot increase your fatigue points any further"))
		return
	}
	if g.Player.charSheet.HasCharPointsLeft() {
		g.Player.charSheet.Increment(stat)
		g.updateUIStatus()
	} else {
		g.msg(foundation.Msg("You have no more character points to spend"))
	}
}

func (g *GameState) ItemAt(loc geometry.Point) foundation.ItemForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsItemAt(loc) {
		itemAt := g.gridMap.ItemAt(loc)
		return itemAt
	}
	return nil
}

func (g *GameState) ObjectAt(loc geometry.Point) foundation.ObjectCategory {
	if g.gridMap.Contains(loc) && g.gridMap.IsObjectAt(loc) {
		objectAt := g.gridMap.ObjectAt(loc)
		return objectAt.ObjectIcon()
	}
	return -1
}

func (g *GameState) ActorAt(loc geometry.Point) foundation.ActorForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsActorAt(loc) {
		actorAt := g.gridMap.ActorAt(loc)
		return actorAt
	}
	return nil
}

func (g *GameState) AimedShot() {
	equipment := g.Player.GetEquipment()
	if !equipment.HasMissileQuivered() {
		g.msg(foundation.Msg("You have no quivered missile"))
		return
	}
	item := equipment.GetNextQuiveredMissile()
	g.startRangedAttackWithMissile(item)
}

func (g *GameState) QuickShot() {
	equipment := g.Player.GetEquipment()
	if !equipment.HasMissileQuivered() {
		g.msg(foundation.Msg("You have no quivered missile"))
		return
	}
	item := equipment.GetNextQuiveredMissile()

	enemies := g.playerVisibleEnemiesByDistance()
	preselectedTarget := g.Player.Position()
	if len(enemies) == 0 {
		g.msg(foundation.Msg("No enemies in sight"))
		return
	}
	preselectedTarget = enemies[0].Position()
	g.actorRangedAttackWithMissile(g.Player, item, g.Player.Position(), preselectedTarget)
}

func (g *GameState) OpenTacticsMenu() {
	var menuItems []foundation.MenuItem
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Aimed Attack",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "All-Out Attack",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "All-Out Defense",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Feint",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name: "Charge Attack",
		Action: func() {
			g.startAimZapEffect("charge_attack")
		},
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name: "Start Sprinting",
		Action: func() {
			g.startSprint(g.Player)
		},
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Toggle Acrobatic Dodge",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Defend & Retreat",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Dive for cover",
		Action:     nil,
		CloseMenus: true,
	})

	g.ui.OpenMenu(menuItems)
}

func (g *GameState) OpenHitLocationMenu() {
	var menuItems []foundation.MenuItem
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Torso (0)",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Vitals (-3) -> 3x DMG w/ piercing",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Skull (-7, +2 DR) -> 4x DMG w/ criticals against head",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Eye (-9) -> Like skull hit without +2DR",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Legs (-2) -> limb loss at 1/2 MAX HP DMG",
		Action:     nil,
		CloseMenus: true,
	})
	menuItems = append(menuItems, foundation.MenuItem{
		Name:       "Whatever location presents itself",
		Action:     nil,
		CloseMenus: true,
	})

	g.ui.OpenMenu(menuItems)
}

func (g *GameState) GetCharacterSheet() []string {

	actor := g.Player
	basicActorInfo := actor.GetDetailInfo()

	return basicActorInfo
}

func (g *GameState) GetPlayerPosition() geometry.Point {
	return g.Player.Position()
}

func (g *GameState) QueueActionAfterAnimation(action func()) {
	g.afterAnimationActions = append(g.afterAnimationActions, action)
}
func (g *GameState) IsLit(pos geometry.Point) bool {
	return g.gridMap.IsTileLit(pos)
}

func (g *GameState) IsExplored(loc geometry.Point) bool {
	if !g.gridMap.Contains(loc) {
		return false
	}
	return g.gridMap.IsExplored(loc)
}

func (g *GameState) IsVisibleToPlayer(loc geometry.Point) bool {
	if !g.gridMap.Contains(loc) {
		return false
	}

	// special abilities
	canSeeFood := g.Player.HasFlag(foundation.FlagSeeFood)
	if g.IsFoodAt(loc) && canSeeFood {
		return true
	}

	canSeeMonsters := g.Player.HasFlag(foundation.FlagSeeMonsters)
	if g.gridMap.IsActorAt(loc) && canSeeMonsters {
		return true
	}

	canSeeMagic := g.Player.HasFlag(foundation.FlagSeeMagic)
	if g.gridMap.IsItemAt(loc) && canSeeMagic && g.gridMap.ItemAt(loc).IsMagic() {
		return true
	}

	if !g.gridMap.IsExplored(loc) {
		return false
	}
	isVisibleToPlayer := g.canPlayerSee(loc) || g.showEverything

	return isVisibleToPlayer
}

func (g *GameState) ChooseItemForUseOrZap() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsUsableOrZappable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything usable."))
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Use what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.playerUseOrZapItem(item)
	})
}

func (g *GameState) IsSomethingBlockingTargetingAtLoc(point geometry.Point) bool {
	return !g.gridMap.IsCurrentlyPassable(point)
}

func (g *GameState) OpenWizardMenu() {
	g.ui.OpenMenu([]foundation.MenuItem{
		{
			Name:       "Show Map",
			Action:     g.revealAll,
			CloseMenus: true,
		},
		{
			Name: "Load Test Map",
			Action: func() {
				g.GotoNamedLevel("line_room")
			},
			CloseMenus: true,
		},
		{
			Name: "Goto Town",
			Action: func() {
				g.GotoNamedLevel("town")
			},
			CloseMenus: true,
		},
		{
			Name: "250 Char Points",
			Action: func() {
				g.Player.AddCharacterPoints(250)
			},
		},
		{
			Name:   "Create Item",
			Action: g.openWizardCreateItemMenu,
		},
		{
			Name:   "Create Monster",
			Action: g.openWizardCreateMonsterMenu,
		},
		{
			Name:   "Create Trap",
			Action: g.openWizardCreateTrapMenu,
		},
	})
}

func NewGameState(ui foundation.GameUI, config *foundation.Configuration) *GameState {
	g := &GameState{
		config:              config,
		playerName:          config.PlayerName,
		playerColor:         "White",
		playerIcon:          '@',
		maximumDungeonLevel: 26,
		ui:                  ui,
		tileStyle:           0,
		dataDefinitions:     GetDataDefinitions(),
		playerFoV:           geometry.NewFOV(geometry.NewRect(0, 0, config.MapWidth, config.MapHeight)),
		visionRange:         14,
	}
	g.init()
	ui.SetGame(g)

	return g
}
func (g *GameState) giveAndEquipItem(actor *Actor, item *Item) {
	actor.GetInventory().Add(item)
	actor.GetEquipment().Equip(item)
}
func (g *GameState) init() {
	g.Player = NewPlayer(g.playerName, g.playerIcon, g.playerColor)

	g.giveAndEquipItem(g.Player, g.NewItemFromName("main_gauche"))
	g.giveAndEquipItem(g.Player, g.NewItemFromName("leather_armor"))
	for i := 0; i < 20; i++ {
		g.giveAndEquipItem(g.Player, g.NewItemFromName("arrow"))
	}
	g.giveAndEquipItem(g.Player, g.NewItemFromName("short_bow"))

	equipment := g.Player.GetEquipment()
	g.Player.GetFlags().SetOnChangeHandler(func(flag foundation.ActorFlag, value int) {
		g.ui.UpdateStats()
	})
	g.Player.charSheet.SetStatChangedHandler(g.ui.UpdateStats)
	g.Player.charSheet.SetResourceChangedHandler(g.ui.UpdateStats)

	g.Player.GetInventory().SetOnChangeHandler(g.ui.UpdateInventory)

	g.Player.GetInventory().SetOnBeforeRemove(equipment.UnEquip)

	equipment.SetOnChangeHandler(g.updateUIStatus)

	g.identification = NewIdentificationKnowledge()
	g.identification.SetOnIdChanged(g.ui.UpdateInventory)

	g.identification.MixScrolls(g.dataDefinitions.GetScrollInternalNames())
	g.identification.MixPotions(g.dataDefinitions.GetPotionInternalNames())
	g.identification.MixWands(g.dataDefinitions.GetWandInternalNames())
	g.identification.MixRings(g.dataDefinitions.GetRingInternalNames())

	g.identification.SetAlwaysIDOnUse(g.dataDefinitions.AlwaysIDOnUseInternalNames())

	g.TurnsTaken = 0
	g.logBuffer = []foundation.HiLiteString{}
	g.currentDungeonLevel = 0
	g.deepestDungeonLevelPlayerReached = 0
	g.showEverything = false
}

func (g *GameState) NewItemFromName(name string) *Item {
	def := g.dataDefinitions.GetItemDefByName(name)
	return NewItem(def, g.identification)
}
func (g *GameState) NewGold(amount int) *Item {
	def := ItemDef{
		Name:         "gold",
		InternalName: "gold",
		Category:     foundation.ItemCategoryGold,
		Charges:      rpg.NewDice(1, 1, 1),
	}
	item := NewItem(def, g.identification)
	item.SetCharges(amount)
	return item
}
func (g *GameState) Reset() {
	g.init()
	g.moveIntoDungeon()
	g.ui.UpdateInventory()
}

func (g *GameState) IsSomethingInterestingAtLoc(loc geometry.Point) bool {
	gridMap := g.gridMap

	if g.Player.Position() == loc {
		return true
	}

	if gridMap.IsActorAt(loc) {
		return true
	}

	if gridMap.IsItemAt(loc) {
		return true
	}

	return false
}

func (g *GameState) IsEquipped(items foundation.ItemForUI) bool {
	itemStack, isItem := items.(*InventoryStack)
	if !isItem {
		return false
	}
	return g.Player.GetEquipment().IsEquipped(itemStack.First())
}

func (g *GameState) GetVisibleEnemies() []foundation.ActorForUI {
	return actorsForUI(g.playerVisibleEnemiesByDistance())
}

func (g *GameState) UIReady() {
	g.moveIntoDungeon()
	// ADD Banner
	//g.ui.ShowTextFileFullscreen(path.Join("data","banner.txt"), g.moveIntoDungeon)
}

func (g *GameState) moveIntoDungeon() {
	g.ui.InitDungeonUI()
	g.GotoDungeonLevel(1, StairsBoth, true)
}

func (g *GameState) updateUIStatus() {
	g.ui.UpdateVisibleEnemies()
	g.ui.UpdateStats()
	g.ui.UpdateLogWindow()
	g.ui.UpdateInventory()
}
func (g *GameState) GetHudFlags() map[foundation.ActorFlag]int {
	return g.Player.GetFlags().UnderlyingCopy()
}

func (g *GameState) GetHudStats() map[foundation.HudValue]int {
	uiStats := make(map[foundation.HudValue]int)
	//g.Player.stats

	uiStats[foundation.HudTurnsTaken] = g.TurnsTaken
	uiStats[foundation.HudDungeonLevel] = g.currentDungeonLevel
	uiStats[foundation.HudGold] = g.Player.GetGold()

	uiStats[foundation.HudHitPoints] = g.Player.GetHitPoints()
	uiStats[foundation.HudHitPointsMax] = g.Player.GetHitPointsMax()

	uiStats[foundation.HudFatiguePoints] = g.Player.GetFatiguePoints()
	uiStats[foundation.HudFatiguePointsMax] = g.Player.GetFatiguePointsMax()

	uiStats[foundation.HudStrength] = g.Player.GetStrength()
	uiStats[foundation.HudDexterity] = g.Player.GetDexterity()
	uiStats[foundation.HudIntelligence] = g.Player.GetIntelligence()
	uiStats[foundation.HudMeleeSkill] = g.Player.getMeleeSkillInUse()
	uiStats[foundation.HudRangedSkill] = g.Player.GetSkill(rpg.SkillNameMissileWeapons)
	uiStats[foundation.HudDamageResistance] = g.Player.GetDamageResistance()

	return uiStats
}

func (g *GameState) GetLog() []foundation.HiLiteString {
	return g.logBuffer
}
func (g *GameState) GetMapInfo(pos geometry.Point) foundation.HiLiteString {
	return g.QueryMap(pos, false)
}
func (g *GameState) GetMapInfoForMovement(pos geometry.Point) foundation.HiLiteString {
	return g.QueryMap(pos, true)
}
func (g *GameState) QueryMap(pos geometry.Point, isMovement bool) foundation.HiLiteString {
	if !g.gridMap.Contains(pos) {
		return foundation.NoMsg()
	}
	if g.gridMap.IsActorAt(pos) && g.Player.Position() != pos {
		actor := g.gridMap.ActorAt(pos)
		return foundation.HiLite("You see %s here", actor.Name())
	}
	if g.gridMap.IsItemAt(pos) {
		item := g.gridMap.ItemAt(pos)
		return foundation.HiLite("You see %s here", item.Name())
	}
	if g.gridMap.IsObjectAt(pos) {
		object := g.gridMap.ObjectAt(pos)
		return foundation.HiLite("You see %s here", object.Name())
	}

	cell := g.gridMap.GetCell(pos)
	if !cell.TileType.IsSpecial() && isMovement {
		return foundation.NoMsg()
	}
	tileDesc := cell.TileType.DefinedDescription
	return foundation.HiLite("You see %s here", tileDesc)
}

func (g *GameState) getPlayerRoom() *dungen.DungeonRoom {
	if g.dungeonLayout == nil {
		return nil
	}
	playerRoom := g.dungeonLayout.GetRoomAt(g.Player.Position())
	return playerRoom
}

func (g *GameState) msg(message foundation.HiLiteString) {
	if !message.IsEmpty() {
		g.appendLogMessage(message)
		g.ui.UpdateLogWindow()
	}
}

func (g *GameState) appendLogMessage(message foundation.HiLiteString) {
	if len(g.logBuffer) == 0 {
		g.logBuffer = append(g.logBuffer, message)
		return
	}

	lastLogMessageIndex := len(g.logBuffer) - 1
	lastLogMessage := g.logBuffer[lastLogMessageIndex]

	if message.IsEqual(lastLogMessage) {
		lastLogMessage.Repetitions++
		g.logBuffer[lastLogMessageIndex] = lastLogMessage
		return
	}

	g.logBuffer = append(g.logBuffer, message)
}

// adapted from https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/command.c#L35
func (g *GameState) endPlayerTurn() {
	// player has changed the game state..

	g.identification.SetCurrentItemInUse("") // reset item in use

	g.TurnsTaken++

	didCancel := g.ui.AnimatePending() // animate player actions..

	playerTimeTakeForTurn := 100 / (g.Player.GetBasicSpeed())

	g.enemyMovement(playerTimeTakeForTurn)

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

func (g *GameState) OpenInventory() {
	inventory := g.GetInventory()
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything."))
		return
	}
	g.ui.OpenInventoryForManagement(inventory)
}
func (g *GameState) ChooseItemForThrow() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsThrowable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything throwable."))
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Throw what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.startRangedAttackWithMissile(item)
	})
}

func (g *GameState) ChooseItemForMissileLaunch() {
	equipment := g.Player.GetEquipment()
	if !equipment.HasMissileLauncherEquipped() {
		g.msg(foundation.Msg("You are not carrying a missile launcher."))
		return
	}

	launcher := equipment.GetMissileLauncher()
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsMissile() && item.GetWeapon().IsLaunchedWith(launcher.GetWeapon().GetWeaponType())
	})

	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything that can be launched with this launcher."))
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Launch what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.startRangedAttackWithMissile(item)
	})
}

func (g *GameState) GetInventory() []foundation.ItemForUI {
	return itemStacksForUI(g.Player.GetInventory().StackedItems())
}

func (g *GameState) MapAt(mapPos geometry.Point) foundation.TileType {
	if !g.gridMap.Contains(mapPos) {
		return foundation.TileEmpty
	}
	mapCell := g.gridMap.GetCell(mapPos)
	return mapCell.TileType.Icon()
}
func (g *GameState) TopEntityAt(mapPos geometry.Point) foundation.EntityType {
	if !g.gridMap.Contains(mapPos) {
		return foundation.EntityTypeOther
	}

	mapCell := g.gridMap.GetCell(mapPos)

	if mapCell.Actor != nil {
		actor := *mapCell.Actor
		if actor.IsDrawn(g.Player.HasFlag(foundation.FlagSeeInvisible)) {
			return foundation.EntityTypeActor
		}
	}

	if mapCell.Item != nil {
		return foundation.EntityTypeItem
	}

	if mapCell.Object != nil {
		object := *mapCell.Object
		if object.IsDrawn() {
			return foundation.EntityTypeObject
		}
	}

	return foundation.EntityTypeWorldTile
}
func (g *GameState) addItemToMap(item *Item, mapPos geometry.Point) {
	g.gridMap.AddItemWithDisplacement(item, mapPos)
}

func (g *GameState) removeItemFromInventory(holder *Actor, item *Item) {
	equipment := holder.GetEquipment()
	inventory := holder.GetInventory()
	if equipment.IsQuiveredItem(item) && item.IsMissile() {
		nextInStack := inventory.RemoveAndGetNextInStack(item)
		if nextInStack != nil {
			equipment.Equip(nextInStack)
		}
	} else {
		inventory.Remove(item)
	}
}

func (g *GameState) playerVisibleEnemiesByDistance() []*Actor {
	var enemies []*Actor
	if g.Player == nil || g.gridMap == nil {
		return enemies
	}
	playerPos := g.Player.Position()
	for _, actor := range g.gridMap.Actors() {
		if actor == g.Player {
			continue
		}
		if g.canPlayerSee(actor.Position()) {
			enemies = append(enemies, actor)
		}
	}
	slices.SortStableFunc(enemies, func(i, j *Actor) int {
		distI := geometry.Distance(playerPos, i.Position())
		distJ := geometry.Distance(playerPos, j.Position())
		return cmp.Compare(distI, distJ)
	})
	return enemies
}
func (g *GameState) GetVisibleItems() []foundation.ItemForUI {
	return itemStacksForUI(StacksFromItems(g.playerVisibleItemsByDistance()))
}
func (g *GameState) playerVisibleItemsByDistance() []*Item {
	playerPos := g.Player.Position()
	var visibleItems []*Item
	for _, item := range g.gridMap.Items() {
		if g.canPlayerSee(item.Position()) {
			visibleItems = append(visibleItems, item)
		}
	}
	slices.SortStableFunc(visibleItems, func(i, j *Item) int {
		distI := geometry.Distance(playerPos, i.Position())
		distJ := geometry.Distance(playerPos, j.Position())
		return cmp.Compare(distI, distJ)
	})
	return visibleItems
}

// Should be functionally identical to bool cansee(y, x) in chase.c
// https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/chase.c#L387
func (g *GameState) canPlayerSee(pos geometry.Point) bool {
	if g.currentDungeonLevel == 0 {
		return true
	}
	playerRoom := g.getPlayerRoom()
	playerPos := g.Player.Position()
	if g.dungeonLayout == nil {
		return g.playerFoV.Visible(pos)
	}
	isAtDoor := g.dungeonLayout.IsDoorAt(playerPos)
	isInRoom := playerRoom != nil
	isRoomLit := playerRoom != nil && playerRoom.IsLit()
	if pos == playerPos {
		return true
	}
	chebyDist := geometry.DistanceChebyshev(playerPos, pos)
	nextToUs := chebyDist == 1
	if nextToUs { // we will always see something right next us, save invisibility checks, blindness, etc.
		return true
	}

	// Case 1 corridor - we only see stuff right next to us
	if !isInRoom {
		return false
	}

	// Case 2 Special Case doors, we can see inside a lit room
	actorInSameRoom := isInRoom && playerRoom.ContainsIncludingWalls(pos)
	if isAtDoor {
		if actorInSameRoom && isRoomLit {
			return true
		} else {
			return false
		}
	}

	// Case 3 in a room
	if !actorInSameRoom {
		return false
	}

	if isRoomLit {
		return true
	} else {
		return false
	}
}

func (g *GameState) GetFilteredInventory(filter func(item *Item) bool) []foundation.ItemForUI {
	items := g.Player.GetInventory().StackedItemsWithFilter(filter)
	return itemStacksForUI(items)

}
func (g *GameState) hasPaidWithCharge(user *Actor, item *Item) bool {
	if item == nil { // no item = intrinsic effect
		return true
	}
	if item.charges == 0 {
		g.msg(foundation.Msg("The item is out of charges"))
		return false
	}
	if item.charges > 0 {
		item.charges--
		if item.charges == 0 { // destroy
			g.removeItemFromInventory(user, item)
		}
	}
	return true
}

func (g *GameState) defaultStyleIcon(icon rune) foundation.TextIcon {
	return foundation.TextIcon{
		Rune: icon,
		Fg:   g.defaultForeground,
		Bg:   g.defaultBackground,
	}
}

func (g *GameState) NewEnemyFromDef(def MonsterDef) *Actor {
	actor := NewActor(def.Name, def.Icon, def.Color)
	actor.GetFlags().Init(def.Flags.UnderlyingCopy())
	actor.SetIntrinsicZapEffects(def.ZapEffects)
	actor.SetIntrinsicUseEffects(def.UseEffects)
	actor.SetInternalName(def.InternalName)

	actor.SetSizeModifier(def.SizeModifier)

	actor.charSheet.SetStat(rpg.Strength, def.Strength)
	actor.charSheet.SetStat(rpg.Dexterity, def.Dexterity)
	actor.charSheet.SetStat(rpg.Intelligence, def.Intelligence)
	actor.charSheet.SetStat(rpg.Health, def.Health)
	actor.charSheet.SetStat(rpg.Will, def.Will)
	actor.charSheet.SetStat(rpg.Perception, def.Perception)
	actor.charSheet.SetStat(rpg.FatiguePoints, def.FatiguePoints)
	actor.charSheet.SetStat(rpg.HitPoints, max(1, def.HitPoints)) // hack to avoid 0 hp actors which would despawn immediately
	actor.charSheet.SetStat(rpg.BasicSpeed, def.BasicSpeed)
	actor.charSheet.ResetResources()

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	if random.Intn(100) < def.CarryChance {
		if random.Intn(100) < def.CarryChance {
			itemDef := g.dataDefinitions.PickItemForLevel(random, def.DungeonLevel)
			item := NewItem(itemDef, g.identification)
			actor.GetInventory().Add(item)
		} else {
			actor.AddGold(def.Gold.Roll())
		}
	}

	actor.SetIntrinsicAttacks(def.Attacks)
	return actor
}

func (g *GameState) actorKilled(causeOfDeath string, victim *Actor) {
	if victim == g.Player {
		g.QueueActionAfterAnimation(func() {
			g.gameOver(causeOfDeath)
		})
		return
	}
	g.msg(foundation.HiLite("%s killed %s", causeOfDeath, victim.Name()))

	g.dropInventory(victim)
}

func (g *GameState) revealAll() {
	g.gridMap.SetAllExplored()
	g.showEverything = true
}

func (g *GameState) isInPlayerRoom(position geometry.Point) bool {
	playerRoom := g.getPlayerRoom()
	if playerRoom == nil {
		return false
	}
	return playerRoom.ContainsIncludingWalls(position)
}

func (g *GameState) spawnEntities(random *rand.Rand, level int, newMap *gridmap.GridMap[*Actor, *Item, *Object], dungeon *dungen.DungeonMap) {

	playerRoom := dungeon.GetRoomAt(g.Player.Position())

	mustSpawnAmuletOfYendor := level == 26 && !g.Player.GetInventory().HasItemWithName("amulet_of_yendor")

	spawnItemsInRoom := func(room *dungen.DungeonRoom, itemCount int) {
		for i := 0; i < itemCount; i++ {
			spawnPos, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			itemDef := g.dataDefinitions.PickItemForLevel(random, level)
			item := NewItem(itemDef, g.identification)
			newMap.AddItem(item, spawnPos)
		}
	}

	spawnMonstersInRoom := func(room *dungen.DungeonRoom, monsterCount int) {
		for i := 0; i < monsterCount; i++ {
			spawnPos, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			monsterDef := g.dataDefinitions.PickMonsterForLevel(random, level)
			monster := g.NewEnemyFromDef(monsterDef)
			if random.Intn(26) >= level {
				monster.GetFlags().Set(foundation.FlagSleep)
			}
			if monster.HasFlag(foundation.FlagWallCrawl) {
				walls := room.GetWalls()
				spawnPos = walls[random.Intn(len(walls))]
			}
			newMap.AddActor(monster, spawnPos)
		}
	}

	spawnObjectsInRoom := func(room *dungen.DungeonRoom, objectCount int) {
		for i := 0; i < objectCount; i++ {
			spawnPos, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			trapEffects := foundation.GetAllTrapCategories()
			randomEffect := trapEffects[random.Intn(len(trapEffects))]
			object := g.NewTrap(randomEffect)
			newMap.AddObject(object, spawnPos)
		}
	}

	allRooms := dungeon.AllRooms()
	randomRoomOrder := random.Perm(len(allRooms))
	for _, roomIndex := range randomRoomOrder {
		room := allRooms[roomIndex]

		itemCount := random.Intn(3)
		spawnItemsInRoom(room, itemCount)

		if random.Intn(2) == 0 || itemCount > 0 {
			monsterCount := random.Intn(max(2, itemCount)) + 1
			spawnMonstersInRoom(room, monsterCount)
		}

		if level > 1 && room != playerRoom && random.Intn(4) == 0 {
			objectCount := random.Intn(3) + 1
			spawnObjectsInRoom(room, objectCount)
		}

		if mustSpawnAmuletOfYendor {
			spawnPos, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			amulet := g.NewItemFromName("amulet_of_yendor")
			newMap.AddItem(amulet, spawnPos)
			mustSpawnAmuletOfYendor = false
		}
	}
}

func (g *GameState) openWizardCreateItemMenu() {
	allCategories := []foundation.ItemCategory{
		foundation.ItemCategoryFood,
		foundation.ItemCategoryWeapons,
		foundation.ItemCategoryArmor,
		foundation.ItemCategoryAmulets,
		foundation.ItemCategoryPotions,
		foundation.ItemCategoryScrolls,
		foundation.ItemCategoryRings,
		foundation.ItemCategoryWands,
	}
	var menuActions []foundation.MenuItem

	for _, c := range allCategories {
		category := c
		menuActions = append(menuActions, foundation.MenuItem{
			Name: category.String(),
			Action: func() {
				g.openWizardCreateItemSelectionMenu(g.dataDefinitions.Items[category])
			},
			CloseMenus: true,
		})
	}

	g.ui.OpenMenu(menuActions)
}

func (g *GameState) openWizardCreateItemSelectionMenu(defs []ItemDef) {
	var menuActions []foundation.MenuItem
	for _, def := range defs {
		itemDef := def
		menuActions = append(menuActions, foundation.MenuItem{
			Name: itemDef.Name,
			Action: func() {
				newItem := NewItem(itemDef, g.identification)
				inv := g.Player.GetInventory()
				if inv.IsFull() {
					g.gridMap.AddItemWithDisplacement(newItem, g.Player.Position())
				} else {
					inv.Add(newItem)
				}
			},
			CloseMenus: true,
		})
	}
	g.ui.OpenMenu(menuActions)
}
func (g *GameState) openWizardCreateMonsterMenu() {
	defs := g.dataDefinitions.Monsters
	var menuActions []foundation.MenuItem
	for _, def := range defs {
		monsterDef := def
		menuActions = append(menuActions, foundation.MenuItem{
			Name: monsterDef.Name,
			Action: func() {
				if monsterDef.Flags.IsSet(foundation.FlagWallCrawl) {
					g.spawnCrawlerInWall(monsterDef)
				} else {
					newActor := g.NewEnemyFromDef(monsterDef)
					g.gridMap.AddActorWithDisplacement(newActor, g.Player.Position())
				}
			},
			CloseMenus: true,
		})
	}
	g.ui.OpenMenu(menuActions)
}

func (g *GameState) openWizardCreateTrapMenu() {
	trapTypes := foundation.GetAllTrapCategories()
	var menuActions []foundation.MenuItem
	for _, def := range trapTypes {
		trapType := def
		menuActions = append(menuActions, foundation.MenuItem{
			Name: trapType.String(),
			Action: func() {
				random := rand.New(rand.NewSource(time.Now().UnixNano()))
				trapPos := g.gridMap.GetRandomFreeAndSafeNeighbor(random, g.Player.Position())
				newTrap := g.NewTrap(trapType)
				newTrap.SetHidden(false)
				g.gridMap.AddObject(newTrap, trapPos)
			},
			CloseMenus: true,
		})
	}
	g.ui.OpenMenu(menuActions)
}

func (g *GameState) spawnCrawlerInWall(monsterDef MonsterDef) {
	playerRoom := g.getPlayerRoom()
	if playerRoom == nil {
		return
	}
	walls := playerRoom.GetWalls()
	spawnPos := walls[rand.Intn(len(walls))]
	newActor := g.NewEnemyFromDef(monsterDef)
	g.gridMap.ForceSpawnActorInWall(newActor, spawnPos)
}
func (g *GameState) calculateTotalNetWorth() int {
	return g.Player.GetGold()
}
func (g *GameState) gameWon() {
	scoreInfo := foundation.ScoreInfo{
		PlayerName:         g.Player.Name(),
		Gold:               g.calculateTotalNetWorth(),
		MaxLevel:           g.deepestDungeonLevelPlayerReached,
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
		MaxLevel:           g.deepestDungeonLevelPlayerReached,
		DescriptiveMessage: death,
		Escaped:            false,
	}
	highScores := g.writePlayerScore(scoreInfo)
	g.ui.ShowGameOver(scoreInfo, highScores)
}

func (g *GameState) IsFoodAt(loc geometry.Point) bool {
	return g.gridMap.IsItemAt(loc) && g.gridMap.ItemAt(loc).IsFood()
}

func (g *GameState) IsBlockingRay(point geometry.Point) bool {
	return !g.gridMap.IsCurrentlyPassable(point)
}

func (g *GameState) updatePlayerFoVAndApplyExploration() {
	g.gridMap.UpdateFieldOfView(g.playerFoV, g.Player.Position(), g.visionRange)
	for _, pos := range g.playerFoV.Visibles {
		g.gridMap.SetExplored(pos)
	}
}

func (g *GameState) checkPlayerCanAct() {
	// idea
	// check if the player can act before giving back control to him
	// if he cannot act, eg, he is stunned and forced to do nothing,
	// then check the end condition for this status effect
	// if it's not reached, we want the UI to show a message about the situation
	// the player has to confirm it and then we can end the turn
	if !g.Player.HasFlag(foundation.FlagStun) && !g.Player.HasFlag(foundation.FlagHeld) {
		return
	}

	if g.Player.HasFlag(foundation.FlagStun) {
		_, result, _ := rpg.SuccessRoll(g.Player.GetIntelligence() + g.Player.GetFlags().Get(foundation.FlagStun) - 1)

		if result.IsSuccess() {
			g.msg(foundation.Msg("You shake off the stun"))
			g.Player.GetFlags().Unset(foundation.FlagStun)
			return
		}
		g.Player.GetFlags().Increment(foundation.FlagStun)

		g.msg(foundation.Msg("You are stunned and cannot act"))

		// TODO: animate a small delay here?
		g.endPlayerTurn()
	}
	if g.Player.HasFlag(foundation.FlagHeld) {
		_, result, marginOfSuccess := rpg.SuccessRoll(g.Player.GetStrength())

		if result.IsCriticalSucces() {
			g.msg(foundation.Msg("You break free from the hold"))
			g.Player.GetFlags().Unset(foundation.FlagHeld)
			return
		} else if result.IsSuccess() {
			g.Player.GetFlags().Decrease(foundation.FlagHeld, marginOfSuccess)
		}

		g.msg(foundation.Msg("You are held and cannot act"))

		// TODO: animate a small delay here?
		g.endPlayerTurn()
	}
}

func (g *GameState) customBehaviours(internalName string) (func(actor *Actor), bool) {
	switch internalName {
	case "xeroc_2":
		return g.aiWallMimic, true
	}
	return nil, false

}

func (g *GameState) aiWallMimic(actor *Actor) {
	stillCloaked := actor.HasFlag(foundation.FlagInvisible)

	if !stillCloaked {
		g.defaultBehaviour(actor)
		return
	}

	sameRoom := g.isInPlayerRoom(actor.Position())

	if !sameRoom {
		return
	}

	if stillCloaked && rand.Intn(5) == 0 {
		uncloakAnim := uncloakAndCharge(g, actor, g.Player.Position())
		g.ui.AddAnimations(uncloakAnim)
		return
	}
}

func (g *GameState) writePlayerScore(info foundation.ScoreInfo) []foundation.ScoreInfo {
	scoresFile := "scores.bin"

	scoreTable := LoadHighScoreTable(scoresFile)

	scoreTable = append(scoreTable, info) // add score

	slices.SortStableFunc(scoreTable, func(i, j foundation.ScoreInfo) int {
		if i.Escaped && !j.Escaped {
			return -1
		}
		if !i.Escaped && j.Escaped {
			return 1
		}
		if i.Escaped && j.Escaped {
			return cmp.Compare(j.Gold, i.Gold)
		}
		if i.MaxLevel == j.MaxLevel {
			return cmp.Compare(j.Gold, i.Gold)
		}
		return cmp.Compare(j.MaxLevel, i.MaxLevel)
	})

	if len(scoreTable) > 15 {
		scoreTable = scoreTable[:15]
	}

	saveHighScoreTable(scoresFile, scoreTable)

	return scoreTable
}

func (g *GameState) triggerTileEffectsAfterMovement(actor *Actor, oldPos, newPos geometry.Point) []foundation.Animation {
	isPlayer := actor == g.Player
	cell := g.gridMap.GetCell(newPos)
	if cell.TileType.IsVendor() && isPlayer {
		itemsForVendor := []util.Tuple[foundation.ItemForUI, int]{
			{g.NewItemFromName("mace"), 100},
		}
		g.ui.OpenVendorMenu(itemsForVendor, g.buyItemFromVendor)
	}
	if g.gridMap.IsObjectAt(newPos) {
		objectAt := g.gridMap.ObjectAt(newPos)
		var animations []foundation.Animation
		if isPlayer {
			playerMoveAnim := g.ui.GetAnimMove(g.Player, oldPos, newPos)
			playerMoveAnim.RequestMapUpdateOnFinish()
			animations = append(animations, playerMoveAnim)
		}
		triggeredEffectAnimations := objectAt.OnWalkOver()
		animations = append(animations, triggeredEffectAnimations...)
		return animations
	}
	return nil
}

func (g *GameState) buyItemFromVendor(item foundation.ItemForUI, price int) {
	player := g.Player
	if !player.HasGold(price) {
		g.msg(foundation.Msg("You cannot afford that"))
		return
	}
	if player.GetInventory().IsFull() {
		g.msg(foundation.Msg("You cannot carry more items"))
		return
	}
	player.RemoveGold(price)
	i := item.(*Item)
	player.GetInventory().Add(i)
}

func (g *GameState) newLevelReached(level int) {
	g.Player.AddCharacterPoints(10)
	g.msg(foundation.HiLite("You've been awarded 10 character points for reaching level %s", fmt.Sprint(level)))
}

func (g *GameState) checkTilesForHiddenObjects(tiles []geometry.Point) {
	var noticedSomething bool
	for _, tile := range tiles {
		if g.gridMap.IsObjectAt(tile) {
			object := g.gridMap.ObjectAt(tile)
			if object.IsHidden() {
				perception := g.Player.GetPerception()
				_, result, _ := rpg.SuccessRoll(perception)
				if result.IsSuccess() {
					noticedSomething = true
				}
				if result.IsCriticalSucces() {
					object.SetHidden(false)
				}
			}
		}
	}

	if noticedSomething {
		g.msg(foundation.Msg("you feel like something is wrong with this room"))
	}
}

func (g *GameState) dropInventory(victim *Actor) {
	goldAmount := victim.GetGold()
	if goldAmount > 0 {
		g.addItemToMap(g.NewGold(goldAmount), victim.Position())
	}
	for _, item := range victim.GetInventory().Items() {
		g.addItemToMap(item, victim.Position())
	}
}

func (g *GameState) startSprint(actor *Actor) {
	if actor.GetFatiguePoints() < 1 {
		return
	}
	actor.LooseFatigue(1)
	haste(g, actor)
}

func (g *GameState) unstableStairs() bool {
	if g.currentDungeonLevel >= 26 {
		return false
	}
	if g.Player.GetInventory().HasItemWithName("amulet_of_yendor") {
		return false
	}
	chance := util.Clamp(0, 0.9, float64(min(10, g.ascensionsWithoutAmulet))/10.0)
	return rand.Float64() < chance
}

func saveHighScoreTable(scoresFile string, scoreTable []foundation.ScoreInfo) {
	file := util.CreateFile(scoresFile)
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err := encoder.Encode(scoreTable)
	if err != nil {
		log.Fatal(err)
	}
}

func LoadHighScoreTable(scoresFile string) []foundation.ScoreInfo {
	var scoreTable []foundation.ScoreInfo
	if util.FileExists(scoresFile) { // read from file
		file, err := os.Open(scoresFile)
		if err != nil {
			log.Fatal(err)
		}
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&scoreTable)
		if err != nil {
			log.Fatal(err)
		}

		file.Close()
	}
	return scoreTable
}

func NewItem(def ItemDef, id *IdentificationKnowledge) *Item {
	charges := 1
	if def.Charges.NotZero() {
		charges = def.Charges.Roll()
	}
	item := &Item{
		name:         def.Name,
		internalName: def.InternalName,
		category:     def.Category,
		charges:      charges,
		slot:         def.Slot,
		flags:        foundation.NewBitFlags(),
		id:           id,
		stat:         def.Stat,
		statBonus:    def.StatBonus,
		skill:        def.Skill,
		skillBonus:   def.SkillBonus,
		equipFlag:    def.EquipFlag,
		thrownDamage: def.ThrowDamageDice,
	}

	if def.IsValidWeapon() {
		item.weapon = &WeaponInfo{
			damageDice:       def.WeaponDef.DamageDice,
			weaponType:       def.WeaponDef.Type,
			launchedWithType: def.WeaponDef.LaunchedWithType,
			skillUsed:        def.WeaponDef.SkillUsed,
			damagePlus:       0,
		}
	}

	if def.IsValidArmor() {
		item.armor = &ArmorInfo{
			damageResistance: def.ArmorDef.DamageResistance,
			plus:             0,
			encumbrance:      def.ArmorDef.Encumbrance,
		}
	}

	item.zapEffectName = def.ZapEffect
	item.useEffectName = def.UseEffect

	return item

}
