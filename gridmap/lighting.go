package gridmap

import (
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"strconv"
	"strings"
	"time"
)

type LightSource struct {
	Pos          geometry.Point
	Radius       int
	Color        fxtools.HDRColor
	MaxIntensity float64
}

func (s LightSource) ToRecord() []recfile.Field {
	return []recfile.Field{
		{Name: "Position", Value: s.Pos.String()},
		{Name: "Radius", Value: strconv.Itoa(s.Radius)},
		{Name: "Color", Value: s.Color.EncodeAsString()},
		{Name: "Max_Intensity", Value: strconv.FormatFloat(s.MaxIntensity, 'f', 2, 64)},
	}
}

func NewLightSourceFromRecord(record []recfile.Field) *LightSource {
	var result LightSource
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "position":
			result.Pos, _ = geometry.NewPointFromString(field.Value)
		case "radius":
			result.Radius, _ = strconv.Atoi(field.Value)
		case "color":
			result.Color = fxtools.NewColorFromString(field.Value)
		case "max_intensity":
			result.MaxIntensity, _ = strconv.ParseFloat(field.Value, 64)
		}
	}
	return &result
}

// AddDynamicLightSource adds a light source to the map. It will automatically call UpdateDynamicLights.
func (m *GridMap[ActorType, ItemType, ObjectType]) AddDynamicLightSource(pos geometry.Point, light *LightSource) {
	if m.IsDynamicLightSource(pos) {
		return
	}
	m.DynamicLights[pos] = light
	light.Pos = pos
}

