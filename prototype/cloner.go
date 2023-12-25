package prototype

import (
	"encoding"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// clonerFunc 通用克隆方法
type clonerFunc func(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error

// valueCloner 基于 reflect.Value 获取 clonerFunc
func valueCloner(srcVal reflect.Value, opts *options) clonerFunc {
	if !srcVal.IsValid() {
		return func(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
			return nil
		}
	}
	return typeCloner(srcVal.Type(), true, opts)
}

// typeCloner 基于 reflect.Type 获取 clonerFunc
func typeCloner(srcType reflect.Type, allowAddr bool, opts *options) clonerFunc {
	if srcType.Kind() != reflect.Pointer && allowAddr && reflect.PointerTo(srcType).Implements(clonerToType) {
		return addrClonerToCloner()
	}
	if srcType.Implements(clonerToType) {
		return clonerToCloner
	}
	switch srcType.Kind() {
	case reflect.Bool:
		return boolCloner
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intCloner
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintCloner
	case reflect.Float32, reflect.Float64:
		return floatCloner
	case reflect.String:
		return stringCloner
	case reflect.Struct:
		return structCloner
	case reflect.Map:
		return mapCloner
	case reflect.Slice:
		return sliceCloner
	case reflect.Array:
		return arrayCloner
	case reflect.Interface:
		return interfaceCloner
	case reflect.Pointer:
		return pointerCloner
	default:
		return unsupportedTypeCloner
	}
}

// clonerToCloner 实现了 ClonerTo 接口，直接调用
func clonerToCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.Kind() == reflect.Pointer && srcVal.IsNil() {
		return nil
	}
	cloner, ok := srcVal.Interface().(ClonerTo)
	if !ok {
		return typeCloner(srcVal.Type(), false, opts)(e, fks, tgtVal, srcVal, opts)
	}
	if tgtVal.Kind() == reflect.Pointer {
		return cloner.CloneTo(tgtVal.Interface())
	}
	if tgtVal.CanAddr() {
		tgtAddr := tgtVal.Addr()
		return cloner.CloneTo(tgtAddr.Interface())
	}
	return typeCloner(srcVal.Type(), false, opts)(e, fks, tgtVal, srcVal, opts)
}

// addrClonerToCloner 实现了 ClonerTo 接口，直接调用
func addrClonerToCloner() clonerFunc {
	return func(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
		if !srcVal.CanAddr() {
			return typeCloner(srcVal.Type(), false, opts)(e, fks, tgtVal, srcVal, opts)
		}
		srcAddr := srcVal.Addr()
		if srcAddr.IsNil() {
			return nil
		}
		cloner, ok := srcAddr.Interface().(ClonerTo)
		if !ok {
			return typeCloner(srcVal.Type(), false, opts)(e, fks, tgtVal, srcVal, opts)
		}
		if tgtVal.Kind() == reflect.Pointer {
			return cloner.CloneTo(tgtVal.Interface())
		}
		if tgtVal.CanAddr() {
			return cloner.CloneTo(tgtVal.Addr().Interface())
		}
		return typeCloner(srcVal.Type(), false, opts)(e, fks, tgtVal, srcVal, opts)
	}
}

/*
boolCloner 克隆bool类型
bool ----> ClonerFrom
bool ----> bool(true, false)
bool ----> any(true, false)
bool ----> string("true", "false")
bool ----> int(true:1, false:0)
bool ----> uint(true:1, false:0)
bool ----> float(true:1, false:0)
bool ----> pointer
bool ----> struct
*/
func boolCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	b := srcVal.Bool()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(b)
	}
	switch tv.Kind() {
	case reflect.Bool:
		tv.SetBool(b)
		return nil
	case reflect.Interface:
		return setAnyValue(e, fks, tgtVal, srcVal, opts, tv, b)
	case reflect.String:
		tv.SetString(strconv.FormatBool(b))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(fks, tv, int64(boolIntMap[b]))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setInt2Uint(fks, tv, int64(boolIntMap[b]))
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, float64(boolIntMap[b]))
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, boolCloner)
	case reflect.Struct:
		return setBoolToStruct(e, fks, tv, srcVal, opts, b)
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
intCloner 克隆int类型
int ----> ClonerFrom
int ----> int(i)
int ----> any(i)
int ----> uint(i)
int ----> float(i)
int ----> bool(0->false, !0->true)
int ----> string("i")
int ----> pointer
int ----> struct
*/
func intCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	i := srcVal.Int()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(i)
	}
	switch tv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(fks, tv, i)
	case reflect.Interface:
		return setAnyValue(e, fks, tgtVal, srcVal, opts, tv, i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setInt2Uint(fks, tv, i)
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, float64(i))
	case reflect.Bool:
		tv.SetBool(i != 0)
		return nil
	case reflect.String:
		tv.SetString(strconv.FormatInt(i, 10))
		return nil
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, intCloner)
	case reflect.Struct:
		return setIntToStruct(e, fks, tv, srcVal, opts, i)
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
uintCloner 克隆uint类型
uint ----> ClonerFrom
uint ----> uint(u)
uint ----> any(u)
uint ----> int(u)
uint ----> float(u)
uint ----> bool(0->false, !0->true)
uint ----> string("u")
uint ----> pointer
uint ----> struct
*/
func uintCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	u := srcVal.Uint()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(u)
	}
	switch tv.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setUint(fks, tv, u)
	case reflect.Interface:
		return setAnyValue(e, fks, tgtVal, srcVal, opts, tv, u)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setUint2Int(fks, tv, u)
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, float64(u))
	case reflect.Bool:
		tv.SetBool(u != 0)
		return nil
	case reflect.String:
		tv.SetString(strconv.FormatUint(u, 10))
		return nil
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, uintCloner)
	case reflect.Struct:
		return setUintToStruct(e, fks, tv, srcVal, opts, u)
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
floatCloner 克隆float类型
float ----> ClonerFrom
float ----> float(f)
float ----> any(f)
float ----> int(f)
float ----> uint(f)
float ----> bool(0->false, !0->true)
float ----> string("f")
float ----> pointer
float ----> struct
*/
func floatCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	f := srcVal.Float()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(f)
	}
	switch tv.Kind() {
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, f)
	case reflect.Interface:
		return setAnyValue(e, fks, tgtVal, srcVal, opts, tv, f)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setFloat2Int(fks, tv, f)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setFloat2Uint(fks, tv, f)
	case reflect.Bool:
		tv.SetBool(f != 0)
		return nil
	case reflect.String:
		tv.SetString(strconv.FormatFloat(f, 'f', -1, 64))
		return nil
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, floatCloner)
	case reflect.Struct:
		return setFloatToStruct(e, fks, tv, srcVal, opts, f)
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
stringCloner 克隆float类型
string ----> ClonerFrom
string ----> string
string ----> []byte
string ----> any(string)
string ----> bool(strconv.ParseBool)
string ----> int(strconv.ParseInt)
string ----> uint(strconv.ParseUint)
string ----> float(strconv.ParseFloat)
string ----> pointer
string ----> struct
*/
func stringCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	s := srcVal.String()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(s)
	}
	switch tv.Kind() {
	case reflect.String:
		tv.SetString(s)
		return nil
	case reflect.Slice:
		if tv.Type().Elem().Kind() == reflect.Uint8 {
			tv.SetBytes([]byte(s))
			return nil
		}
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Interface:
		return setAnyValue(e, fks, tgtVal, srcVal, opts, tv, s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newParseError(fks, tv.Type(), s, err)
		}
		tv.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newParseError(fks, tv.Type(), s, err)
		}
		return setInt(fks, tv, i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newParseError(fks, tv.Type(), s, err)
		}
		return setUint(fks, tv, u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newParseError(fks, tv.Type(), s, err)
		}
		return setFloat(fks, tv, f)
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, stringCloner)
	case reflect.Struct:
		return setStringToStruct(e, fks, tv, srcVal, opts, s)
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
bytesCloner 克隆[]byte类型
[]byte ----> ClonerFrom
[]byte ----> []byte
[]byte ----> string(base64)
[]byte ----> any([]byte)
string ----> bool(strconv.ParseBool)
string ----> int(strconv.ParseInt)
string ----> uint(strconv.ParseUint)
string ----> float(strconv.ParseFloat)
[]byte ----> pointer
[]byte ----> struct
*/
func bytesCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	bs := srcVal.Bytes()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(bs)
	}
	switch tv.Kind() {
	case reflect.Slice:
		if tv.Type().Elem().Kind() == reflect.Uint8 {
			tv.SetBytes(bs)
			return nil
		}
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.String:
		tv.SetString(base64.StdEncoding.EncodeToString(bs))
		return nil
	case reflect.Interface:
		return setAnyValue(e, fks, tgtVal, srcVal, opts, tv, bs)
	case reflect.Bool:
		s := string(bs)
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newParseError(fks, tv.Type(), s, err)
		}
		tv.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := string(bs)
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newParseError(fks, tv.Type(), s, err)
		}
		return setInt(fks, tv, i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		s := string(bs)
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newParseError(fks, tv.Type(), s, err)
		}
		return setUint(fks, tv, u)
	case reflect.Float32, reflect.Float64:
		s := string(bs)
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newParseError(fks, tv.Type(), s, err)
		}
		return setFloat(fks, tv, f)
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, stringCloner)
	case reflect.Struct:
		return setBytesToStruct(e, fks, tv, srcVal, opts, bs)
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
interfaceCloner 克隆interface类型
interface ----> reflect.Value ----> valueCloner
*/
func interfaceCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
	}
	srcVal = srcVal.Elem()
	cloner := valueCloner(srcVal, opts)
	return cloner(e, fks, tgtVal, srcVal, opts)
}

