package game

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/gridmap"
	"RogueUI/recfile"
	"RogueUI/special"
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type MapLoader interface {
	LoadMap(mapName string) *gridmap.GridMap[*Actor, *Item, Object]
}

type TextMapLoader struct {
	gameState *GameState
	random    *rand.Rand
}

func NewTextMapLoader(g *GameState) *TextMapLoader {
	return &TextMapLoader{gameState: g, random: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (t *TextMapLoader) LoadMap(levelName string) *gridmap.GridMap[*Actor, *Item, Object] {
	g := t.gameState
	rawMapData := ReadMapFile(path.Join(g.config.DataRootDir, "maps", levelName+".txt"))
	if rawMapData.IsEmpty() {
		return nil
	}

	didInitFlagName := fmt.Sprintf("DidInitLevel(%s)", levelName)
	actorRecords := make(map[string]recfile.Record)
	objectRecords := make(map[string]recfile.Record)
	if !g.gameFlags.HasFlag(didInitFlagName) {
		for _, flagLine := range rawMapData.Flags {
			parts := strings.Split(flagLine, "=")
			if len(parts) != 2 {
				continue
			}
			flagName := strings.TrimSpace(parts[0])
			flagValue, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			g.gameFlags.Set(flagName, flagValue)
		}
		g.gameFlags.SetFlag(didInitFlagName)

		actorRecords = rawMapData.ActorRecords
		objectRecords = rawMapData.ObjectRecords
	}

	mapping := t.generateTileMapping(rawMapData.Legend, actorRecords, objectRecords)
	simpleMapper := func(gridMap *gridmap.GridMap[*Actor, *Item, Object], icon rune, pos geometry.Point) {
		mapFunc, exists := mapping[icon]
		if !exists {
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileFloor,
				DefinedDescription: "floor",
				IsWalkable:         true,
				IsTransparent:      true,
			})
			fmt.Fprintf(os.Stderr, "No mapping function for %c\n", icon)
			return
		}
		mapFunc(g, gridMap, pos)
	}

	newMap := gridmap.NewMapFromString[*Actor, *Item, Object](rawMapData.Dimensions.X, rawMapData.Dimensions.Y, rawMapData.MapData, simpleMapper)
	newMap.SetCardinalMovementOnly(!g.config.DiagonalMovementEnabled)
	newMap.SetName(levelName)
	return newMap
}

func (t *TextMapLoader) generateTileMapping(legend []string, actorRecords map[string]recfile.Record, objectRecords map[string]recfile.Record) map[rune]func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
	mapping := make(map[rune]func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point))
	for _, line := range legend {
		if len(line) == 0 {
			continue
		}
		if !strings.Contains(line, "=") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		keyString := strings.TrimSpace(parts[0])
		key := []rune(keyString)
		icon := key[0]
		if len(key) > 1 && keyString == "<space>" {
			icon = ' '
		}

		function := t.getMappingFunction(mapping, actorRecords, objectRecords, strings.TrimSpace(parts[1]))
		if function == nil {
			fmt.Fprintf(os.Stderr, "No mapping function for %s\n", parts[1])
			continue
		}
		mapping[icon] = function
	}
	return mapping

}

