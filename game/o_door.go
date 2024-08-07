package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type DoorState int

const (
	DoorLocked DoorState = iota
	DoorClosed
	DoorOpen
	DoorBroken
)

type Door struct {
	*BaseObject
	state      DoorState
	lockedFlag string
	lockDiff   foundation.Difficulty
}

func (b *Door) GetCategory() foundation.ObjectCategory {
	switch b.state {
	case DoorLocked:
		return foundation.ObjectLockedDoor
	case DoorClosed:
		return foundation.ObjectClosedDoor
	case DoorOpen:
		return foundation.ObjectOpenDoor
	case DoorBroken:
		return foundation.ObjectBrokenDoor
	}
	return foundation.ObjectClosedDoor
}

func (b *Door) IsTransparent() bool {
	switch b.state {
	case DoorOpen:
		return true
	case DoorBroken:
		return true
	}
	return false
}

func (b *Door) IsWalkable(actor *Actor) bool {
	return b.state != DoorLocked || (actor != nil && actor.HasKey(b.lockedFlag))
}

func (b *Door) SetLockedByFlag(flag string) {
	b.lockedFlag = flag
}
func (b *Door) SetLockDifficulty(difficulty foundation.Difficulty) {
	b.lockDiff = difficulty
}

func (b *Door) SetStateFromCategory(cat foundation.ObjectCategory) {
	switch cat {
	case foundation.ObjectLockedDoor:
		b.state = DoorLocked
	case foundation.ObjectClosedDoor:
		b.state = DoorClosed
	case foundation.ObjectOpenDoor:
		b.state = DoorOpen
	case foundation.ObjectBrokenDoor:
		b.state = DoorBroken
	}
}

func (g *GameState) NewDoor(rec recfile.Record, palette textiles.ColorPalette) *Door {
	door := &Door{
		BaseObject: &BaseObject{
			category:    foundation.ObjectClosedDoor,
			isAlive:     true,
			isDrawn:     true,
			displayName: "a door",
		},
		lockDiff: foundation.Easy,
	}
	door.SetWalkable(false)
	door.SetHidden(false)
	door.SetTransparent(true)

	door.onBump = func(actor *Actor) {
		if actor == g.Player && door.state == DoorLocked {

			if door.lockedFlag != "" && actor.HasKey(door.lockedFlag) {
				door.state = DoorClosed
				g.msg(foundation.Msg("You unlocked the door"))
				return
			}

			g.ui.StartLockpickGame(door.lockDiff, g.Player.GetInventory().GetLockpickCount, g.Player.GetInventory().RemoveLockpick, func(result foundation.InteractionResult) {
				if result == foundation.Success {
					door.state = DoorClosed
					g.msg(foundation.Msg("You picked the lock deftly"))
				}
			})
		}
	}

	var icon textiles.TextIcon
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			switch strings.ToLower(field.Value) {
			case "lockeddoor":
				door.state = DoorLocked
			case "closeddoor":
				door.state = DoorClosed
			case "opendoor":
				door.state = DoorOpen
			case "brokendoor":
				door.state = DoorBroken
			}
		case "icon":
			icon.Char = field.AsRune()
		case "foreground":
			icon.Fg = palette.Get(field.Value)
		case "background":
			icon.Bg = palette.Get(field.Value)
		case "lockflag":
			door.lockedFlag = field.Value
		case "lockdifficulty":
			door.lockDiff = foundation.DifficultyFromString(field.Value)
		case "position":
			door.position, _ = geometry.NewPointFromEncodedString(field.Value)
		}
	}
	door.icon = icon

	return door
}
