package console

import (
	"contractor/foundation"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"path/filepath"
)

// Main Menu

func (u *UI) showMainMenu() {
	var items []foundation.MenuItem
	if u.saveGamesExist() {
		loadGame := foundation.MenuItem{
			Name: "Continue",
			Action: func() {
				panelName := "loadMenu"
				loadMenu := u.openSimpleMenuNotClosable(panelName, chooseSubDirMenuItems(u.settings.SaveGameDir, func(savegameSubdir string) {
					u.pages.RemovePanel(panelName)
					u.mapOverlay.ClearAll()
					u.game.LoadGame(savegameSubdir)
					u.moveInGame()
				}))
				loadMenu.SetInputCapture(u.directionalWrapper(func(event *tcell.EventKey) *tcell.EventKey {
					if event.Key() == tcell.KeyEscape {
						u.pages.RemovePanel(panelName)
						_, prim := u.pages.GetFrontPanel()
						u.lockFocusToPrimitive(prim)
					}
					return event
				}))
				offsetVertically(loadMenu, 6)
			},
			CloseMenus: false,
		}
		items = append(items, loadGame)
	}

	newGame := foundation.MenuItem{
		Name:   "New Game",
		Action: u.newGame,
	}
	items = append(items, newGame)

	quitGame := foundation.MenuItem{
		Name:       "Quit Game",
		Action:     u.QuitGame,
		CloseMenus: true,
	}
	items = append(items, quitGame)

	panelName := "mainMenu"

	mainMenu, longestItem := u.createSimpleMenu(panelName, items)
	mainMenu.SetInputCapture(u.directionalWrapper(mainMenu.GetInputCapture()))

	u.makeCenteredModal(panelName, mainMenu, longestItem, len(items))

	// move 6 lines down
	offsetVertically(mainMenu, 6)
}

func (u *UI) newGame() {
	modeDialogue := OpenChoiceDialogue(u.application, u.pages, "Mode?", "Choose game mode", []string{"Save anytime", "Ironman"}, func(index int, choice string) {
		if index == 0 || index == 1 {
			if index == 1 {
				u.game.SetIronMan()
			}
			u.moveInGame()
		} else {
			u.pages.SendToFront("mainMenu")
			_, frontPanel := u.pages.GetFrontPanel()
			u.lockFocusToPrimitive(frontPanel)
		}
	})
	for i := 0; i < modeDialogue.GetForm().GetButtonCount(); i++ {
		button := modeDialogue.GetForm().GetButton(i)
		button.SetInputCapture(u.directionalWrapper(button.GetInputCapture()))
	}
}

// Pause/System Menu

func (u *UI) OpenSystemMenu() {
	if u.game.IsIronMan() {
		u.OpenMenu([]foundation.MenuItem{
			{
				Name: "Save & Quit",
				Action: func() {
					u.game.SaveGame(filepath.Join(u.settings.SaveGameDir, "iron_man"))
				},
				CloseMenus: true,
			},
			{
				Name: "Quit Game",
				Action: func() {
					u.AskForConfirmation("Really?", "Do you want to quit this game?", func(confirmed bool) {
						if confirmed {
							u.QuitGame()
						}
					})
				},
				CloseMenus: true,
			},
		})
		return
	}
	u.OpenMenu([]foundation.MenuItem{
		{
			Name:       "Save Game",
			Action:     u.SelectSaveName,
			CloseMenus: true,
		},
		{
			Name:       "Load Game",
			Action:     u.SelectLoadName,
			CloseMenus: true,
		},
		{
			Name: "Quit Game",
			Action: func() {
				u.AskForConfirmation("Really?", "Do you want to quit this game?", func(confirmed bool) {
					if confirmed {
						u.QuitGame()
					}
				})
			},
			CloseMenus: true,
		},
	})
}

func (u *UI) SelectSaveName() {
	u.ChooseSaveDir(u.settings.SaveGameDir, u.game.SaveGame)
}

