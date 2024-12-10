package game

import (
	"contractor/d100"
	"contractor/foundation"
	"cmp"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"slices"
)

func (g *GameState) addOneWarPointAgainst(victim *Actor) {
	g.gameFlags.Increment("WarPoints")
	g.gameFlags.Increment(fmt.Sprintf("WarPoints(%s)", victim.GetTeam()))
	g.gameFlags.Increment(fmt.Sprintf("WarPoints(%s)", victim.GetInternalName()))
}

func (g *GameState) addOneSupportPointFor(supported *Actor) {
	g.gameFlags.Increment("SupportPoints")
	g.gameFlags.Increment(fmt.Sprintf("SupportPoints(%s)", supported.GetTeam()))
	g.gameFlags.Increment(fmt.Sprintf("SupportPoints(%s)", supported.GetInternalName()))
}

func (g *GameState) getIntimidateModifiers(actor *Actor, opponent *Actor, situationalMods d100.ModList) d100.ModList {
	modifiers := situationalMods

	if actor == g.Player { // these mods are only for the player
		globalWarPoints := g.gameFlags.Get("WarPoints")
		factionWarPoints := g.gameFlags.Get(fmt.Sprintf("WarPoints(%s)", actor.GetTeam()))
		charWarPoints := g.gameFlags.Get(fmt.Sprintf("WarPoints(%s)", opponent.GetInternalName()))
		if charWarPoints > 0 {
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    "Personal Notoriety",
				Modifier:  +20,
				IsPercent: true,
			})
		} else if factionWarPoints > 1 {
			bonus := min(10, factionWarPoints)
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    "Faction Notoriety",
				Modifier:  +bonus,
				IsPercent: true,
			})
		} else if globalWarPoints > 2 {
			bonus := min(5, globalWarPoints)
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    "Global Notoriety",
				Modifier:  +bonus,
				IsPercent: true,
			})
		}
	}

	pCool := actor.GetCharSheet().GetStat(d100.Cool)
	pStr := actor.GetCharSheet().GetStat(d100.Strength)
	pSpd := actor.GetBasicSpeed()

	oCool := opponent.GetCharSheet().GetStat(d100.Cool)
	oStr := opponent.GetCharSheet().GetStat(d100.Strength)
	oSpd := opponent.GetBasicSpeed()

	if pCool > oCool {
		bonus := (pCool - oCool) * 2
		modifiers = append(modifiers, d100.DefaultModifier{
			Source:    "Cooler",
			Modifier:  +bonus,
			IsPercent: true,
		})
	} else if pCool < oCool {
		malus := (oCool - pCool) * 2
		modifiers = append(modifiers, d100.DefaultModifier{
			Source:    "Less cool",
			Modifier:  -malus,
			IsPercent: true,
		})
	} else {
		if pStr > oStr {
			bonus := (pStr - oStr) * 2
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    "Stronger",
				Modifier:  +bonus,
				IsPercent: true,
			})
		} else if pStr < oStr {
			malus := (oStr - pStr) * 2
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    "Weaker",
				Modifier:  -malus,
				IsPercent: true,
			})
		}

		if pSpd > oSpd {
			bonus := (pSpd - oSpd) * 2
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    "Faster",
				Modifier:  +bonus,
				IsPercent: true,
			})
		} else if pSpd < oSpd {
			malus := (oSpd - pSpd) * 2
			modifiers = append(modifiers, d100.DefaultModifier{
				Source:    "Slower",
				Modifier:  -malus,
				IsPercent: true,
			})
		}
	}

	return modifiers
}

func (g *GameState) getPlayerRangedChanceToHitForUI(mods d100.CombatModifiers) func(target foundation.ActorForUI) foundation.AttackInfo {
	return func(target foundation.ActorForUI) foundation.AttackInfo {
		defender := target.(*Actor)
		attacker := g.Player
		weapon, hasWeapon := g.Player.GetInventory().GetEquippedWeapon()
		if !hasWeapon || !weapon.IsLoaded() {
			return foundation.RangedAttackInfo{}
		}

		weaponSkill := weapon.GetSkillUsed()
		baseSkill := attacker.GetCharSheet().GetSkill(weaponSkill)

		cth, modifiers := g.getRangedChanceToHit(attacker, weapon, defender, g.NewAmmo(weapon.GetLoadedAmmo().GetInternalName(), weapon.BulletCountForCurrentAttackMode()), d100.NoModifierList)

		newMods := mods.WithChanceToHit(modifiers)

		return foundation.RangedAttackInfo{
			SkillUsed:       weaponSkill,
			SkillBaseValue:  baseSkill,
			Mods:            newMods,
			CtH:             cth,
			DamageBaseValue: weapon.GetWeaponDamageForCurrentAttackMode(),
			Victim:          defender,
		}
	}
}

