package d100

type BodyStructure []BodyPart

type BodyPart int

var HumanBodyParts = BodyStructure{Body, Eyes, Head, Arms, Groin, Legs}

const (
	Body BodyPart = iota
	Eyes
	Head
	Arms
	Groin
	Legs
)

func (b BodyPart) AimPenalty() int {
	switch b {
	case Body:
		return 0
	case Eyes:
		return -60
	case Head:
		return -40
	case Arms:
		return -30
	case Groin:
		return -30
	case Legs:
		return -20
	}
	return 0
}

func (b BodyPart) String() string {
	switch b {
	case Body:
		return "Body"
	case Eyes:
		return "Eyes"
	case Head:
		return "Head"
	case Arms:
		return "Arms"
	case Groin:
		return "Groin"
	case Legs:
		return "Legs"
	}
	return "Unknown"
}

func (b BodyPart) DamageForCrippled(maxHitpointsOfActor int) int {
	// 30hp, 80hp, 150, 240
	switch b {
	case Body:
		return maxHitpointsOfActor
	case Eyes:
		return maxHitpointsOfActor / 5 // 6, 16, 30, 48
	case Head:
		return maxHitpointsOfActor / 4 // 8, 20, 38, 60
	case Arms:
		return maxHitpointsOfActor / 3 // 10, 26, 50, 80
	case Groin:
		return maxHitpointsOfActor / 6 // 5, 13, 25, 40
	case Legs:
		return maxHitpointsOfActor / 3 // 10, 26, 50, 80
	}
	return maxHitpointsOfActor
}
