//go:build ebiten
package main

import "RogueUI/console"

func init() {
    graphicsModes["ebiten"] = console.NewEbitenUI()
}
