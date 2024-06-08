package game

import (
	"RogueUI/dice_curve"
	"RogueUI/foundation"
)

type Equipment struct {
	slots     map[foundation.EquipSlot]*Item
	onChanged func()
}

func NewEquipment() *Equipment {
	return &Equipment{
		slots: make(map[foundation.EquipSlot]*Item),
	}
}

func (e *Equipment) IsEquipped(item *Item) bool {
	slotName := item.SlotName()

	if slotName == foundation.SlotNameMainHand || slotName == foundation.SlotNameOffHand {
		if slotItem, exists := e.slots[foundation.SlotNameMainHand]; exists && slotItem == item {
			return true
		}
		if slotItem, exists := e.slots[foundation.SlotNameOffHand]; exists && slotItem == item {
			return true
		}
	}
	if slotName == foundation.SlotNameRing {
		if slotItem, exists := e.slots[foundation.SlotNameRingLeft]; exists && slotItem == item {
			return true
		}
		if slotItem, exists := e.slots[foundation.SlotNameRingRight]; exists && slotItem == item {
			return true
		}
	}
	if slotItem, exists := e.slots[slotName]; exists && slotItem == item {
		return true
	}
	return false
}

func (e *Equipment) Equip(item *Item) {
	defer e.changed()
	slotName := item.SlotName()
	if slotName == foundation.SlotNameMainHand || slotName == foundation.SlotNameOffHand {
		_, hasOffHandItem := e.slots[foundation.SlotNameOffHand]
		if !hasOffHandItem {
			e.slots[foundation.SlotNameOffHand] = item
		} else {
			e.slots[foundation.SlotNameMainHand] = item
		}
		return
	}

	if slotName == foundation.SlotNameRing {
		if _, exists := e.slots[foundation.SlotNameRingLeft]; !exists {
			e.slots[foundation.SlotNameRingLeft] = item
			return
		}
		if _, exists := e.slots[foundation.SlotNameRingRight]; !exists {
			e.slots[foundation.SlotNameRingRight] = item
			return
		}
	}

	e.slots[slotName] = item
}

func (e *Equipment) UnEquip(item *Item) {
	defer e.changed()
	slotName := item.SlotName()
	if slotName == foundation.SlotNameMainHand || slotName == foundation.SlotNameOffHand {
		if slotItem, exists := e.slots[foundation.SlotNameMainHand]; exists && slotItem == item {
			e.unEquipBySlot(foundation.SlotNameMainHand)
			return
		}
		if slotItem, exists := e.slots[foundation.SlotNameOffHand]; exists && slotItem == item {
			e.unEquipBySlot(foundation.SlotNameOffHand)
			return
		}
	}

	if slotName == foundation.SlotNameRing {
		if slotItem, exists := e.slots[foundation.SlotNameRingLeft]; exists && slotItem == item {
			e.unEquipBySlot(foundation.SlotNameRingLeft)
			return
		}
		if slotItem, exists := e.slots[foundation.SlotNameRingRight]; exists && slotItem == item {
			e.unEquipBySlot(foundation.SlotNameRingRight)
			return
		}
	}

	e.unEquipBySlot(slotName)
}

func (e *Equipment) unEquipBySlot(hand foundation.EquipSlot) {
	delete(e.slots, hand)
}

func (e *Equipment) GetBySlot(hand foundation.EquipSlot) *Item {
	return e.slots[hand]
}

func (e *Equipment) GetArmor() []*Item {
	var armor []*Item
	for _, item := range e.slots {
		if item.IsArmor() {
			armor = append(armor, item)
		}
	}
	return armor

}

func (e *Equipment) CanEquip(item *Item) bool {
	toBeReplaced := e.GetItemsToReplace(item)
	for _, itemToRemove := range toBeReplaced {
		if !e.CanUnequip(itemToRemove) {
			return false
		}
	}
	return true
}
func (e *Equipment) GetItemsToReplace(item *Item) []*Item {
	var items []*Item
	if item.SlotName() == foundation.SlotNameOffHand || item.SlotName() == foundation.SlotNameMainHand {
		_, hasItemInMainHand := e.slots[foundation.SlotNameMainHand]
		_, hasItemInOffHand := e.slots[foundation.SlotNameOffHand]

		if !hasItemInOffHand {
			return items
		}
		if !hasItemInMainHand {
			return items
		}
		items = appendIfNotNil(items, e.slots[foundation.SlotNameMainHand])
		return items
	}

	if item.SlotName() == foundation.SlotNameRing {
		if e.isOneSlotAvailable(foundation.SlotNameRingLeft, foundation.SlotNameRingRight) {
			return items
		} else {
			items = appendIfNotNil(items, e.slots[foundation.SlotNameRingLeft])
			return items
		}
	}
	items = appendIfNotNil(items, e.slots[item.SlotName()])
	return items
}

