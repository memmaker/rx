package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/text"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/memmaker/go/geometry"
    "golang.org/x/image/font"
    "image"
    "image/color"
    "regexp"
    "strconv"
    "strings"
)

// what do we need?

// draw tiles from an atlas
// at any scale
// at any screen position
// from any atlas
// with any tint color

type TileRenderer struct {
    op          *ebiten.DrawImageOptions
    deviceScale float64
    tileScale   float64
    fontScale   float64
    font        BitmapFont

    defaultAtlas        TextureAtlas
    currentRenderTarget *ebiten.Image

    whiteTile int32
    ttfFont   font.Face

    colorFromName      func(name string) color.RGBA
    defaultBorderColor color.Color
    defaultFillColor   color.Color
    globalScaleColor   [4]float32
}

func (g *TileRenderer) SetRenderTarget(screen *ebiten.Image) {
    g.currentRenderTarget = screen
}
func (g *TileRenderer) IsInitialized() bool {
    return g.currentRenderTarget != nil
}
func (g *TileRenderer) SetColorFromName(colorFromName func(name string) color.RGBA) {
    g.colorFromName = colorFromName
}
func (g *TileRenderer) SetTTF(font font.Face) {
    g.ttfFont = font
}
func (g *TileRenderer) SetDefaultColors(borderColor, fillColor color.Color) {
    g.defaultBorderColor = borderColor
    g.defaultFillColor = fillColor
}
func (g *TileRenderer) MeasureString(textToMeasure string) (float64, float64) {
    _, advance := font.BoundString(g.ttfFont, textToMeasure)
    height := g.ttfFont.Metrics().Ascent + g.ttfFont.Metrics().Descent
    //return float64((bounds.Max.X - bounds.Min.X).Round()) / g.deviceScale, float64((bounds.Max.Y - bounds.Min.Y).Round()) / g.deviceScale
    return float64((advance).Ceil()) / g.deviceScale, float64((height).Round()) / g.deviceScale
}
func (g *TileRenderer) SetDefaultAtlas(atlas TextureAtlas) {
    g.defaultAtlas = atlas
}

func NewTileRenderer(tileScale float64) *TileRenderer {
    return &TileRenderer{
        op:          &ebiten.DrawImageOptions{},
        tileScale:   tileScale,
        deviceScale: 1.0,
        fontScale:   1,
        globalScaleColor: [4]float32{
            1, 1, 1, 1,
        },
    }
}
func (g *TileRenderer) SetWhiteTile(index int32) {
    g.whiteTile = index
}
func (g *TileRenderer) SetTileScale(scale float64) {
    g.tileScale = scale
}
func (g *TileRenderer) SetDeviceScale(scale float64) {
    g.deviceScale = scale
}
func (g *TileRenderer) SetFontScale(scale float64) {
    g.fontScale = scale
}

