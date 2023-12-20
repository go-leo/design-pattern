package prototype

import (
	"fmt"
	"reflect"
	"strings"
)

type Code int

const (
	NonPointer     Code = 1
	Nil                 = 2
	Overflow            = 3
	NegativeNumber      = 4
	StringParse         = 5
)

type Error struct {
	Code       Code
	FullKeys   []string
	TargetType reflect.Type
	Value      string
	err        error
}

func (e Error) Error() string {
	keys := strings.Join(e.FullKeys, ".")
	switch e.Code {
	case NonPointer:
		return fmt.Sprintf("prototype: non-pointer error, target type is %s", e.TargetType.String())
	case Nil:
		if e.TargetType == nil {
			return "prototype: nil error, target is nil"
		}
		return fmt.Sprintf("prototype: nil error, target is (%s)(nil)", e.TargetType.String())
	case Overflow:
		return fmt.Sprintf("prototype: overflow error, %s, cannot Clone %s into Go value of type %s", keys, e.Value, e.TargetType.String())
	case NegativeNumber:
		return fmt.Sprintf("prototype: negative number error, %s cannot Clone %s into Go value of type %s", keys, e.Value, e.TargetType.String())
	case StringParse:
		return fmt.Sprintf("prototype: parse string error, %s cannot Clone %s into Go value of type %s, %v", keys, e.Value, e.TargetType.String(), e.err)
	default:
		return ""
	}
}

func (e Error) Unwrap() error {
	return e.err
}

func newNonPointerError(tgtType reflect.Type) error {
	return Error{Code: Overflow, TargetType: tgtType}
}

func newNilError(tgtType reflect.Type) error {
	return Error{Code: Nil, TargetType: tgtType}
}

func newOverflowError(fks []string, tgtType reflect.Type, value string) error {
	return Error{Code: Overflow, FullKeys: fks, TargetType: tgtType, Value: value}
}

func newNegativeNumberError(fks []string, tgtType reflect.Type, value string) error {
	return Error{Code: NegativeNumber, FullKeys: fks, TargetType: tgtType, Value: value}
}

func newStringParseError(fks []string, tgtType reflect.Type, value string, err error) error {
	return Error{Code: StringParse, FullKeys: fks, TargetType: tgtType, Value: value, err: err}
}

type UnsupportedTypeError struct {
	SourceType reflect.Type
	TargetType reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "prototype: unsupported type: " + e.SourceType.String() + " to " + e.TargetType.String()
}

// An UnsupportedValueError is returned by Clone when attempting
// to encode an unsupported value.
type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "prototype: unsupported value: " + e.Str
}

// An CloneError describes a JSON value that was
// not appropriate for a value of a specific Go type.
type CloneError struct {
	TargetValue reflect.Value
	SourceValue reflect.Value

	Value string       // description of JSON value - "bool", "array", "number -5"
	Type  reflect.Type // type of Go value it could not be assigned to
}

func (e *CloneError) Error() string {
	return "prototype: cannot Clone " + e.Value + " into Go value of type " + e.Type.String()
}

type InvalidTargetError struct {
	Type reflect.Type
}

func (e *InvalidTargetError) Error() string {
	if e.Type == nil {
		return "prototype: Clone(nil, src)"
	}
	if e.Type.Kind() != reflect.Pointer {
		return "prototype: Clone(non-pointer " + e.Type.String() + ", src)"
	}
	return "prototype: Clone(nil " + e.Type.String() + ", src)"
}
