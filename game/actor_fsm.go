package game

import "RogueUI/fsmai"

type ActorBehavior interface {
	AssociatedState() fsmai.StateName
	Init(state *GameState, actor *Actor, event fsmai.TransitionEvent)
	Execute(state *GameState, actor *Actor) (fsmai.TransitionEvent, int)
}

type BehaviorFactory func(state fsmai.StateName) ActorBehavior

type ActorFSM struct {
	eventListener   func(event fsmai.TransitionEvent)
	currentBehavior ActorBehavior
	transition      *fsmai.TransitionTable
	behaviorFactory BehaviorFactory
	gameState       *GameState
	actor           *Actor
	defaultState    fsmai.StateName
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

func (p *ActorFSM) SetEventListener(listener func(event fsmai.TransitionEvent)) {
	p.eventListener = listener
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
	p.currentBehavior = p.behaviorFactory(nextState)
	p.currentBehavior.Init(p.gameState, p.actor, event)
}

func (p *ActorFSM) State() fsmai.StateName {
	return p.currentBehavior.AssociatedState()
}

func (p *ActorFSM) DefaultState() fsmai.StateName {
	return p.defaultState
}
