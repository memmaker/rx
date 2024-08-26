package game

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"path"
	"strings"
	"text/template"
)

type BaseObject struct {
	position        geometry.Point
	category        foundation.ObjectCategory
	customIcon      textiles.TextIcon
	onDamage        func(actor SourcedDamage) []foundation.Animation
	onWalkOver      func(actor *Actor) []foundation.Animation
	isAlive         bool
	isDrawn         bool
	isWalkable      bool
	isHidden        bool
	triggerOnDamage bool
	onBump          func(actor *Actor)
	iconForObject   func(string) textiles.TextIcon
	internalName    string
	displayName     string
	isTransparent   bool
	useCustomIcon   bool
	contextActions  []foundation.MenuItem
}

func (b *BaseObject) GetCategory() foundation.ObjectCategory {
	return b.category
}

func (g *GameState) NewTrap(trapType foundation.ObjectCategory, iconForObject func(objectType string) textiles.TextIcon) *BaseObject {
	trap := NewObject(trapType, iconForObject)
	walkOverEffect := func(actor *Actor) []foundation.Animation {
		if trap.isAlive {
			trap.isAlive = false
			zapEffect := ZapEffectFromName(trapType.ZapEffect())
			consequences := zapEffect(g, nil, trap.Position())
			return consequences
		}
		return nil
	}
	damageEffect := func(dmg SourcedDamage) []foundation.Animation {
		return walkOverEffect(dmg.Attacker)
	}

	trap.SetHidden(true)
	trap.SetOnDamage(damageEffect)
	trap.SetOnWalkOver(walkOverEffect)
	trap.SetWalkable(true)
	return trap
}

func (g *GameState) NewTerminal(rec recfile.Record, iconForObject func(objectType string) textiles.TextIcon) *BaseObject {
	terminal := NewObject(foundation.ObjectTerminal, iconForObject)
	terminal.SetWalkable(false)
	terminal.SetHidden(false)
	terminal.SetTransparent(true)

	var scriptName string
	declareAsTerminal := true
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			terminal.customIcon = iconForObject(field.Value)
			terminal.useCustomIcon = true
		case "description":
			terminal.displayName = field.Value
		case "dialogue":
			scriptName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			terminal.SetPosition(spawnPos)
		case "tags":
			if strings.ToLower(field.Value) == "no_sound" {
				declareAsTerminal = false
			}
		}
	}
	terminal.internalName = scriptName

	terminal.onBump = func(actor *Actor) {
		if actor == g.Player {
			g.StartDialogue(scriptName, terminal, declareAsTerminal)
		}
	}
	return terminal
}

func (g *GameState) NewReadable(rec recfile.Record, iconForObject func(objectType string) textiles.TextIcon) *BaseObject {
	sign := NewObject(foundation.ObjectReadable, nil)
	sign.SetWalkable(false)
	sign.SetHidden(false)
	sign.SetTransparent(true)

	var text string
	var customIcon textiles.TextIcon
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			customIcon = iconForObject(field.Value)
		case "description":
			sign.displayName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			sign.SetPosition(spawnPos)
		case "text":
			text = g.fillTemplatedText(strings.TrimSpace(field.Value))
		case "textfile":
			text = fxtools.ReadFile(path.Join(g.config.DataRootDir, "text", field.Value+".txt"))
		}
	}

	showText := func() {
		g.ui.OpenTextWindow(g.fillTemplatedText(text))
	}
	sign.customIcon = customIcon
	sign.useCustomIcon = true
	sign.internalName = "readable"

	sign.onBump = func(actor *Actor) {
		if actor == g.Player {
			showText()
		}
	}
	sign.SetContextActions([]foundation.MenuItem{
		{
			Name:       "Read",
			Action:     showText,
			CloseMenus: true,
		},
	})
	return sign
}

func (g *GameState) fillTemplatedText(text string) string {
	parsedTemplate, err := template.New("text").Parse(text)
	if err != nil {
		panic(err)
	}
	replaceValues := map[string]string{"pcname": g.Player.Name()}

	var filledText strings.Builder
	err = parsedTemplate.Execute(&filledText, replaceValues)
	if err != nil {
		panic(err)
	}
	return filledText.String()
}

func NewObject(icon foundation.ObjectCategory, iconForObject func(objectType string) textiles.TextIcon) *BaseObject {
	return &BaseObject{
		category:      icon,
		isAlive:       true,
		isDrawn:       true,
		iconForObject: iconForObject,
	}
}

func (b *BaseObject) SetOnDamage(onDamage func(damage SourcedDamage) []foundation.Animation) {
	b.onDamage = onDamage
}
func (b *BaseObject) Position() geometry.Point {
	return b.position
}

func (b *BaseObject) SetPosition(pos geometry.Point) {
	b.position = pos
}
func (b *BaseObject) OnDamage(damage SourcedDamage) []foundation.Animation {
	if b.onDamage != nil && b.triggerOnDamage {
		return b.onDamage(damage)
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

func (b *BaseObject) Icon() textiles.TextIcon {
	if b.useCustomIcon {
		return b.customIcon
	}
	return b.iconForObject(b.GetCategory().LowerString())
}

func (b *BaseObject) AppendContextActions(items []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	return append(items, b.contextActions...)
}

func (b *BaseObject) SetContextActions(items []foundation.MenuItem) {
	b.contextActions = items
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
	OnDamage(dmg SourcedDamage) []foundation.Animation
	OnWalkOver(actor *Actor) []foundation.Animation
	OnBump(actor *Actor)
	Icon() textiles.TextIcon
	AppendContextActions(items []foundation.MenuItem, g *GameState) []foundation.MenuItem
}
