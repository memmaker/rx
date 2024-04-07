package dungen

import (
	"RogueUI/geometry"
	"RogueUI/util"
	"cmp"
	"math"
	"math/rand"
	"slices"
)

type DungeonRoom struct {
	roomPositionOffset geometry.Point

	floorTiles map[geometry.Point]bool

	availableConnections    map[geometry.Point]RoomConnection
	connectedRooms          map[geometry.Point]*DungeonRoom
	rotationMatrix          [2][2]int
	inverseRotationMatrix   [2][2]int
	rotationOffset          geometry.CompassDirection
	usedTemplate            string
	roomId                  int
	wallTiles               []geometry.Point
	poiPositions            []geometry.Point
	usedDecorationLevelName string
	slotId                  int
	plugsIntoSlotId         int
	canBeLit                bool
	bounds                  geometry.Rect
	doors                   map[geometry.Point]bool
}

func (r *DungeonRoom) GetFreeConnectionsWithRotatedDirection() map[geometry.Point]RoomConnection {
	result := make(map[geometry.Point]RoomConnection)
	for _, point := range getSortedKeys(r.availableConnections) {
		roomConnection := r.availableConnections[point]
		if _, ok := r.connectedRooms[point]; !ok {
			result[point] = RoomConnection{
				OutwardDirection: roomConnection.OutwardDirection + r.rotationOffset,
				ConnectionID:     roomConnection.ConnectionID,
			}
		}
	}
	return result
}

func (r *DungeonRoom) HasMatchingConnection(otherConnection RoomConnection) (geometry.Point, bool) {
	neededDirection := otherConnection.OutwardDirection.Opposite()
	neededId := otherConnection.ConnectionID
	for _, pos := range getSortedKeys(r.availableConnections) {
		if _, alreadyConnected := r.connectedRooms[pos]; alreadyConnected {
			continue
		}
		roomConnection := r.availableConnections[pos]

		dir := roomConnection.OutwardDirection
		dir = dir + r.rotationOffset
		if dir >= 360 {
			dir = dir - 360
		}
		if dir == neededDirection && ((neededId == 0 && roomConnection.ConnectionID == 0) || roomConnection.ConnectionID == neededId) {
			return pos, true
		}
	}
	return geometry.Point{}, false
}

func (r *DungeonRoom) SetPositionOffset(offset geometry.Point) {
	r.roomPositionOffset = offset
}

func (r *DungeonRoom) GetAbsoluteDoorPosition(doorPos geometry.Point) geometry.Point {
	return r.ToAbsolutePosition(doorPos)
}

func (r *DungeonRoom) ToAbsolutePosition(relativePos geometry.Point) geometry.Point {
	rotMat := r.rotationMatrix
	rotatedPos := matrixMul(relativePos, rotMat)
	return r.roomPositionOffset.Add(rotatedPos)
}

func (r *DungeonRoom) ToRelativePosition(absolutePos geometry.Point) geometry.Point {
	translatedPos := absolutePos.Sub(r.roomPositionOffset)
	rotMat := r.inverseRotationMatrix
	unrotatedPos := matrixMul(translatedPos, rotMat)
	return unrotatedPos
}

func (r *DungeonRoom) GetAbsoluteFloorTiles() []geometry.Point {
	result := make([]geometry.Point, 0)
	for point, _ := range r.floorTiles {
		result = append(result, r.ToAbsolutePosition(point))
	}
	return result
}
func (r *DungeonRoom) GetRelativeFloorTiles() []geometry.Point {
	result := make([]geometry.Point, 0)
	for point, _ := range r.floorTiles {
		result = append(result, point)
	}
	//REMOVE THIS; debug needs only
	slices.SortStableFunc(result, func(i, j geometry.Point) int {
		idxI := util.XYToIndex(i.X, i.Y, 20)
		idxJ := util.XYToIndex(j.X, j.Y, 20)
		return cmp.Compare(idxI, idxJ)
	})
	return result
}
func (r *DungeonRoom) GetAbsoluteRoomTiles() []geometry.Point {
	result := make([]geometry.Point, 0)
	for point, _ := range r.floorTiles {
		result = append(result, r.ToAbsolutePosition(point))
	}
	for _, point := range r.wallTiles {
		result = append(result, r.ToAbsolutePosition(point))
	}
	for point, _ := range r.doors {
		result = append(result, r.ToAbsolutePosition(point))
	}
	return result
}

