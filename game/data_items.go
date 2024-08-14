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

	Slot foundation.EquipSlot

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

	Skill      dice_curve.SkillName
	SkillBonus dice_curve.Dice
	Tags       foundation.ItemTags
	Position   geometry.Point
	TextFile   string
	LockFlag   string
}

func (i ItemDef) IsValidArmor() bool {
	return i.Slot.IsArmorSlot()
}

func (i ItemDef) IsValidWeapon() bool {
	return i.WeaponDef.IsValid()
}

func (i ItemDef) IsValidAmmo() bool {
	return i.AmmoDef.IsValid()
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
	itemDef := ItemDef{}
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
		case "tags":
			itemDef.Tags |= foundation.ItemTagFromString(field.Value)
		case "slot":
			itemDef.Slot = foundation.ItemSlotFromString(field.Value)
		case "ammo_damage":
			itemDef.AmmoDef.Damage = fxtools.ParseInterval(field.Value)
		case "ammo_type":
			itemDef.AmmoDef.Kind = field.Value
		case "weapon_type":
			itemDef.WeaponDef.Type = WeaponTypeFromString(field.Value)
		case "weapon_uses_ammo":
			itemDef.WeaponDef.UsesAmmo = field.Value
		case "weapon_skill_used":
			itemDef.WeaponDef.SkillUsed = special.SkillFromName(field.Value)
		case "weapon_damage":
			itemDef.WeaponDef.Damage = fxtools.ParseInterval(field.Value)
		case "weapon_magazine_size":
			itemDef.WeaponDef.MagazineSize = field.AsInt()
		case "weapon_burst_rounds":
			itemDef.WeaponDef.BurstRounds = field.AsInt()
		case "weapon_targeting_mode":
			itemDef.WeaponDef.TargetingMode |= TargetingModeFromString(field.Value)
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
			itemDef.ArmorDef.Protection[special.Physical] = Protection{
				damageThreshold: values[0].AsInt(),
				damageReduction: values[1].AsInt(),
			}
		case "armor_energy":
			if itemDef.ArmorDef.Protection == nil {
				itemDef.ArmorDef.Protection = make(map[special.DamageType]Protection)
			}
			values := field.AsList(",")
			itemDef.ArmorDef.Protection[special.Energy] = Protection{
				damageThreshold: values[0].AsInt(),
				damageReduction: values[1].AsInt(),
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
		case "lockflag":
			itemDef.LockFlag = field.Value
		}
	}

	return itemDef
}
