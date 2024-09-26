package foundation

import (
    "bytes"
    "encoding/gob"
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
    case FlagSeeTraps:
        return "See Traps"
    case FlagSeeInvisible:
        return "See Invisible"
    case FlagCanConfuse:
        return "Can Confuse"
    case FlagCurseStuck:
        return "Curse of sticking"
    case FlagCurseTeleportitis:
        return "Curse of teleportation"
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
    case FlagAwareOfPlayer:
        return "Aware of Player"
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
    case FlagSeeTraps:
        return "STr"
    case FlagSeeMagic:
        return "SMg"
    case FlagSeeInvisible:
        return "SIn"
    case FlagCanConfuse:
        return "CCf"
    case FlagCurseStuck:
        return "Stk"
    case FlagCurseTeleportitis:
        return "Tpt"
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
    }
    return "Unk"

}

func (f ActorFlag) ShowInHud() bool {
    switch f {
    case FlagSleep:
        return true
    case FlagHunger:
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
    case FlagWallCrawl:
        return true
    case FlagGold:
        return true
    case FlagScared:
        return true
    case FlagCancel:
        return true
    case FlagBlind:
        return true
    case FlagConfused:
        return true
    case FlagInvisible:
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
    case FlagCanConfuse:
        return true
    case FlagCurseStuck:
        return true
    case FlagCurseTeleportitis:
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
    case FlagConcentratedAiming:
        return true
    }
    return false
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
    FlagRunning
    FlagHeld
    FlagFly
    FlagRegenerating
    FlagWallCrawl
    FlagGold
    FlagScared
    FlagChase
    FlagCancel
    FlagBlind
    FlagConfused
    FlagInvisible
    FlagSeeFood
    FlagSeeMonsters
    FlagSeeTraps
    FlagSeeMagic
    FlagSeeInvisible
    FlagCanConfuse
    FlagCurseStuck
    FlagCurseTeleportitis
    FlagHallucinating
    FlagSlowDigestion
    FlagKnockedDown
    FlagZombie
    FlagAnimal
    FlagConcentratedAiming
    FlagTurnsSinceLastIdleChatter
    FlagCount
)

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
    case "see_traps":
        return FlagSeeTraps
    case "chase":
        return FlagChase
    case "can_confuse":
        return FlagCanConfuse
    case "curse_stuck":
        return FlagCurseStuck
    case "curse_teleportitis":
        return FlagCurseTeleportitis
    case "hallucinating":
        return FlagHallucinating
    case "slow_digestion":
        return FlagSlowDigestion
    case "knocked_down":
        return FlagKnockedDown
    case "is_zombie":
        return FlagZombie
    case "is_animal":
        return FlagAnimal
    case "concentrated_aiming":
        return FlagConcentratedAiming
    case "running":
        return FlagRunning
    }
    panic("Invalid actor flag: " + flag)
    return 0

}

type ActorFlags struct {
    values  map[ActorFlag]int
    changed func(flag ActorFlag, value int)
}

func NewActorFlags() *ActorFlags {
    return &ActorFlags{values: make(map[ActorFlag]int)}
}

func (m *ActorFlags) GobEncode() ([]byte, error) {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(m.values)
    return buf.Bytes(), err
}

func (m *ActorFlags) GobDecode(data []byte) error {
    dec := gob.NewDecoder(bytes.NewReader(data))
    return dec.Decode(&m.values)
}

func (m *ActorFlags) Set(flag ActorFlag) {
    m.values[flag] = 1
    m.onChange(flag, m.values[flag])
}

func (m *ActorFlags) Unset(flag ActorFlag) {
    delete(m.values, flag)
    m.onChange(flag, 0)
}

func (m *ActorFlags) IsSet(flag ActorFlag) bool {
    _, ok := m.values[flag]
    return ok
}

func (m *ActorFlags) Increment(flag ActorFlag) {
    m.values[flag]++
    m.onChange(flag, m.values[flag])
}

func (m *ActorFlags) Decrement(flag ActorFlag) {
    if !m.IsSet(flag) {
        return
    }
    m.values[flag]--
    if m.values[flag] <= 0 {
        delete(m.values, flag)
        m.onChange(flag, 0)
    } else {
        m.onChange(flag, m.values[flag])
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
    val, ok := m.values[flag]
    if !ok {
        return 0
    }
    return val
}

func (m *ActorFlags) Decrease(flag ActorFlag, amount int) {
    m.values[flag] = m.Get(flag) - amount
    if m.values[flag] <= 0 {
        delete(m.values, flag)
        m.onChange(flag, 0)
    } else {
        m.onChange(flag, m.values[flag])
    }
}

func (m *ActorFlags) Increase(flag ActorFlag, amount int) {
    m.values[flag] = m.Get(flag) + amount
    m.onChange(flag, m.values[flag])
}

func (m *ActorFlags) UnderlyingCopy() map[ActorFlag]int {
    return maps.Clone[map[ActorFlag]int](m.values)
}

func (m *ActorFlags) Init(underlying map[ActorFlag]int) {
    m.values = underlying
}

func (m *ActorFlags) SetFlagTo(flag ActorFlag, value int) {
    m.values[flag] = value
    m.onChange(flag, value)
}
