package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type Door struct {
	*BaseObject
	lockedFlag string
	lockDiff   foundation.Difficulty
	numberLock []rune
}

func (b *Door) IsTransparent() bool {
	switch b.GetCategory() {
	case foundation.ObjectOpenDoor:
		return true
	case foundation.ObjectBrokenDoor:
		return true
	}
	return false
}

func (b *Door) GetIcon() textiles.TextIcon {
	return b.iconForObject(b.GetCategory().LowerString())
}
func (b *Door) IsWalkable(actor *Actor) bool {
	return b.GetCategory() != foundation.ObjectLockedDoor || (actor != nil && actor.HasKey(b.lockedFlag))
}
func (b *Door) AppendContextActions(items []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	if b.GetCategory() == foundation.ObjectLockedDoor && g.Player.HasKey(b.lockedFlag) {
		items = append(items, foundation.MenuItem{
			Name:       "Unlock",
			Action:     b.Unlock,
			CloseMenus: true,
		})
	}
	if b.GetCategory() == foundation.ObjectOpenDoor {
		items = append(items, foundation.MenuItem{
			Name:       "Close",
			Action:     b.Close,
			CloseMenus: true,
		})
	}
	if b.GetCategory() == foundation.ObjectClosedDoor {
		items = append(items, foundation.MenuItem{
			Name:       "Open",
			Action:     b.Open,
			CloseMenus: true,
		})
	}
	return items
}

func (b *Door) Close() {
	b.category = foundation.ObjectClosedDoor
}
func (b *Door) SetLockedByFlag(flag string) {
	b.lockedFlag = flag
}
func (b *Door) SetLockDifficulty(difficulty foundation.Difficulty) {
	b.lockDiff = difficulty
}

func (b *Door) IsLocked() bool {
	return b.GetCategory() == foundation.ObjectLockedDoor
}

func (b *Door) GetLockFlag() string {
	return b.lockedFlag
}

func (b *Door) Unlock() {
	b.category = foundation.ObjectClosedDoor
}

func (b *Door) Open() {
	b.category = foundation.ObjectOpenDoor
}

func (g *GameState) NewDoor(rec recfile.Record, iconForObject func(category string) textiles.TextIcon) *Door {
	door := &Door{
		BaseObject: &BaseObject{
			category:      foundation.ObjectClosedDoor,
			isAlive:       true,
			isDrawn:       true,
			displayName:   "a door",
			iconForObject: iconForObject,
		},
		lockDiff: foundation.Easy,
	}
	door.SetWalkable(false)
	door.SetHidden(false)
	door.SetTransparent(true)

	door.onBump = func(actor *Actor) {
		if actor == g.Player && door.GetCategory() == foundation.ObjectLockedDoor {

			if door.lockedFlag != "" && actor.HasKey(door.lockedFlag) {
				door.category = foundation.ObjectClosedDoor
				g.msg(foundation.Msg("You unlocked the door"))
				return
			}

			if len(door.numberLock) > 0 {
				g.ui.OpenKeypad(door.numberLock, func(result bool) {
					if result {
						door.category = foundation.ObjectClosedDoor
						g.msg(foundation.Msg("You unlocked the door"))
					}
				})
			} else {
				g.ui.StartLockpickGame(door.lockDiff, g.Player.GetInventory().GetLockpickCount, g.Player.GetInventory().RemoveLockpick, func(result foundation.InteractionResult) {
					if result == foundation.Success {
						door.category = foundation.ObjectClosedDoor
						g.msg(foundation.Msg("You picked the lock deftly"))
					}
				})
			}
		}
	}
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "category":
			switch strings.ToLower(field.Value) {
			case "lockeddoor":
				door.category = foundation.ObjectLockedDoor
			case "closeddoor":
				door.category = foundation.ObjectClosedDoor
			case "opendoor":
				door.category = foundation.ObjectOpenDoor
			case "brokendoor":
				door.category = foundation.ObjectBrokenDoor
			}
		case "description":
			door.displayName = field.Value
		case "lockflag":
			door.lockedFlag = field.Value
		case "numberlock":
			door.numberLock = []rune(field.Value)
		case "lockdifficulty":
			door.lockDiff = foundation.DifficultyFromString(field.Value)
		case "position":
			door.position, _ = geometry.NewPointFromEncodedString(field.Value)
		}
	}
	return door
}
