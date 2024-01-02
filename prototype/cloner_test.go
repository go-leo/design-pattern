package prototype_test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"github.com/go-leo/design-pattern/prototype"
	"github.com/go-leo/gox/errorx"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

type testClonerFromString string

func (s *testClonerFromString) CloneFrom(src any) error {
	switch v := src.(type) {
	case nil:
		*s = "nil"
	case bool:
		*s = testClonerFromString(strconv.FormatBool(v))
	case int64:
		*s = testClonerFromString(strconv.FormatInt(v, 10))
	case uint64:
		*s = testClonerFromString(strconv.FormatUint(v, 10))
	case float64:
		*s = testClonerFromString(strconv.FormatFloat(v, 'g', -1, 64))
	case string:
		*s = testClonerFromString(v)
	}
	return nil
}

func TestClonerFrom(t *testing.T) {
	var err error
	var tgtClonerFrom testClonerFromString

	srcBool := true
	tgtClonerFrom = ""

	err = prototype.Clone(&tgtClonerFrom, srcBool)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatBool(srcBool), tgtClonerFrom)

	var srcInt = math.MaxInt
	tgtClonerFrom = ""
	err = prototype.Clone(&tgtClonerFrom, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatInt(int64(srcInt), 10), tgtClonerFrom)

	var srcUint uint = math.MaxUint
	tgtClonerFrom = ""
	err = prototype.Clone(&tgtClonerFrom, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatUint(uint64(srcUint), 10), string(tgtClonerFrom))

	var srcFloat32 float32 = math.MaxFloat32
	tgtClonerFrom = ""
	err = prototype.Clone(&tgtClonerFrom, srcFloat32)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatFloat(float64(srcFloat32), 'g', -1, 64), tgtClonerFrom)

	var srcFloat64 = math.MaxFloat64
	tgtClonerFrom = ""
	err = prototype.Clone(&tgtClonerFrom, srcFloat64)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatFloat(srcFloat64, 'g', -1, 64), tgtClonerFrom)

	var srcString = "hello prototype"
	tgtClonerFrom = ""
	err = prototype.Clone(&tgtClonerFrom, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, srcString, tgtClonerFrom)

}

type testClonerToString string

func (s *testClonerToString) CloneTo(tgt any) error {
	switch tgt := tgt.(type) {
	case *bool:
		b, err := strconv.ParseBool(string(*s))
		if err != nil {
			return err
		}
		*tgt = b
	case *int64:
		i, err := strconv.ParseInt(string(*s), 10, 64)
		if err != nil {
			return err
		}
		*tgt = i
	case *uint64:
		u, err := strconv.ParseUint(string(*s), 10, 64)
		if err != nil {
			return err
		}
		*tgt = u
	case *float64:
		f, err := strconv.ParseFloat(string(*s), 64)
		if err != nil {
			return err
		}
		*tgt = f
	case *string:
		*tgt = string(*s)
	case *testClonerToString:
		*tgt = *s
	}
	return nil
}

type testClonerToStruct struct {
	S testClonerToString
}

func TestClonerTo(t *testing.T) {
	var err error
	var srcClonerTo testClonerToString

	srcClonerTo = "true"
	var tgtBool bool
	err = prototype.Clone(&tgtBool, srcClonerTo)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, strconv.FormatBool(tgtBool))

	srcClonerTo = testClonerToString(strconv.FormatInt(math.MaxInt64, 10))
	var tgtInt64 int64
	err = prototype.Clone(&tgtInt64, srcClonerTo)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, strconv.FormatInt(tgtInt64, 10))

	srcClonerTo = testClonerToString(strconv.FormatUint(math.MaxUint, 10))
	var tgtUint64 uint64
	err = prototype.Clone(&tgtUint64, srcClonerTo)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, strconv.FormatUint(tgtUint64, 10))

	srcClonerTo = testClonerToString(strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64))
	var tgtFloat64 float64
	err = prototype.Clone(&tgtFloat64, srcClonerTo)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, strconv.FormatFloat(tgtFloat64, 'f', -1, 64))

	srcClonerTo = "hello prototype"
	var tgtString string
	err = prototype.Clone(&tgtString, srcClonerTo)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, tgtString)

	srcClonerToStruct := testClonerToStruct{S: srcClonerTo}
	var tgtClonerToStruct testClonerToStruct
	err = prototype.Clone(&tgtClonerToStruct, srcClonerToStruct)
}

