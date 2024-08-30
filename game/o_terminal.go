package game

import (
	"RogueUI/foundation"
	"bytes"
	"encoding/gob"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"strings"
)

type Terminal struct {
	*BaseObject
	isPlayer          func(*Actor) bool
	startDialogue     func()
	declareAsTerminal bool
}

func (t *Terminal) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := t.BaseObject.gobEncode(enc); err != nil {
		return nil, err
	}

	if err := enc.Encode(t.declareAsTerminal); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *Terminal) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))

	t.BaseObject = &BaseObject{}

	if err := t.BaseObject.gobDecode(dec); err != nil {
		return err
	}
	if err := dec.Decode(&t.declareAsTerminal); err != nil {
		return err
	}

	return nil
}

func (g *GameState) NewTerminal(rec recfile.Record) *Terminal {
	terminal := &Terminal{BaseObject: NewObject(foundation.ObjectTerminal, g.iconForObject)}
	terminal.SetWalkable(false)
	terminal.SetHidden(false)
	terminal.SetTransparent(true)

	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			terminal.customIcon = g.iconForObject(field.Value)
			terminal.useCustomIcon = true
		case "description":
			terminal.displayName = field.Value
		case "dialogue":
			terminal.internalName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			terminal.SetPosition(spawnPos)
		case "tags":
			if strings.ToLower(field.Value) == "no_sound" {
				terminal.declareAsTerminal = false
			}
		case "declared_as_terminal":
			terminal.declareAsTerminal = recfile.StrBool(field.Value)
		}
	}
	terminal.InitWithGameState(g)
	return terminal
}

func (t *Terminal) InitWithGameState(g *GameState) {
	t.iconForObject = g.iconForObject
	t.isPlayer = func(actor *Actor) bool { return actor == g.Player }
	t.startDialogue = func() { g.StartDialogue(t.internalName, t, t.declareAsTerminal) }
}

func (t *Terminal) OnBump(actor *Actor) {
	if t.isPlayer(actor) {
		t.startDialogue()
	}
}

func (t *Terminal) ToRecord() recfile.Record {
	return recfile.Record{
		{Name: "category", Value: t.category.String()},
		{Name: "description", Value: t.displayName},
		{Name: "dialogue", Value: t.internalName},
		{Name: "position", Value: t.position.Encode()},
		{Name: "icon", Value: string(t.customIcon.Char)},
		{Name: "fg", Value: recfile.RGBStr(t.customIcon.Fg)},
		{Name: "bg", Value: recfile.RGBStr(t.customIcon.Bg)},
		{Name: "declared_as_terminal", Value: recfile.BoolStr(t.declareAsTerminal)},
	}
}
