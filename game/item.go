package game

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/rpg"
	"code.rocketnine.space/tslocum/cview"
	"fmt"
	"image/color"
	"math/rand"
)

type WeaponInfo struct {
	damageDice       rpg.Dice
	damagePlus       int
	weaponType       WeaponType
	launchedWithType WeaponType
	vorpalEnemy      string
	skillUsed        rpg.SkillName
}

func (i *WeaponInfo) GetVorpalEnemy() string {
	return i.vorpalEnemy
}

func (i *WeaponInfo) Vorpalize(enemy string) {
	i.vorpalEnemy = enemy
}
func (i *WeaponInfo) GetDamagePlus() int {
	return i.damagePlus
}

func (i *WeaponInfo) GetDamageDice() rpg.Dice {
	return i.damageDice.WithBonus(i.damagePlus)
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
func (i *WeaponInfo) IsEnchantable() bool {
	return i.damagePlus <= 7
}

func (i *WeaponInfo) AddEnchantment() {
	i.damagePlus++
}

func (i *WeaponInfo) IsEnchanted() bool {
	return i.damagePlus > 0
}

func (i *WeaponInfo) IsLaunchedWith(category WeaponType) bool {
	return i.launchedWithType == category
}

func (i *WeaponInfo) GetWeaponType() WeaponType {
	return i.weaponType
}

func (i *WeaponInfo) GetSkillUsed() rpg.SkillName {
	return i.skillUsed
}

func (i *WeaponInfo) IsVorpal() bool {
	return i.vorpalEnemy != ""
}

type ArmorInfo struct {
	damageResistance int
	plus             int
	encumbrance      rpg.Encumbrance
}

func (i *ArmorInfo) GetArmorClass() int {
	return i.damageResistance
}

func (i *ArmorInfo) GetDamageResistanceWithPlus() int {
	return i.damageResistance + i.plus
}

func (i *ArmorInfo) IsEnchantable() bool {
	return i.plus <= 7
}

func (i *ArmorInfo) AddEnchantment() {
	i.plus++
}

func (i *ArmorInfo) IsEnchanted() bool {
	return i.plus > 0
}

func (i *ArmorInfo) GetEncumbrance() rpg.Encumbrance {
	return i.encumbrance
}

type Item struct {
	name          string
	internalName  string
	position      geometry.Point
	category      foundation.ItemCategory
	weapon        *WeaponInfo
	armor         *ArmorInfo
	useEffectName string
	zapEffectName string
	charges       int
	slot          foundation.EquipSlot

	id         *IdentificationKnowledge
	stat       rpg.Stat
	statBonus  int
	skill      rpg.SkillName
	skillBonus int

	equipFlag    foundation.ActorFlag
	thrownDamage rpg.Dice
	isKnown      bool
}

func (i *Item) InventoryNameWithColorsAndShortcut(lineColorCode string) string {
	return fmt.Sprintf("%c - %s", i.Shortcut(), i.InventoryNameWithColors(lineColorCode))
}

func (i *Item) Shortcut() rune {
	return -1
}

func (i *Item) DisplayLength() int {
	return cview.TaggedStringWidth(i.InventoryNameWithColorsAndShortcut(""))
}

func (i *Item) GetListInfo() string {
	return fmt.Sprintf("%s", i.name)
}

func (i *Item) InventoryNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())
	if i.IsWeapon() {
		line = cview.Escape(fmt.Sprintf("%s (%s)", i.Name(), i.weapon.GetDamageDice().ShortString()))
	}
	if i.IsArmor() {
		line = cview.Escape(fmt.Sprintf("%s [%+d]", i.Name(), i.armor.GetDamageResistanceWithPlus()))
	}
	if i.IsRing() && i.charges > 1 && i.id.IsItemIdentified(i.internalName) {
		line = cview.Escape(fmt.Sprintf("%s (%d turns)", i.Name(), i.charges))
	}
	return colorCode + line + "[-]"
}

func (i *Item) SetPosition(pos geometry.Point) {
	i.position = pos
}

func (i *Item) Position() geometry.Point {
	return i.position
}

func (i *Item) Name() string {
	name := i.name
	if i.IsGold() {
		name = fmt.Sprintf("%d gold", i.charges)
	}

	if i.IsPotion() && !i.id.IsItemIdentified(i.internalName) {
		flavor := i.id.GetPotionColor(i.internalName)
		name = fmt.Sprintf("%s potion", flavor)
	}

	if i.IsScroll() && !i.id.IsItemIdentified(i.internalName) {
		flavor := i.id.GetScrollName(i.internalName)
		name = fmt.Sprintf("scroll of '%s'", flavor)
	}

	if i.IsWand() && !i.id.IsItemIdentified(i.internalName) {
		flavor := i.id.GetWandMaterial(i.internalName)
		name = fmt.Sprintf("%s wand", flavor)
	}

	if i.IsRing() && !i.id.IsItemIdentified(i.internalName) {
		flavor := i.id.GetRingStone(i.internalName)
		name = fmt.Sprintf("%s ring", flavor)
	}

	if i.IsStuck() && i.isKnown {
		name = fmt.Sprintf("*%d* %s", i.GetCharges(), name)
	}

	if i.statBonus != 0 && (i.isKnown ||i.id.IsItemIdentified(i.internalName)) {
		name = fmt.Sprintf("%s [%+d %s]", name, i.statBonus, i.stat.ToShortString())
	}

	if i.skillBonus != 0 && (i.isKnown ||i.id.IsItemIdentified(i.internalName)) {
		name = fmt.Sprintf("%s [%+d %s]", name, i.skillBonus, i.skill.ToShortString())
	}

	return name
}

