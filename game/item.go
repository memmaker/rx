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
	"strings"
)

type GenericItem struct {
	description  string
	internalName string
	position     geometry.Point
	category     foundation.ItemCategory

	qualityInPercent special.Percentage

	useEffectName string
	zapEffectName string

	stackSize int

	charges int

	statChanges StatChange

	equipFlag    foundation.ActorFlag
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
	effectParameters       foundation.Params
	invIndex               int
}

func (i *GenericItem) SetInventoryIndex(index int) {
	i.invIndex = index
}
func (i *GenericItem) AddStacks(item foundation.Item) {
	i.stackSize += item.StackSize()
}

func (i *GenericItem) IsStackable() bool {
	return false
}
func (i *GenericItem) IsRepairable() bool {
	return false
}
func (i *GenericItem) Quality() special.Percentage {
	return i.qualityInPercent
}

func (i *GenericItem) String() string {
	return fmt.Sprintf("Item: %s(%d)", i.internalName, i.charges)
}

func (i *GenericItem) ShouldActivate(tickCount int) bool {
	return i.charges == tickCount
}

func (i *GenericItem) IsAlive(tickCount int) bool {
	return tickCount <= i.charges && i.alive
}

func (i *GenericItem) IsMultipleStacks() bool {
	return i.stackSize > 1
}

func (i *GenericItem) StackSize() int {
	if i.IsMultipleStacks() {
		return i.stackSize
	}
	return 1
}

// GobEncode encodes the Item struct into a byte slice.
func (i *GenericItem) GobEncode() ([]byte, error) {
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

	if err := encoder.Encode(i.useEffectName); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.zapEffectName); err != nil {
		return nil, err
	}
	if err := encoder.Encode(i.charges); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.statChanges); err != nil {
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
func (i *GenericItem) GobDecode(data []byte) error {
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

	if err := decoder.Decode(&i.useEffectName); err != nil {
		return err
	}
	if err := decoder.Decode(&i.zapEffectName); err != nil {
		return err
	}
	if err := decoder.Decode(&i.charges); err != nil {
		return err
	}

	if err := decoder.Decode(&i.statChanges); err != nil {
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

func (g *GameState) NewItemFromString(itemName string) foundation.Item {
	if fxtools.LooksLikeAFunction(itemName) {
		name, args := fxtools.GetNameAndArgs(itemName)
		switch name {
		case "key":
			return NewKey(args.Get(0), args.Get(1), g.iconForItem(foundation.ItemCategoryKeys))
		case "note":
			return NewNoteFromFile(args.Get(0), args.Get(1), g.iconForItem(foundation.ItemCategoryReadables))
		default: // parametric item name(charges, quality)
			newItem := g.newItemFromName(name)
			count := args.GetInt(0)
			newItem.SetCharges(count)
			if len(args) > 1 {
				quality := args.GetInt(1)
				newItem.SetQuality(special.Percentage(quality))
			}
			return newItem
		}
	}

	// default item creation from template without parameters
	newItem := g.newItemFromName(itemName)
	if newItem.IsRepairable() && newItem.Quality() == -1 {
		newItem.SetQuality(special.Percentage(rand.Intn(90) + 10))
	}
	return newItem
}

func (g *GameState) newItemFromName(itemName string) foundation.Item {
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
func NewNoteFromFile(fileName, description string, icon textiles.TextIcon) *GenericItem {
	return &GenericItem{
		description:  description,
		internalName: fileName,
		category:     foundation.ItemCategoryReadables,
		textFile:     fileName,
		icon:         icon,
	}
}
func NewKey(keyID, description string, icon textiles.TextIcon) *GenericItem {
	return &GenericItem{
		description:  description,
		internalName: keyID,
		lockFlag:     keyID,
		category:     foundation.ItemCategoryKeys,
		charges:      -1,
		icon:         icon,
	}
}

func (i *GenericItem) InventoryNameWithColorsAndShortcut(lineColorCode string) string {
	return fmt.Sprintf("%c - %s", i.Shortcut(), i.InventoryNameWithColors(lineColorCode))
}

func (i *GenericItem) Shortcut() rune {
	return foundation.ShortCutFromIndex(i.invIndex)
}

func (i *GenericItem) DisplayLength() int {
	return cview.TaggedStringWidth(i.InventoryNameWithColorsAndShortcut(""))
}

func (i *GenericItem) Description() string {
	return i.description
}

func (i *GenericItem) LongNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())
	return colorCode + line + "[-]"
}

func (i *GenericItem) InventoryNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())

	statPairs := i.getStatPairsAsStrings()

	if len(statPairs) > 0 {
		line = fmt.Sprintf("%s [%s]", line, strings.Join(statPairs, "|"))
	}

	lineWithColor := colorCode + line + "[-]"

	return lineWithColor
}

