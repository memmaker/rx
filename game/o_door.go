package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type Door struct {
	*BaseObject
	lockedFlag            string
	lockDiff              foundation.Difficulty
	lockStrengthRemaining int
	numberLock            []rune

	hitpoints        int
	damageThreshold  int
	audioCueBaseName string
	player           foundation.AudioCuePlayer
	onBump           func(actor *Actor)
	updatePlayerFoV  func()
}

func (b *Door) InitWithGameState(g *GameState) {
	b.iconForObject = g.iconForObject
	b.player = g.ui
	b.updatePlayerFoV = g.updatePlayerFoVAndApplyExploration
	b.onBump = func(actor *Actor) {
		if actor == g.Player && b.GetCategory() == foundation.ObjectLockedDoor {
			if b.lockedFlag != "" && actor.HasKey(b.lockedFlag) {
				b.category = foundation.ObjectClosedDoor
				g.msg(foundation.Msg("You unlocked the door"))
				g.ui.PlayCue("world/PICKKEYS")
				return
			}

			if len(b.numberLock) > 0 {
				g.ui.OpenKeypad(b.numberLock, func(result bool) {
					if result {
						b.category = foundation.ObjectClosedDoor
						g.msg(foundation.Msg("You unlocked the door"))
					}
				})
			} else {
				// lockpicking
				lockPickResult := func(success bool) {
					if success {
						b.category = foundation.ObjectClosedDoor
						g.msg(foundation.Msg("You picked the lock deftly"))
						g.ui.PlayCue("world/PICKKEYS")
					} else {
						g.msg(foundation.Msg("You failed to pick the lock"))
					}
				}
				if g.config.UseLockpickingMiniGame {
					g.ui.StartLockpickGame(b.lockDiff, g.Player.GetInventory().GetLockpickCount, g.Player.GetInventory().RemoveLockpick, func(result foundation.InteractionResult) {
						lockPickResult(result == foundation.Success)
					})
				} else if g.config.UseLockpickingDX {
					if g.Player.GetInventory().GetLockpickCount() == 0 {
						g.msg(foundation.Msg("You don't have any lockpicks"))
						return
					}
					g.Player.GetInventory().RemoveLockpick()
					skill := g.Player.GetCharSheet().GetSkill(special.Lockpick)
					reduction := int(float64(skill) * 0.375)
					if b.PickByReduceStrength(reduction) {
						lockPickResult(true)
					} else {
						remaining := b.lockStrengthRemaining
						g.msg(foundation.Msg(fmt.Sprintf("You reduced the lock strength by %d%%, lock strength remaining: %d%%", reduction, remaining)))
					}

				} else {
					skill := g.Player.GetCharSheet().GetSkill(special.Lockpick)
					modifier := b.lockDiff.GetRollModifier()
					chance := skill + modifier
					if chance <= 0 {
						g.msg(foundation.Msg("You don't have the skill to pick this lock"))
						return
					}
					luckPercent := special.Percentage(g.Player.GetCharSheet().GetStat(special.Luck))
					rollResult := special.SuccessRoll(special.Percentage(chance), luckPercent)
					if !rollResult.Success && rollResult.Crit {
						g.Player.GetInventory().RemoveLockpick()
						g.msg(foundation.Msg("Your lockpick broke!"))
						return
					}
					lockPickResult(rollResult.Success)
				}

			}
		}
	}

}

func (b *Door) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := b.BaseObject.gobEncode(enc); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.lockedFlag); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.lockDiff); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.numberLock); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.hitpoints); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.damageThreshold); err != nil {
		return nil, err
	}

	if err := enc.Encode(b.audioCueBaseName); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *Door) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))

	b.BaseObject = &BaseObject{}

	if err := b.BaseObject.gobDecode(dec); err != nil {
		return err
	}

	if err := dec.Decode(&b.lockedFlag); err != nil {
		return err
	}

	if err := dec.Decode(&b.lockDiff); err != nil {
		return err
	}

	if err := dec.Decode(&b.numberLock); err != nil {
		return err
	}

	if err := dec.Decode(&b.hitpoints); err != nil {
		return err
	}

	if err := dec.Decode(&b.damageThreshold); err != nil {
		return err
	}

	if err := dec.Decode(&b.audioCueBaseName); err != nil {
		return err
	}

	return nil
}
func (b *Door) IsTransparent() bool {
	switch b.GetCategory() {
	case foundation.ObjectOpenDoor:
		return true
	case foundation.ObjectBrokenDoor:
		return true
	}
	return false
}

