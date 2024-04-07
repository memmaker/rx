package dungen

import (
	"RogueUI/geometry"
	"RogueUI/util"
	"cmp"
	"fmt"
	"math/rand"
	"slices"
)

type TemplateRules struct {
	RequireTemplates map[string]bool
	ForbidTemplates  map[string]bool
	TemplatePrefix   string
	NoDoors          bool
}

func (r TemplateRules) WithNoDoors() TemplateRules {
	r.NoDoors = true
	return r
}

func (r TemplateRules) WithForbidden(templateName string) TemplateRules {
	r.ForbidTemplates[templateName] = true
	return r
}

func NewTemplateRules(requiredRooms map[int][]string, forLevel int, templatePrefix string) TemplateRules {
	rules := TemplateRules{
		RequireTemplates: make(map[string]bool),
		ForbidTemplates:  make(map[string]bool),
		TemplatePrefix:   templatePrefix,
	}
	for level, rooms := range requiredRooms {
		if level == forLevel {
			for _, room := range rooms {
				rules.RequireTemplates[room] = true
			}
		} else {
			for _, room := range rooms {
				rules.ForbidTemplates[room] = true
			}
		}

	}
	return rules
}

type RoomConnection struct {
	OutwardDirection geometry.CompassDirection
	ConnectionID     int
}
type AccretionGenerator struct {
	random          *rand.Rand
	roomTemplates   []*DungeonRoom
	maxRoomCount    int
	printDebugSteps bool
	currentRoomId   int
	levelRules      TemplateRules
	placedTemplates map[string]bool
}

func NewAccretionGenerator(randomSource *rand.Rand) *AccretionGenerator {
	return &AccretionGenerator{
		random:        randomSource,
		currentRoomId: 0,
	}
}

func (g *AccretionGenerator) getTemplatesForLevel() []*DungeonRoom {
	var templates []*DungeonRoom
	for _, room := range g.roomTemplates {
		if !g.levelRules.ForbidTemplates[room.GetTemplateName()] {
			templates = append(templates, room)
		}
	}
	return templates
}

func (g *AccretionGenerator) getNewRoomFromRandomTemplateForSlot(slotID int) *DungeonRoom {
	var matchingRooms []*DungeonRoom

	missingTemplates := make(map[string]bool)
	for reqRoom := range g.levelRules.RequireTemplates {
		if _, ok := g.placedTemplates[reqRoom]; !ok {
			missingTemplates[reqRoom] = true
		}
	}

	if len(missingTemplates) > 0 { // force one of the missing templates
		missingSlotIds := make(map[int]bool)
		for _, room := range g.getTemplatesForLevel() {
			if _, ok := missingTemplates[room.GetTemplateName()]; !ok {
				continue
			}
			if room.GetPlugsIntoSlotId() == slotID {
				matchingRooms = append(matchingRooms, room)
			} else {
				missingSlotIds[room.GetPlugsIntoSlotId()] = true
			}
		}
		if len(matchingRooms) == 0 {
			// probably a missing template is not available for this slot
			// see what slot we need and try to find a room that provides it
			for _, room := range g.getTemplatesForLevel() {
				if _, ok := missingSlotIds[room.GetSlotId()]; !ok {
					continue
				}
				if room.GetPlugsIntoSlotId() == slotID {
					matchingRooms = append(matchingRooms, room)
				}
			}
		}
	} else { // otherwise, just pick any room
		for _, room := range g.getTemplatesForLevel() {
			if room.GetPlugsIntoSlotId() == slotID {
				if _, ok := g.placedTemplates[room.GetTemplateName()]; ok {
					if _, ok := g.levelRules.RequireTemplates[room.GetTemplateName()]; ok {
						continue
					}
				}
				matchingRooms = append(matchingRooms, room)
			}
		}
	}

	if len(matchingRooms) == 0 {
		println("no matching rooms for slot " + fmt.Sprintf("%d", slotID))
		for _, room := range g.getTemplatesForLevel() {
			println("room " + room.GetTemplateName() + " slot " + fmt.Sprintf("%d", room.GetPlugsIntoSlotId()))
		}
		panic("no matching rooms for slot " + fmt.Sprintf("%d", slotID))
	}

	randomTemplate := matchingRooms[g.random.Intn(len(matchingRooms))]
	clonedRoom := randomTemplate.Clone()
	return clonedRoom
}

