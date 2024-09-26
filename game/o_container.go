package game

import (
	"RogueUI/foundation"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type Container struct {
	*BaseObject

	isKnown        bool
	containedItems []foundation.Item
	show           func()
	isPlayer       func(actor *Actor) bool
	lockFlag       string

	flagRemovalOf string
	flagCall      func()
}

func (b *Container) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := b.BaseObject.gobEncode(enc); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.isKnown); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.containedItems); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *Container) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))

	b.BaseObject = &BaseObject{}

	if err := b.BaseObject.gobDecode(dec); err != nil {
		return err
	}
	if err := dec.Decode(&b.isKnown); err != nil {
		return err
	}

	if err := dec.Decode(&b.containedItems); err != nil {
		return err
	}

	return nil
}

func (b *Container) GetCategory() foundation.ObjectCategory {
	if b.isKnown {
		if len(b.containedItems) == 0 {
			return foundation.ObjectKnownEmptyContainer
		} else {
			return foundation.ObjectKnownContainer
		}
	}
	return foundation.ObjectUnknownContainer
}
func (b *Container) Icon() textiles.TextIcon {
	return b.iconForObject(b.GetCategory().LowerString())
}
func (b *Container) OnBump(actor *Actor) {
	if b.isPlayer(actor) {
		b.show()
		b.isKnown = true
	}
}

func (b *Container) RemoveItem(item foundation.Item) {
	for i, containedItem := range b.containedItems {
		if containedItem == item {
			if b.flagCall != nil && b.flagRemovalOf != "" && containedItem.InternalName() == b.flagRemovalOf {
				b.flagCall()
			}
			b.containedItems = append(b.containedItems[:i], b.containedItems[i+1:]...)
			return
		}
	}
}

func (b *Container) ContainsItems() bool {
	return len(b.containedItems) > 0
}

func (b *Container) AddItem(item foundation.Item) {
	for _, containedItem := range b.containedItems {
		if containedItem.CanStackWith(item) {
			containedItem.SetCharges(containedItem.Charges() + item.Charges())
			return
		}
	}
	item.SetPosition(b.position)
	b.containedItems = append(b.containedItems, item)
}

func (g *GameState) NewContainer(rec recfile.Record) Object {
	container := &Container{
		BaseObject: &BaseObject{
			category: foundation.ObjectUnknownContainer,
			isAlive:  true,
		},
	}
	container.SetWalkable(false)
	container.SetHidden(false)
	container.SetTransparent(true)
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			container.internalName = field.Value
		case "description":
			container.displayName = field.Value
		case "position":
			container.position, _ = geometry.NewPointFromEncodedString(field.Value)
		case "item":
			item := g.NewItemFromString(field.Value)
			if item != nil {
				container.AddItem(item)
			}
		case "flag_removal_of":
			container.flagRemovalOf = field.Value
		case "lockflag":
			container.lockFlag = field.Value
		}
	}
	container.InitWithGameState(g)
	return container
}

func (b *Container) InitWithGameState(g *GameState) {
	b.iconForObject = g.iconForObject
	b.isPlayer = func(actor *Actor) bool { return actor == g.Player }
	b.show = func() {
		g.openContainer(b)
	}
	if b.flagRemovalOf != "" {
		b.flagCall = func() {
			g.gameFlags.Increment(fmt.Sprintf("ContainerRemoved(%s, %s)", b.internalName, b.flagRemovalOf))
		}
	}
}

func (b *Container) HasItemsWithName(name string, stackSize int) bool {
	for _, item := range b.containedItems {
		if item.InternalName() == name {
			stackSize -= item.StackSize()
			if stackSize <= 0 {
				return true
			}
		}
	}
	return stackSize <= 0
}

func (b *Container) RemoveItemsWithName(name string, count int) []foundation.Item {
	var itemsRemoved []foundation.Item
	for i := 0; i < len(b.containedItems); i++ {
		item := b.containedItems[i]
		if item.InternalName() == name {
			if item.StackSize() <= count {
				b.containedItems = append(b.containedItems[:i], b.containedItems[i+1:]...)
				itemsRemoved = append(itemsRemoved, item)
				count -= item.StackSize()
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
	for _, containedItem := range b.containedItems {
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

func (g *GameState) openContainer(container *Container) {
	containerItems := StackedFilteredAndSortedItems(container.containedItems, func(item foundation.Item) bool { return true })
	playerItems := g.Player.GetInventory().Items()

	transferToPlayer := func(itemTaken foundation.Item, amount int) {
		itemName := itemTaken.Name()

		if amount > 0 {
			g.stackTransfer(container, g.Player.GetInventory(), itemTaken, amount)

			g.ui.PlayCue("world/pickup")

			g.msg(foundation.HiLite("You take %s from %s.", itemName, container.Name()))
		}

		g.openContainer(container)
	}
	transferToContainer := func(itemTaken foundation.Item, amount int) {
		itemName := itemTaken.Name()

		if amount > 0 {
			g.stackTransfer(g.Player.GetInventory(), container, itemTaken, amount)

			g.ui.PlayCue("world/drop")

			g.msg(foundation.HiLite("You place %s in %s.", itemName, container.Name()))
		}

		g.openContainer(container)
	}
	g.ui.ShowGiveAndTakeContainer(g.Player.Name(), playerItems, container.Name(), containerItems, transferToPlayer, transferToContainer)
}

type ItemContainer interface {
	AddItem(item foundation.Item)
	RemoveItem(item foundation.Item)
}

func (g *GameState) stackTransfer(from ItemContainer, to ItemContainer, item foundation.Item, splitAmount int) {
	if splitAmount == 0 {
		return
	}

	multiItem := item
	totalAmount := multiItem.StackSize()

	splitAmount = min(splitAmount, totalAmount)

	if splitAmount == totalAmount {
		from.RemoveItem(multiItem)
		to.AddItem(multiItem)
		return
	}

	splitItem := multiItem.Split(splitAmount)

	multiItem.SetCharges(totalAmount - splitAmount)

	to.AddItem(splitItem)
}
