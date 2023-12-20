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
	SourceTagKey string
	TargetTagKey string
	DeepClone    bool
	NameComparer func(t, s string) bool
	UnixTime     func(t time.Time) int64
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
	if o.UnixTime == nil {
		o.UnixTime = func(t time.Time) int64 { return t.Unix() }
	}
	return o
}

type Option func(o *options)
