package dice_curve

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

type Dice struct {
	dice  int
	sides int
	bonus int
}

func NewDice(dice, sides, bonus int) Dice {
	return Dice{dice, sides, bonus}
}

func (r Dice) Max() int {
	return (r.dice * r.sides) + r.bonus
}

func (r Dice) Min() int {
	return r.dice + r.bonus
}

func (r Dice) String() string {
	if r.isFixed() {
		flatValue := r.Min()
		return fmt.Sprintf("%d", flatValue)
	}
	if r.bonus == 0 {
		return fmt.Sprintf("%dd%d (%d-%d)", r.dice, r.sides, r.Min(), r.Max())
	}
	return fmt.Sprintf("%dd%d%+d (%d-%d)", r.dice, r.sides, r.bonus, r.Min(), r.Max())
}

func (r Dice) isFixed() bool {
	return r.Min() == r.Max()
}

func (r Dice) ShortString() string {
	if r.isFixed() {
		flatValue := r.Min()
		return fmt.Sprintf("%d", flatValue)
	}
	if r.bonus == 0 {
		return fmt.Sprintf("%dd%d", r.dice, r.sides)
	}
	return fmt.Sprintf("%dd%d%+d", r.dice, r.sides, r.bonus)
}
func (r Dice) Roll() int {
	if r.isFixed() {
		flatValue := r.Min()
		return flatValue
	}
	if r.sides == 0 {
		return r.bonus
	}
	result := r.bonus
	for i := 0; i < r.dice; i++ {
		result += 1 + rand.Intn(r.sides)
	}
	return result
}

func (r Dice) RollWithBonus(bonus int) int {
	return r.Roll() + bonus
}

func (r Dice) DiceCount() int {
	return r.dice
}

func (r Dice) Avg() int {
	return r.Min() + (r.Max()-r.Min())/2
}

func (r Dice) RollPyramid() int {
	return int((float64(r.Roll()) + float64(r.Roll())) * 0.5)
}

func (r Dice) WithBonus(plus int) Dice {
	return NewDice(r.dice, r.sides, r.bonus+plus)
}

func (r Dice) NotZero() bool {
	return r.dice != 0 || r.sides != 0 || r.bonus != 0
}

func (r Dice) ExpectedValue() int {
	if r.isFixed() {
		return r.Min()
	}
	// erwartungswert
	totalExpectedValue := float64(r.dice)*r.expectedValueForOneDice() + float64(r.bonus)
	return int(totalExpectedValue)
}

func (r Dice) expectedValueForOneDice() float64 {
	// erwartungswert
	sum := 1.0
	p := 1.0 / float64(r.sides)
	for i := 1; i < r.sides; i++ {
		sum += float64(i) * p
	}
	return sum
}

func (r Dice) WithAddedDice(bonus int) Dice {
	return NewDice(r.dice+bonus, r.sides, r.bonus)
}

func ParseDice(dice string) Dice {
	// check if it's an interval, eg. 12-24
	dice = strings.ToLower(strings.Replace(dice, " ", "", -1))

	intervalPattern := regexp.MustCompile(`^(\d+)-(\d+)$`)

	flatNumberPattern := regexp.MustCompile(`^(-?\d+)$`)

	hasDiceNotation := strings.ContainsAny(dice, "d")

	isInterval := intervalPattern.MatchString(dice)

	isFlatNumber := flatNumberPattern.MatchString(dice)

	if !hasDiceNotation && !isInterval && isFlatNumber {
		match := flatNumberPattern.FindStringSubmatch(dice)
		if match == nil {
			return Dice{0, 0, 0}
		}
		numVal, _ := strconv.Atoi(match[1])
		return NewDice(0, 0, numVal)
	}

	if !hasDiceNotation && !isFlatNumber && isInterval {
		match := intervalPattern.FindStringSubmatch(dice)
		if match == nil {
			return Dice{0, 0, 0}
		}
		minVal, _ := strconv.Atoi(match[1])
		maxVal, _ := strconv.Atoi(match[2])
		dieRange := (maxVal - minVal) + 1
		dieBonus := minVal - 1
		return NewDice(1, dieRange, dieBonus)
	}

	if !hasDiceNotation {
		return Dice{0, 0, 0}
	}

	// dice is a string in the format "NdM" where N is the number of dice to roll and M is the number of sides on each die
	if strings.HasPrefix(dice, "d") {
		dice = "1" + dice
	}
	if !strings.ContainsAny(dice, "-+") {
		dice = dice + "+0"
	}
	pattern := regexp.MustCompile(`(\d+)d(\d+)\s*[+\-]\s*(\d+)`)
	match := pattern.FindStringSubmatch(dice)
	if match == nil {
		return Dice{0, 0, 0}
	}
	sign := 1
	if strings.Contains(dice, "-") {
		sign = -1
	}
	numDice, _ := strconv.Atoi(match[1])
	numSides, _ := strconv.Atoi(match[2])
	bonus, _ := strconv.Atoi(match[3])
	return NewDice(numDice, numSides, bonus*sign)
}

func InSix(count int) bool {
	return rand.Intn(6)+1 <= count
}

func Spread(value int, spread float64) int {
	fluctuation := float64(value) * spread
	interval := int(fluctuation * 2)
	if interval == 0 {
		return value
	}
	return value + rand.Intn(interval) - int(fluctuation)
}
