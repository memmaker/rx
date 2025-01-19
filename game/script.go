package game

import (
	"contractor/foundation"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"math"
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
	if f.condition == nil {
		return "<no condition>"
	}
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

func (g *GameState) IterateItemsInAllInventories(mapName string, action func(*Actor, foundation.Item)) {
	searchedMap, loaded := g.activeMaps[mapName]
	if !loaded {
		return
	}
	for _, actor := range searchedMap.Actors() {
		for _, item := range actor.GetInventory().GetItems() {
			action(actor, item)
		}
	}
}
func (g *GameState) IterateItemsOnMapZoneRecursively(mapName string, zoneName string, action func(foundation.Item)) {
	searchedMap, loaded := g.activeMaps[mapName]
	if !loaded {
		return
	}
	for _, item := range searchedMap.Items() {
		if !searchedMap.IsZoneAt(item.Position(), zoneName) {
			continue
		}
		action(item)
	}
	for _, obj := range searchedMap.Objects() {
		if !searchedMap.IsZoneAt(obj.Position(), zoneName) {
			continue
		}
		if container, isContainer := obj.(*Container); isContainer {
			container.IterateItems(action)
		}
	}
}
func (g *GameState) IterateItemsOnMapRecursively(mapName string, action func(foundation.Item)) {
	searchedMap, loaded := g.activeMaps[mapName]
	if !loaded {
		return
	}
	for _, item := range searchedMap.Items() {
		action(item)
	}
	for _, obj := range searchedMap.Objects() {
		if container, isContainer := obj.(*Container); isContainer {
			container.IterateItems(action)
		}
	}
}

func (g *GameState) playerAddCyberware(cyberware CyberWare) {
	g.ui.FadeToBlack()
	g.advanceTime(time.Hour * 24)
	if g.Player.HasWatch() {
		g.ShowDateTime()
	}
	g.ui.FadeFromBlack()

	g.Player.AddCyberWare(cyberware)
	g.msg(foundation.HiLite("%s installed.", cyberware.String()))
}

func moveAwayFromActor(g *GameState, a *Actor, target *Actor) (TransitionEvent, int) {
	nextMovePos := g.currentMap().GetMoveAwayFromActor(a, target)
	return g.actorMakeMove(a, nextMovePos, true)
}

func LoadScript(dataDir string, name string, condFuncs map[string]govaluate.ExpressionFunction) ActionScript {
	filePath := path.Join(dataDir, "scripts", name+".rec")
	records, _ := recfile.ReadMultiAndClose(fxtools.MustOpen(filePath))
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
					var parseErr error
					varValue, parseErr = govaluate.NewEvaluableExpressionWithFunctions(f.Value, condFuncs)
					if parseErr != nil {
						panic(parseErr)
					}
				}
			}
			if varName != "" && varValue != nil {
				evaledVarValue, evalErr := varValue.Evaluate(nil)
				if evalErr != nil {
					panic(evalErr)
				}
				if evaledVarValue == nil {
					return ActionScript{} // could not evaluate the required variable
				}
				script.Variables[varName] = evaledVarValue
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

func (g *GameState) NewScriptLeaveMap(leaver *Actor, running bool) ActionScript {
	if leaver == nil {
		g.msg(foundation.HiLite("NewScriptLeaveMap: leaver is nil"))
		return ActionScript{}
	}
	transitionPos := ""

	minDist := math.MaxInt
	for reachablePos, dist := range leaver.GetDijkstraMap() {
		if g.currentMap().IsTransitionAt(reachablePos) && dist < minDist {
			transitionPos = g.currentMap().GetNamedLocationByPos(reachablePos)
			minDist = dist
		}
	}

	if transitionPos == "" {
		g.msg(foundation.Msg("No transition found"))
		return ActionScript{}
	}

	return g.NewScriptLeaveMapAt(leaver, running, transitionPos)
}

func (g *GameState) NewScriptLeaveMapAtLocation(leaver *Actor, running bool, location string) ActionScript {
	return g.NewScriptLeaveMapAt(leaver, running, location)
}

func (g *GameState) NewScriptLeaveMapAt(leaver *Actor, running bool, locationName string) ActionScript {
	gMap := g.currentMap()
	return ActionScript{
		Name: fmt.Sprintf("leaves_map_%s", leaver.GetInternalName()),
		Frames: []ScriptFrame{
			FrameSetState(StateScripted, LocationEvent{Event: EventNone, Location: MapPosition{
				MapName:      gMap.GetName(),
				LocationName: locationName,
				Position:     gMap.GetNamedLocation(locationName),
			}}, leaver),
			FrameRemoveFromMap(leaver, func(actor *Actor) {
				transitionPos := gMap.GetNamedLocation(locationName)
				transitionTarget, exists := gMap.GetTransitionAt(transitionPos)
				if exists {
					g.actorTransition(gMap, actor, transitionTarget)
				}
			}).WithCondition(func() bool {
				transitionPos := gMap.GetNamedLocation(locationName)
				return leaver.Position() == transitionPos
			}),
		},
		Outcomes: []ScriptFrame{
			BasicScriptFrame{
				condition: func() bool {
					transitionPos := gMap.GetNamedLocation(locationName)
					return leaver.Position() == transitionPos
				},
			},
			BasicScriptFrame{
				condition: func() bool { return !leaver.IsAlive() },
			},
		},
		CancelFrame: FrameRemoveFromMap(leaver, func(actor *Actor) {
			transitionPos := gMap.GetNamedLocation(locationName)
			transitionTarget, exists := gMap.GetTransitionAt(transitionPos)
			if exists {
				g.actorTransition(gMap, actor, transitionTarget)
			}
		}),
	}
}

func (g *GameState) NewScriptKill(killer, victim *Actor) ActionScript {
	if killer == nil || victim == nil {
		g.msg(foundation.HiLite("NewScriptKill: killer or victim is nil"))
		return ActionScript{}
	}

	killer.TryEquipRangedWeaponFirst()

	return ActionScript{
		Name: fmt.Sprintf("%s_kills_%s", killer.GetInternalName(), victim.GetInternalName()),
		Frames: []ScriptFrame{
			FrameSetState(StateKill, ActorEvent{Event: EventProvoked, Actor: victim}, killer).WithAction(func() {
				g.tryAddRandomChatter(killer, foundation.ChatterOnTheWayToAKill)
			}),
			FrameSetState(StateKill, ActorEvent{Event: EventProvoked, Actor: victim}, killer).
				WithCondition(func() bool {
					return g.IsInShootingRange(killer, victim)
				}).
				WithAction(func() {
					g.tryAddRandomChatter(killer, foundation.ChatterKillOneLiner)
				}),
		},

		Outcomes: []ScriptFrame{
			FrameSetState(StateIdle, NoEvent, killer).WithCondition(func() bool {
				return killer.IsAlive() && !victim.IsAlive()
			}),
			FrameSetState(StateIdle, NoEvent, victim).WithCondition(func() bool {
				return !killer.IsAlive() && victim.IsAlive()
			}),
			FrameSetState(StateIdle, NoEvent, victim).WithCondition(func() bool {
				return !killer.IsAlive() && !victim.IsAlive()
			}),
		},
		CancelFrame: FrameSetState(StateIdle, NoEvent, killer, victim),
	}
}

func (g *GameState) RunScriptByName(scriptName string) {
	if fxtools.LooksLikeAFunction(scriptName) {
		name, args := fxtools.GetNameAndArgs(scriptName)
		switch name {
		case "LeaveMapAt":
			actorName := args.Get(0)
			locationName := args.Get(1)
			running := false
			if len(args) > 2 {
				running = args.GetBool(2)
			}
			actor := g.actorWithName(actorName)
			leaveMapAtLocation := g.NewScriptLeaveMapAtLocation(actor, running, locationName)
			g.Scripts.Run(leaveMapAtLocation)
		}
	} else {
		g.Scripts.RunScriptByName(g.config.DataRootDir, scriptName, g.GetScriptFuncs())
	}
}

type SetStateFrame struct {
	actors      []*Actor
	cond        func() bool
	moreActions []func()
	state       StateName
	initEvent   TransitionEvent
}

func (m SetStateFrame) IsEmpty() bool {
	return len(m.actors) == 0
}

func (m SetStateFrame) WithAction(action func()) SetStateFrame {
	m.moreActions = append(m.moreActions, action)
	return m
}

func (m SetStateFrame) Condition(vars map[string]interface{}) bool {
	if m.cond != nil {
		return m.cond()
	}
	return true
}

func (m SetStateFrame) ExecuteActions(vars map[string]interface{}) {
	for _, actor := range m.actors {
		actor.FSM.SetState(m.state, m.initEvent)
	}
	for _, action := range m.moreActions {
		action()
	}
}

func (m SetStateFrame) String() string {
	names := fxtools.MapSlice(m.actors, func(t1 *Actor) string {
		return t1.Name()
	})
	return fmt.Sprintf("SetStateFrame{actors: %s}", strings.Join(names, ", "))
}

func (m SetStateFrame) WithCondition(cond func() bool) SetStateFrame {
	m.cond = cond
	return m
}

func FrameSetState(state StateName, initEvent TransitionEvent, actors ...*Actor) SetStateFrame {
	return SetStateFrame{actors: actors, state: state, initEvent: initEvent}
}

type RemoveFromMapFrame struct {
	actor  *Actor
	cond   func() bool
	remove func(actor *Actor)
}

func (r RemoveFromMapFrame) IsEmpty() bool {
	return r.actor == nil
}

func (r RemoveFromMapFrame) Condition(m map[string]interface{}) bool {
	if r.cond != nil {
		return r.cond()
	}
	return true
}

func (r RemoveFromMapFrame) ExecuteActions(m map[string]interface{}) {
	r.remove(r.actor)
}

func (r RemoveFromMapFrame) String() string {
	return fmt.Sprintf("RemoveFromMapFrame{actor: %s}", r.actor.Name())
}

func (r RemoveFromMapFrame) WithCondition(f func() bool) RemoveFromMapFrame {
	r.cond = f
	return r
}

func FrameRemoveFromMap(leaver *Actor, remove func(actor *Actor)) RemoveFromMapFrame {
	return RemoveFromMapFrame{actor: leaver, remove: remove}
}

type BasicScriptFrame struct {
	name      string
	condition func() bool
	action    func()
}

func (b BasicScriptFrame) IsEmpty() bool {
	return b.condition == nil && b.action == nil
}

func (b BasicScriptFrame) Condition(m map[string]interface{}) bool {
	if b.condition != nil {
		return b.condition()
	}
	return true
}

func (b BasicScriptFrame) ExecuteActions(m map[string]interface{}) {
	if b.action != nil {
		b.action()
	}
}

func (b BasicScriptFrame) String() string {
	return b.name
}
