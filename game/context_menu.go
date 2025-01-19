package game

import (
	"contractor/d100"
	"contractor/foundation"
	"contractor/gridmap"
	"fmt"
)

func (g *GameState) animatedActionFromMenu(action func()) func() {
	return func() {
		g.ui.ForceUIRedraw()
		action()
	}
}

func (g *GameState) canPlayerTalkToActor(actor *Actor) bool {
	distance := g.currentMap().MoveDistance(g.Player.Position(), actor.Position())
	return g.Player.CanSee(actor.Position()) && actor.HasDialogue() && !actor.IsSleeping() && distance <= 6
}

func (g *GameState) appendContextActionsForActor(buffer []foundation.MenuItem, actor *Actor) []foundation.MenuItem {
	distance := g.currentMap().MoveDistance(g.Player.Position(), actor.Position())

	if g.canPlayerTalkToActor(actor) {
		buffer = append(buffer, foundation.MenuItem{
			Name:       "[white]Talk To[-]",
			Action:     func() { g.PlayerStartDialogue(actor.GetDialogueFile(), actor) },
			CloseMenus: true,
		})
	}
	buffer = append(buffer, foundation.MenuItem{
		Name:       "Look at",
		Action:     func() { g.ui.OpenTextWindow(actor.GetDetailInfo()) },
		CloseMenus: true,
	})

	if g.Player.HasPerk(d100.PerkDisarm) && actor.GetInventory().HasWeaponEquipped() && distance == 1 {
		chances := formatContestSkSk(g.Player, actor, d100.SkillForUnarmed, d100.SkillForUnarmed)
		label := fmt.Sprintf("Disarm (%s)", chances)
		buffer = append(buffer, foundation.MenuItem{
			Name: label,
			Action: g.animatedActionFromMenu(func() {
				g.actorDisarm(g.Player, actor)
			}),
			CloseMenus: true,
		})
	}

	if actor.IsHostileTowards(g.Player) || distance > 1 {
		return buffer
	}
	if g.Player.HasPerk(d100.PerkPickpocket) {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Pickpocket",
			Action: g.animatedActionFromMenu(func() {
				g.StartPickpocket(actor)
			}),
			CloseMenus: true,
		})
	}

	if !actor.IsSleeping() {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Melee Attack",
			Action: g.animatedActionFromMenu(func() {
				g.playerMeleeAttack(actor)
			}),
			CloseMenus: true,
		})
		if g.Player.HasPerk(d100.PerkNonLethalTakeDown) {
			nonLethalChanceString := formatContestStSt(g.Player, actor, d100.Strength, d100.Strength)
			buffer = append(buffer, foundation.MenuItem{
				Name: fmt.Sprintf("Non-Lethal Takedown (%s)", nonLethalChanceString),
				Action: g.animatedActionFromMenu(func() {
					g.playerNonLethalTakedown(actor)
				}),
				CloseMenus: true,
			})
		}
	} else {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Wake Up",
			Action: func() {
				g.msg(foundation.Msg("You shake the sleeping figure awake."))
				actor.WakeUp()
				g.endPlayerTurn(g.Player.TimeNeededForActions())
			},
			CloseMenus: true,
		})
	}

	if g.currentMap().IsPositionNextToTileWithFlag(actor.Position(), gridmap.TileFlagWater) {
		label := "Drown"
		if !actor.IsSleeping() {
			drownChanceString := formatContestStSt(g.Player, actor, d100.Strength, d100.Strength)
			label = fmt.Sprintf("Drown (%s)", drownChanceString)
		}

		buffer = append(buffer, foundation.MenuItem{
			Name: label,
			Action: g.animatedActionFromMenu(func() {
				g.playerDrown(actor)
			}),
			CloseMenus: true,
		})
	}

	if g.Player.HasPerk(d100.PerkBackstab) && g.Player.GetInventory().HasMeleeWeaponEquipped() {
		label := "Backstab"
		if !actor.IsSleeping() {
			stabChanceString := formatContestSkSt(g.Player, actor, d100.SkillForBackstabbing, d100.Perception)
			label = fmt.Sprintf("Backstab (%s)", stabChanceString)
		}
		buffer = append(buffer, foundation.MenuItem{
			Name: label,
			Action: g.animatedActionFromMenu(func() {
				g.playerBackstab(actor)
			}),
			CloseMenus: true,
		})
	}
	return buffer
}

func formatContestStSt(one *Actor, two *Actor, statOne d100.Stat, statTwo d100.Stat) string {
	percentOne := d100.Percentage(one.GetCharSheet().GetStat(statOne) * 10)
	percentTwo := d100.Percentage(two.GetCharSheet().GetStat(statTwo) * 10)
	return fmt.Sprintf("%d%% vs %d%%", int(percentOne), int(percentTwo))
}

func formatContestSkSt(one *Actor, two *Actor, skillOne d100.Skill, statTwo d100.Stat) string {
	percentOne := d100.Percentage(two.GetCharSheet().GetSkill(skillOne))
	percentTwo := d100.Percentage(one.GetCharSheet().GetStat(statTwo) * 10)
	return fmt.Sprintf("%d%% vs %d%%", int(percentOne), int(percentTwo))
}

func formatContestSkSk(one *Actor, two *Actor, skillOne d100.Skill, skillTwo d100.Skill) string {
	percentOne := d100.Percentage(two.GetCharSheet().GetSkill(skillOne))
	percentTwo := d100.Percentage(one.GetCharSheet().GetSkill(skillTwo))
	return fmt.Sprintf("%d%% vs %d%%", int(percentOne), int(percentTwo))
}

func (g *GameState) OpenContextMenuForItem(uiItem foundation.Item, done func()) {
	item := uiItem
	contextActions := []foundation.MenuItem{
		{Name: "Inspect", Action: func() {
			g.PlayerExamineItem(item)
		}},
	}

	if len(contextActions) == 0 {
		return
	}
	if len(contextActions) == 1 {
		contextActions[0].Action()
		return
	}
	g.ui.OpenMenu(contextActions)
}
