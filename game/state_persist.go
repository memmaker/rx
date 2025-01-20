package game

import (
	"contractor/foundation"
	"contractor/gridmap"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"os"
	"path/filepath"
	"strings"
)

func (g *GameState) Save(directory string) error {
	os.MkdirAll(directory, os.ModePerm)
	// Global game state
	globalRecord := recfile.Record{
		recfile.Field{Name: "CurrentMap", Value: g.currentMapName},
		recfile.Field{Name: "TurnsTaken", Value: recfile.IntStr(g.TurnsTaken())},
		recfile.Field{Name: "ActorID", Value: recfile.UInt64Str(uint64(gridmap.PeekAtNextActorID()))},
		recfile.Field{Name: "ItemID", Value: recfile.UInt64Str(uint64(gridmap.PeekAtNextItemID()))},
		recfile.Field{Name: "GameTime", Value: recfile.TimeStr(g.gameTime.Time)},
		recfile.Field{Name: "ShowEverything", Value: recfile.BoolStr(g.showEverything)},
	}
	globalFile := fxtools.MustCreate(filepath.Join(directory, "global.rec"))
	err := recfile.WriteMulti(globalFile, map[string][]recfile.Record{
		"global": {globalRecord},
		"flags":  g.gameFlags.ToRecord(),
	})
	if err != nil {
		return err
	}
	err = globalFile.Close()
	if err != nil {
		return err
	}

	// Journal
	journalRecords := g.journal.ToRecords()
	journalFile := fxtools.MustCreate(filepath.Join(directory, "journal.rec"))
	err = recfile.WriteMulti(journalFile, journalRecords)
	if err != nil {
		return err
	}
	err = journalFile.Close()
	if err != nil {
		return err
	}

	// Loaded Map States
	mapDirectory := filepath.Join(directory, "maps")
	for mapName, gameMap := range g.activeMaps {
		mapDirName := filepath.Join(mapDirectory, mapName)
		os.MkdirAll(mapDirName, os.ModePerm)
		err = gameMap.Save(mapDirName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *GameState) Load(directory string) {
	// Global game state
	globalFile := fxtools.MustOpen(filepath.Join(directory, "global.rec"))
	globalRecords, _ := recfile.ReadMulti(globalFile)
	globalFile.Close()

	globalRecord := globalRecords["global"][0]

	for _, field := range globalRecord {
		switch strings.ToLower(field.Name) {
		case "currentmap":
			g.currentMapName = field.Value
		case "turnstaken":
			g.gameTime = g.gameTime.WithTurns(recfile.StrInt(field.Value))
		case "gametime":
			g.gameTime = g.gameTime.WithTime(recfile.StrTime(field.Value))
		case "showeverything":
			g.showEverything = recfile.StrBool(field.Value)
		case "actorid":
			gridmap.SetNextActorID(gridmap.ActorID(recfile.StrUInt64(field.Value)))
		case "itemid":
			gridmap.SetNextItemID(gridmap.ItemID(recfile.StrUInt64(field.Value)))
		}
	}

	flagRecords := globalRecords["flags"]
	if len(flagRecords) > 0 {
		g.gameFlags = fxtools.NewStringFlagsFromRecord(flagRecords)
	}
	g.logBuffer = make([]foundation.HiLiteString, 0)

	// Journal
	journalFile := fxtools.MustOpen(filepath.Join(directory, "journal.rec"))
	journalRecords, _ := recfile.ReadMulti(journalFile)
	journalFile.Close()
	g.journal = NewJournalFromRecords(journalRecords, g.GetScriptFuncs())

	// Loaded Map States
	mapEntries, err := os.ReadDir(filepath.Join(directory, "maps"))
	if err != nil {
		panic(err)
	}
	loadedMaps := make(map[string]*gridmap.GridMap[*Actor, foundation.Item, Object])
	for _, mapEntry := range mapEntries {
		if mapEntry.IsDir() {
			mapName := mapEntry.Name()
			gameMap := gridmap.Load[*Actor, foundation.Item, Object](directory, mapName)
			for _, actor := range gameMap.Actors() {
				// TODO: restore poshandlers for inventory items and other attach stuff, fsm.RestoreState
				actor.InitWithGameState(g)
			}
			for _, obj := range gameMap.Objects() {
				obj.InitWithGameState(g)
			}
			loadedMaps[mapName] = gameMap
		}
	}
	g.activeMaps = loadedMaps

	filteredActors := g.currentMap().GetFilteredActors(func(actor *Actor) bool {
		return actor.GetInternalName() == "player"
	})
	g.Player = filteredActors[0]

	// Restore missing glue
	g.hookupJournalAndFlags()
	g.playerAttachHooks()

	// CheckAndRunFrames lights & player position
	g.currentMap().UpdateBakedLights()
	g.afterPlayerMoved(geometry.Point{}, true)

	g.updateUIStatus()
}
