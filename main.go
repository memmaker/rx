package main

import (
	"RogueUI/console"
	"RogueUI/dungen"
	"RogueUI/foundation"
	"RogueUI/game"
	"RogueUI/util"
	"bufio"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"golang.org/x/term"
	"math/rand"
	"os"
	"path"
	"strings"
)

func main() {
	//testMapGen()
	//return

	//setKeypadToApplicationMode()   // set application mode
	//estKeyCodes()
	//return
	if !term.IsTerminal(0) {
		fmt.Println("This program must be run in a terminal.")
		return
	}
	width, _, err := term.GetSize(0)
	if err != nil {
		return
	}

	util.SetKeypadToNumericMode()
	bannerLines := util.ReadFileAsLines(path.Join("data", "banner.txt"))
	for _, line := range bannerLines {
		length := len(line)
		startX := (width - length) / 2
		if width == 0 {
			startX = 0
		}
		linePadded := util.LeftPadCount(line, startX)
		fmt.Println(linePadded)
	}

	var playerName string
	if len(os.Args) > 2 {
		if os.Args[1] == "-n" {
			playerName = os.Args[2]
		}
	} else {
		playerName = askForName()
	}
	config := foundation.NewConfigurationFromFile("config.rec")
	config.PlayerName = playerName
	gameUI := console.NewTextUI(config)
	game.NewGameState(gameUI, config)
	gameUI.StartGameLoop()

}

func testKeyCodes() {
	screen, _ := tcell.NewScreen()
	screen.Init()
	tty, _ := screen.Tty()
	//tty.Write([]byte{0x1B, 0x3D}) // set application mode
	tty.Write([]byte{0x1B, 0x3E}) // set numeric mode
	//util.SetKeypadToApplicationMode()       // unset application mode
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
	dunGen := dungen.NewRogueGenerator(random, 80, 23)
	for i := 0; i < 10; i++ {

		dungeon := dunGen.Generate()
		dungeon.Print()
		println()
	}

}
