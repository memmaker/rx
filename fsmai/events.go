package fsmai

import (
	"github.com/memmaker/go/geometry"
)

type TransitionEventName int

const (
	EventNone TransitionEventName = iota
	EventProvoked
	EventCalmed
	EventHeavilyInjured
	EventEnemySighted
	EventLeaderJoined
	EventLeaderLeft
	EventTargetDied
	EventTargetLost
	EventThreatNeutralized
	EventMinorCrimeWitnessed
	EventMajorCrimeWitnessed
)

func (e TransitionEventName) ToString() string {
	switch e {
	case EventNone:
		return "None"
	case EventProvoked:
		return "Provoked"
	case EventHeavilyInjured:
		return "HeavilyInjured"
	case EventEnemySighted:
		return "EnemySighted"
	case EventTargetDied:
		return "TargetDied"
	case EventTargetLost:
		return "TargetLost"
	case EventThreatNeutralized:
		return "ThreatNeutralized"
	case EventCalmed:
		return "Calmed"
	default:
		return "Unknown"
	}
}

type TransitionEvent interface {
	Name() TransitionEventName
}

var NoEvent EmptyEvent = EmptyEvent{EventNone}

type EmptyEvent struct {
	Event TransitionEventName
}

func (e EmptyEvent) Name() TransitionEventName {
	return e.Event
}

type DirectionalEvent struct {
	name      TransitionEventName
	Direction geometry.CompassDirection
}

func (d DirectionalEvent) Name() TransitionEventName {
	return d.name
}

func NewEvent(name TransitionEventName) TransitionEvent {
	return EmptyEvent{name}
}

func NewDirectionalEvent(name TransitionEventName, direction geometry.CompassDirection) TransitionEvent {
	return DirectionalEvent{name, direction}
}
