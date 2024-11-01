package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"slices"
	"strings"
	"text/template"
)

// state changes / animations
// act_move(actor, from, to) & anim_move(actor, from, to)
// act_melee(actorOne, actorTwo) & anim_melee(actorOne, actorTwo)

// map window size in a console : 23x80

type MapLoader interface {
	LoadMap(mapName string) gridmap.MapLoadResult[*Actor, foundation.Item, Object]
}

func (g *GameState) giveAndTryEquipItem(actor *Actor, item foundation.Item) {
	actor.GetInventory().AddItem(item)
	if item.IsEquippable() {
		actor.GetEquipment().Equip(item)
	}
}

func (g *GameState) updateUIStatus() {
	g.ui.UpdateVisibleActors()
	g.ui.UpdateStats()
	g.ui.UpdateLogWindow()
	g.ui.UpdateInventory()
}

func (g *GameState) msg(message foundation.HiLiteString) {
	if !message.IsEmpty() {
		g.appendLogMessage(message)
		g.ui.UpdateLogWindow()
	}
}

func (g *GameState) appendLogMessage(message foundation.HiLiteString) {
	if len(g.logBuffer) == 0 {
		g.logBuffer = append(g.logBuffer, message)
		return
	}

	lastLogMessageIndex := len(g.logBuffer) - 1
	lastLogMessage := g.logBuffer[lastLogMessageIndex]

	if message.IsEqual(lastLogMessage) {
		lastLogMessage.Repetitions++
		g.logBuffer[lastLogMessageIndex] = lastLogMessage
		return
	}

	g.logBuffer = append(g.logBuffer, message)
}

func (g *GameState) removeItemFromInventory(holder *Actor, item foundation.Item) {
	inventory := holder.GetInventory()
	inventory.RemoveItem(item)
}

func (g *GameState) hasPaidWithCharge(user *Actor, item foundation.Item) bool {
	if item == nil { // no item = intrinsic effect
		return true
	}
	if item.Charges() == 0 {
		g.msg(foundation.Msg("The item is out of charges"))
		return false
	}
	if item.Charges() == 1 {
		if item.IsMultipleStacks() {
			item.RemoveStacks(1)
		} else {
			g.removeItemFromInventory(user, item)
		}
		return true
	}

	item.ConsumeCharge()
	if item.Charges() == 0 { // destroy
		g.removeItemFromInventory(user, item)
	}
	return true
}

func (g *GameState) actorKilled(causeOfDeath SourcedDamage, victim *Actor) {
	if victim == g.Player {
		g.msg(foundation.HiLite("You have died"))
		g.QueueActionAfterAnimation(func() {
			g.gameOver(causeOfDeath.String())
		})
		return
	}

	killedFlag := fmt.Sprintf("Killed(%s)", victim.GetInternalName())
	g.gameFlags.SetFlag(killedFlag)

	if causeOfDeath.IsActor() && causeOfDeath.Attacker == g.Player {
		killedByPlayerFlag := fmt.Sprintf("KilledByPlayer(%s)", victim.GetInternalName())
		g.gameFlags.SetFlag(killedByPlayerFlag)
		//g.awardXP(victim.GetXP(), fmt.Sprintf("for killing %s", victim.Name()))
		g.gameFlags.Increment("PlayerKillCount")
	}

	//g.dropInventory(victim)
	g.currentMap().SetActorToDowned(victim)

	delete(g.chatterCache, victim)
}

func (g *GameState) revealAll() {
	g.currentMap().SetAllExplored()
	g.showEverything = true
}
func (g *GameState) makeMapBloody(mapPos geometry.Point) {
	// we need a random integer between 5 and 15
	if !g.currentMap().Contains(mapPos) {
		return
	}
	var bloodColorFgInt int
	var bloodColorBgInt int
	if g.currentMap().IsTileWalkable(mapPos) { // blood on the floor is darker
		// range 10-15
		bloodColorFgInt = rand.Intn(6) + 10
		bloodColorBgInt = rand.Intn(6) + 10
	} else {
		// range 5-10
		bloodColorFgInt = rand.Intn(6) + 5
		bloodColorBgInt = rand.Intn(6) + 5
	}

	currentTileIcon := g.currentMap().GetTileIconAt(mapPos)
	g.currentMap().SetTileIcon(mapPos, currentTileIcon.WithBg(g.palette.Get(fmt.Sprintf("red_%d", bloodColorBgInt))).WithFg(g.palette.Get(fmt.Sprintf("red_%d", bloodColorFgInt))))
	return
}

