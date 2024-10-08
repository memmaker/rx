package console

import (
	"RogueUI/foundation"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"strings"
)

type InputCapturer interface {
	SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey)
	GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey
	SetRect(x int, y int, w int, h int)
}

func wrapPrimitivesSideBySide(p, q cview.Primitive, width, height int) cview.Primitive {
	containerLeft := cview.NewFlex()
	containerLeft.SetDirection(cview.FlexRow)
	containerLeft.AddItem(nil, 0, 1, false)
	containerLeft.AddItem(p, height, 1, false)
	containerLeft.AddItem(nil, 0, 1, false)

	containerRight := cview.NewFlex()
	containerRight.SetDirection(cview.FlexRow)
	containerRight.AddItem(nil, 0, 1, false)
	containerRight.AddItem(q, height, 1, true)
	containerRight.AddItem(nil, 0, 1, false)

	flex := cview.NewFlex()
	flex.AddItem(containerLeft, width, 1, false)
	flex.AddItem(nil, 0, 1, false)
	flex.AddItem(containerRight, width, 1, true)
	return flex
}

func wrapPrimitivesTopToBottom(p, q cview.Primitive) cview.Primitive {

	container := cview.NewFlex()
	container.SetDirection(cview.FlexRow)
	container.AddItem(p, 0, 1, false)
	container.AddItem(q, 0, 1, true)

	return container
}

func widthAndHeightFromString(description string) (int, int) {
	longest := 0
	height := 0
	for _, line := range strings.Split(description, "\n") {
		withoutColors := cview.TaggedStringWidth(line)
		longest = max(longest, withoutColors)
		height++
	}
	return longest, height
}

func longestInventoryLineWithoutColorCodes(items []foundation.Item) int {
	longest := 0
	for _, line := range items {
		withoutColors := line.DisplayLength()
		longest = max(longest, withoutColors)
	}
	return longest
}

func drawBackgroundAndBorderWithTitleForInventory(screen tcell.Screen, x int, y int, width int, height int, title string, style tcell.Style, runes []rune) {
	horizontal := runes[0]
	vertical := runes[1]
	topLeft := runes[2]
	topRight := runes[3]
	//bottomRight := runes[4]
	bottomLeft := runes[5]
	fg, _, _ := style.Decompose()
	// fill the background
	for i := x; i < x+width; i++ {
		for j := y; j < y+height; j++ {
			screen.SetContent(i, j, ' ', nil, style)
		}
	}

	// Draw the corners
	cview.Print(screen, []byte(string(topLeft)), x, y, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(topRight)), x+width-1, y, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(bottomLeft)), x, y+height-1, width, cview.AlignLeft, fg)

	// center title
	startTitleX := x + 1 + (width-1-len(title))/2
	endTitleX := startTitleX + len(title)
	// Draw the horizontal borders
	for i := x + 1; i < x+width-1; i++ {
		if title != "" && i >= startTitleX && i < endTitleX {
			cview.Print(screen, []byte(string(title[i-startTitleX])), i, y, width, cview.AlignLeft, fg)
		} else {
			cview.Print(screen, []byte(string(horizontal)), i, y, width, cview.AlignLeft, fg)
		}
	}

	// Draw the vertical borders
	for i := y + 1; i < y+height-1; i++ {
		cview.Print(screen, []byte(string(vertical)), x, i, width, cview.AlignLeft, fg)
	}
}

func drawBackgroundAndBorderWithTitle(screen tcell.Screen, x int, y int, width int, height int, title string, style tcell.Style, runes []rune) {
	horizontal := runes[0]
	vertical := runes[1]
	topLeft := runes[2]
	topRight := runes[3]
	bottomRight := runes[4]
	bottomLeft := runes[5]
	fg, _, _ := style.Decompose()
	// fill the background
	for i := x; i < x+width; i++ {
		for j := y; j < y+height; j++ {
			screen.SetContent(i, j, ' ', nil, style)
		}
	}

	// Draw the corners
	cview.Print(screen, []byte(string(topLeft)), x, y, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(topRight)), x+width-1, y, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(bottomRight)), x+width-1, y+height-1, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(bottomLeft)), x, y+height-1, width, cview.AlignLeft, fg)

	// center title
	startTitleX := x + 1 + (width-1-len(title))/2
	endTitleX := startTitleX + len(title)
	// Draw the horizontal borders
	for i := x + 1; i < x+width-1; i++ {
		if title != "" && i >= startTitleX && i < endTitleX {
			cview.Print(screen, []byte(string(title[i-startTitleX])), i, y, width, cview.AlignLeft, fg)
		} else {
			cview.Print(screen, []byte(string(horizontal)), i, y, width, cview.AlignLeft, fg)
		}
		cview.Print(screen, []byte(string(horizontal)), i, y+height-1, width, cview.AlignLeft, fg)
	}

	// Draw the vertical borders
	for i := y + 1; i < y+height-1; i++ {
		cview.Print(screen, []byte(string(vertical)), x, i, width, cview.AlignLeft, fg)
		cview.Print(screen, []byte(string(vertical)), x+width-1, i, width, cview.AlignLeft, fg)
	}
}

func iterateBorderTiles(x int, y int, width int, height int, runes []rune, setTile func(r rune, x, y int)) {
	horizontal := runes[0]
	vertical := runes[1]
	topLeft := runes[2]
	topRight := runes[3]
	bottomRight := runes[4]
	bottomLeft := runes[5]
	// fill the background

	// Draw the corners
	setTile(topLeft, x, y)
	setTile(topRight, x+width-1, y)
	setTile(bottomRight, x+width-1, y+height-1)
	setTile(bottomLeft, x, y+height-1)

	// center title
	// Draw the horizontal borders
	for i := x + 1; i < x+width-1; i++ {
		setTile(horizontal, i, y)
		setTile(horizontal, i, y+height-1)
	}

	// Draw the vertical borders
	for i := y + 1; i < y+height-1; i++ {
		setTile(vertical, x, i)
		setTile(vertical, x+width-1, i)
	}
}
