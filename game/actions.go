package game

import (
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"path"
)

var NoModifiers []dice_curve.Modifier

// MISC ACTIONS

func (g *GameState) SwitchWeapons() {
	equipment := g.Player.GetEquipment()
	equipment.SwitchWeapons()
}

func (g *GameState) CycleTargetMode() {
	equippedWeapon, hasWeapon := g.Player.GetEquipment().GetMainHandItem()
	if !hasWeapon || !equippedWeapon.IsWeapon() {
		g.msg(foundation.Msg("You have no weapon equipped"))
		return
	}
	equippedWeapon.GetWeapon().CycleTargetMode()
	g.ui.UpdateStats()
}

func (g *GameState) Wait() {
	g.endPlayerTurn(10)
}

func (g *GameState) ReloadWeapon() {
	weapon, hasWeapon := g.Player.GetEquipment().GetMainHandItem()
	if !hasWeapon {
		g.msg(foundation.Msg("You have no weapon equipped"))
		return
	}
	weaponPart := weapon.GetWeapon()
	if weaponPart == nil {
		g.msg(foundation.Msg("You have no weapon equipped"))
		return
	}
	caliber := weaponPart.UsesAmmo()
	if caliber == "" {
		g.msg(foundation.Msg("This weapon does not use ammo"))
		return
	}
	bulletsNeededForFullClip, ammoKindLoaded := weaponPart.BulletsNeededForFullClip()
	if bulletsNeededForFullClip <= 0 {
		g.msg(foundation.Msg("This weapon needs no reloading"))
		return
	}
	inventory := g.Player.GetInventory()
	var ammo *Item
	if ammoKindLoaded != "" && inventory.HasAmmo(caliber, ammoKindLoaded) {
		ammo = inventory.RemoveAmmoByName(ammoKindLoaded, bulletsNeededForFullClip)
	} else {
		bulletsNeededForFullClip = weaponPart.GetMagazineSize()
		ammo = inventory.RemoveAmmoByCaliber(caliber, bulletsNeededForFullClip)
	}

	if ammo == nil {
		g.msg(foundation.Msg("You have no ammo for this weapon"))
		return
	}
	unloadedAmmo := weaponPart.LoadAmmo(ammo)
	if unloadedAmmo != nil {
		inventory.Add(unloadedAmmo)
	}
	g.ui.UpdateStats()
	g.ui.UpdateInventory()
}

// ADDITIONAL MENUS

func (g *GameState) PlayerApplySkill() {
	//TODO implement me
	panic("implement me")
}

func (g *GameState) OpenTacticsMenu() {
	var menuItems []foundation.MenuItem
	/*
		menuItems = append(menuItems, foundation.MenuItem{
			Name:       "Aimed Attack",
			Action:     nil,
			CloseMenus: true,
		})
		menuItems = append(menuItems, foundation.MenuItem{
			Name:       "All-Out Attack",
			Action:     nil,
			CloseMenus: true,
		})
		menuItems = append(menuItems, foundation.MenuItem{
			Name:       "All-Out Defense",
			Action:     nil,
			CloseMenus: true,
		})
		menuItems = append(menuItems, foundation.MenuItem{
			Name:       "Feint",
			Action:     nil,
			CloseMenus: true,
		})

		menuItems = append(menuItems, foundation.MenuItem{
			Name:       "Toggle Acrobatic Dodge",
			Action:     nil,
			CloseMenus: true,
		})
		menuItems = append(menuItems, foundation.MenuItem{
			Name:       "Defend & Retreat",
			Action:     nil,
			CloseMenus: true,
		})
		menuItems = append(menuItems, foundation.MenuItem{
			Name:       "Dive for cover",
			Action:     nil,
			CloseMenus: true,
		})
	*/

	menuItems = append(menuItems, foundation.MenuItem{
		Name: "Charge Attack",
		Action: func() {
			g.startZapEffect("charge_attack", nil)
		},
		CloseMenus: true,
	})

	charSheet := g.Player.GetCharSheet()
	if charSheet.GetActionPoints() > 0 {
		menuItems = append(menuItems, foundation.MenuItem{
			Name: "Heroic Charge",
			Action: func() {
				if charSheet.GetActionPoints() > 0 {
					payFatigue := func() {
						charSheet.LooseActionPoints(1)
					}
					g.startZapEffect("heroic_charge", payFatigue)
				} else {
					g.msg(foundation.Msg("You are too fatigued to perform a heroic charge"))
				}
			},
			CloseMenus: true,
		})
		menuItems = append(menuItems, foundation.MenuItem{
			Name: "Start Sprinting",
			Action: func() {
				g.startSprint(g.Player)
			},
			CloseMenus: true,
		})

	}
	g.ui.OpenMenu(menuItems)
}

// ITEM MANAGEMENT & APPLICATION

func (g *GameState) startZapItem(item *Item) {
	g.ui.SelectTarget(func(targetPos geometry.Point, hitZone int) {
		g.playerZapItemAndEndTurn(item, targetPos)
	})
}

