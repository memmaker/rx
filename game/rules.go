package game

import (
	"contractor/d100"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"path/filepath"
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
	rulesFile := filepath.Join(definitionDirectory, "rules.rec")
	rulesRecords, _ := recfile.ReadMultiAndClose(fxtools.MustOpen(rulesFile))

	// Skill Definitions are mandatory
	skillDefs := rulesRecords["Skills"]
	for _, skillDef := range skillDefs {
		d100.AddSkill(SkillDefFromRecord(skillDef))
	}

	skillMap := rulesRecords["SkillMap"][0].ToMap(",")
	d100.LoadSkillMap(skillMap)

	// Perk Requirements are mandatory
	perkRequirements := rulesRecords["CharacterRequirement"]
	perkReqMap := make(map[d100.Perk]d100.CharacterRequirement)
	for _, perkReqRecord := range perkRequirements {
		perkName := d100.PerkFromString(perkReqRecord.FindValueForKeyIgnoreCase("perk"))
		reqs := NewPerkRequirements(perkReqRecord)
		perkReqMap[perkName] = reqs
	}
	d100.LoadPerkRequirements(perkReqMap)

	// The rest are optional
	if _, ok := rulesRecords["Global"]; ok {
		globalRules := rulesRecords["Global"][0].ToMap(",")

		d100.SkillCap = globalRules.GetIntOrDefault("SkillCap", 200)
		d100.ChanceForCriticalFailure = globalRules.GetIntOrDefault("ChanceForCriticalFailure", 5)
		d100.SuccessChanceCap = globalRules.GetIntOrDefault("SuccessChanceCap", 95)
		d100.LockStrengthReductionPerSkill = globalRules.GetFloatOrDefault("LockStrengthReductionPerSkill", 0.375)

		maxArmorDR = globalRules.GetIntOrDefault("MaxArmorDR", 25)
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

func NewPerkRequirements(record recfile.Record) d100.CharacterRequirement {
	reqs := d100.CharacterRequirement{}
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "requirelevel":
			reqs.Level = field.AsInt()
		case "requirestat":
			if fxtools.LooksLikeAFunction(field.Value) {
				if reqs.Stats == nil {
					reqs.Stats = make(map[d100.Stat]int)
				}

				name, args := fxtools.GetNameAndArgs(field.Value)
				stat := d100.StatFromString(name)
				reqs.Stats[stat] = args.GetInt(0)
			}
		case "requireskill":
			if fxtools.LooksLikeAFunction(field.Value) {
				if reqs.Skills == nil {
					reqs.Skills = make(map[d100.Skill]int)
				}

				name, args := fxtools.GetNameAndArgs(field.Value)
				skill := d100.SkillFromString(name)
				reqs.Skills[skill] = args.GetInt(0)
			}
		case "requirederivedstat":
			if fxtools.LooksLikeAFunction(field.Value) {
				if reqs.DerivedStats == nil {
					reqs.DerivedStats = make(map[d100.DerivedStat]int)
				}

				name, args := fxtools.GetNameAndArgs(field.Value)
				derivedStat := d100.DerivedStatFromString(name)
				reqs.DerivedStats[derivedStat] = args.GetInt(0)
			}
		case "requireperk":
			if fxtools.LooksLikeAFunction(field.Value) {
				if reqs.Perks == nil {
					reqs.Perks = make(map[d100.Perk]int)
				}

				name, args := fxtools.GetNameAndArgs(field.Value)
				perk := d100.PerkFromString(name)
				reqs.Perks[perk] = args.GetInt(0)
			}
		}
	}
	return reqs
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
