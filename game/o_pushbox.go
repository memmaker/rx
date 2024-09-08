package game

import (
	"RogueUI/foundation"
	"bytes"
	"encoding/gob"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"strings"
)

type PushBox struct {
	*BaseObject
	onBump   func(actor *Actor)
	onDamage func(dmg SourcedDamage) []foundation.Animation
}

func (b *PushBox) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := b.BaseObject.gobEncode(enc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *PushBox) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))

	b.BaseObject = &BaseObject{}

	if err := b.BaseObject.gobDecode(dec); err != nil {
		return err
	}

	return nil
}
func (g *GameState) NewPushBox(record recfile.Record) *PushBox {
	box := &PushBox{
		BaseObject: &BaseObject{
			category:      foundation.ObjectPushBox,
			isAlive:       true,
			useCustomIcon: false,
			internalName:  "pushbox",
		},
	}
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "category":
			box.category = foundation.ObjectCategoryFromString(field.Value)
		case "position":
			box.position, _ = geometry.NewPointFromEncodedString(field.Value)
		case "description":
			box.displayName = field.Value
		}
	}

	box.SetWalkable(false)
	box.SetHidden(false)
	box.SetTransparent(false)

	box.InitWithGameState(g)
	return box
}

func (b *PushBox) InitWithGameState(g *GameState) {
	b.iconForObject = g.iconForObject
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
		}
	}
	b.onDamage = func(dmg SourcedDamage) []foundation.Animation {
		if b.category == foundation.ObjectExplodingPushBox {
			g.currentMap().RemoveObject(b)
			return explosion(g, dmg.Attacker, b.Position(), NewParams(map[string]string{
				"radius": "3",
				"damage": "10-20",
			}))
		}
		return nil
	}
}
func (b *PushBox) SetExploding() {
	b.category = foundation.ObjectExplodingPushBox
}

func (b *PushBox) ToRecord() recfile.Record {
	return recfile.Record{
		{Name: "category", Value: b.category.String()},
		{Name: "description", Value: b.displayName},
		{Name: "position", Value: b.position.Encode()},
	}
}

func (b *PushBox) OnBump(actor *Actor) {
	b.onBump(actor)
}

func (b *PushBox) OnDamage(dmg SourcedDamage) []foundation.Animation {
	return b.onDamage(dmg)
}
