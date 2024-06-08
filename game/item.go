package game

import (
	"RogueUI/cview"
	"RogueUI/dice_curve"
	"RogueUI/foundation"
	"RogueUI/geometry"
	"fmt"
	"image/color"
)

type Item struct {
	name          string
	internalName  string
	position      geometry.Point
	category      foundation.ItemCategory
	weapon        *WeaponInfo
	armor         *ArmorInfo
	ammo          *AmmoInfo
	useEffectName string
	zapEffectName string
	charges       int
	slot          foundation.EquipSlot

	stat       dice_curve.Stat
	statBonus  int
	skill      dice_curve.SkillName
	skillBonus int

	equipFlag    foundation.ActorFlag
	thrownDamage dice_curve.Dice
}

func (g *GameState) NewItemFromName(name string) *Item {
	def := g.dataDefinitions.GetItemDefByName(name)
	return NewItem(def)
}
func NewItem(def ItemDef) *Item {
	charges := 1
	if def.Charges.NotZero() {
		charges = def.Charges.Roll()
	}
	item := &Item{
		name:         def.Name,
		internalName: def.InternalName,
		category:     def.Category,
		charges:      charges,
		slot:         def.Slot,
		stat:         def.Stat,
		statBonus:    def.StatBonus.Roll(),
		skill:        def.Skill,
		skillBonus:   def.SkillBonus.Roll(),
		equipFlag:    def.EquipFlag,
		thrownDamage: def.ThrowDamageDice,
	}

	if def.IsValidAmmo() {
		item.ammo = &AmmoInfo{
			damage: def.AmmoDef.Damage,
			kind:   def.AmmoDef.Kind,
		}
	}

	if def.IsValidWeapon() {
		item.weapon = &WeaponInfo{
			damageDice:       def.WeaponDef.Damage,
			weaponType:       def.WeaponDef.Type,
			usesAmmo:         def.WeaponDef.UsesAmmo,
			skillUsed:        def.WeaponDef.SkillUsed,
			magazineSize:     def.WeaponDef.MagazineSize,
			burstRounds:      def.WeaponDef.BurstRounds,
			loadedInMagazine: nil,
			qualityInPercent: 100,
			targetingMode:    def.WeaponDef.TargetingMode,
		}
		item.weapon.CycleTargetMode()
		item.slot = foundation.SlotNameMainHand
	}

	if def.IsValidArmor() {
		item.armor = &ArmorInfo{
			protection:         def.ArmorDef.Protection,
			radiationReduction: def.ArmorDef.RadiationReduction,
			encumbrance:        def.ArmorDef.Encumbrance,
			durability:         100,
		}
	}

	item.zapEffectName = def.ZapEffect
	item.useEffectName = def.UseEffect

	return item

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

func (i *Item) LongNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())
	if i.IsWeapon() {
		weapon := i.weapon
		targetMode := weapon.GetCurrentTargetingMode().ToString()
		timeNeeded := weapon.GetTimeNeeded()
		bullets := fmt.Sprintf("%d/%d", weapon.GetLoadedBullets(), weapon.GetMagazineSize())
		line = cview.Escape(fmt.Sprintf("%s (%s: %d TU / %s Dmg.) - %s", i.Name(), targetMode, timeNeeded, weapon.GetDamage().ShortString(), bullets))
	}
	if i.IsArmor() {
		line = cview.Escape(fmt.Sprintf("%s [%s]", i.Name(), i.armor.GetProtectionValueAsString()))
	}
	if i.IsRing() && i.charges > 1 {
		line = cview.Escape(fmt.Sprintf("%s (%d turns)", i.Name(), i.charges))
	}
	return colorCode + line + "[-]"
}

func (i *Item) InventoryNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())
	if i.IsWeapon() {
		weapon := i.weapon
		line = cview.Escape(fmt.Sprintf("%s (%s)", i.Name(), weapon.GetDamage().ShortString()))
	}
	if i.IsArmor() {
		line = cview.Escape(fmt.Sprintf("%s [%s]", i.Name(), i.armor.GetProtectionValueAsString()))
	}
	if i.IsRing() && i.charges > 1 {
		line = cview.Escape(fmt.Sprintf("%s (%d turns)", i.Name(), i.charges))
	}
	if i.IsAmmo() {
		line = cview.Escape(fmt.Sprintf("%s x%d", i.Name(), i.GetCharges()))
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

	if i.statBonus != 0 {
		name = fmt.Sprintf("%s [%+d %s]", name, i.statBonus, i.stat.ToShortString())
	}

	if i.skillBonus != 0 {
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

	if i.internalName != other.internalName {
		return false
	}
	if (i.IsWeapon() && !i.IsMissile()) || i.IsArmor() || i.IsRing() || (other.IsWeapon() && !other.IsMissile()) || other.IsArmor() || other.IsRing() {
		return false
	}

	if i.useEffectName != other.useEffectName || i.zapEffectName != other.zapEffectName {
		return false
	}

	if i.IsAmmo() && other.IsAmmo() && i.ammo.damage == other.ammo.damage {
		return true
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
	return i.IsWeapon() && i.GetWeapon().IsMelee()
}

func (i *Item) IsRangedWeapon() bool {
	return i.IsWeapon() && i.GetWeapon().IsRanged()
}

func (i *Item) IsArmor() bool {
	return i.armor != nil
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

	isMagicRing := i.IsRing()

	return isConsumableMagic || isMagicRing
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

func (i *Item) GetStatBonus(stat dice_curve.Stat) int {

	if i.stat == stat {
		return i.statBonus
	}
	return 0
}
func (i *Item) GetSkillBonus(skill dice_curve.SkillName) int {
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

func (i *Item) GetThrowDamageDice() dice_curve.Dice {
	return i.thrownDamage
}

func (i *Item) ConsumeCharge() {
	i.charges--
}

func (i *Item) SetCharges(amount int) {
	i.charges = amount
}

func (i *Item) AfterEquippedTurn() {
	if (i.IsRing()) && i.charges > 0 {
		i.charges--
	}
}

func (i *Item) IsAmmo() bool {
	return i.ammo != nil
}

func (i *Item) IsAmmoOfCaliber(ammo string) bool {
	return i.IsAmmo() && i.ammo.kind == ammo
}

func (i *Item) RemoveCharges(spent int) {
	i.charges -= spent
	if i.charges < 0 {
		i.charges = 0
	}
}

func (i *Item) Split(bullets int) *Item {
	if bullets >= i.charges {
		return i
	}
	clone := *i
	clone.charges = bullets
	i.charges -= bullets
	return &clone
}

func (i *Item) Merge(ammo *Item) {
	i.charges += ammo.charges
}

func (i *Item) GetAmmo() *AmmoInfo {
	return i.ammo
}
