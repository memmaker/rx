package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
	"time"
)

type Container struct {
	*BaseObject

	Known                 bool
	ContainedItems        []foundation.Item
	show                  func(actor *Actor)
	isPlayer              func(actor *Actor) bool
	LockedFlag            string
	LockDiff              d100.Difficulty
	LockStrengthRemaining int
	NumberLock            []rune

	VendingMachine bool
	refill         func()
	LastRefillTime time.Time

	FlagRemovalOf string
	flagCall      func()
	Locked        bool
}

func (b *Container) ReduceStrength(reduction int) (int, bool) {
	if b.LockStrengthRemaining <= 0 {
		return 0, false
	}
	realReduction := int(float64(reduction) * b.LockDiff.LockReductionFactor())
	b.LockStrengthRemaining -= realReduction
	if b.LockStrengthRemaining <= 0 {
		b.Unlock()
		return realReduction, true
	}
	return realReduction, false
}

func (b *Container) Unlock() {
	b.Locked = false
	b.LockStrengthRemaining = 0
}

func (b *Container) RemainingStrength() int {
	return b.LockStrengthRemaining
}

func (b *Container) GetCategory() foundation.ObjectCategory {
	if b.Known {
		if len(b.ContainedItems) == 0 {
			return foundation.ObjectKnownEmptyContainer
		} else {
			return foundation.ObjectKnownContainer
		}
	}
	return foundation.ObjectUnknownContainer
}
func (b *Container) GetIcon() textiles.TextIcon {
	return b.iconForObject(b.GetCategory().LowerString())
}
func (b *Container) OnBump(actor *Actor) {
	if b.isPlayer(actor) {
		b.show(actor)
		b.Known = true
	}
}

func (b *Container) RemoveItem(item foundation.Item) {
	for i, containedItem := range b.ContainedItems {
		if containedItem == item {
			if b.flagCall != nil && b.FlagRemovalOf != "" && containedItem.GetInternalName() == b.FlagRemovalOf {
				b.flagCall()
			}
			b.ContainedItems = append(b.ContainedItems[:i], b.ContainedItems[i+1:]...)
			return
		}
	}
}

func (b *Container) ContainsItems() bool {
	return len(b.ContainedItems) > 0
}

func (b *Container) AddItem(item foundation.Item) {
	for _, containedItem := range b.ContainedItems {
		if containedItem.CanStackWith(item) {
			containedItem.AddStacks(item)
			return
		}
	}
	item.SetPosition(b.RawPosition)
	b.ContainedItems = append(b.ContainedItems, item)
}

func (g *GameState) NewContainer(rec recfile.Record, iconForObject func(objectType string) textiles.TextIcon) Object {
	container := &Container{
		BaseObject: &BaseObject{
			iconForObject: iconForObject,
			Category:      foundation.ObjectUnknownContainer,
		},
		LockStrengthRemaining: 100,
	}
	var randomNumberCode bool
	container.SetWalkable(false)
	container.SetHidden(false)
	container.SetTransparent(true)
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			container.InternalName = field.Value
		case "description":
			container.DisplayName = field.Value
		case "iconoverride":
			container.CustomIcon = container.iconForObject(field.Value)
		case "position":
			container.RawPosition, _ = geometry.NewPointFromEncodedString(field.Value)
		case "item":
			item := g.NewItemFromString(field.Value)
			if item != nil {
				container.AddItem(item)
			}
		case "isvendingmachine":
			container.VendingMachine = recfile.StrBool(field.Value)
		case "flag_removal_of":
			container.FlagRemovalOf = field.Value
		case "lockflag":
			container.Locked = true
			container.LockedFlag = field.Value
		case "numberlock":
			container.Locked = true
			if strings.ToLower(field.Value) == "random" {
				// random 4-digit code
				randomNumberCode = true
			} else {
				container.NumberLock = []rune(field.Value)
			}
		case "lockdifficulty":
			container.LockDiff = d100.DifficultyFromString(field.Value)
		}
	}
	if randomNumberCode {
		container.NumberLock = g.getRandomNumberCode(container.InternalName)
	}
	container.InitWithGameState(g)
	return container
}

func (b *Container) InitWithGameState(g *GameState) {
	b.isPlayer = func(actor *Actor) bool { return actor == g.Player }
	b.show = func(actor *Actor) {
		if b.Locked {
			if b.LockedFlag != "" && actor.HasKey(b.LockedFlag) {
				b.Unlock()
				g.msg(foundation.Msg("You unlocked the container using the key"))
				g.ui.PlayCue("world/PICKKEYS")
				g.endPlayerTurn(g.Player.TimeNeededForActions())
				return
			}
			if len(b.NumberLock) > 0 {
				picksNeeded := d100.ElectronicLockpicksNeeded(b.LockDiff, g.Player.GetCharSheet().GetSkill(d100.SkillForPickLocks))
				actionInfo := fmt.Sprintf("Hack (%d)", picksNeeded)
				g.ui.OpenKeypad(actionInfo, b.NumberLock, func() bool {
					success := g.playerPickElectronicLock(b.LockDiff)
					if success {
						b.Unlock()
					}
					return success
				}, func(result bool) {
					if result {
						b.Unlock()
						g.msg(foundation.Msg("You unlocked the container using the code"))
						g.endPlayerTurn(g.Player.TimeNeededForActions())
					}
				})
			} else {
				// lockpicking
				if g.Player.GetInventory().GetLockpickCount(LockTypeMechanical) == 0 {
					g.msg(foundation.Msg("You don't have any lockpicks"))
					return
				}
				if b.IsLockAtFullStrength() {
					g.ui.AskForConfirmation("Confirm", "Do you want to pick the lock?", func(result bool) {
						if result {
							g.playerTryPickLock(b)
						}
					})
				} else {
					g.playerTryPickLock(b)
				}
			}
			return
		}

		if b.VendingMachine {
			g.openVendingMachineMenu(b)
		} else {
			g.openContainer(b)
		}
	}
	if b.FlagRemovalOf != "" {
		b.flagCall = func() {
			g.gameFlags.Increment(fmt.Sprintf("ContainerRemoved(%s, %s)", b.InternalName, b.FlagRemovalOf))
		}
	}
}

