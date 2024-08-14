package foundation

import (
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"math/rand"
	"strings"
)

type ObjectForUI interface {
	GetCategory() ObjectCategory
	Position() geometry.Point
	GetIcon() textiles.TextIcon
}
type ObjectCategory int

const (
	ObjectExplodingTrap ObjectCategory = iota
	ObjectSlowTrap
	ObjectTeleportTrap
	ObjectDartTrap
	ObjectArrowTrap
	ObjectDescendTrap
	ObjectBearTrap
	ObjectTerminal
	ObjectUnknownContainer
	ObjectKnownContainer
	ObjectKnownEmptyContainer
	ObjectLockedDoor
	ObjectClosedDoor
	ObjectOpenDoor
	ObjectBrokenDoor
	ObjectReadable
	ObjectElevator
)

func RandomObjectCategory() ObjectCategory {
	return ObjectCategory(rand.Intn(int(ObjectBearTrap) + 1))
}

func GetAllTrapCategories() []ObjectCategory {
	return []ObjectCategory{
		ObjectExplodingTrap,
		ObjectSlowTrap,
		ObjectTeleportTrap,
		ObjectDartTrap,
		ObjectArrowTrap,
		ObjectDescendTrap,
		ObjectBearTrap,
	}
}

func (o ObjectCategory) String() string {
	switch o {
	case ObjectExplodingTrap:
		return "Exploding Trap"
	case ObjectSlowTrap:
		return "Slow Trap"
	case ObjectTeleportTrap:
		return "Teleport Trap"
	case ObjectDartTrap:
		return "Dart Trap"
	case ObjectArrowTrap:
		return "Arrow Trap"
	case ObjectDescendTrap:
		return "Descend Trap"
	case ObjectBearTrap:
		return "Bear Trap"
	case ObjectLockedDoor:
		return "Locked Door"
	case ObjectClosedDoor:
		return "Closed Door"
	case ObjectOpenDoor:
		return "Open Door"
	case ObjectBrokenDoor:
		return "Broken Door"
	case ObjectTerminal:
		return "Terminal"
	case ObjectKnownContainer:
		return "Known Container"
	case ObjectKnownEmptyContainer:
		return "Known Empty Container"
	case ObjectUnknownContainer:
		return "Unknown Container"
	case ObjectReadable:
		return "Readable"
	case ObjectElevator:
		return "Elevator"
	default:
		return "Unknown"
	}
}

func ObjectCategoryFromString(s string) ObjectCategory {
	s = strings.ToLower(s)
	switch s {
	case "explodingtrap":
		return ObjectExplodingTrap
	case "slowtrap":
		return ObjectSlowTrap
	case "teleporttrap":
		return ObjectTeleportTrap
	case "darttrap":
		return ObjectDartTrap
	case "arrowtrap":
		return ObjectArrowTrap
	case "descendtrap":
		return ObjectDescendTrap
	case "beartrap":
		return ObjectBearTrap
	case "knowncontainer":
		return ObjectKnownContainer
	case "knownemptycontainer":
		return ObjectKnownEmptyContainer
	case "unknowncontainer":
		return ObjectUnknownContainer
	case "lockeddoor":
		return ObjectLockedDoor
	case "closeddoor":
		return ObjectClosedDoor
	case "opendoor":
		return ObjectOpenDoor
	case "brokendoor":
		return ObjectBrokenDoor
	case "readable":
		return ObjectReadable
	case "terminal":
		return ObjectTerminal
	case "elevator":
		return ObjectElevator
	default:
		return -1
	}
}

func (o ObjectCategory) ZapEffect() string {
	switch o {
	case ObjectExplodingTrap:
		return "explode"
	case ObjectSlowTrap:
		return "slow_target"
	case ObjectTeleportTrap:
		return "teleport_target_away"
	case ObjectDartTrap:
		return "magic_dart"
	case ObjectArrowTrap:
		return "magic_arrow"
	case ObjectDescendTrap:
		return "force_descend_target"
	case ObjectBearTrap:
		return "hold_target"
	default:
		return ""
	}
}

func (o ObjectCategory) IsTrap() bool {
	return o >= ObjectExplodingTrap && o <= ObjectBearTrap
}

func (o ObjectCategory) LowerString() string {
	// reversal of ObjectCategoryFromString
	switch o {
	case ObjectExplodingTrap:
		return "explodingtrap"
	case ObjectSlowTrap:
		return "slowtrap"
	case ObjectTeleportTrap:
		return "teleporttrap"
	case ObjectDartTrap:
		return "darttrap"
	case ObjectArrowTrap:
		return "arrowtrap"
	case ObjectDescendTrap:
		return "descendtrap"
	case ObjectBearTrap:
		return "beartrap"
	case ObjectKnownContainer:
		return "knowncontainer"
	case ObjectKnownEmptyContainer:
		return "knownemptycontainer"
	case ObjectUnknownContainer:
		return "unknowncontainer"
	case ObjectLockedDoor:
		return "lockeddoor"
	case ObjectClosedDoor:
		return "closeddoor"
	case ObjectOpenDoor:
		return "opendoor"
	case ObjectBrokenDoor:
		return "brokendoor"
	case ObjectReadable:
		return "readable"
	case ObjectTerminal:
		return "terminal"
	case ObjectElevator:
		return "elevator"
	default:
		return ""
	}
}
