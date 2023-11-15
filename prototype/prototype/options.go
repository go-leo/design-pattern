package prototype

import "reflect"

// Hook is a hook.
type Hook func(fullKeys []string, tgtVal reflect.Value, srcVal reflect.Value) error

type options struct {
	ValueHook    map[reflect.Value]map[reflect.Value]Hook
	TypeHooks    map[reflect.Type]map[reflect.Type]Hook
	KindHooks    map[reflect.Kind]map[reflect.Kind]Hook
	SourceTagKey string
	TargetTagKey string
	DeepClone    bool
	EqualFold    func(tgtKey, srcKey string) bool
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o *options) correct() *options {
	return o
}

type Option func(o *options)
