package prototype

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func stringify(v reflect.Value) (string, error) {
	if tm, ok := v.Interface().(encoding.TextMarshaler); ok {
		if v.Kind() == reflect.Pointer && v.IsNil() {
			return "", nil
		}
		buf, err := tm.MarshalText()
		return string(buf), err
	}
	if str, ok := v.Interface().(fmt.Stringer); ok {
		if v.Kind() == reflect.Pointer && v.IsNil() {
			return "", nil
		}
		return str.String(), nil
	}
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Interface, reflect.Pointer:
		return stringify(v.Elem())
	default:
		return "", errors.New("unexpected map key type")
	}
}

func indirect(v reflect.Value) (ClonerFrom, reflect.Value) {
	if v.CanAddr() {
		v = v.Addr()
	}
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

func indirectValue(v reflect.Value) reflect.Value {
	for (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) && !v.IsNil() {
		v = v.Elem()
	}
	return v
}
