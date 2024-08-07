package game

import (
    "RogueUI/dice_curve"
    "RogueUI/foundation"
    "RogueUI/special"
    "fmt"
    "github.com/memmaker/go/fxtools"
    "github.com/memmaker/go/geometry"
    "math/rand"
    "strings"
)

// melee attacks with and without weapons
// firearms with different modes of fire
// throwing weapons

// Logic & Animation
// - Melee Attack (with and without weapons)
// - Throw Item
// - Ranged Attack with fire mode

// OFFENSIVE ACTIONS

func (g *GameState) PlayerRangedAttack() {
    mainHandItem, hasWeapon := g.Player.GetEquipment().GetMainHandItem()
    if !hasWeapon || !mainHandItem.IsRangedWeapon() {
        g.msg(foundation.Msg("You have no suitable weapon equipped"))
        return
    }
    weapon := mainHandItem.GetWeapon()
    if !weapon.HasAmmo() {
        g.msg(foundation.Msg("You have no ammo"))
        return
    }
    targetMode := weapon.GetCurrentTargetingMode()
    if targetMode == TargetingModeAimed {
        g.ui.SelectBodyPart(func(victim foundation.ActorForUI, bodyPart int) {
            g.msg(foundation.HiLite("You aim at %s's %s", victim.Name(), victim.GetBodyPartByIndex(bodyPart)))
            target := victim.(*Actor)
            shotAnim := g.actorRangedAttack(g.Player, mainHandItem, targetMode, target, bodyPart)
            g.ui.AddAnimations(shotAnim)
            g.endPlayerTurn(10)
        })
    } else {
        g.ui.SelectTarget(func(targetPos geometry.Point, bodyPart int) {
            if g.gridMap.IsActorAt(targetPos) {
                target := g.gridMap.ActorAt(targetPos)
                shotAnim := g.actorRangedAttack(g.Player, mainHandItem, targetMode, target, bodyPart)
                g.ui.AddAnimations(shotAnim)
                g.endPlayerTurn(10)
            }
        })
    }

}

func (g *GameState) PlayerQuickRangedAttack() {
    enemies := g.playerVisibleEnemiesByDistance()
    equipment := g.Player.GetEquipment()
    mainHandItem, hasWeapon := equipment.GetMainHandItem()

    if !hasWeapon || !mainHandItem.IsRangedWeapon() {
        g.msg(foundation.Msg("You have no suitable weapon equipped"))
        return
    }
    weapon := mainHandItem.GetWeapon()
    if !weapon.HasAmmo() {
        g.msg(foundation.Msg("You have no ammo"))
        return
    }
    if len(enemies) == 0 {
        g.msg(foundation.Msg("No enemies in sight"))
        return
    }

    mode := weapon.GetCurrentTargetingMode()
    if mode == TargetingModeAimed {
        mode = TargetingModeSingle
    }
    shotAnim := g.actorRangedAttack(g.Player, mainHandItem, mode, enemies[0], 2)
    g.ui.AddAnimations(shotAnim)
    g.endPlayerTurn(10)
}

func (g *GameState) QuickThrow() {
    enemies := g.playerVisibleEnemiesByDistance()
    preselectedTarget := g.Player.Position()
    equipment := g.Player.GetEquipment()
    weapon, hasWeapon := equipment.GetMainHandItem()
    if hasWeapon || !weapon.IsMissile() {
        g.msg(foundation.Msg("You have no suitable weapon equipped"))
        return
    }
    if len(enemies) == 0 {
        g.msg(foundation.Msg("No enemies in sight"))
        return
    }
    preselectedTarget = enemies[0].Position()
    g.actorThrowItem(g.Player, weapon, g.Player.Position(), preselectedTarget)
}

func (g *GameState) playerMeleeAttack(defender *Actor) {
    consequences := g.actorMeleeAttack(g.Player, defender)
    if !g.Player.HasFlag(foundation.FlagInvisible) {
        defender.GetFlags().Set(foundation.FlagAwareOfPlayer)
    }
    g.ui.AddAnimations(consequences)
    g.endPlayerTurn(10)
}

func (g *GameState) actorMeleeAttack(attacker *Actor, defender *Actor) []foundation.Animation {
    if !defender.IsAlive() {
        return nil
    }
    var afterAttackAnimations []foundation.Animation

    itemInHand, hasItem := attacker.GetEquipment().GetMeleeWeapon()
    attackerSkill := special.Unarmed
    damage := fxtools.NewInterval(1, 4)
    attackAudioCue := ""
    if attacker.IsHuman() {
        attackAudioCue = "weapons/unarmed"
    } else {
        attackAudioCue = fmt.Sprintf("enemies/%s/Attack", attacker.GetInternalName())
    }
    if hasItem && itemInHand.IsMeleeWeapon() {
        weapon := itemInHand.GetWeapon()
        attackerSkill = special.MeleeWeapons
        damage = weapon.GetDamage()
        if attacker.IsHuman() {
            attackAudioCue = fmt.Sprintf("weapons/%s", itemInHand.GetInternalName())
        }
    }

    chanceToHit := special.MeleeChanceToHit(attacker.GetCharSheet(), attackerSkill, defender.GetCharSheet(), special.Body)
    //attackerMeleeSkill, attackerMeleeDamageDice := attacker.GetMelee(defender.GetInternalName())

    //defenseScore := defender.GetActiveDefenseScore()

    //attackerMeleeSkill, defenseScore = g.applyAttackAndDefenseMods(attackerMeleeSkill, attackMod, defenseScore, defenseMod)

    //outcome := dice_curve.Attack(attackerMeleeSkill, attackerMeleeDamageDice, defenseScore, defender.GetDamageResistance())

    //_, damageDone := outcome.TypeOfHit, outcome.DamageDone

    /*
    	for _, message := range outcome.String(attacker.Name(), defender.Name()) {
    		g.msg(foundation.Msg(message))
    	}

    */

    animAttackerIndicator := g.ui.GetAnimBackgroundColor(attacker.Position(), "dark_gray_6", 4, nil)
    animAttackerIndicator.SetAudioCue(attackAudioCue)

    afterAttackAnimations = append(afterAttackAnimations, animAttackerIndicator)

    if rand.Intn(100) < chanceToHit {
        // hit
        damageDone := damage.Roll()
        damageAnims := g.damageActor(attacker.Name(), defender, damageDone)
        afterAttackAnimations = append(afterAttackAnimations, damageAnims...)
    } else {
        afterAttackAnimations = append(afterAttackAnimations, g.ui.GetAnimEvade(defender, nil))
    }
    /*
    	if outcome.IsHit() {
    		//animDamage := g.damageActor(attacker.Name(), defender, damageDone)
    		//animAttack := g.ui.GetAnimAttack(attacker, defender) // currently no attack animation
    		//afterAttackAnimations = append(afterAttackAnimations, animDamage...)
    		//animAttack.SetFollowUp(afterAttackAnimations)
    	} else {
    		animMiss := g.ui.GetAnimDamage(defender.Position(), 0, nil)
    		afterAttackAnimations = append(afterAttackAnimations, animMiss)
    	}
    */

    return afterAttackAnimations
}

