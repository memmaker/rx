package game

import (
	"RogueUI/dungen"
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/gridmap"
	"bufio"
	"math/rand"
	"os"
	"path"
	"time"
)

func (g *GameState) GotoNamedLevel(levelName string) {
	mapData := ReadFileAsOneStringWithoutNewLines(path.Join("data", "prefabs", levelName+".txt"))
	if mapData == "" {
		return
	}
	var spawnPos geometry.Point
	simpleMapper := func(gridMap *gridmap.GridMap[*Actor, *Item, *Object], icon rune, pos geometry.Point) {
		switch icon {
		case '@':
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileFloor,
				DefinedDescription: "floor",
				IsWalkable:         true,
				IsTransparent:      true,
			})
			spawnPos = pos
		case '^':
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileMountain,
				DefinedDescription: "wall",
				IsWalkable:         false,
				IsTransparent:      false,
			})
		case '#':
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileWall,
				DefinedDescription: "wall",
				IsWalkable:         false,
				IsTransparent:      true,
			})
		case ' ':
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileFloor,
				DefinedDescription: "floor",
				IsWalkable:         true,
				IsTransparent:      true,
			})
		case '>':
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileStairsDown,
				DefinedDescription: "stairs down",
				IsWalkable:         true,
				IsTransparent:      true,
			})
		case '1':
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileVendorGeneral,
				DefinedDescription: "a general store",
				IsWalkable:         true,
				IsTransparent:      true,
			})
		case '2':
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileVendorArmor,
				DefinedDescription: "an armor store",
				IsWalkable:         true,
				IsTransparent:      true,
			})
		case '3':
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileVendorWeapons,
				DefinedDescription: "a weapon store",
				IsWalkable:         true,
				IsTransparent:      true,
			})
		case '4':
			gridMap.SetTile(pos, gridmap.Tile{
				Feature:            foundation.TileVendorAlchemist,
				DefinedDescription: "an alchemist outlet",
				IsWalkable:         true,
				IsTransparent:      true,
			})
		}
	}
	w, h := g.gridMap.GetWidth(), g.gridMap.GetHeight()
	newMap := gridmap.NewMapFromString[*Actor, *Item, *Object](w, h, mapData, simpleMapper)
	newMap.SetCardinalMovementOnly(!g.config.DiagonalMovementEnabled)

	if g.gridMap != nil {
		g.gridMap.RemoveActor(g.Player)
		g.Player.RemoveLevelStatusEffects()
	}

	g.dungeonLayout = nil

	newMap.AddActor(g.Player, spawnPos)

	g.gridMap = newMap

	g.afterPlayerMoved()

	g.updateUIStatus()
}

