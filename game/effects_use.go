package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

func GetAllUseEffects() map[string]func(g *GameState, user *Actor) (bool, []foundation.Animation) {
	return map[string]func(g *GameState, user *Actor) (endsTurnDirectly bool, animations []foundation.Animation){
		"phase_door":                     endTurn(true, phaseDoor),
		"confuse":                        endTurn(true, confuse),
		"haste":                          endTurn(true, noAnim(haste)),
		"blindness":                      endTurn(true, noAnim(blindness)),
		"hallucination":                  endTurn(true, noAnim(hallucination)),
		"levitation":                     endTurn(true, noAnim(levitation)),
		"see_invisible":                  endTurn(true, noAnim(seeInvisible)),
		"confuse_monster_on_next_attack": endTurn(true, noAnim(confuseEnemyOnNextAttack)),
		"reveal_map":                     endTurn(true, revealMap),
		"freeze_monsters_in_room":        endTurn(true, holdAllVisibleMonsters),
		"sleep_monsters_in_room":         endTurn(true, sleepAllVisibleMonsters),
		"scare_monsters_in_room":         endTurn(true, scareAllVisibleMonsters),
		"enchant_armor":                  endTurn(false, playerEnchantArmor),
		"enchant_weapon":                 endTurn(false, playerEnchantWeapon),
		"aggravate_monsters":             endTurn(true, aggroMonsters),
		"detect_food":                    endTurn(true, noAnim(playerDetectFood)),
		"detect_magic":                   endTurn(true, noAnim(playerDetectMagic)),
		"detect_monsters":                endTurn(true, noAnim(playerDetectMonsters)),
		"detect_traps":                   endTurn(true, noAnim(playerDetectTraps)),
		"drain_life":                     endTurn(true, drainLife),
		"heal":                           endTurn(true, heal),
		"extra_heal":                     endTurn(true, extraHeal),
		"show_time":                      endTurn(false, showTime),
		"raise_level":                    endTurn(true, noAnim(raiseLevel)),
		"uncloak":                        endTurn(true, uncloak),
		"satiate_fully":                  endTurn(true, satiateFully),
	}
}

func showTime(g *GameState, user *Actor) []foundation.Animation {
	if user != g.Player {
		return nil
	}
	g.ShowDateTime()
	return nil
}

func uncloak(g *GameState, user *Actor) []foundation.Animation {
	user.GetFlags().Unset(special.FlagInvisible)
	user.GetFlags().Unset(special.FlagSleep)
	uncloakAnim, _ := g.ui.GetAnimUncloakAtPosition(user, user.Position())
	return []foundation.Animation{uncloakAnim}
}

func raiseLevel(g *GameState, user *Actor) {
	g.msg(foundation.Msg("you suddenly feel much more skillful"))
	//g.Player.LevelUp()
}

func heal(g *GameState, actor *Actor) []foundation.Animation {
	amount := actor.GetHitPointsMax() / 2
	actor.Heal(amount)
	g.msg(foundation.Msg("you begin to feel better"))
	return nil
}

func extraHeal(g *GameState, actor *Actor) []foundation.Animation {
	amount := actor.GetHitPointsMax()
	actor.Heal(amount)
	g.msg(foundation.Msg("you begin to feel much better"))
	return nil
}

func satiateFully(g *GameState, actor *Actor) []foundation.Animation {
	actor.Satiate()
	g.msg(foundation.Msg("you don't feel hungry anymore"))
	return nil
}

func drainLife(g *GameState, user *Actor) []foundation.Animation {
	userHealth := user.GetHitPoints()
	damageDone := max(1, userHealth/2)
	var affectedActors []*Actor
	for _, actor := range g.currentMap().Actors() {
		if actor == user {
			continue
		}
		if geometry.DistanceChebyshev(user.Position(), actor.Position()) <= 1 {
			affectedActors = append(affectedActors, actor)
		}
	}
	if len(affectedActors) == 0 {
		g.msg(foundation.Msg("Nothing happens."))
		return nil
	}
	ballPos := user.Position()

	flyFromUserAnim, _ := g.ui.GetAnimProjectile('☼', "LightRed", user.Position(), ballPos, nil)

	damage := SourcedDamage{
		NameOfThing:     "drain life",
		Attacker:        user,
		IsObviousAttack: true,
		TargetingMode:   special.TargetingModeFireSingle,
		DamageType:      special.DamageTypeNormal,
		DamageAmount:    damageDone,
		BodyPart:        special.Body,
	}
	userDamageAnim := g.damageActorWithFollowUp(damage, user, nil, []foundation.Animation{flyFromUserAnim})

	var enemyAnims []foundation.Animation

	damage = SourcedDamage{
		NameOfThing:     "drain life",
		Attacker:        user,
		IsObviousAttack: true,
		TargetingMode:   special.TargetingModeFireSingle,
		DamageType:      special.DamageTypeRadiation,
		DamageAmount:    max(1, damageDone/len(affectedActors)),
		BodyPart:        special.Body,
	}
	for _, actor := range affectedActors {
		flyToEnemyAnim, _ := g.ui.GetAnimProjectile('☼', "LightRed", ballPos, actor.Position(), nil)
		damageAnims := g.damageActor(damage, actor)
		flyToEnemyAnim.SetFollowUp(damageAnims)
		enemyAnims = append(enemyAnims, flyToEnemyAnim)
	}

	flyFromUserAnim.SetFollowUp(enemyAnims)

	return userDamageAnim
}

