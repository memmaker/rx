package gridmap

type ItemID uint64
type ActorID uint64

var nextItemID ItemID
var nextActorID ActorID

func NextItemID() ItemID {
	next := nextItemID
	nextItemID++
	return next
}

func NextActorID() ActorID {
	next := nextActorID
	nextActorID++
	return next
}

func PeekAtNextItemID() ItemID {
	return nextItemID
}

func PeekAtNextActorID() ActorID {
	return nextActorID
}

func SetNextItemID(id ItemID) {
	nextItemID = id
}

func SetNextActorID(id ActorID) {
	nextActorID = id
}