func (g *GameState) makeMapBurned(mapPos geometry.Point) {
	// we need a random integer between 5 and 15
	if !g.currentMap().Contains(mapPos) || g.currentMap().IsTileWithFlagAt(mapPos, gridmap.TileFlagWater) {
		return
	}
	// black, dark_gray_7, brown_12
	var factor float64
	if g.currentMap().IsTileWalkable(mapPos) { // burn on the floor is darker
		factor = 0.2
	} else {
		factor = 0.5
	}

	currentTileIcon := g.currentMap().GetTileIconAt(mapPos)

	bgNew := multiplyWithRandomJitter(currentTileIcon.Bg, factor)
	fgNew := multiplyWithRandomJitter(currentTileIcon.Fg, factor)

	g.currentMap().SetTileIcon(mapPos, currentTileIcon.WithBg(bgNew).WithFg(fgNew))
	return
}

func multiplyWithRandomJitter(color color.RGBA, amount float64) color.RGBA {
	rAmount := amount + (rand.Float64() * 0.1) - 0.05
	gAmount := amount + (rand.Float64() * 0.1) - 0.05
	bAmount := amount + (rand.Float64() * 0.1) - 0.05
	newC := color
	newC.G = uint8(float64(color.G) * gAmount)
	newC.R = uint8(float64(color.R) * rAmount)
	newC.B = uint8(float64(color.B) * bAmount)
	return newC
}
func (g *GameState) spreadBloodAround(mapPos geometry.Point) {
	spreadArea := g.currentMap().GetDijkstraMap(mapPos, 2, g.currentMap().IsCurrentlyPassable)
	randomIndex := rand.Intn(len(spreadArea))
	currIndex := 0
	for pos, _ := range spreadArea {
		if currIndex == randomIndex {
			g.makeMapBloody(pos)
			break
		}
		currIndex++
	}

	rayHits := g.currentMap().RayCast(mapPos.ToCenteredPointF(), geometry.RandomDirection().ToPoint().ToCenteredPointF(), func(point geometry.Point) bool {
		return !g.currentMap().IsTileWalkable(point)
	})
	if rayHits.Distance <= 3 {
		wallPos := geometry.Point{X: int(rayHits.ColliderGridPosition[0]), Y: int(rayHits.ColliderGridPosition[1])}
		if !g.currentMap().IsTileWalkable(wallPos) {
			g.makeMapBloody(wallPos)
		}
	}
}
func (g *GameState) checkPlayerCanAct() {
	// idea
	// check if the player can act before giving back control to him
	// if he cannot act, eg, he is stunned and forced to do nothing,
	// then check the end condition for this status effect
	// if it's not reached, we want the UI to show a message about the situation
	// the player has to confirm it and then we can end the turn
	if !g.Player.HasFlag(foundation.FlagStun) && !g.Player.HasFlag(foundation.FlagHeld) {
		return
	}

	if g.Player.HasFlag(foundation.FlagStun) {
		result := g.Player.GetCharSheet().StatRoll(d100.Strength, 0)

		if result.Success {
			g.msg(foundation.Msg("You shake off the stun"))
			g.Player.GetFlags().Unset(foundation.FlagStun)
			return
		}
		g.Player.GetFlags().Increment(foundation.FlagStun)

		g.msg(foundation.Msg("You are stunned and cannot act"))

		// TODO: animate a small delay here?
		g.endPlayerTurn(g.Player.TimeNeededForActions())
	}
	if g.Player.HasFlag(foundation.FlagHeld) {
		result := g.Player.GetCharSheet().StatRoll(d100.Strength, 0)

		if result.Crit {
			g.msg(foundation.Msg("You break free from the hold"))
			g.Player.GetFlags().Unset(foundation.FlagHeld)
			return
		} else if result.Success {
			g.Player.GetFlags().Decrease(foundation.FlagHeld, 10)
			if !g.Player.HasFlag(foundation.FlagHeld) {
				g.msg(foundation.Msg("You break free from the hold"))
				return
			}
		}

		g.msg(foundation.Msg("You are held and cannot act"))

		// TODO: animate a small delay here?
		g.endPlayerTurn(g.Player.TimeNeededForActions())
	}
}

