package game

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"cmp"
	"code.rocketnine.space/tslocum/cview"
	"fmt"
	"image/color"
	"slices"
)

type Inventory struct {
	items          []*Item
	maxItemStacks  int
	onChanged      func()
	onBeforeRemove func(*Item)
}

func NewInventory(maxItemStacks int) *Inventory {
	return &Inventory{
		items:         make([]*Item, 0),
		maxItemStacks: maxItemStacks,
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
func (i *Inventory) StackedItems() []*InventoryStack {
	return i.StackedItemsWithFilter(func(item *Item) bool { return true })
}
func (i *Inventory) StackedItemsWithFilter(filter func(*Item) bool) []*InventoryStack {
	if len(i.items) == 0 {
		return []*InventoryStack{}
	}
	stacks := make([]*InventoryStack, 0)
	for _, item := range i.items {
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

func (i *Inventory) Remove(item *Item) {
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

func (i *Inventory) Add(item *Item) {
	defer i.changed()
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
	if i.onBeforeRemove != nil {
		i.onBeforeRemove(item)
	}
}

func (i *Inventory) RemoveAndGetNextInStack(item *Item) *Item {
	i.Remove(item)
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
			expectedDamageI := itemI.GetWeapon().GetDamageDice().ExpectedValue()
			expectedDamageJ := itemJ.GetWeapon().GetDamageDice().ExpectedValue()
			return cmp.Compare(expectedDamageI, expectedDamageJ)
		}
		if itemI.IsArmor() && itemJ.IsArmor() {
			return cmp.Compare(itemI.GetArmor().GetDamageResistanceWithPlus(), itemJ.GetArmor().GetDamageResistanceWithPlus())
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
