package gridmap

type ItemID uint64
type ActorID uint64

var currentItemID ItemID
var currentActorID ActorID

func NextItemID() ItemID {
	next := currentItemID
	currentItemID++
	return next
}

func NextActorID() ActorID {
	next := currentActorID
	currentActorID++
	return next
}

func CurrentItemID() ItemID {
	return currentItemID
}

func CurrentActorID() ActorID {
	return currentActorID
}

func SetCurrentItemID(id ItemID) {
	currentItemID = id
}

func SetCurrentActorID(id ActorID) {
	currentActorID = id
}
