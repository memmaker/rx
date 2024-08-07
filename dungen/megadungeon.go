package dungen

import (
    "github.com/memmaker/go/fxtools"
    "github.com/memmaker/go/geometry"
    "math/rand"
)

// based on https://journal.stuffwithstuff.com/2014/12/21/rooms-and-mazes/
// https://github.com/munificent/hauberk/blob/db360d9efa714efb6d937c31953ef849c7394a39/lib/src/content/dungeon.dart
type Region interface {
    GetAbsoluteFloorTiles() []geometry.Point
}

type CorridorRegion map[geometry.Point]bool

func (c CorridorRegion) GetAbsoluteFloorTiles() []geometry.Point {
    var result []geometry.Point
    for pos := range c {
        result = append(result, pos)
    }
    return result
}

type MegaDungeonGenerator struct {
    randomSource           *rand.Rand
    roomTries              int
    regionLookup           map[geometry.Point]Region
    allRegions             []Region
    removeDeadEnds         bool
    minRoomSize            int
    straightPassageChance  float64
    imperfectConnectChance float64
    roomRatioInterval      float64
    allConnections         []geometry.Point
    doorDistance           int
    requiredConnections    []geometry.Point
    optionalConnections    []geometry.Point
}

func NewMegaDungeonGenerator(source *rand.Rand) *MegaDungeonGenerator {
    return &MegaDungeonGenerator{
        randomSource:           source,
        roomTries:              source.Intn(2500) + 16,
        minRoomSize:            source.Intn(6) + 2,
        straightPassageChance:  source.Float64(),
        removeDeadEnds:         true,
        imperfectConnectChance: source.Float64(),
        doorDistance:           source.Intn(6) + 2,
        roomRatioInterval:      source.Float64() * 0.5, // ==> between 1 and 1.2
        regionLookup:           make(map[geometry.Point]Region),
    }
}
func (c *MegaDungeonGenerator) SetRoomRatioInterval(interval float64) {
    c.roomRatioInterval = interval
}
func (c *MegaDungeonGenerator) SetRemoveDeadEnds(remove bool) {
    c.removeDeadEnds = remove

}
func (c *MegaDungeonGenerator) SetImperfectConnectChance(chance float64) {
    c.imperfectConnectChance = chance
}

func (c *MegaDungeonGenerator) SetDoorDistance(distance int) {
    c.doorDistance = distance

}
func (c *MegaDungeonGenerator) SetRoomTries(tries int) {
    c.roomTries = tries
}

func (c *MegaDungeonGenerator) SetMinRoomSize(size int) {
    c.minRoomSize = size
}

func (c *MegaDungeonGenerator) SetStraightPassageChance(chance float64) {
    c.straightPassageChance = chance
}

func (c *MegaDungeonGenerator) Generate(width, height int) *DungeonMap {
    c.regionLookup = make(map[geometry.Point]Region)
    c.allRegions = make([]Region, 0)
    c.allConnections = make([]geometry.Point, 0)

    if width%2 == 0 { // must be odd
        width--
    }
    if height%2 == 0 { // must be odd
        height--
    }
    emptyMap := NewDungeonMap(width, height)
    c.addRooms(emptyMap)
    c.addCorridors(emptyMap)

    c.connectRegions(emptyMap)
    c.placeDoorsAtConnections(emptyMap)
    //emptyMap.AddBorder()
    if c.removeDeadEnds {
        emptyMap.FillDeadEnds(c.randomSource)
    }
    //emptyMap.addMoreDoors(c.randomSource, 50, 0.3)

    return emptyMap
}

func (c *MegaDungeonGenerator) addRooms(emptyMap *DungeonMap) {
    for i := 0; i < c.roomTries; i++ {
        roomSize := makeOddForSize(c.randomSource, c.minRoomSize, c.randomSource.Intn(5)+c.minRoomSize)
        // we want a ration between 1 and 1.2
        aspectRatio := c.randomSource.Float64()*c.roomRatioInterval + 1
        roomWidth := roomSize
        roomHeight := roomSize
        if c.randomSource.Intn(2) == 0 {
            roomWidth = makeOddForSize(c.randomSource, c.minRoomSize, int(float64(roomWidth)*aspectRatio))
        } else {
            roomHeight = makeOddForSize(c.randomSource, c.minRoomSize, int(float64(roomHeight)*aspectRatio))
        }
        // pick a random spot
        x := c.randomSource.Intn(max(2, emptyMap.width-roomWidth-1)) + 1
        y := c.randomSource.Intn(max(2, emptyMap.height-roomHeight-1)) + 1
        x = makeOdd(c.randomSource, x)
        y = makeOdd(c.randomSource, y)

        emptyRoom := NewDungeonRoomFromRect(c.randomSource, geometry.NewRect(x, y, x+roomWidth, y+roomHeight))

        if !emptyMap.CanPlaceRoomRestrictive(emptyRoom) {
            continue
        }
        emptyMap.AddRoomAndSetTiles(emptyRoom)
        c.addRegion(emptyRoom)
    }
}

