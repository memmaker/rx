package console

import (
	"RogueUI/cview"
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/util"
	"cmp"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"math/rand"
	"slices"
	"strings"
	"unicode/utf8"
)

type HackingGame struct {
	*cview.Box
	style                     tcell.Style
	correctPassword           string
	wordsToInsert             []string
	guesses                   []string
	wordMatrixDim             geometry.Point
	wordIndices               []int
	random                    *rand.Rand
	currentlyHoveredWordIndex int
	randomSeed                int64
	allowedAttempts           int
	doneGuessing              bool
	closeFunc                 func(previousGuesses []string, result foundation.InteractionResult)
	borders                   cview.BorderDef
	audioPlayer               AudioCuePlayer
}

func NewHackingGame(correctPassword string, fakePasswords []string, close func(previousGuesses []string, result foundation.InteractionResult)) *HackingGame {
	box := cview.NewBox()
	//box.SetBorder(true)

	intSum := util.StringSum(correctPassword)
	rnd := rand.New(rand.NewSource(intSum))
	wordsToInsert := []string{correctPassword}
	wordsToInsert = append(wordsToInsert, fakePasswords...)
	rnd.Shuffle(len(wordsToInsert), func(i, j int) {
		wordsToInsert[i], wordsToInsert[j] = wordsToInsert[j], wordsToInsert[i]
	})
	for i := 0; i < len(wordsToInsert); i++ {
		wordsToInsert[i] = strings.ToUpper(wordsToInsert[i])
	}
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorForestGreen).Background(tcell.ColorBlack)
	h := &HackingGame{
		Box:                       box,
		random:                    rnd,
		randomSeed:                intSum,
		correctPassword:           strings.ToUpper(correctPassword),
		wordsToInsert:             wordsToInsert,
		allowedAttempts:           4,
		style:                     greenStyle,
		currentlyHoveredWordIndex: -1,
		closeFunc:                 close,
		borders: cview.BorderDef{
			Horizontal:  '─',
			Vertical:    '│',
			TopLeft:     '╭',
			TopRight:    '╮',
			BottomLeft:  '╰',
			BottomRight: '╯',
		},
	}

	box.SetBorder(false)
	box.SetDrawFunc(h.drawInside)
	box.SetBackgroundTransparent(false)
	box.SetInputCapture(h.handleInput)
	box.SetMouseCapture(h.handleMouse)
	return h
}
func (hg *HackingGame) SetAmberStyle() {
	hg.style = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
}
func (hg *HackingGame) layout(width, height int) {
	if width < 40 || height < 10 {
		return
	}
	wordLength := utf8.RuneCountInString(hg.correctPassword)

	rightBarWidth := 36
	headerHeight := 4
	widthLeft := width - rightBarWidth
	heightLeft := height - headerHeight
	screenChars := widthLeft * heightLeft
	maxIndex := screenChars - wordLength

	hg.wordIndices = make([]int, len(hg.wordsToInsert))
	isIndexAvailable := func(index int) bool {
		for _, i := range hg.wordIndices {
			if i == index || util.AbsInt(i-index) <= wordLength {
				return false
			}
		}
		return true
	}

	for i := 0; i < len(hg.wordsToInsert); i++ {
		// pick a random index
		index := hg.random.Intn(maxIndex)
		for !isIndexAvailable(index) {
			index = hg.random.Intn(maxIndex)
		}
		// insert the word
		hg.wordIndices[i] = index
	}

	slices.SortStableFunc(hg.wordIndices, func(i, j int) int {
		return cmp.Compare(i, j)
	})

	hg.wordMatrixDim = geometry.Point{X: widthLeft, Y: heightLeft}

}
func (hg *HackingGame) evaluateGuess(guess string) int {
	var correct int
	corrRunes := []rune(hg.correctPassword)
	asRunes := []rune(guess)
	for i, r := range asRunes {
		if i >= len(corrRunes) {
			break
		}
		if r == corrRunes[i] {
			correct++
		}
	}
	return correct
}

