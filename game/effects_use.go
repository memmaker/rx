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
		"create_monster":                 endTurn(true, noAnim(createMonster)),
		"light":                          endTurn(true, noAnim(light)),
		"drain_life":                     endTurn(true, drainLife),
		"heal":                           endTurn(true, heal),
		"extra_heal":                     endTurn(true, extraHeal),
		"raise_level":                    endTurn(true, noAnim(raiseLevel)),
		"uncloak":                        endTurn(true, uncloak),
		"vorpalize":                      endTurn(false, playerVorpalizeWeapon),
		"satiate_fully":                  endTurn(true, satiateFully),
	}
}

func uncloak(g *GameState, user *Actor) []foundation.Animation {
	user.GetFlags().Unset(foundation.FlagInvisible)
	user.GetFlags().Unset(foundation.FlagSleep)
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
	userRoom := g.dungeonLayout.GetRoomAt(user.Position())
	isInRoom := userRoom != nil
	var affectedActors []*Actor
	for _, actor := range g.gridMap.Actors() {
		if actor == user {
			continue
		}
		if isInRoom && userRoom.Contains(actor.Position()) {
			affectedActors = append(affectedActors, actor)
		} else if !isInRoom && geometry.DistanceChebyshev(user.Position(), actor.Position()) <= 1 {
			affectedActors = append(affectedActors, actor)
		}
	}
	if len(affectedActors) == 0 {
		g.msg(foundation.Msg("Nothing happens."))
		return nil
	}
	ballPos := user.Position()
	if isInRoom {
		ballPos = userRoom.GetCenter()
	}

	flyFromUserAnim, _ := g.ui.GetAnimProjectile('☼', "LightRed", user.Position(), ballPos, nil)

	damage := SourcedDamage{
		NameOfThing:     "drain life",
		Attacker:        user,
		IsObviousAttack: true,
		AttackMode:      special.TargetingModeFireSingle,
		DamageType:      special.DamageTypeNormal,
		DamageAmount:    damageDone,
	}
	userDamageAnim := g.damageActorWithFollowUp(damage, user, nil, []foundation.Animation{flyFromUserAnim})

	var enemyAnims []foundation.Animation

	damage = SourcedDamage{
		NameOfThing:     "drain life",
		Attacker:        user,
		IsObviousAttack: true,
		AttackMode:      special.TargetingModeFireSingle,
		DamageType:      special.DamageTypeRadiation,
		DamageAmount:    max(1, damageDone/len(affectedActors)),
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

func light(g *GameState, user *Actor) {
	room := g.getPlayerRoom()
	if room == nil {
		g.msg(foundation.Msg("Nothing happens."))
		return
	}

	roomTiles := room.GetAbsoluteRoomTiles()
	room.SetLit(true)
	g.gridMap.SetLitMulti(roomTiles)
	g.gridMap.SetListExplored(roomTiles, true)
}

func createMonster(g *GameState, user *Actor) {
	freePositions := g.gridMap.GetFilteredNeighbors(user.Position(), func(point geometry.Point) bool {
		return g.gridMap.CanPlaceActorHere(point)
	})

	if len(freePositions) == 0 {
		g.msg(foundation.Msg("You hear a distant roar."))
		return
	}

	monster := g.NewActorFromDef(g.dataDefinitions.RandomMonsterDef())
	g.msg(foundation.Msg("A monster appears!"))

	randomPos := freePositions[rand.Intn(len(freePositions))]

	g.gridMap.AddActor(monster, randomPos)

	appearAnim := g.ui.GetAnimAppearance(monster, randomPos, nil)

	g.ui.AddAnimations([]foundation.Animation{appearAnim})
}

func playerDetectFood(g *GameState, user *Actor) {
	g.Player.GetFlags().Set(foundation.FlagSeeFood)
	g.msg(foundation.Msg("You feel a sudden awareness of your surroundings"))
}

func playerDetectMagic(g *GameState, user *Actor) {
	g.Player.GetFlags().Set(foundation.FlagSeeMagic)
	g.msg(foundation.Msg("You feel a sudden awareness of your surroundings"))
}

func playerDetectMonsters(g *GameState, user *Actor) {
	tunsUntilUnsee := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(foundation.FlagSeeMonsters, tunsUntilUnsee)
	g.msg(foundation.Msg("You feel a sudden awareness of your surroundings"))
}

func playerDetectTraps(g *GameState, user *Actor) {
	tunsUntilUnsee := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(foundation.FlagSeeTraps, tunsUntilUnsee)
	g.msg(foundation.Msg("You feel a sudden awareness of your surroundings"))
}

func seeInvisible(g *GameState, user *Actor) {
	user.GetFlags().Set(foundation.FlagSeeInvisible)

	if g.Player != user {
		return
	}
	g.msg(foundation.Msg("You feel your sight sharpen"))
	tunsUntilUnsee := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(foundation.FlagSeeInvisible, tunsUntilUnsee)
}
func makeInvisible(g *GameState, user *Actor) {
	user.GetFlags().Set(foundation.FlagInvisible)

	if g.Player != user {
		g.msg(foundation.HiLite("%s vanishes", user.Name()))
		return
	}
	g.msg(foundation.Msg("You vanish"))
	turnsUntilVisible := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(foundation.FlagInvisible, turnsUntilVisible)
}
func haste(g *GameState, user *Actor) {
	if user.GetFlags().IsSet(foundation.FlagSlow) {
		user.GetFlags().Unset(foundation.FlagHaste)
		return
	}
	user.GetFlags().Set(foundation.FlagHaste)

	if g.Player != user {
		g.msg(foundation.HiLite("%s speeds up", user.Name()))
		return
	}

	g.msg(foundation.Msg("The world around you slows down"))

	tunsUntilUnhasted := rand.Intn(user.GetBasicSpeed()/2) + user.GetBasicSpeed()/2
	g.Player.GetFlags().Increase(foundation.FlagHaste, tunsUntilUnhasted)
}

func slow(g *GameState, user *Actor) {
	if user.GetFlags().IsSet(foundation.FlagHaste) {
		user.GetFlags().Unset(foundation.FlagHaste)
		return
	}

	user.GetFlags().Set(foundation.FlagSlow)

	if g.Player != user {
		g.msg(foundation.HiLite("%s slows down", user.Name()))
		return
	}
	g.msg(foundation.Msg("The world around you speeds up"))
	tunsUntilUnslowed := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(foundation.FlagSlow, tunsUntilUnslowed)
}

func cancel(g *GameState, user *Actor) {
	user.GetFlags().Set(foundation.FlagCancel)

	if g.Player != user {
		return
	}
	tunsUntilUncancelled := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(foundation.FlagCancel, tunsUntilUncancelled)
}

func blindness(g *GameState, user *Actor) {
	user.GetFlags().Set(foundation.FlagBlind)

	if g.Player != user {
		return
	}
	g.msg(foundation.Msg("You are blinded!"))
	tunsUntilUnhasted := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(foundation.FlagBlind, tunsUntilUnhasted)
}

func hallucination(g *GameState, user *Actor) {
	user.GetFlags().Set(foundation.FlagHallucinating)

	if g.Player != user {
		return
	}
	g.msg(foundation.Msg("You are hallucinating!"))
	turns := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(foundation.FlagHallucinating, turns)
}

func levitation(g *GameState, user *Actor) {
	user.GetFlags().Set(foundation.FlagFly)

	if g.Player != user {
		return
	}
	g.msg(foundation.Msg("You feel lighter"))
	turnsUntilEarthbound := rand.Intn(8) + 8
	g.Player.GetFlags().Increase(foundation.FlagFly, turnsUntilEarthbound)
}

func aggroMonsters(g *GameState, actor *Actor) []foundation.Animation {
	dMap := g.gridMap.GetDijkstraMap(actor.Position(), 1000, func(point geometry.Point) bool {
		return g.gridMap.Contains(point)
	})
	waveEffect := g.ui.GetAnimRadialAlert(actor.Position(), dMap, nil)

	for _, monster := range g.gridMap.Actors() {
		if monster == g.Player {
			continue
		}
		monster.GetFlags().Unset(foundation.FlagSleep)
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

	//playerInventory := g.Player.GetInventory()
	//playerEquipment := g.Player.GetEquipment()

	onSelected := func(item foundation.ItemForUI) {
		armorItem := item.(*InventoryStack).First()

		//wasEquipped := playerEquipment.IsEquipped(armorItem)
		//playerInventory.Remove(armorItem)
		//playerInventory.Add(armorItem)
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
func playerVorpalizeWeapon(g *GameState, actor *Actor) []foundation.Animation {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsWeapon() && !item.GetWeapon().IsVorpal()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any suitable weapons."))
		return nil
	}

	onWeaponSelected := func(item foundation.ItemForUI) {
		weaponItem := item.(*InventoryStack).First()

		// select monster
		defs := g.dataDefinitions.Monsters
		var menuActions []foundation.MenuItem
		for _, def := range defs {
			monsterDef := def
			menuActions = append(menuActions, foundation.MenuItem{
				Name: monsterDef.Description,
				Action: func() {
					weaponItem.GetWeapon().Vorpalize(monsterDef.Name)
					//playerInventory.Add(armorItem)
					//if wasEquipped {playerEquipment.Equip(armorItem)}
					g.msg(foundation.HiLite("Your %s gives off a flash of intense white light", weaponItem.Name()))

					g.ui.UpdateInventory()

					origin := g.Player.Position()
					animations := g.ui.GetAnimVorpalizeWeapon(origin, nil)

					g.ui.AddAnimations(animations)
					g.endPlayerTurn(g.Player.timeNeededForActions())
				},
				CloseMenus: true,
			})
		}
		g.ui.OpenMenu(menuActions)
	}

	g.ui.OpenInventoryForSelection(inventory, "Vorpalize which weapon?", onWeaponSelected)

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
		//playerInventory.Remove(armorItem)
		//weaponItem.GetWeapon().AddEnchantment()
		//playerInventory.Add(armorItem)
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
	targetPos := g.gridMap.RandomSpawnPosition()

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

	g.afterPlayerMoved()

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
		g.Player.GetFlags().Increase(foundation.FlagConfused, turnsUntilUnconfuse)
	} else {
		flags := target.GetFlags()
		flags.Set(foundation.FlagConfused)
	}

	confuseAnim := g.ui.GetAnimConfuse(target.Position(), nil)
	return OneAnimation(confuseAnim)
}

func confuseEnemyOnNextAttack(g *GameState, user *Actor) {
	user.GetFlags().Set(foundation.FlagCanConfuse)
	var msg foundation.HiLiteString
	if user == g.Player {
		msg = foundation.HiLite("%s start glowing red", "You")
	} else {
		msg = foundation.HiLite("%s starts glowing red", user.Name())
	}
	g.msg(msg)
}
func revealMap(g *GameState, user *Actor) []foundation.Animation {
	dMap := g.gridMap.GetDijkstraMap(user.Position(), 1000, func(point geometry.Point) bool {
		return g.gridMap.IsTileWalkable(point) || g.gridMap.HasWalkableNeighbor(point)
	})
	litFilter := func(pos geometry.Point) bool {
		return g.dungeonLayout.IsCorridor(pos) || g.dungeonLayout.IsDoorAt(pos) || (!g.gridMap.IsTileWalkable(pos) && g.gridMap.HasWalkableNeighbor(pos))
	}
	g.gridMap.SetAllExplored()
	g.gridMap.SetLitByFilter(litFilter)

	waveEffect := g.ui.GetAnimRadialReveal(user.Position(), dMap, nil)

	//g.QueueActionAfterAnimation(reveal)

	return []foundation.Animation{waveEffect}
}
func holdAllVisibleMonsters(g *GameState, user *Actor) []foundation.Animation {
	affectedMonsters := g.playerVisibleEnemiesByDistance()

	for _, actor := range affectedMonsters {
		if actor == g.Player {
			continue
		}
		actor.GetFlags().Set(foundation.FlagHeld)
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
	affectedMonsters := g.playerVisibleEnemiesByDistance()

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
	affectedMonsters := g.playerVisibleEnemiesByDistance()

	for _, actor := range affectedMonsters {
		if actor == g.Player {
			continue
		}
		actor.GetFlags().Set(foundation.FlagScared)
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
	g.Player.GetFlags().Unset(foundation.FlagCanConfuse)
}

// Adapted from: https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/potions.c#L47
// and
// https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/potions.c#L288C15-L288C31

// adapted from: https://github.com/memmaker/rogue-pc-modern-C/blob/582340fcaef32dd91595721efb2d5db41ff3cb05/src/daemons.c#L79
func (g *GameState) unconfusePlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(foundation.FlagConfused)
	g.msg(foundation.Msg("you feel less confused now"))
}

func (g *GameState) unseeMonstersPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(foundation.FlagSeeMonsters)
	g.msg(foundation.Msg("your senses return to normal"))
}

func (g *GameState) unhastePlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(foundation.FlagHaste)
	g.msg(foundation.Msg("The world around you speeds up"))
}

func (g *GameState) makeVisiblePlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(foundation.FlagInvisible)
	g.msg(foundation.Msg("You can see your hands again"))
}
func (g *GameState) unslowPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(foundation.FlagSlow)
	g.msg(foundation.Msg("The world around you slows down"))
}

func (g *GameState) unseeInvisiblePlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(foundation.FlagSeeInvisible)
	g.msg(foundation.Msg("Your sight returns to normal"))
}

func (g *GameState) unflyPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(foundation.FlagFly)
	g.msg(foundation.Msg("You feel gravity's pull"))
}

func (g *GameState) unblindPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(foundation.FlagBlind)
	g.msg(foundation.Msg("You can see again"))
}

func (g *GameState) uncancelPlayer() {
	player := g.Player
	flags := player.GetFlags()
	flags.Unset(foundation.FlagCancel)
	g.msg(foundation.Msg("You feel your powers return"))
}
