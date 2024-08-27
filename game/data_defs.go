package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/textiles"
	"image/color"
	"path"
)

func getExperienceTable() []int {
	eLevels := make([]int, 20)
	eLevels[0] = 10
	for i := 1; i < len(eLevels); i++ {
		eLevels[i] = eLevels[i-1] << 1
	}
	return eLevels
}

func getLevelForExperience(experience int) int {
	eLevels := getExperienceTable()
	var i int
	for i = 0; eLevels[i] != 0; i++ {
		if eLevels[i] > experience {
			break
		}
		i++
	}
	return i
}

func loadIconsForItems(dataDirectory string, colors textiles.ColorPalette) (map[foundation.ItemCategory]textiles.TextIcon, map[foundation.ItemCategory]color.RGBA) {
	convertItemCategories := func(r map[string]textiles.IconRecord) map[foundation.ItemCategory]textiles.TextIcon {
		convertMap := make(map[foundation.ItemCategory]textiles.TextIcon)
		for name, rec := range r {
			category := foundation.ItemCategoryFromString(name)
			icon := textiles.NewTextIconFromNamedColorChar(rec.Icon, colors)
			convertMap[category] = icon
		}
		return convertMap
	}

	itemCategoryFile := path.Join(dataDirectory, "iconsForItems.rec")
	itemCatRecords := fxtools.MustOpen(itemCategoryFile)
	iconsForItems := textiles.ReadIconRecordsIntoMap(itemCatRecords)

	return convertItemCategories(iconsForItems), loadInventoryColors(iconsForItems, colors)
}

func loadInventoryColors(records map[string]textiles.IconRecord, palette textiles.ColorPalette) map[foundation.ItemCategory]color.RGBA {
	inventoryItemColors := make(map[foundation.ItemCategory]color.RGBA)
	for name, rec := range records {
		if field, exists := rec.Meta.FindFieldIgnoreCase("InventoryColor"); exists {
			inventoryItemColors[foundation.ItemCategoryFromString(name)] = palette.Get(field.Value)
		}
	}
	return inventoryItemColors
}
