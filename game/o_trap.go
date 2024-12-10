package game

import (
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"strings"
)

type TrapState uint8

func (s TrapState) CanTrigger() bool {
	return s == TrapArmed
}

func (s TrapState) CanActivate() bool {
	return s == TrapTriggered || s == TrapArmed || s == TrapDisarmed
}

func (s TrapState) String() string {
	switch s {
	case TrapArmed:
		return "armed"
	case TrapDisarmed:
		return "disarmed"
	case TrapTriggered:
		return "triggered"
	case TrapActivated:
		return "activated"
	}
	return "Unknown"
}

const (
	TrapArmed TrapState = iota
	TrapDisarmed
	TrapTriggered
	TrapActivated
)

type Trap struct {
	*BaseObject
	trigger                 func()
	zapEffect               string
	triggerOnProximity      bool
	placedByPlayer          bool
	explode                 func() []foundation.Animation
	state                   TrapState
	minSkillNeededForDisarm int
	unTrigger               func()
}

func (t *Trap) Name() string {
	baseName := t.BaseObject.Name()

	return fmt.Sprintf("%s (%s)", baseName, t.state.String())
}

func (t *Trap) GetIcon() textiles.TextIcon {
	baseIcon := t.BaseObject.GetIcon()
	if t.UseCustomIcon {
		baseIcon = t.CustomIcon
	}
	if t.state == TrapTriggered {
		return baseIcon.WithFg(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}
	if t.state == TrapDisarmed {
		return baseIcon.WithFg(color.RGBA{R: 0, G: 255, B: 0, A: 255})
	}
	return baseIcon
}
func (t *Trap) ShouldActivate(tickCount int) bool {
	return tickCount == 1 && t.state.CanActivate()
}

func (t *Trap) IsTimerTicking(tickCount int) bool {
	return tickCount <= 1
}

func (t *Trap) String() string {
	return t.DisplayName
}

func (t *Trap) SetPlacedByPlayer() {
	t.placedByPlayer = true
}

func (t *Trap) InitWithGameState(g *GameState) {
	t.trigger = func() {
		if !t.state.CanTrigger() {
			return
		}
		t.state = TrapTriggered
		g.metronome.AddTimed(t, g.currentMapName, func() {
			g.ui.AddAnimations(t.explode())
		})
	}
	t.unTrigger = func() {
		g.metronome.RemoveTimed(t)
	}
	t.explode = func() []foundation.Animation {
		if !t.state.CanActivate() {
			return nil
		}
		t.state = TrapActivated

		g.currentMap().RemoveObject(t)
		zapEffect := ZapEffectFromName(t.zapEffect)
		anims := zapEffect(g, nil, t.Position(), foundation.Params{})

		return anims
	}
}
func (t *Trap) ToRecord() recfile.Record {
	return recfile.Record{
		{Name: "category", Value: t.Category.String()},
		{Name: "position", Value: t.RawPosition.Encode()},
	}
}

func (g *GameState) NewTrap(record recfile.Record, resolver func(objType string) textiles.TextIcon) *Trap {
	trap := &Trap{BaseObject: NewObject(foundation.ObjectTrap, resolver), state: TrapArmed}
	trap.SetHidden(false)
	trap.SetWalkable(true)
	trap.minSkillNeededForDisarm = 10

	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "name":
			trap.InternalName = field.Value
		case "iconoverride":
			trap.CustomIcon = trap.iconForObject(field.Value)
			trap.UseCustomIcon = true
		case "position":
			trap.RawPosition, _ = geometry.NewPointFromEncodedString(field.Value)
		case "description":
			trap.DisplayName = field.Value
		case "triggeronproximity":
			trap.triggerOnProximity = recfile.StrBool(field.Value)
		case "zapeffect":
			trap.zapEffect = field.Value
		case "ishidden":
			trap.SetHidden(recfile.StrBool(field.Value))
		case "minskillneededfordisarm":
			trap.minSkillNeededForDisarm = recfile.StrInt(field.Value)
		}
	}

	trap.InitWithGameState(g)
	return trap
}

func (t *Trap) IsTransparent() bool {
	return true
}

func (t *Trap) IsProximityTriggered() bool {
	return t.triggerOnProximity && t.state == TrapArmed
}

func (t *Trap) OnProximity(actor *Actor) []foundation.Animation {
	if t.state == TrapDisarmed {
		return nil
	}
	t.trigger()
	return nil
}
func (t *Trap) IsHidden() bool {
	return t.Hidden
}
func (t *Trap) OnDamage(damage SourcedDamage) []foundation.Animation {
	return t.explode()
}

func (t *Trap) OnWalkOver(actor *Actor) []foundation.Animation {
	if t.state == TrapDisarmed {
		return nil
	}
	t.trigger()
	return nil
}

func (t *Trap) OnItemPlaced(item foundation.Item) {
	if t.state == TrapDisarmed {
		return
	}
	t.trigger()
}

func (t *Trap) Activate() []foundation.Animation {
	return t.explode()
}

func (t *Trap) Disarm() {
	t.state = TrapDisarmed
	t.unTrigger()
}

func (t *Trap) IsArmedOrTriggered() bool {
	return t.state == TrapArmed || t.state == TrapTriggered
}

func (t *Trap) MinimalSkillNeededForDisarm() int {
	return t.minSkillNeededForDisarm
}

func (t *Trap) IsPlacedByPlayer() bool {
	return t.placedByPlayer
}
