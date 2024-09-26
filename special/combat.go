package special

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

type PosInfo struct {
	Distance            int
	ObstacleCount       int
	IlluminationPenalty int // 0 for bright, -40 for darkness
}

func MeleeChanceToHit(attacker *CharSheet, attackerSkill Skill, defender *CharSheet) int {
	s := attacker.GetSkill(attackerSkill)
	str := attacker.GetStat(Strength)
	mws := 0 // TODO: minimum STR for weapon
	h1 := 0  // Set to 1 for 1H weapon
	h2 := 1  // Set to 1 for 2H weapon
	oh := 0  // Set to 1 for One-Handed PERK
	hand := oh * (-40*h2 + 20*h1)

	wh := 0 // Set To 1 for Weapon Handling PERK
	wa := 0 // Set to 1 for Weapon Accuracy PERK

	obstacle := 0
	b := -10 * obstacle
	blind := boolAsInt(false)

	// insufficient strength penalty
	t := -20 * max(0, mws-str-3*wh)

	// illumination penatly
	d := 0

	defenderDodge := defender.GetDerivedStat(Dodge)

	computedCtH := s +
		b +
		t +
		hand +
		20*wa -
		max(0, defenderDodge) +
		d -
		25*blind

	hitChance := min(95, computedCtH)

	return hitChance
}
func RangedChanceToHit(positionInfos PosInfo, attacker *CharSheet, attackerSkill Skill, minWeaponStr int, defender *CharSheet, defenderIsHelpless bool) int {
	s := attacker.GetSkill(attackerSkill)
	p := attacker.GetStat(Perception)
	str := attacker.GetStat(Strength)
	h1 := 0 // Set to 1 for 1H weapon
	h2 := 1 // Set to 1 for 2H weapon
	oh := 0 // Set to 1 for One-Handed PERK
	hand := oh * (-40*h2 + 20*h1)

	wh := 0 // Set To 1 for Weapon Handling PERK
	wa := 0 // Set to 1 for Weapon Accuracy PERK
	lr := 0 // Set to 1 for Long Range PERK
	sr := 0 // Set to 1 for Scope Range PERK

	h := positionInfos.Distance
	obstacle := positionInfos.ObstacleCount
	ranged := boolAsInt(attackerSkill.IsRangedAttackSkill())
	knocked := boolAsInt(defenderIsHelpless)
	b := -10 * obstacle
	blind := boolAsInt(false)

	// range penalty
	rp := sr*(boolAsInt(h < 8)*8+boolAsInt(h >= 8)*-5*(p-2)) +
		sr*(-2-2*lr)*(p-2)

	// perception bonus
	pb := rp + boolAsInt(h+rp < -2*p)*-2*p

	sharp := 0 // TODO: Level of sharp shooter perk

	// insufficient strength penalty
	t := -20 * max(0, minWeaponStr-str-3*wh)

	// illumination penatly
	d := positionInfos.IlluminationPenalty

	rang := ranged * (b + (-4-8*blind)*(h+pb-2*sharp))

	defenderDodge := defender.GetDerivedStat(Dodge)

	relativeSize := 0 // TODO: relative size of attacker and defender

	computedCtH := s +
		rang +
		t +
		hand +
		20*wa -
		max(0, defenderDodge) +
		10*relativeSize +
		d -
		25*blind +
		40*knocked

	hitChance := min(95, computedCtH)

	return hitChance
}

func boolAsInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
