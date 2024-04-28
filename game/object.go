package game

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
)

type Object struct {
	position        geometry.Point
	category        foundation.ObjectCategory
	onDamage        func() []foundation.Animation
	onWalkOver      func() []foundation.Animation
	isAlive         bool
	isDrawn         bool
	isWalkable      bool
	isHidden        bool
	triggerOnDamage bool
}

func (g *GameState) NewTrap(trapType foundation.ObjectCategory) *Object {
	trap := NewObject(trapType)
	triggerEffect := func() []foundation.Animation {
		if trap.isAlive {
			trap.isAlive = false
			zapEffect := ZapEffectFromName(trapType.ZapEffect())
			consequences := zapEffect(g, nil, trap.Position())
			return consequences
		}
		return nil
	}

	trap.SetHidden(true)
	trap.SetOnDamage(triggerEffect)
	trap.SetOnWalkOver(triggerEffect)
	trap.SetWalkable(true)
	return trap
}
func NewObject(icon foundation.ObjectCategory) *Object {
	return &Object{
		category: icon,
		isAlive:  true,
		isDrawn:  true,
	}
}

func (b *Object) SetOnDamage(onDamage func() []foundation.Animation) {
	b.onDamage = onDamage
}
func (b *Object) Position() geometry.Point {
	return b.position
}

func (b *Object) ObjectIcon() foundation.ObjectCategory {
	return b.category
}

func (b *Object) SetPosition(pos geometry.Point) {
	b.position = pos
}
func (b *Object) OnDamage() []foundation.Animation {
	if b.onDamage != nil && b.triggerOnDamage {
		return b.onDamage()
	}
	return nil
}
func (b *Object) OnWalkOver() []foundation.Animation {
	if b.onWalkOver != nil {
		return b.onWalkOver()
	}
	return nil
}
func (b *Object) IsWalkable(actor *Actor) bool {
	return b.isWalkable
}

func (b *Object) IsTransparent() bool {
	return false
}
func (b *Object) IsPassableForProjectile() bool {
	return false
}

func (b *Object) IsAlive() bool {
	return b.isAlive
}

func (b *Object) SetDrawOnMap(drawOnMap bool) {
	b.isDrawn = drawOnMap
}

func (b *Object) IsDrawn() bool {
	return b.isDrawn && !b.isHidden
}

func (b *Object) SetOnWalkOver(handler func() []foundation.Animation) {
	b.onWalkOver = handler
}

func (b *Object) SetWalkable(isWalkable bool) {
	b.isWalkable = isWalkable
}

func (b *Object) IsHidden() bool {
	return b.isHidden
}

func (b *Object) SetHidden(isHidden bool) {
	b.isHidden = isHidden
}

func (b *Object) Name() string {
	return b.category.String()
}

func (b *Object) IsTrap() bool {
	return b.category.IsTrap()
}
