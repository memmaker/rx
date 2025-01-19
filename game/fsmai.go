package game

type TransitionEventName int

const (
	EventNone TransitionEventName = iota
	EventProvoked
	EventCalmed
	EventHeavilyInjured
	EventEnemySighted
	EventLeaderJoined
	EventLeaderLeft
	EventTargetDied
	EventTargetLost
	EventApproach
	EventSuspiciousActivity
)

func (e TransitionEventName) ToString() string {
	switch e {
	case EventNone:
		return "None"
	case EventProvoked:
		return "Provoked"
	case EventHeavilyInjured:
		return "HeavilyInjured"
	case EventEnemySighted:
		return "EnemySighted"
	case EventTargetDied:
		return "TargetDied"
	case EventTargetLost:
		return "TargetLost"
	case EventApproach:
		return "Approach"
	case EventCalmed:
		return "Calmed"
	default:
		return "Unknown"
	}
}

type TransitionEvent interface {
	Name() TransitionEventName
}

var NoEvent EmptyEvent = EmptyEvent{EventNone}

type EmptyEvent struct {
	Event TransitionEventName
}

func (e EmptyEvent) Name() TransitionEventName {
	return e.Event
}

type StateName int

const (
	StateIdle StateName = iota
	StatePanic
	StateKill
	StateFollow
	StateHunt
	StateInvestigate
	StateScripted
	StateDead
	StateCount
	// Also change NewTransitionTable() below, if you add new states at the end or the beginning
)

func (s StateName) ToString() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StatePanic:
		return "Panic"
	case StateHunt:
		return "Hunt"
	case StateInvestigate:
		return "Investigate"
	case StateFollow:
		return "Follow"
	case StateScripted:
		return "Scripted"
	case StateKill:
		return "Kill"
	case StateDead:
		return "Dead"
	default:
		return "Unknown"
	}
}

type TransitionTable map[StateName]map[TransitionEventName]StateName

func NewTransitionTable() *TransitionTable {
	t := make(TransitionTable)
	for state := StateIdle; state < StateCount; state++ {
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
func contains(states []StateName, state StateName) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

func (t *TransitionTable) Exists(currentState StateName, event TransitionEventName) bool {
	_, ok := (*t)[currentState][event]
	return ok
}

func (t *TransitionTable) GetNextState(currentState StateName, event TransitionEventName) StateName {
	return (*t)[currentState][event]
}
