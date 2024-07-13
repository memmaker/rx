package foundation

import "strings"

type EquipSlot int

func (n EquipSlot) ToString() string {
	switch n {
	case SlotNameMainHand:
		return "Wielding"
	case SlotNameOffHand:
		return "Off Hand"
	case SlotNameLightSource:
		return "Light source"
	case SlotNameAmulet:
		return "Around neck"
	case SlotNameRing:
		return "Ring"
	case SlotNameRingRight:
		return "On right hand"
	case SlotNameRingLeft:
		return "On left hand"
	case SlotNameArmorTorso:
		return "On body"
	case SlotNameArmorHead:
		return "On head"
	case SlotNameArmorFeet:
		return "On feet"
	case SlotNameArmorHands:
		return "On hands"
	case SlotNameArmorBack:
		return "On back"
	}
	return "Unknown"
}

func (n EquipSlot) IsArmorSlot() bool {
	switch n {
	case SlotNameArmorTorso, SlotNameArmorHead, SlotNameArmorFeet, SlotNameArmorHands, SlotNameArmorBack, SlotNameShield:
		return true
	}
	return false
}

const (
	SlotNameNotEquippable EquipSlot = iota

	// On Body & Items

	SlotNameArmorTorso
	SlotNameArmorHead
	SlotNameArmorHands
	SlotNameArmorFeet
	SlotNameArmorBack
	SlotNameAmulet
	SlotNameLightSource

	// On Items only
	SlotNameShield
	SlotNameRing

	// On Body only
	SlotNameMainHand
	SlotNameOffHand
	SlotNameRingRight
	SlotNameRingLeft
)

func ItemSlotFromString(s string) EquipSlot {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "light_source":
		return SlotNameLightSource
	case "amulet":
		return SlotNameAmulet
	case "ring":
		return SlotNameRing
	case "torso":
		return SlotNameArmorTorso
	case "helmet":
		return SlotNameArmorHead
	case "boots":
		return SlotNameArmorFeet
	case "gloves":
		return SlotNameArmorHands
	case "cape":
		return SlotNameArmorBack
	}
	panic("Invalid slot: " + s)
	return SlotNameNotEquippable
}

type ItemTags uint32

func (t ItemTags) Contains(tag ItemTags) bool {
	return t&tag != 0
}

const (
	TagNone   ItemTags = 0
	TagNoLoot ItemTags = 1 << iota
)

func ItemTagFromString(s string) ItemTags {
	s = strings.ToLower(s)
	switch s {
	case "no_loot":
		return TagNoLoot
	}
	panic("Unknown item tag: " + s)
	return TagNone
}
