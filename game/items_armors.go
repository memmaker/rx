package game

import (
    "RogueUI/foundation"
    "bytes"
    "encoding/gob"
    "fmt"
    "github.com/memmaker/go/cview"
    "github.com/memmaker/go/fxtools"
    "strings"
)

type Armor struct {
    *GenericItem
    protection         map[DamageType]Protection
    encumbrance        int
    radiationReduction int

    fashionStyle FashionStyle
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

    appendIfNotZero := func(value int, name string) {
        if value != 0 {
            basicRows = append(basicRows, fxtools.NewTableRow(name, fmt.Sprintf("%+d", value)))
        }
    }
    basicRows = append(basicRows, fxtools.NewTableRow("Quality", fmt.Sprintf("%d%%", int(i.qualityInPercent))))
    basicRows = append(basicRows, fxtools.NewTableRow("Encumbrance", fmt.Sprintf("%d", i.encumbrance)))

    for dType := DamageType(0); dType < DamageTypeCount; dType++ {
        if protection, exists := i.protection[dType]; exists {
            protection.Scaled(i.qualityInPercent.Normalized())
            protLabel := fmt.Sprintf("Protection vs. %s", dType.String())
            basicRows = append(basicRows, fxtools.NewTableRow(protLabel, protection.String()))
        }
    }
    appendIfNotZero(i.radiationReduction, "Radiation Reduction")

    lines := fxtools.TableLayout(basicRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
    lines = append([]string{i.InventoryNameWithColors(colorCode), i.category.String(), i.fashionStyle.String()}, lines...)
    return strings.Join(lines, "\n")
}
func (i *Armor) InventoryNameWithColorsAndShortcut(lineColorCode string) string {
    return fmt.Sprintf("%c - %s", i.Shortcut(), i.InventoryNameWithColors(lineColorCode))
}

func (i *Armor) LongNameWithColors(colorCode string) string {
    var baseName string
    if len(i.protection) == 0 {
        baseName = i.Name()
    } else {
        baseName = fmt.Sprintf("%s [%s]", i.Name(), i.GetArmorProtectionValueAsString())
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
    var baseName string
    if len(i.protection) == 0 {
        baseName = i.Name()
    } else {
        baseName = fmt.Sprintf("%s [%s]", i.Name(), i.GetArmorProtectionValueAsString())
    }

    lineWithColor := colorCode + baseName + "[-]"

    qIcon := getQualityIcon(i.qualityInPercent)
    lineWithColor = fmt.Sprintf("%s %s", qIcon, lineWithColor)

    return lineWithColor
}
func (i *Armor) DisplayLength() int {
    return cview.TaggedStringWidth(i.InventoryNameWithColorsAndShortcut("[red]"))
}
func (i *Armor) GetArmorProtection(damageType DamageType) Protection {
    if damageType != DamageTypeNormal {
        return i.getRawProtection(DamageTypeEnergy)
    }
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
    physical := i.getRawProtection(DamageTypeNormal)
    energy := i.getRawProtection(DamageTypeLaser)

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

type FashionStyle int8

func (s FashionStyle) PriceMultiplier() int {
    switch s {
    case FashionStyleCombatGear:
        return 1
    case FashionStyleLowLife:
        return 1
    case FashionStyleChromeGang:
        return 1
    case FashionStyleNomadLeather:
        return 1
    case FashionStyleGenericChic:
        return 1
    case FashionStyleEdgerunner:
        return 3
    case FashionStyleBusiness:
        return 3
    case FashionStyleHighFashion:
        return 4
    }
    return 1
}

func (s FashionStyle) String() string {
    switch s {
    case FashionStyleCombatGear:
        return "Combat Gear"
    case FashionStyleLowLife:
        return "Low Life"
    case FashionStyleChromeGang:
        return "Gang Colors"
    case FashionStyleNomadLeather:
        return "Nomad Leather"
    case FashionStyleGenericChic:
        return "Generic Chic"
    case FashionStyleEdgerunner:
        return "Edgerunner"
    case FashionStyleBusiness:
        return "Business"
    case FashionStyleHighFashion:
        return "High Fashion"
    }
    return "Unknown"

}

const FashionStyleRandom FashionStyle = -1
const (
    FashionStyleCombatGear = iota
    FashionStyleLowLife
    FashionStyleChromeGang
    FashionStyleNomadLeather
    FashionStyleGenericChic
    FashionStyleEdgerunner
    FashionStyleBusiness
    FashionStyleHighFashion
    FashionStyleCount
)

func FashionStyleFromString(str string) FashionStyle {
    switch strings.ToLower(str) {
    case "random":
        return FashionStyleRandom
    case "combat_gear":
        return FashionStyleCombatGear
    case "low_life":
        return FashionStyleLowLife
    case "gang_colors":
        return FashionStyleChromeGang
    case "nomad_leather":
        return FashionStyleNomadLeather
    case "generic_chic":
        return FashionStyleGenericChic
    case "edgerunner":
        return FashionStyleEdgerunner
    case "business":
        return FashionStyleBusiness
    case "high_fashion":
        return FashionStyleHighFashion
    default:
        return FashionStyleRandom
    }
}