func (b *Door) Icon() textiles.TextIcon {
	return b.iconForObject(b.GetCategory().LowerString())
}
func (b *Door) IsWalkable(actor *Actor) bool {
	return b.GetCategory() != foundation.ObjectLockedDoor || (actor != nil && actor.HasKey(b.lockedFlag))
}
func (b *Door) AppendContextActions(items []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	if b.GetCategory() == foundation.ObjectLockedDoor && g.Player.HasKey(b.lockedFlag) {
		items = append(items, foundation.MenuItem{
			Name: "Unlock",
			Action: func() {
				b.Unlock()
				g.ui.PlayCue("world/PICKKEYS")
			},
			CloseMenus: true,
		})
	}
	if b.GetCategory() == foundation.ObjectOpenDoor {
		items = append(items, foundation.MenuItem{
			Name:       "Close",
			Action:     b.Close,
			CloseMenus: true,
		})
	}
	if b.GetCategory() == foundation.ObjectClosedDoor {
		items = append(items, foundation.MenuItem{
			Name:       "Open",
			Action:     b.Open,
			CloseMenus: true,
		})
	}
	return items
}

func (b *Door) SetLockedByFlag(flag string) {
	b.lockedFlag = flag
}
func (b *Door) SetLockDifficulty(difficulty foundation.Difficulty) {
	b.lockDiff = difficulty
	b.lockStrengthRemaining = difficulty.GetStrength()
}

func (b *Door) IsLocked() bool {
	return b.GetCategory() == foundation.ObjectLockedDoor
}

func (b *Door) GetLockFlag() string {
	return b.lockedFlag
}

func (b *Door) Unlock() {
	if b.IsBroken() {
		return
	}
	b.category = foundation.ObjectClosedDoor
}

func (b *Door) Close() {
	if b.IsBroken() {
		return
	}
	b.category = foundation.ObjectClosedDoor
	b.PlayCloseSfx()
	b.updatePlayerFoV()
}

func (b *Door) PlayCloseSfx() {
	cueName := fmt.Sprintf("world/%s_close", b.audioCueBaseName)
	b.player.PlayCue(cueName)
}
func (b *Door) Open() {
	if b.IsBroken() {
		return
	}
	b.category = foundation.ObjectOpenDoor
	b.PlayOpenSfx()
	b.updatePlayerFoV()
}

func (b *Door) PlayOpenSfx() {
	cueName := fmt.Sprintf("world/%s_open", b.audioCueBaseName)
	b.player.PlayCue(cueName)
}
func (b *Door) IsBroken() bool {
	return b.GetCategory() == foundation.ObjectBrokenDoor
}
func (g *GameState) NewDoor(rec recfile.Record) *Door {
	door := &Door{
		BaseObject: &BaseObject{
			category:      foundation.ObjectClosedDoor,
			isAlive:       true,
			displayName:   "a door",
			iconForObject: g.iconForObject,
		},
		hitpoints:       10,
		damageThreshold: 2,
	}
	door.SetLockDifficulty(foundation.Easy)
	door.SetWalkable(false)
	door.SetHidden(false)
	door.SetTransparent(true)

	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			door.internalName = field.Value
		case "category":
			switch strings.ToLower(field.Value) {
			case "lockeddoor":
				door.category = foundation.ObjectLockedDoor
			case "closeddoor":
				door.category = foundation.ObjectClosedDoor
			case "opendoor":
				door.category = foundation.ObjectOpenDoor
			case "brokendoor":
				door.category = foundation.ObjectBrokenDoor
			}
		case "description":
			door.displayName = field.Value
		case "lockflag":
			door.lockedFlag = field.Value
		case "numberlock":
			door.numberLock = []rune(field.Value)
		case "lockdifficulty":
			door.SetLockDifficulty(foundation.DifficultyFromString(field.Value))
		case "position":
			door.position, _ = geometry.NewPointFromEncodedString(field.Value)
		case "hitpoints":
			door.hitpoints = field.AsInt()
		case "damage_threshold":
			door.damageThreshold = field.AsInt()
		case "audiocue":
			door.audioCueBaseName = field.Value
		}
	}

	door.InitWithGameState(g)
	return door
}

