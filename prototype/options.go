package prototype

import (
	"reflect"
	"strings"
	"time"
)

// Hook is a hook.
type Hook func(fullKeys []string, tgtVal reflect.Value, srcVal reflect.Value) error

type options struct {
	ValueHook    map[reflect.Value]map[reflect.Value]Hook
	TypeHooks    map[reflect.Type]map[reflect.Type]Hook
	KindHooks    map[reflect.Kind]map[reflect.Kind]Hook
	TagKey       string
	DeepClone    bool
	NameComparer func(t, s string) bool
	TimeToInt    func(t time.Time) int64
	IntToTime    func(i int64) time.Time
	StringToTime func(s string) time.Time
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
	if o.TimeToInt == nil {
		o.TimeToInt = func(t time.Time) int64 { return t.Unix() }
	}
	if o.IntToTime == nil {
		o.IntToTime = func(i int64) time.Time { return time.Unix(i, 0) }
	}
	return o
}

type Option func(o *options)

func TagKey(key string) Option {
	return func(o *options) {
		o.TagKey = key
	}
}
