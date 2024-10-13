package foundation

import (
    "github.com/memmaker/go/geometry"
    "github.com/memmaker/go/textiles"
    "strings"
)

type ObjectForUI interface {
    GetCategory() ObjectCategory
    Position() geometry.Point
    Icon() textiles.TextIcon
}
type ObjectCategory int

const (
    ObjectTrap ObjectCategory = iota
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
    ObjectPushBox
    ObjectExplodingPushBox
)

func (o ObjectCategory) String() string {
    switch o {
    case ObjectTrap:
        return "Trap"
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
    case ObjectPushBox:
        return "Push Box"
    case ObjectExplodingPushBox:
        return "Exploding Push Box"
    default:
        return "Unknown"
    }
}

func ObjectCategoryFromString(s string) ObjectCategory {
    s = strings.ToLower(s)
    switch s {
    case "trap":
        return ObjectTrap
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
    case "pushbox":
        return ObjectPushBox
    case "explodingpushbox":
        return ObjectExplodingPushBox
    default:
        return -1
    }
}

func (o ObjectCategory) IsTrap() bool {
    return o >= ObjectTrap && o <= ObjectTrap
}

func (o ObjectCategory) LowerString() string {
    // reversal of ObjectCategoryFromString
    switch o {
    case ObjectTrap:
        return "trap"
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
    case ObjectPushBox:
        return "pushbox"
    case ObjectExplodingPushBox:
        return "explodingpushbox"
    default:
        return ""
    }
}
