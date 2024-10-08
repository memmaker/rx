package game

import (
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"RogueUI/special"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"image/color"
	"math/rand"
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
	if item.Charges() > 0 {
		item.ConsumeCharge()
		if item.Charges() == 0 { // destroy
			g.removeItemFromInventory(user, item)
		}
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
func (g *GameState) updatePlayerFoVAndApplyExploration() {
	g.currentMap().UpdateFieldOfView(g.playerFoV, g.Player.Position(), g.visionRange)
	g.playerFoV.RemoveFromVisibles(func(p geometry.Point) bool {
		return g.currentMap().IsDarknessAt(g.gameTime.Time, p) && g.Player.Position() != p
	})

	for _, pos := range g.playerFoV.Visibles {
		g.currentMap().SetExplored(pos)
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
		result := g.Player.GetCharSheet().StatRoll(special.Strength, 0)

		if result.Success {
			g.msg(foundation.Msg("You shake off the stun"))
			g.Player.GetFlags().Unset(foundation.FlagStun)
			return
		}
		g.Player.GetFlags().Increment(foundation.FlagStun)

		g.msg(foundation.Msg("You are stunned and cannot act"))

		// TODO: animate a small delay here?
		g.endPlayerTurn(g.Player.timeNeededForActions())
	}
	if g.Player.HasFlag(foundation.FlagHeld) {
		result := g.Player.GetCharSheet().StatRoll(special.Strength, 0)

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
		g.endPlayerTurn(g.Player.timeNeededForActions())
	}
}

func (g *GameState) triggerTileEffectsAfterMovement(actor *Actor, oldPos, newPos geometry.Point) []foundation.Animation {
	isPlayer := actor == g.Player
	if g.currentMap().IsObjectAt(newPos) {
		var animations []foundation.Animation
		objectAt := g.currentMap().ObjectAt(newPos)
		if objectAt.IsTrap() {
			if isPlayer {
				playerMoveAnim := g.ui.GetAnimMove(g.Player, oldPos, newPos)
				playerMoveAnim.RequestMapUpdateOnFinish()
				animations = append(animations, playerMoveAnim)
			}
			triggeredEffectAnimations := objectAt.OnWalkOver(actor)
			animations = append(animations, triggeredEffectAnimations...)
		}
		return animations
	}
	return nil
}

func (g *GameState) buyItemFromVendor(item foundation.Item, price int) {
	player := g.Player
	if !player.HasGold(price) {
		g.msg(foundation.Msg("You cannot afford that"))
		return
	}
	if player.GetInventory().IsFull() {
		g.msg(foundation.Msg("You cannot carry more items"))
		return
	}
	player.RemoveGold(price)
	i := item.(*GenericItem)
	player.GetInventory().AddItem(i)
}

func (g *GameState) newLevelReached(level int) {
	//g.Player.AddCharacterPoints(10)
	g.msg(foundation.HiLite("You've been awarded 10 character points for reaching level %s", fmt.Sprint(level)))
}

func (g *GameState) checkTilesForHiddenObjects(tiles []geometry.Point) {
	var noticedSomething bool
	for _, tile := range tiles {
		if g.currentMap().IsObjectAt(tile) {
			object := g.currentMap().ObjectAt(tile)
			if object.IsHidden() {
				perceptionResult := g.Player.GetCharSheet().StatRoll(special.Perception, 0)
				if perceptionResult.Success {
					noticedSomething = true
				}
				if perceptionResult.Crit {
					object.SetHidden(false)
				}
			}
		}
	}

	if noticedSomething {
		g.msg(foundation.Msg("you feel like something is wrong with this room"))
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
	case special.DamageTypePlasma:
		flightPath := g.getFlightPath(sourcePos, targetPos)
		attackAnim, _ = g.ui.GetAnimProjectileWithLight('*', "green_2", flightPath, nil)
		isProjectile = true
	case special.DamageTypeExplosive:
		flightPath := g.getFlightPath(sourcePos, targetPos)
		attackAnim, _ = g.ui.GetAnimProjectileWithLight('°', "white", flightPath, nil)
		isProjectile = true
	case special.DamageTypeLaser:
		flightPath := g.getFlightPath(sourcePos, targetPos)
		attackAnim = g.ui.GetAnimLaser(flightPath, fxtools.NewColorFromRGBA(g.palette.Get("red_8")).MultiplyWithScalar(2), nil)
	default:
		attackAnim = g.ui.GetAnimMuzzleFlash(sourcePos, fxtools.NewColorFromRGBA(g.palette.Get("White")).MultiplyWithScalar(0.7), 2, bulletCount, nil)
	}

	attackAnim.SetAudioCue(weapon.GetFireAudioCue(attackMode.Mode))
	return attackAnim, isProjectile
}

func (g *GameState) getFlightPath(sourcePos geometry.Point, targetPos geometry.Point) []geometry.Point {
	flightPath := geometry.BresenhamLine(sourcePos, targetPos, func(x, y int) bool {
		if x == sourcePos.X && y == sourcePos.Y {
			return true
		}
		return !g.IsSomethingBlockingTargetingAtLoc(geometry.Point{X: x, Y: y})
	})
	return flightPath
}