type testTextUnmarshaler struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (tu *testTextUnmarshaler) UnmarshalText(text []byte) error {
	split := strings.Split(string(text), ",")
	tu.ID, _ = strconv.Atoi(split[0])
	tu.Name = split[1]
	return nil
}

func TestTextUnmarshaler(t *testing.T) {
	expected := testTextUnmarshaler{
		ID:   7,
		Name: "prototype",
	}
	var err error

	srcJsonStr := "7,prototype"
	var tgtTextUnmarshaler *testTextUnmarshaler
	err = prototype.Clone(&tgtTextUnmarshaler, srcJsonStr)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, *tgtTextUnmarshaler)

	srcJsonBytes := []byte("7,prototype")
	var tgtTextUnmarshalerByte *testTextUnmarshaler
	err = prototype.Clone(&tgtTextUnmarshalerByte, srcJsonBytes)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, *tgtTextUnmarshalerByte)

}

func TestBoolCloner(t *testing.T) {
	var err error

	var srcBool bool
	var tgtBool bool

	err = prototype.Clone(&tgtBool, srcBool)
	assert.NoError(t, err)
	assert.Equal(t, srcBool, tgtBool)

	srcBool = true
	tgtBool = false
	err = prototype.Clone(&tgtBool, srcBool)
	assert.NoError(t, err)
	assert.Equal(t, srcBool, tgtBool)

	var tgtErr error
	err = prototype.Clone(&tgtErr, srcBool)
	var utErr prototype.Error
	assert.ErrorAs(t, err, &utErr)

	var tgtSqlNullBool sql.NullBool
	err = prototype.Clone(&tgtSqlNullBool, srcBool)
	assert.NoError(t, err)
	assert.Equal(t, tgtBool, tgtSqlNullBool.Bool)

	var tgtSqlNullBoolPtr *sql.NullBool
	err = prototype.Clone(&tgtSqlNullBoolPtr, srcBool)
	assert.NoError(t, err)
	assert.Equal(t, tgtBool, tgtSqlNullBoolPtr.Bool)

	var tgtSqlNullInt64 sql.NullInt64
	err = prototype.Clone(&tgtSqlNullInt64, srcBool)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, tgtSqlNullInt64.Int64)

	var tgtSqlNullString sql.NullString
	err = prototype.Clone(&tgtSqlNullString, srcBool)
	assert.NoError(t, err)
	assert.EqualValues(t, "true", tgtSqlNullString.String)

	var tgtWrappersPBBool wrapperspb.BoolValue
	err = prototype.Clone(&tgtWrappersPBBool, srcBool)
	assert.NoError(t, err)
	assert.EqualValues(t, true, tgtWrappersPBBool.Value)

	var tgtWrappersPBBoolPtr *wrapperspb.BoolValue
	err = prototype.Clone(&tgtWrappersPBBoolPtr, srcBool)
	assert.NoError(t, err)
	assert.EqualValues(t, true, tgtWrappersPBBoolPtr.Value)

	var tgtAnypb anypb.Any
	err = prototype.Clone(&tgtAnypb, srcBool)
	assert.NoError(t, err)
	m := wrapperspb.BoolValue{}
	err = tgtAnypb.UnmarshalTo(&m)
	assert.NoError(t, err)
	assert.EqualValues(t, true, m.Value)

	var structPBBoolValue structpb.Value_BoolValue
	err = prototype.Clone(&structPBBoolValue, srcBool)
	assert.NoError(t, err)
	assert.EqualValues(t, true, structPBBoolValue.BoolValue)

	var structPBValue structpb.Value
	err = prototype.Clone(&structPBValue, srcBool)
	assert.NoError(t, err)
	assert.EqualValues(t, true, structPBValue.GetBoolValue())

	var structPBNumberValue structpb.Value_NumberValue
	err = prototype.Clone(&structPBNumberValue, srcBool)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, structPBNumberValue.NumberValue)

	var structPBStringValue structpb.Value_StringValue
	err = prototype.Clone(&structPBStringValue, srcBool)
	assert.NoError(t, err)
	assert.EqualValues(t, "true", structPBStringValue.StringValue)

}

