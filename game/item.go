package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"strconv"
	"strings"
)

type Item struct {
	description  string
	internalName string
	position     geometry.Point
	category     foundation.ItemCategory

	qualityInPercent special.Percentage

	weapon        *WeaponInfo
	armor         *ArmorInfo
	ammo          *AmmoInfo
	useEffectName string
	zapEffectName string
	charges       int

	stat       special.Stat
	statBonus  int
	skill      special.Skill
	skillBonus int

	equipFlag    special.ActorFlag
	thrownDamage fxtools.Interval
	tags         foundation.ItemTags
	textFile     string
	text         string
	lockFlag     string

	icon                   textiles.TextIcon
	chanceToBreakOnThrow   int
	currentAttackModeIndex int
	setFlagOnPickup        string
	setFlagOnDrop          string
	size                   int
	weight                 int
	cost                   int
	posHandler             func() geometry.Point
	alive                  bool
	effectParameters       Params
}

func (i *Item) String() string {
	return fmt.Sprintf("Item: %s(%d)", i.internalName, i.charges)
}

func (i *Item) ShouldActivate(tickCount int) bool {
	return i.charges == tickCount
}

func (i *Item) IsAlive(tickCount int) bool {
	return tickCount <= i.charges && i.alive
}

func (i *Item) IsMultipleStacks() bool {
	return i.charges > 1 &&
		(i.IsStackingWithCharges())
}

func (i *Item) GetStackSize() int {
	if i.IsMultipleStacks() {
		return i.charges
	}
	return 1
}

// GobEncode encodes the Item struct into a byte slice.
func (i *Item) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Encode each field of the struct in order
	if err := encoder.Encode(i.description); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.internalName); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.position); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.category); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.qualityInPercent); err != nil {
		return nil, err
	}

	conditionalEncode := func(cond bool, value any) error {
		if cond {
			err := encoder.Encode(true)
			if err != nil {
				return err
			}
			return encoder.Encode(value)
		} else {
			return encoder.Encode(false)
		}
	}

	if err := conditionalEncode(i.weapon != nil, i.weapon); err != nil {
		return nil, err
	}
	if err := conditionalEncode(i.armor != nil, i.armor); err != nil {
		return nil, err
	}
	if err := conditionalEncode(i.ammo != nil, i.ammo); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.useEffectName); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.zapEffectName); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.charges); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.stat); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.statBonus); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.skill); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.skillBonus); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.equipFlag); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.thrownDamage); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.tags); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.textFile); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.text); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.lockFlag); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.icon); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.chanceToBreakOnThrow); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.currentAttackModeIndex); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.setFlagOnPickup); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.size); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.weight); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.cost); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.alive); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GobDecode decodes a byte slice into an Item struct.
func (i *Item) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)

	// Decode each field of the struct in order
	if err := decoder.Decode(&i.description); err != nil {
		return err
	}
	if err := decoder.Decode(&i.internalName); err != nil {
		return err
	}
	if err := decoder.Decode(&i.position); err != nil {
		return err
	}
	if err := decoder.Decode(&i.category); err != nil {
		return err
	}
	if err := decoder.Decode(&i.qualityInPercent); err != nil {
		return err
	}

	conditionalDecode := func(value any) error {
		hasComponent := false
		if err := decoder.Decode(&hasComponent); err != nil {
			return err
		}
		if hasComponent {
			if err := decoder.Decode(value); err != nil {
				return err
			}
		}
		return nil
	}

	if err := conditionalDecode(&i.weapon); err != nil {
		return err
	}

	if err := conditionalDecode(&i.armor); err != nil {
		return err
	}

	if err := conditionalDecode(&i.ammo); err != nil {
		return err
	}

	if err := decoder.Decode(&i.useEffectName); err != nil {
		return err
	}
	if err := decoder.Decode(&i.zapEffectName); err != nil {
		return err
	}
	if err := decoder.Decode(&i.charges); err != nil {
		return err
	}

	if err := decoder.Decode(&i.stat); err != nil {
		return err
	}
	if err := decoder.Decode(&i.statBonus); err != nil {
		return err
	}
	if err := decoder.Decode(&i.skill); err != nil {
		return err
	}
	if err := decoder.Decode(&i.skillBonus); err != nil {
		return err
	}

	if err := decoder.Decode(&i.equipFlag); err != nil {
		return err
	}
	if err := decoder.Decode(&i.thrownDamage); err != nil {
		return err
	}
	if err := decoder.Decode(&i.tags); err != nil {
		return err
	}
	if err := decoder.Decode(&i.textFile); err != nil {
		return err
	}
	if err := decoder.Decode(&i.text); err != nil {
		return err
	}
	if err := decoder.Decode(&i.lockFlag); err != nil {
		return err
	}

	if err := decoder.Decode(&i.icon); err != nil {
		return err
	}
	if err := decoder.Decode(&i.chanceToBreakOnThrow); err != nil {
		return err
	}
	if err := decoder.Decode(&i.currentAttackModeIndex); err != nil {
		return err
	}
	if err := decoder.Decode(&i.setFlagOnPickup); err != nil {
		return err
	}
	if err := decoder.Decode(&i.size); err != nil {
		return err
	}
	if err := decoder.Decode(&i.weight); err != nil {
		return err
	}
	if err := decoder.Decode(&i.cost); err != nil {
		return err
	}

	if err := decoder.Decode(&i.alive); err != nil {
		return err
	}

	return nil
}

