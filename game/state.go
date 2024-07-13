package game

import (
	"RogueUI/dice_curve"
	"RogueUI/dungen"
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/gridmap"
	"RogueUI/recfile"
	"RogueUI/special"
	"RogueUI/util"
	"cmp"
	"encoding/gob"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"
	"path"
	"slices"
	"strings"
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

	gridMap       *gridmap.GridMap[*Actor, *Item, Object]
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
	afterAnimationActions []func()

	playerFoV               *geometry.FOV
	visionRange             int
	genericWallIconIndex    foundation.TextIcon
	playerIcon              rune
	playerColor             string
	config                  *foundation.Configuration
	ascensionsWithoutAmulet int
	mapLoader               MapLoader
	gameFlags               util.StringFlags
	terminalGuesses         map[string][]string
}

func (g *GameState) GetRangedHitChance(target foundation.ActorForUI) int {
	defender := target.(*Actor)
	var posInfos special.PosInfo
	posInfos.ObstacleCount = 0
	posInfos.Distance = g.gridMap.MoveDistance(g.Player.Position(), defender.Position())
	posInfos.IlluminationPenalty = 0
	equippedWeapon, hasWeapon := g.Player.GetEquipment().GetMainHandItem()
	if !hasWeapon || !equippedWeapon.IsRangedWeapon() {
		return 0
	}
	weaponSkill := equippedWeapon.GetWeapon().GetSkillUsed()
	return special.RangedChanceToHit(posInfos, g.Player.GetCharSheet(), weaponSkill, defender.GetCharSheet(), special.Body)
}

func (g *GameState) GetItemInMainHand() (foundation.ItemForUI, bool) {
	return g.Player.GetEquipment().GetMainHandItem()
}

func (g *GameState) GetBodyPartsAndHitChances(targeted foundation.ActorForUI) []util.Tuple[string, int] {
	victim := targeted.(*Actor)
	attackerSkill, defenderSkill := 0, 0 // TODO
	return victim.GetBodyPartsAndHitChances(attackerSkill, defenderSkill)
}

func (g *GameState) GetRandomEnemyName() string {
	return g.dataDefinitions.RandomMonsterDef().Name
}

func (g *GameState) ItemAt(loc geometry.Point) foundation.ItemForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsItemAt(loc) {
		itemAt := g.gridMap.ItemAt(loc)
		return itemAt
	}
	return nil
}

func (g *GameState) ObjectAt(loc geometry.Point) foundation.ObjectForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsObjectAt(loc) {
		objectAt := g.gridMap.ObjectAt(loc)
		return objectAt
	}
	return nil
}

func (g *GameState) ActorAt(loc geometry.Point) foundation.ActorForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsActorAt(loc) {
		actorAt := g.gridMap.ActorAt(loc)
		return actorAt
	}
	return nil
}
func (g *GameState) DownedActorAt(loc geometry.Point) foundation.ActorForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsDownedActorAt(loc) {
		actorAt := g.gridMap.DownedActorAt(loc)
		return actorAt
	}
	return nil
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

	canSeeTraps := g.Player.HasFlag(foundation.FlagSeeTraps)
	if g.gridMap.IsObjectAt(loc) && canSeeTraps {
		objectAt := g.gridMap.ObjectAt(loc)
		if objectAt.IsTrap() {
			return true
		}
	}

	if !g.gridMap.IsExplored(loc) {
		return false
	}
	isVisibleToPlayer := g.canPlayerSee(loc) || g.showEverything

	return isVisibleToPlayer
}

func (g *GameState) IsSomethingBlockingTargetingAtLoc(point geometry.Point) bool {
	return !g.gridMap.IsCurrentlyPassable(point)
}

