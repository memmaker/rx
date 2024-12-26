package game

import (
	"contractor/d100"
	"contractor/foundation"
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
	DamageDice fxtools.Interval
	WeaponType WeaponType

	SkillUsed        d100.Skill
	MagazineSize     int
	LoadedInMagazine *Ammo

	PelletCount  int
	BurstRounds  int
	CaliberIndex int
	CaliberName  string

	AttackModes []AttackMode
	SoundID     int32
	DamageType  DamageType
	MinSTR      int

	Reliability  d100.Percentage
	RelativeSize WeaponSize
	AccuracyMod  d100.Percentage

	DegradeFactor float64
	Jammed        bool

	CurrentAttackModeIndex int
}

func (i *Weapon) Degrade(degrade float64) {
	ammoFactor := i.ammoDegradeFactor()
	weaponFactor := i.DegradeFactor
	i.QualityInPercent = max(0, i.QualityInPercent-d100.Percentage(degrade*weaponFactor*ammoFactor))
}

func (i *Weapon) FullDescription(colorCode string) string {
	basicRows := i.GenericItem.fullDescriptionRows()

	basicRows = append(basicRows, fxtools.NewTableRow("Type", i.WeaponType.String()))
	basicRows = append(basicRows, fxtools.NewTableRow("Caliber", i.GetCaliberName()))
	basicRows = append(basicRows, fxtools.NewTableRow("Quality", fmt.Sprintf("%d%%", int(i.QualityInPercent))))
	basicRows = append(basicRows, fxtools.NewTableRow("Concealability", i.RelativeSize.String()))

	if i.DamageType == DamageTypeNormal {
		basicRows = append(basicRows, fxtools.NewTableRow("Damage", fmt.Sprintf("%s", i.DamageDice.Scaled(i.QualityInPercent.Normalized()).ShortString())))
	} else {
		basicRows = append(basicRows, fxtools.NewTableRow("Damage", fmt.Sprintf("%s (%s)", i.DamageDice.Scaled(i.QualityInPercent.Normalized()).ShortString(), i.DamageType.String())))
	}

	if i.AccuracyMod != 0 {
		basicRows = append(basicRows, fxtools.NewTableRow("Accuracy Modifier", fmt.Sprintf("%+d%%", int(i.AccuracyMod))))
	}
	if i.MinSTR > 0 {
		basicRows = append(basicRows, fxtools.NewTableRow("Min. Strength", fmt.Sprintf("%d", i.MinSTR)))
	}

	basicRows = append(basicRows, fxtools.NewTableRow("Reliability", fmt.Sprintf("%d%%", int(i.Reliability))))

	for _, attackMode := range i.AttackModes {
		modeVal := fmt.Sprintf("Range: %d, TU: %d", attackMode.MaxRange, attackMode.TUCost)
		modeLabel := attackMode.String()
		basicRows = append(basicRows, fxtools.NewTableRow(modeLabel, modeVal))
	}

	lines := fxtools.TableLayout(basicRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	lines = append([]string{i.InventoryNameWithColors(colorCode), i.Category.String()}, lines...)

	lines = i.appendText(lines)

	return strings.Join(lines, "\n")
}
func (i *Weapon) InventoryNameWithColorsAndShortcut(lineColorCode string) string {
	return fmt.Sprintf("%c - %s", i.Shortcut(), i.InventoryNameWithColors(lineColorCode))
}
func (i *Weapon) InventoryNameWithColors(colorCode string) string {
	line := cview.Escape(fmt.Sprintf("%s [%s]", i.Name(), i.GetWeaponDamage().ShortString()))

	lineWithColor := colorCode + line + "[-]"

	qIcon := getQualityIcon(i.QualityInPercent)
	lineWithColor = fmt.Sprintf("%s %s", qIcon, lineWithColor)

	return lineWithColor
}
func (i *Weapon) LongNameWithColors(colorCode string) string {
	weapon := i
	attackMode := weapon.GetAttackMode(i.CurrentAttackModeIndex)
	targetMode := attackMode.String()
	if i.Jammed {
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
	return i.WeaponType.IsRanged()
}

func (i *Weapon) IsMeleeWeapon() bool {
	return i.WeaponType.IsMelee()
}

func (i *Weapon) DoesJam() bool {
	if i.Jammed {
		return true
	}
	if !i.IsAutomaticWeapon() {
		return false
	}

	reliability := int(i.Reliability)
	isReliableOnThisShot := rand.Intn(100)+1 <= reliability
	i.Jammed = !isReliableOnThisShot
	return i.Jammed
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
	return i.GetAttackMode(i.CurrentAttackModeIndex)
}

func (i *Weapon) CycleTargetMode() {
	i.CurrentAttackModeIndex++
	if i.CurrentAttackModeIndex >= len(i.AttackModes) {
		i.CurrentAttackModeIndex = 0
	}
}

func (i *Weapon) IsLoadedWeapon() bool {
	return i.IsWeapon() && i.IsLoaded()
}

func (i *Weapon) GetWeaponDamage() fxtools.Interval {
	return i.getRawDamage().Scaled(i.QualityInPercent.Normalized())
}

func (i *Weapon) GetWeaponDamageForCurrentAttackMode() fxtools.Interval {
	perBullet := i.getRawDamage().Scaled(i.QualityInPercent.Normalized())
	return perBullet.Scaled(float64(i.BulletCountForCurrentAttackMode()))
}

func (i *Weapon) getRawDamage() fxtools.Interval {
	return i.DamageDice
}

func (i *Weapon) GetWeaponType() WeaponType {
	return i.WeaponType
}

func (i *Weapon) GetSkillUsed() d100.Skill {
	return i.SkillUsed
}

func (i *Weapon) GetCaliber() int {
	return i.CaliberIndex
}

func (i *Weapon) BulletsNeededForFullClip() (int, string) {
	if i.LoadedInMagazine == nil {
		return i.MagazineSize, ""
	}
	ammoKind := i.LoadedInMagazine
	return i.MagazineSize - i.GetLoadedBullets(), ammoKind.GetInternalName()
}

func (i *Weapon) LoadAmmo(ammo *Ammo) *Ammo {
	if i.LoadedInMagazine == nil {
		i.LoadedInMagazine = ammo
		return nil
	}
	if i.LoadedInMagazine.CanStackWith(ammo) {
		i.LoadedInMagazine.AddStacks(ammo)
		return nil
	}
	oldAmmo := i.LoadedInMagazine
	i.LoadedInMagazine = ammo
	if oldAmmo.GetStackSize() > 0 {
		return oldAmmo
	}
	return nil
}

func (i *Weapon) IsRanged() bool {
	return i.WeaponType.IsRanged()
}

func (i *Weapon) IsMelee() bool {
	return i.WeaponType.IsMelee()
}

func (i *Weapon) HasAmmo() bool {
	return i.GetLoadedBullets() > 0 || !i.NeedsAmmo()
}

func (i *Weapon) GetLoadedBullets() int {
	if i.LoadedInMagazine == nil {
		return 0
	}
	return i.LoadedInMagazine.GetStackSize()
}

func (i *Weapon) GetMagazineSize() int {
	return i.MagazineSize
}

func (i *Weapon) RemoveBullets(spent int) *Ammo {
	if i.LoadedInMagazine == nil {
		return nil
	}
	if spent >= i.LoadedInMagazine.GetStackSize() {
		spentBullets := i.LoadedInMagazine
		i.LoadedInMagazine = nil
		return spentBullets
	}
	spentBullets := i.LoadedInMagazine.Split(spent)
	return spentBullets.(*Ammo)
}

func (i *Weapon) GetBurstRounds() int {
	return i.BurstRounds
}

func (i *Weapon) NeedsAmmo() bool {
	return i.CaliberIndex > 0
}

func (i *Weapon) GetFireAudioCue(mode TargetingMode) string {
	strMode := "single"
	if (mode == TargetingModeFireBurst || mode == TargetingModeFireFullAuto) &&
		len(i.AttackModes) > 1 {
		strMode = "burst"
	}
	return fmt.Sprintf("weapons/%d_%s", i.SoundID, strMode)
}

func (i *Weapon) GetReloadAudioCue() string {
	return fmt.Sprintf("weapons/%d_reload", i.SoundID)
}
func (i *Weapon) GetOutOfAmmoAudioCue() string {
	return fmt.Sprintf("weapons/%d_out_of_ammo", i.SoundID)
}
func (i *Weapon) GetMissAudioCue() string {
	return fmt.Sprintf("weapons/%d_hit_surface", i.SoundID)
}

func (i *Weapon) GetDamageType() DamageType {
	return i.DamageType
}

func (i *Weapon) GetAttackMode(index int) AttackMode {
	return i.AttackModes[index]
}

func (i *Weapon) IsValid() bool {
	return i.WeaponType != WeaponTypeUnknown
}

func (i *Weapon) GetLoadedAmmo() *Ammo {
	return i.LoadedInMagazine

}

func (i *Weapon) IsLoaded() bool {
	return i.LoadedInMagazine != nil && i.LoadedInMagazine.GetStackSize() > 0
}

func (i *Weapon) Unload() *Ammo {
	ammo := i.LoadedInMagazine
	i.LoadedInMagazine = nil
	return ammo
}

func (i *Weapon) GetTargetDTModifier() int {
	if i.LoadedInMagazine == nil {
		return 0
	}
	ammo := i.LoadedInMagazine
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
	return i.QualityInPercent <= 0
}

func (i *Weapon) IsJammed() bool {
	return i.Jammed
}

func (i *Weapon) Unjam() {
	i.Jammed = false
}

func (i *Weapon) GetAmmoTypeShortString() string {
	if i.LoadedInMagazine == nil {
		return ""
	}
	if i.LoadedInMagazine.ShortIdentifier == "" {
		return ""
	}
	return fmt.Sprintf(" %s", i.LoadedInMagazine.ShortIdentifier)
}

func (i *Weapon) GetCaliberName() string {
	return i.CaliberName
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

type TargetingMode int

const (
	TargetingModeNone TargetingMode = iota
	TargetingModePunch
	TargetingModeKick
	TargetingModeSwing
	TargetingModeThrust
	TargetingModeThrow
	TargetingModeFireSingle
	TargetingModeFireBurst
	TargetingModeFireFullAuto
	TargetingModeFlame

	// unique for bowel disruptor
	TargetingModeBDLoose
	TargetingModeBDWatery
	TargetingModeBDFiery
	TargetingModeBDBurningAnalGeyser
	TargetingModeBDRectalVolcano
	TargetingModeBDProlapse
	TargetingModeBDUnspeakableGutHorror
	TargetingModeBDShatIntoUnconsciousness
	TargetingModeBDFatalIntestinalMaelstrom

	TargetingModeCount
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

	case "bd_loose":
		return TargetingModeBDLoose
	case "bd_watery":
		return TargetingModeBDWatery
	case "bd_fiery":
		return TargetingModeBDFiery
	case "bd_burning_anal_geyser":
		return TargetingModeBDBurningAnalGeyser
	case "bd_rectal_volcano":
		return TargetingModeBDRectalVolcano
	case "bd_prolapse":
		return TargetingModeBDProlapse
	case "bd_unspeakable_gut_horror":
		return TargetingModeBDUnspeakableGutHorror
	case "bd_shat_into_unconsciousness":
		return TargetingModeBDShatIntoUnconsciousness
	case "bd_fatal_intestinal_maelstrom":
		return TargetingModeBDFatalIntestinalMaelstrom
	}
	panic("Invalid targeting mode: " + value)
	return TargetingModeNone
}
func (t TargetingMode) Next() TargetingMode {
	if t == TargetingModeNone {
		return TargetingModePunch
	}
	nextVal := t + 1
	if nextVal >= TargetingModeCount {
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
		return "Fire Single"
	case TargetingModeFireBurst:
		return "Fire Burst"
	case TargetingModeFlame:
		return "Flame"
	case TargetingModeBDLoose:
		return "Loose"
	case TargetingModeBDWatery:
		return "Watery"
	case TargetingModeBDFiery:
		return "Fiery"
	case TargetingModeBDBurningAnalGeyser:
		return "Burning Anal Geyser"
	case TargetingModeBDRectalVolcano:
		return "Rectal Volcano"
	case TargetingModeBDProlapse:
		return "Prolapse"
	case TargetingModeBDUnspeakableGutHorror:
		return "Unspeakable Gut Horror"
	case TargetingModeBDShatIntoUnconsciousness:
		return "Shat Into Unconsciousness"
	case TargetingModeBDFatalIntestinalMaelstrom:
		return "Fatal Intestinal Maelstrom"
	}
	return "Unknown"
}

func (t TargetingMode) IsMelee() bool {
	return t == TargetingModePunch || t == TargetingModeKick || t == TargetingModeSwing || t == TargetingModeThrust || t == TargetingModeThrow
}

func (t TargetingMode) IsAimable() bool {
	return t == TargetingModeFireSingle || t == TargetingModePunch || t == TargetingModeKick || t == TargetingModeSwing || t == TargetingModeThrust || t == TargetingModeThrow
}

func (t TargetingMode) IsBurstOrFullAuto() bool {
	return t == TargetingModeFireBurst || t == TargetingModeFireFullAuto
}

func (t TargetingMode) IsFullAuto() bool {
	return t == TargetingModeFireFullAuto
}

type DamageType int32

func (t DamageType) IsEnergy() bool {
	return t == DamageTypeLaser || t == DamageTypePlasma || t == DamageTypeElectrical || t == DamageTypeEMP

}

func (t DamageType) String() string {
	switch t {
	case DamageTypeNormal:
		return "Normal"
	case DamageTypeLaser:
		return "Laser"
	case DamageTypeFire:
		return "Fire"
	case DamageTypePlasma:
		return "Plasma"
	case DamageTypeElectrical:
		return "Electrical"
	case DamageTypeEMP:
		return "EMP"
	case DamageTypeExplosive:
		return "Explosive"
	case DamageTypeRadiation:
		return "Radiation"
	case DamageTypePoison:
		return "Poison"
	}
	return "Unknown"
}

const (
	DamageTypeNormal DamageType = iota
	DamageTypeLaser
	DamageTypeFire
	DamageTypePlasma
	DamageTypeElectrical
	DamageTypeEMP
	DamageTypeExplosive
	DamageTypeRadiation
	DamageTypePoison
	// DamageTypeEnergy is a catch-all for all energy damage types on armors
	DamageTypeEnergy
	DamageTypeCount
)

func DamageTypeFromString(value string) DamageType {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "normal":
		return DamageTypeNormal
	case "laser":
		return DamageTypeLaser
	case "fire":
		return DamageTypeFire
	case "plasma":
		return DamageTypePlasma
	case "electrical":
		return DamageTypeElectrical
	case "emp":
		return DamageTypeEMP
	case "explosive":
		return DamageTypeExplosive
	case "radiation":
		return DamageTypeRadiation
	case "poison":
		return DamageTypePoison
	}
	panic("Invalid damage type: " + value)
	return DamageTypeNormal
}
