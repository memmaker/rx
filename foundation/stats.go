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
	HudFatiguePoints    HudValue = "FatiguePoints"
	HudFatiguePointsMax HudValue = "FatiguePointsMax"
	HudStrength         HudValue = "Strength"
	HudDexterity        HudValue = "Dexterity"
	HudIntelligence     HudValue = "Intelligence"
	HudActiveDefense    HudValue = "AD"
	HudDamageResistance HudValue = "DR"
	HudDungeonLevel     HudValue = "Dungeon Level"
	HudTurnsTaken       HudValue = "Turns Taken"
)
