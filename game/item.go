package game

import (
	"contractor/d100"
	"contractor/foundation"
	"contractor/gridmap"
	"contractor/util"
	"encoding/gob"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"strings"
)

func init() {
	// This is a hack to make sure that the foundation package is imported
	// so that the gob.Register function is called
	gob.Register(&GenericItem{})
	gob.Register(&Weapon{})
	gob.Register(&Ammo{})
	gob.Register(&Armor{})
}

type GenericItem struct {
	UID gridmap.ItemID

	// Configuration
	DisplayName     string
	InternalName    string
	UseEffectName   string
	ZapEffectName   string
	StatChanges     StatChange
	EquipFlag       foundation.ActorFlag
	ThrownDamage    fxtools.Interval
	TextFile        string
	LongDescription string

	TextValue string
	TextVar   string

	Weight int
	Cost   int

	LockFlag             string
	ChanceToBreakOnThrow int
	SetFlagOnPickup      string
	SetFlagOnDrop        string
	EffectParameters     foundation.Params

	// State
	RawPosition      geometry.Point
	Category         foundation.ItemCategory
	QualityInPercent d100.Percentage
	StackSize        int
	Charges          int

	Tags foundation.ItemTags

	Icon textiles.TextIcon

	posHandler func() geometry.Point

	Alive bool

	InvIndex int

	Hidden bool
}

func (i *GenericItem) SetInventoryIndex(index int) {
	i.InvIndex = index
}
func (i *GenericItem) AddStacks(item foundation.Item) {
	i.StackSize += item.GetStackSize()
}
func (i *GenericItem) ID() gridmap.ItemID {
	return i.UID
}
func (i *GenericItem) SetStackSize(count int) {
	i.StackSize = count
}
func (i *GenericItem) IsStackable() bool {
	return i.IsGold() || i.IsLockpick() || i.IsFood() || i.IsConsumable() || i.IsReadable()
}

func (i *GenericItem) GetPrice() int {
	return i.Cost
}
func (i *GenericItem) IsHidden() bool {
	return i.Hidden
}
func (i *GenericItem) SetHidden(hidden bool) {
	i.Hidden = hidden
}
func (i *GenericItem) IsRepairable() bool {
	return false
}
func (i *GenericItem) GetQuality() d100.Percentage {
	return i.QualityInPercent
}
func (i *GenericItem) String() string {
	return fmt.Sprintf("Item: %s(%d)", i.InternalName, i.Charges)
}

func (i *GenericItem) ShouldActivate(tickCount int) bool {
	return i.Charges == tickCount
}

func (i *GenericItem) IsTimerTicking(tickCount int) bool {
	return tickCount <= i.Charges && i.Alive
}

func (i *GenericItem) IsMultipleStacks() bool {
	return i.StackSize > 1
}

func (i *GenericItem) GetStackSize() int {
	if i.IsMultipleStacks() {
		return i.StackSize
	}
	return 1
}

func (g *GameState) NewItemFromString(itemName string) foundation.Item {
	if fxtools.LooksLikeAFunction(itemName) {
		name, args := fxtools.GetNameAndArgs(itemName)
		switch name {
		case "key":
			return NewKey(args.Get(0), args.Get(1), g.iconForItem(foundation.ItemCategoryKeys))
		case "note":
			return NewNoteFromFile(args.Get(0), args.Get(1), g.iconForItem(foundation.ItemCategoryReadables))
		default: // parametric item name(stackSize, quality)
			newItem := g.newItemFromName(name)
			valOne := args.GetInt(0)
			if newItem.IsStackable() {
				newItem.SetStackSize(valOne)
				if len(args) > 1 {
					quality := args.GetInt(1)
					newItem.SetQuality(d100.Percentage(quality))
				}
			} else if newItem.IsRepairable() {
				newItem.SetQuality(d100.Percentage(valOne))
			}

			return newItem
		}
	}

	// default item creation from template without parameters
	newItem := g.newItemFromName(itemName)
	if newItem.IsRepairable() && newItem.GetQuality() == -1 {
		newItem.SetQuality(d100.Percentage(rand.Intn(90) + 10))
	}
	return newItem
}

