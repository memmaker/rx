package game

import (
	"github.com/kelindar/binary"
)

func NewDefaultTransitionTable() *TransitionTable {
	t := NewTransitionTable()

	// how to leave idle
	t.AddTransition(StateIdle, EventHeavilyInjured, StatePanic)
	t.AddTransition(StateIdle, EventProvoked, StateKill)
	t.AddTransition(StateIdle, EventEnemySighted, StateKill)
	t.AddTransition(StateIdle, EventLeaderJoined, StateFollow)
	t.AddTransition(StateIdle, EventSuspiciousActivity, StateInvestigate)

	// investigating
	t.AddTransition(StateInvestigate, EventEnemySighted, StateKill)
	t.AddTransition(StateInvestigate, EventProvoked, StateKill)
	t.AddTransition(StateInvestigate, EventCalmed, StateIdle)

	// how to leave follow
	t.AddTransition(StateFollow, EventLeaderLeft, StateIdle)
	t.AddTransition(StateFollow, EventEnemySighted, StateKill)
	t.AddTransition(StateFollow, EventProvoked, StateKill)

	// hunting
	t.AddTransition(StateHunt, EventEnemySighted, StateKill)
	t.AddTransition(StateHunt, EventProvoked, StateKill)
	t.AddTransition(StateHunt, EventCalmed, StateIdle)

	// killing
	t.AddTransition(StateKill, EventHeavilyInjured, StatePanic)
	t.AddTransition(StateKill, EventTargetLost, StateHunt)
	t.AddTransition(StateKill, EventTargetDied, StateIdle)
	t.AddTransition(StateKill, EventCalmed, StateIdle)

	// panic
	t.AddTransition(StatePanic, EventCalmed, StateIdle)

	return t
}

type ActorBehavior interface {
	AssociatedState() StateName
	WithInitEvent(event TransitionEvent) ActorBehavior
	Init(state *GameState, actor *Actor)
	Execute(state *GameState, actor *Actor) (TransitionEvent, int)
	IsCombatBehavior() bool
	IsHostilityTowards(other *Actor) bool
}

type BehaviorFactory func(state StateName) ActorBehavior

type ActorFSM struct {
	transition      *TransitionTable
	behaviorFactory BehaviorFactory
	gameState       *GameState
	actor           *Actor
	currentBehavior ActorBehavior
}

func NewActorFSM(state *GameState, actor *Actor, behaviorFactory BehaviorFactory) *ActorFSM {
	return &ActorFSM{
		transition:      NewDefaultTransitionTable(),
		behaviorFactory: behaviorFactory,
		currentBehavior: behaviorFactory(StateIdle),
		gameState:       state,
		actor:           actor,
	}
}
func (p *ActorFSM) GobEncode() ([]byte, error) {
	state := p.currentBehavior.AssociatedState()
	return binary.Marshal(state)
}

func (p *ActorFSM) GobDecode(data []byte) error {
	var state StateName
	if err := binary.Unmarshal(data, &state); err != nil {
		return err
	}
	return nil
}

func (p *ActorFSM) RestoreState(state *GameState, actor *Actor, currentBehavior ActorBehavior, behaviorFactory BehaviorFactory) {
	p.transition = NewDefaultTransitionTable()
	p.behaviorFactory = behaviorFactory
	p.gameState = state
	p.actor = actor
	p.currentBehavior = currentBehavior
}

func (p *ActorFSM) ExecuteBehavior() int {
	eventFromCurrentState, timeUsed := p.currentBehavior.Execute(p.gameState, p.actor)
	p.SendEvent(eventFromCurrentState)
	return timeUsed
}

func (p *ActorFSM) SendEvent(eventFromCurrentState TransitionEvent) {
	currentState := p.currentBehavior.AssociatedState()
	if p.transition.Exists(currentState, eventFromCurrentState.Name()) {
		nextState := p.transition.GetNextState(currentState, eventFromCurrentState.Name())
		p.SetState(nextState, eventFromCurrentState)
	}
}

func (p *ActorFSM) SetState(nextState StateName, event TransitionEvent) {
	p.currentBehavior = p.behaviorFactory(nextState).WithInitEvent(event)
	p.currentBehavior.Init(p.gameState, p.actor)
}

func (p *ActorFSM) State() StateName {
	return p.currentBehavior.AssociatedState()
}

func (p *ActorFSM) IsInCombat() bool {
	return p.currentBehavior.IsCombatBehavior()
}

func (p *ActorFSM) IsHostileTowards(other *Actor) bool {
	return p.currentBehavior.IsHostilityTowards(other)
}
