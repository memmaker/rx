package game

import (
    "RogueUI/foundation"
    "github.com/memmaker/go/geometry"
    "math/rand"
)

func (g *GameState) aiAct(enemy *Actor) {
    distanceToPlayer := geometry.DistanceChebyshev(enemy.Position(), g.Player.Position())

    isHostile := enemy.IsHostile()
    if !isHostile {
        return
    }
    sameRoom := g.isInPlayerRoom(enemy.Position()) || distanceToPlayer <= 1

    if enemy.HasFlag(foundation.FlagStun) {
        stunCounter := enemy.GetFlags().Get(foundation.FlagStun)
        if stunCounter == 1 {
            //g.msg(foundation.HiLite("%s is stunned", enemy.Name()))
            enemy.GetFlags().Increment(foundation.FlagStun)
            return
        } else {
            //turnMod := stunCounter - 1
            //_, result, _ := dice_curve.SuccessRoll(enemy.GetIntelligence() + turnMod)
            if true { // result.IsFailure() { TODO
                enemy.GetFlags().Increment(foundation.FlagStun)
                return
            }
            g.msg(foundation.HiLite("%s clears its mind", enemy.Name()))
        }
    }

    if enemy.HasFlag(foundation.FlagHeld) {
        if rand.Intn(10) == 0 {
            enemy.GetFlags().Unset(foundation.FlagHeld)
            g.msg(foundation.HiLite("%s breaks free", enemy.Name()))
        } else {
            return
        }
    }

    if enemy.IsSleeping() {
        if sameRoom && CanPerceive(enemy, g.Player) && rand.Intn(10) == 0 {
            enemy.WakeUp()
            g.ui.AddAnimations(OneAnimation(g.ui.GetAnimWakeUp(enemy.Position(), nil)))
            g.msg(foundation.HiLite("%s wakes up", enemy.Name()))
        } else {
            return
        }
    }

    if enemy.HasFlag(foundation.FlagCanConfuse) && rand.Intn(4) == 0 {
        enemy.GetFlags().Unset(foundation.FlagCanConfuse)
        g.msg(foundation.HiLite("%s stops glowing red", enemy.Name()))
    }

    consequencesOfConfusion := g.doesActConfused(enemy)
    if len(consequencesOfConfusion) > 0 {
        g.ui.AddAnimations(consequencesOfConfusion)
        return
    }

    if enemy.HasFlag(foundation.FlagScared) {
        if !sameRoom && rand.Intn(3) == 0 {
            enemy.GetFlags().Unset(foundation.FlagScared)
            g.msg(foundation.HiLite("%s regains its courage", enemy.Name()))
        } else {
            newPos := g.gridMap.GetMoveOnPlayerDijkstraMap(enemy.Position(), false, g.playerDijkstraMap)
            consequencesOfMonsterMove := g.actorMoveAnimated(enemy, newPos)
            g.ui.AddAnimations(consequencesOfMonsterMove)
            return
        }
    }

    losToPlayer := g.canPlayerSee(enemy.Position())
    if !enemy.HasFlag(foundation.FlagAwareOfPlayer) && sameRoom && losToPlayer && CanPerceive(enemy, g.Player) {
        enemy.GetFlags().Set(foundation.FlagAwareOfPlayer)
        g.msg(foundation.HiLite("%s notices you", enemy.Name()))
    }

    if !enemy.HasFlag(foundation.FlagAwareOfPlayer) {
        return
    }

    wantToChase := sameRoom || enemy.HasFlag(foundation.FlagChase)
    if !wantToChase {
        return
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
        consequencesOfMonsterAttack := g.actorMeleeAttack(enemy, g.Player)
        g.ui.AddAnimations(consequencesOfMonsterAttack)
        return
    }

    // has skills?
    zaps := enemy.GetIntrinsicZapEffects()
    canZap := len(zaps) > 0 && !enemy.HasFlag(foundation.FlagCancel)
    if canZap && sameRoom { //rand.Intn(3) == 0 {
        // zap
        zap := zaps[rand.Intn(len(zaps))]
        targetPos := g.Player.Position()
        consequencesOfMonsterZap := g.actorInvokeZapEffect(enemy, zap, targetPos)
        g.ui.AddAnimations(consequencesOfMonsterZap)
        return
    }

    aiUseEffects := enemy.GetIntrinsicUseEffects()
    canUse := len(aiUseEffects) > 0 && !enemy.HasFlag(foundation.FlagCancel)
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
    if enemy.HasFlag(foundation.FlagConfused) {
        if rand.Intn(6) == 0 {
            enemy.GetFlags().Unset(foundation.FlagConfused)
        } else if rand.Intn(5) != 0 {
            actionDirection := geometry.RandomDirection()
            targetPos := enemy.Position().Add(actionDirection.ToPoint())
            if g.gridMap.IsActorAt(targetPos) {
                return g.actorMeleeAttack(enemy, g.gridMap.ActorAt(targetPos))
            } else if g.gridMap.IsCurrentlyPassable(targetPos) {
                return g.actorMoveAnimated(enemy, targetPos)
            } else {
                return nil
            }
        }
    }
    return nil
}
