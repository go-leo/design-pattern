package prototype

import (
	"errors"
	"reflect"
)

type InvalidToArgError struct {
	Type reflect.Type
}

func (e *InvalidToArgError) Error() string {
	if e.Type == nil {
		return "prototype: Clone(nil, from)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "prototype: Clone(non-pointer " + e.Type.String() + ", from)"
	}
	return "prototype: Clone(nil " + e.Type.String() + ", from)"
}

type InvalidFromArgError struct {
	Type reflect.Type
}

func (e *InvalidFromArgError) Error() string {
	if e.Type == nil {
		return "prototype: Clone(from, nil)"
	}
	return "prototype: Clone(to, invalid" + e.Type.String() + ")"
}

var (
	ErrInvalidCopyDestination        = errors.New("copy destination must be non-nil and addressable")
	ErrInvalidCopyFrom               = errors.New("copy from must be non-nil and addressable")
	ErrMapKeyNotMatch                = errors.New("map's key type doesn't match")
	ErrNotSupported                  = errors.New("not supported")
	ErrFieldNameTagStartNotUpperCase = errors.New("cloneAny field name tag must be start upper case")
)