func TestIntCloner(t *testing.T) {
	var err error

	var srcInt = 1
	var tgtInt int

	err = prototype.Clone(&tgtInt, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, tgtInt)

	srcInt = 300
	var tgtInt8 uint8
	err = prototype.Clone(&tgtInt8, srcInt)
	var overflowErr prototype.Error
	assert.ErrorAs(t, err, &overflowErr)

	var tgtInt16 uint16
	err = prototype.Clone(&tgtInt16, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, tgtInt16)

	var tgtFloat32 float32
	err = prototype.Clone(&tgtFloat32, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, tgtFloat32)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, tgtAny)

	srcInt = math.MaxInt64

	sqlNullByte := sql.NullByte{}
	err = prototype.Clone(&sqlNullByte, srcInt)
	var e prototype.Error
	assert.ErrorAs(t, err, &e)

	sqlNullInt64 := sql.NullInt64{}
	err = prototype.Clone(&sqlNullInt64, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, sqlNullInt64.Int64)

	srcInt = int(time.Now().Unix())

	sqlNullTime := sql.NullTime{}
	err = prototype.Clone(&sqlNullTime, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, sqlNullTime.Time.Unix())

	var timestampPBTimestamp timestamppb.Timestamp
	err = prototype.Clone(&timestampPBTimestamp, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, timestampPBTimestamp.AsTime().Unix())

	srcInt = int(time.Hour)

	var durationPBDuration durationpb.Duration
	err = prototype.Clone(&durationPBDuration, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, durationPBDuration.AsDuration())

	srcInt = math.MinInt64

	var wrappersPBUint64 wrapperspb.UInt64Value
	err = prototype.Clone(&wrappersPBUint64, srcInt)
	e = prototype.Error{}
	assert.ErrorAs(t, err, &e)

	var wrappersPBBytes wrapperspb.BytesValue
	err = prototype.Clone(&wrappersPBBytes, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.Itoa(srcInt), string(wrappersPBBytes.Value))

	var anyPBAny anypb.Any
	err = prototype.Clone(&anyPBAny, srcInt)
	assert.NoError(t, err)
	var int64pb wrapperspb.Int64Value
	err = anyPBAny.UnmarshalTo(&int64pb)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, int64pb.Value)

	var structPBNumberValue structpb.Value_NumberValue
	err = prototype.Clone(&structPBNumberValue, srcInt)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(srcInt), structPBNumberValue.NumberValue)

}

func TestUIntCloner(t *testing.T) {
	var err error

	var srcUint uint = 1

	var tgtInt int
	err = prototype.Clone(&tgtInt, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, tgtInt)

	srcUint = 300
	var tgtInt8 uint8
	err = prototype.Clone(&tgtInt8, srcUint)
	var overflowErr prototype.Error
	assert.ErrorAs(t, err, &overflowErr)

	var tgtInt16 uint16
	err = prototype.Clone(&tgtInt16, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, tgtInt16)

	var tgtFloat32 float32
	err = prototype.Clone(&tgtFloat32, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, tgtFloat32)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, tgtAny)

	var sqlNullInt32 sql.NullInt32
	err = prototype.Clone(&sqlNullInt32, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, sqlNullInt32.Int32)

	var sqlNullFloat64 sql.NullFloat64
	err = prototype.Clone(&sqlNullFloat64, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(srcUint), sqlNullFloat64.Float64)

	srcUint = uint(time.Now().Unix())

	var sqlNullTime sql.NullTime
	err = prototype.Clone(&sqlNullTime, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, sqlNullTime.Time.Unix())

	var timestampPBTimestamp timestamppb.Timestamp
	err = prototype.Clone(&timestampPBTimestamp, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, timestampPBTimestamp.AsTime().Unix())

	srcUint = uint(time.Hour)

	var durationPBDuration durationpb.Duration
	err = prototype.Clone(&durationPBDuration, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, durationPBDuration.AsDuration())

	var wrappersPBDouble wrapperspb.DoubleValue
	err = prototype.Clone(&wrappersPBDouble, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(srcUint), wrappersPBDouble.Value)

	var wrappersPBString wrapperspb.StringValue
	err = prototype.Clone(&wrappersPBString, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatUint(uint64(srcUint), 10), wrappersPBString.Value)

	var anyPBAny anypb.Any
	err = prototype.Clone(&anyPBAny, srcUint)
	assert.NoError(t, err)
	uInt64Value := wrapperspb.UInt64Value{}
	err = anyPBAny.UnmarshalTo(&uInt64Value)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, uInt64Value.Value)

	var structPBValue structpb.Value
	err = prototype.Clone(&structPBValue, srcUint)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(srcUint), structPBValue.GetNumberValue())

}

