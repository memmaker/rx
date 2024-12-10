package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

/*
%rec: recipes

name: meld_production
input: liquid_methane
input: lidocaine
output: meld
req_skill: biochemistry(30)

name: meld_maker
category: maker
description: A machine that can produce Meld from raw materials. It has to be pre-configured by an expert with a blueprint.
recipe: meld_production

*/

type MakerBlueprint struct {
	Ingredients  []string
	Results      []string
	Requirements d100.CharacterRequirement
}

type ItemMaker struct {
	*BaseObject
	isPlayer          func(*Actor) bool
	recipeFromName    func(string) MakerBlueprint
	recipes           []MakerBlueprint
	isBroken          bool
	isRiggedToExplode bool
	outputSlots       []foundation.Item
}

func (g *GameState) NewItemMaker(rec recfile.Record, iconForObject func(objectType string) textiles.TextIcon) *ItemMaker {
	maker := &ItemMaker{
		BaseObject: NewObject(foundation.ObjectReadable, iconForObject),
	}

	maker.SetWalkable(false)
	maker.SetHidden(false)
	maker.SetTransparent(true)

	var customIcon textiles.TextIcon
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			maker.InternalName = field.Value
		case "iconoverride":
			customIcon = maker.iconForObject(field.Value)
		case "icon":
			customIcon.Char = field.AsRune()
		case "fg":
			customIcon.Fg = field.AsRGB(",")
		case "bg":
			customIcon.Bg = field.AsRGB(",")
		case "description":
			maker.DisplayName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			maker.SetPosition(spawnPos)
		}
	}

	maker.CustomIcon = customIcon
	maker.UseCustomIcon = true
	maker.InternalName = "readable"
	maker.InitWithGameState(g)
	return maker
}

func (r *ItemMaker) InitWithGameState(g *GameState) {
	r.isPlayer = func(actor *Actor) bool { return actor == g.Player }
}
func (r *ItemMaker) AppendContextActions(actions []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	return append(actions, foundation.MenuItem{
		Name:       "Read",
		CloseMenus: true,
	})
}

func (r *ItemMaker) OnBump(actor *Actor) {
	if r.isPlayer(actor) {
	}
}

func (r *ItemMaker) ToRecord() recfile.Record {
	rec := recfile.Record{
		{Name: "category", Value: r.Category.String()},
		{Name: "position", Value: r.RawPosition.Encode()},
		{Name: "description", Value: r.DisplayName},
		{Name: "icon", Value: string(r.CustomIcon.Char)},
		{Name: "fg", Value: recfile.RGBStr(r.CustomIcon.Fg)},
		{Name: "bg", Value: recfile.RGBStr(r.CustomIcon.Bg)},
	}
	return rec
}
