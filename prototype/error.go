package prototype

import (
	"fmt"
	"reflect"
	"strings"
)

type Code int

const (
	NonPointer               Code = 1
	Nil                           = 2
	Overflow                      = 3
	NegativeNumber                = 4
	FailedParse                   = 5
	PointerCycle                  = 6
	UnsupportedType               = 7
	FailedConvertScanner          = 8
	FailedUnmarshalNew            = 9
	FailedStringify               = 10
	FailedSetEmbeddedPointer      = 11
)

type Error struct {
	Code       Code
	Labels     []string
	TargetType reflect.Type
	SourceType reflect.Type
	Value      string
	err        error
}

func (e Error) Error() string {
	labels := strings.Join(e.Labels, ".")
	switch e.Code {
	case NonPointer:
		return fmt.Sprintf("prototype: non-pointer error, target type is %s", e.TargetType.String())
	case Nil:
		if e.TargetType == nil {
			return "prototype: nil error, target is nil"
		}
		return fmt.Sprintf("prototype: nil error, target is (%s)(nil)", e.TargetType.String())
	case Overflow:
		return fmt.Sprintf("prototype: overflow error, %s, cannot Clone %s into target type %s", labels, e.Value, e.TargetType.String())
	case NegativeNumber:
		return fmt.Sprintf("prototype: negative number error, %s, cannot Clone %s into target type %s", labels, e.Value, e.TargetType.String())
	case FailedParse:
		return fmt.Sprintf("prototype: failed to parse string error, %s, cannot Clone %s into target type %s, %v", labels, e.Value, e.TargetType.String(), e.err)
	case PointerCycle:
		return fmt.Sprintf("prototype: pointer cycle error, encountered a cycle via %s", e.SourceType.String())
	case UnsupportedType:
		return fmt.Sprintf("prototype: unsupported type error, %s, cannot Clone source type %s into  target type %s", labels, e.SourceType.String(), e.TargetType.String())
	case FailedConvertScanner:
		return fmt.Sprintf("prototype: failed to convert scanner error, %s, target type %s", labels, e.TargetType.String())
	case FailedUnmarshalNew:
		return fmt.Sprintf("prototype: failed to unmarshal new error, %s, target type %s, %v", labels, e.SourceType.String(), e.err)
	case FailedStringify:
		return fmt.Sprintf("prototype: failed to stringify, target type %s, %v", e.SourceType.String(), e.err)
	case FailedSetEmbeddedPointer:
		return fmt.Sprintf("prototype: failed to set embedded pointer, %s, unexported type %v", labels, e.TargetType.String())
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

func newOverflowError(labels []string, tgtType reflect.Type, value string) error {
	return Error{Code: Overflow, Labels: labels, TargetType: tgtType, Value: value}
}

func newNegativeNumberError(labels []string, tgtType reflect.Type, value string) error {
	return Error{Code: NegativeNumber, Labels: labels, TargetType: tgtType, Value: value}
}

func newParseError(labels []string, tgtType reflect.Type, value string, err error) error {
	return Error{Code: FailedParse, Labels: labels, TargetType: tgtType, Value: value, err: err}
}

func newPointerCycleError(labels []string, srcType reflect.Type) error {
	return Error{Code: PointerCycle, Labels: labels, SourceType: srcType}
}

func newUnsupportedTypeError(labels []string, tgtType reflect.Type, srcType reflect.Type) error {
	return Error{Code: UnsupportedType, Labels: labels, TargetType: tgtType, SourceType: srcType}
}

func newConvertScannerError(labels []string, tgtType reflect.Type) error {
	return Error{Code: FailedConvertScanner, Labels: labels, TargetType: tgtType}
}

func newUnmarshalNewError(labels []string, srcType reflect.Type, err error) error {
	return Error{Code: FailedUnmarshalNew, Labels: labels, SourceType: srcType, err: err}
}

func newStringifyError(srcType reflect.Type, err error) error {
	return Error{Code: FailedStringify, SourceType: srcType, err: err}
}

func newSetEmbeddedPointerError(labels []string, tgtType reflect.Type) error {
	return Error{Code: FailedSetEmbeddedPointer, Labels: labels, TargetType: tgtType}
}
