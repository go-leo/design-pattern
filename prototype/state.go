package prototype

import (
	"reflect"
	"sync"
)

const startDetectingCyclesAfter = 1000

// stackOverflowGuard 记录递归调用路径中的指针，避免可能导致的堆栈溢出
type stackOverflowGuard struct {
	ptrLevel uint
	ptrSeen  map[any]struct{}
}

func (g *stackOverflowGuard) forward() {
	g.ptrLevel++
}

func (g *stackOverflowGuard) back() {
	g.ptrLevel--
}

func (g *stackOverflowGuard) isTooDeep() bool {
	return g.ptrLevel > startDetectingCyclesAfter
}

func (g *stackOverflowGuard) isSeen(ptr any) bool {
	_, ok := g.ptrSeen[ptr]
	return ok
}

func (g *stackOverflowGuard) remember(ptr any) {
	g.ptrSeen[ptr] = struct{}{}
}

func (g *stackOverflowGuard) forget(ptr any) {
	delete(g.ptrSeen, ptr)
}

func (g *stackOverflowGuard) checkPointerCycle(ptrFunc func(srcVal reflect.Value) any, cloner clonerFunc) clonerFunc {
	return func(g *stackOverflowGuard, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
		if g.forward(); g.isTooDeep() {
			ptr := ptrFunc(srcVal)
			if g.isSeen(ptr) {
				return newPointerCycleError(labels, srcVal.Type())
			}
			g.remember(ptr)
			defer g.forget(ptr)
		}
		defer g.back()
		return cloner(g, labels, tgtVal, srcVal, opts)
	}
}

var cloneContextPool sync.Pool

func newCloneContext() *stackOverflowGuard {
	if v := cloneContextPool.Get(); v != nil {
		g := v.(*stackOverflowGuard)
		if len(g.ptrSeen) > 0 {
			panic("pointerCloner.encode should have emptied ptrSeen via defers")
		}
		g.ptrLevel = 0
		return g
	}
	return &stackOverflowGuard{ptrSeen: make(map[any]struct{})}
}

func freeCloneContext(ctx *stackOverflowGuard) {
	cloneContextPool.Put(ctx)
}
