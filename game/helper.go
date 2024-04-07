package game

import (
	"RogueUI/foundation"
	"math/rand"
)

func itemsForUI(stack []*Item) []foundation.ItemForUI {
	displayStack := make([]foundation.ItemForUI, len(stack))
	for index, item := range stack {
		displayStack[index] = item
	}
	return displayStack
}

func itemStacksForUI(stack []*InventoryStack) []foundation.ItemForUI {
	displayStack := make([]foundation.ItemForUI, len(stack))
	for index, item := range stack {
		displayStack[index] = item
	}
	return displayStack
}

func actorsForUI(stack []*Actor) []foundation.ActorForUI {
	displayStack := make([]foundation.ActorForUI, len(stack))
	for index, actor := range stack {
		displayStack[index] = actor
	}
	return displayStack
}

// Adapted from: https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/misc.c#L485
func spread(nm int) int {
	return nm - nm/10 + rand.Intn(nm/5)
}

func confuseDuration() int {
	return spread(20)
}

func strengthDamageBonus(str int) int {
	add := 6
	if str < 8 {
		return str - 7
	}
	if str < 31 {
		add--
	}
	if str < 22 {
		add--
	}
	if str < 20 {
		add--
	}
	if str < 18 {
		add--
	}
	if str < 17 {
		add--
	}
	if str < 16 {
		add--
	}
	return add
}
