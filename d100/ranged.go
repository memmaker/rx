package d100

func NewRangedModsFromSituation(mods []Modifier) RangedModifiers {
	return RangedModifiers(mods)
}

type RangedModifiers Modifiers

func (r RangedModifiers) WithRangeMods(distance int, weaponRange int, fullAuto bool, bulletCount int) RangedModifiers {
	mods := r
	isClose := false
	isMedium := false

	switch {
	case distance <= 1:
		mods = appendIfNonZero(r, DefaultRangedMod(RngModRangePointBlank))
		isClose = true
		isMedium = true
	case distance <= (weaponRange / 4):
		delta := (weaponRange - distance) / 2
		mods = appendIfNonZero(r, DefaultModifier{
			Source:    RngModRangeNearPerMeter.String(),
			Modifier:  definedRangedModifiers[RngModRangeNearPerMeter] * delta,
			IsPercent: true,
		})
		isClose = true
		isMedium = true
	case distance <= (weaponRange / 2):
		delta := (weaponRange - distance) / 2
		mods = appendIfNonZero(r, DefaultModifier{
			Source:    RngModRangeNearPerMeter.String(),
			Modifier:  definedRangedModifiers[RngModRangeNearPerMeter] * delta,
			IsPercent: true,
		})
		isMedium = true
	case distance <= weaponRange:
		delta := (weaponRange - distance) / 2
		mods = appendIfNonZero(r, DefaultModifier{
			Source:    RngModRangeNearPerMeter.String(),
			Modifier:  definedRangedModifiers[RngModRangeNearPerMeter] * delta,
			IsPercent: true,
		})
	default:
		delta := (distance - weaponRange) / 2
		mods = appendIfNonZero(r, DefaultModifier{
			Source:    RngModRangeFarPerMeter.String(),
			Modifier:  definedRangedModifiers[RngModRangeFarPerMeter] * delta,
			IsPercent: true,
		})
	}

	if !fullAuto && bulletCount == 3 && isMedium {
		mods = appendIfNonZero(r, DefaultRangedMod(RngModThreeRoundBurstMediumRange))
	}

	if fullAuto {
		if isClose {
			mods = appendIfNonZero(r, DefaultModifier{
				Source:    RngModFullAutoCloseRangePerTenBullets.String(),
				Modifier:  definedRangedModifiers[RngModFullAutoCloseRangePerTenBullets] * (bulletCount / 10),
				IsPercent: true,
			})
		} else {
			mods = appendIfNonZero(r,
				DefaultModifier{
					Source:    RngModFullAutoFarRangePerTenBullets.String(),
					Modifier:  definedRangedModifiers[RngModFullAutoFarRangePerTenBullets] * (bulletCount / 10),
					IsPercent: true,
				})
		}
	}
	return mods
}

func DefaultRangedMod(mod RangedAttackModifier) Modifier {
	return DefaultModifier{
		Source:    mod.String(),
		Modifier:  definedRangedModifiers[mod],
		IsPercent: true,
	}
}

func (r RangedModifiers) With(modActive RangedAttackModifier) RangedModifiers {
	if modActive == 0 {
		return r
	}
	return appendIfNonZero(r, DefaultRangedMod(modActive))
}

func (r RangedModifiers) WithAimTurns(turns int) RangedModifiers {
	return appendIfNonZero(r, DefaultModifier{
		Source:    "Aiming",
		Modifier:  definedRangedModifiers[RngModPerAimTurn] * turns,
		IsPercent: true,
	})
}

func (r RangedModifiers) WithObstacles(obstacles int) RangedModifiers {
	return appendIfNonZero(r, DefaultModifier{
		Source:    "Obstacles",
		Modifier:  definedRangedModifiers[RngModPerObstacle] * obstacles,
		IsPercent: true,
	})
}

func (r RangedModifiers) WithSpeedMods(speedDelta int) RangedModifiers {
	return appendIfNonZero(r, DefaultModifier{
		Source:    "Target Speed",
		Modifier:  definedRangedModifiers[RngModPerTargetVelocityDelta] * speedDelta,
		IsPercent: true,
	})
}

type RangedAttackModifier uint32

