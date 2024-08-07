package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type Container struct {
	*BaseObject

	isKnown        bool
	containedItems []*Item
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

func (b *Container) OnBump(actor *Actor) {
	if b.onBump != nil {
		b.onBump(actor)
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

func (g *GameState) NewContainer(rec recfile.Record, palette textiles.ColorPalette) Object {
	container := &Container{
		BaseObject: &BaseObject{
			category: foundation.ObjectUnknownContainer,
			isAlive:  true,
			isDrawn:  true,
		},
	}
	container.SetWalkable(false)
	container.SetHidden(false)
	container.SetTransparent(true)

	container.onBump = func(actor *Actor) {
		if actor == g.Player {
			if !container.ContainsItems() {
				g.msg(foundation.HiLite("The %s is empty", container.Name()))
				return
			}
			g.openContainer(container)
		}
	}
	var icon textiles.TextIcon
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "description":
			container.displayName = field.Value
		case "icon":
			icon.Char = field.AsRune()
		case "foreground":
			icon.Fg = palette.Get(field.Value)
		case "background":
			icon.Bg = palette.Get(field.Value)
		case "position":
			container.position, _ = geometry.NewPointFromEncodedString(field.Value)
		case "item":
			item := g.NewItemFromName(field.Value)
			container.AddItem(item)
		}
	}
	container.icon = icon
	return container
}

func (g *GameState) openContainer(container *Container) {
	itemsUI := itemsForUI(container.containedItems)
	transferItem := func(itemTaken foundation.ItemForUI) {
		item := itemTaken.(*Item)
		container.RemoveItem(item)
		g.Player.GetInventory().Add(item)
		if container.ContainsItems() {
			g.openContainer(container)
		}
	}
	g.msg(foundation.HiLite("You search %s", container.Name()))
	g.ui.ShowContainer(container.Name(), itemsUI, transferItem)
}
