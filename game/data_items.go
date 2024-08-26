package game

import (
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"RogueUI/special"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"strings"
)

type ItemDef struct {
	Description string
	Name        string

	WeaponDef WeaponDef
	ArmorDef  ArmorDef
	AmmoDef   AmmoDef

	ThrowDamageDice dice_curve.Dice

	UseEffect string
	ZapEffect string

	Stat      dice_curve.Stat
	StatBonus dice_curve.Dice

	Charges  dice_curve.Dice
	Category foundation.ItemCategory

	AlwaysIDOnUse bool
	EquipFlag     foundation.ActorFlag

	Skill                dice_curve.SkillName
	SkillBonus           dice_curve.Dice
	Tags                 foundation.ItemTags
	Position             geometry.Point
	TextFile             string
	Text                 string
	LockFlag             string
	Size                 int
	Cost                 int
	Weight               int
	ChanceToBreakOnThrow int
	SetFlagOnPickup      string
}

func (i ItemDef) IsValidArmor() bool {
	return i.ArmorDef.IsValid()
}

func (i ItemDef) IsValidWeapon() bool {
	return i.WeaponDef.IsValid()
}

func (i ItemDef) IsValidAmmo() bool {
	return i.AmmoDef.IsValid()
}

func (i ItemDef) GetAttackModes() []AttackMode {
	noAim := i.Tags.Contains(foundation.TagNoAim)
	var modes []AttackMode
	if i.WeaponDef.TargetingModeOne != special.TargetingModeNone {
		modes = append(modes, AttackMode{
			Mode:     i.WeaponDef.TargetingModeOne,
			TUCost:   i.WeaponDef.TUCostOne,
			MaxRange: i.WeaponDef.MaxRangeOne,
			IsAimed:  false,
		})
		if !noAim && i.WeaponDef.TargetingModeOne.IsAimable() {
			modes = append(modes, AttackMode{
				Mode:     i.WeaponDef.TargetingModeOne,
				TUCost:   i.WeaponDef.TUCostOne + 2,
				MaxRange: i.WeaponDef.MaxRangeOne,
				IsAimed:  true,
			})
		}
	}
	if i.WeaponDef.TargetingModeTwo != special.TargetingModeNone {
		modes = append(modes, AttackMode{
			Mode:     i.WeaponDef.TargetingModeTwo,
			TUCost:   i.WeaponDef.TUCostTwo,
			MaxRange: i.WeaponDef.MaxRangeTwo,
			IsAimed:  false,
		})
		if !noAim && i.WeaponDef.TargetingModeTwo.IsAimable() {
			modes = append(modes, AttackMode{
				Mode:     i.WeaponDef.TargetingModeTwo,
				TUCost:   i.WeaponDef.TUCostTwo + 2,
				MaxRange: i.WeaponDef.MaxRangeTwo,
				IsAimed:  true,
			})
		}
	}
	return modes
}

func ItemDefsFromRecords(otherRecords []recfile.Record) []ItemDef {
	var items []ItemDef
	for _, record := range otherRecords {
		itemDef := NewItemDefFromRecord(record)
		items = append(items, itemDef)
	}
	return items
}

