package game

import "RogueUI/foundation"

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
				g.startAimItem(item)
			},
			CloseMenus: true,
		}
		contextActions = append(contextActions, zapAction)
	}

	if item.IsThrowable() {
		throwAction := foundation.MenuItem{
			Name: "Throw",
			Action: func() {
				g.startRangedAttackWithMissile(item)
			},
			CloseMenus: true,
		}
		contextActions = append(contextActions, throwAction)
	}

	g.ui.OpenMenu(contextActions)
}
