package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"bytes"
	"cmp"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"math"
	"slices"
	"strings"
)

type Armor struct {
	*GenericItem
	protection         map[DamageType]Protection
	encumbrance        int
	radiationReduction int

	fashionStyle foundation.FashionStyle

	concealSlots []WeaponSize
}

func (i *Armor) CanConceal(weapons []*Weapon) bool {
	slotsAvailable := slices.Clone(i.concealSlots)
	// smallest slots first
	slices.SortStableFunc(slotsAvailable, func(i, j WeaponSize) int {
		return cmp.Compare(i, j)
	})
	// biggest weapons first
	slices.SortStableFunc(weapons, func(i, j *Weapon) int {
		return cmp.Compare(j.relativeSize, i.relativeSize)
	})
	hasSlot := func(weaponOfSize WeaponSize) int {
		for index, availableSlot := range slotsAvailable {
			if availableSlot.CanFit(weaponOfSize) {
				return index
			}
		}
		return -1
	}
	popSlot := func(index int) {
		slotsAvailable = append(slotsAvailable[:index], slotsAvailable[index+1:]...)
	}
	for _, weapon := range weapons {
		size := weapon.relativeSize
		if index := hasSlot(size); index != -1 {
			popSlot(index) // we can fit this weapon
		} else {
			return false // we can't fit this weapon
		}
	}
	return true
}

func (i *Armor) IsArmor() bool {
	return true
}

func (i *Armor) IsEquippable() bool {
	return true
}
func (i *Armor) IsRepairable() bool {
	return true
}

func (i *Armor) FullDescription(colorCode string) string {
	basicRows := i.GenericItem.fullDescriptionRows()

	basicRows = append([]fxtools.TableRow{fxtools.NewTableRow("Style", i.fashionStyle.String())}, basicRows...)

	appendIfNotZero := func(value int, name string) {
		if value != 0 {
			basicRows = append(basicRows, fxtools.NewTableRow(name, fmt.Sprintf("%+d", value)))
		}
	}
	basicRows = append(basicRows, fxtools.NewTableRow("Quality", fmt.Sprintf("%d%%", int(i.qualityInPercent))))
	appendIfNotZero(i.GetEncumbrance(), "Encumbrance")

	for dType := DamageType(0); dType < DamageTypeCount; dType++ {
		if protection, exists := i.protection[dType]; exists {
			protection = protection.Scaled(i.qualityInPercent.Normalized())
			protLabel := fmt.Sprintf("Protection vs. %s", dType.String())
			basicRows = append(basicRows, fxtools.NewTableRow(protLabel, protection.String()))
		}
	}
	appendIfNotZero(i.radiationReduction, "Radiation Reduction")

	lines := fxtools.TableLayout(basicRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	lines = append([]string{i.InventoryNameWithColors(colorCode), i.category.String()}, lines...)

	if len(i.concealSlots) > 0 {
		lines = append(lines, "Conceals:")
		lines = append(lines, i.concealSlotsAsStrings()...)
	}

	lines = i.appendText(lines)

	return strings.Join(lines, "\n")
}
func (i *Armor) InventoryNameWithColorsAndShortcut(lineColorCode string) string {
	return fmt.Sprintf("%c - %s", i.Shortcut(), i.InventoryNameWithColors(lineColorCode))
}

func (i *Armor) LongNameWithColors(colorCode string) string {
	var baseName string
	if len(i.protection) == 0 || !i.HasProtectionValue() {
		baseName = i.Name()
	} else {
		baseName = fmt.Sprintf("%s [%d]", i.Name(), i.GetProtectionRating())
	}

	line := cview.Escape(baseName)

	statPairs := i.getStatPairsAsStrings()

	if len(statPairs) > 0 {
		withStats := fmt.Sprintf("%s [%s]", line, strings.Join(statPairs, "|"))
		line = cview.Escape(withStats)
	}

	lineWithColor := colorCode + line + "[-]"

	qIcon := getQualityIcon(i.qualityInPercent)
	lineWithColor = fmt.Sprintf("%s %s", qIcon, lineWithColor)

	return lineWithColor
}

func (i *Armor) InventoryNameWithColors(colorCode string) string {
	baseName := i.InventoryName()

	lineWithColor := colorCode + baseName + "[-]"

	qIcon := getQualityIcon(i.qualityInPercent)
	lineWithColor = fmt.Sprintf("%s %s", qIcon, lineWithColor)

	return lineWithColor
}

func (i *Armor) InventoryName() string {
	var baseName string
	if len(i.protection) == 0 || !i.HasProtectionValue() {
		baseName = i.Name()
	} else {
		baseName = fmt.Sprintf("%s [%d]", i.Name(), i.GetProtectionRating())
	}
	return baseName
}
func (i *Armor) DisplayLength() int {
	return cview.TaggedStringWidth(i.InventoryNameWithColorsAndShortcut("[red]"))
}
func (i *Armor) GetArmorProtection(damageType DamageType) Protection {
	/*
		if damageType != DamageTypeNormal {
			return i.getRawProtection(DamageTypeEnergy)
		}
	*/
	return i.getRawProtection(damageType).Scaled(i.qualityInPercent.Normalized())
}

func (i *Armor) GetArmorProtectionValueAsString() string {
	physical := i.GetArmorProtection(DamageTypeNormal)
	energy := i.GetArmorProtection(DamageTypeEnergy)
	return fmt.Sprintf("%s %s", physical.String(), energy.String())

}

func (i *Armor) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Encode each field of the struct in order

	if err := encoder.Encode(i.protection); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.encumbrance); err != nil {
		return nil, err
	}

	if err := encoder.Encode(i.radiationReduction); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (i *Armor) GobDecode(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// Decode each field of the struct in order

	if err := decoder.Decode(&i.protection); err != nil {
		return err
	}

	if err := decoder.Decode(&i.encumbrance); err != nil {
		return err
	}

	if err := decoder.Decode(&i.radiationReduction); err != nil {
		return err
	}

	return nil
}

