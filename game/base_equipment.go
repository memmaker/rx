package game

import (
	"RogueUI/foundation"
	"RogueUI/rpg"
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
	if slotName == foundation.SlotNameTwoHandedWeapon || slotName == foundation.SlotNameOneHandedWeapon {
		slotName = foundation.SlotNameMainHand
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
	if slotName == foundation.SlotNameTwoHandedWeapon {
		e.slots[foundation.SlotNameMainHand] = item
		delete(e.slots, foundation.SlotNameOffHand)
		return
	}

	if slotName == foundation.SlotNameOneHandedWeapon {
		e.slots[foundation.SlotNameMainHand] = item
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
	if slotName == foundation.SlotNameTwoHandedWeapon {
		e.unEquipBySlot(foundation.SlotNameMainHand)
		e.unEquipBySlot(foundation.SlotNameOffHand)
		return
	}
	if slotName == foundation.SlotNameOneHandedWeapon {
		slotName = foundation.SlotNameMainHand
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

type AttackType int

const (
	MeleeAttack AttackType = iota
	RangedAttack
)

func (e *Equipment) GetMainWeapon(attackType AttackType) *Item {
	isCorrectWeaponType := func(item *Item) bool {
		if attackType == MeleeAttack {
			return item.IsMeleeWeapon()
		}
		if attackType == RangedAttack {
			return item.IsRangedWeapon()
		}
		return false
	}
	if mainHand, exists := e.slots[foundation.SlotNameMainHand]; exists && isCorrectWeaponType(mainHand) {
		return mainHand
	}
	if ranged, exists := e.slots[foundation.SlotNameMissileLauncher]; exists && isCorrectWeaponType(ranged) {
		return ranged
	}
	if twoHanded, exists := e.slots[foundation.SlotNameTwoHandedWeapon]; exists && isCorrectWeaponType(twoHanded) {
		return twoHanded
	}
	if offHand, exists := e.slots[foundation.SlotNameOffHand]; exists && isCorrectWeaponType(offHand) {
		return offHand
	}

	return nil
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

func (e *Equipment) GetItemsToReplace(item *Item) []*Item {
	var items []*Item
	if item.SlotName() == foundation.SlotNameOffHand {
		mainHandItem, hasItemInMainHand := e.slots[foundation.SlotNameMainHand]
		if hasItemInMainHand && mainHandItem.IsTwoHandedWeapon() {
			items = appendIfNotNil(items, e.slots[foundation.SlotNameMainHand])
			items = appendIfNotNil(items, e.slots[foundation.SlotNameOffHand])
			return items
		}
		items = appendIfNotNil(items, e.slots[foundation.SlotNameOffHand])
		return items
	}
	if item.SlotName() == foundation.SlotNameTwoHandedWeapon {
		items = appendIfNotNil(items, e.slots[foundation.SlotNameMainHand])
		items = appendIfNotNil(items, e.slots[foundation.SlotNameOffHand])
		return items
	}
	if item.SlotName() == foundation.SlotNameOneHandedWeapon {
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

func (e *Equipment) HasMissileLauncherEquipped() bool {
	slotMain := e.GetBySlot(foundation.SlotNameMissileLauncher)
	return slotMain != nil && slotMain.IsRangedWeapon()
}

func (e *Equipment) HasMeleeWeaponEquipped() bool {
	slotMain := e.GetBySlot(foundation.SlotNameMainHand)
	slotOff := e.GetBySlot(foundation.SlotNameOffHand)
	return (slotMain != nil && slotMain.IsMeleeWeapon()) || (slotOff != nil && slotOff.IsMeleeWeapon())
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

func (e *Equipment) HasThrowableQuivered() bool {
	slot := e.GetBySlot(foundation.SlotNameQuiver)
	return slot != nil && slot.IsThrowable()
}

func (e *Equipment) GetNextQuiveredThrowable() *Item {
	slot := e.GetBySlot(foundation.SlotNameQuiver)
	if slot == nil {
		return nil
	}
	return slot
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
	return true
}
func (e *Equipment) GetFlagsCombined() uint32 {
	var flags uint32
	for _, item := range e.slots {
		flags |= item.GetEquipFlag()
	}
	return flags
}

func (e *Equipment) ContainsFlag(flag uint32) bool {
	for _, item := range e.slots {
		itemFlags := item.GetEquipFlag()
		if itemFlags&flag != 0 {
			return true
		}
	}
	return false
}

func (e *Equipment) GetStatModifier(stat rpg.Stat) int {
	modifier := 0
	for _, item := range e.slots {
		modifier += item.GetStatBonus(stat)
	}
	return modifier
}
func (e *Equipment) GetSkillModifier(skill rpg.SkillName) int {
	modifier := 0
	for _, item := range e.slots {
		modifier += item.GetSkillBonus(skill)
	}
	return modifier
}

func (e *Equipment) GetEncumbranceFromArmor() rpg.Encumbrance {
	armor := e.GetBySlot(foundation.SlotNameArmorTorso)
	encumbrance := rpg.EncumbranceNone
	if armor != nil {
		encumbrance = armor.GetArmor().GetEncumbrance()
	}
	return encumbrance
}

func (e *Equipment) GetMissileLauncher() *Item {
	return e.GetBySlot(foundation.SlotNameMissileLauncher)
}

func (e *Equipment) HasMissileLauncherEquippedForMissile(missile *WeaponInfo) bool {
	launcher := e.GetMissileLauncher()
	if launcher == nil {
		return false
	}
	return missile.IsLaunchedWith(launcher.GetWeapon().GetWeaponType())
}

func (e *Equipment) GetNextQuiveredMissile() *Item {
	slot := e.GetBySlot(foundation.SlotNameQuiver)
	if slot == nil {
		return nil
	}
	return slot
}

func (e *Equipment) HasMissileQuivered() bool {
	slot := e.GetBySlot(foundation.SlotNameQuiver)
	return slot != nil && slot.IsMissile()
}

func (e *Equipment) IsQuiveredItem(item *Item) bool {
	slot := e.GetBySlot(foundation.SlotNameQuiver)
	return slot == item
}
