package dungen

import (
	"RogueUI/geometry"
	"math/rand"
)

type VaultGenerator struct {
	random           *rand.Rand
	mapWidth         int
	mapHeight        int
	minRoomSize      int
	maxRoomSize      int
	minCorridorWidth int
	maxCorridorWidth int
}

func NewVaultGenerator(random *rand.Rand, mapWidth, mapHeight int) *VaultGenerator {
	return &VaultGenerator{
		random:           random,
		mapWidth:         mapWidth,
		mapHeight:        mapHeight,
		minRoomSize:      3,
		maxRoomSize:      min(mapWidth, mapHeight),
		minCorridorWidth: 2,
		maxCorridorWidth: 5,
	}
}

func (v *VaultGenerator) Generate() *DungeonMap {
	emptyMap := NewDungeonMap(v.mapWidth, v.mapHeight)
	v.carveCorridor(emptyMap)
	return emptyMap
}

func (v *VaultGenerator) carveCorridor(emptyMap *DungeonMap) {
	//startLocation, startDirection := v.chooseValidCorridorStartLocation()
	//minMargin := v.minRoomSize + v.minCorridorWidth
	/*
	   safeZone := geometry.NewRect(minMargin, minMargin, v.mapWidth-minMargin, v.mapHeight-minMargin)
	   currentLocation := startLocation
	   currentDirection := startDirection
	   xBound := 0
	   if currentDirection.X > 0 {
	       xBound = v.mapWidth - minMargin
	   } else if currentDirection.X < 0 {
	       xBound = minMargin
	   }

	*/
}

func (v *VaultGenerator) chooseValidCorridorStartLocation() (geometry.Point, geometry.Point) {
	// decide for each axis whether to start at the border of the map
	// if we don't start at the border, we have to keep a margin of
	// at least minRoomSize + minCorridorWidth
	minMargin := v.minRoomSize + v.minCorridorWidth
	var x, y int
	var xDirection, yDirection int
	if v.random.Intn(2) == 0 { // start the x axis on the border
		if v.random.Intn(2) == 0 { // start at the left border
			x = 0
			xDirection = 1
		} else { // start at the right border
			x = v.mapWidth - 1
			xDirection = -1
		}
	} else { // start the x axis not on the border
		x = v.random.Intn(v.mapWidth-2*minMargin) + minMargin
		if x < v.mapWidth/2 {
			xDirection = 1
		} else {
			xDirection = -1
		}
	}

	if v.random.Intn(2) == 0 { // start the y axis on the border
		if v.random.Intn(2) == 0 { // start at the top border
			y = 0
			yDirection = 1
		} else { // start at the bottom border
			y = v.mapHeight - 1
			yDirection = -1
		}
	} else { // start the y axis not on the border
		y = v.random.Intn(v.mapHeight-2*minMargin) + minMargin
		if y < v.mapHeight/2 {
			yDirection = 1
		} else {
			yDirection = -1
		}
	}

	return geometry.Point{X: x, Y: y}, geometry.Point{X: xDirection, Y: yDirection}
}

func (v *VaultGenerator) carveCorridorBetween(emptyMap *DungeonMap, one geometry.Point, two geometry.Point) {
	x, y := one.X, one.Y
	dx := two.X - x
	dy := two.Y - y
	emptyMap.SetCorridor(x, y)
	emptyMap.SetCorridor(two.X, two.Y)
	goHorizontalFirst := dx > dy
	if goHorizontalFirst {
		for x != two.X {
			x += geometry.Sign(dx)
			emptyMap.SetCorridor(x, y)
		}
		for y != two.Y {
			y += geometry.Sign(dy)
			emptyMap.SetCorridor(x, y)
		}
	} else {
		for y != two.Y {
			y += geometry.Sign(dy)
			emptyMap.SetCorridor(x, y)
		}
		for x != two.X {
			x += geometry.Sign(dx)
			emptyMap.SetCorridor(x, y)
		}
	}
}
