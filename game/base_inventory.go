package game

import (
	"bytes"
	"cmp"
	"contractor/d100"
	"contractor/foundation"
	"contractor/gridmap"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/geometry"
	"maps"
	"slices"
)

type Inventory struct {
	items          map[gridmap.ItemID]foundation.Item
	maxItemStacks  int
	displayName    string
	onChanged      func()
	onBeforeRemove func(equippableItem foundation.Equippable)
	getCarrierPos  func() geometry.Point

	equipSlots map[foundation.EquipSlot]gridmap.ItemID
}

func (i *Inventory) GobEncode() ([]byte, error) {
	buffer := &bytes.Buffer{}
	gobber := gob.NewEncoder(buffer)

	if err := gobber.Encode(i.items); err != nil {
		return nil, err
	}

	if err := gobber.Encode(i.maxItemStacks); err != nil {
		return nil, err
	}

	if err := gobber.Encode(i.displayName); err != nil {
		return nil, err
	}

	if err := gobber.Encode(i.equipSlots); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (i *Inventory) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	gobber := gob.NewDecoder(buffer)

	if err := gobber.Decode(&i.items); err != nil {
		return err
	}

	if err := gobber.Decode(&i.maxItemStacks); err != nil {
		return err
	}

	if err := gobber.Decode(&i.displayName); err != nil {
		return err
	}

	if err := gobber.Decode(&i.equipSlots); err != nil {
		return err
	}

	return nil
}

func (i *Inventory) Name() string {
	return i.displayName
}

func NewInventory(maxItemStacks int) *Inventory {
	return &Inventory{
		items:         make(map[gridmap.ItemID]foundation.Item),
		maxItemStacks: maxItemStacks,
		equipSlots:    make(map[foundation.EquipSlot]gridmap.ItemID),
	}
}

func (i *Inventory) SetCarrierPosition(position func() geometry.Point) {
	i.getCarrierPos = position
}

func (i *Inventory) EnsurePositionHandlers() {
	for _, item := range i.items {
		item.SetPositionHandler(i.getCarrierPos)
	}
}

func (i *Inventory) SetName(name string) {
	i.displayName = name
}
func (i *Inventory) SetOnBeforeRemove(onBeforeRemove func(equippable foundation.Equippable)) {
	i.onBeforeRemove = onBeforeRemove
}
func (i *Inventory) GetItems() []foundation.Item {
	var items []foundation.Item
	for v := range maps.Values(i.items) {
		items = append(items, v.(foundation.Item))
	}
	return StackedFilteredAndSortedItems(items, func(item foundation.Item) bool {
		return true
	})
}

func StackedFilteredAndSortedItems(items []foundation.Item, filter func(foundation.Item) bool) []foundation.Item {
	if len(items) == 0 {
		return []foundation.Item{}
	}
	stacks := make([]foundation.Item, 0)
	for _, item := range items {
		if !filter(item) {
			continue
		}
		found := false
		for stackIndex, stack := range stacks {
			if stack.CanStackWith(item) {
				stack.AddStacks(item)
				stacks[stackIndex] = stack
				found = true
				break
			}
		}
		if !found {
			stacks = append(stacks, item)
		}
	}

	SortInventory(stacks)

	for i, stack := range stacks {
		stack.SetInventoryIndex(i)
	}

	return stacks
}

func (i *Inventory) RemoveItem(item foundation.Item) {
	defer i.changed()
	i.removeItemInternal(item)
}

func (i *Inventory) removeItemInternal(item foundation.Item) {
	i.beforeRemove(item)
	delete(i.items, item.ID())
}

func (i *Inventory) Has(item foundation.Item) bool {
	for _, invItem := range i.items {
		if invItem == item {
			return true
		}
	}
	return false
}

func (i *Inventory) AddItem(item foundation.Item) {
	defer i.changed()
	i.addItemInternally(item)
}

func (i *Inventory) addItemInternally(item foundation.Item) {
	for _, invItem := range i.items {
		if invItem.CanStackWith(item) {
			invItem.AddStacks(item)
			return
		}
	}

	item.SetPositionHandler(i.getCarrierPos)
	i.items[item.ID()] = item
}

