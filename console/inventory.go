package console

import (
	"RogueUI/cview"
	"RogueUI/foundation"
	"RogueUI/geometry"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"image/color"
	"strings"
	"unicode"
)

type TextInventory struct {
	*cview.Box
	items                []foundation.ItemForUI
	defaultSelection     func(item foundation.ItemForUI)
	shiftSelection       func(item foundation.ItemForUI)
	controlSelection     func(item foundation.ItemForUI)
	listWidth            int
	listHeight           int
	closeHandler         func()
	isEquipped           func(item foundation.ItemForUI) bool
	style                tcell.Style
	closeOnSelect        bool
	closeOnShiftSelect   bool
	closeOnControlSelect bool
	ourTitle             string
	selectionOnly        bool
	lineColor            func(foundation.ItemCategory) color.RGBA
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
	runes := []rune{cview.Borders.Horizontal, cview.Borders.Vertical, cview.Borders.TopLeft, cview.Borders.TopRight, cview.Borders.BottomRight, cview.Borders.BottomLeft}
	drawBackgroundAndBorderWithTitleForInventory(screen, startX, startY, i.listWidth+2, i.listHeight+2, i.ourTitle, i.style, runes)

	listOffset := geometry.Point{X: 2, Y: 1}
	var equipRunes []rune
	var unequipRunes []rune
	var useRunes []rune
	var dropRunes []rune

	for lineIndex, invItem := range i.items {
		item := invItem
		shortcut := invItem.Shortcut()
		line := invItem.InventoryNameWithColorsAndShortcut(RGBAToFgColorCode(i.lineColor(invItem.GetCategory())))
		taggedStringWidth := cview.TaggedStringWidth(line)
		if taggedStringWidth < i.listWidth {
			line = RightPadColored(line, i.listWidth)
		}
		if i.isEquipped != nil && i.isEquipped(item) {
			line = line[:2] + "+" + line[3:]
		}
		drawX := startX + listOffset.X
		drawY := startY + listOffset.Y + lineIndex
		cview.Print(screen, []byte(line), drawX, drawY, width, cview.AlignLeft, fg)
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
	var infoLines []string
	if len(unequipRunes) > 0 {
		infoLines = append(infoLines, fmt.Sprintf("[%s] Unequip", string(unequipRunes)))
	}
	if len(equipRunes) > 0 {
		infoLines = append(infoLines, fmt.Sprintf("[%s] Equip", string(equipRunes)))
	}
	if len(useRunes) > 0 {
		infoLines = append(infoLines, fmt.Sprintf("[<CTRL> + %s] Use", string(useRunes)))
	}
	if len(dropRunes) > 0 {
		infoLines = append(infoLines, fmt.Sprintf("[%s] Drop", strings.ToUpper(string(dropRunes))))
	}
	if i.selectionOnly {
		return x, y, width, height
	}
	additionalLines := len(infoLines)
	// draw more background

	for lineY := 0; lineY < additionalLines; lineY++ {
		for lineX := 0; lineX < i.listWidth+2; lineX++ {
			screen.SetContent(startX+lineX, lineAfterList+lineY, ' ', nil, i.style)
		}
	}

	for idx, line := range infoLines {
		cview.Print(screen, []byte(cview.Escape(line)), startX, lineAfterList+idx, width, cview.AlignLeft, fg)
	}

	return x, y, width, height
}

func NewTextInventory() *TextInventory {
	box := cview.NewBox()
	//box.SetBorder(true)
	t := &TextInventory{
		Box:   box,
		items: []foundation.ItemForUI{},
		lineColor: func(category foundation.ItemCategory) color.RGBA {
			return color.RGBA{R: 170, G: 170, B: 170, A: 255}
		},
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
func (i *TextInventory) SetDefaultSelection(onSelect func(item foundation.ItemForUI)) {
	i.defaultSelection = onSelect
}

func (i *TextInventory) SetShiftSelection(onSelect func(item foundation.ItemForUI)) {
	i.shiftSelection = onSelect
}

func (i *TextInventory) SetControlSelection(onSelect func(item foundation.ItemForUI)) {
	i.controlSelection = onSelect
}

func (i *TextInventory) SetItems(invItem []foundation.ItemForUI) {
	i.items = invItem
	i.updateListBounds()
}

func (i *TextInventory) updateListBounds() {
	width := longestInventoryLineWithoutColorCodes(i.items)
	height := len(i.items)
	i.listWidth = max(i.listWidth, width)
	i.listHeight = height
}

func (i *TextInventory) SetEquippedTest(isEquipped func(item foundation.ItemForUI) bool) {
	i.isEquipped = isEquipped
}

func (i *TextInventory) SetCloseHandler(escapeHandler func()) {
	i.closeHandler = escapeHandler
}

func (i *TextInventory) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if i.closeHandler != nil && event.Key() == tcell.KeyEscape {
		i.closeHandler()
		return nil
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
						i.closeHandler()
					}
					i.shiftSelection(invItem)
				}
			} else if modCtrl {
				if i.controlSelection != nil {
					if i.closeOnControlSelect {
						i.closeHandler()
					}
					i.controlSelection(invItem)
				}
			} else if i.defaultSelection != nil {
				if i.closeOnSelect {
					i.closeHandler()
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