const (
	RngModRangePointBlank RangedAttackModifier = 1 << iota
	RngModRangeNearPerMeter
	RngModRangeFarPerMeter

	RngModSizeVerySmall
	RngModSizeSmall
	RngModSizeNormal
	RngModSizeLarge
	RngModSizeVeryLarge

	RngModLightDim
	RngModLightDarkness

	RngModTargetVelocityStationary
	RngModPerTargetVelocityDelta
	RngModTargetCamouflaged

	RngModPerObstacle
	RngModPerAimTurn
	RngModFullAutoCloseRangePerTenBullets
	RngModFullAutoFarRangePerTenBullets

	RngModThreeRoundBurstMediumRange
	RngModFastDraw
	RngModAmbush
	RngModCalledShot
	RngModFiringWhileRunning
	RngModSurprised
	RngModLaserSight
	RngModTelescopicSight
	RngModTargetingScope
	RngModSmartgun
	RngModSmartgoggles
)

func (m RangedAttackModifier) String() string {
	switch m {
	case RngModRangePointBlank:
		return "Point Blank"
	case RngModRangeNearPerMeter:
		return "Distance"
	case RngModRangeFarPerMeter:
		return "Distance"
	case RngModSizeVerySmall:
		return "Very Small Target"
	case RngModSizeSmall:
		return "Small Target"
	case RngModSizeLarge:
		return "Large Target"
	case RngModSizeVeryLarge:
		return "Very Large Target"
	case RngModLightDim:
		return "Dim Light"
	case RngModLightDarkness:
		return "Darkness"
	case RngModTargetVelocityStationary:
		return "Stationary Target"
	case RngModPerTargetVelocityDelta:
		return "Very Fast Target"
	case RngModTargetCamouflaged:
		return "Camouflaged Target"
	case RngModPerObstacle:
		return "Obstacles"
	case RngModPerAimTurn:
		return "Aiming"
	case RngModFastDraw:
		return "Fast Draw"
	case RngModSurprised:
		return "Surprised Target"
	case RngModAmbush:
		return "Ambush"
	case RngModCalledShot:
		return "Called Shot"
	case RngModFiringWhileRunning:
		return "Firing While Running"
	case RngModLaserSight:
		return "Laser Sight"
	case RngModTelescopicSight:
		return "Telescopic Sight"
	case RngModTargetingScope:
		return "Targeting Scope"
	case RngModSmartgun:
		return "Smartgun"
	case RngModSmartgoggles:
		return "Smartgoggles"
	case RngModThreeRoundBurstMediumRange:
		return "Three Round Burst Up to Medium Range"
	case RngModFullAutoCloseRangePerTenBullets:
		return "Full Auto Close Range"
	case RngModFullAutoFarRangePerTenBullets:
		return "Full Auto Far Range"
	default:
		return "Unknown Modifier"
	}
}

var definedRangedModifiers = map[RangedAttackModifier]int{
	RngModRangePointBlank:   20,
	RngModRangeNearPerMeter: 1,
	RngModRangeFarPerMeter:  -5,

	RngModSizeVerySmall: -20,
	RngModSizeSmall:     -10,
	RngModSizeNormal:    0,
	RngModSizeLarge:     10,
	RngModSizeVeryLarge: 20,

	RngModLightDim:      -20,
	RngModLightDarkness: -40,

	RngModTargetVelocityStationary: 0,
	RngModSurprised:                5,

	RngModTargetCamouflaged: -60,

	RngModPerTargetVelocityDelta: -4,
	RngModPerObstacle:            -10,
	RngModPerAimTurn:             10,

	RngModFullAutoCloseRangePerTenBullets: 10,
	RngModFullAutoFarRangePerTenBullets:   -10,

	RngModThreeRoundBurstMediumRange: 10,

	RngModFastDraw:           -20,
	RngModAmbush:             20,
	RngModCalledShot:         -20,
	RngModFiringWhileRunning: -20,
	RngModLaserSight:         20,
	RngModTelescopicSight:    20,
	RngModTargetingScope:     20,
	RngModSmartgun:           20,
	RngModSmartgoggles:       20,
}
