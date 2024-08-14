package game

import (
	"RogueUI/gridmap"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"math/rand"
	"path"
	"strings"
	"time"
)

type RecMapLoader struct {
	gameState *GameState
	random    *rand.Rand
	palette   textiles.ColorPalette
}

func NewRecMapLoader(g *GameState, palette textiles.ColorPalette) *RecMapLoader {
	return &RecMapLoader{gameState: g, random: rand.New(rand.NewSource(time.Now().UnixNano())), palette: palette}
}
func (t *RecMapLoader) LoadMap(levelName string) *gridmap.GridMap[*Actor, *Item, Object] {
	g := t.gameState
	mapDir := path.Join(g.config.DataRootDir, "maps", levelName)
	if !fxtools.DirExists(mapDir) {
		return nil
	}

	objTypes := loadIconsForObjects(mapDir, t.palette)
	g.iconsForObjects = objTypes

	tileSet := textiles.ReadTilesFile(fxtools.MustOpen(path.Join(mapDir, "tileSet.rec")), t.palette)
	mapSize, tileMap := textiles.ReadTileMap16(path.Join(mapDir, "tiles.bin"))

	actorRecords := recfile.Read(fxtools.MustOpen(path.Join(mapDir, "actors.rec")))
	objectRecords := recfile.Read(fxtools.MustOpen(path.Join(mapDir, "objects.rec")))
	itemRecords := recfile.Read(fxtools.MustOpen(path.Join(mapDir, "items.rec")))

	// Optional: Read init flags, if they exist and haven't been loaded before
	flagsFile := path.Join(mapDir, "initFlags.rec")
	didInitFlagName := fmt.Sprintf("DidInitLevel(%s)", levelName)
	if fxtools.FileExists(flagsFile) && !g.gameFlags.HasFlag(didInitFlagName) {
		flags := recfile.Read(fxtools.MustOpen(flagsFile))[0]
		for _, flag := range flags {
			g.gameFlags.Set(flag.Name, flag.AsInt())
		}
	}

	newMap := gridmap.NewEmptyMap[*Actor, *Item, Object](mapSize.X, mapSize.Y)
	newMap.SetCardinalMovementOnly(!g.config.DiagonalMovementEnabled)
	newMap.SetName(levelName)

	// Set Tiles
	for y := 0; y < mapSize.Y; y++ {
		for x := 0; x < mapSize.X; x++ {
			tileIndex := tileMap[y*mapSize.X+x]
			tile := tileSet[tileIndex]
			newMap.SetTile(geometry.Point{X: x, Y: y}, toGridMapTile(tile))
		}
	}

	// Set Actors
	for _, record := range actorRecords {
		actorDef := NewActorDefFromRecord(record, t.palette)
		newActor := g.NewEnemyFromDef(actorDef)
		if newActor != nil {
			actorPos := newActor.Position()
			newMap.AddActor(newActor, actorPos)
		}
	}

	// Set Items
	for _, record := range itemRecords {
		itemDef := NewItemDefFromRecord(record)
		newItem := NewItem(itemDef, g.iconForItem(itemDef.Category))
		if newItem != nil {
			itemPos := newItem.Position()
			newMap.AddItem(newItem, itemPos)
		}
	}

	// Set Objects
	for _, record := range objectRecords {
		objCategory := record.FindValueForKeyIgnoreCase("category")
		if tryHandleAsPseudoObject(objCategory, record, newMap) {
			continue
		}
		object := g.NewObjectFromRecord(record, t.palette, newMap)
		if object != nil {
			objectPos := object.Position()
			newMap.AddObject(object, objectPos)
		}
	}

	return newMap
}

func tryHandleAsPseudoObject(objName string, rec recfile.Record, newMap *gridmap.GridMap[*Actor, *Item, Object]) bool {
	switch strings.ToLower(objName) {
	case "transition":
		// Add a named location
		name := rec.FindValueForKeyIgnoreCase("Location")
		pos, _ := geometry.NewPointFromEncodedString(rec.FindValueForKeyIgnoreCase("position"))
		newMap.AddNamedLocation(name, pos)

		// Add the transition
		targetMap := rec.FindValueForKeyIgnoreCase("TargetMap")
		targetLocation := rec.FindValueForKeyIgnoreCase("TargetLocation")
		newMap.AddTransitionAt(pos, gridmap.Transition{
			TargetMap:      targetMap,
			TargetLocation: targetLocation,
		})
		return true
	case "namedlocation":
		name := rec.FindValueForKeyIgnoreCase("identifier")
		pos, _ := geometry.NewPointFromEncodedString(rec.FindValueForKeyIgnoreCase("position"))
		newMap.AddNamedLocation(name, pos)
		return true
	}
	return false
}

func toGridMapTile(tile textiles.TextTile) gridmap.Tile {
	return gridmap.Tile{
		Icon:               tile.Icon,
		DefinedDescription: tile.Name,
		IsWalkable:         tile.IsWalkable,
		IsTransparent:      tile.IsTransparent,
		IsDamaging:         false,
	}
}
