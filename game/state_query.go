package game

import (
	"RogueUI/dungen"
	"RogueUI/foundation"
	"RogueUI/special"
	"cmp"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"slices"
)

func (g *GameState) GetRangedChanceToHitForUI(target foundation.ActorForUI) int {
	defender := target.(*Actor)
	return g.getRangedChanceToHit(defender)
}

func (g *GameState) getRangedChanceToHit(defender *Actor) int {
	var posInfos special.PosInfo
	posInfos.ObstacleCount = 0
	posInfos.Distance = g.gridMap.MoveDistance(g.Player.Position(), defender.Position())
	posInfos.IlluminationPenalty = 0
	equippedWeapon, hasWeapon := g.Player.GetEquipment().GetMainHandItem()
	if !hasWeapon || !equippedWeapon.IsRangedWeapon() {
		return 0
	}
	weaponSkill := equippedWeapon.GetWeapon().GetSkillUsed()
	return special.RangedChanceToHit(posInfos, g.Player.GetCharSheet(), weaponSkill, defender.GetCharSheet(), special.Body)
}

func (g *GameState) GetItemInMainHand() (foundation.ItemForUI, bool) {
	return g.Player.GetEquipment().GetMainHandItem()
}

func (g *GameState) GetBodyPartsAndHitChances(targeted foundation.ActorForUI) []fxtools.Tuple[string, int] {
	victim := targeted.(*Actor)
	attackerSkill, defenderSkill := 0, 0 // TODO
	return victim.GetBodyPartsAndHitChances(attackerSkill, defenderSkill)
}

func (g *GameState) GetRandomEnemyName() string {
	return g.dataDefinitions.RandomMonsterDef().Description
}

func (g *GameState) ItemAt(loc geometry.Point) foundation.ItemForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsItemAt(loc) {
		itemAt := g.gridMap.ItemAt(loc)
		return itemAt
	}
	return nil
}

func (g *GameState) ObjectAt(loc geometry.Point) foundation.ObjectForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsObjectAt(loc) {
		objectAt := g.gridMap.ObjectAt(loc)
		return objectAt
	}
	return nil
}

func (g *GameState) ActorAt(loc geometry.Point) foundation.ActorForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsActorAt(loc) {
		actorAt := g.gridMap.ActorAt(loc)
		return actorAt
	}
	return nil
}
func (g *GameState) DownedActorAt(loc geometry.Point) foundation.ActorForUI {
	if g.gridMap.Contains(loc) && g.gridMap.IsDownedActorAt(loc) {
		actorAt := g.gridMap.DownedActorAt(loc)
		return actorAt
	}
	return nil
}
func (g *GameState) GetCharacterSheet() []string {

	actor := g.Player
	basicActorInfo := actor.GetDetailInfo()

	return basicActorInfo
}

func (g *GameState) GetPlayerPosition() geometry.Point {
	return g.Player.Position()
}

func (g *GameState) IsLit(pos geometry.Point) bool {
	return g.gridMap.IsTileLit(pos)
}

func (g *GameState) IsExplored(loc geometry.Point) bool {
	if !g.gridMap.Contains(loc) {
		return false
	}
	return g.gridMap.IsExplored(loc)
}

func (g *GameState) IsVisibleToPlayer(loc geometry.Point) bool {
	if !g.gridMap.Contains(loc) {
		return false
	}

	// special abilities
	canSeeFood := g.Player.HasFlag(foundation.FlagSeeFood)
	if g.IsFoodAt(loc) && canSeeFood {
		return true
	}

	canSeeMonsters := g.Player.HasFlag(foundation.FlagSeeMonsters)
	if g.gridMap.IsActorAt(loc) && canSeeMonsters {
		return true
	}

	canSeeMagic := g.Player.HasFlag(foundation.FlagSeeMagic)
	if g.gridMap.IsItemAt(loc) && canSeeMagic && g.gridMap.ItemAt(loc).IsMagic() {
		return true
	}

	canSeeTraps := g.Player.HasFlag(foundation.FlagSeeTraps)
	if g.gridMap.IsObjectAt(loc) && canSeeTraps {
		objectAt := g.gridMap.ObjectAt(loc)
		if objectAt.IsTrap() {
			return true
		}
	}

	if !g.gridMap.IsExplored(loc) {
		return false
	}
	isVisibleToPlayer := g.canPlayerSee(loc) || g.showEverything

	return isVisibleToPlayer
}

