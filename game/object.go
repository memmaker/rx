package game

import (
	"contractor/foundation"
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
	gob.Register(&Bed{})
	gob.Register(&ItemMaker{})
}

type BaseObject struct {
	RawPosition           geometry.Point
	Category              foundation.ObjectCategory
	CustomIcon            textiles.TextIcon
	iconForObject         func(string) textiles.TextIcon
	InternalName          string
	DisplayName           string
	Walkable              bool
	Hidden                bool
	Transparent           bool
	UseCustomIcon         bool
	PassableForProjectile bool
}

func (b *BaseObject) GetCategory() foundation.ObjectCategory {
	return b.Category
}

func NewObject(icon foundation.ObjectCategory, iconForObject func(objectType string) textiles.TextIcon) *BaseObject {
	return &BaseObject{
		Category:      icon,
		iconForObject: iconForObject,
	}
}
func (b *BaseObject) GetInternalName() string {
	return b.InternalName
}
func (b *BaseObject) Position() geometry.Point {
	return b.RawPosition
}

func (b *BaseObject) SetPosition(pos geometry.Point) {
	b.RawPosition = pos
}
func (b *BaseObject) OnDamage(damage SourcedDamage) []foundation.Animation {
	return nil
}
func (b *BaseObject) OnWalkOver(actor *Actor) []foundation.Animation {
	return nil
}
func (b *BaseObject) OnProximity(actor *Actor) []foundation.Animation {
	return nil
}
func (b *BaseObject) IsWalkable(actor *Actor) bool {
	return b.Walkable
}

func (b *BaseObject) IsTransparent() bool {
	return b.Transparent
}
func (b *BaseObject) IsPassableForProjectile() bool {
	return b.PassableForProjectile
}
func (b *BaseObject) IsProximityTriggered() bool {
	return false
}

func (b *BaseObject) SetWalkable(isWalkable bool) {
	b.Walkable = isWalkable
}

func (b *BaseObject) IsHidden() bool {
	return b.Hidden
}

func (b *BaseObject) SetHidden(isHidden bool) {
	b.Hidden = isHidden
}

func (b *BaseObject) Name() string {
	if b.DisplayName != "" {
		return b.DisplayName
	}
	return b.Category.String()
}

func (b *BaseObject) IsTrap() bool {
	return b.Category.IsTrap()
}
func (b *BaseObject) IsBed() bool {
	return b.Category == foundation.ObjectBed
}
func (b *BaseObject) OnBump(actor *Actor) {

}

func (b *BaseObject) SetTransparent(transparent bool) {
	b.Transparent = transparent
}

func (b *BaseObject) SetDisplayName(name string) {
	b.DisplayName = name
}

func (b *BaseObject) GetIcon() textiles.TextIcon {
	if b.UseCustomIcon {
		return b.CustomIcon
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
	IsTrap() bool
	IsBed() bool
	IsProximityTriggered() bool
	OnDamage(dmg SourcedDamage) []foundation.Animation
	OnWalkOver(actor *Actor) []foundation.Animation
	OnProximity(actor *Actor) []foundation.Animation
	OnBump(actor *Actor)
	GetIcon() textiles.TextIcon
	AppendContextActions(items []foundation.MenuItem, g *GameState) []foundation.MenuItem
	SetIconResolver(object func(objectType string) textiles.TextIcon)
	InitWithGameState(g *GameState)
	GetInternalName() string
}
