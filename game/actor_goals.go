package game

import (
	"contractor/foundation"
	"contractor/fsmai"
	"contractor/gridmap"
	"encoding/gob"
	"github.com/memmaker/go/geometry"
)

type ActorGoal interface {
	Description() string
	IsCombatGoal() bool
	Action(ga *GameState, a *Actor) (fsmai.TransitionEvent, int)
	IsHostilityTowards(attacker *Actor) bool
	Achieved(ga *GameState, a *Actor) bool
}

func GoalMoveToSpawn() ActorGoalMoveToSpawn {
	return ActorGoalMoveToSpawn{}
}

type ActorGoalMoveToSpawn struct{}

func (g ActorGoalMoveToSpawn) Description() string {
	return "walking home"
}

func (g ActorGoalMoveToSpawn) IsCombatGoal() bool {
	return false
}

func (g ActorGoalMoveToSpawn) Action(ga *GameState, a *Actor) (fsmai.TransitionEvent, int) {
	targetPos := a.SpawnPosition
	return runTowards(ga, a, targetPos)
}

func (g ActorGoalMoveToSpawn) IsHostilityTowards(victim *Actor) bool {
	return false
}

func (g ActorGoalMoveToSpawn) Achieved(ga *GameState, a *Actor) bool {
	return a.Position() == a.SpawnPosition
}

func GoalFollowLeader(leader *Actor) ActorGoalFollowLeader {
	return ActorGoalFollowLeader{leaderID: leader.ID()}
}

type ActorGoalFollowLeader struct {
	leaderID gridmap.ActorID
}

func (g ActorGoalFollowLeader) Description() string {
	return "following"
}

func (g ActorGoalFollowLeader) IsCombatGoal() bool {
	return false
}

func (g ActorGoalFollowLeader) Action(ga *GameState, a *Actor) (fsmai.TransitionEvent, int) {
	leader := ga.resolveActorID(g.leaderID)
	if leader == nil {
		return fsmai.NoEvent, 0
	}
	return moveTowardsActor(ga, a, leader, 2)
}

func (g ActorGoalFollowLeader) IsHostilityTowards(victim *Actor) bool {
	return false
}

func (g ActorGoalFollowLeader) Achieved(ga *GameState, a *Actor) bool {
	leader := ga.resolveActorID(g.leaderID)
	return leader == nil || !leader.IsAlive() || !a.IsAlive()
}

func GoalMoveIntoShootingRange(target *Actor) ActorGoalMoveIntoShootingRange {
	return ActorGoalMoveIntoShootingRange{targetID: target.ID()}
}

type ActorGoalMoveIntoShootingRange struct {
	targetID gridmap.ActorID
}

func (g ActorGoalMoveIntoShootingRange) Description() string {
	return "moving aggressively"
}

func (g ActorGoalMoveIntoShootingRange) IsCombatGoal() bool {
	return true
}

func (g ActorGoalMoveIntoShootingRange) Action(ga *GameState, a *Actor) (fsmai.TransitionEvent, int) {
	targetActor := ga.resolveActorID(g.targetID)
	return moveTowardsActor(ga, a, targetActor, 1)
}

func (g ActorGoalMoveIntoShootingRange) IsHostilityTowards(victim *Actor) bool {
	return victim.ID() == g.targetID
}

func (g ActorGoalMoveIntoShootingRange) Achieved(ga *GameState, a *Actor) bool {
	targetActor := ga.resolveActorID(g.targetID)
	return targetActor == nil || ga.IsInShootingRange(a, targetActor)
}

func GoalFleeFromActor(threat *Actor) ActorGoalFleeFromActor {
	return ActorGoalFleeFromActor{
		threatID: threat.ID(),
	}
}

type ActorGoalFleeFromActor struct {
	threatID gridmap.ActorID
}

func (g ActorGoalFleeFromActor) Description() string {
	return "fleeing"
}

func (g ActorGoalFleeFromActor) IsCombatGoal() bool {
	return false
}

func (g ActorGoalFleeFromActor) Action(ga *GameState, a *Actor) (fsmai.TransitionEvent, int) {
	threat := ga.resolveActorID(g.threatID)
	return moveAwayFromActor(ga, a, threat)
}

func (g ActorGoalFleeFromActor) IsHostilityTowards(victim *Actor) bool {
	return false
}

func (g ActorGoalFleeFromActor) Achieved(ga *GameState, a *Actor) bool {
	threatActor := ga.resolveActorID(g.threatID)
	return threatActor == nil || !threatActor.IsAlive() || !a.IsAlive()
}

func GoalKillActor(target *Actor) ActorGoalKillActor {
	return ActorGoalKillActor{
		victimID: target.ID(),
	}
}

type ActorGoalKillActor struct {
	victimID gridmap.ActorID
}

func (g ActorGoalKillActor) Description() string {
	return "attacking"
}

