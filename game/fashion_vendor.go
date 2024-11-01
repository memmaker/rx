package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"fmt"
	"github.com/memmaker/go/textiles"
	"math/rand"
)

type ClothingDescription struct {
	Name         string
	BasePrice    int
	CanBeLong    bool
	CanBeLeather bool
	CanBeHeavy   bool
}

func newFashionInventory(style foundation.FashionStyle) []foundation.Item {
	// Types and base prices
	var armorTypes = []ClothingDescription{
		{
			Name:         "Coat",
			BasePrice:    100,
			CanBeLong:    true,
			CanBeLeather: true,
			CanBeHeavy:   true,
		},
		{
			Name:         "Duster",
			BasePrice:    300,
			CanBeLong:    true,
			CanBeLeather: true,
			CanBeHeavy:   true,
		},
		{
			Name:         "Trenchcoat",
			BasePrice:    250,
			CanBeLong:    true,
			CanBeLeather: true,
			CanBeHeavy:   true,
		},
		{
			Name:         "Jacket",
			BasePrice:    150,
			CanBeLong:    true,
			CanBeLeather: true,
			CanBeHeavy:   true,
		},
		{
			Name:         "Vest",
			BasePrice:    85,
			CanBeLong:    false,
			CanBeLeather: true,
			CanBeHeavy:   true,
		},
		{
			Name:         "T-Shirt",
			BasePrice:    30,
			CanBeLong:    false,
			CanBeLeather: false,
			CanBeHeavy:   false,
		},
		{
			Name:         "Tank Top",
			BasePrice:    25,
			CanBeLong:    false,
			CanBeLeather: false,
			CanBeHeavy:   false,
		},
		{
			Name:         "Bodysuit",
			BasePrice:    110,
			CanBeLong:    false,
			CanBeLeather: false,
			CanBeHeavy:   false,
		},
		{
			Name:         "Jumpsuit",
			BasePrice:    125,
			CanBeLong:    false,
			CanBeLeather: false,
			CanBeHeavy:   true,
		},
		{
			Name:         "Suit",
			BasePrice:    450,
			CanBeLong:    false,
			CanBeLeather: false,
			CanBeHeavy:   false,
		},
		{
			Name:         "Tuxedo",
			BasePrice:    500,
			CanBeLong:    false,
			CanBeLeather: false,
			CanBeHeavy:   false,
		},
		{
			Name:         "Dress",
			BasePrice:    300,
			CanBeLong:    true,
			CanBeLeather: false,
			CanBeHeavy:   true,
		},
	}

	var fashionInventory []foundation.Item

	for i := 0; i < 10; i++ {
		randomArmorIndex := rand.Intn(len(armorTypes))
		chosenType := armorTypes[randomArmorIndex]

		newFashionItem := generateNewFashionItem(chosenType, style)
		fashionInventory = append(fashionInventory, newFashionItem)
	}

	return fashionInventory
}

func generateNewFashionItem(desc ClothingDescription, forceStyle foundation.FashionStyle) *Armor {
	brandsForStyles := map[foundation.FashionStyle][]string{
		foundation.FashionStyleGenericChic: {
			"Levi",
			"Nu-Tek",
			"Uniwear",
			"Gibson Battlegear",
			"Jordash/Boy",
		},
		foundation.FashionStyleEdgerunner: {
			"Image Fashionware",
			"Icon America",
			"Cryo-Max",
		},
		foundation.FashionStyleHighFashion: {
			"Image Fashionware",
			"Cryo-Max",
		},
		foundation.FashionStyleBusiness: {
			"Takanaka",
		},
	}

	baseNormalDT := 0
	pieceName := desc.Name
	basePrice := desc.BasePrice

	isLeather := desc.CanBeLeather && rand.Intn(2) == 0

	if isLeather {
		pieceName = fmt.Sprintf("Leather %s", pieceName)
		basePrice *= 2
		baseNormalDT += 1
	}

	isLong := desc.CanBeLong && rand.Intn(2) == 0

	if isLong {
		pieceName = fmt.Sprintf("Long %s", pieceName)
		basePrice = int(float64(basePrice) * 1.2)
		baseNormalDT += 1
	}

	isHeavy := desc.CanBeHeavy && rand.Intn(2) == 0

	if isHeavy {
		pieceName = fmt.Sprintf("Heavy %s", pieceName)
		basePrice = int(float64(basePrice) * 1.2)
		baseNormalDT += 1
	}

	chosenStyle := forceStyle
	if chosenStyle < 0 || chosenStyle >= foundation.FashionStyleCount {
		chosenStyle = foundation.FashionStyle(rand.Intn(int(foundation.FashionStyleCount)))
	}

	brands := brandsForStyles[chosenStyle]
	chosenBrand := brands[rand.Intn(len(brands))]

	pieceName = fmt.Sprintf("%s %s", chosenBrand, pieceName)

	basePrice *= chosenStyle.PriceMultiplier()
	basePrice = int(float64(basePrice))

	return &Armor{
		GenericItem: &GenericItem{
			name:             pieceName,
			internalName:     "fashion_piece",
			category:         foundation.ItemCategoryArmor,
			qualityInPercent: d100.Percentage(100),
			stackSize:        1,
			statChanges:      StatChange{},
			icon:             textiles.TextIcon{},
			weight:           0,
			cost:             basePrice,
		},
		fashionStyle: chosenStyle,

		protection: map[DamageType]Protection{
			DamageTypeNormal: Protection{
				DamageReduction: 0,
				DamageThreshold: baseNormalDT,
			},
			DamageTypeEnergy: Protection{
				DamageReduction: 0,
				DamageThreshold: 0,
			},
		},

		encumbrance: 0,
	}
}