func (g *TileRenderer) SetFont(font BitmapFont) {
    g.font = font
}
func (g *TileRenderer) DrawDefaultBorder(topLeftScreen geometry.Point, size geometry.Point) {

    //scale := g.deviceScale

    drawX := float64(topLeftScreen.X) * g.deviceScale
    drawY := float64(topLeftScreen.Y) * g.deviceScale

    targetWidth := float64(size.X) * g.deviceScale
    targetHeight := float64(size.Y) * g.deviceScale

    tileSize := g.defaultAtlas.GetTileSize()

    scaleX := targetWidth / float64(tileSize.X)
    scaleY := targetHeight / float64(tileSize.Y)

    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(g.defaultBorderColor)
    g.op.GeoM.Reset()
    borderSize := 1 * g.deviceScale
    widthWithBorder := targetWidth + (borderSize * 2)
    heightWithBorder := targetHeight + (borderSize * 2)
    scaleWithBorderX := widthWithBorder / float64(tileSize.X)
    scaleWithBorderY := heightWithBorder / float64(tileSize.Y)
    g.op.GeoM.Scale(scaleWithBorderX, scaleWithBorderY)
    g.op.GeoM.Translate(float64(drawX)-borderSize, float64(drawY)-borderSize)
    //g.op.GeoM.Translate(float64(drawX), float64(drawY))
    g.currentRenderTarget.DrawImage(ExtractSubImageFromAtlas(g.whiteTile, g.defaultAtlas), g.op)

    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(g.defaultFillColor)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(scaleX, scaleY)
    g.op.GeoM.Translate(float64(drawX), float64(drawY))
    g.currentRenderTarget.DrawImage(ExtractSubImageFromAtlas(g.whiteTile, g.defaultAtlas), g.op)
}
func (g *TileRenderer) DrawImageOnScreen(screenX int, screenY int, size geometry.Point, image *ebiten.Image) {
    drawX := float64(screenX) * g.deviceScale
    drawY := float64(screenY) * g.deviceScale

    targetWidth := float64(size.X)
    targetHeight := float64(size.Y)

    imgSizeX := float64(image.Bounds().Dx()) / g.deviceScale
    imgSizeY := float64(image.Bounds().Dy()) / g.deviceScale

    scaleX := targetWidth / float64(imgSizeX)
    scaleY := targetHeight / float64(imgSizeY)

    g.op.ColorScale.Reset()

    g.op.GeoM.Reset()
    g.op.GeoM.Scale(scaleX, scaleY)
    g.op.GeoM.Translate(float64(drawX), float64(drawY))
    g.currentRenderTarget.DrawImage(image, g.op)
}

func (g *TileRenderer) DrawColoredBorder(topLeftScreen geometry.Point, size geometry.Point, fillColor, borderColor color.Color) {

    //scale := g.deviceScale

    drawX := float64(topLeftScreen.X) * g.deviceScale
    drawY := float64(topLeftScreen.Y) * g.deviceScale

    targetWidth := float64(size.X) * g.deviceScale
    targetHeight := float64(size.Y) * g.deviceScale

    tileSize := g.defaultAtlas.GetTileSize()

    scaleX := targetWidth / float64(tileSize.X)
    scaleY := targetHeight / float64(tileSize.Y)

    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(borderColor)
    borderSize := 1 * g.deviceScale
    widthWithBorder := targetWidth + (borderSize * 2)
    heightWithBorder := targetHeight + (borderSize * 2)
    scaleWithBorderX := widthWithBorder / float64(tileSize.X)
    scaleWithBorderY := heightWithBorder / float64(tileSize.Y)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(scaleWithBorderX, scaleWithBorderY)
    g.op.GeoM.Translate(float64(drawX)-borderSize, float64(drawY)-borderSize)
    //g.op.GeoM.Translate(float64(drawX), float64(drawY))
    g.currentRenderTarget.DrawImage(ExtractSubImageFromAtlas(g.whiteTile, g.defaultAtlas), g.op)

    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(fillColor)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(scaleX, scaleY)
    g.op.GeoM.Translate(float64(drawX), float64(drawY))
    g.currentRenderTarget.DrawImage(ExtractSubImageFromAtlas(g.whiteTile, g.defaultAtlas), g.op)
}

func (g *TileRenderer) DrawColoredRect(topLeftScreen geometry.Point, size geometry.Point, fillColor color.Color) {
    drawX := float64(topLeftScreen.X) * g.deviceScale
    drawY := float64(topLeftScreen.Y) * g.deviceScale

    targetWidth := float64(size.X) * g.deviceScale
    targetHeight := float64(size.Y) * g.deviceScale

    tileSize := g.defaultAtlas.GetTileSize()

    scaleX := targetWidth / float64(tileSize.X)
    scaleY := targetHeight / float64(tileSize.Y)

    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(fillColor)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(scaleX, scaleY)
    g.op.GeoM.Translate(float64(drawX), float64(drawY))
    g.currentRenderTarget.DrawImage(ExtractSubImageFromAtlas(g.whiteTile, g.defaultAtlas), g.op)
}

func (g *TileRenderer) DrawStringOnGrid(gridX int, gridY int, text string, color color.Color) {
    tileSize := g.font.atlas.GetTileSize()
    drawX := float64(gridX) * float64(tileSize.X) * g.deviceScale * g.fontScale
    drawY := float64(gridY) * float64(tileSize.Y) * g.deviceScale * g.fontScale
    g.DrawString(int(drawX), int(drawY), text, color)
}

