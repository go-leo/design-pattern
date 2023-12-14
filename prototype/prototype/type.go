package prototype

import (
	"database/sql"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
)

var (
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	stringerType      = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	numberType          = reflect.TypeOf(json.Number(""))

	scannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

	sqlNullBoolType    = reflect.TypeOf(sql.NullBool{})
	sqlNullInt16Type   = reflect.TypeOf(sql.NullInt16{})
	sqlNullInt32Type   = reflect.TypeOf(sql.NullInt32{})
	sqlNullInt64Type   = reflect.TypeOf(sql.NullInt64{})
	sqlNullFloat64Type = reflect.TypeOf(sql.NullFloat64{})
	sqlNullStringType  = reflect.TypeOf(sql.NullString{})
	sqlNullByteType    = reflect.TypeOf(sql.NullByte{})
	sqlNullTimeType    = reflect.TypeOf(sql.NullTime{})
)

func indirectValue(v reflect.Value) (sql.Scanner, reflect.Value) {
	for {
		if v.Type().NumMethod() > 0 && v.CanInterface() {
			if scanner, ok := v.Interface().(sql.Scanner); ok {
				return scanner, v
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
