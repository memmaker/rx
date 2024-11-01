package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"RogueUI/fsmai"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"path"
	"strings"
	"time"
)

type ScriptFrame interface {
	IsEmpty() bool
	Condition(map[string]interface{}) bool
	ExecuteActions(map[string]interface{})
	String() string
}

type UserScriptFrame struct {
	// The condition for the frame.
	condition *govaluate.EvaluableExpression
	// The actions for the frame.
	actions []*govaluate.EvaluableExpression
}

func (f UserScriptFrame) String() string {
	return f.condition.String()
}

func (f UserScriptFrame) Condition(vars map[string]interface{}) bool {
	if f.condition == nil {
		return true
	}
	condition, err := f.condition.Evaluate(vars)
	if err != nil {
		panic(err)
	}
	return condition.(bool)
}

func (f UserScriptFrame) ExecuteActions(vars map[string]interface{}) {
	for _, action := range f.actions {
		_, err := action.Evaluate(vars)
		if err != nil {
			panic(err)
		}
	}
}

func (f UserScriptFrame) IsEmpty() bool {
	return f.condition == nil && len(f.actions) == 0
}

type ActionScript struct {
	Name string

	Variables map[string]interface{}

	Frames []ScriptFrame

	Outcomes []ScriptFrame

	CancelFrame ScriptFrame
}

func (s ActionScript) CanRunFrame(frame ScriptFrame) bool {
	return frame.Condition(s.Variables)
}

func (s ActionScript) IsEmpty() bool {
	return len(s.Frames) == 0
}