func (g *GameState) afterActorMovedOnMap(actor *Actor, oldPos geometry.Point) []foundation.Animation {
	newPos := actor.Position()
	isPlayer := actor == g.Player

	g.updateFoVAndDijkstraMap(actor)

	var animations []foundation.Animation

	if objectAt, hasObj := g.currentMap().TryGetObjectAt(newPos); hasObj {
		if objectAt.IsWalkable(actor) {
			triggeredEffectAnimations := objectAt.OnWalkOver(actor)
			animations = append(animations, triggeredEffectAnimations...)
		}
	}

	neighbors := actor.GetAllNeighbors()
	for _, neighbor := range neighbors {
		if objectAt, hasObj := g.currentMap().TryGetObjectAt(neighbor); hasObj {
			if objectAt.IsProximityTriggered() {
				triggeredEffectAnimations := objectAt.OnProximity(actor)
				animations = append(animations, triggeredEffectAnimations...)
			}

			if isPlayer && objectAt.IsHidden() && g.Player.GetCharSheet().GetStat(d100.Perception) > 1 {
				objectAt.SetHidden(false)
				g.msg(foundation.HiLite("You notice %s", objectAt.Name()))
			}
		}
		if itemAt, hasItem := g.currentMap().TryGetItemAt(neighbor); isPlayer && hasItem && itemAt.IsHidden() && g.Player.GetCharSheet().GetStat(d100.Perception) > 1 {
			itemAt.SetHidden(false)
			g.msg(foundation.HiLite("You notice %s", itemAt.Name()))
		}
	}

	currentZone := g.currentMap().FirstZoneAt(newPos)
	observers := g.getObservers(newPos)
	isSneaking := actor.HasFlag(foundation.FlagSneaking)

	if len(observers) > 0 {
		for _, observer := range observers {
			if observer.teamName == actor.teamName {
				continue
			}

			if !g.isDetectedByObserver(actor, observer) {
				continue
			}

			if observer.IsAggressive() {
				observer.FSM.SendEvent(NewEnemySightedEvent(actor))
				if isSneaking {
					g.msg(foundation.HiLite("You have been detected by %s", observer.Name()))
				}
			} else if observer.IsGuarding(currentZone) {
				g.reactToMinorCrime(observer, foundation.ChatterTrespassing)
			}
		}
	}

	if isPlayer {
		g.afterPlayerMoved(oldPos, false)
	}

	return animations
}

func (g *GameState) afterPlayerMoved(oldPos geometry.Point, wasMapTransition bool) {
	// explore the map
	// print "You see.." message
	if g.currentMap().IsItemAt(g.Player.Position()) && g.config.AutoPickup {
		g.PlayerPickupItem()
	}
	g.updatePlayerLightSource()

	g.msg(g.GetMapInfoForMovement(g.Player.Position()))

	g.Player.GetFlags().Unset(foundation.FlagConcentratedAiming)

	// automatic door opening/closing sfx handling
	if !wasMapTransition {
		if door, exists := g.TryGetDoorAt(g.Player.Position()); exists {
			if door.IsClosedButNotLocked() {
				door.PlayOpenSfx()
			}
		}
		if door, exists := g.TryGetDoorAt(oldPos); exists {
			if door.IsClosedButNotLocked() {
				door.PlayCloseSfx()
			}
		}
	}

	// check transition
	if !wasMapTransition {
		g.CheckTransition()
	}
}

func (g *GameState) updatePlayerLightSource() {
	if g.Player.GetInventory().HasLightSource() || g.Player.IsCyberWareActive(CyberWareLight) {
		wasOff := g.playerLightSource.MaxIntensity == 0
		wasAtPos := g.playerLightSource.Pos == g.Player.Position()
		g.playerLightSource.MaxIntensity = 1
		g.currentMap().MoveLightSource(g.playerLightSource, g.Player.Position())
		if wasAtPos && wasOff {
			g.currentMap().UpdateDynamicLights()
		}
	} else if g.playerLightSource.MaxIntensity > 0 {
		g.playerLightSource.MaxIntensity = 0
		g.currentMap().UpdateDynamicLights()
	}
}

func (g *GameState) openCyberwareMenu(vendor *Actor, onClose func()) {
	itemsForSale := vendor.OffersCyberWare
	if len(itemsForSale) == 0 {
		g.msg(foundation.Msg("Nothing for sale"))
		return
	}
	var tableRows []fxtools.TableRow
	var menuItems []foundation.MenuItem
	for _, item := range itemsForSale {
		tableRows = append(tableRows, fxtools.NewTableRow(item.GetItem1().String(), fmt.Sprintf("$%d", item.GetItem2())))
	}

	labelLines := fxtools.TableLayoutLastRight(tableRows)

	for index, item := range itemsForSale {
		label := labelLines[index]
		if item.GetItem2() > g.Player.GetGold() {
			label = fmt.Sprintf("%s%s[-]", textiles.RGBAToFgColorCode(g.palette.Get("dark_gray_3")), label)
		}
		menuItems = append(menuItems, foundation.MenuItem{
			Name: label,
			Action: func() {
				if item.GetItem2() > g.Player.GetGold() {
					g.msg(foundation.Msg("You cannot afford that"))
					g.openCyberwareMenu(vendor, onClose)
				} else {
					g.Player.RemoveGold(item.GetItem2())
					g.playerAddCyberware(item.GetItem1())
				}
			},
		})
	}

	title := fmt.Sprintf("Installed by %s", vendor.Name())
	g.ui.OpenMenuWithTitleAndClose(title, menuItems, onClose)
}

