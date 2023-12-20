package prototype

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
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
	var tgtIClonerFromString testClonerFromString

	srcBool := true
	tgtIClonerFromString = ""
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIClonerFromString), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatBool(srcBool), tgtIClonerFromString)

	var srcInt = math.MaxInt
	tgtIClonerFromString = ""
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIClonerFromString), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatInt(int64(srcInt), 10), tgtIClonerFromString)

	var srcUint uint = math.MaxUint
	tgtIClonerFromString = ""
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIClonerFromString), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatUint(uint64(srcUint), 10), string(tgtIClonerFromString))

	var srcFloat32 float32 = math.MaxFloat32
	tgtIClonerFromString = ""
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIClonerFromString), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatFloat(float64(srcFloat32), 'g', -1, 64), tgtIClonerFromString)

	var srcFloat64 = math.MaxFloat64
	tgtIClonerFromString = ""
	err = floatCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIClonerFromString), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, strconv.FormatFloat(srcFloat64, 'g', -1, 64), tgtIClonerFromString)

	var srcString = "hello prototype"
	tgtIClonerFromString = ""
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIClonerFromString), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcString, tgtIClonerFromString)

}

func TestClonerTo(t *testing.T) {

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
	assert.EqualValues(t, tgtString, srcString)

	var tgtBytes []byte
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtBytes), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtBytes, srcString)

	var tgtAny any
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtAny, srcString)
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

	var srcBoolSlice = []bool{true, false, false, true}
	var tgtBoolSlice []bool
	tgtBoolSliceVal := reflect.ValueOf(&tgtBoolSlice)
	srcBoolSliceVal := reflect.ValueOf(srcBoolSlice)
	err = sliceCloner(new(cloneContext), []string{}, tgtBoolSliceVal, srcBoolSliceVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtBoolSlice, srcBoolSlice)

	var srcInt16Slice = []int16{math.MinInt16, math.MinInt8, math.MaxInt8, math.MaxInt16}
	var tgtInt8Slice []int8
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt8Slice), reflect.ValueOf(srcInt16Slice), new(options))
	var ofErr Error
	assert.ErrorAs(t, err, &ofErr)

	var tgtInt32Slice []int32
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt32Slice), reflect.ValueOf(srcInt16Slice), new(options))
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtInt32Slice, srcInt16Slice)

	var srcUint16Slice = []uint16{0, math.MaxInt8, math.MaxInt16}
	var tgtUint32Slice []uint32
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtUint32Slice), reflect.ValueOf(srcUint16Slice), new(options))
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtUint32Slice, srcUint16Slice)

	var srcStringSlice = []string{"1", "2", "3", "4"}
	var tgtStringSlice []string
	err = sliceCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStringSlice), reflect.ValueOf(srcStringSlice), new(options))
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtStringSlice, srcStringSlice)

}

func TestArrayCloner(t *testing.T) {
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var err error

	var srcBoolArray = [4]bool{true, false, false, true}
	var tgtBoolArray [6]bool
	tgtBoolArrayVal := reflect.ValueOf(&tgtBoolArray)
	srcBoolArrayVal := reflect.ValueOf(srcBoolArray)
	err = arrayCloner(new(cloneContext), []string{}, tgtBoolArrayVal, srcBoolArrayVal, opts)
	assert.NoError(t, err)
	assert.EqualValues(t, tgtBoolArray[:4], srcBoolArray[:])

	var srcInt16Array = [8]int16{math.MinInt16, math.MinInt8, math.MaxInt8, math.MaxInt16}
	var tgtInt8Array [8]int8
	err = arrayCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt8Array), reflect.ValueOf(srcInt16Array), new(options))
	var ofErr Error
	assert.ErrorAs(t, err, &ofErr)

	var tgtInt32Array [2]int32
	err = arrayCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt32Array), reflect.ValueOf(srcInt16Array), new(options))
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtInt32Array, srcInt16Array[:2])

	var srcUint16Array = [3]uint16{0, math.MaxInt8, math.MaxInt16}
	var tgtUint32Array [3]uint32
	err = arrayCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtUint32Array), reflect.ValueOf(srcUint16Array), new(options))
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtUint32Array, srcUint16Array)

	var srcStringArray = [6]string{"1", "2", "3", "4"}
	var tgtStringArray [6]string
	err = arrayCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStringArray), reflect.ValueOf(srcStringArray), new(options))
	assert.NoError(t, err)
	assert.ObjectsAreEqualValues(tgtStringArray, srcStringArray)

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
