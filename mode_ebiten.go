//go:build ebiten

package main

import "contractor/ui_console"

func init() {
	textModeLifecycles["text_gl"] = ui_console.NewEbitenUI()
}
