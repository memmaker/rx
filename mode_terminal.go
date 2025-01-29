//go:build ebiten

package main

import (
	"contractor/ui_console"
)

func init() {
	textModeLifecycles["terminal"] = ui_console.NewTerminalUI()
}