func printToScreen(screen tcell.Screen, x int, y int, text string, style tcell.Style) {
	for i, r := range []rune(text) {
		screen.SetContent(x+i, y, r, nil, style)
	}
}
func (hg *HackingGame) playCue(cue string) {
	if hg.audioPlayer != nil {
		hg.audioPlayer.PlayCue(cue)
	}
}
func (hg *HackingGame) drawInside(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	if width < 40 || height < 10 {
		return x, y, width, height
	}
	cview.DrawBox(screen, x, y, x+width-1, y+height-1, hg.style, hg.borders, ' ')

	x = x + 1
	y = y + 1
	width = width - 2
	height = height - 2
	hg.random.Seed(hg.randomSeed)
	if hg.wordMatrixDim.X != width-36 || hg.wordMatrixDim.Y != height-4 {
		hg.layout(width, height)
	}
	hg.random.Seed(hg.randomSeed)
	charCount := hg.wordMatrixDim.X * hg.wordMatrixDim.Y
	currentIndex := 0
	currentWordIndex := hg.wordIndices[currentIndex]
	currentWord := []rune(hg.wordsToInsert[currentIndex])
	guessesLeft := hg.guessesLeft()

	printToScreen(screen, x, y, strings.Repeat("*", width), hg.style)
	printToScreen(screen, x, y+3, strings.Repeat("*", width), hg.style)
	hg.printHeaderLineCentered(screen, x, y+1, "Vault-Tec Security System", width, hg.style)
	if hg.doneGuessing {
		if hg.wasLastGuessCorrect() {
			hg.printHeaderLineCentered(screen, x, y+2, ">> Access Granted <<", width, hg.style.Foreground(tcell.ColorGreen).Reverse(true))
		} else {
			hg.printHeaderLineCentered(screen, x, y+2, "<< Access Denied >>", width, hg.style.Foreground(tcell.ColorRed).Reverse(true))
		}
	} else if guessesLeft <= 0 {
		hg.printHeaderLineCentered(screen, x, y+2, "<< LOCKED >>", width, hg.style.Foreground(tcell.ColorRed))
	} else {
		guessesLeftStr := strings.TrimSpace(strings.Repeat("█ ", guessesLeft))
		hg.printHeaderLineCentered(screen, x, y+2, fmt.Sprintf("Attempts remaining: %s", guessesLeftStr), width, hg.style.Foreground(tcell.ColorYellow))
	}
	matrixX := x + 8
	matrixY := y + 4
	for i := 0; i < hg.wordMatrixDim.Y; i++ {
		var addressCounter uint16
		addressCounter = uint16(i) * uint16(hg.wordMatrixDim.X)
		asHex := fmt.Sprintf("0x%04X", addressCounter)
		printToScreen(screen, x, y+4+i, asHex, hg.style)
	}
	for i := 0; i < charCount; i++ {
		isInsideWord := i >= currentWordIndex && i < currentWordIndex+utf8.RuneCountInString(hg.correctPassword)
		var r rune
		style := hg.style
		if isInsideWord {
			charIndex := i - currentWordIndex
			r = currentWord[charIndex]
			//style = style.Bold(true)
			if currentIndex == hg.currentlyHoveredWordIndex {
				style = style.Reverse(true)
			}

			if charIndex == utf8.RuneCountInString(hg.correctPassword)-1 {
				currentIndex++
				if currentIndex < len(hg.wordsToInsert) {
					currentWordIndex = hg.wordIndices[currentIndex]
					currentWord = []rune(hg.wordsToInsert[currentIndex])
				} else {
					currentWordIndex = -1
				}
			}

		} else {
			// special chars
			special := []rune{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '+', '=', '{', '}', '[', ']', '|', '\\', ':', ';', '"', '\'', '<', '>', ',', '.', '?', '/', '`', '~'}
			r = special[hg.random.Intn(len(special))]
		}
		screen.SetContent(matrixX+i%hg.wordMatrixDim.X, matrixY+i/hg.wordMatrixDim.X, r, nil, style)
	}

	rightbarX := 8 + hg.wordMatrixDim.X + 2
	rightbarY := 4 + hg.wordMatrixDim.Y

	for i := 0; i < len(hg.guesses); i++ {
		reverseIndex := len(hg.guesses) - i - 1
		guessedWord := hg.guesses[reverseIndex]
		correct := hg.evaluateGuess(guessedWord)
		result := fmt.Sprintf("> %s: %d/%d", guessedWord, correct, utf8.RuneCountInString(hg.correctPassword))
		style := hg.style.Foreground(tcell.ColorLightGray)
		if correct == utf8.RuneCountInString(hg.correctPassword) {
			style = style.Foreground(tcell.ColorLimeGreen)
		} else if correct == 0 {
			style = style.Foreground(tcell.ColorRed)
		}
		printToScreen(screen, rightbarX, rightbarY-i, result, style)
	}

	return x, y, width, height
}

