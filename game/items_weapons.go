package game

import (
	"RogueUI/special"
	"github.com/memmaker/go/fxtools"
	"strings"
)

type AmmoInfo struct {
	DamageMultiplier int
	DamageDivisor    int
	ACModifier       int
	DRModifier       int
	RoundsInMagazine int
	CaliberIndex     int
}

func (i AmmoInfo) Equals(other *AmmoInfo) bool {
	return i.DamageMultiplier == other.DamageMultiplier &&
		i.DamageDivisor == other.DamageDivisor &&
		i.ACModifier == other.ACModifier &&
		i.DRModifier == other.DRModifier &&
		i.RoundsInMagazine == other.RoundsInMagazine &&
		i.CaliberIndex == other.CaliberIndex

}

type TargetingMode int

/*
	case AttackModeNone:
		return "None"
	case AttackModePunch:
		return "Punch"
	case AttackModeKick:
		return "Kick"
	case AttackModeSwing:
		return "Swing"
	case AttackModeThrust:
		return "Thrust"
	case AttackModeThrow:
		return "Throw"
	case AttackModeFireSingle:
		return "Fire_Single"
	case AttackModeFireBurst:
		return "Fire_Burst"
	case AttackModeFlame:
		return "Flame"
*/
// as bitmask
const (
	TargetingModeNone         TargetingMode = 0
	TargetingModePunch        TargetingMode = 1
	TargetingModeKick         TargetingMode = 2
	TargetingModeSwing        TargetingMode = 4
	TargetingModeThrust       TargetingMode = 8
	TargetingModeThrow        TargetingMode = 16
	TargetingModeFireSingle   TargetingMode = 32
	TargetingModeFireAimed    TargetingMode = 64
	TargetingModeFireBurst    TargetingMode = 128
	TargetingModeFireFullAuto TargetingMode = 256
	TargetingModeFlame        TargetingMode = 512
)

func TargetingModeFromString(value string) TargetingMode {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "none":
		return TargetingModeNone
	case "punch":
		return TargetingModePunch
	case "kick":
		return TargetingModeKick
	case "swing":
		return TargetingModeSwing
	case "thrust":
		return TargetingModeThrust
	case "throw":
		return TargetingModeThrow
	case "fire_single":
		return TargetingModeFireSingle
	case "fire_burst":
		return TargetingModeFireBurst
	case "flame":
		return TargetingModeFlame
	}
	panic("Invalid targeting mode: " + value)
	return TargetingModeNone
}
func (t TargetingMode) Next() TargetingMode {
	if t == TargetingModeNone {
		return TargetingModePunch
	}
	nextVal := t << 1
	if nextVal > TargetingModeFlame {
		nextVal = TargetingModePunch
	}
	return nextVal
}

func (t TargetingMode) ToString() string {
	switch t {
	case TargetingModeNone:
		return "None"
	case TargetingModePunch:
		return "Punch"
	case TargetingModeKick:
		return "Kick"
	case TargetingModeSwing:
		return "Swing"
	case TargetingModeThrust:
		return "Thrust"
	case TargetingModeThrow:
		return "Throw"
	case TargetingModeFireSingle:
		return "Fire_Single"
	case TargetingModeFireBurst:
		return "Fire_Burst"
	case TargetingModeFlame:
		return "Flame"
	}
	return "Unknown"
}

/*
Name: flamethrower_fuel
Description: Flamethrower Fuel
Category: Ammo
Size: 0
Weight: 10
Cost: 250
ammo_dmg_multiplier: 1
ammo_dmg_divisor: 1
ammo_ac_modifier: -20
ammo_dr_modifier: 25
ammo_rounds_in_magazine: 10
ammo_caliber_index: 2
*/
type AmmoDef struct {
	DamageMultiplier int
	DamageDivisor    int
	ACModifier       int
	DRModifier       int
	RoundsInMagazine int
	CaliberIndex     int
}

