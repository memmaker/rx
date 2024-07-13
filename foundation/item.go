package foundation

import (
	"RogueUI/geometry"
	"math/rand"
	"strings"
)

type ItemForUI interface {
	Name() string
	GetCategory() ItemCategory
	InventoryNameWithColors(lineColorCode string) string
	InventoryNameWithColorsAndShortcut(invItemColorCode string) string
	IsEquippable() bool
	IsUsableOrZappable() bool
	GetListInfo() string
	Shortcut() rune
	DisplayLength() int
	Position() geometry.Point
	LongNameWithColors(colorCode string) string
}

type ItemCategory int

func (c ItemCategory) String() string {
	switch c {
	case ItemCategoryGold:
		return "Gold"
	case ItemCategoryFood:
		return "Food"
	case ItemCategoryWeapons:
		return "Weapons"
	case ItemCategoryArmor:
		return "Armor"
	case ItemCategoryRings:
		return "Rings"
	case ItemCategoryScrolls:
		return "Scrolls"
	case ItemCategoryPotions:
		return "Potions"
	case ItemCategoryWands:
		return "Wands"
	case ItemCategoryAmulets:
		return "Amulets"
	case ItemCategoryAmmo:
		return "Ammo"
	case ItemCategoryBooks:
		return "Books"
	case ItemCategoryLockpicks:
		return "Lockpicks"
	case ItemCategoryConsumables:
		return "Consumables"
	case ItemCategoryKeys:
		return "Keys"
	case ItemCategoryOther:
		return "Other"
	}
	panic("Unknown item category")
}

const (
	ItemCategoryGold ItemCategory = iota
	ItemCategoryFood
	ItemCategoryWeapons
	ItemCategoryArmor
	ItemCategoryRings
	ItemCategoryScrolls
	ItemCategoryPotions
	ItemCategoryWands
	ItemCategoryAmulets
	ItemCategoryAmmo
	ItemCategoryBooks
	ItemCategoryConsumables
	ItemCategoryLockpicks
	ItemCategoryKeys
	ItemCategoryOther
)

func RandomItemCategory() ItemCategory {
	return ItemCategory(rand.Intn(int(ItemCategoryOther) + 1))
}
func ItemCategoryFromString(s string) ItemCategory {
	s = strings.TrimPrefix(strings.ToLower(s), "item")
	switch s {
	case "gold":
		return ItemCategoryGold
	case "food":
		return ItemCategoryFood
	case "weapons":
		return ItemCategoryWeapons
	case "armor":
		return ItemCategoryArmor
	case "rings":
		return ItemCategoryRings
	case "scrolls":
		return ItemCategoryScrolls
	case "potions":
		return ItemCategoryPotions
	case "wands":
		return ItemCategoryWands
	case "amulets":
		return ItemCategoryAmulets
	case "ammo":
		return ItemCategoryAmmo
	case "lockpicks":
		return ItemCategoryLockpicks
	case "books":
		return ItemCategoryBooks
	case "consumables":
		return ItemCategoryConsumables
	case "keys":
		return ItemCategoryKeys
	case "other":
		return ItemCategoryOther
	}
	panic("Unknown item category: " + s)
}