func (g *GameState) IsSomethingBlockingTargetingAtLoc(point geometry.Point) bool {
	return !g.gridMap.IsCurrentlyPassable(point)
}

func (g *GameState) IsSomethingInterestingAtLoc(loc geometry.Point) bool {
	gridMap := g.gridMap

	if g.Player.Position() == loc {
		return true
	}

	if gridMap.IsActorAt(loc) {
		return true
	}

	if gridMap.IsItemAt(loc) {
		return true
	}

	return false
}

func (g *GameState) IsEquipped(items foundation.ItemForUI) bool {
	itemStack, isItem := items.(*InventoryStack)
	if !isItem {
		return false
	}
	return g.Player.GetEquipment().IsEquipped(itemStack.First())
}

func (g *GameState) GetVisibleEnemies() []foundation.ActorForUI {
	return actorsForUI(g.playerVisibleEnemiesByDistance())
}

func (g *GameState) GetHudFlags() map[foundation.ActorFlag]int {
	flagSet := g.Player.GetFlags().UnderlyingCopy()
	equipFlags := g.Player.GetEquipment().GetAllFlags()
	for flag, _ := range equipFlags {
		flagSet[flag] = 1
	}
	return flagSet
}

func (g *GameState) GetHudStats() map[foundation.HudValue]int {
	uiStats := make(map[foundation.HudValue]int)
	//g.Player.stats

	uiStats[foundation.HudTurnsTaken] = g.TurnsTaken
	uiStats[foundation.HudGold] = g.Player.GetGold()

	uiStats[foundation.HudHitPoints] = max(0, g.Player.GetHitPoints())
	uiStats[foundation.HudHitPointsMax] = g.Player.GetHitPointsMax()

	uiStats[foundation.HudFatiguePoints] = g.Player.GetCharSheet().GetActionPoints()
	uiStats[foundation.HudFatiguePointsMax] = g.Player.GetCharSheet().GetActionPointsMax()

	uiStats[foundation.HudDamageResistance] = g.Player.GetDamageResistance()

	return uiStats
}

func (g *GameState) GetLog() []foundation.HiLiteString {
	return g.logBuffer
}
func (g *GameState) GetMapInfo(pos geometry.Point) foundation.HiLiteString {
	if g.canPlayerSee(pos) {
		return g.QueryMap(pos, false)
	}
	return foundation.NoMsg()
}
func (g *GameState) GetMapInfoForMovement(pos geometry.Point) foundation.HiLiteString {
	return g.QueryMap(pos, true)
}
func (g *GameState) QueryMap(pos geometry.Point, isMovement bool) foundation.HiLiteString {
	if !g.gridMap.Contains(pos) {
		return foundation.NoMsg()
	}
	if g.gridMap.IsActorAt(pos) && g.Player.Position() != pos {
		actor := g.gridMap.ActorAt(pos)
		return foundation.HiLite("You see %s here", actor.Name())
	}
	if g.gridMap.IsDownedActorAt(pos) && g.Player.Position() != pos {
		actor := g.gridMap.DownedActorAt(pos)
		return foundation.HiLite("You see %s here", actor.Name())
	}
	if g.gridMap.IsItemAt(pos) {
		item := g.gridMap.ItemAt(pos)
		return foundation.HiLite("You see %s here", item.Name())
	}
	if g.gridMap.IsObjectAt(pos) {
		object := g.gridMap.ObjectAt(pos)
		return foundation.HiLite("You see %s here", object.Name())
	}

	cell := g.gridMap.GetCell(pos)
	if isMovement {
		return foundation.NoMsg()
	}
	tileDesc := cell.TileType.DefinedDescription
	return foundation.HiLite("You see %s here", tileDesc)
}

func (g *GameState) getPlayerRoom() *dungen.DungeonRoom {
	if g.dungeonLayout == nil {
		return nil
	}
	playerRoom := g.dungeonLayout.GetRoomAt(g.Player.Position())
	return playerRoom
}

func (g *GameState) GetInventory() []foundation.ItemForUI {
	return itemStacksForUI(g.Player.GetInventory().StackedItems())
}