func noAnim(h func(g *GameState, user *Actor)) func(*GameState, *Actor) []foundation.Animation {
	return func(g *GameState, user *Actor) []foundation.Animation {
		h(g, user)
		return nil
	}
}
func endTurn(endsTurnDirectly bool, h func(*GameState, *Actor) []foundation.Animation) func(*GameState, *Actor) (bool, []foundation.Animation) {
	return func(g *GameState, user *Actor) (bool, []foundation.Animation) {
		return endsTurnDirectly, h(g, user)
	}
}

func playerDetectFood(g *GameState, user *Actor) {
	g.Player.GetFlags().Set(special.FlagSeeFood)
	g.msg(foundation.Msg("You feel a sudden awareness of your surroundings"))
}

func playerDetectMagic(g *GameState, user *Actor) {
	g.Player.GetFlags().Set(special.FlagSeeMagic)
	g.msg(foundation.Msg("You feel a sudden awareness of your surroundings"))
}

func playerDetectMonsters(g *GameState, user *Actor) {
	tunsUntilUnsee := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(special.FlagSeeMonsters, tunsUntilUnsee)
	g.msg(foundation.Msg("You feel a sudden awareness of your surroundings"))
}

func playerDetectTraps(g *GameState, user *Actor) {
	tunsUntilUnsee := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(special.FlagSeeTraps, tunsUntilUnsee)
	g.msg(foundation.Msg("You feel a sudden awareness of your surroundings"))
}

func seeInvisible(g *GameState, user *Actor) {
	user.GetFlags().Set(special.FlagSeeInvisible)

	if g.Player != user {
		return
	}
	g.msg(foundation.Msg("You feel your sight sharpen"))
	tunsUntilUnsee := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(special.FlagSeeInvisible, tunsUntilUnsee)
}
func makeInvisible(g *GameState, user *Actor) {
	user.GetFlags().Set(special.FlagInvisible)

	if g.Player != user {
		g.msg(foundation.HiLite("%s vanishes", user.Name()))
		return
	}
	g.msg(foundation.Msg("You vanish"))
	turnsUntilVisible := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(special.FlagInvisible, turnsUntilVisible)
}
func haste(g *GameState, user *Actor) {
	if user.GetFlags().IsSet(special.FlagSlow) {
		user.GetFlags().Unset(special.FlagHaste)
		return
	}
	user.GetFlags().Set(special.FlagHaste)

	if g.Player != user {
		g.msg(foundation.HiLite("%s speeds up", user.Name()))
		return
	}

	g.msg(foundation.Msg("The world around you slows down"))

	tunsUntilUnhasted := rand.Intn(user.GetBasicSpeed()/2) + user.GetBasicSpeed()/2
	g.Player.GetFlags().Increase(special.FlagHaste, tunsUntilUnhasted)
}

func slow(g *GameState, user *Actor) {
	if user.GetFlags().IsSet(special.FlagHaste) {
		user.GetFlags().Unset(special.FlagHaste)
		return
	}

	user.GetFlags().Set(special.FlagSlow)

	if g.Player != user {
		g.msg(foundation.HiLite("%s slows down", user.Name()))
		return
	}
	g.msg(foundation.Msg("The world around you speeds up"))
	tunsUntilUnslowed := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(special.FlagSlow, tunsUntilUnslowed)
}

func cancel(g *GameState, user *Actor) {
	user.GetFlags().Set(special.FlagCancel)

	if g.Player != user {
		return
	}
	tunsUntilUncancelled := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(special.FlagCancel, tunsUntilUncancelled)
}

func blindness(g *GameState, user *Actor) {
	user.GetFlags().Set(special.FlagBlind)

	if g.Player != user {
		return
	}
	g.msg(foundation.Msg("You are blinded!"))
	tunsUntilUnhasted := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(special.FlagBlind, tunsUntilUnhasted)
}

func hallucination(g *GameState, user *Actor) {
	user.GetFlags().Set(special.FlagHallucinating)

	if g.Player != user {
		return
	}
	g.msg(foundation.Msg("You are hallucinating!"))
	turns := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(special.FlagHallucinating, turns)
}

