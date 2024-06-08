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
	isContainer     bool
	containedItems  *[]foundation.ItemForUI
	isAlive         bool
	isDrawn         bool
	isWalkable      bool
	isHidden        bool
	triggerOnDamage bool
	onBump          func(actor *Actor)
	name            string
	isKnown         bool
}

func (b *Object) GetCategory() foundation.ObjectCategory {
	if b.isContainer {
		if b.isKnown {
			if len(*b.containedItems) == 0 {
				return foundation.ObjectKnownEmptyContainer
			} else {
				return foundation.ObjectKnownContainer
			}
		}
		return foundation.ObjectUnknownContainer
	}
	return b.category
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
func (g *GameState) NewContainer(name string, items []*Item) *Object {
	iui := itemsForUI(items)
	container := NewObject(foundation.ObjectUnknownContainer)
	container.isContainer = true
	container.SetWalkable(false)
	container.SetHidden(false)
	itemsUI := &iui
	container.containedItems = itemsUI
	container.name = name

	transferItem := func(itemTaken foundation.ItemForUI) {
		for i, item := range *itemsUI {
			if item == itemTaken {
				*itemsUI = append((*itemsUI)[:i], (*itemsUI)[i+1:]...)
				g.Player.GetInventory().Add(itemTaken.(*Item))
				break
			}
		}
	}
	container.onBump = func(actor *Actor) {
		if actor == g.Player {
			g.msg(foundation.HiLite("You search %s", container.Name()))
			g.ui.ShowContainer(container.Name(), itemsUI, transferItem)
		}
	}
	return container
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
	if b.name != "" {
		return b.name
	}
	return b.category.String()
}

func (b *Object) IsTrap() bool {
	return b.category.IsTrap()
}

func (b *Object) OnBump(actor *Actor) {
	if b.onBump != nil {
		b.onBump(actor)
		if b.isContainer {
			b.isKnown = true
		}
	}
}
