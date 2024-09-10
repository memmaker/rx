package gridmap

import (
	"encoding/binary"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"io"
)

type ColoredIcon struct {
	Rune   rune
	Fg, Bg string
}
type TileFlags uint8

const TileNoFlags TileFlags = 0

const ( // bitwise flags
	TileFlagHazardous TileFlags = 1 << iota
	TileFlagWater
	TileFlagRadiated
	TileFlagMountable
	TileFlagCrawlable
)

func (t TileFlags) Has(tag TileFlags) bool {
	return t&tag != 0
}

func (t TileFlags) With(tag TileFlags) TileFlags {
	return t | tag
}

type Tile struct {
	Icon               textiles.TextIcon
	DefinedDescription string
	IsWalkable         bool // this
	IsTransparent      bool // this

	Flags TileFlags
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

func (t Tile) ToRecord() recfile.Record {
	return recfile.Record{
		recfile.Field{Name: "Icon", Value: string(t.Icon.Char)},
		recfile.Field{Name: "Fg", Value: recfile.RGBStr(t.Icon.Fg)},
		recfile.Field{Name: "Bg", Value: recfile.RGBStr(t.Icon.Bg)},
		recfile.Field{Name: "IsWalkable", Value: recfile.BoolStr(t.IsWalkable)},
		recfile.Field{Name: "IsTransparent", Value: recfile.BoolStr(t.IsTransparent)},
		recfile.Field{Name: "Flags", Value: recfile.Int64Str(int64(t.Flags))},
	}
}

func NewTileFromRecord(record recfile.Record) Tile {
	tile := Tile{}
	for _, field := range record {
		switch field.Name {
		case "Icon":
			tile.Icon.Char = []rune(field.Value)[0]
		case "Fg":
			tile.Icon.Fg = field.AsRGB(",")
		case "Bg":
			tile.Icon.Bg = field.AsRGB(",")
		case "IsWalkable":
			tile.IsWalkable = field.AsBool()
		case "IsTransparent":
			tile.IsTransparent = field.AsBool()
		case "Flags":
			tile.Flags = TileFlags(field.AsInt())
		}
	}
	return tile
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
	TileType      Tile
	IsExplored    bool
	Actor         *ActorType
	DownedActor   *ActorType
	Item          *ItemType
	Object        *ObjectType
	BakedLighting fxtools.HDRColor
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