func mergeMaps(maps ...map[string]govaluate.ExpressionFunction) map[string]govaluate.ExpressionFunction {
	result := make(map[string]govaluate.ExpressionFunction)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func (g *GameState) getScriptFuncs() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		// Player Only
		"IsWounded": func(args ...interface{}) (interface{}, error) {
			return g.Player.IsWounded(), nil
		},
		"Skill": func(args ...interface{}) (interface{}, error) {
			skillName := args[0].(string)
			skillValue := g.Player.GetCharSheet().GetSkill(d100.SkillFromString(skillName))
			return (float64)(skillValue), nil
		},
		"RollSkill": func(args ...interface{}) (interface{}, error) {
			skillName := args[0].(string)
			diff := d100.Medium
			if len(args) > 1 {
				diff = d100.DifficultyFromString(args[1].(string))
			}
			result := g.Player.GetCharSheet().SkillRollVsDiff(d100.SkillFromString(skillName), diff)
			return (bool)(result.Success), nil
		},

		// Player Inventory & Equipment
		"RemoveItem": func(args ...interface{}) (interface{}, error) {
			count := 1
			itemName := args[0].(string)
			if len(args) > 1 {
				count = int(args[1].(float64))
			}
			removedItem := g.Player.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			if len(removedItem) > 0 {
				g.msg(foundation.HiLite("%s removed.", removedItem[0].Name()))
			}
			return nil, nil
		},
		// eg. StackTransForTo(NPC, 'gold', 500)
		"StackTransferTo": func(args ...interface{}) (interface{}, error) {
			count := 1
			targetActor := args[0].(*Actor)
			itemName := args[1].(string)
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			removedItem := g.Player.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			if len(removedItem) > 0 {
				targetActor.GetInventory().AddItems(removedItem)
				g.msg(foundation.HiLite("%s removed.", removedItem[0].Name()))
			}
			return nil, nil
		},

		"HasItem": func(args ...interface{}) (interface{}, error) {
			itemName := args[0].(string)
			count := 1
			if len(args) > 1 {
				count = int(args[1].(float64))
			}
			return g.Player.GetInventory().HasItemWithNameAndCount(itemName, count), nil
		},
		"HasArmorEquipped": func(args ...interface{}) (interface{}, error) {
			return g.Player.GetEquipment().HasArmorEquipped(), nil
		},
		"HasArmorEquippedWithName": func(args ...interface{}) (interface{}, error) {
			armorName := args[0].(string)
			return g.Player.GetEquipment().HasArmorWithNameEquipped(armorName), nil
		},
		"HasWeaponEquippedWithName": func(args ...interface{}) (interface{}, error) {
			weaponName := args[0].(string)
			mainHandItem, hasMainHandItem := g.Player.GetEquipment().GetMainHandItem()
			if !hasMainHandItem {
				return false, nil
			}
			return mainHandItem.InternalName() == weaponName, nil
		},

		// Global Queries & Actions

		// Flags
		"HasFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			return (bool)(g.gameFlags.HasFlag(flagName)), nil
		},
		"SetFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			g.gameFlags.SetFlag(flagName)
			return nil, nil
		},
		"ClearFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			g.gameFlags.ClearFlag(flagName)
			return nil, nil
		},
		"IsMap": func(args ...interface{}) (interface{}, error) {
			mapName := args[0].(string)
			return g.currentMap().GetName() == mapName, nil
		},
		"RandomNumberCode": func(args ...interface{}) (interface{}, error) {
			nameOfDoor := args[0].(string)
			code := g.getRandomNumberCode(nameOfDoor)
			return string(code), nil
		},
		// Time / Turns
		"Turns": func(args ...interface{}) (interface{}, error) {
			return (float64)(g.TurnsTaken()), nil
		},
		"IsTurnsAfter": func(args ...interface{}) (interface{}, error) {
			namedTime := args[0].(string)
			turns := args[1].(float64)
			return g.IsTurnsAfter(namedTime, int(turns)), nil
		},
		"IsMinutesAfter": func(args ...interface{}) (interface{}, error) {
			namedTime := args[0].(string)
			minutes := args[1].(float64)
			return g.IsMinutesAfter(namedTime, int(minutes)), nil
		},
		"IsHoursAfter": func(args ...interface{}) (interface{}, error) {
			namedTime := args[0].(string)
			hours := args[1].(float64)
			return g.IsHoursAfter(namedTime, int(hours)), nil
		},
		"IsDaysAfter": func(args ...interface{}) (interface{}, error) {
			namedTime := args[0].(string)
			days := args[1].(float64)
			return g.IsDaysAfter(namedTime, int(days)), nil
		},

		// Scripts
		"IsScriptRunning": func(args ...interface{}) (interface{}, error) {
			scriptName := args[0].(string)
			return g.scriptRunner.IsScriptRunning(scriptName), nil
		},
		"RunScript": func(args ...interface{}) (interface{}, error) {
			scriptName := args[0].(string)
			g.RunScriptByName(scriptName)
			return nil, nil
		},
		"StopScript": func(args ...interface{}) (interface{}, error) {
			scriptName := args[0].(string)
			g.scriptRunner.StopScript(g.currentMap().GetName(), scriptName)
			return nil, nil
		},
		"RestartScript": func(args ...interface{}) (interface{}, error) {
			scriptName := args[0].(string)
			g.scriptRunner.StopScript(g.currentMap().GetName(), scriptName)
			g.RunScriptByName(scriptName)
			return nil, nil
		},
		"RunScriptKill": func(args ...interface{}) (interface{}, error) {
			killerName := args[0].(string)
			victimName := args[1].(string)
			killer := g.actorWithName(killerName)
			victim := g.actorWithName(victimName)
			g.msg(foundation.HiLite("%s kills %s.", killerName, victimName))
			killScript := g.NewScriptKill(killer, victim)
			g.RunScript(killScript)
			return nil, nil
		},
		"RunScriptLeave": func(args ...interface{}) (interface{}, error) {
			leaverName := args[0].(string)
			isRunning := false
			if len(args) > 1 {
				isRunning = args[1].(bool)
			}
			leaver := g.actorWithName(leaverName)
			killScript := g.NewScriptLeaveMap(leaver, isRunning)
			g.RunScript(killScript)
			return nil, nil
		},

		// Query Containers
		"ContainerWithName": func(args ...interface{}) (interface{}, error) {
			containerName := args[0].(string)
			containers := g.currentMap().GetFilteredObjects(func(c Object) bool {
				return c.GetInternalName() == containerName
			})
			if len(containers) > 0 {
				return containers[0], nil
			}
			return nil, nil
		},

		"IsItemInContainer": func(args ...interface{}) (interface{}, error) {
			container := args[0].(*Container)
			count := 1
			itemName := args[1].(string)
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			return container.HasItemsWithName(itemName, count), nil
		},

		// Query Actors
		"ActorWithName": func(args ...interface{}) (interface{}, error) {
			actorName := args[0].(string)
			actors := g.currentMap().GetFilteredActors(func(a *Actor) bool {
				return a.GetInternalName() == actorName
			})
			if len(actors) > 0 {
				return actors[0], nil
			}
			return nil, nil
		},
		"IsActorWounded": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			return actor.IsWounded(), nil
		},
		"IsActorInShootingRange": func(args ...interface{}) (interface{}, error) {
			attacker := args[0].(*Actor)
			defender := args[1].(*Actor)
			return g.IsInShootingRange(attacker, defender), nil
		},
		"IsActorAtNamedLocation": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			locName := args[1].(string)
			loc := g.currentMap().GetNamedLocation(locName)
			return actor.Position() == loc, nil
		},
		"IsActorDead": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			return !actor.IsAlive(), nil
		},
		"IsActorInCombat": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			if actor.IsInCombat() {
				return true, nil
			}
			return false, nil
		},
		"IsActorInTalkingRange": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			target := args[1].(*Actor)
			return g.IsInTalkingRange(actor, target), nil
		},
		"IsActorInCombatWithPlayer": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			if actor.IsHostileTowards(g.Player) {
				return true, nil
			}
			return false, nil
		},
		// Actor Actions,
		"ActorDropItem": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			count := 1
			itemName := args[1].(string)
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			removedItems := actor.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			for _, item := range removedItems {
				g.addItemToMap(item, actor.Position())
			}
			if len(removedItems) > 0 {
				first := removedItems[0]
				displayName := first.Name()
				g.msg(foundation.HiLite("%s dropped %s.", actor.Name(), displayName))
			}
			return nil, nil
		},
		"PlayerAddCyberware": func(args ...interface{}) (interface{}, error) {
			cyberwareName := args[0].(string)
			g.playerAddCyberware(NewCyberWareFromString(cyberwareName))
			return nil, nil
		},
		// Global Actions
		"SaveTimeNow": func(args ...interface{}) (interface{}, error) {
			nameForTime := args[0].(string)
			g.SaveTimeNow(nameForTime)
			return nil, nil
		},
		"AdvanceTimeByMinutes": func(args ...interface{}) (interface{}, error) {
			minutes := int(args[0].(float64))
			g.advanceTime(time.Minute * time.Duration(minutes))
			return nil, nil
		},
		"AddChatter": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			chatter := args[1].(string)
			g.tryAddChatter(actor, chatter)
			return nil, nil
		},
		"Hilite": func(args ...interface{}) (interface{}, error) {
			text := args[0].(string)
			g.msg(foundation.HiLite(text))
			return nil, nil
		},

		// Actors Goals
		"SetGoalMoveToNamedLocation": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			locName := args[1].(string)
			loc := g.currentMap().GetNamedLocation(locName)
			actor.SetGoal(GoalWalkToLocation(loc))
			return nil, nil
		},
		"SetGoalMoveToSpawn": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			actor.SetGoal(GoalMoveToSpawn())
			return nil, nil
		},
		"SetGoalMoveIntoShootingRange": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			target := args[1].(*Actor)
			actor.tryEquipRangedWeapon()
			actor.SetGoal(GoalMoveIntoShootingRange(target))
			return nil, nil
		},
		"SetGoalKill": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			target := args[1].(*Actor)
			actor.SetGoal(GoalKillActor(actor, target))
			return nil, nil
		},

		// Container Actions
		"ContainerRemoveItem": func(args ...interface{}) (interface{}, error) {
			container := args[0].(*Container)
			itemName := args[1].(string)
			count := 1
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			return container.RemoveItemsWithName(itemName, count), nil
		},
		"ContainerAddItem": func(args ...interface{}) (interface{}, error) {
			container := args[0].(*Container)
			if item, isItem := args[1].(foundation.Item); isItem {
				container.AddItem(item)
				return nil, nil
			} else if items, isItems := args[1].([]foundation.Item); isItems {
				container.AddItems(items)
				return nil, nil
			}
			newItem := g.NewItemFromString(args[1].(string))
			container.AddItem(newItem)
			return nil, nil
		},
		"ActorRemoveItem": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			count := 1
			itemName := args[1].(string)
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			removedItems := actor.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			return removedItems, nil
		},
		"ActorAddItem": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			if item, isItem := args[1].(foundation.Item); isItem {
				actor.GetInventory().AddItem(item)
				return nil, nil
			} else if items, isItems := args[1].([]foundation.Item); isItems {
				actor.GetInventory().AddItems(items)
				return nil, nil
			}
			newItem := g.NewItemFromString(args[1].(string))
			actor.GetInventory().AddItem(newItem)
			return nil, nil
		},
	}
}

