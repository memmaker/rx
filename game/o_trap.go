package game

import (
    "RogueUI/foundation"
    "bytes"
    "encoding/gob"
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

func (t *Trap) Icon() textiles.TextIcon {
    if t.state == TrapTriggered {
        return t.customIcon.WithFg(color.RGBA{R: 255, G: 0, B: 0, A: 255})
    }
    if t.state == TrapDisarmed {
        return t.customIcon.WithFg(color.RGBA{R: 0, G: 255, B: 0, A: 255})
    }
    return t.BaseObject.Icon()
}
func (t *Trap) ShouldActivate(tickCount int) bool {
    return tickCount == 1 && t.state.CanActivate()
}

func (t *Trap) IsTimerTicking(tickCount int) bool {
    return tickCount <= 1
}

func (t *Trap) String() string {
    return t.displayName
}

func (t *Trap) SetPlacedByPlayer() {
    t.placedByPlayer = true
}

func (t *Trap) InitWithGameState(g *GameState) {
    t.iconForObject = g.iconForObject
    t.trigger = func() {
        if !t.state.CanTrigger() {
            return
        }
        t.state = TrapTriggered
        g.metronome.AddTimed(t, true, func() {
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
        {Name: "category", Value: t.category.String()},
        {Name: "position", Value: t.position.Encode()},
    }
}

func (t *Trap) GobEncode() ([]byte, error) {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)

    if err := t.BaseObject.gobEncode(enc); err != nil {
        return nil, err
    }

    if err := enc.Encode(t.state); err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}

func (t *Trap) GobDecode(data []byte) error {
    dec := gob.NewDecoder(bytes.NewReader(data))

    t.BaseObject = &BaseObject{}

    if err := t.BaseObject.gobDecode(dec); err != nil {
        return err
    }
    if err := dec.Decode(&t.state); err != nil {
        return err
    }

    return nil
}
func (g *GameState) NewTrap(record recfile.Record) *Trap {
    trap := &Trap{BaseObject: NewObject(foundation.ObjectTrap, g.iconForObject), state: TrapArmed}
    trap.SetHidden(false)
    trap.SetWalkable(true)
    trap.minSkillNeededForDisarm = 10

    for _, field := range record {
        switch strings.ToLower(field.Name) {
        case "name":
            trap.customIcon = g.iconForObject(field.Value)
            trap.useCustomIcon = true
        case "position":
            trap.position, _ = geometry.NewPointFromEncodedString(field.Value)
        case "description":
            trap.displayName = field.Value
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
    return t.isHidden
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
