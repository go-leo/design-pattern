package prototype

import (
	"errors"
	"reflect"
)

// phasePanicMsg is used as a panic message when we end up with something that
// shouldn't happen. It can indicate a bug in the JSON decoder, or that
// something is editing the data slice while the decoder executes.
var ErrPhase = errors.New("JSON decoder out of sync - data changing underfoot?")

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "prototype: unsupported type: " + e.Type.String()
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

// A MarshalerError represents an error from calling a MarshalJSON or MarshalText method.
type MarshalerError struct {
	Type       reflect.Type
	Err        error
	sourceFunc string
}

func (e *MarshalerError) Error() string {
	srcFunc := e.sourceFunc
	if srcFunc == "" {
		srcFunc = "MarshalJSON"
	}
	return "json: error calling " + srcFunc +
		" for type " + e.Type.String() +
		": " + e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *MarshalerError) Unwrap() error { return e.Err }

type OverflowError struct {
	FullKeys   []string
	TargetType reflect.Type
	Value      string
}

func (e *OverflowError) Error() string {
	return "prototype: overflow error, cannot Clone " + e.Value + " into Go value of type " + e.TargetType.String()
}

type NegativeNumberError struct {
	FullKeys    []string
	TargetValue reflect.Value
	Value       string
}

func (e *NegativeNumberError) Error() string {
	return "prototype: negative number error, cannot Clone " + e.Value + " into Go value of type " + e.TargetValue.Type().String()
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

type InvalidSourceError struct {
	Type reflect.Type
}

func (e *InvalidSourceError) Error() string {
	if e.Type == nil {
		return "prototype: Clone(tgt, nil)"
	}
	if e.Type.Kind() != reflect.Pointer {
		return "prototype: Clone(non-pointer " + e.Type.String() + ", src)"
	}
	return "prototype: Clone(nil " + e.Type.String() + ", src)"
}
