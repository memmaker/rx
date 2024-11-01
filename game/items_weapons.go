package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"math/rand"
	"strings"
)

type WeaponSize uint8

func (c WeaponSize) String() string {
	switch c {
	case SizeHidden:
		return "Hidden"
	case SizeHandWeapon:
		return "Hand Weapon"
	case SizeShortWeapon:
		return "Short Weapon"
	case SizeLongWeapon:
		return "Long Weapon"
	case SizeBig:
		return "Big Weapon"
	}
	return "Big Weapon"
}

func (c WeaponSize) CanFit(weaponOfSize WeaponSize) bool {
	if weaponOfSize == SizeHidden {
		return true
	}
	return c >= weaponOfSize
}

const (
	SizeHidden WeaponSize = iota
	SizeHandWeapon
	SizeShortWeapon
	SizeLongWeapon
	SizeBig
	SizeCount
)

func WeaponSizeFromString(input string) WeaponSize {
	switch strings.ToLower(input) {
	case "big":
		return SizeBig
	case "long":
		return SizeLongWeapon
	case "short":
		return SizeShortWeapon
	case "hand":
		return SizeHandWeapon
	case "hidden":
		return SizeHidden
	}
	return SizeBig
}

type Weapon struct {
	*GenericItem
	damageDice fxtools.Interval
	weaponType WeaponType

	skillUsed        d100.Skill
	magazineSize     int
	loadedInMagazine *Ammo

	PelletCount  int
	burstRounds  int
	caliberIndex int
	caliberName  string

	attackModes []AttackMode
	soundID     int32
	damageType  DamageType
	MinSTR      int

	reliability  d100.Percentage
	relativeSize WeaponSize
	accuracyMod  d100.Percentage

	degradeFactor float64
	jammed        bool
}

func (i *Weapon) Degrade(degrade float64) {
	ammoFactor := i.ammoDegradeFactor()
	weaponFactor := i.degradeFactor
	i.qualityInPercent = max(0, i.qualityInPercent-d100.Percentage(degrade*weaponFactor*ammoFactor))
}