func (r *DungeonRoom) GetPOIPositions() []geometry.Point {
	result := make([]geometry.Point, 0)
	for _, point := range r.poiPositions {
		result = append(result, r.ToAbsolutePosition(point))
	}
	return result
}
func (r *DungeonRoom) AddConnectedRoom(doorPos geometry.Point, room *DungeonRoom) {
	r.connectedRooms[doorPos] = room
}

func (r *DungeonRoom) GetUsedAbsoluteDoorTiles() []geometry.Point {
	result := make([]geometry.Point, 0)
	for point, _ := range r.availableConnections {
		if _, ok := r.connectedRooms[point]; !ok {
			continue
		}
		result = append(result, r.ToAbsolutePosition(point))
	}
	return result
}

/*
func (r *DungeonRoom) AddCorridor(direction geometry.CompassDirection, length int) {
    // remove door tile from available door tiles
    doorPos, ok := r.HasFreeRelativeDoorInDirection(direction)
    if !ok {
        return
    }
    delete(r.availableConnections, doorPos)

    // add corridor tiles
    for i := 0; i < length; i++ {
        r.floorTiles[doorPos.Add(direction.ToPoint().Mul(i))] = true
    }

    // add door tile at the end of the corridor
    newDoorPos := doorPos.Add(direction.ToPoint().Mul(length))
    r.availableConnections[newDoorPos] = direction

    if length > 1 {
        // add another set of doors at 90 degrees to the corridor
        startPos := doorPos.Add(direction.ToPoint().Mul(length - 1))

        secondDoor := startPos.Add(direction.TurnRightBy90().ToPoint())
        r.availableConnections[secondDoor] = direction.TurnRightBy90()

        thirdDoor := startPos.Add(direction.TurnLeftBy90().ToPoint())
        r.availableConnections[thirdDoor] = direction.TurnLeftBy90()
    }
}

*/

func (r *DungeonRoom) Contains(absolutePosition geometry.Point) bool {
	relative := r.ToRelativePosition(absolutePosition)
	if _, ok := r.floorTiles[relative]; ok {
		return true
	}
	if _, ok := r.doors[relative]; ok {
		return true
	}
	return false
}

func (r *DungeonRoom) ContainsIncludingWalls(absolutePosition geometry.Point) bool {
	relative := r.ToRelativePosition(absolutePosition)
	if _, ok := r.floorTiles[relative]; ok {
		return true
	}
	if _, ok := r.doors[relative]; ok {
		return true
	}
	for _, wallPos := range r.wallTiles {
		if wallPos == relative {
			return true
		}
	}
	return false
}

func (r *DungeonRoom) FloorContains(absolutePosition geometry.Point) bool {
	relative := r.ToRelativePosition(absolutePosition)
	if _, ok := r.floorTiles[relative]; ok {
		return true
	}
	return false
}

func (r *DungeonRoom) SetRotationCount(count int) {
	switch count {
	case 0:
		r.rotationMatrix = [2][2]int{
			{1, 0},
			{0, 1},
		}
		r.inverseRotationMatrix = [2][2]int{
			{1, 0},
			{0, 1},
		}
		r.rotationOffset = 0
	case 1:
		r.rotationMatrix = [2][2]int{
			{0, -1},
			{1, 0},
		}
		r.inverseRotationMatrix = [2][2]int{
			{0, 1},
			{-1, 0},
		}
		r.rotationOffset = 90
	case 2:
		r.rotationMatrix = [2][2]int{
			{-1, 0},
			{0, -1},
		}
		r.inverseRotationMatrix = [2][2]int{
			{-1, 0},
			{0, -1},
		}
		r.rotationOffset = 180
	case 3:
		r.rotationMatrix = [2][2]int{
			{0, 1},
			{-1, 0},
		}
		r.inverseRotationMatrix = [2][2]int{
			{0, -1},
			{1, 0},
		}
		r.rotationOffset = 270
	}
}