func TestFloatCloner(t *testing.T) {
	var err error

	var srcFloat32 float32 = 1.1

	var tgtInt int
	err = prototype.Clone(&tgtInt, srcFloat32)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtInt)

	srcFloat32 = 300.5
	var tgtInt8 uint8
	err = prototype.Clone(&tgtInt8, srcFloat32)
	var overflowErr prototype.Error
	assert.ErrorAs(t, err, &overflowErr)

	var tgtInt16 uint16
	err = prototype.Clone(&tgtInt16, srcFloat32)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtInt16)

	var tgtFloat32 float32
	err = prototype.Clone(&tgtFloat32, srcFloat32)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtFloat32)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcFloat32)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtAny)

	var srcFloat64 = 120.4

	tgtInt = 0
	err = prototype.Clone(&tgtInt, srcFloat64)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtInt)

	srcFloat64 = 300.5
	tgtInt8 = 0
	err = prototype.Clone(&tgtInt8, srcFloat64)
	var e prototype.Error
	assert.ErrorAs(t, err, &e)

	tgtInt16 = 0
	err = prototype.Clone(&tgtInt16, srcFloat64)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtInt16)

	tgtFloat32 = 0
	err = prototype.Clone(&tgtFloat32, srcFloat64)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtFloat32)

	var tgtOtherAny any
	err = prototype.Clone(&tgtOtherAny, srcFloat64)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtOtherAny)

}

func TestStringCloner(t *testing.T) {
	var err error

	var srcString = "120.4"

	var tgtString string
	err = prototype.Clone(&tgtString, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, srcString, tgtString)

	var tgtBytes []byte
	err = prototype.Clone(&tgtBytes, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, srcString, tgtBytes)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, srcString, tgtAny)

	srcString = "true"
	var tgtBool bool
	err = prototype.Clone(&tgtBool, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, true, tgtBool)

	srcString = "-1000000000"
	var tgtInt int
	err = prototype.Clone(&tgtInt, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, -1000000000, tgtInt)

	srcString = "1000000000"
	var tgtUint uint
	err = prototype.Clone(&tgtUint, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, 1000000000, tgtUint)

	srcString = "3.1415836"
	var tgtFloat64 float64
	err = prototype.Clone(&tgtFloat64, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, 3.1415836)

	srcString = "3.1415836"
	var tgtStrPtr *string
	err = prototype.Clone(&tgtStrPtr, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, 3.1415836)

}

func TestBytesCloner(t *testing.T) {
	var err error

	var srcBytes = []byte("120.4")

	var tgtBytes []byte
	err = prototype.Clone(&tgtBytes, srcBytes)
	assert.NoError(t, err)
	assert.EqualValues(t, srcBytes, tgtBytes)

	var tgtString string
	err = prototype.Clone(&tgtString, srcBytes)
	assert.NoError(t, err)
	assert.EqualValues(t, string(base64.StdEncoding.EncodeToString(srcBytes)), tgtString)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcBytes)
	assert.NoError(t, err)
	assert.EqualValues(t, srcBytes, tgtAny)

	srcBytes = []byte("true")
	var tgtBool bool
	err = prototype.Clone(&tgtBool, srcBytes)
	assert.NoError(t, err)
	assert.EqualValues(t, true, tgtBool)

	srcBytes = []byte("-1000000000")
	var tgtInt int
	err = prototype.Clone(&tgtInt, srcBytes)
	assert.NoError(t, err)
	assert.EqualValues(t, -1000000000, tgtInt)

	srcBytes = []byte("1000000000")
	var tgtUint uint
	err = prototype.Clone(&tgtUint, srcBytes)
	assert.NoError(t, err)
	assert.EqualValues(t, 1000000000, tgtUint)

	srcBytes = []byte("3.1415836")
	var tgtFloat64 float64
	err = prototype.Clone(&tgtFloat64, srcBytes)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, 3.1415836)

	srcBytes = []byte("3.1415836")
	var tgtStrPtr *string
	err = prototype.Clone(&tgtStrPtr, srcBytes)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, 3.1415836)
}

