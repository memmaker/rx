package game

import (
	"RogueUI/foundation"
	"encoding/gob"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
)

func init() {
	// This is a hack to make sure that the foundation package is imported
	// so that the gob.Register function is called
	gob.Register(&Terminal{})
	gob.Register(&ReadableObject{})
	gob.Register(&Door{})
	gob.Register(&Trap{})
	gob.Register(&Elevator{})
	gob.Register(&Container{})
	gob.Register(&PushBox{})
}

type BaseObject struct {
	position      geometry.Point
	category      foundation.ObjectCategory
	customIcon    textiles.TextIcon
	iconForObject func(string) textiles.TextIcon
	internalName  string
	displayName   string
	isWalkable    bool
	isHidden      bool
	isTransparent bool
	useCustomIcon bool
	isAlive       bool
}

func (b *BaseObject) gobEncode(enc *gob.Encoder) error {
	if err := enc.Encode(b.position); err != nil {
		return err
	}

	if err := enc.Encode(b.category); err != nil {
		return err
	}

	if err := enc.Encode(b.customIcon); err != nil {
		return err
	}

	if err := enc.Encode(b.internalName); err != nil {
		return err
	}

	if err := enc.Encode(b.displayName); err != nil {
		return err
	}

	if err := enc.Encode(b.isWalkable); err != nil {
		return err
	}

	if err := enc.Encode(b.isHidden); err != nil {
		return err
	}

	if err := enc.Encode(b.isTransparent); err != nil {
		return err
	}

	if err := enc.Encode(b.useCustomIcon); err != nil {
		return err
	}

	if err := enc.Encode(b.isAlive); err != nil {
		return err
	}

	return nil
}

func (b *BaseObject) gobDecode(dec *gob.Decoder) error {
	if err := dec.Decode(&b.position); err != nil {
		return err
	}

	if err := dec.Decode(&b.category); err != nil {
		return err
	}

	if err := dec.Decode(&b.customIcon); err != nil {
		return err
	}

	if err := dec.Decode(&b.internalName); err != nil {
		return err
	}

	if err := dec.Decode(&b.displayName); err != nil {
		return err
	}

	if err := dec.Decode(&b.isWalkable); err != nil {
		return err
	}

	if err := dec.Decode(&b.isHidden); err != nil {
		return err
	}

	if err := dec.Decode(&b.isTransparent); err != nil {
		return err
	}

	if err := dec.Decode(&b.useCustomIcon); err != nil {
		return err
	}

	if err := dec.Decode(&b.isAlive); err != nil {
		return err
	}

	return nil
}

func (b *BaseObject) GetCategory() foundation.ObjectCategory {
	return b.category
}

func NewObject(icon foundation.ObjectCategory, iconForObject func(objectType string) textiles.TextIcon) *BaseObject {
	return &BaseObject{
		category:      icon,
		iconForObject: iconForObject,
		isAlive:       true,
	}
}

func (b *BaseObject) Position() geometry.Point {
	return b.position
}

func (b *BaseObject) SetPosition(pos geometry.Point) {
	b.position = pos
}
func (b *BaseObject) OnDamage(damage SourcedDamage) []foundation.Animation {
	return nil
}
func (b *BaseObject) OnWalkOver(actor *Actor) []foundation.Animation {
	return nil
}
func (b *BaseObject) IsWalkable(actor *Actor) bool {
	return b.isWalkable
}

func (b *BaseObject) IsTransparent() bool {
	return b.isTransparent
}
func (b *BaseObject) IsPassableForProjectile() bool {
	return false
}

func (b *BaseObject) IsAlive() bool {
	return b.isAlive
}

func (b *BaseObject) SetWalkable(isWalkable bool) {
	b.isWalkable = isWalkable
}

func (b *BaseObject) IsHidden() bool {
	return b.isHidden
}

func (b *BaseObject) SetHidden(isHidden bool) {
	b.isHidden = isHidden
}

func (b *BaseObject) Name() string {
	if b.displayName != "" {
		return b.displayName
	}
	return b.category.String()
}

func (b *BaseObject) IsTrap() bool {
	return b.category.IsTrap()
}

func (b *BaseObject) OnBump(actor *Actor) {

}

func (b *BaseObject) SetTransparent(transparent bool) {
	b.isTransparent = transparent
}

func (b *BaseObject) SetDisplayName(name string) {
	b.displayName = name
}

func (b *BaseObject) Icon() textiles.TextIcon {
	if b.useCustomIcon {
		return b.customIcon
	}
	return b.iconForObject(b.GetCategory().LowerString())
}

func (b *BaseObject) AppendContextActions(items []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	return items
}

func (b *BaseObject) SetIconResolver(object func(objectType string) textiles.TextIcon) {
	b.iconForObject = object
}

type Object interface {
	Name() string
	GetCategory() foundation.ObjectCategory
	Position() geometry.Point
	SetPosition(pos geometry.Point)
	SetHidden(isHidden bool)
	IsHidden() bool
	IsWalkable(actor *Actor) bool
	IsTransparent() bool
	IsPassableForProjectile() bool
	IsAlive() bool
	IsTrap() bool
	OnDamage(dmg SourcedDamage) []foundation.Animation
	OnWalkOver(actor *Actor) []foundation.Animation
	OnBump(actor *Actor)
	Icon() textiles.TextIcon
	AppendContextActions(items []foundation.MenuItem, g *GameState) []foundation.MenuItem
	SetIconResolver(object func(objectType string) textiles.TextIcon)
	InitWithGameState(g *GameState)
}