func NewItemDefFromRecord(record recfile.Record) ItemDef {
	itemDef := ItemDef{
		WeaponDef: WeaponDef{
			CaliberIndex: -1,
		},
	}
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "name":
			itemDef.Name = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			itemDef.Position = spawnPos
		case "description":
			itemDef.Description = field.Value
		case "category":
			itemDef.Category = foundation.ItemCategoryFromString(field.Value)
		case "size":
			itemDef.Size = field.AsInt()
		case "cost":
			itemDef.Cost = field.AsInt()
		case "weight":
			itemDef.Weight = field.AsInt()
		case "chance_to_break_on_throw":
			itemDef.ChanceToBreakOnThrow = field.AsInt()
		case "tags":
			itemDef.Tags |= foundation.ItemTagFromString(field.Value)
		case "ammo_dmg_multiplier":
			itemDef.AmmoDef.DamageMultiplier = field.AsInt()
		case "ammo_dmg_divisor":
			itemDef.AmmoDef.DamageDivisor = field.AsInt()
		case "ammo_ac_modifier":
			itemDef.AmmoDef.ACModifier = field.AsInt()
		case "ammo_dr_modifier":
			itemDef.AmmoDef.DRModifier = field.AsInt()
		case "ammo_rounds_in_magazine":
			itemDef.AmmoDef.RoundsInMagazine = field.AsInt()
		case "ammo_caliber_index":
			itemDef.AmmoDef.CaliberIndex = field.AsInt()
		case "weapon_type":
			itemDef.WeaponDef.Type = WeaponTypeFromString(field.Value)
		case "weapon_damage_type":
			itemDef.WeaponDef.DamageType = special.DamageTypeFromString(field.Value)
		case "weapon_caliber_index":
			itemDef.WeaponDef.CaliberIndex = field.AsInt()
		case "weapon_sound_id":
			itemDef.WeaponDef.SoundID = field.AsInt32()
		case "weapon_skill_used":
			itemDef.WeaponDef.SkillUsed = special.SkillFromName(field.Value)
		case "weapon_damage":
			itemDef.WeaponDef.Damage = fxtools.ParseInterval(field.Value)
		case "weapon_magazine_size":
			itemDef.WeaponDef.MagazineSize = field.AsInt()
		case "weapon_burst_rounds":
			itemDef.WeaponDef.BurstRounds = field.AsInt()
		case "weapon_attack_mode_one":
			itemDef.WeaponDef.TargetingModeOne = special.TargetingModeFromString(field.Value)
		case "weapon_attack_mode_two":
			itemDef.WeaponDef.TargetingModeTwo = special.TargetingModeFromString(field.Value)
		case "weapon_ap_cost_one":
			itemDef.WeaponDef.TUCostOne = field.AsInt() * 2
		case "weapon_ap_cost_two":
			itemDef.WeaponDef.TUCostTwo = field.AsInt() * 2
		case "weapon_max_range_one":
			itemDef.WeaponDef.MaxRangeOne = field.AsInt()
		case "weapon_max_range_two":
			itemDef.WeaponDef.MaxRangeTwo = field.AsInt()
		case "thrown_damage":
			itemDef.ThrowDamageDice = dice_curve.ParseDice(field.Value)
		case "shot_max_range":
			itemDef.WeaponDef.ShotMaxRange = field.AsInt()
		case "shot_min_range":
			itemDef.WeaponDef.ShotMinRange = field.AsInt()
		case "shot_half_damage_range":
			itemDef.WeaponDef.ShotHalfDamageRange = field.AsInt()
		case "shot_accuracy":
			itemDef.WeaponDef.ShotAccuracy = field.AsInt()
		case "use_effect":
			if useEffectExists(field.Value) {
				itemDef.UseEffect = field.Value
			} else {
				panic("Invalid use effect: " + field.Value)
			}
		case "zap_effect":
			if zapEffectExists(field.Value) {
				itemDef.ZapEffect = field.Value
			} else {
				panic("Invalid zap effect: " + field.Value)
			}
		case "charges":
			itemDef.Charges = dice_curve.ParseDice(field.Value)
		case "always_id_on_use":
			itemDef.AlwaysIDOnUse = field.AsBool()
		case "armor_encumbrance":
			itemDef.ArmorDef.Encumbrance = field.AsInt()
		case "armor_radiation_reduction":
			itemDef.ArmorDef.RadiationReduction = field.AsInt()
		case "armor_physical":
			if itemDef.ArmorDef.Protection == nil {
				itemDef.ArmorDef.Protection = make(map[special.DamageType]Protection)
			}
			values := field.AsList(",")
			itemDef.ArmorDef.Protection[special.DamageTypeNormal] = Protection{
				DamageThreshold: values[0].AsInt(),
				DamageReduction: values[1].AsInt(),
			}
		case "armor_energy":
			if itemDef.ArmorDef.Protection == nil {
				itemDef.ArmorDef.Protection = make(map[special.DamageType]Protection)
			}
			values := field.AsList(",")
			itemDef.ArmorDef.Protection[special.DamageTypeLaser] = Protection{
				DamageThreshold: values[0].AsInt(),
				DamageReduction: values[1].AsInt(),
			}
		case "stat":
			itemDef.Stat = dice_curve.StatFromString(field.Value)
		case "stat_bonus":
			itemDef.StatBonus = dice_curve.ParseDice(field.Value)
		case "skill":
			itemDef.Skill = dice_curve.SkillNameFromString(field.Value)
		case "skill_bonus":
			itemDef.SkillBonus = dice_curve.ParseDice(field.Value)
		case "equip_flag":
			itemDef.EquipFlag = foundation.ActorFlagFromString(field.Value)
		case "textfile":
			itemDef.TextFile = field.Value
		case "text":
			itemDef.Text = field.Value
		case "lockflag":
			itemDef.LockFlag = field.Value
		case "pickupflag":
			itemDef.SetFlagOnPickup = field.Value
		}
	}

	return itemDef
}