func (g *GameState) newItemFromName(itemName string) foundation.Item {
	if itemName == "gold" {
		return g.NewGold(1)
	}

	if itemName == "bowel_disruptor" {
		return g.NewBowelDisruptor()
	}

	itemDef := g.getItemTemplateByName(itemName)

	if len(itemDef) == 0 {
		panic(fmt.Sprintf("Item not found: %s", itemName))
	}

	newItem := NewItemFromRecord(itemDef, g.NewItemFromString, g.iconForItem)

	if newItem == nil {
		panic(fmt.Sprintf("Item not found: %s", itemName))
	}
	return newItem
}
func NewNoteFromFile(fileName, description string, icon textiles.TextIcon) *GenericItem {
	return &GenericItem{
		UID:          gridmap.NextItemID(),
		DisplayName:  description,
		InternalName: fileName,
		Category:     foundation.ItemCategoryReadables,
		TextFile:     fileName,
		Icon:         icon,
	}
}
func NewKey(keyID, description string, icon textiles.TextIcon) *GenericItem {
	return &GenericItem{
		UID:          gridmap.NextItemID(),
		DisplayName:  description,
		InternalName: keyID,
		LockFlag:     keyID,
		Category:     foundation.ItemCategoryKeys,
		Charges:      -1,
		Icon:         icon,
	}
}

func (i *GenericItem) InventoryNameWithColorsAndShortcut(lineColorCode string) string {
	return fmt.Sprintf("%c - %s", i.Shortcut(), i.InventoryNameWithColors(lineColorCode))
}

func (i *GenericItem) Shortcut() rune {
	return foundation.ShortCutFromIndex(i.InvIndex)
}

func (i *GenericItem) DisplayLength() int {
	return cview.TaggedStringWidth(i.InventoryNameWithColorsAndShortcut("[red]"))
}

func (i *GenericItem) FullDescription(colorCode string) string {
	rows := i.fullDescriptionRows()
	lines := fxtools.TableLayout(rows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	lines = append([]string{i.InventoryNameWithColors(colorCode), i.Category.String()}, lines...)

	lines = i.appendText(lines)

	return strings.Join(lines, "\n")
}

func (i *GenericItem) fullDescriptionRows() []fxtools.TableRow {
	var rows []fxtools.TableRow
	statPairs := i.getStatPairsAsRows()
	rows = append(rows, statPairs...)

	if i.IsConsumable() {
		rows = append(rows, fxtools.NewTableRow("Duration", fmt.Sprintf("%d turns", i.Charges)))
	}
	return rows
}

func (i *GenericItem) LongNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())
	statPairs := i.getStatPairsAsStrings()

	if len(statPairs) > 0 {
		line = cview.Escape(fmt.Sprintf("%s [%s]", line, strings.Join(statPairs, "|")))
	}

	return colorCode + line + "[-]"
}
func (i *GenericItem) ShortNameWithColors(colorCode string) string {
	return colorCode + cview.Escape(i.Name()) + "[-]"
}
func (i *GenericItem) InventoryNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())

	if i.GetStackSize() > 1 && !i.IsGold() {
		line = fmt.Sprintf("%s (x%d)", line, i.GetStackSize())
	}

	lineWithColor := colorCode + line + "[-]"

	return lineWithColor
}

func (i *GenericItem) getStatPairsAsStrings() []string {
	var statPairs []string
	if len(i.StatChanges.StatChanges) > 0 {
		for stat := d100.Stat(0); stat < d100.StatCount; stat++ {
			if chg, hasChg := i.StatChanges.StatChanges[stat]; hasChg {
				var statName string
				statName = stat.ToShortString()
				statPairs = append(statPairs, fmt.Sprintf("%+d %s", chg, statName))
			}
		}
	}

	if len(i.StatChanges.SkillChanges) > 0 {
		for skill := d100.Skill(0); skill < d100.Skill(d100.SkillCount()); skill++ {
			if chg, hasChg := i.StatChanges.SkillChanges[skill]; hasChg {
				var skillName string
				skillName = skill.ToShortString()

				statPairs = append(statPairs, fmt.Sprintf("%+d %s", chg, skillName))
			}
		}
	}

	if len(i.StatChanges.DerivedStatChanges) > 0 {
		for stat := d100.DerivedStat(0); stat < d100.DerivedStatCount; stat++ {
			if chg, hasChg := i.StatChanges.DerivedStatChanges[stat]; hasChg {
				var statName string
				statName = stat.ToShortString()
				statPairs = append(statPairs, fmt.Sprintf("%+d %s", chg, statName))
			}
		}
	}
	return statPairs
}
func (i *GenericItem) getStatPairsAsRows() []fxtools.TableRow {
	var statPairs []fxtools.TableRow
	if len(i.StatChanges.StatChanges) > 0 {
		for stat := d100.Stat(0); stat < d100.StatCount; stat++ {
			if chg, hasChg := i.StatChanges.StatChanges[stat]; hasChg {
				var statName string
				statName = stat.String()
				statPairs = append(statPairs, fxtools.NewTableRow(statName, fmt.Sprintf("%+d", chg)))
			}
		}
	}

	if len(i.StatChanges.SkillChanges) > 0 {
		for skill := d100.Skill(0); skill < d100.Skill(d100.SkillCount()); skill++ {
			if chg, hasChg := i.StatChanges.SkillChanges[skill]; hasChg {
				var skillName string
				skillName = skill.String()
				statPairs = append(statPairs, fxtools.NewTableRow(skillName, fmt.Sprintf("%+d", chg)))
			}
		}
	}

	if len(i.StatChanges.DerivedStatChanges) > 0 {
		for stat := d100.DerivedStat(0); stat < d100.DerivedStatCount; stat++ {
			if chg, hasChg := i.StatChanges.DerivedStatChanges[stat]; hasChg {
				var statName string
				statName = stat.String()
				statPairs = append(statPairs, fxtools.NewTableRow(statName, fmt.Sprintf("%+d", chg)))
			}
		}
	}
	return statPairs
}
func getQualityIcon(quality d100.Percentage) string {
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
	i.RawPosition = pos
}

