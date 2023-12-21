package prototype

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
	"reflect"
	"strconv"
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
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error
	var tgtClonerFrom testClonerFromString

	srcBool := true
	tgtClonerFrom = ""
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtClonerFrom), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatBool(srcBool), tgtClonerFrom)

	var srcInt = math.MaxInt
	tgtClonerFrom = ""
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtClonerFrom), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatInt(int64(srcInt), 10), tgtClonerFrom)

	var srcUint uint = math.MaxUint
	tgtClonerFrom = ""
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtClonerFrom), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatUint(uint64(srcUint), 10), string(tgtClonerFrom))

	var srcFloat32 float32 = math.MaxFloat32
	tgtClonerFrom = ""
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtClonerFrom), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatFloat(float64(srcFloat32), 'g', -1, 64), tgtClonerFrom)

	var srcFloat64 = math.MaxFloat64
	tgtClonerFrom = ""
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtClonerFrom), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatFloat(srcFloat64, 'g', -1, 64), tgtClonerFrom)

	var srcString = "hello prototype"
	tgtClonerFrom = ""
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtClonerFrom), reflect.ValueOf(srcString), opts)
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
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error
	var srcClonerTo testClonerToString

	srcClonerTo = "true"
	srcClonerToVal := reflect.ValueOf(&srcClonerTo)
	cloner := valueCloner(srcClonerToVal, opts)
	var tgtBool bool
	err = cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBool), srcClonerToVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, strconv.FormatBool(tgtBool))

	srcClonerTo = testClonerToString(strconv.FormatInt(math.MaxInt64, 10))
	srcClonerToVal = reflect.ValueOf(&srcClonerTo)
	cloner = valueCloner(srcClonerToVal, opts)
	var tgtInt64 int64
	err = cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt64), srcClonerToVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, strconv.FormatInt(tgtInt64, 10))

	srcClonerTo = testClonerToString(strconv.FormatUint(math.MaxUint, 10))
	srcClonerToVal = reflect.ValueOf(&srcClonerTo)
	cloner = valueCloner(srcClonerToVal, opts)
	var tgtUint64 uint64
	err = cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtUint64), srcClonerToVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, strconv.FormatUint(tgtUint64, 10))

	srcClonerTo = testClonerToString(strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64))
	srcClonerToVal = reflect.ValueOf(&srcClonerTo)
	cloner = valueCloner(srcClonerToVal, opts)
	var tgtFloat64 float64
	err = cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat64), srcClonerToVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, strconv.FormatFloat(tgtFloat64, 'f', -1, 64))

	srcClonerTo = "hello prototype"
	srcClonerToVal = reflect.ValueOf(&srcClonerTo)
	cloner = valueCloner(srcClonerToVal, opts)
	var tgtString string
	err = cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtString), srcClonerToVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcClonerTo, tgtString)

	srcClonerToStruct := testClonerToStruct{S: srcClonerTo}
	srcClonerToStructVal := reflect.ValueOf(&srcClonerToStruct)
	cloner = valueCloner(srcClonerToStructVal, opts)
	var tgtClonerToStruct testClonerToStruct
	err = cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtClonerToStruct), srcClonerToStructVal, opts)

}

func TestBoolCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcBool bool
	var tgtBool bool

	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBool), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	assert.Equal(t, srcBool, tgtBool)

	srcBool = true
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBool), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	assert.Equal(t, srcBool, tgtBool)

	var srcAny any
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&srcAny), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	assert.Equal(t, srcAny, tgtBool)

	var tgtErr error
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtErr), reflect.ValueOf(srcBool), opts)
	var utErr *UnsupportedTypeError
	assert.ErrorAs(t, err, &utErr)
	assert.Equal(t, nil, tgtErr)

}

func TestIntCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcInt = 1
	var tgtInt int

	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, tgtInt)

	srcInt = 300
	var tgtInt8 uint8
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt8), reflect.ValueOf(srcInt), opts)
	var overflowErr Error
	assert.ErrorAs(t, err, &overflowErr)

	var tgtInt16 uint16
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt16), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, tgtInt16)

	var tgtFloat32 float32
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat32), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, tgtFloat32)

	var tgtAny any
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcInt, tgtAny)
}

func TestUIntCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcUint uint = 1

	var tgtInt int
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, tgtInt)

	srcUint = 300
	var tgtInt8 uint8
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt8), reflect.ValueOf(srcUint), opts)
	var overflowErr Error
	assert.ErrorAs(t, err, &overflowErr)

	var tgtInt16 uint16
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt16), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, tgtInt16)

	var tgtFloat32 float32
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat32), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, tgtFloat32)

	var tgtAny any
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcUint, tgtAny)
}

func TestFloat32Cloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcFloat32 float32 = 1.1

	var tgtInt int
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtInt)

	srcFloat32 = 300.5
	var tgtInt8 uint8
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt8), reflect.ValueOf(srcFloat32), opts)
	var overflowErr Error
	assert.ErrorAs(t, err, &overflowErr)

	var tgtInt16 uint16
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt16), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtInt16)

	var tgtFloat32 float32
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat32), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtFloat32)

	var tgtAny any
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtAny)
}

func TestFloat64Cloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcFloat64 = 120.4

	var tgtInt int
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtInt)

	srcFloat64 = 300.5
	var tgtInt8 uint8
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt8), reflect.ValueOf(srcFloat64), opts)
	var overflowErr Error
	assert.ErrorAs(t, err, &overflowErr)

	var tgtInt16 uint16
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt16), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtInt16)

	var tgtFloat32 float32
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat32), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtFloat32)

	var tgtAny any
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtAny)
}

func TestStringCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcString = "120.4"

	var tgtString string
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtString), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcString, tgtString)

	var tgtBytes []byte
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBytes), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcString, tgtBytes)

	var tgtAny any
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcString, tgtAny)

	srcString = "true"
	var tgtBool bool
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBool), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, true, tgtBool)

	srcString = "-1000000000"
	var tgtInt int
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, -1000000000, tgtInt)

	srcString = "1000000000"
	var tgtUint uint
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtUint), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, 1000000000, tgtUint)

	srcString = "3.1415836"
	var tgtFloat64 float64
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat64), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, 3.1415836)

	srcString = "3.1415836"
	var tgtStrPtr *string
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStrPtr), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, 3.1415836)
}

func TestBytesCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcBytes = []byte("120.4")

	var tgtBytes []byte
	err = bytesCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBytes), reflect.ValueOf(srcBytes), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcBytes, tgtBytes)

	var tgtString string
	err = bytesCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtString), reflect.ValueOf(srcBytes), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcBytes, tgtString)

	var tgtAny any
	err = bytesCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcBytes), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcBytes, tgtAny)

	srcBytes = []byte("true")
	var tgtBool bool
	err = bytesCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBool), reflect.ValueOf(srcBytes), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, true, tgtBool)

	srcBytes = []byte("-1000000000")
	var tgtInt int
	err = bytesCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(srcBytes), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, -1000000000, tgtInt)

	srcBytes = []byte("1000000000")
	var tgtUint uint
	err = bytesCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtUint), reflect.ValueOf(srcBytes), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, 1000000000, tgtUint)

	srcBytes = []byte("3.1415836")
	var tgtFloat64 float64
	err = bytesCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat64), reflect.ValueOf(srcBytes), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, 3.1415836)

	srcBytes = []byte("3.1415836")
	var tgtStrPtr *string
	err = bytesCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStrPtr), reflect.ValueOf(srcBytes), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, 3.1415836)
}

func TestTimeCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
		TargetTagKey: "",
		DeepClone:    false,
		NameComparer: nil,
		UnixTime: func(t time.Time) int64 {
			return t.Unix()
		},
	}

	var err error

	var srcTime = time.Now()

	var tgtStruct time.Time
	err = timeCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStruct), reflect.ValueOf(srcTime), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtStruct, srcTime)

	var tgtAny any
	err = timeCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcTime), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtAny, srcTime)

	var tgtString string
	err = timeCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtString), reflect.ValueOf(srcTime), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtString, srcTime.Format(time.RFC3339))

	var tgtInt int
	err = timeCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(srcTime), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtInt, srcTime.Unix())

	var tgtUint uint
	err = timeCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtUint), reflect.ValueOf(srcTime), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtUint, srcTime.Unix())

	var tgtFloat32 float32
	err = timeCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat32), reflect.ValueOf(srcTime), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat32, float32(srcTime.Unix()))

	var tgtPtr *time.Time
	err = timeCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtPtr), reflect.ValueOf(srcTime), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, *tgtPtr, srcTime)

	var tgtPtrPtr **time.Time
	err = timeCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtPtrPtr), reflect.ValueOf(srcTime), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, **tgtPtrPtr, srcTime)

}

func TestStructCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcSqlNullBool sql.NullBool
	_ = srcSqlNullBool.Scan(true)
	var tgtBool bool
	err = structCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBool), reflect.ValueOf(srcSqlNullBool), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtBool, srcSqlNullBool.Bool)

	var srcSqlNullInt64 sql.NullInt64
	_ = srcSqlNullInt64.Scan(math.MaxInt64)
	var tgtInt32 int32
	err = structCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt32), reflect.ValueOf(srcSqlNullInt64), opts)
	var overflowErr Error
	assert.ErrorAs(t, err, &overflowErr)

	_ = srcSqlNullInt64.Scan(math.MinInt64)
	var tgtUint64 uint64
	err = structCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtUint64), reflect.ValueOf(srcSqlNullInt64), opts)
	var negativeErr Error
	assert.ErrorAs(t, err, &negativeErr)

	var tgtInt64 int64
	err = structCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt64), reflect.ValueOf(srcSqlNullInt64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtInt64, srcSqlNullInt64.Int64)

	var srcSqlNullByte sql.NullByte
	_ = srcSqlNullByte.Scan(math.MaxUint8)
	var tgtUint8 uint8
	err = structCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtUint8), reflect.ValueOf(srcSqlNullByte), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtUint8, srcSqlNullByte.Byte)

	var srcSqlNullFloat sql.NullFloat64
	_ = srcSqlNullFloat.Scan(math.MaxFloat64)
	err = structCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt64), reflect.ValueOf(srcSqlNullFloat), opts)
	assert.ErrorAs(t, err, &overflowErr)

	var tgtFloat32 float32
	err = structCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat32), reflect.ValueOf(srcSqlNullFloat), opts)
	assert.ErrorAs(t, err, &overflowErr)

	var tgtFloat64 float64
	err = structCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat64), reflect.ValueOf(srcSqlNullFloat), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtFloat64, srcSqlNullFloat.Float64)

	now := time.Now()
	srcTimestampPB := timestamppb.New(now)
	var tgtTime time.Time
	err = pointerCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtTime), reflect.ValueOf(srcTimestampPB), opts)
	assert.EqualValues(t, now.Local().Format(time.DateTime), tgtTime.Local().Format(time.DateTime))

	hour := time.Hour
	srcDurationPB := durationpb.New(hour)
	var tgtDuration time.Duration
	err = pointerCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtDuration), reflect.ValueOf(srcDurationPB), opts)
	assert.EqualValues(t, hour, tgtDuration)

	type Error struct {
		Course string
		Msg    string
	}

	type Code struct {
		*Error
		Msg  string
		Code int
		//Details []string
	}

	type Int int

	type Response struct {
		Code
		Int
		Msg      string
		Username string
		IsPass   bool
	}

	var (
		srcResponse = Response{
			Code: Code{
				Error: &Error{
					Course: "Course",
					Msg:    "error msg",
				},
				Msg:  "ok",
				Code: 200,
				//Details: []string{"fast", "slow"},
			},
			Msg:      "success",
			Username: "Forest",
			IsPass:   false,
		}
	)
	srcVal := reflect.ValueOf(srcResponse)

	var tgtResp Response
	tgtStructVal := reflect.ValueOf(&tgtResp)
	err = structCloner(new(cloneContext), []string{}, tgtStructVal, srcVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, &tgtResp, &srcResponse)

	var tgtAny any
	tgtAnyVal := reflect.ValueOf(&tgtAny)
	err = structCloner(new(cloneContext), []string{}, tgtAnyVal, srcVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtAny, map[string]any{
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
	})

	var tgtMap map[string]any
	tgtMapVal := reflect.ValueOf(&tgtMap)
	err = structCloner(new(cloneContext), []string{}, tgtMapVal, srcVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtAny, map[string]any{
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
	})
}

func TestMapCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error
	var tgtVal reflect.Value
	var srcVal reflect.Value

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
	}
	var tgtMap map[string]any
	tgtVal = reflect.ValueOf(&tgtMap)
	srcVal = reflect.ValueOf(srcMap)
	err = mapCloner(new(cloneContext), []string{}, tgtVal, srcVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtMap, srcMap)

	var tgtAny any
	tgtVal = reflect.ValueOf(&tgtAny)
	srcVal = reflect.ValueOf(srcMap)
	err = mapCloner(new(cloneContext), []string{}, tgtVal, srcVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtMap, srcMap)

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
	}
	var tgtStruct Response
	tgtVal = reflect.ValueOf(&tgtStruct)
	srcVal = reflect.ValueOf(srcMap)
	err = mapCloner(new(cloneContext), []string{}, tgtVal, srcVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtStruct, Response{
		Code: Code{
			Error: &Error{
				Course: "Course",
				Msg:    "error msg",
			},
			Msg:  "ok",
			Code: 200,
		},
		Msg:      "success",
		Username: "Forest",
		IsPass:   false,
	})

	var tgtStructPointer *Response
	tgtVal = reflect.ValueOf(&tgtStructPointer)
	srcVal = reflect.ValueOf(srcMap)
	err = mapCloner(new(cloneContext), []string{}, tgtVal, srcVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, *tgtStructPointer, Response{
		Code: Code{
			Error: &Error{
				Course: "Course",
				Msg:    "error msg",
			},
			Msg:  "ok",
			Code: 200,
		},
		Msg:      "success",
		Username: "Forest",
		IsPass:   false,
	})
}

func TestSliceCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcBytesSlice = []byte{'1', '2', 'a', 'b'}
	var tgtString string
	tgtStringVal := reflect.ValueOf(&tgtString)
	srcBytesSliceVal := reflect.ValueOf(srcBytesSlice)
	err = sliceCloner(new(cloneContext), []string{}, tgtStringVal, srcBytesSliceVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, "12ab", tgtString)

	var srcInt16Slice = []int16{math.MinInt16, math.MinInt8, math.MaxInt8, math.MaxInt16}
	var tgtInt8Slice []int8
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt8Slice), reflect.ValueOf(srcInt16Slice), opts)
	var ofErr Error
	assert.ErrorAs(t, err, &ofErr)

	var tgtInt32Slice []int32
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt32Slice), reflect.ValueOf(srcInt16Slice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtInt32Slice, srcInt16Slice)

	var srcStringSlice = []string{"1", "2", "3", "4"}
	var tgtStringSlice []string
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStringSlice), reflect.ValueOf(srcStringSlice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtStringSlice, srcStringSlice)

	var tgtStringArray [10]string
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStringArray), reflect.ValueOf(srcStringSlice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtStringArray[:4])

	var tgtAny any
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcStringSlice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtAny)

	var tgtStringArrayPtr *[10]string
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStringArrayPtr), reflect.ValueOf(srcStringSlice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtStringArrayPtr[:4])

}

func TestArrayCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcInt16Slice = [...]int16{math.MinInt16, math.MinInt8, math.MaxInt8, math.MaxInt16}
	var tgtInt32Slice [3]int32
	err = arrayCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt32Slice), reflect.ValueOf(srcInt16Slice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcInt16Slice[:3], tgtInt32Slice[:])

	var tgtInt64Slice [6]int64
	err = arrayCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt64Slice), reflect.ValueOf(srcInt16Slice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcInt16Slice[:], tgtInt64Slice[:4])

	var srcStringSlice = [...]string{"1", "2", "3", "4"}
	var tgtStringSlice []string
	err = arrayCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStringSlice), reflect.ValueOf(srcStringSlice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtStringSlice, srcStringSlice)

	var tgtAny any
	err = arrayCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcStringSlice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtAny)

	var tgtStringArrayPtr *[10]string
	err = arrayCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStringArrayPtr), reflect.ValueOf(srcStringSlice), opts)
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(srcStringSlice, tgtStringArrayPtr[:4])

}

