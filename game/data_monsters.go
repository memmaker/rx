package game

import (
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"RogueUI/special"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

/*
strength: 18
dexterity: 10
intelligence: 9
health: 12
hp: 18
will: 9
perception: 10
fp: 12
speed: 6
size_modifier: 1
dodge: 8
dr: 0
attack: 12 | 1d2 | bite | claw
*/
type IntrinsicAttack struct {
	BaseSkill  int
	DamageDice dice_curve.Dice
	AttackName string
}
type ActorDef struct {
	Description      string
	Name             string
	Icon             textiles.TextIcon
	SpecialStats     map[special.Stat]int
	DerivedStat      map[special.DerivedStat]int
	SkillAdjustments map[special.Skill]int
	SizeModifier     int
	ZapEffects       []string
	UseEffects       []string
	Flags            *foundation.MapFlags

	Equipment       []string
	XP              int
	DefaultRelation PlayerRelation

	Position     geometry.Point
	DialogueFile string
}

func (d ActorDef) WithIcon(icon textiles.TextIcon) ActorDef {
	d.Icon = icon
	return d
}

func ActorDefsFromRecords(records []recfile.Record, palette textiles.ColorPalette) []ActorDef {
	var monsters []ActorDef
	for _, record := range records {
		monsterDef := NewActorDefFromRecord(record, palette)
		monsters = append(monsters, monsterDef)
	}
	return monsters
}

func NewActorDefFromRecord(record recfile.Record, palette textiles.ColorPalette) ActorDef {
	monsterDef := ActorDef{
		Flags:            foundation.NewMapFlags(),
		SpecialStats:     make(map[special.Stat]int),
		DerivedStat:      make(map[special.DerivedStat]int),
		SkillAdjustments: make(map[special.Skill]int),
	}
	monsterDef = fillDefinitionFromRecord(monsterDef, record, palette)
	return monsterDef
}

func fillDefinitionFromRecord(monsterDef ActorDef, record recfile.Record, palette textiles.ColorPalette) ActorDef {
	var icon textiles.TextIcon
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "name":
			monsterDef.Name = field.Value
		case "description":
			monsterDef.Description = field.Value
		case "icon":
			icon.Char = field.AsRune()
		case "foreground":
			icon.Fg = palette.Get(field.Value)
		case "expvalue":
			monsterDef.XP = field.AsInt()
		case "zap_effect":
			monsterDef.ZapEffects = append(monsterDef.ZapEffects, field.Value)
		case "use_effect":
			monsterDef.UseEffects = append(monsterDef.UseEffects, field.Value)
		case "strength":
			monsterDef.SpecialStats[special.Strength] = field.AsInt()
		case "perception":
			monsterDef.SpecialStats[special.Perception] = field.AsInt()
		case "endurance":
			monsterDef.SpecialStats[special.Endurance] = field.AsInt()
		case "charisma":
			monsterDef.SpecialStats[special.Charisma] = field.AsInt()
		case "intelligence":
			monsterDef.SpecialStats[special.Intelligence] = field.AsInt()
		case "agility":
			monsterDef.SpecialStats[special.Agility] = field.AsInt()
		case "luck":
			monsterDef.SpecialStats[special.Luck] = field.AsInt()
		case "hitpoints":
			monsterDef.DerivedStat[special.HitPoints] = field.AsInt()
		case "actionpoints":
			monsterDef.DerivedStat[special.ActionPoints] = field.AsInt()
		case "speed":
			monsterDef.DerivedStat[special.Speed] = field.AsInt()
		case "skillbonusunarmed":
			monsterDef.SkillAdjustments[special.Unarmed] = field.AsInt()
		case "skillbonusmeleeweapons":
			monsterDef.SkillAdjustments[special.MeleeWeapons] = field.AsInt()
		case "skillbonussmallguns":
			monsterDef.SkillAdjustments[special.SmallGuns] = field.AsInt()
		case "skillbonusbigguns":
			monsterDef.SkillAdjustments[special.BigGuns] = field.AsInt()
		case "skillbonusenergyweapons":
			monsterDef.SkillAdjustments[special.EnergyWeapons] = field.AsInt()
		case "skillbonusthrowing":
			monsterDef.SkillAdjustments[special.Throwing] = field.AsInt()
		case "size_modifier":
			monsterDef.SizeModifier = field.AsInt()
		case "dodge":
			monsterDef.DerivedStat[special.Dodge] = field.AsInt()
		case "equipment":
			monsterDef.Equipment = append(monsterDef.Equipment, field.Value)
		case "default_relation":
			monsterDef.DefaultRelation = PlayerRelationFromString(field.Value)
		case "position":
			pos, _ := geometry.NewPointFromEncodedString(field.Value)
			monsterDef.Position = pos
		case "dialogue":
			monsterDef.DialogueFile = field.Value
		case "flags":
			for _, mFlag := range field.AsList("|") {
				monsterDef.Flags.Set(foundation.ActorFlagFromString(mFlag.Value))
			}
		default:
			println("WARNING: Unknown field: " + field.Name)
		}
	}

	return monsterDef.WithIcon(icon)
}
