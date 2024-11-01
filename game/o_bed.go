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

type Bed struct {
	*BaseObject
	isPlayer    func(*Actor) bool
	nameOfOwner string
	isOccupied  bool
	sleepAction func()
}

func (b *Bed) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := b.BaseObject.gobEncode(enc); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.nameOfOwner); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.isOccupied); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *Bed) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))

	b.BaseObject = &BaseObject{}

	if err := b.BaseObject.gobDecode(dec); err != nil {
		return err
	}
	if err := dec.Decode(&b.nameOfOwner); err != nil {
		return err
	}

	if err := dec.Decode(&b.isOccupied); err != nil {
		return err
	}

	return nil
}
func (g *GameState) NewBed(rec recfile.Record) *Bed {
	bed := &Bed{
		BaseObject: NewObject(foundation.ObjectBed, g.iconForObject),
	}

	bed.SetWalkable(true)
	bed.SetHidden(false)
	bed.SetTransparent(true)

	var customIcon textiles.TextIcon
	useCustomIcon := false
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			bed.internalName = field.Value
		case "iconoverride":
			customIcon = g.iconForObject(field.Value)
			useCustomIcon = true
		case "icon":
			customIcon.Char = field.AsRune()
			useCustomIcon = true
		case "fg":
			customIcon.Fg = field.AsRGB(",")
		case "bg":
			customIcon.Bg = field.AsRGB(",")
		case "description":
			bed.displayName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			bed.SetPosition(spawnPos)
		case "owner":
			bed.nameOfOwner = field.Value
		}
	}

	bed.customIcon = customIcon
	bed.useCustomIcon = useCustomIcon

	if bed.internalName == "" {
		if bed.nameOfOwner != "" {
			bed.internalName = fmt.Sprintf("bed_of_%s", bed.nameOfOwner)
		} else {
			bed.internalName = "bed"
		}
	}

	bed.InitWithGameState(g)
	return bed
}

func (b *Bed) InitWithGameState(g *GameState) {
	b.iconForObject = g.iconForObject
	b.isPlayer = func(actor *Actor) bool { return actor == g.Player }
	b.sleepAction = func() {
		g.openRestMenu(true)
	}
}
func (b *Bed) AppendContextActions(actions []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	if b.nameOfOwner != "" && b.nameOfOwner != "player" {
		return actions
	}

	return append(actions, foundation.MenuItem{
		Name:       "Sleep",
		Action:     b.sleepAction,
		CloseMenus: true,
	})
}

func (b *Bed) OnWalkOver(actor *Actor) []foundation.Animation {
	if b.isPlayer(actor) {
		b.sleepAction()
	}
	return nil
}

func (b *Bed) ToRecord() recfile.Record {
	rec := recfile.Record{
		{Name: "category", Value: b.category.String()},
		{Name: "position", Value: b.position.Encode()},
		{Name: "description", Value: b.displayName},
		{Name: "icon", Value: string(b.customIcon.Char)},
		{Name: "fg", Value: recfile.RGBStr(b.customIcon.Fg)},
		{Name: "bg", Value: recfile.RGBStr(b.customIcon.Bg)},
		{Name: "isoccupied", Value: recfile.BoolStr(b.isOccupied)},
	}
	if b.nameOfOwner != "" {
		rec = append(rec, recfile.Field{Name: "owner", Value: b.nameOfOwner})
	}
	return rec
}
