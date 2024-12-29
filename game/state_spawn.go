package game

import (
	"contractor/foundation"
	"contractor/fsmai"
	"contractor/gridmap"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

func (g *GameState) TestTeamSpawn() {
	teamName := "ebi_security"
	aggressor := g.Player
	victim := g.actorWithName("daniel_harker")
	g.SpawnTeamForHelping(victim, aggressor, teamName)
}

func (g *GameState) SpawnTeamForHelping(victim *Actor, aggressor *Actor, teamName string) {
	// determine a spawn position for the team, that is both plausible and outside the player's vision range
	// make it the goal of the leader to attack the aggressor

	// we really want a dedicated search for threat behavior
	// it should handle moving to the last known location of the aggressor
	// and then searching for the aggressor
	// if the aggressor is not found, the team should return to their spawn position and despawn
	// if the aggressor is found, the team should attack the aggressor
	gMap := g.currentMap()
	for transPos, _ := range gMap.Transitions() {

		_, isReachable := aggressor.GetDijkstraMap()[transPos]

		if isReachable && !aggressor.CanSee(transPos) {
			locationName := gMap.GetNamedLocationByPos(transPos)
			leader, _ := g.SpawnTeam(teamName, MapPosition{MapName: gMap.GetName(), LocationName: locationName, Position: transPos})
			if leader != nil {
				leader.FSM.SetState(fsmai.StateSearch, ActorEvent{
					Event: fsmai.EventMajorCrimeWitnessed,
					Actor: aggressor,
				})
			}
			break
		}
	}
}

func (g *GameState) SpawnTeam(teamName string, teamPos MapPosition) (*Actor, []*Actor) {
	targetMap := g.ensureMapIsLoaded(teamPos.MapName)
	targetPos := teamPos.Position
	template, exists := g.globalTeamTemplates[teamName]
	if !exists {
		return nil, nil
	}
	var leader *Actor
	var members []*Actor
	for _, field := range template {
		switch strings.ToLower(field.Name) {
		case "leader":
			leader = g.NewActorFromName(field.Value, teamPos.MapName)
		case "member":
			members = append(members, g.NewActorFromName(field.Value, teamPos.MapName))
		}
	}
	if leader == nil {
		return nil, nil
	}

	targetMap.AddActorWithDisplacement(leader, targetPos)

	leader.SpawnPosition = leader.Position()

	for _, member := range members {
		targetMap.AddActorWithDisplacement(member, targetPos)
		member.SpawnPosition = member.Position()
		member.FSM.SetState(fsmai.StateFollow, NewLeaderJoinedEvent(leader))
	}

	return leader, members
}

func (g *GameState) NewObjectFromRecord(record recfile.Record, newMap *gridmap.GridMap[*Actor, foundation.Item, Object], iconResolver func(objType string) textiles.TextIcon) Object {
	objectType := record.FindValueForKeyIgnoreCase("category")

	switch strings.ToLower(objectType) {
	case "explodingpushbox":
		box := g.NewPushBox(record, iconResolver)
		box.SetExploding()
		return box
	case "pushbox":
		return g.NewPushBox(record, iconResolver)
	case "elevator":
		elevator := g.NewElevator(record, iconResolver)
		newMap.AddNamedLocation(elevator.GetIdentifier(), elevator.Position())
		return elevator
	case "unknowncontainer":
		return g.NewContainer(record, iconResolver)
	case "trap":
		return g.NewTrap(record, iconResolver)
	case "terminal":
		return g.NewTerminal(record, iconResolver)
	case "bed":
		return g.NewBed(record, iconResolver)
	case "readable":
		return g.NewReadable(record, iconResolver)
	case "lockeddoor":
		fallthrough
	case "closeddoor":
		fallthrough
	case "brokendoor":
		fallthrough
	case "opendoor":
		return g.NewDoor(record, iconResolver)
	}
	return nil
}

func (g *GameState) iconForItem(itemCategory foundation.ItemCategory) textiles.TextIcon {
	if icon, exists := g.iconsForItems[itemCategory]; exists {
		return icon
	}
	return textiles.TextIcon{}
}
func (g *GameState) addItemToMap(item foundation.Item, mapPos geometry.Point) {
	g.currentMap().AddItemWithDisplacement(item, mapPos)
}

func (g *GameState) NewGold(amount int) *GenericItem {
	icon := g.iconForItem(foundation.ItemCategoryGold)
	gold := &GenericItem{
		UID:          gridmap.NextItemID(),
		DisplayName:  "gold",
		InternalName: "gold",
		Category:     foundation.ItemCategoryGold,
		StackSize:    amount,
		Icon:         icon,
	}
	return gold
}
