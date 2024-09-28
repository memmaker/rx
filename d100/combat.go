package d100

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

	hitChance := min(SuccessChanceCap, computedCtH)

	return hitChance
}

func boolAsInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