func TestTimeCloner(t *testing.T) {
	var err error

	var srcTime = time.Now()

	var tgtStruct time.Time
	err = prototype.Clone(&tgtStruct, srcTime)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtStruct, srcTime)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcTime)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtAny, srcTime)

	var tgtString string
	err = prototype.Clone(&tgtString, srcTime)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtString, srcTime.Format(time.RFC3339))

	var tgtInt int
	err = prototype.Clone(&tgtInt, srcTime)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtInt, srcTime.Unix())

	var tgtUint uint
	err = prototype.Clone(&tgtUint, srcTime)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtUint, srcTime.Unix())

	var tgtFloat32 float32
	err = prototype.Clone(&tgtFloat32, srcTime)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat32, float32(srcTime.Unix()))

	var tgtPtr *time.Time
	err = prototype.Clone(&tgtPtr, srcTime)
	assert.NoError(t, err)
	assert.EqualValues(t, *tgtPtr, srcTime)

	var tgtPtrPtr **time.Time
	err = prototype.Clone(&tgtPtrPtr, srcTime)
	assert.NoError(t, err)
	assert.EqualValues(t, **tgtPtrPtr, srcTime)

}

func TestSliceCloner(t *testing.T) {
	var err error

	var srcBytesSlice = []byte{'1', '2', 'a', 'b'}
	var tgtString string
	err = prototype.Clone(&tgtString, srcBytesSlice)
	assert.NoError(t, err)
	assert.EqualValues(t, base64.StdEncoding.EncodeToString(srcBytesSlice), tgtString)

	var srcInt16Slice = []int16{math.MinInt16, math.MinInt8, math.MaxInt8, math.MaxInt16}
	var tgtInt8Slice []int8
	err = prototype.Clone(&tgtInt8Slice, srcInt16Slice)
	var ofErr prototype.Error
	assert.ErrorAs(t, err, &ofErr)

	var tgtInt32Slice []int32
	err = prototype.Clone(&tgtInt32Slice, srcInt16Slice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtInt32Slice, srcInt16Slice)

	var srcStringSlice = []string{"1", "2", "3", "4"}
	var tgtStringSlice []string
	err = prototype.Clone(&tgtStringSlice, srcStringSlice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtStringSlice, srcStringSlice)

	var tgtStringArray [10]string
	err = prototype.Clone(&tgtStringArray, srcStringSlice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtStringArray[:4])

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcStringSlice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtAny)

	var tgtStringArrayPtr *[10]string
	err = prototype.Clone(&tgtStringArrayPtr, srcStringSlice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtStringArrayPtr[:4])

}

func TestArrayCloner(t *testing.T) {
	var err error

	var srcInt16Slice = [...]int16{math.MinInt16, math.MinInt8, math.MaxInt8, math.MaxInt16}
	var tgtInt32Slice [3]int32
	err = prototype.Clone(&tgtInt32Slice, srcInt16Slice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcInt16Slice[:3], tgtInt32Slice[:])

	var tgtInt64Slice [6]int64
	err = prototype.Clone(&tgtInt64Slice, srcInt16Slice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcInt16Slice[:], tgtInt64Slice[:4])

	var srcStringSlice = [...]string{"1", "2", "3", "4"}
	var tgtStringSlice []string
	err = prototype.Clone(&tgtStringSlice, srcStringSlice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtStringSlice, srcStringSlice)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcStringSlice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtAny)

	var tgtStringArrayPtr *[10]string
	err = prototype.Clone(&tgtStringArrayPtr, srcStringSlice)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtStringArrayPtr[:4])

}

func TestInterfaceCloner(t *testing.T) {
	var err error

	var srcString = "120.4"

	var tgtString string
	err = prototype.Clone(&tgtString, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtString, srcString)

	var tgtBytes []byte
	err = prototype.Clone(&tgtBytes, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtBytes, srcString)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcString)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtAny, srcString)
}

func TestDuration(t *testing.T) {
	src := time.Hour
	var tgt time.Duration
	err := prototype.Clone(&tgt, src)
	assert.NoError(t, err)
	assert.EqualValues(t, src, tgt)
}

func TestRawBytes(t *testing.T) {
	src := sql.RawBytes("hello prototype")
	var tgt sql.RawBytes
	err := prototype.Clone(&tgt, src)
	assert.NoError(t, err)
	assert.EqualValues(t, src, tgt)
}

