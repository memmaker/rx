package game

import (
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
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
		recfile.Field{Name: "TurnsTaken", Value: recfile.IntStr(g.TurnsTaken())},
		recfile.Field{Name: "GameTime", Value: recfile.TimeStr(g.gameTime.Time)},
		recfile.Field{Name: "ShowEverything", Value: recfile.BoolStr(g.showEverything)},
	}
	globalFile := fxtools.MustCreate(path.Join(directory, "global.rec"))
	err := recfile.WriteMulti(globalFile, map[string][]recfile.Record{
		"global":           {globalRecord},
		"flags":            g.gameFlags.ToRecord(),
		"terminal_guesses": g.terminalGuessesToRecords(),
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
			g.gameTime = g.gameTime.WithTurns(recfile.StrInt(field.Value))
		case "gametime":
			g.gameTime = g.gameTime.WithTime(recfile.StrTime(field.Value))
		case "showeverything":
			g.showEverything = recfile.StrBool(field.Value)
		}
	}

	flagRecords := globalRecords["flags"]
	if len(flagRecords) > 0 {
		g.gameFlags = fxtools.NewStringFlagsFromRecord(flagRecords)
	}
	g.logBuffer = make([]foundation.HiLiteString, 0)
	g.terminalGuesses = g.terminalGuessesFromRecords(globalRecords["terminal_guesses"])

	// Journal
	journalFile := fxtools.MustOpen(path.Join(directory, "journal.rec"))
	journalRecords := recfile.ReadMulti(journalFile)
	journalFile.Close()
	g.journal = NewJournalFromRecords(journalRecords, g.getScriptFuncs())

	// Loaded Map States
	mapEntries, err := os.ReadDir(path.Join(directory, "maps"))
	if err != nil {
		panic(err)
	}
	loadedMaps := make(map[string]*gridmap.GridMap[*Actor, foundation.Item, Object])
	for _, mapEntry := range mapEntries {
		if mapEntry.IsDir() {
			mapName := mapEntry.Name()
			gameMap := gridmap.Load[*Actor, foundation.Item, Object](directory, mapName)
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
	g.iconsForObjects = gridmap.LoadIconsForObjects(path.Join(g.config.DataRootDir, "maps", g.currentMapName), g.palette)

	g.hookupJournalAndFlags()
	g.attachHooksToPlayer()

	// CheckAndRunFrames lights & player position
	g.currentMap().UpdateBakedLights()
	g.afterPlayerMoved(geometry.Point{}, true)

	g.updateUIStatus()
}

func (g *GameState) terminalGuessesToRecords() []recfile.Record {
	var recs []recfile.Record
	for key, values := range g.terminalGuesses {
		record := recfile.Record{
			recfile.Field{Name: "terminal", Value: key},
		}
		for _, value := range values {
			record = append(record, recfile.Field{Name: "guess", Value: value})
		}
		recs = append(recs, record)
	}
	return recs
}

func (g *GameState) terminalGuessesFromRecords(records []recfile.Record) map[string][]string {
	result := make(map[string][]string)
	for _, record := range records {
		var terminal string
		var guesses []string
		for _, field := range record {
			if field.Name == "terminal" {
				terminal = field.Value
			} else if field.Name == "guess" {
				guesses = append(guesses, field.Value)
			}
		}
		result[terminal] = guesses
	}
	return result
}
