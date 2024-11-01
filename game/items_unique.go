package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"github.com/memmaker/go/fxtools"
)

func (g *GameState) NewBowelDisruptor() *Weapon {
	return &Weapon{
		GenericItem: &GenericItem{
			name:             "bowel disruptor",
			internalName:     "bowel_disruptor",
			icon:             g.iconForItem(foundation.ItemCategoryWeapons),
			category:         foundation.ItemCategoryWeapons,
			qualityInPercent: 100,
			zapEffectName:    "disrupt_bowels",
			stackSize:        1,
			charges:          -1,
			statChanges: StatChange{
				SkillChanges: map[d100.Skill]int{
					d100.SkillForIntimidation: 15,
				},
			},
			thrownDamage:         fxtools.Interval{Min: 2, Max: 4},
			text:                 "Will cause instant and painful loss of bowel control. It is an illegal weapon and is completely untraceable.",
			chanceToBreakOnThrow: 10,
			setFlagOnPickup:      "UniqueTaken(bowel_disruptor)",
			weight:               4,
			cost:                 1000,
			alive:                true,
		},
		damageDice:       fxtools.Interval{Min: 1, Max: 6},
		weaponType:       WeaponTypePistol,
		skillUsed:        d100.SkillForRangedAttack,
		magazineSize:     20,
		loadedInMagazine: nil,
		PelletCount:      1,
		burstRounds:      1,
		caliberIndex:     4,
		caliberName:      "Micro Fusion Cell",
		attackModes: []AttackMode{
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
		soundID:       70,
		damageType:    DamageTypeElectrical,
		MinSTR:        2,
		reliability:   99,
		relativeSize:  SizeHandWeapon,
		degradeFactor: 0.2,
		jammed:        false,
	}
}