func (g *GameState) NewItemFromString(itemName string) *Item {
	charges := 1
	setCharges := false
	if fxtools.LooksLikeAFunction(itemName) {
		var item *Item
		name, args := fxtools.GetNameAndArgs(itemName)
		switch name {
		case "key":
			item = NewKey(args.Get(0), args.Get(1), g.iconForItem(foundation.ItemCategoryKeys))
		case "note":
			item = NewNoteFromFile(args.Get(0), args.Get(1), g.iconForItem(foundation.ItemCategoryReadables))
		}
		return item
	} else if strings.Contains(itemName, "|") {
		parts := strings.Split(itemName, "|")
		itemName = strings.TrimSpace(parts[0])
		charges, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		setCharges = true
	}

	newItem := g.newItemFromName(itemName)
	if setCharges {
		newItem.SetCharges(charges)
	}

	return newItem
}

func (g *GameState) newItemFromName(itemName string) *Item {
	if itemName == "gold" {
		return g.NewGold(1)
	}

	itemDef := g.getItemTemplateByName(itemName)

	if len(itemDef) == 0 {
		panic(fmt.Sprintf("Item not found: %s", itemName))
	}

	newItem := NewItemFromRecord(itemDef, g.iconForItem)

	if newItem == nil {
		panic(fmt.Sprintf("Item not found: %s", itemName))
	}
	return newItem
}
func NewNoteFromFile(fileName, description string, icon textiles.TextIcon) *Item {
	return &Item{
		description:  description,
		internalName: fileName,
		category:     foundation.ItemCategoryReadables,
		textFile:     fileName,
		icon:         icon,
	}
}
func NewKey(keyID, description string, icon textiles.TextIcon) *Item {
	return &Item{
		description:  description,
		internalName: keyID,
		lockFlag:     keyID,
		category:     foundation.ItemCategoryKeys,
		charges:      -1,
		icon:         icon,
	}
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
	return fmt.Sprintf("%s", i.description)
}

func (i *Item) LongNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())
	if i.IsWeapon() {
		weapon := i.weapon
		attackMode := weapon.GetAttackMode(i.currentAttackModeIndex)
		targetMode := attackMode.String()
		timeNeeded := attackMode.TUCost
		bullets := fmt.Sprintf("%d/%d", weapon.GetLoadedBullets(), weapon.GetMagazineSize())
		line = cview.Escape(fmt.Sprintf("%s (%s: %d TU / %s Dmg.) - %s", i.Name(), targetMode, timeNeeded, i.GetWeaponDamage().ShortString(), bullets))
	}
	if i.IsArmor() {
		line = cview.Escape(fmt.Sprintf("%s [%s]", i.Name(), i.GetArmorProtectionValueAsString()))
	}
	return colorCode + line + "[-]"
}