func (g *AccretionGenerator) getFirstRoom() *DungeonRoom {
	return g.getNewRoomFromRandomTemplateWithZeroIDs()
}
func (g *AccretionGenerator) getNewRoomFromRandomTemplateWithZeroIDs() *DungeonRoom {
	var subsetWithZeroIDConnections []*DungeonRoom
	for _, room := range g.getTemplatesForLevel() {
		if room.HasOnlyZeroIDConnections() {
			subsetWithZeroIDConnections = append(subsetWithZeroIDConnections, room)
		}
	}
	randomTemplate := subsetWithZeroIDConnections[g.random.Intn(len(subsetWithZeroIDConnections))]
	clonedRoom := randomTemplate.Clone()
	return clonedRoom
}

func (g *AccretionGenerator) nextRoomForGenerator(slotID int) *DungeonRoom {
	var newRoom *DungeonRoom
	newRoom = g.getNewRoomFromRandomTemplateForSlot(slotID)
	newRoom.SetId(g.currentRoomId)
	g.currentRoomId++
	return newRoom
}

func (g *AccretionGenerator) Generate(width, height int) *DungeonMap {
	g.placedTemplates = make(map[string]bool)
	dMap := NewDungeonMap(width, height)

	firstRoom := g.getFirstRoom()
	firstRoom.SetId(g.currentRoomId)
	g.currentRoomId++

	roomSize := firstRoom.GetRelativeBoundingRectIncludingConnectors().Size()
	centerMapPos := geometry.Point{X: width - roomSize.X, Y: height - roomSize.Y}.Div(2)
	firstRoom.SetPositionOffset(centerMapPos)

	if !dMap.CanPlaceRoom(firstRoom) {
		panic("first room cannot be placed") // TODO
	}

	dMap.AddRoomAndSetTiles(firstRoom)
	g.placedTemplates[firstRoom.GetTemplateName()] = true

	if g.printDebugSteps {
		println("=== first room ===")
		dMap.Print()
	}
	maxRooms := g.maxRoomCount
	repCounter := 0
	currentSlotID := 0
	for len(dMap.rooms) < maxRooms && repCounter < 100 {
		nextRoom := g.nextRoomForGenerator(currentSlotID)
		if g.printDebugSteps {
			println(fmt.Sprintf("=== room %d ===", nextRoom.roomId))
			nextRoom.PrintUntransformed()
		}
		for _, existingRoom := range dMap.rooms {
			if g.tryConnectTo(dMap, nextRoom, existingRoom) {
				repCounter = 0
				dMap.AddRoomAndSetTiles(nextRoom)
				g.placedTemplates[nextRoom.GetTemplateName()] = true
				currentSlotID = nextRoom.GetSlotId()
				if g.printDebugSteps {
					println(fmt.Sprintf("=== added room %d ===", nextRoom.roomId))
					dMap.PrintPlannedDoors()
				}
				break
			}
		}

		repCounter++
	}
	if len(dMap.rooms) < maxRooms {
		println("=== COULD NOT PLACE ALL ROOMS ===")
	}

	if g.printDebugSteps {
		println("=== BEFORE TRIM ===")
		dMap.PrintPlannedDoors()
		dMap.Trim()
		println("=== AFTER TRIM ===")
		dMap.PrintPlannedDoors()
	}

	if g.levelRules.NoDoors {
		dMap.ConnectRoomsOnMap(false)
	} else {
		dMap.ConnectRoomsOnMap(true)
		dMap.addMoreDoors(g.random, 10, 0.8)
	}

	//dMap.FillDeadEnds(g.random)

	return dMap
}

func (g *AccretionGenerator) randomRect(width int, height int) geometry.Rect {
	randWidth := max(3, g.random.Intn(width))
	randHeight := max(3, g.random.Intn(height))
	rect := geometry.NewRect(0, 0, randWidth, randHeight)
	return rect
}

