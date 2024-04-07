package foundation

import "strings"

func ActorFlagFromString(flag string) uint32 {
	flag = strings.ToLower(strings.TrimSpace(flag))
	switch flag {
	case "blind":
		return IsBlind
	case "see_monsters":
		return SeeMonsters
	case "running":
		return IsScared
	case "found":
		return IsFound
	case "invisible":
		return IsInvisible
	case "mean":
		return IsMean
	case "greedy":
		return IsGreedy
	case "held":
		return IsHeld
	case "confused":
		return IsConfused
	case "regenerating":
		return IsRegenerating
	case "can_confuse":
		return CanConfuse
	case "can_see_invisible":
		return CanSeeInvisible
	case "cancelled":
		return IsCancelled
	case "slow":
		return IsSlow
	case "hasted":
		return IsHasted
	case "flying":
		return IsFlying
	case "mimic":
		return IsMimic
	case "wall_crawler":
		return IsWallCrawler

	}
	panic("Invalid actor flag: " + flag)
	return 0

}

func ActorStatusToString(status uint32) string {
	switch status {
	case IsBlind:
		return "Blind"
	case SeeMonsters:
		return "See Monsters"
	case IsScared:
		return "Running"
	case IsFound:
		return "Found"
	case IsInvisible:
		return "Invisible"
	case IsMean:
		return "Mean"
	case IsGreedy:
		return "Greedy"
	case IsHeld:
		return "Held"
	case IsConfused:
		return "Confused"
	case IsRegenerating:
		return "Regenerating"
	case CanConfuse:
		return "Can Confuse"
	case CanSeeInvisible:
		return "Can See Invisible"
	case IsCancelled:
		return "Cancelled"
	case IsSlow:
		return "Slow"
	case IsHasted:
		return "Hasted"
	case IsFlying:
		return "Flying"
	}
	return "Unknown"
}

func ActorStatusToAbbreviation(status uint32) string {
	switch status {
	case IsBlind:
		return "Bl"
	case SeeMonsters:
		return "See"
	case IsScared:
		return "Run"
	case IsFound:
		return "Fnd"
	case IsInvisible:
		return "Inv"
	case IsMean:
		return "Me"
	case IsGreedy:
		return "Gr"
	case IsHeld:
		return "Hld"
	case IsConfused:
		return "Cf"
	case IsRegenerating:
		return "Rg"
	case CanConfuse:
		return "CC"
	case CanSeeInvisible:
		return "SI"
	case IsCancelled:
		return "Cld"
	case IsSlow:
		return "Slw"
	case IsHasted:
		return "Hst"
	case IsFlying:
		return "Fly"
	}
	return "Unknown"
}

func ActorStatusToAbbreviationToLongString(flagAbbr string) string {
	// Slw -> Slow, etc.
	switch flagAbbr {
	case "Bl":
		return "Blind"
	case "See":
		return "See Monsters"
	case "Run":
		return "Running"
	case "Fnd":
		return "Found"
	case "Inv":
		return "Invisible"
	case "Me":
		return "Mean"
	case "Gr":
		return "Greedy"
	case "Hld":
		return "Held"
	case "Cf":
		return "Confused"
	case "Rg":
		return "Regenerating"
	case "CC":
		return "Can Confuse"
	case "SI":
		return "Can See Invisible"
	case "Cld":
		return "Cancelled"
	case "Slw":
		return "Slow"
	case "Hst":
		return "Hasted"
	case "Fly":
		return "Flying"
	}
	return "Unknown"
}

const ( // as bitmask
	IsBlind uint32 = 1 << iota
	SeeMonsters
	SeeFood
	SeeMagic
	IsSleeping
	IsScared
	IsFound
	IsInvisible
	IsMean
	IsGreedy
	IsHeld
	IsConfused
	IsRegenerating
	CanConfuse
	CanSeeInvisible
	IsCancelled
	IsSlow
	IsHasted
	IsFlying
	IsStunned
	IsMimic
	IsWallCrawler
)

type Flags struct {
	flags         uint32
	changeHandler func(flag uint32, currentValue bool)
}

func NewFlags() *Flags {
	return &Flags{}
}
func NewFlagsFromValue(value uint32) Flags {
	return Flags{flags: value}
}
func (f *Flags) Set(flag uint32) {
	f.flags |= flag
	f.onStateChanged(flag, true)
}

func (f *Flags) Unset(flag uint32) {
	f.flags &^= flag
	f.onStateChanged(flag, false)
}
func (f *Flags) IsSet(flag uint32) bool {
	return f.flags&flag != 0
}

func (f *Flags) AsStrings(toString func(uint32) string) []string {
	result := make([]string, 0)
	for _, flag := range []uint32{IsBlind, SeeMonsters, IsScared, IsFound, IsInvisible, IsMean, IsGreedy, IsHeld, IsConfused, IsRegenerating, CanConfuse, CanSeeInvisible, IsCancelled, IsSlow, IsHasted, IsFlying} {
		if f.IsSet(flag) {
			result = append(result, toString(flag))
		}
	}
	return result
}

func (f *Flags) Underlying() uint32 {
	return f.flags
}

func (f *Flags) SetOnChangeHandler(change func(flag uint32, currentValue bool)) {
	f.changeHandler = change
}

func (f *Flags) onStateChanged(flag uint32, currentValue bool) {
	if f.changeHandler != nil {
		f.changeHandler(flag, currentValue)
	}
}

func (f *Flags) Init(flags uint32) {
	f.flags = flags
}
