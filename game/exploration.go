package game

import (
	"RogueUI/dungen"
	"RogueUI/geometry"
)

func (g *GameState) applyCorridorExploration() {
	// in corridor = always dark, light the corridor/doors while walking
	corridorTiles := g.gridMap.GetFilteredNeighbors(g.Player.Position(), func(pos geometry.Point) bool {
		return g.gridMap.Contains(pos) && (g.dungeonLayout.IsCorridor(pos) || g.dungeonLayout.IsDoorAt(pos))
	})
	corridorTiles = append(corridorTiles, g.Player.Position())
	for _, pos := range corridorTiles {
		g.gridMap.SetExplored(pos)
		g.gridMap.SetLit(pos, true)
	}
}

func (g *GameState) applyLightRoomExploration(playerRoom *dungen.DungeonRoom) {
	allTiles := playerRoom.GetAbsoluteRoomTiles()
	g.gridMap.SetListExplored(allTiles, true)
}

func (g *GameState) applyDarkRoomExploration() {
	g.gridMap.SetSelfAndAllNeighborsExplored(g.Player.Position())
	wallTiles := g.gridMap.GetFilteredNeighbors(g.Player.Position(), func(pos geometry.Point) bool {
		return g.dungeonLayout.IsWallAt(pos) || g.dungeonLayout.IsDoorAt(pos)
	})
	for _, pos := range wallTiles {
		g.gridMap.SetLit(pos, true)
	}
}