func (i *Item) IsThrowable() bool {
	return true
}

func (i *Item) IsUsableOrZappable() bool {
	return i.useEffectName != "" || i.zapEffectName != ""
}

func (i *Item) IsUsable() bool {
	return i.useEffectName != ""
}

func (i *Item) GetUseEffectName() string {
	return i.useEffectName
}

func (i *Item) GetZapEffectName() string {
	return i.zapEffectName
}

func (i *Item) IsZappable() bool {
	return i.zapEffectName != ""
}

func (i *Item) Color() color.RGBA {
	return color.RGBA{255, 255, 255, 255}
}

func (i *Item) CanStackWith(other *Item) bool {
	if i.name != other.name || i.category != other.category {
		return false
	}

	if (i.IsWeapon() && !i.IsMissile()) || i.IsArmor() || i.IsRing() || (other.IsWeapon() && !other.IsMissile()) || other.IsArmor() || other.IsRing() {
		return false
	}

	if i.useEffectName != other.useEffectName || i.zapEffectName != other.zapEffectName {
		return false
	}

	if i.charges != other.charges {
		return false
	}

	return true
}

func (i *Item) SlotName() foundation.EquipSlot {
	return i.slot
}

func (i *Item) IsEquippable() bool {
	return i.slot != foundation.SlotNameNotEquippable
}

func (i *Item) IsMeleeWeapon() bool {
	return i.IsWeapon() && (i.slot == foundation.SlotNameOneHandedWeapon || i.slot == foundation.SlotNameTwoHandedWeapon)
}

func (i *Item) IsRangedWeapon() bool {
	return i.IsWeapon() && i.slot == foundation.SlotNameMissileLauncher
}

func (i *Item) IsArmor() bool {
	return i.armor != nil
}

func (i *Item) IsTwoHandedWeapon() bool {
	return i.IsWeapon() && i.slot == foundation.SlotNameTwoHandedWeapon
}

func (i *Item) IsWeapon() bool {
	return i.weapon != nil
}

func (i *Item) IsShield() bool {
	return i.IsArmor() && i.slot == foundation.SlotNameShield
}

func (i *Item) GetCategory() foundation.ItemCategory {
	return i.category
}

func (i *Item) GetWeapon() *WeaponInfo {
	return i.weapon
}

func (i *Item) GetArmor() *ArmorInfo {
	return i.armor
}

func (i *Item) IsPotion() bool {
	return i.category == foundation.ItemCategoryPotions
}

func (i *Item) IsGold() bool {
	return i.category == foundation.ItemCategoryGold
}

func (i *Item) GetCharges() int {
	return i.charges
}

func (i *Item) IsFood() bool {
	return i.category == foundation.ItemCategoryFood
}

func (i *Item) IsMagic() bool { // potions, scrolls, wands & weapons/armor with plusses

	isConsumableMagic := i.IsPotion() || i.IsWand() || i.IsScroll()

	isEnchantedWeapon := i.IsWeapon() && i.weapon.IsEnchanted()

	isEnchantedArmor := i.IsArmor() && i.armor.IsEnchanted()

	isMagicRing := i.IsRing()

	return isConsumableMagic || isEnchantedWeapon || isEnchantedArmor || isMagicRing
}

func (i *Item) IsWand() bool {
	return i.category == foundation.ItemCategoryWands
}

func (i *Item) IsScroll() bool {
	return i.category == foundation.ItemCategoryScrolls
}

func (i *Item) IsRing() bool {
	return i.category == foundation.ItemCategoryRings
}

func (i *Item) GetInternalName() string {
	return i.internalName
}

func (i *Item) GetStatBonus(stat rpg.Stat) int {

	if i.stat == stat {
		return i.statBonus
	}
	return 0
}
func (i *Item) GetSkillBonus(skill rpg.SkillName) int {
	if i.skill == skill {
		return i.skillBonus
	}
	return 0
}
func (i *Item) GetEquipFlag() foundation.ActorFlag {
	if i.IsRing() && i.charges == 0 {
		return foundation.FlagNone
	}
	return i.equipFlag
}

func (i *Item) IsMissile() bool {
	return i.IsWeapon() && i.GetWeapon().GetWeaponType().IsMissile()
}

func (i *Item) GetThrowDamageDice() rpg.Dice {
	return i.thrownDamage
}

func (i *Item) ConsumeCharge() {
	i.charges--
}

func (i *Item) SetCharges(amount int) {
	i.charges = amount
}

func (i *Item) AfterEquippedTurn() {
	if (i.IsRing() || i.IsStuck()) && i.charges > 0 {
		i.charges--
		i.isKnown = true
	}
}

func (i *Item) IsStuck() bool {
	return i.equipFlag == foundation.FlagCurseStuck && i.charges > 0
}

func (i *Item) IsCursed() bool {
	return i.IsStuck() || i.statBonus < 0 || i.skillBonus < 0
}

func (i *Item) RemoveCurse() {
	if !i.IsCursed() {
		return
	}
	blessing := rand.Intn(100) < 4

	if i.IsStuck() {
		i.equipFlag = foundation.FlagNone
	}
	if i.statBonus < 0 {
		if blessing {
			i.statBonus = rand.Intn(3) + 1
		} else {
			i.statBonus = 0
		}
	}
	if i.skillBonus < 0 {
		if blessing {
			i.skillBonus = rand.Intn(3) + 1
		} else {
			i.skillBonus = 0
		}
	}
}
