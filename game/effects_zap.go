package game

import (
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"RogueUI/special"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"math/rand"
)

func GetAllZapEffects() map[string]func(g *GameState, zapper *Actor, aimPos geometry.Point, params Params) []foundation.Animation {
	var zapEffects = map[string]func(g *GameState, zapper *Actor, aimPos geometry.Point, params Params) []foundation.Animation{
		"fire_breath":    fireBreath,
		"explode":        explosion,
		"plasma_explode": plasmaExplosion,
		//"magic_missile":        magicMissile,
		//"haste_target":         hasteTarget,
		//"slow_target":          slowTarget,
		//"teleport_target_away": teleportTargetAway,
		//"teleport_target_to":   teleportTargetTo,
		//"cancel_target":        cancelTarget,
		//"invisibility_target":  invisibilityTarget,
		//"cold_ray":             coldRay,
		//"lightning_ray":        lightningRay,
		//"fire_ray":             fireRay,
		//"charge_attack":        chargeAttack,
		//"heroic_charge":        heroicCharge,
		//"magic_dart":           magicDart,
		//"magic_arrow":          magicArrow,
		//"hold_target":          holdTarget,
	}
	return zapEffects
}

/*
func forceDescendTarget(g *GameState, zapper *Actor, pos geometry.Point) []foundation.Animation {
    if !g.currentMap().IsActorAt(pos) {
        return nil
    }

    descendingActor := g.currentMap().ActorAt(pos)
    if descendingActor == g.Player {
        g.QueueActionAfterAnimation(g.descendToRandomLocation)
    } else {
        g.currentMap().RemoveActor(descendingActor)
    }
    return nil
}

*/

func magicDart(g *GameState, zapper *Actor, pos geometry.Point) []foundation.Animation {
	return magicItemProjectile(g, zapper, pos, "dart", "a dart")
}

func magicArrow(g *GameState, zapper *Actor, pos geometry.Point) []foundation.Animation {
	return magicItemProjectile(g, zapper, pos, "arrow", "an arrow")
}

func uncloakAndCharge(g *GameState, zapper *Actor, pos geometry.Point) []foundation.Animation {
	zapper.GetFlags().Unset(special.FlagInvisible)
	zapper.SetAware()
	uncloakAnim, _ := g.ui.GetAnimUncloakAtPosition(zapper, zapper.Position())
	chargeAnim, _ := charge(g, zapper, pos, false, g.getLine)
	//tileIcon := g.currentMap().GetTileIconAt(targetPos)
	uncloakAnim.SetFollowUp([]foundation.Animation{chargeAnim})

	return []foundation.Animation{uncloakAnim}
}
func chargeAttack(g *GameState, zapper *Actor, pos geometry.Point) []foundation.Animation {
	moveAnim, _ := charge(g, zapper, pos, false, g.getLineOfSight)
	return []foundation.Animation{moveAnim}
}
func heroicCharge(g *GameState, zapper *Actor, pos geometry.Point) []foundation.Animation {
	moveAnim, _ := charge(g, zapper, pos, true, g.getLineOfSight)
	return []foundation.Animation{moveAnim}
}
func charge(g *GameState, zapper *Actor, pos geometry.Point, isHeroic bool, getPath func(src, dst geometry.Point) []geometry.Point) (foundation.Animation, geometry.Point) {
	origin := zapper.Position()
	pathOfFlight := getPath(origin, pos)

	var hitActor *Actor
	targetPos := pathOfFlight[len(pathOfFlight)-1]
	if !g.currentMap().IsCurrentlyPassable(targetPos) && len(pathOfFlight) > 1 {
		if g.currentMap().IsActorAt(targetPos) {
			hitActor = g.currentMap().ActorAt(targetPos)
		}
		targetPos = pathOfFlight[len(pathOfFlight)-2]
	}
	g.actorMove(zapper, targetPos)
	pathOfFlight = append([]geometry.Point{origin}, pathOfFlight...)
	moveAnim := g.ui.GetAnimQuickMove(zapper, pathOfFlight)

	if hitActor != nil {
		attackAnims := g.actorMeleeAttack(zapper, hitActor, 0) // TODO: apply -4 to hit, cap effective skill at 9
		moveAnim.SetFollowUp(attackAnims)
	}
	return moveAnim, targetPos
}

