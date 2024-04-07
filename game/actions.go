package game

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/rpg"
	"fmt"
	"strings"
)

var NoModifiers []rpg.Modifier

func (g *GameState) playerAttack(defender *Actor) {
	consequences := g.actorMeleeAttack(g.Player, NoModifiers, defender, NoModifiers)
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
		g.playerZapItemAndEndTurn(item, targetPos)
	})
}
func (g *GameState) startAimZapEffect(zapEffectName string) {
	g.ui.SelectTarget(g.Player.Position(), func(targetPos geometry.Point) {
		g.playerInvokeZapEffectAndEndTurn(zapEffectName, targetPos)
	})
}
func (g *GameState) PlayerUseOrZapItem(uiItem foundation.ItemForUI) {
	itemStack, isItem := uiItem.(*InventoryStack)
	if !isItem {
		return
	}
	item := itemStack.First()
	g.playerUseOrZapItem(item)
}
func (g *GameState) playerUseOrZapItem(item *Item) {
	if item.IsUsable() {
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
	var moveAnim foundation.Animation
	if g.couldPlayerSeeActor(actor) && (g.canPlayerSee(newPos) || g.canPlayerSee(oldPos)) && actor != g.Player {
		moveAnim = g.ui.GetAnimMove(actor, oldPos, newPos)
	}
	g.actorMove(actor, newPos)
	return OneAnimation(moveAnim)
}
func (g *GameState) couldPlayerSeeActor(actor *Actor) bool {
	if actor.HasFlag(foundation.IsInvisible) && !g.Player.HasFlag(foundation.CanSeeInvisible) {
		return false
	}

	return true
}
func (g *GameState) actorMove(actor *Actor, newPos geometry.Point) {
	g.gridMap.MoveActor(actor, newPos)
}

func (g *GameState) PlayerInteractWithMap() {
	pos := g.Player.Position()
	cell := g.gridMap.GetCell(pos)

	if cell.TileType.IsStairsDown() {
		g.Descend()
	} else if cell.TileType.IsStairsUp() {
		g.Ascend()
	}
}

func (g *GameState) Descend() {
	g.GotoDungeonLevel(g.currentDungeonLevel + 1)
}

func (g *GameState) Ascend() {
	if g.currentDungeonLevel == 0 {
		return
	}
	if g.currentDungeonLevel == 1 {
		g.currentDungeonLevel = 0
		g.GotoNamedLevel("town")
		g.gridMap.SetAllExplored()
		g.gridMap.SetAllLit()
	} else {
		g.GotoDungeonLevel(g.currentDungeonLevel - 1)
	}
}

func (g *GameState) actorMeleeAttack(attacker *Actor, attackMod []rpg.Modifier, defender *Actor, defenseMod []rpg.Modifier) []foundation.Animation {
	if !defender.IsAlive() {
		return nil
	}
	var afterAttackAnimations []foundation.Animation

	if attacker.HasFlag(foundation.CanConfuse) {
		attacker.GetFlags().Unset(foundation.CanConfuse)
		confuseAnim := confuse(g, defender)
		afterAttackAnimations = append(afterAttackAnimations, confuseAnim...)
		g.msg(foundation.HiLite("%s stops glowing red", attacker.Name()))
	}

	attackerMeleeSkill, attackerMeleeDamageDice := attacker.GetMelee(defender.GetInternalName())

	defenseScore := defender.GetActiveDefenseScore()

	attackerMeleeSkill, defenseScore = g.applyAttackAndDefenseMods(attackerMeleeSkill, attackMod, defenseScore, defenseMod)

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
	if g.Player.GetEquipment().IsEquipped(item) {
		g.actorUnequipItem(g.Player, item)
	} else {
		g.actorEquipItem(g.Player, item)
	}
}

func (g *GameState) actorEquipItem(wearer *Actor, item *Item) {
	equipment := wearer.GetEquipment()
	equipment.Equip(item)
}

func (g *GameState) actorUnequipItem(wearer *Actor, item *Item) {
	equipment := wearer.GetEquipment()
	equipment.UnEquip(item)
}
