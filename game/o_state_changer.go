package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type PushBox struct {
	*BaseObject
	onBump   func(actor *Actor)
	onDamage func(dmg SourcedDamage) []foundation.Animation
}

func (g *GameState) NewPushBox(record recfile.Record, resolver func(objType string) textiles.TextIcon) *PushBox {
	box := &PushBox{
		BaseObject: &BaseObject{
			Category:      foundation.ObjectPushBox,
			UseCustomIcon: false,
			InternalName:  "pushbox",
			iconForObject: resolver,
		},
	}
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "category":
			box.Category = foundation.ObjectCategoryFromString(field.Value)
		case "position":
			box.RawPosition, _ = geometry.NewPointFromEncodedString(field.Value)
		case "description":
			box.DisplayName = field.Value
		}
	}

	box.SetWalkable(false)
	box.SetHidden(false)
	box.SetTransparent(false)

	box.InitWithGameState(g)
	return box
}

func (b *PushBox) InitWithGameState(g *GameState) {
	b.onBump = func(actor *Actor) {
		// push direction
		pushDir := b.Position().Sub(actor.Position())
		targetPos := b.Position().Add(pushDir)
		if g.currentMap().CanPlaceObjectHere(targetPos) {
			g.currentMap().MoveObject(b, targetPos)
			g.ui.PlayCue("world/BOX2")
			if g.currentMap().IsHazardousTileAt(targetPos) { // TODO: maybe ask for specific tile type?
				b.SetTransparent(true)
				b.SetWalkable(true)
			}
			g.endPlayerTurn(g.Player.TimeNeededForActions())
		}
	}
	b.onDamage = func(dmg SourcedDamage) []foundation.Animation {
		if b.Category == foundation.ObjectExplodingPushBox {
			g.currentMap().RemoveObject(b)
			return explosion(g, dmg.Attacker, b.Position(), foundation.Params{
				"damage_interval": fxtools.Interval{Min: 10, Max: 20},
			})
		}
		return nil
	}
}
func (b *PushBox) SetExploding() {
	b.Category = foundation.ObjectExplodingPushBox
}

func (b *PushBox) ToRecord() recfile.Record {
	return recfile.Record{
		{Name: "category", Value: b.Category.String()},
		{Name: "description", Value: b.DisplayName},
		{Name: "position", Value: b.RawPosition.Encode()},
	}
}

func (b *PushBox) OnBump(actor *Actor) {
	b.onBump(actor)
}

func (b *PushBox) OnDamage(dmg SourcedDamage) []foundation.Animation {
	return b.onDamage(dmg)
}