func coldRay(g *GameState, zapper *Actor, aimPos geometry.Point) []foundation.Animation {
	damage := max(1, dice_curve.Spread(8, 0.35))
	trailLead := '☼'
	trailColors := []string{"White", "White", "LightCyan", "light_blue_3", "Blue"}
	hitEntityHandler := func(hitPos geometry.Point) []foundation.Animation {
		if g.currentMap().IsActorAt(hitPos) {
			actor := g.currentMap().ActorAt(hitPos)
			if actor.IsAlive() {
				freeze := func() {
					g.msg(foundation.HiLite("%s is frozen", actor.Name()))
					actor.GetFlags().Set(special.FlagHeld)
				}
				damageWithSource := SourcedDamage{
					NameOfThing:     "ice ray",
					Attacker:        zapper,
					IsObviousAttack: true,
					TargetingMode:   special.TargetingModeFireSingle,
					DamageType:      special.DamageTypePoison,
					DamageAmount:    damage,
				}
				damageAnim := g.damageActorWithFollowUp(damageWithSource, actor, freeze, nil)
				return damageAnim
			}
		}
		return nil
	}

	if rand.Intn(20) == 0 { // 1 in 20 chance to just bounce off
		bounceCount := rand.Intn(30) + 3
		return g.bouncingRay(zapper, aimPos, bounceCount, trailLead, trailColors, hitEntityHandler)
	}

	return g.singleRay(zapper.Position(), aimPos, trailLead, trailColors, hitEntityHandler)

}
func plasmaExplosion(g *GameState, zapper *Actor, loc geometry.Point, params Params) []foundation.Animation {
	radius := params.GetIntOrDefault("radius", 3)
	damageAmount := params.GetDamageOrDefault(35)

	affected := g.currentMap().GetDijkstraMap(loc, radius, func(p geometry.Point) bool {
		return g.currentMap().IsTileWalkable(p) || g.currentMap().IsTileWithFlagAt(p, gridmap.TileFlagDestroyable)
	})

	var animationsForThisFrame []foundation.Animation
	var deferredAnimations []foundation.Animation
	damage := SourcedDamage{
		NameOfThing:     "plasma_explosion",
		Attacker:        zapper,
		IsObviousAttack: true,
		TargetingMode:   special.TargetingModeFireSingle,
		DamageType:      special.DamageTypePlasma,
		DamageAmount:    damageAmount,
	}
	for point, _ := range affected {
		damageAnims := g.damageLocation(damage, point)
		deferredAnimations = append(deferredAnimations, damageAnims...)
	}

	plasmaColor := fxtools.HDRColor{R: 0.2, G: 1.0, B: 0.2, A: 1}
	explosionAnim := g.ui.GetAnimRadialExplosion(affected, plasmaColor, nil)
	explosionAnim.SetAudioCue("world/explosion")
	if len(deferredAnimations) > 0 && explosionAnim != nil {
		explosionAnim.SetFollowUp(deferredAnimations)
	}

	animationsForThisFrame = append(animationsForThisFrame, explosionAnim)

	return animationsForThisFrame
}
func explosion(g *GameState, zapper *Actor, loc geometry.Point, params Params) []foundation.Animation {
	radius := params.GetIntOrDefault("radius", 3)
	damageAmount := params.GetDamageOrDefault(35)

	affected := g.currentMap().GetDijkstraMap(loc, radius, func(p geometry.Point) bool {
		return g.currentMap().IsTileWalkable(p) || g.currentMap().IsTileWithFlagAt(p, gridmap.TileFlagDestroyable)
	})

	var animationsForThisFrame []foundation.Animation
	var deferredAnimations []foundation.Animation
	var affectedPoints []geometry.Point
	damage := SourcedDamage{
		NameOfThing:     "explosion",
		Attacker:        zapper,
		IsObviousAttack: true,
		TargetingMode:   special.TargetingModeFireSingle,
		DamageType:      special.DamageTypeExplosive,
		DamageAmount:    damageAmount,
	}
	for point, _ := range affected {
		damageAnims := g.damageLocation(damage, point)
		deferredAnimations = append(deferredAnimations, damageAnims...)
		affectedPoints = append(affectedPoints, point)
	}

	explosionAnim := g.ui.GetAnimExplosion(affectedPoints, nil)
	explosionAnim.SetAudioCue("world/explosion")
	if len(deferredAnimations) > 0 && explosionAnim != nil {
		explosionAnim.SetFollowUp(deferredAnimations)
	}

	animationsForThisFrame = append(animationsForThisFrame, explosionAnim)

	return animationsForThisFrame
}
func (g *GameState) singleRay(origin, target geometry.Point, leadIcon rune, trailColors []string, hitEntityHandler func(hitPos geometry.Point) []foundation.Animation) []foundation.Animation {
	direction := target.ToCenteredPointF().Sub(origin.ToCenteredPointF())
	rayHitInfo := g.currentMap().RayCast(origin.ToCenteredPointF(), direction, g.IsBlockingRay)

	collisionAt := valuePairToPoint(rayHitInfo.ColliderGridPosition)

	hitEntity := g.currentMap().IsActorAt(collisionAt) || g.currentMap().IsObjectAt(collisionAt)

	flightPath := ArraysToPoints(rayHitInfo.TraversedGridCells)

	if len(flightPath) <= 1 {
		return nil
	}
	// remove start
	flightPath = flightPath[1:]

	projAnim, _ := g.ui.GetAnimProjectileWithTrail(leadIcon, trailColors, flightPath, nil)

	if hitEntity {
		animationsFromEntityHit := hitEntityHandler(collisionAt)
		projAnim.SetFollowUp(animationsFromEntityHit)
	}

	return []foundation.Animation{projAnim}
}

