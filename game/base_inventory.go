package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
	"bytes"
	"cmp"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"slices"
)

type Inventory struct {
	items          []*Item
	maxItemStacks  int
	onChanged      func()
	onBeforeRemove func(*Item)
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
		items:         make([]*Item, 0),
		maxItemStacks: maxItemStacks,
		getCarrierPos: position,
	}
}
func (i *Inventory) SetOnBeforeRemove(onBeforeRemove func(*Item)) {
	i.onBeforeRemove = onBeforeRemove
}
func (i *Inventory) Items() []*Item {
	return i.items
}

type InventoryStack struct {
	items    []*Item
	invIndex int
}

func (i InventoryStack) IsMultipleStacks() bool {
	return len(i.items) > 1 || i.First().IsMultipleStacks()
}

func (i InventoryStack) GetStackSize() int {
	if len(i.items) == 1 && i.First().IsMultipleStacks() {
		return i.First().GetStackSize()
	}
	return len(i.items)
}

func (i InventoryStack) GetCarryWeight() int {
	return i.First().GetCarryWeight() * len(i.items)
}

func (i InventoryStack) GetIcon() textiles.TextIcon {
	return i.items[0].GetIcon()
}

func (i InventoryStack) LongNameWithColors(colorCode string) string {
	return appendStacks(i.items[0].LongNameWithColors(colorCode), len(i.items))
}

func (i InventoryStack) Position() geometry.Point {
	return i.items[0].Position()
}

func (i InventoryStack) Shortcut() rune {
	return foundation.ShortCutFromIndex(i.invIndex)
}
func (i InventoryStack) GetCategory() foundation.ItemCategory {
	return i.items[0].GetCategory()
}
func (i InventoryStack) DisplayLength() int {
	return cview.TaggedStringWidth(i.InventoryNameWithColorsAndShortcut(""))

}
func (i InventoryStack) InventoryNameWithColors(lineColor string) string {
	return appendStacks(i.items[0].InventoryNameWithColors(lineColor), len(i.items))
}

func (i InventoryStack) InventoryNameWithColorsAndShortcut(lineColor string) string {
	return fmt.Sprintf("%c - %s", i.Shortcut(), appendStacks(i.items[0].InventoryNameWithColors(lineColor), len(i.items)))
}
func (i InventoryStack) GetListInfo() string {
	return appendStacks(i.items[0].GetListInfo(), len(i.items))
}
func (i InventoryStack) Name() string {
	return appendStacks(i.items[0].Name(), len(i.items))
}

func appendStacks(name string, stackCount int) string {
	if stackCount > 1 {
		name = fmt.Sprintf("%s (x%d)", name, stackCount)
	}
	return name
}

func (i InventoryStack) String() string {
	return i.Name()
}

func (i InventoryStack) Color() color.RGBA {
	return i.items[0].Color()
}

func (i InventoryStack) Counter() int {
	return len(i.items)
}

func (i InventoryStack) IsUsableOrZappable() bool {
	return i.items[0].IsUsableOrZappable()
}
func (i InventoryStack) IsEquippable() bool {
	return i.items[0].IsEquippable()
}

func (i InventoryStack) First() *Item {
	return i.items[0]

}

func (i InventoryStack) GetItems() []*Item {
	return i.items
}
func (i *Inventory) StackedItems() []*InventoryStack {
	return StackedItemsWithFilter(i.items, func(item *Item) bool { return true })
}
func StackedItemsWithFilter(items []*Item, filter func(*Item) bool) []*InventoryStack {
	if len(items) == 0 {
		return []*InventoryStack{}
	}
	stacks := make([]*InventoryStack, 0)
	for _, item := range items {
		found := false
		for stackIndex, stack := range stacks {
			if stack.items[0].CanStackWith(item) {
				stack.items = append(stack.items, item)
				stacks[stackIndex] = stack
				found = true
				break
			}
		}
		if !found {
			stacks = append(stacks, &InventoryStack{items: []*Item{item}})
		}
	}

	SortInventory(stacks)

	for invIndex, stack := range stacks {
		stack.invIndex = invIndex
	}

	for index := len(stacks) - 1; index >= 0; index-- {
		curStack := stacks[index]
		if !filter(curStack.First()) {
			stacks = append(stacks[:index], stacks[index+1:]...)
		}
	}

	return stacks
}

