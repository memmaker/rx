package foundation

import "strings"

type EquipSlot int

func (n EquipSlot) ToString() string {
	switch n {
	case SlotNameMainHand:
		return "Wielding"
	case SlotNameLightSource:
		return "Light source"
	case SlotNameArmorTorso:
		return "On body"
	case SlotNameArmorHead:
		return "On head"
	}
	return "Unknown"
}

func (n EquipSlot) IsArmorSlot() bool {
	switch n {
	case SlotNameArmorTorso, SlotNameArmorHead:
		return true
	}
	return false
}

const (
	SlotNameNotEquippable EquipSlot = iota

	// On Body & Items

	SlotNameArmorTorso
	SlotNameArmorHead
	SlotNameLightSource

	// On Body only
	SlotNameMainHand
)

func ItemSlotFromString(s string) EquipSlot {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "light_source":
		return SlotNameLightSource
	case "torso":
		return SlotNameArmorTorso
	case "helmet":
		return SlotNameArmorHead
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
	TagNoAim
	TagNoSound
	TagLightSource
	TagTimed
)

func ItemTagFromString(s string) ItemTags {
	s = strings.ToLower(s)
	switch s {
	case "no_loot":
		return TagNoLoot
	case "no_aim":
		return TagNoAim
	case "no_sound":
		return TagNoSound
	case "light_source":
		return TagLightSource
	case "timed":
		return TagTimed
	}
	panic("Unknown item tag: " + s)
	return TagNone
}
