package gridmap

import (
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"io"
	"math"
	"math/rand"
	"sort"
	"time"
)

type NamedLocation struct {
	LocationName string
	Pos          geometry.Point
}
type Transition struct {
	TargetMap      string
	TargetLocation string
}

func (t Transition) IsEmpty() bool {
	return t.TargetMap == "" && t.TargetLocation == ""
}

func (t Transition) Encode() string {
	return fmt.Sprintf("transition(%s, %s)", t.TargetMap, t.TargetLocation)
}

func MustDecodeTransition(str string) Transition {
	var t Transition
	_, err := fmt.Sscanf(str, "transition(%s, %s)", &t.TargetMap, &t.TargetLocation)
	if err != nil {
		panic(err)
	}
	return t
}

type MapObject interface {
	Position() geometry.Point
	SetPosition(geometry.Point)
}

type MapActor interface {
	MapObject
}
type MapObjectWithProperties[ActorType interface {
	comparable
	MapActor
}] interface {
	MapObject
	IsWalkable(person ActorType) bool
	IsTransparent() bool
	IsPassableForProjectile() bool
}
type ZoneType int

func (t ZoneType) ToString() string {
	switch t {
	case ZoneTypePublic:
		return "Public"
	case ZoneTypePrivate:
		return "Private"
	case ZoneTypeHighSecurity:
		return "High Security"
	case ZoneTypeDropOff:
		return "Drop Off"
	}
	return "Unknown"
}

func NewZoneTypeFromString(str string) ZoneType {
	switch str {
	case "Public":
		return ZoneTypePublic
	case "Private":
		return ZoneTypePrivate
	case "High Security":
		return ZoneTypeHighSecurity
	case "Drop Off":
		return ZoneTypeDropOff
	}
	return ZoneTypePublic
}

const (
	ZoneTypePublic ZoneType = iota
	ZoneTypePrivate
	ZoneTypeHighSecurity
	ZoneTypeDropOff
)

type ZoneInfo struct {
	Name        string
	Type        ZoneType
	AmbienceCue string
}

const PublicZoneName = "Public Space"

func (i ZoneInfo) IsDropOff() bool {
	return i.Type == ZoneTypeDropOff
}

func (i ZoneInfo) IsHighSecurity() bool {
	return i.Type == ZoneTypeHighSecurity || i.Type == ZoneTypeDropOff
}

func (i ZoneInfo) IsPublic() bool {
	return i.Type == ZoneTypePublic
}

func (i ZoneInfo) IsPrivate() bool {
	return i.Type == ZoneTypePrivate
}

func (i ZoneInfo) ToString() string {
	return fmt.Sprintf("%s (%s)", i.Name, i.Type.ToString())
}

func NewZone(name string) *ZoneInfo {
	return &ZoneInfo{
		Name: name,
	}
}
func NewPublicZone(name string) *ZoneInfo {
	return &ZoneInfo{
		Name: name,
		Type: ZoneTypePublic,
	}
}