func TestInterfaceCloner(t *testing.T) {
	var err error

	var srcString = "120.4"

	var tgtString string
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtString), reflect.ValueOf(srcString), new(options))
	assert.NoError(t, err)
	assert.EqualValues(t, tgtString, srcString)

	var tgtBytes []byte
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBytes), reflect.ValueOf(srcString), new(options))
	assert.NoError(t, err)
	assert.EqualValues(t, tgtBytes, srcString)

	var tgtAny any
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcString), new(options))
	assert.NoError(t, err)
	assert.EqualValues(t, tgtAny, srcString)
}

func TestDuration(t *testing.T) {
	src := time.Hour
	var tgt time.Duration
	err := Clone(&tgt, src)
	assert.NoError(t, err)
	assert.EqualValues(t, src, tgt)
}

//
//type User struct {
//	Name     string
//	Birthday *time.Time
//	Nickname string
//	Role     string
//	Age      int32
//	FakeAge  *int32
//	Notes    []string
//	flags    []byte
//}
//
//type Employee struct {
//	_User     *User
//	Name      string
//	Birthday  *time.Time
//	NickName  *string
//	Age       int64
//	FakeAge   int
//	EmployeID int64
//	DoubleAge int32
//	SuperRule string
//	Notes     []*string
//	flags     []byte
//}
//
//func TestCopyStruct(t *testing.T) {
//	var fakeAge int32 = 12
//	user := User{
//		Name:     "Jinzhu",
//		Nickname: "jinzhu",
//		Age:      18,
//		FakeAge:  &fakeAge,
//		Role:     "Admin",
//		Notes:    []string{"hello world", "welcome"},
//		flags:    []byte{'x'},
//	}
//	employee := Employee{}
//
//	if err := Clone(employee, &user); err == nil {
//		t.Errorf("Copy to unaddressable value should get error")
//	}
//
//	Clone(&employee, &user)
//	checkEmployee(employee, user, t, "Copy From Ptr To Ptr")
//
//	employee2 := Employee{}
//	Clone(&employee2, user)
//	checkEmployee(employee2, user, t, "Copy From Struct To Ptr")
//
//	employee3 := Employee{}
//	ptrToUser := &user
//	Clone(&employee3, &ptrToUser)
//	checkEmployee(employee3, user, t, "Copy From Double Ptr To Ptr")
//
//	employee4 := &Employee{}
//	Clone(&employee4, user)
//	checkEmployee(*employee4, user, t, "Copy From Ptr To Double Ptr")
//
//	employee5 := &Employee{}
//	Clone(&employee5, &employee)
//	checkEmployee(*employee5, user, t, "Copy From Employee To Employee")
//}
//
//func checkEmployee(employee Employee, user User, t *testing.T, testCase string) {
//	if employee.Name != user.Name {
//		t.Errorf("%v: Name haven't been copied correctly.", testCase)
//	}
//	if employee.NickName == nil || *employee.NickName != user.Nickname {
//		t.Errorf("%v: NickName haven't been copied correctly.", testCase)
//	}
//	if employee.Birthday == nil && user.Birthday != nil {
//		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
//	}
//	if employee.Birthday != nil && user.Birthday == nil {
//		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
//	}
//	if employee.Birthday != nil && user.Birthday != nil &&
//		!employee.Birthday.Equal(*(user.Birthday)) {
//		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
//	}
//	if employee.Age != int64(user.Age) {
//		t.Errorf("%v: Age haven't been copied correctly.", testCase)
//	}
//	if user.FakeAge != nil && employee.FakeAge != int(*user.FakeAge) {
//		t.Errorf("%v: FakeAge haven't been copied correctly.", testCase)
//	}
//
//	if len(employee.Notes) != len(user.Notes) {
//		t.Fatalf("%v: Copy from slice doesn't work, employee notes len: %v, user: %v", testCase, len(employee.Notes), len(user.Notes))
//	}
//
//	for idx, note := range user.Notes {
//		if note != *employee.Notes[idx] {
//			t.Fatalf("%v: Copy from slice doesn't work, notes idx: %v employee: %v user: %v", testCase, idx, *employee.Notes[idx], note)
//		}
//	}
//	if employee.SuperRule != "Super "+user.Role {
//		t.Errorf("%v: Copy to method doesn't work", testCase)
//	}
//}