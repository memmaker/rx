package game

import (
	"RogueUI/foundation"
	"RogueUI/recfile"
	"RogueUI/rpg"
	"regexp"
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
	EffectiveSkill int
	DamageDice     rpg.Dice
	AttackName     string
}
type MonsterDef struct {
	Name          string
	InternalName  string
	Icon          rune
	Color         string
	Strength      int
	Dexterity     int
	Intelligence  int
	Health        int
	Will          int
	Perception    int
	FatiguePoints int

	HitPoints        int
	Dodge            int
	SizeModifier     int
	BasicSpeed       int
	Attacks          []IntrinsicAttack
	DamageResistance int
	ZapEffects       []string
	UseEffects       []string
	DungeonLevel     int
	Flags            foundation.Flags
}

func MonsterDefsFromRecords(records []recfile.Record) []MonsterDef {
	var monsters []MonsterDef
	for _, record := range records {
		monsterDef := NewMonsterDefFromRecord(record)
		monsters = append(monsters, monsterDef)
	}
	return monsters
}

func NewMonsterDefFromRecord(record recfile.Record) MonsterDef {
	monsterDef := MonsterDef{}
	// an attack key looks like this "attack_1", "attack_2", etc..
	attackRegex, _ := regexp.Compile("attack_([0-9]+)")
	for _, field := range record {
		if attackRegex.MatchString(field.Name) {
			fields := field.AsList("|")
			monsterDef.Attacks = append(monsterDef.Attacks, IntrinsicAttack{
				EffectiveSkill: fields[0].AsInt(),
				DamageDice:     rpg.ParseDice(fields[1].Value),
				AttackName:     fields[2].Value,
			})
			continue
		}

		switch field.Name {
		case "name":
			monsterDef.Name = field.Value
		case "internal_name":
			monsterDef.InternalName = field.Value
		case "letter":
			monsterDef.Icon = []rune(field.Value)[0]
		case "color":
			monsterDef.Color = field.Value
		case "hp":
			monsterDef.HitPoints = field.AsInt()
		case "attack":
			fields := field.AsList("|")
			monsterDef.Attacks = append(monsterDef.Attacks, IntrinsicAttack{
				EffectiveSkill: fields[0].AsInt(),
				DamageDice:     rpg.ParseDice(fields[1].Value),
				AttackName:     fields[2].Value,
			})
		case "dr":
			monsterDef.DamageResistance = field.AsInt()
		case "zap_effect":
			monsterDef.ZapEffects = append(monsterDef.ZapEffects, field.Value)
		case "use_effect":
			monsterDef.UseEffects = append(monsterDef.UseEffects, field.Value)
		case "strength":
			monsterDef.Strength = field.AsInt()
		case "dexterity":
			monsterDef.Dexterity = field.AsInt()
		case "intelligence":
			monsterDef.Intelligence = field.AsInt()
		case "health":
			monsterDef.Health = field.AsInt()
		case "will":
			monsterDef.Will = field.AsInt()
		case "perception":
			monsterDef.Perception = field.AsInt()
		case "fp":
			monsterDef.FatiguePoints = field.AsInt()
		case "speed":
			monsterDef.BasicSpeed = field.AsInt()
		case "size_modifier":
			monsterDef.SizeModifier = field.AsInt()
		case "dodge":
			monsterDef.Dodge = field.AsInt()
		case "dlvl":
			monsterDef.DungeonLevel = field.AsInt()
		case "flags":
			var flagValue uint32
			for _, flag := range field.AsList("|") {
				flagValue |= foundation.ActorFlagFromString(flag.Value)
			}
			monsterDef.Flags = foundation.NewFlagsFromValue(flagValue)
		default:
			println("WARNING: Unknown field: " + field.Name)
		}
	}
	return monsterDef
}
