package prototype

import (
	"encoding"
	"encoding/base64"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"strconv"
	"time"
)

// clonerFunc 通用克隆方法
type clonerFunc func(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error

// valueCloner 基于 reflect.Value 获取 clonerFunc
func valueCloner(srcVal reflect.Value, opts *options) clonerFunc {
	if !srcVal.IsValid() {
		return func(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
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
	return kindCloner(srcType.Kind(), opts)
}

func kindCloner(srcKind reflect.Kind, opts *options) clonerFunc {
	switch srcKind {
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

// addrClonerToCloner 实现了 ClonerTo 接口，直接调用
func addrClonerToCloner() clonerFunc {
	return func(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
		if !srcVal.CanAddr() {
			return typeCloner(srcVal.Type(), false, opts)(g, labels, tgtVal, srcVal, opts)
		}
		srcAddr := srcVal.Addr()
		if srcAddr.IsNil() {
			return nil
		}
		if !srcAddr.CanInterface() {
			return typeCloner(srcVal.Type(), false, opts)(g, labels, tgtVal, srcVal, opts)
		}
		cloner, ok := srcAddr.Interface().(ClonerTo)
		if !ok {
			return typeCloner(srcVal.Type(), false, opts)(g, labels, tgtVal, srcVal, opts)
		}
		var tgtPtr any
		if tgtVal.Kind() == reflect.Pointer {
			tgtPtr = tgtVal.Interface()
		} else if tgtVal.CanAddr() {
			tgtPtr = tgtVal.Addr().Interface()
		} else {
			return kindCloner(srcVal.Type().Kind(), opts)(g, labels, tgtVal, srcVal, opts)
		}
		ok, err := cloner.CloneTo(tgtPtr)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		return kindCloner(srcVal.Type().Kind(), opts)(g, labels, tgtVal, srcVal, opts)
	}
}

// clonerToCloner 实现了 ClonerTo 接口，直接调用
func clonerToCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.Kind() == reflect.Pointer && srcVal.IsNil() {
		return nil
	}
	if !srcVal.CanInterface() {
		return kindCloner(srcVal.Type().Kind(), opts)(g, labels, tgtVal, srcVal, opts)
	}
	cloner, ok := srcVal.Interface().(ClonerTo)
	if !ok {
		return kindCloner(srcVal.Type().Kind(), opts)(g, labels, tgtVal, srcVal, opts)
	}
	var tgtPtr any
	if tgtVal.Kind() == reflect.Pointer && tgtVal.CanInterface() {
		tgtPtr = tgtVal.Interface()
	} else if tgtVal.CanAddr() {
		tgtPtr = tgtVal.Addr().Interface()
	} else {
		return kindCloner(srcVal.Type().Kind(), opts)(g, labels, tgtVal, srcVal, opts)
	}
	ok, err := cloner.CloneTo(tgtPtr)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return kindCloner(srcVal.Type().Kind(), opts)(g, labels, tgtVal, srcVal, opts)
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
func boolCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	b := srcVal.Bool()
	cloner, tv := indirect(tgtVal)
	if cloner != nil {
		ok, err := cloner.CloneFrom(b)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		tv = indirectValue(tv)
	}
	switch tv.Kind() {
	case reflect.Bool:
		tv.SetBool(b)
		return nil
	case reflect.Interface:
		return setAnyValue(g, labels, tv, srcVal, opts, b)
	case reflect.String:
		tv.SetString(strconv.FormatBool(b))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(labels, tv, int64(boolIntMap[b]))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setInt2Uint(labels, tv, int64(boolIntMap[b]))
	case reflect.Float32, reflect.Float64:
		return setFloat(labels, tv, float64(boolIntMap[b]))
	case reflect.Pointer:
		return setPointer(g, labels, tv, srcVal, opts)
	case reflect.Struct:
		return setBoolToStruct(g, labels, tv, srcVal, opts, b)
	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
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
func intCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	i := srcVal.Int()
	cloner, tv := indirect(tgtVal)
	if cloner != nil {
		ok, err := cloner.CloneFrom(i)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		tv = indirectValue(tv)
	}
	switch tv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(labels, tv, i)
	case reflect.Interface:
		return setAnyValue(g, labels, tv, srcVal, opts, i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setInt2Uint(labels, tv, i)
	case reflect.Float32, reflect.Float64:
		return setFloat(labels, tv, float64(i))
	case reflect.Bool:
		tv.SetBool(i != 0)
		return nil
	case reflect.String:
		tv.SetString(strconv.FormatInt(i, 10))
		return nil
	case reflect.Pointer:
		return setPointer(g, labels, tv, srcVal, opts)
	case reflect.Struct:
		return setIntToStruct(g, labels, tv, srcVal, opts, i)
	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
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
func uintCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	u := srcVal.Uint()
	cloner, tv := indirect(tgtVal)
	if cloner != nil {
		ok, err := cloner.CloneFrom(u)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		tv = indirectValue(tv)
	}
	switch tv.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setUint(labels, tv, u)
	case reflect.Interface:
		return setAnyValue(g, labels, tv, srcVal, opts, u)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setUint2Int(labels, tv, u)
	case reflect.Float32, reflect.Float64:
		return setFloat(labels, tv, float64(u))
	case reflect.Bool:
		tv.SetBool(u != 0)
		return nil
	case reflect.String:
		tv.SetString(strconv.FormatUint(u, 10))
		return nil
	case reflect.Pointer:
		return setPointer(g, labels, tv, srcVal, opts)
	case reflect.Struct:
		return setUintToStruct(g, labels, tv, srcVal, opts, u)
	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
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
func floatCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	f := srcVal.Float()
	cloner, tv := indirect(tgtVal)
	if cloner != nil {
		ok, err := cloner.CloneFrom(f)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		tv = indirectValue(tv)
	}
	switch tv.Kind() {
	case reflect.Float32, reflect.Float64:
		return setFloat(labels, tv, f)
	case reflect.Interface:
		return setAnyValue(g, labels, tv, srcVal, opts, f)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setFloat2Int(labels, tv, f)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setFloat2Uint(labels, tv, f)
	case reflect.Bool:
		tv.SetBool(f != 0)
		return nil
	case reflect.String:
		tv.SetString(strconv.FormatFloat(f, 'f', -1, 64))
		return nil
	case reflect.Pointer:
		return setPointer(g, labels, tv, srcVal, opts)
	case reflect.Struct:
		return setFloatToStruct(g, labels, tv, srcVal, opts, f)
	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
}

/*
stringCloner 克隆float类型
string ----> ClonerFrom
string ----> string
string ----> []byte
string ----> []rune
string ----> any(string)
string ----> bool(strconv.ParseBool)
string ----> int(strconv.ParseInt)
string ----> uint(strconv.ParseUint)
string ----> float(strconv.ParseFloat)
string ----> pointer
string ----> struct
*/
func stringCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	s := srcVal.String()
	cloner, tv := indirect(tgtVal)
	if cloner != nil {
		ok, err := cloner.CloneFrom(s)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		tv = indirectValue(tv)
	}

	if reflect.PointerTo(tv.Type()).Implements(textUnmarshalerType) && tv.CanAddr() {
		if unmarshaler, ok := tv.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return unmarshaler.UnmarshalText([]byte(s))
		}
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
		if tv.Type().Elem().Kind() == reflect.Int32 {
			tv.Set(reflect.ValueOf([]rune(s)))
			return nil
		}
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	case reflect.Interface:
		return setAnyValue(g, labels, tv, srcVal, opts, s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newParseError(labels, tv.Type(), s, err)
		}
		tv.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newParseError(labels, tv.Type(), s, err)
		}
		return setInt(labels, tv, i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newParseError(labels, tv.Type(), s, err)
		}
		return setUint(labels, tv, u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newParseError(labels, tv.Type(), s, err)
		}
		return setFloat(labels, tv, f)
	case reflect.Pointer:
		return setPointer(g, labels, tv, srcVal, opts)
	case reflect.Struct:
		return setStringToStruct(g, labels, tv, srcVal, opts, s)
	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
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
func bytesCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	bs := srcVal.Bytes()
	cloner, tv := indirect(tgtVal)
	if cloner != nil {
		ok, err := cloner.CloneFrom(bs)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		tv = indirectValue(tv)
	}

	if reflect.PointerTo(tv.Type()).Implements(textUnmarshalerType) && tv.CanAddr() {
		if unmarshaler, ok := tv.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return unmarshaler.UnmarshalText(bs)
		}
	}

	switch tv.Kind() {
	case reflect.Slice:
		if tv.Type().Elem().Kind() == reflect.Uint8 {
			tv.SetBytes(bs)
			return nil
		}
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	case reflect.String:
		tv.SetString(base64.StdEncoding.EncodeToString(bs))
		return nil
	case reflect.Interface:
		return setAnyValue(g, labels, tv, srcVal, opts, bs)
	case reflect.Bool:
		s := string(bs)
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newParseError(labels, tv.Type(), s, err)
		}
		tv.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := string(bs)
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newParseError(labels, tv.Type(), s, err)
		}
		return setInt(labels, tv, i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		s := string(bs)
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newParseError(labels, tv.Type(), s, err)
		}
		return setUint(labels, tv, u)
	case reflect.Float32, reflect.Float64:
		s := string(bs)
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newParseError(labels, tv.Type(), s, err)
		}
		return setFloat(labels, tv, f)
	case reflect.Pointer:
		return setPointer(g, labels, tv, srcVal, opts)
	case reflect.Struct:
		return setBytesToStruct(g, labels, tv, srcVal, opts, bs)
	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
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
func timeCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	t := srcVal.Interface().(time.Time)

	cloner, tv := indirect(tgtVal)
	if cloner != nil {
		ok, err := cloner.CloneFrom(t)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		tv = indirectValue(tv)
	}
	switch tv.Kind() {
	case reflect.Struct:
		return setTimeToStruct(g, labels, tv, srcVal, opts, t)
	case reflect.Interface:
		return setAnyValue(g, labels, tv, srcVal, opts, t)
	case reflect.String:
		tv.SetString(opts.TimeToString(t))
		return nil
	case reflect.Slice:
		if tv.Type().Elem().Kind() == reflect.Uint8 {
			tv.SetBytes([]byte(opts.TimeToString(t)))
			return nil
		}
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(labels, tv, opts.TimeToInt(t))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setUint(labels, tv, uint64(opts.TimeToInt(t)))
	case reflect.Float32, reflect.Float64:
		return setFloat(labels, tv, float64(opts.TimeToInt(t)))
	case reflect.Pointer:
		return setPointer(g, labels, tv, srcVal, opts)
	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
}

/*
interfaceCloner 克隆interface类型
interface ----> reflect.Value ----> valueCloner
*/
func interfaceCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
	}
	srcVal = srcVal.Elem()
	cloner := valueCloner(srcVal, opts)
	return cloner(g, labels, tgtVal, srcVal, opts)
}

/*
pointerCloner 克隆pointer类型
pointer ----> noDeepClone ----> tgtPtr = srcPtr
pointer ----> DeepClone ----> valueCloner
*/
func pointerCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
	}
	// 浅克隆，并且类型相同
	if !opts.DeepClone && tgtVal.Kind() == reflect.Pointer && indirectType(tgtVal.Type()) == indirectType(srcVal.Type()) {
		srcVal = indirectValue(srcVal)
		if srcVal.Kind() == reflect.Pointer && srcVal.IsNil() {
			return nil
		}
		for tgtVal.Type().Elem().Kind() == reflect.Pointer {
			if tgtVal.IsNil() {
				tgtVal.Set(reflect.New(tgtVal.Type().Elem()))
			}
			tgtVal = tgtVal.Elem()
		}
		tgtVal.Set(srcVal.Addr())
		return nil
	}

	srcVal = srcVal.Elem()
	cloner := g.checkPointerCycle(
		func(srcVal reflect.Value) any { return srcVal.Interface() },
		valueCloner(srcVal, opts),
	)
	return cloner(g, labels, tgtVal, srcVal, opts)
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
func sliceCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
	}
	if srcVal.Type().Elem().Kind() == reflect.Uint8 {
		return bytesCloner(g, labels, tgtVal, srcVal, opts)
	}
	cloner := g.checkPointerCycle(
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
	return cloner(g, labels, tgtVal, srcVal, opts)
}