func (g *GameState) OpenDialogueNode(conversation *Conversation, node ConversationNode, isTerminal bool) {
	endConversation := false
	var effectCalls []func()
	for _, effect := range node.Effects {
		if effect == "EndConversation" {
			endConversation = true
		} else {
			if util.LooksLikeAFunction(effect) {
				name, args := util.GetNameAndArgs(effect)
				switch name {
				case "Hacking":
					terminalID := args.Get(0)
					difficulty := args.Get(1)
					flagName := args.Get(2)
					followUpNode := args.Get(3)

					effectCalls = append(effectCalls, func() {
						previousGuesses := g.terminalGuesses[terminalID]
						g.ui.StartHackingGame(util.MurmurHash(flagName), foundation.DifficultyFromString(difficulty), previousGuesses, func(pGuesses []string, result foundation.InteractionResult) {
							g.terminalGuesses[terminalID] = pGuesses
							if result == foundation.Success {
								g.gameFlags.SetFlag(flagName)
							}
							nextNode := conversation.GetNodeByName(followUpNode)
							g.OpenDialogueNode(conversation, nextNode, isTerminal)
							return
						})
					})
				default:
					g.ApplyEffect(name, args)
				}
			}
		}
	}

	nodeText := node.NpcText
	var nodeOptions []foundation.MenuItem
	if endConversation {
		nodeOptions = append(nodeOptions, foundation.MenuItem{
			Name:       "<End Conversation>",
			Action:     g.ui.CloseConversation,
			CloseMenus: true,
		})
	} else {
		for _, o := range node.Options {
			option := o
			if option.displayCondition != nil {
				evalResult, err := option.displayCondition.Evaluate(nil)
				asBool := evalResult.(bool)
				if err != nil || !asBool {
					continue
				}
			}
			nodeOptions = append(nodeOptions, foundation.MenuItem{
				Name: option.playerText,
				Action: func() {
					nextNode := conversation.GetNextNode(option)
					g.OpenDialogueNode(conversation, nextNode, isTerminal)
				},
				CloseMenus: true,
			})
		}
	}

	g.ui.SetConversationState(nodeText, nodeOptions, isTerminal)
	for _, effectCall := range effectCalls {
		effectCall()
	}
}

