package game

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/cview"
	"strings"
)

type Armor struct {
	*GenericItem
	protection         map[DamageType]Protection
	encumbrance        int
	radiationReduction int
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
func (i *Armor) InventoryNameWithColorsAndShortcut(lineColorCode string) string {
	return fmt.Sprintf("%c - %s", i.Shortcut(), i.InventoryNameWithColors(lineColorCode))
}
func (i *Armor) InventoryNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())

	line = cview.Escape(fmt.Sprintf("%s [%s]", i.Name(), i.GetArmorProtectionValueAsString()))

	statPairs := i.getStatPairsAsStrings()

	if len(statPairs) > 0 {
		line = fmt.Sprintf("%s [%s]", line, strings.Join(statPairs, "|"))
	}

	lineWithColor := colorCode + line + "[-]"

	qIcon := getQualityIcon(i.qualityInPercent)
	lineWithColor = fmt.Sprintf("%s %s", qIcon, lineWithColor)

	return lineWithColor
}
func (i *Armor) LongNameWithColors(colorCode string) string {
	line := cview.Escape(fmt.Sprintf("%s [%s]", i.Name(), i.GetArmorProtectionValueAsString()))
	return colorCode + line + "[-]"
}

func (i *Armor) GetArmorProtection(damageType DamageType) Protection {
	return i.getRawProtection(damageType).Scaled(i.qualityInPercent.Normalized())
}

func (i *Armor) GetArmorProtectionValueAsString() string {
	physical := i.GetArmorProtection(DamageTypeNormal)
	energy := i.GetArmorProtection(DamageTypeLaser)
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
	protection := i.protection[dType]
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
	return len(i.protection) != 0
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
