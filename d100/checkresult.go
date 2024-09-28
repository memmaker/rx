package d100

type CheckResult struct {
	Success bool // Did the check succeed?
	Crit    bool // Critical success or failure
	DieRoll int  // The result of the die roll (D100)
	Degrees int
}

func (r CheckResult) IsCriticalSuccess() bool {
	return r.Crit && r.Success
}