func (i *Armor) getRawProtection(dType DamageType) Protection {
	protection, exists := i.protection[dType]
	if !exists {
		return Protection{}
	}
	return protection
}

func (i *Armor) GetEncumbrance() int {
	return i.encumbrance
}

func (i *Armor) GetProtectionRating() int {
	physical := i.GetArmorProtection(DamageTypeNormal)
	energy := i.GetArmorProtection(DamageTypeLaser)

	return (physical.DamageReduction + energy.DamageReduction) + (physical.DamageThreshold + energy.DamageThreshold)
}

func (i *Armor) IsValid() bool {
	return i.category == foundation.ItemCategoryArmor || i.category == foundation.ItemCategoryHeadgear
}

func (i *Armor) DegradeDT(ablation int) {
	for dmg, protection := range i.protection {
		i.protection[dmg] = protection.WithDTReducedBy(ablation)
	}
}

func (i *Armor) HasProtectionValue() bool {
	for _, protection := range i.protection {
		protection = protection.Scaled(i.qualityInPercent.Normalized())
		if protection.DamageReduction != 0 || protection.DamageThreshold != 0 {
			return true
		}
	}
	return false
}

func (i *Armor) concealSlotsAsStrings() []string {
	slotCount := make(map[WeaponSize]int)
	for _, slot := range i.concealSlots {
		slotCount[slot]++
	}
	result := make([]string, 0, len(slotCount))
	for weaponSize := WeaponSize(1); weaponSize < SizeCount; weaponSize++ {
		if slotCount[weaponSize] > 0 {
			result = append(result, fmt.Sprintf(" %dx %s", slotCount[weaponSize], weaponSize.String()))
		}
	}
	return result
}

type Protection struct {
	DamageReduction int
	DamageThreshold int
}

func (p Protection) String() string {
	return fmt.Sprintf("%d|%d%%", p.DamageThreshold, p.DamageReduction)
}

func (p Protection) Scaled(float float64) Protection {
	return Protection{
		DamageReduction: int(float * float64(p.DamageReduction)),
		DamageThreshold: int(float * float64(p.DamageThreshold)),
	}
}

func (p Protection) WithDTReducedBy(ablation int) Protection {
	return Protection{
		DamageReduction: p.DamageReduction,
		DamageThreshold: max(0, p.DamageThreshold-ablation),
	}
}

func (p Protection) WithReduction(reduction int) Protection {
	return Protection{
		DamageReduction: reduction,
		DamageThreshold: p.DamageThreshold,
	}
}

func (p Protection) WithThreshold(threshold int) Protection {
	return Protection{
		DamageReduction: p.DamageReduction,
		DamageThreshold: threshold,
	}
}

type ArmorWeight int

const (
	ArmorWeightLight ArmorWeight = iota
	ArmorWeightMedium
	ArmorWeightHeavy
	ArmorWeightVeryHeavy
)

type ArmorCondition int

const (
	ArmorConditionNew            ArmorCondition = 100
	ArmorConditionUsed                          = 75
	ArmorConditionDamaged                       = 50
	ArmorConditionHeavilyDamaged                = 25
)

// random_armor(light, new)

