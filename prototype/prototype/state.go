package prototype

import "sync"

const startDetectingCyclesAfter = 1000

type cloneState struct {
	// Keep track of what pointers we've seen in the current recursive call
	// path, to avoid cycles that could lead to a stack overflow. Only do
	// the relatively expensive map operations if ptrLevel is larger than
	// startDetectingCyclesAfter, so that we skip the work if we're within a
	// reasonable amount of nested pointers deep.
	ptrLevel uint
	ptrSeen  map[any]struct{}
}

func (e *cloneState) forward() {
	e.ptrLevel++
}

func (e *cloneState) back() {
	e.ptrLevel--
}

func (e *cloneState) isTooDeep() bool {
	return e.ptrLevel > startDetectingCyclesAfter
}

func (e *cloneState) isSeen(ptr any) bool {
	_, ok := e.ptrSeen[ptr]
	return ok
}

func (e *cloneState) remember(ptr any) {
	e.ptrSeen[ptr] = struct{}{}
}

func (e *cloneState) forget(ptr any) {
	delete(e.ptrSeen, ptr)
}

var cloneStatePool sync.Pool

func newCloneState() *cloneState {
	if v := cloneStatePool.Get(); v != nil {
		e := v.(*cloneState)
		if len(e.ptrSeen) > 0 {
			panic("ptrEncoder.encode should have emptied ptrSeen via defers")
		}
		e.ptrLevel = 0
		return e
	}
	return &cloneState{ptrSeen: make(map[any]struct{})}
}
