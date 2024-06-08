package dice_curve

import "fmt"

func SuccessRoll(effectiveSkill int) (rollResult int, result ResultType, marginOfSuccess int) {
	roll := NewDice(3, 6, 0).Roll()

	isSuccess := roll <= effectiveSkill
	marginOfSuccess = effectiveSkill - roll

	var resultType ResultType
	if isSuccess {
		resultType = TypeSuccess
		if isCriticalSuccess(roll, effectiveSkill) {
			resultType = TypeCriticalSuccess
		}
	} else {
		resultType = TypeFail
		if isCriticalFailure(roll, effectiveSkill) {
			resultType = TypeCriticalFail
		}
	}

	return roll, resultType, marginOfSuccess
}

// SkillContest returns the winner of a skill contest between two contenders.
// Will return 0 on draw (shouldn't happen (often)), 1 if contenderOne wins, and 2 if contenderTwo wins.
func SkillContest(contenderOneEffectiveSkill, contenderTwoEffectiveSkill int) (winner int) {
	_, rOne, _ := SuccessRoll(contenderOneEffectiveSkill)
	_, rTwo, _ := SuccessRoll(contenderTwoEffectiveSkill)
	maxTries := 100
	for i := 0; i < maxTries; i++ {
		if rOne == rTwo {
			continue
		}
		if rOne.IsSuccess() && rTwo.IsFailure() {
			return 1
		}
		if rOne.IsFailure() && rTwo.IsSuccess() {
			return 2
		}
		_, rOne, _ = SuccessRoll(contenderOneEffectiveSkill)
		_, rTwo, _ = SuccessRoll(contenderTwoEffectiveSkill)
	}
	return 0
}

type ResultType int

func (t ResultType) IsFailure() bool {
	return t == TypeFail || t == TypeCriticalFail
}

func (t ResultType) IsCriticalSucces() bool {
	return t == TypeCriticalSuccess
}

func (t ResultType) IsSuccess() bool {
	return t == TypeSuccess || t == TypeCriticalSuccess
}

const (
	TypeFail ResultType = iota
	TypeSuccess
	TypeCriticalSuccess
	TypeCriticalFail
)

type AttackOutcome struct {
	TypeOfHit  ResultType
	DamageDone int

	WasActiveDefenseRolled bool
	AttackRoll             int
	AttackEffectiveSkill   int
	AttackMarginOfSuccess  int
	AttackSuccessful       bool
	DamageDice             Dice
	DefenseRoll            int
	DefenseEffectiveSkill  int
	DefenseMarginOfSuccess int
	DefenseSuccessful      bool
	DamageBeforeDR         int
	DefenseDR              int
	CritTableRoll          int
	CritEffectMessage      string
}

func (a AttackOutcome) IsHit() bool {
	return a.TypeOfHit == TypeSuccess || a.TypeOfHit == TypeCriticalSuccess
}

func (a AttackOutcome) IsCriticalHit() bool {
	return a.TypeOfHit == TypeCriticalSuccess
}

func (a AttackOutcome) IsMiss() bool {
	return a.TypeOfHit == TypeFail || a.TypeOfHit == TypeCriticalFail
}
func (a AttackOutcome) IsCriticalMiss() bool {
	return a.TypeOfHit == TypeCriticalFail
}

func (a AttackOutcome) HitTypeMessage() string {
	if a.IsCriticalHit() {
		return a.CritEffectMessage
	}
	if a.IsCriticalMiss() {
		return "Critical miss!"
	}
	if a.IsMiss() {
		return "Miss!"
	}

	return "Hit!"
}

