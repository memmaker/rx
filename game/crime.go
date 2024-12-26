package game

import "contractor/foundation"

func (g *GameState) playerMinorCrimeIsDenied() bool {
	location := g.Player.Position()
	observers := g.getObservers(location)
	for _, observer := range observers {
		if observer.TeamName == g.Player.TeamName {
			continue
		}

		if observer.HasFlag(foundation.FlagIgnoresCrime) || observer.HasFlag(foundation.FlagAnimal) || observer.HasFlag(foundation.FlagZombie) {
			continue
		}

		if g.isDetectedByObserver(g.Player, observer) {
			g.reactToMinorCrime(observer, foundation.ChatterMinorCrimeNoticed)
			return true
		}
	}
	return false
}

func (g *GameState) reactToMinorCrime(observer *Actor, warning foundation.ChatterTopic) {
	if observer.GetFlags().Get(foundation.FlagMinorCrimeWarningsGiven) >= 2 {
		g.tryAddRandomChatter(observer, foundation.ChatterIWarnedYou)
		g.trySetHostile(observer, g.Player)
		observer.GetFlags().Unset(foundation.FlagMinorCrimeWarningsGiven)
	} else {
		g.tryAddRandomChatter(observer, warning)
		observer.GetFlags().Increment(foundation.FlagMinorCrimeWarningsGiven)
	}
}
