package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

func DefaultBehaviorFactory(state StateName) ActorBehavior {
	var BehaviorTable = map[StateName]ActorBehavior{
		StateIdle:        IdleBehaviour{},
		StateFollow:      FollowBehaviour{},
		StateScripted:    ScriptedBehaviour{},
		StateInvestigate: InvestigateBehaviour{},
		StateHunt:        HuntBehaviour{},
		StateKill:        KillBehaviour{},
		StatePanic:       PanicBehaviour{},
	}
	return BehaviorTable[state]
}

func (g *GameState) TryAIAction(enemy *Actor) int {
	enemy.GetFlags().Increment(foundation.FlagTurnsSinceLastIdleChatter)

	if enemy.RawTimeEnergy <= 0 || enemy.RawTimeEnergy < enemy.maximalTimeNeededForActions() {
		return 0 // not enough time energy for any action, spend 0 to accumulate
	}

	if enemy.HasFlag(foundation.FlagKnockedDown) {
		if enemy.GetFlags().Decrement(foundation.FlagKnockedDown) {
			return enemy.RawTimeEnergy
		} else {
			g.msg(foundation.HiLite("%s is standing up", enemy.Name()))
			standupTime := enemy.TimeNeededForActions() // get this before resetting knocked down flag
			return standupTime
		}
	}

	if enemy.HasFlag(foundation.FlagHeld) {
		if rand.Intn(10) == 0 {
			enemy.GetFlags().Unset(foundation.FlagHeld)
			g.msg(foundation.HiLite("%s breaks free", enemy.Name()))
		} else {
			return enemy.RawTimeEnergy
		}
	}

	if enemy.IsSleeping() {
		return enemy.RawTimeEnergy
	}

	return enemy.FSM.ExecuteBehavior()

	// Status Effects
	if enemy.HasFlag(foundation.FlagStun) {
		stunCounter := enemy.GetFlags().Get(foundation.FlagStun)
		if stunCounter == 1 {
			//g.msg(foundation.HiLite("%s is stunned", enemy.Name()))
			enemy.GetFlags().Increment(foundation.FlagStun)
			return enemy.RawTimeEnergy
		} else {
			if true { // result.IsFailure() { TODO
				enemy.GetFlags().Increment(foundation.FlagStun)
				return enemy.RawTimeEnergy
			}
			g.msg(foundation.HiLite("%s clears its mind", enemy.Name()))
		}
	}

	distanceToPlayer := geometry.DistanceChebyshev(enemy.Position(), g.Player.Position())

	nearEachOther := distanceToPlayer <= 7

	if enemy.HasFlag(foundation.FlagConfused) {
		consequencesOfConfusion := g.actConfused(enemy)
		if len(consequencesOfConfusion) > 0 {
			g.ui.AddAnimations(consequencesOfConfusion)
			return enemy.maximalTimeNeededForActions()
		}
	}

	inCombat := enemy.IsInCombat()
	if !inCombat {

		// IDLE STUFF HERE
		if nearEachOther && g.Player.CanSee(enemy.Position()) && enemy.ChatterFile != "" && enemy.GetFlags().Get(foundation.FlagTurnsSinceLastIdleChatter) > 40 && rand.Intn(4) == 0 {
			if g.tryAddRandomChatter(enemy, foundation.ChatterBeingAroundPlayer) {
				enemy.GetFlags().Unset(foundation.FlagTurnsSinceLastIdleChatter)
			}
		}

		if enemy.HasFlag(foundation.FlagAnimal) {
			consequencesOfConfusion := g.actConfused(enemy)
			if len(consequencesOfConfusion) > 0 {
				g.ui.AddAnimations(consequencesOfConfusion)
				return enemy.maximalTimeNeededForActions()
			}
		}

		return enemy.RawTimeEnergy // just wait and spend all time energy
	}

	if !enemy.IsHostileTowards(g.Player) {
		return enemy.RawTimeEnergy
	}

	wantToChase := nearEachOther || enemy.HasFlag(foundation.FlagChase)
	if !wantToChase {
		return enemy.RawTimeEnergy
	}

	return enemy.RawTimeEnergy
}

func (g *GameState) actConfused(enemy *Actor) []foundation.Animation {
	if rand.Intn(6) == 0 {
		enemy.GetFlags().Unset(foundation.FlagConfused)
	} else if rand.Intn(5) != 0 {
		actionDirection := geometry.RandomDirection()
		targetPos := enemy.Position().Add(actionDirection.ToPoint())
		if g.currentMap().IsActorAt(targetPos) {
			return g.actorMeleeAttack(enemy, g.currentMap().ActorAt(targetPos), d100.Body, d100.NoCombatModifier)
		} else if g.currentMap().IsCurrentlyPassable(targetPos) {
			return g.actorMove(enemy, targetPos)
		} else {
			return nil
		}
	}
	return nil
}
