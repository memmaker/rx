package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type Door struct {
	*BaseObject
	LockedFlag            string
	LockDiff              d100.Difficulty
	LockStrengthRemaining int
	NumberLock            []rune

	HitPoints        int
	DamageThreshold  int
	AudioCueBaseName string
	player           foundation.AudioCuePlayer
	onBump           func(actor *Actor)
	updatePlayerFoV  func()
	setUnlockedFlag  func()
}

func (b *Door) RemainingStrength() int {
	return b.LockStrengthRemaining
}

func (b *Door) InitWithGameState(g *GameState) {
	b.player = g.ui

	b.updatePlayerFoV = func() {
		g.updateFoVAndDijkstraMap(g.Player)
	}
	b.setUnlockedFlag = func() {
		g.gameFlags.SetFlag(fmt.Sprintf("DoorUnlocked(%s)", b.InternalName))
		g.updateFoVAndDijkstraMap(g.Player)
	}
	b.onBump = func(actor *Actor) {
		if actor == g.Player && b.GetCategory() == foundation.ObjectLockedDoor {
			if b.LockedFlag != "" && actor.HasKey(b.LockedFlag) {
				b.Unlock()
				g.msg(foundation.Msg("You unlocked the door using the key"))
				g.ui.PlayCue("world/PICKKEYS")
				g.endPlayerTurn(g.Player.TimeNeededForActions())
				return
			}

			if len(b.NumberLock) > 0 {
				picksNeeded := d100.ElectronicLockpicksNeeded(b.LockDiff, g.Player.GetCharSheet().GetSkill(d100.SkillForPickLocks))
				actionInfo := fmt.Sprintf("Hack (%d)", picksNeeded)
				g.ui.OpenKeypad(actionInfo, b.NumberLock, func() bool {
					success := g.playerPickElectronicLock(b.LockDiff)
					if success {
						b.Unlock()
					}
					return success
				}, func(result bool) {
					if result {
						b.Unlock()
						g.msg(foundation.Msg("You unlocked the door using the code"))
						g.endPlayerTurn(g.Player.TimeNeededForActions())
					}
				})
			} else {
				// lockpicking
				if g.Player.GetInventory().GetLockpickCount(LockTypeMechanical) == 0 {
					g.msg(foundation.Msg("You don't have any lockpicks"))
					return
				}
				if b.IsLockAtFullStrength() {
					g.ui.AskForConfirmation("Confirm", "Do you want to pick the lock?", func(result bool) {
						if result {
							g.playerTryPickLock(b)
						}
					})
				} else {
					g.playerTryPickLock(b)
				}
			}
		}
	}

}

func (g *GameState) playerPickElectronicLock(difficulty d100.Difficulty) bool {
	if g.playerMinorCrimeIsDenied() {
		return false
	}

	picksNeeded := d100.ElectronicLockpicksNeeded(difficulty, g.Player.GetCharSheet().GetSkill(d100.SkillForHacking))

	if g.Player.GetInventory().GetLockpickCount(LockTypeElectronic) >= picksNeeded {
		g.Player.GetInventory().RemoveLockpicks(LockTypeElectronic, picksNeeded)
		g.ui.PlayCue("world/confirmed")
		g.msg(foundation.Msg(fmt.Sprintf("%d e-pick(s) removed", picksNeeded)))
		g.msg(foundation.Msg("You have defeated the electronic lock"))
		g.endPlayerTurn(g.Player.TimeNeededForActions())
		return true // should close?
	} else {
		g.msg(foundation.Msg("You don't have enough e-picks"))
		return false
	}
}

type Pickable interface {
	ReduceStrength(reduction int) (int, bool)
	Unlock()
	RemainingStrength() int
}

