package game

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/rpg"
	"math/rand"
)

func (g *GameState) executeAI(enemy *Actor) {
	if !enemy.IsAlive() {
		return
	}

	if enemy.HasFlag(foundation.IsStunned) {
		stunCounter := enemy.GetFlagCounter(foundation.IsStunned)
		if stunCounter == 1 {
			//g.msg(foundation.HiLite("%s is stunned", enemy.Name()))
			enemy.IncrementFlagCounter(foundation.IsStunned)
			return
		} else {
			turnMod := stunCounter - 1
			_, success, _ := rpg.SuccessRoll(enemy.GetIntelligence() + turnMod)
			if !success {
				enemy.IncrementFlagCounter(foundation.IsStunned)
				return
			}
			g.msg(foundation.HiLite("%s clears its mind", enemy.Name()))
		}
	}

	if g.currentDungeonLevel <= 2 || enemy.HasFlag(foundation.IsSlow) {
		if g.TurnsTaken%2 == 0 {
			return
		}
	}

	g.aiAct(enemy)

	if enemy.HasFlag(foundation.IsHasted) {
		g.aiAct(enemy)
	}
}

func (g *GameState) aiAct(enemy *Actor) {
	distanceToPlayer := geometry.DistanceChebyshev(enemy.Position(), g.Player.Position())

	sameRoom := g.isInPlayerRoom(enemy.Position()) || distanceToPlayer <= 1

	if enemy.HasFlag(foundation.IsHeld) {
		if rand.Intn(10) == 0 {
			enemy.GetFlags().Unset(foundation.IsHeld)
			g.msg(foundation.HiLite("%s breaks free", enemy.Name()))
		} else {
			return
		}
	}

	if enemy.IsSleeping() {
		if sameRoom && enemy.HasFlag(foundation.IsMean) && rand.Intn(3) != 0 {
			//if (!on(*tp, ISRUN) && rnd(3) != 0 && on(*tp, ISMEAN) && !on(*tp, ISHELD)
			//		&& !ISWEARING(R_STEALTH))
			enemy.WakeUp()
		} else if enemy.IsSleeping() {
			return
		}
	}

	if enemy.HasFlag(foundation.CanConfuse) && rand.Intn(4) == 0 {
		enemy.GetFlags().Unset(foundation.CanConfuse)
		g.msg(foundation.HiLite("%s stops glowing red", enemy.Name()))
	}

	consequencesOfConfusion := g.doesActConfused(enemy)
	if len(consequencesOfConfusion) > 0 {
		g.ui.AddAnimations(consequencesOfConfusion)
		return
	}

	if enemy.HasFlag(foundation.IsScared) {
		if !sameRoom && rand.Intn(3) == 0 {
			enemy.GetFlags().Unset(foundation.IsScared)
			g.msg(foundation.HiLite("%s regains its courage", enemy.Name()))
		} else {
			newPos := g.gridMap.GetMoveOnPlayerDijkstraMap(enemy.Position(), false, g.playerDijkstraMap)
			consequencesOfMonsterMove := g.actorMoveAnimated(enemy, newPos)
			g.ui.AddAnimations(consequencesOfMonsterMove)
			return
		}
	}

	if customBehaviour, exists := g.customBehaviours(enemy.GetInternalName()); exists {
		customBehaviour(enemy)
	} else {
		g.defaultBehaviour(enemy)
	}
}

func (g *GameState) defaultBehaviour(enemy *Actor) {
	distanceToPlayer := g.gridMap.MoveDistance(enemy.Position(), g.Player.Position())

	sameRoom := g.isInPlayerRoom(enemy.Position()) || distanceToPlayer <= 1

	if distanceToPlayer <= 1 {
		consequencesOfMonsterAttack := g.actorMeleeAttack(enemy, NoModifiers, g.Player, NoModifiers)
		g.ui.AddAnimations(consequencesOfMonsterAttack)
		return
	}

	// has skills?
	zaps := enemy.GetIntrinsicZapEffects()
	canZap := len(zaps) > 0 && !enemy.HasFlag(foundation.IsCancelled)
	if canZap && sameRoom { //rand.Intn(3) == 0 {
		// zap
		zap := zaps[rand.Intn(len(zaps))]
		targetPos := g.Player.Position()
		consequencesOfMonsterZap := g.actorInvokeZapEffect(enemy, zap, targetPos)
		g.ui.AddAnimations(consequencesOfMonsterZap)
		return
	}

	aiUseEffects := enemy.GetIntrinsicUseEffects()
	canUse := len(aiUseEffects) > 0 && !enemy.HasFlag(foundation.IsCancelled)
	if canUse && sameRoom {
		useEffect := aiUseEffects[rand.Intn(len(aiUseEffects))]
		_, consequencesOfMonsterUseEffect := g.actorInvokeUseEffect(enemy, useEffect)
		g.ui.AddAnimations(consequencesOfMonsterUseEffect)
		return
	}

	gridMap := g.gridMap
	var newPos geometry.Point
	if !gridMap.IsTileWalkable(enemy.Position()) {
		newPos = gridMap.GetRandomFreeAndSafeNeighbor(rand.New(rand.NewSource(23)), enemy.Position())
	} else {
		newPos = gridMap.GetMoveOnPlayerDijkstraMap(enemy.Position(), true, g.playerDijkstraMap)
	}
	consequencesOfMonsterMove := g.actorMoveAnimated(enemy, newPos)
	g.ui.AddAnimations(consequencesOfMonsterMove)
}

func (g *GameState) doesActConfused(enemy *Actor) []foundation.Animation {
	if enemy.HasFlag(foundation.IsConfused) {
		if rand.Intn(6) == 0 {
			enemy.GetFlags().Unset(foundation.IsConfused)
		} else if rand.Intn(5) != 0 {
			actionDirection := geometry.RandomDirection()
			targetPos := enemy.Position().Add(actionDirection.ToPoint())
			if g.gridMap.IsActorAt(targetPos) {
				return g.actorMeleeAttack(enemy, ModFlatAndCap(-2, 10, "confused"), g.gridMap.ActorAt(targetPos), NoModifiers)
			} else if g.gridMap.IsCurrentlyPassable(targetPos) {
				return g.actorMoveAnimated(enemy, targetPos)
			} else {
				return nil
			}
		}
	}
	return nil
}
