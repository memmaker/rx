package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type PushBox struct {
	*BaseObject
}

func (g *GameState) NewPushBox(record recfile.Record, iconForObject func(objectType string) textiles.TextIcon) *PushBox {

	var spawnPos geometry.Point
	var description string
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "position":
			spawnPos, _ = geometry.NewPointFromEncodedString(field.Value)
		case "description":
			description = field.Value
		}
	}

	box := &PushBox{
		BaseObject: &BaseObject{
			category:      foundation.ObjectPushBox,
			isAlive:       true,
			isDrawn:       true,
			iconForObject: iconForObject,
			position:      spawnPos,
			useCustomIcon: false,
			internalName:  "pushbox",
			displayName:   description,
		},
	}
	box.SetWalkable(false)
	box.SetHidden(false)
	box.SetTransparent(false)

	box.onBump = func(actor *Actor) {
		if actor == g.Player {
			// push direction
			pushDir := box.Position().Sub(actor.Position())
			targetPos := box.Position().Add(pushDir)
			if g.gridMap.CanPlaceObjectHere(targetPos) {
				g.gridMap.MoveObject(box, targetPos)
				g.ui.PlayCue("world/BOX2")
				if g.gridMap.IsDamagingTileAt(targetPos) { // TODO: maybe ask for specific tile type?
					box.SetTransparent(true)
					box.SetWalkable(true)
				}
			}
		}
	}
	box.onDamage = func(dmg SourcedDamage) []foundation.Animation {
		if box.category == foundation.ObjectExplodingPushBox {
			g.gridMap.RemoveObject(box)
			return explosion(g, dmg.Attacker, box.Position())
		}
		return nil
	}
	return box
}

func (b *PushBox) SetExploding() {
	b.category = foundation.ObjectExplodingPushBox
	b.triggerOnDamage = true
}