func (b *Door) OnDamage(dmg SourcedDamage) []foundation.Animation {
	if dmg.DamageType == special.DamageTypeEMP || dmg.DamageType == special.DamageTypeRadiation || dmg.DamageType == special.DamageTypePoison {
		return nil
	}

	reducedDamage := max(0, dmg.DamageAmount-b.damageThreshold)
	if reducedDamage > 0 {
		b.hitpoints -= reducedDamage
		if b.hitpoints <= 0 {
			b.category = foundation.ObjectBrokenDoor
		}
	}

	return nil
}

func (b *Door) Name() string {
	if b.IsBroken() {
		return fmt.Sprintf("%s (broken)", b.displayName)
	}

	if b.IsOpen() {
		return fmt.Sprintf("%s (open)", b.displayName)
	}

	if b.IsLocked() {
		var lockedWith string
		if b.lockedFlag != "" && len(b.numberLock) == 0 {
			lockedWith = fmt.Sprintf("a %s mechanical lock", b.lockDiff.String())
		} else if len(b.numberLock) > 0 {
			lockedWith = "a numeric keypad"
		}
		strengthString := fmt.Sprintf("DT: %d HP: %d", b.damageThreshold, b.hitpoints)
		return fmt.Sprintf("%s (locked with %s, %s)", b.displayName, lockedWith, strengthString)
	}

	return fmt.Sprintf("%s (closed)", b.displayName)
}

func (b *Door) IsOpen() bool {
	return b.GetCategory() == foundation.ObjectOpenDoor
}

func (b *Door) IsClosedButNotLocked() bool {
	return b.GetCategory() == foundation.ObjectClosedDoor
}

func (b *Door) OnBump(actor *Actor) {
	if b.onBump != nil {
		b.onBump(actor)
	}
}

func (b *Door) ToRecord() recfile.Record {
	rec := recfile.Record{}
	rec = append(rec, recfile.Field{Name: "position", Value: b.position.Encode()})
	rec = append(rec, recfile.Field{Name: "category", Value: b.GetCategory().String()})
	rec = append(rec, recfile.Field{Name: "description", Value: b.displayName})
	if b.lockedFlag != "" {
		rec = append(rec, recfile.Field{Name: "lockflag", Value: b.lockedFlag})
	}
	if len(b.numberLock) > 0 {
		rec = append(rec, recfile.Field{Name: "numberlock", Value: string(b.numberLock)})
	}
	if b.lockDiff != foundation.Easy {
		rec = append(rec, recfile.Field{Name: "lockdifficulty", Value: b.lockDiff.String()})
	}
	if b.hitpoints != 10 {
		rec = append(rec, recfile.Field{Name: "hitpoints", Value: recfile.IntStr(b.hitpoints)})
	}
	if b.damageThreshold != 2 {
		rec = append(rec, recfile.Field{Name: "damage_threshold", Value: recfile.IntStr(b.damageThreshold)})
	}
	if b.audioCueBaseName != "" {
		rec = append(rec, recfile.Field{Name: "audiocue", Value: b.audioCueBaseName})
	}
	return rec
}

func (b *Door) PickByReduceStrength(reduction int) bool {
	b.lockStrengthRemaining -= reduction
	if b.lockStrengthRemaining <= 0 {
		b.Unlock()
		return true
	}
	return false
}
