package game

import (
	"RogueUI/foundation"
	"RogueUI/special"
)

func CanPerceive(observer *Actor, observed *Actor) bool {
	return true
	/*
	   distance := geometry.Distance(observer.Position(), observed.Position())
	   observedChar := observed.GetCharSheet()
	   observerChar := observer.GetCharSheet()

	   sneakSkill := observedChar.GetSkill(special.Sneak)
	   isObserverSneaking := observedChar.HasFlag(special.FlagSneaking)

	*/

}

func CheckStrength(actor *Actor) foundation.CheckResult {
	critChance := actor.GetCharSheet().GetStat(special.Luck)
	strengthSkill := actor.GetCharSheet().GetStat(special.Strength)
	return special.SuccessRoll(special.Percentage(strengthSkill*10), special.Percentage(critChance))
}

func CheckPerception(actor *Actor) foundation.CheckResult {
	critChance := actor.GetCharSheet().GetStat(special.Luck)
	perceptionSkill := actor.GetCharSheet().GetStat(special.Perception)
	return special.SuccessRoll(special.Percentage(perceptionSkill*10), special.Percentage(critChance))
}
