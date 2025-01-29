package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
    "image/color"
    "math"
)

type MapRenderer struct {
    mapWindow       *MapWindow
    gridRenderer    *TileRenderer
    getMiniMapColor func(mapPos geometry.Point, outTileColor *[4]byte)
    miniMapImage    *ebiten.Image
}

func (r *MapRenderer) MeasureString(text string) (float64, float64) {
    return r.gridRenderer.MeasureString(text)
}

func NewMapRenderer(gridRenderer *TileRenderer, mapWindow *MapWindow) *MapRenderer {
    m := &MapRenderer{
        mapWindow:    mapWindow,
        gridRenderer: gridRenderer,
        getMiniMapColor: func(mapPos geometry.Point, outTileColor *[4]byte) {
            outTileColor[0] = 255
            outTileColor[1] = 255
            outTileColor[2] = 255
            outTileColor[3] = 255
        },
    }
    m.initMinimap()
    return m
}

func (r *MapRenderer) SetMiniMapColorHandler(textureIndexToMinimapColor func(mapPos geometry.Point, outTileColor *[4]byte)) {
    r.getMiniMapColor = textureIndexToMinimapColor
}

func (r *MapRenderer) GetVisibleMap() geometry.Rect {
    return r.mapWindow.GetVisibleMap()
}

func (r *MapRenderer) Draw(tick uint64) {
    scrollOffset := r.mapWindow.GetScrollOffset()

    gridSize := r.mapWindow.GetGridSize()
    tileScale := r.mapWindow.GetTileScale()

    //scaledTileSize := gridSize.MulF(tileScale)
    scaledTileSizeX := float64(gridSize.X) * tileScale
    scaledTileSizeY := float64(gridSize.Y) * tileScale

    firstTileX := int(float64(scrollOffset.X) / scaledTileSizeX)
    firstTileY := int(float64(scrollOffset.Y) / scaledTileSizeY)

    screenSize := r.mapWindow.GetWindowSizeInPixels()

    screenWidth := screenSize.X
    screenHeight := screenSize.Y
    tileCountX := int(math.Ceil(float64(screenWidth) / float64(scaledTileSizeX)))
    tileCountY := int(math.Ceil(float64(screenHeight) / float64(scaledTileSizeY)))

    drawOffsetFromEdgeX := (scrollOffset.X % int(scaledTileSizeX)) * -1
    drawOffsetFromEdgeY := (scrollOffset.Y % int(scaledTileSizeY)) * -1

    for yStep := 0; yStep < tileCountY; yStep++ {
        for xStep := 0; xStep < tileCountX; xStep++ {
            mapPosX := firstTileX + xStep
            mapPosY := firstTileY + yStep
            drawScreenX := float64(drawOffsetFromEdgeX) + float64(xStep)*float64(scaledTileSizeX)
            drawScreenY := float64(drawOffsetFromEdgeY) + float64(yStep)*float64(scaledTileSizeY)
            drawInfos := r.mapWindow.GetTextureIndexAt(mapPosX, mapPosY, tick)
            for _, drawInfo := range drawInfos {
                r.gridRenderer.DrawDefaultScaleTile(drawScreenX, drawScreenY, drawInfo.Atlas, drawInfo.Icon, drawInfo.Color)
            }
        }
    }
}
func (r *MapRenderer) DrawOnMapWithAtlas(mapPos geometry.Point, atlas TextureAtlas, icon int32, tintColor color.Color) {
    if !r.mapWindow.IsMapCellVisible(mapPos) {
        return
    }
    x, y := r.mapToScreen(mapPos)
    r.gridRenderer.DrawDefaultScaleTile(x, y, atlas, icon, tintColor)
}

func (r *MapRenderer) DrawOnMap(mapPos geometry.Point, icon int32, tintColor color.Color) {
    if !r.mapWindow.IsMapCellVisible(mapPos) {
        return
    }
    x, y := r.mapToScreen(mapPos)
    r.gridRenderer.DrawDefaultScaleTile(x, y, r.gridRenderer.defaultAtlas, icon, tintColor)
}
func (r *MapRenderer) DrawOnMapF(mapPos geometry.PointF, icon int32, tintColor color.Color) {
    if !r.mapWindow.IsMapCellVisible(mapPos.ToPoint()) { // could be too rough..
        return
    }
    x, y := r.mapFloatToScreen(mapPos)
    r.gridRenderer.DrawDefaultScaleTile(x, y, r.gridRenderer.defaultAtlas, icon, tintColor)
}
func (r *MapRenderer) DrawOnMapFWithAtlas(mapPos geometry.PointF, atlas TextureAtlas, icon int32, tintColor color.Color) {
    if !r.mapWindow.IsMapCellVisible(mapPos.ToPoint()) { // could be too rough..
        return
    }
    x, y := r.mapFloatToScreen(mapPos)
    r.gridRenderer.DrawDefaultScaleTile(x, y, atlas, icon, tintColor)
}