func (g *GameState) playerAddCyberware(cyberware CyberWare) {
	g.ui.FadeToBlack()
	g.advanceTime(time.Hour * 24)
	if g.Player.HasWatch() {
		g.printTime()
	}
	g.ui.FadeFromBlack()

	g.Player.AddCyberWare(cyberware)
	g.msg(foundation.HiLite("%s installed.", cyberware.String()))
}

func moveAwayFromActor(g *GameState, a *Actor, target *Actor) (fsmai.TransitionEvent, int) {
	nextMovePos := a.getMoveAwayFromActor(g, target)
	if nextMovePos == a.Position() {
		return fsmai.NoEvent, a.TimeNeededForMovement()
	}

	if !g.currentMap().IsWalkableFor(nextMovePos, a) {
		return fsmai.NoEvent, a.TimeNeededForMovement()
	}

	g.ui.AddAnimations(g.actorMoveAnimated(a, nextMovePos))

	return fsmai.NoEvent, a.TimeNeededForMovement()
}
func moveTowardsActor(g *GameState, a *Actor, target *Actor) (fsmai.TransitionEvent, int) {
	nextMovePos := a.getMoveTowardsActor(g, target)
	if nextMovePos == a.Position() {
		return fsmai.NoEvent, a.TimeNeededForMovement()
	}

	if !g.currentMap().IsWalkableFor(nextMovePos, a) {
		return fsmai.NoEvent, a.TimeNeededForMovement()
	}

	g.ui.AddAnimations(g.actorMoveAnimated(a, nextMovePos))

	return fsmai.NoEvent, a.TimeNeededForMovement()
}
func walkTowards(g *GameState, a *Actor, targetPos geometry.Point) (fsmai.TransitionEvent, int) {
	nextMovePos := a.getMoveTowards(g, targetPos)
	if nextMovePos == a.Position() {
		return fsmai.NoEvent, a.TimeNeededForMovement()
	}

	if !g.currentMap().IsWalkableFor(nextMovePos, a) {
		return fsmai.NoEvent, a.TimeNeededForMovement()
	}

	g.ui.AddAnimations(g.actorMoveAnimated(a, nextMovePos))

	return fsmai.NoEvent, a.TimeNeededForMovement()
}
func strideTowards(g *GameState, a *Actor, targetPos geometry.Point) (fsmai.TransitionEvent, int) {
	nextMovePos := a.getMoveTowards(g, targetPos)
	timeForMove := max(a.TimeNeededForMovement(), 10)
	if nextMovePos == a.Position() {
		return fsmai.NoEvent, timeForMove
	}

	if !g.currentMap().IsWalkableFor(nextMovePos, a) {
		return fsmai.NoEvent, timeForMove
	}

	g.ui.AddAnimations(g.actorMoveAnimated(a, nextMovePos))

	return fsmai.NoEvent, timeForMove
}