func levitation(g *GameState, user *Actor) {
	user.GetFlags().Set(special.FlagFly)

	if g.Player != user {
		return
	}
	g.msg(foundation.Msg("You feel lighter"))
	turnsUntilEarthbound := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(special.FlagFly, turnsUntilEarthbound)
}

func aggroMonsters(g *GameState, actor *Actor) []foundation.Animation {
	dMap := g.currentMap().GetDijkstraMap(actor.Position(), 1000, func(point geometry.Point) bool {
		return g.currentMap().Contains(point)
	})
	waveEffect := g.ui.GetAnimRadialAlert(actor.Position(), dMap, nil)

	for _, monster := range g.currentMap().Actors() {
		if monster == g.Player {
			continue
		}
		monster.GetFlags().Unset(special.FlagSleep)
	}
	g.msg(foundation.Msg("You hear a loud noise"))
	return []foundation.Animation{waveEffect}
}

func playerEnchantArmor(g *GameState, actor *Actor) []foundation.Animation {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsArmor()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any armor to enchant."))
		return nil
	}

	//playerInventory := g.Player.GetInventoryForUI()
	//playerEquipment := g.Player.GetEquipment()

	onSelected := func(item foundation.ItemForUI) {
		armorItem := item.(*InventoryStack).First()

		//wasEquipped := playerEquipment.IsEquipped(armorItem)
		//playerInventory.RemoveItem(armorItem)
		//playerInventory.AddItem(armorItem)
		//if wasEquipped {playerEquipment.Equip(armorItem)}
		g.msg(foundation.HiLite("Your %s glows silver for a moment.", armorItem.Name()))

		g.ui.UpdateInventory()
		animation := g.ui.GetAnimEnchantArmor(g.Player, g.Player.Position(), nil)

		g.ui.AddAnimations([]foundation.Animation{animation})

		g.endPlayerTurn(g.Player.timeNeededForActions())

	}

	g.ui.OpenInventoryForSelection(inventory, "Enchant which armor?", onSelected)

	return nil
}

func playerEnchantWeapon(g *GameState, actor *Actor) []foundation.Animation {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsWeapon()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any weapons to enchant."))
		return nil
	}

	onSelected := func(item foundation.ItemForUI) {
		weaponItem := item.(*InventoryStack).First()

		//wasEquipped := playerEquipment.IsEquipped(armorItem)
		//playerInventory.RemoveItem(armorItem)
		//weaponItem.GetWeapon().AddEnchantment()
		//playerInventory.AddItem(armorItem)
		//if wasEquipped {playerEquipment.Equip(armorItem)}
		g.msg(foundation.HiLite("Your %s glows blue for a moment.", weaponItem.Name()))

		g.ui.UpdateInventory()

		animation := g.ui.GetAnimEnchantWeapon(g.Player, g.Player.Position(), nil)

		g.ui.AddAnimations([]foundation.Animation{animation})

		g.endPlayerTurn(g.Player.timeNeededForActions())
	}

	g.ui.OpenInventoryForSelection(inventory, "Enchant which weapon?", onSelected)

	return nil
}

func phaseDoor(g *GameState, user *Actor) []foundation.Animation {
	targetPos := g.currentMap().RandomSpawnPosition()

	teleportAnimation := teleportWithAnimation(g, user, targetPos)
	teleportAnimation.RequestMapUpdateOnFinish()

	if user != g.Player || rand.Intn(5) == 0 {
		animateConfuse := confuse(g, user)
		teleportAnimation.SetFollowUp(animateConfuse)
	}
	return OneAnimation(teleportAnimation)
}

func teleportWithAnimation(g *GameState, actor *Actor, targetPos geometry.Point) foundation.Animation {

	origin := actor.Position()

	if actor == g.Player {
		g.msg(foundation.Msg("You teleport"))
	} else {
		g.msg(foundation.HiLite("%s teleports", actor.Name()))
	}
	g.actorMove(actor, targetPos)

	g.afterPlayerMoved(origin, false)

	vanishAnim, _ := g.ui.GetAnimTeleport(actor, origin, targetPos, nil)

	return vanishAnim
}

func confuse(g *GameState, target *Actor) []foundation.Animation {

	if target == g.Player {
		g.msg(foundation.Msg("wait, what's going on? Huh? What? Who?"))
	} else {
		g.msg(foundation.HiLite("%s looks confused", target.Name()))
	}
	if target == g.Player {
		// Monsters have a chance to get unconfused when they take their turn
		// So this fuse is only used for tracking the time the player is confused.
		turnsUntilUnconfuse := rand.Intn(8) + confuseDuration()
		g.Player.GetFlags().Increase(special.FlagConfused, turnsUntilUnconfuse)
	} else {
		flags := target.GetFlags()
		flags.Set(special.FlagConfused)
	}

	confuseAnim := g.ui.GetAnimConfuse(target.Position(), nil)
	return OneAnimation(confuseAnim)
}

