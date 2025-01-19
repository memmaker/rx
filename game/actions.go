package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"path"
	"strconv"
	"time"
)

// MISC ACTIONS

func (g *GameState) CycleTargetMode() {
	equippedWeapon, hasWeapon := g.Player.GetInventory().GetEquippedWeapon()
	if !hasWeapon || !equippedWeapon.IsWeapon() {
		g.msg(foundation.Msg("You have no weapon equipped"))
		return
	}
	equippedWeapon.CycleTargetMode()
	g.ui.UpdateStats()
}

func (g *GameState) Wait() {
	g.msg(foundation.Msg("Time passes"))
	g.endPlayerTurn(10)
}
func (g *GameState) PlayerReloadWeapon() {
	g.actorReloadMainHandWeapon(g.Player)
}
func (g *GameState) actorReloadMainHandWeapon(actor *Actor) bool {
	weaponPart, hasItem := actor.GetInventory().GetEquippedWeapon()
	if !hasItem {
		if actor == g.Player {
			g.msg(foundation.Msg("You have no weapon equipped"))
		}
		return false
	}
	if weaponPart.HasTag(foundation.TagNoReload) {
		if actor == g.Player {
			g.msg(foundation.Msg("This weapon cannot be reloaded"))
		}
		return false
	}
	if !weaponPart.NeedsAmmo() {
		if actor == g.Player {
			g.msg(foundation.Msg("This weapon does not use ammo"))
		}
		return false
	}

	bulletsNeededForFullClip, ammoKindLoaded := weaponPart.BulletsNeededForFullClip()
	if bulletsNeededForFullClip <= 0 {
		if actor == g.Player {
			g.msg(foundation.Msg("This weapon needs no reloading"))
		}
		return false
	}
	caliber := weaponPart.GetCaliber()
	inventory := actor.GetInventory()
	var ammo *Ammo
	if ammoKindLoaded != "" && inventory.HasAmmo(caliber, ammoKindLoaded) {
		ammo = inventory.RemoveAmmoByName(ammoKindLoaded, bulletsNeededForFullClip)
	} else {
		bulletsNeededForFullClip = weaponPart.GetMagazineSize()
		ammo = inventory.RemoveAmmoByCaliber(caliber, bulletsNeededForFullClip)
	}

	if ammo == nil {
		if actor == g.Player {
			g.msg(foundation.Msg("You have no ammo for this weapon"))
		}
		return false
	}
	unloadedAmmo := weaponPart.LoadAmmo(ammo)
	if unloadedAmmo != nil {
		inventory.AddItem(unloadedAmmo)
	}

	g.ui.PlayCue(weaponPart.GetReloadAudioCue())

	if actor == g.Player {
		g.ui.UpdateStats()
		g.ui.UpdateInventory()
	}
	return true
}

// ADDITIONAL MENUS

