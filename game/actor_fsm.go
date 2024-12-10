package game

import (
	"contractor/fsmai"
	"github.com/kelindar/binary"
)

type ActorBehavior interface {
	AssociatedState() fsmai.StateName
	WithInitEvent(event fsmai.TransitionEvent) ActorBehavior
	Init(state *GameState, actor *Actor)
	Execute(state *GameState, actor *Actor) (fsmai.TransitionEvent, int)
}

type BehaviorFactory func(state fsmai.StateName) ActorBehavior

type ActorFSM struct {
	transition      *fsmai.TransitionTable
	behaviorFactory BehaviorFactory
	gameState       *GameState
	actor           *Actor
	defaultState    fsmai.StateName
	currentBehavior ActorBehavior
}

func NewActorFSM(state *GameState, actor *Actor, defaultState fsmai.StateName, behaviorFactory BehaviorFactory) *ActorFSM {
	return &ActorFSM{
		transition:      fsmai.NewDefaultTransitionTable(defaultState),
		behaviorFactory: behaviorFactory,
		currentBehavior: behaviorFactory(defaultState),
		gameState:       state,
		actor:           actor,
		defaultState:    defaultState,
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
	p.defaultState = state
	return nil
}

func (p *ActorFSM) RestoreState(state *GameState, actor *Actor, defaultState fsmai.StateName, behaviorFactory BehaviorFactory) {
	currentState := p.defaultState
	p.transition = fsmai.NewDefaultTransitionTable(defaultState)
	p.behaviorFactory = behaviorFactory
	p.gameState = state
	p.actor = actor
	p.defaultState = defaultState
	p.currentBehavior = p.behaviorFactory(currentState)
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

func (p *ActorFSM) DefaultState() fsmai.StateName {
	return p.defaultState
}
