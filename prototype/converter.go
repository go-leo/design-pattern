package prototype

import (
	"reflect"
)

type BoolConverter interface {
	ToInt(b bool) (int64, error)
	ToUint(b bool) (uint64, error)
	ToFloat(b bool) (float64, error)
	ToString(b bool) (string, error)
	ToStruct(b bool, val reflect.Value) error
	ToMap(b bool, val reflect.Value) error
	ToSlice(b bool, val reflect.Value) error
	ToArray(b bool, val reflect.Value) error
	ToPointer(b bool, val reflect.Value) error
	ToInterface(b bool, val reflect.Value) error
}

type DefaultBoolConverter struct {
}

func (DefaultBoolConverter) BoolToInt(b bool) int64 {
	if !b {
		return 0
	}
	return 1
}

func (DefaultBoolConverter) BoolToUint(b bool) uint64 {
	if !b {
		return 0
	}
	return 1
}
