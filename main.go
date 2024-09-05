package main

import (
    "RogueUI/console"
    "RogueUI/dungen"
    "RogueUI/foundation"
    "RogueUI/game"
    "bufio"
    "fmt"
    "github.com/gdamore/tcell/v2"
    "github.com/memmaker/go/fxtools"
    "math/rand"
    "os"
    "strings"
)

func main() {
    fxtools.SetKeypadToNumericMode()

    config := foundation.NewConfigurationFromFile("config.rec")

    showIntro := false
    if config.PlayerName == "" {
        config.PlayerName = askForName()
        showIntro = true
        //config.WriteToFile("config.rec")
    }
    gameUI := console.NewTextUI(config)
    game.NewGameState(gameUI, config)

    if showIntro {
        gameUI.StartWithIntro()
    } else {
        gameUI.StartGameLoop()
    }
}
func showBanner(filename string, width int) {
    bannerLines := fxtools.ReadFileAsLines(filename)
    for _, line := range bannerLines {
        length := len(line)
        startX := (width - length) / 2
        if width == 0 {
            startX = 0
        }
        linePadded := fxtools.LeftPadCount(line, startX)
        fmt.Println(linePadded)
    }
}

func testKeyCodes() {
    screen, _ := tcell.NewScreen()
    screen.Init()
    tty, _ := screen.Tty()
    //tty.Write([]byte{0x1B, 0x3D}) // set application mode
    tty.Write([]byte{0x1B, 0x3E}) // set numeric mode
    //fxtools.SetKeypadToApplicationMode()       // unset application mode
    quit := false
    for !quit {
        event := screen.PollEvent()
        switch typedEvent := event.(type) {
        case *tcell.EventKey:
            keyMessage := fmt.Sprintf("KeyID: %d, KeyName: %s, Rune: %c, Mods: %s", typedEvent.Key(), tcell.KeyNames[typedEvent.Key()], typedEvent.Rune(), modAsString(typedEvent.Modifiers()))
            screenPrint(screen, keyMessage)
            if typedEvent.Key() == tcell.KeyCtrlC {
                quit = true
            }
        }
    }
}

func modAsString(modifiers tcell.ModMask) string {
    var mods []string
    if modifiers&tcell.ModCtrl != 0 {
        mods = append(mods, "Ctrl")
    }
    if modifiers&tcell.ModAlt != 0 {
        mods = append(mods, "Alt")
    }
    if modifiers&tcell.ModShift != 0 {
        mods = append(mods, "Shift")
    }
    return strings.Join(mods, "|")
}

func screenPrint(screen tcell.Screen, text string) {
    screen.Clear()
    startY := 0
    startX := 0
    style := tcell.StyleDefault
    for i, r := range []rune(text) {
        screen.SetContent(startX+i, startY, r, nil, style)
    }
    screen.Show()
}

func askForName() string {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Who are you? ")
    userInput, _ := reader.ReadString('\n')
    return strings.TrimSpace(userInput)
}

func testMapGen() {
    random := rand.New(rand.NewSource(42))
    dunGen := dungen.NewVaultGenerator(random, 80, 23)
    for i := 0; i < 10; i++ {
        dungeon := dunGen.Generate()
        dungeon.Print()
        println()
    }
}
