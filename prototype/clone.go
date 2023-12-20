package prototype

import (
	"reflect"
)

// Clone 将值从 src 克隆到 tgt
func Clone(tgt any, src any, opts ...Option) error {
	// 获取目标值的反射值
	tgtVal := reflect.ValueOf(tgt)
	// 如果目标不是一个指针或者为nil，返回错误
	if tgtVal.Kind() != reflect.Pointer {
		return newNonPointerError(reflect.TypeOf(tgt))
	}
	if tgtVal.IsNil() {
		return newNilError(reflect.TypeOf(tgt))
	}

	// 获取原值的反射值
	srcVal := reflect.ValueOf(src)

	// 从对象池中获取克隆状态上下文
	ctx := newCloneContext()
	// 克隆结束后，将克隆状态上下文放入对象池中
	defer freeCloneContext(ctx)

	// 初始化options
	o := new(options).apply(opts...).correct()

	// 处理对象克隆逻辑
	return clone(ctx, tgtVal, srcVal, o)
}

func clone(e *cloneContext, tgtVal, srcVal reflect.Value, opts *options) error {
	return clonerByValue(srcVal, opts)(e, []string{}, tgtVal, srcVal, opts)
}