/*
pointerCloner 克隆pointer类型
pointer ----> reflect.Type ----> typeCloner
*/
func pointerCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
	}
	cloner := e.checkPointerCycle(
		func(srcVal reflect.Value) any { return srcVal.Interface() },
		typeCloner(srcVal.Type().Elem(), true, opts),
	)
	return cloner(e, fks, tgtVal, srcVal.Elem(), opts)
}

/*
timeCloner 克隆time.Time类型
time.IntToTime ----> ClonerFrom
time.IntToTime ----> struct(time.IntToTime)
time.IntToTime ----> any(time.IntToTime)
time.IntToTime ----> string(time.RFC3339)
time.IntToTime ----> []byte(time.RFC3339)
time.IntToTime ----> int(time.Unix)
time.IntToTime ----> uint(time.Unix)
time.IntToTime ----> float(time.Unix)
time.IntToTime ----> pointer
*/
func timeCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	t := srcVal.Interface().(time.Time)
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(t)
	}
	switch tv.Kind() {
	case reflect.Struct:
		return setTimeToStruct(e, fks, tv, srcVal, opts, t)
	case reflect.Interface:
		return setAnyValue(e, fks, tgtVal, srcVal, opts, tv, t)
	case reflect.String:
		tv.SetString(t.Format(time.RFC3339))
		return nil
	case reflect.Slice:
		if tv.Type().Elem().Kind() == reflect.Uint8 {
			tv.SetBytes([]byte(t.Format(time.RFC3339)))
			return nil
		}
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(fks, tv, opts.TimeToInt(t))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setUint(fks, tv, uint64(opts.TimeToInt(t)))
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, float64(opts.TimeToInt(t)))
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, timeCloner)
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
sliceCloner 克隆slice类型,
[]byte ----> bytesCloner
slice ----> ClonerFrom
slice ----> slice,
slice ----> array,
slice ----> any(slice),
slice ----> pointer,
*/
func sliceCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	srcType := srcVal.Type()
	if srcType.Elem().Kind() == reflect.Uint8 {
		return bytesCloner(e, fks, tgtVal, srcVal, opts)
	}
	if srcVal.IsNil() {
		return nil
	}
	cloner := e.checkPointerCycle(
		func(srcVal reflect.Value) any {
			return struct {
				ptr interface{}
				len int
			}{
				ptr: srcVal.UnsafePointer(),
				len: srcVal.Len(),
			}
		},
		arrayCloner,
	)
	return cloner(e, fks, tgtVal, srcVal, opts)
}

