package foundation

import "strings"

type EquipSlot int

func (n EquipSlot) ToString() string {
	switch n {
	case SlotNameMainHand:
		return "Wielding"
	case SlotNameMissileLauncher:
		return "Shooting"
	case SlotNameOffHand:
		return "Off Hand"
	case SlotNameLightSource:
		return "Light source"
	case SlotNameOneHandedWeapon:
		return "One Handed Weapon"
	case SlotNameAmulet:
		return "Around neck"
	case SlotNameRing:
		return "Ring"
	case SlotNameRingRight:
		return "On right hand"
	case SlotNameRingLeft:
		return "On left hand"
	case SlotNameTwoHandedWeapon:
		return "Two Handed Weapon"
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
	case SlotNameQuiver:
		return "Quivered"
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

func (n EquipSlot) IsWeaponSlot() bool {
	switch n {
	case SlotNameOneHandedWeapon, SlotNameTwoHandedWeapon, SlotNameMissileLauncher, SlotNameQuiver, SlotNameShield, SlotNameArmorHands:
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
	SlotNameMissileLauncher
	SlotNameQuiver

	// On Items only
	SlotNameOneHandedWeapon
	SlotNameShield
	SlotNameTwoHandedWeapon
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
	case "missile_launcher":
		return SlotNameMissileLauncher
	case "light_source":
		return SlotNameLightSource
	case "one_handed_weapon":
		return SlotNameOneHandedWeapon
	case "amulet":
		return SlotNameAmulet
	case "ring":
		return SlotNameRing
	case "two_handed_weapon":
		return SlotNameTwoHandedWeapon
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
	case "quiver":
		return SlotNameQuiver
	}
	panic("Invalid slot: " + s)
	return SlotNameNotEquippable
}