func (i *Inventory) IsEmpty() bool {
	return len(i.items) == 0
}

func (i *Inventory) IsFull() bool {
	return len(i.items) == i.maxItemStacks
}

func (i *Inventory) SetOnChangeHandler(onChanged func()) {
	i.onChanged = onChanged
}

func (i *Inventory) changed() {
	if i.onChanged != nil {
		i.onChanged()
	}
}

func (i *Inventory) beforeRemove(item foundation.Item) {
	item.SetPositionHandler(nil)
	if i.onBeforeRemove != nil {
		i.onBeforeRemove(item)
	}
}

func (i *Inventory) RemoveAndGetNextInStack(item *GenericItem) foundation.Item {
	i.RemoveItem(item)
	for _, invItem := range i.items {
		if invItem.CanStackWith(item) {
			return invItem
		}
	}
	return nil
}

func (i *Inventory) HasItemWithName(internalName string) bool {
	for _, invItem := range i.items {
		if invItem.GetInternalName() == internalName {
			return true
		}
	}
	return false
}

func (i *Inventory) RemoveAmmoByCaliber(caliberIndex int, neededBullets int) *Ammo {
	for _, invItem := range i.items {
		ammo, isAmmo := invItem.(*Ammo)
		if !isAmmo {
			continue
		}
		if ammo.IsAmmoOfCaliber(caliberIndex) {
			availableBullets := invItem.GetStackSize()
			if availableBullets > neededBullets {
				splitBullets := ammo.Split(neededBullets)
				return splitBullets.(*Ammo)
			} else {
				i.RemoveItem(ammo)
				return ammo
			}
			break
		}
	}
	return nil
}

func (i *Inventory) RemoveAmmoByName(name string, amount int) *Ammo {
	for _, invItem := range i.items {
		ammo, isAmmo := invItem.(*Ammo)
		if !isAmmo {
			continue
		}
		if ammo.GetInternalName() == name {
			availableBullets := ammo.GetStackSize()
			if availableBullets > amount {
				splitBullets := ammo.Split(amount)
				i.changed()
				return splitBullets.(*Ammo)
			} else {
				i.RemoveItem(ammo)
				return ammo
			}
			break
		}
	}
	return nil

}

func (i *Inventory) HasAmmo(caliber int, name string) bool {
	for _, invItem := range i.items {
		ammo, isAmmo := invItem.(*Ammo)
		if !isAmmo {
			continue
		}
		if ammo.IsAmmoOfCaliber(caliber) && ammo.GetInternalName() == name {
			return true
		}
	}
	return false
}

type LockType bool

const (
	LockTypeMechanical LockType = false
	LockTypeElectronic LockType = true
)

func (l LockType) String() string {
	if l == LockTypeMechanical {
		return "mechanical"
	}
	return "electronic"
}

func (i *Inventory) RemoveLockpicks(lock LockType, count int) {
	pickName := lockpickName(lock)
	for _, invItem := range i.items {
		if invItem.IsLockpick() && invItem.GetInternalName() == pickName {
			if invItem.GetStackSize() > count {
				invItem.RemoveStacks(count)
			} else {
				i.removeItemInternal(invItem)
			}
			break
		}
	}
	i.changed()
}

func lockpickName(lock LockType) string {
	pickName := fmt.Sprintf("%s_lockpick", lock.String())
	return pickName
}

func (i *Inventory) GetLockpickCount(lock LockType) int {
	count := 0
	pickName := lockpickName(lock)
	for _, invItem := range i.items {
		if invItem.IsLockpick() && invItem.GetInternalName() == pickName {
			count += invItem.GetStackSize()
		}
	}
	return count
}
func (i *Inventory) HasKey(identifier string) bool {
	for _, invItem := range i.items {
		if invItem.IsKey() && invItem.GetLockFlag() == identifier {
			return true
		}
	}
	return false
}

