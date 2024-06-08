import os

nameMap = {
	"MAANTT": "ant",
	"MABRAN": "king_rat",
	"MACLAW": "adult_deathclaw",
	"MACLW2": "baby_deathclaw",
	"MACYBR": "robo_dog",
	"MADDOG": "dog",
	"MADEGG": "egg",
	"MADETH": "grey_deathclaw",
	"MAFEYE": "eye_robot",
	"MAFIRE": "fire_gecko",
	"MAGCKO": "golden_gecko",
	"MAGUN2": "plasma_turret",
	"MAGUNN": "mini_turret",
	"MAHAND": "mr_handy",
	"MALIEN": "alien_wanamingo",
	"MAMANT": "mantis",
	"MAMRAT": "mole_rat",
	"MAMTN2": "super_mutant_leather",
	"MAMTNT": "super_mutant",
	"MAMURT": "pig_rat",
	"MAPLNT": "mutant_plant",
	"MAQUEN": "queen_wanamingo",
	"MAROBE": "goris",
	"MAROBO": "brain_bot",
	"MAROBT": "assault_bot",
	"MASC2": "baby_radscorpion",
	"MASCRP": "radscorpion",
	"MASPHN": "floater",
	"MASRAT": "rat",
	"MATHNG": "centaur"
}

codeMap = {
	"AA": "Idle",
	"AN": "Dodge",
	"AO": "Hit",
	"AQ": "Attack",
	"LK": "Attack",
	"KL": "Attack",
	"JK": "Attack",
	"MJ": "Attack",
	"IK": "Attack",
	"HJ": "Attack",
	"GM": "Attack",
	"FG": "Attack",
	"DM": "Attack",
	"BA": "Falling",
	"BB": "Falling",
	"BC": "Falling",
	"BD": "HoleInBody",
	"BE": "Burned",
	"BK": "Burned",
	"BI": "SlicedInTwo",
	"BL": "Exploded",
	"BG": "Perforated",
	"BF": "RippedApart",
	"BM": "Meltdown",
	"BH": "Electrocuted",
	"BO": "Bleeding",
	"BP": "Bleeding",
	"CJ": "GetUp",
	"YA": "Death",
	"ZA": "Death",
	"ZB": "Death",
	"ZR": "Death",
	"ZQ": "Death",
}

def ensureDirExists(dirname):
	print("Ensuring directory exists:", dirname)
	if not os.path.exists(dirname):
		os.makedirs(dirname)

def moveFileToDir(oldName, newName):
	print("Moving", oldName, "to", newName)
	os.rename(oldName, newName)

def main(dirname):
	files = os.listdir(dirname)
	for file in files:
		if file.endswith(".ogg"):
			withoutExtension = file.split(".")[0]
			namePart = withoutExtension[0:6]
			if namePart in nameMap:
				realName = nameMap[namePart]
				newDir = os.path.join(dirname, realName)
				ensureDirExists(newDir)
				cueCode = withoutExtension[6:8]
				if cueCode in codeMap:
					cueName = codeMap[cueCode]
					subDir = os.path.join(newDir, cueName)
					ensureDirExists(subDir)
					moveFileToDir(os.path.join(dirname, file), os.path.join(subDir, file))
				else:
					print("Cue code not found", cueCode)
if __name__ == "__main__":
	currentWorkingDirectory = "/Users/felix/Documents/Fallout2_Export/sound/SFX"
	main(currentWorkingDirectory)