func (c *MegaDungeonGenerator) addCorridors(emptyMap *DungeonMap) {
    start := findEmptySpot(c.randomSource, emptyMap)
    for start != (geometry.Point{}) {
        region := fillMaze(emptyMap, start, c.straightPassageChance, c.randomSource)
        c.addRegion(region)
        start = findEmptySpot(c.randomSource, emptyMap)
    }
}

func (c *MegaDungeonGenerator) addRegion(region Region) {
    c.allRegions = append(c.allRegions, region)
    for _, pos := range region.GetAbsoluteFloorTiles() {
        c.regionLookup[pos] = region
    }
}

func (c *MegaDungeonGenerator) connectRegions(emptyMap *DungeonMap) {
    // don't we want a connector matrix table..?
    // map[Region]map[Region]geometry.Point

    var mainRegion Region
    mainRegion = emptyMap.GetRandomRoom(c.randomSource)
    // we want to get all the tiles that are adjacent walls and have a floor tile of another region on the other side

    for len(c.allRegions) > 1 {
        availableConnectors := c.getConnectors(emptyMap)
        connectorsToOther, exists := availableConnectors[mainRegion]
        if !exists {
            return
        }

        connectingRegion, connectorPos := c.chooseRandom(connectorsToOther)
        if connectingRegion == nil {
            println("ERR: no connecting region")
            return
        }
        c.requiredConnections = append(c.requiredConnections, connectorPos)
        c.allConnections = append(c.allConnections, connectorPos)
        addTries := 5
        if c.randomSource.Float64() < c.imperfectConnectChance && len(c.allRegions) > 2 {
            found, additionalConnectorPos := c.chooseRandomConnector(mainRegion, availableConnectors)
            for geometry.DistanceManhattan(additionalConnectorPos, connectorPos) < c.doorDistance && found && addTries > 0 {
                // we want to connect to another random region
                found, additionalConnectorPos = c.chooseRandomConnector(mainRegion, availableConnectors)
                addTries--
            }
            if found && c.hasMinDistToConnectors(c.allConnections, additionalConnectorPos, c.doorDistance) {
                c.allConnections = append(c.allConnections, additionalConnectorPos)
                c.optionalConnections = append(c.optionalConnections, additionalConnectorPos)
            }
        }

        mainRegion = c.mergeRegions(mainRegion, connectingRegion, connectorPos)
    }
}

func (c *MegaDungeonGenerator) chooseRandomConnector(regionOne Region, connectors map[Region]map[Region][]geometry.Point) (bool, geometry.Point) {
    allPos := make([]geometry.Point, 0)
    for _, conns := range connectors[regionOne] {
        for _, pos := range conns {
            if !c.hasMinDistToConnectors(c.allConnections, pos, c.doorDistance) {
                continue
            }
            allPos = append(allPos, conns...)
        }
    }
    if len(allPos) == 0 {
        return false, geometry.Point{}
    }
    randomIndex := rand.Intn(len(allPos))

    return true, allPos[randomIndex]
}

func (c *MegaDungeonGenerator) chooseRandom(other map[Region][]geometry.Point) (Region, geometry.Point) {
    var flatList []fxtools.Tuple[Region, geometry.Point]

    for reg, pos := range other {
        for _, p := range pos {
            flatList = append(flatList, fxtools.NewTuple(reg, p))
        }
    }
    if len(flatList) > 0 {
        randomIndex := rand.Intn(len(flatList))
        return flatList[randomIndex].Item1, flatList[randomIndex].Item2
    }
    return nil, geometry.Point{}

}

func (c *MegaDungeonGenerator) getConnectors(emptyMap *DungeonMap) map[Region]map[Region][]geometry.Point {
    availableConnectors := make(map[Region]map[Region][]geometry.Point)
    emptyMap.TraverseTilesRandomly(c.randomSource, func(pos geometry.Point) {
        if regions, areDifferent := c.isConnectingRegions(emptyMap, pos); areDifferent {
            regionOne := regions.Item1
            regionTwo := regions.Item2
            if _, ok := availableConnectors[regionOne]; !ok {
                availableConnectors[regionOne] = make(map[Region][]geometry.Point)
            }
            if _, ok := availableConnectors[regionTwo]; !ok {
                availableConnectors[regionTwo] = make(map[Region][]geometry.Point)
            }
            availableConnectors[regionOne][regionTwo] = append(availableConnectors[regionOne][regionTwo], pos)
            availableConnectors[regionTwo][regionOne] = append(availableConnectors[regionTwo][regionOne], pos)
        }
    })
    return availableConnectors
}

