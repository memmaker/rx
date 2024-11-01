package foundation

type HudValue string

const (
	HudLevel           HudValue = "Level"
	HudExperience      HudValue = "Experience"
	HudGold            HudValue = "Gold"
	HudMeleeSkill      HudValue = "EffectiveMeleeSkill"
	HudRangedSkill     HudValue = "EffectiveRangedSkill"
	HudHitPoints       HudValue = "HitPoints"
	HudHitPointsMax    HudValue = "HitPointsMax"
	HudActionPoints    HudValue = "ActionPoints"
	HudActionPointsMax HudValue = "ActionPointsMax"
	HudStrength        HudValue = "Strength"
	HudDexterity       HudValue = "Dexterity"
	HudIntelligence    HudValue = "Intelligence"
	HudArmorString     HudValue = "DR"
	HudDungeonLevel    HudValue = "Dungeon Level"
	HudTurnsTaken      HudValue = "Turns Taken"
)

type HudValueMap map[HudValue]interface{}

func (hvm HudValueMap) GetInt(val HudValue) int {
	return hvm[val].(int)
}

func (hvm HudValueMap) GetString(val HudValue) string {
	return hvm[val].(string)
}

func (hvm HudValueMap) GetBool(val HudValue) bool {
	return hvm[val].(bool)
}

func (hvm HudValueMap) GetFloat(val HudValue) float64 {
	return hvm[val].(float64)
}
