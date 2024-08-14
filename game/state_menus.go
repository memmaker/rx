package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"math/rand"
	"path"
	"strconv"
	"time"
)

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

// TODO: ContextMenu Actions don't take up any time!
func (g *GameState) OpenContextMenuFor(mapPos geometry.Point) {
	if g.gridMap.MoveDistance(g.Player.Position(), mapPos) > 1 {
		return
	}
	var menuItems []foundation.MenuItem
	if g.gridMap.IsActorAt(mapPos) {
		actor := g.gridMap.ActorAt(mapPos)
		if actor == g.Player {
			// TODO: Self actions..
		} else {
			menuItems = g.appendContextActionsForActor(menuItems, actor)
		}
	}
	if g.gridMap.IsObjectAt(mapPos) {
		object := g.gridMap.ObjectAt(mapPos)
		menuItems = object.AppendContextActions(menuItems, g)
	}
	if len(menuItems) == 0 {
		return
	}
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
			Name: "Test Keypad",
			Action: func() {
				g.ui.OpenKeypad([]rune{'1', '2', '3', '4'}, func(sequence bool) {
					g.msg(foundation.HiLite("Keypad result: %s", strconv.FormatBool(sequence)))
				})
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

func (g *GameState) StartDialogue(name string, conversationPartnerName string, isTerminal bool) {
	conversationFilename := path.Join(g.config.DataRootDir, "dialogues", name+".txt")
	if !fxtools.FileExists(conversationFilename) {
		return
	}
	conversation, err := g.ParseConversation(conversationFilename)
	if err != nil {
		panic(err)
		return
	}
	rootNode := conversation.GetRootNode()
	g.OpenDialogueNode(conversation, rootNode, conversationPartnerName, isTerminal)
}

func (g *GameState) OpenDialogueNode(conversation *Conversation, node ConversationNode, conversationPartnerName string, isTerminal bool) {
	endConversation := false
	var effectCalls []func()
	for _, effect := range node.Effects {
		if effect == "EndConversation" {
			endConversation = true
		} else {
			if fxtools.LooksLikeAFunction(effect) {
				name, args := fxtools.GetNameAndArgs(effect)
				switch name {
				case "RemoveItem":
					itemName := args.Get(0)
					removedItem := g.Player.GetInventory().RemoveItemByName(itemName)
					if removedItem != nil {
						g.msg(foundation.HiLite("%s removed.", removedItem.Name()))
					}
				case "Hacking":
					terminalID := args.Get(0)
					difficulty := args.Get(1)
					flagName := args.Get(2)
					followUpNode := args.Get(3)

					effectCalls = append(effectCalls, func() {
						previousGuesses := g.terminalGuesses[terminalID]
						g.ui.StartHackingGame(fxtools.MurmurHash(flagName), foundation.DifficultyFromString(difficulty), previousGuesses, func(pGuesses []string, result foundation.InteractionResult) {
							g.terminalGuesses[terminalID] = pGuesses
							if result == foundation.Success {
								g.gameFlags.SetFlag(flagName)
							}
							nextNode := conversation.GetNodeByName(followUpNode)
							g.OpenDialogueNode(conversation, nextNode, conversationPartnerName, isTerminal)
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
			if option.CanDisplay() {
				nodeOptions = append(nodeOptions, foundation.MenuItem{
					Name: option.playerText,
					Action: func() {
						nextNode := conversation.GetNextNode(option)
						g.OpenDialogueNode(conversation, nextNode, conversationPartnerName, isTerminal)
					},
					CloseMenus: true,
				})
			}
		}
	}

	g.ui.SetConversationState(nodeText, nodeOptions, conversationPartnerName, isTerminal)
	for _, effectCall := range effectCalls {
		effectCall()
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
			Name: itemDef.Description,
			Action: func() {
				newItem := NewItem(itemDef, g.iconForItem(itemDef.Category))
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
			Name: monsterDef.Description,
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
				newTrap := g.NewTrap(trapType, g.iconForObject)
				newTrap.SetHidden(false)
				g.gridMap.AddObject(newTrap, trapPos)
			},
			CloseMenus: true,
		})
	}
	g.ui.OpenMenu(menuActions)
}
