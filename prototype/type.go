package prototype

import (
	"context"
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

var (
	clonerFromType = reflect.TypeOf((*ClonerFrom)(nil)).Elem()
	clonerToType   = reflect.TypeOf((*ClonerTo)(nil)).Elem()

	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	stringerType      = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	scannerType         = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

	timeType = reflect.TypeOf(time.Time{})

	sqlNullInt16Type   = reflect.TypeOf(sql.NullInt16{})
	sqlNullInt32Type   = reflect.TypeOf(sql.NullInt32{})
	sqlNullInt64Type   = reflect.TypeOf(sql.NullInt64{})
	sqlNullByteType    = reflect.TypeOf(sql.NullByte{})
	sqlNullFloat64Type = reflect.TypeOf(sql.NullFloat64{})
	sqlNullBoolType    = reflect.TypeOf(sql.NullBool{})
	sqlNullStringType  = reflect.TypeOf(sql.NullString{})
	sqlNullTimeType    = reflect.TypeOf(sql.NullTime{})

	timestampPBTimestampType = reflect.TypeOf(timestamppb.Timestamp{})
	durationPBDurationType   = reflect.TypeOf(durationpb.Duration{})

	wrappersPBBoolType   = reflect.TypeOf(wrapperspb.BoolValue{})
	wrappersPBInt32Type  = reflect.TypeOf(wrapperspb.Int32Value{})
	wrappersPBInt64Type  = reflect.TypeOf(wrapperspb.Int64Value{})
	wrappersPBUint32Type = reflect.TypeOf(wrapperspb.UInt32Value{})
	wrappersPBUint64Type = reflect.TypeOf(wrapperspb.UInt64Value{})
	wrappersPBFloatType  = reflect.TypeOf(wrapperspb.FloatValue{})
	wrappersPBDoubleType = reflect.TypeOf(wrapperspb.DoubleValue{})
	wrappersPBStringType = reflect.TypeOf(wrapperspb.StringValue{})
	wrappersPBBytesType  = reflect.TypeOf(wrapperspb.BytesValue{})

	emptyPBEmptyType = reflect.TypeOf(emptypb.Empty{})
	anyPBAnyType     = reflect.TypeOf(anypb.Any{})

	structPBStructType      = reflect.TypeOf(structpb.Struct{})
	structPBListType        = reflect.TypeOf(structpb.ListValue{})
	structPBValueType       = reflect.TypeOf(structpb.Value{})
	structPBNullValueType   = reflect.TypeOf(structpb.Value_NullValue{})
	structPBNumberValueType = reflect.TypeOf(structpb.Value_NumberValue{})
	structPBBoolValueType   = reflect.TypeOf(structpb.Value_BoolValue{})
	structPBStringValueType = reflect.TypeOf(structpb.Value_StringValue{})
	structPBStructValueType = reflect.TypeOf(structpb.Value_StructValue{})
	structPBListValueType   = reflect.TypeOf(structpb.Value_ListValue{})
)

var (
	intKinds       = []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64}
	uintKinds      = []reflect.Kind{reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr}
	floatKinds     = []reflect.Kind{reflect.Float32, reflect.Float64}
	stringKinds    = []reflect.Kind{reflect.String}
	boolKinds      = []reflect.Kind{reflect.Bool}
	allSampleKinds []reflect.Kind
)

func init() {
	allSampleKinds = append(allSampleKinds, intKinds...)
	allSampleKinds = append(allSampleKinds, uintKinds...)
	allSampleKinds = append(allSampleKinds, floatKinds...)
	allSampleKinds = append(allSampleKinds, stringKinds...)
	allSampleKinds = append(allSampleKinds, boolKinds...)
}

var boolIntMap = map[bool]int{
	false: 0,
	true:  1,
}