func (g *GameState) PlayerApplySkill() {
	//TODO implement me
	panic("implement me")
}
func (g *GameState) OpenCyberWareMenu() {
	var menuItems []foundation.MenuItem
	for c := CyberWare(0); c < CyberWareCount; c++ {
		if g.Player.HasCyberWare(c) {

			if c == CyberWareClock { // clock handles its own activation
				menuItems = append(menuItems, foundation.MenuItem{
					Name:       fmt.Sprintf("Activate %s", c.String()),
					Action:     g.ShowDateTime,
					CloseMenus: true,
				})
				continue
			}

			activateCall := func() {
				g.Player.SetCyberWareActive(c, true)
				g.msg(foundation.HiLite("You activate %s", c.String()))
			}
			deactivateCall := func() {
				g.Player.SetCyberWareActive(c, false)
				g.msg(foundation.HiLite("You deactivate %s", c.String()))
			}

			if c == CyberWareThermopticCamouflage {
				activateCall = func() {
					if g.Player.GetCharSheet().GetActionPoints() > 0 {
						g.Player.GetCharSheet().LooseActionPoints(1)
						g.Player.SetCyberWareActive(c, true)
						g.Player.GetFlags().Set(foundation.FlagActiveCamouflage)
						g.msg(foundation.HiLite("You activate %s", c.String()))
					} else {
						g.msg(foundation.Msg("You do not have enough action points"))
					}
				}
				deactivateCall = func() {
					g.Player.SetCyberWareActive(c, false)
					g.Player.GetFlags().Unset(foundation.FlagActiveCamouflage)
					g.msg(foundation.HiLite("You deactivate %s", c.String()))
				}
			}

			if g.Player.IsCyberWareActive(c) {
				menuItems = append(menuItems, foundation.MenuItem{
					Name: fmt.Sprintf("Deactivate %s", c.String()),
					Action: func() {
						deactivateCall()
						g.OpenCyberWareMenu()
					},
					CloseMenus: true,
				})
			} else {
				menuItems = append(menuItems, foundation.MenuItem{
					Name: fmt.Sprintf("Activate %s", c.String()),
					Action: func() {
						activateCall()
						g.OpenCyberWareMenu()
					},
					CloseMenus: true,
				})
			}
		}
	}
	if len(menuItems) == 0 {
		g.ui.OpenTextWindow("You have no cyberware")
		return
	}
	g.ui.OpenMenu(menuItems)
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
	currentAimLevel := g.Player.GetFlags().Get(foundation.FlagConcentratedAiming)
	if currentAimLevel < 3 {
		menuItems = append(menuItems, foundation.MenuItem{
			Name: "Concentrate on target",
			Action: func() {
				g.Player.GetFlags().Increment(foundation.FlagConcentratedAiming)
				g.endPlayerTurn(g.Player.TimeNeededForActions())
			},
			CloseMenus: true,
		})
	}
	if g.Player.HasPerk(d100.PerkCharge) {
		menuItems = append(menuItems, foundation.MenuItem{
			Name: "Charge Attack",
			Action: func() {
				g.startZapEffect("charge_attack", nil, foundation.Params{})
			},
			CloseMenus: true,
		})
	}
	charSheet := g.Player.GetCharSheet()
	if g.Player.HasPerk(d100.PerkCharge) && charSheet.GetActionPoints() > 0 {
		menuItems = append(menuItems, foundation.MenuItem{
			Name: "Heroic Charge",
			Action: func() {
				if charSheet.GetActionPoints() > 0 {
					payFatigue := func() {
						charSheet.LooseActionPoints(1)
					}
					g.startZapEffect("heroic_charge", payFatigue, foundation.Params{})
				} else {
					g.msg(foundation.Msg("You are too fatigued to perform a heroic charge"))
				}
			},
			CloseMenus: true,
		})

	}
	g.ui.OpenMenu(menuItems)
}

// ITEM MANAGEMENT & APPLICATION
func (g *GameState) startZapItem(item foundation.Zappable) {
	g.ui.SelectTarget(g.getThrownChanceToHitForUI, func(targetPos geometry.Point) {
		g.playerZapItemAndEndTurn(item, targetPos)
	})
}

func (g *GameState) startZapEffect(zapEffectName string, payCost func(), params foundation.Params) {

	// TODO: This needs to depend on the zap effect
	targetingFunc := g.getThrownChanceToHitForUI

	g.ui.SelectTarget(targetingFunc, func(targetPos geometry.Point) {
		if payCost != nil {
			payCost()
		}
		g.playerInvokeZapEffectAndEndTurn(zapEffectName, targetPos, params)
	})
}

func (g *GameState) PlayerApplyItem(item foundation.Item) {
	if item.IsLockpick() {
		g.PlayerExamineItem(item)
	} else if item.IsDrug() {
		g.actorConsumeDrug(g.Player, item.(*GenericItem))
	} else if item.IsReadable() {
		g.playerReadItem(item)
	} else if item.IsUsable() {
		g.actorUseItem(g.Player, item)
	} else if item.IsZappable() {
		if item.HasTag(foundation.TagTimed) {
			g.actorSetItemCountdown(g.Player, item)
		} else {
			g.startZapItem(item)
		}
	}
}

func (g *GameState) OpenRepairMenu() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.NeedsRepair() && item.GetQuality() < g.Player.GetMaxRepairQuality()
	})
	if len(inventory) == 0 {
		g.ui.OpenTextWindow("You have nothing to repair")
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Repair what?", func(itemStack foundation.Item) {
		g.playerRepairItem(itemStack)
	})
}

func (g *GameState) openNPCRepairMenu(npcVendor *Actor, whenDone func()) {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.NeedsRepair() && item.GetQuality() < npcVendor.GetMaxRepairQuality()
	})
	if len(inventory) == 0 {
		g.ui.OpenTextWindow("You have nothing to repair")
		return
	}
	g.ui.OpenInventoryForSelectionWithClose(inventory, "Repair what?", func(itemStack foundation.Item) {
		g.actorRepairItem(npcVendor, itemStack)
	}, whenDone)
}