func (g *GameState) startZapEffect(zapEffectName string, payCost func()) {
	g.ui.SelectTarget(func(targetPos geometry.Point, hitZone int) {
		if payCost != nil {
			payCost()
		}
		g.playerInvokeZapEffectAndEndTurn(zapEffectName, targetPos)
	})
}

func (g *GameState) PlayerApplyItem(uiItem foundation.ItemForUI) {
	itemStack, isItem := uiItem.(*InventoryStack)
	if !isItem {
		return
	}
	item := itemStack.First()
	g.playerUseOrZapItem(item)
}

func (g *GameState) playerUseOrZapItem(item *Item) {
	if item.IsReadable() {
		g.playerReadItem(item)
	} else if item.IsUsable() {
		g.actorUseItem(g.Player, item)
	} else if item.IsZappable() {
		g.startZapItem(item)
	}
}

func (g *GameState) playerReadItem(item *Item) {
	file := path.Join(g.config.DataRootDir, "text", item.GetTextFile()+".txt")
	lines := fxtools.ReadFileAsLines(file)
	if len(lines) > 0 {
		g.ui.OpenTextWindow(g.fillTemplatedTexts(lines))
		return
	}
}

func (g *GameState) actorUseItem(user *Actor, item *Item) {
	useEffectName := item.GetUseEffectName()

	if useEffectName == "" {
		g.msg(foundation.Msg("You cannot use this item"))
		return
	}

	if !g.hasPaidWithCharge(user, item) {
		return
	}

	actionEndsTurn, consequencesOfEffect := g.actorInvokeUseEffect(user, useEffectName)

	g.ui.AddAnimations(consequencesOfEffect)

	if user == g.Player {
		if actionEndsTurn {
			g.endPlayerTurn(10)
		}
	}
}

func (g *GameState) actorInvokeUseEffect(user *Actor, useEffectName string) (endsTurn bool, animations []foundation.Animation) {
	if effect, exists := GetAllUseEffects()[useEffectName]; exists {
		return effect(g, user)
	}
	return false, nil
}

func (g *GameState) PickupItem() {
	inventory := g.Player.GetInventory()
	if inventory.IsFull() {
		g.msg(foundation.Msg("You cannot carry any more items"))
		return
	}
	if item, exists := g.gridMap.TryGetItemAt(g.Player.Position()); exists {
		g.gridMap.RemoveItem(item)
		if item.IsGold() {
			g.Player.AddGold(item.GetCharges())
		} else {
			inventory.Add(item)
		}

		g.msg(foundation.HiLite("You picked up %s", item.Name()))
		//g.endPlayerTurn()
	}
}

func (g *GameState) DropItem(uiItem foundation.ItemForUI) {
	itemStack, isItem := uiItem.(*InventoryStack)
	if !isItem {
		return
	}
	item := itemStack.First()
	g.actorDropItem(g.Player, item)
}

func (g *GameState) actorDropItem(holder *Actor, item *Item) {
	equipment := holder.GetEquipment()
	if equipment.IsEquipped(item) {
		if equipment.CanUnequip(item) {
			g.actorUnequipItem(holder, item)
		} else {
			g.msg(foundation.Msg("You cannot remove this item"))
			return
		}
	}

	g.removeItemFromInventory(holder, item)
	g.addItemToMap(item, holder.Position())
	g.msg(foundation.HiLite("You dropped %s", item.Name()))
	if holder == g.Player {
		g.endPlayerTurn(g.Player.timeNeededForActions() / 2)
	}
}

func (g *GameState) inspectItem(item *Item) func() {
	return func() {
		g.ui.OpenTextWindow([]string{"Item", item.description})
	}
}

// EQUIP / UNEQUIP

func (g *GameState) EquipToggle(uiItem foundation.ItemForUI) {
	itemStack, isItem := uiItem.(*InventoryStack)
	if !isItem {
		return
	}
	item := itemStack.First()
	if !item.IsEquippable() {
		g.msg(foundation.Msg("You cannot equip this item"))
		return
	}
	equipment := g.Player.GetEquipment()
	if equipment.IsEquipped(item) {
		if equipment.CanUnequip(item) {
			g.actorUnequipItem(g.Player, item)
		} else {
			g.msg(foundation.Msg("You cannot remove this item"))
		}
	} else {
		if equipment.CanEquip(item) {
			g.actorEquipItem(g.Player, item)
		} else {
			g.msg(foundation.Msg("You cannot equip this item"))
		}
	}
}

func (g *GameState) actorEquipItem(wearer *Actor, item *Item) {
	equipment := wearer.GetEquipment()
	equipment.Equip(item)
	if wearer == g.Player {
		g.msg(foundation.HiLite("You equipped %s", item.Name()))
	}
}

func (g *GameState) actorUnequipItem(wearer *Actor, item *Item) {
	equipment := wearer.GetEquipment()
	equipment.UnEquip(item)
	if wearer == g.Player {
		g.msg(foundation.HiLite("You unequipped %s", item.Name()))
	}
}

// SPECIAL INVENTORY INTERACTIONS

func (g *GameState) ChooseItemForApply() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsUsableOrZappable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything usable."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.playerUseOrZapItem(item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Use what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.playerUseOrZapItem(item)
	})
}

