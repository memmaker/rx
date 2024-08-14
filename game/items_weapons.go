package game

import (
	"RogueUI/special"
	"github.com/memmaker/go/fxtools"
	"strings"
)

type AmmoInfo struct {
	damage fxtools.Interval
	kind   string
}

func (i AmmoInfo) GetKind() string {
	return i.kind
}

type TargetingMode int

// as bitmask
const (
	TargetingModeNone     TargetingMode = 0
	TargetingModeSingle   TargetingMode = 1
	TargetingModeAimed    TargetingMode = 2
	TargetingModeBurst    TargetingMode = 4
	TargetingModeFullAuto TargetingMode = 8
)

func TargetingModeFromString(value string) TargetingMode {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "single":
		return TargetingModeSingle
	case "aimed":
		return TargetingModeAimed
	case "burst":
		return TargetingModeBurst
	case "full_auto":
		return TargetingModeFullAuto
	}
	panic("Invalid targeting mode: " + value)
	return TargetingModeNone
}
func (t TargetingMode) Next() TargetingMode {
	if t == TargetingModeNone {
		return TargetingModeSingle
	}
	nextVal := t << 1
	if nextVal > TargetingModeFullAuto {
		nextVal = TargetingModeSingle
	}
	return nextVal
}

func (t TargetingMode) ToString() string {
	switch t {
	case TargetingModeSingle:
		return "Single"
	case TargetingModeAimed:
		return "Aimed"
	case TargetingModeBurst:
		return "Burst"
	case TargetingModeFullAuto:
		return "Full Auto"
	}
	return "Unknown"
}

type AmmoDef struct {
	Damage fxtools.Interval
	Kind   string
}

func (d AmmoDef) IsValid() bool {
	return d.Damage.NotZero()
}

type WeaponDef struct {
	Damage              fxtools.Interval
	Type                WeaponType
	SkillUsed           special.Skill
	ShotMaxRange        int
	ShotMinRange        int
	ShotHalfDamageRange int
	ShotAccuracy        int
	TargetingMode       TargetingMode
	// linking to another weapon type
	UsesAmmo     string
	MagazineSize int
	BurstRounds  int
}

func (w WeaponDef) IsValid() bool {
	return w.Type != WeaponTypeUnknown && w.Damage.NotZero() && w.TargetingMode != TargetingModeNone
}

type WeaponInfo struct {
	damageDice           fxtools.Interval
	weaponType           WeaponType
	usesAmmo             string
	vorpalEnemy          string
	skillUsed            special.Skill
	magazineSize         int
	loadedInMagazine     *Item
	qualityInPercent     int
	targetingMode        TargetingMode
	currentTargetingMode TargetingMode
	burstRounds          int
}

func (i *WeaponInfo) GetVorpalEnemy() string {
	return i.vorpalEnemy
}

func (i *WeaponInfo) Vorpalize(enemy string) {
	i.vorpalEnemy = enemy
}

func (i *WeaponInfo) GetDamage() fxtools.Interval {
	if i.loadedInMagazine != nil {
		return i.loadedInMagazine.GetAmmo().damage
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

func (i *WeaponInfo) UsesAmmo() string {
	return i.usesAmmo
}

func (i *WeaponInfo) BulletsNeededForFullClip() (int, string) {
	if i.loadedInMagazine == nil {
		return i.magazineSize, ""
	}
	ammoKind := i.loadedInMagazine.GetInternalName()
	return i.magazineSize - i.GetLoadedBullets(), ammoKind
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
	return i.targetingMode&mode != 0
}

func (i *WeaponInfo) HasAmmo() bool {
	return i.GetLoadedBullets() > 0 || i.usesAmmo == ""
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

type WeaponType int

func (t WeaponType) IsMissile() bool {
	return t == WeaponTypeArrow || t == WeaponTypeBolt || t == WeaponTypeDart || t == WeaponTypeMissile || t == WepaonTypeBullet
}

func (t WeaponType) IsRanged() bool {
	return t.IsMissile() || t == WeaponTypeBow || t == WeaponTypeCrossbow || t == WeaponTypePistol || t == WeaponTypeRifle || t == WeaponTypeShotgun
}

func (t WeaponType) IsMelee() bool {
	return t == WeaponTypeSword || t == WeaponTypeClub || t == WeaponTypeAxe || t == WeaponTypeDagger || t == WeaponTypeSpear
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
	WepaonTypeBullet
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

	}
	panic("Invalid weapon type: " + value)
	return WeaponTypeUnknown
}