func TestMapCloner(t *testing.T) {
	var err error

	srcMap := map[string]any{
		"Msg":      "success",
		"Username": "Forest",
		"IsPass":   false,
		"Int":      int64(0),
		"Code": map[string]any{
			"Msg":  "ok",
			"Code": int64(200),
			"Error": map[string]any{
				"Course": "Course",
				"Msg":    "error msg",
			},
		},
		"Info": map[any]any{
			10:     "10",
			false:  "FALSE",
			"TRUE": true,
			23.4:   []int{2, 3, 4},
		},
		"nil": nil,
	}
	var tgtMap map[string]any
	err = prototype.Clone(&tgtMap, srcMap)
	assert.NoError(t, err)
	expected := string(errorx.Ignore(jsoniter.Marshal(srcMap)))
	actual := string(errorx.Ignore(jsoniter.Marshal(tgtMap)))
	assert.JSONEq(t, expected, actual)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcMap)
	assert.NoError(t, err)
	expected = string(errorx.Ignore(jsoniter.Marshal(srcMap)))
	actual = string(errorx.Ignore(jsoniter.Marshal(tgtAny)))
	assert.JSONEq(t, expected, actual)
}

func TestMapStructCloner(t *testing.T) {
	var err error
	var expected string
	var actual string

	srcMap := map[string]any{
		"Msg":      "success",
		"Username": "Forest",
		"IsPass":   false,
		"Int":      int64(0),
		"Code": map[string]any{
			"Msg":  "ok",
			"Code": int64(200),
			"Error": map[string]any{
				"Course": "Course",
				"Msg":    "error msg",
			},
		},
		"Info": map[any]any{
			10:     "10",
			false:  "FALSE",
			"TRUE": true,
			23.4:   []int{2, 3, 4},
		},
	}

	type Error struct {
		Course string
		Msg    string
	}

	type Code struct {
		Error *Error
		Msg   string
		Code  int
	}

	type Int int

	type Response struct {
		Code     Code
		Int      Int
		Msg      string
		Username string
		IsPass   bool
		Info     map[any]any
	}
	var tgtStruct Response
	err = prototype.Clone(&tgtStruct, srcMap)
	assert.NoError(t, err)
	expected = string(errorx.Ignore(jsoniter.Marshal(srcMap)))
	actual = string(errorx.Ignore(jsoniter.Marshal(tgtStruct)))
	assert.JSONEq(t, expected, actual)

	var tgtStructPointer *Response
	err = prototype.Clone(&tgtStructPointer, srcMap)
	assert.NoError(t, err)
	expected = string(errorx.Ignore(jsoniter.Marshal(srcMap)))
	actual = string(errorx.Ignore(jsoniter.Marshal(tgtStructPointer)))
	assert.JSONEq(t, expected, actual)
}

type testMapSetter struct {
	msg  string
	code int
}

func (x *testMapSetter) Code(c int) {
	x.code = 1000 + c
}

func (x *testMapSetter) Msg(_ context.Context, msg string) error {
	x.msg = "msg: " + msg
	return nil
}

func TestMapSetterCloner(t *testing.T) {
	srcMap := map[string]any{
		"msg":  "success",
		"code": 200,
	}
	var tgtMapSetter testMapSetter
	err := prototype.Clone(&tgtMapSetter, srcMap)
	assert.NoError(t, err)
	assert.EqualValues(t, testMapSetter{
		msg:  "msg: success",
		code: 1200,
	}, tgtMapSetter)
}

