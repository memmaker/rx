package console

import (
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"strconv"
)

type AmountWidget struct {
	*cview.Box
	maxAmount    int
	textEntered  string
	close        func(confirmed bool, amount int)
	doneButton   *cview.Button
	cancelButton *cview.Button
}

func NewAmountWidget(maxAmount int, close func(confirmed bool, amount int)) *AmountWidget {
	box := cview.NewBox()

	a := &AmountWidget{
		Box:       box,
		maxAmount: maxAmount,
		close:     close,
	}

	a.SetBorder(true)
	a.SetTitle("Move items")
	a.SetTitleAlign(cview.AlignCenter)

	a.SetInputCapture(a.handleInput)

	a.cancelButton = cview.NewButton("Cancel")
	a.cancelButton.SetSelectedFunc(func() {
		a.close(false, 0)
	})

	a.doneButton = cview.NewButton("Done")
	a.doneButton.SetSelectedFunc(func() {
		amount, err := strconv.Atoi(a.textEntered)
		if err == nil {
			a.close(true, amount)
		}
	})
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
	a.doneButton.SetRect(x, y+height-3, width/2, 3)
	a.cancelButton.SetRect(x+width/2, y+height-3, width/2, 3)
}

func (a *AmountWidget) Draw(screen tcell.Screen) {
	a.Box.Draw(screen)
	x, y, width, _ := a.GetInnerRect()

	// right align the text label, at the top right
	amountStartX := x + width - len(a.textEntered)
	if amountStartX < x {
		amountStartX = x
	}
	amountStartY := y + 1
	cview.PrintStyle(screen, []byte(a.textEntered), amountStartX, amountStartY, width, cview.AlignRight, tcell.StyleDefault)

	upIcon := "^"
	downIcon := "v"

	upIconStartX := x + width - len(upIcon)
	upIconStartY := y + 2

	downIconStartX := x + width - len(downIcon)
	downIconStartY := y + 4

	cview.PrintStyle(screen, []byte(upIcon), upIconStartX, upIconStartY, width, cview.AlignRight, tcell.StyleDefault)
	cview.PrintStyle(screen, []byte(downIcon), downIconStartX, downIconStartY, width, cview.AlignRight, tcell.StyleDefault)

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
				a.close(true, amount)
			}
		}
	case tcell.KeyEscape:
		a.close(false, 0)
	case tcell.KeyUp:
		amount := a.GetAmount()
		amount++
		a.SetAmount(amount)
	case tcell.KeyDown:
		amount := a.GetAmount()
		amount--
		a.SetAmount(amount)
	case tcell.KeyRune:
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
