package prototype

import (
	"reflect"
	"sync"
)

// Clone 将值从src克隆到tgt
func Clone(tgt any, src any, opts ...Option) error {
	// 获取目标值的反射值
	tgtVal := reflect.ValueOf(tgt)
	// 如果目标不是一个指针或者为nil，返回InvalidTargetError错误
	if tgtVal.Kind() != reflect.Pointer || tgtVal.IsNil() {
		return &InvalidTargetError{Type: reflect.TypeOf(tgt)}
	}

	// 获取原值的反射值
	srcVal := reflect.ValueOf(src)

	// 从对象池中获取克隆状态上下文
	ctx := newCloneContext()
	// 克隆结束后，将克隆状态上下文放入对象池中
	defer cloneContextPool.Put(ctx)

	// 初始化options
	o := new(options).apply(opts...).correct()

	// 处理对象克隆逻辑
	return clone(ctx, tgtVal, srcVal, o)
}

func clone(e *cloneContext, tgtVal, srcVal reflect.Value, opts *options) error {
	return valueCloner(srcVal, opts)(e, []string{}, tgtVal, srcVal, opts)
}

var clonerCache sync.Map

func valueCloner(srvVal reflect.Value, opts *options) ClonerFunc {
	if !srvVal.IsValid() {
		return emptyValueCloner
	}
	return typeCloner(srvVal.Type(), opts)
}

func typeCloner(srcType reflect.Type, opts *options) ClonerFunc {
	if fi, ok := clonerCache.Load(srcType); ok {
		return fi.(ClonerFunc)
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it. This indirect
	// func is only used for recursive types.
	var (
		wg sync.WaitGroup
		f  ClonerFunc
	)
	wg.Add(1)
	fi, loaded := clonerCache.LoadOrStore(srcType, ClonerFunc(func(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
		wg.Wait()
		return f(e, fks, tgtVal, srcVal, opts)
	}))
	if loaded {
		return fi.(ClonerFunc)
	}

	// Compute the real clonerFunc and replace the indirect func with it.
	f = newTypeCloner(srcType)
	wg.Done()
	clonerCache.Store(srcType, f)
	return f
}

// newTypeCloner constructs an ClonerFunc for a type.
func newTypeCloner(srcType reflect.Type) ClonerFunc {
	switch srcType.Kind() {
	case reflect.Bool:
		return boolCloner
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intCloner
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintCloner
	case reflect.Float32:
		return float32Cloner
	case reflect.Float64:
		return float64Cloner
	case reflect.String:
		return stringCloner
	case reflect.Interface:
		return interfaceCloner
	case reflect.Struct:
		return structCloner
	case reflect.Map:
		return mapCloner
	case reflect.Slice:
		return sliceCloner
	case reflect.Array:
		return arrayCloner
	case reflect.Pointer:
		return ptrCloner
	default:
		return unsupportedTypeCloner
	}
}
