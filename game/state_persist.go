package game

import (
	"RogueUI/gridmap"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"os"
	"path"
	"strings"
)

func (g *GameState) Save(directory string) error {
	os.MkdirAll(directory, os.ModePerm)
	// Global game state
	globalRecord := recfile.Record{
		recfile.Field{Name: "CurrentMap", Value: g.currentMapName},
		recfile.Field{Name: "TurnsTaken", Value: recfile.IntStr(g.TurnsTaken)},
		recfile.Field{Name: "GameTime", Value: recfile.TimeStr(g.gameTime)},
		recfile.Field{Name: "ShowEverything", Value: recfile.BoolStr(g.showEverything)},
	}
	globalFile := fxtools.MustCreate(path.Join(directory, "global.rec"))
	err := recfile.WriteMulti(globalFile, map[string][]recfile.Record{
		"global": {globalRecord},
		"flags":  {g.gameFlags.ToRecord()},
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
	journalFile := fxtools.MustCreate(path.Join(directory, "journal.rec"))
	err = recfile.WriteMulti(journalFile, journalRecords)
	if err != nil {
		return err
	}
	err = journalFile.Close()
	if err != nil {
		return err
	}

	// Loaded Map States
	mapDirectory := path.Join(directory, "maps")
	for mapName, gameMap := range g.activeMaps {
		mapDirName := path.Join(mapDirectory, mapName)
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
	globalFile := fxtools.MustOpen(path.Join(directory, "global.rec"))
	globalRecords := recfile.ReadMulti(globalFile)
	globalFile.Close()

	globalRecord := globalRecords["global"][0]
	for _, field := range globalRecord {
		switch strings.ToLower(field.Name) {
		case "currentmap":
			g.currentMapName = field.Value
		case "turnstaken":
			g.TurnsTaken = recfile.StrInt(field.Value)
		case "gametime":
			g.gameTime = recfile.StrTime(field.Value)
		case "showeverything":
			g.showEverything = recfile.StrBool(field.Value)
		}
	}

	flagRecords := globalRecords["flags"]
	if len(flagRecords) > 0 {
		g.gameFlags = fxtools.NewStringFlagsFromRecord(flagRecords[0])
	}

	// Journal
	journalFile := fxtools.MustOpen(path.Join(directory, "journal.rec"))
	journalRecords := recfile.ReadMulti(journalFile)
	journalFile.Close()
	g.journal = NewJournalFromRecords(journalRecords, g.getConditionFuncs())

	// Loaded Map States
	mapEntries, err := os.ReadDir(path.Join(directory, "maps"))
	if err != nil {
		panic(err)
	}
	loadedMaps := make(map[string]*gridmap.GridMap[*Actor, *Item, Object])
	for _, mapEntry := range mapEntries {
		if mapEntry.IsDir() {
			mapName := mapEntry.Name()
			mapFileName := path.Join(directory, "maps", mapName)
			gameMap := gridmap.Load[*Actor, *Item, Object](mapFileName)
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

	g.hookupJournalAndFlags()
	g.attachHooksToPlayer()

	g.currentMap().UpdateBakedLights()
	g.currentMap().UpdateDynamicLights()
}
