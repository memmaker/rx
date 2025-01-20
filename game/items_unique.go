package game

import (
	"contractor/d100"
	"contractor/foundation"
	"contractor/gridmap"
	"github.com/memmaker/go/fxtools"
)

func (g *GameState) NewBowelDisruptor() *Weapon {
	return &Weapon{
		GenericItem: &GenericItem{
			UID:              gridmap.NextItemID(),
			DisplayName:      "bowel disruptor",
			InternalName:     "bowel_disruptor",
			Icon:             g.iconForItem(foundation.ItemCategoryWeapons),
			Category:         foundation.ItemCategoryWeapons,
			QualityInPercent: 100,
			ZapEffectName:    "disrupt_bowels",
			StackSize:        1,
			Charges:          -1,
			StatChanges: StatChange{
				SkillChanges: map[d100.Skill]int{
					d100.SkillForIntimidation: 15,
				},
			},
			ThrownDamage:         fxtools.Interval{Min: 2, Max: 4},
			LongDescription:      "Will cause instant and painful loss of bowel control. It is an illegal weapon and is completely untraceable.",
			ChanceToBreakOnThrow: 10,
			SetFlagOnPickup:      "UniqueTaken(bowel_disruptor)",
			Weight:               4,
			Cost:                 1000,
			Alive:                true,
		},
		DamageDice:       fxtools.Interval{Min: 1, Max: 6},
		WeaponType:       WeaponTypePistol,
		SkillUsed:        d100.SkillForRangedAttack,
		MagazineSize:     20,
		LoadedInMagazine: nil,
		PelletCount:      1,
		BurstRounds:      1,
		CaliberIndex:     4,
		CaliberName:      "Micro Fusion Cell",
		AttackModes: []AttackMode{
			{Mode: TargetingModeBDLoose, TUCost: 10, MaxRange: 10, IsAimed: false},
			{Mode: TargetingModeBDWatery, TUCost: 10, MaxRange: 10, IsAimed: false},
			{Mode: TargetingModeBDFiery, TUCost: 10, MaxRange: 10, IsAimed: false},
			{Mode: TargetingModeBDBurningAnalGeyser, TUCost: 12, MaxRange: 7, IsAimed: false},
			{Mode: TargetingModeBDRectalVolcano, TUCost: 12, MaxRange: 7, IsAimed: false},
			{Mode: TargetingModeBDProlapse, TUCost: 12, MaxRange: 7, IsAimed: false},
			{Mode: TargetingModeBDUnspeakableGutHorror, TUCost: 14, MaxRange: 5, IsAimed: false},
			{Mode: TargetingModeBDShatIntoUnconsciousness, TUCost: 14, MaxRange: 5, IsAimed: false},
			{Mode: TargetingModeBDFatalIntestinalMaelstrom, TUCost: 16, MaxRange: 5, IsAimed: false},
		},
		SoundID:       "plasma_pistol",
		DamageType:    DamageTypeElectrical,
		MinSTR:        2,
		Reliability:   99,
		RelativeSize:  SizeHandWeapon,
		DegradeFactor: 0.2,
		Jammed:        false,
	}
}