func (a AttackOutcome) String(attackerName, defenderName string) []string {
	// tell explicitly what happened, who rolled against what and if he failed and by how much
	attSuccString := fmt.Sprintf("%t", a.AttackSuccessful)
	attackRoll := fmt.Sprintf("%s attacks %s: %d vs. %d => %s by (%d)", attackerName, defenderName, a.AttackRoll, a.AttackEffectiveSkill, attSuccString, a.AttackMarginOfSuccess)

	if !a.AttackSuccessful {
		return []string{attackRoll, a.HitTypeMessage()}
	}

	damageInfo := fmt.Sprintf("Damage rolled on %s: %d - %d DR => %d effective damage", a.DamageDice.ShortString(), a.DamageBeforeDR, a.DefenseDR, a.DamageDone)

	if !a.WasActiveDefenseRolled {
		return []string{attackRoll, damageInfo, a.HitTypeMessage()}
	}

	defSuccString := fmt.Sprintf("%t", a.DefenseSuccessful)
	defenseRoll := fmt.Sprintf("%s defense: %d vs. %d => %s by (%d)", defenderName, a.DefenseRoll, a.DefenseEffectiveSkill, defSuccString, a.DefenseMarginOfSuccess)

	if a.DefenseSuccessful {
		return []string{attackRoll, defenseRoll, a.HitTypeMessage()}
	} else {
		return []string{attackRoll, defenseRoll, damageInfo, a.HitTypeMessage()}
	}
}
func Attack(attackerEffectiveSkill int, attackerDamageDice Dice, activeDefenseScore int, damageResistance int) AttackOutcome {
	attackRoll, attackResult, attMarginOfSucc := SuccessRoll(attackerEffectiveSkill)
	if attackResult.IsFailure() {
		return AttackOutcome{
			TypeOfHit:              attackResult,
			WasActiveDefenseRolled: false,
			AttackSuccessful:       false,
			AttackRoll:             attackRoll,
			AttackEffectiveSkill:   attackerEffectiveSkill,
			AttackMarginOfSuccess:  attMarginOfSucc,
		}
	}

	rolledDefense := false
	var defenseRoll int
	var defenseSucceeds bool
	var defMarginOfSucc int

	if !attackResult.IsCriticalSucces() && activeDefenseScore >= 0 {
		rolledDefense = true
		var defenseResult ResultType
		if defenseRoll, defenseResult, defMarginOfSucc = SuccessRoll(activeDefenseScore); defenseSucceeds {
			hitType := TypeFail
			if defenseResult.IsCriticalSucces() {
				hitType = TypeCriticalFail
			}
			return AttackOutcome{
				TypeOfHit:              hitType,
				WasActiveDefenseRolled: rolledDefense,
				AttackRoll:             attackRoll,
				AttackEffectiveSkill:   attackerEffectiveSkill,
				AttackMarginOfSuccess:  attMarginOfSucc,
				AttackSuccessful:       attackResult.IsSuccess(),
				DefenseRoll:            defenseRoll,
				DefenseEffectiveSkill:  activeDefenseScore,
				DefenseMarginOfSuccess: defMarginOfSucc,
				DefenseSuccessful:      true,
			}
		}
	}

	damageMultiplier := 1.0
	var critTableRoll int
	var critMessage string
	maxDamage := attackRoll == 3
	hitType := TypeSuccess
	if attackResult.IsCriticalSucces() {
		hitType = TypeCriticalSuccess
		threeD6 := NewDice(3, 6, 0)
		critTableRoll = threeD6.Roll()

		switch critTableRoll {
		case 3:
			damageMultiplier = 3
			critMessage = "Triple damage!"
		case 4:
			damageResistance = damageResistance / 2
			critMessage = "Armor halved!"
		case 5:
			damageMultiplier = 2
			critMessage = "Double damage!"
		case 6:
			maxDamage = true
			critMessage = "Max damage!"
		case 7:
			damageMultiplier = 1.5 // TODO: inflict major wound/bleeding
		case 8:
			damageMultiplier = 1.5 // TODO: loss of limb
		case 12:
			damageMultiplier = 1.5 // TODO: drop equipment
		case 16:
			damageMultiplier = 2
			critMessage = "Double damage!"
		case 17:
			damageResistance = damageResistance / 2
			critMessage = "Armor halved!"
		case 18:
			damageMultiplier = 3
			critMessage = "Triple damage!"
		}
	}

	damageBeforeDR, damage := rollDamageWithMultiplier(maxDamage, damageMultiplier, attackerDamageDice, damageResistance)
	return AttackOutcome{
		TypeOfHit:              hitType,
		WasActiveDefenseRolled: rolledDefense,
		AttackRoll:             attackRoll,
		AttackEffectiveSkill:   attackerEffectiveSkill,
		AttackMarginOfSuccess:  attMarginOfSucc,
		AttackSuccessful:       attackResult.IsSuccess(),
		DamageDice:             attackerDamageDice,
		DefenseRoll:            defenseRoll,
		DefenseEffectiveSkill:  activeDefenseScore,
		DefenseMarginOfSuccess: defMarginOfSucc,
		DefenseSuccessful:      false,
		DamageBeforeDR:         damageBeforeDR,
		DefenseDR:              damageResistance,
		DamageDone:             damage,
		CritTableRoll:          critTableRoll,
		CritEffectMessage:      critMessage,
	}
}

