package foundation

import (
	"maps"
	"strings"
)

type ActorFlag int

const (
	FlagNone ActorFlag = iota
	FlagGold

	// Temporary Status Flags
	FlagSleep
	FlagStun
	FlagSlow
	FlagFly
	FlagHaste
	FlagRunning
	FlagSneaking
	FlagHeld
	FlagOpenCarry

	FlagCancel

	FlagBlind
	FlagConfused
	FlagActiveCamouflage
	FlagCurseStuck
	FlagHallucinating
	FlagKnockedDown
	FlagUnconscious
	FlagIgnoresCrime
	FlagDistracted

	// Bookkeeping Flags
	FlagHunger
	FlagStarving
	FlagTurnsSinceEating
	FlagConcentratedAiming
	FlagTurnsSinceLastIdleChatter
	FlagMinorCrimeWarningsGiven

	// Permanent Status Flags
	FlagZombie
	FlagAnimal
	FlagRobot
	FlagChase
	FlagSpawnDead

	// Perks
	FlagSlowDigestion
	FlagSeeFood
	FlagSeeMonsters
	FlagSeeTraps
	FlagSeeMagic
	FlagSeeInvisible
	FlagRegenerating

	FlagCount
)

func (f ActorFlag) String() string { // Nice strings for display
	switch f {
	case FlagSleep:
		return "Sleep"
	case FlagHunger:
		return "Hunger"
	case FlagStarving:
		return "Starving"
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
	case FlagGold:
		return "Gold"
	case FlagCancel:
		return "Cancel"
	case FlagBlind:
		return "Blind"
	case FlagConfused:
		return "Confused"
	case FlagActiveCamouflage:
		return "Active Camouflage"
	case FlagSeeFood:
		return "See Food"
	case FlagSeeMonsters:
		return "See Monsters"
	case FlagSeeMagic:
		return "See Magic"
	case FlagSeeTraps:
		return "See Traps"
	case FlagSeeInvisible:
		return "See Invisible"
	case FlagCurseStuck:
		return "Curse of sticking"
	case FlagHallucinating:
		return "Hallucinating"
	case FlagSlowDigestion:
		return "Slow Digestion"
	case FlagKnockedDown:
		return "Knocked Down"
	case FlagZombie:
		return "Zombie"
	case FlagAnimal:
		return "Animal"
	case FlagRunning:
		return "Running"
	case FlagNone:
		return "None"
	case FlagTurnsSinceEating:
		return "Turns Since Eating"
	case FlagChase:
		return "Chase"
	case FlagTurnsSinceLastIdleChatter:
		return "Turns Since Last Idle Chatter"
	case FlagConcentratedAiming:
		return "Concentrated Aiming"
	case FlagCount:
		return "Count"
	case FlagSneaking:
		return "Sneaking"
	case FlagIgnoresCrime:
		return "Ignores Crime"
	case FlagOpenCarry:
		return "Open Carry"
	}
	return "Unknown"
}

func (f ActorFlag) StringShort() string { // short abbreviated strings (2-3 letters)
	switch f {
	case FlagSleep:
		return "Slp"
	case FlagHunger:
		return "Hng"
	case FlagStarving:
		return "Sta"
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
	case FlagChase:
		return "Chs"
	case FlagGold:
		return "Gld"
	case FlagCancel:
		return "Cnl"
	case FlagBlind:
		return "Bld"
	case FlagConfused:
		return "Cnf"
	case FlagActiveCamouflage:
		return "ACm"
	case FlagSeeFood:
		return "SFd"
	case FlagSeeMonsters:
		return "SMn"
	case FlagSeeTraps:
		return "STr"
	case FlagSeeMagic:
		return "SMg"
	case FlagSeeInvisible:
		return "SIn"
	case FlagCurseStuck:
		return "Stk"
	case FlagHallucinating:
		return "Hlc"
	case FlagSlowDigestion:
		return "SDg"
	case FlagKnockedDown:
		return "Knd"
	case FlagZombie:
		return "Zmb"
	case FlagAnimal:
		return "Anm"
	case FlagRunning:
		return "Run"
	case FlagConcentratedAiming:
		return "CAm"
	case FlagSneaking:
		return "Sne"
	case FlagOpenCarry:
		return "OCr"
	case FlagIgnoresCrime:
		return "IgC"
	case FlagRobot:
		return "Rbt"
	}
	return "Unk"

}

