package game

import (
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"bufio"
	"fmt"
	"github.com/memmaker/go/geometry"
	"os"
	"path"
)

func (g *GameState) GotoNamedLevel(levelName string, location string) {

	if g.metronome.LeavingMapEvents() {
		g.ui.AnimatePending()
	}

	var loadedMap *gridmap.GridMap[*Actor, foundation.Item, Object]
	var ok bool
	var firstTimeInit func()
	if loadedMap, ok = g.activeMaps[levelName]; !ok {
		result := g.mapLoader.LoadMap(levelName)
		loadedMap = result.Map

		if loadedMap == nil {
			g.msg(foundation.Msg("It's impossible to move there.."))
			return
		}

		g.iconsForObjects = result.IconsForObjects

		firstTimeInit = func() {
			flags := result.FlagsOfMap
			for flagName, flagValue := range flags {
				g.gameFlags.Set(flagName, flagValue)
			}

			scripts := result.ScriptsToRun
			for _, script := range scripts {
				g.RunScriptByName(script)
			}
		}

	} else {
		g.iconsForObjects = gridmap.LoadIconsForObjects(path.Join(g.config.DataRootDir, "maps", levelName), g.palette)
	}

	if g.currentMap() != nil && g.Player != nil { // RemoveItem Player from Old Map
		g.currentMap().RemoveActor(g.Player)
		g.Player.RemoveLevelStatusEffects()
	}

	namedLocation := loadedMap.GetNamedLocation(location)
	loadedMap.AddActor(g.Player, namedLocation)

	mapVisited := fmt.Sprintf("PlayerVisited(%s)", levelName)
	g.gameFlags.Increment(mapVisited)

	g.setCurrentMap(loadedMap)

	g.afterPlayerMoved(geometry.Point{}, true)

	if firstTimeInit != nil {
		firstTimeInit()
	}

	g.updateUIStatus()

	g.ui.PlayMusic(path.Join(g.config.DataRootDir, "audio", "music", loadedMap.GetMeta().MusicFile+".ogg"))
}

/*
func (g *GameState) GotoDungeonLevel(level int, stairs StairsInLevel, placePlayerOnStairs bool) {
    if g.currentMap() != nil {
        g.currentMap().RemoveActor(g.Player)
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

    newMap := gridmap.NewEmptyMap[*Actor, *Item, Object](mapWidth, mapHeight)
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

    g.currentMap() = newMap

    g.afterPlayerMoved()

    g.ui.AfterPlayerMoved(foundation.MoveInfo{
        Direction: geometry.North,
        OldPos:    spawnPos,
        NewPos:    spawnPos,
        Mode:      foundation.PlayerMoveModeManual,
    })
    g.updateUIStatus()

}
*/
/*
func (g *GameState) decorateMapWithTiles(newMap *gridmap.GridMap[*Actor, *Item, Object], dungeon *dungen.DungeonMap, stairs StairsInLevel) (geometry.Point, geometry.Point) {
    mapWidth, mapHeight := dungeon.GetSize()

    floorTile := g.dataDefinitions.GetTileByName("RoomFloor")

    corridorTile := g.dataDefinitions.GetTileByName("CorridorFloor")

    horizWallTile := g.dataDefinitions.GetTileByName("RoomWallHorizontal")

    vertWallTile := g.dataDefinitions.GetTileByName("RoomWallVertical")

    tlWallTile := g.dataDefinitions.GetTileByName("RoomWallCornerTopLeft")
    trWallTile := g.dataDefinitions.GetTileByName("RoomWallCornerTopRight")
    blWallTile := g.dataDefinitions.GetTileByName("RoomWallCornerBottomLeft")
    brWallTile := g.dataDefinitions.GetTileByName("RoomWallCornerBottomRight")

    fakeDoorTile := g.dataDefinitions.GetTileByName("DoorClosed")
    corridorWall := g.dataDefinitions.GetTileByName("CorridorWallVertical")
    stairsUp := g.dataDefinitions.GetTileByName("StairsUp")
    stairsDown := g.dataDefinitions.GetTileByName("StairsDown")

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
            eastN := pos.AddItem(geometry.East.ToPoint())
            westN := pos.AddItem(geometry.West.ToPoint())
            if room.FloorContains(eastN) || room.FloorContains(westN) {
                newMap.SetTile(pos, vertWallTile)
                continue
            }
            northN := pos.AddItem(geometry.North.ToPoint())
            southN := pos.AddItem(geometry.South.ToPoint())
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
*/
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
