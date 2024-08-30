package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
)

// MOVEMENT

func (g *GameState) playerMove(oldPos geometry.Point, newPos geometry.Point) {
	directConsequencesOfMove := g.actorMove(g.Player, newPos)

	g.afterPlayerMoved(oldPos, false)

	g.ui.AddAnimations(directConsequencesOfMove)

	g.endPlayerTurn(g.Player.timeNeededForActions())
}
func (g *GameState) actorMoveAnimated(actor *Actor, newPos geometry.Point) []foundation.Animation {
	oldPos := actor.Position()
	var moveAnims []foundation.Animation
	if g.couldPlayerSeeActor(actor) && (g.canPlayerSee(newPos) || g.canPlayerSee(oldPos)) && actor != g.Player {
		move := g.ui.GetAnimMove(actor, oldPos, newPos)
		move.RequestMapUpdateOnFinish()
		moveAnims = append(moveAnims, move)
	}
	moveAnims = append(moveAnims, g.actorMove(actor, newPos)...)
	return moveAnims
}
func (g *GameState) actorMove(actor *Actor, newPos geometry.Point) []foundation.Animation {
	oldPos := actor.Position()
	if oldPos == newPos {
		return nil
	}
	g.currentMap().MoveActor(actor, newPos)
	if actor.Position() == newPos {
		return g.triggerTileEffectsAfterMovement(actor, oldPos, newPos)
	}
	return nil
}
