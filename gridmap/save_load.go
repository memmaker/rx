package gridmap

import (
	"encoding/gob"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"path"
)

type TileDataOnDisk struct {
	TileChar      rune
	TileFg        color.RGBA
	TileBg        color.RGBA
	Description   string
	IsWalkable    bool
	IsTransparent bool
	IsExplored    bool
	Flags         TileFlags
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Save(directory string) error {
	metaData := fxtools.MustCreate(path.Join(directory, "metaData.rec"))
	defer metaData.Close()

	metaRecord := recfile.Record{
		recfile.Field{Name: "mapWidth", Value: recfile.IntStr(m.mapWidth)},
		recfile.Field{Name: "mapHeight", Value: recfile.IntStr(m.mapHeight)},

		recfile.Field{Name: "isOutdoor", Value: recfile.BoolStr(m.meta.IsOutdoor)},
		recfile.Field{Name: "friendlyName", Value: m.meta.DisplayName},
		recfile.Field{Name: "indoorAmbientLight", Value: m.meta.IndoorAmbientLight.EncodeAsString()},
		recfile.Field{Name: "musicFile", Value: m.meta.MusicFile},
	}

	locationRecords := make([]recfile.Record, 0)
	for name, location := range m.namedLocations {
		locationRecords = append(locationRecords, recfile.Record{
			recfile.Field{Name: "Name", Value: name},
			recfile.Field{Name: "Location", Value: location.Encode()},
		})
	}
	transitionRecords := make([]recfile.Record, 0)
	for pos, transition := range m.transitionMap {
		transitionRecords = append(transitionRecords, recfile.Record{
			recfile.Field{Name: "Location", Value: pos.Encode()},
			recfile.Field{Name: "TransitionToMap", Value: transition.TargetMap},
			recfile.Field{Name: "TransitionToLocation", Value: transition.TargetLocation},
		})
	}

	err := recfile.WriteMulti(metaData, map[string][]recfile.Record{
		"meta":        {metaRecord},
		"locations":   locationRecords,
		"transitions": transitionRecords,
	})
	if err != nil {
		return err
	}

	tilesOnDisk := make([]TileDataOnDisk, len(m.cells))
	for i, cell := range m.cells {
		tilesOnDisk[i] = TileDataOnDisk{
			TileChar:      cell.TileType.Icon.Char,
			TileFg:        cell.TileType.Icon.Fg,
			TileBg:        cell.TileType.Icon.Bg,
			Description:   cell.TileType.DefinedDescription,
			IsWalkable:    cell.TileType.IsWalkable,
			IsTransparent: cell.TileType.IsTransparent,
			Flags:         cell.TileType.Flags,
			IsExplored:    cell.IsExplored,
		}
	}

	cellFile := fxtools.MustCreate(path.Join(directory, "cells.bin"))
	defer cellFile.Close()
	gobber := gob.NewEncoder(cellFile)
	err = gobber.Encode(tilesOnDisk)
	if err != nil {
		return err
	}

	if len(m.allItems) > 0 {
		itemFile := fxtools.MustCreate(path.Join(directory, "items.bin"))
		defer itemFile.Close()
		gobber = gob.NewEncoder(itemFile)
		err = gobber.Encode(m.allItems)
		if err != nil {
			return err
		}
	}

	if len(m.allObjects) > 0 {
		objectFile := fxtools.MustCreate(path.Join(directory, "objects.bin"))
		defer objectFile.Close()
		gobber = gob.NewEncoder(objectFile)
		err = gobber.Encode(m.allObjects)
		if err != nil {
			return err
		}
	}

	if len(m.allActors) > 0 {
		actorFile := fxtools.MustCreate(path.Join(directory, "actors.bin"))
		defer actorFile.Close()
		gobber = gob.NewEncoder(actorFile)
		if err = gobber.Encode(m.allActors); err != nil {
			return err
		}
	}

	if len(m.allDownedActors) > 0 {
		downedActorFile := fxtools.MustCreate(path.Join(directory, "downedActors.bin"))
		defer downedActorFile.Close()
		gobber = gob.NewEncoder(downedActorFile)
		if err = gobber.Encode(m.allDownedActors); err != nil {
			return err
		}
	}

	if len(m.BakedLights) > 0 {
		lightFile := fxtools.MustCreate(path.Join(directory, "bakedLights.bin"))
		defer lightFile.Close()
		gobber = gob.NewEncoder(lightFile)
		if err = gobber.Encode(m.BakedLights); err != nil {
			return err
		}
	}

	return nil
}

func Load[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}](mapDirectory, mapName string) *GridMap[ActorType, ItemType, ObjectType] {

	directory := path.Join(mapDirectory, "maps", mapName)

	metaData := fxtools.MustOpen(path.Join(directory, "metaData.rec"))
	defer metaData.Close()
	metaRecords := recfile.ReadMulti(metaData)
	metaRecord := metaRecords["meta"][0]

	var mapWidth, mapHeight int
	var meta MapMeta
	for _, field := range metaRecord {
		switch field.Name {
		case "mapWidth":
			mapWidth = field.AsInt()
		case "mapHeight":
			mapHeight = field.AsInt()
		case "friendlyName":
			meta.DisplayName = field.Value
		case "indoorAmbientLight":
			meta.IndoorAmbientLight = fxtools.NewColorFromString(field.Value)
		case "musicFile":
			meta.MusicFile = field.Value
		case "isOutdoor":
			meta.IsOutdoor = field.AsBool()
		}
	}

	cellFile := fxtools.MustOpen(path.Join(directory, "cells.bin"))
	defer cellFile.Close()

	gobber := gob.NewDecoder(cellFile)
	var cells []TileDataOnDisk
	err := gobber.Decode(&cells)
	if err != nil {
		panic(err)
	}

	restoredMap := NewEmptyMap[ActorType, ItemType, ObjectType](mapWidth, mapHeight)
	for i, cell := range cells {
		restoredMap.SetCellByIndex(i, MapCell[ActorType, ItemType, ObjectType]{
			TileType: Tile{
				Icon: textiles.TextIcon{
					Char: cell.TileChar,
					Fg:   cell.TileFg,
					Bg:   cell.TileBg,
				},
				DefinedDescription: cell.Description,
				IsWalkable:         cell.IsWalkable,
				IsTransparent:      cell.IsTransparent,
				Flags:              cell.Flags,
			},
			IsExplored:    cell.IsExplored,
			Actor:         nil,
			DownedActor:   nil,
			Item:          nil,
			Object:        nil,
			BakedLighting: fxtools.HDRColor{},
		})
	}
	restoredMap.meta = meta
	restoredMap.name = mapName

	for _, record := range metaRecords["locations"] {
		var name string
		var location geometry.Point
		for _, field := range record {
			switch field.Name {
			case "Name":
				name = field.Value
			case "Location":
				location, _ = geometry.NewPointFromEncodedString(field.Value)
			}
		}
		restoredMap.namedLocations[name] = location
	}

	for _, record := range metaRecords["transitions"] {
		var pos geometry.Point
		var transition Transition
		for _, field := range record {
			switch field.Name {
			case "Location":
				pos, _ = geometry.NewPointFromEncodedString(field.Value)
			case "TransitionToMap":
				transition.TargetMap = field.Value
			case "TransitionToLocation":
				transition.TargetLocation = field.Value
			}
		}
		restoredMap.transitionMap[pos] = transition
	}

	if fxtools.FileExists(path.Join(directory, "items.bin")) {
		itemFile := fxtools.MustOpen(path.Join(directory, "items.bin"))
		defer itemFile.Close()
		gobber = gob.NewDecoder(itemFile)
		var items []ItemType
		err = gobber.Decode(&items)
		if err != nil {
			panic(err)
		}
		for _, item := range items {
			restoredMap.AddItem(item, item.Position())
		}
	}

	if fxtools.FileExists(path.Join(directory, "objects.bin")) {
		objectFile := fxtools.MustOpen(path.Join(directory, "objects.bin"))
		defer objectFile.Close()
		gobber = gob.NewDecoder(objectFile)
		var objects []ObjectType
		err = gobber.Decode(&objects)
		if err != nil {
			panic(err)
		}
		for _, object := range objects {
			restoredMap.AddObject(object, object.Position())
		}
	}

	if fxtools.FileExists(path.Join(directory, "actors.bin")) {
		actorFile := fxtools.MustOpen(path.Join(directory, "actors.bin"))
		defer actorFile.Close()
		gobber = gob.NewDecoder(actorFile)
		var actors []ActorType
		err = gobber.Decode(&actors)
		if err != nil {
			panic(err)
		}
		for _, actor := range actors {
			restoredMap.AddActor(actor, actor.Position())
		}
	}

	if fxtools.FileExists(path.Join(directory, "downedActors.bin")) {
		downedActorFile := fxtools.MustOpen(path.Join(directory, "downedActors.bin"))
		defer downedActorFile.Close()
		gobber = gob.NewDecoder(downedActorFile)
		var downedActors []ActorType
		err = gobber.Decode(&downedActors)
		if err != nil {
			panic(err)
		}
		for _, downedActor := range downedActors {
			restoredMap.AddDownedActor(downedActor, downedActor.Position())
		}
	}

	if fxtools.FileExists(path.Join(directory, "bakedLights.bin")) {
		lightFile := fxtools.MustOpen(path.Join(directory, "bakedLights.bin"))
		defer lightFile.Close()
		gobber = gob.NewDecoder(lightFile)
		var lights map[geometry.Point]*LightSource
		err = gobber.Decode(&lights)
		if err != nil {
			panic(err)
		}
		restoredMap.BakedLights = lights
	}

	return restoredMap
}