func (g *TileRenderer) DrawCharOnGrid(gridX int, gridY int, icon rune, color color.Color) {
    tileSize := g.font.atlas.GetTileSize()
    drawX := float64(gridX) * float64(tileSize.X) * g.deviceScale * g.fontScale
    drawY := float64(gridY) * float64(tileSize.Y) * g.deviceScale * g.fontScale
    g.DrawDefaultScaleCharOnScreen(drawX, drawY, icon, color)
}

func (g *TileRenderer) DrawOnGrid(gridX int, gridY int, icon int32) {
    tileSize := g.defaultAtlas.GetTileSize()
    drawX := float64(gridX) * float64(tileSize.X) * g.deviceScale * g.tileScale
    drawY := float64(gridY) * float64(tileSize.Y) * g.deviceScale * g.tileScale
    g.DrawDefaultScaleTile(drawX, drawY, g.defaultAtlas, icon, color.White)
}

func (g *TileRenderer) DrawOnFontGrid(gridX int, gridY int, icon int32) {
    tileSize := g.font.atlas.GetTileSize()
    drawX := float64(gridX) * float64(tileSize.X) * g.deviceScale * g.fontScale
    drawY := float64(gridY) * float64(tileSize.Y) * g.deviceScale * g.fontScale
    g.DrawTileWithDefaultOrientation(drawX, drawY, g.defaultAtlas, icon, geometry.PointF{X: g.fontScale, Y: g.fontScale}, color.White)
}

func (g *TileRenderer) DrawOnScreen(screenX int, screenY int, icon int32) {
    g.DrawDefaultScaleTile(float64(screenX), float64(screenY), g.defaultAtlas, icon, color.White)
}

func (g *TileRenderer) DrawOnScreenWithScale(screenX int, screenY int, icon int32, scale float64) {
    g.DrawTileWithDefaultOrientation(float64(screenX), float64(screenY), g.defaultAtlas, icon, geometry.PointF{X: scale, Y: scale}, color.White)
}

func (g *TileRenderer) DrawDefaultScaleTile(screenX float64, screenY float64, atlas TextureAtlas, index int32, tintColor color.Color) {
    g.DrawTileWithDefaultOrientation(screenX, screenY, atlas, index, geometry.PointF{X: g.tileScale, Y: g.tileScale}, tintColor)
}
func (g *TileRenderer) DrawDefaultScaleTileWithFlip(screenX float64, screenY float64, atlas TextureAtlas, index int32, tintColor color.Color, flipX bool) {
    g.DrawTile(screenX, screenY, atlas, index, geometry.PointF{X: g.tileScale, Y: g.tileScale}, tintColor, flipX)
}

func (g *TileRenderer) DrawScaledTile(screenX float64, screenY float64, atlas TextureAtlas, index int32, scale geometry.PointF, tintColor color.Color) {
    g.DrawTileWithDefaultOrientation(screenX, screenY, atlas, index, geometry.PointF{X: g.tileScale * scale.X, Y: g.tileScale * scale.Y}, tintColor)
}

func (g *TileRenderer) DrawScaledTileWithFlip(screenX float64, screenY float64, atlas TextureAtlas, index int32, scale geometry.PointF, tintColor color.Color, flipX bool) {
    g.DrawTile(screenX, screenY, atlas, index, geometry.PointF{X: g.tileScale * scale.X, Y: g.tileScale * scale.Y}, tintColor, flipX)
}
func (g *TileRenderer) DrawTileWithDefaultOrientation(screenX float64, screenY float64, atlas TextureAtlas, index int32, scale geometry.PointF, tintColor color.Color) {
    g.DrawTile(screenX, screenY, atlas, index, scale, tintColor, false)
}

