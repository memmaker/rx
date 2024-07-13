package game

import "RogueUI/foundation"

type ElevatorButton struct {
	Label     string
	LevelName string
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

func (b *Elevator) SetLockedByFlag(flag string) {
	b.lockedFlag = flag
}
func (g *GameState) NewElevator(internalName, displayName string, levels []ElevatorButton) *Elevator {
	ele := &Elevator{
		BaseObject: &BaseObject{
			category:     foundation.ObjectElevator,
			isAlive:      true,
			isDrawn:      true,
			displayName:  displayName,
			internalName: internalName,
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
