package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"RogueUI/fsmai"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

type Behaviour struct {
	StateName       fsmai.StateName
	BehaviourAction func(g *GameState, actor *Actor, event fsmai.TransitionEvent) (fsmai.TransitionEvent, int)
	InitAction      func(g *GameState, actor *Actor, event fsmai.TransitionEvent)
	InitEvent       fsmai.TransitionEvent
}

func (b *Behaviour) AssociatedState() fsmai.StateName {
	return b.StateName
}

func (b *Behaviour) Init(state *GameState, actor *Actor, event fsmai.TransitionEvent) {
	b.InitEvent = event
	if b.InitAction != nil {
		b.InitAction(state, actor, event)
	}
}

func (b *Behaviour) Execute(state *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
	return b.BehaviourAction(state, actor, b.InitEvent)
}

func DefaultBehaviorFactory(state fsmai.StateName) ActorBehavior {
	var BehaviorTable = map[fsmai.StateName]ActorBehavior{
		fsmai.StateNeutral: &Behaviour{
			StateName:       fsmai.StateNeutral,
			BehaviourAction: BehaviourNeutralIdle,
			InitAction:      BehaviourNeutralIdleInit,
		},
		fsmai.StateAggressive: &Behaviour{
			StateName:       fsmai.StateAggressive,
			BehaviourAction: BehaviourAggressiveIdle,
			InitAction:      BehaviourAggressiveIdleInit,
		},
		fsmai.StateKill: &Behaviour{
			StateName:       fsmai.StateKill,
			BehaviourAction: BehaviourKill,
			InitAction:      BehaviourKillInit,
		},
		fsmai.StatePanic: &Behaviour{
			StateName:       fsmai.StatePanic,
			BehaviourAction: BehaviourPanic,
			InitAction:      BehaviourPanicInit,
		},
	}
	return BehaviorTable[state]
}

func (g *GameState) TryAIAction(enemy *Actor) int {
	enemy.GetFlags().Increment(foundation.FlagTurnsSinceLastIdleChatter)

	if enemy.timeEnergy <= 0 || enemy.timeEnergy < enemy.maximalTimeNeededForActions() {
		return 0 // not enough time energy for any action, spend 0 to accumulate
	}

	return enemy.FSM.ExecuteBehavior()

	// Status Effects
	if enemy.HasFlag(foundation.FlagStun) {
		stunCounter := enemy.GetFlags().Get(foundation.FlagStun)
		if stunCounter == 1 {
			//g.msg(foundation.HiLite("%s is stunned", enemy.Name()))
			enemy.GetFlags().Increment(foundation.FlagStun)
			return enemy.timeEnergy
		} else {
			if true { // result.IsFailure() { TODO
				enemy.GetFlags().Increment(foundation.FlagStun)
				return enemy.timeEnergy
			}
			g.msg(foundation.HiLite("%s clears its mind", enemy.Name()))
		}
	}

	if enemy.HasFlag(foundation.FlagHeld) {
		if rand.Intn(10) == 0 {
			enemy.GetFlags().Unset(foundation.FlagHeld)
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
		if nearEachOther && g.canPlayerSee(enemy.Position()) && enemy.chatterFile != "" && enemy.GetFlags().Get(foundation.FlagTurnsSinceLastIdleChatter) > 40 && rand.Intn(4) == 0 {
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

		if slot, move := enemy.MoveToNextTimeSlot(g.gameTime.Time); move {
			loc := g.currentMap().GetNamedLocation(slot.Location)
			g.msg(foundation.HiLite("%s moves to %s", enemy.Name(), slot.Location))
			enemy.SetGoal(GoalMoveToLocation(loc))
		}

		return enemy.timeEnergy // just wait and spend all time energy
	}

	if !enemy.IsHostileTowards(g.Player) {
		return enemy.timeEnergy
	}

	wantToChase := nearEachOther || enemy.HasFlag(foundation.FlagChase)
	if !wantToChase {
		return enemy.timeEnergy
	}

	return enemy.timeEnergy
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
			return g.actorMoveAnimated(enemy, targetPos)
		} else {
			return nil
		}
	}
	return nil
}
