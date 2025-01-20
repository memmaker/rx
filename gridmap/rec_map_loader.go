package gridmap

import (
    "github.com/memmaker/go/fxtools"
    "github.com/memmaker/go/geometry"
    "github.com/memmaker/go/recfile"
    "github.com/memmaker/go/textiles"
    "math/rand"
    "path/filepath"
    "strings"
    "time"
)

type RecMapLoader[ActorType interface {
    comparable
    MapActor
}, ItemType interface {
    comparable
    MapItem
}, ObjectType interface {
    comparable
    MapObjectWithProperties[ActorType]
}] struct {
    random        *rand.Rand
    palette       textiles.ColorPalette
    actorFactory  func(rec recfile.Record, mapName string) (ActorType, geometry.Point)
    itemFactory   func(rec recfile.Record) (ItemType, geometry.Point)
    objectFactory func(rec recfile.Record, newMap *GridMap[ActorType, ItemType, ObjectType], iconResolver func(objType string) textiles.TextIcon) (ObjectType, geometry.Point)
    mapBaseDir    string
    diagonalMove  bool
}

func NewRecMapLoader[ActorType interface {
    comparable
    MapActor
}, ItemType interface {
    comparable
    MapItem
}, ObjectType interface {
    comparable
    MapObjectWithProperties[ActorType]
}](
    mapBaseDir string,
    palette textiles.ColorPalette,
    actorFactory func(rec recfile.Record, mapName string) (ActorType, geometry.Point),
    itemFactory func(rec recfile.Record) (ItemType, geometry.Point),
    objectFactory func(rec recfile.Record, newMap *GridMap[ActorType, ItemType, ObjectType], iconsForObjects func(objType string) textiles.TextIcon) (ObjectType, geometry.Point),
) *RecMapLoader[ActorType, ItemType, ObjectType] {
    return &RecMapLoader[ActorType, ItemType, ObjectType]{
        random:        rand.New(rand.NewSource(time.Now().UnixNano())),
        palette:       palette,
        actorFactory:  actorFactory,
        itemFactory:   itemFactory,
        objectFactory: objectFactory,
        mapBaseDir:    mapBaseDir,
        diagonalMove:  true,
    }
}

type MapLoadResult[ActorType interface {
    comparable
    MapActor
}, ItemType interface {
    comparable
    MapItem
}, ObjectType interface {
    comparable
    MapObjectWithProperties[ActorType]
}] struct {
    Map          *GridMap[ActorType, ItemType, ObjectType]
    FlagsOfMap   map[string]int
    ScriptsToRun []string
}

func (t *RecMapLoader[ActorType, ItemType, ObjectType]) LoadMap(mapName string) MapLoadResult[ActorType, ItemType, ObjectType] {
    mapDir := filepath.Join(t.mapBaseDir, mapName)
    if !fxtools.DirExists(mapDir) {
        return MapLoadResult[ActorType, ItemType, ObjectType]{}
    }

    objTypes := LoadIconsForObjects(mapDir, t.palette)
    iconresolver := func(objType string) textiles.TextIcon {
        if icon, exists := objTypes[strings.ToLower(objType)]; exists {
            return icon
        }
        return textiles.TextIcon{}
    }

    tileSet := textiles.ReadTilesFileAndClose(fxtools.MustOpen(filepath.Join(mapDir, "tileSet.rec")), t.palette)
    mapSize, tileMap := textiles.ReadTileMap16(filepath.Join(mapDir, "tiles.bin"))

    scriptRecs, _ := recfile.ReadAndClose(fxtools.MustOpen(filepath.Join(mapDir, "meta.rec")))
    metaData, scriptsToRun := NewMapMetaData(scriptRecs[0])

    var zones map[string]map[geometry.Point]bool
    if fxtools.FileExists(filepath.Join(mapDir, "zones.bin")) {
        zones = textiles.ReadZones(filepath.Join(mapDir, "zones.bin"))
    }
    var zoneRecords, actorRecords, objectRecords, itemRecords []recfile.Record
    var zoneMeta map[string]ZoneMetadata
    if fxtools.FileExists(filepath.Join(mapDir, "zones.rec")) {
        zoneRecords, _ = recfile.ReadAndClose(fxtools.MustOpen(filepath.Join(mapDir, "zones.rec")))
        zoneMeta = NewZoneMetadata(zoneRecords)
    }

    if fxtools.FileExists(filepath.Join(mapDir, "actors.rec")) {
        actorRecords, _ = recfile.ReadAndClose(fxtools.MustOpen(filepath.Join(mapDir, "actors.rec")))
    }
    if fxtools.FileExists(filepath.Join(mapDir, "objects.rec")) {
        objectRecords, _ = recfile.ReadAndClose(fxtools.MustOpen(filepath.Join(mapDir, "objects.rec")))
    }
    if fxtools.FileExists(filepath.Join(mapDir, "items.rec")) {
        itemRecords, _ = recfile.ReadAndClose(fxtools.MustOpen(filepath.Join(mapDir, "items.rec")))
    }

    // Optional: Read init flags, if they exist and haven't been loaded before
    flagsFile := filepath.Join(mapDir, "initFlags.rec")
    flagsOfMap := make(map[string]int)
    if fxtools.FileExists(flagsFile) {
        flagRecs, _ := recfile.ReadAndClose(fxtools.MustOpen(flagsFile))
        flags := flagRecs[0]
        for _, flag := range flags {
            flagsOfMap[flag.Name] = flag.AsInt()
        }
    }

    newMap := NewEmptyMap[ActorType, ItemType, ObjectType](mapSize.X, mapSize.Y)
    newMap.SetCardinalMovementOnly(!t.diagonalMove)
    newMap.SetName(mapName)
    newMap.SetMeta(metaData)
    newMap.SetZones(zones)
    newMap.SetZoneMetadata(zoneMeta)

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
        spawnedActor, spawnPos := t.actorFactory(record, mapName)
        if spawnedActor.IsAlive() {
            newMap.AddActor(spawnedActor, spawnPos)
        } else {
            newMap.AddDownedActor(spawnedActor, spawnPos)
        }
    }

    // Set GetItems
    for _, record := range itemRecords {
        newMap.AddItem(t.itemFactory(record))
    }

    // Set Objects
    for _, record := range objectRecords {
        objCategory := record.FindValueForKeyIgnoreCase("category")
        if tryHandleAsPseudoObject(objCategory, record, newMap) {
            continue
        }
        newMap.AddObject(t.objectFactory(record, newMap, iconresolver))
    }
    newMap.UpdateBakedLights()

    return MapLoadResult[ActorType, ItemType, ObjectType]{
        Map:          newMap,
        FlagsOfMap:   flagsOfMap,
        ScriptsToRun: scriptsToRun,
    }
}

