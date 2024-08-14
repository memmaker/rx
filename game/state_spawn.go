package game

import (
	"RogueUI/dice_curve"
	"RogueUI/dungen"
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"RogueUI/special"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"math/rand"
	"strings"
)

func (g *GameState) spawnEntities(random *rand.Rand, level int, newMap *gridmap.GridMap[*Actor, *Item, Object], dungeon *dungen.DungeonMap) {

	playerRoom := dungeon.GetRoomAt(g.Player.Position())

	mustSpawnAmuletOfYendor := level == 26 && !g.Player.GetInventory().HasItemWithName("amulet_of_yendor")

	spawnItemsInRoom := func(room *dungen.DungeonRoom, itemCount int) {
		for i := 0; i < itemCount; i++ {
			spawnPos, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			itemDef := g.dataDefinitions.PickItemForLevel(random, level)
			item := NewItem(itemDef, g.iconForItem(itemDef.Category))

			newMap.AddItem(item, spawnPos)
		}
	}

	spawnMonstersInRoom := func(room *dungen.DungeonRoom, monsterCount int) {
		for i := 0; i < monsterCount; i++ {
			spawnPos, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			monsterDef := g.dataDefinitions.PickMonsterForLevel(random, level)
			monster := g.NewEnemyFromDef(monsterDef)
			if random.Intn(26) >= level {
				monster.GetFlags().Set(foundation.FlagSleep)
			}
			if monster.HasFlag(foundation.FlagWallCrawl) {
				walls := room.GetWalls()
				spawnPos = walls[random.Intn(len(walls))]
			}
			newMap.AddActor(monster, spawnPos)
		}
	}

	spawnObjectsInRoom := func(room *dungen.DungeonRoom, objectCount int) {
		for i := 0; i < objectCount; i++ {
			_, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			/*
				trapEffects := foundation.GetAllTrapCategories()
				randomEffect := trapEffects[random.Intn(len(trapEffects))]
				object := g.NewTrap(randomEffect)
			*/

			//newMap.AddObject(object, spawnPos)
		}
	}

	allRooms := dungeon.AllRooms()
	randomRoomOrder := random.Perm(len(allRooms))
	for _, roomIndex := range randomRoomOrder {
		room := allRooms[roomIndex]

		itemCount := random.Intn(3)
		spawnItemsInRoom(room, itemCount)

		if random.Intn(2) == 0 || itemCount > 0 {
			monsterCount := random.Intn(max(2, itemCount)) + 1
			spawnMonstersInRoom(room, monsterCount)
		}

		if level > 1 && room != playerRoom && random.Intn(4) < 4 {
			objectCount := random.Intn(3) + 1
			spawnObjectsInRoom(room, objectCount)
		}

		if mustSpawnAmuletOfYendor {
			spawnPos, exists := room.GetRandomAbsoluteFloorPositionWithFilter(random, newMap.IsEmptyNonSpecialFloor)
			if !exists {
				break
			}
			amulet := g.NewItemFromName("amulet_of_yendor")
			newMap.AddItem(amulet, spawnPos)
			mustSpawnAmuletOfYendor = false
		}
	}
}

func (g *GameState) spawnCrawlerInWall(monsterDef ActorDef) {
	playerRoom := g.getPlayerRoom()
	if playerRoom == nil {
		return
	}
	walls := playerRoom.GetWalls()
	spawnPos := walls[rand.Intn(len(walls))]
	newActor := g.NewEnemyFromDef(monsterDef)
	g.gridMap.ForceSpawnActorInWall(newActor, spawnPos)
}

func (g *GameState) NewObjectFromRecord(record recfile.Record, palette textiles.ColorPalette, newMap *gridmap.GridMap[*Actor, *Item, Object]) Object {
	objectType := record.FindValueForKeyIgnoreCase("category")
	switch strings.ToLower(objectType) {
	case "elevator":
		elevator := g.NewElevator(record, g.iconForObject)
		newMap.AddNamedLocation(elevator.GetIdentifier(), elevator.Position())
		return elevator
	case "unknowncontainer":
		return g.NewContainer(record, g.iconForObject)
	case "terminal":
		return g.NewTerminal(record, g.iconForObject)
	case "readable":
		return g.NewReadable(record, g.iconForObject)
	case "lockeddoor":
		fallthrough
	case "closeddoor":
		fallthrough
	case "brokendoor":
		fallthrough
	case "opendoor":
		return g.NewDoor(record, g.iconForObject)
	}
	return nil
}

func (g *GameState) iconForObject(objectType string) textiles.TextIcon {
	if icon, exists := g.iconsForObjects[strings.ToLower(objectType)]; exists {
		return icon
	}
	return textiles.TextIcon{}
}

func (g *GameState) iconForItem(itemCategory foundation.ItemCategory) textiles.TextIcon {
	if icon, exists := g.iconsForItems[itemCategory]; exists {
		return icon
	}
	return textiles.TextIcon{}
}
func (g *GameState) addItemToMap(item *Item, mapPos geometry.Point) {
	g.gridMap.AddItemWithDisplacement(item, mapPos)
}

func (g *GameState) NewEnemyFromDef(def ActorDef) *Actor {
	charSheet := special.NewCharSheet()

	for stat, statValue := range def.SpecialStats {
		charSheet.SetStat(stat, statValue)
	}

	for derivedStat, derivedStatValue := range def.DerivedStat {
		charSheet.SetDerivedStatAbsoluteValue(derivedStat, derivedStatValue)
	}

	for skill, skillValue := range def.SkillAdjustments {
		charSheet.SetSkillAdjustment(skill, skillValue)
	}

	charSheet.HealCompletely()

	actor := NewActor(def.Description, def.Icon, charSheet)
	actor.GetFlags().Init(def.Flags.UnderlyingCopy())
	actor.SetIntrinsicZapEffects(def.ZapEffects)
	actor.SetIntrinsicUseEffects(def.UseEffects)
	actor.SetInternalName(def.Name)

	actor.SetSizeModifier(def.SizeModifier)
	actor.SetRelationToPlayer(def.DefaultRelation)
	actor.SetPosition(def.Position)
	actor.SetDialogueFile(def.DialogueFile)

	for _, itemName := range def.Equipment {
		item := g.NewItemFromName(itemName)
		actor.GetInventory().Add(item)
	}

	return actor
}

func (g *GameState) NewGold(amount int) *Item {
	def := ItemDef{
		Description: "gold",
		Name:        "gold",
		Category:    foundation.ItemCategoryGold,
		Charges:     dice_curve.NewDice(1, 1, 1),
	}
	item := NewItem(def, g.iconForItem(def.Category))
	item.SetCharges(amount)
	return item
}