func (g *GameState) MapAt(loc geometry.Point) textiles.TextIcon {
	if !g.gridMap.Contains(loc) {
		return textiles.TextIcon{}
	}
	mapCell := g.gridMap.GetCell(loc)
	return mapCell.TileType.Icon
}
func (g *GameState) TopEntityAt(mapPos geometry.Point) foundation.EntityType {
	if !g.gridMap.Contains(mapPos) {
		return foundation.EntityTypeOther
	}

	mapCell := g.gridMap.GetCell(mapPos)

	if mapCell.Actor != nil {
		actor := *mapCell.Actor
		if actor.IsDrawn(g.Player.HasFlag(foundation.FlagSeeInvisible)) {
			return foundation.EntityTypeActor
		}
	}

	if mapCell.DownedActor != nil {
		actor := *mapCell.DownedActor
		if actor.IsDrawn(g.Player.HasFlag(foundation.FlagSeeInvisible)) {
			return foundation.EntityTypeDownedActor
		}
	}

	if mapCell.Item != nil {
		return foundation.EntityTypeItem
	}

	if mapCell.Object != nil {
		object := *mapCell.Object
		if object.IsDrawn() {
			return foundation.EntityTypeObject
		}
	}

	return foundation.EntityTypeWorldTile
}

func (g *GameState) playerVisibleEnemiesByDistance() []*Actor {
	var enemies []*Actor
	if g.Player == nil || g.gridMap == nil {
		return enemies
	}
	playerPos := g.Player.Position()
	for _, actor := range g.gridMap.Actors() {
		if actor == g.Player || !actor.IsHostile() {
			continue
		}
		if g.canPlayerSee(actor.Position()) && g.couldPlayerSeeActor(actor) {
			enemies = append(enemies, actor)
		}
	}
	slices.SortStableFunc(enemies, func(i, j *Actor) int {
		distI := geometry.Distance(playerPos, i.Position())
		distJ := geometry.Distance(playerPos, j.Position())
		return cmp.Compare(distI, distJ)
	})
	return enemies
}
func (g *GameState) GetVisibleItems() []foundation.ItemForUI {
	return itemStacksForUI(StacksFromItems(g.playerVisibleItemsByDistance()))
}
func (g *GameState) playerVisibleItemsByDistance() []*Item {
	playerPos := g.Player.Position()
	var visibleItems []*Item
	for _, item := range g.gridMap.Items() {
		if g.canPlayerSee(item.Position()) {
			visibleItems = append(visibleItems, item)
		}
	}
	slices.SortStableFunc(visibleItems, func(i, j *Item) int {
		distI := geometry.Distance(playerPos, i.Position())
		distJ := geometry.Distance(playerPos, j.Position())
		return cmp.Compare(distI, distJ)
	})
	return visibleItems
}

func (g *GameState) canPlayerSee(pos geometry.Point) bool {
	playerRoom := g.getPlayerRoom()
	playerPos := g.Player.Position()
	if g.dungeonLayout == nil {
		return g.playerFoV.Visible(pos)
	}
	isAtDoor := g.dungeonLayout.IsDoorAt(playerPos)
	isInRoom := playerRoom != nil
	isRoomLit := playerRoom != nil && playerRoom.IsLit()
	if pos == playerPos {
		return true
	}
	chebyDist := geometry.DistanceChebyshev(playerPos, pos)
	nextToUs := chebyDist == 1
	if nextToUs { // we will always see something right next us, save invisibility checks, blindness, etc.
		return true
	}

	// Case 1 corridor - we only see stuff right next to us
	if !isInRoom {
		return false
	}

	// Case 2 Special Case doors, we can see inside a lit room
	actorInSameRoom := isInRoom && playerRoom.ContainsIncludingWalls(pos)
	if isAtDoor {
		if actorInSameRoom && isRoomLit {
			return true
		} else {
			return false
		}
	}

	// Case 3 in a room
	if !actorInSameRoom {
		return false
	}

	if isRoomLit {
		return true
	} else {
		return false
	}
}

func (g *GameState) GetFilteredInventory(filter func(item *Item) bool) []foundation.ItemForUI {
	items := g.Player.GetInventory().StackedItemsWithFilter(filter)
	return itemStacksForUI(items)

}

func (g *GameState) isInPlayerRoom(position geometry.Point) bool {
	if g.dungeonLayout == nil {
		return true
	}
	playerRoom := g.getPlayerRoom()
	if playerRoom == nil {
		return false
	}
	return playerRoom.ContainsIncludingWalls(position)
}

func (g *GameState) IsFoodAt(loc geometry.Point) bool {
	return g.gridMap.IsItemAt(loc) && g.gridMap.ItemAt(loc).IsFood()
}

func (g *GameState) IsBlockingRay(point geometry.Point) bool {
	return !g.gridMap.IsCurrentlyPassable(point)
}
