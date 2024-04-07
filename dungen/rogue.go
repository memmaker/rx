package dungen

import (
	"RogueUI/geometry"
	"RogueUI/util"
	"math"
	"math/rand"
)

type RogueGenerator struct {
	mapWidth, mapHeight       int
	maxRoomSize               geometry.Point
	maxRooms                  int
	random                    *rand.Rand
	minRoomSize               geometry.Point
	gridDivisions             geometry.Point
	roomLitChance             float64
	additionalRoomConnections int
}

func NewRogueGenerator(random *rand.Rand, mapCols, mapRows int) *RogueGenerator {
	gridDivision := geometry.Point{X: 3, Y: 3}
	r := &RogueGenerator{
		random:        random,
		mapWidth:      mapCols,
		mapHeight:     mapRows,
		roomLitChance: 0.5,
	}
	r.SetGridDivision(gridDivision)
	return r
}

func (r *RogueGenerator) SetGridDivision(dim geometry.Point) {
	r.gridDivisions = dim
	r.maxRooms = dim.X * dim.Y
	maxWidth := int(math.Floor(float64(r.mapWidth-2-(dim.X-1)*3) / float64(dim.X)))
	maxHeight := int(math.Floor(float64(r.mapHeight-2-(dim.Y-1)*3) / float64(dim.Y)))
	r.minRoomSize = geometry.Point{X: 3, Y: 3}
	r.maxRoomSize = geometry.Point{X: max(4, maxWidth), Y: max(4, maxHeight)}
}

func (r *RogueGenerator) SetRoomLitChance(chance float64) {
	r.roomLitChance = chance
}
func (r *RogueGenerator) Generate() *DungeonMap {
	emptyMap := NewDungeonMap(r.mapWidth, r.mapHeight)
	r.placeRogueRooms(emptyMap)
	connectInfos := r.connectRooms(emptyMap)
	emptyMap.SetFirstAndLastRooms(connectInfos.firstRoom, connectInfos.lastRoom)
	r.createCorridors(emptyMap, connectInfos)
	r.placeStairs(r.random, emptyMap)

	return emptyMap
}

func (r *RogueGenerator) placeRogueRooms(emptyMap *DungeonMap) {
	for i := 0; i < r.maxRooms; i++ {
		roomSize := geometry.Point{
			X: r.random.Intn(r.maxRoomSize.X-r.minRoomSize.X) + 1 + r.minRoomSize.X,
			Y: r.random.Intn(r.maxRoomSize.Y-r.minRoomSize.Y) + 1 + r.minRoomSize.Y,
		}

		spaceLeft := r.maxRoomSize.Sub(roomSize)
		topLeftOfRoom := geometry.Point{X: 0, Y: 0}
		if spaceLeft.X > 0 {
			topLeftOfRoom.X = r.random.Intn(spaceLeft.X)
		}
		if spaceLeft.Y > 0 {
			topLeftOfRoom.Y = r.random.Intn(spaceLeft.Y)
		}

		xGrid, yGrid := util.IndexToXY(i, r.gridDivisions.X)
		xOff := (xGrid * (r.maxRoomSize.X + 3)) + 1
		yOff := (yGrid * (r.maxRoomSize.Y + 3)) + 1
		newRoom := NewDungeonRoomFromRect(r.random, geometry.NewRect(topLeftOfRoom.X + +xOff, topLeftOfRoom.Y+yOff, topLeftOfRoom.X+xOff+roomSize.X, topLeftOfRoom.Y+yOff+roomSize.Y))
		emptyMap.AddRoomAndSetTiles(newRoom)
		if r.random.Float64() < r.roomLitChance {
			newRoom.SetLit(true)
		}
	}
}

type ConnectionInfo struct {
	edges     [][2]int
	firstRoom int
	lastRoom  int
}

func (r *RogueGenerator) connectRooms(emptyMap *DungeonMap) ConnectionInfo {
	connectedRooms := make(map[int]bool)
	getRandomConnectedRoom := func() int {
		randomIndex := r.random.Intn(len(connectedRooms))
		counter := 0
		for roomIndex := range connectedRooms {
			if counter == randomIndex {
				return roomIndex
			}
			counter++
		}
		return -1
	}
	connections := make([][2]int, 0)

	addConnection := func(room1, room2 int) {
		connections = append(connections, [2]int{room1, room2})
	}

	roomCount := len(emptyMap.AllRooms())

	getAdjacentRoomIndices := func(room int) []int {
		x, y := util.IndexToXY(room, r.gridDivisions.X)
		var adjacentRooms []int
		if x > 0 {
			adjacentRooms = append(adjacentRooms, room-1)
		}
		if x < r.gridDivisions.X-1 {
			adjacentRooms = append(adjacentRooms, room+1)
		}
		if y > 0 {
			adjacentRooms = append(adjacentRooms, room-r.gridDivisions.X)
		}
		if y < r.gridDivisions.Y-1 {
			adjacentRooms = append(adjacentRooms, room+r.gridDivisions.X)
		}
		return adjacentRooms
	}
	startRoom := r.random.Intn(roomCount)
	connectedRooms[startRoom] = true
	var lastRoom int
	for len(connectedRooms) < roomCount {
		connectedRoom := getRandomConnectedRoom()
		adjacentRooms := getAdjacentRoomIndices(connectedRoom)
		for _, adjacentRoom := range adjacentRooms {
			if _, isConn := connectedRooms[adjacentRoom]; isConn {
				continue
			}
			addConnection(connectedRoom, adjacentRoom)
			connectedRooms[adjacentRoom] = true
			lastRoom = adjacentRoom
		}
	}

	for i := 0; i < r.additionalRoomConnections; i++ {
		room1 := r.random.Intn(roomCount)
		possibleConnections := getAdjacentRoomIndices(room1)
		randomIndices := r.random.Perm(len(possibleConnections))

	outer:
		for _, index := range randomIndices {
			room2 := possibleConnections[index]
			for _, edge := range connections {
				if (edge[0] == room1 && edge[1] == room2) || (edge[0] == room2 && edge[1] == room1) {
					continue outer
				}
			}

			addConnection(room1, room2)
		}
	}

	return ConnectionInfo{
		edges:     connections,
		firstRoom: startRoom,
		lastRoom:  lastRoom,
	}
}

