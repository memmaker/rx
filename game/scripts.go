package game

import (
	"RogueUI/foundation"
	"fmt"
)

func (g *GameState) NewScriptKill(killer, victim *Actor) ActionScript {
	if killer == nil || victim == nil {
		g.msg(foundation.HiLite("NewScriptKill: killer or victim is nil"))
		return ActionScript{}
	}
	return ActionScript{
		Name: fmt.Sprintf("%s_kills_%s", killer.GetInternalName(), victim.GetInternalName()),
		Frames: []ScriptFrame{
			FrameSetGoal(GoalMoveIntoShootingRange(victim), killer).WithAction(func() {
				g.tryAddRandomChatter(killer, foundation.ChatterOnTheWayToAKill)
			}),
			FrameSetGoal(GoalKillActor(killer, victim), killer).
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
	return fmt.Sprintf("SetGoalFrame{actors: %v}", m.actors)
}

func (m SetGoalFrame) WithCondition(cond func() bool) SetGoalFrame {
	m.cond = cond
	return m
}

func FrameSetGoal(goal ActorGoal, actors ...*Actor) SetGoalFrame {
	return SetGoalFrame{actors: actors, goal: goal}
}
