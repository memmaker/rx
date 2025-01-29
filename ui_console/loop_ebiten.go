//go:build ebiten

package ui_console

import (
    "contractor/foundation"
    "contractor/tcell_ebiten"
    "github.com/gdamore/tcell/v2"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/memmaker/go/cview"
    "github.com/memmaker/go/fxtools"
    "golang.org/x/image/font"
    "golang.org/x/image/font/opentype"
    "path/filepath"
    "time"
)

func NewEbitenUI() UILifeCycler {
    return EbitenUI{}
}

type EbitenUI struct{}

func (u EbitenUI) StartGameLoop(config *foundation.Configuration, application *cview.Application, onScreenReady func()) {
    ebiten.SetWindowSize(1280, 800)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    ebiten.SetWindowTitle("CONTRACTOR")
    ebiten.SetWindowDecorated(true)
    ebiten.SetWindowFloating(false)

    mainFont, closeMainFont := mustLoadFontByName(config.MainFontName)
    //mainFont, closeMainFont := mustLoadFontByName("PerfectDOSVGA437Unicode")
    fallbackFont, closeFallbackFont := mustLoadFontByName(config.FallbackFontName)
    defer closeMainFont()
    defer closeFallbackFont()

    gs := tcell_ebiten.NewGameScreen(text.NewGoXFace(mainFont))
    gs.Clear()
    gs.SetFallbackFont(text.NewGoXFace(fallbackFont))
    gs.ForcedFallbacks([]rune(config.ForcedFallbackRunes))
    gs.Init()
    defer gs.Fini()

    application.SetScreen(gs)

    go func() {
        screen := application.GetScreen()
        w, h := screen.Size()
        for w == 0 || h == 0 {
            time.Sleep(100 * time.Millisecond)
            w, h = screen.Size()
        }
        application.QueueUpdate(func() {
            resize := tcell.NewEventResize(w, h)
            application.QueueEvent(resize)

            if onScreenReady != nil {
                onScreenReady()
            }
        })
        if err := application.Run(); err != nil {
            panic(err)
        }
    }()

    ebiten.RunGameWithOptions(gs, &ebiten.RunGameOptions{
        GraphicsLibrary: ebiten.GraphicsLibraryOpenGL,
    })
}

func (u EbitenUI) QuitGame(application *cview.Application) {
    gs := application.GetScreen().(tcell_ebiten.GameScreen)
    gs.Close()
}

func mustLoadFontByName(fontName string) (fontFace font.Face, close func() error) {
    filename := filepath.Join("data_atom", "glfonts", fontName+".ttf")
    file := fxtools.MustOpen(filename)
    //defer file.Close()
    tt, err := opentype.ParseReaderAt(file)
    if err != nil {
        println(err.Error())
        return nil, nil
    }
    size := 24.0
    deviceDPIScale := ebiten.Monitor().DeviceScaleFactor()
    dpi := 72 * deviceDPIScale
    fontFace, faceErr := opentype.NewFace(tt, &opentype.FaceOptions{
        Size:    size,
        DPI:     dpi,
        Hinting: font.HintingVertical,
    })
    if faceErr != nil {
        println(faceErr.Error())
    }
    return fontFace, file.Close
}
