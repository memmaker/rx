package foundation

import (
    "contractor/d100"
    "contractor/gridmap"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"strings"
)

type Readable interface {
	IsSkillBook() bool
	GetSkillBookValues() (d100.Skill, int)
	GetTextFile() string
	GetText() string
	TextVariables(funcs map[string]govaluate.ExpressionFunction) map[string]string
}

type Usable interface {
	Name() string
	GetEffectParameters() Params
	UseEffect() string
}

type Zappable interface {
	Name() string
	ZapEffect() string
	GetEffectParameters() Params
}

type Repairable interface {
	Name() string
	CanBeRepairedWith(parts Repairable) bool
	GetQuality() d100.Percentage
	SetQuality(quality d100.Percentage)
	NeedsRepair() bool
	GetCategory() ItemCategory
	GetInternalName() string
}
type Equippable interface {
	Name() string
	IsArmor() bool
	IsWeapon() bool
	IsLightSource() bool
	IsRangedWeapon() bool
	IsMeleeWeapon() bool
	IsMissile() bool
	GetEquipFlag() ActorFlag
	GetCharges() int
	AfterEquippedTurn()
	GetInternalName() string
	IsHeadGear() bool
	ID() gridmap.ItemID
}

type Timable interface {
	Zappable
	Name() string
	HasTag(timed ItemTags) bool
	SetCharges(turns int)

	ShouldActivate(tickCount int) bool
	IsTimerTicking(tickCount int) bool
	String() string
	Position() geometry.Point
}
type Item interface {
	ID() gridmap.ItemID

	GetCategory() ItemCategory
	Name() string
	String() string
	InventoryNameWithColors(lineColorCode string) string
	InventoryNameWithColorsAndShortcut(invItemColorCode string) string
	LongNameWithColors(colorCode string) string
	FullDescription(colorCode string) string
	Shortcut() rune
	DisplayLength() int
	Position() geometry.Point
	SetPosition(position geometry.Point)
	GetIcon() textiles.TextIcon
	GetCarryWeight() int
	GetDerivedStatMod(stat d100.DerivedStat) (int, bool)
	ShouldActivate(tickCount int) bool
	IsTimerTicking(tickCount int) bool

	// Type Queries
	IsLightSource() bool
	IsRangedWeapon() bool
	IsMeleeWeapon() bool
	IsDrug() bool
	IsWeapon() bool
	IsArmor() bool
	IsAmmo() bool
	IsConsumable() bool
	IsLockpick() bool
	IsMissile() bool
	IsEquippable() bool
	IsUsableOrZappable() bool
	IsReadable() bool
	IsUsable() bool
	IsZappable() bool
	IsKey() bool
	IsWatch() bool
	IsFood() bool
	IsHeadGear() bool

	// Stacking
	IsMultipleStacks() bool
	GetStackSize() int
	Split(count int) Item
	CanStackWith(item Item) bool
	AddStacks(item Item)

	GetLockFlag() string
	GetThrowDamage() fxtools.Interval
	GetInternalName() string

	GetEffectParameters() Params
	ZapEffect() string
	UseEffect() string
	GetCharges() int
	SetCharges(count int)
	SetQuality(qualityInPercent d100.Percentage)

	GetQuality() d100.Percentage
	GetEquipFlag() ActorFlag
	NeedsRepair() bool
	Color() color.RGBA
	SetPositionHandler(pos func() geometry.Point)
	ConsumeCharge()
	HasTag(loot ItemTags) bool

	AfterEquippedTurn()
	GetDropFlag() string
	GetPickupFlag() string

	CanBeRepairedWith(parts Repairable) bool

	IsSkillBook() bool
	GetSkillBookValues() (d100.Skill, int)
	GetTextFile() string
	GetText() string
	TextVariables(funcs map[string]govaluate.ExpressionFunction) map[string]string
	SetAlive(isAlive bool)
	GetStatMod(stat d100.Stat) (int, bool)
	GetSkillMod(skill d100.Skill) (int, bool)
	IsBreakingNow() bool
	IsThrowable() bool
	IsStackable() bool
	SetInventoryIndex(i int)
	IsRepairable() bool
	IsHidden() bool
	SetHidden(isHidden bool)
	SetStackSize(count int)
	GetPrice() int
	RemoveStacks(amount int)
	IsGold() bool
}

type ItemCategory int

func (c ItemCategory) String() string {
	switch c {
	case ItemCategoryGold:
		return "Cash"
	case ItemCategoryFood:
		return "Food"
	case ItemCategoryWeapons:
		return "Weapon"
	case ItemCategoryArmor:
		return "Armor"
	case ItemCategoryAmmo:
		return "Ammo"
	case ItemCategoryReadables:
		return "Book"
	case ItemCategoryLockpicks:
		return "Lockpick"
	case ItemCategoryConsumables:
		return "Consumable"
	case ItemCategoryHeadgear:
		return "Headgear"
	case ItemCategoryKeys:
		return "Key"
	case ItemCategoryOther:
		return "Other"
	}
	panic("Unknown item category")
}

func (c ItemCategory) IsEasySteal() bool {
	switch c {
	case ItemCategoryGold, ItemCategoryFood, ItemCategoryAmmo, ItemCategoryReadables, ItemCategoryLockpicks, ItemCategoryKeys:
		return true
	}
	return false
}

func (c ItemCategory) IsHardSteal() bool {
	switch c {
	case ItemCategoryWeapons, ItemCategoryArmor, ItemCategoryHeadgear:
		return true
	}
	return false
}

func (c ItemCategory) IsArmor() bool {
	return c == ItemCategoryArmor || c == ItemCategoryHeadgear
}

const (
	ItemCategoryGold ItemCategory = iota
	ItemCategoryFood
	ItemCategoryWeapons
	ItemCategoryHeadgear
	ItemCategoryArmor
	ItemCategoryAmmo
	ItemCategoryReadables
	ItemCategoryConsumables
	ItemCategoryLockpicks
	ItemCategoryKeys
	ItemCategoryOther
)

func RandomItemCategory() ItemCategory {
	return ItemCategory(rand.Intn(int(ItemCategoryOther) + 1))
}
func ItemCategoryFromString(s string) ItemCategory {
	s = strings.TrimPrefix(strings.ToLower(s), "item")
	switch s {
	case "gold":
		return ItemCategoryGold
	case "food":
		return ItemCategoryFood
	case "weapons":
		return ItemCategoryWeapons
	case "armor":
		return ItemCategoryArmor
	case "ammo":
		return ItemCategoryAmmo
	case "lockpicks":
		return ItemCategoryLockpicks
	case "books":
		return ItemCategoryReadables
	case "drinks":
		return ItemCategoryConsumables
	case "headgear":
		return ItemCategoryHeadgear
	case "consumables":
		return ItemCategoryConsumables
	case "keys":
		return ItemCategoryKeys
	case "notes":
		return ItemCategoryReadables
	case "other":
		return ItemCategoryOther
	}
	panic("Unknown item category: " + s)
}
