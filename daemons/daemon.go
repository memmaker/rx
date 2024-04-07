package daemons

import "unsafe"

// Adapted from
// https://github.com/memmaker/rogue-pc-modern-C/blob/main/src/daemon.c

const DAEMON = -1

type DelayedAction struct {
	Call func()
	Time int
}

var daemons []*DelayedAction

// Start should only ever be called with a named function.
func Start(callback func()) {
	daemons = append(daemons, &DelayedAction{Call: callback, Time: DAEMON})
}

// Fuse should only ever be called with a named function.
func Fuse(callback func(), time int) {
	daemons = append(daemons, &DelayedAction{Call: callback, Time: time})
}

// Lengthen should only ever be called with a named function
func Lengthen(callback func(), time int) {
	for i := range daemons {
		daemon := daemons[i]
		if functionPointersAreEqual(daemon.Call, callback) {
			daemon.Time = time
		}
	}
}
func UpdateDeamons() {
	for i := range daemons {
		daemon := daemons[i]
		if daemon.Time == DAEMON && daemon.Call != nil {
			daemon.Call()
		}
	}
}

func UpdateFuses() {
	for i := len(daemons) - 1; i >= 0; i-- {
		daemon := daemons[i]
		if daemon.Time > 0 {
			daemon.Time--
		}
		if daemon.Time == 0 {
			daemon.Call()
			daemons = append(daemons[:i], daemons[i+1:]...)
		}
	}
}

// Extinguish should only ever be called with a named function
func Extinguish(callback func()) {
	for i := len(daemons) - 1; i >= 0; i-- {
		if functionPointersAreEqual(daemons[i].Call, callback) {
			daemons = append(daemons[:i], daemons[i+1:]...)
		}
	}
}
func ClearAll() {
	daemons = nil
}

// functionPointersAreEqual IMPORTANT: This will only work with named functions and not with anonymous functions.
func functionPointersAreEqual(x, y func()) bool {
	px := *(*unsafe.Pointer)(unsafe.Pointer(&x))
	py := *(*unsafe.Pointer)(unsafe.Pointer(&y))
	return px == py
}
