package prototype

import (
	"fmt"
	"reflect"
	"strings"
)

type Code int

const (
	NonPointer           Code = 1
	Nil                       = 2
	Overflow                  = 3
	NegativeNumber            = 4
	FailedParse               = 5
	PointerCycle              = 6
	UnsupportedType           = 7
	FailedConvertScanner      = 8
	FailedUnmarshalNew        = 9
)

type Error struct {
	Code       Code
	FullKeys   []string
	TargetType reflect.Type
	SourceType reflect.Type
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
		return fmt.Sprintf("prototype: overflow error, %s, cannot Clone %s into target type %s", keys, e.Value, e.TargetType.String())
	case NegativeNumber:
		return fmt.Sprintf("prototype: negative number error, %s, cannot Clone %s into target type %s", keys, e.Value, e.TargetType.String())
	case FailedParse:
		return fmt.Sprintf("prototype: failed to parse string error, %s, cannot Clone %s into target type %s, %v", keys, e.Value, e.TargetType.String(), e.err)
	case PointerCycle:
		return fmt.Sprintf("prototype: pointer cycle error, encountered a cycle via %s", e.SourceType.String())
	case UnsupportedType:
		return fmt.Sprintf("prototype: unsupported type error, %s, cannot Clone source type %s into  target type %s", keys, e.SourceType.String(), e.TargetType.String())
	case FailedConvertScanner:
		return fmt.Sprintf("prototype: failed to convert scanner error, %s, target type %s", keys, e.TargetType.String())
	case FailedUnmarshalNew:
		return fmt.Sprintf("prototype: failed to unmarshal new error, %s, target type %s, %v", keys, e.SourceType.String(), e.err)
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
	return Error{Code: FailedParse, FullKeys: fks, TargetType: tgtType, Value: value, err: err}
}

func newPointerCycleError(fks []string, srcType reflect.Type) error {
	return Error{Code: PointerCycle, FullKeys: fks, SourceType: srcType}
}

func newUnsupportedTypeError(fks []string, tgtType reflect.Type, srcType reflect.Type) error {
	return Error{Code: UnsupportedType, FullKeys: fks, TargetType: tgtType, SourceType: srcType}
}

func newFailedConvertScannerError(fks []string, tgtType reflect.Type) error {
	return Error{Code: FailedConvertScanner, FullKeys: fks, TargetType: tgtType}
}

func newUnmarshalNewError(fks []string, srcType reflect.Type, err error) error {
	return Error{Code: FailedUnmarshalNew, FullKeys: fks, SourceType: srcType, err: err}
}
