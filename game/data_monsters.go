package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
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

	charSheet := special.NewCharSheet()

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
				skill := special.SkillFromBonusString(field.Value)
				if skill != -1 {
					charSheet.SetSkillAdjustment(skill, field.AsInt())
				}
			}
		}
	}

	actor.GetFlags().Init(flags.UnderlyingCopy())

	if hitpoints != -1 {
		charSheet.SetDerivedStatAbsoluteValue(special.HitPoints, hitpoints)
	}
	if actionpoints != -1 {
		charSheet.SetDerivedStatAbsoluteValue(special.ActionPoints, actionpoints)
	}
	if speed != -1 {
		charSheet.SetDerivedStatAbsoluteValue(special.Speed, speed)
	}
	if dodge != -1 {
		charSheet.SetDerivedStatAbsoluteValue(special.Dodge, dodge)
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
