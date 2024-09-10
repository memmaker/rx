package console

import (
	"github.com/memmaker/go/cview"
)

func OpenConfirmDialogue(app *cview.Application, panels *cview.Panels, title string, msg string, result func(didConfirm bool)) *cview.Modal {
	oldBeforeFocusFunc := app.GetBeforeFocusFunc()
	oldFocus := app.GetFocus()

	modal := NewConfirmDialogue(msg, result, func() {
		panels.RemovePanel("confirm")
		// Reset focus changes
		app.SetBeforeFocusFunc(nil)
		app.SetFocus(oldFocus)
		app.SetBeforeFocusFunc(oldBeforeFocusFunc)
	})

	modal.SetTitle(title)
	panels.AddPanel("confirm", modal, false, true)

	// force focus on the modal
	app.SetBeforeFocusFunc(nil)
	app.SetFocus(modal.GetForm())
	// deny any focus change
	app.SetBeforeFocusFunc(func(p cview.Primitive) bool {
		if p == modal || p == modal.GetForm() {
			return true
		}
		x, y, w, h := p.GetRect()
		if modal.InRect(x, y) && modal.InRect(x+w, y+h) {
			return true
		}
		return false
	})

	return modal
}
func NewConfirmDialogue(msg string, result func(didConfirm bool), close func()) *cview.Modal {
	modal := cview.NewModal()
	modal.SetText(msg)
	modal.AddButtons([]string{"Cancel", "Confirm"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if result == nil {
			close()
			return
		}
		close()
		if buttonLabel == "Confirm" {
			result(true)
		} else {
			result(false)
		}
	})
	return modal
}