// AddBakedLightSource adds a light source to the map. It will automatically call UpdateBakedLights and UpdateDynamicLights.
func (m *GridMap[ActorType, ItemType, ObjectType]) AddBakedLightSource(pos geometry.Point, light *LightSource) {
	if m.IsBakedLightSource(pos) {
		return
	}
	m.BakedLights[pos] = light
}
func (m *GridMap[ActorType, ItemType, ObjectType]) SetAmbientLight(color fxtools.HDRColor) {
	m.meta = m.meta.WithAmbientLight(color)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsDynamicLightSource(pos geometry.Point) bool {
	_, ok := m.DynamicLights[pos]
	return ok
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsBakedLightSource(pos geometry.Point) bool {
	_, ok := m.BakedLights[pos]
	return ok
}

// MoveLightSource moves a light source to a new position. It will automatically call UpdateDynamicLights.
func (m *GridMap[ActorType, ItemType, ObjectType]) MoveLightSource(lightSource *LightSource, to geometry.Point) {
	if m.IsDynamicLightSource(to) {
		return
	}
	delete(m.DynamicLights, lightSource.Pos)
	lightSource.Pos = to
	m.DynamicLights[to] = lightSource
	m.UpdateDynamicLights()
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IndoorLightAt(p geometry.Point) fxtools.HDRColor {
	bakedLighting := m.cells[p.X+p.Y*m.mapWidth].BakedLighting
	if m.meta.IndoorAmbientLight.Brightness() > bakedLighting.Brightness() {
		bakedLighting = m.meta.IndoorAmbientLight
	}

	if dynamicLightAt, ok := m.dynamicallyLitCells[p]; ok {
		if dynamicLightAt.Brightness() > bakedLighting.Brightness() {
			return dynamicLightAt
		}
		return bakedLighting
	}
	return bakedLighting
}

func (m *GridMap[ActorType, ItemType, ObjectType]) OutdoorLightAt(p geometry.Point, timeOfDay time.Time) fxtools.HDRColor {
	ambientLight := GetAmbientLightFromDayTime(timeOfDay)
	bakedLighting := m.cells[p.X+p.Y*m.mapWidth].BakedLighting
	if ambientLight.Brightness() > bakedLighting.Brightness() {
		bakedLighting = ambientLight
	}
	if dynamicLightAt, ok := m.dynamicallyLitCells[p]; ok {
		if dynamicLightAt.Brightness() > bakedLighting.Brightness() {
			return dynamicLightAt
		}
		return bakedLighting
	}
	return bakedLighting
}

func GetAmbientLightFromDayTime(timeOfDay time.Time) fxtools.HDRColor {
	morning := fxtools.HDRColor{R: 0.4, G: 0.368, B: 0.3466666666666667, A: 1.0}
	noon := fxtools.HDRColor{R: 1.8, G: 1.7966666666666666, B: 1.7666666666666666, A: 1.0}
	evening := fxtools.HDRColor{R: 1.4177777777777778, G: 1.4177777777777778, B: 1.46666666666666666, A: 1.0}
	night := fxtools.HDRColor{R: 0.2111111111111111, G: 0.2111111111111111, B: 0.33333333333333333, A: 1.0}

	secondsSinceMidnight := timeOfDay.Hour()*3600 + timeOfDay.Minute()*60 + timeOfDay.Second()

	// we need to determine the percentage of the interval between the two times
	// for example, if it's 9:00, we need to know how far we are between 6:00 and 12:00
	// 9:00 is 50% of the way between 6:00 and 12:00
	// BUT: we want to be precise, so we need to know how many seconds are in the interval
	// 6:00 - 12:00 is 6 hours, so 6 * 3600 = 21600 seconds
	// that means we got these intervals:
	// 0:00 - 6:00 = 0 - 21600
	// 6:00 - 12:00 = 21600 - 43200
	// 12:00 - 18:00 = 43200 - 64800
	// 18:00 - 24:00 = 64800 - 86400

	var ambientLightColor fxtools.HDRColor
	startColor := morning
	endColor := noon
	// 6:00 - 12:00
	if secondsSinceMidnight >= 21600 && secondsSinceMidnight < 43200 {
		intervalPercentage := float64(secondsSinceMidnight-21600) / 21600.0
		ambientLightColor = startColor.Lerp(endColor, intervalPercentage)
	} else if secondsSinceMidnight >= 43200 && secondsSinceMidnight < 64800 {
		startColor = noon
		endColor = evening
		intervalPercentage := float64(secondsSinceMidnight-43200) / 21600.0
		ambientLightColor = startColor.Lerp(endColor, intervalPercentage)
	} else if secondsSinceMidnight >= 64800 && secondsSinceMidnight < 86400 {
		startColor = evening
		endColor = night
		intervalPercentage := float64(secondsSinceMidnight-64800) / 21600.0
		ambientLightColor = startColor.Lerp(endColor, intervalPercentage)
	} else if secondsSinceMidnight >= 0 && secondsSinceMidnight < 21600 {
		startColor = night
		endColor = morning
		intervalPercentage := float64(secondsSinceMidnight) / 21600.0
		ambientLightColor = startColor.Lerp(endColor, intervalPercentage)
	}
	return ambientLightColor
}

// we use the value stored in cell.Lighting for lighting the tile later on..
func (m *GridMap[ActorType, ItemType, ObjectType]) UpdateDynamicLights() {
	for key, _ := range m.dynamicallyLitCells {
		delete(m.dynamicallyLitCells, key)
	}
	if len(m.DynamicLights) == 0 {
		return
	}
	setLightAt := func(point geometry.Point, light fxtools.HDRColor) {
		if light.R > 0 || light.G > 0 || light.B > 0 {
			m.dynamicallyLitCells[point] = light
		}
	}
	m.updateLightMap(m.DynamicLights, setLightAt)
	m.DynamicLightsChanged = false
}

func (m *GridMap[ActorType, ItemType, ObjectType]) UpdateBakedLights() {
	setLightAt := func(point geometry.Point, light fxtools.HDRColor) {
		m.cells[point.X+point.Y*m.mapWidth].BakedLighting = light
	}
	m.updateLightMap(m.BakedLights, setLightAt)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) updateLightMap(lightSources map[geometry.Point]*LightSource, setLightAt func(p geometry.Point, light fxtools.HDRColor)) {
	lightAt := make(map[geometry.Point]fxtools.HDRColor)
	isTransparent := func(p geometry.Point) bool {
		if lightSources[p] != nil {
			return true
		}
		return m.IsTransparent(p) && !m.IsActorAt(p)
	}
	for _, lightSource := range lightSources {
		for _, nodePos := range m.lightfov.SSCVisionMap(lightSource.Pos, lightSource.Radius, true, isTransparent) {

			//for _, node := range m.lightfov.LightMap(&MapLighter[VictimType, ItemType, ObjectType]{gridmap: m, sources: lightSources}, []geometry.Point{lightSource.Pos}) {
			pos := nodePos
			dist := geometry.Distance(lightSource.Pos, pos)
			//pos := node.P
			//dist := node.Cost
			if dist < 0 {
				dist = 0
			}
			if _, hasValue := lightAt[pos]; !hasValue {
				lightAt[pos] = fxtools.HDRColor{R: 0, G: 0, B: 0, A: 1}
			}
			colorOfLight := lightAt[pos]

			//intensityWithFalloff := fxtools.Clamp(0, lightSource.MaxIntensity, (float64(lightSource.Radius)-dist)/dist)
			// flat intensity in radius > lightSource.MaxIntensity
			flatIntensity := lightSource.MaxIntensity
			if dist > float64(lightSource.Radius) {
				flatIntensity = 0
			}
			sourceLightColor := lightSource.Color.MultiplyWithScalar(flatIntensity)
			lightAt[pos] = colorOfLight.Add(sourceLightColor)
		}
	}
	for pos, light := range lightAt {
		setLightAt(pos, light)
	}
}

type MapLighter[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}] struct {
	gridmap *GridMap[ActorType, ItemType, ObjectType]
	sources map[geometry.Point]*LightSource
}

func (m *MapLighter[ActorType, ItemType, ObjectType]) Cost(src geometry.Point, from geometry.Point, to geometry.Point) float64 {
	if src == from {
		return 1
		return geometry.Distance(from, to)
	}
	currentMap := m.gridmap
	switch {
	case !currentMap.cells[to.Y*currentMap.mapWidth+to.X].TileType.IsTransparent || !currentMap.cells[from.Y*currentMap.mapWidth+from.X].TileType.IsTransparent:
		return 1000
	case currentMap.IsActorAt(from):
		return geometry.Distance(from, to) + 2
	}

	return geometry.Distance(from, to)
}

// needed for lighting
func (m *MapLighter[ActorType, ItemType, ObjectType]) MaxCost(src geometry.Point) float64 {
	if light, ok := m.sources[src]; ok {
		return float64(light.Radius) + 0.5
	}
	return 0
}
