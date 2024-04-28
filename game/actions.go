package game

import (
    "RogueUI/foundation"
    "RogueUI/geometry"
    "RogueUI/rpg"
    "fmt"
    "math/rand"
    "strings"
)

var NoModifiers []rpg.Modifier

func (g *GameState) Wait() {
	g.endPlayerTurn()
}

func (g *GameState) playerAttack(defender *Actor) {
	consequences := g.actorMeleeAttack(g.Player, NoModifiers, defender, NoModifiers)
	if !g.Player.HasFlag(foundation.FlagInvisible) {
		defender.GetFlags().Set(foundation.FlagAwareOfPlayer)
	}
	g.ui.AddAnimations(consequences)
	g.endPlayerTurn()
}

func (g *GameState) playerMove(newPos geometry.Point) {
	directConsequencesOfMove := g.actorMoveAnimated(g.Player, newPos)

	g.afterPlayerMoved()

	g.ui.AddAnimations(directConsequencesOfMove)

	g.endPlayerTurn()
}

func (g *GameState) startRangedAttackWithMissile(item *Item) {
	g.ui.SelectTarget(g.Player.Position(), func(targetPos geometry.Point) {
		g.actorRangedAttackWithMissile(g.Player, item, g.Player.Position(), targetPos)
	})
}

func (g *GameState) startAimItem(item *Item) {
	g.ui.SelectTarget(g.Player.Position(), func(targetPos geometry.Point) {
		g.identification.SetCurrentItemInUse(item.GetInternalName())
		g.playerZapItemAndEndTurn(item, targetPos)
	})
}
func (g *GameState) startAimZapEffect(zapEffectName string, payCost func()) {
	g.ui.SelectTarget(g.Player.Position(), func(targetPos geometry.Point) {
		if payCost != nil {
			payCost()
		}
		g.playerInvokeZapEffectAndEndTurn(zapEffectName, targetPos)
	})
}
func (g *GameState) PlayerApplyItem(uiItem foundation.ItemForUI) {
	itemStack, isItem := uiItem.(*InventoryStack)
	if !isItem {
		return
	}
	item := itemStack.First()
	g.playerUseOrZapItem(item)
}
func (g *GameState) playerUseOrZapItem(item *Item) {
	if item.IsUsable() {
		g.identification.SetCurrentItemInUse(item.GetInternalName())
		g.actorUseItem(g.Player, item)
	} else if item.IsZappable() {
		g.startAimItem(item)
	}
}
func (g *GameState) actorUseItem(user *Actor, item *Item) {
	useEffectName := item.GetUseEffectName()

	if useEffectName == "" {
		g.msg(foundation.Msg("You cannot use this item"))
		return
	}

	if !g.hasPaidWithCharge(user, item) {
		return
	}

	actionEndsTurn, consequencesOfEffect := g.actorInvokeUseEffect(user, useEffectName)

	g.ui.AddAnimations(consequencesOfEffect)

	if user == g.Player {
		if g.identification.CanBeIdentifiedByUsing(item.GetInternalName()) {
			g.identification.IdentifyItem(item.GetInternalName())
			g.ui.UpdateInventory()
		}
		if actionEndsTurn {
			g.endPlayerTurn()
		}
	}
}

func useEffectExists(effectName string) bool {
	_, exists := GetAllUseEffects()[effectName]
	return exists
}
func (g *GameState) actorInvokeUseEffect(user *Actor, useEffectName string) (endsTurn bool, animations []foundation.Animation) {
	if effect, exists := GetAllUseEffects()[useEffectName]; exists {
		return effect(g, user)
	}
	return false, nil
}
func (g *GameState) actorMoveAnimated(actor *Actor, newPos geometry.Point) []foundation.Animation {
	oldPos := actor.Position()
	var moveAnims []foundation.Animation
	if g.couldPlayerSeeActor(actor) && (g.canPlayerSee(newPos) || g.canPlayerSee(oldPos)) && actor != g.Player {
		move := g.ui.GetAnimMove(actor, oldPos, newPos)
		move.RequestMapUpdateOnFinish()
		moveAnims = append(moveAnims, move)
	}
	moveAnims = append(moveAnims, g.actorMove(actor, newPos)...)
	return moveAnims
}
func (g *GameState) couldPlayerSeeActor(actor *Actor) bool {
	if actor.HasFlag(foundation.FlagInvisible) && !g.Player.HasFlag(foundation.FlagSeeInvisible) {
		return false
	}

	return true
}
func (g *GameState) actorMove(actor *Actor, newPos geometry.Point) []foundation.Animation {
	oldPos := actor.Position()
	if oldPos == newPos {
		return nil
	}
	g.gridMap.MoveActor(actor, newPos)
	if actor.Position() == newPos {
		return g.triggerTileEffectsAfterMovement(actor, oldPos, newPos)
	}
	return nil
}