func (i *Weapon) FullDescription(colorCode string) string {
	basicRows := i.GenericItem.fullDescriptionRows()

	basicRows = append(basicRows, fxtools.NewTableRow("Type", i.weaponType.String()))
	basicRows = append(basicRows, fxtools.NewTableRow("Caliber", i.GetCaliberName()))
	basicRows = append(basicRows, fxtools.NewTableRow("Quality", fmt.Sprintf("%d%%", int(i.qualityInPercent))))
	basicRows = append(basicRows, fxtools.NewTableRow("Concealability", i.relativeSize.String()))

	if i.damageType == DamageTypeNormal {
		basicRows = append(basicRows, fxtools.NewTableRow("Damage", fmt.Sprintf("%s", i.damageDice.Scaled(i.qualityInPercent.Normalized()).ShortString())))
	} else {
		basicRows = append(basicRows, fxtools.NewTableRow("Damage", fmt.Sprintf("%s (%s)", i.damageDice.Scaled(i.qualityInPercent.Normalized()).ShortString(), i.damageType.String())))
	}

	if i.accuracyMod != 0 {
		basicRows = append(basicRows, fxtools.NewTableRow("Accuracy Modifier", fmt.Sprintf("%+d%%", int(i.accuracyMod))))
	}
	if i.MinSTR > 0 {
		basicRows = append(basicRows, fxtools.NewTableRow("Min. Strength", fmt.Sprintf("%d", i.MinSTR)))
	}

	basicRows = append(basicRows, fxtools.NewTableRow("Reliability", fmt.Sprintf("%d%%", int(i.reliability))))

	for _, attackMode := range i.attackModes {
		modeVal := fmt.Sprintf("Range: %d, TU: %d", attackMode.MaxRange, attackMode.TUCost)
		modeLabel := attackMode.String()
		basicRows = append(basicRows, fxtools.NewTableRow(modeLabel, modeVal))
	}

	lines := fxtools.TableLayout(basicRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	lines = append([]string{i.InventoryNameWithColors(colorCode), i.category.String()}, lines...)

	lines = i.appendText(lines)

	return strings.Join(lines, "\n")
}
func (i *Weapon) InventoryNameWithColorsAndShortcut(lineColorCode string) string {
	return fmt.Sprintf("%c - %s", i.Shortcut(), i.InventoryNameWithColors(lineColorCode))
}
func (i *Weapon) InventoryNameWithColors(colorCode string) string {
	line := cview.Escape(fmt.Sprintf("%s [%s]", i.Name(), i.GetWeaponDamage().ShortString()))

	lineWithColor := colorCode + line + "[-]"

	qIcon := getQualityIcon(i.qualityInPercent)
	lineWithColor = fmt.Sprintf("%s %s", qIcon, lineWithColor)

	return lineWithColor
}
func (i *Weapon) LongNameWithColors(colorCode string) string {
	weapon := i
	attackMode := weapon.GetAttackMode(i.currentAttackModeIndex)
	targetMode := attackMode.String()
	if i.jammed {
		targetMode = "*JAMMED*"
	}
	bullets := fmt.Sprintf("%d/%d", weapon.GetLoadedBullets(), weapon.GetMagazineSize())
	ammoShortString := weapon.GetAmmoTypeShortString()
	line := cview.Escape(fmt.Sprintf("%s - %s - %s%s", i.Name(), targetMode, bullets, ammoShortString))
	return colorCode + line + "[-]"
}
func (i *Weapon) DisplayLength() int {
	return cview.TaggedStringWidth(i.InventoryNameWithColorsAndShortcut("[red]"))
}
func (i *Weapon) ammoDegradeFactor() float64 {
	factor := 1.0
	weapon := i
	if weapon.NeedsAmmo() && weapon.HasAmmo() {
		ammo := weapon.GetLoadedAmmo()
		ammoInfo := ammo
		factor = ammoInfo.ConditionFactor
	}
	return factor
}

func (i *Weapon) IsEquippable() bool {
	return true
}
func (i *Weapon) IsRepairable() bool {
	return true
}
func (i *Weapon) IsWeapon() bool {
	return true
}

func (i *Weapon) IsRangedWeapon() bool {
	return i.weaponType.IsRanged()
}

func (i *Weapon) IsMeleeWeapon() bool {
	return i.weaponType.IsMelee()
}

func (i *Weapon) DoesJam() bool {
	if i.jammed {
		return true
	}
	if !i.IsAutomaticWeapon() {
		return false
	}

	reliability := int(i.reliability)
	isReliableOnThisShot := rand.Intn(100)+1 <= reliability
	i.jammed = !isReliableOnThisShot
	return i.jammed
}

func (i *Weapon) GetEffectParameters() foundation.Params {
	parameters := i.GenericItem.GetEffectParameters()
	weapon := i
	if !parameters.HasDamage() && i.IsWeapon() {
		damageInterval := i.GetWeaponDamage()
		parameters["damage_interval"] = damageInterval
		parameters["damage"] = damageInterval.Roll()
	}
	if i.IsRangedWeapon() && weapon.NeedsAmmo() && weapon.HasAmmo() {
		ammo := weapon.GetLoadedAmmo()
		ammoInfo := ammo
		if ammoInfo.BonusRadius > 0 {
			parameters["bonus_radius"] = ammoInfo.BonusRadius
		}
	}
	// effect_delay
	// effect_duration
	return parameters
}

func (i *Weapon) GetCurrentAttackMode() AttackMode {
	return i.GetAttackMode(i.currentAttackModeIndex)
}

func (i *Weapon) CycleTargetMode() {
	i.currentAttackModeIndex++
	if i.currentAttackModeIndex >= len(i.attackModes) {
		i.currentAttackModeIndex = 0
	}
}

func (i *Weapon) IsLoadedWeapon() bool {
	return i.IsWeapon() && i.IsLoaded()
}

func (i *Weapon) GetWeaponDamage() fxtools.Interval {
	return i.getRawDamage().Scaled(i.qualityInPercent.Normalized())
}

func (i *Weapon) GetWeaponDamageForCurrentAttackMode() fxtools.Interval {
	perBullet := i.getRawDamage().Scaled(i.qualityInPercent.Normalized())
	return perBullet.Scaled(float64(i.BulletCountForCurrentAttackMode()))
}

func (i *Weapon) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Encode each field of the struct in order
	if err := encoder.Encode(i.damageDice); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.weaponType); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.skillUsed); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.magazineSize); err != nil {
		return nil, err
	}
	if i.loadedInMagazine == nil {
		encoder.Encode(false)
	} else {
		encoder.Encode(true)
		if err := encoder.Encode(i.loadedInMagazine); err != nil {
			return nil, err
		}
	}

	if err := encoder.Encode(i.burstRounds); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.caliberIndex); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.attackModes); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.soundID); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.damageType); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (i *Weapon) GobDecode(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// Decode each field of the struct in order
	if err := decoder.Decode(&i.damageDice); err != nil {
		return err
	}

	if err := decoder.Decode(&i.weaponType); err != nil {
		return err
	}

	if err := decoder.Decode(&i.skillUsed); err != nil {
		return err
	}

	if err := decoder.Decode(&i.magazineSize); err != nil {
		return err
	}

	var hasAmmo bool
	if err := decoder.Decode(&hasAmmo); err != nil {
		return err
	}

	if hasAmmo {
		i.loadedInMagazine = &Ammo{}
		if err := decoder.Decode(i.loadedInMagazine); err != nil {
			return err
		}
	}

	if err := decoder.Decode(&i.burstRounds); err != nil {
		return err
	}

	if err := decoder.Decode(&i.caliberIndex); err != nil {
		return err
	}

	if err := decoder.Decode(&i.attackModes); err != nil {
		return err
	}

	if err := decoder.Decode(&i.soundID); err != nil {
		return err
	}

	if err := decoder.Decode(&i.damageType); err != nil {
		return err
	}

	return nil
}

