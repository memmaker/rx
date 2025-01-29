package ui_console

import (
	"contractor/foundation"
	"github.com/memmaker/go/cview"
)

func (u *UI) SetConversationState(starterText string, starterOptions []foundation.MenuItem, chatterSource foundation.ChatterSource, isTerminal bool) {
	u.dialogueIsTerminal = isTerminal

	// text field
	if u.dialogueText == nil {
		textField := cview.NewTextView()
		textField.SetBorder(true)
		textField.SetBorderColor(toTcellColor(u.uiTheme.GetColorByName("neon_green_2")))
		textField.SetScrollable(true)
		textField.SetDynamicColors(true)
		textField.SetWordWrap(true)
		u.dialogueText = textField
		if isTerminal {
			u.audioPlayer.PlayCue("ui/terminal_poweron")
		}
	}
	u.dialogueText.SetTitle(chatterSource.Name())
	u.dialogueText.SetText(starterText)

	// menu
	if u.dialogueOptions == nil {
		choicesMenu := cview.NewList()
		u.applyListStyle(choicesMenu)
		u.dialogueOptions = choicesMenu
	}
	panelName := "conversation"
	u.dialogueOptions.SetSelectedFunc(func(index int, listItem *cview.ListItem) {
		action := starterOptions[index]
		if action.CloseMenus {
			u.popPanel(panelName)
		}
		action.Action()
	})

	u.makeTopAndBottomModal(panelName, u.dialogueText, u.dialogueOptions)
	u.lockFocusToPrimitive(u.dialogueOptions)

	originalCapture := u.dialogueOptions.GetInputCapture()

	if u.settings.DialogueShortcutsAreNumbers {
		u.dialogueOptions.SetInputCapture(u.directionalWrapperWithoutNumbers(originalCapture))
		setListItemsFromMenuItemsWithNumbers(u.dialogueOptions, starterOptions)
	} else {
		u.dialogueOptions.SetInputCapture(u.directionalWrapperWithoutAlphabet(originalCapture))
		setListItemsFromMenuItems(u.dialogueOptions, starterOptions)
	}

}

func (u *UI) CloseConversation() {
	if u.dialogueIsTerminal {
		u.audioPlayer.PlayCue("ui/terminal_poweroff")
	}
	u.pages.RemovePanel("conversation")

	u.dialogueOptions = nil
	u.dialogueText = nil
	u.dialogueIsTerminal = false

	u.resetFocusToMain()
}
