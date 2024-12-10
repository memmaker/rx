package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

func NewActorFromRecord(record recfile.Record, palette textiles.ColorPalette, newItemFromString func(string) foundation.Item) *Actor {
	actor := NewActor()
	var icon textiles.TextIcon
	var zapEffects []string
	var useEffects []string
	var equipment []string
	var vendorInv []string
	var cyberWareInstalls []string
	fashionVendorStyle := foundation.FashionStyle(-2)

	flags := foundation.NewActorFlags()

	charSheet := d100.NewCharSheet()

	dodge := -1
	hitpoints := -1
	actionpoints := -1
	speed := -1

	for _, field := range record {
		lowerName := strings.ToLower(field.Name)
		switch lowerName {
		case "name":
			actor.SetInternalName(field.Value)
		case "description":
			actor.SetDisplayName(field.Value)
		case "icon":
			icon.Char = field.AsRune()
		case "foreground":
			icon.Fg = palette.Get(field.Value)
		case "xp":
			actor.SetXP(field.AsInt())
		case "zap_effect":
			zapEffects = append(zapEffects, field.Value)
		case "use_effect":
			useEffects = append(useEffects, field.Value)
		case "strength":
			charSheet.SetStat(d100.Strength, field.AsInt())
		case "perception":
			charSheet.SetStat(d100.Perception, field.AsInt())
		case "endurance":
			charSheet.SetStat(d100.Endurance, field.AsInt())
		case "cool":
			charSheet.SetStat(d100.Cool, field.AsInt())
		case "intelligence":
			charSheet.SetStat(d100.Intelligence, field.AsInt())
		case "agility":
			charSheet.SetStat(d100.Agility, field.AsInt())
		case "hitpoints":
			hitpoints = field.AsInt()
		case "dodge":
			dodge = field.AsInt()
		case "actionpoints":
			actionpoints = field.AsInt()
		case "speed":
			speed = field.AsInt()
		case "size_modifier":
			actor.SetSizeModifier(field.AsInt())
		case "equipment":
			equipment = append(equipment, field.Value)
		case "cyberwareinstall":
			cyberWareInstalls = append(cyberWareInstalls, field.Value)
		case "selling":
			vendorInv = append(vendorInv, field.Value)
		case "fashion_vendor_style":
			fashionVendorStyle = foundation.FashionStyleFromString(field.Value)
		case "aggressive":
			actor.Aggressive = field.AsBool()
		case "guarding_zone":
			actor.GuardingZone = field.Value
		case "position":
			pos, _ := geometry.NewPointFromEncodedString(field.Value)
			actor.SetPosition(pos)
		case "audio":
			actor.AudioBaseName = field.Value
		case "dialogue":
			actor.SetDialogueFile(field.Value)
		case "chatter":
			actor.SetChatterFile(field.Value)
		case "faction":
			actor.TeamName = field.Value
		case "flags":
			for _, mFlag := range field.AsList("|") {
				flags.Set(foundation.ActorFlagFromString(mFlag.Value))
			}
		default:
			//println("WARNING: Unknown field: " + field.Name)
			if strings.HasPrefix(lowerName, "skillbonus") {
				skill := d100.SkillFromString(strings.TrimPrefix(lowerName, "skillbonus"))
				if skill != -1 {
					charSheet.SetSkillAdjustment(skill, field.AsInt())
				}
			}
		}
	}

	actor.GetFlags().Init(flags.UnderlyingCopy())

	if hitpoints != -1 {
		charSheet.SetDerivedStatAbsoluteValue(d100.HitPoints, hitpoints)
	}
	if actionpoints != -1 {
		charSheet.SetDerivedStatAbsoluteValue(d100.ActionPoints, actionpoints)
	}
	if speed != -1 {
		charSheet.SetDerivedStatAbsoluteValue(d100.Speed, speed)
	}
	if dodge != -1 {
		charSheet.SetDerivedStatAbsoluteValue(d100.Dodge, dodge)
	}

	charSheet.HealAPAndHPCompletely()

	if actor.HasFlag(foundation.FlagZombie) {
		actor.Aggressive = true
	}

	actor.SetCharSheet(charSheet)
	actor.SetIcon(icon)
	actor.SetIntrinsicZapEffects(zapEffects)
	actor.SetIntrinsicUseEffects(useEffects)

	for _, itemName := range equipment {
		item := newItemFromString(itemName)
		if item != nil {
			actor.GetInventory().AddItem(item)
			if item.IsArmor() {
				actor.GetInventory().Equip(item)
			}
		}
	}

	if len(vendorInv) > 0 {
		actor.VendorInv = NewInventory(40)
		for _, itemName := range vendorInv {
			item := newItemFromString(itemName)
			if item != nil {
				actor.GetVendorInventory().AddItem(item)
			}
		}
	}

	if len(cyberWareInstalls) > 0 {
		offers := make([]fxtools.Tuple[CyberWare, int], len(cyberWareInstalls))
		for i, cyberWareName := range cyberWareInstalls {
			var cyberWare CyberWare
			var price int
			if fxtools.LooksLikeAFunction(cyberWareName) {
				name, args := fxtools.GetNameAndArgs(cyberWareName)
				cyberWare = NewCyberWareFromString(name)
				price = args.GetInt(0)
			} else {
				cyberWare = NewCyberWareFromString(cyberWareName)
				price = cyberWare.DefaultPrice()
			}
			offers[i] = fxtools.Tuple[CyberWare, int]{Item1: cyberWare, Item2: price}
		}
		actor.OffersCyberWare = offers
	}

	if fashionVendorStyle != -2 {
		actor.VendorInv = NewInventory(40)
		actor.GetVendorInventory().AddItems(newFashionInventory(fashionVendorStyle))
	}

	actor.secondaryInit()

	if actor.HasFlag(foundation.FlagSpawnDead) {
		actor.Kill()
	}

	return actor
}
