package game

import (
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"path"
	"strings"
)

type DataDefinitions struct {
	Items    map[foundation.ItemCategory][]ItemDef
	Monsters []ActorDef
}

func (d DataDefinitions) HasItems(category foundation.ItemCategory) bool {
	defines, ok := d.Items[category]
	return ok && len(defines) > 0
}

func GetDataDefinitions(rootDir string, palette textiles.ColorPalette) DataDefinitions {
	dataDir := path.Join(rootDir, "definitions")

	readCloser := fxtools.MustOpen(path.Join(dataDir, "armor.rec"))
	armorRecords := recfile.Read(readCloser)
	readCloser.Close()

	readCloser = fxtools.MustOpen(path.Join(dataDir, "weapons.rec"))
	weaponRecords := recfile.Read(readCloser)
	readCloser.Close()

	readCloser = fxtools.MustOpen(path.Join(dataDir, "food.rec"))
	foodRecords := recfile.Read(readCloser)
	readCloser.Close()

	readCloser = fxtools.MustOpen(path.Join(dataDir, "miscItems.rec"))
	miscRecords := recfile.Read(readCloser)
	readCloser.Close()

	items := make(map[foundation.ItemCategory][]ItemDef)

	if len(weaponRecords) > 0 {
		items[foundation.ItemCategoryWeapons] = ItemDefsFromRecords(weaponRecords)
	}
	if len(armorRecords) > 0 {
		items[foundation.ItemCategoryArmor] = ItemDefsFromRecords(armorRecords)
	}
	if len(miscRecords) > 0 {
		miscItems := ItemDefsFromRecords(miscRecords)
		for _, i := range miscItems {
			category := i.Category
			items[category] = append(items[category], i)
		}
	}
	if len(foodRecords) > 0 {
		items[foundation.ItemCategoryFood] = ItemDefsFromRecords(foodRecords)
	}

	readCloser = fxtools.MustOpen(path.Join(dataDir, "actors.rec"))
	monsterRecords := recfile.Read(readCloser)
	readCloser.Close()

	var monsters []ActorDef

	if len(monsterRecords) > 0 {
		monsters = ActorDefsFromRecords(monsterRecords, palette)
	}

	return DataDefinitions{
		Items:    items,
		Monsters: monsters,
	}
}

func (d DataDefinitions) PickItemForLevel(random *rand.Rand, level int) ItemDef {
	var allCategories []foundation.ItemCategory
	for category, defs := range d.Items {
		if len(defs) == 0 {
			continue
		}
		allCategories = append(allCategories, category)
	}
	randomCategory := allCategories[random.Intn(len(allCategories))]

	if randomCategory == foundation.ItemCategoryGold {
		return ItemDef{
			Description: "gold",
			Name:        "gold",
			Category:    foundation.ItemCategoryGold,
			Charges:     dice_curve.NewDice(min(10, level+1), 10, 0),
		}
	}
	items := filterItemDefs(d.Items[randomCategory], func(def ItemDef) bool {
		return !def.Tags.Contains(foundation.TagNoLoot)
	})

	return items[random.Intn(len(items))]
}

func filterItemDefs(defs []ItemDef, keep func(def ItemDef) bool) []ItemDef {
	var filtered []ItemDef
	for _, def := range defs {
		if keep(def) {
			filtered = append(filtered, def)
		}
	}
	return filtered

}

func (d DataDefinitions) PickMonsterForLevel(random *rand.Rand, level int) ActorDef {
	var filteredMonsters []ActorDef

	for _, monster := range d.Monsters {
		if monster.DungeonLevel <= level {
			filteredMonsters = append(filteredMonsters, monster)
		}
	}

	if len(filteredMonsters) == 0 {
		panic(fmt.Sprintf("No monsters found for level %d", level))
	}
	return filteredMonsters[random.Intn(len(filteredMonsters))]
}

func (d DataDefinitions) RandomMonsterDef() ActorDef {
	return d.Monsters[rand.Intn(len(d.Monsters))]
}

func (d DataDefinitions) GetScrollInternalNames() []string {
	return mapItemDefs(d.Items[foundation.ItemCategoryScrolls], func(def ItemDef) string {
		return def.Name
	})
}

func (d DataDefinitions) GetPotionInternalNames() []string {
	return mapItemDefs(d.Items[foundation.ItemCategoryPotions], func(def ItemDef) string {
		return def.Name
	})
}

func (d DataDefinitions) GetWandInternalNames() []string {
	return mapItemDefs(d.Items[foundation.ItemCategoryWands], func(def ItemDef) string {
		return def.Name
	})
}

func (d DataDefinitions) GetRingInternalNames() []string {
	return mapItemDefs(d.Items[foundation.ItemCategoryRings], func(def ItemDef) string {
		return def.Name
	})
}

func (d DataDefinitions) AlwaysIDOnUseInternalNames() []string {
	var names []string

	mapToInternalNames := func(def ItemDef) string {
		return def.Name
	}
	filter := func(def ItemDef) bool {
		return def.AlwaysIDOnUse
	}
	scrolls := mapAndFilterItemDefs(d.Items[foundation.ItemCategoryScrolls], filter, mapToInternalNames)

	potions := mapAndFilterItemDefs(d.Items[foundation.ItemCategoryPotions], filter, mapToInternalNames)

	wands := mapAndFilterItemDefs(d.Items[foundation.ItemCategoryWands], filter, mapToInternalNames)

	rings := mapAndFilterItemDefs(d.Items[foundation.ItemCategoryRings], filter, mapToInternalNames)

	names = append(names, scrolls...)
	names = append(names, potions...)
	names = append(names, wands...)
	names = append(names, rings...)

	return names
}

func (d DataDefinitions) GetItemDefByName(name string) ItemDef {
	for _, defs := range d.Items {
		for _, def := range defs {
			if def.Name == name {
				return def
			}
		}
	}
	panic("Item not found: " + name)
}

func (d DataDefinitions) GetMonsterByName(name string) ActorDef {
	for _, monster := range d.Monsters {
		if monster.Name == name {
			return monster
		}
	}
	panic("Monster not found: " + name)
}

func mapItemDefs(defs []ItemDef, mapper func(ItemDef) string) []string {
	var names []string
	for _, def := range defs {
		names = append(names, mapper(def))
	}
	return names
}

func mapAndFilterItemDefs(defs []ItemDef, keep func(ItemDef) bool, mapper func(ItemDef) string) []string {
	var names []string
	for _, def := range defs {
		if !keep(def) {
			continue
		}
		names = append(names, mapper(def))
	}
	return names
}

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

func loadIconsForObjects(dataDirectory string, colors textiles.ColorPalette) map[string]textiles.TextIcon {
	convertObjectCategories := func(r map[string]textiles.IconRecord) map[string]textiles.TextIcon {
		convertMap := make(map[string]textiles.TextIcon)
		for name, rec := range r {
			icon := textiles.NewTextIconFromNamedColorChar(rec.Icon, colors)
			convertMap[strings.ToLower(name)] = icon
		}
		return convertMap
	}

	objectTypeFile := path.Join(dataDirectory, "iconsForObjects.rec")
	iconsForObjects := textiles.ReadIconRecordsIntoMap(fxtools.MustOpen(objectTypeFile))

	return convertObjectCategories(iconsForObjects)
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
