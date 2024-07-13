package game

import "RogueUI/foundation"

type DoorState int

const (
	DoorLocked DoorState = iota
	DoorClosed
	DoorOpen
	DoorBroken
)

type Door struct {
	*BaseObject
	state         DoorState
	keyIdentifier string
	lockedFlag    string
	lockDiff      foundation.Difficulty
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
	return b.state != DoorLocked || (actor != nil && actor.HasKey(b.keyIdentifier))
}

func (b *Door) SetLockedByFlag(flag string) {
	b.lockedFlag = flag
}
func (b *Door) SetLockedByKey(keyIdentifier string) {
	b.keyIdentifier = keyIdentifier
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

func (g *GameState) NewDoor(displayName string) *Door {
	door := &Door{
		BaseObject: &BaseObject{
			category:    foundation.ObjectClosedDoor,
			isAlive:     true,
			isDrawn:     true,
			displayName: displayName,
		},
		lockDiff: foundation.Easy,
	}
	door.SetWalkable(false)
	door.SetHidden(false)
	door.SetTransparent(true)

	door.onBump = func(actor *Actor) {
		if actor == g.Player && door.state == DoorLocked {

			if door.keyIdentifier != "" && actor.HasKey(door.keyIdentifier) {
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
	return door
}