func (i *Inventory) RemoveItem(item *Item) {
	defer i.changed()
	for idx, invItem := range i.items {
		if invItem == item {
			i.beforeRemove(item)
			i.items = append(i.items[:idx], i.items[idx+1:]...)
			return
		}
	}
}

func (i *Inventory) Has(item *Item) bool {
	for _, invItem := range i.items {
		if invItem == item {
			return true
		}
	}
	return false
}

func (i *Inventory) AddItem(item *Item) {
	defer i.changed()
	i.addItemInternally(item)
}

func (i *Inventory) addItemInternally(item *Item) {
	if item.IsStackingWithCharges() {
		for _, invItem := range i.items {
			if invItem.CanStackWith(item) {
				invItem.SetCharges(invItem.GetCharges() + item.GetCharges())
				return
			}
		}
	}

	item.SetPositionHandler(i.getCarrierPos)
	i.items = append(i.items, item)
}

func (i *Inventory) IsEmpty() bool {
	return len(i.items) == 0
}

func (i *Inventory) IsFull() bool {
	return len(i.StackedItems()) == i.maxItemStacks
}

func (i *Inventory) SetOnChangeHandler(onChanged func()) {
	i.onChanged = onChanged
}

func (i *Inventory) changed() {
	if i.onChanged != nil {
		i.onChanged()
	}
}

func (i *Inventory) beforeRemove(item *Item) {
	item.SetPositionHandler(nil)
	if i.onBeforeRemove != nil {
		i.onBeforeRemove(item)
	}
}

