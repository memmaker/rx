package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

func NewActorFromRecord(record recfile.Record, palette textiles.ColorPalette, newItemFromString func(string) *Item) *Actor {
	actor := NewActor()

	var icon textiles.TextIcon
	var zapEffects []string
	var useEffects []string
	var equipment []string

	flags := foundation.NewActorFlags()

	charSheet := special.NewCharSheet()

	for _, field := range record {
		switch strings.ToLower(field.Name) {
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
			charSheet.SetStat(special.Strength, field.AsInt())
		case "perception":
			charSheet.SetStat(special.Perception, field.AsInt())
		case "endurance":
			charSheet.SetStat(special.Endurance, field.AsInt())
		case "charisma":
			charSheet.SetStat(special.Charisma, field.AsInt())
		case "intelligence":
			charSheet.SetStat(special.Intelligence, field.AsInt())
		case "agility":
			charSheet.SetStat(special.Agility, field.AsInt())
		case "luck":
			charSheet.SetStat(special.Luck, field.AsInt())
		case "hitpoints":
			charSheet.SetDerivedStatAbsoluteValue(special.HitPoints, field.AsInt())
		case "actionpoints":
			charSheet.SetDerivedStatAbsoluteValue(special.ActionPoints, field.AsInt())
		case "speed":
			charSheet.SetDerivedStatAbsoluteValue(special.Speed, field.AsInt())
		case "skillbonusunarmed":
			charSheet.SetSkillAdjustment(special.Unarmed, field.AsInt())
		case "skillbonusmeleeweapons":
			charSheet.SetSkillAdjustment(special.MeleeWeapons, field.AsInt())
		case "skillbonussmallguns":
			charSheet.SetSkillAdjustment(special.SmallGuns, field.AsInt())
		case "skillbonusbigguns":
			charSheet.SetSkillAdjustment(special.BigGuns, field.AsInt())
		case "skillbonusenergyweapons":
			charSheet.SetSkillAdjustment(special.EnergyWeapons, field.AsInt())
		case "skillbonusthrowing":
			charSheet.SetSkillAdjustment(special.Throwing, field.AsInt())
		case "size_modifier":
			actor.SetSizeModifier(field.AsInt())
		case "dodge":
			charSheet.SetDerivedStatAbsoluteValue(special.Dodge, field.AsInt())
		case "equipment":
			equipment = append(equipment, field.Value)
		case "default_relation":
			actor.SetRelationToPlayer(PlayerRelationFromString(field.Value))
		case "position":
			pos, _ := geometry.NewPointFromEncodedString(field.Value)
			actor.SetPosition(pos)
		case "dialogue":
			actor.SetDialogueFile(field.Value)
		case "flags":
			for _, mFlag := range field.AsList("|") {
				flags.Set(foundation.ActorFlagFromString(mFlag.Value))
			}
		default:
			//println("WARNING: Unknown field: " + field.Name)
		}
	}

	actor.GetFlags().Init(flags.UnderlyingCopy())

	charSheet.HealCompletely()

	actor.SetCharSheet(charSheet)
	actor.SetIcon(icon)
	actor.SetIntrinsicZapEffects(zapEffects)
	actor.SetIntrinsicUseEffects(useEffects)

	for _, itemName := range equipment {
		item := newItemFromString(itemName)
		if item != nil {
			actor.GetInventory().Add(item)
		}
	}
	return actor
}
