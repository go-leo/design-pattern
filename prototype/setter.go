package prototype

import (
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

func setAny(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, tv reflect.Value, i any) error {
	if tv.NumMethod() == 0 {
		tv.Set(reflect.ValueOf(i))
		return nil
	}
	return hookCloner(e, fks, tgtVal, srcVal, opts)
}

func setStruct(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, v any) error {
	switch v := v.(type) {
	case bool:
	case int64:
	case uint:
	case float64:
	case string:
	case []byte:

	case time.Time:
		if tgtVal.Type() == timeType {
			tgtVal.Set(reflect.ValueOf(v))
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
	return nil
}