func (c *MegaDungeonGenerator) isConnectingRegions(emptyMap *DungeonMap, pos geometry.Point) (fxtools.Tuple[Region, Region], bool) {
    if !emptyMap.IsWallAt(pos) {
        return fxtools.Tuple[Region, Region]{}, false
    }
    // check north/south
    northOf := pos.Add(geometry.Point{X: 0, Y: -1})
    southOf := pos.Add(geometry.Point{X: 0, Y: 1})

    // check east/west
    eastOf := pos.Add(geometry.Point{X: 1, Y: 0})
    westOf := pos.Add(geometry.Point{X: -1, Y: 0})

    if regions, areDifferent := c.areInDifferentRegions(emptyMap, northOf, southOf); areDifferent && emptyMap.IsWallAt(westOf) && emptyMap.IsWallAt(eastOf) {
        return regions, true
    }

    if regions, areDifferent := c.areInDifferentRegions(emptyMap, eastOf, westOf); areDifferent && emptyMap.IsWallAt(northOf) && emptyMap.IsWallAt(southOf) {
        return regions, true
    }

    return fxtools.Tuple[Region, Region]{}, false
}

func (c *MegaDungeonGenerator) areInDifferentRegions(emptyMap *DungeonMap, posOne geometry.Point, posTwo geometry.Point) (fxtools.Tuple[Region, Region], bool) {
    if !emptyMap.Contains(posOne) || !emptyMap.Contains(posTwo) {
        return fxtools.Tuple[Region, Region]{}, false
    }

    if emptyMap.IsWallAt(posOne) || emptyMap.IsWallAt(posTwo) {
        return fxtools.Tuple[Region, Region]{}, false
    }

    northernRegion, northOk := c.regionLookup[posOne]
    southernRegion, southOk := c.regionLookup[posTwo]

    if !northOk || !southOk || northernRegion == southernRegion {
        return fxtools.Tuple[Region, Region]{}, false
    }
    if roomNorth, ok := northernRegion.(*DungeonRoom); ok {
        if roomNorth.IsCornerPosition(posOne) {
            return fxtools.Tuple[Region, Region]{}, false
        }
    }
    if roomSouth, ok := southernRegion.(*DungeonRoom); ok {
        if roomSouth.IsCornerPosition(posTwo) {
            return fxtools.Tuple[Region, Region]{}, false
        }
    }
    return fxtools.NewTuple(northernRegion, southernRegion), true
}

type MergedRegion struct {
    regions   []Region
    connector geometry.Point
}

func (m MergedRegion) GetAbsoluteFloorTiles() []geometry.Point {
    var result []geometry.Point
    for _, region := range m.regions {
        result = append(result, region.GetAbsoluteFloorTiles()...)
    }
    result = append(result, m.connector)
    return result
}

func NewMergedRegion(regionOne, regionTwo Region, connector geometry.Point) *MergedRegion {
    return &MergedRegion{
        regions:   []Region{regionOne, regionTwo},
        connector: connector,
    }
}
func (c *MegaDungeonGenerator) mergeRegions(regionOne Region, regionTwo Region, pos geometry.Point) *MergedRegion {
    merged := NewMergedRegion(regionOne, regionTwo, pos)
    // rebuild map
    clear(c.regionLookup)

    var newRegionList []Region
    for _, region := range c.allRegions {
        if region == regionOne || region == regionTwo {
            continue
        }
        newRegionList = append(newRegionList, region)
    }
    newRegionList = append(newRegionList, merged)
    c.allRegions = newRegionList

    for _, region := range c.allRegions {
        for _, regionPos := range region.GetAbsoluteFloorTiles() {
            c.regionLookup[regionPos] = region
        }
    }
    return merged
}

func (c *MegaDungeonGenerator) placeDoorsAtConnections(emptyMap *DungeonMap) {
    var placedDoors []geometry.Point
    for _, connector := range c.requiredConnections {
        emptyMap.SetDoor(connector.X, connector.Y)
        placedDoors = append(placedDoors, connector)
    }

    for _, connector := range c.optionalConnections {
        if c.hasMinDistToConnectors(placedDoors, connector, c.doorDistance) {
            emptyMap.SetDoor(connector.X, connector.Y)
            placedDoors = append(placedDoors, connector)
        }
    }
}

func (c *MegaDungeonGenerator) findConnectors() []geometry.Point {
    rootRegion := c.allRegions[0]
    var connectors []geometry.Point
    return c.connectorsFromRegionRecursive(rootRegion, connectors)
}

func (c *MegaDungeonGenerator) connectorsFromRegionRecursive(currentRegion Region, connectors []geometry.Point) []geometry.Point {
    if merged, ok := currentRegion.(*MergedRegion); ok {
        connectors = append(connectors, merged.connector)
        for _, region := range merged.regions {
            connectors = c.connectorsFromRegionRecursive(region, connectors)
        }
    }
    return connectors
}