/*
arrayCloner 克隆array类型,
array ----> ClonerFrom
array ----> array,
array ----> slice,
array ----> any(slice),
array ----> pointer,
*/
func arrayCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	cloner, tv := indirect(tgtVal)
	if cloner != nil {
		ok, err := cloner.CloneFrom(srcVal.Interface())
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		tv = indirectValue(tv)
	}
	switch tv.Kind() {
	case reflect.Array, reflect.Slice:
		return setSliceArray(g, labels, tv, srcVal, opts)
	case reflect.Interface:
		return setAnySlice(g, labels, tv, srcVal, opts)
	case reflect.Pointer:
		return setPointer(g, labels, tv, srcVal, opts)
	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
}

func mapCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
	}
	cloner := g.checkPointerCycle(
		func(srcVal reflect.Value) any { return srcVal.UnsafePointer() },
		_mapCloner,
	)
	return cloner(g, labels, tgtVal, srcVal, opts)
}

func _mapCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	cloner, tv := indirect(tgtVal)
	if cloner != nil {
		ok, err := cloner.CloneFrom(srcVal.Interface())
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		tv = indirectValue(tv)
	}

	switch tv.Kind() {
	case reflect.Map:
		return setMapToMap(g, labels, tv, srcVal, opts)
	case reflect.Interface:
		return setMapToAny(g, labels, tv, srcVal, opts)
	case reflect.Struct:
		return setMapToStruct(g, labels, tv, srcVal, opts)
	case reflect.Pointer:
		return setPointer(g, labels, tv, srcVal, opts)
	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
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
func structCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}

	switch srcVal.Type() {
	case timeType:
		return timeCloner(g, labels, tgtVal, srcVal, opts)

	case sqlNullBoolType:
		return boolCloner(g, labels, tgtVal, srcVal.FieldByName("Bool"), opts)
	case sqlNullByteType:
		return uintCloner(g, labels, tgtVal, srcVal.FieldByName("Byte"), opts)
	case sqlNullInt16Type:
		return intCloner(g, labels, tgtVal, srcVal.FieldByName("Int16"), opts)
	case sqlNullInt32Type:
		return intCloner(g, labels, tgtVal, srcVal.FieldByName("Int32"), opts)
	case sqlNullInt64Type:
		return intCloner(g, labels, tgtVal, srcVal.FieldByName("Int64"), opts)
	case sqlNullFloat64Type:
		return floatCloner(g, labels, tgtVal, srcVal.FieldByName("Float64"), opts)
	case sqlNullStringType:
		return stringCloner(g, labels, tgtVal, srcVal.FieldByName("String"), opts)
	case sqlNullTimeType:
		return timeCloner(g, labels, tgtVal, srcVal.FieldByName("Time"), opts)

	case timestampPBTimestampType:
		timestamp, _ := srcVal.Interface().(timestamppb.Timestamp)
		return timeCloner(g, labels, tgtVal, reflect.ValueOf(timestamp.AsTime()), opts)
	case durationPBDurationType:
		duration, _ := srcVal.Interface().(durationpb.Duration)
		return intCloner(g, labels, tgtVal, reflect.ValueOf(duration.AsDuration()), opts)

	case wrappersPBBoolType:
		return boolCloner(g, labels, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBInt32Type, wrappersPBInt64Type:
		return intCloner(g, labels, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBUint32Type, wrappersPBUint64Type:
		return uintCloner(g, labels, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBFloatType, wrappersPBDoubleType:
		return floatCloner(g, labels, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBStringType:
		return stringCloner(g, labels, tgtVal, srcVal.FieldByName("Value"), opts)
	case wrappersPBBytesType:
		return bytesCloner(g, labels, tgtVal, srcVal.FieldByName("Value"), opts)

	case anyPBAnyType:
		if srcVal.CanAddr() {
			srcVal = srcVal.Addr()
			srcPtr := srcVal.Interface().(*anypb.Any)
			message, err := srcPtr.UnmarshalNew()
			if err != nil {
				return newUnmarshalNewError(labels, tgtVal.Type(), srcVal.Type(), err)
			}
			return interfaceCloner(g, labels, tgtVal, reflect.ValueOf(message), opts)
		}
		return bytesCloner(g, labels, tgtVal, srcVal.FieldByName("Value"), opts)
	case emptyPBEmptyType:
		return nil

	case structPBStructType:
		structPB, _ := srcVal.Interface().(structpb.Struct)
		return mapCloner(g, labels, tgtVal, reflect.ValueOf(structPB.AsMap()), opts)
	case structPBListType:
		structPB, _ := srcVal.Interface().(structpb.ListValue)
		return sliceCloner(g, labels, tgtVal, reflect.ValueOf(structPB.AsSlice()), opts)
	case structPBValueType:
		valuePB, _ := srcVal.Interface().(structpb.Value)
		return interfaceCloner(g, labels, tgtVal, reflect.ValueOf(valuePB.AsInterface()), opts)
	case structPBNullValueType:
		return nil
	case structPBNumberValueType:
		numberPB, _ := srcVal.Interface().(structpb.Value_NumberValue)
		return floatCloner(g, labels, tgtVal, reflect.ValueOf(numberPB.NumberValue), opts)
	case structPBStringValueType:
		stringPB, _ := srcVal.Interface().(structpb.Value_StringValue)
		return stringCloner(g, labels, tgtVal, reflect.ValueOf(stringPB.StringValue), opts)
	case structPBBoolValueType:
		boolPB, _ := srcVal.Interface().(structpb.Value_BoolValue)
		return boolCloner(g, labels, tgtVal, reflect.ValueOf(boolPB.BoolValue), opts)
	case structPBStructValueType:
		structPB, _ := srcVal.Interface().(structpb.Value_StructValue)
		return mapCloner(g, labels, tgtVal, reflect.ValueOf(structPB.StructValue.AsMap()), opts)
	case structPBListValueType:
		listPB, _ := srcVal.Interface().(structpb.Value_ListValue)
		return sliceCloner(g, labels, tgtVal, reflect.ValueOf(listPB.ListValue.AsSlice()), opts)

	default:
		cloner, tv := indirect(tgtVal)
		if cloner != nil {
			ok, err := cloner.CloneFrom(srcVal.Interface())
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			tv = indirectValue(tv)
		}

		switch tv.Kind() {
		case reflect.Struct:
			return setStructToStruct(g, labels, tv, srcVal, opts)
		case reflect.Interface:
			return setStructToAny(g, labels, tv, srcVal, opts)
		case reflect.Map:
			return setStructToMap(g, labels, tv, srcVal, opts)
		case reflect.Pointer:
			return setPointer(g, labels, tv, srcVal, opts)
		default:
			return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
		}
	}
}

func unsupportedTypeCloner(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	for _, hook := range opts.Cloners {
		ok, err := hook(labels, tgtVal, srcVal)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return newUnsupportedTypeError(labels, tgtVal.Type(), srcVal.Type())
}