func (i *Item) InventoryNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())

	if i.IsWeapon() {
		line = cview.Escape(fmt.Sprintf("%s (%s Dmg.)", i.Name(), i.GetWeaponDamage().ShortString()))
	}
	if i.IsArmor() {
		line = cview.Escape(fmt.Sprintf("%s [%s]", i.Name(), i.GetArmorProtectionValueAsString()))
	}
	if i.IsAmmo() {
		line = cview.Escape(fmt.Sprintf("%s (x%d)", i.Name(), i.GetCharges()))
	}

	if i.statBonus != 0 {
		line = fmt.Sprintf("%s [%+d %s]", line, i.statBonus, i.stat.ToShortString())
	}

	if i.skillBonus != 0 {
		line = fmt.Sprintf("%s [%+d %s]", line, i.skillBonus, i.skill.ToShortString())
	}

	lineWithColor := colorCode + line + "[-]"

	if i.IsWeapon() || i.IsArmor() {
		qIcon := getQualityIcon(i.qualityInPercent)
		lineWithColor = fmt.Sprintf("%s %s", qIcon, lineWithColor)
	}

	return lineWithColor
}

func getQualityIcon(quality special.Percentage) string {
	colorCode := "[green]"
	// Lower one eighth block
	char := ""
	if quality < 13 {
		colorCode = "[red]"
		char = "▁"
	} else if quality < 25 {
		colorCode = "[red]"
		char = "▂"
	} else if quality < 38 {
		colorCode = "[red]"
		char = "▃"
	} else if quality < 50 {
		colorCode = "[yellow]"
		char = "▄"
	} else if quality < 63 {
		colorCode = "[yellow]"
		char = "▅"
	} else if quality < 75 {
		colorCode = "[yellow]"
		char = "▆"
	} else if quality < 88 {
		char = "▇"
	} else {
		char = "█"
	}

	return fmt.Sprintf("%s%s[-]", colorCode, char)
}

func (i *Item) SetPosition(pos geometry.Point) {
	i.position = pos
}

func (i *Item) Position() geometry.Point {
	if i.posHandler != nil {
		return i.posHandler()
	}
	return i.position
}

func (i *Item) Name() string {
	name := i.description
	if i.IsGold() {
		name = fmt.Sprintf("$%d", i.charges)
	}

	return name
}

func (i *Item) IsThrowable() bool {
	return true
}

func (i *Item) IsUsableOrZappable() bool {
	return i.useEffectName != "" || i.zapEffectName != ""
}

func (i *Item) IsReadable() bool {
	isRealText := i.IsBook()
	isSkillBook := i.IsSkillBook()
	return isRealText || isSkillBook
}

func (i *Item) IsBook() bool {
	return i.textFile != "" || i.text != ""
}

func (i *Item) IsSkillBook() bool {
	return i.skillBonus != 0 && i.category == foundation.ItemCategoryReadables
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
	if i.description != other.description || i.category != other.category {
		return false
	}

	if i.internalName != other.internalName {
		return false
	}
	if (i.IsWeapon() && !i.IsMissile()) || i.IsArmor() || (other.IsWeapon() && !other.IsMissile()) || other.IsArmor() {
		return false
	}

	if i.useEffectName != other.useEffectName || i.zapEffectName != other.zapEffectName {
		return false
	}

	if i.category == foundation.ItemCategoryGold && other.category == foundation.ItemCategoryGold {
		return true
	}

	if i.IsAmmo() && other.IsAmmo() && i.GetAmmo().Equals(other.GetAmmo()) {
		return true
	}

	if i.charges != other.charges {
		return false
	}

	return true
}

func (i *Item) SlotName() foundation.EquipSlot {
	switch {
	case i.IsArmor():
		return foundation.SlotNameArmorTorso
	case i.IsWeapon():
		return foundation.SlotNameMainHand
	}
	return foundation.SlotNameNotEquippable
}

