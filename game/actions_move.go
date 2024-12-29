package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/geometry"
)

// MOVEMENT

func (g *GameState) playerMove(newPos geometry.Point) {
	if g.Player.Position() == newPos {
		return
	}
	directConsequencesOfMove := g.actorMove(g.Player, newPos)

	g.ui.AddAnimations(directConsequencesOfMove)

	g.endPlayerTurn(g.Player.TimeNeededForMovement())
}

func (g *GameState) actorMove(actor *Actor, newPos geometry.Point) []foundation.Animation {
	oldPos := actor.Position()
	if oldPos == newPos {
		return nil
	}
	g.currentMap().MoveActor(actor, newPos)
	if actor.Position() == newPos {
		return g.afterActorMovedOnMap(actor, oldPos)
	}
	return nil
}
