package game

import (
	"strings"
)

type TargetingMode int

const (
	TargetingModeNone TargetingMode = iota
	TargetingModePunch
	TargetingModeKick
	TargetingModeSwing
	TargetingModeThrust
	TargetingModeThrow
	TargetingModeFireSingle
	TargetingModeFireBurst
	TargetingModeFireFullAuto
	TargetingModeFlame

	// unique for bowel disruptor
	TargetingModeBDLoose
	TargetingModeBDWatery
	TargetingModeBDFiery
	TargetingModeBDBurningAnalGeyser
	TargetingModeBDRectalVolcano
	TargetingModeBDProlapse
	TargetingModeBDUnspeakableGutHorror
	TargetingModeBDShatIntoUnconsciousness
	TargetingModeBDFatalIntestinalMaelstrom

	TargetingModeCount
)

func TargetingModeFromString(value string) TargetingMode {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "none":
		return TargetingModeNone
	case "punch":
		return TargetingModePunch
	case "kick":
		return TargetingModeKick
	case "swing":
		return TargetingModeSwing
	case "thrust":
		return TargetingModeThrust
	case "throw":
		return TargetingModeThrow
	case "fire_single":
		return TargetingModeFireSingle
	case "fire_burst":
		return TargetingModeFireBurst
	case "flame":
		return TargetingModeFlame

	case "bd_loose":
		return TargetingModeBDLoose
	case "bd_watery":
		return TargetingModeBDWatery
	case "bd_fiery":
		return TargetingModeBDFiery
	case "bd_burning_anal_geyser":
		return TargetingModeBDBurningAnalGeyser
	case "bd_rectal_volcano":
		return TargetingModeBDRectalVolcano
	case "bd_prolapse":
		return TargetingModeBDProlapse
	case "bd_unspeakable_gut_horror":
		return TargetingModeBDUnspeakableGutHorror
	case "bd_shat_into_unconsciousness":
		return TargetingModeBDShatIntoUnconsciousness
	case "bd_fatal_intestinal_maelstrom":
		return TargetingModeBDFatalIntestinalMaelstrom
	}
	panic("Invalid targeting mode: " + value)
	return TargetingModeNone
}
func (t TargetingMode) Next() TargetingMode {
	if t == TargetingModeNone {
		return TargetingModePunch
	}
	nextVal := t + 1
	if nextVal >= TargetingModeCount {
		nextVal = TargetingModePunch
	}
	return nextVal
}

func (t TargetingMode) ToString() string {
	switch t {
	case TargetingModeNone:
		return "None"
	case TargetingModePunch:
		return "Punch"
	case TargetingModeKick:
		return "Kick"
	case TargetingModeSwing:
		return "Swing"
	case TargetingModeThrust:
		return "Thrust"
	case TargetingModeThrow:
		return "Throw"
	case TargetingModeFireSingle:
		return "Fire Single"
	case TargetingModeFireBurst:
		return "Fire Burst"
	case TargetingModeFlame:
		return "Flame"
	case TargetingModeBDLoose:
		return "Loose"
	case TargetingModeBDWatery:
		return "Watery"
	case TargetingModeBDFiery:
		return "Fiery"
	case TargetingModeBDBurningAnalGeyser:
		return "Burning Anal Geyser"
	case TargetingModeBDRectalVolcano:
		return "Rectal Volcano"
	case TargetingModeBDProlapse:
		return "Prolapse"
	case TargetingModeBDUnspeakableGutHorror:
		return "Unspeakable Gut Horror"
	case TargetingModeBDShatIntoUnconsciousness:
		return "Shat Into Unconsciousness"
	case TargetingModeBDFatalIntestinalMaelstrom:
		return "Fatal Intestinal Maelstrom"
	}
	return "Unknown"
}

func (t TargetingMode) IsMelee() bool {
	return t == TargetingModePunch || t == TargetingModeKick || t == TargetingModeSwing || t == TargetingModeThrust || t == TargetingModeThrow
}

func (t TargetingMode) IsAimable() bool {
	return t == TargetingModeFireSingle || t == TargetingModePunch || t == TargetingModeKick || t == TargetingModeSwing || t == TargetingModeThrust || t == TargetingModeThrow
}

func (t TargetingMode) IsBurstOrFullAuto() bool {
	return t == TargetingModeFireBurst || t == TargetingModeFireFullAuto
}

func (t TargetingMode) IsFullAuto() bool {
	return t == TargetingModeFireFullAuto
}

type DamageType int32

func (t DamageType) IsEnergy() bool {
	return t == DamageTypeLaser || t == DamageTypePlasma || t == DamageTypeElectrical || t == DamageTypeEMP

}

func (t DamageType) String() string {
	switch t {
	case DamageTypeNormal:
		return "Normal"
	case DamageTypeLaser:
		return "Laser"
	case DamageTypeFire:
		return "Fire"
	case DamageTypePlasma:
		return "Plasma"
	case DamageTypeElectrical:
		return "Electrical"
	case DamageTypeEMP:
		return "EMP"
	case DamageTypeExplosive:
		return "Explosive"
	case DamageTypeRadiation:
		return "Radiation"
	case DamageTypePoison:
		return "Poison"
	}
	return "Unknown"
}

const (
	DamageTypeNormal DamageType = iota
	DamageTypeLaser
	DamageTypeFire
	DamageTypePlasma
	DamageTypeElectrical
	DamageTypeEMP
	DamageTypeExplosive
	DamageTypeRadiation
	DamageTypePoison
	// DamageTypeEnergy is a catch-all for all energy damage types on armors
	DamageTypeEnergy
	DamageTypeCount
)

func DamageTypeFromString(value string) DamageType {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "normal":
		return DamageTypeNormal
	case "laser":
		return DamageTypeLaser
	case "fire":
		return DamageTypeFire
	case "plasma":
		return DamageTypePlasma
	case "electrical":
		return DamageTypeElectrical
	case "emp":
		return DamageTypeEMP
	case "explosive":
		return DamageTypeExplosive
	case "radiation":
		return DamageTypeRadiation
	case "poison":
		return DamageTypePoison
	}
	panic("Invalid damage type: " + value)
	return DamageTypeNormal
}