func (g *TileRenderer) SetGlobalScaleColor(color [4]float32) {
    g.globalScaleColor = color
}
func (g *TileRenderer) DrawTile(screenX float64, screenY float64, atlas TextureAtlas, index int32, scale geometry.PointF, tintColor color.Color, flipX bool) {
    g.op.ColorScale.Reset()
    //g.op.ColorScale.SetR()

    // CORRECT: g.op.ColorScale.ScaleWithColor(tintColor)

    // WEIRD TEST:
    rVal, gVal, bVal, aVal := tintColor.RGBA()
    rAsFloat := float32(float64(rVal) / 65535.0)
    gAsFloat := float32(float64(gVal) / 65535.0)
    bAsFloat := float32(float64(bVal) / 65535.0)
    aAsFloat := float32(float64(aVal) / 65535.0)
    scaleColor := g.globalScaleColor
    scaleRVal, scaleGVal, scaleBVal, scaleAVal := scaleColor[0], scaleColor[1], scaleColor[2], scaleColor[3]
    finalR := rAsFloat * scaleRVal
    finalG := gAsFloat * scaleGVal
    finalB := bAsFloat * scaleBVal
    finalA := aAsFloat * scaleAVal

    g.op.ColorScale.Scale(finalR, finalG, finalB, finalA)

    g.op.GeoM.Reset()
    tileScale := scale.Mul(g.deviceScale)
    tx := screenX * g.deviceScale
    if flipX {
        g.op.GeoM.Scale(-tileScale.X, tileScale.Y)
        tx += float64(atlas.tileSizeX) * tileScale.X
    } else {
        g.op.GeoM.Scale(tileScale.X, tileScale.Y)
    }
    g.op.GeoM.Translate(tx, screenY*g.deviceScale)
    g.currentRenderTarget.DrawImage(ExtractSubImageFromAtlas(index, atlas), g.op)
}
func (g *TileRenderer) DrawDefaultScaleCharOnScreen(screenX, screenY float64, char rune, textColor color.Color) {
    g.DrawCharOnScreen(int(screenX), int(screenY), char, g.fontScale, textColor)
}
func (g *TileRenderer) DrawTTFOnScreen(screenX, screenY float64, textToDraw string, textColor color.Color) {
    screenX = screenX * g.deviceScale
    screenY = screenY * g.deviceScale
    text.Draw(g.currentRenderTarget, textToDraw, g.ttfFont, int(screenX), int(screenY), textColor)
}

func (g *TileRenderer) DrawTTFOnScreenWithColorCodes(screenX, screenY float64, textToDraw []ColoredTextPart) {
    drawOptions := &ebiten.DrawImageOptions{}
    screenXF := screenX * g.deviceScale
    screenYF := screenY * g.deviceScale
    for _, textPart := range textToDraw {
        offset := textPart.XOffset * (g.deviceScale)
        drawOptions.GeoM.Reset()
        drawOptions.GeoM.Translate(screenXF+offset, screenYF)
        drawOptions.ColorScale.Reset()
        drawOptions.ColorScale.ScaleWithColor(textPart.Color)
        text.DrawWithOptions(g.currentRenderTarget, textPart.Text, g.ttfFont, drawOptions)
        //text.Draw(g.currentRenderTarget, textPart.Text, g.ttfFont, screenX+offset, screenY, textPart.Color)
    }

}
func (g *TileRenderer) DrawCharOnScreen(screenX, screenY int, char rune, scale float64, textColor color.Color) {
    if !g.font.IsLoaded() {
        println("font not loaded")
        return
    }
    textureIndex, ok := g.font.fontMap[char]
    if !ok {
        return
    }
    fontScale := g.deviceScale * scale
    g.op.ColorScale.Reset()
    //g.op.ColorScale.ScaleWithColor(textColor)
    _, _, _, alphaByte := textColor.RGBA()
    alphaFloat := float32(float64(alphaByte) / 65535.0)
    g.op.ColorScale.ScaleAlpha(alphaFloat)
    g.op.ColorScale.ScaleWithColor(textColor)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(fontScale, fontScale)
    g.op.GeoM.Translate(float64(screenX), float64(screenY))
    g.currentRenderTarget.DrawImage(ExtractSubImageFromAtlas(int32(textureIndex), g.font.atlas), g.op)
}

