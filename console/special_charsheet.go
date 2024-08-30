package console

import (
    "RogueUI/special"
    "fmt"
    "github.com/gdamore/tcell/v2"
    "github.com/memmaker/go/cview"
)

type CharsheetViewer struct {
    *cview.Box
    sheet              *special.CharSheet
    close              func()
    nameAgeSexBar      *cview.TextView
    derivedStatsWindow *cview.TextView
    descriptionWindow  *cview.TextView
    statList           *cview.List
    skillList          *cview.List
    charPointsDisplay  *cview.TextView
    traitsList         *cview.List
    buttonBar          *cview.Form
    charName           string
}

func NewCharsheetViewer(name string, sheet *special.CharSheet, close func()) *CharsheetViewer {
    c := &CharsheetViewer{
        Box:      cview.NewBox(),
        sheet:    sheet,
        close:    close,
        charName: name,
    }
    c.setupUI()
    return c
}

func (c *CharsheetViewer) Draw(screen tcell.Screen) {
    c.Box.Draw(screen)
    c.drawChildren(screen)
}

func (c *CharsheetViewer) setupUI() {
    grid := cview.NewGrid()
    grid.SetRows(1, 10, 8, 9, 1)
    grid.SetColumns(20, 20, 40)

    nameAgeSexBar := cview.NewTextView()
    nameAgeSexBar.SetBorder(false)

    derivedStatsWindow := cview.NewTextView()
    derivedStatsWindow.SetBorder(true)

    descriptionWindow := cview.NewTextView()
    descriptionWindow.SetBorder(false)

    statList := cview.NewList()
    statList.SetBorder(true)

    skillList := cview.NewList()
    skillList.SetBorder(true)

    charPointsDisplay := cview.NewTextView()
    charPointsDisplay.SetBorder(false)

    traitsList := cview.NewList()
    traitsList.SetBorder(true)

    buttonBar := cview.NewForm()
    buttonBar.SetBorder(false)
    buttonBar.AddButton("Done", func() {
        c.Close()
    })
    buttonBar.AddButton("Cancel", func() {
        c.Close()
    })

    grid.AddItem(nameAgeSexBar, 0, 0, 1, 2, 0, 0, false)
    grid.AddItem(statList, 1, 0, 2, 1, 0, 0, false)
    grid.AddItem(charPointsDisplay, 2, 0, 1, 1, 0, 0, false)
    grid.AddItem(traitsList, 3, 0, 1, 2, 0, 0, false)
    grid.AddItem(derivedStatsWindow, 1, 1, 2, 1, 0, 0, false)
    grid.AddItem(skillList, 0, 2, 2, 1, 0, 0, false)
    grid.AddItem(descriptionWindow, 2, 2, 1, 1, 0, 0, false)
    grid.AddItem(buttonBar, 4, 2, 1, 1, 0, 0, false)

    c.nameAgeSexBar = nameAgeSexBar
    c.derivedStatsWindow = derivedStatsWindow
    c.descriptionWindow = descriptionWindow
    c.statList = statList
    c.skillList = skillList
    c.charPointsDisplay = charPointsDisplay
    c.traitsList = traitsList
    c.buttonBar = buttonBar
}

func (c *CharsheetViewer) Close() {
    c.close()
}

func (c *CharsheetViewer) updateUIFromSheet() {
    c.nameAgeSexBar.SetText(c.charName)

    hpString := fmt.Sprintf("Hitpoints: %s", c.sheet.GetHitPointsString())
    c.derivedStatsWindow.SetText(hpString)

    c.descriptionWindow.SetText("")
    /*
       c.statList.AddItem()
       c.skillList.AddItem()

       c.charPointsDisplay.SetText(c.sheet.CharPointsDisplay)

       c.traitsList.SetItems(c.sheet.TraitsList)

    */
}

func (c *CharsheetViewer) drawChildren(screen tcell.Screen) {
    c.nameAgeSexBar.Draw(screen)
    c.derivedStatsWindow.Draw(screen)
    c.descriptionWindow.Draw(screen)
    c.statList.Draw(screen)
    c.skillList.Draw(screen)
    c.charPointsDisplay.Draw(screen)
    c.traitsList.Draw(screen)
    c.buttonBar.Draw(screen)
}