func fireRay(g *GameState, zapper *Actor, aimPos geometry.Point) []foundation.Animation {

	damageRolled := max(1, dice_curve.Spread(10, 0.8))

	trailColors := []string{"White", "Yellow", "LightRed", "Red"}

	hitEntityHandler := func(hitPos geometry.Point) []foundation.Animation {
		damage := SourcedDamage{
			NameOfThing:     "fire ray",
			Attacker:        zapper,
			IsObviousAttack: true,
			TargetingMode:   special.TargetingModeFireSingle,
			DamageType:      special.DamageTypeFire,
			DamageAmount:    damageRolled,
		}
		return g.damageLocation(damage, hitPos)
	}
	if rand.Intn(20) == 0 { // 1 in 20 chance to just bounce off
		bounceCount := rand.Intn(30) + 3
		return g.bouncingRay(zapper, aimPos, bounceCount, ' ', trailColors, hitEntityHandler)
	}
	return g.singleRay(zapper.Position(), aimPos, ' ', trailColors, hitEntityHandler)
}

func lightningRay(g *GameState, zapper *Actor, aimPos geometry.Point) []foundation.Animation {
	damageRolled := max(1, dice_curve.Spread(7, 0.5))

	trailColors := []string{
		"White",
		"Yellow_3",
		"Yellow_3",
		"Yellow_3",
	}
	dontHitThese := make(map[*Actor]bool)
	dontHitThese[zapper] = true

	hitEntityHandler := func(hitPos geometry.Point) []foundation.Animation {
		damage := SourcedDamage{
			NameOfThing:     "lightning ray",
			Attacker:        zapper,
			IsObviousAttack: true,
			TargetingMode:   special.TargetingModeFireSingle,
			DamageType:      special.DamageTypeElectrical,
			DamageAmount:    damageRolled,
		}
		damageAnims := g.damageLocation(damage, hitPos)
		if g.currentMap().IsActorAt(hitPos) {
			dontHitThese[g.currentMap().ActorAt(hitPos)] = true
		}
		return damageAnims
	}

	if rand.Intn(20) == 0 { // 1 in 20 chance to just bounce off
		bounceCount := rand.Intn(30) + 3
		return g.bouncingRay(zapper, aimPos, bounceCount, ' ', trailColors, hitEntityHandler)
	}

	nextTarget := func(curPos geometry.Point) (bool, geometry.Point) {
		nearbyActors := g.nearbyActors(curPos, dontHitThese)

		if len(nearbyActors) == 0 || rand.Intn(2) == 0 {
			g.msg(foundation.Msg("The lightning fizzles out"))
			return false, curPos
		}
		nextTargetActor := nearbyActors[rand.Intn(len(nearbyActors))]
		g.msg(foundation.HiLite("The lightning arcs towards %s", nextTargetActor.Name()))
		dontHitThese[nextTargetActor] = true
		return true, nextTargetActor.Position()
	}
	return g.chainedRay(zapper, aimPos, ' ', trailColors, nextTarget, hitEntityHandler)
}
func (g *GameState) nearbyActors(curPos geometry.Point, exclude map[*Actor]bool) []*Actor {
	return g.currentMap().GetFilteredActorsInRadius(curPos, 10, func(actor *Actor) bool {
		if actor.Position() == curPos {
			return false
		}
		if _, donthit := exclude[actor]; donthit {
			return false
		}
		hasLos := g.currentMap().IsLineOfSightClear(curPos, actor.Position(), g.IsSomethingBlockingTargetingAtLoc)
		if !hasLos {
			return false
		}
		return true
	})
}
func invisibilityTarget(g *GameState, zapper *Actor, targetPos geometry.Point) []foundation.Animation {
	var animations []foundation.Animation

	origin := zapper.Position()
	pathOfFlight := g.getLineOfSight(origin, targetPos)

	targetPos = pathOfFlight[len(pathOfFlight)-1]

	projAnim, _ := g.ui.GetAnimProjectile('*', "light_gray_5", zapper.Position(), targetPos, nil)
	animations = append(animations, projAnim)

	if g.currentMap().IsActorAt(targetPos) {
		targetActor := g.currentMap().ActorAt(targetPos)
		//actorIcon := targetActor.Icon()
		//coverAnim := g.ui.GetAnimCover(targetPos, actorIcon, dist, nil)
		//animations = append(animations, coverAnim)
		makeInvisible(g, targetActor)
	}

	return animations
}