func TestSampleStructCloner(t *testing.T) {
	var err error

	var srcSqlNullBool sql.NullBool
	_ = srcSqlNullBool.Scan(true)
	var tgtBool bool
	err = prototype.Clone(&tgtBool, srcSqlNullBool)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtBool, srcSqlNullBool.Bool)

	var srcSqlNullInt64 sql.NullInt64
	_ = srcSqlNullInt64.Scan(math.MaxInt64)
	var tgtInt32 int32
	err = prototype.Clone(&tgtInt32, srcSqlNullInt64)
	var overflowErr prototype.Error
	assert.ErrorAs(t, err, &overflowErr)

	_ = srcSqlNullInt64.Scan(math.MinInt64)
	var tgtUint64 uint64
	err = prototype.Clone(&tgtUint64, srcSqlNullInt64)
	var negativeErr prototype.Error
	assert.ErrorAs(t, err, &negativeErr)

	var tgtInt64 int64
	err = prototype.Clone(&tgtInt64, srcSqlNullInt64)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtInt64, srcSqlNullInt64.Int64)

	var srcSqlNullByte sql.NullByte
	_ = srcSqlNullByte.Scan(math.MaxUint8)
	var tgtUint8 uint8
	err = prototype.Clone(&tgtUint8, srcSqlNullByte)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtUint8, srcSqlNullByte.Byte)

	var srcSqlNullFloat sql.NullFloat64
	_ = srcSqlNullFloat.Scan(math.MaxFloat64)
	err = prototype.Clone(&tgtUint8, srcSqlNullFloat)
	assert.ErrorAs(t, err, &overflowErr)

	var tgtFloat32 float32
	err = prototype.Clone(&tgtFloat32, srcSqlNullFloat)
	assert.ErrorAs(t, err, &overflowErr)

	var tgtFloat64 float64
	err = prototype.Clone(&tgtFloat64, srcSqlNullFloat)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, srcSqlNullFloat.Float64)

	now := time.Now()
	srcTimestampPB := timestamppb.New(now)
	var tgtTime time.Time
	err = prototype.Clone(&tgtTime, srcTimestampPB)
	assert.EqualValues(t, now.Local().Format(time.DateTime), tgtTime.Local().Format(time.DateTime))

	hour := time.Hour
	srcDurationPB := durationpb.New(hour)
	var tgtDuration time.Duration
	err = prototype.Clone(&tgtDuration, srcDurationPB)
	assert.EqualValues(t, hour, tgtDuration)

}

func TestCustomStructCloner(t *testing.T) {
	var err error

	type Error struct {
		Course string
		Msg    string
	}

	type Code struct {
		*Error
		Msg     string
		Code    int
		Details []string
	}

	type Int int

	type Response struct {
		Code
		Int
		Msg      string
		Username string
		IsPass   bool
	}

	var srcResponse = Response{
		Code: Code{
			Error: &Error{
				Course: "Course",
				Msg:    "error msg",
			},
			Msg:     "ok",
			Code:    200,
			Details: []string{"fast", "slow"},
		},
		Msg:      "success",
		Username: "Forest",
		IsPass:   false,
	}
	var tgtResp Response
	err = prototype.Clone(&tgtResp, srcResponse)
	assert.NoError(t, err)
	assert.EqualValues(t, &tgtResp, &srcResponse)

	var tgtAny any
	err = prototype.Clone(&tgtAny, srcResponse)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtAny, map[any]any{
		"Msg":      "success",
		"Username": "Forest",
		"IsPass":   false,
		"Int":      int64(0),
		"Code": map[any]any{
			"Msg":  "ok",
			"Code": int64(200),
			"Error": map[any]any{
				"Course": "Course",
				"Msg":    "error msg",
			},
			"Details": []any{"fast", "slow"},
		},
	})

	var tgtMap map[string]any
	err = prototype.Clone(&tgtMap, srcResponse)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtMap, map[string]any{
		"Msg":      "success",
		"Username": "Forest",
		"IsPass":   false,
		"Int":      int64(0),
		"Code": map[any]any{
			"Msg":  "ok",
			"Code": int64(200),
			"Error": map[any]any{
				"Course": "Course",
				"Msg":    "error msg",
			},
			"Details": []any{"fast", "slow"},
		},
	})
}

type testGetter struct {
	id      string
	name    string
	Age     int    `prototype:"age"`
	Address string `prototype:"address"`
}

func (t testGetter) Id() string {
	return "id:" + t.id
}

func (t testGetter) Name(ctx context.Context) (string, error) {
	return "name:" + t.name, nil
}

type testSetter struct {
	Id      string `prototype:"id"`
	Name    string `prototype:"name"`
	age     int
	address string
}

func (t *testSetter) SetAge(age int) {
	t.age = age * 2
}

func (t *testSetter) SetAddress(ctx context.Context, address string) error {
	t.address = "china-" + address
	return nil
}

