package gridmap

import (
	"RogueUI/foundation"
	"encoding/binary"
	"fmt"
	"io"
)

type Tile struct {
	Feature            foundation.TileType // we need this
	DefinedDescription string
	IsWalkable         bool // this
	IsTransparent      bool // this
}

func (t Tile) ToBinary(out io.Writer) {
	// we want to serialize the tile
	// icon, iswalkable, istransparent

	must(binary.Write(out, binary.LittleEndian, t.Feature))
	must(binary.Write(out, binary.LittleEndian, t.IsWalkable))
	must(binary.Write(out, binary.LittleEndian, t.IsTransparent))
}

func NewTileFromBinary(in io.Reader) Tile {
	var icon foundation.TileType
	var isWalkable bool
	var isTransparent bool

	must(binary.Read(in, binary.LittleEndian, &icon))
	must(binary.Read(in, binary.LittleEndian, &isWalkable))
	must(binary.Read(in, binary.LittleEndian, &isTransparent))

	return Tile{
		Feature:       icon,
		IsWalkable:    isWalkable,
		IsTransparent: isTransparent,
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func (t Tile) Icon() foundation.TileType {
	return t.Feature
}

func (t Tile) Description() string {
	return t.DefinedDescription
}

func (t Tile) EncodeAsString() string {
	return fmt.Sprintf("%c: %s", t.Feature, t.DefinedDescription)
}

func (t Tile) WithIsWalkable(isWalkable bool) Tile {
	t.IsWalkable = isWalkable
	return t
}

func (t Tile) WithIcon(icon foundation.TileType) Tile {
	t.Feature = icon
	return t
}

func (t Tile) WithIsTransparent(value bool) Tile {
	t.IsTransparent = value
	return t
}

func (t Tile) IsTree() bool {
	return t.Feature == foundation.TileTree
}

func (t Tile) IsWater() bool {
	return t.Feature == foundation.TileWater
}

func (t Tile) IsLand() bool {
	return !t.IsWater() && !t.IsVoid()
}

func (t Tile) IsMountain() bool {
	return t.Feature == foundation.TileMountain
}

func (t Tile) IsVoid() bool {
	return t.Feature == foundation.TileEmpty
}

func (t Tile) IsStairsUp() bool {
	return t.Feature == foundation.TileStairsUp
}

func (t Tile) IsStairsDown() bool {
	return t.Feature == foundation.TileStairsDown
}

func (t Tile) IsChasm() bool {
	return t.Feature == foundation.TileChasm
}

func (t Tile) IsStairs() bool {
	return t.IsStairsUp() || t.IsStairsDown()
}

func (t Tile) IsLava() bool {
	return t.Feature == foundation.TileLava
}

func (t Tile) IsSpecial() bool {
	return t.IsStairs() || t.IsChasm() || t.IsLava() || t.IsVoid() || t.IsWater() || t.IsMountain() || t.IsTree() || t.IsDoor()
}

func (t Tile) IsDoor() bool {
	return t.Feature == foundation.TileDoorClosed || t.Feature == foundation.TileDoorOpen
}

func (t Tile) IsVendor() bool {
	return t.Feature == foundation.TileVendorGeneral || t.Feature == foundation.TileVendorWeapons || t.Feature == foundation.TileVendorArmor || t.Feature == foundation.TileVendorAlchemist
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