func teleportTargetTo(g *GameState, zapper *Actor, targetPos geometry.Point) []foundation.Animation {
	var animations []foundation.Animation

	origin := zapper.Position()

	freePositions := g.currentMap().GetFreeCellsForDistribution(origin, 1, g.currentMap().CanPlaceActorHere)
	if len(freePositions) == 0 {
		g.msg(foundation.Msg("There is no place to teleport to"))
		return nil
	}
	teleportTargetPos := freePositions[rand.Intn(len(freePositions))]

	pathOfFlight := g.getLineOfSight(origin, targetPos)

	targetPos = pathOfFlight[len(pathOfFlight)-1]

	projAnim, _ := g.ui.GetAnimProjectile('°', "LightCyan", zapper.Position(), targetPos, nil)
	animations = append(animations, projAnim)

	if g.currentMap().IsActorAt(targetPos) {
		targetActor := g.currentMap().ActorAt(targetPos)
		//hitActorIcon := targetActor.Icon()
		//coverAnim := g.ui.GetAnimCover(targetPos, hitActorIcon, dist, nil)
		//animations = append(animations, coverAnim)

		teleportAnim := teleportWithAnimation(g, targetActor, teleportTargetPos)
		projAnim.SetFollowUp([]foundation.Animation{teleportAnim})
	}

	return animations
}

func teleportTargetAway(g *GameState, zapper *Actor, targetPos geometry.Point) []foundation.Animation {
	var animations []foundation.Animation
	origin := originFromZapperOrWall(g, zapper, targetPos)

	var projAnim foundation.Animation
	if origin != targetPos {
		pathOfFlight := g.getLineOfSight(origin, targetPos)

		targetPos = pathOfFlight[len(pathOfFlight)-1]

		projAnim, _ = g.ui.GetAnimProjectile('°', "LightCyan", origin, targetPos, nil)
		animations = append(animations, projAnim)
	}

	if g.currentMap().IsActorAt(targetPos) {
		targetActor := g.currentMap().ActorAt(targetPos)
		//hitActorIcon := targetActor.Icon()
		//coverAnim := g.ui.GetAnimCover(targetPos, hitActorIcon, dist, nil)
		//animations = append(animations, coverAnim)

		teleportAnim := phaseDoor(g, targetActor)
		if projAnim != nil {
			projAnim.SetFollowUp(teleportAnim)
		} else {
			animations = teleportAnim
		}
	}

	return animations
}

func originFromZapperOrWall(g *GameState, zapper *Actor, targetPos geometry.Point) geometry.Point {
	var origin geometry.Point
	if zapper != nil {
		origin = zapper.Position()
	} else {
		origin = targetPos
	}
	return origin
}

