package fsmai

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