/*
arrayCloner 克隆array类型,
array ----> ClonerFrom
array ----> array,
array ----> slice,
array ----> any(slice),
array ----> pointer,
*/
func arrayCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(srcVal.Interface())
	}
	switch tv.Kind() {
	case reflect.Array, reflect.Slice:
		return setSliceArray(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.Interface:
		return setAnySlice(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, arrayCloner)
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func mapCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
	}
	cloner := e.checkPointerCycle(
		func(srcVal reflect.Value) any { return srcVal.UnsafePointer() },
		_mapCloner,
	)
	return cloner(e, fks, tgtVal, srcVal, opts)
}

func _mapCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(srcVal.Interface())
	}

	switch tv.Kind() {
	case reflect.Map:
		return setMapToMap(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.Interface:
		return setMapToAny(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.Struct:
		return setMapToStruct(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, _mapCloner)
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func kvPairs(srcVal reflect.Value) ([]kvPair, error) {
	// Extract and sort the keys.
	kayValPairs := make([]kvPair, srcVal.Len())
	mapIter := srcVal.MapRange()
	for i := 0; mapIter.Next(); i++ {
		kayValPairs[i].kVal = mapIter.Key()
		kayValPairs[i].vVal = mapIter.Value()
		if err := kayValPairs[i].resolve(); err != nil {
			return nil, fmt.Errorf("prototype: map resolve error for type %q: %q", srcVal.Type().String(), err.Error())
		}
	}
	sort.Slice(kayValPairs, func(i, j int) bool { return strings.Compare(kayValPairs[i].keyStr, kayValPairs[j].keyStr) < 0 })
	return kayValPairs, nil
}

/*
structCloner 克隆 struct 类型
sql.NullBool ----> boolCloner
sql.NullByte ----> uintCloner
sql.NullInt16 ----> intCloner
sql.NullInt32 ----> intCloner
sql.NullInt64 ----> intCloner
sql.NullFloat64 ----> floatCloner
sql.NullString ----> stringCloner
sql.NullTime ----> timeCloner
struct ----> ClonerFrom
struct ----> struct
struct ----> any(map[string]any)
struct ----> map
struct ----> pointer
*/
func structCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}

	switch srcVal.Type() {
	case timeType:
		return timeCloner(e, fks, tgtVal, srcVal, opts)

	case sqlNullBoolType:
		return boolCloner(e, fks, tgtVal, srcVal.FieldByName("Bool"), opts)
	case sqlNullByteType:
		return uintCloner(e, fks, tgtVal, srcVal.FieldByName("Byte"), opts)
	case sqlNullInt16Type:
		return intCloner(e, fks, tgtVal, srcVal.FieldByName("Int16"), opts)
	case sqlNullInt32Type:
		return intCloner(e, fks, tgtVal, srcVal.FieldByName("Int32"), opts)
	case sqlNullInt64Type:
		return intCloner(e, fks, tgtVal, srcVal.FieldByName("Int64"), opts)
	case sqlNullFloat64Type:
		return floatCloner(e, fks, tgtVal, srcVal.FieldByName("Float64"), opts)
	case sqlNullStringType:
		return stringCloner(e, fks, tgtVal, srcVal.FieldByName("String"), opts)
	case sqlNullTimeType:
		return timeCloner(e, fks, tgtVal, srcVal.FieldByName("IntToTime"), opts)

	case timestampPBTimestampType:
		timestamp, _ := srcVal.Interface().(timestamppb.Timestamp)
		return timeCloner(e, fks, tgtVal, reflect.ValueOf(timestamp.AsTime()), opts)
	case durationPBDurationType:
		duration, _ := srcVal.Interface().(durationpb.Duration)
		return intCloner(e, fks, tgtVal, reflect.ValueOf(duration.AsDuration()), opts)

	case wrappersPBBoolType:
		return boolCloner(e, fks, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBInt32Type, wrappersPBInt64Type:
		return intCloner(e, fks, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBUint32Type, wrappersPBUint64Type:
		return uintCloner(e, fks, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBFloatType, wrappersPBDoubleType:
		return floatCloner(e, fks, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBStringType:
		return stringCloner(e, fks, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBBytesType:
		return bytesCloner(e, fks, tgtVal, srcVal.FieldByName("Value"), opts)

	case anyPBAnyType:
		if srcVal.CanAddr() {
			srcPtr := srcVal.Addr().Interface().(*anypb.Any)
			message, err := srcPtr.UnmarshalNew()
			if err != nil {
				return newUnmarshalNewError(fks, srcVal.Type(), err)
			}
			return interfaceCloner(e, fks, tgtVal, reflect.ValueOf(message), opts)
		}
		return bytesCloner(e, fks, tgtVal, srcVal.FieldByName("Value"), opts)
	case emptyPBEmptyType:
		return nil

	case structPBStructType:
		structPB, _ := srcVal.Interface().(structpb.Struct)
		return mapCloner(e, fks, tgtVal, reflect.ValueOf(structPB.AsMap()), opts)
	case structPBListType:
		structPB, _ := srcVal.Interface().(structpb.ListValue)
		return sliceCloner(e, fks, tgtVal, reflect.ValueOf(structPB.AsSlice()), opts)
	case structPBValueType:
		valuePB, _ := srcVal.Interface().(structpb.Value)
		return interfaceCloner(e, fks, tgtVal, reflect.ValueOf(valuePB.AsInterface()), opts)
	case structPBNullValueType:
		return nil
	case structPBNumberValueType:
		numberPB, _ := srcVal.Interface().(structpb.Value_NumberValue)
		return floatCloner(e, fks, tgtVal, reflect.ValueOf(numberPB.NumberValue), opts)
	case structPBStringValueType:
		stringPB, _ := srcVal.Interface().(structpb.Value_StringValue)
		return stringCloner(e, fks, tgtVal, reflect.ValueOf(stringPB.StringValue), opts)
	case structPBBoolValueType:
		boolPB, _ := srcVal.Interface().(structpb.Value_BoolValue)
		return boolCloner(e, fks, tgtVal, reflect.ValueOf(boolPB.BoolValue), opts)
	case structPBStructValueType:
		structPB, _ := srcVal.Interface().(structpb.Value_StructValue)
		return mapCloner(e, fks, tgtVal, reflect.ValueOf(structPB.StructValue.AsMap()), opts)
	case structPBListValueType:
		listPB, _ := srcVal.Interface().(structpb.Value_ListValue)
		return sliceCloner(e, fks, tgtVal, reflect.ValueOf(listPB.ListValue.AsSlice()), opts)

	default:
		cloner, tv := indirectValue(tgtVal)
		if cloner != nil {
			return cloner.CloneFrom(srcVal.Interface())
		}
		switch tv.Kind() {
		case reflect.Struct:
			return struct2StructCloner(e, fks, tv, srcVal, opts)
		case reflect.Interface:
			if tv.NumMethod() == 0 {
				return struct2AnyCloner(e, fks, tv, srcVal, opts)
			}
			return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
		case reflect.Map:
			tgtType := tv.Type()
			tgtKeyType := tgtType.Key()
			if slices.Contains(allSampleKinds, tgtKeyType.Kind()) ||
				reflect.PointerTo(tgtKeyType).Implements(textUnmarshalerType) {
				if tv.IsNil() {
					tv.Set(reflect.MakeMap(tgtType))
				}
				return struct2MapCloner(e, fks, tv, srcVal, opts)
			}
			return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
		case reflect.Pointer:
			return setPointer(e, fks, tgtVal, srcVal, opts, tv, structCloner)
		default:
			return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
		}
	}
}

func struct2StructCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	tgtType := tgtVal.Type()
	tgtFields := cachedTypeFields(tgtType, opts, opts.TagKey)
	srcType := srcVal.Type()
	srcFields := cachedTypeFields(srcType, opts, opts.TagKey)
	if err := struct2StructDominantFieldCloner(e, fks, tgtVal, srcVal, tgtType, srcType, tgtFields, srcFields, opts); err != nil {
		return err
	}
	if err := struct2StructRecessivesFieldCloner(e, fks, tgtVal, srcVal, tgtType, srcType, tgtFields, srcFields, opts); err != nil {
		return err
	}
	return nil
}

func struct2StructDominantFieldCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, tgtType, srcType reflect.Type, tgtFields, srcFields jstructFields, opts *options) error {
	// 复制字段, 循环src字段
	for srcName, srcIdx := range srcFields.dominantsNameIndex {
		srcDominantField := srcFields.dominants[srcIdx]
		// 查找src字段值
		srcDominantFieldVal, ok := findValue(srcVal, srcDominantField)
		if !ok {
			continue
		}

		// 查找 tgt 主要字段
		tgtDominantField, ok := findDominantField(tgtFields, opts, srcName)
		if !ok {
			// 没有找到目标，则跳过
			continue
		}

		// 查找tgt字段值
		tgtDominantFieldVal, err := findSettableValue(tgtVal, tgtDominantField)
		if err != nil {
			return err
		}

		// 克隆src字段到tgt字段
		if err := srcDominantField.clonerFunc(e, append(slices.Clone(fks), srcName), tgtDominantFieldVal, srcDominantFieldVal, opts); err != nil {
			return err
		}
	}
	return nil
}

func struct2StructRecessivesFieldCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, tgtType, srcType reflect.Type, tgtFields, srcFields jstructFields, opts *options) error {
	// 复制字段, 循环src字段
	for srcKey, srcIdxs := range srcFields.recessivesNameIndex {
		srcRecessiveFieldValMap := make(map[string]reflect.Value)
		for _, srcIdx := range srcIdxs {
			srcRecessiveField := srcFields.recessives[srcIdx]
			// 查找src字段值
			srcRecessiveFieldVal, ok := findValue(srcVal, srcRecessiveField)
			if !ok {
				continue
			}
			srcRecessiveFieldValMap[srcRecessiveField.fullName] = srcRecessiveFieldVal
		}
		if len(srcRecessiveFieldValMap) <= 0 {
			continue
		}

		tgtRecessiveFields, ok := findRecessiveField(tgtFields, opts, srcKey)
		if !ok {
			continue
		}
		if len(tgtRecessiveFields) <= 0 {
			continue
		}

		tgtRecessiveFieldValMap := make(map[string]reflect.Value)
		for _, recessiveField := range tgtRecessiveFields {
			// 查找tgt字段值
			tgtFieldVal, err := findSettableValue(tgtVal, recessiveField)
			if err != nil {
				return err
			}
			tgtRecessiveFieldValMap[recessiveField.fullName] = tgtFieldVal
		}

		for fullName, srcRecessiveFieldVal := range srcRecessiveFieldValMap {
			tgtRecessiveFieldVal, ok := tgtRecessiveFieldValMap[fullName]
			if !ok {
				continue
			}

			// 克隆src字段到tgt字段
			srcRecessiveField := srcFields.recessives[srcFields.recessivesFullNameIndex[fullName]]
			if err := srcRecessiveField.clonerFunc(e, append(slices.Clone(fks), srcKey), tgtRecessiveFieldVal, srcRecessiveFieldVal, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func struct2AnyCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	m := make(map[string]any)
	srcType := srcVal.Type()
	srcFields := cachedTypeFields(srcType, opts, opts.TagKey)
	for _, selfField := range srcFields.selfFields {
		// 查找src字段值
		srcDominantFieldVal, ok := findValue(srcVal, selfField)
		if !ok {
			continue
		}

		var vVal reflect.Value
		switch srcDominantFieldVal.Kind() {
		case reflect.String:
			vVal = reflect.ValueOf(new(string))
		case reflect.Bool:
			vVal = reflect.ValueOf(new(bool))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			vVal = reflect.ValueOf(new(int64))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			vVal = reflect.ValueOf(new(uint64))
		case reflect.Float32, reflect.Float64:
			vVal = reflect.ValueOf(new(float64))
		case reflect.Slice, reflect.Array:
			vVal = reflect.ValueOf(new([]any))
		case reflect.Map, reflect.Struct:
			vVal = reflect.ValueOf(new(map[string]any))
		case reflect.Interface, reflect.Pointer:
			vVal = reflect.New(srcType)
		default:
			return errors.New("prototype: Unexpected type")
		}

		// 克隆src字段到tgt字段
		if err := selfField.clonerFunc(e, append(slices.Clone(fks), selfField.name), vVal, srcDominantFieldVal, opts); err != nil {
			return err
		}

		m[selfField.name] = vVal.Elem().Interface()
	}

	tgtVal.Set(reflect.ValueOf(m))
	return nil
}

func struct2MapCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	srcType := srcVal.Type()
	srcFields := cachedTypeFields(srcType, opts, opts.TagKey)
	tgtType := tgtVal.Type()

	// 复制字段, 循环src字段
	var mapElem reflect.Value
	elemType := tgtType.Elem()
	for _, selfField := range srcFields.selfFields {
		fieldFks := append(slices.Clone(fks), selfField.name)

		// 查找src字段值
		srcDominantFieldVal, ok := findValue(srcVal, selfField)
		if !ok {
			continue
		}

		if !mapElem.IsValid() {
			mapElem = reflect.New(elemType).Elem()
		} else {
			mapElem.Set(reflect.Zero(elemType))
		}
		vVal := mapElem

		// 克隆src字段到tgt字段
		if err := selfField.clonerFunc(e, fieldFks, vVal, srcDominantFieldVal, opts); err != nil {
			return err
		}

		kType := tgtType.Key()
		var kVal reflect.Value
		switch {
		case reflect.PointerTo(kType).Implements(textUnmarshalerType):
			kVal = reflect.New(kType)
			if u, ok := kVal.Interface().(encoding.TextUnmarshaler); ok {
				err := u.UnmarshalText([]byte(selfField.name))
				if err != nil {
					return err
				}
			}
			kVal = kVal.Elem()
		case kType.Kind() == reflect.String:
			kVal = reflect.ValueOf(selfField.name).Convert(kType)
		case kType.Kind() == reflect.Bool:
			b, err := strconv.ParseBool(selfField.name)
			if err != nil {
				return err
			}
			kVal = reflect.ValueOf(b).Convert(kType)
		case slices.Contains(intKinds, kType.Kind()):
			i, err := strconv.ParseInt(selfField.name, 10, 64)
			if err != nil {
				return err
			}
			kVal = reflect.Zero(kType)
			if err := setInt(fks, kVal, i); err != nil {
				return err
			}
		case slices.Contains(uintKinds, kType.Kind()):
			u, err := strconv.ParseUint(selfField.name, 10, 64)
			if err != nil {
				return err
			}
			kVal = reflect.Zero(kType)
			if err := setUint(fks, kVal, u); err != nil {
				return err
			}
		case slices.Contains(floatKinds, kType.Kind()):
			f, err := strconv.ParseFloat(selfField.name, 64)
			if err != nil {
				return err
			}
			kVal = reflect.Zero(kType)
			if err := setFloat(fieldFks, kVal, f); err != nil {
				return err
			}
		default:
			return errors.New("prototype: Unexpected key type")
		}
		if !kVal.IsValid() {
			continue
		}
		tgtVal.SetMapIndex(kVal, vVal)
	}
	return nil
}

type kvPair struct {
	kVal   reflect.Value
	vVal   reflect.Value
	keyStr string
}

func (p *kvPair) resolve() error {
	if tm, ok := p.kVal.Interface().(encoding.TextMarshaler); ok {
		if p.kVal.Kind() == reflect.Pointer && p.kVal.IsNil() {
			return nil
		}
		buf, err := tm.MarshalText()
		p.keyStr = string(buf)
		return err
	}
	if str, ok := p.kVal.Interface().(fmt.Stringer); ok {
		if p.kVal.Kind() == reflect.Pointer && p.kVal.IsNil() {
			return nil
		}
		p.keyStr = str.String()
		return nil
	}
	switch p.kVal.Kind() {
	case reflect.String:
		p.keyStr = p.kVal.String()
		return nil
	case reflect.Bool:
		p.keyStr = strconv.FormatBool(p.kVal.Bool())
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p.keyStr = strconv.FormatInt(p.kVal.Int(), 10)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p.keyStr = strconv.FormatUint(p.kVal.Uint(), 10)
		return nil
	case reflect.Float32, reflect.Float64:
		p.keyStr = strconv.FormatFloat(p.kVal.Float(), 'f', -1, 64)
		return nil
	default:
		return errors.New("unexpected map key type")
	}
}

func unsupportedTypeCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	return hookCloner(e, fks, tgtVal, srcVal, opts)
}

func hookCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	valueHooks, ok := opts.ValueHook[srcVal]
	if ok {
		hook, ok := valueHooks[tgtVal]
		if ok {
			return hook(fks, tgtVal, srcVal)
		}
	}
	typeHooks, ok := opts.TypeHooks[srcVal.Type()]
	if ok {
		hook, ok := typeHooks[tgtVal.Type()]
		if ok {
			return hook(fks, tgtVal, srcVal)
		}
	}
	kindHooks, ok := opts.KindHooks[srcVal.Kind()]
	if ok {
		hook, ok := kindHooks[tgtVal.Kind()]
		if ok {
			return hook(fks, tgtVal, srcVal)
		}
	}
	return newUnsupportedTypeError(fks, srcVal.Type(), tgtVal.Type())
}
