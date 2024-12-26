package game

import (
	"contractor/foundation"
	"contractor/gridmap"
)

func (g *GameState) ensureMapIsLoaded(levelName string) *gridmap.GridMap[*Actor, foundation.Item, Object] {
	if _, ok := g.activeMaps[levelName]; !ok {

		// actually load a map
		result := g.mapLoader.LoadMap(levelName)

		if result.Map == nil {
			g.msg(foundation.Msg("It's impossible to move there.."))
			return nil
		}

		g.activeMaps[levelName] = result.Map

		g.setFlagsAndRunScripts(result)
	}
	return g.activeMaps[levelName]
}

func (g *GameState) setFlagsAndRunScripts(result gridmap.MapLoadResult[*Actor, foundation.Item, Object]) {
	flags := result.FlagsOfMap
	for flagName, flagValue := range flags {
		g.gameFlags.Set(flagName, flagValue)
	}

	scripts := result.ScriptsToRun
	for _, script := range scripts {
		g.RunScriptOnMap(result.Map.GetName(), script)
	}
}
