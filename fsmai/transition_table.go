package fsmai

func NewDefaultTransitionTable(defaultState StateName) *TransitionTable {
	t := NewTransitionTable()

	// leaving neutral
	t.AddTransition(defaultState, EventHeavilyInjured, StatePanic)
	t.AddTransition(defaultState, EventProvoked, StateKill)
	t.AddTransition(StateAggressive, EventEnemySighted, StateKill)

	// killing
	t.AddTransition(StateKill, EventHeavilyInjured, StatePanic)
	t.AddTransition(StateKill, EventTargetLost, defaultState)
	t.AddTransition(StateKill, EventTargetDied, defaultState)
	t.AddTransition(StateKill, EventCalmed, defaultState)

	// panic
	t.AddTransition(StatePanic, EventThreatNeutralized, defaultState)

	return t
}

var NeutralActorTransitionTable = NewDefaultTransitionTable(StateNeutral)
var AggressiveActorTransitionTable = NewDefaultTransitionTable(StateAggressive)

type TransitionTable map[StateName]map[TransitionEventName]StateName

func NewTransitionTable() *TransitionTable {
	t := make(TransitionTable)
	for state := StateNeutral; state < StateCount; state++ {
		t[state] = make(map[TransitionEventName]StateName)
	}
	return &t
}

func (t *TransitionTable) AddTransition(fromState StateName, event TransitionEventName, toState StateName) {
	(*t)[fromState][event] = toState
}

func (t *TransitionTable) AddTransitionFromAllExcept(excludedStates []StateName, event TransitionEventName, toState StateName) {
	for state := range *t {
		if !contains(excludedStates, state) {
			(*t)[state][event] = toState
		}
	}
}

func (t *TransitionTable) Exists(currentState StateName, event TransitionEventName) bool {
	_, ok := (*t)[currentState][event]
	return ok
}

func (t *TransitionTable) GetNextState(currentState StateName, event TransitionEventName) StateName {
	return (*t)[currentState][event]
}
