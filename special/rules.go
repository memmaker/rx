package special

import (
	"RogueUI/foundation"
	"math/rand"
)

type Percentage uint8

func SuccessRoll(chanceOfSuccess, critChance Percentage) foundation.CheckResult {
	roll := rand.Intn(100)
	var result foundation.CheckResult
	result.Success = roll < int(chanceOfSuccess)
	result.Crit = roll < int(critChance)
	return result
}

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