func cancelTarget(g *GameState, zapper *Actor, targetPos geometry.Point) []foundation.Animation {
	var animations []foundation.Animation

	origin := zapper.Position()
	pathOfFlight := g.getLineOfSight(origin, targetPos)

	targetPos = pathOfFlight[len(pathOfFlight)-1]

	projAnim, _ := g.ui.GetAnimProjectile('*', "light_gray_5", origin, targetPos, nil)
	animations = append(animations, projAnim)

	if g.currentMap().IsActorAt(targetPos) {
		targetActor := g.currentMap().ActorAt(targetPos)
		cancel(g, targetActor)
	}

	return animations
}
func holdTarget(g *GameState, zapper *Actor, targetPos geometry.Point) []foundation.Animation {
	var animations []foundation.Animation

	origin := originFromZapperOrWall(g, zapper, targetPos)

	pathOfFlight := g.getLineOfSight(origin, targetPos)

	targetPos = pathOfFlight[len(pathOfFlight)-1]

	projAnim, _ := g.ui.GetAnimProjectile('*', "White", origin, targetPos, nil)
	animations = append(animations, projAnim)

	if g.currentMap().IsActorAt(targetPos) {
		targetActor := g.currentMap().ActorAt(targetPos)
		targetActor.GetFlags().Increase(special.FlagHeld, rand.Intn(10)+5)
	}

	return animations
}
func slowTarget(g *GameState, zapper *Actor, targetPos geometry.Point) []foundation.Animation {
	var animations []foundation.Animation

	origin := originFromZapperOrWall(g, zapper, targetPos)

	pathOfFlight := g.getLineOfSight(origin, targetPos)

	targetPos = pathOfFlight[len(pathOfFlight)-1]

	projAnim, _ := g.ui.GetAnimProjectile('*', "light_gray_5", origin, targetPos, nil)
	animations = append(animations, projAnim)

	if g.currentMap().IsActorAt(targetPos) {
		targetActor := g.currentMap().ActorAt(targetPos)
		slow(g, targetActor)
	}

	return animations
}

func hasteTarget(g *GameState, zapper *Actor, targetPos geometry.Point) []foundation.Animation {
	var animations []foundation.Animation

	origin := zapper.Position()
	pathOfFlight := g.getLineOfSight(origin, targetPos)

	targetPos = pathOfFlight[len(pathOfFlight)-1]

	projAnim, _ := g.ui.GetAnimProjectile('*', "light_gray_5", zapper.Position(), targetPos, nil)
	animations = append(animations, projAnim)

	if g.currentMap().IsActorAt(targetPos) {
		targetActor := g.currentMap().ActorAt(targetPos)
		haste(g, targetActor)
	}

	return animations
}

func magicMissile(g *GameState, zapper *Actor, targetPos geometry.Point) []foundation.Animation {
	origin := zapper.Position()
	pathOfFlight := g.getLineOfSight(origin, targetPos)

	targetPos = pathOfFlight[len(pathOfFlight)-1]
	if !g.currentMap().IsTileWalkable(targetPos) && len(pathOfFlight) > 1 {
		targetPos = pathOfFlight[len(pathOfFlight)-2]
	}

	var onHitAnimations []foundation.Animation

	projAnim, _ := g.ui.GetAnimProjectile('°', "LightGreen", zapper.Position(), targetPos, nil)
	damage := SourcedDamage{
		NameOfThing:     "magic missile",
		Attacker:        zapper,
		IsObviousAttack: false,
		TargetingMode:   special.TargetingModeFireSingle,
		DamageType:      special.DamageTypeRadiation,
		DamageAmount:    5,
	}
	damageConsequences := g.damageLocation(damage, targetPos)
	onHitAnimations = append(onHitAnimations, damageConsequences...)

	if projAnim != nil {
		projAnim.SetFollowUp(onHitAnimations)
	}

	return OneAnimation(projAnim)
}

func magicItemProjectile(g *GameState, zapper *Actor, targetPos geometry.Point, itemName, friendlyName string) []foundation.Animation {
	origin := originFromZapperOrWall(g, zapper, targetPos)
	sourceName := nameOfDamageSource(zapper, friendlyName)
	pathOfFlight := g.getLineOfSight(origin, targetPos)

	targetPos = pathOfFlight[len(pathOfFlight)-1]
	if !g.currentMap().IsTileWalkable(targetPos) && len(pathOfFlight) > 1 {
		targetPos = pathOfFlight[len(pathOfFlight)-2]
	}

	var onHitAnimations []foundation.Animation

	dart := g.NewItemFromString(itemName)

	g.addItemToMap(dart, targetPos)

	projAnim, _ := g.ui.GetAnimThrow(dart, origin, targetPos)
	damage := SourcedDamage{
		NameOfThing:     sourceName,
		Attacker:        zapper,
		IsObviousAttack: true,
		TargetingMode:   special.TargetingModeFireSingle,
		DamageType:      special.DamageTypePlasma,
		DamageAmount:    dart.GetThrowDamage().Roll(),
	}
	damageConsequences := g.damageLocation(damage, targetPos)

	onHitAnimations = append(onHitAnimations, damageConsequences...)

	if projAnim != nil {
		projAnim.SetFollowUp(onHitAnimations)
	}

	return OneAnimation(projAnim)
}

