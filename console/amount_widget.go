package console

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"strconv"
)

type AmountWidget struct {
	*cview.Box
	maxAmount              int
	textEntered            string
	close                  func(amount int)
	doneButton             *cview.Button
	cancelButton           *cview.Button
	madeFirstKeyboardInput bool
	amountStyle            tcell.Style
	setFocus               func(p cview.Primitive)
}

func NewAmountWidget(itemName string, maxAmount int, close func(amount int), setFocus func(p cview.Primitive)) *AmountWidget {
	box := cview.NewBox()

	a := &AmountWidget{
		Box:         box,
		maxAmount:   maxAmount,
		close:       close,
		setFocus:    setFocus,
		amountStyle: tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack).Bold(true),
	}

	a.SetBorder(true)
	a.SetTitle(fmt.Sprintf("Move %s", itemName))
	a.SetTitleAlign(cview.AlignCenter)

	a.SetInputCapture(a.handleInput)
	a.SetMouseCapture(a.handleMouse)
	a.cancelButton = cview.NewButton("Cancel")
	a.cancelButton.SetSelectedFunc(func() {
		a.close(0)
	})
	a.cancelButton.SetBlurFunc(func(key tcell.Key) {
		a.setFocus(a.doneButton)
	})

	a.doneButton = cview.NewButton("Done")
	a.doneButton.SetBlurFunc(func(key tcell.Key) {
		a.setFocus(a.cancelButton)
	})
	a.doneButton.SetSelectedFunc(func() {
		amount, err := strconv.Atoi(a.textEntered)
		if err == nil {
			a.close(amount)
		}
	})
	a.SetAmount(maxAmount)
	return a
}

func (a *AmountWidget) SetAmount(amount int) {
	a.textEntered = strconv.Itoa(min(max(amount, 0), a.maxAmount))
}
func (a *AmountWidget) SetAmountAsText(text string) {
	atoi, _ := strconv.Atoi(text)
	a.SetAmount(atoi)
}

func (a *AmountWidget) SetRect(x, y, width, height int) {
	a.Box.SetRect(x, y, width, height)
	a.cancelButton.SetRect(x+1, y+height-2, width/2-2, 1)
	a.doneButton.SetRect(x+width/2+1, y+height-2, width/2-2, 1)
}

func (a *AmountWidget) Draw(screen tcell.Screen) {
	a.Box.Draw(screen)
	x, y, width, _ := a.GetInnerRect()

	// right align the text label, at the top right
	amountStartX := x + width - len(a.textEntered)

	amountStartY := y
	cview.PrintStyle(screen, []byte(a.textEntered), amountStartX, amountStartY, width, cview.AlignLeft, a.amountStyle)

	lessIcon := " < "
	moreIcon := " > "

	upIconStartX := x
	upIconStartY := y + 2

	downIconStartX := x + width - len(moreIcon)
	downIconStartY := y + 2

	cview.PrintStyle(screen, []byte(lessIcon), upIconStartX, upIconStartY, width, cview.AlignLeft, tcell.StyleDefault.Background(tcell.ColorGray))
	cview.PrintStyle(screen, []byte(moreIcon), downIconStartX, downIconStartY, width, cview.AlignLeft, tcell.StyleDefault.Background(tcell.ColorGray))

	a.doneButton.Draw(screen)
	a.cancelButton.Draw(screen)
}

func (a *AmountWidget) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(a.textEntered) > 0 {
			a.textEntered = a.textEntered[:len(a.textEntered)-1]
		}
	case tcell.KeyEnter:
		if len(a.textEntered) > 0 {
			amount, err := strconv.Atoi(a.textEntered)
			if err == nil {
				a.close(amount)
			}
		}
	case tcell.KeyTab:
		if a.doneButton.HasFocus() {
			a.setFocus(a.cancelButton)
		} else {
			a.setFocus(a.doneButton)
		}
		return nil
	case tcell.KeyEscape:
		a.close(0)
	case tcell.KeyLeft:
		amount := a.GetAmount()
		amount--
		a.SetAmount(amount)
	case tcell.KeyRight:
		amount := a.GetAmount()
		amount++
		a.SetAmount(amount)
	case tcell.KeyRune:
		if !a.madeFirstKeyboardInput {
			a.textEntered = ""
			a.madeFirstKeyboardInput = true
		}
		switch event.Rune() {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			a.SetAmountAsText(a.textEntered + string(event.Rune()))
		}
	}
	return event
}

func (a *AmountWidget) GetAmount() int {
	amount, err := strconv.Atoi(a.textEntered)
	if err != nil {
		return 0
	}
	return amount
}

func (a *AmountWidget) handleMouse(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
	x, y := event.Position()
	xOffset, yOffset, w, _ := a.GetInnerRect()
	relativeX := x - xOffset
	relativeY := y - yOffset
	isMoreOrLess := relativeY == 2
	isButton := relativeY == 4
	if action == cview.MouseLeftClick {
		if isMoreOrLess {
			if relativeX < w/2 {
				amount := a.GetAmount()
				amount--
				a.SetAmount(amount)
			} else {
				amount := a.GetAmount()
				amount++
				a.SetAmount(amount)
			}
			return action, nil
		}
		if isButton {
			if relativeX < w/2 {
				a.cancelButton.MouseHandler()(action, event, a.setFocus)
			} else {
				a.doneButton.MouseHandler()(action, event, a.setFocus)
			}
			return action, nil
		}
	}
	return action, event
}
