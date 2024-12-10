package game

import (
	"fmt"
	"github.com/memmaker/go/fxtools"
	"strings"
)

type Timed interface {
	ShouldActivate(tickCount int) bool
	IsTimerTicking(tickCount int) bool
	String() string
}
type TimedFunc struct {
	Timed
	ActionOnThing func()
	TickCount     int
	BoundToMap    string
}

func (f TimedFunc) WithTickCount(i int) TimedFunc {
	f.TickCount = i
	return f
}

type Metronome struct {
	timed []TimedFunc
}

func (m *Metronome) AddTimed(timed Timed, boundToMap string, action func()) {
	m.timed = append(m.timed, TimedFunc{
		Timed:         timed,
		ActionOnThing: action,
		TickCount:     0,
		BoundToMap:    boundToMap,
	})
}

func (m *Metronome) RemoveTimed(timed Timed) {
	m.timed = fxtools.FilterSlice(m.timed, func(timedFunc TimedFunc) bool {
		return timedFunc.Timed != timed
	})
}

type Mapper interface {
	ExecuteOnMap(mapName string, action func())
}

func (m *Metronome) Tick(mapper Mapper) {
	m.timed = fxtools.FilterSlice(m.timed, func(timed TimedFunc) bool {
		return timed.IsTimerTicking(timed.TickCount)
	})
	for i, timed := range m.timed {
		if timed.ShouldActivate(timed.TickCount) {
			mapper.ExecuteOnMap(timed.BoundToMap, timed.ActionOnThing)
		}
		m.timed[i] = timed.WithTickCount(timed.TickCount + 1)
	}
}

func (m *Metronome) HasTimed(item Timed) bool {
	for _, timed := range m.timed {
		if timed.Timed == item {
			return true
		}
	}
	return false
}

func (m *Metronome) String() string {
	if len(m.timed) == 0 {
		return "No timed things"
	}
	out := make([]string, len(m.timed))
	for i, timed := range m.timed {
		out[i] = fmt.Sprintf("%d: %s", timed.TickCount, timed.Timed.String())
	}
	return fmt.Sprintf("Timed things:\n%s", strings.Join(out, "\n"))
}