func (g *GameState) playerTryPickLock(pickable Pickable) {
	if g.playerMinorCrimeIsDenied() {
		return
	}
	g.Player.GetInventory().RemoveLockpicks(LockTypeMechanical, 1)
	skill := g.Player.GetCharSheet().GetSkill(d100.SkillForPickLocks)
	if skill <= 0 {
		g.msg(foundation.Msg("You don't have the skill to pick locks"))
		return
	}
	reduction := int(float64(skill) * d100.LockStrengthReductionPerSkill)
	if reducedBy, didOpen := pickable.ReduceStrength(reduction); didOpen {
		pickable.Unlock()
		g.msg(foundation.Msg("You picked the lock deftly"))
		g.ui.PlayCue("world/PICKKEYS")
	} else {
		remaining := pickable.RemainingStrength()
		g.msg(foundation.Msg(fmt.Sprintf("You reduced the lock strength by %d%%, lock strength remaining: %d%%", reducedBy, remaining)))
	}
	g.endPlayerTurn(g.Player.TimeNeededForActions())
}
func (b *Door) IsTransparent() bool {
	if b.Transparent {
		return true
	}
	switch b.GetCategory() {
	case foundation.ObjectOpenDoor:
		return true
	case foundation.ObjectBrokenDoor:
		return true
	}
	return false
}

func (b *Door) GetIcon() textiles.TextIcon {
	return b.iconForObject(b.GetCategory().LowerString())
}
func (b *Door) IsWalkable(actor *Actor) bool {
	return b.GetCategory() != foundation.ObjectLockedDoor || (actor != nil && actor.HasKey(b.LockedFlag))
}
func (b *Door) AppendContextActions(items []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	if b.GetCategory() == foundation.ObjectLockedDoor {
		if g.Player.HasKey(b.LockedFlag) {
			items = append(items, foundation.MenuItem{
				Name: "Unlock",
				Action: func() {
					b.Unlock()
					g.msg(foundation.Msg("You unlocked the door using the key"))
					g.ui.PlayCue("world/PICKKEYS")
				},
				CloseMenus: true,
			})
		}
		if len(b.NumberLock) > 0 {

		}

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
	b.LockedFlag = flag
}
func (b *Door) SetLockDifficulty(difficulty d100.Difficulty) {
	b.LockDiff = difficulty
	b.LockStrengthRemaining = 100
}

func (b *Door) IsLocked() bool {
	return b.GetCategory() == foundation.ObjectLockedDoor
}

func (b *Door) GetLockFlag() string {
	return b.LockedFlag
}

func (b *Door) Unlock() {
	if b.IsBroken() {
		return
	}
	b.Category = foundation.ObjectClosedDoor
	b.setUnlockedFlag()
}

func (b *Door) Close() {
	if b.IsBroken() {
		return
	}
	b.Category = foundation.ObjectClosedDoor
	b.PlayCloseSfx()
	b.updatePlayerFoV()
}

func (b *Door) PlayCloseSfx() {
	cueName := fmt.Sprintf("world/%s_close", b.AudioCueBaseName)
	b.player.PlayCue(cueName)
}
func (b *Door) Open() {
	if b.IsBroken() {
		return
	}
	b.Category = foundation.ObjectOpenDoor
	b.PlayOpenSfx()
	b.updatePlayerFoV()
}

func (b *Door) PlayOpenSfx() {
	cueName := fmt.Sprintf("world/%s_open", b.AudioCueBaseName)
	b.player.PlayCue(cueName)
}
func (b *Door) IsBroken() bool {
	return b.GetCategory() == foundation.ObjectBrokenDoor
}
func (g *GameState) NewDoor(rec recfile.Record, resolver func(objType string) textiles.TextIcon) *Door {
	door := &Door{
		BaseObject: &BaseObject{
			Category:      foundation.ObjectClosedDoor,
			DisplayName:   "a door",
			iconForObject: resolver,
		},
		HitPoints:       10,
		DamageThreshold: 2,
	}
	door.SetLockDifficulty(d100.Easy)
	door.SetWalkable(false)
	door.SetHidden(false)
	door.SetTransparent(false)
	randomNumberCode := false
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			door.InternalName = field.Value
		case "category":
			switch strings.ToLower(field.Value) {
			case "lockeddoor":
				door.Category = foundation.ObjectLockedDoor
			case "closeddoor":
				door.Category = foundation.ObjectClosedDoor
			case "opendoor":
				door.Category = foundation.ObjectOpenDoor
			case "brokendoor":
				door.Category = foundation.ObjectBrokenDoor
			}
		case "description":
			door.DisplayName = field.Value
		case "lockflag":
			door.LockedFlag = field.Value
		case "numberlock":
			if strings.ToLower(field.Value) == "random" {
				// random 4-digit code
				randomNumberCode = true
			} else {
				door.NumberLock = []rune(field.Value)
			}
		case "lockdifficulty":
			door.SetLockDifficulty(d100.DifficultyFromString(field.Value))
		case "position":
			door.RawPosition, _ = geometry.NewPointFromEncodedString(field.Value)
		case "hitpoints":
			door.HitPoints = field.AsInt()
		case "damage_threshold":
			door.DamageThreshold = field.AsInt()
		case "audiocue":
			door.AudioCueBaseName = field.Value
		case "istransparent":
			door.SetTransparent(field.AsBool())
		case "ispassableforprojectile":
			door.PassableForProjectile = field.AsBool()
		}
	}
	if randomNumberCode {
		door.NumberLock = g.getRandomNumberCode(door.InternalName)
	}
	door.InitWithGameState(g)
	return door
}

