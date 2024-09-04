package game

import (
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

	flags := special.NewActorFlags()

	charSheet := special.NewCharSheet()

	var dodge, hitpoints, actionpoints, speed int
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
			hitpoints = field.AsInt()
		case "dodge":
			dodge = field.AsInt()
		case "actionpoints":
			actionpoints = field.AsInt()
		case "speed":
			speed = field.AsInt()
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
		case "equipment":
			equipment = append(equipment, field.Value)
		case "default_relation":
			actor.SetAIState(PlayerRelationFromString(field.Value))
		case "position":
			pos, _ := geometry.NewPointFromEncodedString(field.Value)
			actor.SetPosition(pos)
		case "dialogue":
			actor.SetDialogueFile(field.Value)
		case "flags":
			for _, mFlag := range field.AsList("|") {
				flags.Set(special.ActorFlagFromString(mFlag.Value))
			}
		default:
			//println("WARNING: Unknown field: " + field.Name)
		}
	}

	actor.GetFlags().Init(flags.UnderlyingCopy())

	charSheet.SetDerivedStatAbsoluteValue(special.HitPoints, hitpoints)
	charSheet.SetDerivedStatAbsoluteValue(special.ActionPoints, actionpoints)
	charSheet.SetDerivedStatAbsoluteValue(special.Speed, speed)
	charSheet.SetDerivedStatAbsoluteValue(special.Dodge, dodge)

	charSheet.HealCompletely()

	if actor.HasFlag(special.FlagZombie) {
		actor.SetHostile()
	}

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