func (t *TextMapLoader) getMappingFunction(currentMappings map[rune]func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point), actorRecords map[string]recfile.Record, objectRecords map[string]recfile.Record, mappingDescription string) func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
	parts := strings.Split(mappingDescription, "|")
	typeDesc := strings.ToLower(strings.TrimSpace(parts[0]))
	setDefaultFloor, hasDefaultFloor := currentMappings[' ']
	if !hasDefaultFloor {
		setDefaultFloor = func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
			m.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileFloor,
				DefinedDescription: "floor",
				IsWalkable:         true,
				IsTransparent:      true,
			})
		}

	}
	switch {
	case typeDesc == "tile":
		tileName := strings.TrimSpace(parts[1])
		desc := strings.TrimSpace(parts[2])
		return func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
			mapFeature := foundation.FeatureFromName(tileName)
			tile := gridmap.Tile{
				Feature:            mapFeature,
				DefinedDescription: desc,
				IsWalkable:         mapFeature.IsWalkable(),
				IsTransparent:      mapFeature.IsTransparent(),
			}
			m.SetTile(pos, tile)
		}
	case strings.Contains(typeDesc, "transition"):
		locationName := strings.TrimSpace(parts[1])
		locationDesc := strings.TrimSpace(parts[2])
		targetMap := strings.TrimSpace(parts[3])
		targetLocation := strings.TrimSpace(parts[4])
		return func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
			m.AddTransitionAt(pos, gridmap.Transition{
				TargetMap:      targetMap,
				TargetLocation: targetLocation,
			})
			m.AddNamedLocation(locationName, pos)
			m.SetTile(pos, gridmap.Tile{
				Feature:            foundation.FeatureFromName(typeDesc),
				DefinedDescription: locationDesc,
				IsWalkable:         true,
				IsTransparent:      true,
			})
		}
	case typeDesc == "actor":
		dataDefTemplate := strings.TrimSpace(parts[1])
		internalName := dataDefTemplate
		if len(parts) > 2 {
			internalName = strings.TrimSpace(parts[2])
		}

		return func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
			setDefaultFloor(g, m, pos)
			actorDef := g.dataDefinitions.GetMonsterByName(dataDefTemplate)
			if actorDetailRecord, exists := actorRecords[internalName]; exists {
				actorDef = fillDefinitionFromRecord(actorDef, actorDetailRecord)
			}
			newActor := g.NewEnemyFromDef(actorDef)
			newActor.SetInternalName(internalName)
			if len(parts) > 3 {
				displayName := strings.TrimSpace(parts[3])
				newActor.SetDisplayName(displayName)
			}

			m.AddActor(newActor, pos)
		}
	case typeDesc == "container":
		internalName := strings.TrimSpace(parts[1])
		displayName := strings.TrimSpace(parts[2])

		return func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
			setDefaultFloor(g, m, pos)
			container := g.NewContainer(displayName)
			if containerDetailRecord, exists := objectRecords[internalName]; exists {
				fillContainerFromRecord(g, containerDetailRecord, container)
			}
			m.AddObject(container, pos)
		}
	case typeDesc == "elevator":
		internalName := strings.TrimSpace(parts[1])
		displayName := strings.TrimSpace(parts[2])
		var levels []ElevatorButton
		for _, levelString := range parts[3:] {
			levelParts := strings.Split(levelString, ":")
			if len(levelParts) != 2 {
				continue
			}
			levels = append(levels, ElevatorButton{
				LevelName: strings.TrimSpace(levelParts[0]),
				Label:     strings.TrimSpace(levelParts[1]),
			})
		}
		return func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
			setDefaultFloor(g, m, pos)
			elevator := g.NewElevator(internalName, displayName, levels)
			m.AddObject(elevator, pos)
			m.AddNamedLocation(internalName, pos)
		}
	case typeDesc == "terminal":
		terminalName := strings.TrimSpace(parts[1])
		terminalDesc := strings.TrimSpace(parts[2])
		return func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
			setDefaultFloor(g, m, pos)
			terminal := g.NewTerminal(terminalName)
			terminal.SetDisplayName(terminalDesc)
			m.AddObject(terminal, pos)
		}
	case strings.Contains(typeDesc, "door"):
		doorCat := foundation.ObjectCategoryFromString(typeDesc)
		displayName := strings.TrimSpace(parts[1])
		return func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
			setDefaultFloor(g, m, pos)
			door := g.NewDoor(displayName)
			door.SetStateFromCategory(doorCat)
			if len(parts) > 2 {
				lockDiff := foundation.DifficultyFromString(strings.TrimSpace(parts[2]))
				door.SetLockDifficulty(lockDiff)
			}
			if len(parts) > 3 {
				keyID := strings.TrimSpace(parts[3])
				door.SetLockedByKey(keyID)
			}
			m.AddObject(door, pos)
		}
	case typeDesc == "deadactor":
		internalName := strings.TrimSpace(parts[1])
		return func(g *GameState, m *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) {
			setDefaultFloor(g, m, pos)
			actorDef := g.dataDefinitions.GetMonsterByName(internalName)
			newActor := g.NewEnemyFromDef(actorDef)
			newActor.Kill()
			m.AddDownedActor(newActor, pos)
		}

	}
	return nil
}