func (u *UI) ChooseSaveDir(savegameBaseDirectory string, onSubDirConfirmed func(savegameSubdir string)) {
	menuItems := chooseSubDirMenuItems(savegameBaseDirectory, func(savegameSubdir string) {
		u.AskForConfirmation("Overwrite?", fmt.Sprintf("Overwrite savegame %s?", savegameSubdir), func(didConfirm bool) {
			if didConfirm {
				onSubDirConfirmed(savegameSubdir)
			}
		})
	})
	if len(menuItems) == 0 {
		u.AskForString("Enter new savegame name", "", func(entered string) {
			onSubDirConfirmed(filepath.Join(savegameBaseDirectory, entered))
		})
		return
	}
	newEntryItem := foundation.MenuItem{
		Name: "<New Savegame..>",
		Action: func() {
			u.AskForString("Enter new savegame name", "", func(entered string) {
				onSubDirConfirmed(filepath.Join(savegameBaseDirectory, entered))
			})
		},
		CloseMenus: true,
	}
	menuItems = append([]foundation.MenuItem{newEntryItem}, menuItems...)
	u.OpenMenu(menuItems)
}

func (u *UI) SelectLoadName() {
	u.ChooseLoadDir(u.settings.SaveGameDir, func(savegameSubdir string) {
		u.mapOverlay.ClearAll()
		u.game.LoadGame(savegameSubdir)
	})
}

func (u *UI) ChooseLoadDir(savegameBaseDirectory string, onSubDirConfirmed func(savegameSubdir string)) {
	menuItems := chooseSubDirMenuItems(savegameBaseDirectory, onSubDirConfirmed)
	if len(menuItems) == 0 {
		u.OpenTextWindow("No savegames found in the savegame directory.")
	} else {
		u.OpenMenu(menuItems)
	}
}

// Groundwork

func (u *UI) OpenMenuWithTitle(title string, actions []foundation.MenuItem) {
	menu := u.openSimpleMenu(actions, nil)
	menu.SetTitle(title)
}

func (u *UI) OpenMenuWithTitleAndClose(title string, actions []foundation.MenuItem, onClose func()) {
	menu := u.openSimpleMenu(actions, onClose)
	menu.SetTitle(title)
}

func (u *UI) OpenMenu(actions []foundation.MenuItem) {
	u.openSimpleMenu(actions, nil)
}

func (u *UI) openSimpleMenu(menuItems []foundation.MenuItem, onClose func()) *cview.List {
	panelName := "menu"
	list := u.openSimpleMenuNotClosable(panelName, menuItems)
	originalCapture := list.GetInputCapture()
	list.SetInputCapture(u.directionalWrapper(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			u.popPanel(panelName)
			if onClose != nil {
				onClose()
			}
			return nil
		}
		if originalCapture != nil {
			return originalCapture(event)
		}
		return event
	}))
	list.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		if action == cview.MouseRightClick {
			u.popPanel(panelName)
			if onClose != nil {
				onClose()
			}
			return action, nil
		}
		return action, event
	})
	return list
}

func (u *UI) openSimpleMenuNotClosable(panelName string, menuItems []foundation.MenuItem) *cview.List {
	list, longestItem := u.createSimpleMenu(panelName, menuItems)

	u.makeCenteredModal(panelName, list, longestItem, len(menuItems))

	return list
}

func (u *UI) createSimpleMenu(panelName string, menuItems []foundation.MenuItem) (*cview.List, int) {
	list := cview.NewList()
	u.applyListStyle(list)
	list.SetSelectedFunc(func(index int, listItem *cview.ListItem) {
		item := menuItems[index]
		if item.CloseMenus {
			u.popPanel(panelName)
		}
		item.Action()
	})

	longestItem := setListItemsFromMenuItems(list, menuItems)
	return list, longestItem
}
