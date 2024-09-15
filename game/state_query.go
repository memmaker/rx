package game

import (
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
	attacker := g.Player
	weapon, hasWeapon := g.Player.GetEquipment().GetMainHandItem()
	if !hasWeapon {
		return 0
	}
	return g.getRangedChanceToHit(attacker, weapon, defender)
}

func (g *GameState) getRangedChanceToHit(attacker *Actor, equippedWeapon *Item, defender *Actor) int {
	var posInfos special.PosInfo
	posInfos.ObstacleCount = 0
	posInfos.Distance = g.currentMap().MoveDistance(attacker.Position(), defender.Position())
	posInfos.IlluminationPenalty = 0

	weaponSkill := equippedWeapon.GetWeapon().GetSkillUsed()

	defenderIsHelpless := defender.IsSleeping() || defender.IsStunned() || defender.IsKnockedDown()

	acModifier := 0

	if equippedWeapon != nil && equippedWeapon.IsRangedWeapon() && equippedWeapon.GetWeapon().NeedsAmmo() {
		ammoItem := equippedWeapon.GetWeapon().GetAmmo()
		if ammoItem != nil {
			ammo := ammoItem.GetAmmo()
			acModifier = ammo.ACModifier
		}
	}

	return special.RangedChanceToHit(posInfos, attacker.GetCharSheet(), weaponSkill, defender.GetCharSheet(), defenderIsHelpless, acModifier, special.Body)
}

func (g *GameState) GetItemInMainHand() (foundation.ItemForUI, bool) {
	return g.Player.GetEquipment().GetMainHandItem()
}

func (g *GameState) GetBodyPartsAndHitChances(targeted foundation.ActorForUI) []fxtools.Tuple3[special.BodyPart, bool, int] {
	victim := targeted.(*Actor)
	mainHandItem, hasMainHandItem := g.Player.GetEquipment().GetMainHandItem()
	if !hasMainHandItem {
		return victim.GetBodyPartsAndHitChances(g.Player.GetCharSheet().GetSkill(special.Unarmed))
	}
	baseChance := 0
	if mainHandItem.IsRangedWeapon() {
		baseChance = g.getRangedChanceToHit(g.Player, mainHandItem, victim)
	} else if mainHandItem.IsMeleeWeapon() {
		baseChance = g.getMeleeChanceToHit(g.Player, mainHandItem, victim)
	}
	return victim.GetBodyPartsAndHitChances(baseChance)
}

func (g *GameState) ItemAt(loc geometry.Point) foundation.ItemForUI {
	if g.currentMap().Contains(loc) && g.currentMap().IsItemAt(loc) {
		itemAt := g.currentMap().ItemAt(loc)
		return itemAt
	}
	return nil
}

func (g *GameState) ObjectAt(loc geometry.Point) foundation.ObjectForUI {
	if g.currentMap().Contains(loc) && g.currentMap().IsObjectAt(loc) {
		objectAt := g.currentMap().ObjectAt(loc)
		return objectAt
	}
	return nil
}

func (g *GameState) ActorAt(loc geometry.Point) foundation.ActorForUI {
	if g.currentMap().Contains(loc) && g.currentMap().IsActorAt(loc) {
		actorAt := g.currentMap().ActorAt(loc)
		return actorAt
	}
	return nil
}
func (g *GameState) DownedActorAt(loc geometry.Point) foundation.ActorForUI {
	if g.currentMap().Contains(loc) && g.currentMap().IsDownedActorAt(loc) {
		actorAt := g.currentMap().DownedActorAt(loc)
		return actorAt
	}
	return nil
}
func (g *GameState) GetCharacterSheet() string {

	actor := g.Player
	basicActorInfo := actor.GetDetailInfo()

	return basicActorInfo
}

func (g *GameState) GetPlayerPosition() geometry.Point {
	return g.Player.Position()
}

func (g *GameState) IsExplored(loc geometry.Point) bool {
	if !g.currentMap().Contains(loc) {
		return false
	}
	return g.currentMap().IsExplored(loc)
}

func (g *GameState) IsVisibleToPlayer(loc geometry.Point) bool {
	if !g.currentMap().Contains(loc) {
		return false
	}

	// special abilities
	canSeeFood := g.Player.HasFlag(special.FlagSeeFood)
	if g.IsFoodAt(loc) && canSeeFood {
		return true
	}

	canSeeMonsters := g.Player.HasFlag(special.FlagSeeMonsters)
	if g.currentMap().IsActorAt(loc) && canSeeMonsters {
		return true
	}

	canSeeTraps := g.Player.HasFlag(special.FlagSeeTraps)
	if g.currentMap().IsObjectAt(loc) && canSeeTraps {
		objectAt := g.currentMap().ObjectAt(loc)
		if objectAt.IsTrap() {
			return true
		}
	}

	if !g.currentMap().IsExplored(loc) {
		return false
	}
	isVisibleToPlayer := g.canPlayerSee(loc) || g.showEverything

	return isVisibleToPlayer
}

func (g *GameState) IsSomethingBlockingTargetingAtLoc(point geometry.Point) bool {
	currentMap := g.currentMap()
	if currentMap.IsActorAt(point) {
		return true
	}
	if currentMap.IsObjectAt(point) {
		object := currentMap.ObjectAt(point)
		if !object.IsPassableForProjectile() {
			return true
		}
	}
	if !currentMap.IsTileWalkable(point) && !currentMap.IsTransparent(point) {
		return true
	}
	return false
}

