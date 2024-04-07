package game

import "math/rand"

type IdentificationKnowledge struct {
	identifiedItemTypes map[string]bool

	potionMap map[string]string // potion color -> real name
	ringMap   map[string]string // potion color -> real name
	scrollMap map[string]string // potion color -> real name
	wandMap   map[string]string // potion color -> real name

	alwaysIDOnUse map[string]bool
}

func NewIdentificationKnowledge() *IdentificationKnowledge {
	return &IdentificationKnowledge{
		identifiedItemTypes: make(map[string]bool),
	}
}

func (i *IdentificationKnowledge) SetAlwaysIDOnUse(names []string) {
	i.alwaysIDOnUse = make(map[string]bool)
	for _, name := range names {
		i.alwaysIDOnUse[name] = true
	}
}

func (i *IdentificationKnowledge) IsItemIdentified(name string) bool {
	return i.identifiedItemTypes[name]
}

func (i *IdentificationKnowledge) IdentifyItem(name string) {
	i.identifiedItemTypes[name] = true
}
func (i *IdentificationKnowledge) MixPotions(potionNames []string) {

	var colorNames = []string{ // for potions
		"amber",
		"aquamarine",
		"black",
		"blue",
		"brown",
		"clear",
		"crimson",
		"cyan",
		"ecru",
		"gold",
		"green",
		"grey",
		"magenta",
		"orange",
		"pink",
		"plaid",
		"purple",
		"red",
		"silver",
		"tan",
		"tangerine",
		"topaz",
		"turquoise",
		"vermilion",
		"violet",
		"white",
		"yellow",
	}

	// colors
	randomIndexes := rand.Perm(len(colorNames))
	popNextColor := func() string {
		nextColor := colorNames[randomIndexes[0]]
		randomIndexes = randomIndexes[1:]
		return nextColor
	}
	i.potionMap = make(map[string]string)
	for _, name := range potionNames {
		i.potionMap[name] = popNextColor()
	}
}

func (i *IdentificationKnowledge) GetPotionColor(name string) string {
	return i.potionMap[name]
}

func (i *IdentificationKnowledge) MixRings(ringNames []string) {

	var stoneNames = []string{ // for rings
		"agate",
		"alexandrite",
		"amethyst",
		"carnelian",
		"diamond",
		"emerald",
		"germanium",
		"granite",
		"garnet",
		"jade",
		"kryptonite",
		"lapis lazuli",
		"moonstone",
		"obsidian",
		"onyx",
		"opal",
		"pearl",
		"peridot",
		"ruby",
		"sapphire",
		"stibotantalite",
		"tiger eye",
		"topaz",
		"turquoise",
		"taaffeite",
		"zircon",
	}

	// colors
	randomIndexes := rand.Perm(len(stoneNames))
	popNextStone := func() string {
		nextStone := stoneNames[randomIndexes[0]]
		randomIndexes = randomIndexes[1:]
		return nextStone
	}
	i.ringMap = make(map[string]string)
	for _, name := range ringNames {
		i.ringMap[name] = popNextStone()
	}
}

func (i *IdentificationKnowledge) GetRingStone(name string) string {
	return i.ringMap[name]
}

func (i *IdentificationKnowledge) MixScrolls(internalNames []string) {
	var scrollNames = []string{ // idea: transport a message by randomly ceasar shifting the text
		"udug hul",
		"klaatu barada nikto",
		"sazun hera duoder",
		"eiris sazun idisi",
		"sic gorgiamus allos subjectatos nunc",
		"secht tonna tacid dom dorodailter",
		"adeochosa inna husci do chongnam frim",
		"cris nathrach mo chris",
		"vas ort flam",
		"an lor xen",
		"vas des sanct",
		"scientia est potentia",
		"sine ira et studio",
		"tempus fugit",
		"fiat lux",
		"homo homini lupus",
		"nihil ex nihilo",
		"verbum terroris",
		"timor locum istum permeat",
		"aquafaxius scaturient aquae",
		"archosphaero pugnus aereus",
		"armatura scutum et praesidium",
		"auris illusionis aurium fallacia",
		"bannbaladin ego sum amicus tuus",
		"globus tonitrui",
		"ardebit in crepitantibus ira",
		"contra daemones et prolem",
		"tutela contra daemones",
		"vinculum infirmitatis",
		"dissolve sicut folia arida",
		"virtus lucis abscessit",
		"duplicatus duplex dolor",
		"maledictus oculus",
	}
	// colors
	randomIndexes := rand.Perm(len(scrollNames))
	popNextScroll := func() string {
		nextScroll := scrollNames[randomIndexes[0]]
		randomIndexes = randomIndexes[1:]
		return nextScroll
	}
	i.scrollMap = make(map[string]string)
	for _, name := range internalNames {
		i.scrollMap[name] = popNextScroll()
	}
}

func (i *IdentificationKnowledge) GetScrollName(name string) string {
	return i.scrollMap[name]
}

func (i *IdentificationKnowledge) MixWands(wandNames []string) {

	var woodNames = []string{ // for wands
		"avocado wood",
		"balsa",
		"bamboo",
		"banyan",
		"birch",
		"cedar",
		"cherry",
		"cinnibar",
		"cypress",
		"dogwood",
		"driftwood",
		"ebony",
		"elm",
		"eucalyptus",
		"fall",
		"hemlock",
		"holly",
		"ironwood",
		"kukui wood",
		"mahogany",
		"manzanita",
		"maple",
		"oaken",
		"persimmon wood",
		"pecan",
		"pine",
		"poplar",
		"redwood",
		"rosewood",
		"spruce",
		"teak",
		"walnut",
		"zebrawood",
	}

	var metalNames = []string{ // for wands
		"aluminum",
		"beryllium",
		"bone",
		"brass",
		"bronze",
		"copper",
		"electrum",
		"gold",
		"iron",
		"lead",
		"magnesium",
		"mercury",
		"nickel",
		"pewter",
		"platinum",
		"steel",
		"silver",
		"silicon",
		"tin",
		"titanium",
		"tungsten",
		"zinc",
	}

	woodIndexes := rand.Perm(len(woodNames))
	metalIndexes := rand.Perm(len(metalNames))
	popNextMaterial := func() string {
		if rand.Intn(2) == 0 {
			nextWood := woodNames[woodIndexes[0]]
			woodIndexes = woodIndexes[1:]
			return nextWood
		} else {
			nextMetal := metalNames[metalIndexes[0]]
			metalIndexes = metalIndexes[1:]
			return nextMetal
		}
	}
	i.wandMap = make(map[string]string)
	for _, name := range wandNames {
		i.wandMap[name] = popNextMaterial()
	}
}

func (i *IdentificationKnowledge) GetWandMaterial(name string) string {
	return i.wandMap[name]
}

func (i *IdentificationKnowledge) CanBeIdentifiedByUsing(name string) bool {
	if i.IsItemIdentified(name) {
		return false
	}

	if i.alwaysIDOnUse[name] {
		return true
	}

	return false
}