func confuseEnemyOnNextAttack(g *GameState, user *Actor) {
	user.GetFlags().Set(special.FlagCanConfuse)
	var msg foundation.HiLiteString
	if user == g.Player {
		msg = foundation.HiLite("%s start glowing red", "You")
	} else {
		msg = foundation.HiLite("%s starts glowing red", user.Name())
	}
	g.msg(msg)
}
func revealMap(g *GameState, user *Actor) []foundation.Animation {
	dMap := g.currentMap().GetDijkstraMap(user.Position(), 1000, func(point geometry.Point) bool {
		return g.currentMap().IsTileWalkable(point) || g.currentMap().HasWalkableNeighbor(point)
	})
	g.currentMap().SetAllExplored()

	waveEffect := g.ui.GetAnimRadialReveal(user.Position(), dMap, nil)

	//g.QueueActionAfterAnimation(reveal)

	return []foundation.Animation{waveEffect}
}
func holdAllVisibleMonsters(g *GameState, user *Actor) []foundation.Animation {
	affectedMonsters := g.playerVisibleActorsByDistance()

	for _, actor := range affectedMonsters {
		if actor == g.Player {
			continue
		}
		actor.GetFlags().Set(special.FlagHeld)
	}
	var animations []foundation.Animation
	for _, actor := range affectedMonsters {
		//originalActorIcon := actor.Icon()
		// cover up anim

		flightAnim, _ := g.ui.GetAnimProjectile('☼', "White", user.Position(), actor.Position(), nil)
		//coverAnim := g.ui.GetAnimCover(actor.Position(), originalActorIcon, dist, nil)
		animations = append(animations, flightAnim)
		//animations = append(animations, coverAnim)
	}

	return animations
}
func sleepAllVisibleMonsters(g *GameState, user *Actor) []foundation.Animation {
	affectedMonsters := g.playerVisibleActorsByDistance()

	for _, actor := range affectedMonsters {
		if actor == g.Player {
			continue
		}
		actor.SetSleeping()
	}
	var animations []foundation.Animation
	for _, actor := range affectedMonsters {
		//originalActorIcon := actor.Icon()
		// cover up anim

		flightAnim, _ := g.ui.GetAnimProjectile('Z', "Yellow", user.Position(), actor.Position(), nil)
		//coverAnim := g.ui.GetAnimCover(actor.Position(), originalActorIcon, dist, nil)
		animations = append(animations, flightAnim)
		//animations = append(animations, coverAnim)
	}

	return animations
}

func scareAllVisibleMonsters(g *GameState, user *Actor) []foundation.Animation {
	affectedMonsters := g.playerVisibleActorsByDistance()

	for _, actor := range affectedMonsters {
		if actor == g.Player {
			continue
		}
		actor.GetFlags().Set(special.FlagScared)
	}
	var animations []foundation.Animation
	for _, actor := range affectedMonsters {
		//originalActorIcon := actor.Icon()
		// cover up anim

		flightAnim, _ := g.ui.GetAnimProjectile('☼', "Red", user.Position(), actor.Position(), nil)
		//coverAnim := g.ui.GetAnimCover(actor.Position(), originalActorIcon, dist, nil)
		animations = append(animations, flightAnim)
		//animations = append(animations, coverAnim)
	}

	return animations
}

func (g *GameState) removePlayerCanConfuse() {
	g.Player.GetFlags().Unset(special.FlagCanConfuse)
}

// Adapted from: https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/potions.c#L47
// and
// https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/potions.c#L288C15-L288C31

// adapted from: https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/daemons.c#L79
func (g *GameState) unconfusePlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(special.FlagConfused)
	g.msg(foundation.Msg("you feel less confused now"))
}

func (g *GameState) unseeMonstersPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(special.FlagSeeMonsters)
	g.msg(foundation.Msg("your senses return to normal"))
}

func (g *GameState) unhastePlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(special.FlagHaste)
	g.msg(foundation.Msg("The world around you speeds up"))
}

func (g *GameState) makeVisiblePlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(special.FlagInvisible)
	g.msg(foundation.Msg("You can see your hands again"))
}
func (g *GameState) unslowPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(special.FlagSlow)
	g.msg(foundation.Msg("The world around you slows down"))
}

func (g *GameState) unseeInvisiblePlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(special.FlagSeeInvisible)
	g.msg(foundation.Msg("Your sight returns to normal"))
}

func (g *GameState) unflyPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(special.FlagFly)
	g.msg(foundation.Msg("You feel gravity's pull"))
}

func (g *GameState) unblindPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(special.FlagBlind)
	g.msg(foundation.Msg("You can see again"))
}

func (g *GameState) uncancelPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(special.FlagCancel)
	g.msg(foundation.Msg("You feel your powers return"))
}
