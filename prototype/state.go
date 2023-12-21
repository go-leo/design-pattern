package prototype

import (
	"reflect"
	"sync"
)

const startDetectingCyclesAfter = 1000

type cloneContext struct {
	// Keep track of what pointers we've seen in the current recursive call
	// path, to avoid cycles that could lead to a stack overflow. Only do
	// the relatively expensive map operations if ptrLevel is larger than
	// startDetectingCyclesAfter, so that we skip the work if we're within a
	// reasonable amount of nested pointers deep.
	ptrLevel uint
	ptrSeen  map[any]struct{}
}

func (e *cloneContext) forward() {
	e.ptrLevel++
}

func (e *cloneContext) back() {
	e.ptrLevel--
}

func (e *cloneContext) isTooDeep() bool {
	return e.ptrLevel > startDetectingCyclesAfter
}

func (e *cloneContext) isSeen(ptr any) bool {
	_, ok := e.ptrSeen[ptr]
	return ok
}

func (e *cloneContext) remember(ptr any) {
	e.ptrSeen[ptr] = struct{}{}
}

func (e *cloneContext) forget(ptr any) {
	delete(e.ptrSeen, ptr)
}

func (e *cloneContext) checkPointerCycle(ptrFunc func(srcVal reflect.Value) any, cloner clonerFunc) clonerFunc {
	return func(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
		if e.forward(); e.isTooDeep() {
			ptr := ptrFunc(srcVal)
			if e.isSeen(ptr) {
				return newPointerCycleError(fks, srcVal.Type())
			}
			e.remember(ptr)
			defer e.forget(ptr)
		}
		defer e.back()
		return cloner(e, fks, tgtVal, srcVal, opts)
	}
}

var cloneContextPool sync.Pool

func newCloneContext() *cloneContext {
	if v := cloneContextPool.Get(); v != nil {
		e := v.(*cloneContext)
		if len(e.ptrSeen) > 0 {
			panic("pointerCloner.encode should have emptied ptrSeen via defers")
		}
		e.ptrLevel = 0
		return e
	}
	return &cloneContext{ptrSeen: make(map[any]struct{})}
}

func freeCloneContext(ctx *cloneContext) {
	cloneContextPool.Put(ctx)
}
