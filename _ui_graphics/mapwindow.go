package ui_graphics

import (
    "github.com/memmaker/go/geometry"
    "math"
)

type MapWindow struct {
    lookup       func(x, y int, tick uint64) []CellDrawInfo
    tileScale    func() float64
    windowSize   geometry.Point // in pixels
    mapSize      geometry.Point // in cells
    scrollOffset geometry.Point // in pixels - offset of the top left corner of the screen
    gridSize     geometry.Point // in pixels - size of a cell
}

func (m *MapWindow) GetTileScale() float64 {
    return m.tileScale()
}

func NewMapWindow(
    windowSize geometry.Point,
    mapSize geometry.Point,
    gridSize geometry.Point,
    getTileScale func() float64,
    lookup func(x, y int, tick uint64) []CellDrawInfo,
) *MapWindow {
    return &MapWindow{
        lookup:     lookup,
        tileScale:  getTileScale,
        windowSize: windowSize,
        gridSize:   gridSize,
        mapSize:    mapSize,
    }
}

func (m *MapWindow) GetMapSizeInPixels() geometry.Point {
    mapSize := m.mapSize
    tileScale := m.tileScale()
    sizeInPixels := geometry.Point{
        X: int(float64(mapSize.X*m.gridSize.X) * tileScale),
        Y: int(float64(mapSize.Y*m.gridSize.Y) * tileScale),
    }
    return sizeInPixels
}
func (m *MapWindow) GetGridSize() geometry.Point {
    return m.gridSize
}

func (m *MapWindow) GetTextureIndexAt(cellX, cellY int, tick uint64) []CellDrawInfo {
    return m.lookup(cellX, cellY, tick)
}

func (m *MapWindow) GetScrollOffset() geometry.Point {
    return m.scrollOffset
}

func (m *MapWindow) ScrollBy(point geometry.Point) {
    newScrollX := m.scrollOffset.X + point.X
    newScrollY := m.scrollOffset.Y + point.Y
    m.setScrollOffset(newScrollX, newScrollY)
}

func (m *MapWindow) CenterOn(mapPos geometry.Point) {
    pixelPos := m.PixelOffsetCenter(mapPos)
    newScrollX := pixelPos.X - m.windowSize.X/2
    newScrollY := pixelPos.Y - m.windowSize.Y/2
    m.setScrollOffset(newScrollX, newScrollY)
}

func (m *MapWindow) setScrollOffset(newScrollX, newScrollY int) {
    mapSize := m.GetMapSizeInPixels()
    scaledWindowSize := m.windowSize

    shouldScrollHorizontally := mapSize.X > scaledWindowSize.X
    shouldScrollVertically := mapSize.Y > scaledWindowSize.Y

    if shouldScrollHorizontally {
        if newScrollX > mapSize.X-scaledWindowSize.X {
            newScrollX = mapSize.X - scaledWindowSize.X
        } else if newScrollX < 0 {
            newScrollX = 0
        }
    }

    if shouldScrollVertically {
        if newScrollY > mapSize.Y-scaledWindowSize.Y {
            newScrollY = mapSize.Y - scaledWindowSize.Y
        } else if newScrollY < 0 {
            newScrollY = 0
        }
    }

    m.scrollOffset = geometry.Point{X: newScrollX, Y: newScrollY}
}

func (m *MapWindow) GetVisibleMap() geometry.Rect {

    scrollOffset := m.GetScrollOffset()

    gridSize := m.GetGridSize()
    tileScale := m.GetTileScale()

    scaledTileSize := gridSize.ToPointF().Mul(tileScale)

    firstTileX := int(float64(scrollOffset.X) / scaledTileSize.X)
    firstTileY := int(float64(scrollOffset.Y) / scaledTileSize.Y)

    screenSize := m.GetWindowSizeInPixels()

    screenWidth := screenSize.X
    screenHeight := screenSize.Y
    tileCountX := int(math.Ceil(float64(screenWidth) / float64(scaledTileSize.X)))
    tileCountY := int(math.Ceil(float64(screenHeight) / float64(scaledTileSize.Y)))

    return geometry.Rect{
        Min: geometry.Point{X: firstTileX, Y: firstTileY},
        Max: geometry.Point{X: firstTileX + tileCountX + 1, Y: firstTileY + tileCountY + 1},
    }
}
func (m *MapWindow) IsMapCellVisible(cell geometry.Point) bool {
    visibleMap := m.GetVisibleMap()
    return visibleMap.Contains(cell)
}
func (m *MapWindow) MapToScreen(mapPosition geometry.Point) geometry.Point {
    inPixel := m.PixelOffsetTopLeft(mapPosition)
    scrollOffset := m.GetScrollOffset()
    return inPixel.Sub(scrollOffset)
}

func (m *MapWindow) MapFloatToScreen(mapPosition geometry.PointF) geometry.Point {
    tileSize := m.gridSize.ToPointF().Mul(m.tileScale())
    screenPos := geometry.Point{
        X: int(mapPosition.X * float64(tileSize.X)),
        Y: int(mapPosition.Y * float64(tileSize.Y)),
    }
    scrollOffset := m.GetScrollOffset()
    return screenPos.Sub(scrollOffset)
}

func (m *MapWindow) PixelOffsetTopLeft(mapPosition geometry.Point) geometry.Point {
    tileSize := m.gridSize.ToPointF().Mul(m.tileScale())
    screenPos := geometry.Point{
        X: int(float64(mapPosition.X) * float64(tileSize.X)),
        Y: int(float64(mapPosition.Y) * float64(tileSize.Y)),
    }
    return screenPos
}
func (m *MapWindow) PixelOffsetCenter(mapPosition geometry.Point) geometry.Point {
    tileSize := m.gridSize.ToPointF().Mul(m.tileScale())
    screenPos := geometry.Point{
        X: int((float64(mapPosition.X) + 0.5) * float64(tileSize.X)),
        Y: int((float64(mapPosition.Y) + 0.5) * float64(tileSize.Y)),
    }
    return screenPos
}
func (m *MapWindow) GetMapCellAtScreenPos(pixels geometry.Point) geometry.Point {
    gridSize := m.gridSize.ToPointF().Mul(m.tileScale())
    return geometry.Point{
        X: int(float64(pixels.X+m.scrollOffset.X) / gridSize.X),
        Y: int(float64(pixels.Y+m.scrollOffset.Y) / gridSize.Y),
    }
}

func (m *MapWindow) GetExactMapPositionFromScreenPos(pixels geometry.Point) geometry.PointF {
    gridSize := m.gridSize.ToPointF().Mul(m.tileScale())
    return geometry.PointF{
        X: float64(pixels.X+m.scrollOffset.X) / gridSize.X,
        Y: float64(pixels.Y+m.scrollOffset.Y) / gridSize.Y,
    }
}
func (m *MapWindow) GetWindowSizeInPixels() geometry.Point {
    return m.windowSize
}

func (m *MapWindow) CenterOnFloat(pos geometry.PointF) {
    gridSize := m.gridSize.ToPointF().Mul(m.tileScale())
    screenPos := geometry.PointF{
        X: (pos.X * gridSize.X) - float64(m.windowSize.X/2),
        Y: (pos.Y * gridSize.Y) - float64(m.windowSize.Y/2),
    }
    m.setScrollOffset(int(screenPos.X), int(screenPos.Y))
}

func (m *MapWindow) OnScreenSizeChanged(newWindowSize geometry.Point) {
    m.windowSize = newWindowSize
}

func (m *MapWindow) ResetScrolling() {
    m.scrollOffset = geometry.Point{}
}
