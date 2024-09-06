package game

import (
	"RogueUI/foundation"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"math/rand"
	"path"
	"strconv"
	"strings"
	"time"
)

func (g *GameState) OpenInventory() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return !item.IsAmmo()
	})
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
	if g.currentMap().MoveDistance(g.Player.Position(), mapPos) > 1 {
		return
	}
	var menuItems []foundation.MenuItem
	if g.currentMap().IsActorAt(mapPos) {
		actor := g.currentMap().ActorAt(mapPos)
		if actor == g.Player {
			// TODO: Self actions..
		} else {
			menuItems = g.appendContextActionsForActor(menuItems, actor)
		}
	}
	if g.currentMap().IsObjectAt(mapPos) {
		object := g.currentMap().ObjectAt(mapPos)
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

func (g *GameState) PlayerRest(duration time.Duration) {
	g.ui.FadeToBlack()
	g.gameTime = g.gameTime.Add(duration)
	g.msg(foundation.Msg(fmt.Sprintf("Time is now %s", g.gameTime.Format("15:04"))))
	g.ui.FadeFromBlack()
}

func (g *GameState) OpenRestMenu() {
	g.ui.OpenMenu([]foundation.MenuItem{
		{
			Name: "Rest for ten minutes",
			Action: func() {
				g.PlayerRest(10 * time.Minute)
			},
			CloseMenus: true,
		},
		{
			Name: "Rest for thirty minutes",
			Action: func() {
				g.PlayerRest(30 * time.Minute)
			},
			CloseMenus: true,
		},
		{
			Name: "Rest for an hour",
			Action: func() {
				g.PlayerRest(time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: "Rest for two hours",
			Action: func() {
				g.PlayerRest(2 * time.Hour)
			},

			CloseMenus: true,
		},
		{
			Name: "Rest for three hours",
			Action: func() {
				g.PlayerRest(3 * time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: "Rest for four hours",
			Action: func() {
				g.PlayerRest(4 * time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: "Rest for five hours",
			Action: func() {
				g.PlayerRest(5 * time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: "Rest for six hours",
			Action: func() {
				g.PlayerRest(6 * time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: "Rest until morning (0600)",
			Action: func() {
				now := g.gameTime
				morning := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, now.Location())
				if now.After(morning) {
					morning = morning.AddDate(0, 0, 1)
				}
				g.PlayerRest(morning.Sub(now))
			},
			CloseMenus: true,
		},
		{
			Name: "Rest until noon (1200)",
			Action: func() {
				now := g.gameTime
				noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
				if now.After(noon) {
					noon = noon.AddDate(0, 0, 1)
				}
				g.PlayerRest(noon.Sub(now))
			},
			CloseMenus: true,
		},
		{
			Name: "Rest until evening (1800)",
			Action: func() {
				now := g.gameTime
				evening := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
				if now.After(evening) {
					evening = evening.AddDate(0, 0, 1)
				}
				g.PlayerRest(evening.Sub(now))
			},
			CloseMenus: true,
		},
		{
			Name: "Rest until midnight (0000)",
			Action: func() {
				now := g.gameTime
				midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				if now.After(midnight) {
					midnight = midnight.AddDate(0, 0, 1)
				}
				g.PlayerRest(midnight.Sub(now))
			},
			CloseMenus: true,
		},
		{
			Name: "Rest until healed",
			Action: func() {
				g.PlayerRest(time.Hour * 48) // TODO: change this?
			},
			CloseMenus: true,
		},
	})
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
			Name: "10 Skill Points",
			Action: func() {
				g.Player.GetCharSheet().AddSkillPoints(10)
			},
		},
		{
			Name: "1000 XP",
			Action: func() {
				g.awardXP(1000, "for testing")
			},
		},
		{
			Name: "Show Flags",
			Action: func() {
				g.ui.OpenTextWindow(g.gameFlags.String())
			},
		},
		{
			Name: "Set Flag",
			Action: func() {
				g.ui.AskForString("Flag name", "", func(flagName string) {
					g.gameFlags.SetFlag(flagName)
				})
			},
		},
		{
			Name:   "Create Trap",
			Action: g.openWizardCreateTrapMenu,
		},
		{
			Name: "Game Save",
			Action: func() {
				err := g.Save("savegame")
				if err != nil {
					panic(err)
				}
			},
		},
		{
			Name: "Game Load",
			Action: func() {
				g.Load("savegame")
			},
		},
		{
			Name: "Run Test Script",
			Action: func() {
				g.RunScript("jeff_kills_winters")
			},
		},
	})
}

func (g *GameState) RunScript(scriptName string) {
	g.scriptRunner.RunScript(g.config.DataRootDir, scriptName, g.getScriptFuncs())
}

func (g *GameState) StartDialogue(name string, partner foundation.ChatterSource, isTerminal bool) {
	conversationFilename := path.Join(g.config.DataRootDir, "dialogues", name+".txt")
	if !fxtools.FileExists(conversationFilename) {
		return
	}
	conversation, err := ParseConversation(conversationFilename, g.getConditionFuncs())
	if err != nil {
		panic(err)
		return
	}

	var npcName string
	if actor, isActor := partner.(*Actor); isActor {
		npcName = actor.GetInternalName()
		talkedFlagName := fmt.Sprintf("TalkedTo(%s)", npcName)
		g.gameFlags.Increment(talkedFlagName)
	} else {
		npcName = partner.Name()
	}

	params := map[string]interface{}{
		"NPC_NAME": npcName,
	}
	rootNode := conversation.GetRootNode(params)
	g.OpenDialogueNode(conversation, rootNode, partner, isTerminal)
}

func (g *GameState) OpenDialogueNode(conversation *Conversation, node ConversationNode, conversationPartner foundation.ChatterSource, isTerminal bool) {
	endConversation := false
	instantEndWithChatter := false
	var effectCalls []func()
	for _, effect := range node.Effects {
		if effect == "StartCombat" {
			if actor, isActor := conversationPartner.(*Actor); isActor {
				actor.AddToEnemyActors(g.Player.GetInternalName())
				actor.SetHostile()
			}
			instantEndWithChatter = true
		} else if effect == "EndHostility" {
			if actor, isActor := conversationPartner.(*Actor); isActor {
				actor.RemoveEnemy(g.Player)
				actor.SetNeutral()
			}
		} else if effect == "EndConversation" {
			endConversation = true
		} else {
			if fxtools.LooksLikeAFunction(effect) {
				name, args := fxtools.GetNameAndArgs(effect)
				switch name {
				case "AdvanceTimeByMinutes":
					minutes, err := strconv.Atoi(args.Get(0))
					if err != nil {
						panic(err)
					}
					g.advanceTime(time.Minute * time.Duration(minutes))
				case "RunScript":
					scriptName := args.Get(0)
					g.RunScript(scriptName)
				case "DriverTransition":
					mapName := args.Get(0)
					locationName := args.Get(1)
					g.ui.FadeToBlack()

					var taxiDriver *Actor
					if actor, isActor := conversationPartner.(*Actor); isActor {
						taxiDriver = actor
						g.currentMap().RemoveActor(taxiDriver)
					}

					g.GotoNamedLevel(mapName, locationName)

					tdLoc := g.currentMap().GetNamedLocation("taxi_driver")
					g.currentMap().AddActor(taxiDriver, tdLoc)

					g.ui.FadeFromBlack()
					instantEndWithChatter = true
				case "Transition":
					mapName := args.Get(0)
					locationName := args.Get(1)
					g.ui.FadeToBlack()
					g.GotoNamedLevel(mapName, locationName)
					g.ui.FadeFromBlack()
					instantEndWithChatter = true
				case "RemoveItem":
					itemName := args.Get(0)
					removedItem := g.Player.GetInventory().RemoveItemByName(itemName)
					if removedItem != nil {
						g.msg(foundation.HiLite("%s removed.", removedItem.Name()))
					}
				case "StopScript":
					scriptName := args.Get(0)
					g.scriptRunner.StopScript(scriptName)
				case "GiveItem":
					itemName := args.Get(0)
					actor, isActor := conversationPartner.(*Actor)
					var itemForPlayer *Item
					if isActor {
						itemForPlayer = actor.GetInventory().GetItemByName(itemName)
						if itemForPlayer != nil {
							actor.GetInventory().Remove(itemForPlayer)
						}
					} else {
						itemForPlayer = g.NewItemFromString(itemName)
					}
					if itemForPlayer != nil {
						g.Player.GetInventory().Add(itemForPlayer)
						g.msg(foundation.HiLite("%s received.", itemForPlayer.Name()))
					}
				case "Hacking":
					terminalID := args.Get(0)
					difficulty := args.Get(1)
					flagName := args.Get(2)
					successNode := args.Get(3)
					failNode := args.Get(4)

					effectCalls = append(effectCalls, func() {
						previousGuesses := g.terminalGuesses[terminalID]
						g.ui.StartHackingGame(fxtools.MurmurHash(flagName), foundation.DifficultyFromString(difficulty), previousGuesses, func(pGuesses []string, result foundation.InteractionResult) {
							g.terminalGuesses[terminalID] = pGuesses
							followUpNode := failNode
							if result == foundation.Success {
								g.gameFlags.SetFlag(flagName)
								followUpNode = successNode
							}
							nextNode := conversation.GetNodeByName(followUpNode)
							g.OpenDialogueNode(conversation, nextNode, conversationPartner, isTerminal)
							return
						})
					})
				default:
					g.ApplyEffect(name, args)
				}
			}
		}
	}

	nodeText := g.fillTemplatedText(node.NpcText)

	if otherActor, isActor := conversationPartner.(*Actor); isActor && instantEndWithChatter {
		g.ui.CloseConversation()
		g.tryAddChatter(otherActor, nodeText)
		return
	}

	var nodeOptions []foundation.MenuItem
	if endConversation {
		nodeOptions = append(nodeOptions, foundation.MenuItem{
			Name:       "<Leave>",
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
						g.OpenDialogueNode(conversation, nextNode, conversationPartner, isTerminal)
					},
					CloseMenus: true,
				})
			}
		}
	}

	g.ui.SetConversationState(nodeText, nodeOptions, conversationPartner, isTerminal)
	for _, effectCall := range effectCalls {
		effectCall()
	}
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
				trapPos := g.currentMap().GetRandomFreeAndSafeNeighbor(random, g.Player.Position())
				newTrap := g.NewTrap(trapType)
				newTrap.SetHidden(false)
				g.currentMap().AddObject(newTrap, trapPos)
			},
			CloseMenus: true,
		})
	}
	g.ui.OpenMenu(menuActions)
}

func (g *GameState) OpenJournal() {
	entries := g.journal.GetEntriesForViewing("default")
	g.ui.OpenTextWindow(strings.Join(entries, "\n\n"))
}
