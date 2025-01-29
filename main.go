package main

import (
    "bufio"
    "contractor/dungen"
    "contractor/foundation"
    "contractor/game"
    "contractor/ui_console"
    "contractor/validation"
    "encoding/csv"
    "fmt"
    "github.com/gdamore/tcell/v2"
    "github.com/memmaker/go/fxtools"
    "github.com/memmaker/go/recfile"
    "math/rand"
    "os"
    "path/filepath"
    "strings"
)

var textModeLifecycles = make(map[string]ui_console.UILifeCycler)

func cleanUpWeaponSounds() {
    soundDir := "data_atom/audio/weapons"
    suffixes := []string{"_single", "_burst", "_reload", "_hit_surface", "_out_of_ammo"}
    csvFile := fxtools.MustOpen("sound_id_mapper.csv")
    reader := csv.NewReader(csvFile)
    records, err := reader.ReadAll()
    if err != nil {
        panic(err)
    }
    mapper := make(map[string]string)
    for row, record := range records {
        if row == 0 {
            continue
        }
        soundId := record[0]
        soundName := record[1]
        mapper[soundId] = soundName

        for _, suffix := range suffixes {
            soundFileName := fmt.Sprintf("%s%s", soundId, suffix)
            newName := fmt.Sprintf("%s%s", soundName, suffix)
            soundFile := filepath.Join(soundDir, soundFileName)
            if fxtools.DirExists(soundFile) {
                os.Rename(soundFile, filepath.Join(soundDir, newName))
            }
        }
    }

    weaponFile := "data_atom/definitions/weapons.rec"
    weaponRecords, _ := recfile.ReadAndClose(fxtools.MustOpen(weaponFile))
    for wpnIdx, weaponRecord := range weaponRecords {
        field, hasField := weaponRecord.FindFieldIgnoreCase("weapon_sound_id")
        if !hasField {
            continue
        }
        soundName, hasSound := mapper[field.Value]
        if !hasSound {
            continue
        }
        weaponRecord.WithKeyValue(field.Name, soundName)
        weaponRecords[wpnIdx] = weaponRecord
    }

    recfile.WriteAndClose(fxtools.MustCreate(weaponFile+".new"), weaponRecords)
}

func main() {

    devStart := false

    config := foundation.NewConfigurationFromFile("config.rec")
    mode := "graphics"

    if len(os.Args) > 1 {
        if os.Args[1] == "dev" {
            devStart = true
        } else if os.Args[1] == "val_dialogue" {
            gameState := game.NewGameState(config)
            validation.ValidateDialogue(config.DataRootDir, gameState)
            return
        } else if os.Args[1] == "graph_dialogue" {
            hideBackLinksToNode := ""
            filename := ""
            if len(os.Args) > 3 {
                hideBackLinksToNode = os.Args[2]
                filename = os.Args[3]
            } else {
                filename = os.Args[2]
            }
            gameState := game.NewGameState(config)
            validation.GraphDialogue(config.DataRootDir, gameState, filename, hideBackLinksToNode)
            return
        } else if os.Args[1] == "val_ammo" {
            validation.ValidateWeaponAndAmmoPairings(config.DataRootDir)
            return
        } else {
            mode = os.Args[1]
        }
    }

    fxtools.SetKeypadToNumericMode()

    if config.PlayerName == "" {
        config.PlayerName = askForName()
        config.WriteToFile("config.rec")
    }
    // create an ui independent game state
    gameState := game.NewGameState(config)

    // create the game UI and link it to the game state
    var gameUI foundation.GameUI

    if false && strings.ToLower(mode) == "graphics" {
        /*
           		tileRenderer := ui_graphics.NewTileRenderer(config.TileScale)
                   tileRenderer.SetColorFromName(gameState.Palette().Get)
                   atlasPath := filepath.Join(config.DataRootDir, "cp437.png")
                   tileset := ui_graphics.NewTextureAtlas(atlasPath, config.TileWidth, config.TileHeight)
                   tileRenderer.SetDefaultAtlas(tileset)
                   tileRenderer.SetDefaultColors(color.RGBA{R: 200, G: 20, B: 20, A: 255}, color.Black)
                   tileRenderer.SetWhiteTile(config.WhiteTileIndex)

                   gameUI = ui_graphics.NewUIController(tileRenderer, gameState, config)

        */
    } else {
        var textModeLifecycle ui_console.UILifeCycler
        if lifeCyle, ok := textModeLifecycles[mode]; ok {
            textModeLifecycle = lifeCyle
            devStart = true
        }
        if textModeLifecycle == nil {
            textModeLifecycle = chooseGraphicsMode()
        }

        gameUI = ui_console.NewTextUI(gameState, textModeLifecycle, config)
    }

    if devStart {
        gameUI.StartGameLoop()
    } else {
        gameUI.StartWithIntro()
    }
}

func chooseGraphicsMode() ui_console.UILifeCycler {
    if len(textModeLifecycles) == 1 {
        for _, mode := range textModeLifecycles {
            return mode
        }
    }
    if len(textModeLifecycles) == 0 {
        panic("No graphics modes available")
    }
    fallbackModes := []string{"text_gl", "terminal"}

    for _, mode := range fallbackModes {
        if lifeCyle, ok := textModeLifecycles[mode]; ok {
            return lifeCyle
        }
    }
    panic("No fallback graphics mode available")
    return nil
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