type StairsInLevel int

func (l StairsInLevel) AllowsUp() bool {
	return l == StairsUpOnly || l == StairsBoth
}
func (l StairsInLevel) AllowsDown() bool {
	return l == StairsDownOnly || l == StairsBoth
}

const (
	StairsNone StairsInLevel = iota
	StairsUpOnly
	StairsDownOnly
	StairsBoth
)

func (g *GameState) PlayerInteractWithMap() {
	pos := g.Player.Position()
	cell := g.gridMap.GetCell(pos)

	isDescending := cell.TileType.IsStairsDown()
	isAscending := cell.TileType.IsStairsUp()
	if !isDescending && !isAscending {
		g.msg(foundation.Msg("There are no stairs here"))
		return
	}
	if isDescending && isAscending {
		g.msg(foundation.Msg("There are both up and down stairs here."))
		return
	}

	if isDescending {
		g.PlayerTryDescend()
	} else if isAscending {
		g.PlayerTryAscend()
	}
}

func (g *GameState) PlayerTryDescend() {
	pos := g.Player.Position()
	cell := g.gridMap.GetCell(pos)
	stairs := StairsBoth
	if cell.TileType.IsStairsDown() {
		g.descendWithStairs(stairs)
	}
}

func (g *GameState) PlayerTryAscend() {
	pos := g.Player.Position()
	cell := g.gridMap.GetCell(pos)
	stairs := StairsBoth
	if cell.TileType.IsStairsUp() {
		if g.currentDungeonLevel == 1 {
			if g.Player.GetInventory().HasItemWithName("amulet_of_yendor") {
				g.gameWon()
			} else {
				g.msg(foundation.Msg("you are not leaving this place without that amulet."))
			}
		} else {
			g.ascendWithStairs(stairs)
			if !g.Player.GetInventory().HasItemWithName("amulet_of_yendor") {
				if g.unstableStairs() {
					if g.currentDungeonLevel < 26 {
						stairs = StairsDownOnly
					}
					if rand.Intn(10) == 0 {
						g.msg(foundation.Msg("the stairs are crumbling beneath you, you fall deep down."))
						g.GotoDungeonLevel(g.currentDungeonLevel+2, stairs, true)
					} else {
						g.msg(foundation.Msg("the stairs are crumbling beneath you, you fall down."))
						g.descendWithStairs(stairs)
					}

					return
				}
				g.ascensionsWithoutAmulet++
				g.msg(foundation.Msg("you feel that the dungeon becomes unstable."))
			}
		}
	}
}
func (g *GameState) Descend() {
	g.descendWithStairs(StairsBoth)
}

func (g *GameState) descendWithStairs(stairs StairsInLevel) {
	g.GotoDungeonLevel(g.currentDungeonLevel+1, stairs, true)
}

func (g *GameState) descendToRandomLocation() {
	g.GotoDungeonLevel(g.currentDungeonLevel+1, StairsBoth, false)
}

func (g *GameState) Ascend() {
	g.ascendWithStairs(StairsBoth)
}

func (g *GameState) ascendWithStairs(stairs StairsInLevel) {
	if g.currentDungeonLevel == 0 {
		return
	}
	if g.currentDungeonLevel == 1 {
		g.currentDungeonLevel = 0
		g.GotoNamedLevel("town")
		g.gridMap.SetAllExplored()
		g.gridMap.SetAllLit()
	} else {
		g.GotoDungeonLevel(g.currentDungeonLevel-1, stairs, true)
	}
}

