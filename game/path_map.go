package game

import (
	"contractor/foundation"
	"contractor/gridmap"
	"github.com/memmaker/go/geometry"
	"slices"
)

type MapPosition struct {
	MapName  string
	Position geometry.Point
	// LocationName is optional and may be empty, will only be set if the position is a named location
	LocationName string
}

type Pathfinder struct {
	getMap         func(string) *gridmap.GridMap[*Actor, foundation.Item, Object]
	neighborsCache map[*Actor]map[MapPosition][]MapPosition
}

func NewPathfinder(getMap func(string) *gridmap.GridMap[*Actor, foundation.Item, Object]) *Pathfinder {
	return &Pathfinder{
		getMap:         getMap,
		neighborsCache: make(map[*Actor]map[MapPosition][]MapPosition),
	}
}

func (p *Pathfinder) getNeighbors(actor *Actor, from MapPosition, to MapPosition) []MapPosition {
	neighbors, cacheExists := p.neighborsCache[actor][from]

	if !cacheExists {
		gMap := p.getMap(from.MapName)
		location := gMap.GetNamedLocation(from.LocationName)

		if transitionTo, exists := gMap.GetTransitionAt(location); exists {
			targetMap := p.getMap(transitionTo.TargetMap)
			targetLocation := targetMap.GetNamedLocation(transitionTo.TargetLocation)
			neighbors = append(neighbors, MapPosition{MapName: transitionTo.TargetMap, LocationName: transitionTo.TargetLocation, Position: targetLocation})
		}

		for locationOfTransition, _ := range gMap.Transitions() {
			if locationOfTransition == location {
				continue
			}
			pathToOtherTransition := gMap.GetJPSPath(location, locationOfTransition, func(point geometry.Point) bool {
				return gMap.IsWalkableFor(point, actor)
			})

			if len(pathToOtherTransition) > 0 {
				nameOfLocation := gMap.GetNamedLocationByPos(locationOfTransition)
				neighbors = append(neighbors, MapPosition{MapName: from.MapName, LocationName: nameOfLocation, Position: locationOfTransition})
			}
		}

		if to != from && to.MapName == from.MapName {
			pathToTarget := gMap.GetJPSPath(location, to.Position, func(point geometry.Point) bool {
				return gMap.IsWalkableFor(point, actor)
			})
			if len(pathToTarget) > 0 {
				neighbors = append(neighbors, to)
			}
		}

		if p.neighborsCache[actor] == nil {
			p.neighborsCache[actor] = make(map[MapPosition][]MapPosition)
		}
		p.neighborsCache[actor][from] = neighbors
	}

	return neighbors
}

// FindPath returns a path of transitions from the current map to the target map.
// It expects the actor to have a DijkstraMap that has been updated with the current map.
func (p *Pathfinder) FindPath(actor *Actor, currentMap string, to MapPosition) []MapPosition {
	p.neighborsCache[actor] = make(map[MapPosition][]MapPosition)

	reachable := actor.GetDijkstraMap()
	gMap := p.getMap(currentMap)
	reachableTransitions := make(map[MapPosition]int)
	openTransitions := make(map[MapPosition]int)
	for pos, _ := range reachable {
		if _, exists := gMap.GetTransitionAt(pos); exists {

			locationName := gMap.GetNamedLocationByPos(pos)

			transition := MapPosition{MapName: currentMap, LocationName: locationName, Position: pos}

			openTransitions[transition] = 1
			reachableTransitions[transition] = 1
		}
	}
	closedTransitions := make(map[MapPosition]int)

	for {
		if len(openTransitions) == 0 {
			break
		}
		for transition, dist := range openTransitions {
			closedTransitions[transition] = dist
			delete(openTransitions, transition)

			neighbors := p.getNeighbors(actor, transition, to)
			for _, neighbor := range neighbors {
				if _, exists := closedTransitions[neighbor]; exists {
					continue
				}
				if _, exists := openTransitions[neighbor]; exists {
					continue
				}
				openTransitions[neighbor] = dist + 1
			}
		}
	}

	if _, exists := closedTransitions[to]; !exists {
		return nil
	}

	path := make([]MapPosition, 0)

	current := to
	_, isReachable := reachableTransitions[current]
	path = append(path, current)
	for !isReachable {
		minDist := closedTransitions[current]

		for _, neighbor := range p.getNeighbors(actor, current, MapPosition{}) {
			if dist, exists := closedTransitions[neighbor]; exists {
				if dist < minDist {
					minDist = dist
					current = neighbor
				}
			}
		}

		isReachable = reachableTransitions[current] == 1
		if current.MapName != path[len(path)-1].MapName {
			path = append(path, current)
		}
	}

	slices.Reverse(path)

	return path
}