func (i *Inventory) RemoveItemByName(itemName string) foundation.Item {
	for _, invItem := range i.items {
		if invItem.GetInternalName() == itemName {
			i.RemoveItem(invItem)
			return invItem
		}
	}
	return nil
}

func (i *Inventory) GetBestWeapon() *Weapon {
	maxDamage := 0
	var bestWeapon *Weapon
	for _, invItem := range i.items {
		if invItem.IsWeapon() {
			wep := invItem.(*Weapon)
			damage := wep.GetWeaponDamage().ExpectedValue()
			if damage > maxDamage {
				maxDamage = damage
				bestWeapon = wep
			}
		}
	}
	return bestWeapon
}

func (i *Inventory) GetBestRangedWeapon() *Weapon {
	maxDamage := 0
	var bestWeapon *Weapon
	for _, invItem := range i.items {
		if invItem.IsRangedWeapon() {
			weapon := invItem.(*Weapon)
			damage := weapon.GetWeaponDamage().ExpectedValue()
			if damage > maxDamage {
				maxDamage = damage
				bestWeapon = weapon
			}
		}
	}
	return bestWeapon
}

func (i *Inventory) GetBestMeleeWeapon() *Weapon {
	maxDamage := 0
	var bestWeapon *Weapon
	for _, invItem := range i.items {
		if invItem.IsMeleeWeapon() {
			weapon := invItem.(*Weapon)
			damage := weapon.GetWeaponDamage().ExpectedValue()
			if damage > maxDamage {
				maxDamage = damage
				bestWeapon = weapon
			}
		}
	}
	return bestWeapon
}

func (i *Inventory) HasStealableItems(isStealable func(item foundation.Equippable) bool) bool {
	for _, invItem := range i.items {
		if isStealable(invItem) {
			return true
		}
	}
	return false
}

func (i *Inventory) HasLightSource() bool {
	for _, invItem := range i.items {
		if invItem.IsLightSource() {
			return true
		}
	}
	return false
}

func (i *Inventory) GetTotalWeight() int {
	totalWeight := 0
	for _, invItem := range i.items {
		totalWeight += invItem.GetCarryWeight()
	}
	return totalWeight
}

func (i *Inventory) GetAmmoWeight() int {
	totalWeight := 0
	for _, invItem := range i.items {
		if invItem.IsAmmo() {
			totalWeight += invItem.GetCarryWeight()
		}
	}
	return totalWeight
}

func (i *Inventory) GetNonAmmoWeight() int {
	totalWeight := 0
	for _, invItem := range i.items {
		if !invItem.IsAmmo() {
			totalWeight += invItem.GetCarryWeight()
		}
	}
	return totalWeight
}

func (i *Inventory) GetItemByName(name string) foundation.Item {
	for _, invItem := range i.items {
		if invItem.GetInternalName() == name {
			return invItem
		}
	}
	return nil
}

func (i *Inventory) GetSkillModifiersFromItems(skill d100.Skill) []d100.Modifier {
	var modifiers []d100.Modifier
	for _, invItem := range i.items {
		first := invItem
		if first.IsConsumable() || first.IsSkillBook() || first.IsEquippable() {
			continue
		}
		if modValue, hasValue := first.GetSkillMod(skill); hasValue {
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    invItem.Name(),
				Modifier:  modValue,
				Order:     0,
				IsPercent: true,
			})
		}
	}
	return modifiers
}

func (i *Inventory) GetStatModifiersFromItems(stat d100.Stat) []d100.Modifier {
	var modifiers []d100.Modifier
	for _, invItem := range i.items {
		first := invItem
		if first.IsConsumable() || first.IsSkillBook() || first.IsEquippable() {
			continue
		}
		if modValue, hasValue := first.GetStatMod(stat); hasValue {
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:   invItem.Name(),
				Modifier: modValue,
				Order:    0,
			})
		}
	}
	return modifiers
}

func (i *Inventory) GetDerivedStatModifiersFromItems(stat d100.DerivedStat) []d100.Modifier {
	var modifiers []d100.Modifier
	for _, invItem := range i.items {
		first := invItem
		if first.IsConsumable() || first.IsSkillBook() || first.IsEquippable() {
			continue
		}
		if modValue, hasValue := first.GetDerivedStatMod(stat); hasValue {
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:   invItem.Name(),
				Modifier: modValue,
				Order:    0,
			})
		}
	}
	return modifiers
}

