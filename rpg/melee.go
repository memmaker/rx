package rpg

type MeleeDamage struct {
	StrengthThreshold int
	ThrustDice        Dice
	SwingDice         Dice
}

var basicDamageTable = []MeleeDamage{
	{1, ParseDice("1d6-6"), ParseDice("1d6-5")},
	{2, ParseDice("1d6-6"), ParseDice("1d6-5")},
	{3, ParseDice("1d6-5"), ParseDice("1d6-4")},
	{4, ParseDice("1d6-5"), ParseDice("1d6-4")},
	{5, ParseDice("1d6-4"), ParseDice("1d6-3")},
	{6, ParseDice("1d6-4"), ParseDice("1d6-3")},
	{7, ParseDice("1d6-3"), ParseDice("1d6-2")},
	{8, ParseDice("1d6-3"), ParseDice("1d6-2")},
	{9, ParseDice("1d6-2"), ParseDice("1d6-1")},
	{10, ParseDice("1d6-2"), ParseDice("1d6")},
	{11, ParseDice("1d6-1"), ParseDice("1d6+1")},
	{12, ParseDice("1d6-1"), ParseDice("1d6+2")},
	{13, ParseDice("1d6"), ParseDice("2d6-1")},
	{14, ParseDice("1d6"), ParseDice("2d6")},
	{15, ParseDice("1d6+1"), ParseDice("2d6+1")},
	{16, ParseDice("1d6+1"), ParseDice("2d6+2")},
	{17, ParseDice("1d6+2"), ParseDice("3d6-1")},
	{18, ParseDice("1d6+2"), ParseDice("3d6")},
	{19, ParseDice("2d6-1"), ParseDice("3d6+1")},
	{20, ParseDice("2d6-1"), ParseDice("3d6+2")},
	{21, ParseDice("2d6"), ParseDice("4d6-1")},
	{22, ParseDice("2d6"), ParseDice("4d6")},
	{23, ParseDice("2d6+1"), ParseDice("4d6+1")},
	{24, ParseDice("2d6+1"), ParseDice("4d6+2")},
	{25, ParseDice("2d6+2"), ParseDice("5d6-1")},
	{26, ParseDice("2d6+2"), ParseDice("5d6")},
	{27, ParseDice("3d6-1"), ParseDice("5d6+1")},
	{28, ParseDice("3d6-1"), ParseDice("5d6+1")},
	{29, ParseDice("3d6"), ParseDice("5d6+2")},
	{30, ParseDice("3d6"), ParseDice("5d6+2")},
	{31, ParseDice("3d6+1"), ParseDice("6d6-1")},
	{32, ParseDice("3d6+1"), ParseDice("6d6-1")},
	{33, ParseDice("3d6+2"), ParseDice("6d6")},
	{34, ParseDice("3d6+2"), ParseDice("6d6")},
	{35, ParseDice("4d6-1"), ParseDice("6d6+1")},
	{36, ParseDice("4d6-1"), ParseDice("6d6+1")},
	{37, ParseDice("4d6"), ParseDice("6d6+2")},
	{38, ParseDice("4d6"), ParseDice("6d6+2")},
	{39, ParseDice("4d6+1"), ParseDice("7d6-1")},
	{40, ParseDice("4d6+1"), ParseDice("7d6-1")},
	{45, ParseDice("5d6"), ParseDice("7d6+1")},
	{50, ParseDice("5d6+2"), ParseDice("8d6-1")},
	{55, ParseDice("6d6"), ParseDice("8d6+1")},
	{60, ParseDice("7d6-1"), ParseDice("9d6")},
	{65, ParseDice("7d6+1"), ParseDice("9d6+2")},
	{70, ParseDice("8d6"), ParseDice("10d6")},
	{75, ParseDice("8d6+2"), ParseDice("10d6+2")},
	{80, ParseDice("9d6"), ParseDice("11d6")},
	{85, ParseDice("9d6+2"), ParseDice("11d6+2")},
	{90, ParseDice("10d6"), ParseDice("12d6")},
	{95, ParseDice("10d6+2"), ParseDice("12d6+2")},
	{100, ParseDice("11d6"), ParseDice("13d6")},
}

func GetBasicMeleeDamageFromStrength(strength int) MeleeDamage {
	for _, damage := range basicDamageTable {
		if strength <= damage.StrengthThreshold {
			return damage
		}
	}
	return basicDamageTable[len(basicDamageTable)-1]
}