func fillContainerFromRecord(g *GameState, record recfile.Record, container *Container) {
	randomLootCount := 0
	randomLootQuality := 1

	for _, field := range record {
		if field.Name == "item" {
			addedItem := g.NewItemFromName(field.Value)
			container.AddItem(addedItem)
		} else if field.Name == "random_loot_count" {
			randomLootCount = special.ParseInterval(field.Value).Roll()
		} else if field.Name == "random_loot_quality" {
			randomLootQuality = field.AsInt()
		}
	}
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < randomLootCount; i++ {
		itemDef := g.dataDefinitions.PickItemForLevel(random, randomLootQuality)
		randomLoot := NewItem(itemDef)
		container.AddItem(randomLoot)
	}
}

type RawMapDescription struct {
	Dimensions geometry.Point
	MapData    []rune
	Legend     []string
	Flags      []string

	ObjectRecords map[string]recfile.Record
	ActorRecords  map[string]recfile.Record
}

func (d RawMapDescription) IsEmpty() bool {
	return len(d.MapData) == 0 || d.Dimensions.X == 0 || d.Dimensions.Y == 0
}

func ReadMapFile(filename string) RawMapDescription {
	type ParseState uint8
	const (
		Map ParseState = iota
		Legend
		Flags
		ActorRecords
		ObjectRecords
	)
	currentParseState := Map
	file, err := os.Open(filename)
	if err != nil {
		return RawMapDescription{}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var mapChars []rune
	var legendLines []string
	var flagLines []string
	objectRecords := make(map[string]recfile.Record)
	actorRecords := make(map[string]recfile.Record)
	var currentObjectRecord []string
	var currentActorRecord []string
	dataEdgeReached := false
	width := 0
	height := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		curLineWidth := utf8.RuneCountInString(line)
		if width == 0 && len(mapChars) == 0 {
			width = curLineWidth
		}

		if curLineWidth == width && currentParseState == Map {
			mapChars = append(mapChars, []rune(line)...)
			height++
		}
		if line == "## Legend" {
			currentParseState = Legend
			dataEdgeReached = true
		} else if line == "## InitFlags" {
			currentParseState = Flags
			dataEdgeReached = true
		} else if line == "## ActorRecords" {
			currentParseState = ActorRecords
			dataEdgeReached = true
		} else if line == "## ObjectRecords" {
			currentParseState = ObjectRecords
			dataEdgeReached = true
		} else if curLineWidth == 0 {
			dataEdgeReached = true
		}
		if dataEdgeReached {
			if currentParseState == ObjectRecords && len(currentObjectRecord) > 0 {
				// commit the current record
				record, recordName := recfile.RecordFromSlice(currentObjectRecord).WithPoppedValue("name")
				objectRecords[recordName] = record
				currentObjectRecord = make([]string, 0)
			} else if currentParseState == ActorRecords && len(currentActorRecord) > 0 {
				// commit the current record
				record, recordName := recfile.RecordFromSlice(currentActorRecord).WithPoppedValue("name")
				actorRecords[recordName] = record
				currentActorRecord = make([]string, 0)
			}
			dataEdgeReached = false
			continue
		}
		if currentParseState == Legend {
			legendLines = append(legendLines, line)
		} else if currentParseState == Flags {
			flagLines = append(flagLines, line)
		} else if currentParseState == ActorRecords {
			currentActorRecord = append(currentActorRecord, line)
		} else if currentParseState == ObjectRecords {
			currentObjectRecord = append(currentObjectRecord, line)
		}
	}
	if currentParseState == ObjectRecords && len(currentObjectRecord) > 0 {
		// commit the current record
		record, recordName := recfile.RecordFromSlice(currentObjectRecord).WithPoppedValue("name")
		objectRecords[recordName] = record
	} else if currentParseState == ActorRecords && len(currentActorRecord) > 0 {
		// commit the current record
		record, recordName := recfile.RecordFromSlice(currentActorRecord).WithPoppedValue("name")
		actorRecords[recordName] = record
	}
	//return geometry.Point{X: width, Y: height}, mapChars, legendLines, flagLines, equipmentLines
	return RawMapDescription{
		Dimensions:    geometry.Point{X: width, Y: height},
		MapData:       mapChars,
		Legend:        legendLines,
		Flags:         flagLines,
		ActorRecords:  actorRecords,
		ObjectRecords: objectRecords,
	}
}
