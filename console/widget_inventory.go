package console

import (
	"contractor/foundation"
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

func (u *UI) openInventory(items []foundation.Item) *TextInventory {
	panelName := "inventory"
	screenWidth, screenHeight := u.application.GetScreen().Size()
	inventory := NewTextInventory(screenWidth, screenHeight, u.game.IsPlayerOverEncumbered)
	fg := u.uiTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.uiTheme.GetUIColorForTcell(UIColorUIBackground)

	borderFgFocus := u.uiTheme.GetUIColorForTcell(UIColorBorderForegroundFocus)

	inventory.SetBorder(false)
	inventory.SetPadding(1, 1, 1, 1)
	inventory.SetWrapAround(false)
	inventory.SetHover(true)
	inventory.ShowSecondaryText(false)

	inventory.SetScrollBarColor(fg)
	//list.SetHighlightFullLine(true)

	inventory.SetTitleColor(fg)
	inventory.SetMainTextColor(fg)
	inventory.SetSecondaryTextColor(fg)

	inventory.SetBorderColor(fg)
	inventory.SetBorderColorFocused(borderFgFocus)

	inventory.SetBackgroundColor(bg)

	inventory.SetShortcutColor(fg)

	inventory.SetSelectedTextColor(bg)
	inventory.SetSelectedBackgroundColor(fg)
	inventory.SetForceSelectedTextColor(true)

	inventory.SetLineColor(u.uiTheme.GetInventoryItemColor)
	inventory.SetEquippedTest(u.game.IsEquipped)
	inventory.SetStyle(u.uiTheme.defaultStyle)
	inventory.SetContextMenu(u.game.OpenContextMenuForItem)
	inventory.SetItems(items)

	inventory.SetCloseHandler(func() {
		u.popPanel(panelName)
	})
	u.pages.AddPanel(panelName, inventory, false, true)
	u.pages.ShowPanel(panelName)

	u.lockFocusToPrimitive(inventory)

	originalInputCapture := inventory.GetInputCapture()
	inventory.SetInputCapture(u.directionalWrapperWithoutAlphabet(originalInputCapture))

	//u.makeTopRightModal(panelName, list, len(inventoryItems), longestItem)
	return inventory
}
func (u *UI) dropItemWithAmountSelection(item foundation.Item, after func()) {
	if item.IsMultipleStacks() {
		u.openAmountWidget(item.Name(), item.GetStackSize(), func(amount int) {
			if amount <= 0 || amount > item.GetStackSize() {
				return
			}
			if amount == item.GetStackSize() {
				u.game.PlayerDropItem(item)
				return
			}
			splitItem := item.Split(amount)
			u.game.PlayerDropItem(splitItem)
			if after != nil {
				after()
			}
		})
	} else {
		u.game.PlayerDropItem(item)
		if after != nil {
			after()
		}
	}
}
func (u *UI) OpenInventoryForManagement() {
	getInventory := u.game.GetNonAmmoPlayerInventory

	inv := u.openInventory(getInventory())
	inv.SetTitle("Inventory")
	inv.SetDefaultSelection(func(item foundation.Item) {
		// inventory won't close when doing this
		if item.IsEquippable() {
			u.game.EquipToggle(item)
		} else {
			//inv.Close()
			u.game.PlayerApplyItem(item) // PROBLEM: When consuming the last item, the list won't update correctly
			inv.SetItems(getInventory())
		}
	})

	inv.SetShiftSelection(func(item foundation.Item) {
		// inventory is closed beforehand
		u.dropItemWithAmountSelection(item, func() {
			inv.SetItems(getInventory())
		})
	})
	inv.SetControlSelection(u.game.PlayerExamineItem)

	inv.SetCloseOnControlSelection(false)
	inv.SetCloseOnShiftSelection(false)
}
func (u *UI) OpenInventoryForSelection(itemStacks []foundation.Item, prompt string, onSelected func(item foundation.Item)) {
	u.rightPanel.Clear()
	inv := u.openInventory(itemStacks)
	inv.SetSelectionMode()
	inv.SetTitle(prompt)
	inv.SetDefaultSelection(onSelected)
	inv.SetCloseOnSelection(true)
	inv.SetAfterClose(func() {
		u.UpdateInventory()
	})
}

func (u *UI) OpenInventoryForSelectionWithClose(itemStacks []foundation.Item, prompt string, onSelected func(item foundation.Item), afterClose func()) {
	u.rightPanel.Clear()
	inv := u.openInventory(itemStacks)
	inv.SetSelectionMode()
	inv.SetTitle(prompt)
	inv.SetDefaultSelection(onSelected)
	inv.SetCloseOnSelection(true)
	inv.SetAfterClose(func() {
		u.UpdateInventory()
		afterClose()
	})
}

type TextInventory struct {
	*cview.List
	items                  []foundation.Item
	defaultSelection       func(item foundation.Item)
	shiftSelection         func(item foundation.Item)
	controlSelection       func(item foundation.Item)
	contextMenu            func(item foundation.Item, done func())
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
	stringLabelsWithWeight []string
	isOverEncumbered       func() bool
	afterClose             func()
	infoLines              []string
	screenHeight           int
	screenWidth            int
}

func (i *TextInventory) SetContextMenu(contextMenu func(item foundation.Item, done func())) {
	i.contextMenu = contextMenu
}
func (i *TextInventory) SetLineColor(lineColor func(foundation.ItemCategory) color.RGBA) {
	i.lineColor = lineColor
}
func (i *TextInventory) Draw(screen tcell.Screen) {
	x, y, w, h := i.GetRect()
	i.drawOutside(screen, x, y, w, h)
	i.List.Draw(screen)
}
func (i *TextInventory) drawOutside(screen tcell.Screen, x int, y int, width int, height int) {
	// align top right
	startX := x
	startY := y
	fg, _, _ := i.style.Decompose()
	//cview.Borders.Cross
	runes := []rune{cview.Borders.HorizontalFocus, cview.Borders.VerticalFocus, cview.Borders.TopLeftFocus, cview.Borders.TopRightFocus, cview.Borders.BottomRightFocus, cview.Borders.BottomLeftFocus}
	drawBackgroundAndBorderWithTitleForInventory(screen, startX, startY, i.listWidth+2, i.listHeight+2, i.ourTitle, i.style, runes)

	listOffset := geometry.Point{X: 2, Y: 1}

	lineAfterList := startY + listOffset.Y + i.listHeight + 1

	additionalLines := len(i.infoLines)

	for lineY := 0; lineY < additionalLines; lineY++ {
		for lineX := 0; lineX < i.listWidth+2; lineX++ {
			screen.SetContent(startX+lineX, lineAfterList+lineY, ' ', nil, i.style)
		}
	}

	for idx, line := range i.infoLines {
		cview.Print(screen, []byte(line), startX, lineAfterList+idx, width, cview.AlignLeft, fg)
	}
}

func NewTextInventory(screenWidth int, screenHeight int, isOverEncumbered func() bool) *TextInventory {
	list := cview.NewList()
	list.SetBorder(false)
	t := &TextInventory{
		List:  list,
		items: []foundation.Item{},
		lineColor: func(category foundation.ItemCategory) color.RGBA {
			return color.RGBA{R: 170, G: 170, B: 170, A: 255}
		},
		isOverEncumbered: isOverEncumbered,
		screenWidth:      screenWidth,
		screenHeight:     screenHeight,
	}
	list.SetBackgroundTransparent(true)
	list.SetInputCapture(t.handleInput)
	//list.SetMouseCapture(t.handleMouse)
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
	i.updateListItems()
}

func (i *TextInventory) updateListItems() {
	currentItem := i.GetCurrentItemIndex()
	labels := make([]fxtools.TableRow, len(i.items))
	for lineIndex, invItem := range i.items {
		namePart := invItem.InventoryNameWithColorsAndShortcut(textiles.RGBAToFgColorCode(i.lineColor(invItem.GetCategory())))
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

	i.SetRect(0, 0, i.listWidth+2, i.listHeight+2)

	i.Clear()

	var totalWeight int
	var equipAndUseRunes []rune
	var inspectRunes []rune
	var dropRunes []rune

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

		//cview.PrintStyle(screen, []byte(line), drawX, drawY, width, cview.AlignLeft, drawStyle)

		listItem := cview.NewListItem(line)
		//listItem.SetShortcut(shortcut)
		i.AddItem(listItem)

		if item.IsEquippable() || item.IsUsableOrZappable() || item.IsConsumable() || item.IsReadable() {
			equipAndUseRunes = append(equipAndUseRunes, shortcut)
		}
		inspectRunes = append(inspectRunes, shortcut)
		dropRunes = append(dropRunes, shortcut)
	}

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

	var infoLines []string

	infoLines = append(infoLines, getWeightLine())

	if !i.selectionOnly {
		if len(equipAndUseRunes) > 0 {
			infoLines = append(infoLines, cview.Escape(fmt.Sprintf("[%s] (Un)Equip / Use", string(equipAndUseRunes))))
		}
		if len(inspectRunes) > 0 {
			infoLines = append(infoLines, cview.Escape(fmt.Sprintf("[<CTRL> + letter] Examine")))
		}
		if len(dropRunes) > 0 {
			infoLines = append(infoLines, cview.Escape(fmt.Sprintf("[<SHFT> + letter] Drop")))
		}
	}

	i.infoLines = infoLines
	widthNeeded := i.listWidth + 2

	heightNeeded := i.listHeight + len(infoLines) + 2

	rightX := i.screenWidth - widthNeeded
	i.SetRect(rightX, 0, widthNeeded, heightNeeded)

	if currentItem >= 0 && currentItem < len(i.items) {
		i.SetCurrentItem(currentItem)
	}
}

func (i *TextInventory) SetEquippedTest(isEquipped func(item foundation.Item) bool) {
	i.isEquipped = isEquipped
}

func (i *TextInventory) SetCloseHandler(escapeHandler func()) {
	i.closeHandler = escapeHandler
}
func (i *TextInventory) previousItem() {
	itemIndex := i.GetCurrentItemIndex()
	i.SetCurrentItem(itemIndex - 1)
}

func (i *TextInventory) nextItem() {
	itemIndex := i.GetCurrentItemIndex()
	i.SetCurrentItem(itemIndex + 1)
}
func (i *TextInventory) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if i.closeHandler != nil && event.Key() == tcell.KeyEscape {
		i.Close()
		return nil
	}

	if event.Key() == tcell.KeyUp {
		i.previousItem()
		return nil
	} else if event.Key() == tcell.KeyDown {
		i.nextItem()
		return nil
	}

	if event.Key() == tcell.KeyEnter {
		currentIndex := i.GetCurrentItemIndex()
		if i.defaultSelection != nil && currentIndex >= 0 && currentIndex < len(i.items) {
			if i.closeOnSelect {
				i.Close()
			}
			i.defaultSelection(i.items[currentIndex])
			i.updateListItems()
			return nil
		}
	}

	if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
		currentIndex := i.GetCurrentItemIndex()
		if i.defaultSelection != nil && currentIndex >= 0 && currentIndex < len(i.items) {
			i.contextMenu(i.items[currentIndex], func() {
				i.updateListItems()
			})
			return nil
		}
	}

	event = parseControlCodes(event)

	runeReceived := event.Rune()

	modCtrl := event.Modifiers() == tcell.ModAlt || event.Modifiers() == tcell.ModCtrl || event.Modifiers() == tcell.ModMeta

	if modCtrl && runeReceived < 97 { // 1 == a, 2 == b, etc
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
			i.updateListItems()
			return nil
		}
	}
	return event
}