func (i *Inventory) HasSkillModifier(skill d100.Skill) bool {
	return len(i.GetSkillModifiersFromItems(skill)) > 0
}

func (i *Inventory) StackedItemsWithFilter(filter func(item foundation.Item) bool) []foundation.Item {
	var items []foundation.Item
	for v := range maps.Values(i.items) {
		items = append(items, v.(foundation.Item))
	}
	return StackedFilteredAndSortedItems(items, filter)
}

func (i *Inventory) HasWeapon() bool {
	for _, invItem := range i.items {
		if invItem.IsWeapon() {
			return true
		}
	}
	return false
}

func (i *Inventory) HasWatch() bool {
	for _, invItem := range i.items {
		if invItem.IsWatch() {
			return true
		}
	}
	return false
}

func (i *Inventory) RemoveItemsByNameAndCount(name string, count int) []foundation.Item {
	defer i.changed()

	itemsToRemove := make([]foundation.Item, 0)
	splitItems := make([]foundation.Item, 0)
	for _, invItem := range i.items {
		if invItem.GetInternalName() == name {
			if invItem.IsMultipleStacks() && invItem.GetStackSize() > count {
				splitItems = append(splitItems, invItem.Split(count))
				count = 0
			} else {
				itemsToRemove = append(itemsToRemove, invItem)
				count -= invItem.GetStackSize()
			}
			if count <= 0 {
				break
			}
		}
	}

	for _, item := range itemsToRemove {
		i.removeItemInternal(item)
	}
	return append(itemsToRemove, splitItems...)
}

func (i *Inventory) HasItemWithNameAndCount(name string, count int) bool {
	for _, invItem := range i.items {
		if invItem.GetInternalName() == name {
			count -= invItem.GetStackSize()
			if count <= 0 {
				return true
			}
		}
	}
	return false
}
func (i *Inventory) AddItems(player []foundation.Item) {
	for _, item := range player {
		i.addItemInternally(item)
	}
	i.changed()
}

func (i *Inventory) HasAmmoWithCaliber(caliber int) bool {
	for _, invItem := range i.items {
		if invItem.IsAmmo() {
			ammo := invItem.(*Ammo)
			if ammo.IsAmmoOfCaliber(caliber) {
				return true
			}
		}
	}
	return false
}

func (i *Inventory) HasExactlyOneRangedWeapon() bool {
	weaponCount := 0
	for _, invItem := range i.items {
		if invItem.IsRangedWeapon() {
			weaponCount++
		}
	}
	return weaponCount == 1
}

func (i *Inventory) HasArmorEquipped() bool {
	armorID, hasArmor := i.equipSlots[foundation.SlotNameArmorTorso]
	_, hasInInventory := i.items[armorID]

	if !hasArmor || !hasInInventory {
		return false
	}
	return true
}

func (i *Inventory) GetArmor() *Armor {
	armorIndex, hasArmor := i.equipSlots[foundation.SlotNameArmorTorso]
	armor, exists := i.items[armorIndex]
	if !hasArmor || !exists {
		return nil
	}
	asArmor, canCast := armor.(*Armor)
	if !canCast {
		return nil
	}
	return asArmor
}

func (i *Inventory) GetEquippedWeapon() (*Weapon, bool) {
	weaponID, hasWeapon := i.equipSlots[foundation.SlotNameMainHand]
	wep, exists := i.items[weaponID]
	if !hasWeapon || !exists {
		return nil, false
	}
	asWeapon, canCast := wep.(*Weapon)
	if !canCast {
		return nil, false
	}
	return asWeapon, true
}

func (i *Inventory) IsEquipped(item foundation.Equippable) bool {
	slot := slotFromItem(item)
	itemID, exists := i.equipSlots[slot]
	foundItem, inInv := i.items[itemID]

	if !exists || !inInv {
		return false
	}
	return foundItem == item
}

