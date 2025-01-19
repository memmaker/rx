package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/geometry"
)

// TODO: we need to have major/minor and also obvious vs stealthy crimes
func (g *GameState) onActorCriminalActivity(perpetrator *Actor, victim *Actor, topic foundation.ChatterTopic) {
	if victim == nil && perpetrator == nil {
		return
	}

	isPerpetratorBeingObvious := (perpetrator != nil && !perpetrator.IsSneaky()) || topic == foundation.ChatterAttackNoticed

	if victim != nil {

		// crimes against the player
		if victim == g.Player && perpetrator != nil && victim.Faction != perpetrator.Faction {
			// just make allies react to crimes against the player
			g.makeObservingAlliesReactToMajorCrime(victim, perpetrator, true)
			return
		}

		// impossible to know who did it
		if perpetrator == nil {
			//raise suspicions in actors that have the same faction as the victim
			g.makeObservingAlliesReactToSuspiciousActivity(victim.Position(), victim.Faction)
			return
		}

		// crimes only exist between actors of different factions
		if victim.Faction == perpetrator.Faction {
			return
		}

		// major crimes against NPCs will alert all observing allies
		// NPC crimes are always major
		if topic.IsMajorCrime() || perpetrator != g.Player {
			g.makeObservingAlliesReactToMajorCrime(victim, perpetrator, isPerpetratorBeingObvious)

			if g.isDetectedByObserver(perpetrator, victim) || isPerpetratorBeingObvious {
				victim.FSM.SendEvent(NewProvokedEvent(perpetrator))
			} else {
				victim.FSM.SendEvent(NewSuspiciousActivityEvent(g.currentMapName, perpetrator.Position()))
			}
			return
		}

		// minor crimes against NPCs will trigger a warning first
		if g.isDetectedByObserver(perpetrator, victim) || isPerpetratorBeingObvious {
			g.reactToMinorCrimeOfPlayer(victim, topic)
		}
		return
	}

	// there is no victim, so the crime is against an object or the environment
	// check every observer of the perpetrator
	for _, observer := range g.getObservers(perpetrator.Position()) {
		if observer.Faction == perpetrator.Faction {
			continue
		}

		if !isPerpetratorBeingObvious && !g.isDetectedByObserver(perpetrator, observer) {
			continue
		}

		// just have the first observer react to a minor crime of the player
		if !topic.IsMajorCrime() && perpetrator == g.Player {
			g.reactToMinorCrimeOfPlayer(observer, topic)
			return
		}

		// major crimes against the environment or objects will alert all observing npcs
		if topic.IsMajorCrime() || perpetrator != g.Player {
			observer.FSM.SendEvent(NewProvokedEvent(perpetrator))
		}
	}

	if perpetrator == g.Player {
		g.ui.UpdateVisibleActors()
	}
}

func (g *GameState) makeObservingAlliesReactToSuspiciousActivity(position geometry.Point, faction string) {
	for _, actor := range g.getObservers(position) {
		if actor.Faction == faction && actor.IsIdle() && actor.CanDetect(position) {
			actor.FSM.SendEvent(NewSuspiciousActivityEvent(g.currentMapName, position))
			return
		}
	}
}

func (g *GameState) playerMinorCrimeIsDenied() bool {
	location := g.Player.Position()
	observers := g.getObservers(location)
	for _, observer := range observers {
		if observer.Faction == g.Player.Faction {
			continue
		}

		if observer.HasFlag(foundation.FlagIgnoresCrime) || observer.HasFlag(foundation.FlagAnimal) || observer.HasFlag(foundation.FlagZombie) {
			continue
		}

		if g.isDetectedByObserver(g.Player, observer) {
			g.reactToMinorCrimeOfPlayer(observer, foundation.ChatterMinorCrimeNoticed)
			return true
		}
	}
	return false
}

func (g *GameState) reactToMinorCrimeOfPlayer(observer *Actor, warning foundation.ChatterTopic) {
	if observer.GetFlags().Get(foundation.FlagMinorCrimeWarningsGiven) >= 2 {
		g.tryAddRandomChatter(observer, foundation.ChatterIWarnedYou)
		observer.FSM.SendEvent(NewProvokedEvent(g.Player))
		observer.GetFlags().Unset(foundation.FlagMinorCrimeWarningsGiven)
	} else {
		g.tryAddRandomChatter(observer, warning)
		observer.GetFlags().Increment(foundation.FlagMinorCrimeWarningsGiven)
	}
}

func (g *GameState) makeObservingAlliesReactToMajorCrime(affected *Actor, sourceOfTrouble *Actor, isObvious bool) {
	victimObservers := g.getObservers(affected.Position())
	aggressorObserver := g.getObservers(sourceOfTrouble.Position())
	allObservers := make(map[*Actor]bool)

	for _, observer := range append(victimObservers, aggressorObserver...) {
		if observer == g.Player || observer == affected || observer == sourceOfTrouble || observer.Faction != affected.Faction {
			continue
		}
		if _, ok := allObservers[observer]; ok {
			continue
		}
		allObservers[observer] = true

		if g.isDetectedByObserver(sourceOfTrouble, observer) || isObvious {
			observer.FSM.SendEvent(NewProvokedEvent(sourceOfTrouble))
		} else {
			observer.FSM.SendEvent(NewSuspiciousActivityEvent(g.currentMapName, affected.Position()))
		}
	}
}
