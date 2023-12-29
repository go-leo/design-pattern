package internal

import "reflect"

const (
	Invalid uint = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	String
	Array
	Slice
	Map
	Struct
	Pointer
	Interface
	Complex64
	Complex128
	Chan
	Func
	UnsafePointer
)

var KindOrder = map[reflect.Kind]uint{
	reflect.Invalid:       Invalid,
	reflect.Bool:          Bool,
	reflect.Int:           Int,
	reflect.Int8:          Int8,
	reflect.Int16:         Int16,
	reflect.Int32:         Int32,
	reflect.Int64:         Int64,
	reflect.Uint:          Uint,
	reflect.Uint8:         Uint8,
	reflect.Uint16:        Uint16,
	reflect.Uint32:        Uint32,
	reflect.Uint64:        Uint64,
	reflect.Uintptr:       Uintptr,
	reflect.Float32:       Float32,
	reflect.Float64:       Float64,
	reflect.String:        String,
	reflect.Array:         Array,
	reflect.Slice:         Slice,
	reflect.Map:           Map,
	reflect.Struct:        Struct,
	reflect.Pointer:       Pointer,
	reflect.Interface:     Interface,
	reflect.Complex64:     Complex64,
	reflect.Complex128:    Complex128,
	reflect.Chan:          Chan,
	reflect.Func:          Func,
	reflect.UnsafePointer: UnsafePointer,
}