func (g *GameState) openVendorMenu(vendor *Actor, onClose func()) {
	itemsForSale := vendor.GetVendorInventory().Items()
	if len(itemsForSale) == 0 {
		g.msg(foundation.Msg("Nothing for sale"))
		return
	}
	title := fmt.Sprintf("Buy from %s", vendor.Name())
	g.ui.OpenVendorMenu(title, itemsForSale, g.buyItemFromVendor(vendor, onClose), g.inspectItem, onClose)
}

func (g *GameState) buyItemFromVendor(vendor *Actor, onClose func()) func(item foundation.Item, amount int, price int) {
	return func(item foundation.Item, amount int, price int) {
		player := g.Player
		if !player.HasGold(price) {
			g.msg(foundation.Msg("You cannot afford that"))
			g.openVendorMenu(vendor, onClose)
			return
		}
		if player.GetInventory().IsFull() {
			g.msg(foundation.Msg("You cannot carry more items"))
			g.openVendorMenu(vendor, onClose)
			return
		}

		if amount == 0 {
			g.openVendorMenu(vendor, onClose)
			return
		}

		stackTransfer(vendor.GetVendorInventory(), player.GetInventory(), item, amount)

		vendor.GetInventory().AddItem(player.RemoveGold(price))

		g.msg(foundation.HiLite("You bought %s for $%s", item.Name(), fmt.Sprint(price)))

		g.ui.PlayCue("world/pickup")

		g.openVendorMenu(vendor, onClose)
	}
}

func (g *GameState) openVendingMachineMenu(machine *Container) {
	itemsForSale := machine.ItemsFiltered(func(item foundation.Item) bool {
		return !item.IsGold()
	})
	if len(itemsForSale) == 0 {
		g.msg(foundation.Msg("Nothing for sale"))
		return
	}
	title := fmt.Sprintf("Buy from %s", machine.Name())
	g.ui.OpenVendorMenu(title, itemsForSale, g.buyItemFromVendingMachine(machine), g.inspectItem, nil)
}

func (g *GameState) buyItemFromVendingMachine(machine *Container) func(item foundation.Item, amount int, price int) {
	return func(item foundation.Item, amount int, price int) {
		player := g.Player
		if !player.HasGold(price) {
			g.msg(foundation.Msg("You cannot afford that"))
			g.openVendingMachineMenu(machine)
			return
		}
		if player.GetInventory().IsFull() {
			g.msg(foundation.Msg("You cannot carry more items"))
			g.openVendingMachineMenu(machine)
			return
		}
		if amount == 0 {
			g.openVendingMachineMenu(machine)
			return
		}

		stackTransfer(machine, player.GetInventory(), item, amount)

		machine.AddItem(player.RemoveGold(price))

		g.msg(foundation.HiLite("You bought %s for $%s", item.Name(), fmt.Sprint(price)))

		g.ui.PlayCue("world/pickup")

		g.openVendingMachineMenu(machine)
	}
}
func (g *GameState) dropInventory(victim *Actor) {
	goldAmount := victim.GetGold()
	if goldAmount > 0 {
		g.addItemToMap(g.NewGold(goldAmount), victim.Position())
	}
	for _, item := range victim.GetInventory().Items() {
		g.addItemToMap(item, victim.Position())
	}
}
func (g *GameState) fillTemplatedTextCustom(text string, vars map[string]string) string {
	parsedTemplate, err := template.New("text").Parse(text)
	if err != nil {
		panic(err)
	}
	vars["pcname"] = g.Player.Name()
	var filledText strings.Builder
	err = parsedTemplate.Execute(&filledText, vars)
	if err != nil {
		panic(err)
	}
	return filledText.String()
}
func (g *GameState) fillTemplatedText(text string) string {
	parsedTemplate, err := template.New("text").Parse(text)
	if err != nil {
		panic(err)
	}
	replaceValues := map[string]string{
		"pcname":         g.Player.Name(), // use as {{ .pcname }}
		"keys_move":      g.ui.GetKeybindingsAsString("move"),
		"keys_wait":      g.ui.GetKeybindingsAsString("wait"),
		"keys_look":      g.ui.GetKeybindingsAsString("look"),
		"keys_action":    g.ui.GetKeybindingsAsString("map_interaction"),
		"keys_inventory": g.ui.GetKeybindingsAsString("inventory"),
	}

	var filledText strings.Builder
	err = parsedTemplate.Execute(&filledText, replaceValues)
	if err != nil {
		panic(err)
	}
	return filledText.String()
}

