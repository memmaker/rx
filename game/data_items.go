package game

import (
	"RogueUI/foundation"
	"RogueUI/recfile"
	"RogueUI/rpg"
	"strings"
)

type WeaponType int

func (t WeaponType) IsMissile() bool {
	return t == ItemTypeArrow || t == ItemTypeBolt || t == ItemTypeDart
}

const (
	ItemTypeUnknown WeaponType = iota
	ItemTypeSword
	ItemTypeClub
	ItemTypeAxe
	ItemTypeDagger
	ItemTypeSpear
	ItemTypeBow
	ItemTypeArrow
	ItemTypeCrossbow
	ItemTypeBolt
	ItemTypeDart
)

type WeaponDef struct {
	DamageDice          rpg.Dice
	Type                WeaponType
	LaunchedWithType    WeaponType
	SkillUsed           rpg.SkillName
	ShotMaxRange        int
	ShotMinRange        int
	ShotHalfDamageRange int
	ShotAccuracy        int
}

func (w WeaponDef) IsValid() bool {
	return w.Type != ItemTypeUnknown && w.DamageDice.NotZero()
}

type ArmorDef struct {
	DamageResistance int
	Encumbrance      rpg.Encumbrance
}

type ItemDef struct {
	Name         string
	InternalName string

	Slot foundation.EquipSlot

	WeaponDef WeaponDef
	ArmorDef  ArmorDef

	ThrowDamageDice rpg.Dice

	UseEffect string
	ZapEffect string

	Stat      rpg.Stat
	StatBonus rpg.Dice

	Charges  rpg.Dice
	Category foundation.ItemCategory

	AlwaysIDOnUse bool
	EquipFlag     foundation.ActorFlag

	Skill      rpg.SkillName
	SkillBonus rpg.Dice
}

func (i ItemDef) IsValidArmor() bool {
	return i.Slot.IsArmorSlot()
}

func (i ItemDef) IsValidWeapon() bool {
	return i.Slot.IsWeaponSlot() && i.WeaponDef.IsValid()
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
		switch field.Name {
		case "name":
			itemDef.Name = field.Value
		case "internal_name":
			itemDef.InternalName = field.Value
		case "category":
			itemDef.Category = foundation.ItemCategoryFromString(field.Value)
		case "slot":
			itemDef.Slot = foundation.ItemSlotFromString(field.Value)
		case "weapon_type":
			itemDef.WeaponDef.Type = WeaponTypeFromString(field.Value)
		case "weapon_launched_with_type":
			itemDef.WeaponDef.LaunchedWithType = WeaponTypeFromString(field.Value)
		case "weapon_skill_used":
			itemDef.WeaponDef.SkillUsed = rpg.SkillNameFromString(field.Value)
		case "weapon_damage":
			itemDef.WeaponDef.DamageDice = rpg.ParseDice(field.Value)
		case "damage_resistance":
			itemDef.ArmorDef.DamageResistance = field.AsInt()
		case "thrown_damage":
			itemDef.ThrowDamageDice = rpg.ParseDice(field.Value)
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
			itemDef.Charges = rpg.ParseDice(field.Value)
		case "always_id_on_use":
			itemDef.AlwaysIDOnUse = field.AsBool()
		case "encumbrance":
			itemDef.ArmorDef.Encumbrance = rpg.EncumbranceFromString(field.Value)
		case "stat":
			itemDef.Stat = rpg.StatFromString(field.Value)
		case "stat_bonus":
			itemDef.StatBonus = rpg.ParseDice(field.Value)
		case "skill":
			itemDef.Skill = rpg.SkillNameFromString(field.Value)
		case "skill_bonus":
			itemDef.SkillBonus = rpg.ParseDice(field.Value)
		case "equip_flag":
			itemDef.EquipFlag = foundation.ActorFlagFromString(field.Value)
		}
	}

	return itemDef
}

func WeaponTypeFromString(value string) WeaponType {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "sword":
		return ItemTypeSword
	case "club":
		return ItemTypeClub
	case "axe":
		return ItemTypeAxe
	case "dagger":
		return ItemTypeDagger
	case "spear":
		return ItemTypeSpear
	case "bow":
		return ItemTypeBow
	case "arrow":
		return ItemTypeArrow
	case "crossbow":
		return ItemTypeCrossbow
	case "bolt":
		return ItemTypeBolt
	case "dart":
		return ItemTypeDart
	}
	panic("Invalid weapon type: " + value)
	return ItemTypeUnknown
}