func (i *Inventory) CanUnequip(item foundation.Item) bool {
	return true
}

func (i *Inventory) CanEquip(item foundation.Item) bool {
	targetSlot := slotFromItem(item.(foundation.Equippable))
	itemInSlot, exists := i.equipSlots[targetSlot]
	if exists && !i.CanUnequip(i.items[itemInSlot]) {
		return false
	}
	return true
}

func (i *Inventory) Equip(item foundation.Item) {
	if item == nil {
		return
	}
	defer i.changed()
	slot := slotFromItem(item.(foundation.Equippable))
	i.equipSlots[slot] = item.ID()
}

func (i *Inventory) UnEquip(item foundation.Equippable) {
	if item == nil {
		return
	}

	slotName := slotFromItem(item)

	if i.equipSlots[slotName] == item.ID() {
		i.unEquipBySlot(slotName)
		i.changed()
	}
}
func (i *Inventory) unEquipBySlot(hand foundation.EquipSlot) {
	delete(i.equipSlots, hand)
}

func (i *Inventory) GetMainHandItem() (foundation.Item, bool) {
	itemID, exists := i.equipSlots[foundation.SlotNameMainHand]
	item, inInv := i.items[itemID]
	if !exists || !inInv {
		return nil, false
	}
	return item, true
}

func (i *Inventory) GetMeleeWeapon() (*Weapon, bool) {
	weaponID, hasWeapon := i.equipSlots[foundation.SlotNameMainHand]
	wep, exists := i.items[weaponID]
	if !hasWeapon || !exists {
		return nil, false
	}
	asWeapon, canCast := wep.(*Weapon)
	if !canCast || !asWeapon.IsMeleeWeapon() {
		return nil, false
	}
	return asWeapon, true
}

func (i *Inventory) HasWeaponEquipped() bool {
	_, hasWeapon := i.GetEquippedWeapon()
	return hasWeapon
}

func (i *Inventory) HasRangedWeaponEquipped() bool {
	weapon, hasWeapon := i.GetEquippedWeapon()
	return hasWeapon && weapon.IsRangedWeapon()
}

func (i *Inventory) GetRangedWeapon() (*Weapon, bool) {
	weapon, hasWeapon := i.GetEquippedWeapon()
	if !hasWeapon || !weapon.IsRangedWeapon() {
		return nil, false
	}
	return weapon, true
}

func (i *Inventory) HasMeleeWeaponEquipped() bool {
	weapon, hasWeapon := i.GetEquippedWeapon()
	return hasWeapon && weapon.IsMeleeWeapon()
}

func (i *Inventory) GetHelmet() *Armor {
	armorID, hasArmor := i.equipSlots[foundation.SlotNameArmorHead]
	armor, exists := i.items[armorID]
	if !hasArmor || !exists {
		return nil
	}
	asArmor, canCast := armor.(*Armor)
	if !canCast {
		return nil
	}
	return asArmor
}

func (i *Inventory) IsNotEquipped(item foundation.Equippable) bool {
	return !i.IsEquipped(item)
}

func (i *Inventory) AfterTurn() {
	for _, id := range i.equipSlots {
		item := i.items[id]
		item.AfterEquippedTurn()
	}
}

func (i *Inventory) GetStatModifiersFromEquippedItems(stat d100.Stat) []d100.Modifier {
	var modifiers []d100.Modifier
	for _, itemID := range i.equipSlots {
		item := i.items[itemID]
		if modValue, hasValue := item.GetStatMod(stat); hasValue {
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    item.Name(),
				Modifier:  modValue,
				Order:     0,
				IsPercent: true,
			})
		}
	}
	slices.SortStableFunc(modifiers, func(i, j d100.Modifier) int {
		return cmp.Compare(i.Description(), j.Description())
	})
	return modifiers

}