func (i *GenericItem) Position() geometry.Point {
	if i.posHandler != nil {
		return i.posHandler()
	}
	return i.RawPosition
}

func (i *GenericItem) Name() string {
	name := i.DisplayName
	if i.IsGold() {
		name = fmt.Sprintf("%d sat", i.StackSize)
	}

	return name
}

func (i *GenericItem) IsThrowable() bool {
	return true
}

func (i *GenericItem) IsUsableOrZappable() bool {
	return i.UseEffectName != "" || i.ZapEffectName != ""
}

func (i *GenericItem) IsReadable() bool {
	isRealText := i.IsBook()
	isSkillBook := i.IsSkillBook()
	return isRealText || isSkillBook
}

func (i *GenericItem) TextVariables(scriptFuncs map[string]govaluate.ExpressionFunction) map[string]string {
	//TextValue && TextVar
	if i.TextValue == "" || i.TextVar == "" {
		return make(map[string]string)
	}
	valueExpr, _ := govaluate.NewEvaluableExpressionWithFunctions(i.TextValue, scriptFuncs)
	value, _ := valueExpr.Evaluate(nil)
	return map[string]string{
		i.TextVar: value.(string),
	}
}

func (i *GenericItem) IsBook() bool {
	return i.Category == foundation.ItemCategoryReadables && (i.TextFile != "" || i.LongDescription != "")
}

func (i *GenericItem) IsSkillBook() bool {
	return len(i.StatChanges.SkillChanges) == 1 &&
		len(i.StatChanges.StatChanges) == 0 &&
		len(i.StatChanges.DerivedStatChanges) == 0 &&
		i.Category == foundation.ItemCategoryReadables
}

func (i *GenericItem) IsUsable() bool {
	return i.UseEffectName != ""
}

func (i *GenericItem) UseEffect() string {
	return i.UseEffectName
}

func (i *GenericItem) ZapEffect() string {
	return i.ZapEffectName
}

func (i *GenericItem) IsZappable() bool {
	return i.ZapEffectName != ""
}

func (i *GenericItem) Color() color.RGBA {
	return color.RGBA{255, 255, 255, 255}
}