type MapMeta struct {
    DisplayName        string
    IsOutdoor          bool
    MusicFile          string
    IndoorAmbientLight fxtools.HDRColor
    LastVisited        time.Time
}

func (m MapMeta) WithAmbientLight(color fxtools.HDRColor) MapMeta {
    m.IndoorAmbientLight = color
    return m
}

func (m MapMeta) WithLastVisited(t time.Time) MapMeta {
    m.LastVisited = t
    return m
}

func NewMapMetaData(record recfile.Record) (MapMeta, []string) {
    var result MapMeta
    var scriptsToRun []string
    for _, field := range record {
        switch strings.ToLower(field.Name) {
        case "name":
            result.DisplayName = field.Value
        case "isoutdoor":
            result.IsOutdoor = field.AsBool()
        case "musicfile":
            result.MusicFile = field.Value
        case "indoorambientlight":
            result.IndoorAmbientLight = fxtools.NewColorFromString(field.Value)
        case "runscript":
            scriptsToRun = append(scriptsToRun, field.Value)
        }
    }
    return result, scriptsToRun
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

    objectTypeFile := filepath.Join(dataDirectory, "iconsForObjects.rec")
    iconsForObjects := textiles.ReadIconRecordsIntoMap(fxtools.MustOpen(objectTypeFile))

    return convertObjectCategories(iconsForObjects)
}
func tryHandleAsPseudoObject[ActorType interface {
    comparable
    MapActor
}, ItemType interface {
    comparable
    MapItem
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
        // AddItem a named location
        name := rec.FindValueForKeyIgnoreCase("Location")
        pos, _ := geometry.NewPointFromEncodedString(rec.FindValueForKeyIgnoreCase("position"))
        newMap.AddNamedLocation(name, pos)

        // AddItem the transition
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
        Flags:              TileFlags(tile.Flags),
    }
}

type ZoneMetadata struct {
    Lighting  fxtools.HDRColor
    Music     string
    IsPrivate bool
    IsIndoor  bool
}

func NewZoneMetadata(records []recfile.Record) map[string]ZoneMetadata {
    result := make(map[string]ZoneMetadata)
    for _, record := range records {
        meta := ZoneMetadata{}
        var zoneName string
        for _, field := range record {
            switch strings.ToLower(field.Name) {
            case "name":
                zoneName = field.Value
            case "lighting":
                meta.Lighting = fxtools.NewColorFromString(field.Value)
            case "music":
                meta.Music = field.Value
            case "isprivate":
                meta.IsPrivate = field.AsBool()
            case "isindoor":
                meta.IsIndoor = field.AsBool()
            }
        }
        result[zoneName] = meta
    }
    return result
}
