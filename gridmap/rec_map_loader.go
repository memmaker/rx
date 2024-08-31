package gridmap

import (
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"math/rand"
	"path"
	"strings"
	"time"
)

type RecMapLoader[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}] struct {
	random                     *rand.Rand
	palette                    textiles.ColorPalette
	actorFactory               func(rec recfile.Record) (ActorType, geometry.Point)
	itemFactory                func(rec recfile.Record) (ItemType, geometry.Point)
	objectFactory              func(rec recfile.Record, newMap *GridMap[ActorType, ItemType, ObjectType]) (ObjectType, geometry.Point)
	mapBaseDir                 string
	diagonalMove               bool
	setIconsResolverForObjects func(iconsForObject map[string]textiles.TextIcon)
}

func NewRecMapLoader[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}](
	mapBaseDir string,
	palette textiles.ColorPalette,
	setIconResolver func(iconsForObject map[string]textiles.TextIcon),
	actorFactory func(rec recfile.Record) (ActorType, geometry.Point),
	itemFactory func(rec recfile.Record) (ItemType, geometry.Point),
	objectFactory func(rec recfile.Record, newMap *GridMap[ActorType, ItemType, ObjectType]) (ObjectType, geometry.Point),
) *RecMapLoader[ActorType, ItemType, ObjectType] {
	return &RecMapLoader[ActorType, ItemType, ObjectType]{
		random:                     rand.New(rand.NewSource(time.Now().UnixNano())),
		palette:                    palette,
		actorFactory:               actorFactory,
		itemFactory:                itemFactory,
		objectFactory:              objectFactory,
		mapBaseDir:                 mapBaseDir,
		diagonalMove:               true,
		setIconsResolverForObjects: setIconResolver,
	}
}

type MapLoadResult[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}] struct {
	Map             *GridMap[ActorType, ItemType, ObjectType]
	IconsForObjects map[string]textiles.TextIcon
	FlagsOfMap      map[string]int
}

func (t *RecMapLoader[ActorType, ItemType, ObjectType]) LoadMap(mapName string) MapLoadResult[ActorType, ItemType, ObjectType] {
	mapDir := path.Join(t.mapBaseDir, mapName)
	if !fxtools.DirExists(mapDir) {
		return MapLoadResult[ActorType, ItemType, ObjectType]{}
	}

	objTypes := LoadIconsForObjects(mapDir, t.palette)
	t.setIconsResolverForObjects(objTypes)

	tileSet := textiles.ReadTilesFile(fxtools.MustOpen(path.Join(mapDir, "tileSet.rec")), t.palette)
	mapSize, tileMap := textiles.ReadTileMap16(path.Join(mapDir, "tiles.bin"))

	metaData := NewMapMetaData(recfile.Read(fxtools.MustOpen(path.Join(mapDir, "meta.rec")))[0])

	actorRecords := recfile.Read(fxtools.MustOpen(path.Join(mapDir, "actors.rec")))
	objectRecords := recfile.Read(fxtools.MustOpen(path.Join(mapDir, "objects.rec")))
	itemRecords := recfile.Read(fxtools.MustOpen(path.Join(mapDir, "items.rec")))

	// Optional: Read init flags, if they exist and haven't been loaded before
	flagsFile := path.Join(mapDir, "initFlags.rec")
	flagsOfMap := make(map[string]int)
	if fxtools.FileExists(flagsFile) {
		flags := recfile.Read(fxtools.MustOpen(flagsFile))[0]
		for _, flag := range flags {
			flagsOfMap[flag.Name] = flag.AsInt()
		}
	}

	newMap := NewEmptyMap[ActorType, ItemType, ObjectType](mapSize.X, mapSize.Y)
	newMap.SetCardinalMovementOnly(!t.diagonalMove)
	newMap.SetName(mapName)
	newMap.SetMeta(metaData)

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
		newMap.AddActor(t.actorFactory(record))
	}

	// Set Items
	for _, record := range itemRecords {
		newMap.AddItem(t.itemFactory(record))
	}

	// Set Objects
	for _, record := range objectRecords {
		objCategory := record.FindValueForKeyIgnoreCase("category")
		if tryHandleAsPseudoObject(objCategory, record, newMap) {
			continue
		}
		newMap.AddObject(t.objectFactory(record, newMap))
	}
	newMap.UpdateBakedLights()

	return MapLoadResult[ActorType, ItemType, ObjectType]{
		Map:             newMap,
		IconsForObjects: objTypes,
		FlagsOfMap:      flagsOfMap,
	}
}

type MapMeta struct {
	Name               string
	IsOutdoor          bool
	MusicFile          string
	IndoorAmbientLight fxtools.HDRColor
}

func NewMapMetaData(record recfile.Record) MapMeta {
	var result MapMeta
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "name":
			result.Name = field.Value
		case "isoutdoor":
			result.IsOutdoor = field.AsBool()
		case "musicfile":
			result.MusicFile = field.Value
		case "indoorambientlight":
			result.IndoorAmbientLight = fxtools.NewColorFromString(field.Value)
		}
	}
	return result
}

func LoadIconsForObjects(dataDirectory string, colors textiles.ColorPalette) map[string]textiles.TextIcon {
	convertObjectCategories := func(r map[string]textiles.IconRecord) map[string]textiles.TextIcon {
		convertMap := make(map[string]textiles.TextIcon)
		for name, rec := range r {
			icon := textiles.NewTextIconFromNamedColorChar(rec.Icon, colors)
			convertMap[strings.ToLower(name)] = icon
		}
		return convertMap
	}

	objectTypeFile := path.Join(dataDirectory, "iconsForObjects.rec")
	iconsForObjects := textiles.ReadIconRecordsIntoMap(fxtools.MustOpen(objectTypeFile))

	return convertObjectCategories(iconsForObjects)
}
func tryHandleAsPseudoObject[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}](objName string, rec recfile.Record, newMap *GridMap[ActorType, ItemType, ObjectType]) bool {
	switch strings.ToLower(objName) {
	case "bakedlight":
		pos, _ := geometry.NewPointFromEncodedString(rec.FindValueForKeyIgnoreCase("position"))
		light := NewLightSourceFromRecord(rec)
		newMap.AddBakedLightSource(pos, light)
		return true
	case "transition":
		// Add a named location
		name := rec.FindValueForKeyIgnoreCase("Location")
		pos, _ := geometry.NewPointFromEncodedString(rec.FindValueForKeyIgnoreCase("position"))
		newMap.AddNamedLocation(name, pos)

		// Add the transition
		targetMap := rec.FindValueForKeyIgnoreCase("TargetMap")
		targetLocation := rec.FindValueForKeyIgnoreCase("TargetLocation")
		newMap.AddTransitionAt(pos, Transition{
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

func toGridMapTile(tile textiles.TextTile) Tile {
	return Tile{
		Icon:               tile.Icon,
		DefinedDescription: tile.Name,
		IsWalkable:         tile.IsWalkable,
		IsTransparent:      tile.IsTransparent,
		IsDamaging:         false,
	}
}
