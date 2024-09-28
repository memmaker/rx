package game

import (
	"RogueUI/d100"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"path"
	"strconv"
	"strings"
)

func CanPerceive(observer *Actor, observed *Actor) bool {
	return true
	/*
	   distance := geometry.Distance(observer.Position(), observed.Position())
	   observedChar := observed.GetCharSheet()
	   observerChar := observer.GetCharSheet()

	   sneakSkill := observedChar.GetSkill(special.Sneak)
	   isObserverSneaking := observedChar.HasFlag(special.FlagSneaking)

	*/

}

func loadD100Rules(definitionDirectory string) {
	rulesFile := path.Join(definitionDirectory, "rules.rec")
	rulesRecords := recfile.ReadMulti(fxtools.MustOpen(rulesFile))

	// Skill Definitions are mandatory
	skillDefs := rulesRecords["Skills"]
	for _, skillDef := range skillDefs {
		d100.AddSkill(SkillDefFromRecord(skillDef))
	}

	skillMap := rulesRecords["SkillMap"][0].ToMap(",")
	d100.LoadSkillMap(skillMap)

	// The rest are optional
	if _, ok := rulesRecords["Global"]; ok {
		globalRules := rulesRecords["Global"][0].ToMap(",")

		d100.SkillCap = globalRules.GetIntOrDefault("SkillCap", 200)
		d100.ChanceForCriticalFailure = globalRules.GetIntOrDefault("ChanceForCriticalFailure", 5)
		d100.SuccessChanceCap = globalRules.GetIntOrDefault("SuccessChanceCap", 95)
		maxArmorDR = globalRules.GetIntOrDefault("MaxArmorDR", 85)
		maxArmorDT = globalRules.GetIntOrDefault("MaxArmorDT", 30)
		ablationWithoutPenetration = globalRules.GetIntOrDefault("AblationWithoutPenetration", 1)
		ablationWithPenetration = globalRules.GetIntOrDefault("AblationWithPenetration", 2)
	}

	if _, ok := rulesRecords["SkillDifficulty"]; ok {
		skillDiffs := rulesRecords["SkillDifficulty"][0].ToMap(",")
		d100.LoadSkillDifficulties(skillDiffs)
	}

	if _, ok := rulesRecords["LevelTable"]; ok {
		levelUpTable := fxtools.MapSlice(rulesRecords["LevelTable"][0].ToValueList(), func(i string) int {
			val, _ := strconv.Atoi(i)
			return val
		})

		afterTable := rulesRecords["AfterTable"][0].FindValueForKeyIgnoreCase("xp")
		d100.LoadLevelUpTable(levelUpTable, afterTable)
	}

	if _, ok := rulesRecords["DerivedStats"]; ok {
		derivedStatsBaseValues := rulesRecords["DerivedStats"][0].ToMap(",")
		d100.LoadDerivedBaseValues(derivedStatsBaseValues)
	}
}

func SkillDefFromRecord(def recfile.Record) (string, string, d100.CombatSkillType, string) {
	var name string
	var shortName string
	var combat d100.CombatSkillType
	var baseValue string
	for _, field := range def {
		switch strings.ToLower(field.Name) {
		case "name":
			name = field.Value
		case "short":
			shortName = field.Value
		case "combat":
			combat = d100.CombatSkillTypeFromString(field.Value)
		case "basevalue":
			baseValue = field.Value
		}
	}
	return name, shortName, combat, baseValue
}