func (g *GameState) ChooseItemForDrop() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return true
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.DropItem(stack)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Drop what?", func(itemStack foundation.ItemForUI) {
		g.DropItem(itemStack)
	})
}

func (g *GameState) ChooseItemForQuaff() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsPotion()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any potions."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Quaff what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
	})
}

func (g *GameState) ChooseItemForEat() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsFood()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any food."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Eat what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
	})
}

func (g *GameState) ChooseItemForRead() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsScroll()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any scrolls."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Read what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
	})
}

func (g *GameState) ChooseItemForZap() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsZappable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any wands."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.startZapItem(item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Zap what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.startZapItem(item)
	})
}

func (g *GameState) ChooseItemForUse() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsUsable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything usable."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Use what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
	})
}

func (g *GameState) ChooseWeaponForWield() {
	equipment := g.Player.GetEquipment()
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsWeapon() && item.IsEquippable() && !equipment.IsEquipped(item)
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any weapons."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorEquipItem(g.Player, stack.First())
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Wield what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorEquipItem(g.Player, item)
	})
}

func (g *GameState) ChooseArmorForWear() {
	equipment := g.Player.GetEquipment()
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsArmor() && item.IsEquippable() && !equipment.IsEquipped(item)
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any armor."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorEquipItem(g.Player, stack.First())
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Wear what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorEquipItem(g.Player, item)
	})
}

func (g *GameState) ChooseRingToPutOn() {
	equipment := g.Player.GetEquipment()
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsRing() && item.IsEquippable() && !equipment.IsEquipped(item)
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any rings."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorEquipItem(g.Player, stack.First())
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Put on what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorEquipItem(g.Player, item)
	})
}

func (g *GameState) ChooseArmorToTakeOff() {
	equipment := g.Player.GetEquipment()
	wornArmor := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsArmor() && item.IsEquippable() && equipment.IsEquipped(item)
	})

	if len(wornArmor) == 0 {
		g.msg(foundation.Msg("You are not wearing any armor."))
		return
	}
	if len(wornArmor) == 1 {
		stack, isStack := wornArmor[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorUnequipItem(g.Player, stack.First())
		return
	}
	g.ui.OpenInventoryForSelection(wornArmor, "Take off what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUnequipItem(g.Player, item)
	})
}

func (g *GameState) ChooseRingToRemove() {
	equipment := g.Player.GetEquipment()
	wornRings := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsRing() && item.IsEquippable() && equipment.IsEquipped(item)
	})
	if len(wornRings) == 0 {
		g.msg(foundation.Msg("You are not wearing any rings."))
		return
	}
	if len(wornRings) == 1 {
		stack, isStack := wornRings[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorUnequipItem(g.Player, stack.First())
		return

	}
	g.ui.OpenInventoryForSelection(wornRings, "Remove what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUnequipItem(g.Player, item)
	})
}

// ASCEND / DESCEND

type StairsInLevel int

func (l StairsInLevel) AllowsUp() bool {
	return l == StairsUpOnly || l == StairsBoth
}
func (l StairsInLevel) AllowsDown() bool {
	return l == StairsDownOnly || l == StairsBoth
}

const (
	StairsNone StairsInLevel = iota
	StairsUpOnly
	StairsDownOnly
	StairsBoth
)

func (g *GameState) PlayerInteractWithMap() {
	pos := g.Player.Position()

	transition, transitionExists := g.gridMap.GetTransitionAt(pos)
	if transitionExists {
		currentMapName := g.gridMap.GetName()
		location := g.gridMap.GetNamedLocationByPos(pos)
		lockFlagName := fmt.Sprintf("lock(%s/%s)", currentMapName, location)
		if g.gameFlags.HasFlag(lockFlagName) {
			g.msg(foundation.Msg("The way is blocked"))
			return
		}
		g.GotoNamedLevel(transition.TargetMap, transition.TargetLocation)
		return
	}
}

// HELPER STUFF

func ModRangedDefault(origin geometry.Point, target *Actor) []dice_curve.Modifier {
	// TODO: Light
	var attackMods []dice_curve.Modifier

	// step 1. size modifier
	sizeMod := target.GetSizeModifier()
	if sizeMod != 0 {
		attackMods = append(attackMods, ModFlat(sizeMod, "size"))
	}

	// step 2. range modifier
	dist := geometry.Distance(origin, target.Position())
	rangeUsed := int(dist)
	distMod := dice_curve.GetDistanceModifier(rangeUsed)
	if distMod != 0 {
		attackMods = append(attackMods, ModFlat(distMod, fmt.Sprintf("range(%d)", rangeUsed)))
	}

	return attackMods
}

func OneAnimation(anim foundation.Animation) []foundation.Animation {
	if anim == nil {
		return nil
	}
	return []foundation.Animation{anim}
}

func useEffectExists(effectName string) bool {
	_, exists := GetAllUseEffects()[effectName]
	return exists
}

func (g *GameState) couldPlayerSeeActor(actor *Actor) bool {
	if actor.HasFlag(foundation.FlagInvisible) && !g.Player.HasFlag(foundation.FlagSeeInvisible) {
		return false
	}

	return true
}
