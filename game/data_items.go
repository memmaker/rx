package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"math/rand"
	"strings"
)

func NewItemFromRecord(record recfile.Record, icon func(itemCategory foundation.ItemCategory) textiles.TextIcon) *Item {
	NoQualityDefined := special.Percentage(-1)
	item := &Item{
		qualityInPercent: NoQualityDefined,
		alive:            true,
		effectParameters: make(Params),
	}

	charges := 1

	itemAmmo := &AmmoInfo{
		CaliberIndex:                    -1,
		BonusDamageAgainstActorWithTags: make(map[special.ActorFlag]int),
		DamageFactor:                    1,
		ConditionFactor:                 1,
		SpreadFactor:                    1,
	}
	itemWeapon := &WeaponInfo{
		loadedInMagazine: nil,
	}
	var targetModes [2]special.TargetingMode
	var tuCosts [2]int
	var maxRanges [2]int

	itemArmor := &ArmorInfo{}
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		// GLOBAL FIELDS
		case "name":
			item.internalName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			item.position = spawnPos
		case "description":
			item.description = field.Value
		case "category":
			item.category = foundation.ItemCategoryFromString(field.Value)
			item.icon = icon(item.category)
		case "size":
			item.size = field.AsInt()
		case "cost":
			item.cost = field.AsInt()
		case "weight":
			item.weight = field.AsInt()
		case "quality":
			item.qualityInPercent = special.Percentage(field.AsInt())
		case "chance_to_break_on_throw":
			item.chanceToBreakOnThrow = field.AsInt()
		case "tags":
			item.tags |= foundation.ItemTagFromString(field.Value)
		case "thrown_damage":
			item.thrownDamage = fxtools.ParseInterval(field.Value)
		case "use_effect":
			if useEffectExists(field.Value) {
				item.useEffectName = field.Value
			} else {
				panic("Invalid use effect: " + field.Value)
			}
		case "zap_effect":
			if zapEffectExists(field.Value) {
				item.zapEffectName = field.Value
			} else {
				panic("Invalid zap effect: " + field.Value)
			}
		case "effect_damage":
			item.effectParameters["damage"] = field.AsInt()
		case "effect_damage_interval":
			item.effectParameters["damage_interval"] = fxtools.ParseInterval(field.Value)
		case "effect_radius":
			item.effectParameters["radius"] = field.AsInt()
		case "charges":
			charges = fxtools.ParseInterval(field.Value).Roll()
		case "stat":
			item.stat = special.StatFromString(field.Value)
		case "stat_bonus":
			item.statBonus = fxtools.ParseInterval(field.Value).Roll()
		case "skill":
			item.skill = special.SkillFromString(field.Value)
		case "skill_bonus":
			item.skillBonus = fxtools.ParseInterval(field.Value).Roll()
		case "equip_flag":
			item.equipFlag = special.ActorFlagFromString(field.Value)
		case "textfile":
			item.textFile = field.Value
		case "text":
			item.text = field.Value
		case "lockflag":
			item.lockFlag = field.Value
		case "pickupflag":
			item.setFlagOnPickup = field.Value
		case "dropflag":
			item.setFlagOnDrop = field.Value

		// AMMO FIELDS
		case "ammo_dmg_factor":
			itemAmmo.DamageFactor = field.AsFloat()
		case "ammo_condition_factor":
			itemAmmo.ConditionFactor = field.AsFloat()
		case "ammo_spread_factor":
			itemAmmo.SpreadFactor = field.AsFloat()
		case "ammo_dt_modifier":
			itemAmmo.DTModifier = field.AsInt()
		case "ammo_bonus_radius":
			itemAmmo.BonusRadius = field.AsInt()
		case "ammo_bonus_dmg_against":
			if fxtools.LooksLikeAFunction(field.Value) {
				name, args := fxtools.GetNameAndArgs(field.Value)
				itemAmmo.BonusDamageAgainstActorWithTags[special.ActorFlagFromString(name)] = args.GetInt(0)
			}
		case "ammo_rounds_in_magazine":
			itemAmmo.RoundsInMagazine = field.AsInt()
		case "ammo_caliber_index":
			itemAmmo.CaliberIndex = field.AsInt()

		// WEAPON FIELDS
		case "weapon_type":
			itemWeapon.weaponType = WeaponTypeFromString(field.Value)
		case "weapon_damage_type":
			itemWeapon.damageType = special.DamageTypeFromString(field.Value)
		case "weapon_caliber_index":
			itemWeapon.caliberIndex = field.AsInt()
		case "weapon_sound_id":
			itemWeapon.soundID = field.AsInt32()
		case "weapon_skill_used":
			itemWeapon.skillUsed = special.SkillFromString(field.Value)
		case "weapon_damage":
			itemWeapon.damageDice = fxtools.ParseInterval(field.Value)
		case "weapon_magazine_size":
			itemWeapon.magazineSize = field.AsInt()
		case "weapon_burst_rounds":
			itemWeapon.burstRounds = field.AsInt()
		case "weapon_attack_mode_one":
			targetModes[0] = special.TargetingModeFromString(field.Value)
		case "weapon_attack_mode_two":
			targetModes[1] = special.TargetingModeFromString(field.Value)
		case "weapon_ap_cost_one":
			tuCosts[0] = field.AsInt() * 2
		case "weapon_ap_cost_two":
			tuCosts[1] = field.AsInt() * 2
		case "weapon_max_range_one":
			maxRanges[0] = field.AsInt()
		case "weapon_max_range_two":
			maxRanges[1] = field.AsInt()
		case "weapon_min_str":
			itemWeapon.MinSTR = field.AsInt()

		// ARMOR FIELDS
		case "armor_encumbrance":
			itemArmor.encumbrance = field.AsInt()
		case "armor_radiation_reduction":
			itemArmor.radiationReduction = field.AsInt()
		case "armor_physical":
			if itemArmor.protection == nil {
				itemArmor.protection = make(map[special.DamageType]Protection)
			}
			values := field.AsList(",")
			itemArmor.protection[special.DamageTypeNormal] = Protection{
				DamageThreshold: values[0].AsInt(),
				DamageReduction: values[1].AsInt(),
			}
		case "armor_energy":
			if itemArmor.protection == nil {
				itemArmor.protection = make(map[special.DamageType]Protection)
			}
			values := field.AsList(",")
			itemArmor.protection[special.DamageTypeLaser] = Protection{
				DamageThreshold: values[0].AsInt(),
				DamageReduction: values[1].AsInt(),
			}
		}
	}

	item.charges = charges

	if itemAmmo.IsValid() {
		item.ammo = itemAmmo
		item.charges = item.ammo.RoundsInMagazine
	}

	if itemWeapon.IsValid() {
		noAim := item.tags.Contains(foundation.TagNoAim)
		itemWeapon.attackModes = GetAttackModes(targetModes, tuCosts, maxRanges, noAim)
		item.weapon = itemWeapon
	}

	if itemArmor.IsValid() {
		item.armor = itemArmor
	}

	if item.qualityInPercent == NoQualityDefined && (item.IsWeapon() || item.IsArmor()) {
		item.qualityInPercent = max(10, special.Percentage(rand.Intn(100)+1))
	}

	return item
}
