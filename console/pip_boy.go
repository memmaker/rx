package console

import (
	"RogueUI/cview"
	"github.com/gdamore/tcell/v2"
)

type PipBoy struct {
	*cview.Grid
	statusBar *cview.TextView
	tabs      *cview.TabbedPanels
	onClose   func()
}

func (b PipBoy) handleKey(event *tcell.EventKey) *tcell.EventKey {
	return event
}

func (b PipBoy) SetOnClose(onClose func()) {
	b.onClose = onClose
}

func NewPipBoy() *PipBoy {
	grid := cview.NewGrid()
	grid.SetColumns(0)
	grid.SetRows(0, 1)

	statTab := cview.NewTextView()
	statTab.SetText("here be stats")

	invTab := cview.NewList()
	invTab.AddItem(cview.NewListItem("here be item #1"))
	invTab.AddItem(cview.NewListItem("here be item #2"))

	mainTabs := cview.NewTabbedPanels()

	mainTabs.AddTab("STAT", "STAT", statTab)
	mainTabs.AddTab("INV", "INV", invTab)

	statusBar := cview.NewTextView()
	statusBar.SetText("here be status")

	grid.AddItem(mainTabs, 0, 0, 1, 1, 0, 0, true)
	grid.AddItem(statusBar, 1, 0, 1, 1, 0, 0, false)

	p := &PipBoy{
		Grid:      grid,
		statusBar: statusBar,
		tabs:      mainTabs,
	}

	//grid.SetDrawFunc(p.drawInside)
	//grid.SetBackgroundTransparent(true)
	grid.SetInputCapture(p.handleKey)

	return p
}
