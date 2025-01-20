package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/geometry"
	"os"
	"strconv"
	"strings"
	"time"
)

func (g *GameState) GetNonAmmoPlayerInventory() []foundation.Item {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return !item.IsAmmo()
	})
	return inventory
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
		if g.Player.CanSee(mapPos) && g.currentMap().IsActorAt(mapPos) {
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
func (g *GameState) PlayerRest(isHealing bool, duration time.Duration) {
	g.ui.FadeToBlack()
	g.advanceTime(duration)
	if g.Player.HasWatch() {
		g.ShowDateTime()
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
				g.transitionToMapLocation("v84_cave", "vault_84")
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
			Name:       "Teleport",
			Action:     g.OpenTeleportMenu,
			CloseMenus: true,
		},
		{
			Name:       "Test Team Spawn",
			Action:     g.TestTeamSpawn,
			CloseMenus: true,
		},
		{
			Name:       "Test Forced Monologue",
			Action:     g.TestForcedMonologue,
			CloseMenus: true,
		},
		{
			Name: "Test Pathfinder",
			Action: func() {
				pf := NewPathfinder(g.ensureMapIsLoaded)
				mapPath := pf.FindPath(g.Player, g.currentMapName, MapPosition{MapName: "zone_corporate", LocationName: "street_to_ebi", Position: g.ensureMapIsLoaded("zone_corporate").GetNamedLocation("street_to_ebi")})
				g.msg(foundation.Msg(fmt.Sprintf("Path: %v", mapPath)))
			},
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
				g.ui.OpenTextWindow(g.Scripts.String())
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
func (g *GameState) OpenJournal() {
	entries := g.journal.GetEntriesForViewing("default")
	g.ui.OpenTextWindow(strings.Join(entries, "\n\n"))
}