func (g *GameState) GotoDungeonLevel(level int, stairs StairsInLevel, placePlayerOnStairs bool) {
	if g.gridMap != nil {
		g.gridMap.RemoveActor(g.Player)
		g.Player.RemoveLevelStatusEffects()
	}
	isDown := level > g.currentDungeonLevel

	g.currentDungeonLevel = level
	if g.deepestDungeonLevelPlayerReached < level {
		g.deepestDungeonLevelPlayerReached = level
		if level > 1 {
			g.newLevelReached(level)
		}
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	dunGen := dungen.NewRogueGenerator(random, g.config.MapWidth, g.config.MapHeight)
	dunGen.SetRoomLitChance(0.9)
	dunGen.SetAdditionalRoomConnections(random.Intn(5))
	dungeon := dunGen.Generate()

	mapWidth, mapHeight := dungeon.GetSize()

	g.dungeonLayout = dungeon

	newMap := gridmap.NewEmptyMap[*Actor, *Item, *Object](mapWidth, mapHeight)
	newMap.SetCardinalMovementOnly(!g.config.DiagonalMovementEnabled)

	stairsUp, stairsDown := g.decorateMapWithTiles(newMap, dungeon, stairs)

	// place player
	//var otherEndPos geometry.Point
	if placePlayerOnStairs && isDown && stairs.AllowsUp() {
		newMap.AddActor(g.Player, stairsUp)
		//otherEndPos = stairsDown
	} else if placePlayerOnStairs && !isDown && stairs.AllowsDown() {
		newMap.AddActor(g.Player, stairsDown)
		//otherEndPos = stairsUp
	} else {
		randomPos := newMap.RandomSpawnPosition()
		newMap.AddActor(g.Player, randomPos)
	}

	g.spawnEntities(random, level, newMap, dungeon)

	spawnPos := g.Player.Position()

	g.gridMap = newMap

	g.afterPlayerMoved()

	g.ui.AfterPlayerMoved(foundation.MoveInfo{
		Direction: geometry.North,
		OldPos:    spawnPos,
		NewPos:    spawnPos,
		Mode:      foundation.PlayerMoveModeManual,
	})
	g.updateUIStatus()
}

func (g *GameState) decorateMapWithTiles(newMap *gridmap.GridMap[*Actor, *Item, *Object], dungeon *dungen.DungeonMap, stairs StairsInLevel) (geometry.Point, geometry.Point) {
	mapWidth, mapHeight := dungeon.GetSize()

	floorTile := gridmap.Tile{
		Feature:            foundation.TileRoomFloor,
		DefinedDescription: "floor",
		IsWalkable:         true,
		IsTransparent:      true,
	}

	corridorTile := gridmap.Tile{
		Feature:            foundation.TileCorridorFloor,
		DefinedDescription: "corridor",
		IsWalkable:         true,
		IsTransparent:      true,
	}

	horizWallTile := gridmap.Tile{
		Feature:            foundation.TileRoomWallHorizontal,
		DefinedDescription: "wall",
		IsWalkable:         false,
		IsTransparent:      false,
	}

	vertWallTile := gridmap.Tile{
		Feature:            foundation.TileRoomWallVertical,
		DefinedDescription: "wall",
		IsWalkable:         false,
		IsTransparent:      false,
	}

	tlWallTile := gridmap.Tile{
		Feature:            foundation.TileRoomWallCornerTopLeft,
		DefinedDescription: "wall",
		IsWalkable:         false,
		IsTransparent:      false,
	}

	trWallTile := gridmap.Tile{
		Feature:            foundation.TileRoomWallCornerTopRight,
		DefinedDescription: "wall",
		IsWalkable:         false,
		IsTransparent:      false,
	}
	blWallTile := gridmap.Tile{
		Feature:            foundation.TileRoomWallCornerBottomLeft,
		DefinedDescription: "wall",
		IsWalkable:         false,
		IsTransparent:      false,
	}
	brWallTile := gridmap.Tile{
		Feature:            foundation.TileRoomWallCornerBottomRight,
		DefinedDescription: "wall",
		IsWalkable:         false,
		IsTransparent:      false,
	}

	fakeDoorTile := gridmap.Tile{
		Feature:            foundation.TileDoorClosed,
		DefinedDescription: "door",
		IsWalkable:         true,
		IsTransparent:      true,
	}

	corridorWall := gridmap.Tile{
		Feature:            foundation.TileCorridorWallVertical,
		DefinedDescription: "wall",
		IsWalkable:         false,
		IsTransparent:      false,
	}

	stairsUp := gridmap.Tile{
		Feature:            foundation.TileStairsUp,
		DefinedDescription: "stairs up",
		IsWalkable:         true,
		IsTransparent:      true,
	}
	stairsDown := gridmap.Tile{
		Feature:            foundation.TileStairsDown,
		DefinedDescription: "stairs down",
		IsWalkable:         true,
		IsTransparent:      true,
	}
	newMap.FillTile(corridorWall)

	var stairsUpLoc geometry.Point
	var stairsDownLoc geometry.Point
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			tile := dungeon.GetTile(x, y)
			pos := geometry.Point{X: x, Y: y}
			if tile == dungen.Room {
				newMap.SetTile(pos, floorTile)
			} else if tile == dungen.Corridor {
				newMap.SetTile(pos, corridorTile)
			} else if tile == dungen.Door {
				newMap.SetTile(pos, fakeDoorTile)
			} else if tile == dungen.StairsUp && stairs.AllowsUp() {
				newMap.SetTile(pos, stairsUp)
				stairsUpLoc = pos
			} else if tile == dungen.StairsDown && stairs.AllowsDown() {
				newMap.SetTile(pos, stairsDown)
				stairsDownLoc = pos
			}
		}
	}

	// decorate walls & light up the rooms
	for _, room := range dungeon.AllRooms() {
		for _, pos := range room.GetWalls() {
			eastN := pos.Add(geometry.East.ToPoint())
			westN := pos.Add(geometry.West.ToPoint())
			if room.FloorContains(eastN) || room.FloorContains(westN) {
				newMap.SetTile(pos, vertWallTile)
				continue
			}
			northN := pos.Add(geometry.North.ToPoint())
			southN := pos.Add(geometry.South.ToPoint())
			if room.FloorContains(northN) || room.FloorContains(southN) {
				newMap.SetTile(pos, horizWallTile)
				continue
			}
			// corners
			if room.IsTopLeftWallCorner(pos) {
				newMap.SetTile(pos, tlWallTile)
				continue
			}
			if room.IsTopRightWallCorner(pos) {
				newMap.SetTile(pos, trWallTile)
				continue
			}
			if room.IsBottomLeftWallCorner(pos) {
				newMap.SetTile(pos, blWallTile)
				continue
			}
			if room.IsBottomRightWallCorner(pos) {
				newMap.SetTile(pos, brWallTile)
				continue
			}
		}
		if room.IsLit() {
			for _, pos := range room.GetAbsoluteRoomTiles() {
				newMap.SetLit(pos, true)
			}
		}
	}
	return stairsUpLoc, stairsDownLoc
}

func ReadFileAsOneStringWithoutNewLines(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var result string
	for scanner.Scan() {
		result += scanner.Text()
	}
	return result
}
