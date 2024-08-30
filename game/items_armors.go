package game

import (
	"RogueUI/special"
	"bytes"
	"encoding/gob"
	"fmt"
)

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

type ArmorInfo struct {
	protection         map[special.DamageType]Protection
	encumbrance        int
	radiationReduction int
	durability         int
}

func (i *ArmorInfo) GobEncode() ([]byte, error) {
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

	if err := encoder.Encode(i.durability); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (i *ArmorInfo) GobDecode(data []byte) error {
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

	if err := decoder.Decode(&i.durability); err != nil {
		return err
	}

	return nil
}

func (i *ArmorInfo) GetProtection(dType special.DamageType) Protection {
	durabilityAsPercentFloat := float64(i.durability) / 100.0
	protection := i.protection[dType]
	return protection.Scaled(durabilityAsPercentFloat)
}

func (i *ArmorInfo) GetProtectionValueAsString() string {
	physical := i.GetProtection(special.DamageTypeNormal)
	energy := i.GetProtection(special.DamageTypeLaser)
	return fmt.Sprintf("%s %s", physical.String(), energy.String())

}

func (i *ArmorInfo) GetEncumbrance() int {
	return i.encumbrance
}

func (i *ArmorInfo) GetProtectionRating() int {
	physical := i.GetProtection(special.DamageTypeNormal)
	energy := i.GetProtection(special.DamageTypeLaser)

	return (physical.DamageReduction + energy.DamageReduction) + (physical.DamageThreshold + energy.DamageThreshold)
}

func (i *ArmorInfo) IsValid() bool {
	return len(i.protection) != 0
}

type ArmorDef struct {
	Protection         map[special.DamageType]Protection
	Encumbrance        int
	RadiationReduction int
}

func (d ArmorDef) IsValid() bool {
	return len(d.Protection) != 0
}