func (g *GameState) getThrownChanceToHitForUI(target foundation.ActorForUI) foundation.AttackInfo {
	defender := target.(*Actor)
	attacker := g.Player

	distance := int(geometry.Distance(attacker.Position(), defender.Position()))
	strength := attacker.GetCharSheet().GetStat(d100.Strength)

	perception := attacker.GetCharSheet().GetStat(d100.Perception)

	agility := attacker.GetCharSheet().GetStat(d100.Agility)

	baseChance := min(perception, agility) * 10

	maxDist := strength * 2

	var cthMods []d100.Modifier
	cth := baseChance
	if distance > 1 && distance > maxDist {
		delta := distance - maxDist
		cth = baseChance - (delta * 3)
		cthMods = append(cthMods, d100.DefaultModifier{
			Source:    "Distance",
			Modifier:  -(delta * 3),
			Order:     0,
			IsPercent: true,
		})
	} else if distance == 1 {
		cth = baseChance + 100
		cthMods = append(cthMods, d100.DefaultModifier{
			Source:    "Point Blank",
			Modifier:  +100,
			Order:     0,
			IsPercent: true,
		})
	}

	return foundation.RangedAttackInfo{
		SkillUsed:      d100.SkillForThrowing,
		SkillBaseValue: baseChance,
		Mods: d100.CombatModifiers{
			ChanceToHitMods: cthMods,
			DamageMods:      nil,
		},
		CtH:             cth,
		DamageBaseValue: fxtools.Interval{},
		Victim:          defender,
	}
}

func (g *GameState) GetItemInMainHand() (foundation.Item, bool) {
	item, b := g.Player.GetInventory().GetMainHandItem()
	if !b {
		return nil, false
	}
	return item.(foundation.Item), b
}

func (g *GameState) GetBodyPartsAndHitChances(targeted foundation.ActorForUI) []fxtools.Tuple3[d100.BodyPart, bool, int] {
	victim := targeted.(*Actor)
	mainHandItem, hasMainHandItem := g.Player.GetInventory().GetEquippedWeapon()
	if !hasMainHandItem {
		return victim.GetBodyPartsAndHitChances(g.Player.GetCharSheet().GetSkill(d100.SkillForUnarmed), true)
	}
	baseChance := 0
	isMelee := true
	if mainHandItem.IsRangedWeapon() && mainHandItem.IsLoaded() {
		isMelee = false
		baseChance, _ = g.getRangedChanceToHit(g.Player, mainHandItem, victim, g.NewAmmo(mainHandItem.GetLoadedAmmo().GetInternalName(), mainHandItem.BulletCountForCurrentAttackMode()), d100.NoModifierList)
	} else if mainHandItem.IsMeleeWeapon() {
		baseChance, _ = g.getMeleeChanceToHit(g.Player, mainHandItem, victim, d100.NoModifierList)
	}
	return victim.GetBodyPartsAndHitChances(baseChance, isMelee)
}

