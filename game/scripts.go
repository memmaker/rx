package game

import (
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"math"
	"strings"
)

func (g *GameState) NewScriptLeaveMap(leaver *Actor, running bool) ActionScript {
	if leaver == nil {
		g.msg(foundation.HiLite("NewScriptLeaveMap: leaver is nil"))
		return ActionScript{}
	}
	transitionPos := leaver.Position()

	g.updateFoVAndDijkstraMap(leaver)

	minDist := math.MaxInt
	for reachablePos, dist := range leaver.DijkstraMap {
		if g.currentMap().IsTransitionAt(reachablePos) && dist < minDist {
			transitionPos = reachablePos
			minDist = dist
		}
	}

	if transitionPos == leaver.Position() && !g.currentMap().IsTransitionAt(transitionPos) {
		g.msg(foundation.Msg("No transition found"))
		return ActionScript{}
	}

	return g.NewScriptLeaveMapAt(leaver, running, transitionPos)
}

func (g *GameState) NewScriptLeaveMapAtLocation(leaver *Actor, running bool, location string) ActionScript {
	transitionPos := g.currentMap().GetNamedLocation(location)
	return g.NewScriptLeaveMapAt(leaver, running, transitionPos)
}

func (g *GameState) NewScriptLeaveMapAt(leaver *Actor, running bool, transitionPos geometry.Point) ActionScript {
	transition, _ := g.currentMap().GetTransitionAt(transitionPos)

	var goToTransition ActorGoal
	if running {
		goToTransition = GoalRunToLocation(transitionPos)
	} else {
		goToTransition = GoalWalkToLocation(transitionPos)
	}

	return ActionScript{
		Name: fmt.Sprintf("leaves_map_%s", leaver.GetInternalName()),
		Frames: []ScriptFrame{
			FrameSetGoal(goToTransition, leaver),
			FrameRemoveFromMap(leaver, func(actor *Actor) {
				g.actorTransition(g.currentMap(), actor, transition)
			}).WithCondition(func() bool {
				return leaver.Position() == transitionPos
			}),
		},
		Outcomes: []ScriptFrame{
			BasicScriptFrame{
				condition: func() bool {
					return leaver.Position() == transitionPos && leaver.HasFlag(foundation.FlagWantsToTransition)
				},
			},
			BasicScriptFrame{
				condition: func() bool { return !leaver.IsAlive() },
			},
		},
		CancelFrame: FrameRemoveFromMap(leaver, func(actor *Actor) {
			g.actorTransition(g.currentMap(), actor, transition)
		}),
	}
}

func (g *GameState) NewScriptKill(killer, victim *Actor) ActionScript {
	if killer == nil || victim == nil {
		g.msg(foundation.HiLite("NewScriptKill: killer or victim is nil"))
		return ActionScript{}
	}
	g.updateFoVAndDijkstraMap(victim)
	killer.TryEquipRangedWeaponFirst()

	return ActionScript{
		Name: fmt.Sprintf("%s_kills_%s", killer.GetInternalName(), victim.GetInternalName()),
		Frames: []ScriptFrame{
			FrameSetGoal(GoalMoveIntoShootingRange(victim), killer).WithAction(func() {
				g.tryAddRandomChatter(killer, foundation.ChatterOnTheWayToAKill)
			}),
			FrameSetGoal(GoalKillActor(victim), killer).
				WithCondition(func() bool {
					return g.IsInShootingRange(killer, victim)
				}).
				WithAction(func() {
					g.tryAddRandomChatter(killer, foundation.ChatterKillOneLiner)
				}),
		},

		Outcomes: []ScriptFrame{
			FrameSetGoal(GoalMoveToSpawn(), killer).WithCondition(func() bool {
				return killer.IsAlive() && !victim.IsAlive()
			}),
			FrameSetGoal(GoalMoveToSpawn(), victim).WithCondition(func() bool {
				return !killer.IsAlive() && victim.IsAlive()
			}),
			FrameSetGoal(GoalMoveToSpawn(), victim).WithCondition(func() bool {
				return !killer.IsAlive() && !victim.IsAlive()
			}),
		},
		CancelFrame: FrameSetGoal(GoalMoveToSpawn(), killer, victim),
	}
}

type SetGoalFrame struct {
	actors      []*Actor
	goal        ActorGoal
	cond        func() bool
	moreActions []func()
}

func (m SetGoalFrame) IsEmpty() bool {
	return len(m.actors) == 0
}

func (m SetGoalFrame) WithAction(action func()) SetGoalFrame {
	m.moreActions = append(m.moreActions, action)
	return m
}

func (m SetGoalFrame) Condition(vars map[string]interface{}) bool {
	if m.cond != nil {
		return m.cond()
	}
	return true
}

func (m SetGoalFrame) ExecuteActions(vars map[string]interface{}) {
	for _, actor := range m.actors {
		actor.SetGoal(m.goal)
	}
	for _, action := range m.moreActions {
		action()
	}
}

func (m SetGoalFrame) String() string {
	names := fxtools.MapSlice(m.actors, func(t1 *Actor) string {
		return t1.Name()
	})
	return fmt.Sprintf("SetGoalFrame{actors: %s}", strings.Join(names, ", "))
}

func (m SetGoalFrame) WithCondition(cond func() bool) SetGoalFrame {
	m.cond = cond
	return m
}

func FrameSetGoal(goal ActorGoal, actors ...*Actor) SetGoalFrame {
	return SetGoalFrame{actors: actors, goal: goal}
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
