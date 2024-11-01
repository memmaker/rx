package d100

import "math/rand"

var SkillCap = 200
var ChanceForCriticalFailure = 5
var SuccessChanceCap = 95
var LockStrengthReductionPerSkill = 0.375

func Die() int {
	return 1 + rand.Intn(100)
}

func ElectronicLockpicksNeeded(lockDifficulty Difficulty, actorSkill int) int {
	basePickCount := lockDifficulty.EPicksNeeded()
	// skill can range from 0% to 200%.
	// at 200% we want to have a 90% reduction in picks needed

	// skill = 20
	// 20 / 200 = 0.1
	// 0.1 * 0.9 = 0.09

	skillFactor := float64(actorSkill) / float64(SkillCap) // 0.0 to 1.0
	skillFactor = skillFactor * 0.9                        // 0.0 to 0.9

	picksReducedBy := float64(basePickCount) * skillFactor
	picksNeeded := max(1, basePickCount-int(picksReducedBy))
	return picksNeeded
}