func (i *GenericItem) CanStackWith(other foundation.Item) bool {
	if i.Category != other.GetCategory() {
		return false
	}

	if i.InternalName != other.GetInternalName() {
		return false
	}
	if (i.IsWeapon() && !i.IsMissile()) || i.IsArmor() || (other.IsWeapon() && !other.IsMissile()) || other.IsArmor() {
		return false
	}

	if i.UseEffectName != other.UseEffect() || i.ZapEffectName != other.ZapEffect() {
		return false
	}

	if i.Category == foundation.ItemCategoryGold && other.GetCategory() == foundation.ItemCategoryGold {
		return true
	}

	if i.Charges != other.GetCharges() {
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

func (i *GenericItem) GetCategory() foundation.ItemCategory {
	return i.Category
}

func (i *GenericItem) IsGold() bool {
	return i.Category == foundation.ItemCategoryGold
}

func (i *GenericItem) GetCharges() int {
	return i.Charges
}

func (i *GenericItem) IsFood() bool {
	return i.Category == foundation.ItemCategoryFood
}

func (i *GenericItem) IsHeadGear() bool {
	return i.Category == foundation.ItemCategoryHeadgear
}

func (i *GenericItem) IsConsumable() bool {
	return i.Category == foundation.ItemCategoryFood || i.Category == foundation.ItemCategoryConsumables
}
func (i *GenericItem) IsDrug() bool {
	return i.Charges > 0 && i.Category == foundation.ItemCategoryConsumables && (len(i.StatChanges.StatChanges) > 0 || len(i.StatChanges.SkillChanges) > 0 || len(i.StatChanges.DerivedStatChanges) > 0)
}

func (i *GenericItem) GetInternalName() string {
	return i.InternalName
}

func (i *GenericItem) GetEquipFlag() foundation.ActorFlag {
	return i.EquipFlag
}

func (i *GenericItem) GetThrowDamage() fxtools.Interval {
	return i.ThrownDamage
}

func (i *GenericItem) ConsumeCharge() {
	i.Charges--
}

func (i *GenericItem) SetCharges(amount int) {
	i.Charges = amount
}

func (i *GenericItem) AfterEquippedTurn() {

}

func (i *GenericItem) RemoveStacks(spent int) {
	i.StackSize -= spent
	if i.StackSize < 0 {
		i.StackSize = 0
	}
}

func (i *GenericItem) Split(bullets int) foundation.Item {
	if bullets >= i.StackSize {
		return i
	}
	clone := *i
	clone.StackSize = bullets
	clone.UID = gridmap.NextItemID()
	i.StackSize -= bullets
	return &clone
}

func (i *GenericItem) IsMissile() bool {
	return false
}

func (i *GenericItem) IsAmmo() bool {
	return false
}

func (i *GenericItem) IsLockpick() bool {
	return i.Category == foundation.ItemCategoryLockpicks
}

func (i *GenericItem) IsKey() bool {
	return i.Category == foundation.ItemCategoryKeys && i.LockFlag != ""
}

func (i *GenericItem) GetLockFlag() string {
	return i.LockFlag
}

func (i *GenericItem) HasTag(tag foundation.ItemTags) bool {
	return i.Tags.Contains(tag)
}

func (i *GenericItem) GetTextFile() string {
	return i.TextFile
}

func (i *GenericItem) GetIcon() textiles.TextIcon {
	return i.Icon
}

func (i *GenericItem) IsBreakingNow() bool {
	return rand.Intn(100) < i.ChanceToBreakOnThrow
}

func (i *GenericItem) GetPickupFlag() string {
	return i.SetFlagOnPickup
}

func (i *GenericItem) GetDropFlag() string {
	return i.SetFlagOnDrop
}

func (i *GenericItem) GetText() string {
	return i.LongDescription
}

func (i *GenericItem) IsLightSource() bool {
	return i.HasTag(foundation.TagLightSource)
}

func (i *GenericItem) GetCarryWeight() int {
	return i.Weight
}

func (i *GenericItem) NeedsRepair() bool {
	if !i.IsWeapon() && !i.IsArmor() {
		return false
	}
	return i.QualityInPercent < 100
}

func (i *GenericItem) CanBeRepairedWith(other foundation.Repairable) bool {
	if !other.NeedsRepair() || i == other {
		return false
	}
	return i.Category == other.GetCategory() && i.InternalName == other.GetInternalName()
}

func (i *GenericItem) IsWatch() bool {
	return i.UseEffectName == "show_time"
}

func (i *GenericItem) SetPositionHandler(handler func() geometry.Point) {
	i.posHandler = handler
}

func (i *GenericItem) SetAlive(value bool) {
	i.Alive = value
}

func (i *GenericItem) GetEffectParameters() foundation.Params {
	if i.EffectParameters == nil {
		return make(foundation.Params)
	}
	return i.EffectParameters
}

func (i *GenericItem) SetQuality(quality d100.Percentage) {
	i.QualityInPercent = d100.Percentage(quality)
}

func (i *GenericItem) Degrade(degrade float64) {
	i.QualityInPercent -= d100.Percentage(degrade)
}

func (i *GenericItem) GetSkillMod(skill d100.Skill) (int, bool) {
	mod, hasMod := i.StatChanges.SkillChanges[skill]
	return mod, hasMod
}

func (i *GenericItem) GetStatMod(stat d100.Stat) (int, bool) {
	mod, hasMod := i.StatChanges.StatChanges[stat]
	return mod, hasMod
}

func (i *GenericItem) GetDerivedStatMod(stat d100.DerivedStat) (int, bool) {
	mod, hasMod := i.StatChanges.DerivedStatChanges[stat]
	return mod, hasMod
}

func (i *GenericItem) GetSkillBookValues() (d100.Skill, int) {
	for skill, value := range i.StatChanges.SkillChanges {
		return skill, value
	}
	return d100.Skill(-1), 0
}

func (i *GenericItem) appendText(lines []string) []string {
	width := max(longestLine(lines), 26)
	if i.LongDescription != "" {
		lines = append(lines, "")
		lines = append(lines, util.WrapString(i.LongDescription, uint(width)))
	}
	return lines
}

func longestLine(lines []string) int {
	longest := 0
	for _, line := range lines {
		if cview.TaggedStringWidth(line) > longest {
			longest = cview.TaggedStringWidth(line)
		}
	}
	return longest
}