func (hg *HackingGame) guessesLeft() int {
	return hg.allowedAttempts - len(hg.guesses)
}

func (hg *HackingGame) printHeaderLineCentered(screen tcell.Screen, xStart int, yCoord int, text string, width int, style tcell.Style) {
	//strLen := utf8.RuneCountInString(text)
	strLen := runewidth.StringWidth(text)
	centerX := (width - strLen) / 2
	leftPad := strings.Repeat(" ", centerX-1)
	rightPad := strings.Repeat(" ", width-centerX-strLen)
	printToScreen(screen, xStart+1, yCoord, leftPad, hg.style)
	printToScreen(screen, xStart+centerX+strLen, yCoord, rightPad, hg.style)

	screen.SetContent(xStart, yCoord, '*', nil, hg.style)
	screen.SetContent(width, yCoord, '*', nil, hg.style)

	printToScreen(screen, xStart+centerX, yCoord, text, style)
}

func (hg *HackingGame) handleInput(event *tcell.EventKey) *tcell.EventKey {
	//uikey := toUIKey(event)
	if hg.doneGuessing || hg.guessesLeft() <= 0 {
		if hg.wasLastGuessCorrect() {
			hg.closeFunc(hg.guesses, foundation.Success)
		} else {
			hg.closeFunc(hg.guesses, foundation.Failure)
		}
		return nil
	}
	switch event.Key() {
	case tcell.KeyUp:
		if hg.currentlyHoveredWordIndex > 0 {
			hg.currentlyHoveredWordIndex--
		} else {
			hg.currentlyHoveredWordIndex = len(hg.wordsToInsert) - 1
		}
		hg.playCue("ui/hacking_updown")
	case tcell.KeyDown:
		if hg.currentlyHoveredWordIndex < len(hg.wordsToInsert)-1 {
			hg.currentlyHoveredWordIndex++
		} else {
			hg.currentlyHoveredWordIndex = 0
		}
		hg.playCue("ui/hacking_updown")
	case tcell.KeyEnter:
		hg.confirmSelection()
	case tcell.KeyEsc:
		hg.closeFunc(hg.guesses, foundation.Cancel)
	}
	return event
}
func (hg *HackingGame) confirmSelection() {
	if hg.currentlyHoveredWordIndex >= 0 && hg.currentlyHoveredWordIndex < len(hg.wordsToInsert) {
		guessedWord := hg.wordsToInsert[hg.currentlyHoveredWordIndex]
		hg.guesses = append(hg.guesses, guessedWord)
		if guessedWord == hg.correctPassword || hg.allowedAttempts == len(hg.guesses) {
			hg.doneGuessing = true
		}
		hg.playCue("ui/hacking_confirm")
	}
}
func (hg *HackingGame) SetAudioPlayer(player AudioCuePlayer) {
	hg.audioPlayer = player
}
func (hg *HackingGame) wasLastGuessCorrect() bool {
	lastGuess := hg.guesses[len(hg.guesses)-1]
	guessedCorrectly := lastGuess == hg.correctPassword
	return guessedCorrectly
}

func (hg *HackingGame) handleMouse(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
	if hg.doneGuessing || hg.guessesLeft() <= 0 {
		if action == cview.MouseLeftClick {
			if hg.wasLastGuessCorrect() {
				hg.closeFunc(hg.guesses, foundation.Success)
			} else {
				hg.closeFunc(hg.guesses, foundation.Failure)
			}
		}
		return action, nil
	}

	rawMouseX, rawMouseY := event.Position()
	mouseX := rawMouseX - 8 - 1
	mouseY := rawMouseY - 4 - 1
	wordLength := utf8.RuneCountInString(hg.correctPassword)
	if mouseX >= 0 && mouseY >= 0 && mouseX < hg.wordMatrixDim.X && mouseY < hg.wordMatrixDim.Y {
		indexInWordMatrix := mouseY*hg.wordMatrixDim.X + mouseX
		for i, index := range hg.wordIndices {
			if indexInWordMatrix >= index && indexInWordMatrix < index+wordLength {
				hg.currentlyHoveredWordIndex = i
				hg.playCue("ui/hacking_updown")
				break
			}
		}
	}
	if action == cview.MouseLeftClick {
		hg.confirmSelection()
	}
	return cview.MouseLeftClick, event
}

func (hg *HackingGame) SetAttemptsLeft(left int) {
	hg.allowedAttempts = left
}

func (hg *HackingGame) SetGuesses(guesses []string) {
	hg.guesses = guesses
}