func nameOfDamageSource(zapper *Actor, otherName string) string {
	if zapper == nil {
		return otherName
	}
	return zapper.Name()
}

func (g *GameState) damageLocation(damage SourcedDamage, targetPos geometry.Point) []foundation.Animation {
	if g.currentMap().IsActorAt(targetPos) {
		defender := g.currentMap().ActorAt(targetPos)
		return g.damageActor(damage, defender)
	} else if g.currentMap().IsObjectAt(targetPos) {
		object := g.currentMap().ObjectAt(targetPos)
		return object.OnDamage(damage)
	} else if g.currentMap().IsTileWithFlagAt(targetPos, gridmap.TileFlagDestroyable) {
		tileDamageThreshold := 10
		if damage.DamageAmount >= tileDamageThreshold {
			tileAt, exists := g.currentMap().TryGetTileAt(targetPos)
			if exists {
				g.currentMap().SetTile(targetPos, tileAt.Destroyed())
			}
		}
	}
	return nil
}

type SourcedDamage struct {
	NameOfThing     string
	Attacker        *Actor
	IsObviousAttack bool
	TargetingMode   special.TargetingMode
	DamageType      special.DamageType
	DamageAmount    int
	BodyPart        special.BodyPart
}

func (d SourcedDamage) IsActor() bool {
	return d.Attacker != nil
}
func (d SourcedDamage) String() string {
	if d.Attacker != nil {
		return d.Attacker.Name()
	}
	return d.NameOfThing
}

func (g *GameState) damageActorWithFollowUp(
	damage SourcedDamage,
	victim *Actor,
	done func(),
	followUps []foundation.Animation,
) []foundation.Animation {

	didCripple := victim.TakeDamage(damage)

	if damage.IsObviousAttack {
		g.trySetHostile(victim, damage.Attacker)
	}
	isKill := victim.GetHitPoints() <= 0
	isOverKill := victim.GetHitPoints() <= (-victim.GetHitPointsMax() / 2)
	var damageAnim foundation.Animation
	var damageAudioCue string

	hurtFlag := fmt.Sprintf("WasHurt(%s)", victim.GetInternalName())
	g.gameFlags.SetFlag(hurtFlag)

	if damage.Attacker == g.Player {
		hurtByPlayerFlag := fmt.Sprintf("WasHurtByPlayer(%s)", victim.GetInternalName())
		g.gameFlags.SetFlag(hurtByPlayerFlag)
	}

	g.actorHitMessage(victim, damage, didCripple, isKill, isOverKill)

	if isKill {
		g.actorKilled(damage, victim)
		if isOverKill {
			damageAudioCue = victim.GetDeathCriticalAudioCue(damage.TargetingMode, damage.DamageType)
		} else {
			damageAudioCue = victim.GetDeathAudioCue()
		}
		// TODO: replace this with cool matching death animations
		g.makeMapBloody(victim.Position())
		damageAnim = g.ui.GetAnimDamage(g.spreadBloodAround, victim.Position(), damage.DamageAmount, 4, done)
		damageAnim.SetFollowUp(followUps)
	} else { // only a flesh wound
		damageAudioCue = victim.GetHitAudioCue(damage.TargetingMode.IsMelee())

		//
		bullets := 1
		if damage.TargetingMode.IsBurstOrFullAuto() {
			bullets = 3
		}
		damageAnim = g.ui.GetAnimDamage(g.spreadBloodAround, victim.Position(), damage.DamageAmount, bullets, done)
		damageAnim.SetFollowUp(followUps)

		if victim != g.Player && rand.Intn(5) == 0 {
			g.tryAddRandomChatter(victim, foundation.ChatterBeingDamaged)
		}
	}

	damageAnim.SetAudioCue(damageAudioCue)
	return []foundation.Animation{damageAnim}
}