func (i *Inventory) RemoveAndGetNextInStack(item *Item) *Item {
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

func (i *Inventory) RemoveAmmoByCaliber(ammo int, neededBullets int) *Item {
	for _, invItem := range i.items {
		if invItem.IsAmmoOfCaliber(ammo) {
			availableBullets := invItem.GetCharges()
			if availableBullets > neededBullets {
				splitBullets := invItem.Split(neededBullets)
				invItem.SetCharges(availableBullets - neededBullets)
				return splitBullets
			} else {
				i.RemoveItem(invItem)
				return invItem
			}
			break
		}
	}
	return nil
}

func (i *Inventory) RemoveAmmoByName(name string, amount int) *Item {
	for _, invItem := range i.items {
		if invItem.GetInternalName() == name {
			availableBullets := invItem.GetCharges()
			if availableBullets > amount {
				splitBullets := invItem.Split(amount)
				invItem.SetCharges(availableBullets - amount)
				return splitBullets
			} else {
				i.RemoveItem(invItem)
				return invItem
			}
			break
		}
	}
	return nil

}

func (i *Inventory) HasAmmo(caliber int, name string) bool {
	for _, invItem := range i.items {
		if invItem.IsAmmoOfCaliber(caliber) && invItem.GetInternalName() == name {
			return true
		}
	}
	return false
}

func (i *Inventory) GetPointer() *[]*Item {
	return &i.items
}

func (i *Inventory) GetLockpickCount() int {
	count := 0
	for _, invItem := range i.items {
		if invItem.IsLockpick() {
			count += invItem.GetCharges()
		}
	}
	return count
}

func (i *Inventory) RemoveLockpick() {
	for _, invItem := range i.items {
		if invItem.IsLockpick() {
			invItem.ConsumeCharge()
			if invItem.GetCharges() == 0 {
				i.RemoveItem(invItem)
			}
			break
		}
	}
}

func (i *Inventory) HasKey(identifier string) bool {
	for _, invItem := range i.items {
		if invItem.IsKey() && invItem.GetLockFlag() == identifier {
			return true
		}
	}
	return false
}

func (i *Inventory) RemoveItemByName(itemName string) *Item {
	for _, invItem := range i.items {
		if invItem.GetInternalName() == itemName {
			i.RemoveItem(invItem)
			return invItem
		}
	}
	return nil
}

func (i *Inventory) GetBestWeapon() *Item {
	maxDamage := 0
	var bestWeapon *Item
	for _, invItem := range i.items {
		if invItem.IsWeapon() {
			damage := invItem.GetWeaponDamage().ExpectedValue()
			if damage > maxDamage {
				maxDamage = damage
				bestWeapon = invItem
			}
		}
	}
	return bestWeapon
}

func (i *Inventory) GetBestRangedWeapon() *Item {
	maxDamage := 0
	var bestWeapon *Item
	for _, invItem := range i.items {
		if invItem.IsRangedWeapon() {
			damage := invItem.GetWeaponDamage().ExpectedValue()
			if damage > maxDamage {
				maxDamage = damage
				bestWeapon = invItem
			}
		}
	}
	return bestWeapon
}

func (i *Inventory) HasStealableItems(isStealable func(item *Item) bool) bool {
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

func (i *Inventory) GetItemByName(name string) *Item {
	for _, invItem := range i.items {
		if invItem.GetInternalName() == name {
			return invItem
		}
	}
	return nil
}

func (i *Inventory) GetSkillModifiersFromItems(skill special.Skill) []special.Modifier {
	var modifiers []special.Modifier
	for _, invItem := range i.StackedItems() {
		if invItem.First().skill == skill && invItem.First().skillBonus != 0 {
			modifiers = append(modifiers, special.DefaultModifier{
				Source:    invItem.Name(),
				Modifier:  invItem.First().GetSkillBonus(skill),
				Order:     0,
				IsPercent: true,
			})
		}
	}
	return modifiers
}

func (i *Inventory) GetStatModifiersFromItems(stat special.Stat) []special.Modifier {
	var modifiers []special.Modifier
	for _, invItem := range i.StackedItems() {
		if invItem.First().stat == stat && invItem.First().statBonus != 0 {
			modifiers = append(modifiers, special.DefaultModifier{
				Source:   invItem.Name(),
				Modifier: invItem.First().GetStatBonus(stat),
				Order:    0,
			})
		}
	}
	return modifiers
}

func (i *Inventory) HasSkillModifier(skill special.Skill) bool {
	return len(i.GetSkillModifiersFromItems(skill)) > 0
}

func (i *Inventory) StackedItemsWithFilter(filter func(item *Item) bool) []*InventoryStack {
	return StackedItemsWithFilter(i.items, filter)
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

func (i *Inventory) RemoveItemsByNameAndCount(name string, count int) []*Item {
	itemsToRemove := make([]*Item, 0)
	splitItems := make([]*Item, 0)
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
		i.RemoveItem(item)
	}
	return append(itemsToRemove, splitItems...)
}

func (i *Inventory) AddItems(player []*Item) {
	for _, item := range player {
		i.addItemInternally(item)
	}
	i.changed()
}

func SortInventory(stacks []*InventoryStack) {
	slices.SortStableFunc(stacks, func(i, j *InventoryStack) int {
		itemI := i.items[0]
		itemJ := j.items[0]
		if itemI.GetCategory() != itemJ.GetCategory() {
			return cmp.Compare(itemI.GetCategory(), itemJ.GetCategory())
		}
		if itemI.IsWeapon() && itemJ.IsWeapon() {
			if itemI.GetWeapon().GetWeaponType() != itemJ.GetWeapon().GetWeaponType() {
				return cmp.Compare(itemI.GetWeapon().GetWeaponType(), itemJ.GetWeapon().GetWeaponType())
			}
			expectedDamageI := itemI.GetWeaponDamage().ExpectedValue()
			expectedDamageJ := itemJ.GetWeaponDamage().ExpectedValue()
			return cmp.Compare(expectedDamageJ, expectedDamageI)
		}
		if itemI.IsArmor() && itemJ.IsArmor() {
			return cmp.Compare(itemI.GetArmor().GetProtectionRating(), itemJ.GetArmor().GetProtectionRating())
		}
		return cmp.Compare(itemI.Name(), itemJ.Name())
	})
}

func StacksFromItems(items []*Item) []*InventoryStack {
	stacks := make([]*InventoryStack, 0)
	for _, item := range items {
		stacks = append(stacks, &InventoryStack{items: []*Item{item}})
	}
	return stacks
}
