package d100

import (
    "github.com/Knetic/govaluate"
    "strings"
)

func AddSkill(name string, short string, combatType CombatSkillType, baseValue string) {
    customSkills = append(customSkills, NewSkillDef(name, short, combatType, baseValue))
}
func LoadDerivedBaseValues(m map[string]string) {
    for derivedStat, baseValue := range m {
        expr, err := govaluate.NewEvaluableExpressionWithFunctions(baseValue, standardFunctions())
        if err != nil {
            panic(err)
        }
        derivedBaseValues[DerivedStatFromString(derivedStat)] = expr
    }
}

func LoadSkillMap(m map[string]string) {
    for skillFunction, skillName := range m {
        switch strings.ToLower(skillFunction) {
        case "unarmed":
            SkillForUnarmed = SkillFromString(skillName)
        case "repairs":
            SkillForRepairs = SkillFromString(skillName)
        case "picklocks":
            SkillForPickLocks = SkillFromString(skillName)
        case "pickpockets":
            SkillForPickPockets = SkillFromString(skillName)
        case "sneak":
            SkillForSneak = SkillFromString(skillName)
        case "hacking":
            SkillForHacking = SkillFromString(skillName)
        case "backstabbing":
            SkillForBackstabbing = SkillFromString(skillName)
        case "throwing":
            SkillForThrowing = SkillFromString(skillName)
        case "traps":
            SkillForTraps = SkillFromString(skillName)
        default:
            panic("Unknown skill function: " + skillFunction)
        }
    }
}

func LoadSkillDifficulties(diffs map[string]string) {
    for skillName, modExpr := range diffs {
        skill := DifficultyFromString(skillName)
        expr, err := govaluate.NewEvaluableExpressionWithFunctions(modExpr, standardFunctions())
        if err != nil {
            panic(err)
        }
        skillDiffs[skill] = expr
    }
}