func (g *GameState) ItemAt(loc geometry.Point) foundation.Item {
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
	if g.showEverything {
		return true
	}
	// special abilities
	canSeeFood := g.Player.HasFlag(foundation.FlagSeeFood)
	if g.IsFoodAt(loc) && canSeeFood {
		return true
	}

	canSeeMonsters := g.Player.HasFlag(foundation.FlagSeeMonsters)
	if g.currentMap().IsActorAt(loc) && canSeeMonsters {
		return true
	}

	canSeeTraps := g.Player.HasFlag(foundation.FlagSeeTraps)
	if g.currentMap().IsObjectAt(loc) && canSeeTraps {
		objectAt := g.currentMap().ObjectAt(loc)
		if objectAt.IsTrap() {
			return true
		}
	}

	if !g.currentMap().IsExplored(loc) {
		return false
	}
	isVisibleToPlayer := g.canPlayerSee(loc)

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

func (g *GameState) IsEquipped(items foundation.Item) bool {
	return g.Player.GetInventory().IsEquipped(items)
}

func (g *GameState) GetVisibleActors() []foundation.ActorForUI {
	return actorsForUI(g.playerVisibleActorsByDistance())
}

func (g *GameState) GetVisibleEnemies() []*Actor {
	return fxtools.FilterSlice(g.playerVisibleActorsByDistance(), func(actor *Actor) bool {
		return actor.IsHostileTowards(g.Player)
	})
}

func (g *GameState) GetHudFlags() map[foundation.ActorFlag]int {
	flagSet := g.Player.GetFlags().UnderlyingCopy()
	equipFlags := g.Player.GetInventory().GetAllEquipmentFlags()
	for flag, _ := range equipFlags {
		flagSet[flag] = 1
	}
	if g.Player.IsOpenCarryWeapon() {
		flagSet[foundation.FlagOpenCarry] = 1
	}
	return flagSet
}

func (g *GameState) GetHudStats() foundation.HudValueMap {
	uiStats := make(foundation.HudValueMap)
	if g.Player == nil {
		return uiStats
	}
	//g.Player.stats

	uiStats[foundation.HudTurnsTaken] = g.TurnsTaken()
	uiStats[foundation.HudGold] = g.Player.GetGold()

	uiStats[foundation.HudHitPoints] = max(0, g.Player.GetHitPoints())
	uiStats[foundation.HudHitPointsMax] = g.Player.GetHitPointsMax()

	uiStats[foundation.HudActionPoints] = g.Player.GetCharSheet().GetActionPoints()
	uiStats[foundation.HudActionPointsMax] = g.Player.GetCharSheet().GetActionPointsMax()

	uiStats[foundation.HudArmorString] = g.Player.OutfitStyle().String()

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
	if g.currentMap().IsActorAt(pos) && !isMovement {
		actor := g.currentMap().ActorAt(pos)
		return foundation.HiLite(actor.LookInfo())
	}
	if g.currentMap().IsDownedActorAt(pos) && g.Player.Position() != pos {
		actor := g.currentMap().DownedActorAt(pos)
		return foundation.HiLite(actor.LookInfo())
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

func (g *GameState) GetInventoryForUI() []foundation.Item {
	if g.Player == nil {
		return []foundation.Item{}
	}
	return g.Player.GetInventory().StackedItemsWithFilter(func(item foundation.Item) bool { return !item.IsAmmo() })
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
		if actor == g.Player || actor.IsVisible(g.Player.HasFlag(foundation.FlagSeeInvisible)) {
			return foundation.EntityTypeActor
		}
	}

	if mapCell.DownedActor != nil {
		actor := *mapCell.DownedActor
		if actor.IsVisible(g.Player.HasFlag(foundation.FlagSeeInvisible)) {
			return foundation.EntityTypeDownedActor
		}
	}

	if mapCell.Item != nil {
		item := *mapCell.Item
		if !item.IsHidden() {
			return foundation.EntityTypeItem
		}
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
func (g *GameState) GetVisibleItems() []foundation.Item {
	return g.playerVisibleItemsByDistance()
}
func (g *GameState) playerVisibleItemsByDistance() []foundation.Item {
	playerPos := g.Player.Position()
	var visibleItems []foundation.Item
	for _, item := range g.currentMap().Items() {
		if g.canPlayerSee(item.Position()) {
			visibleItems = append(visibleItems, item)
		}
	}
	slices.SortStableFunc(visibleItems, func(i, j foundation.Item) int {
		distI := geometry.Distance(playerPos, i.Position())
		distJ := geometry.Distance(playerPos, j.Position())
		return cmp.Compare(distI, distJ)
	})
	return visibleItems
}

func (g *GameState) canPlayerSee(pos geometry.Point) bool {
	return g.Player.CanSee(pos)
}

func (g *GameState) GetFilteredInventory(filter func(item foundation.Item) bool) []foundation.Item {
	items := g.Player.GetInventory().StackedItemsWithFilter(filter)
	return items
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