func (r *DungeonRoom) findCenter() geometry.Point {
	minX := math.MaxInt
	minY := math.MaxInt
	maxX := -math.MaxInt
	maxY := -math.MaxInt
	for point, _ := range r.floorTiles {
		if point.X < minX {
			minX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	return geometry.Point{
		X: (minX + maxX) / 2,
		Y: (minY + maxY) / 2,
	}
}

func NewDungeonRoomFromRect(random *rand.Rand, bounds geometry.Rect) *DungeonRoom {
	rectRoom := &DungeonRoom{
		floorTiles:     roomTilesFromRect(bounds),
		wallTiles:      wallTilesFromRect(bounds),
		bounds:         bounds,
		doors:          make(map[geometry.Point]bool),
		connectedRooms: make(map[geometry.Point]*DungeonRoom),
		availableConnections: map[geometry.Point]RoomConnection{
			bounds.GetRandomPointOnEdge(random, geometry.North).Add(geometry.North.ToPoint()): {
				OutwardDirection: geometry.North,
				ConnectionID:     0,
			},
			bounds.GetRandomPointOnEdge(random, geometry.South).Add(geometry.South.ToPoint()): {
				OutwardDirection: geometry.South,
				ConnectionID:     0,
			},
			bounds.GetRandomPointOnEdge(random, geometry.East).Add(geometry.East.ToPoint()): {
				OutwardDirection: geometry.East,
				ConnectionID:     0,
			},
			bounds.GetRandomPointOnEdge(random, geometry.West).Add(geometry.West.ToPoint()): {
				OutwardDirection: geometry.West,
				ConnectionID:     0,
			},
		},
	}
	// choose a random direction or none

	/*
	   switch random.Intn(5) {
	   case 0:
	       rectRoom.AddCorridor(geometry.North, random.Intn(4)+1)
	   case 1:
	       rectRoom.AddCorridor(geometry.South, random.Intn(4)+1)
	   case 2:
	       rectRoom.AddCorridor(geometry.East, random.Intn(4)+1)
	   case 3:
	       rectRoom.AddCorridor(geometry.West, random.Intn(4)+1)
	   }

	*/
	rectRoom.SetRotationCount(0)
	return rectRoom
}

func (r *DungeonRoom) SetFloorTiles(tiles []geometry.Point) {
	r.floorTiles = make(map[geometry.Point]bool)
	for _, tile := range tiles {
		r.floorTiles[tile] = true
	}
}

func (r *DungeonRoom) SetConnections(tiles map[geometry.Point]RoomConnection) {
	r.availableConnections = tiles
}

func (r *DungeonRoom) Clone() *DungeonRoom {
	newRoom := &DungeonRoom{
		roomPositionOffset:      r.roomPositionOffset,
		connectedRooms:          make(map[geometry.Point]*DungeonRoom),
		doors:                   r.doors,
		rotationMatrix:          r.rotationMatrix,
		inverseRotationMatrix:   r.inverseRotationMatrix,
		rotationOffset:          r.rotationOffset,
		usedTemplate:            r.usedTemplate,
		usedDecorationLevelName: r.usedDecorationLevelName,
		wallTiles:               make([]geometry.Point, 0),
		floorTiles:              make(map[geometry.Point]bool),
		availableConnections:    make(map[geometry.Point]RoomConnection),
		poiPositions:            make([]geometry.Point, 0),
		slotId:                  r.slotId,
		plugsIntoSlotId:         r.plugsIntoSlotId,
		canBeLit:                r.canBeLit,
	}
	for point, _ := range r.floorTiles {
		newRoom.floorTiles[point] = true
	}
	for _, point := range r.wallTiles {
		newRoom.wallTiles = append(newRoom.wallTiles, point)
	}
	for point, direction := range r.availableConnections {
		newRoom.availableConnections[point] = direction
	}
	for _, point := range r.poiPositions {
		newRoom.poiPositions = append(newRoom.poiPositions, point)
	}
	return newRoom
}

func (r *DungeonRoom) PrintUntransformed() {
	bounds := r.GetRelativeBoundingRectIncludingConnectors()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			relativeRoomPos := geometry.Point{X: x, Y: y}
			if _, ok := r.floorTiles[relativeRoomPos]; ok {
				print(".")
			} else if _, ok = r.availableConnections[relativeRoomPos]; ok {
				print("+")
			} else {
				print(" ")
			}
		}
		println()
	}
}