func LoadScript(dataDir string, name string, condFuncs map[string]govaluate.ExpressionFunction) ActionScript {
	filePath := path.Join(dataDir, "scripts", name+".rec")
	records := recfile.ReadMultiAndClose(fxtools.MustOpen(filePath))
	return NewActionScript(name, records, condFuncs)
}

func NewActionScript(name string, records map[string][]recfile.Record, condFuncs map[string]govaluate.ExpressionFunction) ActionScript {
	frames := FramesFromRecords(records["frames"], condFuncs)

	script := ActionScript{
		Name:      name,
		Frames:    frames,
		Variables: make(map[string]interface{}),
	}
	if len(records["cancel"]) > 0 {
		cancelRecord := records["cancel"][0]
		script.CancelFrame = NewScriptFrameFromRecord(cancelRecord, condFuncs)
	}

	if len(records["outcomes"]) > 0 {
		for _, outcome := range records["outcomes"] {
			outcomeFrame := NewScriptFrameFromRecord(outcome, condFuncs)
			if outcomeFrame.condition != nil {
				script.Outcomes = append(script.Outcomes, outcomeFrame)
			}
		}
	}

	if len(records["definitions"]) > 0 {
		for _, defNode := range records["definitions"] {
			varName := ""
			var varValue *govaluate.EvaluableExpression
			for _, f := range defNode {
				switch strings.ToLower(f.Name) {
				case "var":
					varName = f.Value
				case "set":
					varValue, _ = govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
				}
			}
			if varName != "" && varValue != nil {
				script.Variables[varName], _ = varValue.Evaluate(nil)
			}
		}

	}
	return script
}

func NewScriptFrameFromRecord(outcome recfile.Record, condFuncs map[string]govaluate.ExpressionFunction) UserScriptFrame {
	outcomeFrame := UserScriptFrame{}
	for _, f := range outcome {
		switch strings.ToLower(f.Name) {
		case "if":
			cond, parseErr := govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
			if parseErr != nil {
				panic(parseErr)
			} else {
				outcomeFrame.condition = cond
			}
		case "do":
			action, parseErr := govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
			if parseErr != nil {
				panic(parseErr)
			} else {
				outcomeFrame.actions = append(outcomeFrame.actions, action)
			}
		}
	}
	return outcomeFrame
}

func FramesFromRecords(records []recfile.Record, condFuncs map[string]govaluate.ExpressionFunction) []ScriptFrame {
	frames := make([]ScriptFrame, len(records))
	for i, record := range records {
		frames[i] = NewScriptFrame(record, condFuncs)
	}
	return frames
}

func NewScriptFrame(record recfile.Record, condFuncs map[string]govaluate.ExpressionFunction) UserScriptFrame {
	frame := UserScriptFrame{}
	for _, f := range record {
		switch strings.ToLower(f.Name) {
		case "if":
			cond, parseErr := govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
			if parseErr != nil {
				panic(parseErr)
			} else {
				frame.condition = cond
			}
		case "do":
			action, parseErr := govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
			if parseErr != nil {
				panic(parseErr)
			} else {
				frame.actions = append(frame.actions, action)
			}
		}
	}
	return frame
}