func AttackAgainstDoubleActiveDefense(attackerEffectiveSkill int, attackerDamageDice Dice, activeDefenseScores [2]int, damageResistance int) (ResultType, int) {
	attackRoll, attack, _ := SuccessRoll(attackerEffectiveSkill)
	if attack.IsFailure() {
		return attack, 0
	}

	isCrit := attack.IsCriticalSucces()

	if !isCrit {
		if _, defense, _ := SuccessRoll(activeDefenseScores[0]); defense.IsSuccess() {
			return defense, 0
		}
		if _, defense, _ := SuccessRoll(activeDefenseScores[1]); defense.IsSuccess() {
			return defense, 0
		}
	}

	_, damage := rollDamage(attackRoll, attackerDamageDice, damageResistance)

	return attack, damage
}

func checkDefenseCritSuccess(activeDefenseScore int, defRoll int) ResultType {
	hitType := TypeFail
	if isCriticalSuccess(activeDefenseScore, defRoll) {
		hitType = TypeCriticalFail
	}
	return hitType
}

func checkHitCritFail(attackerEffectiveSkill int, attackRoll int) ResultType {
	hitType := TypeFail
	if isCriticalFailure(attackerEffectiveSkill, attackRoll) {
		hitType = TypeCriticalFail
	}
	return hitType
}

func isCriticalSuccess(effectiveSkill int, roll int) bool {
	return roll <= 4 ||
		(effectiveSkill == 15 && roll <= 5) ||
		(effectiveSkill >= 16 && roll <= 6)
}

func isCriticalFailure(effectiveSkill int, roll int) bool {
	return roll >= 18 ||
		(effectiveSkill >= 7 && roll >= 17) ||
		(effectiveSkill == 6 && roll >= 16) ||
		(effectiveSkill == 5 && roll >= 15) ||
		(effectiveSkill == 4 && roll >= 14) ||
		(effectiveSkill == 3 && roll >= 13)
}

func rollDamage(attackRoll int, attackerDamageDice Dice, damageResistance int) (rawDamage int, damageAfterDR int) {
	willDoMaxDamage := attackRoll == 3

	var damage int
	if willDoMaxDamage {
		damage = attackerDamageDice.Max()
	} else {
		damage = attackerDamageDice.Roll()
	}

	return damage, max(1, damage-damageResistance)
}

func rollDamageWithMultiplier(willDoMaxDamage bool, multiplier float64, attackerDamageDice Dice, damageResistance int) (rawDamage int, damageAfterDR int) {
	var damage int
	if willDoMaxDamage {
		damage = int(float64(attackerDamageDice.Max()) * multiplier)
	} else {
		damage = int(float64(attackerDamageDice.Roll()) * multiplier)
	}

	return damage, max(1, damage-damageResistance)
}

type RollModifier int

const (
	NoModifier                                               RollModifier = 0
	AttackingWhileMoving                                     RollModifier = -4
	AttackingWithBadFooting                                  RollModifier = -2
	AttackingWithMajorDistraction                            RollModifier = -3
	AttackingWithMinorDistraction                            RollModifier = -2
	AttackingWithWeaponThatRequiresMoreStrengthPerDifference RollModifier = -1
	AttackingWhileBlind                                      RollModifier = -10
	AttackingWhileFoeHardToSee                               RollModifier = -5
	MeleeAttackingWithManeuverAllOutAttack                   RollModifier = 4
	MeleeAttackingWithLargeShield                            RollModifier = -2
	RangeAttackingWithManeuverAllOutAttack                   RollModifier = 1
	BaseModifierVeryEasyTask                                 RollModifier = 6
	BaseModifierEasyTask                                     RollModifier = 4
	BaseModifierNormalTask                                   RollModifier = 0
	BaseModifierHardTask                                     RollModifier = -4
	BaseModifierVeryHardTask                                 RollModifier = -6
)
