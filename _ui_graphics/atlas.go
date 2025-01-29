package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
    "image"
    "image/color"
    "log"
    "os"
)

type CellDrawInfo struct {
    Icon  int32
    Color color.Color
    Atlas TextureAtlas
}
type TextureAtlas struct {
    imageData *ebiten.Image
    tileSizeX int
    tileSizeY int
}

func (a TextureAtlas) GetTileSize() geometry.Point {
    return geometry.Point{X: a.tileSizeX, Y: a.tileSizeY}
}

func NewTextureAtlas(imageFilename string, tileSizeX, tileSizeY int) TextureAtlas {
    return TextureAtlas{
        imageData: ebiten.NewImageFromImage(mustLoadImage(imageFilename)),
        tileSizeX: tileSizeX,
        tileSizeY: tileSizeY,
    }
}

func mustLoadImage(filename string) image.Image {
    openFile, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    }
    defer openFile.Close()
    img, _, err := image.Decode(openFile)
    if err != nil {
        log.Fatal(err)
    }
    return img
}