func (r *RogueGenerator) createCorridors(emptyMap *DungeonMap, infos ConnectionInfo) {
	allRooms := emptyMap.AllRooms()
	for _, edge := range infos.edges {
		r1Index := edge[0]
		r2Index := edge[1]
		room1 := allRooms[r1Index]
		room2 := allRooms[r2Index]
		if geometry.Abs(r1Index-r2Index) == 1 { // horizontal
			left := room1
			right := room2
			if r1Index > r2Index {
				left = room2
				right = room1
			}
			r.createHorizontalCorridorBetweenRooms(emptyMap, left, right)
		} else { // vertical
			top := room1
			bottom := room2
			if r1Index > r2Index {
				top = room2
				bottom = room1
			}
			r.createVerticalCorridorBetweenRooms(emptyMap, top, bottom)
		}
	}
}

func (r *RogueGenerator) createHorizontalCorridorBetweenRooms(emptyMap *DungeonMap, left *DungeonRoom, right *DungeonRoom) {
	center1 := left.GetRandomConnectorPosition(r.random, geometry.East)
	center2 := right.GetRandomConnectorPosition(r.random, geometry.West)
	smallestX := min(center1.X, center2.X)
	largestX := max(center1.X, center2.X)
	smallestY := min(center1.Y, center2.Y)
	largestY := max(center1.Y, center2.Y)

	meetAtX := (smallestX + largestX) / 2
	if largestX-smallestX > 3 {
		meetAtX = r.random.Intn(largestX-smallestX-2) + smallestX + 1
	}

	rooms := []*DungeonRoom{left, right}

	for x := smallestX; x <= largestX; x++ {
		if x <= meetAtX {
			emptyMap.SetCorridorSelective(x, center1.Y, rooms)
		} else if x > meetAtX {
			emptyMap.SetCorridorSelective(x, center2.Y, rooms)
		}
	}
	for y := smallestY; y <= largestY; y++ {
		emptyMap.SetCorridorSelective(meetAtX, y, rooms)
	}
}

func (r *RogueGenerator) createVerticalCorridorBetweenRooms(emptyMap *DungeonMap, top *DungeonRoom, bottom *DungeonRoom) {
	center1 := top.GetRandomConnectorPosition(r.random, geometry.South)
	center2 := bottom.GetRandomConnectorPosition(r.random, geometry.North)
	smallestX := min(center1.X, center2.X)
	largestX := max(center1.X, center2.X)
	smallestY := min(center1.Y, center2.Y)
	largestY := max(center1.Y, center2.Y)

	meetAtY := (smallestY + largestY) / 2
	if largestY-smallestY > 3 {
		meetAtY = r.random.Intn(largestY-smallestY-2) + smallestY + 1
	}

	rooms := []*DungeonRoom{top, bottom}

	for y := smallestY; y <= largestY; y++ {
		if y <= meetAtY {
			emptyMap.SetCorridorSelective(center1.X, y, rooms)
		} else if y > meetAtY {
			emptyMap.SetCorridorSelective(center2.X, y, rooms)
		}
	}

	for x := smallestX; x <= largestX; x++ {
		emptyMap.SetCorridorSelective(x, meetAtY, rooms)
	}
}

func (r *RogueGenerator) placeStairs(random *rand.Rand, emptyMap *DungeonMap) {
	firstRoom, lastRoom := emptyMap.GetFirstAndLastRoom()
	stairsUp := firstRoom.GetRandomAbsoluteFloorPosition(random)
	stairsDown := lastRoom.GetRandomAbsoluteFloorPosition(random)
	emptyMap.SetStairsUp(stairsUp)
	emptyMap.SetStairsDown(stairsDown)
}

func (r *RogueGenerator) SetAdditionalRoomConnections(count int) {
	r.additionalRoomConnections = count
}
