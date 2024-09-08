package game

import (
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"path"
	"strings"
)

type ScriptFrame struct {
	// The condition for the frame.
	Condition *govaluate.EvaluableExpression
	// The actions for the frame.
	Actions []*govaluate.EvaluableExpression
}
type ActionScript struct {
	Name string

	Variables map[string]interface{}

	Frames []ScriptFrame

	Outcomes []ScriptFrame

	CancelFrame ScriptFrame
}

func (g *GameState) getScriptFuncs() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		// Queries
		"HasFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			return (bool)(g.gameFlags.HasFlag(flagName)), nil
		},
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
		"Turns": func(args ...interface{}) (interface{}, error) {
			return (float64)(g.TurnsTaken), nil
		},
		"IsInShootingRange": func(args ...interface{}) (interface{}, error) {
			attacker := args[0].(*Actor)
			defender := args[1].(*Actor)
			return g.IsInShootingRange(attacker, defender), nil
		},
		"IsDead": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			return !actor.IsAlive(), nil
		},
		"IsInCombat": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			if actor.IsInCombat() {
				return true, nil
			}
			return false, nil
		},
		"IsInCombatWithPlayer": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			if actor.IsHostileTowards(g.Player) {
				return true, nil
			}
			return false, nil
		},
		// Actions
		"AddChatter": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			chatter := args[1].(string)
			g.tryAddChatter(actor, chatter)
			return nil, nil
		},
		"SetFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			g.gameFlags.SetFlag(flagName)
			return nil, nil
		},
		"SetGoalMoveToNamedLocation": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			locName := args[1].(string)
			loc := g.currentMap().GetNamedLocation(locName)
			actor.SetGoal(ActorGoal{
				Action: func(g *GameState, a *Actor) int {
					return moveTowards(g, a, loc)
				},
				Achieved: func(g *GameState, a *Actor) bool {
					return a.Position() == loc
				},
			})
			return nil, nil
		},
		"SetGoalMoveToSpawn": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			actor.SetGoal(ActorGoal{
				Action: func(g *GameState, a *Actor) int {
					targetPos := a.SpawnPosition
					return moveTowards(g, a, targetPos)
				},
				Achieved: func(g *GameState, a *Actor) bool {
					return a.Position() == a.SpawnPosition
				},
			})
			return nil, nil
		},
		"SetGoalMoveIntoShootingRange": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			target := args[1].(*Actor)
			actor.tryEquipRangedWeapon()
			actor.SetGoal(ActorGoal{
				Action: func(g *GameState, a *Actor) int {
					return moveIntoShootingRange(g, a, target)
				},
				Achieved: func(g *GameState, a *Actor) bool {
					return g.IsInShootingRange(a, target)
				},
			})
			return nil, nil
		},
		"SetGoalKill": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			target := args[1].(*Actor)
			actor.SetGoal(g.getKillGoal(actor, target))
			return nil, nil
		},
	}
}

func (g *GameState) getKillGoal(attacker *Actor, victim *Actor) ActorGoal {
	return ActorGoal{
		Action: func(g *GameState, a *Actor) int {
			return tryKill(g, a, victim)
		},
		Achieved: func(g *GameState, a *Actor) bool {
			return !victim.IsAlive() || !attacker.IsAlive()
		},
	}
}
func tryKill(g *GameState, a *Actor, target *Actor) int {
	if !g.IsInShootingRange(a, target) {
		return moveIntoShootingRange(g, a, target)
	}
	if !a.GetEquipment().HasRangedWeaponInMainHand() {
		a.tryEquipRangedWeapon()
	}
	mainHandItem, hasMainHandItem := a.GetEquipment().GetMainHandItem()
	if hasMainHandItem && mainHandItem.IsRangedWeapon() {

		if !mainHandItem.GetWeapon().HasAmmo() && mainHandItem.GetWeapon().NeedsAmmo() {
			g.actorReloadMainHandWeapon(a)
			return a.timeNeededForActions()
		}

		g.ui.AddAnimations(g.actorRangedAttack(a, mainHandItem, mainHandItem.GetCurrentAttackMode(), target, 0))
		return mainHandItem.GetCurrentAttackMode().TUCost
	}

	return a.timeNeededForActions()
}
func moveIntoShootingRange(g *GameState, a *Actor, target *Actor) int {
	weaponRange := a.GetWeaponRange()
	targetPos := g.getShootingRangePosition(a, weaponRange, target)

	return moveTowards(g, a, targetPos)
}

func moveTowards(g *GameState, a *Actor, targetPos geometry.Point) int {
	nextMovePos := a.getMoveTowards(g, targetPos)
	if nextMovePos == a.Position() {
		return a.timeEnergy
	}

	if !g.currentMap().IsCurrentlyPassable(nextMovePos) {
		return a.timeEnergy
	}

	g.ui.AddAnimations(g.actorMoveAnimated(a, nextMovePos))

	return a.timeNeededForMovement()
}

func LoadScript(dataDir string, name string, condFuncs map[string]govaluate.ExpressionFunction) ActionScript {
	filePath := path.Join(dataDir, "scripts", name+".rec")
	records := recfile.ReadMulti(fxtools.MustOpen(filePath))
	return NewActionScript(name, records, condFuncs)
}

func NewActionScript(name string, records map[string][]recfile.Record, condFuncs map[string]govaluate.ExpressionFunction) ActionScript {
	frames := FramesFromRecords(records["frames"], condFuncs)

	script := ActionScript{
		Name:        name,
		Frames:      frames,
		Variables:   make(map[string]interface{}),
		CancelFrame: NewScriptFrameFromRecord(records["cancel"][0], condFuncs),
	}

	if len(records["outcomes"]) > 0 {
		for _, outcome := range records["outcomes"] {
			outcomeFrame := NewScriptFrameFromRecord(outcome, condFuncs)
			if outcomeFrame.Condition != nil {
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

func NewScriptFrameFromRecord(outcome recfile.Record, condFuncs map[string]govaluate.ExpressionFunction) ScriptFrame {
	outcomeFrame := ScriptFrame{}
	for _, f := range outcome {
		switch strings.ToLower(f.Name) {
		case "if":
			cond, parseErr := govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
			if parseErr != nil {
				panic(parseErr)
			} else {
				outcomeFrame.Condition = cond
			}
		case "do":
			action, parseErr := govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
			if parseErr != nil {
				panic(parseErr)
			} else {
				outcomeFrame.Actions = append(outcomeFrame.Actions, action)
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

func NewScriptFrame(record recfile.Record, condFuncs map[string]govaluate.ExpressionFunction) ScriptFrame {
	frame := ScriptFrame{}
	for _, f := range record {
		switch strings.ToLower(f.Name) {
		case "if":
			cond, parseErr := govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
			if parseErr != nil {
				panic(parseErr)
			} else {
				frame.Condition = cond
			}
		case "do":
			action, parseErr := govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
			if parseErr != nil {
				panic(parseErr)
			} else {
				frame.Actions = append(frame.Actions, action)
			}
		}
	}
	return frame
}