func (r *MapRenderer) mapToScreen(mapPos geometry.Point) (float64, float64) {
    scrollOffset := r.mapWindow.GetScrollOffset()

    gridSize := r.mapWindow.GetGridSize()
    tileScale := r.mapWindow.GetTileScale()

    scaledTileSize := gridSize.ToPointF().Mul(tileScale)

    drawOffsetFromEdgeX := (scrollOffset.X % int(scaledTileSize.X)) * -1
    drawOffsetFromEdgeY := (scrollOffset.Y % int(scaledTileSize.Y)) * -1

    offsetInTilesX := scrollOffset.X / int(scaledTileSize.X)
    offsetInTilesY := scrollOffset.Y / int(scaledTileSize.Y)

    x := float64(drawOffsetFromEdgeX) + float64(mapPos.X-offsetInTilesX)*float64(scaledTileSize.X)
    y := float64(drawOffsetFromEdgeY) + float64(mapPos.Y-offsetInTilesY)*float64(scaledTileSize.Y)
    return x, y
}

func (r *MapRenderer) mapFloatToScreen(mapPos geometry.PointF) (float64, float64) {
    scrollOffset := r.mapWindow.GetScrollOffset()

    gridSize := r.mapWindow.GetGridSize()
    tileScale := r.mapWindow.GetTileScale()

    scaledTileSize := gridSize.ToPointF().Mul(tileScale)

    drawOffsetFromEdgeX := (scrollOffset.X % int(scaledTileSize.X)) * -1
    drawOffsetFromEdgeY := (scrollOffset.Y % int(scaledTileSize.Y)) * -1

    offsetInTilesX := float64(scrollOffset.X / int(scaledTileSize.X))
    offsetInTilesY := float64(scrollOffset.Y / int(scaledTileSize.Y))

    x := float64(drawOffsetFromEdgeX) + float64(mapPos.X-offsetInTilesX)*float64(scaledTileSize.X)
    y := float64(drawOffsetFromEdgeY) + float64(mapPos.Y-offsetInTilesY)*float64(scaledTileSize.Y)
    return x, y
}
func (r *MapRenderer) DrawStringOnMapWithOffset(mapPos, screenOffset geometry.Point, text string, textColor color.Color) {
    if !r.mapWindow.IsMapCellVisible(mapPos) {
        return
    }
    x, y := r.mapToScreen(mapPos)
    r.gridRenderer.DrawTTFOnScreen(x+float64(screenOffset.X), y+float64(screenOffset.Y), text, textColor)
}
func (r *MapRenderer) CenterOn(pos geometry.Point) {
    r.mapWindow.CenterOn(pos)
}

func (r *MapRenderer) CenterOnFloat(pos geometry.PointF) {
    r.mapWindow.CenterOnFloat(pos)
}

func (r *MapRenderer) MapFloatToScreen(floatPos geometry.PointF) geometry.Point {
    return r.mapWindow.MapFloatToScreen(floatPos)
}

func (r *MapRenderer) GetMapCellAtScreenPos(pixels geometry.Point) geometry.Point {
    return r.mapWindow.GetMapCellAtScreenPos(pixels)
}

func (r *MapRenderer) ScrollBy(direction geometry.Point) {
    r.mapWindow.ScrollBy(direction)
}

func (r *MapRenderer) OnScreenSizeChanged(newWindowSize geometry.Point) {
    r.mapWindow.OnScreenSizeChanged(newWindowSize)
}

func (r *MapRenderer) GetScaledTileSize() geometry.PointF {
    return r.mapWindow.GetGridSize().ToPointF().Mul(r.mapWindow.GetTileScale())
}
func (r *MapRenderer) initMinimap() {
    mapSize := r.mapWindow.mapSize
    r.miniMapImage = ebiten.NewImage(mapSize.X, mapSize.Y)
    r.UpdateMiniMap()
}
func (r *MapRenderer) UpdateMiniMap() {
    mapSize := r.mapWindow.mapSize
    dataLength := mapSize.X * mapSize.Y * 4
    pixelData := make([]byte, dataLength)
    for x := 0; x < mapSize.X; x++ {
        for y := 0; y < mapSize.Y; y++ {
            var miniMapTileColor [4]byte
            mapPos := geometry.Point{X: x, Y: y}
            r.getMiniMapColor(mapPos, &miniMapTileColor)
            pixelData[(y*mapSize.X+x)*4] = miniMapTileColor[0]
            pixelData[(y*mapSize.X+x)*4+1] = miniMapTileColor[1]
            pixelData[(y*mapSize.X+x)*4+2] = miniMapTileColor[2]
            pixelData[(y*mapSize.X+x)*4+3] = miniMapTileColor[3]
        }
    }
    r.miniMapImage.WritePixels(pixelData)
}

func (r *MapRenderer) GetMiniMap() *ebiten.Image {
    return r.miniMapImage
}

func (r *MapRenderer) GetExactMapPositionFromScreenPos(point geometry.Point) geometry.PointF {
    return r.mapWindow.GetExactMapPositionFromScreenPos(point)
}

func (r *MapRenderer) ResetScrolling() {
    r.mapWindow.ResetScrolling()
}
