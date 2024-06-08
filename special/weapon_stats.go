package special

import (
	"math/rand"
	"regexp"
	"strconv"
)

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

type Interval struct {
	min int
	max int
}

func (i Interval) NotZero() bool {
	return !i.IsZero()
}

func (i Interval) IsZero() bool {
	return i.min == 0 && i.max == 0
}

func (i Interval) ExpectedValue() int {
	return (i.min + i.max) / 2
}

func (i Interval) ShortString() string {
	if i.min == i.max {
		return strconv.Itoa(i.min)
	}
	return strconv.Itoa(i.min) + "-" + strconv.Itoa(i.max)
}

func (i Interval) Roll() int {
	if i.min == i.max {
		return i.min
	}
	return rand.Intn(i.max-i.min) + i.min
}

func NewInterval(min, max int) Interval {
	if min > max {
		min, max = max, min
	}
	return Interval{min, max}
}

func ParseInterval(s string) Interval {
	// looks like "1-6" or possibly "1", "2 - 6"
	pattern := `(\d+)(?:\s*-\s*(\d+))?`
	reg := regexp.MustCompile(pattern)
	matches := reg.FindStringSubmatch(s)
	var minVal, maxVal int
	if len(matches) == 3 {
		minVal = ParseInt(matches[1])
		maxVal = ParseInt(matches[2])
	} else {
		minVal = ParseInt(matches[1])
		maxVal = minVal
	}
	return NewInterval(minVal, maxVal)
}

func ParseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

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
	damage          Interval
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
