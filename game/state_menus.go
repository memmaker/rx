package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func (g *GameState) OpenInventory() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return !item.IsAmmo()
	})
	if len(inventory) == 0 {
		g.ui.OpenTextWindow("You are not carrying anything.")
		return
	}
	g.ui.OpenInventoryForManagement(inventory)
}
func (g *GameState) ChooseItemForThrow() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsThrowable()
	})
	if len(inventory) == 0 {
		g.ui.OpenTextWindow("You are not carrying anything throwable.")
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Throw what?", func(itemStack foundation.Item) {
		g.startThrowItem(itemStack)
	})
}

// TODO: ContextMenu Actions don't take up any time!
func (g *GameState) OpenContextMenuFor(mapPos geometry.Point) bool {
	var menuItems []foundation.MenuItem
	distance := g.currentMap().MoveDistance(g.Player.Position(), mapPos)
	if distance > 1 {
		if g.canPlayerSee(mapPos) && g.currentMap().IsActorAt(mapPos) {
			actorAt := g.currentMap().ActorAt(mapPos)
			menuItems = g.appendContextActionsForActor(menuItems, actorAt)
			if len(menuItems) > 0 {
				g.ui.OpenMenuWithTitle(actorAt.Name(), menuItems)
				return true
			}
		}
		return false
	}
	if g.currentMap().IsActorAt(mapPos) {
		actor := g.currentMap().ActorAt(mapPos)
		if actor == g.Player {
			// TODO: Self actions..
		} else {
			menuItems = g.appendContextActionsForActor(menuItems, actor)
			if len(menuItems) > 0 {
				g.ui.OpenMenuWithTitle(actor.Name(), menuItems)
				return true
			}
		}
	}
	if g.currentMap().IsObjectAt(mapPos) {
		object := g.currentMap().ObjectAt(mapPos)
		menuItems = object.AppendContextActions(menuItems, g)
	}
	if len(menuItems) == 0 {
		return false
	}
	g.ui.OpenMenu(menuItems)
	return true
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

func (g *GameState) PlayerRest(isHealing bool, duration time.Duration) {
	g.ui.FadeToBlack()
	g.advanceTime(duration)
	if g.Player.HasWatch() {
		g.printTime()
	}
	g.ui.FadeFromBlack()

	if isHealing && g.Player.GetCharSheet().NeedsHealing() {
		hours := int(duration.Hours())
		healingRate := g.Player.GetCharSheet().GetDerivedStat(d100.HealingRate)
		healedPoints := hours * healingRate
		g.Player.Heal(healedPoints)
		g.ui.UpdateStats()
		g.msg(foundation.HiLite("You have recovered %s hit points.", strconv.Itoa(healedPoints)))
	}
}

func (g *GameState) SaveGame(toDirectory string) {
	if toDirectory != "" {
		err := g.Save(toDirectory)
		if err != nil {
			panic(err)
		} else {
			g.msg(foundation.Msg("Game saved."))
			if g.IsIronMan() {
				g.ui.QuitGame()
			}
		}
	}
}

func (g *GameState) LoadGame(fromDirectory string) {
	if fromDirectory != "" {
		g.Load(fromDirectory)
		if g.IsIronMan() {
			os.RemoveAll(fromDirectory)
		}
		g.msg(foundation.Msg("Game loaded."))
	}
}
func (g *GameState) OpenWaitMenu() {
	g.openRestMenu(false)
}
func (g *GameState) openRestMenu(isHealing bool) {
	waitWord := "Wait"
	if isHealing {
		waitWord = "Rest"
	}
	g.ui.OpenMenu([]foundation.MenuItem{
		{
			Name: fmt.Sprintf("%s for ten minutes", waitWord),
			Action: func() {
				g.PlayerRest(isHealing, 10*time.Minute)
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s for thirty minutes", waitWord),
			Action: func() {
				g.PlayerRest(isHealing, 30*time.Minute)
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s for one hour", waitWord),
			Action: func() {
				g.PlayerRest(isHealing, time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s for two hours", waitWord),
			Action: func() {
				g.PlayerRest(isHealing, 2*time.Hour)
			},

			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s for three hours", waitWord),
			Action: func() {
				g.PlayerRest(isHealing, 3*time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s for four hours", waitWord),
			Action: func() {
				g.PlayerRest(isHealing, 4*time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s for five hours", waitWord),
			Action: func() {
				g.PlayerRest(isHealing, 5*time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s for six hours", waitWord),
			Action: func() {
				g.PlayerRest(isHealing, 6*time.Hour)
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s until morning (0600)", waitWord),
			Action: func() {
				now := g.gameTime.Time
				morning := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, now.Location())
				if !morning.After(now) {
					morning = morning.AddDate(0, 0, 1)
				}
				g.PlayerRest(isHealing, morning.Sub(now))
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s until noon (1200)", waitWord),
			Action: func() {
				now := g.gameTime.Time
				noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
				if !noon.After(now) {
					noon = noon.AddDate(0, 0, 1)
				}
				g.PlayerRest(isHealing, noon.Sub(now))
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s until evening (1800)", waitWord),
			Action: func() {
				now := g.gameTime.Time
				evening := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
				if !evening.After(now) {
					evening = evening.AddDate(0, 0, 1)
				}
				g.PlayerRest(isHealing, evening.Sub(now))
			},
			CloseMenus: true,
		},
		{
			Name: fmt.Sprintf("%s until midnight (0000)", waitWord),
			Action: func() {
				now := g.gameTime.Time
				midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				if !midnight.After(now) {
					midnight = midnight.AddDate(0, 0, 1)
				}
				g.PlayerRest(isHealing, midnight.Sub(now))
			},
			CloseMenus: true,
		},
	})
}
func (g *GameState) OpenWizardMenu() {
	g.ui.OpenMenu([]foundation.MenuItem{
		{
			Name:   "Show Map",
			Action: g.revealAll,
		},
		{
			Name: "Load Test Map",
			Action: func() {
				g.GotoNamedLevel("v84_cave", "vault_84")
			},
			CloseMenus: true,
		},
		{
			Name: "10 Skill Points",
			Action: func() {
				g.Player.GetCharSheet().AddSkillPoints(10)
			},
		},
		{
			Name:       "Test Team Spawn",
			Action:     g.TestTeamSpawn,
			CloseMenus: true,
		},
		{
			Name: "All the lockpicks",
			Action: func() {
				for i := 0; i < 200; i++ {
					g.Player.GetInventory().AddItem(g.newItemFromName("mechanical_lockpick"))
					g.Player.GetInventory().AddItem(g.newItemFromName("electronic_lockpick"))
				}
			},
		},
		{
			Name: "Filthy Rich",
			Action: func() {
				g.Player.GetInventory().AddItem(g.NewGold(1000000))
			},
		},
		{
			Name: "God Like",
			Action: func() {
				for p := CyberWare(1); p < CyberWareCount; p++ {
					g.Player.AddCyberWare(p)
				}
				g.Player.GetCharSheet().AddPerkPoints(int(d100.PerkCount))
				for p := d100.Perk(0); p < d100.PerkCount; p++ {
					g.Player.GetCharSheet().AddPerk(p)
				}
				g.Player.GetCharSheet().SetGodLike()
				g.Player.GetCharSheet().HealAPAndHPCompletely()
				g.ui.UpdateStats()
			},
		},
		{
			Name: "1000 XP",
			Action: func() {
				g.awardXP(1000, "for testing")
			},
		},
		{
			Name: "Add all Perks",
			Action: func() {
				g.Player.GetCharSheet().AddPerkPoints(int(d100.PerkCount))
				for p := d100.Perk(0); p < d100.PerkCount; p++ {
					g.Player.GetCharSheet().AddPerk(p)
				}
			},
		},
		{
			Name: "Add all cyberware",
			Action: func() {
				for p := CyberWare(1); p < CyberWareCount; p++ {
					g.Player.AddCyberWare(p)
				}
			},
		},
		{
			Name: "Show Flags",
			Action: func() {
				g.ui.OpenTextWindow(g.gameFlags.String())
			},
		},
		{
			Name: "Show TimeTracker",
			Action: func() {
				g.ui.OpenTextWindow(g.timeTracker.String())
			},
		},
		{
			Name: "Show Scripts",
			Action: func() {
				g.ui.OpenTextWindow(g.scriptRunner.String())
			},
		},
		{
			Name: "Show Metronome",
			Action: func() {
				g.ui.OpenTextWindow(g.metronome.String())
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
	})
}

func (g *GameState) RunScriptByName(scriptName string) {
	if fxtools.LooksLikeAFunction(scriptName) {
		name, args := fxtools.GetNameAndArgs(scriptName)
		switch name {
		case "LeaveMapAt":
			actorName := args.Get(0)
			locationName := args.Get(1)
			running := false
			if len(args) > 2 {
				running = args.GetBool(2)
			}
			actor := g.actorWithName(actorName)
			leaveMapAtLocation := g.NewScriptLeaveMapAtLocation(actor, running, locationName)
			g.RunScript(leaveMapAtLocation)
		}
	} else {
		mapName := g.currentMap().GetName()

		g.RunScriptOnMap(mapName, scriptName)
	}
}

func (g *GameState) RunScriptOnMap(mapName string, scriptName string) {
	pathToMaps := path.Join(g.config.DataRootDir, "maps")
	g.ExecuteOnMap(mapName, func() {
		g.scriptRunner.RunScriptByName(pathToMaps, mapName, scriptName, g.GetScriptFuncs())
	})
}

func (g *GameState) RunScript(script ActionScript) {
	mapName := g.currentMap().GetName()
	g.scriptRunner.RunScript(mapName, script)
}

func (g *GameState) StartDialogue(name string, partner foundation.ChatterSource, isTerminal bool) {
	conversationFilename := path.Join(g.config.DataRootDir, "dialogues", name+".rec")
	if !fxtools.FileExists(conversationFilename) {
		g.msg(foundation.HiLite("%s has nothing to say.", partner.Name()))
		return
	}
	conversation, err := ParseConversation(conversationFilename, g.GetScriptFuncs())
	if err != nil {
		panic(err)
		return
	}

	var npcName string
	params := make(map[string]interface{})
	if actor, isActor := partner.(*Actor); isActor {
		npcName = actor.GetInternalName()
		talkedFlagName := fmt.Sprintf("TalkedTo(%s)", npcName)
		g.gameFlags.Increment(talkedFlagName)
		params["NPC"] = actor
	} else {
		npcName = partner.Name()
	}
	params["NPC_NAME"] = npcName

	rootNode := conversation.GetRootNode(params)
	g.OpenDialogueNode(conversation, ConversationNode{}, rootNode, partner, isTerminal)
}

func (g *GameState) OpenDialogueNode(conversation *Conversation, prevNode ConversationNode, currentNode ConversationNode, conversationPartner foundation.ChatterSource, isTerminal bool) {
	endConversation := false
	instantEndWithChatter := false

	nodeText := g.fillTemplatedText(currentNode.NpcText)

	var nodeOptions []foundation.MenuItem
	var effectCalls []func()
	for _, effect := range currentNode.Effects {
		if effect == "StartCombat" { // simple parameterless dialogue only effects
			if actor, isActor := conversationPartner.(*Actor); isActor {
				actor.FSM.SendEvent(NewProvokedEvent(g.Player))
			}
			instantEndWithChatter = true
		} else if effect == "EndHostility" {
			if actor, isActor := conversationPartner.(*Actor); isActor {
				actor.FSM.SendEvent(NewCalmedEvent(g.Player))
			}
		} else if effect == "EndWithChatter" {
			instantEndWithChatter = true
		} else if effect == "EndConversation" {
			endConversation = true
		} else if effect == "ReturnToPreviousNode" {
			if !prevNode.IsEmpty() {
				currentNode = prevNode
			}
		} else {
			if fxtools.LooksLikeAFunction(effect) {
				name, args := fxtools.GetNameAndArgs(effect)
				switch name { // these effects are also only possible here, because they directly influence the conversation flow
				case "GotoNode":
					nodeName := args.Get(0)
					nextNode := conversation.GetNodeByName(nodeName)
					if !nextNode.IsEmpty() {
						currentNode = nextNode
					}
				case "TransitionWithDriver":
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
				case "TakeItemFromPlayer":
					itemName := args.Get(0)
					count := 1
					if len(args) > 1 {
						count = args.GetInt(1)
					}
					actor, isActor := conversationPartner.(*Actor)
					itemsForPartner := g.Player.GetInventory().RemoveItemsByNameAndCount(itemName, count)
					if len(itemsForPartner) > 0 {
						first := itemsForPartner[0]
						niceItemName := first.Name()
						g.msg(foundation.HiLite("%s removed.", niceItemName))
						if isActor {
							actor.GetInventory().AddItems(itemsForPartner)
						}
					}
				case "GiveItemToPlayer":
					itemName := args.Get(0)
					count := 1
					if len(args) > 1 {
						count = args.GetInt(1)
					}
					actor, isActor := conversationPartner.(*Actor)
					var itemsForPlayer []foundation.Item
					if isActor {
						itemsForPlayer = actor.GetInventory().RemoveItemsByNameAndCount(itemName, count)
					} else {
						oneItem := g.newItemFromName(itemName)
						if oneItem.IsStackable() {
							oneItem.SetCharges(count)
						} else {
							for i := 0; i < count; i++ {
								itemsForPlayer = append(itemsForPlayer, oneItem)
								oneItem = g.newItemFromName(itemName)
							}
						}
					}
					if len(itemsForPlayer) > 0 {
						first := itemsForPlayer[0]
						itemStackName := first.Name()
						g.Player.GetInventory().AddItems(itemsForPlayer)
						g.msg(foundation.HiLite("%s received.", itemStackName))
					}

				default: // parse as generic expression and effect
					expr, parseErr := govaluate.NewEvaluableExpressionWithFunctions(effect, g.GetScriptFuncs())
					if parseErr != nil {
						panic(parseErr)
					}
					_, evalErr := expr.Evaluate(conversation.Variables)
					if evalErr != nil {
						panic(evalErr)
					}
				}
			}
		}
	}

	if otherActor, isActor := conversationPartner.(*Actor); isActor && instantEndWithChatter {
		g.ui.CloseConversation()
		g.tryAddChatter(otherActor, nodeText)
		return
	}

	if endConversation {
		nodeOptions = append(nodeOptions, foundation.MenuItem{
			Name:       "<Leave>",
			Action:     g.ui.CloseConversation,
			CloseMenus: true,
		})
	} else {
		for _, o := range currentNode.Options {
			option := o
			if option.CanDisplay(conversation.Variables) {
				onChoice := func() {
					nextNode := conversation.GetNextNode(option)
					g.OpenDialogueNode(conversation, currentNode, nextNode, conversationPartner, isTerminal)
				}

				if len(option.effects) > 0 {
					onChoice = func() {
						for _, effect := range option.effects {
							switch effect {
							case "StartTrading":
								followUp := func() { g.ui.SetConversationState(nodeText, nodeOptions, conversationPartner, isTerminal) }
								g.openVendorMenu(conversationPartner.(*Actor), followUp)
							case "StartRepair":
								followUp := func() { g.ui.SetConversationState(nodeText, nodeOptions, conversationPartner, isTerminal) }
								g.openNPCRepairMenu(conversationPartner.(*Actor), followUp)
							case "StartCyberware":
								followUp := func() { g.ui.SetConversationState(nodeText, nodeOptions, conversationPartner, isTerminal) }
								g.openCyberwareMenu(conversationPartner.(*Actor), followUp)

							}
						}
					}
				}
				nodeOptions = append(nodeOptions, foundation.MenuItem{
					Name:       g.fillTemplatedText(option.playerText) + option.RollInfo(),
					Action:     onChoice,
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

func (g *GameState) OpenJournal() {
	entries := g.journal.GetEntriesForViewing("default")
	g.ui.OpenTextWindow(strings.Join(entries, "\n\n"))
}
