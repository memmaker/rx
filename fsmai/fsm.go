package fsmai

func contains(states []StateName, state StateName) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}
