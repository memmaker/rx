package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"math/rand"
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
		g.msg(foundation.Msg("You are not carrying anything."))
		return
	}
	g.ui.OpenInventoryForManagement(inventory)
}
func (g *GameState) ChooseItemForThrow() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsThrowable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything throwable."))
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

func (g *GameState) PlayerRest(duration time.Duration) {
	g.ui.FadeToBlack()
	g.advanceTime(duration)
	if g.Player.GetInventory().HasWatch() {
		g.msg(foundation.Msg(fmt.Sprintf("Time is now %s", g.gameTime.Time.Format("15:04"))))
	}
	g.ui.FadeFromBlack()
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
				now := g.gameTime.Time
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
				now := g.gameTime.Time
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
				now := g.gameTime.Time
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
				now := g.gameTime.Time
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
				g.RunScriptByName("jeff_kills_winters")
			},
		},
	})
}

func (g *GameState) RunScriptByName(scriptName string) {
	g.scriptRunner.RunScriptByName(path.Join(g.config.DataRootDir, "maps"), g.currentMap().GetName(), scriptName, g.getScriptFuncs())
}

func (g *GameState) RunScript(script ActionScript) {
	g.scriptRunner.RunScript(g.currentMap().GetName(), script)
}

func (g *GameState) StartDialogue(name string, partner foundation.ChatterSource, isTerminal bool) {
	conversationFilename := path.Join(g.config.DataRootDir, "dialogues", name+".rec")
	if !fxtools.FileExists(conversationFilename) {
		g.msg(foundation.HiLite("%s has nothing to say.", partner.Name()))
		return
	}
	conversation, err := ParseConversation(conversationFilename, g.getScriptFuncs())
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

	var effectCalls []func()
	for _, effect := range currentNode.Effects {
		if effect == "StartCombat" { // simple parameterless dialogue only effects
			if actor, isActor := conversationPartner.(*Actor); isActor {
				g.trySetHostile(actor, g.Player)
			}
			instantEndWithChatter = true
		} else if effect == "EndHostility" {
			if actor, isActor := conversationPartner.(*Actor); isActor {
				actor.RemoveEnemy(g.Player)
				actor.SetNeutral()
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
					expr, parseErr := govaluate.NewEvaluableExpressionWithFunctions(effect, g.getScriptFuncs())
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

	var nodeOptions []foundation.MenuItem
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
				nodeOptions = append(nodeOptions, foundation.MenuItem{
					Name: g.fillTemplatedText(option.playerText) + option.RollInfo(),
					Action: func() {
						nextNode := conversation.GetNextNode(option)
						g.OpenDialogueNode(conversation, currentNode, nextNode, conversationPartner, isTerminal)
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

func (g *GameState) playerHackingRoll(difficulty foundation.Difficulty) special.CheckResult {
	scienceSkill := g.Player.GetCharSheet().GetSkill(special.Technology)
	luck := 5

	modifier := difficulty.GetRollModifier()
	effectiveSkill := scienceSkill + modifier
	rollResult := special.SuccessRoll(special.Percentage(effectiveSkill), special.Percentage(luck))
	return rollResult
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
