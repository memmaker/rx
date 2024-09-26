package console

import (
	"RogueUI/foundation"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"strings"
	"unicode"
)

type TextInventory struct {
	*cview.Box
	items                  []foundation.Item
	defaultSelection       func(item foundation.Item)
	shiftSelection         func(item foundation.Item)
	controlSelection       func(item foundation.Item)
	listWidth              int
	listHeight             int
	closeHandler           func()
	isEquipped             func(item foundation.Item) bool
	style                  tcell.Style
	closeOnSelect          bool
	closeOnShiftSelect     bool
	closeOnControlSelect   bool
	ourTitle               string
	selectionOnly          bool
	lineColor              func(foundation.ItemCategory) color.RGBA
	cursorAtIndex          int
	stringLabelsWithWeight []string
	isOverEncumbered       func() bool
	afterClose             func()
}

func (i *TextInventory) SetLineColor(lineColor func(foundation.ItemCategory) color.RGBA) {
	i.lineColor = lineColor
}
func (i *TextInventory) drawInside(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	// align top right
	startX := x + width - i.listWidth - 2
	startY := y
	fg, _, _ := i.style.Decompose()
	//cview.Borders.Cross
	runes := []rune{cview.Borders.HorizontalFocus, cview.Borders.VerticalFocus, cview.Borders.TopLeftFocus, cview.Borders.TopRightFocus, cview.Borders.BottomRightFocus, cview.Borders.BottomLeftFocus}
	drawBackgroundAndBorderWithTitleForInventory(screen, startX, startY, i.listWidth+2, i.listHeight+2, i.ourTitle, i.style, runes)

	listOffset := geometry.Point{X: 2, Y: 1}
	var equipRunes []rune
	var unequipRunes []rune
	var useRunes []rune
	var dropRunes []rune
	var totalWeight int
	for lineIndex, invItem := range i.items {
		totalWeight += invItem.GetCarryWeight()
		item := invItem
		shortcut := invItem.Shortcut()
		line := i.stringLabelsWithWeight[lineIndex]
		taggedStringWidth := cview.TaggedStringWidth(line)
		if taggedStringWidth < i.listWidth {
			line = RightPadColored(line, i.listWidth)
		}
		if i.isEquipped != nil && i.isEquipped(item) {
			line = line[:2] + "+" + line[3:]
		}
		drawX := startX + listOffset.X
		drawY := startY + listOffset.Y + lineIndex
		drawStyle := i.style.Foreground(fg)
		if i.cursorAtIndex == lineIndex {
			drawStyle = drawStyle.Reverse(true)
		}
		cview.PrintStyle(screen, []byte(line), drawX, drawY, width, cview.AlignLeft, drawStyle)

		if item.IsEquippable() && i.isEquipped != nil {
			if i.isEquipped(item) {
				unequipRunes = append(unequipRunes, shortcut)
			} else {
				equipRunes = append(equipRunes, shortcut)
			}
		}
		if item.IsUsableOrZappable() {
			useRunes = append(useRunes, shortcut)
		}
		dropRunes = append(dropRunes, shortcut)
	}

	lineAfterList := startY + listOffset.Y + i.listHeight + 1

	getWeightLine := func() string {
		weightLeft := "Weight:"
		weightColor := "[#00FF00]"
		if i.isOverEncumbered() {
			weightColor = "[#FF0000]"
		}
		weightRight := fmt.Sprintf("%s%d[-]lbs", weightColor, totalWeight)
		centerSpaceCount := (i.listWidth + 2) - len(weightLeft) - cview.TaggedStringWidth(weightRight)
		weightLine := fmt.Sprintf("%s%s%s", weightLeft, strings.Repeat(" ", centerSpaceCount), weightRight)

		return weightLine
	}

	if i.selectionOnly {
		for lineX := 0; lineX < i.listWidth+2; lineX++ {
			screen.SetContent(startX+lineX, lineAfterList, ' ', nil, i.style)
		}
		cview.Print(screen, []byte(getWeightLine()), startX, lineAfterList, width, cview.AlignLeft, fg)
		return x, y, width, height
	}
	var infoLines []string

	infoLines = append(infoLines, getWeightLine())

	if len(unequipRunes) > 0 {
		infoLines = append(infoLines, cview.Escape(fmt.Sprintf("[%s] Unequip", string(unequipRunes))))
	}
	if len(equipRunes) > 0 {
		infoLines = append(infoLines, cview.Escape(fmt.Sprintf("[%s] Equip", string(equipRunes))))
	}
	if len(useRunes) > 0 {
		infoLines = append(infoLines, cview.Escape(fmt.Sprintf("[<CTRL> + %s] Use", string(useRunes))))
	}
	if len(dropRunes) > 0 {
		infoLines = append(infoLines, cview.Escape(fmt.Sprintf("[%s] Drop", strings.ToUpper(string(dropRunes)))))
	}

	additionalLines := len(infoLines)

	for lineY := 0; lineY < additionalLines; lineY++ {
		for lineX := 0; lineX < i.listWidth+2; lineX++ {
			screen.SetContent(startX+lineX, lineAfterList+lineY, ' ', nil, i.style)
		}
	}

	for idx, line := range infoLines {
		cview.Print(screen, []byte(line), startX, lineAfterList+idx, width, cview.AlignLeft, fg)
	}

	return x, y, width, height
}

