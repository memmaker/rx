package foundation

func ShortCutFromIndex(itemIndex int) rune {
	shortcut := rune(97 + itemIndex)
	if itemIndex > 25 {
		shortcut = rune(39 + itemIndex)
	}
	return shortcut
}
