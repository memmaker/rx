package d100

import (
	"math/rand"
)

func SuccessRoll(chanceOfSuccess, successCritChange Percentage) CheckResult {
	var result CheckResult
	dieRoll := rand.Intn(100) + 1
	result.DieRoll = dieRoll

	critSucc := dieRoll <= int(successCritChange)
	critFail := dieRoll > (100 - ChanceForCriticalFailure)

	result.Success = dieRoll < int(chanceOfSuccess) || critSucc

	if result.Success {
		result.Crit = critSucc
		result.Degrees = int(chanceOfSuccess) - dieRoll
	} else {
		result.Crit = critFail
		result.Degrees = dieRoll - int(chanceOfSuccess)
	}

	return result
}

// SkillContest returns 0 if the first actor wins, 1 if the second actor wins, or a random number if the contest is a tie.
func SkillContest(firstActor, firstCritChance, secondActor, secondCritChance Percentage) int {
	firstRoll := SuccessRoll(firstActor, firstCritChance)
	secondRoll := SuccessRoll(secondActor, secondCritChance)
	maxTries := 100
	for i := 0; i < maxTries; i++ {
		if (firstRoll.Success && !secondRoll.Success) || (firstRoll.IsCriticalSuccess() && !secondRoll.IsCriticalSuccess()) {
			return 0
		}
		if (!firstRoll.Success && secondRoll.Success) || (!firstRoll.IsCriticalSuccess() && secondRoll.IsCriticalSuccess()) {
			return 1
		}

		// blackjack tiebreaker
		if firstRoll.DieRoll > secondRoll.DieRoll {
			return 0
		} else if firstRoll.DieRoll < secondRoll.DieRoll {
			return 1
		}

		firstRoll = SuccessRoll(firstActor, firstCritChance)
		secondRoll = SuccessRoll(secondActor, secondCritChance)
	}
	return rand.Intn(2)
}