func (g *GameState) actorRepairItem(npcVendor *Actor, item foundation.Repairable) {
	if !item.NeedsRepair() {
		g.msg(foundation.Msg("You do not need to repair this item"))
		return
	}

	firstQuality := item.GetQuality()
	secondQuality := d100.Percentage(50)

	newQuality := npcVendor.GetRepairQuality(firstQuality, secondQuality)

	item.SetQuality(newQuality)

	g.msg(foundation.HiLite("%s was repaired to %s", item.Name(), fmt.Sprintf("%d%%", int(newQuality))))

	g.ui.UpdateInventory()
}

func (g *GameState) playerRepairItem(item foundation.Repairable) {
	if !item.NeedsRepair() {
		g.msg(foundation.Msg("You do not need to repair this item"))
		return
	}

	redundantCopiesOfTheItemForRepair := g.GetFilteredInventory(func(i foundation.Item) bool {
		return item.CanBeRepairedWith(i) && i.GetQuality() < g.Player.GetMaxRepairQuality()
	})

	if len(redundantCopiesOfTheItemForRepair) == 0 {
		g.msg(foundation.Msg("You have nothing to repair this item with"))
		return
	}

	g.ui.OpenInventoryForSelection(redundantCopiesOfTheItemForRepair, "Repair with what?", func(uiItem foundation.Item) {
		g.playerRepairItemWith(item, uiItem)
	})
}

func (g *GameState) playerRepairItemWith(toRepair foundation.Repairable, spareParts foundation.Repairable) {
	if !toRepair.CanBeRepairedWith(spareParts) {
		g.msg(foundation.Msg("You cannot repair this item with that"))
		return
	}
	if weapon, isWeapon := spareParts.(*Weapon); isWeapon {
		ammo := weapon.Unload()
		g.Player.GetInventory().AddItem(ammo)
	}

	g.Player.GetInventory().RemoveItem(spareParts.(foundation.Item))

	firstQuality := toRepair.GetQuality()
	secondQuality := spareParts.GetQuality()

	newQuality := g.Player.GetRepairQuality(firstQuality, secondQuality)

	toRepair.SetQuality(newQuality)

	g.msg(foundation.HiLite("You repaired %s to %s", toRepair.Name(), fmt.Sprintf("%d%%", int(newQuality))))

	g.ui.UpdateInventory()
}

func (g *GameState) playerReadItem(item foundation.Readable) {
	if item.IsSkillBook() {
		skill, increase := item.GetSkillBookValues()
		if g.Player.GetCharSheet().GetUnmodifiedSkill(skill) >= d100.SkillCap {
			g.msg(foundation.Msg("You're already a master of this skill"))
			return
		}
		g.advanceTime(2 * time.Duration(time.Hour))
		g.Player.GetCharSheet().AddSkillPointsTo(skill, increase)
		g.msg(foundation.HiLite("Your %s skill increased by %s", skill.String(), strconv.Itoa(increase)))
		g.Player.GetInventory().RemoveItem(item.(foundation.Item))
		return
	}
	var lines string
	if item.GetTextFile() != "" {
		file := path.Join(g.config.DataRootDir, "text", item.GetTextFile()+".txt")
		lines = fxtools.ReadFile(file)
	} else {
		lines = item.GetText()
	}
	if len(lines) > 0 {
		g.ui.OpenTextWindow(g.fillTemplatedTextCustom(lines, item.TextVariables(g.GetScriptFuncs())))
		return
	}
}

func (g *GameState) actorUseItem(user *Actor, item foundation.Usable) {
	useEffectName := item.UseEffect()

	if useEffectName == "" {
		g.msg(foundation.Msg("You cannot use this item"))
		return
	}

	if !g.hasPaidWithCharge(user, item.(foundation.Item)) {
		return
	}

	actionEndsTurn, consequencesOfEffect := g.actorInvokeUseEffect(user, useEffectName)

	g.ui.AddAnimations(consequencesOfEffect)

	if user == g.Player {
		if actionEndsTurn {
			g.endPlayerTurn(g.Player.TimeNeededForActions())
		}
	}
}

