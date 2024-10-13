package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"bytes"
	"cmp"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/geometry"
	"slices"
)

type Inventory struct {
	items          []foundation.Item
	maxItemStacks  int
	onChanged      func()
	onBeforeRemove func(equippableItem foundation.Equippable)
	getCarrierPos  func() geometry.Point
}

func (i *Inventory) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Encode each field of the struct in order
	if err := encoder.Encode(i.items); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.maxItemStacks); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (i *Inventory) GobDecode(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// Decode each field of the struct in order
	if err := decoder.Decode(&i.items); err != nil {
		return err
	}
	if err := decoder.Decode(&i.maxItemStacks); err != nil {
		return err
	}

	return nil
}

func NewInventory(maxItemStacks int, position func() geometry.Point) *Inventory {
	return &Inventory{
		items:         make([]foundation.Item, 0),
		maxItemStacks: maxItemStacks,
		getCarrierPos: position,
	}
}
func (i *Inventory) SetOnBeforeRemove(onBeforeRemove func(equippable foundation.Equippable)) {
	i.onBeforeRemove = onBeforeRemove
}
func (i *Inventory) Items() []foundation.Item {
	return i.items
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
	for idx, invItem := range i.items {
		if invItem == item {
			i.beforeRemove(item)
			i.items = append(i.items[:idx], i.items[idx+1:]...)
			return
		}
	}
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
	i.items = append(i.items, item)
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
		if invItem.InternalName() == internalName {
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
			availableBullets := invItem.StackSize()
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
		if ammo.InternalName() == name {
			availableBullets := ammo.StackSize()
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
		if ammo.IsAmmoOfCaliber(caliber) && ammo.InternalName() == name {
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
		if invItem.IsLockpick() && invItem.InternalName() == pickName {
			if invItem.StackSize() > count {
				invItem.RemoveStacks(count)
				i.changed()
			} else {
				i.RemoveItem(invItem)
			}
			break
		}
	}
}

func lockpickName(lock LockType) string {
	pickName := fmt.Sprintf("%s_lockpick", lock.String())
	return pickName
}

func (i *Inventory) GetLockpickCount(lock LockType) int {
	count := 0
	pickName := lockpickName(lock)
	for _, invItem := range i.items {
		if invItem.IsLockpick() && invItem.InternalName() == pickName {
			count += invItem.StackSize()
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
		if invItem.InternalName() == itemName {
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
		if invItem.InternalName() == name {
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
	return StackedFilteredAndSortedItems(i.items, filter)
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
	itemsToRemove := make([]foundation.Item, 0)
	splitItems := make([]foundation.Item, 0)
	for _, invItem := range i.items {
		if invItem.InternalName() == name {
			if invItem.IsMultipleStacks() && invItem.StackSize() > count {
				splitItems = append(splitItems, invItem.Split(count))
				count = 0
			} else {
				itemsToRemove = append(itemsToRemove, invItem)
				count -= invItem.StackSize()
			}
			if count <= 0 {
				break
			}
		}
	}
	for _, item := range itemsToRemove {
		i.RemoveItem(item)
	}
	return append(itemsToRemove, splitItems...)
}

func (i *Inventory) HasItemWithNameAndCount(name string, count int) bool {
	for _, invItem := range i.items {
		if invItem.InternalName() == name {
			count -= invItem.StackSize()
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

func SortInventory(stacks []foundation.Item) {
	slices.SortStableFunc(stacks, func(i, j foundation.Item) int {
		itemI := i
		itemJ := j
		if itemI.Category() != itemJ.Category() {
			return cmp.Compare(itemI.Category(), itemJ.Category())
		}
		if itemI.IsWeapon() && itemJ.IsWeapon() {
			weapI := itemI.(*Weapon)
			weapJ := itemJ.(*Weapon)
			if weapI.GetWeaponType() != weapJ.GetWeaponType() {
				return cmp.Compare(weapI.GetWeaponType(), weapJ.GetWeaponType())
			}
			expectedDamageI := weapI.GetWeaponDamage().ExpectedValue()
			expectedDamageJ := weapJ.GetWeaponDamage().ExpectedValue()
			return cmp.Compare(expectedDamageJ, expectedDamageI)
		}
		if itemI.IsArmor() && itemJ.IsArmor() {
			armorI := itemI.(*Armor)
			armorJ := itemJ.(*Armor)
			return cmp.Compare(armorI.GetProtectionRating(), armorJ.GetProtectionRating())
		}
		return cmp.Compare(itemI.Name(), itemJ.Name())
	})
}