func (b *Container) IsLockAtFullStrength() bool {
	return b.LockStrengthRemaining == 100
}

func (b *Container) AppendContextActions(items []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	if b.VendingMachine {
		items = append(items, foundation.MenuItem{
			Name: "Buy",
			Action: func() {
				g.openVendingMachineMenu(b)
			},
			CloseMenus: true,
		})

		hackDifficulty := d100.Hard

		picksNeeded := d100.ElectronicLockpicksNeeded(hackDifficulty, g.Player.GetCharSheet().GetSkill(d100.SkillForPickLocks))
		hackLabel := fmt.Sprintf("Hack (%d)", picksNeeded)
		items = append(items, foundation.MenuItem{
			Name: hackLabel,
			Action: func() {
				if g.playerPickElectronicLock(hackDifficulty) {
					g.openContainer(b) // TODO: Remove the money instead..
				}
			},
			CloseMenus: true,
		})
	}
	return items
}

func (b *Container) HasItemsWithName(name string, stackSize int) bool {
	for _, item := range b.ContainedItems {
		if item.GetInternalName() == name {
			stackSize -= item.GetStackSize()
			if stackSize <= 0 {
				return true
			}
		}
	}
	return stackSize <= 0
}

func (b *Container) RemoveItemsWithName(name string, count int) []foundation.Item {
	var itemsRemoved []foundation.Item
	for i := 0; i < len(b.ContainedItems); i++ {
		item := b.ContainedItems[i]
		if item.GetInternalName() == name {
			if item.GetStackSize() <= count {
				b.ContainedItems = append(b.ContainedItems[:i], b.ContainedItems[i+1:]...)
				itemsRemoved = append(itemsRemoved, item)
				count -= item.GetStackSize()
				i--
			} else {
				splitItem := item.Split(count)
				itemsRemoved = append(itemsRemoved, splitItem)
				break
			}
		}
	}
	return itemsRemoved
}

func (b *Container) Has(item foundation.Item) bool {
	for _, containedItem := range b.ContainedItems {
		if containedItem == item {
			return true
		}
	}
	return false
}

func (b *Container) AddItems(items []foundation.Item) {
	for _, item := range items {
		b.AddItem(item)
	}
}

func (b *Container) GetItems() []foundation.Item {
	return b.ContainedItems

}

func (b *Container) ItemsFiltered(keep func(item foundation.Item) bool) []foundation.Item {
	return StackedFilteredAndSortedItems(b.ContainedItems, keep)
}

func (b *Container) ItemCountByPrefix(prefix string) int {
	count := 0
	for _, item := range b.ContainedItems {
		if strings.HasPrefix(item.GetInternalName(), prefix) {
			count += item.GetStackSize()
		}
	}
	return count
}

func (b *Container) IterateItems(action func(foundation.Item)) {
	for _, item := range b.ContainedItems {
		action(item)
	}
}

func (g *GameState) openContainer(container ItemContainer) {
	containerItems := StackedFilteredAndSortedItems(container.GetItems(), func(item foundation.Item) bool { return true })
	playerItems := StackedFilteredAndSortedItems(g.Player.GetInventory().GetItems(), func(item foundation.Item) bool { return true })

	// PROBLEM: For "Take All", we are calling this function multiple times..
	// Re-Opening the container multiple times, is not a good idea.
	transferToPlayer := func(itemTaken foundation.Item, amount int) {
		itemName := itemTaken.Name()

		if amount > 0 {
			stackTransfer(container, g.Player.GetInventory(), itemTaken, amount)

			g.ui.PlayCue("world/pickup")

			g.msg(foundation.HiLite("You take %s from %s.", itemName, container.Name()))
		}

		g.openContainer(container)
	}
	transferToContainer := func(itemTaken foundation.Item, amount int) {
		itemName := itemTaken.Name()

		if amount > 0 {
			stackTransfer(g.Player.GetInventory(), container, itemTaken, amount)

			g.ui.PlayCue("world/drop")

			g.msg(foundation.HiLite("You place %s in %s.", itemName, container.Name()))
		}

		g.openContainer(container)
	}
	takeAll := func() {
		for _, item := range container.GetItems() {
			stackTransfer(container, g.Player.GetInventory(), item, item.GetStackSize())
		}
		g.openContainer(container)
	}
	g.ui.ShowGiveAndTakeContainer(g.Player.Name(), playerItems, container.Name(), containerItems, transferToPlayer, transferToContainer, takeAll)
}

type ItemContainer interface {
	AddItem(item foundation.Item)
	RemoveItem(item foundation.Item)
	GetItems() []foundation.Item
	Name() string
}

func stackTransfer(from ItemContainer, to ItemContainer, item foundation.Item, splitAmount int) {
	if splitAmount == 0 {
		return
	}

	multiItem := item
	totalAmount := multiItem.GetStackSize()

	splitAmount = min(splitAmount, totalAmount)

	if splitAmount == totalAmount {
		from.RemoveItem(multiItem)
		to.AddItem(multiItem)
		return
	}

	splitItem := multiItem.Split(splitAmount)

	to.AddItem(splitItem)
}