func (i *Weapon) getRawDamage() fxtools.Interval {
	return i.damageDice
}

func (i *Weapon) GetWeaponType() WeaponType {
	return i.weaponType
}

func (i *Weapon) GetSkillUsed() d100.Skill {
	return i.skillUsed
}

func (i *Weapon) GetCaliber() int {
	return i.caliberIndex
}

func (i *Weapon) BulletsNeededForFullClip() (int, string) {
	if i.loadedInMagazine == nil {
		return i.magazineSize, ""
	}
	ammoKind := i.loadedInMagazine
	return i.magazineSize - i.GetLoadedBullets(), ammoKind.InternalName()
}

func (i *Weapon) LoadAmmo(ammo *Ammo) *Ammo {
	if i.loadedInMagazine == nil {
		i.loadedInMagazine = ammo
		return nil
	}
	if i.loadedInMagazine.CanStackWith(ammo) {
		i.loadedInMagazine.AddStacks(ammo)
		return nil
	}
	oldAmmo := i.loadedInMagazine
	i.loadedInMagazine = ammo
	if oldAmmo.StackSize() > 0 {
		return oldAmmo
	}
	return nil
}

func (i *Weapon) IsRanged() bool {
	return i.weaponType.IsRanged()
}

func (i *Weapon) IsMelee() bool {
	return i.weaponType.IsMelee()
}

func (i *Weapon) HasAmmo() bool {
	return i.GetLoadedBullets() > 0 || !i.NeedsAmmo()
}

func (i *Weapon) GetLoadedBullets() int {
	if i.loadedInMagazine == nil {
		return 0
	}
	return i.loadedInMagazine.StackSize()
}

func (i *Weapon) GetMagazineSize() int {
	return i.magazineSize
}

func (i *Weapon) RemoveBullets(spent int) *Ammo {
	if i.loadedInMagazine == nil {
		return nil
	}
	if spent >= i.loadedInMagazine.StackSize() {
		spentBullets := i.loadedInMagazine
		i.loadedInMagazine = nil
		return spentBullets
	}
	spentBullets := i.loadedInMagazine.Split(spent)
	return spentBullets.(*Ammo)
}

func (i *Weapon) GetBurstRounds() int {
	return i.burstRounds
}

func (i *Weapon) NeedsAmmo() bool {
	return i.caliberIndex > 0
}

func (i *Weapon) GetFireAudioCue(mode TargetingMode) string {
	strMode := "single"
	if (mode == TargetingModeFireBurst || mode == TargetingModeFireFullAuto) &&
		len(i.attackModes) > 1 {
		strMode = "burst"
	}
	return fmt.Sprintf("weapons/%d_%s", i.soundID, strMode)
}

func (i *Weapon) GetReloadAudioCue() string {
	return fmt.Sprintf("weapons/%d_reload", i.soundID)
}
func (i *Weapon) GetOutOfAmmoAudioCue() string {
	return fmt.Sprintf("weapons/%d_out_of_ammo", i.soundID)
}
func (i *Weapon) GetMissAudioCue() string {
	return fmt.Sprintf("weapons/%d_hit_surface", i.soundID)
}

func (i *Weapon) GetDamageType() DamageType {
	return i.damageType
}

func (i *Weapon) GetAttackMode(index int) AttackMode {
	return i.attackModes[index]
}

func (i *Weapon) IsValid() bool {
	return i.weaponType != WeaponTypeUnknown
}

func (i *Weapon) GetLoadedAmmo() *Ammo {
	return i.loadedInMagazine

}

func (i *Weapon) IsLoaded() bool {
	return i.loadedInMagazine != nil && i.loadedInMagazine.StackSize() > 0
}

func (i *Weapon) Unload() *Ammo {
	ammo := i.loadedInMagazine
	i.loadedInMagazine = nil
	return ammo
}

func (i *Weapon) GetTargetDTModifier() int {
	if i.loadedInMagazine == nil {
		return 0
	}
	ammo := i.loadedInMagazine
	if ammo == nil {
		return 0
	}
	return ammo.DTModifier
}

func (i *Weapon) Range() int {
	return i.GetCurrentAttackMode().MaxRange
}