func (r *DungeonRoom) GetRelativeBoundingRectIncludingConnectors() geometry.Rect {
	minX := math.MaxInt
	minY := math.MaxInt
	maxX := -math.MaxInt
	maxY := -math.MaxInt
	for point, _ := range r.floorTiles {
		if point.X < minX {
			minX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	for point, _ := range r.availableConnections {
		if point.X < minX {
			minX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	return geometry.NewRect(minX, minY, maxX+1, maxY+1)
}

func (r *DungeonRoom) PrintRotated() {
	bounds := r.GetRelativeBoundingRectIncludingConnectors()
	transformedBoundsMin := matrixMul(bounds.Min, r.rotationMatrix)
	transformedBoundsMax := matrixMul(bounds.Max, r.rotationMatrix)
	transformedBounds := geometry.NewRect(transformedBoundsMin.X, transformedBoundsMin.Y, transformedBoundsMax.X, transformedBoundsMax.Y)
	transformedBounds = transformedBounds.Shift(0, 0, 1, 1)
	zero := geometry.Point{X: 0, Y: 0}
	for y := transformedBounds.Min.Y; y < transformedBounds.Max.Y; y++ {
		for x := transformedBounds.Min.X; x < transformedBounds.Max.X; x++ {
			absTransformedPos := geometry.Point{X: x, Y: y}
			relativePos := matrixMul(absTransformedPos, r.inverseRotationMatrix)
			if relativePos == zero {
				print("0")
			} else if _, ok := r.floorTiles[relativePos]; ok {
				print(".")
			} else if _, ok = r.availableConnections[relativePos]; ok {
				print("+")
			} else if _, ok = r.connectedRooms[relativePos]; ok {
				print("=")
			} else {
				print(" ")
			}
		}
		println()
	}
}
func (r *DungeonRoom) PrintTransformed() {
	bounds := r.GetRelativeBoundingRectIncludingConnectors()
	transformedBoundsMin := r.ToAbsolutePosition(bounds.Min)
	transformedBoundsMax := r.ToAbsolutePosition(bounds.Max)
	transformedBounds := geometry.NewRect(transformedBoundsMin.X, transformedBoundsMin.Y, transformedBoundsMax.X, transformedBoundsMax.Y)
	transformedBounds = transformedBounds.Shift(0, 0, 1, 1)
	zero := geometry.Point{X: 0, Y: 0}
	for y := transformedBounds.Min.Y; y < transformedBounds.Max.Y; y++ {
		xStart := min(0, transformedBounds.Min.X)
		for x := xStart; x < transformedBounds.Max.X; x++ {
			absolutePos := geometry.Point{X: x, Y: y}
			relativePos := r.ToRelativePosition(absolutePos)
			if absolutePos == zero {
				print("0")
			} else if relativePos == zero {
				print("Q")
			} else if _, ok := r.floorTiles[relativePos]; ok {
				print(".")
			} else if _, ok = r.availableConnections[relativePos]; ok {
				print("+")
			} else if _, ok = r.connectedRooms[relativePos]; ok {
				print("=")
			} else {
				print(" ")
			}
		}
		println()
	}
}

func (r *DungeonRoom) PrintTransformedWithHighlight(relativePositionForHighlight geometry.Point) {
	bounds := r.GetRelativeBoundingRectIncludingConnectors()
	transformedBoundsMin := r.ToAbsolutePosition(bounds.Min)
	transformedBoundsMax := r.ToAbsolutePosition(bounds.Max)
	transformedBounds := geometry.NewRect(transformedBoundsMin.X, transformedBoundsMin.Y, transformedBoundsMax.X, transformedBoundsMax.Y)
	transformedBounds = transformedBounds.Shift(0, 0, 1, 1)
	zero := geometry.Point{X: 0, Y: 0}
	for y := transformedBounds.Min.Y; y < transformedBounds.Max.Y; y++ {
		xStart := min(0, transformedBounds.Min.X)
		for x := xStart; x < transformedBounds.Max.X; x++ {
			absolutePos := geometry.Point{X: x, Y: y}
			relativePos := r.ToRelativePosition(absolutePos)
			if relativePos == relativePositionForHighlight {
				print("X")
			} else if absolutePos == zero {
				print("0")
			} else if relativePos == zero {
				print("Q")
			} else if _, ok := r.floorTiles[relativePos]; ok {
				print(".")
			} else if _, ok = r.availableConnections[relativePos]; ok {
				print("+")
			} else if _, ok = r.connectedRooms[relativePos]; ok {
				print("=")
			} else {
				print(" ")
			}
		}
		println()
	}
}

func (r *DungeonRoom) SetId(id int) {
	r.roomId = id
}

func (r *DungeonRoom) GetRotatedRelativeDoorPosition(pos geometry.Point) geometry.Point {
	return matrixMul(pos, r.rotationMatrix)
}

func (r *DungeonRoom) GetPositionOffset() geometry.Point {
	return r.roomPositionOffset
}

func (r *DungeonRoom) SetIncludedWallTiles(tiles []geometry.Point) {
	r.wallTiles = tiles
}

func (r *DungeonRoom) SetPOIPositions(points []geometry.Point) {
	r.poiPositions = points
}

func (r *DungeonRoom) HasDecoratedTemplate() bool {
	return r.usedTemplate != "" && r.usedDecorationLevelName != ""
}

func (r *DungeonRoom) GetTemplateName() string {
	return r.usedTemplate
}

func (r *DungeonRoom) SetDecorationLevelName(levelName string) {
	r.usedDecorationLevelName = levelName
}

func (r *DungeonRoom) GetDecorationLevelName() string {
	return r.usedDecorationLevelName
}

func (r *DungeonRoom) HasOnlyZeroIDConnections() bool {
	for _, connection := range r.availableConnections {
		if connection.ConnectionID != 0 {
			return false
		}
	}
	return true
}

func (r *DungeonRoom) SetSlotId(id int) {
	r.slotId = id
}
func (r *DungeonRoom) GetSlotId() int {
	return r.slotId
}
func (r *DungeonRoom) SetPlugsIntoSlotId(plugsIn int) {
	r.plugsIntoSlotId = plugsIn
}
func (r *DungeonRoom) GetPlugsIntoSlotId() int {
	return r.plugsIntoSlotId
}

func (r *DungeonRoom) GetFloorTileCount() int {
	return len(r.floorTiles)
}

func (r *DungeonRoom) SetLit(lit bool) {
	r.canBeLit = lit
}

func (r *DungeonRoom) IsLit() bool {
	return r.canBeLit
}

func (r *DungeonRoom) IsCornerPosition(posOne geometry.Point) bool {
	shifted := posOne.Sub(r.roomPositionOffset)
	if r.bounds.IsOnCorner(shifted) {
		return true
	}
	return false
}

func (r *DungeonRoom) IsConnectorPosition(pos geometry.Point) bool {
	shifted := pos.Sub(r.roomPositionOffset)
	if _, ok := r.availableConnections[shifted]; ok {
		return true
	}
	return false
}

func (r *DungeonRoom) GetRandomConnectorPosition(random *rand.Rand, direction geometry.CompassDirection) geometry.Point {
	return r.bounds.GetRandomPointOnEdge(random, direction).Add(direction.ToPoint())
}

func (r *DungeonRoom) GetRandomAbsoluteFloorPosition(random *rand.Rand) geometry.Point {
	allFloorTiles := r.GetAbsoluteFloorTiles()
	return allFloorTiles[random.Intn(len(allFloorTiles))]
}
func (r *DungeonRoom) GetRandomAbsoluteFloorPositionWithFilter(random *rand.Rand, isFree func(geometry.Point) bool) (geometry.Point, bool) {
	allFloorTiles := r.GetAbsoluteFloorTiles()
	// randomize the order
	randomIndices := random.Perm(len(allFloorTiles))
	for _, idx := range randomIndices {
		if isFree(allFloorTiles[idx]) {
			return allFloorTiles[idx], true
		}
	}
	return geometry.Point{}, false
}

func (r *DungeonRoom) HasWallTile(pos geometry.Point) bool {
	for _, wallPos := range r.wallTiles {
		if wallPos == pos {
			return true
		}
	}
	return false
}

func (r *DungeonRoom) SetDoor(pos geometry.Point) {
	r.doors[pos] = true
	for i := len(r.wallTiles) - 1; i >= 0; i-- {
		wallPos := r.wallTiles[i]
		if wallPos == pos {
			r.wallTiles = append(r.wallTiles[:i], r.wallTiles[i+1:]...)
		}
	}
}

func (r *DungeonRoom) GetWalls() []geometry.Point {
	return r.wallTiles
}

func (r *DungeonRoom) IsTopLeftWallCorner(pos geometry.Point) bool {
	if pos.X == r.bounds.Min.X-1 && pos.Y == r.bounds.Min.Y-1 {
		return true
	}
	return false
}
func (r *DungeonRoom) IsTopRightWallCorner(pos geometry.Point) bool {
	if pos.X == r.bounds.Max.X && pos.Y == r.bounds.Min.Y-1 {
		return true
	}
	return false
}
func (r *DungeonRoom) IsBottomLeftWallCorner(pos geometry.Point) bool {
	if pos.X == r.bounds.Min.X-1 && pos.Y == r.bounds.Max.Y {
		return true
	}
	return false
}
func (r *DungeonRoom) IsBottomRightWallCorner(pos geometry.Point) bool {
	if pos.X == r.bounds.Max.X && pos.Y == r.bounds.Max.Y {
		return true
	}
	return false
}

func (r *DungeonRoom) GetCenter() geometry.Point {
	return r.bounds.Center()
}
func NewDungeonRoomFromTemplate(templateName string) *DungeonRoom {
	rectRoom := &DungeonRoom{
		usedTemplate:   templateName,
		doors:          make(map[geometry.Point]bool),
		connectedRooms: make(map[geometry.Point]*DungeonRoom),
	}
	rectRoom.SetRotationCount(0)
	return rectRoom
}

func roomTilesFromRect(bounds geometry.Rect) map[geometry.Point]bool {
	result := make(map[geometry.Point]bool)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			result[geometry.Point{X: x, Y: y}] = true
		}
	}
	return result
}

func wallTilesFromRect(bounds geometry.Rect) []geometry.Point {
	result := make([]geometry.Point, 0)
	minX := bounds.Min.X - 1
	maxX := bounds.Max.X
	minY := bounds.Min.Y - 1
	maxY := bounds.Max.Y
	for x := minX; x < maxX; x++ {
		result = append(result, geometry.Point{X: x, Y: minY})
		result = append(result, geometry.Point{X: x, Y: maxY})
	}
	for y := minY; y < maxY; y++ {
		result = append(result, geometry.Point{X: minX, Y: y})
		result = append(result, geometry.Point{X: maxX, Y: y})
	}
	result = append(result, geometry.Point{X: maxX, Y: maxY})
	return result
}
func matrixMul(point geometry.Point, b [2][2]int) geometry.Point {
	return geometry.Point{
		X: point.X*b[0][0] + point.Y*b[0][1],
		Y: point.X*b[1][0] + point.Y*b[1][1],
	}
}