func (g *GameState) actorMeleeAttack(attacker *Actor, attackMod []rpg.Modifier, defender *Actor, defenseMod []rpg.Modifier) []foundation.Animation {
	if !defender.IsAlive() {
		return nil
	}
	var afterAttackAnimations []foundation.Animation

	if attacker.HasFlag(foundation.FlagCanConfuse) {
		attacker.GetFlags().Unset(foundation.FlagCanConfuse)
		confuseAnim := confuse(g, defender)
		afterAttackAnimations = append(afterAttackAnimations, confuseAnim...)
		g.msg(foundation.HiLite("%s stops glowing red", attacker.Name()))
	}

	attackerMeleeSkill, attackerMeleeDamageDice := attacker.GetMelee(defender.GetInternalName())

	defenseScore := defender.GetActiveDefenseScore()

	attackerMeleeSkill, defenseScore = g.applyAttackAndDefenseMods(attackerMeleeSkill, attackMod, defenseScore, defenseMod)

	if defender.HasFlag(foundation.FlagSleep) {
		defenseScore = -1
	}

	outcome := rpg.Attack(attackerMeleeSkill, attackerMeleeDamageDice, defenseScore, defender.GetDamageResistance())

	_, damageDone := outcome.TypeOfHit, outcome.DamageDone

	for _, message := range outcome.String(attacker.Name(), defender.Name()) {
		g.msg(foundation.Msg(message))
	}

	animAttackerIndicator := g.ui.GetAnimBackgroundColor(attacker.Position(), "VeryDarkGray", 4, nil)
	afterAttackAnimations = append(afterAttackAnimations, animAttackerIndicator)

	if outcome.IsHit() {
		animDamage := g.damageActor(attacker.Name(), defender, damageDone)
		//animAttack := g.ui.GetAnimAttack(attacker, defender) // currently no attack animation
		afterAttackAnimations = append(afterAttackAnimations, animDamage...)
		//animAttack.SetFollowUp(afterAttackAnimations)
	} else {
		animMiss := g.ui.GetAnimDamage(defender.Position(), 0, nil)
		afterAttackAnimations = append(afterAttackAnimations, animMiss)
	}

	return afterAttackAnimations
}

func (g *GameState) applyAttackAndDefenseMods(attackerMeleeSkill int, attackMod []rpg.Modifier, defenseScore int, defenseMod []rpg.Modifier) (int, int) {
	// apply situational modifiers
	var attackModDescriptions []string
	for _, mod := range rpg.FilterModifiers(attackMod) {
		attackerMeleeSkill = mod.Apply(attackerMeleeSkill)
		attackModDescriptions = append(attackModDescriptions, mod.Description())
	}
	var defenseModDescriptions []string
	for _, mod := range rpg.FilterModifiers(defenseMod) {
		defenseScore = mod.Apply(defenseScore)
		defenseModDescriptions = append(defenseModDescriptions, mod.Description())
	}
	if len(attackModDescriptions) > 0 {
		g.msg(foundation.HiLite("Attack modifiers\n%s", strings.Join(attackModDescriptions, "\n")))
	}
	if len(defenseModDescriptions) > 0 {
		g.msg(foundation.HiLite("Defense modifiers\n%s", strings.Join(defenseModDescriptions, "\n")))
	}
	g.msg(foundation.Msg(fmt.Sprintf("Attacker Effective Skill: %d", attackerMeleeSkill)))
	g.msg(foundation.Msg(fmt.Sprintf("Defender Effective Skill: %d", defenseScore)))
	return attackerMeleeSkill, defenseScore
}
func (g *GameState) actorRangedAttack(attacker *Actor, attackMod []rpg.Modifier, defender *Actor, defenseMod []rpg.Modifier, missile *Item) []foundation.Animation {
	if !defender.IsAlive() {
		return nil
	}

	var rangedSkill int
	var rangedDamage rpg.Dice

	if attacker.IsLaunching(missile) {
		launcher := attacker.GetEquipment().GetMissileLauncher()
		rangedSkill, rangedDamage = attacker.GetRanged(defender.GetInternalName(), launcher, missile)
	} else {
		rangedSkill, rangedDamage = attacker.GetThrowing(defender.GetInternalName(), missile)
	}

	defenseScore := defender.GetActiveDefenseScore()

	rangedSkill, defenseScore = g.applyAttackAndDefenseMods(rangedSkill, attackMod, defenseScore, defenseMod)

	outcome := rpg.Attack(rangedSkill, rangedDamage, defenseScore, defender.GetDamageResistance())
	_, damageDone := outcome.TypeOfHit, outcome.DamageDone

	for _, message := range outcome.String(attacker.Name(), defender.Name()) {
		g.msg(foundation.Msg(message))
	}

	var afterAttackAnimations []foundation.Animation

	animDamage := g.damageActor(attacker.Name(), defender, damageDone)

	afterAttackAnimations = append(afterAttackAnimations, animDamage...)

	return afterAttackAnimations
}
func (g *GameState) PickupItem() {
	inventory := g.Player.GetInventory()
	if inventory.IsFull() {
		g.msg(foundation.Msg("You cannot carry any more items"))
		return
	}
	if item, exists := g.gridMap.TryGetItemAt(g.Player.Position()); exists {
		g.gridMap.RemoveItem(item)
		if item.IsGold() {
			g.Player.AddGold(item.GetCharges())
		} else {
			inventory.Add(item)
		}

		g.msg(foundation.HiLite("You picked up %s", item.Name()))
		//g.endPlayerTurn()
	}
}

