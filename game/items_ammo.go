package game

import (
	"RogueUI/foundation"
	"fmt"
	"github.com/memmaker/go/cview"
	"strings"
)

type Ammo struct {
	*GenericItem
	DamageFactor                    float64
	ConditionFactor                 float64
	SpreadFactor                    float64
	BonusDamageAgainstActorWithTags map[foundation.ActorFlag]int
	DTModifier                      int
	BonusRadius                     int
	RoundsInMagazine                int
	CaliberIndex                    int
}

func (i Ammo) IsAmmo() bool {
	return true
}

func (i Ammo) Equals(other *Ammo) bool {
	return i.DamageFactor == other.DamageFactor &&
		i.DTModifier == other.DTModifier &&
		i.ConditionFactor == other.ConditionFactor &&
		i.RoundsInMagazine == other.RoundsInMagazine &&
		i.CaliberIndex == other.CaliberIndex &&
		i.SpreadFactor == other.SpreadFactor &&
		i.BonusRadius == other.BonusRadius &&
		i.bonusDamageEquals(other.BonusDamageAgainstActorWithTags)
}

func (i Ammo) IsValid() bool {
	return i.DamageFactor != 0 && i.RoundsInMagazine > 0 && i.CaliberIndex > 0 && i.ConditionFactor != 0 && i.SpreadFactor != 0
}

func (i Ammo) IsAmmoOfCaliber(ammo int) bool {
	return i.IsAmmo() && i.CaliberIndex == ammo
}

func (i Ammo) bonusDamageEquals(tags map[foundation.ActorFlag]int) bool {
	if len(i.BonusDamageAgainstActorWithTags) != len(tags) {
		return false
	}
	for tag, damage := range i.BonusDamageAgainstActorWithTags {
		if tags[tag] != damage {
			return false
		}
	}
	return true
}
func (i Ammo) InventoryNameWithColorsAndShortcut(lineColorCode string) string {
	return fmt.Sprintf("%c - %s", i.Shortcut(), i.InventoryNameWithColors(lineColorCode))
}
func (i Ammo) InventoryNameWithColors(colorCode string) string {
	line := cview.Escape(i.Name())

	ammo := i
	line = cview.Escape(fmt.Sprintf("%s [%s] (x%d)", i.Name(), ammo.ShortString(), i.Charges()))

	statPairs := i.getStatPairsAsStrings()

	if len(statPairs) > 0 {
		line = fmt.Sprintf("%s [%s]", line, strings.Join(statPairs, "|"))
	}

	lineWithColor := colorCode + line + "[-]"

	return lineWithColor
}
func (i Ammo) ShortString() string {
	str := strings.Builder{}
	if i.DamageFactor != 1 {
		str.WriteString(fmt.Sprintf("Dmg: x%.2f ", i.DamageFactor))
	}
	if i.ConditionFactor != 1 {
		str.WriteString(fmt.Sprintf("Cnd: x%.2f ", i.ConditionFactor))
	}
	if i.SpreadFactor != 1 {
		str.WriteString(fmt.Sprintf("Spr: x%.2f ", i.SpreadFactor))
	}
	if i.DTModifier != 0 {
		str.WriteString(fmt.Sprintf("DT: %+d ", i.DTModifier))
	}
	if i.BonusRadius != 0 {
		str.WriteString(fmt.Sprintf("Rad: %+d ", i.BonusRadius))
	}
	if len(i.BonusDamageAgainstActorWithTags) > 0 {
		for tag, damage := range i.BonusDamageAgainstActorWithTags {
			str.WriteString(fmt.Sprintf("%+d vs. %s ", damage, tag.StringShort()))
		}
	}
	return strings.TrimSpace(str.String())
}

func (i Ammo) CanStackWith(other foundation.Item) bool {
	if other.IsAmmo() && i.Equals(other.(*Ammo)) {
		return true
	}

	return false
}
