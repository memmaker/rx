package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
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
func (g *GameState) NewElevator(rec recfile.Record, iconForObject func(objectType string) textiles.TextIcon) *Elevator {
	identifier, description, pos, levels := parseElevatorRecord(rec)

	ele := &Elevator{
		BaseObject: &BaseObject{
			category:      foundation.ObjectElevator,
			isAlive:       true,
			isDrawn:       true,
			position:      pos,
			iconForObject: iconForObject,
			displayName:   description,
			internalName:  identifier,
		},
	}
	ele.SetWalkable(true)
	ele.SetHidden(false)
	ele.SetTransparent(true)

	ele.onWalkOver = func(actor *Actor) []foundation.Animation {
		if actor == g.Player {
			var elevatorActions = make([]foundation.MenuItem, len(levels))
			for i, l := range levels {
				level := l
				elevatorActions[i] = foundation.MenuItem{
					Name: level.Label,
					Action: func() {
						g.GotoNamedLevel(level.LevelName, ele.internalName)
					},
					CloseMenus: true,
				}
			}
			g.ui.OpenMenu(elevatorActions)
		}
		return nil
	}
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
