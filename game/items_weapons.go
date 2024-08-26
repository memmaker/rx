package game

import (
	"RogueUI/special"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"strings"
)

type AttackMode struct {
	Mode     special.TargetingMode
	TUCost   int
	MaxRange int
	IsAimed  bool
}

func (m AttackMode) String() string {
	if m.IsAimed {
		return fmt.Sprintf("%s (Aimed)", m.Mode.ToString())
	}
	return m.Mode.ToString()
}

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
	// linking to another weapon type
	MagazineSize     int
	BurstRounds      int
	CaliberIndex     int
	SoundID          int32
	DamageType       special.DamageType
	TargetingModeOne special.TargetingMode
	TargetingModeTwo special.TargetingMode
	TUCostOne        int
	TUCostTwo        int
	MaxRangeOne      int
	MaxRangeTwo      int
}

func (w WeaponDef) IsValid() bool {
	return w.Type != WeaponTypeUnknown && w.Damage.NotZero() && w.TargetingModeOne != special.TargetingModeNone
}

type WeaponInfo struct {
	damageDice       fxtools.Interval
	weaponType       WeaponType
	vorpalEnemy      string
	skillUsed        special.Skill
	magazineSize     int
	loadedInMagazine *Item
	qualityInPercent int
	burstRounds      int
	caliberIndex     int
	attackModes      []AttackMode
	soundID          int32
	damageType       special.DamageType
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

func (i *WeaponInfo) HasAmmo() bool {
	return i.GetLoadedBullets() > 0 || !i.NeedsAmmo()
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

func (i *WeaponInfo) GetFireAudioCue(mode special.TargetingMode) string {
	strMode := "single"
	if mode == special.TargetingModeFireBurst || mode == special.TargetingModeFireFullAuto {
		strMode = "burst"
	}
	return fmt.Sprintf("weapons/%d_%s", i.soundID, strMode)
}

func (i *WeaponInfo) GetReloadAudioCue() string {
	return fmt.Sprintf("weapons/%d_reload", i.soundID)
}
func (i *WeaponInfo) GetOutOfAmmoAudioCue() string {
	return fmt.Sprintf("weapons/%d_out_of_ammo", i.soundID)
}
func (i *WeaponInfo) GetMissAudioCue() string {
	return fmt.Sprintf("weapons/%d_hit_surface", i.soundID)
}

func (i *WeaponInfo) GetDamageType() special.DamageType {
	return i.damageType
}

func (i *WeaponInfo) GetAttackMode(index int) AttackMode {
	return i.attackModes[index]
}

type WeaponType int

func (t WeaponType) IsMissile() bool {
	return t == WeaponTypeArrow || t == WeaponTypeBolt || t == WeaponTypeDart || t == WeaponTypeMissile || t == WeaponTypeBullet
}

func (t WeaponType) IsRanged() bool {
	return t.IsMissile() || t == WeaponTypeBow || t == WeaponTypeCrossbow || t == WeaponTypePistol || t == WeaponTypeRifle || t == WeaponTypeShotgun || t == WeaponTypeSMG || t == WeaponTypeMinigun || t == WeaponTypeRocketLauncher || t == WeaponTypeBigGun || t == WeaponTypeEnergy
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
