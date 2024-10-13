package d100

import "strings"

type MeleeModifiers []Modifier

func ModsToString(modifiers []Modifier) string {
	var mods []string
	for _, mod := range modifiers {
		mods = append(mods, mod.Description())
	}
	return strings.Join(mods, "\n")
}

func appendIfNonZero(mods []Modifier, mod Modifier) []Modifier {
	if mod.IsZero() {
		return mods
	}
	return append(mods, mod)
}

func (r MeleeModifiers) With(modActive MeleeAttackModifier) MeleeModifiers {
	if modActive == 0 {
		return r
	}
	return appendIfNonZero(r, DefaultMeleeMod(modActive))
}

func (r MeleeModifiers) WithAimTurns(turns int) MeleeModifiers {
	return appendIfNonZero(r, DefaultModifier{
		Source:    "Aiming",
		Modifier:  definedMeleeModifiers[MleModPerAimTurn] * turns,
		IsPercent: true,
	})
}

func (r MeleeModifiers) String() string {
	var mods []string
	for _, mod := range r {
		mods = append(mods, mod.Description())
	}
	return strings.Join(mods, "\n")
}

func (r MeleeModifiers) WithSpeedMods(speedDelta int) MeleeModifiers {
	return appendIfNonZero(r, DefaultModifier{
		Source:    "Target Speed",
		Modifier:  definedMeleeModifiers[MleModPerTargetVelocityDelta] * speedDelta,
		IsPercent: true,
	})
}
func DefaultMeleeMod(mod MeleeAttackModifier) Modifier {
	return DefaultModifier{
		Source:    mod.String(),
		Modifier:  definedMeleeModifiers[mod],
		IsPercent: true,
	}
}

type MeleeAttackModifier uint32

const (
	MleModPerStrength MeleeAttackModifier = 1 << iota
	MleModPerTargetVelocityDelta
	MleModPerAimTurn
)

func (m MeleeAttackModifier) String() string {
	switch m {

	default:
		return "Unknown Modifier"
	}
}

var definedMeleeModifiers = map[MeleeAttackModifier]int{
	MleModPerStrength:            10,
	MleModPerTargetVelocityDelta: 10,
	MleModPerAimTurn:             10,
}