func DefaultRandomizedArmorFromType(armorType ArmorWeight, quality ArmorCondition) *Armor {
	roll := d100.Die()
	prot, encumbrance := protectionAndEncumbranceFromRoll(armorType, roll)
	weight, cost := weightAndCostFromRoll(armorType, roll)
	armorName := randomArmorName(armorType, roll)
	newArmor := &Armor{
		GenericItem: &GenericItem{
			name:             armorName,
			internalName:     "default_randomized_armor",
			category:         foundation.ItemCategoryArmor,
			qualityInPercent: d100.Percentage(quality),
			stackSize:        1,
			charges:          -1,
			thrownDamage:     fxtools.Interval{Min: 1, Max: 2},
			weight:           weight,
			cost:             cost,
			alive:            true,
		},
		protection:   prot,
		encumbrance:  encumbrance,
		fashionStyle: foundation.FashionStyleCombatGear,
	}
	return newArmor
}

func randomArmorName(armorType ArmorWeight, roll int) string {
	switch armorType {
	case ArmorWeightLight:
		lightNames := []string{
			"Padded Shirt",
			"Light Leather Jacket",
			"Light Armor Jacket",
		}
		return lightNames[roll%len(lightNames)]
	case ArmorWeightMedium:
		mediumNames := []string{
			"Kevlar T-Shirt",
			"Kevlar Vest",
			"Medium Armor Jacket",
		}
		return mediumNames[roll%len(mediumNames)]
	case ArmorWeightHeavy:
		heavyNames := []string{
			"Flack Vest",
			"Door Gunner's Vest",
			"Heavy Armor Jacket",
			"Heavy Coat",
			"Heavy Long Coat",
		}
		return heavyNames[roll%len(heavyNames)]
	case ArmorWeightVeryHeavy:
		veryHeavyNames := []string{
			"Full Body Armor",
			"Combat Armor",
			"Powered Armor",
			"ACPA",
			"Exoskeleton",
			"Metal Gear",
		}
		return veryHeavyNames[roll%len(veryHeavyNames)]
	}
	return ""
}

func weightAndCostFromRoll(armorType ArmorWeight, roll int) (int, int) {
	switch armorType {
	case ArmorWeightLight:
		return choicesFromRoll(roll, 1, 2), choicesFromRoll(roll, 100, 200)
	case ArmorWeightMedium:
		return choicesFromRoll(roll, 2, 3), choicesFromRoll(roll, 200, 300)
	case ArmorWeightHeavy:
		return choicesFromRoll(roll, 3, 4), choicesFromRoll(roll, 1300, 1400)
	case ArmorWeightVeryHeavy:
		return choicesFromRoll(roll, 4, 5), choicesFromRoll(roll, 1400, 1500)
	}
	return 0, 0
}

func protectionAndEncumbranceFromRoll(armorType ArmorWeight, roll int) (map[DamageType]Protection, int) {
	switch armorType {
	case ArmorWeightLight:
		return map[DamageType]Protection{
			DamageTypeNormal: {
				DamageReduction: 0,
				DamageThreshold: choicesFromRoll(roll, 1, 2),
			},
			DamageTypeEnergy: {
				DamageReduction: 0,
				DamageThreshold: choicesFromRoll(roll, 1, 2),
			},
		}, choicesFromRoll(roll, 1, 2)
	case ArmorWeightMedium:
		return map[DamageType]Protection{
			DamageTypeNormal: {
				DamageReduction: 0,
				DamageThreshold: choicesFromRoll(roll, 2, 3, 4),
			},
			DamageTypeEnergy: {
				DamageReduction: 0,
				DamageThreshold: choicesFromRoll(roll, 2, 3, 4),
			},
		}, choicesFromRoll(roll, 2, 3, 4)
	case ArmorWeightHeavy:
		return map[DamageType]Protection{
			DamageTypeNormal: {
				DamageReduction: choicesFromRoll(roll, 2, 4, 6),
				DamageThreshold: choicesFromRoll(roll, 6, 8, 10, 12, 14),
			},
			DamageTypeEnergy: {
				DamageReduction: choicesFromRoll(roll, 2, 4),
				DamageThreshold: choicesFromRoll(roll, 5, 8, 10),
			},
		}, choicesFromRoll(roll, 2, 4, 6)
	case ArmorWeightVeryHeavy:
		return map[DamageType]Protection{
			DamageTypeNormal: {
				DamageReduction: choicesFromRoll(roll, 8, 10, 12, 14, 16),
				DamageThreshold: choicesFromRoll(roll, 14, 16, 18, 20, 22),
			},
			DamageTypeEnergy: {
				DamageReduction: choicesFromRoll(roll, 8, 10),
				DamageThreshold: choicesFromRoll(roll, 12, 14, 16),
			},
		}, choicesFromRoll(roll, 8, 10)
	}
	return nil, 0
}

func choicesFromRoll(roll int, values ...int) int {
	choiceCount := len(values)                                // 3
	intervalLength := float64(100) / float64(choiceCount)     // 33.333
	choice := int(math.Round(float64(roll) / intervalLength)) // roll = 10 ->
	if choice >= choiceCount {
		choice = choiceCount - 1
	}
	return values[choice]
}