func (g *GameState) applyAttackAndDefenseMods(attackerMeleeSkill int, attackMod []dice_curve.Modifier, defenseScore int, defenseMod []dice_curve.Modifier) (int, int) {
    // apply situational modifiers
    var attackModDescriptions []string
    for _, mod := range dice_curve.FilterModifiers(attackMod) {
        attackerMeleeSkill = mod.Apply(attackerMeleeSkill)
        attackModDescriptions = append(attackModDescriptions, mod.Description())
    }
    var defenseModDescriptions []string
    for _, mod := range dice_curve.FilterModifiers(defenseMod) {
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

// actorRangedAttack logic and animation of a ranged attack with the equipped weapon
func (g *GameState) actorRangedAttack(attacker *Actor, weaponItem *Item, mode TargetingMode, defender *Actor, bodyPart int) []foundation.Animation {
    if !defender.IsAlive() {
        return nil
    }

    bulletsSpent := 1
    weapon := weaponItem.GetWeapon()
    if mode == TargetingModeBurst {
        bulletsSpent = min(weapon.GetLoadedBullets(), weapon.GetBurstRounds())
    } else if mode == TargetingModeFullAuto {
        bulletsSpent = weapon.GetLoadedBullets()
    }

    weapon.RemoveBullets(bulletsSpent)

    animAttackerIndicator := g.ui.GetAnimBackgroundColor(attacker.Position(), "White", 4, nil)
    animAttackerIndicator.SetAudioCue(fmt.Sprintf("weapons/%s", weaponItem.GetInternalName()))

    var onAttackAnims []foundation.Animation
    onAttackAnims = append(onAttackAnims, animAttackerIndicator)
    var posInfos special.PosInfo
    posInfos.ObstacleCount = 0
    posInfos.Distance = g.gridMap.MoveDistance(attacker.Position(), defender.Position())
    posInfos.IlluminationPenalty = 0
    defenderSheet := defender.GetCharSheet()
    chanceToHit := special.RangedChanceToHit(posInfos, attacker.GetCharSheet(), weapon.GetSkillUsed(), defenderSheet, special.BodyPart(bodyPart))
    damage := weapon.GetDamage()
    var damageAnims []foundation.Animation
    for i := 0; i < bulletsSpent; i++ {
        if rand.Intn(100) < chanceToHit {
            // hit
            damageDone := damage.Roll()
            damageAnims = g.damageActor(attacker.Name(), defender, damageDone)
        }
    }
    if len(damageAnims) == 0 {
        damageAnims = []foundation.Animation{g.ui.GetAnimEvade(defender, nil)}
    }
    return append(onAttackAnims, damageAnims...)
}

// Validation for Player Commands
func (g *GameState) Throw() {
    equipment := g.Player.GetEquipment()
    if !equipment.HasRangedWeaponEquipped() {
        g.msg(foundation.Msg("You have no quivered missile"))
        return
    }
    weapon, hasWeapon := equipment.GetMainHandItem()
    if hasWeapon || !weapon.IsRangedWeapon() {
        g.msg(foundation.Msg("You have no suitable weapon equipped"))
        return
    }
    g.startThrowItem(weapon)
}

// Target Selection Stage
func (g *GameState) startThrowItem(item *Item) {
    g.ui.SelectTarget(func(targetPos geometry.Point, hitZone int) {
        g.actorThrowItem(g.Player, item, g.Player.Position(), targetPos)
    })
}

// Logic And Animation
func (g *GameState) actorThrowItem(thrower *Actor, missile *Item, origin, targetPos geometry.Point) {
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
        consequenceOfActorHit := g.actorRangedAttack(thrower, nil, 0, defender, 0)
        onHitAnimations = append(onHitAnimations, consequenceOfActorHit...)
    } else if g.gridMap.IsObjectAt(targetPos) {
        object := g.gridMap.ObjectAt(targetPos)
        consequenceOfObjectHit := object.OnDamage(thrower)
        onHitAnimations = append(onHitAnimations, consequenceOfObjectHit...)
    }

    if throwAnim != nil {
        throwAnim.SetFollowUp(onHitAnimations)
    }

    g.ui.AddAnimations([]foundation.Animation{throwAnim})

    if thrower == g.Player {
        g.endPlayerTurn(10)
    }
}
