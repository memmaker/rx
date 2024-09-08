package game

import (
	"RogueUI/foundation"
	"bytes"
	"encoding/gob"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type Container struct {
	*BaseObject

	isKnown        bool
	containedItems []*Item
	show           func()
	isPlayer       func(actor *Actor) bool
	lockFlag       string
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

func (b *Container) RemoveItem(item *Item) {
	for i, containedItem := range b.containedItems {
		if containedItem == item {
			b.containedItems = append(b.containedItems[:i], b.containedItems[i+1:]...)
			return
		}
	}
}

func (b *Container) ContainsItems() bool {
	return len(b.containedItems) > 0
}

func (b *Container) AddItem(item *Item) {
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
		case "description":
			container.displayName = field.Value
		case "position":
			container.position, _ = geometry.NewPointFromEncodedString(field.Value)
		case "item":
			item := g.NewItemFromString(field.Value)
			if item != nil {
				container.AddItem(item)
			}
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
}

func (g *GameState) openContainer(container *Container) {
	containerItems := itemStacksForUI(StackedItemsWithFilter(container.containedItems, func(item *Item) bool { return true }))
	playerItems := itemStacksForUI(g.Player.GetInventory().StackedItems())

	transferToPlayer := func(itemTaken foundation.ItemForUI) {
		item := itemTaken.(*InventoryStack).First()
		container.RemoveItem(item)
		g.Player.GetInventory().Add(item)

		g.ui.PlayCue("world/pickup")
		g.msg(foundation.HiLite("You take %s from %s.", item.Name(), container.Name()))
		g.openContainer(container)
	}
	transferToContainer := func(itemTaken foundation.ItemForUI) {
		item := itemTaken.(*InventoryStack).First()
		g.Player.GetInventory().Remove(item)
		container.AddItem(item)

		g.ui.PlayCue("world/drop")

		g.msg(foundation.HiLite("You place %s in %s.", item.Name(), container.Name()))

		g.openContainer(container)
	}
	g.ui.ShowGiveAndTakeContainer(g.Player.Name(), playerItems, container.Name(), containerItems, transferToPlayer, transferToContainer)
}
