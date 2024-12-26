package console

import (
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
)

func (u *UI) AskForConfirmation(title, message string, choice func(didConfirm bool)) {
	dialogue := OpenConfirmDialogue(u, u.application, u.pages, title, message, choice)
	buttonOne := dialogue.GetForm().GetButton(0)
	oldBlurOne := buttonOne.GetInputCapture()
	buttonOne.SetInputCapture(u.directionalWrapper(oldBlurOne))

	buttonTwo := dialogue.GetForm().GetButton(1)
	oldBlurTwo := buttonTwo.GetInputCapture()
	buttonTwo.SetInputCapture(u.directionalWrapper(oldBlurTwo))
}

func (u *UI) AskForString(prompt string, prefill string, result func(entered string)) {
	cview.AskForString(u.application, u.pages, prompt, prefill, result)
}

func (u *UI) FadeToBlack() {
	cview.FadeToBlack(u.application, u.settings.AnimationDelay/2, 10, false)
}

func (u *UI) FadeFromBlack() {
	cview.FadeFromBlack(u.application, u.settings.AnimationDelay/2, 10, false)
}
func (u *UI) popPanel(panelName string) {
	u.pages.RemovePanel(panelName)
	//u.pages.SetCurrentPanel("main")
	u.popFocus()
}

func (u *UI) popFocus() {
	if len(u.focusStack) == 0 {
		u.resetFocusToMain()
		return
	}
	popped := u.focusStack[len(u.focusStack)-1]
	u.focusStack = u.focusStack[:len(u.focusStack)-1]
	u.application.SetBeforeFocusFunc(nil)
	u.application.SetFocus(popped.Primitive)
	u.application.SetBeforeFocusFunc(popped.BeforeFocus)
}

func (u *UI) pushFocus() {
	focus := u.application.GetFocus()
	if focus == nil {
		return
	}
	for _, f := range u.focusStack {
		if f.Primitive == focus {
			return
		}
	}

	u.focusStack = append(u.focusStack, FocusInfo{
		Primitive:   focus,
		BeforeFocus: u.application.GetBeforeFocusFunc(),
	})
}
func (u *UI) defaultFocusHandler(p cview.Primitive) bool {
	if p == u.mainGrid || p == u.mapWindow {
		return true
	}
	return false
}

func (u *UI) makeCenteredModal(panelName string, modal InputPrimitive, w, h int) {
	w = w + 2
	h = h + 2
	screenW, screenH := u.application.GetScreen().Size()
	if h > screenH {
		h = screenH
		w = w + 1 // scrollbar
	}
	if w > screenW {
		w = screenW
	}
	x, y := (screenW-w)/2, (screenH-h)/2
	modal.SetRect(x, y, w, h)
	/*
		originalInputCapture := modal.GetInputCapture()
		modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			command := u.getAdvancedTargetingCommandForKey(toUIKey(event))
			if command == "target_cancel" {
				u.popPanel()
			}
			return originalInputCapture(event)
		})

	*/
	u.pages.AddPanel(panelName, modal, false, true)
	u.lockFocusToPrimitive(modal)
}
func (u *UI) makeTopAndBottomModal(panelName string, primitive, qPrimitive cview.Primitive) {
	modalContainer := wrapPrimitivesTopToBottom(primitive, qPrimitive)

	if inputCapturer, ok := primitive.(InputCapturer); ok {
		inputCapturer.SetInputCapture(u.popOnEscape(panelName))
	}

	if inputCapturer, ok := qPrimitive.(InputCapturer); ok {
		inputCapturer.SetInputCapture(u.popOnEscape(panelName))
	}
	u.pages.AddPanel(panelName, modalContainer, true, true)
	u.pages.ShowPanel(panelName)
}
func (u *UI) resetFocusToMain() {
	u.focusStack = nil
	u.application.SetBeforeFocusFunc(nil)
	u.application.SetFocus(u.mainGrid)
	u.pages.ShowPanel("main")
	u.application.SetBeforeFocusFunc(u.defaultFocusHandler)
}
func (u *UI) lockFocusToPrimitive(p cview.Primitive) {
	u.pushFocus()
	u.application.SetBeforeFocusFunc(nil)
	u.application.SetFocus(p)
	u.application.SetBeforeFocusFunc(func(p cview.Primitive) bool { return false })
}

func (u *UI) popOnEscape(panelName string) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			u.popPanel(panelName)
		}
		return event
	}
}

func (u *UI) yesNoReceiver(yes, no func()) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'y' || event.Rune() == 'Y' {
			yes()
			return nil
		}
		if event.Rune() == 'n' || event.Rune() == 'N' {
			no()
			return nil
		}
		return event
	}
}

func (u *UI) popOnAnyKeyWithNotification(currentPage string, onClose func()) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		u.pages.HidePanel(currentPage)
		onClose()
		return nil
	}
}

func (u *UI) popOnSpaceWithNotification(currentPage string, onClose func()) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == ' ' {
			u.pages.HidePanel(currentPage)
			onClose()
			return nil
		}
		return event
	}
}

func (u *UI) applyStylingToUI() {
	u.uiTheme.SetBorders(&cview.Borders)

	fg := u.uiTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.uiTheme.GetUIColorForTcell(UIColorUIBackground)

	u.statusBar.SetBackgroundColor(u.uiTheme.GetUIColorForTcell(UIColorStatusBarBackground))
	u.statusBar.SetTextColor(u.uiTheme.GetUIColorForTcell(UIColorStatusBarForeground))

	u.messageLabel.SetTextColor(fg)
	u.messageLabel.SetBackgroundColor(bg)
	u.messageLabel.SetBorderColor(fg)
	u.messageLabel.SetScrollBarColor(fg)
	u.messageLabel.SetDynamicColors(true)

	u.rightPanel.SetTextColor(fg)
	u.rightPanel.SetBorderColor(fg)
	u.rightPanel.SetBackgroundColor(bg)
	u.rightPanel.SetDynamicColors(true)
	u.rightPanel.SetTextAlign(cview.AlignRight)

	u.lowerRightPanel.SetTextColor(fg)
	u.lowerRightPanel.SetBorderColor(fg)
	u.lowerRightPanel.SetBackgroundColor(bg)
	u.lowerRightPanel.SetDynamicColors(true)
	u.lowerRightPanel.SetTextAlign(cview.AlignLeft)

	u.mapOverlay.SetDefaultColors(tcellColorToRGBA(bg), tcellColorToRGBA(fg))
}

func (u *UI) applyListStyle(list *cview.List) {
	fg := u.uiTheme.GetUIColorForTcell(UIColorUIForeground)
	bg := u.uiTheme.GetUIColorForTcell(UIColorUIBackground)

	borderFgFocus := u.uiTheme.GetUIColorForTcell(UIColorBorderForegroundFocus)

	list.SetBorder(true)
	list.SetWrapAround(true)
	list.SetHover(true)
	list.ShowSecondaryText(false)

	list.SetScrollBarColor(fg)
	//list.SetHighlightFullLine(true)

	list.SetTitleColor(fg)
	list.SetMainTextColor(fg)
	list.SetSecondaryTextColor(fg)

	list.SetBorderColor(fg)
	list.SetBorderColorFocused(borderFgFocus)

	list.SetBackgroundColor(bg)

	list.SetShortcutColor(fg)

	list.SetSelectedTextColor(bg)
	list.SetSelectedBackgroundColor(fg)
	list.SetForceSelectedTextColor(true)
	//list.SetSelectedTextAttributes(tcell.AttrReverse)
}
