package game

import (
    "RogueUI/foundation"
    "RogueUI/special"
    "github.com/memmaker/go/geometry"
    "math/rand"
)

func (g *GameState) TryAIAction(enemy *Actor) int {
    if enemy.timeEnergy <= 0 {
        return 0 // not enough time energy for any action, spend 0 to accumulate
    }

    if enemy.HasFlag(special.FlagStun) {
        stunCounter := enemy.GetFlags().Get(special.FlagStun)
        if stunCounter == 1 {
            //g.msg(foundation.HiLite("%s is stunned", enemy.Name()))
            enemy.GetFlags().Increment(special.FlagStun)
            return enemy.timeEnergy
        } else {
            //turnMod := stunCounter - 1
            //_, result, _ := dice_curve.SuccessRoll(enemy.GetIntelligence() + turnMod)
            if true { // result.IsFailure() { TODO
                enemy.GetFlags().Increment(special.FlagStun)
                return enemy.timeEnergy
            }
            g.msg(foundation.HiLite("%s clears its mind", enemy.Name()))
        }
    }

    if enemy.HasFlag(special.FlagHeld) {
        if rand.Intn(10) == 0 {
            enemy.GetFlags().Unset(special.FlagHeld)
            g.msg(foundation.HiLite("%s breaks free", enemy.Name()))
        } else {
            return enemy.timeEnergy
        }
    }

    distanceToPlayer := geometry.DistanceChebyshev(enemy.Position(), g.Player.Position())

    nearEachOther := distanceToPlayer <= 7

    if enemy.IsSleeping() {
        if nearEachOther && CanPerceive(enemy, g.Player) && rand.Intn(10) == 0 {
            enemy.WakeUp()
            g.ui.AddAnimations(OneAnimation(g.ui.GetAnimWakeUp(enemy.Position(), nil)))
            g.msg(foundation.HiLite("%s wakes up", enemy.Name()))
        } else {
            return enemy.timeEnergy
        }
    }

    if enemy.HasFlag(special.FlagCanConfuse) && rand.Intn(4) == 0 {
        enemy.GetFlags().Unset(special.FlagCanConfuse)
        g.msg(foundation.HiLite("%s stops glowing red", enemy.Name()))
    }

    if enemy.HasFlag(special.FlagConfused) {
        consequencesOfConfusion := g.actConfused(enemy)
        if len(consequencesOfConfusion) > 0 {
            g.ui.AddAnimations(consequencesOfConfusion)
            return enemy.maximalTimeNeededForActions()
        }
    }

    if enemy.HasFlag(special.FlagScared) {
        if !nearEachOther && rand.Intn(3) == 0 {
            enemy.GetFlags().Unset(special.FlagScared)
            g.msg(foundation.HiLite("%s regains its courage", enemy.Name()))
        } else {
            newPos := g.currentMap().GetMoveOnPlayerDijkstraMap(enemy.Position(), false, g.playerDijkstraMap)
            consequencesOfMonsterMove := g.actorMoveAnimated(enemy, newPos)
            g.ui.AddAnimations(consequencesOfMonsterMove)
            return enemy.timeNeededForMovement()
        }
    }

    if enemy.HasActiveGoal() {
        return enemy.ActOnGoal(g)
    }
    inCombat := enemy.IsInCombat()
    if !inCombat {
        if enemy.HasFlag(special.FlagAnimal) {
            consequencesOfConfusion := g.actConfused(enemy)
            if len(consequencesOfConfusion) > 0 {
                g.ui.AddAnimations(consequencesOfConfusion)
                return enemy.maximalTimeNeededForActions()
            }
        }

        return enemy.timeEnergy // just wait and spend all time energy
    }

    losToPlayer := g.canPlayerSee(enemy.Position())
    if !enemy.HasFlag(special.FlagAwareOfPlayer) && nearEachOther && losToPlayer && CanPerceive(enemy, g.Player) {
        enemy.GetFlags().Set(special.FlagAwareOfPlayer)
        g.msg(foundation.HiLite("%s notices you", enemy.Name()))
    }

    if !enemy.HasFlag(special.FlagAwareOfPlayer) {
        return enemy.timeEnergy
    }

    if !enemy.IsHostileTowards(g.Player) {
        return enemy.timeEnergy
    }

    wantToChase := nearEachOther || enemy.HasFlag(special.FlagChase)
    if !wantToChase {
        return enemy.timeEnergy
    }

    if customBehaviour, exists := g.customBehaviours(enemy.GetInternalName()); exists {
        return customBehaviour(enemy)
    } else if enemy.HasActiveGoal() {
        return enemy.ActOnGoal(g)
    } else {
        return g.defaultBehaviour(enemy)
    }
}