func (g *GameState) IsSomethingInterestingAtLoc(loc geometry.Point) bool {
	gridMap := g.currentMap()

	if !g.canPlayerSee(loc) {
		return false
	}

	if gridMap.IsActorAt(loc) {
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

func (g *GameState) GetVisibleActors() []foundation.ActorForUI {
	return actorsForUI(g.playerVisibleActorsByDistance())
}

func (g *GameState) GetVisibleEnemies() []*Actor {
	return fxtools.FilterSlice(g.playerVisibleActorsByDistance(), func(actor *Actor) bool {
		return actor.IsHostileTowards(g.Player)
	})
}

func (g *GameState) GetHudFlags() map[special.ActorFlag]int {
	flagSet := g.Player.GetFlags().UnderlyingCopy()
	equipFlags := g.Player.GetEquipment().GetAllFlags()
	for flag, _ := range equipFlags {
		flagSet[flag] = 1
	}
	return flagSet
}

func (g *GameState) GetHudStats() map[foundation.HudValue]int {
	uiStats := make(map[foundation.HudValue]int)
	if g.Player == nil {
		return uiStats
	}
	//g.Player.stats

	uiStats[foundation.HudTurnsTaken] = g.TurnsTaken()
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
	if !g.currentMap().Contains(pos) {
		return foundation.NoMsg()
	}
	if g.currentMap().IsActorAt(pos) && g.Player.Position() != pos {
		actor := g.currentMap().ActorAt(pos)
		return foundation.HiLite("You see %s", actor.LookInfo())
	}
	if g.currentMap().IsDownedActorAt(pos) && g.Player.Position() != pos {
		actor := g.currentMap().DownedActorAt(pos)
		return foundation.HiLite("You see %s", actor.LookInfo())
	}
	if g.currentMap().IsItemAt(pos) {
		item := g.currentMap().ItemAt(pos)
		return foundation.HiLite("You see %s", item.Name())
	}
	if g.currentMap().IsObjectAt(pos) {
		object := g.currentMap().ObjectAt(pos)
		return foundation.HiLite("You see %s", object.Name())
	}

	cell := g.currentMap().GetCell(pos)
	if isMovement {
		return foundation.NoMsg()
	}
	tileDesc := cell.TileType.DefinedDescription
	return foundation.HiLite("You see %s", tileDesc)
}

func (g *GameState) GetInventoryForUI() []foundation.ItemForUI {
	if g.Player == nil {
		return []foundation.ItemForUI{}
	}
	return itemStacksForUI(g.Player.GetInventory().StackedItemsWithFilter(func(item *Item) bool { return !item.IsAmmo() }))
}

func (g *GameState) MapAt(loc geometry.Point) textiles.TextIcon {
	if !g.currentMap().Contains(loc) {
		return textiles.TextIcon{}
	}
	mapCell := g.currentMap().GetCell(loc)
	return mapCell.TileType.Icon
}
func (g *GameState) TopEntityAt(mapPos geometry.Point) foundation.EntityType {
	if !g.currentMap().Contains(mapPos) {
		return foundation.EntityTypeOther
	}

	mapCell := g.currentMap().GetCell(mapPos)

	if mapCell.Actor != nil {
		actor := *mapCell.Actor
		if actor.IsVisible(g.Player.HasFlag(special.FlagSeeInvisible)) {
			return foundation.EntityTypeActor
		}
	}

	if mapCell.DownedActor != nil {
		actor := *mapCell.DownedActor
		if actor.IsVisible(g.Player.HasFlag(special.FlagSeeInvisible)) {
			return foundation.EntityTypeDownedActor
		}
	}

	if mapCell.Item != nil {
		return foundation.EntityTypeItem
	}

	if mapCell.Object != nil {
		object := *mapCell.Object
		if !object.IsHidden() {
			return foundation.EntityTypeObject
		}
	}

	return foundation.EntityTypeWorldTile
}

func (g *GameState) playerVisibleActorsByDistance() []*Actor {
	var enemies []*Actor
	if g.Player == nil || g.currentMap() == nil {
		return enemies
	}
	playerPos := g.Player.Position()
	for _, actor := range g.currentMap().Actors() {
		if actor == g.Player {
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
	for _, item := range g.currentMap().Items() {
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
	return g.playerFoV.Visible(pos)
}

func (g *GameState) GetFilteredInventory(filter func(item *Item) bool) []foundation.ItemForUI {
	items := g.Player.GetInventory().StackedItemsWithFilter(filter)
	return itemStacksForUI(items)

}

func (g *GameState) IsFoodAt(loc geometry.Point) bool {
	return g.currentMap().IsItemAt(loc) && g.currentMap().ItemAt(loc).IsFood()
}

func (g *GameState) IsBlockingRay(point geometry.Point) bool {
	return !g.currentMap().IsCurrentlyPassable(point)
}

func (g *GameState) SaveTimeNow(name string) {
	g.timeTracker[name] = g.gameTime
}
func (g *GameState) GetNamedTime(name string) PointInTime {
	if pointInTime, ok := g.timeTracker[name]; ok {
		return pointInTime
	}
	return PointInTime{}
}
func (g *GameState) IsTurnsAfter(name string, turns int) bool {
	pointInTime := g.GetNamedTime(name)
	return g.gameTime.TurnsSince(pointInTime) >= turns
}
func (g *GameState) IsMinutesAfter(name string, minutes int) bool {
	pointInTime := g.GetNamedTime(name)
	return g.gameTime.MinutesSince(pointInTime) >= minutes
}
func (g *GameState) IsHoursAfter(name string, hours int) bool {
	pointInTime := g.GetNamedTime(name)
	return g.gameTime.HoursSince(pointInTime) >= hours
}
func (g *GameState) IsDaysAfter(name string, days int) bool {
	pointInTime := g.GetNamedTime(name)
	return g.gameTime.DaysSince(pointInTime) >= days
}
