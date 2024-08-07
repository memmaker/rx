package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type BaseObject struct {
	position        geometry.Point
	category        foundation.ObjectCategory
	onDamage        func(actor *Actor) []foundation.Animation
	onWalkOver      func(actor *Actor) []foundation.Animation
	isAlive         bool
	isDrawn         bool
	isWalkable      bool
	isHidden        bool
	triggerOnDamage bool
	onBump          func(actor *Actor)
	internalName    string
	displayName     string
	isTransparent   bool
	icon            textiles.TextIcon
}

func (b *BaseObject) GetIcon() textiles.TextIcon {
	return b.icon
}

func (b *BaseObject) GetCategory() foundation.ObjectCategory {
	return b.category
}

func (g *GameState) NewTrap(trapType foundation.ObjectCategory) *BaseObject {
	trap := NewObject(trapType)
	triggerEffect := func(actor *Actor) []foundation.Animation {
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

func (g *GameState) NewTerminal(rec recfile.Record, palette textiles.ColorPalette) *BaseObject {
	terminal := NewObject(foundation.ObjectTerminal)
	terminal.SetWalkable(false)
	terminal.SetHidden(false)
	terminal.SetTransparent(true)

	var scriptName string
	var icon textiles.TextIcon

	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "dialogue":
			scriptName = field.Value
		case "icon":
			icon.Char = field.AsRune()
		case "foreground":
			icon.Fg = palette.Get(field.Value)
		case "background":
			icon.Bg = palette.Get(field.Value)
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			terminal.SetPosition(spawnPos)
		}
	}
	terminal.internalName = scriptName
	terminal.icon = icon

	terminal.onBump = func(actor *Actor) {
		if actor == g.Player {
			g.StartDialogue(scriptName, true)
		}
	}
	return terminal
}
func NewObject(icon foundation.ObjectCategory) *BaseObject {
	return &BaseObject{
		category: icon,
		isAlive:  true,
		isDrawn:  true,
	}
}

func (b *BaseObject) SetOnDamage(onDamage func(actor *Actor) []foundation.Animation) {
	b.onDamage = onDamage
}
func (b *BaseObject) Position() geometry.Point {
	return b.position
}

func (b *BaseObject) SetPosition(pos geometry.Point) {
	b.position = pos
}
func (b *BaseObject) OnDamage(actor *Actor) []foundation.Animation {
	if b.onDamage != nil && b.triggerOnDamage {
		return b.onDamage(actor)
	}
	return nil
}
func (b *BaseObject) OnWalkOver(actor *Actor) []foundation.Animation {
	if b.onWalkOver != nil {
		return b.onWalkOver(actor)
	}
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

func (b *BaseObject) SetDrawOnMap(drawOnMap bool) {
	b.isDrawn = drawOnMap
}

func (b *BaseObject) IsDrawn() bool {
	return b.isDrawn && !b.isHidden
}

func (b *BaseObject) SetOnWalkOver(handler func(actor *Actor) []foundation.Animation) {
	b.onWalkOver = handler
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
	if b.onBump != nil {
		b.onBump(actor)
	}
}

func (b *BaseObject) SetTransparent(transparent bool) {
	b.isTransparent = transparent
}

func (b *BaseObject) SetDisplayName(name string) {
	b.displayName = name
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
	IsDrawn() bool
	IsTrap() bool
	OnDamage(actor *Actor) []foundation.Animation
	OnWalkOver(actor *Actor) []foundation.Animation
	OnBump(actor *Actor)
	GetIcon() textiles.TextIcon
}