func (i *Item) IsEquippable() bool {
	return i.IsWeapon() || i.IsArmor()
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

func (i *Item) GetCategory() foundation.ItemCategory {
	return i.category
}

func (i *Item) GetWeapon() *WeaponInfo {
	return i.weapon
}

func (i *Item) GetArmor() *ArmorInfo {
	return i.armor
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

func (i *Item) GetInternalName() string {
	return i.internalName
}

func (i *Item) GetStatBonus(stat special.Stat) int {

	if i.stat == stat {
		return i.statBonus
	}
	return 0
}
func (i *Item) GetSkillBonus(skill special.Skill) int {
	if i.skill == skill {
		return i.skillBonus
	}
	return 0
}
func (i *Item) GetEquipFlag() special.ActorFlag {
	return i.equipFlag
}

func (i *Item) IsMissile() bool {
	return i.IsWeapon() && i.GetWeapon().GetWeaponType().IsMissile()
}

func (i *Item) GetThrowDamage() fxtools.Interval {
	return i.thrownDamage
}

func (i *Item) ConsumeCharge() {
	i.charges--
}

func (i *Item) SetCharges(amount int) {
	i.charges = amount
}

func (i *Item) AfterEquippedTurn() {

}

func (i *Item) IsAmmo() bool {
	return i.ammo != nil
}

func (i *Item) IsAmmoOfCaliber(ammo int) bool {
	return i.IsAmmo() && i.GetAmmo().CaliberIndex == ammo
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

func (i *Item) IsLockpick() bool {
	return i.category == foundation.ItemCategoryLockpicks
}

func (i *Item) IsKey() bool {
	return i.category == foundation.ItemCategoryKeys && i.lockFlag != ""
}

func (i *Item) GetLockFlag() string {
	return i.lockFlag
}

func (i *Item) HasTag(tag foundation.ItemTags) bool {
	return i.tags.Contains(tag)
}

func (i *Item) GetTextFile() string {
	return i.textFile
}

func (i *Item) GetIcon() textiles.TextIcon {
	return i.icon
}

func (i *Item) IsBreakingNow() bool {
	return rand.Intn(100) < i.chanceToBreakOnThrow
}

func (i *Item) GetCurrentAttackMode() AttackMode {
	return i.weapon.GetAttackMode(i.currentAttackModeIndex)
}

func (i *Item) CycleTargetMode() {
	if !i.IsWeapon() {
		return
	}
	i.currentAttackModeIndex++
	if i.currentAttackModeIndex >= len(i.weapon.attackModes) {
		i.currentAttackModeIndex = 0
	}
}

func (i *Item) PickupFlag() string {
	return i.setFlagOnPickup
}

func (i *Item) DropFlag() string {
	return i.setFlagOnDrop
}

func (i *Item) GetText() string {
	return i.text
}

func (i *Item) IsLightSource() bool {
	return i.HasTag(foundation.TagLightSource)
}

func (i *Item) GetCarryWeight() int {
	return i.weight
}

func (i *Item) NeedsRepair() bool {
	if !i.IsWeapon() && !i.IsArmor() {
		return false
	}
	return i.qualityInPercent < 100
}

func (i *Item) CanBeRepairedWith(other *Item) bool {
	if !other.NeedsRepair() || i == other {
		return false
	}
	return i.category == other.category && i.internalName == other.internalName
}

func (i *Item) IsLoadedWeapon() bool {
	return i.IsWeapon() && i.GetWeapon().IsLoaded()
}

func (i *Item) GetArmorProtection(damageType special.DamageType) Protection {
	return i.armor.getRawProtection(damageType).Scaled(i.qualityInPercent.AsFloat())
}

func (i *Item) GetArmorProtectionValueAsString() string {
	physical := i.GetArmorProtection(special.DamageTypeNormal)
	energy := i.GetArmorProtection(special.DamageTypeLaser)
	return fmt.Sprintf("%s %s", physical.String(), energy.String())

}

func (i *Item) GetWeaponDamage() fxtools.Interval {
	return i.weapon.getRawDamage().Scaled(i.qualityInPercent.AsFloat())
}

func (i *Item) IsStackingWithCharges() bool {
	return i.IsAmmo() || i.IsGold()
}

func (i *Item) IsWatch() bool {
	return i.useEffectName == "show_time"
}

func (i *Item) SetPositionHandler(handler func() geometry.Point) {
	i.posHandler = handler
}

func (i *Item) SetAlive(value bool) {
	i.alive = value
}

func (i *Item) GetEffectParameters() Params {
	parameters := i.effectParameters
	if !parameters.HasDamage() && i.IsWeapon() {
		damageInterval := i.GetWeaponDamage()
		parameters["damage_interval"] = damageInterval
		parameters["damage"] = damageInterval.Roll()
	}
	return parameters
}
