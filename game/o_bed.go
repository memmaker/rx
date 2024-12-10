package game

import (
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type Bed struct {
	*BaseObject
	isPlayer    func(*Actor) bool
	NameOfOwner string
	Occupied    bool
	sleepAction func()
}

func (g *GameState) NewBed(rec recfile.Record, iconForObject func(objectType string) textiles.TextIcon) *Bed {
	bed := &Bed{
		BaseObject: NewObject(foundation.ObjectBed, iconForObject),
	}

	bed.SetWalkable(true)
	bed.SetHidden(false)
	bed.SetTransparent(true)

	var customIcon textiles.TextIcon
	useCustomIcon := false
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			bed.InternalName = field.Value
		case "iconoverride":
			customIcon = bed.iconForObject(field.Value)
			useCustomIcon = true
		case "icon":
			customIcon.Char = field.AsRune()
			useCustomIcon = true
		case "fg":
			customIcon.Fg = field.AsRGB(",")
		case "bg":
			customIcon.Bg = field.AsRGB(",")
		case "description":
			bed.DisplayName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			bed.SetPosition(spawnPos)
		case "owner":
			bed.NameOfOwner = field.Value
		}
	}

	bed.CustomIcon = customIcon
	bed.UseCustomIcon = useCustomIcon

	if bed.InternalName == "" {
		if bed.NameOfOwner != "" {
			bed.InternalName = fmt.Sprintf("bed_of_%s", bed.NameOfOwner)
		} else {
			bed.InternalName = "bed"
		}
	}

	bed.InitWithGameState(g)
	return bed
}

func (b *Bed) InitWithGameState(g *GameState) {
	b.isPlayer = func(actor *Actor) bool { return actor == g.Player }
	b.sleepAction = func() {
		g.openRestMenu(true)
	}
}
func (b *Bed) AppendContextActions(actions []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	if b.NameOfOwner != "" && b.NameOfOwner != "player" {
		return actions
	}

	return append(actions, foundation.MenuItem{
		Name:       "Sleep",
		Action:     b.sleepAction,
		CloseMenus: true,
	})
}

func (b *Bed) OnWalkOver(actor *Actor) []foundation.Animation {
	if b.isPlayer(actor) && (b.NameOfOwner == "" || strings.ToLower(b.NameOfOwner) == "player") {
		b.sleepAction()
	}
	return nil
}

func (b *Bed) ToRecord() recfile.Record {
	rec := recfile.Record{
		{Name: "category", Value: b.Category.String()},
		{Name: "position", Value: b.RawPosition.Encode()},
		{Name: "description", Value: b.DisplayName},
		{Name: "icon", Value: string(b.CustomIcon.Char)},
		{Name: "fg", Value: recfile.RGBStr(b.CustomIcon.Fg)},
		{Name: "bg", Value: recfile.RGBStr(b.CustomIcon.Bg)},
		{Name: "isoccupied", Value: recfile.BoolStr(b.Occupied)},
	}
	if b.NameOfOwner != "" {
		rec = append(rec, recfile.Field{Name: "owner", Value: b.NameOfOwner})
	}
	return rec
}
