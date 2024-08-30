package game

import (
	"RogueUI/foundation"
	"bytes"
	"encoding/gob"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"strings"
)

type ElevatorButton struct {
	Label     string
	LevelName string
}

func (b ElevatorButton) HasValues() bool {
	return b.Label != "" && b.LevelName != ""
}

type Elevator struct {
	*BaseObject
	lockedFlag string
	isPlayer   func(actor *Actor) bool
	activate   func()
	levels     []ElevatorButton
}

func (b *Elevator) InitWithGameState(g *GameState) {
	b.iconForObject = g.iconForObject
	b.isPlayer = func(actor *Actor) bool { return actor == g.Player }

	transitionTo := func(levelName, location string) {
		g.ui.PlayCue("world/elevator")
		g.GotoNamedLevel(levelName, location)
	}
	b.activate = func() {
		var elevatorActions = make([]foundation.MenuItem, len(b.levels))
		for i, l := range b.levels {
			level := l
			elevatorActions[i] = foundation.MenuItem{
				Name: level.Label,
				Action: func() {
					transitionTo(level.LevelName, b.internalName)
				},
				CloseMenus: true,
			}
		}
		g.ui.OpenMenu(elevatorActions)
	}

}
func (b *Elevator) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := b.BaseObject.gobEncode(enc); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.lockedFlag); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.levels); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *Elevator) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))

	b.BaseObject = &BaseObject{}

	if err := b.BaseObject.gobDecode(dec); err != nil {
		return err
	}

	if err := dec.Decode(&b.lockedFlag); err != nil {
		return err
	}

	if err := dec.Decode(&b.levels); err != nil {
		return err
	}

	return nil
}

func (b *Elevator) GetCategory() foundation.ObjectCategory {
	return foundation.ObjectElevator
}

func (b *Elevator) IsTransparent() bool {
	return true
}

func (b *Elevator) IsWalkable(actor *Actor) bool {
	return true
}

func (b *Elevator) GetIdentifier() string {
	return b.internalName
}

func (b *Elevator) SetLockedByFlag(flag string) {
	b.lockedFlag = flag
}
func (g *GameState) NewElevator(rec recfile.Record) *Elevator {
	identifier, description, pos, levels := parseElevatorRecord(rec)

	ele := &Elevator{
		BaseObject: &BaseObject{
			category:      foundation.ObjectElevator,
			isAlive:       true,
			position:      pos,
			iconForObject: g.iconForObject,
			displayName:   description,
			internalName:  identifier,
		},
		levels: levels,
	}
	ele.SetWalkable(true)
	ele.SetHidden(false)
	ele.SetTransparent(true)

	ele.InitWithGameState(g)
	return ele
}

func parseElevatorRecord(rec recfile.Record) (string, string, geometry.Point, []ElevatorButton) {
	var levels []ElevatorButton
	var identifier string
	var description string
	var currentButton ElevatorButton
	var position geometry.Point
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "position":
			position, _ = geometry.NewPointFromEncodedString(field.Value)
		case "description":
			description = field.Value
		case "identifier":
			identifier = field.Value
		case "floordescription":
			currentButton.Label = field.Value
			if currentButton.HasValues() {
				levels = append(levels, currentButton)
				currentButton = ElevatorButton{}
			}
		case "floortarget":
			currentButton.LevelName = field.Value
			if currentButton.HasValues() {
				levels = append(levels, currentButton)
				currentButton = ElevatorButton{}
			}
		}
	}

	return identifier, description, position, levels
}

func (b *Elevator) OnBump(actor *Actor) {
	if b.isPlayer(actor) {
		b.activate()
	}
}

func (b *Elevator) OnWalkOver(actor *Actor) []foundation.Animation {
	if b.isPlayer(actor) {
		b.activate()
	}
	return nil
}

func (b *Elevator) ToRecord() recfile.Record {
	rec := recfile.Record{}
	rec = append(rec, recfile.Field{Name: "position", Value: b.position.Encode()})
	rec = append(rec, recfile.Field{Name: "description", Value: b.displayName})
	rec = append(rec, recfile.Field{Name: "identifier", Value: b.internalName})
	for _, level := range b.levels {
		rec = append(rec, recfile.Field{Name: "floordescription", Value: level.Label})
		rec = append(rec, recfile.Field{Name: "floortarget", Value: level.LevelName})
	}
	return rec
}