func (b *Door) OnDamage(dmg SourcedDamage) []foundation.Animation {
	if dmg.DamageType == DamageTypeEMP || dmg.DamageType == DamageTypeRadiation || dmg.DamageType == DamageTypePoison {
		return nil
	}

	reducedDamage := max(0, dmg.DamageAmount-b.DamageThreshold)
	if reducedDamage > 0 {
		b.HitPoints -= reducedDamage
		if b.HitPoints <= 0 {
			b.Category = foundation.ObjectBrokenDoor
		}
	}

	return nil
}

func (b *Door) Name() string {
	if b.IsBroken() {
		return fmt.Sprintf("%s (broken)", b.DisplayName)
	}

	if b.IsOpen() {
		return fmt.Sprintf("%s (open)", b.DisplayName)
	}

	if b.IsLocked() {
		var lockedWith string
		if b.LockedFlag != "" && len(b.NumberLock) == 0 {
			lockedWith = fmt.Sprintf("a %s mechanical lock", b.LockDiff.String())
		} else if len(b.NumberLock) > 0 {
			lockedWith = "a numeric keypad"
		}
		strengthString := fmt.Sprintf("DT: %d HP: %d", b.DamageThreshold, b.HitPoints)
		return fmt.Sprintf("%s (locked with %s, %s)", b.DisplayName, lockedWith, strengthString)
	}

	return fmt.Sprintf("%s (closed)", b.DisplayName)
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
	rec = append(rec, recfile.Field{Name: "position", Value: b.RawPosition.Encode()})
	rec = append(rec, recfile.Field{Name: "category", Value: b.GetCategory().String()})
	rec = append(rec, recfile.Field{Name: "description", Value: b.DisplayName})
	if b.LockedFlag != "" {
		rec = append(rec, recfile.Field{Name: "lockflag", Value: b.LockedFlag})
	}
	if len(b.NumberLock) > 0 {
		rec = append(rec, recfile.Field{Name: "numberlock", Value: string(b.NumberLock)})
	}
	if b.LockDiff != d100.Easy {
		rec = append(rec, recfile.Field{Name: "lockdifficulty", Value: b.LockDiff.String()})
	}
	if b.HitPoints != 10 {
		rec = append(rec, recfile.Field{Name: "hitpoints", Value: recfile.IntStr(b.HitPoints)})
	}
	if b.DamageThreshold != 2 {
		rec = append(rec, recfile.Field{Name: "damage_threshold", Value: recfile.IntStr(b.DamageThreshold)})
	}
	if b.AudioCueBaseName != "" {
		rec = append(rec, recfile.Field{Name: "audiocue", Value: b.AudioCueBaseName})
	}
	return rec
}

func (b *Door) ReduceStrength(reduction int) (int, bool) {
	if b.IsBroken() || b.IsOpen() || b.IsClosedButNotLocked() || b.LockStrengthRemaining <= 0 {
		return 0, false
	}
	realReduction := int(float64(reduction) * b.LockDiff.LockReductionFactor())
	b.LockStrengthRemaining -= realReduction
	if b.LockStrengthRemaining <= 0 {
		b.Unlock()
		return realReduction, true
	}
	return realReduction, false
}

func (b *Door) IsLockAtFullStrength() bool {
	return b.LockStrengthRemaining == 100
}