func (e *Equipment) isOneSlotAvailable(slotOne foundation.EquipSlot, slotTwo foundation.EquipSlot) bool {
	_, slotOneExists := e.slots[slotOne]
	_, slotTwoExists := e.slots[slotTwo]
	return !slotOneExists || !slotTwoExists
}

func appendIfNotNil(items []*Item, item *Item) []*Item {
	if item != nil {
		return append(items, item)
	}
	return items
}

func (e *Equipment) HasRangedWeaponEquipped() bool {
	slotMain := e.GetBySlot(foundation.SlotNameMainHand)
	return slotMain != nil && slotMain.IsRangedWeapon()
}

func (e *Equipment) HasMeleeWeaponEquipped() bool {
	slotMain := e.GetBySlot(foundation.SlotNameMainHand)
	return slotMain != nil && slotMain.IsMeleeWeapon()
}

func (e *Equipment) AllItems() []*Item {
	var items []*Item
	for _, item := range e.slots {
		items = append(items, item)
	}
	return items
}

func (e *Equipment) GetSlot(item *Item) foundation.EquipSlot {
	for slot, slotItem := range e.slots {
		if slotItem == item {
			return slot
		}
	}
	return foundation.SlotNameNotEquippable
}

func (e *Equipment) HasWeaponEquipped() bool {
	slot := e.GetBySlot(foundation.SlotNameMainHand)
	return slot != nil && slot.IsWeapon()
}

func (e *Equipment) HasArmorEquipped() bool {
	slot := e.GetBySlot(foundation.SlotNameArmorTorso)
	return slot != nil && slot.IsArmor()
}

func (e *Equipment) SetOnChangeHandler(onChanged func()) {
	e.onChanged = onChanged
}

func (e *Equipment) changed() {
	if e.onChanged != nil {
		e.onChanged()
	}
}

func (e *Equipment) HasShieldEquipped() bool {
	slot := e.GetBySlot(foundation.SlotNameOffHand)
	return slot != nil && slot.IsShield()
}

func (e *Equipment) GetShield() *Item {
	return e.GetBySlot(foundation.SlotNameOffHand)
}

func (e *Equipment) CanUnequip(item *Item) bool {
	if item.GetEquipFlag() == foundation.FlagCurseStuck && item.GetCharges() > 0 {
		return false
	}
	return true
}

func (e *Equipment) ContainsFlag(flag foundation.ActorFlag) bool {
	for _, item := range e.slots {
		itemFlags := item.GetEquipFlag()
		if itemFlags == flag {
			return true
		}
	}
	return false
}

func (e *Equipment) GetStatModifier(stat dice_curve.Stat) int {
	modifier := 0
	for _, item := range e.slots {
		modifier += item.GetStatBonus(stat)
	}
	return modifier
}
func (e *Equipment) GetSkillModifier(skill dice_curve.SkillName) int {
	modifier := 0
	for _, item := range e.slots {
		modifier += item.GetSkillBonus(skill)
	}
	return modifier
}

func (e *Equipment) GetEncumbranceFromArmor() int {
	armor := e.GetBySlot(foundation.SlotNameArmorTorso)
	encumbrance := 0
	if armor != nil {
		encumbrance = armor.GetArmor().GetEncumbrance()
	}
	return encumbrance
}

func (e *Equipment) AfterTurn() {
	for _, item := range e.slots {
		item.AfterEquippedTurn()
	}
}

func (e *Equipment) GetAllFlags() map[foundation.ActorFlag]int {
	flags := make(map[foundation.ActorFlag]int)
	for _, item := range e.slots {
		itemFlags := item.GetEquipFlag()
		if itemFlags == foundation.FlagNone {
			continue
		}
		flags[itemFlags] = 1
	}
	return flags
}

func (e *Equipment) GetMainHandItem() (*Item, bool) {
	item, exists := e.slots[foundation.SlotNameMainHand]
	return item, exists
}

func (e *Equipment) SwitchWeapons() {
	mainHand := e.GetBySlot(foundation.SlotNameMainHand)
	offHand := e.GetBySlot(foundation.SlotNameOffHand)
	e.slots[foundation.SlotNameMainHand] = offHand
	e.slots[foundation.SlotNameOffHand] = mainHand
	e.changed()
}

func (e *Equipment) GetMeleeWeapon() (*Item, bool) {
	item, exists := e.slots[foundation.SlotNameMainHand]
	if exists && item.IsMeleeWeapon() {
		return item, true
	}
	offHand, exists := e.slots[foundation.SlotNameOffHand]
	if exists && offHand.IsMeleeWeapon() {
		return offHand, true
	}
	return nil, false

}