func (g *GameState) trySetHostile(affected *Actor, sourceOfTrouble *Actor) {
	if affected == sourceOfTrouble || sourceOfTrouble == nil || affected == nil {
		return
	}
	if affected.IsAlive() &&
		!affected.IsPanicking() &&
		!affected.IsHostileTowards(sourceOfTrouble) &&
		g.canActorSee(affected, sourceOfTrouble.Position()) {
		affected.SetHostileTowards(sourceOfTrouble)
		affected.SetGoal(GoalKillActor(affected, sourceOfTrouble))
		if sourceOfTrouble == g.Player {
			g.ui.UpdateVisibleActors()
		}
		if affected != g.Player {
			if !affected.GetEquipment().HasWeaponEquipped() && affected.GetInventory().HasWeapon() {
				affected.TryEquipRangedWeaponFirst()
				g.actorReloadMainHandWeapon(affected)
			} else if weapon, hasWeapon := affected.GetEquipment().GetRangedWeapon(); hasWeapon {
				g.ui.PlayCue(weapon.GetWeapon().GetReloadAudioCue())
			}
		}
	}
}

func (g *GameState) damageActor(damage SourcedDamage, victim *Actor) []foundation.Animation {
	return g.damageActorWithFollowUp(damage, victim, nil, nil)
}

func (g *GameState) getLineOfSight(origin geometry.Point, targetPos geometry.Point) []geometry.Point {
	pathOfFlight := geometry.BresenhamLine(origin, targetPos, func(x, y int) bool {
		mapPos := geometry.Point{X: x, Y: y}
		if !g.currentMap().Contains(mapPos) {
			return false
		}
		if origin.X == x && origin.Y == y {
			return true
		}
		return g.currentMap().IsCurrentlyPassable(mapPos)
	})
	if len(pathOfFlight) > 1 {
		// remove start
		pathOfFlight = pathOfFlight[1:]
	}
	return pathOfFlight
}

func (g *GameState) getLine(origin geometry.Point, targetPos geometry.Point) []geometry.Point {
	pathOfFlight := geometry.BresenhamLine(origin, targetPos, func(x, y int) bool {
		mapPos := geometry.Point{X: x, Y: y}
		if !g.currentMap().Contains(mapPos) {
			return false
		}
		if origin.X == x && origin.Y == y {
			return true
		}
		return g.currentMap().Contains(mapPos)
	})
	if len(pathOfFlight) > 1 {
		// remove start
		pathOfFlight = pathOfFlight[1:]
	}
	return pathOfFlight
}
func fireBreath(g *GameState, zapper *Actor, pos geometry.Point, params Params) []foundation.Animation {
	damageAmount := params.GetDamageOrDefault(10)
	origin := zapper.Position()
	pathOfFlight := geometry.BresenhamLine(origin, pos, func(x, y int) bool {
		if origin.X == x && origin.Y == y {
			return true
		}
		return g.currentMap().IsCurrentlyPassable(geometry.Point{X: x, Y: y})
	})
	if len(pathOfFlight) > 1 {
		// remove start
		pathOfFlight = pathOfFlight[1:]
	}
	targetPos := pathOfFlight[len(pathOfFlight)-1]
	if !g.currentMap().IsTileWalkable(targetPos) && len(pathOfFlight) > 1 {
		targetPos = pathOfFlight[len(pathOfFlight)-2]
	}

	neigh := g.currentMap().NeighborsCardinal(targetPos, func(p geometry.Point) bool {
		return g.currentMap().Contains(p)
	})

	breathAnim := g.ui.GetAnimBreath(pathOfFlight, nil)

	var onHitAnimations []foundation.Animation
	for _, hitPos := range append(pathOfFlight, neigh...) {
		damage := SourcedDamage{
			NameOfThing:     "flames",
			Attacker:        zapper,
			IsObviousAttack: true,
			TargetingMode:   special.TargetingModeFireSingle,
			DamageType:      special.DamageTypeFire,
			DamageAmount:    damageAmount,
		}
		damageAnims := g.damageLocation(damage, hitPos)
		onHitAnimations = append(onHitAnimations, damageAnims...)
	}

	if len(breathAnim) > 0 {
		breathAnim[0].SetFollowUp(onHitAnimations)
	}

	return breathAnim
}
func zapEffectExists(zapEffectName string) bool {
	zapEffects := GetAllZapEffects()
	_, ok := zapEffects[zapEffectName]
	return ok
}

