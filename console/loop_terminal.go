//go:build terminal

package console

import (
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
)

func NewTerminalUI() UILifeCycler {
	return TerminalUI{}
}

type TerminalUI struct {}

func (u TerminalUI) StartGameLoop(application *cview.Application, onScreenReady func()) {
	/*
	cpuProfileFile, err := os.Create("cpu.prof")
	if err != nil {
		panic(err)
	}
	defer cpuProfileFile.Close()

	// start CPU profiling
	if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()
	 */

	// TEXT MODE
	application.QueueUpdate(func() {
		screen := application.GetScreen()
		tty, isTerm := screen.Tty()
		if isTerm && tty != nil {
			tty.Write([]byte{0x1B, 0x3E})
		}
		w, h := screen.Size()
		resize := tcell.NewEventResize(w, h)
		application.QueueEvent(resize)
		if onScreenReady != nil {
			onScreenReady()
		}
	})

	if err := application.Run(); err != nil {
		panic(err)
	}
}
func (u TerminalUI) QuitGame(application *cview.Application) {
	application.Stop()
}