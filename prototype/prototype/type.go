package prototype

import (
	"encoding"
	"encoding/json"
	"reflect"
)

var (
	textMarshalerType   = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	numberType          = reflect.TypeOf(json.Number(""))
)
