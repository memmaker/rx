package special

import "github.com/memmaker/go/fxtools"

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

type DamageType int

const (
    Physical DamageType = iota
    Energy
)

type AttackMode int

const (
    SingleShot AttackMode = iota
    AimedShot
    BurstShot
    Thrown
    Swing
    Thrust
    Punch
    Placed // for traps
)

type Attack struct {
    mode            AttackMode
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