func NewTextInventory(isOverEncumbered func() bool) *TextInventory {
	box := cview.NewBox()
	//box.SetBorder(true)
	t := &TextInventory{
		Box:   box,
		items: []foundation.Item{},
		lineColor: func(category foundation.ItemCategory) color.RGBA {
			return color.RGBA{R: 170, G: 170, B: 170, A: 255}
		},
		cursorAtIndex:    -1,
		isOverEncumbered: isOverEncumbered,
	}
	box.SetDrawFunc(t.drawInside)
	box.SetBackgroundTransparent(true)
	box.SetInputCapture(t.handleInput)
	return t
}
func (i *TextInventory) SetStyle(style tcell.Style) {
	i.style = style
}
func (i *TextInventory) SetTitle(title string) {
	i.ourTitle = title
	i.listWidth = max(i.listWidth, len(title))
}
func (i *TextInventory) SetDefaultSelection(onSelect func(item foundation.Item)) {
	i.defaultSelection = onSelect
}

func (i *TextInventory) SetShiftSelection(onSelect func(item foundation.Item)) {
	i.shiftSelection = onSelect
}

func (i *TextInventory) SetControlSelection(onSelect func(item foundation.Item)) {
	i.controlSelection = onSelect
}

func (i *TextInventory) SetItems(invItem []foundation.Item) {
	i.items = invItem
	i.updateListBounds()
}

func (i *TextInventory) updateListBounds() {
	labels := make([]fxtools.TableRow, len(i.items))
	for lineIndex, invItem := range i.items {
		namePart := invItem.InventoryNameWithColorsAndShortcut(textiles.RGBAToFgColorCode(i.lineColor(invItem.Category())))
		weightPart := fmt.Sprintf("[#00FF00]%d[-]lbs", invItem.GetCarryWeight())
		row := fxtools.NewTableRow(namePart, weightPart)
		labels[lineIndex] = row
	}
	itemLabels := fxtools.TableLayoutLastRight(labels)
	i.stringLabelsWithWeight = itemLabels

	var width int
	for _, line := range itemLabels {
		withoutColors := cview.TaggedStringWidth(line)
		width = max(width, withoutColors)
	}

	height := len(i.items)
	i.listWidth = max(i.listWidth, width)
	i.listHeight = height
}

func (i *TextInventory) SetEquippedTest(isEquipped func(item foundation.Item) bool) {
	i.isEquipped = isEquipped
}

func (i *TextInventory) SetCloseHandler(escapeHandler func()) {
	i.closeHandler = escapeHandler
}

func (i *TextInventory) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if i.closeHandler != nil && event.Key() == tcell.KeyEscape {
		i.Close()
		return nil
	}

	if event.Key() == tcell.KeyUp {
		i.cursorAtIndex = i.cursorAtIndex - 1
		if i.cursorAtIndex < 0 {
			i.cursorAtIndex = len(i.items) - 1
		}
		return nil
	} else if event.Key() == tcell.KeyDown {
		i.cursorAtIndex = i.cursorAtIndex + 1
		if i.cursorAtIndex >= len(i.items) {
			i.cursorAtIndex = 0
		}
		return nil
	}

	if event.Key() == tcell.KeyEnter {
		if i.defaultSelection != nil && i.cursorAtIndex >= 0 && i.cursorAtIndex < len(i.items) {
			if i.closeOnSelect {
				i.Close()
			}
			i.defaultSelection(i.items[i.cursorAtIndex])
			i.updateListBounds()
			return nil
		}
	}

	runeReceived := event.Rune()
	// to upper
	modCtrl := event.Modifiers() == tcell.ModAlt || event.Modifiers() == tcell.ModCtrl || event.Modifiers() == tcell.ModMeta
	if modCtrl { // 1 == a, 2 == b, etc
		runeReceived = runeReceived + 96
	}
	modShift := unicode.IsUpper(runeReceived)
	if modShift {
		runeReceived = unicode.ToLower(runeReceived)
	}

	for _, invItem := range i.items {
		if runeReceived == invItem.Shortcut() {
			if modShift {
				if i.shiftSelection != nil {
					if i.closeOnShiftSelect {
						i.Close()
					}
					i.shiftSelection(invItem)
				}
			} else if modCtrl {
				if i.controlSelection != nil {
					if i.closeOnControlSelect {
						i.Close()
					}
					i.controlSelection(invItem)
				}
			} else if i.defaultSelection != nil {
				if i.closeOnSelect {
					i.Close()
				}
				i.defaultSelection(invItem)
			}
			i.updateListBounds()
			return nil
		}
	}
	return event
}

func (i *TextInventory) SetCloseOnSelection(value bool) {
	i.closeOnSelect = value
}

func (i *TextInventory) SetCloseOnShiftSelection(value bool) {
	i.closeOnShiftSelect = value
}

func (i *TextInventory) SetCloseOnControlSelection(value bool) {
	i.closeOnControlSelect = value
}

func (i *TextInventory) SetSelectionMode() {
	i.selectionOnly = true
}

func (i *TextInventory) Close() {
	if i.closeHandler != nil {
		i.closeHandler()
	}
	if i.afterClose != nil {
		i.afterClose()
	}
}

func (i *TextInventory) SetAfterClose(afterClose func()) {
	i.afterClose = afterClose
}