func (g *GameState) defaultBehaviour(enemy *Actor) int {
    distanceToPlayer := g.currentMap().MoveDistance(enemy.Position(), g.Player.Position())

    sameRoom := distanceToPlayer <= 1

    rangedWeapon, hasRangedWeapon := enemy.GetEquipment().GetRangedWeapon()
    if hasRangedWeapon {
        attackMode := rangedWeapon.GetCurrentAttackMode()
        weaponRange := attackMode.MaxRange - 1
        if distanceToPlayer <= weaponRange && g.canPlayerSee(enemy.Position()) {
            consequencesOfMonsterRangedAttack := g.actorRangedAttack(enemy, rangedWeapon, attackMode, g.Player, 0)
            g.ui.AddAnimations(consequencesOfMonsterRangedAttack)
            return attackMode.TUCost
        }
    }

    if distanceToPlayer <= 1 {
        consequencesOfMonsterAttack := g.actorMeleeAttack(enemy, g.Player, 0)
        g.ui.AddAnimations(consequencesOfMonsterAttack)
        return enemy.GetMeleeTUCost()
    }

    // has skills?
    zaps := enemy.GetIntrinsicZapEffects()
    canZap := len(zaps) > 0 && !enemy.HasFlag(special.FlagCancel)
    if canZap && sameRoom { //rand.Intn(3) == 0 {
        // zap
        zap := zaps[rand.Intn(len(zaps))]
        targetPos := g.Player.Position()
        consequencesOfMonsterZap := g.actorInvokeZapEffect(enemy, zap, targetPos)
        g.ui.AddAnimations(consequencesOfMonsterZap)
        return enemy.timeNeededForActions()
    }

    aiUseEffects := enemy.GetIntrinsicUseEffects()
    canUse := len(aiUseEffects) > 0 && !enemy.HasFlag(special.FlagCancel)
    if canUse && sameRoom {
        useEffect := aiUseEffects[rand.Intn(len(aiUseEffects))]
        _, consequencesOfMonsterUseEffect := g.actorInvokeUseEffect(enemy, useEffect)
        g.ui.AddAnimations(consequencesOfMonsterUseEffect)
        return enemy.timeNeededForActions()
    }

    gridMap := g.currentMap()
    var newPos geometry.Point
    if !gridMap.IsTileWalkable(enemy.Position()) {
        newPos = gridMap.GetRandomFreeAndSafeNeighbor(rand.New(rand.NewSource(23)), enemy.Position())
    } else {
        newPos = gridMap.GetMoveOnPlayerDijkstraMap(enemy.Position(), true, g.playerDijkstraMap)
    }
    consequencesOfMonsterMove := g.actorMoveAnimated(enemy, newPos)
    g.ui.AddAnimations(consequencesOfMonsterMove)

    return enemy.timeNeededForMovement()
}

func (g *GameState) actConfused(enemy *Actor) []foundation.Animation {
    if rand.Intn(6) == 0 {
        enemy.GetFlags().Unset(special.FlagConfused)
    } else if rand.Intn(5) != 0 {
        actionDirection := geometry.RandomDirection()
        targetPos := enemy.Position().Add(actionDirection.ToPoint())
        if g.currentMap().IsActorAt(targetPos) {
            return g.actorMeleeAttack(enemy, g.currentMap().ActorAt(targetPos), 0)
        } else if g.currentMap().IsCurrentlyPassable(targetPos) {
            return g.actorMoveAnimated(enemy, targetPos)
        } else {
            return nil
        }
    }
    return nil
}

func (g *GameState) customBehaviours(internalName string) (func(actor *Actor) int, bool) {
    switch internalName {
    case "xeroc_2":
        return nil, false
    }
    return nil, false

}
