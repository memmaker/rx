package console

import (
	"RogueUI/foundation"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/geometry"
)

type KeyPad struct {
	*cview.Box
	width           int
	height          int
	currentSequence []rune
	correctSequence []rune
	onCompletion    func(success bool)
	audioPlayer     foundation.AudioCuePlayer
}

func NewKeyPad(screenSize geometry.Point) *KeyPad {
	box := cview.NewBox()
	k := &KeyPad{
		Box:    box,
		width:  (3 * 3) + 4,
		height: (4 * 3) + 5 + 3,
	}
	x := (screenSize.X - k.width) / 2
	y := (screenSize.Y - k.height) / 2
	box.SetRect(x, y, k.width, k.height)
	box.SetBorder(false)
	box.SetInputCapture(k.handleKeys)
	return k
}

func (k *KeyPad) Draw(screen tcell.Screen) {
	k.Box.Draw(screen)
	x, y, width, height := k.GetInnerRect()
	k.drawKeyPad(screen, x, y, width, height)
}

func (k *KeyPad) drawKeyPad(screen tcell.Screen, x int, y int, width int, height int) {
	displayStyle := tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite).Bold(true)
	for i := 0; i < width; i++ {
		screen.SetContent(x+i, y, ' ', nil, displayStyle)
		screen.SetContent(x+i, y+1, ' ', nil, displayStyle)
		screen.SetContent(x+i, y+2, ' ', nil, displayStyle)
	}
	screen.SetCursorStyle(tcell.CursorStyleBlinkingUnderline)
	screen.ShowCursor(x+width-2, y+1)
	for i, _ := range k.currentSequence { // right aligned
		drawX := x + width - 2 - i
		key := k.currentSequence[len(k.currentSequence)-1-i]
		screen.SetContent(drawX, y+1, key, nil, displayStyle)
	}

	k.drawKeys(screen, x, y+3, width, height-3)
}

func (k *KeyPad) drawKeys(screen tcell.Screen, x int, y int, width int, height int) {
	labelRunes := []rune{'7', '8', '9', '4', '5', '6', '1', '2', '3', '*', '0', '#'}
	popRune := func() rune {
		r := labelRunes[0]
		labelRunes = labelRunes[1:]
		return r
	}
	isLabelPos := func(posX, posY int) bool {
		return (posX == x+2 || posX == x+6 || posX == x+10) && (posY == y+2 || posY == y+6 || posY == y+10 || posY == y+14)
	}
	chooseBorderChar := func(posX, posY int) rune {
		hasHorizontalBorder := posY == y || posY == y+4 || posY == y+8 || posY == y+12 || posY == y+16
		hasVerticalBorder := posX == x || posX == x+4 || posX == x+8 || posX == x+12
		if hasHorizontalBorder && hasVerticalBorder {
			if posX == x && posY == y {
				return '┌'
			} else if posX == x+width-1 && posY == y {
				return '┐'
			} else if posX == x && posY == y+height-1 {
				return '└'
			} else if posX == x+width-1 && posY == y+height-1 {
				return '┘'
			} else if posY == y {
				return '┬'
			} else if posY == y+height-1 {
				return '┴'
			} else if posX == x {
				return '├'
			} else if posX == x+width-1 {
				return '┤'
			} else {
				return '┼'
			}
		}
		if hasHorizontalBorder {
			return '─'
		}
		if hasVerticalBorder {
			return '│'
		}
		return ' '
	}
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			// detect border
			if isLabelPos(col, row) {
				screen.SetContent(col, row, popRune(), nil, tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGray))
			} else {
				screen.SetContent(col, row, chooseBorderChar(col, row), nil, tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorBlack))
			}
		}
	}
}

func (k *KeyPad) SetCorrectSequence(sequence []rune) {
	k.correctSequence = sequence
}

func (k *KeyPad) SetOnCompletion(completion func(success bool)) {
	k.onCompletion = completion
}

func (k *KeyPad) handleKeys(event *tcell.EventKey) *tcell.EventKey {
	appendSequence := func(key rune) {
		if len(k.currentSequence) == len(k.correctSequence) {
			k.currentSequence = []rune{key}
			return
		}
		k.currentSequence = append(k.currentSequence, key)
		if len(k.currentSequence) == len(k.correctSequence) {
			k.confirmSequence()
		}
	}
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case '0':
			fallthrough
		case '1':
			fallthrough
		case '2':
			fallthrough
		case '3':
			fallthrough
		case '4':
			fallthrough
		case '5':
			fallthrough
		case '6':
			fallthrough
		case '7':
			fallthrough
		case '8':
			fallthrough
		case '9':
			appendSequence(event.Rune())
		case '*':
			// clear sequence
			k.currentSequence = []rune{}
		case '#':
			// check sequence
			k.onCompletion(false)
		}
	}
	return nil
}

func (k *KeyPad) confirmSequence() {
	if len(k.currentSequence) != len(k.correctSequence) {
		k.currentSequence = []rune{}
		return
	}
	for i := 0; i < len(k.currentSequence); i++ {
		if k.currentSequence[i] != k.correctSequence[i] {
			k.audioPlayer.PlayCue("world/denied")
			return
		}
	}
	if k.onCompletion != nil {
		k.audioPlayer.PlayCue("world/confirmed")
		k.onCompletion(true)
	}
}

func (k *KeyPad) SetAudioPlayer(player foundation.AudioCuePlayer) {
	// do nothing
	k.audioPlayer = player
}
