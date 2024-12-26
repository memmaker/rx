package game

import (
	"contractor/fsmai"
	"github.com/memmaker/go/geometry"
)

type LocationEvent struct {
	Event    fsmai.TransitionEventName
	Location geometry.Point
}

func (l LocationEvent) Name() fsmai.TransitionEventName {
	return l.Event
}

type ActorEvent struct {
	Event fsmai.TransitionEventName
	Actor *Actor
}

func (a ActorEvent) Name() fsmai.TransitionEventName {
	return a.Event
}

func NewProvokedEvent(provoker *Actor) fsmai.TransitionEvent {
	return ActorEvent{Event: fsmai.EventProvoked, Actor: provoker}
}

func NewTargetDiedEvent(target *Actor) fsmai.TransitionEvent {
	return ActorEvent{Event: fsmai.EventTargetDied, Actor: target}
}

func NewTargetLostEvent(target *Actor) fsmai.TransitionEvent {
	return ActorEvent{Event: fsmai.EventTargetLost, Actor: target}
}

func NewHeavilyInjuredEvent(threat *Actor) fsmai.TransitionEvent {
	return ActorEvent{Event: fsmai.EventHeavilyInjured, Actor: threat}
}

func NewEnemySightedEvent(threat *Actor) fsmai.TransitionEvent {
	return ActorEvent{Event: fsmai.EventEnemySighted, Actor: threat}
}

func NewCalmedEvent(threat *Actor) fsmai.TransitionEvent {
	return ActorEvent{Event: fsmai.EventCalmed, Actor: threat}
}

func NewLeaderJoinedEvent(leader *Actor) fsmai.TransitionEvent {
	return ActorEvent{Event: fsmai.EventLeaderJoined, Actor: leader}
}
func NewLeaderLeftEvent() fsmai.TransitionEvent {
	return fsmai.EmptyEvent{Event: fsmai.EventLeaderLeft}
}
func NewMinorCrimeWitnessedEvent(criminal *Actor) fsmai.TransitionEvent {
	return ActorEvent{Event: fsmai.EventMinorCrimeWitnessed, Actor: criminal}
}

func NewMajorCrimeWitnessedEvent(criminal *Actor) fsmai.TransitionEvent {
	return ActorEvent{Event: fsmai.EventMajorCrimeWitnessed, Actor: criminal}
}
