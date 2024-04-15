package foundation

import (
	"maps"
	"strings"
)

type ActorFlag int

func (f ActorFlag) String() string { // Nice strings for display
	switch f {
	case FlagSleep:
		return "Sleep"
	case FlagHunger:
		return "Hunger"
	case FlagStun:
		return "Stun"
	case FlagSlow:
		return "Slow"
	case FlagHaste:
		return "Haste"
	case FlagHeld:
		return "Held"
	case FlagFly:
		return "Fly"
	case FlagRegenerating:
		return "Regenerating"
	case FlagWallCrawl:
		return "Wall Crawl"
	case FlagGold:
		return "Gold"
	case FlagMean:
		return "Mean"
	case FlagScared:
		return "Scared"
	case FlagCancel:
		return "Cancel"
	case FlagBlind:
		return "Blind"
	case FlagConfused:
		return "Confused"
	case FlagInvisible:
		return "Invisible"
	case FlagSeeFood:
		return "See Food"
	case FlagSeeMonsters:
		return "See Monsters"
	case FlagSeeMagic:
		return "See Magic"
	case FlagSeeInvisible:
		return "See Invisible"
	case FlagCanConfuse:
		return "Can Confuse"

	}
	return "Unknown"
}

func (f ActorFlag) StringShort() string { // short abbreviated strings (2-3 letters)
	switch f {
	case FlagSleep:
		return "Slp"
	case FlagHunger:
		return "Hng"
	case FlagStun:
		return "Stn"
	case FlagSlow:
		return "Slw"
	case FlagHaste:
		return "Hst"
	case FlagHeld:
		return "Hld"
	case FlagFly:
		return "Fly"
	case FlagRegenerating:
		return "Reg"
	case FlagWallCrawl:
		return "WCr"
	case FlagChase:
		return "Chs"
	case FlagGold:
		return "Gld"
	case FlagMean:
		return "Mn"
	case FlagScared:
		return "Scd"
	case FlagCancel:
		return "Cnl"
	case FlagBlind:
		return "Bld"
	case FlagConfused:
		return "Cnf"
	case FlagInvisible:
		return "Inv"
	case FlagSeeFood:
		return "SFd"
	case FlagSeeMonsters:
		return "SMn"
	case FlagSeeMagic:
		return "SMg"
	case FlagSeeInvisible:
		return "SIn"
	case FlagCanConfuse:
		return "CCf"

	}
	return "Unk"

}

// Monster can start off awake or asleep
// sleepy monsters can be mean, which means they might wake up if you come near or not
// when an awake monster sees the player it will get the aware of player flag
const (
	FlagNone ActorFlag = iota
	FlagSleep
	FlagAwareOfPlayer
	FlagHunger
	FlagTurnsSinceEating
	FlagStun
	FlagSlow
	FlagHaste
	FlagHeld
	FlagFly
	FlagRegenerating
	FlagWallCrawl
	FlagGold
	FlagMean
	FlagScared
	FlagChase
	FlagCancel
	FlagBlind
	FlagConfused
	FlagInvisible
	FlagSeeFood
	FlagSeeMonsters
	FlagSeeMagic
	FlagSeeInvisible
	FlagCanConfuse
)

func AllFlagsExceptGoldOrdered() []ActorFlag {
	return []ActorFlag{
		FlagSleep,
		FlagHunger,
		FlagStun,
		FlagSlow,
		FlagHaste,
		FlagHeld,
		FlagChase,
		FlagFly,
		FlagRegenerating,
		FlagWallCrawl,
		FlagMean,
		FlagScared,
		FlagCancel,
		FlagBlind,
		FlagConfused,
		FlagInvisible,
		FlagSeeFood,
		FlagSeeMonsters,
		FlagSeeMagic,
		FlagSeeInvisible,
		FlagCanConfuse,
	}
}

func ActorFlagFromString(flag string) ActorFlag {
	flag = strings.ToLower(strings.TrimSpace(flag))
	switch flag {
	case "sleep":
		return FlagSleep
	case "hunger":
		return FlagHunger
	case "stun":
		return FlagStun
	case "slow":
		return FlagSlow
	case "haste":
		return FlagHaste
	case "held":
		return FlagHeld
	case "flying":
		return FlagFly
	case "regenerating":
		return FlagRegenerating
	case "wall_crawler":
		return FlagWallCrawl
	case "gold":
		return FlagGold
	case "mean":
		return FlagMean
	case "scared":
		return FlagScared
	case "cancel":
		return FlagCancel
	case "blind":
		return FlagBlind
	case "confused":
		return FlagConfused
	case "invisible":
		return FlagInvisible
	case "see_food":
		return FlagSeeFood
	case "see_monsters":
		return FlagSeeMonsters
	case "see_magic":
		return FlagSeeMagic
	case "see_invisible":
		return FlagSeeInvisible
	case "chase":
		return FlagChase
	case "can_confuse":
		return FlagCanConfuse

	}
	panic("Invalid actor flag: " + flag)
	return 0

}

type MapFlags struct {
	values  map[ActorFlag]int
	changed func(flag ActorFlag, value int)
}

func NewMapFlags() *MapFlags {
	return &MapFlags{values: make(map[ActorFlag]int)}
}

func (m *MapFlags) Set(flag ActorFlag) {
	m.values[flag] = 1
	m.onChange(flag, m.values[flag])
}

func (m *MapFlags) Unset(flag ActorFlag) {
	delete(m.values, flag)
	m.onChange(flag, 0)
}

func (m *MapFlags) IsSet(flag ActorFlag) bool {
	_, ok := m.values[flag]
	return ok
}

func (m *MapFlags) Increment(flag ActorFlag) {
	m.values[flag]++
	m.onChange(flag, m.values[flag])
}

func (m *MapFlags) Decrement(flag ActorFlag) {
	m.values[flag]--
	if m.values[flag] <= 0 {
		delete(m.values, flag)
		m.onChange(flag, 0)
	} else {
		m.onChange(flag, m.values[flag])
	}
}

func (m *MapFlags) SetOnChangeHandler(change func(flag ActorFlag, value int)) {
	m.changed = change
}

func (m *MapFlags) onChange(flag ActorFlag, value int) {
	if m.changed != nil {
		m.changed(flag, value)
	}
}

func (m *MapFlags) Get(flag ActorFlag) int {
	val, ok := m.values[flag]
	if !ok {
		return 0
	}
	return val
}

func (m *MapFlags) Decrease(flag ActorFlag, amount int) {
	m.values[flag] = m.Get(flag) - amount
	if m.values[flag] <= 0 {
		delete(m.values, flag)
		m.onChange(flag, 0)
	} else {
		m.onChange(flag, m.values[flag])
	}
}

func (m *MapFlags) Increase(flag ActorFlag, amount int) {
	m.values[flag] = m.Get(flag) + amount
	m.onChange(flag, m.values[flag])
}

func (m *MapFlags) UnderlyingCopy() map[ActorFlag]int {
	return maps.Clone[map[ActorFlag]int](m.values)
}

func (m *MapFlags) Init(underlying map[ActorFlag]int) {
	m.values = underlying
}