func (g ActorGoalKillActor) IsCombatGoal() bool {
	return true
}

func (g ActorGoalKillActor) Action(ga *GameState, a *Actor) (fsmai.TransitionEvent, int) {
	victim := ga.resolveActorID(g.victimID)
	return tryKill(ga, a, victim)
}

func (g ActorGoalKillActor) IsHostilityTowards(victim *Actor) bool {
	return victim.ID() == g.victimID
}

func (g ActorGoalKillActor) Achieved(ga *GameState, a *Actor) bool {
	victim := ga.resolveActorID(g.victimID)
	return victim == nil || !victim.IsAlive() || !a.IsAlive()
}

func GoalWalkToLocationOnMap(localTransitionName geometry.Point, targetMap string, targetLocation geometry.Point) ActorGoalMoveToLocationOnMap {
	return ActorGoalMoveToLocationOnMap{
		Transition:     localTransitionName,
		TargetMap:      targetMap,
		TargetLocation: targetLocation,
		Running:        false,
	}
}

type ActorGoalMoveToLocationOnMap struct {
	Transition     geometry.Point
	TargetMap      string
	TargetLocation geometry.Point
	Running        bool
}

func (a ActorGoalMoveToLocationOnMap) Description() string {
	if a.Running {
		return "running"
	}
	return "walking"
}

func (a ActorGoalMoveToLocationOnMap) IsCombatGoal() bool {
	return false
}

func (a ActorGoalMoveToLocationOnMap) Action(g *GameState, actor *Actor) (fsmai.TransitionEvent, int) {
	moveFunc := walkTowards
	if a.Running {
		moveFunc = runTowards
	}
	if g.currentMapName != a.TargetMap {
		if actor.Position() == a.Transition {
			actor.SetFlag(foundation.FlagWantsToTransition)
			return fsmai.NoEvent, 10
		} else {
			return moveFunc(g, actor, a.Transition)
		}
	} else {
		return moveFunc(g, actor, a.TargetLocation)
	}
}

func (a ActorGoalMoveToLocationOnMap) IsHostilityTowards(attacker *Actor) bool {
	return false
}

func (a ActorGoalMoveToLocationOnMap) Achieved(g *GameState, actor *Actor) bool {
	return g.currentMapName == a.TargetMap && actor.Position() == a.TargetLocation
}

func GoalWalkToLocation(loc geometry.Point) ActorGoalMoveToLocation {
	return ActorGoalMoveToLocation{
		Location: loc,
		Running:  false,
	}
}
func GoalRunToLocation(loc geometry.Point) ActorGoalMoveToLocation {
	return ActorGoalMoveToLocation{
		Location: loc,
		Running:  true,
	}
}

type ActorGoalMoveToLocation struct {
	Location geometry.Point
	Running  bool
}

func (g ActorGoalMoveToLocation) Description() string {
	if g.Running {
		return "running"
	}
	return "walking"
}

func (g ActorGoalMoveToLocation) IsCombatGoal() bool {
	return false
}

func (g ActorGoalMoveToLocation) Action(ga *GameState, a *Actor) (fsmai.TransitionEvent, int) {
	if g.Running {
		return runTowards(ga, a, g.Location)
	}
	return walkTowards(ga, a, g.Location)
}

func (g ActorGoalMoveToLocation) IsHostilityTowards(victim *Actor) bool {
	return false
}

func (g ActorGoalMoveToLocation) Achieved(ga *GameState, a *Actor) bool {
	return a.Position() == g.Location
}

func GoalSleepAt(loc geometry.Point) ActorGoalSleepAt {
	return ActorGoalSleepAt{
		Location: loc,
	}
}

type ActorGoalSleepAt struct {
	Location geometry.Point
}

func (g ActorGoalSleepAt) Description() string {
	return "going to bed"
}

func (g ActorGoalSleepAt) IsCombatGoal() bool {
	return false
}

func (g ActorGoalSleepAt) Action(ga *GameState, a *Actor) (fsmai.TransitionEvent, int) {
	if a.Position() == g.Location {
		a.SetSleeping()
		return fsmai.NoEvent, 10
	}
	return walkTowards(ga, a, g.Location)
}

func (g ActorGoalSleepAt) IsHostilityTowards(victim *Actor) bool {
	return false
}

func (g ActorGoalSleepAt) Achieved(ga *GameState, a *Actor) bool {
	return a.Position() == g.Location && a.IsSleeping()
}

func init() {
	gob.Register(ActorGoalMoveToSpawn{})
	gob.Register(ActorGoalMoveIntoShootingRange{})
	gob.Register(ActorGoalFleeFromActor{})
	gob.Register(ActorGoalKillActor{})
	gob.Register(ActorGoalMoveToLocation{})
	gob.Register(ActorGoalSleepAt{})
}
