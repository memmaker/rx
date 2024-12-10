package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type Terminal struct {
	*BaseObject
	isPlayer          func(*Actor) bool
	startDialogue     func()
	DeclareAsTerminal bool
}

func (g *GameState) NewTerminal(rec recfile.Record, resolver func(objType string) textiles.TextIcon) *Terminal {
	terminal := &Terminal{BaseObject: NewObject(foundation.ObjectTerminal, resolver), DeclareAsTerminal: true}
	terminal.SetWalkable(false)
	terminal.SetHidden(false)
	terminal.SetTransparent(true)
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			fallthrough
		case "dialogue":
			terminal.InternalName = field.Value
		case "iconoverride":
			terminal.CustomIcon = terminal.iconForObject(field.Value)
			terminal.UseCustomIcon = true
		case "description":
			terminal.DisplayName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			terminal.SetPosition(spawnPos)
		case "tags":
			if strings.ToLower(field.Value) == "no_sound" {
				terminal.DeclareAsTerminal = false
			}
		case "declared_as_terminal":
			terminal.DeclareAsTerminal = recfile.StrBool(field.Value)
		}
	}
	terminal.InitWithGameState(g)
	return terminal
}

func (t *Terminal) AppendContextActions(actions []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	return append(actions, foundation.MenuItem{
		Name:       "Interact",
		Action:     t.startDialogue,
		CloseMenus: true,
	})
}
func (t *Terminal) InitWithGameState(g *GameState) {
	t.isPlayer = func(actor *Actor) bool { return actor == g.Player }
	t.startDialogue = func() { g.StartDialogue(t.InternalName, t, t.DeclareAsTerminal) }
}

func (t *Terminal) OnBump(actor *Actor) {
	if t.isPlayer(actor) {
		t.startDialogue()
	}
}

func (t *Terminal) ToRecord() recfile.Record {
	return recfile.Record{
		{Name: "category", Value: t.Category.String()},
		{Name: "description", Value: t.DisplayName},
		{Name: "dialogue", Value: t.InternalName},
		{Name: "position", Value: t.RawPosition.Encode()},
		{Name: "icon", Value: string(t.CustomIcon.Char)},
		{Name: "fg", Value: recfile.RGBStr(t.CustomIcon.Fg)},
		{Name: "bg", Value: recfile.RGBStr(t.CustomIcon.Bg)},
		{Name: "declared_as_terminal", Value: recfile.BoolStr(t.DeclareAsTerminal)},
	}
}
