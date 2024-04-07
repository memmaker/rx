package game

import (
	"RogueUI/foundation"
	"RogueUI/recfile"
	"RogueUI/rpg"
	"RogueUI/util"
	"fmt"
	"math/rand"
	"path"
)

type DataDefinitions struct {
	Items    map[foundation.ItemCategory][]ItemDef
	Monsters []MonsterDef
}

func (d DataDefinitions) HasItems(category foundation.ItemCategory) bool {
	defines, ok := d.Items[category]
	return ok && len(defines) > 0
}

func GetDataDefinitions() DataDefinitions {
	dataDir := "data"

	readCloser := util.MustOpen(path.Join(dataDir, "armor.rec"))
	armorRecords := recfile.Read(readCloser)
	readCloser.Close()

	readCloser = util.MustOpen(path.Join(dataDir, "weapons.rec"))
	weaponRecords := recfile.Read(readCloser)
	readCloser.Close()

	readCloser = util.MustOpen(path.Join(dataDir, "scrolls.rec"))
	scrollRecords := recfile.Read(readCloser)
	readCloser.Close()

	readCloser = util.MustOpen(path.Join(dataDir, "potions.rec"))
	potionRecords := recfile.Read(readCloser)
	readCloser.Close()

	readCloser = util.MustOpen(path.Join(dataDir, "wands.rec"))
	wandRecords := recfile.Read(readCloser)
	readCloser.Close()

	readCloser = util.MustOpen(path.Join(dataDir, "rings.rec"))
	ringRecords := recfile.Read(readCloser)
	readCloser.Close()

	readCloser = util.MustOpen(path.Join(dataDir, "other.rec"))
	otherRecords := recfile.Read(readCloser)
	readCloser.Close()

	items := make(map[foundation.ItemCategory][]ItemDef)

	if len(weaponRecords) > 0 {
		items[foundation.ItemCategoryWeapons] = ItemDefsFromRecords(weaponRecords)
	}
	if len(armorRecords) > 0 {
		items[foundation.ItemCategoryArmor] = ItemDefsFromRecords(armorRecords)
	}
	if len(scrollRecords) > 0 {
		items[foundation.ItemCategoryScrolls] = ItemDefsFromRecords(scrollRecords)
	}
	if len(potionRecords) > 0 {
		items[foundation.ItemCategoryPotions] = ItemDefsFromRecords(potionRecords)
	}
	if len(wandRecords) > 0 {
		items[foundation.ItemCategoryWands] = ItemDefsFromRecords(wandRecords)
	}
	if len(ringRecords) > 0 {
		items[foundation.ItemCategoryRings] = ItemDefsFromRecords(ringRecords)
	}
	if len(otherRecords) > 0 {
		items[foundation.ItemCategoryOther] = ItemDefsFromRecords(otherRecords)
	}

	readCloser = util.MustOpen(path.Join(dataDir, "monsters.rec"))
	monsterRecords := recfile.Read(readCloser)
	readCloser.Close()

	var monsters []MonsterDef

	if len(monsterRecords) > 0 {
		monsters = MonsterDefsFromRecords(monsterRecords)
	}

	return DataDefinitions{
		Items:    items,
		Monsters: monsters,
	}
}

func (d DataDefinitions) PickItemForLevel(random *rand.Rand, level int) ItemDef {
	allCategories := []foundation.ItemCategory{
		foundation.ItemCategoryGold,
		foundation.ItemCategoryFood,
		foundation.ItemCategoryWeapons,
		foundation.ItemCategoryArmor,
		foundation.ItemCategoryScrolls,
		foundation.ItemCategoryPotions,
		foundation.ItemCategoryWands,
		foundation.ItemCategoryRings,
	}

	randomCategory := allCategories[random.Intn(len(allCategories))]

	if randomCategory == foundation.ItemCategoryGold {
		return ItemDef{
			Name:         "gold",
			InternalName: "gold",
			Category:     foundation.ItemCategoryGold,
			Charges:      rpg.NewDice(min(10, level+1), 10, 0),
		}
	}
	if randomCategory == foundation.ItemCategoryFood {
		return ItemDef{
			Name:         "food",
			InternalName: "food",
			Category:     foundation.ItemCategoryFood,
			Charges:      rpg.NewDice(0,0,10),
		}
	}
	items := d.Items[randomCategory]

	return items[random.Intn(len(items))]
}

func (d DataDefinitions) PickMonsterForLevel(random *rand.Rand, level int) MonsterDef {
	var filteredMonsters []MonsterDef

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

func (d DataDefinitions) RandomMonsterDef() MonsterDef {
	return d.Monsters[rand.Intn(len(d.Monsters))]
}

func (d DataDefinitions) GetScrollInternalNames() []string {
	return mapItemDefs(d.Items[foundation.ItemCategoryScrolls], func(def ItemDef) string {
		return def.InternalName
	})
}

func (d DataDefinitions) GetPotionInternalNames() []string {
	return mapItemDefs(d.Items[foundation.ItemCategoryPotions], func(def ItemDef) string {
		return def.InternalName
	})
}

func (d DataDefinitions) GetWandInternalNames() []string {
	return mapItemDefs(d.Items[foundation.ItemCategoryWands], func(def ItemDef) string {
		return def.InternalName
	})
}

func (d DataDefinitions) GetRingInternalNames() []string {
	return mapItemDefs(d.Items[foundation.ItemCategoryRings], func(def ItemDef) string {
		return def.InternalName
	})
}

func (d DataDefinitions) AlwaysIDOnUseInternalNames() []string {
	var names []string

	mapToInternalNames := func(def ItemDef) string {
		return def.InternalName
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
			if def.InternalName == name {
				return def
			}
		}
	}
	panic("Item not found: " + name)
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
