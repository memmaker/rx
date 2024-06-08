package game

import (
	"RogueUI/special"
	"fmt"
)

type Protection struct {
	damageReduction int
	damageThreshold int
}

func (p Protection) String() string {
	return fmt.Sprintf("%d|%d%%", p.damageThreshold, p.damageReduction)
}

func (p Protection) Scaled(float float64) Protection {
	return Protection{
		damageReduction: int(float * float64(p.damageReduction)),
		damageThreshold: int(float * float64(p.damageThreshold)),
	}
}

type ArmorInfo struct {
	protection         map[special.DamageType]Protection
	encumbrance        int
	radiationReduction int
	durability         int
}

func (i *ArmorInfo) GetProtection(dType special.DamageType) Protection {
	durabilityAsPercentFloat := float64(i.durability) / 100.0
	protection := i.protection[dType]
	return protection.Scaled(durabilityAsPercentFloat)
}

func (i *ArmorInfo) GetProtectionValueAsString() string {
	physical := i.GetProtection(special.Physical)
	energy := i.GetProtection(special.Energy)
	return fmt.Sprintf("%s %s", physical.String(), energy.String())

}

func (i *ArmorInfo) GetEncumbrance() int {
	return i.encumbrance
}

func (i *ArmorInfo) GetProtectionRating() int {
	physical := i.GetProtection(special.Physical)
	energy := i.GetProtection(special.Energy)

	return (physical.damageReduction + energy.damageReduction) + (physical.damageThreshold + energy.damageThreshold)
}

type ArmorDef struct {
	Protection         map[special.DamageType]Protection
	Encumbrance        int
	RadiationReduction int
}