func (g *GameState) DropItem(uiItem foundation.ItemForUI) {
	itemStack, isItem := uiItem.(*InventoryStack)
	if !isItem {
		return
	}
	item := itemStack.First()
	g.actorDropItem(g.Player, item)
}

func (g *GameState) actorDropItem(holder *Actor, item *Item) {
	equipment := holder.GetEquipment()
	if equipment.IsEquipped(item) {
		if equipment.CanUnequip(item) {
			g.actorUnequipItem(holder, item)
		} else {
			g.msg(foundation.Msg("You cannot remove this item"))
			return
		}
	}

	g.removeItemFromInventory(holder, item)
	g.addItemToMap(item, holder.Position())
	g.msg(foundation.HiLite("You dropped %s", item.Name()))
	if holder == g.Player {
		g.endPlayerTurn()
	}
}

func (g *GameState) inspectItem(item *Item) func() {
	return func() {
		g.ui.OpenTextWindow([]string{"Item", item.name})
	}
}
func (g *GameState) actorRangedAttackWithMissile(thrower *Actor, missile *Item, origin, targetPos geometry.Point) {
	pathOfFlight := geometry.BresenhamLine(origin, targetPos, func(x, y int) bool {
		if origin.X == x && origin.Y == y {
			return true
		}
		return g.gridMap.IsCurrentlyPassable(geometry.Point{X: x, Y: y})
	})
	if len(pathOfFlight) > 1 {
		// remove start
		pathOfFlight = pathOfFlight[1:]
	}
	targetPos = pathOfFlight[len(pathOfFlight)-1]
	if !g.gridMap.IsTileWalkable(targetPos) && len(pathOfFlight) > 1 {
		targetPos = pathOfFlight[len(pathOfFlight)-2]
	}
	var onHitAnimations []foundation.Animation

	g.removeItemFromInventory(thrower, missile)

	g.addItemToMap(missile, targetPos)

	throwAnim, _ := g.ui.GetAnimThrow(missile, origin, targetPos)

	if g.gridMap.IsActorAt(targetPos) {
		defender := g.gridMap.ActorAt(targetPos)
		//isLaunch := thrower.IsLaunching(missile) // otherwise it's a throw
		consequenceOfActorHit := g.actorRangedAttack(thrower, ModRangedDefault(thrower.Position(), defender), defender, NoModifiers, missile)
		onHitAnimations = append(onHitAnimations, consequenceOfActorHit...)
	} else if g.gridMap.IsObjectAt(targetPos) {
		object := g.gridMap.ObjectAt(targetPos)
		consequenceOfObjectHit := object.OnDamage()
		onHitAnimations = append(onHitAnimations, consequenceOfObjectHit...)
	}

	if throwAnim != nil {
		throwAnim.SetFollowUp(onHitAnimations)
	}

	g.ui.AddAnimations([]foundation.Animation{throwAnim})

	if thrower == g.Player {
		g.endPlayerTurn()
	}
}

func ModRangedDefault(origin geometry.Point, target *Actor) []rpg.Modifier {
	// TODO: Light
	var attackMods []rpg.Modifier

	// step 1. size modifier
	sizeMod := target.GetSizeModifier()
	if sizeMod != 0 {
		attackMods = append(attackMods, ModFlat(sizeMod, "size"))
	}

	// step 2. range modifier
	dist := geometry.Distance(origin, target.Position())
	rangeUsed := int(dist)
	distMod := rpg.GetDistanceModifier(rangeUsed)
	if distMod != 0 {
		attackMods = append(attackMods, ModFlat(distMod, fmt.Sprintf("range(%d)", rangeUsed)))
	}

	return attackMods
}

func OneAnimation(anim foundation.Animation) []foundation.Animation {
	if anim == nil {
		return nil
	}
	return []foundation.Animation{anim}
}
func (g *GameState) EquipToggle(uiItem foundation.ItemForUI) {
	itemStack, isItem := uiItem.(*InventoryStack)
	if !isItem {
		return
	}
	item := itemStack.First()
	if !item.IsEquippable() {
		g.msg(foundation.Msg("You cannot equip this item"))
		return
	}
	equipment := g.Player.GetEquipment()
	if equipment.IsEquipped(item) {
		if equipment.CanUnequip(item) {
			g.actorUnequipItem(g.Player, item)
		} else {
			g.msg(foundation.Msg("You cannot remove this item"))
		}
	} else {
		if equipment.CanEquip(item) {
			g.actorEquipItem(g.Player, item)
		} else {
			g.msg(foundation.Msg("You cannot equip this item"))
		}
	}
}

