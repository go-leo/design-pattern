package prototype

import (
	"database/sql"
	"encoding"
	"fmt"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"reflect"
	"time"
)

// ClonerFrom 自定义克隆方法，从源克隆到自己
type ClonerFrom interface {
	CloneFrom(src any) error
}

// ClonerTo 自定义克隆方法，将自己克隆到目标
type ClonerTo interface {
	CloneTo(tgt any) error
}

var (
	clonerFromType = reflect.TypeOf((*ClonerFrom)(nil)).Elem()
	clonerToType   = reflect.TypeOf((*ClonerTo)(nil)).Elem()

	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	stringerType      = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	scannerType         = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

	timeType = reflect.TypeOf(time.Time{})

	sqlNullBoolType    = reflect.TypeOf(sql.NullBool{})
	sqlNullByteType    = reflect.TypeOf(sql.NullByte{})
	sqlNullInt16Type   = reflect.TypeOf(sql.NullInt16{})
	sqlNullInt32Type   = reflect.TypeOf(sql.NullInt32{})
	sqlNullInt64Type   = reflect.TypeOf(sql.NullInt64{})
	sqlNullFloat64Type = reflect.TypeOf(sql.NullFloat64{})
	sqlNullStringType  = reflect.TypeOf(sql.NullString{})
	sqlNullTimeType    = reflect.TypeOf(sql.NullTime{})

	wrappersPBDoubleType = reflect.TypeOf(wrapperspb.DoubleValue{})
	wrappersPBFloatType  = reflect.TypeOf(wrapperspb.FloatValue{})
	wrappersPBInt32Type  = reflect.TypeOf(wrapperspb.Int32Value{})
	wrappersPBInt64Type  = reflect.TypeOf(wrapperspb.Int64Value{})
	wrappersPBUint32Type = reflect.TypeOf(wrapperspb.UInt32Value{})
	wrappersPBUint64Type = reflect.TypeOf(wrapperspb.UInt64Value{})
	wrappersPBBoolType   = reflect.TypeOf(wrapperspb.BoolValue{})
	wrappersPBStringType = reflect.TypeOf(wrapperspb.StringValue{})
	wrappersPBBytesType  = reflect.TypeOf(wrapperspb.BytesValue{})

	timestampPBTimestampType = reflect.TypeOf(timestamppb.Timestamp{})
	durationPBDurationType   = reflect.TypeOf(durationpb.Duration{})

	anyPBAnyType     = reflect.TypeOf(anypb.Any{})
	emptyPBEmptyType = reflect.TypeOf(emptypb.Empty{})

	structPBStructType      = reflect.TypeOf(structpb.Struct{})
	structPBValueType       = reflect.TypeOf(structpb.Value{})
	structPBNullValueType   = reflect.TypeOf(structpb.Value_NullValue{})
	structPBNumberValueType = reflect.TypeOf(structpb.Value_NumberValue{})
	structPBStringValueType = reflect.TypeOf(structpb.Value_StringValue{})
	structPBBoolValueType   = reflect.TypeOf(structpb.Value_BoolValue{})
	structPBStructValueType = reflect.TypeOf(structpb.Value_StructValue{})
	structPBListValueType   = reflect.TypeOf(structpb.Value_ListValue{})
)

var (
	intKinds   = []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64}
	uintKinds  = []reflect.Kind{reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr}
	floatKinds = []reflect.Kind{reflect.Float32, reflect.Float64}
)

func indirectValue(v reflect.Value) (ClonerFrom, reflect.Value) {
	for {
		if v.Type().NumMethod() > 0 && v.CanInterface() {
			if c, ok := v.Interface().(ClonerFrom); ok {
				return c, v
			}
		}
		if v.Kind() == reflect.Pointer && !v.IsNil() {
			v = v.Elem()
		} else {
			break
		}
	}
	return nil, v
}

func indirectType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

var boolMap = map[bool]int{
	false: 0,
	true:  1,
}