func (f ActorFlag) ShowInHud() bool {
	switch f {
	case FlagSleep:
		return true
	case FlagHunger:
		return true
	case FlagStarving:
		return true
	case FlagStun:
		return true
	case FlagSlow:
		return true
	case FlagHaste:
		return true
	case FlagHeld:
		return true
	case FlagFly:
		return true
	case FlagRegenerating:
		return true
	case FlagGold:
		return true
	case FlagCancel:
		return true
	case FlagBlind:
		return true
	case FlagConfused:
		return true
	case FlagActiveCamouflage:
		return true
	case FlagSeeFood:
		return true
	case FlagSeeMonsters:
		return true
	case FlagSeeTraps:
		return true
	case FlagSeeMagic:
		return true
	case FlagSeeInvisible:
		return true
	case FlagCurseStuck:
		return true
	case FlagHallucinating:
		return true
	case FlagSlowDigestion:
		return true
	case FlagKnockedDown:
		return true
	case FlagZombie:
		return true
	case FlagAnimal:
		return true
	case FlagRunning:
		return true
	case FlagSneaking:
		return true
	case FlagConcentratedAiming:
		return true
	case FlagOpenCarry:
		return true
	}
	return false
}

func ActorFlagFromString(flag string) ActorFlag {
	flag = strings.ToLower(strings.TrimSpace(flag))
	switch flag {
	case "sleep":
		return FlagSleep
	case "hunger":
		return FlagHunger
	case "starving":
		return FlagStarving
	case "stun":
		return FlagStun
	case "slow":
		return FlagSlow
	case "haste":
		return FlagHaste
	case "held":
		return FlagHeld
	case "sneaking":
		return FlagSneaking
	case "flying":
		return FlagFly
	case "regenerating":
		return FlagRegenerating
	case "gold":
		return FlagGold
	case "cancel":
		return FlagCancel
	case "blind":
		return FlagBlind
	case "confused":
		return FlagConfused
	case "invisible":
		return FlagActiveCamouflage
	case "see_food":
		return FlagSeeFood
	case "see_monsters":
		return FlagSeeMonsters
	case "see_magic":
		return FlagSeeMagic
	case "see_invisible":
		return FlagSeeInvisible
	case "see_traps":
		return FlagSeeTraps
	case "chase":
		return FlagChase
	case "curse_stuck":
		return FlagCurseStuck
	case "hallucinating":
		return FlagHallucinating
	case "slow_digestion":
		return FlagSlowDigestion
	case "knocked_down":
		return FlagKnockedDown
	case "is_zombie":
		return FlagZombie
	case "ignores_crime":
		return FlagIgnoresCrime
	case "is_animal":
		return FlagAnimal
	case "is_robot":
		return FlagRobot
	case "concentrated_aiming":
		return FlagConcentratedAiming
	case "running":
		return FlagRunning
	case "spawn_dead":
		return FlagSpawnDead
	}
	panic("Invalid actor flag: " + flag)
	return 0

}

type ActorFlags struct {
	Values  map[ActorFlag]int
	changed func(flag ActorFlag, value int)
}

func NewActorFlags() *ActorFlags {
	return &ActorFlags{Values: make(map[ActorFlag]int)}
}

func (m *ActorFlags) Set(flag ActorFlag) {
	m.Values[flag] = 1
	m.onChange(flag, m.Values[flag])
}

func (m *ActorFlags) Unset(flag ActorFlag) {
	delete(m.Values, flag)
	m.onChange(flag, 0)
}

func (m *ActorFlags) IsSet(flag ActorFlag) bool {
	_, ok := m.Values[flag]
	return ok
}

func (m *ActorFlags) Increment(flag ActorFlag) {
	m.Values[flag]++
	m.onChange(flag, m.Values[flag])
}

func (m *ActorFlags) Decrement(flag ActorFlag) {
	if !m.IsSet(flag) {
		return
	}
	m.Values[flag]--
	if m.Values[flag] <= 0 {
		delete(m.Values, flag)
		m.onChange(flag, 0)
	} else {
		m.onChange(flag, m.Values[flag])
	}
}

func (m *ActorFlags) SetOnChangeHandler(change func(flag ActorFlag, value int)) {
	m.changed = change
}

func (m *ActorFlags) onChange(flag ActorFlag, value int) {
	if m.changed != nil {
		m.changed(flag, value)
	}
}

func (m *ActorFlags) Get(flag ActorFlag) int {
	val, ok := m.Values[flag]
	if !ok {
		return 0
	}
	return val
}

func (m *ActorFlags) Decrease(flag ActorFlag, amount int) {
	m.Values[flag] = m.Get(flag) - amount
	if m.Values[flag] <= 0 {
		delete(m.Values, flag)
		m.onChange(flag, 0)
	} else {
		m.onChange(flag, m.Values[flag])
	}
}

func (m *ActorFlags) Increase(flag ActorFlag, amount int) {
	m.Values[flag] = m.Get(flag) + amount
	m.onChange(flag, m.Values[flag])
}

func (m *ActorFlags) UnderlyingCopy() map[ActorFlag]int {
	return maps.Clone[map[ActorFlag]int](m.Values)
}

func (m *ActorFlags) Init(underlying map[ActorFlag]int) {
	m.Values = underlying
}

func (m *ActorFlags) SetFlagTo(flag ActorFlag, value int) {
	m.Values[flag] = value
	m.onChange(flag, value)
}
