package game

import (
	"contractor/d100"
	"contractor/foundation"
	"contractor/gridmap"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"math/rand"
	"strings"
)

func NewItemFromRecord(record recfile.Record, itemFromString func(name string) foundation.Item, icon func(itemCategory foundation.ItemCategory) textiles.TextIcon) foundation.Item {
	NoQualityDefined := d100.Percentage(-1)
	item := &GenericItem{
		UID:              gridmap.NextItemID(),
		QualityInPercent: NoQualityDefined,
		Alive:            true,
		EffectParameters: make(foundation.Params),
		StatChanges:      StatChange{},
		Hidden:           false,
	}

	charges := 1

	itemAmmo := &Ammo{
		CaliberIndex:                    -1,
		BonusDamageAgainstActorWithTags: make(map[foundation.ActorFlag]int),
		DamageFactor:                    1,
		ConditionFactor:                 1,
		SpreadFactor:                    1,
	}
	itemWeapon := &Weapon{
		LoadedInMagazine: nil,
		PelletCount:      1,
	}
	var targetModes [2]TargetingMode
	var tuCosts [2]int
	var maxRanges [2]int
	var preLoadedAmmoName string

	itemArmor := &Armor{}
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		// GLOBAL FIELDS
		case "name":
			item.InternalName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			item.RawPosition = spawnPos
		case "description":
			item.DisplayName = field.Value
		case "longdescription":
			item.Text = field.Value
		case "category":
			item.Category = foundation.ItemCategoryFromString(field.Value)
			item.Icon = icon(item.Category)
		case "cost":
			item.Cost = field.AsInt()
		case "weight":
			item.Weight = field.AsInt()
		case "quality":
			item.QualityInPercent = d100.Percentage(field.AsInt())
		case "chance_to_break_on_throw":
			item.ChanceToBreakOnThrow = field.AsInt()
		case "hidden":
			item.Hidden = field.AsBool()
		case "tags":
			item.Tags |= foundation.ItemTagFromString(field.Value)
		case "thrown_damage":
			item.ThrownDamage = fxtools.ParseInterval(field.Value)
		case "use_effect":
			if useEffectExists(field.Value) {
				item.UseEffectName = field.Value
			} else {
				panic("Invalid use effect: " + field.Value)
			}
		case "zap_effect":
			if zapEffectExists(field.Value) {
				item.ZapEffectName = field.Value
			} else {
				panic("Invalid zap effect: " + field.Value)
			}
		case "effect_damage":
			item.EffectParameters["damage"] = field.AsInt()
		case "effect_damage_interval":
			item.EffectParameters["damage_interval"] = fxtools.ParseInterval(field.Value)
		case "effect_radius":
			item.EffectParameters["radius"] = field.AsInt()
		case "charges":
			charges = fxtools.ParseInterval(field.Value).Roll()
		case "stat_bonus":
			if fxtools.LooksLikeAFunction(field.Value) {
				name, args := fxtools.GetNameAndArgs(field.Value)
				stat := d100.StatFromString(name)
				bonus := args.GetInt(0)
				if item.StatChanges.StatChanges == nil {
					item.StatChanges.StatChanges = make(map[d100.Stat]int)
				}
				item.StatChanges.StatChanges[stat] = bonus
			}
		case "skill_bonus":
			if fxtools.LooksLikeAFunction(field.Value) {
				name, args := fxtools.GetNameAndArgs(field.Value)
				skill := d100.SkillFromString(name)
				bonus := args.GetInt(0)
				if item.StatChanges.SkillChanges == nil {
					item.StatChanges.SkillChanges = make(map[d100.Skill]int)
				}
				item.StatChanges.SkillChanges[skill] = bonus
			}
		case "derived_stat_bonus":
			if fxtools.LooksLikeAFunction(field.Value) {
				name, args := fxtools.GetNameAndArgs(field.Value)
				stat := d100.DerivedStatFromString(name)
				bonus := args.GetInt(0)
				if item.StatChanges.DerivedStatChanges == nil {
					item.StatChanges.DerivedStatChanges = make(map[d100.DerivedStat]int)
				}
				item.StatChanges.DerivedStatChanges[stat] = bonus
			}
		case "equip_flag":
			item.EquipFlag = foundation.ActorFlagFromString(field.Value)
		case "textfile":
			item.TextFile = field.Value
		case "text":
			item.Text = field.Value
		case "textvar":
			item.TextVar = field.Value
		case "textvalue":
			item.TextValue = field.Value
		case "lockflag":
			item.LockFlag = field.Value
		case "pickupflag":
			item.SetFlagOnPickup = field.Value
		case "dropflag":
			item.SetFlagOnDrop = field.Value

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
				itemAmmo.BonusDamageAgainstActorWithTags[foundation.ActorFlagFromString(name)] = args.GetInt(0)
			}
		case "ammo_rounds_in_magazine":
			itemAmmo.RoundsInMagazine = field.AsInt()
		case "ammo_caliber_index":
			itemAmmo.CaliberIndex = field.AsInt()
		case "ammo_short":
			itemAmmo.ShortIdentifier = field.Value

		// WEAPON FIELDS
		case "weapon_type":
			itemWeapon.WeaponType = WeaponTypeFromString(field.Value)
		case "weapon_damage_type":
			itemWeapon.DamageType = DamageTypeFromString(field.Value)
		case "weapon_caliber_index":
			itemWeapon.CaliberIndex = field.AsInt()
		case "weapon_uses_ammo":
			itemWeapon.CaliberName = field.Value
		case "weapon_sound_id":
			itemWeapon.SoundID = field.AsInt32()
		case "weapon_skill_used":
			itemWeapon.SkillUsed = d100.SkillFromString(field.Value)
		case "weapon_damage":
			itemWeapon.DamageDice = fxtools.ParseInterval(field.Value)
		case "weapon_magazine_size":
			itemWeapon.MagazineSize = field.AsInt()
		case "weapon_burst_rounds":
			itemWeapon.BurstRounds = field.AsInt()
		case "weapon_pellet_count":
			itemWeapon.PelletCount = field.AsInt()
		case "weapon_attack_mode_one":
			targetModes[0] = TargetingModeFromString(field.Value)
		case "weapon_attack_mode_two":
			targetModes[1] = TargetingModeFromString(field.Value)
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
		case "weapon_reliability":
			itemWeapon.Reliability = d100.Percentage(field.AsInt())
		case "weapon_concealability":
			itemWeapon.RelativeSize = WeaponSizeFromString(field.Value)
		case "weapon_accuracy":
			itemWeapon.AccuracyMod = d100.Percentage(field.AsInt())
		case "weapon_degrade_factor":
			itemWeapon.DegradeFactor = field.AsFloat()
		case "weapon_always_load":
			preLoadedAmmoName = field.Value
		// ARMOR FIELDS
		case "armor_encumbrance":
			itemArmor.Encumbrance = field.AsInt()
		case "armor_style":
			itemArmor.FashionStyle = foundation.FashionStyleFromString(field.Value)
		case "armor_conceal_slot":
			itemArmor.ConcealSlots = append(itemArmor.ConcealSlots, WeaponSizeFromString(field.Value))
		case "armor_radiation_reduction":
			itemArmor.RadiationReduction = field.AsInt()
		case "armor_physical_reduction":
			protectionValue := field.AsInt()
			if protectionValue == 0 {
				continue
			}
			if itemArmor.Protection == nil {
				itemArmor.Protection = make(map[DamageType]Protection)
			}
			itemArmor.Protection[DamageTypeNormal] = itemArmor.Protection[DamageTypeNormal].WithReduction(protectionValue)
		case "armor_physical_threshold":
			protectionValue := field.AsInt()
			if itemArmor.Protection == nil {
				itemArmor.Protection = make(map[DamageType]Protection)
			}
			if protectionValue == 0 {
				continue
			}
			itemArmor.Protection[DamageTypeNormal] = itemArmor.Protection[DamageTypeNormal].WithThreshold(protectionValue)
		case "armor_energy_reduction":
			protectionValue := field.AsInt()
			if itemArmor.Protection == nil {
				itemArmor.Protection = make(map[DamageType]Protection)
			}
			if protectionValue == 0 {
				continue
			}
			itemArmor.Protection[DamageTypeEnergy] = itemArmor.Protection[DamageTypeEnergy].WithReduction(protectionValue)
		case "armor_energy_threshold":
			protectionValue := field.AsInt()
			if itemArmor.Protection == nil {
				itemArmor.Protection = make(map[DamageType]Protection)
			}
			if protectionValue == 0 {
				continue
			}
			itemArmor.Protection[DamageTypeEnergy] = itemArmor.Protection[DamageTypeEnergy].WithThreshold(protectionValue)
		}
	}

	item.Charges = charges

	if item.QualityInPercent == NoQualityDefined {
		item.QualityInPercent = max(10, d100.Percentage(rand.Intn(100)+1))
	}

	if itemAmmo.IsValid() {
		itemAmmo.GenericItem = item
		itemAmmo.GenericItem.StackSize = itemAmmo.RoundsInMagazine
		return itemAmmo
	}

	if itemWeapon.IsValid() {
		noAim := item.Tags.Contains(foundation.TagNoAim)
		itemWeapon.AttackModes = GetAttackModes(targetModes, tuCosts, maxRanges, noAim)
		itemWeapon.GenericItem = item

		noReload := item.Tags.Contains(foundation.TagNoReload)
		if noReload && preLoadedAmmoName != "" {
			preLoadAmmo := itemFromString(preLoadedAmmoName)
			preLoadAmmo.SetStackSize(itemWeapon.MagazineSize)
			itemWeapon.LoadAmmo(preLoadAmmo.(*Ammo))
		}
		return itemWeapon
	}

	if item.GetCategory().IsArmor() {
		itemArmor.GenericItem = item
		return itemArmor
	}

	return item
}
