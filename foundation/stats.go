package foundation

type HudValue string

const (
	HudLevel            HudValue = "Level"
	HudExperience       HudValue = "Experience"
	HudGold             HudValue = "Gold"
	HudMeleeSkill       HudValue = "EffectiveMeleeSkill"
	HudRangedSkill      HudValue = "EffectiveRangedSkill"
	HudHitPoints        HudValue = "HitPoints"
	HudHitPointsMax     HudValue = "HitPointsMax"
	HudActionPoints     HudValue = "ActionPoints"
	HudActionPointsMax  HudValue = "ActionPointsMax"
	HudStrength         HudValue = "Strength"
	HudDexterity        HudValue = "Dexterity"
	HudIntelligence     HudValue = "Intelligence"
	HudDamageResistance HudValue = "DR"
	HudDungeonLevel     HudValue = "Dungeon Level"
	HudTurnsTaken       HudValue = "Turns Taken"
)