func TestGetterSetterCloner(t *testing.T) {
	id := uuid.NewString()
	name := "prototype"
	age := 30
	address := "shanghai"
	src := testGetter{
		id:      id,
		name:    name,
		Age:     age,
		Address: address,
	}
	var tgt testSetter
	err := prototype.Clone(&tgt, src, prototype.TagKey("prototype"), prototype.SetterPrefix("Set"))
	assert.NoError(t, err)
	assert.EqualValues(t, testSetter{
		Id:      "id:" + id,
		Name:    "name:" + name,
		age:     age * 2,
		address: "china-" + address,
	}, tgt)
}

type testSameLabel struct {
	A    int
	AA   int `prototype:"A"`
	AAA  int `prototype:"A"`
	AAAA int `prototype:"-"`
}

func TestSameLabel(t *testing.T) {
	src := testSameLabel{
		A:    10,
		AA:   20,
		AAA:  30,
		AAAA: 40,
	}
	var tgt testSameLabel
	err := prototype.Clone(&tgt, src, prototype.TagKey("prototype"))
	assert.NoError(t, err)
	assert.EqualValues(t, testSameLabel{
		A:    10,
		AA:   20,
		AAA:  30,
		AAAA: 0,
	}, tgt)
}

type testSameLabelNestedC struct {
	A int
	B int
	C int
}

type testSameLabelNestedA struct {
	A int
	B int
}

type testSameLabelNestedB struct {
	testSameLabelNestedC
	A int
	B int
}

type testSameLabelParent struct {
	testSameLabelNestedA
	testSameLabelNestedB
	A int
}

func TestSameLabelNested(t *testing.T) {
	src := testSameLabelParent{
		testSameLabelNestedA: testSameLabelNestedA{
			A: 2,
			B: 3,
		},
		testSameLabelNestedB: testSameLabelNestedB{
			testSameLabelNestedC: testSameLabelNestedC{
				A: 6,
				B: 7,
				C: 8,
			},
			A: 4,
			B: 5,
		},
		A: 1,
	}

	var tgt testSameLabelParent
	err := prototype.Clone(&tgt, src)
	assert.NoError(t, err)
	assert.EqualValues(t, testSameLabelParent{
		testSameLabelNestedA: testSameLabelNestedA{
			A: 2,
			B: 3,
		},
		testSameLabelNestedB: testSameLabelNestedB{
			testSameLabelNestedC: testSameLabelNestedC{
				A: 6,
				B: 7,
				C: 8,
			},
			A: 4,
			B: 5,
		},
		A: 1,
	}, tgt)
}

type testGetterStructParent struct {
	testGetter
}

type testSetterStructParent struct {
	testSetter
}

func TestNestedGetSetterStruct(t *testing.T) {
	id := uuid.NewString()
	name := "prototype"
	age := 30
	address := "shanghai"
	src := testGetterStructParent{
		testGetter: testGetter{
			id:      id,
			name:    name,
			Age:     age,
			Address: address,
		},
	}
	var tgt testSetterStructParent
	err := prototype.Clone(&tgt, src, prototype.TagKey("prototype"), prototype.SetterPrefix("Set"))
	assert.NoError(t, err)
	assert.EqualValues(t, testSetterStructParent{
		testSetter{
			Id:      "id:" + id,
			Name:    "name:" + name,
			age:     age * 2,
			address: "china-" + address,
		},
	}, tgt)
}

type testGetterPointerParent struct {
	*testGetter
}

type testSetterPointerParent struct {
	*testSetter
}

func TestNestedGetSetterPointer(t *testing.T) {
	id := uuid.NewString()
	name := "prototype"
	age := 30
	address := "shanghai"
	src := testGetterPointerParent{
		testGetter: &testGetter{
			id:      id,
			name:    name,
			Age:     age,
			Address: address,
		},
	}
	var tgt testSetterPointerParent
	err := prototype.Clone(&tgt, src, prototype.TagKey("prototype"), prototype.SetterPrefix("Set"))
	assert.NoError(t, err)

	tgt = testSetterPointerParent{testSetter: new(testSetter)}
	err = prototype.Clone(&tgt, src, prototype.TagKey("prototype"), prototype.SetterPrefix("Set"))
	assert.NoError(t, err)
	assert.EqualValues(t, testSetterPointerParent{
		testSetter: &testSetter{
			Id:      "id:" + id,
			Name:    "name:" + name,
			age:     age * 2,
			address: "china-" + address,
		},
	}, tgt)
}