func (d AmmoDef) IsValid() bool {
	return d.DamageMultiplier != 0 && d.DamageDivisor != 0 && d.RoundsInMagazine > 0
}

type WeaponDef struct {
	Damage              fxtools.Interval
	Type                WeaponType
	SkillUsed           special.Skill
	ShotMaxRange        int
	ShotMinRange        int
	ShotHalfDamageRange int
	ShotAccuracy        int
	TargetingModeOne    TargetingMode
	TargetingModeTwo    TargetingMode
	// linking to another weapon type
	MagazineSize int
	BurstRounds  int
	CaliberIndex int
}

func (w WeaponDef) IsValid() bool {
	return w.Type != WeaponTypeUnknown && w.Damage.NotZero() && w.TargetingModeOne != TargetingModeNone
}

type WeaponInfo struct {
	damageDice           fxtools.Interval
	weaponType           WeaponType
	vorpalEnemy          string
	skillUsed            special.Skill
	magazineSize         int
	loadedInMagazine     *Item
	qualityInPercent     int
	currentTargetingMode TargetingMode
	burstRounds          int
	caliberIndex         int
	targetingModeOne     TargetingMode
	targetingModeTwo     TargetingMode
}

func (i *WeaponInfo) GetVorpalEnemy() string {
	return i.vorpalEnemy
}

func (i *WeaponInfo) Vorpalize(enemy string) {
	i.vorpalEnemy = enemy
}

func (i *WeaponInfo) GetDamage() fxtools.Interval {
	if i.loadedInMagazine != nil {
		//ammo := i.loadedInMagazine.GetAmmo()
		// TODO: APPLY AMMO EFFECTS
	}
	return i.damageDice
}

func (i *WeaponInfo) GetVorpalBonus(enemyName string) (int, int) {
	if i.vorpalEnemy != "" {
		if i.vorpalEnemy == enemyName {
			return 4, 4
		}
		return 1, 1
	}
	return 0, 0
}

func (i *WeaponInfo) GetWeaponType() WeaponType {
	return i.weaponType
}

func (i *WeaponInfo) GetSkillUsed() special.Skill {
	return i.skillUsed
}

func (i *WeaponInfo) IsVorpal() bool {
	return i.vorpalEnemy != ""
}

func (i *WeaponInfo) GetCaliber() int {
	return i.caliberIndex
}

func (i *WeaponInfo) BulletsNeededForFullClip() (int, string) {
	if i.loadedInMagazine == nil {
		return i.magazineSize, ""
	}
	ammoKind := i.loadedInMagazine
	return i.magazineSize - i.GetLoadedBullets(), ammoKind.GetInternalName()
}

func (i *WeaponInfo) LoadAmmo(ammo *Item) *Item {
	if i.loadedInMagazine == nil {
		i.loadedInMagazine = ammo
		return nil
	}
	if i.loadedInMagazine.CanStackWith(ammo) {
		i.loadedInMagazine.Merge(ammo)
		return nil
	}
	oldAmmo := i.loadedInMagazine
	i.loadedInMagazine = ammo
	if oldAmmo.GetCharges() > 0 {
		return oldAmmo
	}
	return nil
}

func (i *WeaponInfo) IsRanged() bool {
	return i.weaponType.IsRanged()
}

func (i *WeaponInfo) IsMelee() bool {
	return i.weaponType.IsMelee()
}

func (i *WeaponInfo) CycleTargetMode() {
	var nextMode TargetingMode
	for nextMode = i.currentTargetingMode.Next(); !i.IsTargetModeSupported(nextMode); nextMode = nextMode.Next() {
	}
	i.currentTargetingMode = nextMode
}

func (i *WeaponInfo) IsTargetModeSupported(mode TargetingMode) bool {
	// does the bitmask in targetingMode contain the mode?
	return i.targetingModeOne == mode || i.targetingModeTwo == mode
}

func (i *WeaponInfo) HasAmmo() bool {
	return i.GetLoadedBullets() > 0 || !i.NeedsAmmo()
}