func (g *GameState) actorSetItemCountdown(user *Actor, item foundation.Timable) {
	useEffectName := item.ZapEffect()

	if useEffectName == "" || !item.HasTag(foundation.TagTimed) {
		if user == g.Player {
			g.msg(foundation.Msg("You cannot set the timer on this item"))
		}
		return
	}

	if g.metronome.HasTimed(item) {
		if user == g.Player {
			g.msg(foundation.Msg("This timer has already been set"))
		}
		return
	}

	setCharge := func(turns int) {
		g.msg(foundation.HiLite("You set the timer of %s to %s", item.Name(), strconv.Itoa(turns)))
		item.SetCharges(turns)
		g.metronome.AddTimed(item, g.currentMapName, func() {
			consequencesOfEffect := g.actorInvokeZapEffect(user, useEffectName, item.Position(), item.GetEffectParameters())
			g.ui.AddAnimations(consequencesOfEffect)
			g.removeItemFromGame(item.(foundation.Item))
		})
	}
	turns := 5
	if g.Player == user {
		g.ui.AskForString("Set countdown", "5", func(input string) {
			turns, _ = strconv.Atoi(input)
			setCharge(turns)
		})
	} else {
		setCharge(turns)
	}
}

func (g *GameState) actorInvokeUseEffect(user *Actor, useEffectName string) (endsTurn bool, animations []foundation.Animation) {
	if effect, exists := GetAllUseEffects()[useEffectName]; exists {
		return effect(g, user)
	}
	return false, nil
}

func (g *GameState) PlayerPickupItem() {
	itemPos := g.Player.Position()
	g.PlayerPickupItemAt(itemPos)
}

func (g *GameState) PlayerPickupItemAt(itemPos geometry.Point) {
	inventory := g.Player.GetInventory()
	if inventory.IsFull() {
		g.msg(foundation.Msg("You cannot carry any more items"))
		return
	}

	if item, exists := g.currentMap().TryGetItemAt(itemPos); exists {
		g.currentMap().RemoveItem(item)
		inventory.AddItem(item)

		g.ui.PlayCue("world/pickup")
		g.msg(foundation.HiLite("You picked up %s", item.Name()))

		if item.GetPickupFlag() != "" {
			g.gameFlags.Increment(item.GetPickupFlag())
		}
		//g.endPlayerTurn()
	}
}

func (g *GameState) DropItemFromInventory(uiItem foundation.Item) {
	g.PlayerDropItem(uiItem)
	g.GetNonAmmoPlayerInventory()
}

func (g *GameState) PlayerDropItem(item foundation.Item) {
	holder := g.Player
	holder.Inventory.RemoveItem(item)

	g.addItemToMap(item, holder.Position())

	g.msg(foundation.HiLite("You dropped %s", item.Name()))
	if item.GetDropFlag() != "" {
		g.gameFlags.Increment(item.GetDropFlag())
	}
	g.endPlayerTurn(g.Player.TimeNeededForActions() / 2)
	g.ui.PlayCue("world/drop")
}

func (g *GameState) PlayerExamineItem(item foundation.Item) {
	g.ui.OpenTextWindow(item.FullDescription(g.inventoryColorCode(item)))
}

// EQUIP / UNEQUIP

func (g *GameState) EquipToggle(item foundation.Item) {
	if !item.IsEquippable() {
		g.msg(foundation.Msg("You cannot equip this item"))
		return
	}
	equipment := g.Player.GetInventory()
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

func (g *GameState) actorEquipItem(wearer *Actor, item foundation.Item) {
	equipment := wearer.GetInventory()
	equipment.Equip(item)
	if wearer == g.Player {
		g.msg(foundation.HiLite("You equipped %s", item.Name()))
		g.forDetectingOtherFactionObserversOf(wearer, func(observer *Actor) (continueIteration bool) {
			if wearer.IsOpenCarryWeapon() && observer.HasFlag(foundation.FlagProhibitsOpenCarry) {
				g.onSeenWithWeapon(wearer, observer)
				return false
			}
			return true
		})
		g.endPlayerTurn(10)
	}
}

func (g *GameState) actorUnequipItem(wearer *Actor, item foundation.Item) {
	if item == nil {
		return
	}
	equipment := wearer.GetInventory()
	equipment.UnEquip(item)
	if wearer == g.Player {
		g.msg(foundation.HiLite("You unequipped %s", item.Name()))
		g.endPlayerTurn(10)
	}
}

// SPECIAL INVENTORY INTERACTIONS

func (g *GameState) ChooseItemForApply() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsUsableOrZappable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything usable."))
		return
	}
	if len(inventory) == 1 {
		g.PlayerApplyItem(inventory[0])
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Use what?", func(itemStack foundation.Item) {
		g.PlayerApplyItem(itemStack)
	})
}