func parseControlCodes(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyCtrlA {
		return tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlB {
		return tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlC {
		return tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlD {
		return tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlE {
		return tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlF {
		return tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlG {
		return tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlH {
		return tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlI {
		return tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlJ {
		return tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlK {
		return tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlL {
		return tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlM {
		return tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlN {
		return tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlO {
		return tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlP {
		return tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlQ {
		return tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlR {
		return tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlS {
		return tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlT {
		return tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlU {
		return tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlV {
		return tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlW {
		return tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlX {
		return tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlY {
		return tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModCtrl)
	} else if event.Key() == tcell.KeyCtrlZ {
		return tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModCtrl)
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

func (i *TextInventory) handleMouse(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
	x, y := event.Position()
	xOffset, yOffset, w, h := i.GetInnerRect()
	relativeX := x - xOffset
	relativeY := y - yOffset
	isOverList := relativeX > 1 && relativeX < w-1 && relativeY > 1 && relativeY < h-1
	if action == cview.MouseLeftClick {
		if isOverList {
			clickedIndex := relativeY - 1
			if clickedIndex >= 0 && clickedIndex < len(i.items) {
				if i.defaultSelection != nil {
					if i.closeOnSelect {
						i.Close()
					}
					i.defaultSelection(i.items[clickedIndex])
					i.updateListItems()
				}
			}
			return action, nil
		}
	}
	return action, event
}
