package game

import (
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

func (g *GameState) NewObjectFromRecord(record recfile.Record, palette textiles.ColorPalette, newMap *gridmap.GridMap[*Actor, *Item, Object]) Object {
	objectType := record.FindValueForKeyIgnoreCase("category")

	switch strings.ToLower(objectType) {
	case "explodingpushbox":
		box := g.NewPushBox(record)
		box.SetExploding()
		return box
	case "pushbox":
		return g.NewPushBox(record)
	case "elevator":
		elevator := g.NewElevator(record)
		newMap.AddNamedLocation(elevator.GetIdentifier(), elevator.Position())
		return elevator
	case "unknowncontainer":
		return g.NewContainer(record)
	case "terminal":
		return g.NewTerminal(record)
	case "readable":
		return g.NewReadable(record)
	case "lockeddoor":
		fallthrough
	case "closeddoor":
		fallthrough
	case "brokendoor":
		fallthrough
	case "opendoor":
		return g.NewDoor(record)
	}
	return nil
}

func (g *GameState) iconForObject(objectType string) textiles.TextIcon {
	if icon, exists := g.iconsForObjects[strings.ToLower(objectType)]; exists {
		return icon
	}
	return textiles.TextIcon{}
}

func (g *GameState) iconForItem(itemCategory foundation.ItemCategory) textiles.TextIcon {
	if icon, exists := g.iconsForItems[itemCategory]; exists {
		return icon
	}
	return textiles.TextIcon{}
}
func (g *GameState) addItemToMap(item *Item, mapPos geometry.Point) {
	g.currentMap().AddItemWithDisplacement(item, mapPos)
}

func (g *GameState) NewGold(amount int) *Item {
	gold := &Item{
		description:  "gold",
		internalName: "gold",
		category:     foundation.ItemCategoryGold,
		charges:      amount,
		icon:         g.iconForItem(foundation.ItemCategoryGold),
	}
	return gold
}
