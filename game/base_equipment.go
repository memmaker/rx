package game

import (
	"RogueUI/foundation"
	"bytes"
	"encoding/gob"
)

type Equipment struct {
	slots     map[foundation.EquipSlot]foundation.Equippable
	onChanged func()
}

func (e *Equipment) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Encode each field of the struct in order
	if err := encoder.Encode(e.slots); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e *Equipment) GobDecode(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// Decode each field of the struct in order
	if err := decoder.Decode(&e.slots); err != nil {
		return err
	}

	return nil
}
func NewEquipment() *Equipment {
	return &Equipment{
		slots: make(map[foundation.EquipSlot]foundation.Equippable),
	}
}

func (e *Equipment) IsEquipped(item foundation.Equippable) bool {
	if item == nil {
		return false
	}
	slot := slotFromItem(item)
	if slot == foundation.SlotNameNotEquippable {
		return false
	}
	slotItem, exists := e.slots[slot]
	if exists && slotItem == item {
		return true
	}
	return false
}

func slotFromItem(item foundation.Equippable) foundation.EquipSlot {
	if item.IsArmor() {
		return foundation.SlotNameArmorTorso
	} else if item.IsLightSource() {
		return foundation.SlotNameLightSource
	} else if item.IsWeapon() {
		return foundation.SlotNameMainHand
	}
	return foundation.SlotNameNotEquippable
}

func (e *Equipment) Equip(item foundation.Equippable) {
	if item == nil {
		return
	}
	defer e.changed()
	slotName := slotFromItem(item)
	e.slots[slotName] = item
}

func (e *Equipment) UnEquip(item foundation.Equippable) {
	if item == nil {
		return
	}
	defer e.changed()
	slotName := slotFromItem(item)
	if e.slots[slotName] == item {
		e.unEquipBySlot(slotName)
	}
}

func (e *Equipment) unEquipBySlot(hand foundation.EquipSlot) {
	delete(e.slots, hand)
}

func (e *Equipment) GetBySlot(hand foundation.EquipSlot) foundation.Equippable {
	return e.slots[hand]
}

func (e *Equipment) GetArmor() *Armor {
	slotItem := e.GetBySlot(foundation.SlotNameArmorTorso)
	if !slotItem.IsArmor() {
		return nil
	}
	return slotItem.(*Armor)
}

func (e *Equipment) CanEquip(item foundation.Equippable) bool {
	toBeReplaced := e.GetItemsToReplace(item)
	for _, itemToRemove := range toBeReplaced {
		if !e.CanUnequip(itemToRemove) {
			return false
		}
	}
	return true
}
func (e *Equipment) GetItemsToReplace(item foundation.Equippable) []foundation.Equippable {
	var items []foundation.Equippable
	return appendIfNotNil(items, e.slots[slotFromItem(item)])
}

func (e *Equipment) isOneSlotAvailable(slotOne foundation.EquipSlot, slotTwo foundation.EquipSlot) bool {
	_, slotOneExists := e.slots[slotOne]
	_, slotTwoExists := e.slots[slotTwo]
	return !slotOneExists || !slotTwoExists
}

func appendIfNotNil(items []foundation.Equippable, item foundation.Equippable) []foundation.Equippable {
	if item != nil {
		return append(items, item)
	}
	return items
}

func (e *Equipment) HasRangedWeaponEquipped() bool {
	return e.HasRangedWeaponInMainHand()
}

func (e *Equipment) HasRangedWeaponInMainHand() bool {
	slotMain := e.GetBySlot(foundation.SlotNameMainHand)
	if slotMain != nil && slotMain.IsRangedWeapon() {
		return true
	}
	return false
}

func (e *Equipment) HasMeleeWeaponEquipped() bool {
	slotMain := e.GetBySlot(foundation.SlotNameMainHand)
	if slotMain != nil && slotMain.IsMeleeWeapon() {
		return true
	}
	return false
}

func (e *Equipment) AllItems() []foundation.Equippable {
	var items []foundation.Equippable
	for _, item := range e.slots {
		items = append(items, item)
	}
	return items
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

func (e *Equipment) CanUnequip(item foundation.Equippable) bool {
	if item.GetEquipFlag() == foundation.FlagCurseStuck && item.Charges() > 0 {
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

func (e *Equipment) GetEncumbranceFromArmor() int {
	armor := e.GetArmor()
	encumbrance := 0
	if armor != nil {
		encumbrance = armor.GetEncumbrance()
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
	for slot, item := range e.slots {
		if item == nil {
			delete(e.slots, slot)
			continue
		}
		itemFlags := item.GetEquipFlag()
		if itemFlags == foundation.FlagNone {
			continue
		}
		flags[itemFlags] = 1
	}
	return flags
}

func (e *Equipment) GetMainHandItem() (foundation.Equippable, bool) {
	item, exists := e.slots[foundation.SlotNameMainHand]
	return item, exists
}

func (e *Equipment) GetMeleeWeapon() (*Weapon, bool) {
	item, exists := e.slots[foundation.SlotNameMainHand]
	if exists && item.IsMeleeWeapon() {
		return item.(*Weapon), true
	}
	return nil, false

}

func (e *Equipment) GetRangedWeapon() (*Weapon, bool) {
	item, exists := e.slots[foundation.SlotNameMainHand]
	if exists && item.IsRangedWeapon() {
		return item.(*Weapon), true
	}
	return nil, false

}

func (e *Equipment) IsNotEquipped(item foundation.Equippable) bool {
	return !e.IsEquipped(item)
}

func (e *Equipment) HasArmorWithNameEquipped(name string) bool {
	armor := e.GetArmor()
	return armor != nil && armor.InternalName() == name
}

func (e *Equipment) GetMainHandWeapon() (*Weapon, bool) {
	item, exists := e.slots[foundation.SlotNameMainHand]
	if exists {
		weapon, isWeapon := item.(*Weapon)
		return weapon, isWeapon
	}
	return nil, false
}