func (g *AccretionGenerator) tryConnectTo(dMap *DungeonMap, newRoom *DungeonRoom, existingRoom *DungeonRoom) bool {
	if g.printDebugSteps {
		println(fmt.Sprintf("=== try connect room %d to room %d ===", newRoom.roomId, existingRoom.roomId))
	}
	freeConnections := existingRoom.GetFreeConnectionsWithRotatedDirection()
	relPositions := getSortedKeys(freeConnections)
	randomRotationOrder := g.random.Perm(4)
	for rotationCount := 0; rotationCount < 4; rotationCount++ {
		newRoom.SetRotationCount(randomRotationOrder[rotationCount])
		if g.printDebugSteps {
			println(fmt.Sprintf("=== rotation %d for room %d ===", rotationCount, newRoom.roomId))
			newRoom.PrintRotated()
		}
		for _, relativeConnectionPos := range relPositions {
			roomConnection := freeConnections[relativeConnectionPos]
			if relativeNewDoorPos, ok := newRoom.HasMatchingConnection(roomConnection); ok {
				rotatedRelativeDoorPos := newRoom.GetRotatedRelativeDoorPosition(relativeNewDoorPos)
				// both have free doors in opposite directions
				// PROBLEM IS HERE-> we don't set the position offset for the new room correctly
				// probably only occurs when the new room is rotated
				absoluteDoorPos := existingRoom.GetAbsoluteDoorPosition(relativeConnectionPos)

				// the new room should be placed so that the new door is at the same position as the existing door
				newRoom.SetPositionOffset(absoluteDoorPos.Sub(rotatedRelativeDoorPos))
				if g.printDebugSteps {
					println(fmt.Sprintf("=== CONNECTION POSSIBLE ==="))
					println(fmt.Sprintf("> NEW ROOM %d", newRoom.roomId))
					newRoom.PrintTransformedWithHighlight(relativeNewDoorPos)
					println(fmt.Sprintf("> EXISTING ROOM %d", existingRoom.roomId))
					existingRoom.PrintTransformedWithHighlight(relativeConnectionPos)
				}
				if dMap.CanPlaceRoom(newRoom) {
					if g.printDebugSteps {
						println(fmt.Sprintf("=== CONNECTED ==="))
					}
					newRoom.AddConnectedRoom(relativeNewDoorPos, existingRoom)
					existingRoom.AddConnectedRoom(relativeConnectionPos, newRoom)
					return true
				} else if newRoom.GetPlugsIntoSlotId() > 0 {
					println(fmt.Sprintf("=== COULD NOT FIT ==="))
					println(fmt.Sprintf("> NEW ROOM %d", newRoom.roomId))
					newRoom.PrintTransformedWithHighlight(relativeNewDoorPos)
					println(fmt.Sprintf("> EXISTING ROOM %d", existingRoom.roomId))
					existingRoom.PrintTransformedWithHighlight(relativeConnectionPos)
					println(fmt.Sprintf("> MAP"))
					dMap.PrintWithHighlight(absoluteDoorPos)
				}
			}
		}
	}

	return false
}

func getSortedKeys(doors map[geometry.Point]RoomConnection) []geometry.Point {
	keys := make([]geometry.Point, len(doors))
	i := 0
	for k := range doors {
		keys[i] = k
		i++
	}
	slices.SortStableFunc(keys, func(i, j geometry.Point) int {
		idxI := util.XYToIndex(i.X, i.Y, 20)
		idxJ := util.XYToIndex(j.X, j.Y, 20)
		return cmp.Compare(idxI, idxJ)
	})
	return keys
}

func (g *AccretionGenerator) SetRoomCount(maxRoomCount int) {
	g.maxRoomCount = maxRoomCount
}

func (g *AccretionGenerator) SetDebugSteps(debug bool) {
	g.printDebugSteps = debug
}

func (g *AccretionGenerator) SetLevelRules(rules TemplateRules) {
	g.levelRules = rules
}

func compassDirectionFromId(identifier string) geometry.CompassDirection {
	switch identifier {
	case "NorthConnector":
		return geometry.North
	case "SouthConnector":
		return geometry.South
	case "EastConnector":
		return geometry.East
	case "WestConnector":
		return geometry.West
	}
	panic("unknown connector identifier: " + identifier)
}
