package foundation

import (
	"RogueUI/special"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"strings"
)

type Readable interface {
	IsSkillBook() bool
	GetSkillBookValues() (special.Skill, int)
	GetTextFile() string
	GetText() string
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
	Quality() special.Percentage
	SetQuality(quality special.Percentage)
	NeedsRepair() bool
	Category() ItemCategory
	InternalName() string
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
	Charges() int
	AfterEquippedTurn()
	InternalName() string
}

type Timable interface {
	Zappable
	Name() string
	HasTag(timed ItemTags) bool
	SetCharges(turns int)

	ShouldActivate(tickCount int) bool
	IsAlive(tickCount int) bool
	String() string
	Position() geometry.Point
}
type Item interface {
	Name() string
	String() string
	Category() ItemCategory
	InventoryNameWithColors(lineColorCode string) string
	InventoryNameWithColorsAndShortcut(invItemColorCode string) string
	Description() string
	Shortcut() rune
	DisplayLength() int
	Position() geometry.Point
	SetPosition(position geometry.Point)
	LongNameWithColors(colorCode string) string
	GetIcon() textiles.TextIcon
	GetCarryWeight() int
	GetDerivedStatMod(stat special.DerivedStat) (int, bool)
	ShouldActivate(tickCount int) bool
	IsAlive(tickCount int) bool

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

	// Stacking
	IsMultipleStacks() bool
	StackSize() int
	Split(count int) Item
	CanStackWith(item Item) bool
	AddStacks(item Item)

	GetLockFlag() string
	GetThrowDamage() fxtools.Interval
	InternalName() string

	GetEffectParameters() Params
	ZapEffect() string
	UseEffect() string
	Charges() int
	SetCharges(count int)
	SetQuality(qualityInPercent special.Percentage)

	Quality() special.Percentage
	GetEquipFlag() ActorFlag
	NeedsRepair() bool
	Color() color.RGBA
	SetPositionHandler(pos func() geometry.Point)
	ConsumeCharge()
	HasTag(loot ItemTags) bool

	AfterEquippedTurn()
	DropFlag() string
	PickupFlag() string

	CanBeRepairedWith(parts Repairable) bool

	IsSkillBook() bool
	GetSkillBookValues() (special.Skill, int)
	GetTextFile() string
	GetText() string
	SetAlive(isAlive bool)
	GetStatMod(stat special.Stat) (int, bool)
	GetSkillMod(skill special.Skill) (int, bool)
	IsBreakingNow() bool
	IsThrowable() bool
	IsStackable() bool
	SetInventoryIndex(i int)
	IsRepairable() bool
}

type ItemCategory int

func (c ItemCategory) String() string {
	switch c {
	case ItemCategoryGold:
		return "Gold"
	case ItemCategoryFood:
		return "Food"
	case ItemCategoryWeapons:
		return "Weapons"
	case ItemCategoryArmor:
		return "Armor"
	case ItemCategoryAmmo:
		return "Ammo"
	case ItemCategoryReadables:
		return "Books"
	case ItemCategoryLockpicks:
		return "Lockpicks"
	case ItemCategoryConsumables:
		return "Consumables"
	case ItemCategoryKeys:
		return "Keys"
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
	case ItemCategoryWeapons, ItemCategoryArmor:
		return true
	}
	return false
}

const (
	ItemCategoryGold ItemCategory = iota
	ItemCategoryFood
	ItemCategoryWeapons
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