type GridMap[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}] struct {
	name            string
	cells           []MapCell[ActorType, ItemType, ObjectType]
	allActors       []ActorType
	allDownedActors []ActorType
	removedActors   []ActorType
	allItems        []ItemType
	allObjects      []ObjectType

	decals map[geometry.Point]int32

	playerSpawn geometry.Point

	mapWidth  int
	mapHeight int

	pathfinder  *geometry.PathRange
	listOfZones []*ZoneInfo
	zoneMap     []*ZoneInfo
	player      ActorType

	namedLocations   map[string]geometry.Point
	ambienceSoundCue string
	noClip           bool

	transitionMap map[geometry.Point]Transition
	secretDoors   map[geometry.Point]bool

	namedRects   map[string]geometry.Rect
	namedTrigger map[string]Trigger

	namedPaths     map[string][]geometry.Point
	displayName    string
	metaInfoString []string

	cardinalMovementOnly bool

	// LIGHTING

	DynamicLights        map[geometry.Point]*LightSource
	BakedLights          map[geometry.Point]*LightSource
	lightfov             *geometry.FOV
	AmbientLight         fxtools.HDRColor
	MaxLightIntensity    float64
	dynamicallyLitCells  map[geometry.Point]fxtools.HDRColor
	DynamicLightsChanged bool
	isIndoor             bool
	meta                 MapMeta
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetCardinalMovementOnly(cardinalMovementOnly bool) {
	m.cardinalMovementOnly = cardinalMovementOnly
}
func (m *GridMap[ActorType, ItemType, ObjectType]) AddZone(zone *ZoneInfo) {
	m.listOfZones = append(m.listOfZones, zone)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetTile(position geometry.Point, mapTile Tile) {
	if !m.Contains(position) {
		return
	}
	index := position.Y*m.mapWidth + position.X
	m.cells[index].TileType = mapTile
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetZone(position geometry.Point, zone *ZoneInfo) {
	if !m.Contains(position) {
		return
	}
	m.zoneMap[position.Y*m.mapWidth+position.X] = zone
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveItemAt(position geometry.Point) {
	m.RemoveItem(m.ItemAt(position))
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveObjectAt(position geometry.Point) {
	m.RemoveObject(m.ObjectAt(position))
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveObject(obj ObjectType) {
	m.cells[obj.Position().Y*m.mapWidth+obj.Position().X] = m.cells[obj.Position().Y*m.mapWidth+obj.Position().X].WithObjectHereRemoved(obj)
	for i := len(m.allObjects) - 1; i >= 0; i-- {
		if m.allObjects[i] == obj {
			m.allObjects = append(m.allObjects[:i], m.allObjects[i+1:]...)
			return
		}
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IterAll(f func(p geometry.Point, c MapCell[ActorType, ItemType, ObjectType])) {
	for y := 0; y < m.mapHeight; y++ {
		for x := 0; x < m.mapWidth; x++ {
			f(geometry.Point{X: x, Y: y}, m.cells[y*m.mapWidth+x])
		}
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IterWindow(window geometry.Rect, f func(p geometry.Point, c MapCell[ActorType, ItemType, ObjectType])) {
	for y := window.Min.Y; y < window.Max.Y; y++ {
		for x := window.Min.X; x < window.Max.X; x++ {
			mapPos := geometry.Point{X: x, Y: y}
			if !m.Contains(mapPos) {
				continue
			}
			f(mapPos, m.cells[y*m.mapWidth+x])
		}
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetPlayerSpawn(position geometry.Point) {
	m.playerSpawn = position
}

func (m *GridMap[ActorType, ItemType, ObjectType]) CellAt(location geometry.Point) MapCell[ActorType, ItemType, ObjectType] {
	return m.cells[m.mapWidth*location.Y+location.X]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ItemAt(location geometry.Point) ItemType {
	return *m.cells[m.mapWidth*location.Y+location.X].Item
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsItemAt(location geometry.Point) bool {
	return m.cells[m.mapWidth*location.Y+location.X].Item != nil
}
func (m *GridMap[ActorType, ItemType, ObjectType]) SetActorToDowned(a ActorType) {
	if !m.RemoveActor(a) {
		println("Could not remove actor from map")
		return
	}
	m.allDownedActors = append(m.allDownedActors, a)
	if m.IsDownedActorAt(a.Position()) && m.DownedActorAt(a.Position()) != a {
		m.displaceDownedActor(a)
		return
	}
	m.cells[a.Position().Y*m.mapWidth+a.Position().X] = m.cells[a.Position().Y*m.mapWidth+a.Position().X].WithDownedActor(a)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetActorToRemoved(person ActorType) {
	m.RemoveActor(person)
	m.RemoveDownedActor(person)
	m.removedActors = append(m.removedActors, person)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetActorToNormal(person ActorType) {
	m.RemoveDownedActor(person)
	m.allActors = append(m.allActors, person)
	if m.IsActorAt(person.Position()) && m.ActorAt(person.Position()) != person {
		m.displaceActor(person, person.Position())
		return
	}
	m.cells[person.Position().Y*m.mapWidth+person.Position().X] = m.cells[person.Position().Y*m.mapWidth+person.Position().X].WithActor(person)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) MoveItem(item ItemType, to geometry.Point) {
	m.cells[item.Position().Y*m.mapWidth+item.Position().X] = m.cells[item.Position().Y*m.mapWidth+item.Position().X].WithItemHereRemoved(item)
	item.SetPosition(to)
	//fxtools.XYToIndex()
	m.cells[to.Y*m.mapWidth+to.X] = m.cells[to.Y*m.mapWidth+to.X].WithItem(item)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetRandomFreeAndSafeNeighbor(source *rand.Rand, location geometry.Point) geometry.Point {
	freeNeighbors := m.GetFilteredNeighbors(location, func(p geometry.Point) bool {
		return m.Contains(p) && m.IsCurrentlyPassable(p) && !m.IsObviousHazardAt(p)
	})
	if len(freeNeighbors) == 0 {
		return location
	}
	return freeNeighbors[source.Intn(len(freeNeighbors))]
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetRandomNeighbor(source *rand.Rand, location geometry.Point) geometry.Point {
	freeNeighbors := m.GetFilteredNeighbors(location, func(p geometry.Point) bool {
		return m.Contains(p)
	})
	if len(freeNeighbors) == 0 {
		return location
	}
	return freeNeighbors[source.Intn(len(freeNeighbors))]
}

type SetOfPoints map[geometry.Point]bool

func (s *SetOfPoints) Pop() geometry.Point {
	for k := range *s {
		delete(*s, k)
		return k
	}
	return geometry.Point{}
}
func (s *SetOfPoints) Contains(p geometry.Point) bool {
	_, ok := (*s)[p]
	return ok
}

func (s *SetOfPoints) ToSlice() []geometry.Point {
	result := make([]geometry.Point, 0)
	for k := range *s {
		result = append(result, k)
	}
	return result
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetNearestUnexploredWalkablePosition(currentPosition geometry.Point) geometry.Point {
	distribution := m.GetFreeCellsForDistribution(currentPosition, 1, func(p geometry.Point) bool {
		return m.Contains(p) && m.IsTileWalkable(p) && !m.IsExplored(p)
	})
	if len(distribution) > 0 {
		return distribution[0]
	}
	return currentPosition
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNearestExploredItem(currentPosition geometry.Point, filter func(item ItemType) bool) geometry.Point {
	distribution := m.GetFreeCellsForDistribution(currentPosition, 1, func(p geometry.Point) bool {
		isValidItemAtPos := m.Contains(p) && m.IsTileWalkable(p) && m.IsExplored(p) && m.IsItemAt(p)
		return isValidItemAtPos && filter(m.ItemAt(p))
	})
	if len(distribution) > 0 {
		return distribution[0]
	}
	return currentPosition
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetFreeCellsForDistribution(position geometry.Point, neededCellCount int, freePredicate func(p geometry.Point) bool) []geometry.Point {
	foundFreeCells := make(SetOfPoints)
	currentPosition := position
	openList := make(SetOfPoints)
	closedList := make(SetOfPoints)
	closedList[currentPosition] = true

	for _, neighbor := range m.GetFilteredNeighbors(currentPosition, m.IsTileWalkable) {
		openList[neighbor] = true
	}
	for len(foundFreeCells) < neededCellCount && len(openList) > 0 {
		freeNeighbors := m.GetFilteredNeighbors(currentPosition, freePredicate)
		for _, neighbor := range freeNeighbors {
			foundFreeCells[neighbor] = true
		}
		// pop from open list
		pop := openList.Pop()
		currentPosition = pop
		for _, neighbor := range m.GetFilteredNeighbors(currentPosition, m.IsTileWalkable) {
			if !closedList.Contains(neighbor) {
				openList[neighbor] = true
			}
		}
		closedList[currentPosition] = true
	}

	freeCells := foundFreeCells.ToSlice()
	sort.Slice(freeCells, func(i, j int) bool {
		return geometry.DistanceSquared(freeCells[i], position) < geometry.DistanceSquared(freeCells[j], position)
	})
	return freeCells
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveItem(item ItemType) {
	m.cells[item.Position().Y*m.mapWidth+item.Position().X] = m.cells[item.Position().Y*m.mapWidth+item.Position().X].WithItemHereRemoved(item)
	for i := len(m.allItems) - 1; i >= 0; i-- {
		if m.allItems[i] == item {
			m.allItems = append(m.allItems[:i], m.allItems[i+1:]...)
			return
		}
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetAllCardinalNeighbors(pos geometry.Point) []geometry.Point {
	neighbors := geometry.Neighbors{}
	allCardinalNeighbors := neighbors.Cardinal(pos, func(p geometry.Point) bool {
		return m.Contains(p)
	})
	return allCardinalNeighbors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetAllDiagonalNeighbors(pos geometry.Point) []geometry.Point {
	neighbors := geometry.Neighbors{}
	allCardinalNeighbors := neighbors.Diagonal(pos, func(p geometry.Point) bool {
		return m.Contains(p)
	})
	return allCardinalNeighbors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) WavePropagationFrom(pos geometry.Point, size int, pressure int) map[int][]geometry.Point {
	soundAnimationMap := make(map[int][]geometry.Point)
	m.pathfinder.DijkstraMap(m.getDijkstraMapperWithActorsNotBlocking(), []geometry.Point{pos}, size)
	for _, v := range m.pathfinder.DijkstraIterNodes {
		cost := v.Cost
		point := v.P
		if soundAnimationMap[cost] == nil {
			soundAnimationMap[cost] = make([]geometry.Point, 0)
		}
		soundAnimationMap[cost] = append(soundAnimationMap[cost], point)
	}
	return soundAnimationMap
}

type DijkstraMapper struct {
	neighbors func(geometry.Point) []geometry.Point
	cost      func(geometry.Point, geometry.Point) int
}

func (d DijkstraMapper) Neighbors(point geometry.Point) []geometry.Point {
	return d.neighbors(point)
}

func (d DijkstraMapper) Cost(point geometry.Point, point2 geometry.Point) int {
	return d.cost(point, point2)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) getDijkstraMapperWithActorsNotBlocking() DijkstraMapper {
	return DijkstraMapper{
		neighbors: func(point geometry.Point) []geometry.Point {
			return m.GetFilteredNeighbors(point, func(p geometry.Point) bool {
				return m.Contains(p) && m.IsWalkable(p)
			})
		},
		cost: func(point geometry.Point, point2 geometry.Point) int {
			return int(geometry.Distance(point, point2) * 10)
		},
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) getDijkstraMapperForExploredWithActorsNotBlocking() DijkstraMapper {
	return DijkstraMapper{
		neighbors: func(point geometry.Point) []geometry.Point {
			return m.GetFilteredNeighbors(point, func(p geometry.Point) bool {
				return m.Contains(p) && m.IsTileWalkable(p) && m.IsExplored(p)
			})
		},
		cost: func(point geometry.Point, point2 geometry.Point) int {
			return int(geometry.Distance(point, point2) * 10)
		},
	}
}
func (m *GridMap[ActorType, ItemType, ObjectType]) getDijkstraMapper(passable func(pos geometry.Point) bool) DijkstraMapper {
	return DijkstraMapper{
		neighbors: func(point geometry.Point) []geometry.Point {
			return m.GetFilteredNeighbors(point, func(p geometry.Point) bool {
				return m.Contains(p) && passable(p)
			})
		},
		cost: func(point geometry.Point, point2 geometry.Point) int {
			return int(geometry.Distance(point, point2) * 10)
		},
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetDijkstraMapWithActorsNotBlocking(start geometry.Point, maxCost int) map[geometry.Point]int {
	nodes := m.pathfinder.DijkstraMap(m.getDijkstraMapperWithActorsNotBlocking(), []geometry.Point{start}, maxCost*10)

	dijkstraMap := make(map[geometry.Point]int)
	for _, v := range nodes {
		dijkstraMap[v.P] = v.Cost
	}
	return dijkstraMap
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetDijkstraMapForExploredWithActorsNotBlocking(start geometry.Point, maxCost int) map[geometry.Point]int {
	nodes := m.pathfinder.DijkstraMap(m.getDijkstraMapperForExploredWithActorsNotBlocking(), []geometry.Point{start}, maxCost*10)

	dijkstraMap := make(map[geometry.Point]int)
	for _, v := range nodes {
		dijkstraMap[v.P] = v.Cost
	}
	return dijkstraMap
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetDijkstraMap(start geometry.Point, maxCost int, passable func(point geometry.Point) bool) map[geometry.Point]int {
	nodes := m.pathfinder.DijkstraMap(m.getDijkstraMapper(passable), []geometry.Point{start}, maxCost*10)

	dijkstraMap := make(map[geometry.Point]int)
	for _, v := range nodes {
		dijkstraMap[v.P] = v.Cost
	}
	return dijkstraMap
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetConnected(startLocation geometry.Point, traverse func(p geometry.Point) bool) []geometry.Point {
	results := make([]geometry.Point, 0)
	for _, node := range m.pathfinder.BreadthFirstMap(MapPather{neighborPredicate: traverse, allNeighbors: m.GetAllCardinalNeighbors}, []geometry.Point{startLocation}, 100) {
		results = append(results, node.P)
	}
	return results
}
func NewMapFromString[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}](width, height int, inputString []rune, mapper func(gridMap *GridMap[ActorType, ItemType, ObjectType], icon rune, pos geometry.Point)) *GridMap[ActorType, ItemType, ObjectType] {
	emptyMap := NewEmptyMap[ActorType, ItemType, ObjectType](width, height)
	size := width * height
	if len(inputString) != size {
		panic("Input string does not match map size")
		return emptyMap
	}
	for i := 0; i < size; i++ {
		icon := inputString[i]
		mapper(emptyMap, icon, geometry.Point{X: i % width, Y: i / width})
	}
	return emptyMap
}

// update for entities:
// call update for every updatable entity (genMap, AllItems, AllObjects, tiles)
// default: just return
// entities have an internal schedule, waiting for ticks to happen

func NewEmptyMap[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}](width, height int) *GridMap[ActorType, ItemType, ObjectType] {
	pathRange := geometry.NewPathRange(geometry.NewRect(0, 0, width, height))
	publicSpaceZone := NewPublicZone(PublicZoneName)
	m := &GridMap[ActorType, ItemType, ObjectType]{
		cells:               make([]MapCell[ActorType, ItemType, ObjectType], width*height),
		allActors:           make([]ActorType, 0),
		allDownedActors:     make([]ActorType, 0),
		allItems:            make([]ItemType, 0),
		allObjects:          make([]ObjectType, 0),
		namedLocations:      map[string]geometry.Point{},
		listOfZones:         []*ZoneInfo{publicSpaceZone},
		zoneMap:             NewZoneMap(publicSpaceZone, width, height),
		mapWidth:            width,
		mapHeight:           height,
		pathfinder:          pathRange,
		secretDoors:         make(map[geometry.Point]bool),
		transitionMap:       make(map[geometry.Point]Transition),
		namedRects:          make(map[string]geometry.Rect),
		namedTrigger:        make(map[string]Trigger),
		namedPaths:          make(map[string][]geometry.Point),
		decals:              make(map[geometry.Point]int32),
		DynamicLights:       make(map[geometry.Point]*LightSource),
		BakedLights:         make(map[geometry.Point]*LightSource),
		lightfov:            geometry.NewFOV(geometry.NewRect(0, 0, width, height)),
		dynamicallyLitCells: make(map[geometry.Point]fxtools.HDRColor),
	}
	return m
}

func (m *GridMap[ActorType, ItemType, ObjectType]) FillTile(tile Tile) {
	for i := range m.cells {
		m.cells[i].TileType = tile
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetCell(p geometry.Point) MapCell[ActorType, ItemType, ObjectType] {
	return m.cells[p.X+p.Y*m.mapWidth]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetCell(p geometry.Point, cell MapCell[ActorType, ItemType, ObjectType]) {
	m.cells[p.X+p.Y*m.mapWidth] = cell
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetCellByIndex(index int, cell MapCell[ActorType, ItemType, ObjectType]) {
	m.cells[index] = cell
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetActor(p geometry.Point) ActorType {
	return *m.cells[p.X+p.Y*m.mapWidth].Actor
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveActor(actor ActorType) bool {
	m.cells[actor.Position().X+actor.Position().Y*m.mapWidth] = m.cells[actor.Position().X+actor.Position().Y*m.mapWidth].WithActorHereRemoved(actor)
	for i := len(m.allActors) - 1; i >= 0; i-- {
		if m.allActors[i] == actor {
			m.allActors = append(m.allActors[:i], m.allActors[i+1:]...)
			return true
		}
	}
	return false
}

// MoveActor Should only be called my the model, so we can ensure that a HUD IsDone will follow
func (m *GridMap[ActorType, ItemType, ObjectType]) MoveActor(actor ActorType, newPos geometry.Point) {
	m.MoveActorFrom(actor, actor.Position(), newPos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) MoveActorFrom(actor ActorType, from, to geometry.Point) {
	if !m.Contains(to) {
		return
	}
	if !m.IsWalkableFor(to, actor) {
		return
	}
	if m.Contains(from) {
		m.cells[from.X+from.Y*m.mapWidth] = m.cells[from.X+from.Y*m.mapWidth].WithActorHereRemoved(actor)
	}
	actor.SetPosition(to)
	m.cells[to.X+to.Y*m.mapWidth] = m.cells[to.X+to.Y*m.mapWidth].WithActor(actor)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) MoveObject(obj ObjectType, newPos geometry.Point) {
	m.cells[obj.Position().X+obj.Position().Y*m.mapWidth] = m.cells[obj.Position().X+obj.Position().Y*m.mapWidth].WithObjectHereRemoved(obj)
	obj.SetPosition(newPos)
	m.cells[newPos.X+newPos.Y*m.mapWidth] = m.cells[newPos.X+newPos.Y*m.mapWidth].WithObject(obj)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Fill(mapCell MapCell[ActorType, ItemType, ObjectType]) {
	for i := range m.cells {
		m.cells[i] = mapCell
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsTransparent(p geometry.Point) bool {
	if !m.Contains(p) {
		return false
	}

	if objectAt, ok := m.TryGetObjectAt(p); ok && !objectAt.IsTransparent() {
		return false
	}

	return m.GetCell(p).TileType.IsTransparent
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsTileWalkable(point geometry.Point) bool {
	if !m.Contains(point) {
		return false
	}
	return m.GetCell(point).TileType.IsWalkable
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Contains(dest geometry.Point) bool {
	return dest.X >= 0 && dest.X < m.mapWidth && dest.Y >= 0 && dest.Y < m.mapHeight
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsActorAt(location geometry.Point) bool {
	if !m.Contains(location) {
		return false
	}
	return m.cells[location.X+location.Y*m.mapWidth].Actor != nil
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ActorAt(location geometry.Point) ActorType {
	return *m.cells[location.X+location.Y*m.mapWidth].Actor
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsDownedActorAt(location geometry.Point) bool {
	if !m.Contains(location) {
		return false
	}
	return m.cells[location.X+location.Y*m.mapWidth].DownedActor != nil
}

func (m *GridMap[ActorType, ItemType, ObjectType]) DownedActorAt(location geometry.Point) ActorType {
	return *m.cells[location.X+location.Y*m.mapWidth].DownedActor
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsObjectAt(location geometry.Point) bool {
	if !m.Contains(location) {
		return false
	}
	return m.cells[location.X+location.Y*m.mapWidth].Object != nil
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ObjectAt(location geometry.Point) ObjectType {
	return *m.cells[location.X+location.Y*m.mapWidth].Object
}
func (m *GridMap[ActorType, ItemType, ObjectType]) ZoneAt(p geometry.Point) *ZoneInfo {
	return m.zoneMap[m.mapWidth*p.Y+p.X]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredCardinalNeighbors(pos geometry.Point, filter func(geometry.Point) bool) []geometry.Point {
	neighbors := geometry.Neighbors{}
	filtered := neighbors.Cardinal(pos, filter)
	return filtered
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Actors() []ActorType {
	return m.allActors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) DownedActors() []ActorType {
	return m.allDownedActors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Items() []ItemType {
	return m.allItems
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Objects() []ObjectType {
	return m.allObjects
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredNeighbors(pos geometry.Point, filter func(geometry.Point) bool) []geometry.Point {
	neighbors := geometry.Neighbors{}
	filtered := neighbors.All(pos, filter)
	return filtered
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredNeighborsForMovement(pos geometry.Point, filter func(geometry.Point) bool) []geometry.Point {
	neighbors := geometry.Neighbors{}
	var result []geometry.Point
	if m.cardinalMovementOnly {
		result = neighbors.Cardinal(pos, filter)
	} else {
		result = neighbors.All(pos, filter)
	}
	return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredDirectionsForMovement(filter func(geometry.Point) bool) []geometry.Point {
	neighbors := geometry.Neighbors{}
	var result []geometry.Point
	if m.cardinalMovementOnly {
		result = neighbors.Cardinal(geometry.Point{}, filter)
	} else {
		result = neighbors.All(geometry.Point{}, filter)
	}
	return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFreeCardinalNeighbors(pos geometry.Point) []geometry.Point {
	neighbors := geometry.Neighbors{}
	freeNeighbors := neighbors.Cardinal(pos, func(p geometry.Point) bool {
		return m.Contains(p) && m.IsWalkable(p) && !m.IsActorAt(p)
	})
	return freeNeighbors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFreeMovementNeighbors(pos geometry.Point) []geometry.Point {
	return m.GetFilteredNeighborsForMovement(pos, func(p geometry.Point) bool {
		return m.Contains(p) && m.IsCurrentlyPassable(p) && !m.IsObviousHazardAt(p)
	})
}

func (m *GridMap[ActorType, ItemType, ObjectType]) displaceDownedActor(a ActorType) {
	free := m.GetFreeCellsForDistribution(a.Position(), 1, func(p geometry.Point) bool {
		return !m.IsDownedActorAt(p) && m.IsWalkable(p)
	})
	if len(free) == 0 {
		return
	}
	freePos := free[0]
	m.MoveDownedActor(a, freePos)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) displaceActor(a ActorType, position geometry.Point) {
	free := m.GetFreeCellsForDistribution(position, 1, func(p geometry.Point) bool {
		return m.CanPlaceActorHere(p)
	})
	if len(free) == 0 {
		return
	}
	freePos := free[0]
	m.MoveActor(a, freePos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddItemWithDisplacement(a ItemType, targetPos geometry.Point) {
	if !m.Contains(targetPos) {
		return
	}
	if m.CanPlaceItemHere(targetPos) {
		m.AddItem(a, targetPos)
		return
	}
	free := m.GetFreeCellsForDistribution(targetPos, 1, func(p geometry.Point) bool {
		return m.CanPlaceItemHere(p)
	})
	if len(free) == 0 {
		println("WARNING: Could not find a free spot for item")
		return
	}
	freePos := free[0]
	m.AddItem(a, freePos)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetJPSPath(start geometry.Point, end geometry.Point, isWalkable func(geometry.Point) bool) []geometry.Point {
	if !isWalkable(end) {
		end = m.getNearestFreeNeighbor(start, end, isWalkable)
	}
	//println(fmt.Sprintf("JPS from %v to %v", start, end))
	return m.pathfinder.JPSPath([]geometry.Point{}, start, end, isWalkable, !m.cardinalMovementOnly)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) getNearestFreeNeighbor(origin, pos geometry.Point, isFree func(geometry.Point) bool) geometry.Point {
	dist := math.MaxInt32
	nearest := pos
	for _, neighbor := range m.NeighborsCardinal(pos, isFree) {
		d := geometry.DistanceManhattan(origin, neighbor)
		if d < dist {
			dist = d
			nearest = neighbor
		}
	}
	return nearest
}

func (m *GridMap[ActorType, ItemType, ObjectType]) getCurrentlyPassableNeighbors(pos geometry.Point) []geometry.Point {
	neighbors := geometry.Neighbors{}
	freeNeighbors := neighbors.All(pos, func(p geometry.Point) bool {
		return m.Contains(p) && m.IsCurrentlyPassable(p)
	})
	return freeNeighbors
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsCurrentlyPassable(p geometry.Point) bool {
	if !m.Contains(p) {
		return false
	}
	return m.IsWalkable(p) && (!m.IsActorAt(p)) //&& !knownAsBlocked
}
func (m *GridMap[ActorType, ItemType, ObjectType]) CurrentlyPassableAndSafeForActor(person ActorType) func(p geometry.Point) bool {
	return func(p geometry.Point) bool {
		if !m.Contains(p) ||
			(m.IsActorAt(p) && m.ActorAt(p) != person) {
			return false
		}
		return m.IsWalkableFor(p, person) && !m.IsObviousHazardAt(p)
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsWalkable(p geometry.Point) bool {
	if !m.Contains(p) {
		return false
	}
	var noActor ActorType
	if m.IsObjectAt(p) && (!m.ObjectAt(p).IsWalkable(noActor)) {
		return false
	}
	cellAt := m.GetCell(p)
	return cellAt.TileType.IsWalkable
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsObviousHazardAt(p geometry.Point) bool {
	return m.IsDamagingTileAt(p)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsWalkableFor(p geometry.Point, person ActorType) bool {
	if !m.Contains(p) {
		return false
	}

	if m.IsActorAt(p) && m.ActorAt(p) != person {
		return false
	}

	if m.noClip {
		return true
	}

	if m.IsObjectAt(p) && (!m.ObjectAt(p).IsWalkable(person)) {
		return false
	}

	cellAt := m.GetCell(p)
	return cellAt.TileType.IsWalkable

}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsInHostileZone(person ActorType) bool {
	ourPos := person.Position()
	zoneAt := m.ZoneAt(ourPos)
	if zoneAt == nil || zoneAt.IsPublic() {
		return false
	}
	return zoneAt.IsHighSecurity()
}

func (m *GridMap[ActorType, ItemType, ObjectType]) CurrentlyPassableForActor(person ActorType) func(p geometry.Point) bool {
	return func(p geometry.Point) bool {
		if !m.Contains(p) ||
			(m.IsActorAt(p) && m.ActorAt(p) != person) {
			return false
		}
		return m.IsWalkableFor(p, person)
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsExplored(pos geometry.Point) bool {
	return m.GetCell(pos).IsExplored
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetExplored(pos geometry.Point) {
	if !m.Contains(pos) {
		return
	}
	m.cells[pos.X+pos.Y*m.mapWidth].IsExplored = true
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetPositionsByFilter(filter func(tile *MapCell[ActorType, ItemType, ObjectType]) bool) []geometry.Point {
	result := make([]geometry.Point, 0)
	for index, c := range m.cells {
		if filter(&c) {
			x := index % m.mapWidth
			y := index / m.mapWidth
			result = append(result, geometry.Point{X: x, Y: y})
		}
	}
	return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNearestTile(pos geometry.Point, filter func(tile Tile) bool) geometry.Point {
	result := pos

	allReachableExplored := m.GetDijkstraMapForExploredWithActorsNotBlocking(pos, 1000)
	minDist := math.MaxInt32

	for loc, dist := range allReachableExplored {
		if filter(m.GetCell(loc).TileType) {
			if dist < minDist {
				result = loc
				minDist = dist
			}
		}
	}
	return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNearestActor(pos geometry.Point, filter func(actor ActorType) bool) ActorType {
	var nearestActor ActorType
	nearestDistance := math.MaxInt32
	for _, actor := range m.allActors {
		if filter(actor) {
			dist := geometry.DistanceSquared(pos, actor.Position())
			if dist < nearestDistance {
				nearestDistance = dist
				nearestActor = actor
			}
		}
	}
	return nearestActor
}

func (m *GridMap[ActorType, ItemType, ObjectType]) currentlyPassablePather() MapPather {
	return MapPather{
		allNeighbors:      m.getCurrentlyPassableNeighbors,
		neighborPredicate: func(pos geometry.Point) bool { return true },
		pathCostFunc:      func(from geometry.Point, to geometry.Point) int { return 1 },
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SwapDownedPositions(downedActorOne ActorType, downedActorTwo ActorType) {
	posTwo := downedActorTwo.Position()
	posOne := downedActorOne.Position()
	downedActorOne.SetPosition(posTwo)
	downedActorTwo.SetPosition(posOne)
	m.cells[posOne.X+posOne.Y*m.mapWidth] = m.cells[posOne.X+posOne.Y*m.mapWidth].WithDownedActor(downedActorTwo)
	m.cells[posTwo.X+posTwo.Y*m.mapWidth] = m.cells[posTwo.X+posTwo.Y*m.mapWidth].WithDownedActor(downedActorOne)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SwapPositions(actorOne ActorType, actorTwo ActorType) {
	posTwo := actorTwo.Position()
	posOne := actorOne.Position()
	actorOne.SetPosition(posTwo)
	actorTwo.SetPosition(posOne)
	m.cells[posOne.X+posOne.Y*m.mapWidth] = m.cells[posOne.X+posOne.Y*m.mapWidth].WithActor(actorTwo)
	m.cells[posTwo.X+posTwo.Y*m.mapWidth] = m.cells[posTwo.X+posTwo.Y*m.mapWidth].WithActor(actorOne)
}

func NewZoneMap(zone *ZoneInfo, width int, height int) []*ZoneInfo {
	zoneMap := make([]*ZoneInfo, width*height)
	for i := 0; i < width*height; i++ {
		zoneMap[i] = zone
	}
	return zoneMap
}

type MapPather struct {
	neighborPredicate func(pos geometry.Point) bool
	allNeighbors      func(pos geometry.Point) []geometry.Point
	pathCostFunc      func(from geometry.Point, to geometry.Point) int
}

func (m MapPather) Neighbors(point geometry.Point) []geometry.Point {
	neighbors := make([]geometry.Point, 0)
	for _, p := range m.allNeighbors(point) {
		if m.neighborPredicate(p) {
			neighbors = append(neighbors, p)
		}
	}
	return neighbors
}
func (m MapPather) Cost(from geometry.Point, to geometry.Point) int {
	return m.pathCostFunc(from, to)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) MoveDownedActor(actor ActorType, newPos geometry.Point) {
	if m.cells[newPos.Y*m.mapWidth+newPos.X].DownedActor != nil {
		return
	}
	if m.cells[actor.Position().Y*m.mapWidth+actor.Position().X].DownedActor != nil && *m.cells[actor.Position().Y*m.mapWidth+actor.Position().X].DownedActor == actor {
		m.cells[actor.Position().Y*m.mapWidth+actor.Position().X] = m.cells[actor.Position().Y*m.mapWidth+actor.Position().X].WithDownedActorHereRemoved(actor)
	}
	actor.SetPosition(newPos)
	m.cells[newPos.Y*m.mapWidth+newPos.X] = m.cells[newPos.Y*m.mapWidth+newPos.X].WithDownedActor(actor)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveDownedActor(actor ActorType) bool {
	m.cells[actor.Position().Y*m.mapWidth+actor.Position().X] = m.cells[actor.Position().Y*m.mapWidth+actor.Position().X].WithDownedActorHereRemoved(actor)
	for i := len(m.allDownedActors) - 1; i >= 0; i-- {
		if m.allDownedActors[i] == actor {
			m.allDownedActors = append(m.allDownedActors[:i], m.allDownedActors[i+1:]...)
			return true
		}
	}
	return false
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Apply(f func(cell MapCell[ActorType, ItemType, ObjectType]) MapCell[ActorType, ItemType, ObjectType]) {
	for i, cell := range m.cells {
		m.cells[i] = f(cell)
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNearestDropOffPosition(pos geometry.Point) geometry.Point {
	nearestLocation := geometry.Point{X: 1, Y: 1}
	shortestDistance := math.MaxInt
	for index, c := range m.zoneMap {
		xPos := index % m.mapWidth
		yPos := index / m.mapWidth
		curPos := geometry.Point{X: xPos, Y: yPos}
		curDist := geometry.DistanceManhattan(curPos, pos)
		isItemHere := m.IsItemAt(curPos)
		isValidZone := c.IsDropOff()
		if !isItemHere && isValidZone && curDist < shortestDistance {
			nearestLocation = curPos
			shortestDistance = curDist
		}
	}
	return nearestLocation
}

func (m *GridMap[ActorType, ItemType, ObjectType]) FindNearestItem(pos geometry.Point, predicate func(item ItemType) bool) ItemType {
	var nearestItem ItemType
	nearestDistance := math.MaxInt
	for _, item := range m.allItems {
		if predicate(item) {
			curDist := geometry.DistanceManhattan(item.Position(), pos)
			if curDist < nearestDistance {
				nearestItem = item
				nearestDistance = curDist
			}
		}
	}
	return nearestItem
}

func (m *GridMap[ActorType, ItemType, ObjectType]) FindAllNearbyActors(source geometry.Point, maxDist int, keep func(actor ActorType) bool) []ActorType {
	results := make([]ActorType, 0)
	for _, actor := range m.allActors {
		if keep(actor) && geometry.DistanceManhattan(actor.Position(), source) <= maxDist {
			results = append(results, actor)
		}
	}
	return results
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsNamedLocationAt(positionInWorld geometry.Point) bool {
	for _, loc := range m.namedLocations {
		if loc == positionInWorld {
			return true
		}
	}
	return false
}
func (m *GridMap[ActorType, ItemType, ObjectType]) ZoneNames() []string {
	result := make([]string, 0)
	for _, zone := range m.listOfZones {
		result = append(result, zone.Name)
	}
	return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Resize(width int, height int, emptyTile Tile) {
	oldWidth := m.mapWidth
	oldHeight := m.mapHeight

	newCells := make([]MapCell[ActorType, ItemType, ObjectType], width*height)
	newZoneMap := make([]*ZoneInfo, width*height)
	for i := 0; i < width*height; i++ {
		newZoneMap[i] = m.listOfZones[0]
	}
	for i := 0; i < width*height; i++ {
		newCells[i] = MapCell[ActorType, ItemType, ObjectType]{
			TileType:   emptyTile,
			IsExplored: true,
		}
	}

	// copy over the old cells into the center of the new map
	for y := 0; y < oldHeight; y++ {
		if y >= height {
			break
		}
		for x := 0; x < oldWidth; x++ {
			if x >= width {
				break
			}
			destIndex := y*width + x
			srcIndex := y*oldWidth + x
			newCells[destIndex] = m.cells[srcIndex]
			newZoneMap[destIndex] = m.zoneMap[srcIndex]
		}
	}

	m.cells = newCells
	m.zoneMap = newZoneMap
	m.mapWidth = width
	m.mapHeight = height
}
func (m *GridMap[ActorType, ItemType, ObjectType]) MapSize() geometry.Point {
	return geometry.Point{X: m.mapWidth, Y: m.mapHeight}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) NeighborsAll(pos geometry.Point, filter func(p geometry.Point) bool) []geometry.Point {
	neighbors := geometry.Neighbors{}
	return neighbors.All(pos, filter)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) NeighborsCardinal(pos geometry.Point, filter func(p geometry.Point) bool) []geometry.Point {
	neighbors := geometry.Neighbors{}
	return neighbors.Cardinal(pos, filter)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNearestWalkableNeighbor(start geometry.Point, dest geometry.Point) geometry.Point {
	minDist := math.MaxInt
	var minPos geometry.Point
	for _, neighbor := range m.NeighborsCardinal(dest, m.IsWalkable) {
		dist := geometry.DistanceManhattan(neighbor, start)
		if dist < minDist {
			minDist = dist
			minPos = neighbor
		}
	}
	return minPos
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsPassableForProjectile(p geometry.Point) bool {
	isTileWalkable := m.IsTileWalkable(p)
	isActorOnTile := m.IsActorAt(p)
	isObjectOnTile := m.IsObjectAt(p)
	isObjectBlocking := false
	if isObjectOnTile {
		objectOnTile := m.ObjectAt(p)
		isObjectBlocking = !objectOnTile.IsPassableForProjectile()
	}
	return isTileWalkable && !isActorOnTile && !isObjectBlocking
}

// BresenhamLine returns a list of points that are on the line between source and destination.
// NOTE: Will remove the source point from the list
func (m *GridMap[ActorType, ItemType, ObjectType]) BresenhamLine(source geometry.Point, destination geometry.Point, isBlocking func(mapPos geometry.Point) bool) []geometry.Point {
	los := geometry.BresenhamLine(source, destination, func(x, y int) bool {
		p := geometry.Point{X: x, Y: y}
		if !m.Contains(p) {
			return false
		}
		if p == source {
			return true
		}
		if isBlocking == nil && isBlocking(p) {
			return false
		}
		return m.IsWalkable(p)
	})
	withoutSource := los[1:]
	return withoutSource
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetAllExplored() {
	for y := 0; y < m.mapHeight; y++ {
		for x := 0; x < m.mapWidth; x++ {
			m.cells[y*m.mapWidth+x].IsExplored = true
		}
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RandomSpawnPosition() geometry.Point {
	for {
		x := rand.Intn(m.mapWidth)
		y := rand.Intn(m.mapHeight)
		pos := geometry.Point{X: x, Y: y}
		if m.IsEmptyNonSpecialFloor(pos) {
			return pos
		}
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddNamedLocation(name string, point geometry.Point) {
	m.namedLocations[name] = point
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNamedLocation(name string) geometry.Point {
	return m.namedLocations[name]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNamedLocationByPos(pos geometry.Point) string {
	for name, location := range m.namedLocations {
		if location == pos {
			return name
		}
	}
	return ""
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RenameLocation(oldName string, newName string) {
	if pos, ok := m.namedLocations[oldName]; ok {
		delete(m.namedLocations, oldName)
		m.namedLocations[newName] = pos
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveNamedLocation(namedLocation string) {
	delete(m.namedLocations, namedLocation)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsDamagingTileAt(p geometry.Point) bool {
	tileType := m.CellAt(p).TileType
	return tileType.IsDamaging
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RandomPosAround(pos geometry.Point) geometry.Point {
	neighbors := m.NeighborsAll(pos, func(p geometry.Point) bool {
		return m.Contains(p)
	})
	if len(neighbors) == 0 {
		return pos
	}
	neighbors = append(neighbors, pos)
	return neighbors[rand.Intn(len(neighbors))]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) TryGetActorAt(pos geometry.Point) (ActorType, bool) {
	var noActor ActorType
	isActorAt := m.IsActorAt(pos)
	if !isActorAt {
		return noActor, false
	}
	return m.ActorAt(pos), isActorAt
}

func (m *GridMap[ActorType, ItemType, ObjectType]) TryGetDownedActorAt(pos geometry.Point) (ActorType, bool) {
	var noActor ActorType
	isDownedActorAt := m.IsDownedActorAt(pos)
	if !isDownedActorAt {
		return noActor, false
	}
	return m.DownedActorAt(pos), isDownedActorAt
}

func (m *GridMap[ActorType, ItemType, ObjectType]) TryGetObjectAt(pos geometry.Point) (ObjectType, bool) {
	var noObject ObjectType
	isObjectAt := m.IsObjectAt(pos)
	if !isObjectAt {
		return noObject, false
	}
	return m.ObjectAt(pos), isObjectAt
}

func (m *GridMap[ActorType, ItemType, ObjectType]) TryGetItemAt(pos geometry.Point) (ItemType, bool) {
	var noItem ItemType
	isItemAt := m.IsItemAt(pos)
	if !isItemAt {
		return noItem, false
	}
	return m.ItemAt(pos), isItemAt
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddActor(actor ActorType, spawnPos geometry.Point) {
	var nilActor ActorType
	if actor == nilActor {
		return
	}
	if m.IsActorAt(spawnPos) {
		return
	}
	m.allActors = append(m.allActors, actor)
	m.MoveActor(actor, spawnPos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddDownedActor(actor ActorType, spawnPos geometry.Point) {
	var nilActor ActorType
	if actor == nilActor {
		return
	}
	if m.IsDownedActorAt(spawnPos) {
		return
	}
	m.allDownedActors = append(m.allDownedActors, actor)
	m.MoveDownedActor(actor, spawnPos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddObject(object ObjectType, spawnPos geometry.Point) {
	var nilObject ObjectType
	if object == nilObject {
		return
	}
	if m.IsObjectAt(spawnPos) {
		return
	}
	m.allObjects = append(m.allObjects, object)
	m.MoveObject(object, spawnPos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddItem(item ItemType, spawnPos geometry.Point) {
	var nilItem ItemType
	if item == nilItem {
		return
	}
	if m.IsItemAt(spawnPos) {
		return
	}
	m.allItems = append(m.allItems, item)
	m.MoveItem(item, spawnPos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) UpdateFieldOfView(fov *geometry.FOV, fovPosition geometry.Point, visionRange int) {
	visionRangeSquared := visionRange * visionRange
	var fovRange = geometry.NewRect(-visionRange, -visionRange, visionRange+1, visionRange+1)
	fov.SetRange(fovRange.Add(fovPosition).Intersect(geometry.NewRect(0, 0, m.mapWidth, m.mapHeight)))

	fov.SSCVisionMap(fovPosition, visionRange, false, func(p geometry.Point) bool {
		if !m.Contains(p) {
			return false
		}
		return m.IsTransparent(p) && geometry.DistanceSquared(p, fovPosition) <= visionRangeSquared
	})
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ToggleNoClip() bool {
	m.noClip = !m.noClip
	return m.noClip
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetSecretDoorAt(pos geometry.Point) {
	m.secretDoors[pos] = true
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsSecretDoorAt(neighbor geometry.Point) bool {
	if _, ok := m.secretDoors[neighbor]; ok {
		return true
	}
	return false
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetName(name string) {
	m.name = name
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddTransitionAt(pos geometry.Point, transition Transition) {
	m.transitionMap[pos] = transition
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetTransitionAt(pos geometry.Point) (Transition, bool) {
	transition, ok := m.transitionMap[pos]
	return transition, ok

}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetName() string {
	return m.name
}

func (m *GridMap[ActorType, ItemType, ObjectType]) WriteTiles(out io.Writer) {
	for _, cell := range m.cells {
		cell.TileType.ToBinary(out)
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Transitions() map[geometry.Point]Transition {
	return m.transitionMap
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SecretDoors() map[geometry.Point]bool {
	return m.secretDoors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddNamedRegion(name string, region geometry.Rect) {
	m.namedRects[name] = region
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNamedRegion(name string) geometry.Rect {
	return m.namedRects[name]
}

type Trigger struct {
	Name    string
	Bounds  geometry.Rect
	OneShot bool
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNamedTriggerAt(pos geometry.Point) (Trigger, bool) {
	for _, trigger := range m.namedTrigger {
		if trigger.Bounds.Contains(pos) {
			return trigger, true
		}
	}
	return Trigger{}, false
}
func (m *GridMap[ActorType, ItemType, ObjectType]) SetTileIcon(pos geometry.Point, index textiles.TextIcon) {
	m.cells[pos.Y*m.mapWidth+pos.X].TileType.Icon = index
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetTileIconAt(pos geometry.Point) textiles.TextIcon {
	return m.cells[pos.Y*m.mapWidth+pos.X].TileType.Icon
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveNamedRegion(regionName string) {
	delete(m.namedRects, regionName)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveNamedTrigger(triggerName string) {
	delete(m.namedTrigger, triggerName)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddNamedTrigger(name string, rect Trigger) {
	m.namedTrigger[name] = rect
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetDisplayName(name string) {
	m.displayName = name
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetDisplayName() string {
	return m.displayName
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredActorsInRadius(location geometry.Point, radius int, filter func(actor ActorType) bool) []ActorType {
	return m.iterateActors(location, radius, filter)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) iterateActors(location geometry.Point, radius int, filter func(actor ActorType) bool) []ActorType {
	result := make([]ActorType, 0)
	for _, actor := range m.allActors {
		if geometry.Distance(location, actor.Position()) <= float64(radius) && filter(actor) {
			result = append(result, actor)
		}
	}
	return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddNamedPath(name string, pathFromRootEntity []geometry.Point) {
	m.namedPaths[name] = pathFromRootEntity
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNamedPath(name string) []geometry.Point {
	return m.namedPaths[name]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetNamedLocations(locations []NamedLocation) {
	m.namedLocations = make(map[string]geometry.Point)
	for _, location := range locations {
		m.namedLocations[location.LocationName] = location.Pos
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetRandomLocation(filter func(location geometry.Point) bool) (geometry.Point, bool) {
	endCounter := 1000000
	for {
		x := rand.Intn(m.mapWidth)
		y := rand.Intn(m.mapHeight)
		pos := geometry.Point{X: x, Y: y}
		if filter(pos) {
			return pos, true
		}
		endCounter--
		if endCounter <= 0 {
			return geometry.Point{}, false
		}
	}
}
func (m *GridMap[ActorType, ItemType, ObjectType]) Print() {
	// walls and floors only
	for y := 0; y < m.mapHeight; y++ {
		for x := 0; x < m.mapWidth; x++ {
			pos := geometry.Point{X: x, Y: y}
			if m.IsTileWalkable(pos) {
				fmt.Printf(".")
			} else {
				fmt.Printf("#")
			}
		}
		fmt.Printf("\n")
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) PrintWithHighlight(hiPos geometry.Point) {
	// walls and floors only
	for y := 0; y < m.mapHeight; y++ {
		for x := 0; x < m.mapWidth; x++ {
			pos := geometry.Point{X: x, Y: y}
			if pos == hiPos {
				fmt.Printf("X")
			} else if m.IsActorAt(pos) {
				fmt.Printf("a")
			} else if m.IsTileWalkable(pos) {
				fmt.Printf(".")
			} else {
				fmt.Printf("#")
			}
		}
		fmt.Printf("\n")
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) CanPlaceActorHere(pos geometry.Point) bool {
	return m.IsWalkable(pos) && !m.IsActorAt(pos) && !m.IsObviousHazardAt(pos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) CanPlaceItemHere(pos geometry.Point) bool {
	return m.IsWalkable(pos) && !m.IsItemAt(pos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetMetaString(info []string) {
	m.metaInfoString = info
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetMetaString() []string {
	return m.metaInfoString
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddToMetaInfo(infoLine string) {
	m.metaInfoString = append(m.metaInfoString, infoLine)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredActors(f func(actor ActorType) bool) []ActorType {
	var result []ActorType
	for _, actor := range m.allActors {
		if f(actor) {
			result = append(result, actor)
		}
	}
	return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) MoveDistance(pOne geometry.Point, pTwo geometry.Point) int {
	if m.cardinalMovementOnly {
		return geometry.DistanceManhattan(pOne, pTwo)
	}
	return geometry.DistanceChebyshev(pOne, pTwo)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) ApplyToActorsAt(positions []geometry.Point, applyFunc func(actor ActorType)) {
	for _, pos := range positions {
		if !m.IsActorAt(pos) {
			continue
		}
		applyFunc(m.ActorAt(pos))
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RayCast(origin, direction geometry.PointF, isBlockingRay func(geometry.Point) bool) fxtools.HitInfo2D {
	direction = direction.Normalize()
	hitInfo := fxtools.Raycast2D(origin.X, origin.Y, direction.X, direction.Y, func(x, y int64) bool {
		currentMapCell := geometry.Point{X: int(x), Y: int(y)}
		if currentMapCell == origin.ToPoint() {
			return false
		}
		return isBlockingRay(currentMapCell)
	})
	return hitInfo
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ReflectingRayCast(origin, direction geometry.PointF, maxReflections int, isBlockingRay func(geometry.Point) bool) []fxtools.HitInfo2D {
	direction = direction.Normalize()
	hitInfos := fxtools.ReflectingRaycast2D(origin.X, origin.Y, direction.X, direction.Y, maxReflections, func(x, y int64) bool {
		currentMapCell := geometry.Point{X: int(x), Y: int(y)}
		return isBlockingRay(currentMapCell)
	})
	return hitInfos
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ChainedRayCast(origin, direction geometry.PointF, isBlockingRay func(geometry.Point) bool, nextTarget func(geometry.Point) (bool, geometry.Point)) []fxtools.HitInfo2D {
	direction = direction.Normalize()
	hitInfos := fxtools.ChainedRaycast2D(origin.X, origin.Y, direction.X, direction.Y, func(x, y int64) bool {
		currentMapCell := geometry.Point{X: int(x), Y: int(y)}
		return isBlockingRay(currentMapCell)

	}, func(x, y int64) (bool, int, int) {
		currentMapCell := geometry.Point{X: int(x), Y: int(y)}
		target, point := nextTarget(currentMapCell)
		return target, point.X, point.Y
	})
	return hitInfos
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddDecal(pos geometry.Point, decalIcon int32) {
	m.decals[pos] = decalIcon
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetDecal(pos geometry.Point) (int32, bool) {
	decal, ok := m.decals[pos]
	if !ok {
		return -1, false
	}
	return decal, true
}
func (m *GridMap[ActorType, ItemType, ObjectType]) JumpOverPositions(origin geometry.Point, target geometry.Point, maxSprint, maxJump int) JumpOverInfo {
	direction := target.Sub(origin)
	beyondTargetPos := target.Add(direction)
	lineOfSprint := m.BresenhamLine(origin, beyondTargetPos, func(pos geometry.Point) bool {
		return !m.Contains(pos)
	})

	var sprint []geometry.Point
	var jump []geometry.Point
	var jumpedPos geometry.Point
	actorFound := false

	for _, pos := range lineOfSprint {
		if m.IsCurrentlyPassable(pos) {
			if actorFound {
				jump = append(jump, pos)
				if len(jump) >= maxJump {
					break
				}
			} else {
				sprint = append(sprint, pos)
				if len(sprint) >= maxSprint {
					break
				}
			}
		} else if m.IsActorAt(pos) {
			jumpedPos = pos
			actorFound = true
		}
	}
	return JumpOverInfo{
		Sprint:     sprint,
		Jump:       jump,
		JumpedPos:  jumpedPos,
		ActorFound: actorFound,
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SprintToPosition(origin geometry.Point, target geometry.Point, maxSprint int) SprintToInfo {
	lineOfSprint := m.BresenhamLine(origin, target, func(pos geometry.Point) bool { return !m.Contains(pos) })

	var sprint []geometry.Point
	var sprintEndPos geometry.Point
	var emptyTileFound bool

	for _, pos := range lineOfSprint {
		if !m.IsActorAt(pos) && m.IsCurrentlyPassable(pos) {
			sprint = append(sprint, pos)
			if len(sprint) >= maxSprint {
				break
			}
		} else {
			break
		}
	}

	if len(sprint) <= maxSprint && len(sprint) > 0 {
		emptyTileFound = true
		sprintEndPos = sprint[len(sprint)-1]
	}

	return SprintToInfo{
		Sprint:         sprint,
		SprintEndPos:   sprintEndPos,
		EmptyTileFound: emptyTileFound,
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetWidth() int {
	return m.mapWidth
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetHeight() int {
	return m.mapHeight
}

func GetLocationsInRadius(origin geometry.Point, radius float64, keep func(point geometry.Point) bool) []geometry.Point {
	result := make([]geometry.Point, 0)
	for y := origin.Y - int(radius); y <= origin.Y+int(radius); y++ {
		for x := origin.X - int(radius); x <= origin.X+int(radius); x++ {
			pos := geometry.Point{X: x, Y: y}
			if geometry.Distance(origin, pos) <= radius && keep(pos) {
				result = append(result, pos)
			}
		}
	}
	return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsValidSprintTarget(attacker geometry.Point, target geometry.Point, maxSprint int) bool {
	sprintInfo := m.SprintToPosition(attacker, target, maxSprint)
	isValid := false
	for _, pos := range sprintInfo.Sprint {
		if pos == target {
			isValid = true
			break
		}
	}
	return isValid
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsBorderWall(point geometry.Point) bool {
	// pos is on the edge of the map
	if point.X == 0 || point.X == m.mapWidth-1 || point.Y == 0 || point.Y == m.mapHeight-1 {
		return true
	}
	return false
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsDownedActor(actor ActorType) bool {
	for _, downedActor := range m.allDownedActors {
		if downedActor == actor {
			return true
		}
	}
	return false
}

func (m *GridMap[ActorType, ItemType, ObjectType]) CanPlaceObjectHere(pos geometry.Point) bool {
	return m.IsWalkable(pos) && !m.IsObjectAt(pos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsEmptyNonSpecialFloor(pos geometry.Point) bool {
	return m.Contains(pos) && m.IsTileWalkable(pos) && !m.IsActorAt(pos) && !m.IsDownedActorAt(pos) && !m.IsItemAt(pos) && !m.IsObjectAt(pos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredAdjacentPositions(positions []geometry.Point, keep func(pos geometry.Point) bool) []geometry.Point {
	results := make(map[geometry.Point]bool, 0)

	for _, pos := range positions {
		neighbors := m.NeighborsAll(pos, m.Contains)
		for _, neighbor := range neighbors {
			if keep(neighbor) {
				results[neighbor] = true
			}
		}
	}

	result := make([]geometry.Point, len(results))
	index := 0
	for pos, _ := range results {
		result[index] = pos
		index++
	}
	return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetSelfAndAllNeighborsExplored(position geometry.Point) {
	neighbors := m.NeighborsAll(position, m.Contains)
	for _, neighbor := range neighbors {
		m.SetExplored(neighbor)
	}
	m.SetExplored(position)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetSelfAndAllWalkableNeighborsExplored(position geometry.Point) {
	neighbors := m.NeighborsAll(position, m.Contains)
	for _, neighbor := range neighbors {
		if !m.IsTileWalkable(neighbor) {
			continue
		}
		m.SetExplored(neighbor)
	}
	m.SetExplored(position)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetListExplored(tiles []geometry.Point, value bool) {
	for _, tile := range tiles {
		m.SetExplored(tile)
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetMoveOnPlayerDijkstraMap(from geometry.Point, towardsPlayer bool, dijkstraMap map[geometry.Point]int) geometry.Point {
	if dijkstraMap == nil {
		return from
	}
	currentDistanceToPlayer, exist := dijkstraMap[from]
	if !exist {
		return from
	}
	var compareFunc func(a, b int) bool
	if towardsPlayer {
		compareFunc = func(a, b int) bool { return a < b }
	} else {
		compareFunc = func(a, b int) bool { return a > b }
	}
	currentMap := m
	neighbors := currentMap.GetFilteredNeighborsForMovement(from, func(pos geometry.Point) bool { // choose a possible next step
		if !currentMap.IsCurrentlyPassable(pos) { // only walk on passable tiles
			return false
		}
		if _, existsTransition := currentMap.GetTransitionAt(pos); existsTransition { // don't walk on transitions
			return false
		}
		neighborDist, neighborExists := dijkstraMap[pos]
		if !neighborExists {
			return false
		}
		neighborsWithTransition := currentMap.GetFilteredNeighbors(pos, func(pos geometry.Point) bool { // in fact, don't walk on tiles next to transitions either
			if !currentMap.IsCurrentlyPassable(pos) {
				return false
			}
			if _, existsTransition := currentMap.GetTransitionAt(pos); existsTransition {
				return true
			}
			return false
		})
		if len(neighborsWithTransition) > 0 {
			return false
		}
		return compareFunc(neighborDist, currentDistanceToPlayer) // depends on rolling direction on our dijkstra map
		//return neighborDist < currentDistanceToPlayer // depends on rolling direction on our dijkstra map
	})

	if len(neighbors) == 0 {
		return from
	}
	// we found some locations we can move to, that also bring us closer to the player
	// which of these is the closest to the player?

	nearestDist := math.MaxInt
	if !towardsPlayer {
		nearestDist = 0
	}
	nearestPos := from

	for _, neighbor := range neighbors {
		neighborDist, _ := dijkstraMap[neighbor]
		//if neighborDist < nearestDist {  // depends on rolling direction on our dijkstra map
		if compareFunc(neighborDist, nearestDist) {
			nearestDist = neighborDist
			nearestPos = neighbor
		}
	}

	return nearestPos
}

func (m *GridMap[ActorType, ItemType, ObjectType]) HasWalkableNeighbor(point geometry.Point) bool {
	neighbors := m.NeighborsAll(point, m.IsTileWalkable)
	return len(neighbors) > 0
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsLineOfSightClear(source geometry.Point, dest geometry.Point) bool {
	direction := dest.Sub(source).ToCenteredPointF()
	hitInfo := m.RayCast(source.ToCenteredPointF(), direction, func(point geometry.Point) bool {
		if point == source {
			return true
		}
		if point == dest {
			return true
		}
		return !m.IsCurrentlyPassable(point)
	})
	return int(hitInfo.ColliderGridPosition[0]) == dest.X && int(hitInfo.ColliderGridPosition[1]) == dest.Y
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddActorWithDisplacement(actor ActorType, position geometry.Point) {
	if m.CanPlaceActorHere(position) {
		m.AddActor(actor, position)
	} else {
		m.allActors = append(m.allActors, actor)
		m.displaceActor(actor, position)
	}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ForceSpawnActorInWall(actor ActorType, to geometry.Point) {
	m.allActors = append(m.allActors, actor)
	actor.SetPosition(to)
	m.cells[to.X+to.Y*m.mapWidth] = m.cells[to.X+to.Y*m.mapWidth].WithActor(actor)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFirstWallCardinalInDirection(origin geometry.Point, dir geometry.CompassDirection) geometry.Point {
	for i := 1; i < max(m.mapWidth, m.mapHeight); i++ {
		pos := origin.Add(dir.ToPoint().Mul(i))
		if !m.Contains(pos) {
			return pos
		}
		if !m.IsTileWalkable(pos) {
			return pos
		}
	}
	return origin
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsTransitionAt(position geometry.Point) bool {
	_, exists := m.GetTransitionAt(position)
	return exists
}

func (m *GridMap[ActorType, ItemType, ObjectType]) LightAt(p geometry.Point, timeOfDay time.Time) fxtools.HDRColor {
	if m.isIndoor {
		return m.IndoorLightAt(p)
	}
	return m.OutdoorLightAt(p, timeOfDay)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetMeta(data MapMeta) {
	m.meta = data
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetMeta() MapMeta {
	return m.meta
}

func (m *GridMap[ActorType, ItemType, ObjectType]) generateTileSetAndMap() ([]Tile, []int16) {
	tileSet := make([]Tile, 0)
	tileMap := make([]int16, m.mapWidth*m.mapHeight)
	for i := 0; i < len(m.cells); i++ {
		tile := m.cells[i].TileType
		tileIndex := -1
		for k, existingTile := range tileSet {
			if existingTile == tile {
				tileIndex = k
				break
			}
		}
		if tileIndex == -1 {
			tileIndex = len(tileSet)
			tileSet = append(tileSet, tile)
		}
		tileMap[i] = int16(tileIndex)
	}
	return tileSet, tileMap

}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetCells(cells []MapCell[ActorType, ItemType, ObjectType]) {
	var allItems []ItemType
	var allObjects []ObjectType
	var allActors []ActorType
	var allDownedActors []ActorType
	for _, cell := range cells {
		if cell.Actor != nil {
			allActors = append(allActors, *cell.Actor)
		}
		if cell.DownedActor != nil {
			allDownedActors = append(allDownedActors, *cell.DownedActor)
		}
		if cell.Item != nil {
			allItems = append(allItems, *cell.Item)
		}
		if cell.Object != nil {
			allObjects = append(allObjects, *cell.Object)
		}
	}
	m.allItems = allItems
	m.allObjects = allObjects
	m.allActors = allActors
	m.allDownedActors = allDownedActors
	m.cells = cells
}

type JumpOverInfo struct {
	Sprint     []geometry.Point
	Jump       []geometry.Point
	JumpedPos  geometry.Point
	ActorFound bool
}

type SprintToInfo struct {
	Sprint         []geometry.Point
	SprintEndPos   geometry.Point
	EmptyTileFound bool
}
