package d100

import (
    "github.com/Knetic/govaluate"
    "strings"
)

type CombatSkillType uint8

const (
    SkillNonCombat CombatSkillType = iota
    SkillMelee
    SkillRanged
)

func CombatSkillTypeFromString(name string) CombatSkillType {
    switch strings.ToLower(name) {
    case "melee":
        return SkillMelee
    case "ranged":
        return SkillRanged
    default:
        return SkillNonCombat
    }
}

type SkillDef struct {
    displayName string
    shortName   string
    combat      CombatSkillType
    baseValue   *govaluate.EvaluableExpression
}

func NewSkillDef(displayName string, shortName string, combat CombatSkillType, baseValue string) SkillDef {
    expression, err := govaluate.NewEvaluableExpressionWithFunctions(baseValue, standardFunctions())
    if err != nil {
        panic(err)
    }
    return SkillDef{
        displayName: displayName,
        shortName:   shortName,
        combat:      combat,
        baseValue:   expression,
    }
}
func (c SkillDef) ToShortString() string {
    return c.shortName
}

func (c SkillDef) ToAdjustmentString() string {
    return c.internalName() + "_Adjustment"
}

func (c SkillDef) String() string {
    return c.displayName
}

func (c SkillDef) IsRangedAttackSkill() bool {
    return c.combat == SkillRanged
}

func (c SkillDef) IsMeleeAttackSkill() bool {
    return c.combat == SkillMelee
}

func (c SkillDef) BaseValue(params map[string]interface{}) int {
    evaluate, err := c.baseValue.Evaluate(params)
    if err != nil {
        panic(err)
    }
    return int(evaluate.(float64))
}

func (c SkillDef) internalName() string {
    return toTechnical(c.displayName)
}

func (c SkillDef) Source() string {
    return c.baseValue.String()
}

func toTechnical(name string) string {
    replacer := strings.NewReplacer("_", "", " ", "", "-", "", "*", "", "'", "", "(", "", ")", "", ",", "", ".", "", ":", "", "?", "", "!", "", "&", "", "/", "", "\\", "", "|", "", "<", "", ">", "", "=", "", "+", "", "#", "", "@", "", "$", "", "%", "", "^", "", "~", "", "`", "", ";", "")
    return strings.ToLower(replacer.Replace(name))
}
