package prototype

import (
	"database/sql"
	"encoding/base64"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"math"
	"reflect"
	"strconv"
	"time"
)

func setInt(fks []string, tgtVal reflect.Value, i int64) error {
	if tgtVal.OverflowInt(i) {
		return newOverflowError(fks, tgtVal.Type(), strconv.FormatInt(i, 10))
	}
	tgtVal.SetInt(i)
	return nil
}

func setUint(fks []string, tgtVal reflect.Value, u uint64) error {
	if tgtVal.OverflowUint(u) {
		return newOverflowError(fks, tgtVal.Type(), strconv.FormatUint(u, 10))
	}
	tgtVal.SetUint(u)
	return nil
}

func setFloat(fks []string, tgtVal reflect.Value, f float64) error {
	if tgtVal.OverflowFloat(f) {
		return newOverflowError(fks, tgtVal.Type(), strconv.FormatFloat(f, 'f', -1, 64))
	}
	tgtVal.SetFloat(f)
	return nil
}

func setInt2Uint(fks []string, tgtVal reflect.Value, i int64) error {
	if i < 0 {
		return newNegativeNumberError(fks, tgtVal.Type(), strconv.FormatInt(i, 10))
	}
	return setUint(fks, tgtVal, uint64(i))
}

func setUint2Int(fks []string, tgtVal reflect.Value, u uint64) error {
	if u > uint64(math.MaxInt64) {
		return newOverflowError(fks, tgtVal.Type(), strconv.FormatUint(u, 10))
	}
	return setInt(fks, tgtVal, int64(u))
}

func setFloat2Int(fks []string, tgtVal reflect.Value, f float64) error {
	if f > float64(math.MaxInt64) || f < float64(math.MinInt64) {
		return newOverflowError(fks, tgtVal.Type(), strconv.FormatFloat(f, 'f', -1, 64))
	}
	return setInt(fks, tgtVal, int64(f))
}

func setFloat2Uint(fks []string, tgtVal reflect.Value, f float64) error {
	if f > float64(math.MaxUint64) {
		return newOverflowError(fks, tgtVal.Type(), strconv.FormatFloat(f, 'f', -1, 64))
	}
	if f < 0 {
		return newNegativeNumberError(fks, tgtVal.Type(), strconv.FormatFloat(f, 'f', -1, 64))
	}
	return setUint(fks, tgtVal, uint64(f))
}

func setPointer(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, tv reflect.Value, cloner func(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options) error) error {
	if tv.IsNil() {
		tv.Set(reflect.New(tv.Type().Elem()))
		return cloner(e, fks, tv, srcVal, opts)
	}
	return cloner(e, fks, tv, srcVal, opts)
}

func setAnyValue(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, tv reflect.Value, i any) error {
	if tv.NumMethod() == 0 {
		tv.Set(reflect.ValueOf(i))
		return nil
	}
	return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
}

func setScanner(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, v any) error {
	if tgtVal.CanAddr() {
		scanner := tgtVal.Addr().Interface().(sql.Scanner)
		return scanner.Scan(v)
	}
	return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
}

func setAnyProtoBuf(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, v proto.Message) error {
	if tgtVal.CanAddr() {
		tgtPtr := tgtVal.Addr().Interface().(*anypb.Any)
		return anypb.MarshalFrom(tgtPtr, v, proto.MarshalOptions{})
	}
	return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
}