func (g *GameState) actorZapItem(zapper *Actor, item *Item, targetPos geometry.Point) []foundation.Animation {
	zapEffectName := item.GetZapEffectName()

	if zapEffectName == "" {
		return nil
	}

	if !g.hasPaidWithCharge(zapper, item) {
		return nil
	}
	parameters := item.GetEffectParameters()
	animations := g.actorInvokeZapEffect(zapper, zapEffectName, targetPos, parameters)

	return animations
}
func (g *GameState) playerZapItemAndEndTurn(item *Item, targetPos geometry.Point) {
	consequences := g.actorZapItem(g.Player, item, targetPos)
	g.ui.AddAnimations(consequences)
	g.endPlayerTurn(g.Player.timeNeededForActions())
}
func (g *GameState) actorInvokeZapEffect(zapper *Actor, zapEffectName string, targetPos geometry.Point, params Params) []foundation.Animation {
	zapFunc := ZapEffectFromName(zapEffectName)
	if zapFunc == nil {
		return nil
	}
	return zapFunc(g, zapper, targetPos, params)
}

func ZapEffectFromName(zapEffectName string) func(g *GameState, zapper *Actor, aimPos geometry.Point, params Params) []foundation.Animation {
	if zapEffectName == "" {
		return nil
	}
	zapEffects := GetAllZapEffects()
	zapFunc, ok := zapEffects[zapEffectName]
	if !ok {
		return nil
	}
	return zapFunc
}

func (g *GameState) playerInvokeZapEffectAndEndTurn(zapEffectName string, targetPos geometry.Point, params Params) {
	consequences := g.actorInvokeZapEffect(g.Player, zapEffectName, targetPos, params)
	g.ui.AddAnimations(consequences)
	g.endPlayerTurn(g.Player.timeNeededForActions())
}

func (g *GameState) bouncingRay(zapper *Actor, aimPos geometry.Point, bounceCount int, lead rune, colors []string, hitEntityHandler func(hitPos geometry.Point) []foundation.Animation) []foundation.Animation {
	origin := zapper.Position()
	aimDirection := aimPos.ToCenteredPointF().Sub(origin.ToCenteredPointF())
	hitinfos := func() []fxtools.HitInfo2D {
		return g.currentMap().ReflectingRayCast(origin.ToCenteredPointF(), aimDirection, bounceCount, g.IsBlockingRay)
	}

	return g.multiRay(lead, colors, hitinfos, hitEntityHandler)
}

func (g *GameState) chainedRay(zapper *Actor, aimPos geometry.Point, lead rune, colors []string, nextTarget func(curPos geometry.Point) (bool, geometry.Point), hitEntityHandler func(hitPos geometry.Point) []foundation.Animation) []foundation.Animation {
	origin := zapper.Position()
	aimDirection := aimPos.ToCenteredPointF().Sub(origin.ToCenteredPointF())
	hitinfos := func() []fxtools.HitInfo2D {
		rayCasts := g.currentMap().ChainedRayCast(origin.ToCenteredPointF(), aimDirection, g.IsBlockingRay, nextTarget)
		return rayCasts
	}

	return g.multiRay(lead, colors, hitinfos, hitEntityHandler)
}

func (g *GameState) multiRay(leadIcon rune, trailColors []string, getHitInfos func() []fxtools.HitInfo2D, hitEntityHandler func(hitPos geometry.Point) []foundation.Animation) []foundation.Animation {
	hitinfos := getHitInfos()

	var rootAnimation foundation.Animation
	var prevAnim foundation.Animation

	for _, rayHitInfo := range hitinfos {
		collisionAt := valuePairToPoint(rayHitInfo.ColliderGridPosition)
		if !g.currentMap().Contains(collisionAt) {
			break
		}
		hitEntity := g.currentMap().IsActorAt(collisionAt) || g.currentMap().IsObjectAt(collisionAt)

		flightPath := ArraysToPoints(rayHitInfo.TraversedGridCells)

		if len(flightPath) <= 1 {
			continue
		}
		// remove start
		flightPath = flightPath[1:]

		projAnim, _ := g.ui.GetAnimProjectileWithTrail(leadIcon, trailColors, flightPath, nil)

		if hitEntity {
			animationsFromEntityHit := hitEntityHandler(collisionAt)
			projAnim.SetFollowUp(animationsFromEntityHit)
		}

		if rootAnimation == nil {
			rootAnimation = projAnim
		} else {
			prevAnim.SetFollowUp(OneAnimation(projAnim))
		}

		prevAnim = projAnim
	}

	return OneAnimation(rootAnimation)
}

func ArraysToPoints(cells [][2]int64) []geometry.Point {
	points := make([]geometry.Point, len(cells))
	for i, cell := range cells {
		points[i] = valuePairToPoint(cell)
	}
	return points
}

func valuePairToPoint(values [2]int64) geometry.Point {
	return geometry.Point{X: int(values[0]), Y: int(values[1])}
}
