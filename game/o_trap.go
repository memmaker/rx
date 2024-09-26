package game

import (
	"RogueUI/foundation"
	"bytes"
	"encoding/gob"
	"github.com/memmaker/go/recfile"
)

type Trap struct {
	*BaseObject
	isLoaded bool
	trigger  func() []foundation.Animation
}

func (t *Trap) InitWithGameState(g *GameState) {
	t.iconForObject = g.iconForObject
	t.trigger = func() []foundation.Animation {
		if t.isLoaded {
			t.isLoaded = false
			zapEffect := ZapEffectFromName(t.category.ZapEffect())
			return zapEffect(g, nil, t.Position(), foundation.Params{})
		}
		return nil
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

	if err := enc.Encode(t.isLoaded); err != nil {
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
	if err := dec.Decode(&t.isLoaded); err != nil {
		return err
	}

	return nil
}
func (g *GameState) NewTrap(trapType foundation.ObjectCategory) *Trap {
	trap := &Trap{BaseObject: NewObject(trapType, g.iconForObject), isLoaded: true}

	trap.SetHidden(true)
	trap.SetWalkable(true)

	trap.InitWithGameState(g)
	return trap
}

func (t *Trap) OnDamage(damage SourcedDamage) []foundation.Animation {
	return t.trigger()
}

func (t *Trap) OnWalkOver(actor *Actor) []foundation.Animation {
	return t.trigger()
}