func (g *GameState) ChooseItemForDrop() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return true
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything."))
		return
	}
	if len(inventory) == 1 {
		g.PlayerDropItem(inventory[0])
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Drop what?", func(itemStack foundation.Item) {
		g.PlayerDropItem(itemStack)
	})
}

func (g *GameState) OpenAmmoInventory() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsAmmo()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any ammunition."))
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Drop what?", func(itemStack foundation.Item) {
		g.PlayerDropItem(itemStack)
	})
}

func (g *GameState) ChooseItemForEat() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsConsumable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any food."))
		return
	}
	if len(inventory) == 1 {

		g.actorUseItem(g.Player, inventory[0])
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Eat what?", func(itemStack foundation.Item) {
		g.actorUseItem(g.Player, itemStack)
	})
}

func (g *GameState) ChooseItemForZap() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsZappable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any wands."))
		return
	}
	if len(inventory) == 1 {
		g.startZapItem(inventory[0])
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Zap what?", func(itemStack foundation.Item) {
		g.startZapItem(itemStack)
	})
}

func (g *GameState) ChooseItemForUse() {
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsUsable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything usable."))
		return
	}
	if len(inventory) == 1 {
		g.actorUseItem(g.Player, inventory[0])
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Use what?", func(itemStack foundation.Item) {
		g.actorUseItem(g.Player, itemStack)
	})
}

func (g *GameState) ChooseWeaponForWield() {
	equipment := g.Player.GetInventory()
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsWeapon() && item.IsEquippable() && !equipment.IsEquipped(item)
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any weapons."))
		return
	}
	if len(inventory) == 1 {
		g.actorEquipItem(g.Player, inventory[0])
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Wield what?", func(itemStack foundation.Item) {
		g.actorEquipItem(g.Player, itemStack)
	})
}

func (g *GameState) ChooseArmorForWear() {
	equipment := g.Player.GetInventory()
	inventory := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsArmor() && item.IsEquippable() && !equipment.IsEquipped(item)
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any armor."))
		return
	}
	if len(inventory) == 1 {
		g.actorEquipItem(g.Player, inventory[0])
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Wear what?", func(itemStack foundation.Item) {
		g.actorEquipItem(g.Player, itemStack)
	})
}

func (g *GameState) ChooseArmorToTakeOff() {
	equipment := g.Player.GetInventory()
	wornArmor := g.GetFilteredInventory(func(item foundation.Item) bool {
		return item.IsArmor() && item.IsEquippable() && equipment.IsEquipped(item)
	})

	if len(wornArmor) == 0 {
		g.msg(foundation.Msg("You are not wearing any armor."))
		return
	}
	if len(wornArmor) == 1 {
		g.actorUnequipItem(g.Player, wornArmor[0])
		return
	}
	g.ui.OpenInventoryForSelection(wornArmor, "Take off what?", func(itemStack foundation.Item) {

		g.actorUnequipItem(g.Player, itemStack)
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

func (g *GameState) CheckTransition() {
	pos := g.Player.Position()
	transition, transitionExists := g.currentMap().GetTransitionAt(pos)

	doTransition := func() {
		title := "Move to another area"
		message := fmt.Sprintf("Do you want to leave %s?", g.currentMap().GetDisplayName())
		g.ui.AskForConfirmation(title, message, func(didConfirm bool) {
			if didConfirm {
				currentMapName := g.currentMap().GetName()
				location := g.currentMap().GetNamedLocationByPos(pos)
				lockFlagName := fmt.Sprintf("lock(%s/%s)", currentMapName, location)
				if g.gameFlags.HasFlag(lockFlagName) {
					g.msg(foundation.Msg("The way is blocked"))
					return
				}
				g.transitionToMapLocation(transition.TargetMap, transition.TargetLocation)
				g.advanceTime(5 * time.Minute)
			}
		})
	}

	if transitionExists {
		g.QueueActionAfterAnimation(doTransition)
	}
}

// HELPER STUFF

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
	if actor.HasFlag(foundation.FlagActiveCamouflage) && !g.Player.HasFlag(foundation.FlagSeeInvisible) {
		return false
	}

	return true
}
