package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"fmt"
)

func (g *GameState) appendContextActionsForActor(buffer []foundation.MenuItem, actor *Actor) []foundation.MenuItem {
	distance := g.currentMap().MoveDistance(g.Player.Position(), actor.Position())

	if actor.HasDialogue() && !actor.IsSleeping() && distance <= 4 {
		buffer = append(buffer, foundation.MenuItem{
			Name:       "Talk To",
			Action:     func() { g.StartDialogue(actor.GetDialogueFile(), actor, false) },
			CloseMenus: true,
		})
	}

	if actor.IsHostileTowards(g.Player) || distance > 1 {
		return buffer
	}

	buffer = append(buffer, foundation.MenuItem{
		Name: "Pickpocket",
		Action: func() {
			g.StartPickpocket(actor)
		},
		CloseMenus: true,
	})

	if !actor.IsSleeping() {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Melee Attack",
			Action: func() {
				g.playerMeleeAttack(actor)
			},
			CloseMenus: true,
		})
		nonLethalChanceString := formatContestStSt(g.Player, actor, d100.Strength, d100.Strength)
		buffer = append(buffer, foundation.MenuItem{
			Name: fmt.Sprintf("Non-Lethal Takedown (%s)", nonLethalChanceString),
			Action: func() {
				g.playerNonLethalTakedown(actor)
			},
			CloseMenus: true,
		})
	} else {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Wake Up",
			Action: func() {
				g.msg(foundation.Msg("You shake the sleeping figure awake."))
				actor.WakeUp()
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
			Action: func() {
				g.playerDrown(actor)
			},
			CloseMenus: true,
		})
	}

	if g.Player.GetEquipment().HasMeleeWeaponEquipped() {
		label := "Backstab"
		if !actor.IsSleeping() {
			stabChanceString := formatContestSkSt(g.Player, actor, d100.SkillForBackstabbing, d100.Perception)
			label = fmt.Sprintf("Backstab (%s)", stabChanceString)
		}
		buffer = append(buffer, foundation.MenuItem{
			Name: label,
			Action: func() {
				g.playerBackstab(actor)
			},
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

// Currently not in use
func (g *GameState) openContextMenuForItem(uiItem foundation.Item) {
	item := uiItem

	contextActions := []foundation.MenuItem{
		{Name: "Inspect", Action: g.inspectItem(item)},
		{
			Name: "Drop",
			Action: func() {
				g.actorDropItem(g.Player, item)
			},
			CloseMenus: true,
		},
	}
	if item.IsEquippable() {
		equipment := g.Player.GetEquipment()

		var equipAction foundation.MenuItem
		if equipment.IsEquipped(item) {
			equipAction = foundation.MenuItem{
				Name: "Unequip",
				Action: func() {
					g.actorUnequipItem(g.Player, item)
				},
				CloseMenus: false,
			}
		} else {
			equipAction = foundation.MenuItem{
				Name: "Equip",
				Action: func() {
					g.actorEquipItem(g.Player, item)
				},
				CloseMenus: false,
			}
		}

		contextActions = append(contextActions, equipAction)
	}
	if item.IsUsable() {
		useAction := foundation.MenuItem{
			Name: "Use",
			Action: func() {
				g.actorUseItem(g.Player, item)
			},
			CloseMenus: true,
		}
		contextActions = append(contextActions, useAction)
	}

	if item.IsZappable() {
		zapAction := foundation.MenuItem{
			Name: "Zap",
			Action: func() {
				g.startZapItem(item)
			},
			CloseMenus: true,
		}
		contextActions = append(contextActions, zapAction)
	}

	if item.IsThrowable() {
		throwAction := foundation.MenuItem{
			Name: "Throw",
			Action: func() {
				g.startThrowItem(item)
			},
			CloseMenus: true,
		}
		contextActions = append(contextActions, throwAction)
	}

	g.ui.OpenMenu(contextActions)
}