func (g *GameState) getWeaponAttackAnim(attacker *Actor, targetPos geometry.Point, item *Weapon, attackMode AttackMode, bulletCount int) (foundation.Animation, bool) {
	weapon := item
	var attackAnim foundation.Animation
	isProjectile := false
	sourcePos := attacker.Position()
	switch weapon.GetDamageType() {
	case DamageTypePlasma:
		flightPath := g.GetFlightPath(sourcePos, targetPos)
		attackAnim, _ = g.ui.GetAnimProjectileWithLight('*', "green_2", flightPath, nil)
		isProjectile = true
	case DamageTypeExplosive:
		flightPath := g.GetFlightPath(sourcePos, targetPos)
		attackAnim, _ = g.ui.GetAnimProjectileWithLight('°', "white", flightPath, nil)
		isProjectile = true
	case DamageTypeLaser:
		flightPath := g.GetFlightPath(sourcePos, targetPos)
		attackAnim = g.ui.GetAnimLaser(flightPath, fxtools.NewColorFromRGBA(g.palette.Get("red_8")).MultiplyWithScalar(2), nil)
	default:
		attackAnim = g.ui.GetAnimMuzzleFlash(sourcePos, fxtools.NewColorFromRGBA(g.palette.Get("White")).MultiplyWithScalar(0.7), 2, bulletCount, nil)
	}

	attackAnim.SetAudioCue(weapon.GetFireAudioCue(attackMode.Mode))
	return attackAnim, isProjectile
}

func (g *GameState) GetFlightPath(sourcePos geometry.Point, targetPos geometry.Point) []geometry.Point {
	isValid := func(los []geometry.Point) bool {
		if len(los) < 2 {
			return false
		}
		startAtOrigin := los[0] == sourcePos
		return startAtOrigin
	}
	var paths [][]geometry.Point
	var conflicts []int

	conflictCount := func(fPath []geometry.Point) int {
		fovConflicts := 0
		for _, p := range fPath {
			if !g.Player.CanSee(p) {
				fovConflicts++
			}
		}
		return fovConflicts
	}

	flightPath := g.originalLine(sourcePos, targetPos)
	fovConflicts := conflictCount(flightPath)
	if isValid(flightPath) {
		if fovConflicts == 0 && flightPath[len(flightPath)-1] == targetPos {
			return flightPath
		}
		paths = append(paths, flightPath)
		conflicts = append(conflicts, fovConflicts)
	}

	flightPath = g.reversedLine(sourcePos, targetPos)
	fovConflicts = conflictCount(flightPath)
	if isValid(flightPath) {
		if fovConflicts == 0 && flightPath[len(flightPath)-1] == targetPos {
			return flightPath
		}
		paths = append(paths, flightPath)
		conflicts = append(conflicts, fovConflicts)
	}

	if len(paths) == 0 {
		return []geometry.Point{}
	}
	if len(paths) == 1 {
		return paths[0]
	}
	minIndex := slices.Index(conflicts, slices.Min(conflicts))
	otherIndex := 1 - minIndex
	if len(paths[minIndex]) < len(paths[otherIndex]) {
		return paths[otherIndex]
	} else {
		return paths[minIndex]
	}
}

func (g *GameState) originalLine(sourcePos geometry.Point, targetPos geometry.Point) []geometry.Point {
	flightPath := geometry.BresenhamLine(sourcePos, targetPos, func(x, y int) bool {
		if x == sourcePos.X && y == sourcePos.Y || x == targetPos.X && y == targetPos.Y {
			return true
		}
		return !g.IsSomethingBlockingTargetingAtLoc(geometry.Point{X: x, Y: y})
	})
	return flightPath
}

func (g *GameState) reversedLine(sourcePos geometry.Point, targetPos geometry.Point) []geometry.Point {
	flightPath := geometry.BresenhamLine(targetPos, sourcePos, func(x, y int) bool {
		if x == sourcePos.X && y == sourcePos.Y || x == targetPos.X && y == targetPos.Y {
			return true
		}
		return !g.IsSomethingBlockingTargetingAtLoc(geometry.Point{X: x, Y: y})
	})
	slices.Reverse(flightPath)
	return flightPath
}
