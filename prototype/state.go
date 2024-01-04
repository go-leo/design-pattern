package prototype

import (
	"reflect"
	"sync"
)

const startDetectingCyclesAfter = 1000

// cloneContext 记录递归调用路径中的指针，避免可能导致的堆栈溢出
// 记录克隆中的错误
type cloneContext struct {
	ptrLevel uint
	ptrSeen  map[any]struct{}
	options  *options
	errors   []error
}

func (g *cloneContext) forward() {
	g.ptrLevel++
}

func (g *cloneContext) back() {
	g.ptrLevel--
}

func (g *cloneContext) isTooDeep() bool {
	return g.ptrLevel > startDetectingCyclesAfter
}

func (g *cloneContext) isSeen(ptr any) bool {
	_, ok := g.ptrSeen[ptr]
	return ok
}

func (g *cloneContext) remember(ptr any) {
	g.ptrSeen[ptr] = struct{}{}
}

func (g *cloneContext) forget(ptr any) {
	delete(g.ptrSeen, ptr)
}

func (g *cloneContext) checkPointerCycle(ptrFunc func(srcVal reflect.Value) any, cloner clonerFunc) clonerFunc {
	return func(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
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

func (g *cloneContext) error(err error) error {
	if err == nil {
		return nil
	}
	if err != nil && g.options.InterruptOnError {
		return err
	}
	g.errors = append(g.errors, err)
	return nil
}

var cloneContextPool sync.Pool

func newCloneContext(o *options) *cloneContext {
	if v := cloneContextPool.Get(); v != nil {
		g := v.(*cloneContext)
		if len(g.ptrSeen) > 0 {
			panic("cloner should have emptied ptrSeen via defers")
		}
		g.ptrLevel = 0
		g.options = o
		return g
	}
	return &cloneContext{ptrSeen: make(map[any]struct{})}
}

func freeCloneContext(ctx *cloneContext) {
	cloneContextPool.Put(ctx)
}