func (i *Weapon) BulletCountForCurrentAttackMode() int {
	switch i.GetCurrentAttackMode().Mode {
	case TargetingModeFireBurst:
		return min(i.GetBurstRounds(), i.GetLoadedBullets())
	case TargetingModeFireFullAuto:
		return min(i.GetMagazineSize(), i.GetLoadedBullets())
	default:
		return 1
	}
}

func (i *Weapon) IsAutomaticWeapon() bool {
	return i.GetBurstRounds() > 1
}

func (i *Weapon) IsBroken() bool {
	return i.qualityInPercent <= 0
}

func (i *Weapon) IsJammed() bool {
	return i.jammed
}

func (i *Weapon) Unjam() {
	i.jammed = false
}

func (i *Weapon) GetAmmoTypeShortString() string {
	if i.loadedInMagazine == nil {
		return ""
	}
	if i.loadedInMagazine.ShortIdentifier == "" {
		return ""
	}
	return fmt.Sprintf(" %s", i.loadedInMagazine.ShortIdentifier)
}

func (i *Weapon) GetCaliberName() string {
	return i.caliberName
}

type WeaponType int

func (t WeaponType) IsMissile() bool {
	return t == WeaponTypeArrow || t == WeaponTypeBolt || t == WeaponTypeDart || t == WeaponTypeMissile || t == WeaponTypeBullet
}

func (t WeaponType) IsRanged() bool {
	return t.IsMissile() || t == WeaponTypeBow || t == WeaponTypeCrossbow || t == WeaponTypePistol || t == WeaponTypeRifle || t == WeaponTypeShotgun || t == WeaponTypeSMG || t == WeaponTypeMinigun || t == WeaponTypeRocketLauncher || t == WeaponTypeBigGun || t == WeaponTypeEnergy
}

func (t WeaponType) IsMelee() bool {
	return t == WeaponTypeSword || t == WeaponTypeClub || t == WeaponTypeAxe || t == WeaponTypeDagger || t == WeaponTypeSpear || t == WeaponTypeKnife || t == WeaponTypeMelee
}

func (t WeaponType) String() string {
	switch t {
	case WeaponTypeSword:
		return "Sword"
	case WeaponTypeClub:
		return "Club"
	case WeaponTypeAxe:
		return "Axe"
	case WeaponTypeDagger:
		return "Dagger"
	case WeaponTypeSpear:
		return "Spear"
	case WeaponTypeBow:
		return "Bow"
	case WeaponTypeArrow:
		return "Arrow"
	case WeaponTypeCrossbow:
		return "Crossbow"
	case WeaponTypeBolt:
		return "Bolt"
	case WeaponTypeDart:
		return "Dart"
	case WeaponTypePistol:
		return "Pistol"
	case WeaponTypeRifle:
		return "Rifle"
	case WeaponTypeShotgun:
		return "Shotgun"
	case WeaponTypeMissile:
		return "Missile"
	case WeaponTypeBullet:
		return "Bullet"
	case WeaponTypeSMG:
		return "SMG"
	case WeaponTypeSledgehammer:
		return "Sledgehammer"
	case WeaponTypeMinigun:
		return "Minigun"
	case WeaponTypeRocketLauncher:
		return "Rocket Launcher"
	case WeaponTypeBigGun:
		return "Big Gun"
	case WeaponTypeKnife:
		return "Knife"
	case WeaponTypeEnergy:
		return "Energy"
	case WeaponTypeThrown:
		return "Thrown"
	case WeaponTypeMelee:
		return "Melee"
	default:
		return "Unknown Weapon Type"
	}
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

type AttackMode struct {
	Mode     TargetingMode
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

func (m AttackMode) IsThrow() bool {
	return m.Mode == TargetingModeThrow
}
func GetAttackModes(targetModes [2]TargetingMode, tuCost [2]int, maxRange [2]int, noAim bool) []AttackMode {
	var modes []AttackMode
	if targetModes[0] != TargetingModeNone {
		modes = append(modes, AttackMode{
			Mode:     targetModes[0],
			TUCost:   tuCost[0],
			MaxRange: maxRange[0],
			IsAimed:  false,
		})
		if !noAim && targetModes[0].IsAimable() {
			modes = append(modes, AttackMode{
				Mode:     targetModes[0],
				TUCost:   tuCost[0] + 2,
				MaxRange: maxRange[0],
				IsAimed:  true,
			})
		}
	}
	if targetModes[1] != TargetingModeNone {
		modes = append(modes, AttackMode{
			Mode:     targetModes[1],
			TUCost:   tuCost[1],
			MaxRange: maxRange[1],
			IsAimed:  false,
		})
		if !noAim && targetModes[1].IsAimable() {
			modes = append(modes, AttackMode{
				Mode:     targetModes[1],
				TUCost:   tuCost[1] + 2,
				MaxRange: maxRange[1],
				IsAimed:  true,
			})
		}
	}
	return modes
}