func (i *Inventory) GetSkillModifiersFromEquippedItems(skill d100.Skill) []d100.Modifier {
	var modifiers []d100.Modifier
	for _, itemID := range i.equipSlots {
		item := i.items[itemID]
		if modValue, hasValue := item.GetSkillMod(skill); hasValue {
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    item.Name(),
				Modifier:  modValue,
				Order:     0,
				IsPercent: true,
			})
		}
	}
	slices.SortStableFunc(modifiers, func(i, j d100.Modifier) int {
		return cmp.Compare(i.Description(), j.Description())
	})
	return modifiers

}

func (i *Inventory) GetDerivedStatModifiersFromEquippedItems(stat d100.DerivedStat) []d100.Modifier {
	var modifiers []d100.Modifier
	for _, itemID := range i.equipSlots {
		item := i.items[itemID]
		if modValue, hasValue := item.GetDerivedStatMod(stat); hasValue {
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    item.Name(),
				Modifier:  modValue,
				Order:     0,
				IsPercent: true,
			})
		}
	}
	slices.SortStableFunc(modifiers, func(i, j d100.Modifier) int {
		return cmp.Compare(i.Description(), j.Description())
	})
	return modifiers

}

func (i *Inventory) HasArmorWithNameEquipped(name string) bool {
	armor := i.GetArmor()
	if armor == nil {
		return false
	}
	return armor.GetInternalName() == name
}

func (i *Inventory) GetAllEquipmentFlags() map[foundation.ActorFlag]int {
	flags := make(map[foundation.ActorFlag]int)
	for _, id := range i.equipSlots {
		item := i.items[id]
		itemFlags := item.GetEquipFlag()
		if itemFlags == foundation.FlagNone {
			continue
		}
		flags[itemFlags] = 1
	}
	return flags
}

func (i *Inventory) EquipmentContainsFlag(flag foundation.ActorFlag) bool {
	for _, id := range i.equipSlots {
		item := i.items[id]
		itemFlags := item.GetEquipFlag()
		if itemFlags == flag {
			return true
		}
	}
	return false
}
func (i *Inventory) GetEncumbranceFromArmor() int {
	armor := i.GetArmor()
	encumbrance := 0
	if armor != nil {
		encumbrance = armor.GetEncumbrance()
	}
	return encumbrance
}

func SortInventory(stacks []foundation.Item) {
	slices.SortStableFunc(stacks, func(i, j foundation.Item) int {
		itemI := i
		itemJ := j
		if itemI.GetCategory() != itemJ.GetCategory() {
			return cmp.Compare(itemI.GetCategory(), itemJ.GetCategory())
		}
		if itemI.IsWeapon() && itemJ.IsWeapon() {
			weapI := itemI.(*Weapon)
			weapJ := itemJ.(*Weapon)
			if weapI.GetWeaponType() != weapJ.GetWeaponType() {
				return cmp.Compare(weapI.GetWeaponType(), weapJ.GetWeaponType())
			}
			expectedDamageI := weapI.GetWeaponDamage().ExpectedValue()
			expectedDamageJ := weapJ.GetWeaponDamage().ExpectedValue()
			if expectedDamageI != expectedDamageJ {
				return cmp.Compare(expectedDamageJ, expectedDamageI)
			}
		}
		if itemI.IsArmor() && itemJ.IsArmor() {
			armorI := itemI.(*Armor)
			armorJ := itemJ.(*Armor)
			if armorI.GetProtectionRating() != armorJ.GetProtectionRating() {
				return cmp.Compare(armorI.GetProtectionRating(), armorJ.GetProtectionRating())
			}
		}
		if itemI.Name() != itemJ.Name() {
			return cmp.Compare(itemI.Name(), itemJ.Name())
		}
		return cmp.Compare(itemI.GetQuality(), itemJ.GetQuality())
	})
}
func slotFromItem(item foundation.Equippable) foundation.EquipSlot {
	if item.IsHeadGear() {
		return foundation.SlotNameArmorHead
	} else if item.IsArmor() {
		return foundation.SlotNameArmorTorso
	} else if item.IsLightSource() {
		return foundation.SlotNameLightSource
	} else if item.IsWeapon() {
		return foundation.SlotNameMainHand
	}
	return foundation.SlotNameNotEquippable
}