func (g *GameState) actorEquipItem(wearer *Actor, item *Item) {
	if item.IsRing() && !g.identification.IsItemIdentified(item.GetInternalName()) && g.identification.CanBeIdentifiedByUsing(item.GetInternalName()) {
		g.identification.IdentifyItem(item.GetInternalName())
	}
	equipment := wearer.GetEquipment()
	equipment.Equip(item)
	if wearer == g.Player {
		g.msg(foundation.HiLite("You equipped %s", item.Name()))
	}
}

func (g *GameState) actorUnequipItem(wearer *Actor, item *Item) {
	equipment := wearer.GetEquipment()
	equipment.UnEquip(item)
	if wearer == g.Player {
		g.msg(foundation.HiLite("You unequipped %s", item.Name()))
	}
}


func (g *GameState) ChooseItemForApply() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsUsableOrZappable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything usable."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.playerUseOrZapItem(item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Use what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.playerUseOrZapItem(item)
	})
}



func (g *GameState) ChooseItemForDrop() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return true
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.DropItem(stack)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Drop what?", func(itemStack foundation.ItemForUI) {
		g.DropItem(itemStack)
	})
}

func (g *GameState) ChooseItemForQuaff() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsPotion()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any potions."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Quaff what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
	})
}

func (g *GameState) ChooseItemForEat() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsFood()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any food."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Eat what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
	})
}

func (g *GameState) ChooseItemForRead() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsScroll()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any scrolls."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Read what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
	})
}

func (g *GameState) ChooseItemForZap() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsZappable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any wands."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.startAimItem(item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Zap what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.startAimItem(item)
	})
}

func (g *GameState) ChooseItemForUse() {
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsUsable()
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying anything usable."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
		return

	}
	g.ui.OpenInventoryForSelection(inventory, "Use what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUseItem(g.Player, item)
	})
}

func (g *GameState) ChooseWeaponForWield() {
	equipment := g.Player.GetEquipment()
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsWeapon() && item.IsEquippable() && !equipment.IsEquipped(item)
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any weapons."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorEquipItem(g.Player, stack.First())
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Wield what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorEquipItem(g.Player, item)
	})
}

func (g *GameState) ChooseArmorForWear() {
	equipment := g.Player.GetEquipment()
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsArmor() && item.IsEquippable() && !equipment.IsEquipped(item)
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any armor."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorEquipItem(g.Player, stack.First())
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Wear what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorEquipItem(g.Player, item)
	})
}

func (g *GameState) ChooseRingToPutOn() {
	equipment := g.Player.GetEquipment()
	inventory := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsRing() && item.IsEquippable() && !equipment.IsEquipped(item)
	})
	if len(inventory) == 0 {
		g.msg(foundation.Msg("You are not carrying any rings."))
		return
	}
	if len(inventory) == 1 {
		stack, isStack := inventory[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorEquipItem(g.Player, stack.First())
		return
	}
	g.ui.OpenInventoryForSelection(inventory, "Put on what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorEquipItem(g.Player, item)
	})
}

func (g *GameState) ChooseArmorToTakeOff() {
	equipment := g.Player.GetEquipment()
	wornArmor := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsArmor() && item.IsEquippable() && equipment.IsEquipped(item)
	})

	if len(wornArmor) == 0 {
		g.msg(foundation.Msg("You are not wearing any armor."))
		return
	}
	if len(wornArmor) == 1 {
		stack, isStack :=  wornArmor[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorUnequipItem(g.Player, stack.First())
		return
	}
	g.ui.OpenInventoryForSelection(wornArmor, "Take off what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUnequipItem(g.Player, item)
	})
}

func (g *GameState) ChooseRingToRemove() {
	equipment := g.Player.GetEquipment()
	wornRings := g.GetFilteredInventory(func(item *Item) bool {
		return item.IsRing() && item.IsEquippable() && equipment.IsEquipped(item)
	})
	if len(wornRings) == 0 {
		g.msg(foundation.Msg("You are not wearing any rings."))
		return
	}
	if len(wornRings) == 1 {
		stack, isStack :=  wornRings[0].(*InventoryStack)
		if !isStack {
			return
		}
		g.actorUnequipItem(g.Player, stack.First())
		return

	}
	g.ui.OpenInventoryForSelection(wornRings, "Remove what?", func(itemStack foundation.ItemForUI) {
		stack, isStack := itemStack.(*InventoryStack)
		if !isStack {
			return
		}
		item := stack.First()
		g.actorUnequipItem(g.Player, item)
	})
}
