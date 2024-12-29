package game

import (
	"contractor/fsmai"
	"github.com/kelindar/binary"
)

func NewDefaultTransitionTable() *fsmai.TransitionTable {
	t := fsmai.NewTransitionTable()

	// how to leave idle
	t.AddTransition(fsmai.StateIdle, fsmai.EventHeavilyInjured, fsmai.StatePanic)
	t.AddTransition(fsmai.StateIdle, fsmai.EventProvoked, fsmai.StateKill)
	t.AddTransition(fsmai.StateIdle, fsmai.EventMajorCrimeWitnessed, fsmai.StateKill)
	t.AddTransition(fsmai.StateIdle, fsmai.EventEnemySighted, fsmai.StateKill)
	t.AddTransition(fsmai.StateIdle, fsmai.EventLeaderJoined, fsmai.StateFollow)

	// how to leave follow
	t.AddTransition(fsmai.StateFollow, fsmai.EventLeaderLeft, fsmai.StateIdle)
	t.AddTransition(fsmai.StateFollow, fsmai.EventProvoked, fsmai.StateKill)

	// searching
	t.AddTransition(fsmai.StateSearch, fsmai.EventEnemySighted, fsmai.StateKill)

	// killing
	t.AddTransition(fsmai.StateKill, fsmai.EventHeavilyInjured, fsmai.StatePanic)
	t.AddTransition(fsmai.StateKill, fsmai.EventTargetLost, fsmai.StateSearch)
	t.AddTransition(fsmai.StateKill, fsmai.EventTargetDied, fsmai.StateIdle)
	t.AddTransition(fsmai.StateKill, fsmai.EventCalmed, fsmai.StateIdle)

	// panic
	t.AddTransition(fsmai.StatePanic, fsmai.EventThreatNeutralized, fsmai.StateIdle)

	return t
}

type ActorBehavior interface {
	AssociatedState() fsmai.StateName
	WithInitEvent(event fsmai.TransitionEvent) ActorBehavior
	Init(state *GameState, actor *Actor)
	Execute(state *GameState, actor *Actor) (fsmai.TransitionEvent, int)
	IsCombatBehavior() bool
	IsHostilityTowards(other *Actor) bool
}

type BehaviorFactory func(state fsmai.StateName) ActorBehavior

type ActorFSM struct {
	transition      *fsmai.TransitionTable
	behaviorFactory BehaviorFactory
	gameState       *GameState
	actor           *Actor
	currentBehavior ActorBehavior
}

func NewActorFSM(state *GameState, actor *Actor, behaviorFactory BehaviorFactory) *ActorFSM {
	return &ActorFSM{
		transition:      NewDefaultTransitionTable(),
		behaviorFactory: behaviorFactory,
		currentBehavior: behaviorFactory(fsmai.StateIdle),
		gameState:       state,
		actor:           actor,
	}
}
func (p *ActorFSM) GobEncode() ([]byte, error) {
	state := p.currentBehavior.AssociatedState()
	return binary.Marshal(state)
}

func (p *ActorFSM) GobDecode(data []byte) error {
	var state fsmai.StateName
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

func (p *ActorFSM) SendEvent(eventFromCurrentState fsmai.TransitionEvent) {
	currentState := p.currentBehavior.AssociatedState()
	if p.transition.Exists(currentState, eventFromCurrentState.Name()) {
		nextState := p.transition.GetNextState(currentState, eventFromCurrentState.Name())
		p.SetState(nextState, eventFromCurrentState)
	}
}

func (p *ActorFSM) SetState(nextState fsmai.StateName, event fsmai.TransitionEvent) {
	p.currentBehavior = p.behaviorFactory(nextState).WithInitEvent(event)
	p.currentBehavior.Init(p.gameState, p.actor)
}

func (p *ActorFSM) State() fsmai.StateName {
	return p.currentBehavior.AssociatedState()
}

func (p *ActorFSM) IsInCombat() bool {
	return p.currentBehavior.IsCombatBehavior()
}

func (p *ActorFSM) IsHostileTowards(other *Actor) bool {
	return p.currentBehavior.IsHostilityTowards(other)
}
