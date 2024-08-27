package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
	"fmt"
)

func (g *GameState) appendContextActionsForActor(buffer []foundation.MenuItem, actor *Actor) []foundation.MenuItem {
	if actor.HasDialogue() && !actor.IsSleeping() {
		buffer = append(buffer, foundation.MenuItem{
			Name:       "Talk To",
			Action:     func() { g.StartDialogue(actor.GetDialogueFile(), actor, false) },
			CloseMenus: true,
		})
	}

	if actor.IsHostileTowards(g.Player) {
		return buffer
	}

	if actor.HasStealableItems() {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Pickpocket",
			Action: func() {
				g.StartPickpocket(actor)
			},
			CloseMenus: true,
		})
	}

	if !actor.IsSleeping() {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Melee Attack",
			Action: func() {
				g.playerMeleeAttack(actor)
			},
			CloseMenus: true,
		})
		nonLethalChanceString := formatContestStSt(g.Player, actor, special.Strength, special.Strength)
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
				actor.WakeUp()
			},
			CloseMenus: true,
		})
	}

	if g.Player.GetEquipment().HasMeleeWeaponEquipped() {
		label := "Backstab"
		if !actor.IsSleeping() {
			stabChanceString := formatContestSkSt(g.Player, actor, special.Sneak, special.Perception)
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

func formatContestStSt(one *Actor, two *Actor, statOne special.Stat, statTwo special.Stat) string {
	percentOne := special.Percentage(one.GetCharSheet().GetStat(statOne) * 10)
	percentTwo := special.Percentage(two.GetCharSheet().GetStat(statTwo) * 10)
	return fmt.Sprintf("%d%% vs %d%%", percentOne, percentTwo)
}

func formatContestSkSt(one *Actor, two *Actor, skillOne special.Skill, statTwo special.Stat) string {
	percentOne := special.Percentage(two.GetCharSheet().GetSkill(skillOne))
	percentTwo := special.Percentage(one.GetCharSheet().GetStat(statTwo) * 10)
	return fmt.Sprintf("%d%% vs %d%%", percentOne, percentTwo)
}

// Currently not in use
func (g *GameState) openContextMenuForItem(uiItem foundation.ItemForUI) {
	itemStack := uiItem.(*InventoryStack)
	item := itemStack.First()

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
