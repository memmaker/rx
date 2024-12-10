package fsmai

type StateName int

const (
	StateNeutral StateName = iota
	StateAggressive
	StatePanic
	StateKill
	StateSearch
	StateDead
	StateCount
	// Also change NewTransitionTable() below, if you add new states at the end or the beginning
)

func (s StateName) ToString() string {
	switch s {
	case StateNeutral:
		return "Neutral"
	case StateAggressive:
		return "Aggressive"
	case StatePanic:
		return "Panic"
	case StateSearch:
		return "Search"
	case StateKill:
		return "Kill"
	case StateDead:
		return "Dead"
	default:
		return "Unknown"
	}
}
