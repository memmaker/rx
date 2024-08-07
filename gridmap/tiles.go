package gridmap

import (
	"encoding/binary"
	"fmt"
	"github.com/memmaker/go/textiles"
	"io"
)

type ColoredIcon struct {
	Rune   rune
	Fg, Bg string
}

type Tile struct {
	Icon               textiles.TextIcon
	DefinedDescription string
	IsWalkable         bool // this
	IsTransparent      bool // this
	IsDamaging         bool
}

func (t Tile) ToBinary(out io.Writer) {
	// we want to serialize the tile
	// icon, iswalkable, istransparent

	must(binary.Write(out, binary.LittleEndian, t.Icon))
	must(binary.Write(out, binary.LittleEndian, t.IsWalkable))
	must(binary.Write(out, binary.LittleEndian, t.IsTransparent))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func (t Tile) Description() string {
	return t.DefinedDescription
}

func (t Tile) EncodeAsString() string {
	return fmt.Sprintf("%c: %s", t.Icon, t.DefinedDescription)
}

func (t Tile) WithIsWalkable(isWalkable bool) Tile {
	t.IsWalkable = isWalkable
	return t
}

func (t Tile) WithIsTransparent(value bool) Tile {
	t.IsTransparent = value
	return t
}

type MapCell[ActorType interface {
	comparable
	MapActor
}, ItemType interface {
	comparable
	MapObject
}, ObjectType interface {
	comparable
	MapObjectWithProperties[ActorType]
}] struct {
	TileType    Tile
	IsExplored  bool
	IsLit       bool // IsLit is true if this tile receives light from a light source and is thus permanently lit if it's explored.
	Actor       *ActorType
	DownedActor *ActorType
	Item        *ItemType
	Object      *ObjectType
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithItemHereRemoved(itemHere ItemType) MapCell[ActorType, ItemType, ObjectType] {
	if c.Item != nil && *c.Item == itemHere {
		c.Item = nil
	}
	return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithObjectHereRemoved(obj ObjectType) MapCell[ActorType, ItemType, ObjectType] {
	if c.Object != nil && *c.Object == obj {
		c.Object = nil
	}
	return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithItemRemoved() MapCell[ActorType, ItemType, ObjectType] {
	c.Item = nil
	return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithObjectRemoved() MapCell[ActorType, ItemType, ObjectType] {
	c.Object = nil
	return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithDownedActor(a ActorType) MapCell[ActorType, ItemType, ObjectType] {
	c.DownedActor = &a
	return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithActor(actor ActorType) MapCell[ActorType, ItemType, ObjectType] {
	c.Actor = &actor
	return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithObject(obj ObjectType) MapCell[ActorType, ItemType, ObjectType] {
	c.Object = &obj
	return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithActorHereRemoved(actorHere ActorType) MapCell[ActorType, ItemType, ObjectType] {
	if c.Actor != nil && *c.Actor == actorHere {
		c.Actor = nil
	}
	return c
}
func (c MapCell[ActorType, ItemType, ObjectType]) WithDownedActorHereRemoved(actorHere ActorType) MapCell[ActorType, ItemType, ObjectType] {
	if c.DownedActor != nil && *c.DownedActor == actorHere {
		c.DownedActor = nil
	}
	return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithItem(item ItemType) MapCell[ActorType, ItemType, ObjectType] {
	c.Item = &item
	return c
}