func (g *GameState) ApplyEffect(name string, args []string) {
	switch name {
	case "SetFlag":
		flagName := strings.Trim(args[0], "'\" ")
		g.gameFlags.SetFlag(flagName)
	case "ClearFlag":
		flagName := strings.Trim(args[0], "'\" ")
		g.gameFlags.ClearFlag(flagName)
	}
	return
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
				g.GotoNamedLevel("v84_cave", "vault_84")
			},
			CloseMenus: true,
		},
		{
			Name: "Test Lockpick (VeryEasy)",
			Action: func() {
				g.ui.StartLockpickGame(foundation.VeryEasy, g.Player.GetInventory().GetLockpickCount, g.Player.GetInventory().RemoveLockpick, func(result foundation.InteractionResult) {
					g.msg(foundation.HiLite("Lockpick result: %s", result.String()))
				})
			},
		},
		{
			Name: "Test Lockpick (Medium)",
			Action: func() {
				g.ui.StartLockpickGame(foundation.Medium, g.Player.GetInventory().GetLockpickCount, g.Player.GetInventory().RemoveLockpick, func(result foundation.InteractionResult) {
					g.msg(foundation.HiLite("Lockpick result: %s", result.String()))
				})
			},
		},
		{
			Name: "Test Lockpick (Very Hard)",
			Action: func() {
				g.ui.StartLockpickGame(foundation.VeryHard, g.Player.GetInventory().GetLockpickCount, g.Player.GetInventory().RemoveLockpick, func(result foundation.InteractionResult) {
					g.msg(foundation.HiLite("Lockpick result: %s", result.String()))
				})
			},
		},
		{
			Name: "250 Skill Points",
			Action: func() {
				g.Player.GetCharSheet().AddSkillPoints(250)
			},
		},
		{
			Name: "Show Flags",
			Action: func() {
				g.ui.OpenTextWindow(g.gameFlags.ToStringArray())
			},
		},
		{
			Name: "Curse all Equipment",
			Action: func() {
				inv := g.Player.GetInventory()
				for _, i := range inv.Items() {
					if i.IsEquippable() {
						g.AddCurseToEquippable(i)
					}
				}
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

func (g *GameState) StartDialogue(name string, isTerminal bool) {
	conversationFilename := path.Join(g.config.DataRootDir, "dialogues", name+".txt")
	if !util.FileExists(conversationFilename) {
		return
	}
	conversation, err := g.ParseConversation(conversationFilename)
	if err != nil {
		panic(err)
		return
	}
	rootNode := conversation.GetRootNode()
	g.OpenDialogueNode(conversation, rootNode, isTerminal)
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
		gameFlags:           make(util.StringFlags),
		dataDefinitions:     GetDataDefinitions(config.DataRootDir),
		playerFoV:           geometry.NewFOV(geometry.NewRect(0, 0, config.MapWidth, config.MapHeight)),
		visionRange:         14,
		terminalGuesses:     make(map[string][]string),
	}
	g.init()
	ui.SetGame(g)

	return g
}
func (g *GameState) giveAndTryEquipItem(actor *Actor, item *Item) {
	actor.GetInventory().Add(item)
	if item.IsEquippable() {
		actor.GetEquipment().Equip(item)
	}
}
func (g *GameState) init() {
	g.mapLoader = NewTextMapLoader(g)
	playerSheet := special.NewCharSheet()
	playerSheet.SetSkill(special.SmallGuns, 50)
	g.Player = NewPlayer(g.playerName, g.playerIcon, g.playerColor, playerSheet)

	gearDataFile := path.Join(g.config.DataRootDir, "definitions", "start_gear.rec")
	if util.FileExists(gearDataFile) {
		startGear := recfile.Read(util.MustOpen(gearDataFile))[0]
		for _, fields := range startGear {
			parts := fields.AsList(",")
			amount := parts[0].AsInt()
			itemName := parts[1].Value
			for i := 0; i < amount; i++ {
				g.giveAndTryEquipItem(g.Player, g.NewItemFromName(itemName))
			}
		}
	}

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

	g.TurnsTaken = 0
	g.logBuffer = []foundation.HiLiteString{}
	g.currentDungeonLevel = 0
	g.deepestDungeonLevelPlayerReached = 0
	g.showEverything = false
}

func (g *GameState) NewGold(amount int) *Item {
	def := ItemDef{
		Name:         "gold",
		InternalName: "gold",
		Category:     foundation.ItemCategoryGold,
		Charges:      dice_curve.NewDice(1, 1, 1),
	}
	item := NewItem(def)
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
	flagSet := g.Player.GetFlags().UnderlyingCopy()
	equipFlags := g.Player.GetEquipment().GetAllFlags()
	for flag, _ := range equipFlags {
		flagSet[flag] = 1
	}
	return flagSet
}

func (g *GameState) GetHudStats() map[foundation.HudValue]int {
	uiStats := make(map[foundation.HudValue]int)
	//g.Player.stats

	uiStats[foundation.HudTurnsTaken] = g.TurnsTaken
	uiStats[foundation.HudDungeonLevel] = g.currentDungeonLevel
	uiStats[foundation.HudGold] = g.Player.GetGold()

	uiStats[foundation.HudHitPoints] = g.Player.GetHitPoints()
	uiStats[foundation.HudHitPointsMax] = g.Player.GetHitPointsMax()

	uiStats[foundation.HudFatiguePoints] = g.Player.GetCharSheet().GetActionPoints()
	uiStats[foundation.HudFatiguePointsMax] = g.Player.GetCharSheet().GetActionPointsMax()

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
	if g.gridMap.IsDownedActorAt(pos) && g.Player.Position() != pos {
		actor := g.gridMap.DownedActorAt(pos)
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
		g.startThrowItem(item)
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

	if mapCell.DownedActor != nil {
		actor := *mapCell.DownedActor
		if actor.IsDrawn(g.Player.HasFlag(foundation.FlagSeeInvisible)) {
			return foundation.EntityTypeDownedActor
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
	if item.IsMissile() {
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
		if g.canPlayerSee(actor.Position()) && g.couldPlayerSeeActor(actor) {
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

func (g *GameState) NewEnemyFromDef(def ActorDef) *Actor {
	charSheet := special.NewCharSheet()

	for stat, statValue := range def.SpecialStats {
		charSheet.SetStat(stat, statValue)
	}

	for derivedStat, derivedStatValue := range def.DerivedStat {
		charSheet.SetDerivedStatAbsoluteValue(derivedStat, derivedStatValue)
	}

	for skill, skillValue := range def.Skills {
		charSheet.SetSkillAbsoluteValue(skill, skillValue)
	}

	charSheet.HealCompletely()

	actor := NewActor(def.Name, def.Icon, def.Color, charSheet)
	actor.GetFlags().Init(def.Flags.UnderlyingCopy())
	actor.SetIntrinsicZapEffects(def.ZapEffects)
	actor.SetIntrinsicUseEffects(def.UseEffects)
	actor.SetInternalName(def.InternalName)

	actor.SetSizeModifier(def.SizeModifier)
	actor.SetRelationToPlayer(def.DefaultRelation)
	for _, itemName := range def.Equipment {
		item := g.NewItemFromName(itemName)
		actor.GetInventory().Add(item)
	}

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

	//g.dropInventory(victim)
	g.gridMap.SetActorToDowned(victim)
}

func (g *GameState) revealAll() {
	g.gridMap.SetAllExplored()
	g.showEverything = true
}

func (g *GameState) isInPlayerRoom(position geometry.Point) bool {
	if g.dungeonLayout == nil {
		return true
	}
	playerRoom := g.getPlayerRoom()
	if playerRoom == nil {
		return false
	}
	return playerRoom.ContainsIncludingWalls(position)
}

func (g *GameState) spawnEntities(random *rand.Rand, level int, newMap *gridmap.GridMap[*Actor, *Item, Object], dungeon *dungen.DungeonMap) {

	playerRoom := dungeon.GetRoomAt(g.Player.Position())

	mustSpawnAmuletOfYendor := level == 26 && !g.Player.GetInventory().HasItemWithName("amulet_of_yendor")

	spawnItemsInRoom := func(room *dungen.DungeonRoom, itemCount int) {
		for i := 0; i < itemCount; i++ {
			spawnPos, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			itemDef := g.dataDefinitions.PickItemForLevel(random, level)
			item := NewItem(itemDef)
			if item.IsEquippable() && random.Intn(5) == 0 {
				g.AddCurseToEquippable(item)
			}
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
			_, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			/*
				trapEffects := foundation.GetAllTrapCategories()
				randomEffect := trapEffects[random.Intn(len(trapEffects))]
				object := g.NewTrap(randomEffect)
			*/

			//newMap.AddObject(object, spawnPos)
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

		if level > 1 && room != playerRoom && random.Intn(4) < 4 {
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
				newItem := NewItem(itemDef)
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

func (g *GameState) spawnCrawlerInWall(monsterDef ActorDef) {
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
		result := CheckStrength(g.Player)

		if result.Success {
			g.msg(foundation.Msg("You shake off the stun"))
			g.Player.GetFlags().Unset(foundation.FlagStun)
			return
		}
		g.Player.GetFlags().Increment(foundation.FlagStun)

		g.msg(foundation.Msg("You are stunned and cannot act"))

		// TODO: animate a small delay here?
		g.endPlayerTurn(g.Player.timeNeededForActions())
	}
	if g.Player.HasFlag(foundation.FlagHeld) {
		result := CheckStrength(g.Player)

		if result.Crit {
			g.msg(foundation.Msg("You break free from the hold"))
			g.Player.GetFlags().Unset(foundation.FlagHeld)
			return
		} else if result.Success {
			g.Player.GetFlags().Decrease(foundation.FlagHeld, 10)
			if !g.Player.HasFlag(foundation.FlagHeld) {
				g.msg(foundation.Msg("You break free from the hold"))
				return
			}
		}

		g.msg(foundation.Msg("You are held and cannot act"))

		// TODO: animate a small delay here?
		g.endPlayerTurn(g.Player.timeNeededForActions())
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
		triggeredEffectAnimations := objectAt.OnWalkOver(actor)
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
	//g.Player.AddCharacterPoints(10)
	g.msg(foundation.HiLite("You've been awarded 10 character points for reaching level %s", fmt.Sprint(level)))
}

func (g *GameState) checkTilesForHiddenObjects(tiles []geometry.Point) {
	var noticedSomething bool
	for _, tile := range tiles {
		if g.gridMap.IsObjectAt(tile) {
			object := g.gridMap.ObjectAt(tile)
			if object.IsHidden() {
				perceptionResult := CheckPerception(g.Player)
				if perceptionResult.Success {
					noticedSomething = true
				}
				if perceptionResult.Crit {
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
	/*
		if actor.GetActionPoints() < 1 {
			g.msg(foundation.Msg("You are too tired to sprint"))
			return
		}
		actor.LooseActionPoints(1)

	*/
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

func (g *GameState) AddCurseToEquippable(item *Item) {
	// item doesn't use equip flag? add a negative one
	// else -> item doesn't use stats? add a negative one
	if item.IsMissile() { // don't curse missiles
		return
	}
	if item.equipFlag == foundation.FlagNone && (item.charges == 0 || item.charges == 1) {
		item.equipFlag = foundation.FlagCurseStuck
		item.charges = rand.Intn(300) + 100
	}
	if item.statBonus == 0 {
		item.stat = dice_curve.GetRandomStat()
		item.statBonus = -(rand.Intn(4) + 1)
	}
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