func (i *GenericItem) getStatPairsAsStrings() []string {
	var statPairs []string
	if len(i.statChanges.StatChanges) > 0 {
		for stat := special.Stat(0); stat < special.StatCount; stat++ {
			if chg, hasChg := i.statChanges.StatChanges[stat]; hasChg {
				statPairs = append(statPairs, fmt.Sprintf("%+d %s", chg, stat.ToShortString()))
			}
		}
	}

	if len(i.statChanges.SkillChanges) > 0 {
		for skill := special.Skill(0); skill < special.SkillCount; skill++ {
			if chg, hasChg := i.statChanges.SkillChanges[skill]; hasChg {
				statPairs = append(statPairs, fmt.Sprintf("%+d %s", chg, skill.ToShortString()))
			}
		}
	}

	if len(i.statChanges.DerivedStatChanges) > 0 {
		for stat := special.DerivedStat(0); stat < special.DerivedStatCount; stat++ {
			if chg, hasChg := i.statChanges.DerivedStatChanges[stat]; hasChg {
				statPairs = append(statPairs, fmt.Sprintf("%+d %s", chg, stat.ToShortString()))
			}
		}
	}
	return statPairs
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

func (i *GenericItem) SetPosition(pos geometry.Point) {
	i.position = pos
}

func (i *GenericItem) Position() geometry.Point {
	if i.posHandler != nil {
		return i.posHandler()
	}
	return i.position
}

func (i *GenericItem) Name() string {
	name := i.description
	if i.IsGold() {
		name = fmt.Sprintf("$%d", i.charges)
	}

	return name
}

func (i *GenericItem) IsThrowable() bool {
	return true
}

func (i *GenericItem) IsUsableOrZappable() bool {
	return i.useEffectName != "" || i.zapEffectName != ""
}

func (i *GenericItem) IsReadable() bool {
	isRealText := i.IsBook()
	isSkillBook := i.IsSkillBook()
	return isRealText || isSkillBook
}

func (i *GenericItem) IsBook() bool {
	return i.textFile != "" || i.text != ""
}

func (i *GenericItem) IsSkillBook() bool {
	return len(i.statChanges.SkillChanges) == 1 &&
		len(i.statChanges.StatChanges) == 0 &&
		len(i.statChanges.DerivedStatChanges) == 0 &&
		i.category == foundation.ItemCategoryReadables
}

func (i *GenericItem) IsUsable() bool {
	return i.useEffectName != ""
}

func (i *GenericItem) UseEffect() string {
	return i.useEffectName
}

func (i *GenericItem) ZapEffect() string {
	return i.zapEffectName
}

func (i *GenericItem) IsZappable() bool {
	return i.zapEffectName != ""
}

func (i *GenericItem) Color() color.RGBA {
	return color.RGBA{255, 255, 255, 255}
}

func (i *GenericItem) CanStackWith(other foundation.Item) bool {
	if i.description != other.Description() || i.category != other.Category() {
		return false
	}

	if i.internalName != other.InternalName() {
		return false
	}
	if (i.IsWeapon() && !i.IsMissile()) || i.IsArmor() || (other.IsWeapon() && !other.IsMissile()) || other.IsArmor() {
		return false
	}

	if i.useEffectName != other.UseEffect() || i.zapEffectName != other.ZapEffect() {
		return false
	}

	if i.category == foundation.ItemCategoryGold && other.Category() == foundation.ItemCategoryGold {
		return true
	}

	if i.charges != other.Charges() {
		return false
	}

	return true
}

func (i *GenericItem) IsEquippable() bool {
	return false
}

func (i *GenericItem) IsMeleeWeapon() bool {
	return false
}

func (i *GenericItem) IsRangedWeapon() bool {
	return false
}

func (i *GenericItem) IsArmor() bool {
	return false
}

func (i *GenericItem) IsWeapon() bool {
	return false
}

func (i *GenericItem) Category() foundation.ItemCategory {
	return i.category
}

func (i *GenericItem) IsGold() bool {
	return i.category == foundation.ItemCategoryGold
}

func (i *GenericItem) Charges() int {
	return i.charges
}

func (i *GenericItem) IsFood() bool {
	return i.category == foundation.ItemCategoryFood
}

func (i *GenericItem) IsConsumable() bool {
	return i.category == foundation.ItemCategoryFood || i.category == foundation.ItemCategoryConsumables
}
func (i *GenericItem) IsDrug() bool {
	return i.charges > 0 && i.category == foundation.ItemCategoryConsumables && (len(i.statChanges.StatChanges) > 0 || len(i.statChanges.SkillChanges) > 0 || len(i.statChanges.DerivedStatChanges) > 0)
}

func (i *GenericItem) InternalName() string {
	return i.internalName
}

func (i *GenericItem) GetEquipFlag() foundation.ActorFlag {
	return i.equipFlag
}

func (i *GenericItem) GetThrowDamage() fxtools.Interval {
	return i.thrownDamage
}

func (i *GenericItem) ConsumeCharge() {
	i.charges--
}

func (i *GenericItem) SetCharges(amount int) {
	i.charges = amount
}

func (i *GenericItem) AfterEquippedTurn() {

}

func (i *GenericItem) RemoveCharges(spent int) {
	i.charges -= spent
	if i.charges < 0 {
		i.charges = 0
	}
}

func (i *GenericItem) Split(bullets int) foundation.Item {
	if bullets >= i.charges {
		return i
	}
	clone := *i
	clone.charges = bullets
	i.charges -= bullets
	return &clone
}

func (i *GenericItem) MergeCharges(ammo foundation.Item) {
	i.charges += ammo.Charges()
}

func (i *GenericItem) IsMissile() bool {
	return false
}

func (i *GenericItem) IsAmmo() bool {
	return false
}

func (i *GenericItem) IsLockpick() bool {
	return i.category == foundation.ItemCategoryLockpicks
}

func (i *GenericItem) IsKey() bool {
	return i.category == foundation.ItemCategoryKeys && i.lockFlag != ""
}

func (i *GenericItem) GetLockFlag() string {
	return i.lockFlag
}

func (i *GenericItem) HasTag(tag foundation.ItemTags) bool {
	return i.tags.Contains(tag)
}

func (i *GenericItem) GetTextFile() string {
	return i.textFile
}

func (i *GenericItem) GetIcon() textiles.TextIcon {
	return i.icon
}

func (i *GenericItem) IsBreakingNow() bool {
	return rand.Intn(100) < i.chanceToBreakOnThrow
}

func (i *GenericItem) PickupFlag() string {
	return i.setFlagOnPickup
}

func (i *GenericItem) DropFlag() string {
	return i.setFlagOnDrop
}

func (i *GenericItem) GetText() string {
	return i.text
}

func (i *GenericItem) IsLightSource() bool {
	return i.HasTag(foundation.TagLightSource)
}

func (i *GenericItem) GetCarryWeight() int {
	return i.weight
}

func (i *GenericItem) NeedsRepair() bool {
	if !i.IsWeapon() && !i.IsArmor() {
		return false
	}
	return i.qualityInPercent < 100
}

func (i *GenericItem) CanBeRepairedWith(other foundation.Repairable) bool {
	if !other.NeedsRepair() || i == other {
		return false
	}
	return i.category == other.Category() && i.internalName == other.InternalName()
}

func (i *GenericItem) IsWatch() bool {
	return i.useEffectName == "show_time"
}

func (i *GenericItem) SetPositionHandler(handler func() geometry.Point) {
	i.posHandler = handler
}

func (i *GenericItem) SetAlive(value bool) {
	i.alive = value
}

func (i *GenericItem) GetEffectParameters() foundation.Params {
	parameters := i.effectParameters
	return parameters
}

func (i *GenericItem) SetQuality(quality special.Percentage) {
	i.qualityInPercent = special.Percentage(quality)
}

func (i *GenericItem) GetDegradationFactorOfAttack() float64 {
	factor := 1.0
	return factor
}

func (i *GenericItem) Degrade(degrade float64) {
	i.qualityInPercent -= special.Percentage(degrade)
}

func (i *GenericItem) GetSkillMod(skill special.Skill) (int, bool) {
	mod, hasMod := i.statChanges.SkillChanges[skill]
	return mod, hasMod
}

func (i *GenericItem) GetStatMod(stat special.Stat) (int, bool) {
	mod, hasMod := i.statChanges.StatChanges[stat]
	return mod, hasMod
}

func (i *GenericItem) GetDerivedStatMod(stat special.DerivedStat) (int, bool) {
	mod, hasMod := i.statChanges.DerivedStatChanges[stat]
	return mod, hasMod
}

func (i *GenericItem) GetSkillBookValues() (special.Skill, int) {
	for skill, value := range i.statChanges.SkillChanges {
		return skill, value
	}
	return special.SkillCount, 0
}
