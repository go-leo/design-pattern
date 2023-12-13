package prototype

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"math"
	"reflect"
	"strconv"
	"testing"
)

type ScannerStruct struct {
	Nil    string
	Bool   bool
	Int    int64
	Uint   uint64
	Float  float64
	String string
}

func (s *ScannerStruct) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		s.Nil = "nil"
	case bool:
		s.Bool = v
	case int64:
		s.Int = v
	case uint64:
		s.Uint = v
	case float64:
		s.Float = v
	case string:
		s.String = v
	}
	return nil
}

type ScannerString string

func (s *ScannerString) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*s = "nil"
	case bool:
		*s = ScannerString(strconv.FormatBool(v))
	case int64:
		*s = ScannerString(strconv.FormatInt(v, 10))
	case uint64:
		*s = ScannerString(strconv.FormatUint(v, 10))
	case float64:
		*s = ScannerString(strconv.FormatFloat(v, 'g', -1, 64))
	case string:
		*s = ScannerString(v)
	}
	return nil
}

func TestEmptyValueCloner(t *testing.T) {
	var err error

	var tgtInt int
	opts := &options{
		ValueHook:    make(map[reflect.Value]map[reflect.Value]Hook),
		TypeHooks:    make(map[reflect.Type]map[reflect.Type]Hook),
		KindHooks:    make(map[reflect.Kind]map[reflect.Kind]Hook),
		SourceTagKey: "",
	}

	var src any

	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(src), opts)
	assert.NoError(t, err)
	assert.Equal(t, 0, tgtInt)

	var tgtStr string
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtStr), reflect.ValueOf(src), opts)
	assert.NoError(t, err)
	assert.Equal(t, "", tgtStr)

	var tgtErr error
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtErr), reflect.ValueOf(src), opts)
	assert.NoError(t, err)
	assert.Equal(t, nil, tgtErr)

	var tgtMap map[string]any
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtMap), reflect.Value{}, new(options))
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}(nil), tgtMap)

	var tgtSlice []int
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtSlice), reflect.Value{}, new(options))
	assert.NoError(t, err)
	assert.Equal(t, []int(nil), tgtSlice)

	var tgtIScannerStruct sql.Scanner = new(ScannerStruct)
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerStruct), reflect.ValueOf(src), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Nil: "nil"}, tgtIScannerStruct)

	var tgtIScannerString sql.Scanner = new(ScannerString)
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerString), reflect.ValueOf(src), opts)
	assert.NoError(t, err)
	scannerString := ScannerString("nil")
	assert.Equal(t, &scannerString, tgtIScannerString)

	var tgtScannerStruct = ScannerStruct{}
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStruct), reflect.ValueOf(src), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerStruct{Nil: "nil"}, tgtScannerStruct)

	var tgtScannerString = ScannerString("")
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerString), reflect.ValueOf(src), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerString("nil"), tgtScannerString)

	var tgtScannerStructPtr = &ScannerStruct{}
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStructPtr), reflect.ValueOf(src), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Nil: "nil"}, tgtScannerStructPtr)

	var tgtScannerStringPtr = new(ScannerString)
	err = emptyValueCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStringPtr), reflect.ValueOf(src), opts)
	assert.NoError(t, err)
	scannerString = "nil"
	assert.Equal(t, &scannerString, tgtScannerStringPtr)

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

	var tgtIScannerStruct sql.Scanner = new(ScannerStruct)
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerStruct), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Bool: srcBool}, tgtIScannerStruct)

	var tgtIScannerString sql.Scanner = new(ScannerString)
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerString), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	scannerString := ScannerString(strconv.FormatBool(srcBool))
	assert.Equal(t, &scannerString, tgtIScannerString)

	var tgtScannerStruct = ScannerStruct{}
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStruct), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerStruct{Bool: srcBool}, tgtScannerStruct)

	var tgtScannerString = ScannerString("")
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerString), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerString(strconv.FormatBool(srcBool)), tgtScannerString)

	var tgtScannerStructPtr = &ScannerStruct{}
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStructPtr), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Bool: srcBool}, tgtScannerStructPtr)

	var tgtScannerStringPtr = new(ScannerString)
	err = boolCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStringPtr), reflect.ValueOf(srcBool), opts)
	assert.NoError(t, err)
	scannerString = ScannerString(strconv.FormatBool(srcBool))
	assert.Equal(t, &scannerString, tgtScannerStringPtr)

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
	var overflowErr *OverflowError
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

	var tgtIScannerStruct sql.Scanner = new(ScannerStruct)
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerStruct), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Int: int64(srcInt)}, tgtIScannerStruct)

	var tgtIScannerString sql.Scanner = new(ScannerString)
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerString), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	scannerString := ScannerString(strconv.Itoa(srcInt))
	assert.Equal(t, &scannerString, tgtIScannerString)

	var tgtScannerStruct = ScannerStruct{}
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStruct), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerStruct{Int: int64(srcInt)}, tgtScannerStruct)

	var tgtScannerString = ScannerString("")
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerString), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerString(strconv.Itoa(srcInt)), tgtScannerString)

	var tgtScannerStructPtr = &ScannerStruct{}
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStructPtr), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Int: int64(srcInt)}, tgtScannerStructPtr)

	var tgtScannerStringPtr = new(ScannerString)
	err = intCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStringPtr), reflect.ValueOf(srcInt), opts)
	assert.NoError(t, err)
	scannerString = ScannerString(strconv.Itoa(srcInt))
	assert.Equal(t, &scannerString, tgtScannerStringPtr)

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
	var overflowErr *OverflowError
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

	var tgtIScannerStruct sql.Scanner = new(ScannerStruct)
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerStruct), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Uint: uint64(srcUint)}, tgtIScannerStruct)

	var tgtIScannerString sql.Scanner = new(ScannerString)
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerString), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	scannerString := ScannerString(strconv.FormatUint(uint64(srcUint), 10))
	assert.Equal(t, &scannerString, tgtIScannerString)

	var tgtScannerStruct = ScannerStruct{}
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStruct), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerStruct{Uint: uint64(srcUint)}, tgtScannerStruct)

	var tgtScannerString = ScannerString("")
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerString), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerString(strconv.FormatUint(uint64(srcUint), 10)), tgtScannerString)

	var tgtScannerStructPtr = &ScannerStruct{}
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStructPtr), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Uint: uint64(srcUint)}, tgtScannerStructPtr)

	var tgtScannerStringPtr = new(ScannerString)
	err = uintCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStringPtr), reflect.ValueOf(srcUint), opts)
	assert.NoError(t, err)
	scannerString = ScannerString(strconv.FormatUint(uint64(srcUint), 10))
	assert.Equal(t, &scannerString, tgtScannerStringPtr)
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
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtInt)

	srcFloat32 = 300.5
	var tgtInt8 uint8
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt8), reflect.ValueOf(srcFloat32), opts)
	var overflowErr *OverflowError
	assert.ErrorAs(t, err, &overflowErr)

	var tgtInt16 uint16
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt16), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtInt16)

	var tgtFloat32 float32
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat32), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtFloat32)

	var tgtAny any
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat32, tgtAny)

	var tgtIScannerStruct sql.Scanner = new(ScannerStruct)
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerStruct), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Float: float64(srcFloat32)}, tgtIScannerStruct)

	var tgtIScannerString sql.Scanner = new(ScannerString)
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerString), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	scannerString := ScannerString(strconv.FormatFloat(float64(srcFloat32), 'g', -1, 32))
	assert.Equal(t, &scannerString, tgtIScannerString)

	var tgtScannerStruct = ScannerStruct{}
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStruct), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerStruct{Float: float64(srcFloat32)}, tgtScannerStruct)

	var tgtScannerString = ScannerString("")
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerString), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerString(strconv.FormatFloat(float64(srcFloat32), 'g', -1, 32)), tgtScannerString)

	var tgtScannerStructPtr = &ScannerStruct{}
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStructPtr), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Float: float64(srcFloat32)}, tgtScannerStructPtr)

	var tgtScannerStringPtr = new(ScannerString)
	err = float32Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStringPtr), reflect.ValueOf(srcFloat32), opts)
	assert.NoError(t, err)
	scannerString = ScannerString(strconv.FormatFloat(float64(srcFloat32), 'g', -1, 32))
	assert.Equal(t, &scannerString, tgtScannerStringPtr)
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
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtInt)

	srcFloat64 = 300.5
	var tgtInt8 uint8
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt8), reflect.ValueOf(srcFloat64), opts)
	var overflowErr *OverflowError
	assert.ErrorAs(t, err, &overflowErr)

	var tgtInt16 uint16
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtInt16), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtInt16)

	var tgtFloat32 float32
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtFloat32), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtFloat32)

	var tgtAny any
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtAny), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.EqualValues(t, srcFloat64, tgtAny)

	var tgtIScannerStruct sql.Scanner = new(ScannerStruct)
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerStruct), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Float: srcFloat64}, tgtIScannerStruct)

	var tgtIScannerString sql.Scanner = new(ScannerString)
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerString), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	scannerString := ScannerString(strconv.FormatFloat(srcFloat64, 'g', -1, 64))
	assert.Equal(t, &scannerString, tgtIScannerString)

	var tgtScannerStruct = ScannerStruct{}
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStruct), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerStruct{Float: float64(srcFloat64)}, tgtScannerStruct)

	var tgtScannerString = ScannerString("")
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerString), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerString(strconv.FormatFloat(float64(srcFloat64), 'g', -1, 64)), tgtScannerString)

	var tgtScannerStructPtr = &ScannerStruct{}
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStructPtr), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{Float: float64(srcFloat64)}, tgtScannerStructPtr)

	var tgtScannerStringPtr = new(ScannerString)
	err = float64Cloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStringPtr), reflect.ValueOf(srcFloat64), opts)
	assert.NoError(t, err)
	scannerString = ScannerString(strconv.FormatFloat(float64(srcFloat64), 'g', -1, 64))
	assert.Equal(t, &scannerString, tgtScannerStringPtr)
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

	var tgtIScannerStruct sql.Scanner = new(ScannerStruct)
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerStruct), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{String: srcString}, tgtIScannerStruct)

	var tgtIScannerString sql.Scanner = new(ScannerString)
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtIScannerString), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	scannerString := ScannerString(srcString)
	assert.Equal(t, &scannerString, tgtIScannerString)

	var tgtScannerStruct = ScannerStruct{}
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStruct), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerStruct{String: srcString}, tgtScannerStruct)

	var tgtScannerString = ScannerString("")
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerString), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.Equal(t, ScannerString(srcString), tgtScannerString)

	var tgtScannerStructPtr = &ScannerStruct{}
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStructPtr), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	assert.Equal(t, &ScannerStruct{String: srcString}, tgtScannerStructPtr)

	var tgtScannerStringPtr = new(ScannerString)
	err = stringCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtScannerStringPtr), reflect.ValueOf(srcString), opts)
	assert.NoError(t, err)
	scannerString = ScannerString(srcString)
	assert.Equal(t, &scannerString, tgtScannerStringPtr)
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
	var overflowErr *OverflowError
	assert.ErrorAs(t, err, &overflowErr)

	_ = srcSqlNullInt64.Scan(math.MinInt64)
	var tgtUint64 uint64
	err = structCloner(new(cloneContext), []string{}, reflect.ValueOf(&tgtUint64), reflect.ValueOf(srcSqlNullInt64), opts)
	var negativeErr *NegativeNumberError
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
	var ofErr *OverflowError
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
	var ofErr *OverflowError
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

// TODO TestInterfaceCloner
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