func (c *MegaDungeonGenerator) hasMinDistToConnectors(existingConnections []geometry.Point, pos geometry.Point, minDist int) bool {
    for _, connector := range existingConnections {
        if geometry.DistanceManhattan(pos, connector) < minDist {
            return false
        }
    }
    return true
}

func fillMaze(emptyMap *DungeonMap, start geometry.Point, straightChance float64, rnd *rand.Rand) *CorridorRegion {
    openList := []geometry.Point{start}
    emptyMap.SetCorridor(start.X, start.Y)
    corridorRegion := make(CorridorRegion)
    corridorRegion[start] = true
    prevDirection := geometry.Point{X: 1, Y: 0}
    for len(openList) > 0 {
        //randomIndexFromOpenList := rnd.Intn(len(openList))
        openListIndex := len(openList) - 1
        pos := openList[openListIndex]
        openList = append(openList[:openListIndex], openList[openListIndex+1:]...)

        neighbours := emptyMap.GetFilteredCardinalNeighbours(pos, func(p geometry.Point) bool {
            return p.X > 0 && p.Y > 0 && p.X < emptyMap.width-1 && p.Y < emptyMap.height-1
        })

        randomIndexes := rnd.Perm(len(neighbours))

        if rnd.Float64() < straightChance {
            neighborInPrevDirection := pos.Add(prevDirection)
            if isCompletelyFreeForCarving(emptyMap, pos, neighborInPrevDirection) {
                emptyMap.SetCorridor(neighborInPrevDirection.X, neighborInPrevDirection.Y)
                corridorRegion[neighborInPrevDirection] = true
                openList = append(openList, pos)
                openList = append(openList, neighborInPrevDirection)
                prevDirection = neighborInPrevDirection.Sub(pos)
                continue
            }
        }

        for _, randomIndex := range randomIndexes {
            randomNeighbor := neighbours[randomIndex]
            if isCompletelyFreeForCarving(emptyMap, pos, randomNeighbor) {
                emptyMap.SetCorridor(randomNeighbor.X, randomNeighbor.Y)
                corridorRegion[randomNeighbor] = true
                openList = append(openList, pos)
                openList = append(openList, randomNeighbor)
                prevDirection = randomNeighbor.Sub(pos)
                break
            }
        }
    }
    return &corridorRegion
}

func findEmptySpot(random *rand.Rand, emptyMap *DungeonMap) geometry.Point {
    var start geometry.Point
    loc, found := emptyMap.GetRandomFiltered(random, func(pos geometry.Point) bool {
        return pos.X >= 1 && pos.Y >= 1 && pos.X < emptyMap.width-1 && pos.Y < emptyMap.height-1 && isCompletelyFree(emptyMap, pos)
    })
    if found {
        start = loc
    }
    return start

}

func isCompletelyFree(emptyMap *DungeonMap, pos geometry.Point) bool {
    neighborsWithWalls := emptyMap.GetAllFilteredNeighbours(pos, func(p geometry.Point) bool {
        return emptyMap.IsWallAt(p)
    })
    return emptyMap.Contains(pos) && emptyMap.IsWallAt(pos) && len(neighborsWithWalls) == 8
}

func isCompletelyFreeForCarving(emptyMap *DungeonMap, from geometry.Point, to geometry.Point) bool {
    direction := to.Sub(from)
    posBehind := to.Add(direction)
    leftOfDirection := direction.RotateLeft()
    rightOfDirection := direction.RotateRight()
    rightOfTo := to.Add(rightOfDirection)
    leftOfTo := to.Add(leftOfDirection)
    rightAndBehind := to.Add(rightOfDirection).Add(direction)
    leftAndBehind := to.Add(leftOfDirection).Add(direction)

    return emptyMap.Contains(to) && emptyMap.IsWallAt(to) &&
        emptyMap.Contains(posBehind) && emptyMap.IsWallAt(posBehind) &&
        emptyMap.Contains(rightOfTo) && emptyMap.IsWallAt(rightOfTo) &&
        emptyMap.Contains(leftOfTo) && emptyMap.IsWallAt(leftOfTo) &&
        emptyMap.Contains(rightAndBehind) && emptyMap.IsWallAt(rightAndBehind) &&
        emptyMap.Contains(leftAndBehind) && emptyMap.IsWallAt(leftAndBehind)
}

func makeOddForSize(rnd *rand.Rand, minValue, value int) int {
    if value%2 == 0 {
        if value-1 < minValue || rnd.Intn(2) == 0 {
            value++
        } else {
            value--
        }
    }
    return value
}

func makeOdd(rnd *rand.Rand, value int) int {
    if value%2 == 0 {
        if rnd.Intn(2) == 0 {
            value++
        } else {
            value--
        }
    }
    return value
}