func setBoolStruct(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, b bool) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullBoolType:
		return setScanner(e, fks, tgtVal, srcVal, opts, b)
	case sqlNullInt16Type, sqlNullInt32Type, sqlNullInt64Type, sqlNullByteType, sqlNullFloat64Type:
		return setScanner(e, fks, tgtVal, srcVal, opts, boolIntMap[b])
	case sqlNullStringType:
		return setScanner(e, fks, tgtVal, srcVal, opts, strconv.FormatBool(b))
	case wrappersPBBoolType:
		tgtVal.FieldByName("Value").SetBool(b)
		return nil
	case wrappersPBInt32Type, wrappersPBInt64Type:
		tgtVal.FieldByName("Value").SetInt(int64(boolIntMap[b]))
		return nil
	case wrappersPBUint32Type, wrappersPBUint64Type:
		tgtVal.FieldByName("Value").SetUint(uint64(boolIntMap[b]))
		return nil
	case wrappersPBDoubleType, wrappersPBFloatType:
		tgtVal.FieldByName("Value").SetFloat(float64(boolIntMap[b]))
		return nil
	case wrappersPBStringType:
		tgtVal.FieldByName("Value").SetString(strconv.FormatBool(b))
		return nil
	case wrappersPBBytesType:
		tgtVal.FieldByName("Value").SetBytes([]byte(strconv.FormatBool(b)))
		return nil

	case emptyPBEmptyType:
		return nil
	case anyPBAnyType:
		return setAnyProtoBuf(e, fks, tgtVal, srcVal, opts, &wrapperspb.BoolValue{Value: b})

	case structPBBoolValueType:
		tgtVal.FieldByName("BoolValue").SetBool(b)
		return nil
	case structPBNumberValueType:
		tgtVal.FieldByName("NumberValue").SetFloat(float64(boolIntMap[b]))
		return nil
	case structPBStringValueType:
		tgtVal.FieldByName("StringValue").SetString(strconv.FormatBool(b))
		return nil
	case structPBValueType:
		value, _ := structpb.NewValue(b)
		tgtVal.Set(reflect.ValueOf(value).Elem())
		return nil
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func setIntStruct(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, i int64) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullInt16Type:
		if i > math.MaxInt16 {
			return newOverflowError(fks, tgtType, strconv.FormatInt(i, 10))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullInt32Type:
		if i > math.MaxInt32 {
			return newOverflowError(fks, tgtType, strconv.FormatInt(i, 10))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullInt64Type:
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullByteType:
		if i > math.MaxUint8 {
			return newOverflowError(fks, tgtType, strconv.FormatInt(i, 10))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullFloat64Type:
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullBoolType:
		return setScanner(e, fks, tgtVal, srcVal, opts, i != 0)
	case sqlNullStringType:
		return setScanner(e, fks, tgtVal, srcVal, opts, strconv.FormatInt(i, 10))
	case sqlNullTimeType:
		return setScanner(e, fks, tgtVal, srcVal, opts, opts.IntToTime(i))

	case timestampPBTimestampType:
		t := opts.IntToTime(i)
		tgtVal.FieldByName("Seconds").SetInt(t.Unix())
		tgtVal.FieldByName("Nanos").SetInt(int64(t.Nanosecond()))
		return nil
	case durationPBDurationType:
		d := time.Duration(i)
		nanos := d.Nanoseconds()
		secs := nanos / 1e9
		nanos -= secs * 1e9
		tgtVal.FieldByName("Seconds").SetInt(secs)
		tgtVal.FieldByName("Nanos").SetInt(nanos)
		return nil

	case wrappersPBInt32Type:
		if i > math.MaxInt32 {
			return newOverflowError(fks, tgtType, strconv.FormatInt(i, 10))
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBInt64Type:
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBUint32Type:
		if i > math.MaxUint32 {
			return newOverflowError(fks, tgtType, strconv.FormatInt(i, 10))
		}
		if i < 0 {
			return newNegativeNumberError(fks, tgtType, strconv.FormatInt(i, 10))
		}
		tgtVal.FieldByName("Value").SetUint(uint64(i))
		return nil
	case wrappersPBUint64Type:
		if i < 0 {
			return newNegativeNumberError(fks, tgtType, strconv.FormatInt(i, 10))
		}
		tgtVal.FieldByName("Value").SetUint(uint64(i))
		return nil
	case wrappersPBFloatType, wrappersPBDoubleType:
		tgtVal.FieldByName("Value").SetFloat(float64(i))
		return nil
	case wrappersPBBoolType:
		tgtVal.FieldByName("Value").SetBool(i != 0)
		return nil
	case wrappersPBStringType:
		tgtVal.FieldByName("Value").SetString(strconv.FormatInt(i, 10))
		return nil
	case wrappersPBBytesType:
		tgtVal.FieldByName("Value").SetBytes([]byte(strconv.FormatInt(i, 10)))
		return nil

	case emptyPBEmptyType:
		return nil
	case anyPBAnyType:
		return setAnyProtoBuf(e, fks, tgtVal, srcVal, opts, &wrapperspb.Int64Value{Value: i})

	case structPBNumberValueType:
		tgtVal.FieldByName("NumberValue").SetFloat(float64(i))
		return nil
	case structPBBoolValueType:
		tgtVal.FieldByName("BoolValue").SetBool(i != 0)
		return nil
	case structPBValueType:
		value, _ := structpb.NewValue(i)
		tgtVal.Set(reflect.ValueOf(value).Elem())
		return nil
	case structPBStringValueType:
		tgtVal.FieldByName("StringValue").SetString(strconv.FormatInt(i, 10))
		return nil
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func setUintStruct(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, u uint64) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullByteType:
		if u > math.MaxUint8 {
			return newOverflowError(fks, tgtType, strconv.FormatUint(u, 10))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, u)
	case sqlNullInt16Type:
		if u > math.MaxInt16 {
			return newOverflowError(fks, tgtType, strconv.FormatUint(u, 10))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, u)
	case sqlNullInt32Type:
		if u > math.MaxInt32 {
			return newOverflowError(fks, tgtType, strconv.FormatUint(u, 10))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, u)
	case sqlNullInt64Type:
		if u > math.MaxInt64 {
			return newOverflowError(fks, tgtType, strconv.FormatUint(u, 10))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, u)
	case sqlNullFloat64Type:
		return setScanner(e, fks, tgtVal, srcVal, opts, u)
	case sqlNullBoolType:
		return setScanner(e, fks, tgtVal, srcVal, opts, u != 0)
	case sqlNullStringType:
		return setScanner(e, fks, tgtVal, srcVal, opts, strconv.FormatUint(u, 10))
	case sqlNullTimeType:
		return setScanner(e, fks, tgtVal, srcVal, opts, opts.IntToTime(int64(u)))

	case timestampPBTimestampType:
		t := opts.IntToTime(int64(u))
		tgtVal.FieldByName("Seconds").SetInt(t.Unix())
		tgtVal.FieldByName("Nanos").SetInt(int64(t.Nanosecond()))
		return nil
	case durationPBDurationType:
		d := time.Duration(u)
		nanos := d.Nanoseconds()
		secs := nanos / 1e9
		nanos -= secs * 1e9
		tgtVal.FieldByName("Seconds").SetInt(secs)
		tgtVal.FieldByName("Nanos").SetInt(nanos)
		return nil

	case wrappersPBUint32Type:
		if u > math.MaxUint32 {
			return newOverflowError(fks, tgtType, strconv.FormatUint(u, 10))
		}
		tgtVal.FieldByName("Value").SetUint(u)
		return nil
	case wrappersPBUint64Type:
		tgtVal.FieldByName("Value").SetUint(u)
		return nil
	case wrappersPBInt32Type:
		if u > math.MaxInt32 {
			return newOverflowError(fks, tgtType, strconv.FormatUint(u, 10))
		}
		tgtVal.FieldByName("Value").SetInt(int64(u))
		return nil
	case wrappersPBInt64Type:
		if u > math.MaxInt64 {
			return newOverflowError(fks, tgtType, strconv.FormatUint(u, 10))
		}
		tgtVal.FieldByName("Value").SetInt(int64(u))
		return nil
	case wrappersPBFloatType, wrappersPBDoubleType:
		tgtVal.FieldByName("Value").SetFloat(float64(u))
		return nil
	case wrappersPBBoolType:
		tgtVal.FieldByName("Value").SetBool(u != 0)
		return nil
	case wrappersPBStringType:
		tgtVal.FieldByName("Value").SetString(strconv.FormatUint(u, 10))
		return nil
	case wrappersPBBytesType:
		tgtVal.FieldByName("Value").SetBytes([]byte(strconv.FormatUint(u, 10)))
		return nil

	case emptyPBEmptyType:
		return nil
	case anyPBAnyType:
		return setAnyProtoBuf(e, fks, tgtVal, srcVal, opts, &wrapperspb.UInt64Value{Value: u})

	case structPBNumberValueType:
		tgtVal.FieldByName("NumberValue").SetFloat(float64(u))
		return nil
	case structPBBoolValueType:
		tgtVal.FieldByName("BoolValue").SetBool(u != 0)
		return nil
	case structPBValueType:
		value, _ := structpb.NewValue(u)
		tgtVal.Set(reflect.ValueOf(value).Elem())
		return nil
	case structPBStringValueType:
		tgtVal.FieldByName("StringValue").SetString(strconv.FormatUint(u, 10))
		return nil
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func setFloatStruct(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, f float64) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullFloat64Type:
		return setScanner(e, fks, tgtVal, srcVal, opts, f)
	case sqlNullInt16Type:
		if f > math.MaxInt16 {
			return newOverflowError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, f)
	case sqlNullInt32Type:
		if f > math.MaxInt32 {
			return newOverflowError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, f)
	case sqlNullInt64Type:
		if f > math.MaxInt64 {
			return newOverflowError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, f)

	case sqlNullByteType:
		if f > math.MaxUint8 {
			return newOverflowError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		if f < 0 {
			return newNegativeNumberError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, f)
	case sqlNullStringType:
		return setScanner(e, fks, tgtVal, srcVal, opts, strconv.FormatFloat(f, 'f', -1, 64))
	case sqlNullBoolType:
		return setScanner(e, fks, tgtVal, srcVal, opts, f != 0)
	case sqlNullTimeType:
		return setScanner(e, fks, tgtVal, srcVal, opts, opts.IntToTime(int64(f)))

	case timestampPBTimestampType:
		t := opts.IntToTime(int64(f))
		tgtVal.FieldByName("Seconds").SetInt(t.Unix())
		tgtVal.FieldByName("Nanos").SetInt(int64(t.Nanosecond()))
		return nil
	case durationPBDurationType:
		d := time.Duration(f)
		nanos := d.Nanoseconds()
		secs := nanos / 1e9
		nanos -= secs * 1e9
		tgtVal.FieldByName("Seconds").SetInt(secs)
		tgtVal.FieldByName("Nanos").SetInt(nanos)
		return nil

	case wrappersPBFloatType:
		if f > math.MaxFloat32 {
			return newOverflowError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		tgtVal.FieldByName("Value").SetFloat(f)
		return nil
	case wrappersPBDoubleType:
		tgtVal.FieldByName("Value").SetFloat(f)
		return nil
	case wrappersPBInt32Type:
		if f > math.MaxInt32 {
			return newOverflowError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		tgtVal.FieldByName("Value").SetInt(int64(f))
		return nil
	case wrappersPBInt64Type:
		if f > math.MaxInt64 {
			return newOverflowError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		tgtVal.FieldByName("Value").SetInt(int64(f))
		return nil
	case wrappersPBUint32Type:
		if f > math.MaxUint32 {
			return newOverflowError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		if f < 0 {
			return newNegativeNumberError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		tgtVal.FieldByName("Value").SetUint(uint64(f))
		return nil
	case wrappersPBUint64Type:
		if f > math.MaxUint64 {
			return newOverflowError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		if f < 0 {
			return newNegativeNumberError(fks, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		tgtVal.FieldByName("Value").SetUint(uint64(f))
		return nil
	case wrappersPBBoolType:
		tgtVal.FieldByName("Value").SetBool(f != 0)
		return nil
	case wrappersPBStringType:
		tgtVal.FieldByName("Value").SetString(strconv.FormatFloat(f, 'f', -1, 64))
		return nil
	case wrappersPBBytesType:
		tgtVal.FieldByName("Value").SetBytes([]byte(strconv.FormatFloat(f, 'f', -1, 64)))
		return nil

	case emptyPBEmptyType:
		return nil
	case anyPBAnyType:
		return setAnyProtoBuf(e, fks, tgtVal, srcVal, opts, &wrapperspb.DoubleValue{Value: f})

	case structPBNumberValueType:
		tgtVal.FieldByName("NumberValue").SetFloat(f)
		return nil
	case structPBBoolValueType:
		tgtVal.FieldByName("BoolValue").SetBool(f != 0)
		return nil
	case structPBValueType:
		value, _ := structpb.NewValue(f)
		tgtVal.Set(reflect.ValueOf(value).Elem())
		return nil
	case structPBStringValueType:
		tgtVal.FieldByName("StringValue").SetString(strconv.FormatFloat(f, 'f', -1, 64))
		return nil
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func setStringStruct(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, s string) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullStringType:
		return setScanner(e, fks, tgtVal, srcVal, opts, s)
	case sqlNullInt16Type:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		if i > math.MaxInt16 {
			return newOverflowError(fks, tgtType, s)
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullInt32Type:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		if i > math.MaxInt32 {
			return newOverflowError(fks, tgtType, s)
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullInt64Type:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		if i > math.MaxInt64 {
			return newOverflowError(fks, tgtType, s)
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullByteType:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		if u > math.MaxUint8 {
			return newOverflowError(fks, tgtType, s)
		}
		if u < 0 {
			return newNegativeNumberError(fks, tgtType, s)
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, u)
	case sqlNullFloat64Type:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, f)
	case sqlNullBoolType:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, b)
	case sqlNullTimeType:
		return setScanner(e, fks, tgtVal, srcVal, opts, opts.StringToTime(s))

	case timestampPBTimestampType:
		t := opts.StringToTime(s)
		tgtVal.FieldByName("Seconds").SetInt(t.Unix())
		tgtVal.FieldByName("Nanos").SetInt(int64(t.Nanosecond()))
		return nil
	case durationPBDurationType:
		d, err := time.ParseDuration(s)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		nanos := d.Nanoseconds()
		secs := nanos / 1e9
		nanos -= secs * 1e9
		tgtVal.FieldByName("Seconds").SetInt(secs)
		tgtVal.FieldByName("Nanos").SetInt(nanos)
		return nil

	case wrappersPBStringType:
		tgtVal.FieldByName("Value").SetString(s)
		return nil
	case wrappersPBBytesType:
		tgtVal.FieldByName("Value").SetBytes([]byte(s))
		return nil
	case wrappersPBInt32Type:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		if i > math.MaxInt32 {
			return newOverflowError(fks, tgtType, s)
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBInt64Type:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		if i > math.MaxInt64 {
			return newOverflowError(fks, tgtType, s)
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBUint32Type:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		if u > math.MaxUint32 {
			return newOverflowError(fks, tgtType, s)
		}
		tgtVal.FieldByName("Value").SetUint(u)
		return nil
	case wrappersPBUint64Type:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		tgtVal.FieldByName("Value").SetUint(u)
		return nil
	case wrappersPBFloatType:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		if f > math.MaxFloat32 {
			return newOverflowError(fks, tgtType, s)
		}
		tgtVal.FieldByName("Value").SetFloat(f)
		return nil
	case wrappersPBDoubleType:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		tgtVal.FieldByName("Value").SetFloat(f)
		return nil
	case wrappersPBBoolType:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		tgtVal.FieldByName("Value").SetBool(b)
		return nil

	case emptyPBEmptyType:
		return nil
	case anyPBAnyType:
		return setAnyProtoBuf(e, fks, tgtVal, srcVal, opts, &wrapperspb.StringValue{Value: s})

	case structPBStringValueType:
		tgtVal.FieldByName("StringValue").SetString(s)
		return nil
	case structPBNumberValueType:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		tgtVal.FieldByName("NumberValue").SetFloat(f)
		return nil
	case structPBBoolValueType:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newStringParseError(fks, tgtType, s, err)
		}
		tgtVal.FieldByName("BoolValue").SetBool(b)
		return nil
	case structPBValueType:
		value, _ := structpb.NewValue(s)
		tgtVal.Set(reflect.ValueOf(value).Elem())
		return nil

	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func setBytesStruct(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, bs []byte) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case wrappersPBBytesType:
		tgtVal.FieldByName("Value").SetBytes(bs)
		return nil
	case wrappersPBStringType:
		tgtVal.FieldByName("Value").SetString(base64.StdEncoding.EncodeToString(bs))
		return nil
	case anyPBAnyType:
		return setAnyProtoBuf(e, fks, tgtVal, srcVal, opts, &wrapperspb.BytesValue{Value: bs})
	case structPBStringValueType:
		tgtVal.FieldByName("StringValue").SetString(base64.StdEncoding.EncodeToString(bs))
		return nil
	case structPBValueType:
		value, _ := structpb.NewValue(bs)
		tgtVal.Set(reflect.ValueOf(value).Elem())
		return nil
	default:
		return setStringStruct(e, fks, tgtVal, srcVal, opts, string(bs))
	}
}

func setTimeStruct(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, t time.Time) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullTimeType:
		return setScanner(e, fks, tgtVal, srcVal, opts, t)

	case sqlNullStringType:
		return setScanner(e, fks, tgtVal, srcVal, opts, opts.TimeToString(t))
	case sqlNullInt16Type:
		i := opts.TimeToInt(t)
		if i > math.MaxInt16 {
			return newOverflowError(fks, tgtType, t.String())
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullInt32Type:
		i := opts.TimeToInt(t)
		if i > math.MaxInt32 {
			return newOverflowError(fks, tgtType, t.String())
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullInt64Type:
		i := opts.TimeToInt(t)
		if i > math.MaxInt64 {
			return newOverflowError(fks, tgtType, t.String())
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullByteType:
		i := opts.TimeToInt(t)
		if i > math.MaxUint8 {
			return newOverflowError(fks, tgtType, t.String())
		}
		if i < 0 {
			return newNegativeNumberError(fks, tgtType, t.String())
		}
		return setScanner(e, fks, tgtVal, srcVal, opts, i)
	case sqlNullFloat64Type:
		i := opts.TimeToInt(t)
		return setScanner(e, fks, tgtVal, srcVal, opts, i)

	case timestampPBTimestampType:
		tgtVal.FieldByName("Seconds").SetInt(t.Unix())
		tgtVal.FieldByName("Nanos").SetInt(int64(t.Nanosecond()))
		return nil

	case wrappersPBStringType:
		tgtVal.FieldByName("Value").SetString(opts.TimeToString(t))
		return nil
	case wrappersPBBytesType:
		tgtVal.FieldByName("Value").SetBytes([]byte(opts.TimeToString(t)))
		return nil
	case wrappersPBInt32Type:
		i := opts.TimeToInt(t)
		if i > math.MaxInt32 {
			return newOverflowError(fks, tgtType, t.String())
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBInt64Type:
		i := opts.TimeToInt(t)
		if i > math.MaxInt64 {
			return newOverflowError(fks, tgtType, t.String())
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBUint32Type:
		i := opts.TimeToInt(t)
		if i > math.MaxUint32 {
			return newOverflowError(fks, tgtType, t.String())
		}
		if i < 0 {
			return newNegativeNumberError(fks, tgtType, t.String())
		}
		tgtVal.FieldByName("Value").SetUint(uint64(i))
		return nil
	case wrappersPBUint64Type:
		i := opts.TimeToInt(t)
		if i < 0 {
			return newNegativeNumberError(fks, tgtType, t.String())
		}
		tgtVal.FieldByName("Value").SetUint(uint64(i))
		return nil
	case wrappersPBFloatType, wrappersPBDoubleType:
		i := opts.TimeToInt(t)
		tgtVal.FieldByName("Value").SetFloat(float64(i))
		return nil

	case emptyPBEmptyType:
		return nil
	case anyPBAnyType:
		return setAnyProtoBuf(e, fks, tgtVal, srcVal, opts, timestamppb.New(t))

	case structPBStringValueType:
		tgtVal.FieldByName("StringValue").SetString(opts.TimeToString(t))
		return nil
	case structPBNumberValueType:
		i := opts.TimeToInt(t)
		tgtVal.FieldByName("NumberValue").SetFloat(float64(i))
		return nil
	case structPBValueType:
		value, _ := structpb.NewValue(opts.TimeToString(t))
		tgtVal.Set(reflect.ValueOf(value).Elem())
		return nil
	default:
		return unsupportedTypeCloner(e, fks, tgtVal, srcVal, opts)
	}
}
