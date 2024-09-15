package special

import (
	"github.com/memmaker/go/fxtools"
	"strings"
)

type TargetingMode int

const (
	TargetingModeNone         TargetingMode = 0
	TargetingModePunch        TargetingMode = 1
	TargetingModeKick         TargetingMode = 2
	TargetingModeSwing        TargetingMode = 4
	TargetingModeThrust       TargetingMode = 8
	TargetingModeThrow        TargetingMode = 16
	TargetingModeFireSingle   TargetingMode = 32
	TargetingModeFireBurst    TargetingMode = 64
	TargetingModeFireFullAuto TargetingMode = 128
	TargetingModeFlame        TargetingMode = 256
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
	}
	panic("Invalid targeting mode: " + value)
	return TargetingModeNone
}
func (t TargetingMode) Next() TargetingMode {
	if t == TargetingModeNone {
		return TargetingModePunch
	}
	nextVal := t << 1
	if nextVal > TargetingModeFlame {
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

type AmmoType int

const (
	Ammo223FMJ AmmoType = iota
	Ammo44MagnumFMJ
	Ammo44MagnumJHP
	Ammo45Caliber
	Ammo2mmEC
	Ammo47mmCaseless
	Ammo5mmAP
	Ammo5mmJHP
	Ammo762mm
	Ammo9mm
	Ammo9mmBall
	Ammo10mmAP
	Ammo10mmJHP
	Ammo14mmAP
	AmmoBBs
	Ammo12Gauge
	AmmoExplosiveRocket
	AmmoRocketAP
	AmmoFlamerFuel
	AmmoFlamerFuelMKII
	AmmoHNNeedler
	AmmoHNNeedlerAP
	AmmoMicroFusionCell
	AmmoSmallEnergyCell
	AmmoSunlight
)

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

type Attack struct {
	mode            TargetingMode
	damage          fxtools.Interval
	damageType      DamageType
	timeNeededInAut int
	maxRange        int
	roundsFired     int
}

type WeaponStats struct {
	attacks         []Attack
	minimumStrength int
	magazineSize    int
	ammoType        AmmoType
	// MISSING: PERKS & CRIT FAIL
}