func (g *TileRenderer) DrawString(screenX int, screenY int, text string, color color.Color) {
    tileSizeX := float64(g.font.atlas.tileSizeX) * g.fontScale * g.deviceScale

    for i, char := range text {
        g.DrawDefaultScaleCharOnScreen(float64(screenX)+float64(i)*tileSizeX, float64(screenY), char, color)
    }
}

func (g *TileRenderer) DrawMultiString(screenX int, screenY int, text []string, color color.Color) {
    tileSizeY := float64(g.font.atlas.tileSizeY) * g.fontScale * g.deviceScale

    for i, line := range text {
        g.DrawString(screenX, screenY+int(float64(i)*tileSizeY), line, color)
    }
}

func (g *TileRenderer) GetFontScale() float64 {
    return g.fontScale
}

func (g *TileRenderer) GetFontGridSize() geometry.Point {
    return g.font.atlas.GetTileSize()
}

func (g *TileRenderer) ParseColorCodedText(draw string) []ColoredTextPart {
    advance, _ := g.ttfFont.GlyphAdvance(' ')
    advanceWidth := float64(advance.Round()) / g.deviceScale
    var textParts []ColoredTextPart
    // color codes look like this [:red] [:blue] [:green] [:yellow] [:white] [:24,54,222]
    regexPattern := `\[:([\w,]+)\]`
    rp, _ := regexp.Compile(regexPattern)
    // split the string into parts
    stringParts := rp.Split(draw, -1)
    if stringParts[0] == "" {
        stringParts = stringParts[1:]
    }
    // find the color codes
    colorCodes := rp.FindAllStringSubmatch(draw, -1)
    if len(colorCodes) < len(stringParts) {
        // prepend [:white]
        colorCodes = append([][]string{{":white", "white"}}, colorCodes...)
    }
    xOffset := 0.0
    for i, textPart := range stringParts {
        textPart = strings.TrimSpace(textPart)
        if textPart == "" {
            continue
        }
        tWidth, tHeight := g.MeasureString(textPart)
        matchingColor := g.toColor(colorCodes, i)
        textParts = append(textParts, ColoredTextPart{
            Text:    textPart,
            Color:   matchingColor,
            XOffset: xOffset,
            Height:  tHeight,
        })
        xOffset += tWidth + advanceWidth
    }
    return textParts
}

var White = color.RGBA{R: 255, G: 255, B: 255, A: 255}

func (g *TileRenderer) toColor(matches [][]string, index int) color.RGBA {
    if len(matches) == 0 {
        return White
    }
    colorName := matches[index][1]
    if strings.ContainsRune(colorName, ',') {
        colorValues := strings.Split(colorName, ",")
        rVal, _ := strconv.Atoi(colorValues[0])
        gVal, _ := strconv.Atoi(colorValues[1])
        bVal, _ := strconv.Atoi(colorValues[2])
        return color.RGBA{R: uint8(rVal), G: uint8(gVal), B: uint8(bVal), A: 255}
    }
    return g.colorFromName(colorName)
}

func (g *TileRenderer) DebugLine(start geometry.Point, end geometry.Point) {
    scale := float32(g.deviceScale)
    vector.StrokeLine(g.currentRenderTarget, float32(start.X)*scale, float32(start.Y)*scale, float32(end.X)*scale, float32(end.Y)*scale, 10, color.White, false)
}

func ExtractSubImageFromAtlas(textureIndex int32, atlas TextureAtlas) *ebiten.Image {
    tileSizeX := atlas.tileSizeX
    tileSizeY := atlas.tileSizeY
    textureData := atlas.imageData

    atlasItemCountX := int32(textureData.Bounds().Size().X / tileSizeX)
    textureRectTopLeft := image.Point{
        X: int((textureIndex % atlasItemCountX) * int32(tileSizeX)),
        Y: int((textureIndex / atlasItemCountX) * int32(tileSizeY)),
    }
    textureRect := image.Rectangle{
        Min: textureRectTopLeft,
        Max: image.Point{
            X: textureRectTopLeft.X + tileSizeX,
            Y: textureRectTopLeft.Y + tileSizeY,
        },
    }

    tileImage := textureData.SubImage(textureRect).(*ebiten.Image)
    return tileImage
}
