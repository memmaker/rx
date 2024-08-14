package game

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
