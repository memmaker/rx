package game

import (
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"RogueUI/special"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"math/rand"
	"strings"
	"text/template"
)

// state changes / animations
// act_move(actor, from, to) & anim_move(actor, from, to)
// act_melee(actorOne, actorTwo) & anim_melee(actorOne, actorTwo)

// map window size in a console : 23x80

type MapLoader interface {
	LoadMap(mapName string) gridmap.MapLoadResult[*Actor, *Item, Object]
}

func (g *GameState) ApplyEffect(name string, args []string) {
	switch name {
	case "SetFlag":
		flagName := strings.Trim(args[0], "'\" ")
		g.gameFlags.SetFlag(flagName)
	case "ClearFlag":
		flagName := strings.Trim(args[0], "'\" ")
		g.gameFlags.ClearFlag(flagName)
	}
	return
}

func (g *GameState) giveAndTryEquipItem(actor *Actor, item *Item) {
	actor.GetInventory().AddItem(item)
	if item.IsEquippable() {
		actor.GetEquipment().Equip(item)
	}
}

func (g *GameState) updateUIStatus() {
	g.ui.UpdateVisibleEnemies()
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

func (g *GameState) removeItemFromInventory(holder *Actor, item *Item) {
	equipment := holder.GetEquipment()
	inventory := holder.GetInventory()
	if item.IsMissile() {
		nextInStack := inventory.RemoveAndGetNextInStack(item)
		if nextInStack != nil {
			equipment.Equip(nextInStack)
		}
	} else {
		inventory.RemoveItem(item)
	}
}

func (g *GameState) hasPaidWithCharge(user *Actor, item *Item) bool {
	if item == nil { // no item = intrinsic effect
		return true
	}
	if item.charges == 0 {
		g.msg(foundation.Msg("The item is out of charges"))
		return false
	}
	if item.charges > 0 {
		item.charges--
		if item.charges == 0 { // destroy
			g.removeItemFromInventory(user, item)
		}
	}
	return true
}

func (g *GameState) actorKilled(causeOfDeath SourcedDamage, victim *Actor) {
	if victim == g.Player {
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
		g.awardXP(victim.GetXP(), fmt.Sprintf("for killing %s", victim.Name()))
	}

	//g.dropInventory(victim)
	g.currentMap().SetActorToDowned(victim)
}

func (g *GameState) revealAll() {
	g.currentMap().SetAllExplored()
	g.showEverything = true
}
func (g *GameState) makeMapBloody(mapPos geometry.Point) {
	// we need a random integer between 5 and 15
	bloodColorFgInt := rand.Intn(11) + 5
	bloodColorBgInt := rand.Intn(11) + 5

	currentTileIcon := g.currentMap().GetTileIconAt(mapPos)
	g.currentMap().SetTileIcon(mapPos, currentTileIcon.WithBg(g.palette.Get(fmt.Sprintf("red_%d", bloodColorBgInt))).WithFg(g.palette.Get(fmt.Sprintf("red_%d", bloodColorFgInt))))
	return
}
func (g *GameState) spreadBloodAround(mapPos geometry.Point) {
	spreadArea := g.currentMap().GetDijkstraMap(mapPos, 2, g.currentMap().IsCurrentlyPassable)
	randomIndex := rand.Intn(len(spreadArea))
	currIndex := 0
	for pos, _ := range spreadArea {
		if currIndex == randomIndex {
			g.makeMapBloody(pos)
			return
		}
		currIndex++
	}
}
func (g *GameState) updatePlayerFoVAndApplyExploration() {
	g.currentMap().UpdateFieldOfView(g.playerFoV, g.Player.Position(), g.visionRange)
	g.playerFoV.RemoveFromVisibles(func(p geometry.Point) bool {
		return g.currentMap().IsDarknessAt(g.gameTime, p) && g.Player.Position() != p
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
	if !g.Player.HasFlag(special.FlagStun) && !g.Player.HasFlag(special.FlagHeld) {
		return
	}

	if g.Player.HasFlag(special.FlagStun) {
		result := g.Player.GetCharSheet().StatRoll(special.Strength, 0)

		if result.Success {
			g.msg(foundation.Msg("You shake off the stun"))
			g.Player.GetFlags().Unset(special.FlagStun)
			return
		}
		g.Player.GetFlags().Increment(special.FlagStun)

		g.msg(foundation.Msg("You are stunned and cannot act"))

		// TODO: animate a small delay here?
		g.endPlayerTurn(g.Player.timeNeededForActions())
	}
	if g.Player.HasFlag(special.FlagHeld) {
		result := g.Player.GetCharSheet().StatRoll(special.Strength, 0)

		if result.Crit {
			g.msg(foundation.Msg("You break free from the hold"))
			g.Player.GetFlags().Unset(special.FlagHeld)
			return
		} else if result.Success {
			g.Player.GetFlags().Decrease(special.FlagHeld, 10)
			if !g.Player.HasFlag(special.FlagHeld) {
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

func (g *GameState) buyItemFromVendor(item foundation.ItemForUI, price int) {
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
	i := item.(*Item)
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

func (g *GameState) startSprint(actor *Actor) {
	/*
		if actor.GetActionPoints() < 1 {
			g.msg(foundation.Msg("You are too tired to sprint"))
			return
		}
		actor.LooseActionPoints(1)

	*/
	haste(g, actor)
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

func (g *GameState) getWeaponAttackAnim(attacker *Actor, targetPos geometry.Point, item *Item, attackMode AttackMode, bulletCount int) (foundation.Animation, bool) {
	weapon := item.GetWeapon()
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
	default:
		attackAnim = g.ui.GetAnimMuzzleFlash(sourcePos, fxtools.NewColorFromRGBA(g.palette.Get("White")).MultiplyWithScalar(5), 4, bulletCount, nil)
	}

	attackAnim.SetAudioCue(weapon.GetFireAudioCue(attackMode.Mode))
	return attackAnim, isProjectile
}

func (g *GameState) getFlightPath(sourcePos geometry.Point, targetPos geometry.Point) []geometry.Point {
	flightPath := geometry.BresenhamLine(sourcePos, targetPos, func(x, y int) bool {
		if x == sourcePos.X && y == sourcePos.Y {
			return true
		}
		return g.currentMap().IsCurrentlyPassable(geometry.Point{X: x, Y: y})
	})
	return flightPath
}
