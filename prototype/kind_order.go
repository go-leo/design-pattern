package prototype

import "reflect"

const (
	_ int = iota
	_Bool
	_Int
	_Int8
	_Int16
	_Int32
	_Int64
	_Uint
	_Uint8
	_Uint16
	_Uint32
	_Uint64
	_Uintptr
	_Float32
	_Float64
	_String
	_Array
	_Slice
	_Map
	_Struct
	_Pointer
	_Interface
	_Complex64
	_Complex128
	_Chan
	_Func
	_UnsafePointer
	_Invalid
)

var _KindOrder = map[reflect.Kind]int{
	reflect.Invalid:       _Invalid,
	reflect.Bool:          _Bool,
	reflect.Int:           _Int,
	reflect.Int8:          _Int8,
	reflect.Int16:         _Int16,
	reflect.Int32:         _Int32,
	reflect.Int64:         _Int64,
	reflect.Uint:          _Uint,
	reflect.Uint8:         _Uint8,
	reflect.Uint16:        _Uint16,
	reflect.Uint32:        _Uint32,
	reflect.Uint64:        _Uint64,
	reflect.Uintptr:       _Uintptr,
	reflect.Float32:       _Float32,
	reflect.Float64:       _Float64,
	reflect.String:        _String,
	reflect.Array:         _Array,
	reflect.Slice:         _Slice,
	reflect.Map:           _Map,
	reflect.Struct:        _Struct,
	reflect.Pointer:       _Pointer,
	reflect.Interface:     _Interface,
	reflect.Complex64:     _Complex64,
	reflect.Complex128:    _Complex128,
	reflect.Chan:          _Chan,
	reflect.Func:          _Func,
	reflect.UnsafePointer: _UnsafePointer,
}
