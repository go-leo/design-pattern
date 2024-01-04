package prototype

import (
	"database/sql"
	"encoding/base64"
	"github.com/go-leo/gox/mathx"
	"golang.org/x/exp/slices"
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

func setInt(labels []string, tgtVal reflect.Value, i int64) error {
	if tgtVal.OverflowInt(i) {
		return newOverflowError(labels, tgtVal.Type(), strconv.FormatInt(i, 10))
	}
	tgtVal.SetInt(i)
	return nil
}

func setUint(labels []string, tgtVal reflect.Value, u uint64) error {
	if tgtVal.OverflowUint(u) {
		return newOverflowError(labels, tgtVal.Type(), strconv.FormatUint(u, 10))
	}
	tgtVal.SetUint(u)
	return nil
}

func setFloat(labels []string, tgtVal reflect.Value, f float64) error {
	if tgtVal.OverflowFloat(f) {
		return newOverflowError(labels, tgtVal.Type(), strconv.FormatFloat(f, 'f', -1, 64))
	}
	tgtVal.SetFloat(f)
	return nil
}

func setInt2Uint(labels []string, tgtVal reflect.Value, i int64) error {
	if i < 0 {
		return newNegativeNumberError(labels, tgtVal.Type(), strconv.FormatInt(i, 10))
	}
	return setUint(labels, tgtVal, uint64(i))
}

func setUint2Int(labels []string, tgtVal reflect.Value, u uint64) error {
	if u > uint64(math.MaxInt64) {
		return newOverflowError(labels, tgtVal.Type(), strconv.FormatUint(u, 10))
	}
	return setInt(labels, tgtVal, int64(u))
}

func setFloat2Int(labels []string, tgtVal reflect.Value, f float64) error {
	if f > float64(math.MaxInt64) || f < float64(math.MinInt64) {
		return newOverflowError(labels, tgtVal.Type(), strconv.FormatFloat(f, 'f', -1, 64))
	}
	return setInt(labels, tgtVal, int64(f))
}

func setFloat2Uint(labels []string, tgtVal reflect.Value, f float64) error {
	if f > float64(math.MaxUint64) {
		return newOverflowError(labels, tgtVal.Type(), strconv.FormatFloat(f, 'f', -1, 64))
	}
	if f < 0 {
		return newNegativeNumberError(labels, tgtVal.Type(), strconv.FormatFloat(f, 'f', -1, 64))
	}
	return setUint(labels, tgtVal, uint64(f))
}

func setPointer(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsNil() {
		return valueCloner(srcVal, opts)(g, labels, tgtVal, srcVal, opts)
	}
	tgtVal.Set(reflect.New(tgtVal.Type().Elem()))
	return valueCloner(srcVal, opts)(g, labels, tgtVal, srcVal, opts)
}

func setAnyValue(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, i any) error {
	if tgtVal.NumMethod() > 0 {
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
	tgtVal.Set(reflect.ValueOf(i))
	return nil
}

func setScanner(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, v any) error {
	if !tgtVal.CanAddr() {
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
	scanner := tgtVal.Addr().Interface().(sql.Scanner)
	return scanner.Scan(v)
}

func setAnyProtoBuf(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, v proto.Message) error {
	if !tgtVal.CanAddr() {
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
	tgtPtr := tgtVal.Addr().Interface().(*anypb.Any)
	return anypb.MarshalFrom(tgtPtr, v, proto.MarshalOptions{})
}

func setBoolToStruct(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, b bool) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullBoolType:
		return setScanner(g, labels, tgtVal, srcVal, opts, b)
	case sqlNullInt16Type, sqlNullInt32Type, sqlNullInt64Type, sqlNullByteType, sqlNullFloat64Type:
		return setScanner(g, labels, tgtVal, srcVal, opts, boolIntMap[b])
	case sqlNullStringType:
		return setScanner(g, labels, tgtVal, srcVal, opts, strconv.FormatBool(b))
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
		return setAnyProtoBuf(g, labels, tgtVal, srcVal, opts, &wrapperspb.BoolValue{Value: b})

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
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
}

func setIntToStruct(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, i int64) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullInt16Type:
		if i > math.MaxInt16 {
			return newOverflowError(labels, tgtType, strconv.FormatInt(i, 10))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullInt32Type:
		if i > math.MaxInt32 {
			return newOverflowError(labels, tgtType, strconv.FormatInt(i, 10))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullInt64Type:
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullByteType:
		if i > math.MaxUint8 {
			return newOverflowError(labels, tgtType, strconv.FormatInt(i, 10))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullFloat64Type:
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullBoolType:
		return setScanner(g, labels, tgtVal, srcVal, opts, i != 0)
	case sqlNullStringType:
		return setScanner(g, labels, tgtVal, srcVal, opts, strconv.FormatInt(i, 10))
	case sqlNullTimeType:
		return setScanner(g, labels, tgtVal, srcVal, opts, opts.IntToTime(i))

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
			return newOverflowError(labels, tgtType, strconv.FormatInt(i, 10))
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBInt64Type:
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBUint32Type:
		if i > math.MaxUint32 {
			return newOverflowError(labels, tgtType, strconv.FormatInt(i, 10))
		}
		if i < 0 {
			return newNegativeNumberError(labels, tgtType, strconv.FormatInt(i, 10))
		}
		tgtVal.FieldByName("Value").SetUint(uint64(i))
		return nil
	case wrappersPBUint64Type:
		if i < 0 {
			return newNegativeNumberError(labels, tgtType, strconv.FormatInt(i, 10))
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
		return setAnyProtoBuf(g, labels, tgtVal, srcVal, opts, &wrapperspb.Int64Value{Value: i})

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
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
}

func setUintToStruct(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, u uint64) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullByteType:
		if u > math.MaxUint8 {
			return newOverflowError(labels, tgtType, strconv.FormatUint(u, 10))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, u)
	case sqlNullInt16Type:
		if u > math.MaxInt16 {
			return newOverflowError(labels, tgtType, strconv.FormatUint(u, 10))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, u)
	case sqlNullInt32Type:
		if u > math.MaxInt32 {
			return newOverflowError(labels, tgtType, strconv.FormatUint(u, 10))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, u)
	case sqlNullInt64Type:
		if u > math.MaxInt64 {
			return newOverflowError(labels, tgtType, strconv.FormatUint(u, 10))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, u)
	case sqlNullFloat64Type:
		return setScanner(g, labels, tgtVal, srcVal, opts, u)
	case sqlNullBoolType:
		return setScanner(g, labels, tgtVal, srcVal, opts, u != 0)
	case sqlNullStringType:
		return setScanner(g, labels, tgtVal, srcVal, opts, strconv.FormatUint(u, 10))
	case sqlNullTimeType:
		return setScanner(g, labels, tgtVal, srcVal, opts, opts.IntToTime(int64(u)))

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
			return newOverflowError(labels, tgtType, strconv.FormatUint(u, 10))
		}
		tgtVal.FieldByName("Value").SetUint(u)
		return nil
	case wrappersPBUint64Type:
		tgtVal.FieldByName("Value").SetUint(u)
		return nil
	case wrappersPBInt32Type:
		if u > math.MaxInt32 {
			return newOverflowError(labels, tgtType, strconv.FormatUint(u, 10))
		}
		tgtVal.FieldByName("Value").SetInt(int64(u))
		return nil
	case wrappersPBInt64Type:
		if u > math.MaxInt64 {
			return newOverflowError(labels, tgtType, strconv.FormatUint(u, 10))
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
		return setAnyProtoBuf(g, labels, tgtVal, srcVal, opts, &wrapperspb.UInt64Value{Value: u})

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
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
}

func setFloatToStruct(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, f float64) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullFloat64Type:
		return setScanner(g, labels, tgtVal, srcVal, opts, f)
	case sqlNullInt16Type:
		if f > math.MaxInt16 {
			return newOverflowError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, f)
	case sqlNullInt32Type:
		if f > math.MaxInt32 {
			return newOverflowError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, f)
	case sqlNullInt64Type:
		if f > math.MaxInt64 {
			return newOverflowError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, f)

	case sqlNullByteType:
		if f > math.MaxUint8 {
			return newOverflowError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		if f < 0 {
			return newNegativeNumberError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, f)
	case sqlNullStringType:
		return setScanner(g, labels, tgtVal, srcVal, opts, strconv.FormatFloat(f, 'f', -1, 64))
	case sqlNullBoolType:
		return setScanner(g, labels, tgtVal, srcVal, opts, f != 0)
	case sqlNullTimeType:
		return setScanner(g, labels, tgtVal, srcVal, opts, opts.IntToTime(int64(f)))

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
			return newOverflowError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		tgtVal.FieldByName("Value").SetFloat(f)
		return nil
	case wrappersPBDoubleType:
		tgtVal.FieldByName("Value").SetFloat(f)
		return nil
	case wrappersPBInt32Type:
		if f > math.MaxInt32 {
			return newOverflowError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		tgtVal.FieldByName("Value").SetInt(int64(f))
		return nil
	case wrappersPBInt64Type:
		if f > math.MaxInt64 {
			return newOverflowError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		tgtVal.FieldByName("Value").SetInt(int64(f))
		return nil
	case wrappersPBUint32Type:
		if f > math.MaxUint32 {
			return newOverflowError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		if f < 0 {
			return newNegativeNumberError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		tgtVal.FieldByName("Value").SetUint(uint64(f))
		return nil
	case wrappersPBUint64Type:
		if f > math.MaxUint64 {
			return newOverflowError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
		}
		if f < 0 {
			return newNegativeNumberError(labels, tgtType, strconv.FormatFloat(f, 'f', -1, 64))
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
		return setAnyProtoBuf(g, labels, tgtVal, srcVal, opts, &wrapperspb.DoubleValue{Value: f})

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
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
}

func setStringToStruct(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, s string) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case sqlNullStringType:
		return setScanner(g, labels, tgtVal, srcVal, opts, s)
	case sqlNullInt16Type:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		if i > math.MaxInt16 {
			return newOverflowError(labels, tgtType, s)
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullInt32Type:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		if i > math.MaxInt32 {
			return newOverflowError(labels, tgtType, s)
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullInt64Type:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		if i > math.MaxInt64 {
			return newOverflowError(labels, tgtType, s)
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullByteType:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		if u > math.MaxUint8 {
			return newOverflowError(labels, tgtType, s)
		}
		if u < 0 {
			return newNegativeNumberError(labels, tgtType, s)
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, u)
	case sqlNullFloat64Type:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, f)
	case sqlNullBoolType:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, b)
	case sqlNullTimeType:
		t, err := opts.StringToTime(s)
		if err != nil {
			return newParseError(labels, tgtVal.Type(), s, err)
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, t)

	case timestampPBTimestampType:
		t, err := opts.StringToTime(s)
		if err != nil {
			return newParseError(labels, tgtVal.Type(), s, err)
		}
		tgtVal.FieldByName("Seconds").SetInt(t.Unix())
		tgtVal.FieldByName("Nanos").SetInt(int64(t.Nanosecond()))
		return nil
	case durationPBDurationType:
		d, err := time.ParseDuration(s)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
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
			return newParseError(labels, tgtType, s, err)
		}
		if i > math.MaxInt32 {
			return newOverflowError(labels, tgtType, s)
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBInt64Type:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		if i > math.MaxInt64 {
			return newOverflowError(labels, tgtType, s)
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBUint32Type:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		if u > math.MaxUint32 {
			return newOverflowError(labels, tgtType, s)
		}
		tgtVal.FieldByName("Value").SetUint(u)
		return nil
	case wrappersPBUint64Type:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		tgtVal.FieldByName("Value").SetUint(u)
		return nil
	case wrappersPBFloatType:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		if f > math.MaxFloat32 {
			return newOverflowError(labels, tgtType, s)
		}
		tgtVal.FieldByName("Value").SetFloat(f)
		return nil
	case wrappersPBDoubleType:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		tgtVal.FieldByName("Value").SetFloat(f)
		return nil
	case wrappersPBBoolType:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		tgtVal.FieldByName("Value").SetBool(b)
		return nil

	case emptyPBEmptyType:
		return nil
	case anyPBAnyType:
		return setAnyProtoBuf(g, labels, tgtVal, srcVal, opts, &wrapperspb.StringValue{Value: s})

	case structPBStringValueType:
		tgtVal.FieldByName("StringValue").SetString(s)
		return nil
	case structPBNumberValueType:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		tgtVal.FieldByName("NumberValue").SetFloat(f)
		return nil
	case structPBBoolValueType:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newParseError(labels, tgtType, s, err)
		}
		tgtVal.FieldByName("BoolValue").SetBool(b)
		return nil
	case structPBValueType:
		value, _ := structpb.NewValue(s)
		tgtVal.Set(reflect.ValueOf(value).Elem())
		return nil

	default:
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
}

func setBytesToStruct(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, bs []byte) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case wrappersPBBytesType:
		tgtVal.FieldByName("Value").SetBytes(bs)
		return nil
	case wrappersPBStringType:
		tgtVal.FieldByName("Value").SetString(base64.StdEncoding.EncodeToString(bs))
		return nil
	case anyPBAnyType:
		return setAnyProtoBuf(g, labels, tgtVal, srcVal, opts, &wrapperspb.BytesValue{Value: bs})
	case structPBStringValueType:
		tgtVal.FieldByName("StringValue").SetString(base64.StdEncoding.EncodeToString(bs))
		return nil
	case structPBValueType:
		value, _ := structpb.NewValue(bs)
		tgtVal.Set(reflect.ValueOf(value).Elem())
		return nil
	default:
		return setStringToStruct(g, labels, tgtVal, srcVal, opts, string(bs))
	}
}

func setTimeToStruct(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, t time.Time) error {
	tgtType := tgtVal.Type()
	switch tgtType {
	case timeType:
		tgtVal.Set(reflect.ValueOf(t))
		return nil

	case sqlNullTimeType:
		return setScanner(g, labels, tgtVal, srcVal, opts, t)
	case sqlNullStringType:
		return setScanner(g, labels, tgtVal, srcVal, opts, opts.TimeToString(t))
	case sqlNullInt16Type:
		i := opts.TimeToInt(t)
		if i > math.MaxInt16 {
			return newOverflowError(labels, tgtType, t.String())
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullInt32Type:
		i := opts.TimeToInt(t)
		if i > math.MaxInt32 {
			return newOverflowError(labels, tgtType, t.String())
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullInt64Type:
		i := opts.TimeToInt(t)
		if i > math.MaxInt64 {
			return newOverflowError(labels, tgtType, t.String())
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullByteType:
		i := opts.TimeToInt(t)
		if i > math.MaxUint8 {
			return newOverflowError(labels, tgtType, t.String())
		}
		if i < 0 {
			return newNegativeNumberError(labels, tgtType, t.String())
		}
		return setScanner(g, labels, tgtVal, srcVal, opts, i)
	case sqlNullFloat64Type:
		i := opts.TimeToInt(t)
		return setScanner(g, labels, tgtVal, srcVal, opts, i)

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
			return newOverflowError(labels, tgtType, t.String())
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBInt64Type:
		i := opts.TimeToInt(t)
		if i > math.MaxInt64 {
			return newOverflowError(labels, tgtType, t.String())
		}
		tgtVal.FieldByName("Value").SetInt(i)
		return nil
	case wrappersPBUint32Type:
		i := opts.TimeToInt(t)
		if i > math.MaxUint32 {
			return newOverflowError(labels, tgtType, t.String())
		}
		if i < 0 {
			return newNegativeNumberError(labels, tgtType, t.String())
		}
		tgtVal.FieldByName("Value").SetUint(uint64(i))
		return nil
	case wrappersPBUint64Type:
		i := opts.TimeToInt(t)
		if i < 0 {
			return newNegativeNumberError(labels, tgtType, t.String())
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
		return setAnyProtoBuf(g, labels, tgtVal, srcVal, opts, timestamppb.New(t))

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
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
}

func setSliceArray(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options) error {
	srcLen := srcVal.Len()
	if tgtVal.Kind() == reflect.Slice {
		tgtVal.Set(reflect.MakeSlice(tgtVal.Type(), srcLen, srcLen))
	}
	tgtLen := tgtVal.Len()
	minLen := mathx.Min(srcLen, tgtLen)
	elemCloner := typeCloner(srcVal.Type().Elem(), true, opts)
	for i := 0; i < minLen; i++ {
		if err := elemCloner(g, append(slices.Clone(labels), strconv.Itoa(i)), tgtVal.Index(i), srcVal.Index(i), opts); err != nil {
			return err
		}
	}
	return nil
}

func setAnySlice(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options) error {
	if tgtVal.NumMethod() > 0 {
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
	srcLen := srcVal.Len()
	// 创建一个切片
	tgtSlice := make([]any, srcLen)
	elemCloner := typeCloner(srcVal.Type().Elem(), true, opts)
	// 将src的元素逐个拷贝到tgt
	for i := 0; i < srcLen; i++ {
		if err := elemCloner(g, append(slices.Clone(labels), strconv.Itoa(i)), reflect.ValueOf(&tgtSlice[i]), srcVal.Index(i), opts); err != nil {
			return err
		}
	}
	// 设置tgtVal
	tgtVal.Set(reflect.ValueOf(tgtSlice))
	return nil
}

func setMapToMap(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options) error {
	tgtType := tgtVal.Type()
	if tgtVal.IsNil() {
		tgtVal.Set(reflect.MakeMap(tgtType))
	}

	srcType := srcVal.Type()
	srcKeyType := srcType.Key()
	srcValType := srcType.Elem()
	srcKeyCloner := typeCloner(srcKeyType, true, opts)
	srcValCloner := typeCloner(srcValType, true, opts)
	srcEntries, err := newMapEntries(srcVal.MapRange())
	if err != nil {
		return err
	}

	tgtKeyType := tgtType.Key()
	tgtValType := tgtType.Elem()

	for _, srcEntry := range srcEntries {
		tgtEntryKeyVal := reflect.New(tgtKeyType)
		if err := srcKeyCloner(g, append(slices.Clone(labels), srcEntry.Label), tgtEntryKeyVal, srcEntry.KeyVal, opts); err != nil {
			return err
		}
		tgtEntryKeyVal = tgtEntryKeyVal.Elem()
		if !tgtEntryKeyVal.IsValid() {
			continue
		}

		tgtEntryValVal := reflect.New(tgtValType)
		if err := srcValCloner(g, append(slices.Clone(labels), srcEntry.Label), tgtEntryValVal, srcEntry.ValVal, opts); err != nil {
			return err
		}
		tgtEntryValVal = tgtEntryValVal.Elem()
		if !tgtEntryValVal.IsValid() {
			continue
		}
		tgtVal.SetMapIndex(tgtEntryKeyVal, tgtEntryValVal)
	}

	return nil
}

func setMapToAny(g *cloneContext, labels []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options) error {
	if tgtVal.NumMethod() > 0 {
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}

	srcType := srcVal.Type()
	srcKeyType := srcType.Key()
	srcValType := srcType.Elem()
	srcKeyCloner := typeCloner(srcKeyType, true, opts)
	srcValCloner := typeCloner(srcValType, true, opts)

	m := make(map[any]any)
	srcEntries, err := newMapEntries(srcVal.MapRange())
	if err != nil {
		return err
	}
	for _, srcEntry := range srcEntries {
		entryLabels := append(slices.Clone(labels), srcEntry.Label)

		var tgtEntryKey any
		if err := srcKeyCloner(g, entryLabels, reflect.ValueOf(&tgtEntryKey), srcEntry.KeyVal, opts); err != nil {
			return err
		}

		var tgtEntryVal any
		if err := srcValCloner(g, entryLabels, reflect.ValueOf(&tgtEntryVal), srcEntry.ValVal, opts); err != nil {
			return err
		}
		m[tgtEntryKey] = tgtEntryVal
	}
	tgtVal.Set(reflect.ValueOf(m))
	return nil
}

func setMapToStruct(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	srcValCloner := typeCloner(srcVal.Type().Elem(), true, opts)
	srcEntries, err := newMapEntries(srcVal.MapRange())
	if err != nil {
		return err
	}
	setFunc := func(g *cloneContext, labels []string, srcVal reflect.Value, opts *options) func(in reflect.Value) error {
		return func(tgtVal reflect.Value) error {
			return srcValCloner(g, labels, tgtVal, srcVal, opts)
		}
	}

	tgtStruct := _CachedStructInfo(tgtVal.Type(), opts)
	for _, srcEntry := range srcEntries {
		field := tgtStruct.FindFieldByLabel(srcEntry.Label, opts)
		if field == nil {
			continue
		}
		if err := field.SetValue(tgtStruct, tgtVal, opts, setFunc(g, append(slices.Clone(labels), srcEntry.Label), srcEntry.ValVal, opts)); err != nil {
			return err
		}
	}
	return nil
}

func setStructToAny(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if tgtVal.NumMethod() > 0 {
		return unsupportedTypeCloner(g, labels, tgtVal, srcVal, opts)
	}
	m := make(map[any]any)
	srcType := srcVal.Type()
	srcStruct := _CachedStructInfo(srcType, opts)
	err := srcStruct.RangeFields(func(label string, srcField *_FieldInfo) error {
		srcFieldVal, err := srcField.GetValue(srcStruct, srcVal, opts)
		if err != nil {
			return err
		}
		if !srcFieldVal.IsValid() {
			return nil
		}
		if srcField.OmitEmpty && srcFieldVal.IsZero() {
			return nil
		}

		var tgtEntryVal any
		cloner := typeCloner(srcFieldVal.Type(), true, opts)
		if err := cloner(g, append(slices.Clone(labels), label),
			reflect.ValueOf(&tgtEntryVal), srcFieldVal, opts); err != nil {
			return err
		}
		m[label] = tgtEntryVal
		return nil
	})
	if err != nil {
		return err
	}
	tgtVal.Set(reflect.ValueOf(m))
	return nil
}

func setStructToMap(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	tgtType := tgtVal.Type()
	tgtKeyType := tgtType.Key()
	tgtValType := tgtType.Elem()

	if tgtVal.IsNil() {
		tgtVal.Set(reflect.MakeMap(tgtType))
	}

	srcType := srcVal.Type()
	srcStruct := _CachedStructInfo(srcType, opts)
	err := srcStruct.RangeFields(func(label string, srcField *_FieldInfo) error {
		srcFieldVal, err := srcField.GetValue(srcStruct, srcVal, opts)
		if err != nil {
			return err
		}
		if !srcFieldVal.IsValid() {
			return nil
		}
		if srcField.OmitEmpty && srcFieldVal.IsZero() {
			return nil
		}
		cloner := typeCloner(srcFieldVal.Type(), true, opts)

		entryLabels := append(slices.Clone(labels), label)
		tgtEntryKeyVal := reflect.New(tgtKeyType)
		if err := stringCloner(g, entryLabels, tgtEntryKeyVal, reflect.ValueOf(label), opts); err != nil {
			return err
		}
		tgtEntryKeyVal = tgtEntryKeyVal.Elem()
		if !tgtEntryKeyVal.IsValid() {
			return nil
		}

		tgtEntryValVal := reflect.New(tgtValType)
		if err := cloner(g, entryLabels, tgtEntryValVal, srcFieldVal, opts); err != nil {
			return err
		}
		tgtEntryValVal = tgtEntryValVal.Elem()
		if !tgtEntryValVal.IsValid() {
			return nil
		}

		tgtVal.SetMapIndex(tgtEntryKeyVal, tgtEntryValVal)
		return nil
	})
	return err
}

func setStructToStruct(g *cloneContext, labels []string, tgtVal, srcVal reflect.Value, opts *options) error {
	srcType := srcVal.Type()
	srcStruct := _CachedStructInfo(srcType, opts)

	tgtType := tgtVal.Type()
	tgtStruct := _CachedStructInfo(tgtType, opts)

	setFunc := func(g *cloneContext, labels []string, srcVal reflect.Value, opts *options) func(in reflect.Value) error {
		return func(tgtVal reflect.Value) error {
			cloner := valueCloner(srcVal, opts)
			return cloner(g, labels, tgtVal, srcVal, opts)
		}
	}

	tgtFields := make(map[*_FieldInfo]struct{})

	srcFunc := func(label string, srcField *_FieldInfo) error {
		srcFieldVal, err := srcField.GetValue(srcStruct, srcVal, opts)
		if err != nil {
			return err
		}
		if !srcFieldVal.IsValid() {
			return nil
		}
		if srcField.OmitEmpty && srcFieldVal.IsZero() {
			return nil
		}

		tgtField := tgtStruct.FindFieldByField(srcField, opts)
		if tgtField == nil {
			return nil
		}
		tgtFields[tgtField] = struct{}{}
		return tgtField.SetValue(tgtStruct, tgtVal, opts, setFunc(g, append(slices.Clone(labels), label), srcFieldVal, opts))
	}
	if err := srcStruct.RangeAllFields(srcFunc); err != nil {
		return err
	}

	tgtFunc := func(label string, tgtField *_FieldInfo) error {
		if _, ok := tgtFields[tgtField]; ok {
			return nil
		}
		tgtFields[tgtField] = struct{}{}

		srcField := srcStruct.FindFieldByField(tgtField, opts)
		if srcField == nil {
			return nil
		}
		srcFieldVal, err := srcField.GetValue(srcStruct, srcVal, opts)
		if err != nil {
			return err
		}
		if !srcFieldVal.IsValid() {
			return nil
		}
		if srcField.OmitEmpty && srcFieldVal.IsZero() {
			return nil
		}

		return tgtField.SetValue(tgtStruct, tgtVal, opts, setFunc(g, append(slices.Clone(labels), label), srcFieldVal, opts))
	}
	return tgtStruct.RangeAllFields(tgtFunc)
}
