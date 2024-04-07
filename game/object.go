package game

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
)

type Object struct {
	position geometry.Point
	icon     foundation.ObjectCategory
	name     string
	onDamage func() []foundation.Animation
	isAlive  bool
	isDrawn  bool
}

func NewBarrel(icon foundation.ObjectCategory, createExplosion func(atLoc geometry.Point) []foundation.Animation) *Object {
	barrel := NewObject("barrel", icon)
	damage := func() []foundation.Animation {
		if barrel.isAlive {
			barrel.isAlive = false
			consequences := createExplosion(barrel.Position())
			return consequences
		}
		return nil
	}
	barrel.SetOnDamage(damage)
	return barrel
}
func NewObject(name string, icon foundation.ObjectCategory) *Object {
	return &Object{
		icon:    icon,
		name:    name,
		isAlive: true,
		isDrawn: true,
	}
}

func (b *Object) SetOnDamage(onDamage func() []foundation.Animation) {
	b.onDamage = onDamage
}
func (b *Object) Position() geometry.Point {
	return b.position
}

func (b *Object) ObjectIcon() foundation.ObjectCategory {
	return b.icon
}

func (b *Object) SetPosition(pos geometry.Point) {
	b.position = pos
}
func (b *Object) OnDamage() []foundation.Animation {
	if b.onDamage != nil {
		return b.onDamage()
	}
	return nil
}
func (b *Object) IsWalkable(actor *Actor) bool {
	return false
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
	return b.isDrawn
}
