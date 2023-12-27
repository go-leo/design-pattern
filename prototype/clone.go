package prototype

import (
	"reflect"
	"strings"
	"time"
)

// Hook is a hook.
type Hook func(labels []string, tgtVal reflect.Value, srcVal reflect.Value) error

type options struct {
	ValueHook    map[reflect.Value]map[reflect.Value]Hook
	TypeHooks    map[reflect.Type]map[reflect.Type]Hook
	KindHooks    map[reflect.Kind]map[reflect.Kind]Hook
	TagKey       string
	DeepClone    bool
	NameComparer func(t, s string) bool
	IntToTime    func(i int64) time.Time
	StringToTime func(s string) (time.Time, error)
	TimeToInt    func(t time.Time) int64
	TimeToString func(t time.Time) string
	GetterPrefix string
	SetterPrefix string
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o *options) correct() *options {
	if o.NameComparer == nil {
		o.NameComparer = strings.EqualFold
	}
	if o.IntToTime == nil {
		o.IntToTime = func(i int64) time.Time { return time.Unix(i, 0) }
	}
	if o.StringToTime == nil {
		o.StringToTime = func(s string) (time.Time, error) { return time.ParseInLocation(time.DateTime, s, time.Local) }
	}
	if o.TimeToInt == nil {
		o.TimeToInt = func(t time.Time) int64 { return t.Unix() }
	}
	if o.TimeToString == nil {
		o.TimeToString = func(t time.Time) string { return t.Local().Format(time.DateTime) }
	}
	return o
}

type Option func(o *options)

func TagKey(key string) Option {
	return func(o *options) {
		o.TagKey = key
	}
}

func GetterPrefix(p string) Option {
	return func(o *options) {
		o.GetterPrefix = p
	}
}

func SetterPrefix(p string) Option {
	return func(o *options) {
		o.SetterPrefix = p
	}
}

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
	return valueCloner(srcVal, opts)(e, []string{}, tgtVal, srcVal, opts)
}