func (i *WeaponInfo) GetCurrentTargetingMode() TargetingMode {
	return i.currentTargetingMode
}

func (i *WeaponInfo) GetTimeNeeded() int {
	return 10
}

func (i *WeaponInfo) GetLoadedBullets() int {
	if i.loadedInMagazine == nil {
		return 0
	}
	return i.loadedInMagazine.GetCharges()
}

func (i *WeaponInfo) GetMagazineSize() int {
	return i.magazineSize
}

func (i *WeaponInfo) RemoveBullets(spent int) {
	if i.loadedInMagazine == nil {
		return
	}

	i.loadedInMagazine.RemoveCharges(spent)
}

func (i *WeaponInfo) GetBurstRounds() int {
	return i.burstRounds
}

func (i *WeaponInfo) NeedsAmmo() bool {
	return i.caliberIndex > 0
}

type WeaponType int

func (t WeaponType) IsMissile() bool {
	return t == WeaponTypeArrow || t == WeaponTypeBolt || t == WeaponTypeDart || t == WeaponTypeMissile || t == WeaponTypeBullet
}

func (t WeaponType) IsRanged() bool {
	return t.IsMissile() || t == WeaponTypeBow || t == WeaponTypeCrossbow || t == WeaponTypePistol || t == WeaponTypeRifle || t == WeaponTypeShotgun || t == WeaponTypeSMG || t == WeaponTypeMinigun || t == WeaponTypeRocketLauncher || t == WeaponTypeBigGun
}

func (t WeaponType) IsMelee() bool {
	return t == WeaponTypeSword || t == WeaponTypeClub || t == WeaponTypeAxe || t == WeaponTypeDagger || t == WeaponTypeSpear || t == WeaponTypeKnife || t == WeaponTypeEnergy || t == WeaponTypeThrown || t == WeaponTypeMelee
}

const (
	WeaponTypeUnknown WeaponType = iota
	WeaponTypeSword
	WeaponTypeClub
	WeaponTypeAxe
	WeaponTypeDagger
	WeaponTypeSpear
	WeaponTypeBow
	WeaponTypeArrow
	WeaponTypeCrossbow
	WeaponTypeBolt
	WeaponTypeDart
	WeaponTypePistol
	WeaponTypeRifle
	WeaponTypeShotgun
	WeaponTypeMissile
	WeaponTypeBullet
	WeaponTypeSMG
	WeaponTypeSledgehammer
	WeaponTypeMinigun
	WeaponTypeRocketLauncher
	WeaponTypeBigGun
	WeaponTypeKnife
	WeaponTypeEnergy
	WeaponTypeThrown
	WeaponTypeMelee
)

func WeaponTypeFromString(value string) WeaponType {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "sword":
		return WeaponTypeSword
	case "club":
		return WeaponTypeClub
	case "axe":
		return WeaponTypeAxe
	case "dagger":
		return WeaponTypeDagger
	case "spear":
		return WeaponTypeSpear
	case "bow":
		return WeaponTypeBow
	case "arrow":
		return WeaponTypeArrow
	case "crossbow":
		return WeaponTypeCrossbow
	case "bolt":
		return WeaponTypeBolt
	case "dart":
		return WeaponTypeDart
	case "pistol":
		return WeaponTypePistol
	case "rifle":
		return WeaponTypeRifle
	case "shotgun":
		return WeaponTypeShotgun
	case "missile":
		return WeaponTypeMissile
	case "smg":
		return WeaponTypeSMG
	case "sledgehammer":
		return WeaponTypeSledgehammer
	case "minigun":
		return WeaponTypeMinigun
	case "rocketlauncher":
		return WeaponTypeRocketLauncher
	case "biggun":
		return WeaponTypeBigGun
	case "knife":
		return WeaponTypeKnife
	case "energy":
		return WeaponTypeEnergy
	case "throwing":
		return WeaponTypeThrown
	case "melee":
		return WeaponTypeMelee
	}
	panic("Invalid weapon type: " + value)
	return WeaponTypeUnknown
}
