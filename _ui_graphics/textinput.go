package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
    "image/color"
    "strings"
)

type EndAction int

const (
    EndActionCancel EndAction = iota
    EndActionConfirm
)

type TextInput struct {
    prompt       string
    currentText  string
    maxLength    int
    gridPos      geometry.Point
    cursorIcon   int32
    cursorFrames int
    shouldClose  bool
    onClose      func(endedWith EndAction, text string)
    uiController *UIController
    drawBorder   bool
    currentTick  uint64
}

func (t *TextInput) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (t *TextInput) OnCommand(command CommandType) bool {
    switch command {
    case PlayerCommandCancel:
        t.ActionCancel()
    case PlayerCommandConfirm:
        t.ActionConfirm()
    case PlayerCommandUp:
        t.ActionUp()
    case PlayerCommandDown:
        t.ActionDown()
    case PlayerCommandLeft:
        t.ActionLeft()
    case PlayerCommandRight:
        t.ActionRight()
    }
    return true
}

func (t *TextInput) ActionUp() {

}

func (t *TextInput) ActionDown() {

}

func (t *TextInput) ActionConfirm() {
    t.onConfirm()
}

func (t *TextInput) OnMouseClicked(button ebiten.MouseButton, x int, y int) bool {
    return false
}

func (t *TextInput) OnAvatarSwitched() {

}

func (t *TextInput) ActionLeft() {

}

func (t *TextInput) ActionRight() {

}

func (t *TextInput) OnMouseMoved(x int, y int) (bool, Tooltip) {
    return false, EmptyTooltip{}
}

func NewTextInput(uiController *UIController, gridPos geometry.Point, maxLength int, cursorIcon int32, cursorFrameCount int, onClose func(endedWith EndAction, text string)) *TextInput {
    return &TextInput{
        gridPos:      gridPos,
        maxLength:    maxLength,
        cursorIcon:   cursorIcon,
        cursorFrames: cursorFrameCount,
        onClose:      onClose,
        uiController: uiController,
    }
}
func (t *TextInput) CenterHorizontallyAtY(y int) {
    neededWidth := t.neededWidth()
    smallScreenSize := t.uiController.GetScreenSize()
    t.gridPos.X = (smallScreenSize.X - neededWidth) / 2
    t.gridPos.Y = y
}

func (t *TextInput) neededWidth() int {
    return len(t.prompt) + t.maxLength + 1 // for the cursor Icon
}
func (t *TextInput) SetPrompt(prompt string) {
    t.prompt = prompt
}

func (t *TextInput) SetDrawBorder(drawBorder bool) {
    t.drawBorder = drawBorder
}
func (t *TextInput) ShouldClose() bool {
    return t.shouldClose
}
func (t *TextInput) Draw() {
    if t.shouldClose {
        return
    }

    renderer := t.uiController.GetRenderer()

    if t.drawBorder {
        //borderTopLeft := t.gridPos.Add(geometry.Point{X: -1, Y: -1})
        //borderBottomRight := t.gridPos.Add(geometry.Point{X: t.neededWidth() + 1, Y: 2})
        //renderer.DrawFilledBorderOnGrid(borderTopLeft, borderBottomRight, "")
    }

    yPos := t.gridPos.Y
    xInputStart := t.gridPos.X
    if t.prompt != "" {
        renderer.DrawStringOnGrid(xInputStart, yPos, t.prompt, color.White)
        xInputStart += len(t.prompt)
    }
    renderer.DrawStringOnGrid(xInputStart, yPos, t.currentText, color.White)
    xCursor := xInputStart + len(t.currentText)
    renderer.DrawOnGrid(xCursor, yPos, t.cursorFromTick(t.currentTick))
}

func (t *TextInput) ActionCancel() {
    t.onCancel()
}
func (t *TextInput) OnKeyPressed(key ebiten.Key, shiftPressed bool) {
    if t.shouldClose {
        return
    }
    if key == ebiten.KeyBackspace && len(t.currentText) > 0 {
        t.currentText = t.currentText[:len(t.currentText)-1]
        return
    }

    if key == ebiten.KeyEnter {
        t.onConfirm()
        return
    }

    if key == ebiten.KeyEscape {
        t.onCancel()
        return
    }
    if key == ebiten.KeySpace {
        t.currentText += " "
        return
    }
    if len(t.currentText) < t.maxLength {
        printable := key.String()
        if strings.HasPrefix(printable, "Digit") {
            printable = printable[5:]
        }
        // only accept standard printable ascii characters
        if len(printable) != 1 || printable[0] < 32 || printable[0] > 126 {
            return
        }
        if ebiten.IsKeyPressed(ebiten.KeyShift) {
            printable = strings.ToUpper(printable)
        } else {
            printable = strings.ToLower(printable)
        }
        t.currentText += printable
    }
}

func (t *TextInput) onCancel() {
    t.onClose(EndActionCancel, "")
    t.shouldClose = true
}

func (t *TextInput) onConfirm() {
    t.onClose(EndActionConfirm, t.currentText)
    t.shouldClose = true
}

func (t *TextInput) cursorFromTick(tick uint64) int32 {
    if t.cursorFrames == 1 {
        return t.cursorIcon
    }
    delays := tick / 20
    return t.cursorIcon + int32(delays%uint64(t.cursorFrames))
}

func (t *TextInput) SetMaxLength(length int) {
    t.maxLength = length
}

func (t *TextInput) SetTick(tick uint64) {
    t.currentTick = tick
}
