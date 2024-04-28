package foundation

import "math/rand"

type ObjectCategory int

const (
	ObjectExplodingTrap ObjectCategory = iota
	ObjectSlowTrap
	ObjectTeleportTrap
	ObjectDartTrap
	ObjectArrowTrap
	ObjectDescendTrap
	ObjectBearTrap
)

func RandomObjectCategory() ObjectCategory {
	return ObjectCategory(rand.Intn(int(ObjectBearTrap)+1))
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

	default:
		return "Unknown"
	}
}

func ObjectCategoryFromString(s string) ObjectCategory {
	switch s {
	case "ExplodingTrap":
		return ObjectExplodingTrap
	case "SlowTrap":
		return ObjectSlowTrap
	case "TeleportTrap":
		return ObjectTeleportTrap
	case "DartTrap":
		return ObjectDartTrap
	case "ArrowTrap":
		return ObjectArrowTrap
	case "DescendTrap":
		return ObjectDescendTrap
	case "BearTrap":
		return ObjectBearTrap
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
	return true
}
