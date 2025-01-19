package game

import (
	"github.com/memmaker/go/geometry"
)

type LocationEvent struct {
	Event    TransitionEventName
	Location MapPosition
}

func (l LocationEvent) Name() TransitionEventName {
	return l.Event
}

type ActorEvent struct {
	Event TransitionEventName
	Actor *Actor
}

func (a ActorEvent) Name() TransitionEventName {
	return a.Event
}
func NewSuspiciousActivityEvent(mapName string, pos geometry.Point) TransitionEvent {
	return LocationEvent{Event: EventSuspiciousActivity, Location: MapPosition{
		MapName:  mapName,
		Position: pos,
	}}
}
func NewProvokedEvent(provoker *Actor) TransitionEvent {
	return ActorEvent{Event: EventProvoked, Actor: provoker}
}

func NewTargetDiedEvent(target *Actor) TransitionEvent {
	return ActorEvent{Event: EventTargetDied, Actor: target}
}

func NewTargetLostEvent(target *Actor) TransitionEvent {
	return ActorEvent{Event: EventTargetLost, Actor: target}
}

func NewHeavilyInjuredEvent(threat *Actor) TransitionEvent {
	return ActorEvent{Event: EventHeavilyInjured, Actor: threat}
}

func NewEnemySightedEvent(threat *Actor) TransitionEvent {
	return ActorEvent{Event: EventEnemySighted, Actor: threat}
}

func NewCalmedEvent(threat *Actor) TransitionEvent {
	return ActorEvent{Event: EventCalmed, Actor: threat}
}

func NewLeaderJoinedEvent(leader *Actor) TransitionEvent {
	return ActorEvent{Event: EventLeaderJoined, Actor: leader}
}
func NewLeaderLeftEvent() TransitionEvent {
	return EmptyEvent{Event: EventLeaderLeft}
}
