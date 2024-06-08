package foundation

type BodyPartFunction int

const (
	BodyPartFunctionNone BodyPartFunction = iota
	BodyPartFunctionMovement
	BodyPartFunctionManipulation
	BodyPartFunctionSensory
	BodyPartFunctionPainCenter
	BodyPartFunctionCognitive
	BodyPartFunctionVital
)

type BodyPart struct {
	Name         string
	SizeModifier int
	HitPoints    int
	HitPointsMax int
	Function     BodyPartFunction
}

func NewBodyPart(name string, sizeModifier, hitPointsMax int, function BodyPartFunction) *BodyPart {
	return &BodyPart{Name: name, SizeModifier: sizeModifier, HitPoints: hitPointsMax, HitPointsMax: hitPointsMax, Function: function}
}

func BodyByName(bodyName string, totalHitPoints int) []*BodyPart {
	oneTwen := float64(totalHitPoints) / 20.0
	switch bodyName {
	default:
		return []*BodyPart{
			NewBodyPart("Head", -2, int(2*oneTwen), BodyPartFunctionCognitive),
			NewBodyPart("Eyes", -3, int(1*oneTwen), BodyPartFunctionSensory),
			NewBodyPart("Torso", 0, int(4*oneTwen), BodyPartFunctionVital),
			NewBodyPart("Left Arm", -1, int(3*oneTwen), BodyPartFunctionManipulation),
			NewBodyPart("Right Arm", -1, int(3*oneTwen), BodyPartFunctionManipulation),
			NewBodyPart("Left Leg", -1, int(3*oneTwen), BodyPartFunctionMovement),
			NewBodyPart("Right Leg", -1, int(3*oneTwen), BodyPartFunctionMovement),
			NewBodyPart("Groin", -2, int(1*oneTwen), BodyPartFunctionPainCenter),
		}
	}
}
