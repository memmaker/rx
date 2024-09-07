package special

import (
	"fmt"
	"math/rand"
)

type CheckResult struct {
	Success bool
	Crit    bool
}

type Percentage int8

func (p Percentage) String() string {
	return fmt.Sprintf("%d%%", p)
}

func (p Percentage) AsFloat() float64 {
	return float64(p) / 100.0
}

func SuccessRoll(chanceOfSuccess, critChance Percentage) CheckResult {
	roll := rand.Intn(100)
	var result CheckResult
	result.Success = roll < int(chanceOfSuccess)
	result.Crit = roll < int(critChance)
	return result
}

// SkillContest returns 0 if the first actor wins, 1 if the second actor wins, or a random number if the contest is a tie.
func SkillContest(firstActor, firstCritChance, secondActor, secondCritChance Percentage) int {
	firstRoll := SuccessRoll(firstActor, firstCritChance)
	secondRoll := SuccessRoll(secondActor, secondCritChance)
	maxTries := 1000
	for i := 0; i < maxTries; i++ {
		if firstRoll.Success && !secondRoll.Success {
			return 0
		}
		if !firstRoll.Success && secondRoll.Success {
			return 1
		}
		firstRoll = SuccessRoll(firstActor, firstCritChance)
		secondRoll = SuccessRoll(secondActor, secondCritChance)
	}
	return rand.Intn(2)
}
