package console

import (
	"github.com/gdamore/tcell/v2"
	"time"
)

type EventJoy struct {
	AxisData   []int8
	Buttons    uint32
	OccurredAt time.Time
}

func (e EventJoy) IsButtonReleased(index int) bool {
	return e.Buttons&(1<<uint(index)) == 0
}
func (e EventJoy) IsButtonDown(index int) bool {
	return e.Buttons&(1<<uint(index)) > 0
}

func (e EventJoy) When() time.Time {
	return e.OccurredAt
}

func NewJoyStickEvent(data []int8, buttons uint32) tcell.Event {
	return &EventJoy{
		AxisData:   data,
		Buttons:    buttons,
		OccurredAt: time.Now(),
	}
}
