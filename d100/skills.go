package d100

func SkillFromString(name string) Skill {
	name = toTechnical(name)
	for i, skill := range customSkills {
		if skill.internalName() == name {
			return Skill(i)
		}
	}
	panic("Unknown skill: " + name)
	return -1
}

type Skill int

var customSkills = make([]SkillDef, 0)

// THESE NEED TO BE SET WHILE LOADING
var SkillForUnarmed Skill = -1
var SkillForRepairs Skill = -1
var SkillForPickLocks Skill = -1
var SkillForPickPockets Skill = -1
var SkillForSneak Skill = -1
var SkillForTraps Skill = -1
var SkillForHacking Skill = -1
var SkillForBackstabbing Skill = -1
var SkillForThrowing Skill = -1
var SkillForIntimidation Skill = -1
var SkillForRangedAttack Skill = -1
var SkillForSmoothTalking Skill = -1

func SkillCount() int {
	return len(customSkills)
}

func (s Skill) ToShortString() string {
	if int(s) < len(customSkills) {
		return customSkills[s].ToShortString()
	}
	return "N/A"
}

func (s Skill) ToAdjustmentString() string {
	if int(s) < len(customSkills) {
		return customSkills[s].ToAdjustmentString()
	}
	return "N/A"
}

func (s Skill) String() string {
	if int(s) < len(customSkills) {
		return customSkills[s].String()
	}
	return "N/A"
}

func (s Skill) IsRangedAttackSkill() bool {
	if int(s) < len(customSkills) {
		return customSkills[s].IsRangedAttackSkill()
	}
	return false
}

func (s Skill) IsMeleeAttackSkill() bool {
	if int(s) < len(customSkills) {
		return customSkills[s].IsMeleeAttackSkill()
	}
	return false
}

func (s Skill) BaseValue(params map[string]interface{}) int {
	if int(s) < len(customSkills) {
		return customSkills[s].BaseValue(params)
	}
	return 0
}

func (s Skill) Source() string {
	if int(s) < len(customSkills) {
		return customSkills[s].Source()
	}
	return "N/A"
}
