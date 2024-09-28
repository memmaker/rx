package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
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

	flags := foundation.NewActorFlags()

	charSheet := d100.NewCharSheet()

	dodge := -1
	hitpoints := -1
	actionpoints := -1
	speed := -1

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
			charSheet.SetStat(d100.Strength, field.AsInt())
		case "perception":
			charSheet.SetStat(d100.Perception, field.AsInt())
		case "endurance":
			charSheet.SetStat(d100.Endurance, field.AsInt())
		case "charisma":
			charSheet.SetStat(d100.Charisma, field.AsInt())
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
		case "default_relation":
			actor.SetAIState(foundation.AIStateFromString(field.Value))
		case "position":
			pos, _ := geometry.NewPointFromEncodedString(field.Value)
			actor.SetPosition(pos)
		case "audio":
			actor.audioBaseName = field.Value
		case "dialogue":
			actor.SetDialogueFile(field.Value)
		case "chatter":
			actor.SetChatterFile(field.Value)
		case "flags":
			for _, mFlag := range field.AsList("|") {
				flags.Set(foundation.ActorFlagFromString(mFlag.Value))
			}
		default:
			//println("WARNING: Unknown field: " + field.Name)
			if strings.HasPrefix(field.Name, "skillbonus") {
				skill := d100.SkillFromString(strings.TrimPrefix(field.Name, "skillbonus"))
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
		actor.SetHostile()
	}

	actor.SetCharSheet(charSheet)
	actor.SetIcon(icon)
	actor.SetIntrinsicZapEffects(zapEffects)
	actor.SetIntrinsicUseEffects(useEffects)

	for _, itemName := range equipment {
		item := newItemFromString(itemName)
		if item != nil {
			actor.GetInventory().AddItem(item)
		}
	}
	return actor
}
