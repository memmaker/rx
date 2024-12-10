//go:build ebiten
package main

import "contractor/console"

func init() {
    graphicsModes["terminal"] = console.NewTerminalUI()
}
